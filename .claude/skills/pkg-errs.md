---
name: Package - errs (Custom Errors)
description: This skill should be used when the user asks to "create a custom error", "add an error type", "create error handling", "add error codes", or mentions anything about the pkg/errs/ package.
version: 1.0.0
---

# Package: errs — Custom Error Types

The `errs` package provides a hierarchy of custom error types that implement a common `BaseError` interface. Each error type carries a status code, error code, description, retry/abort/notification behavior, and can be converted to a consistent JSON DTO for API responses.

**Location:** `pkg/errs/`

---

## BaseError Interface

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
type BaseError interface {
    Code() string                        // Unique error code (e.g., "DOW-00500")
    ErrorDetails() map[string]any        // Structured error details
    Error() string                       // Error message (implements error interface)
    Description() string                 // User-facing description
    StatusCode() int                     // HTTP status code
    SetStatusCode(statusCode int, description string)
    ToDto(transactionId string) Error    // Convert to response DTO
    Is(target error) bool                // Error type matching
    Retries() int                        // Max retry count before escalation
    InternalNotify() bool                // Should notify internal systems?
    Abort() bool                         // Should abort processing? (jobs only)
}
```

---

## Response DTO Structure

```go
type Error struct {
    Error ErrorData `json:"error"`
}

type ErrorData struct {
    Id           string         `json:"id,omitempty"`
    Description  string         `json:"description,omitempty"`
    Code         string         `json:"code,omitempty"`
    ErrorDetails []ErrorDetails `json:"error_details,omitempty"`
}

type ErrorDetails struct {
    Attribute string   `json:"attribute"`
    Messages  []string `json:"messages"`
}
```

---

## Error Types

### BadRequestError (400)

```go
func BadRequestError(description, code string, details ...ErrorDetails) BaseError
```
- **Status:** 400 | **Retries:** 0 | **Abort:** true | **Notify:** true
- Use for: invalid input, validation failures, duplicate records

### NotFoundError (404)

```go
func NotFoundError(message string, code string) BaseError
func IsNotFoundError(err error) bool  // Helper to check type
```
- **Status:** 404 | **Retries:** 0 | **Abort:** true | **Notify:** true
- Use for: resource not found

### InternalServerError (500)

```go
func InternalServerError(err error, code, message string) BaseError
```
- **Status:** 500 | **Retries:** 5 | **Abort:** false | **Notify:** true
- Use for: unexpected internal failures
- Wraps the original error and adds "UNKNOWN_ERROR" attribute to details

### PanicError (500)

```go
func PanicError(message string) BaseError
```
- **Status:** 500 | **Retries:** 0 | **Abort:** false | **Notify:** true
- Use for: recovered panics

### ServerError (configurable status)

```go
func ServerError(description, code string, status int) BaseError
```
- **Status:** configurable | **Retries:** 5 | **Abort:** false | **Notify:** true
- Use for: external API server errors (5xx responses)

### ClientError (configurable status)

```go
func ClientError(description, code string, status int, responseBody ...io.ReadCloser) BaseError
```
- **Status:** configurable | **Retries:** 0 | **Abort:** true | **Notify:** true
- Use for: external API client errors (4xx responses)
- Can parse response body to extract error details from external API

### DbError (auto-detects)

```go
func DbError(err error) BaseError
```
- **Auto-detects** the error type from GORM:
  - `gorm.ErrRecordNotFound` → NotFoundError (404)
  - Unique constraint violation → BadRequestError (400)
  - Everything else → InternalServerError (500)

---

## Implementation Pattern

Each error type follows the same struct pattern:

```go
type badRequestError struct {
    message      string
    description  string
    code         string
    errorDetails []ErrorDetails
    statusCode   int
}

func BadRequestError(description, code string, details ...ErrorDetails) BaseError {
    return badRequestError{
        statusCode:   http.StatusBadRequest,
        message:      "bad request",
        description:  description,
        code:         code,
        errorDetails: details,
    }
}

func (b badRequestError) ToDto(transactionId string) Error {
    return Error{
        Error: ErrorData{
            Description:  b.description,
            Code:         b.code,
            Id:           transactionId,
            ErrorDetails: b.errorDetails,
        },
    }
}

func (b badRequestError) Error() string       { return b.message }
func (b badRequestError) Description() string { return b.description }
func (b badRequestError) StatusCode() int     { return b.statusCode }
func (b badRequestError) Code() string        { return b.code }
func (b badRequestError) Retries() int        { return 0 }
func (b badRequestError) InternalNotify() bool { return true }
func (b badRequestError) Abort() bool         { return true }

func (b badRequestError) Is(target error) bool {
    return errors.As(target, &b)
}

func (b badRequestError) ErrorDetails() map[string]any {
    var details = make(map[string]any)
    for _, detail := range b.errorDetails {
        details["attribute"] = detail.Attribute
        details["messages"] = detail.Messages
    }
    return details
}
```

To add a new error type, copy this pattern and adjust: status code, retries, abort, notify, and the factory function signature.

---

## Key Rules

- Each error type is **unexported** (`badRequestError`, not `BadRequestError`)
- Factory functions are **exported** (`BadRequestError()`, `NotFoundError()`)
- All implement `BaseError` interface
- Error codes are assigned at the **origin** where the error is thrown (see Error Handler skill)
- `DbError` is the only "smart" factory — it auto-detects GORM error types

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
