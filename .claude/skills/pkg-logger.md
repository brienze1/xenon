---
name: Package - logger (Structured Logging)
description: This skill should be used when the user asks to "create a logger", "add logging", "set up structured logs", "create log context", or mentions anything about the pkg/logger/ package.
version: 1.0.0
---

# Package: logger — Structured Logging

The `logger` package provides structured JSON logging with context propagation, log level filtering, correlation/transaction ID tracking, and tracing integration. It exposes both package-level functions and a wrapper interface for dependency injection.

**Location:** `pkg/logger/`

---

## Architecture

```
pkg/logger/
├── context.go    ← Ctx interface, context carrier with correlation/transaction IDs
├── logger.go     ← Core log functions (Error, Warn, Info, Debug, Trace)
├── wrapper.go    ← Logger interface wrapper for DI
├── error.go      ← Error model helper for log output
└── utils.go      ← Utility functions (parseBool, entityName)
```

---

## Context Management

The logger uses a custom `Ctx` interface that extends `context.Context` with logging-specific metadata:

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
type Ctx interface {
    context.Context
    GetRequestId() string
    GetLogGroup() string
    GetLogStream() string
    GetCorrelationId() string
    SetCorrelationId(correlationId string)
    SetReceiveCount(receiveCount int)
    GetReceiveCount() int
    SetOperationUniqueId(operationUniqueId string)
    GetOperationUniqueId() string
}
```

`GetContext()` is a carrier function that accepts multiple context types and returns the logger `Ctx`:

```go
func GetContext(ctx ...any) Ctx {
    // Accepts: fiber.Ctx, context.Context, logger.Ctx
    // Returns logger Ctx with correlation/transaction IDs
}
```

---

## Log Functions (Package-Level)

```go
func Error(ctx context.Context, err error, message string, metadata ...any)
func Warn(ctx context.Context, err error, message string, metadata ...any)
func Info(ctx context.Context, message string, metadata ...any)
func Debug(ctx context.Context, message string, metadata ...any)
func Trace(ctx context.Context, message string, metadata ...any)
```

**Usage throughout the app:**
```go
// Controllers and services use Info level
logger.Info(ctx, "Started", metadata, map[string]any{"source": source})
logger.Info(ctx, "Finished", metadata, map[string]any{"source": source})

// Persistence and webservices use Debug level
logger.Debug(ctx, "Started", map[string]any{"userId": userId})
logger.Debug(ctx, "Finished", map[string]any{"userId": userId})

// Errors use Error level
logger.Error(ctx, err, "Failed to process", metadata)

// Warnings
logger.Warn(ctx, err, "Unexpected state", metadata)
```

---

## Log Output Format (JSON)

```go
type logPattern struct {
    Metadata      []any  `json:"metadata,omitempty"`
    TraceId       string `json:"trace_id,omitempty"`
    SpanId        string `json:"span_id,omitempty"`
    RequestId     string `json:"requestId,omitempty"`
    Timestamp     string `json:"timestamp"`
    Level         string `json:"level"`
    Service       string `json:"service"`
    Operation     string `json:"operation"`
    Exceptions    any    `json:"exceptions,omitempty"`
    Detail        string `json:"detail"`
    CorrelationId string `json:"correlationId,omitempty"`
    TransactionId string `json:"transactionId,omitempty"`
}
```

---

## Logger Wrapper Interface (for DI)

When other packages need logging (e.g., `pkg/http`, `pkg/redis`), they define their own local interface. The logger wrapper implements those interfaces:

```go
type Logger interface {
    Trace(ctx context.Context, message string, metadata ...any)
    Debug(ctx context.Context, message string, metadata ...any)
    Info(ctx context.Context, message string, metadata ...any)
    Warn(ctx context.Context, err error, message string, metadata ...any)
    Error(ctx context.Context, err error, message string, metadata ...any)
}

func LoggerWrapper() Logger {
    return &loggerWrapper{}
}
```

**Used by other packages:**
```go
// pkg/http defines its own interface
type loggerAdapter interface {
    Debug(ctx context.Context, message string, metadata ...any)
    Error(ctx context.Context, err error, message string, metadata ...any)
}

// Constructor accepts it
func HttpClient(timeout time.Duration, logger loggerAdapter) Client { ... }

// Caller passes the wrapper
httpClient := http.HttpClient(20*time.Second, logger.LoggerWrapper())
```

---

## Log Level Filtering

Log level is controlled by the `LOG_LEVEL` environment variable:

```
ERROR > WARN > INFO > DEBUG > TRACE > OFF
```

`shouldPrintLog()` checks if the current log level allows printing.

---

## Key Rules

- **No internal imports** — this is a standalone package
- **No cross-pkg imports** — other packages define their own logger interfaces
- **Context carries metadata** — correlation IDs, transaction IDs, request IDs flow through
- **JSON output** — structured logs with consistent fields
- **Log levels by layer**: Info for controllers/services, Debug for persistence/webservices

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
