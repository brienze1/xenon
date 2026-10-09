---
name: Controllers (Clean Architecture)
description: This skill should be used when the user asks to "create a controller", "add a controller", "create an entrypoint", "create a handler", "add a REST endpoint handler", "add a consumer handler", "create an event handler", or mentions anything about controllers in the internal/integration/entrypoint/controller/ package.
version: 1.0.0
---

# Controllers — Golang Clean Architecture

Controllers are the entry point of the application. They receive requests (HTTP, events, messages), validate input using DTOs, delegate execution to services/use cases, and return responses. Controllers abstract frameworks — the framework-specific code lives in infrastructure, not in the controller.

---

## Core Rules

### Rule 1: Location

Controllers MUST live in `internal/integration/entrypoint/controller/` and declare `package controller`.

```
internal/
└── integration/
    └── entrypoint/
        ├── controller/
        │   ├── handler.go           ← main controller
        │   └── error_handler.go     ← error handling controller
        ├── dto/
        │   ├── event.go             ← input DTO with json/validate tags
        │   └── post_metadata.go     ← nested DTO with ToEntity() method
        ├── errors/
        │   └── validation_error.go  ← validation error formatting
        └── validator/
            └── custom.go            ← custom validation rules
```

---

### Rule 2: Framework Abstraction

Controllers MUST NOT depend on framework-specific types directly. The framework lives in the infrastructure layer and calls the controller through an adapter interface.

```
Infrastructure (framework)          Integration (controller)
─────────────────────────           ────────────────────────
infra/router/ (Fiber/Gin)     →     entrypoint/controller/
infra/broker/ (RabbitMQ/SQS)  →     entrypoint/controller/
infra/handler/ (AWS Lambda)   →     entrypoint/controller/
```

The controller implements an integration adapter (e.g., `adapter.Handler`) and the infrastructure layer calls it. The controller never imports `fiber`, `gin`, `lambda`, etc.

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// GOOD — controller implements adapter.Handler, no framework dependency
type handler struct {
    filterService appAdapter.FilterAndProcessService
}

func NewHandler(filterService appAdapter.FilterAndProcessService) adapter.Handler {
    return &handler{filterService: filterService}
}

func (h *handler) Handle(ctx context.Context, event domain.QueueMessage[dto.Event]) error {
    // framework-free logic
}
```

```go
// BAD — controller importing framework directly
import "github.com/gofiber/fiber/v2"

func (h *handler) Handle(c *fiber.Ctx) error {  // WRONG — framework type in controller
    // ...
}
```

---

### Rule 3: Validate Input Using DTOs

Controllers MUST validate incoming requests using DTOs before passing data to services or use cases. DTOs live in `internal/integration/entrypoint/dto/` and carry `json:` and `validate:` tags.

The flow is:
1. Receive raw input (event, request body)
2. Validate the DTO using the validator
3. Convert DTO to domain entity via `ToEntity()` method
4. Pass domain entity to service/use case

```go
func (h *handler) Handle(ctx context.Context, event domain.QueueMessage[dto.Event]) error {
    // 1. Extract DTO
    eventDto := event.Payload

    // 2. Validate
    err := validator.CustomValidator().Struct("EV-00400", &eventDto)
    if err != nil {
        return err
    }

    // 3. Convert to domain entity and call service
    err = h.service.Process(ctx, eventDto.Source, eventDto.GetEventType(), eventDto.EventData.ToEntity())
    if err != nil {
        return err
    }

    return nil
}
```

---

### Rule 4: Logger — Started and Finished

Every controller method MUST begin with a `Started` log and end with a `Finished` log, including the raw input/output.

```go
func (h *handler) Handle(ctx context.Context, event domain.QueueMessage[dto.Event]) error {
    logger.Info(ctx, "Started", event)

    // ... validation and delegation ...

    logger.Info(ctx, "Finished", event)
    return nil
}
```

Key points:
- `Started` logs the raw input as received (before validation, before conversion).
- `Finished` logs after successful completion.
- If an error occurs and the method returns early, `Finished` is NOT logged.

---

### Rule 5: Controllers Implement Integration Adapters

Controllers MUST implement an adapter interface from `internal/integration/adapter/`. They are never called directly.

```go
// Adapter: internal/integration/adapter/handler.go
package adapter

import (
    "context"
    domain "myproject/pkg/domain/value-objects"
    "myproject/internal/integration/entrypoint/dto"
)

type Handler interface {
    Handle(ctx context.Context, message domain.QueueMessage[dto.Event]) error
}
```

```go
// Controller: internal/integration/entrypoint/controller/handler.go
package controller

func NewHandler(
    errorHandler adapter.ErrorHandler,
    service appAdapter.ProcessOrderService,
) adapter.Handler {
    return &handler{
        errorHandler: errorHandler,
        service:      service,
    }
}
```

---

### Rule 6: Controllers Depend on Application Adapter Interfaces

Controllers call services and use cases through their **application adapter interfaces**, never through concrete types.

```go
type handler struct {
    errorHandler adapter.ErrorHandler                        // integration adapter
    service      appAdapter.FilterAndProcessService          // application adapter
}
```

```go
// BAD — importing concrete service
import "myproject/internal/application/service"

type handler struct {
    svc *service.processOrder  // WRONG — must use adapter interface
}
```

---

### Rule 7: Error Handling — Delegate, Don't Swallow

Controllers handle errors by delegating to an error handler, not by swallowing them. Panic recovery is also the controller's responsibility.

```go
func (h *handler) Handle(ctx context.Context, event domain.QueueMessage[dto.Event]) (err error) {
    logger.Info(ctx, "Started", event)

    defer func(ctx context.Context, event any) {
        if recovered := recover(); recovered != nil {
            err = handlePanic(ctx, event, recovered)
        }
        if err != nil {
            err = h.errorHandler.Handle(ctx, err, event)
        }
    }(ctx, event)

    // ... validation and delegation ...

    logger.Info(ctx, "Finished", event)
    return nil
}
```

Key points:
- Named return (`err error`) enables the defer to modify the return value.
- Panic recovery converts panics to errors.
- Error handler processes all errors uniformly (logging, retry logic, notifications).

---

### Rule 8: Naming and Constructor Conventions

Controllers follow the same constructor pattern as services: `New` + name.

| Element | Pattern | Example |
|---|---|---|
| Controller struct | Unexported, descriptive | `handler`, `createOrder`, `getUser` |
| Constructor | `New{Name}(...)` returning integration adapter | `NewHandler(...)`, `NewCreateOrder(...)`, `NewGetUser(...)` |
| Constructor return type | Integration adapter (`internal/integration/adapter/`) | `adapter.Handler` |
| Package | `controller` | `package controller` |

```go
// Constructor pattern — same as services
func NewHandler(
    errorHandler adapter.ErrorHandler,
    service appAdapter.ProcessOrderService,
) adapter.Handler {
    return &handler{
        errorHandler: errorHandler,
        service:      service,
    }
}

func NewGetUser(
    errorHandler adapter.ErrorHandler,
    findUser appAdapter.FindUserService,
) adapter.GetUserController {
    return &getUser{
        errorHandler: errorHandler,
        findUser:     findUser,
    }
}
```

The adapter interface that the controller implements lives in `internal/integration/adapter/` — this is the integration layer's own adapter folder. The infrastructure layer (routers, lambda handlers, brokers) depends on this adapter to call the controller.

---

## Code Examples

### Example 1 — Event/message handler controller

```go
package controller

import (
    "context"
    "fmt"
    "runtime"

    domain "myproject/pkg/domain/value-objects"

    appAdapter "myproject/internal/application/adapter"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/entrypoint/dto"
    "myproject/internal/integration/entrypoint/validator"
    "myproject/pkg/errs"
    "myproject/pkg/logger"
)

type handler struct {
    errorHandler adapter.ErrorHandler
    service      appAdapter.ProcessOrderService
}

func NewHandler(
    errorHandler adapter.ErrorHandler,
    service appAdapter.ProcessOrderService,
) adapter.Handler {
    return &handler{
        errorHandler: errorHandler,
        service:      service,
    }
}

func (h *handler) Handle(brokerCtx context.Context, event domain.QueueMessage[dto.Event]) (err error) {
    ctx := logger.GetContext(brokerCtx)
    logger.Info(ctx, "Started", event)

    defer func(ctx context.Context, event any) {
        if recovered := recover(); recovered != nil {
            err = handlePanic(ctx, event, recovered)
        }
        if err != nil {
            err = h.errorHandler.Handle(ctx, err, event)
        }
    }(ctx, event)

    eventDto := event.Payload
    ctx.SetCorrelationId(eventDto.GetCorrelationId())

    err = validator.CustomValidator().Struct("EV-00400", &eventDto)
    if err != nil {
        return err
    }

    err = h.service.Process(ctx, eventDto.Source, eventDto.GetEventType(), eventDto.EventData.ToEntity())
    if err != nil {
        return err
    }

    logger.Info(ctx, "Finished", event)
    return nil
}

func handlePanic(ctx context.Context, event any, recovered any) error {
    var message string
    switch e := recovered.(type) {
    case string:
        message = fmt.Sprint("recovered (string) panic:", e)
    case runtime.Error:
        message = fmt.Sprint("recovered (runtime.Error) panic:", e.Error())
    case error:
        message = fmt.Sprint("recovered (error) panic:", e.Error())
    default:
        message = fmt.Sprint("recovered (default) panic:", e)
    }

    err := errs.PanicError(message)
    logger.Error(ctx, err, "Panic occurred", event)
    return err
}
```

---

### Example 2 — REST API controller (conceptual)

```go
package controller

import (
    "context"

    appAdapter "myproject/internal/application/adapter"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/entrypoint/dto"
    "myproject/internal/integration/entrypoint/validator"
    "myproject/pkg/logger"
)

type createOrder struct {
    service appAdapter.CreateOrderService
}

func NewCreateOrder(service appAdapter.CreateOrderService) adapter.CreateOrderController {
    return &createOrder{service: service}
}

func (c *createOrder) Handle(ctx context.Context, request dto.CreateOrderRequest) (*dto.CreateOrderResponse, error) {
    logger.Info(ctx, "Started", request)

    err := validator.CustomValidator().Struct("ORD-00400", &request)
    if err != nil {
        return nil, err
    }

    result, err := c.service.Create(ctx, request.ToEntity())
    if err != nil {
        return nil, err
    }

    response := dto.CreateOrderResponseFromEntity(result)

    logger.Info(ctx, "Finished", response)
    return &response, nil
}
```

Key points:
- Validates input DTO.
- Converts DTO to entity (`request.ToEntity()`) before calling service.
- Converts entity back to response DTO (`dto.CreateOrderResponseFromEntity(result)`).
- The framework (Fiber/Gin) would live in `infra/router/` and call this controller through the adapter.

---

### Example 3 — Controller calling a use case directly (no service needed)

When the controller only needs a single use case (no orchestration), it can depend on the use case interface directly via the application adapter.

```go
package controller

import (
    "context"

    appAdapter "myproject/internal/application/adapter"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/entrypoint/dto"
    "myproject/internal/integration/entrypoint/validator"
    "myproject/pkg/logger"
)

type getUser struct {
    errorHandler adapter.ErrorHandler
    findUser     appAdapter.FindUserService
}

func NewGetUser(
    errorHandler adapter.ErrorHandler,
    findUser appAdapter.FindUserService,
) adapter.GetUserController {
    return &getUser{
        errorHandler: errorHandler,
        findUser:     findUser,
    }
}

func (g *getUser) Handle(ctx context.Context, request dto.GetUserRequest) (response *dto.GetUserResponse, err error) {
    logger.Info(ctx, "Started", request)

    defer func(ctx context.Context, request any) {
        if recovered := recover(); recovered != nil {
            err = handlePanic(ctx, request, recovered)
        }
        if err != nil {
            err = g.errorHandler.Handle(ctx, err, request)
        }
    }(ctx, request)

    err = validator.CustomValidator().Struct("USR-00400", &request)
    if err != nil {
        return nil, err
    }

    user, err := g.findUser.FindById(ctx, request.UserId)
    if err != nil {
        return nil, err
    }

    result := dto.GetUserResponseFromEntity(*user)

    logger.Info(ctx, "Finished", result)
    return &result, nil
}
```

Key points:
- Same pattern as the event handler — defer with panic recovery and error handler.
- Validates input, calls use case/service, converts result to response DTO.
- Error handler is always injected and used in the defer block.

---

## The Controller Flow

```
1. Receive raw input (event, HTTP request, message)
         ↓
2. Log "Started" with raw input
         ↓
3. Extract DTO from input
         ↓
4. Validate DTO (validator.CustomValidator().Struct(...))
         ↓  (validation error → return error → error handler)
5. Convert DTO to domain entity (dto.ToEntity())
         ↓
6. Call service/use case via application adapter interface
         ↓  (error → return error → error handler)
7. Convert result entity to response DTO (if needed)
         ↓
8. Log "Finished" with output
         ↓
9. Return response
```

---

## Anti-Patterns

### Anti-Pattern 1 — Framework dependency in controller (FORBIDDEN)

```go
// BAD — controller importing Fiber
import "github.com/gofiber/fiber/v2"

func (h *handler) Handle(c *fiber.Ctx) error { ... }

// BAD — controller importing Gin
import "github.com/gin-gonic/gin"

func (h *handler) Handle(c *gin.Context) { ... }
```

Framework types belong in `infra/router/`, `infra/handler/`, or `infra/broker/`. The controller receives framework-agnostic types through adapter interfaces.

---

### Anti-Pattern 2 — Passing DTOs to services (FORBIDDEN)

```go
// BAD — passing DTO to service
err = h.service.Process(ctx, eventDto)  // WRONG — service receives domain types

// GOOD — convert to entity first
err = h.service.Process(ctx, eventDto.Source, eventDto.GetEventType(), eventDto.EventData.ToEntity())
```

---

### Anti-Pattern 3 — Business logic in controller (FORBIDDEN)

```go
// BAD — business logic in controller
func (h *handler) Handle(ctx context.Context, event dto.Event) error {
    if event.EventData.Channel == "instagram" && event.EventData.MediaType == "reels" {  // WRONG
        // business logic here...
    }
}
```

Controllers validate input and delegate. Business logic belongs on entities and is orchestrated by services.

---

### Anti-Pattern 4 — Skipping validation (FORBIDDEN)

```go
// BAD — no validation before calling service
func (h *handler) Handle(ctx context.Context, event dto.Event) error {
    return h.service.Process(ctx, event.ToEntity())  // WRONG — not validated
}
```

Always validate DTOs before passing data forward.

---

### Anti-Pattern 5 — Missing Started/Finished logs (FORBIDDEN)

```go
// BAD — no observability
func (h *handler) Handle(ctx context.Context, event dto.Event) error {
    return h.service.Process(ctx, event.ToEntity())
}
```

---

### Anti-Pattern 6 — Calling concrete service (FORBIDDEN)

```go
// BAD — direct dependency on service struct
import "myproject/internal/application/service"

type handler struct {
    svc *service.processOrder  // WRONG
}

// GOOD — depend on adapter interface
import appAdapter "myproject/internal/application/adapter"

type handler struct {
    svc appAdapter.ProcessOrderService
}
```

---

## Quick Reference Checklist

When creating or reviewing a controller, verify every item:

- [ ] File is in `internal/integration/entrypoint/controller/` and declares `package controller`
- [ ] Struct is **unexported** (`handler`, not `Handler`)
- [ ] Constructor follows `New{Name}(...)` pattern returning integration adapter
- [ ] Implements an integration adapter interface from `internal/integration/adapter/` (e.g., `adapter.Handler`)
- [ ] Does NOT import any framework (Fiber, Gin, Lambda SDK, etc.)
- [ ] Begins with `logger.Info(ctx, "Started", ...)` with raw input
- [ ] Ends with `logger.Info(ctx, "Finished", ...)` with raw output
- [ ] Validates input DTOs before processing
- [ ] Converts DTOs to domain entities via `ToEntity()` before calling services
- [ ] Calls services/use cases through application adapter interfaces only
- [ ] Panic recovery is handled via `defer`
- [ ] Error handling is delegated to an error handler
- [ ] Contains NO business logic — only validation, conversion, and delegation
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
