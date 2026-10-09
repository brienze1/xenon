---
name: Error Handler (Clean Architecture)
description: This skill should be used when the user asks to "create an error handler", "add error handling", "handle errors at the edge", "create custom errors", "add error codes", "implement error response formatting", or mentions anything about error handling in the integration/entrypoint/controller/ or pkg/errs/ packages.
version: 1.0.0
---

# Error Handler — Golang Clean Architecture

The error handler is a critical architectural component that lives at the edge of the application. It captures ALL errors before they reach the user/caller and transforms them into consistent, formatted responses. It acts as a controller for errors.

---

## Core Rules

### Rule 1: Location

The error handler MUST live in `internal/integration/entrypoint/controller/` alongside regular controllers.

```
internal/
└── integration/
    └── entrypoint/
        └── controller/
            ├── handler.go           ← main controller
            ├── error_handler.go     ← error handler
            └── ...
```

---

### Rule 2: Edge of the Application

The error handler sits at the edge — it catches every error that propagates up from controllers, services, use cases, and integration layers. It is the LAST stop before the error reaches the infrastructure layer (router/lambda/broker).

**For REST APIs (Fiber/Gin):**
The error handler works alongside a middleware in the infrastructure router. The middleware catches errors from controllers and redirects them to the error handler.

**For Lambdas/Jobs/Batches:**
The infrastructure handler checks if the controller returned an error (`err != nil`) and redirects it to the error handler.

**For both patterns:**
The controller itself uses a `defer` block to catch panics and route all errors through the error handler before returning.

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// Inside a controller — defer delegates errors to error handler
defer func(ctx context.Context, event any) {
    if recovered := recover(); recovered != nil {
        err = handlePanic(ctx, event, recovered)
    }
    if err != nil {
        err = h.errorHandler.Handle(ctx, err, event)
    }
}(ctx, event)
```

---

### Rule 3: Custom Error Type Detection

The error handler MUST check if the error is a known custom error type (`errs.BaseError`). If it is, the handler uses the custom error's methods to generate a formatted response. If the error is unknown, the handler wraps it as an internal server error.

```go
func (e *errorHandler) Handle(ctx context.Context, err error, event any) error {
    logger.Error(ctx, err, "Started")

    var customErr errs.BaseError
    if !errors.As(err, &customErr) {
        // Unknown error → transform into internal server error with consistent format
        customErr = errs.InternalServerError(err, "UNK-00500", "Unknown error")
    }

    // Now customErr has consistent methods: Code(), StatusCode(), ToDto(), etc.
    // ...
}
```

---

### Rule 4: Constructor Pattern

Error handler constructors follow the same pattern as controller constructors: `New{Name}(...)` returning an integration adapter. The struct is **always unexported** — exposed only via the constructor.

```go
type errorHandler struct {  // unexported — only accessible via NewErrorHandler
    // Can inject use cases for notifications (Slack, Discord, error queues, etc.)
    notifySlack usecase.SendNotification
}

func NewErrorHandler(notifySlack usecase.SendNotification) adapter.ErrorHandler {
    return &errorHandler{
        notifySlack: notifySlack,
    }
}
```

The error handler implements `adapter.ErrorHandler` from `internal/integration/adapter/`.

---

### Rule 5: Error Codes — Created at Origin (CRITICAL)

Error codes MUST be written at the exact line where the error is thrown, NOT in the error handler or anywhere else. This makes debugging trivial — search for the code and find the exact line.

```go
// GOOD — error code at the line where the error occurs (in persistence layer)
// File: internal/integration/persistence/user.go
func (p *userPersistence) FindById(ctx context.Context, userId string) (*entity.User, error) {
    var model UserModel
    err := p.db.Where("id = ?", userId).First(&model).Error
    if err != nil {
        return nil, errs.DbError(err)  // Code generated here: "NF-01404" or "IS-01500"
    }
    return model.ToEntity(), nil
}

// GOOD — error code at the line where the error occurs (in webservice layer)
// File: internal/integration/webservice/payment.go
func (w *paymentWebservice) Process(ctx context.Context, ...) error {
    resp, err := w.client.Post(url, body)
    if err != nil {
        return errs.InternalServerError(err, "PAY-00500", "Failed to call payment API")
    }
    if resp.StatusCode >= 400 && resp.StatusCode < 500 {
        return errs.ClientError("Payment rejected", "PAY-00400", resp.StatusCode)
    }
    if resp.StatusCode >= 500 {
        return errs.ServerError("Payment service error", "PAY-01500", resp.StatusCode)
    }
    return nil
}
```

```go
// BAD — error code created in the error handler (impossible to trace origin)
func (e *errorHandler) Handle(ctx context.Context, err error, event any) error {
    if strings.Contains(err.Error(), "payment") {
        return errs.InternalServerError(err, "PAY-00500", "payment error")  // WRONG
    }
}
```

---

### Rule 6: Error Codes Must Be Unique

Every error code MUST be unique across the entire application. This guarantees that searching for a code finds exactly one location.

**Code format convention:** `{PREFIX}-{NUMBER}{HTTP_STATUS}`

Examples:
- `DOW-00500` — Download error, internal server error
- `DOW-01500` — Different download error, same status
- `DOW-00400` — Download error, client error
- `PAY-00500` — Payment error, internal server error
- `NF-01404` — Not found error (from DB)
- `EV-00400` — Event validation error
- `UNK-00500` — Unknown error (only used by error handler as fallback)

```go
// Each error at its origin has a unique code
errs.InternalServerError(err, "DOW-00500", "Failed to create download request")
errs.InternalServerError(err, "DOW-01500", "Failed to download post")
errs.ClientError("Error downloading", "DOW-00400", response.StatusCode)
errs.ServerError("Error downloading", "DOW-02500", response.StatusCode)
errs.BadRequestError("media URL not found", "DOW-04400")
```

---

### Rule 7: Error Handler Can Import Use Cases

The error handler can depend on use case interfaces to send notifications (Slack, Discord, error queues, etc.) when errors meet certain criteria.

```go
type errorHandler struct {
    notifySlack usecase.SendNotification
}

func (e *errorHandler) Handle(ctx context.Context, err error, event any) error {
    logger.Error(ctx, err, "Started")

    var customErr errs.BaseError
    if !errors.As(err, &customErr) {
        customErr = errs.InternalServerError(err, "UNK-00500", "Unknown error")
    }

    // Notify external systems when retries are exhausted
    if ctx.GetReceiveCount() >= customErr.Retries() {
        if customErr.InternalNotify() {
            notifyErr := e.notifySlack.Send(ctx, customErr.Error(), event)
            if notifyErr != nil {
                logger.Error(ctx, notifyErr, "Error while trying to notify", customErr.Error(), event)
            }
        }
    }

    // ...
}
```

---

### Rule 8: Abort Behavior (Jobs/Lambdas Only)

For jobs, lambdas, and batch processors, the error handler can abort processing by returning `nil` when the error type declares it should abort. This prevents infinite retries for known unrecoverable errors.

```go
func (e *errorHandler) Handle(ctx context.Context, err error, event any) error {
    // ...

    if customErr.Abort() {
        // Return nil = job considers the message processed (no retry)
        return nil
    }

    // Return error = job will retry
    logger.Error(ctx, customErr, "Finished", event)
    return customErr
}
```

**Which errors abort?**
- `badRequestError` → Abort: `true` (retrying won't fix bad input)
- `notFoundError` → Abort: `true` (retrying won't create the missing resource)
- `clientError` → Abort: `true` (external client returned a definitive error)
- `internalServerError` → Abort: `false` (retry, might be transient)
- `serverError` → Abort: `false` (retry, might be transient)
- `panicError` → Abort: `false` (retry, might be a race condition)

**For REST APIs:** Abort is not relevant — every error returns an HTTP response regardless.

---

## Custom Error Types (`pkg/errs/`)

The `pkg/errs/` package defines the `BaseError` interface and concrete error types. All custom errors implement this interface.

### BaseError Interface

```go
type BaseError interface {
    Code() string                        // Unique error code (e.g., "DOW-00500")
    Error() string                       // Error message
    Description() string                 // User-facing description
    StatusCode() int                     // HTTP status code
    ErrorDetails() map[string]any        // Structured error details
    ToDto(transactionId string) Error    // Convert to response DTO
    Retries() int                        // Max retry count before escalation
    InternalNotify() bool                // Should notify internal systems?
    Abort() bool                         // Should abort processing? (jobs only)
    Is(target error) bool                // Error type matching
}
```

### Error Types Summary

| Type | Factory | Status | Retries | Abort | Notify | Use for |
|---|---|---|---|---|---|---|
| `badRequestError` | `errs.BadRequestError(desc, code, details...)` | 400 | 0 | Yes | Yes | Invalid input, validation failures |
| `notFoundError` | `errs.NotFoundError(msg, code)` | 404 | 0 | Yes | Yes | Resource not found |
| `clientError` | `errs.ClientError(desc, code, status, body...)` | varies | 0 | Yes | Yes | External API client errors (4xx) |
| `dbError` | `errs.DbError(err)` | varies | 5 | No | Yes | Database errors (auto-detects type) |
| `serverError` | `errs.ServerError(desc, code, status)` | varies | 5 | No | Yes | External API server errors (5xx) |
| `internalServerError` | `errs.InternalServerError(err, code, msg)` | 500 | 5 | No | Yes | Unexpected internal errors |
| `panicError` | `errs.PanicError(msg)` | 500 | 0 | No | Yes | Recovered panics |

### Helper Functions

```go
// Check if error is a specific type
errs.IsNotFoundError(err) // returns bool

// Database error auto-detection
errs.DbError(err)
// → If gorm.ErrRecordNotFound → returns NotFoundError
// → If unique constraint violation → returns BadRequestError
// → Otherwise → returns InternalServerError
```

---

## Code Examples

### Example 1 — Error handler for jobs/lambdas (with notifications)

```go
package controller

import (
    "context"
    "errors"

    "myproject/internal/application/usecase"
    "myproject/internal/integration/adapter"
    "myproject/pkg/errs"
    "myproject/pkg/logger"
)

type errorHandler struct {
    notifySlack usecase.SendNotification
}

func NewErrorHandler(notifySlack usecase.SendNotification) adapter.ErrorHandler {
    return &errorHandler{
        notifySlack: notifySlack,
    }
}

func (e *errorHandler) Handle(ctx context.Context, err error, event any) error {
    logger.Error(ctx, err, "Started")

    var customErr errs.BaseError
    if !errors.As(err, &customErr) {
        customErr = errs.InternalServerError(err, "UNK-00500", "Unknown error")
    }

    if logger.GetContext(ctx).GetReceiveCount() >= customErr.Retries() {
        if customErr.InternalNotify() {
            notifyErr := e.notifySlack.Send(ctx, customErr.Error(), event)
            if notifyErr != nil {
                logger.Error(ctx, notifyErr, "Error while trying to notify", customErr.Error(), event)
            }
        }

        if customErr.Abort() {
            return nil
        }
    }

    logger.Error(ctx, customErr, "Finished", event)
    return customErr
}
```

---

### Example 2 — Error handler for REST APIs

```go
package controller

import (
    "context"
    "errors"

    "myproject/internal/integration/adapter"
    "myproject/pkg/errs"
    "myproject/pkg/logger"
)

type errorHandler struct{}

func NewErrorHandler() adapter.ErrorHandler {
    return &errorHandler{}
}

func (e *errorHandler) Handle(ctx context.Context, err error, event any) error {
    logger.Error(ctx, err, "Started")

    var customErr errs.BaseError
    if !errors.As(err, &customErr) {
        customErr = errs.InternalServerError(err, "UNK-00500", "Unknown error")
    }

    // For REST APIs, always return the formatted error — no abort logic needed
    // The infrastructure router middleware will use customErr.StatusCode() and customErr.ToDto()
    // to build the HTTP response

    logger.Error(ctx, customErr, "Finished", event)
    return customErr
}
```

---

### Example 3 — Error codes at origin (persistence layer)

```go
// File: internal/integration/persistence/order.go
package persistence

import (
    "context"
    "myproject/internal/domain/entity"
    "myproject/pkg/errs"
)

func (p *orderPersistence) FindById(ctx context.Context, orderId int64) (*entity.Order, error) {
    var model OrderModel
    err := p.db.Where("id = ?", orderId).First(&model).Error
    if err != nil {
        return nil, errs.DbError(err)  // Auto-detects: NF-01404 or IS-01500
    }
    return model.ToEntity(), nil
}

func (p *orderPersistence) Save(ctx context.Context, order entity.Order) error {
    model := OrderModelFromEntity(order)
    err := p.db.Create(&model).Error
    if err != nil {
        return errs.DbError(err)  // Auto-detects: BR-01400 (duplicate) or IS-01500
    }
    return nil
}
```

---

### Example 4 — Error codes at origin (webservice layer)

```go
// File: internal/integration/webservice/payment.go
package webservice

import (
    "context"
    "myproject/pkg/errs"
)

func (w *paymentWebservice) Process(ctx context.Context, ...) error {
    resp, err := w.client.Post(url, body)
    if err != nil {
        return errs.InternalServerError(err, "PAY-00500", "Failed to call payment API")
    }
    if resp.StatusCode >= 400 && resp.StatusCode < 500 {
        return errs.ClientError("Payment rejected", "PAY-00400", resp.StatusCode, resp.Body)
    }
    if resp.StatusCode >= 500 {
        return errs.ServerError("Payment service unavailable", "PAY-01500", resp.StatusCode)
    }
    return nil
}
```

---

### Example 5 — Validation error codes at origin (entrypoint)

```go
// File: internal/integration/entrypoint/controller/handler.go

// Validation error code is passed at the controller where validation happens
err = validator.CustomValidator().Struct("EV-00400", &eventDto)
if err != nil {
    return err  // Error already has code "EV-00400" with field-level details
}
```

---

## Anti-Patterns

### Anti-Pattern 1 — Error codes in the error handler (FORBIDDEN)

```go
// BAD — error handler assigning codes (impossible to trace origin)
func (e *errorHandler) Handle(ctx context.Context, err error, event any) error {
    if strings.Contains(err.Error(), "timeout") {
        return errs.ServerError("Timeout", "TIMEOUT-500", 500)  // WRONG
    }
}
```

Error codes MUST be at the origin. The error handler only processes, never creates codes (except `UNK-00500` for truly unknown errors).

---

### Anti-Pattern 2 — Swallowing errors silently (FORBIDDEN)

```go
// BAD — error disappears without logging or formatting
func (e *errorHandler) Handle(ctx context.Context, err error, event any) error {
    return nil  // WRONG — error swallowed
}
```

Every error must at minimum be logged before being returned or aborted.

---

### Anti-Pattern 3 — Duplicate error codes (FORBIDDEN)

```go
// BAD — same code in two different places
// File: persistence/user.go
return errs.InternalServerError(err, "IS-00500", "user lookup failed")

// File: persistence/order.go
return errs.InternalServerError(err, "IS-00500", "order lookup failed")  // WRONG — same code!
```

Each error code must be globally unique. Use different prefixes or numbers.

---

### Anti-Pattern 4 — Creating errors in services/use cases (FORBIDDEN)

```go
// BAD — service creating custom errors with codes
// File: application/service/process_order.go
func (s *processOrder) Process(ctx context.Context, ...) error {
    return errs.BadRequestError("invalid order", "ORD-00400")  // WRONG — belongs in integration layer
}
```

Services pass through errors from use cases. Error creation with codes belongs in the integration layer where the actual failure occurs.

---

## Quick Reference Checklist

When creating or reviewing an error handler, verify every item:

- [ ] Lives in `internal/integration/entrypoint/controller/`
- [ ] Struct is **unexported** (`errorHandler`, not `ErrorHandler`)
- [ ] Implements `adapter.ErrorHandler` from `internal/integration/adapter/`
- [ ] Constructor follows `New{Name}(...)` pattern returning adapter interface
- [ ] Checks `errors.As(err, &customErr)` to detect custom error types
- [ ] Falls back to `errs.InternalServerError` for unknown errors with code `UNK-00500`
- [ ] Logs every error with `logger.Error`
- [ ] Supports abort behavior for jobs/lambdas
- [ ] Can inject use cases for notifications (Slack, Discord, etc.)
- [ ] Error codes are NEVER created in the error handler (except `UNK-00500`)
- [ ] All error codes across the application are unique
- [ ] Error codes exist at the exact line where the error is thrown
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
