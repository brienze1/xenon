---
name: Package - http (HTTP Client)
description: This skill should be used when the user asks to "create an HTTP client", "add an HTTP wrapper", "create a REST client package", or mentions anything about the pkg/http/ package.
version: 1.0.0
---

# Package: http — HTTP Client Wrapper

The `http` package provides a custom HTTP client that wraps Go's `http.Client` with structured logging, sensitive data masking, and tracing support. It exposes a `Client` interface for easy mocking in tests.

**Location:** `pkg/http/`

---

## Interface

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
type Client interface {
    Do(ctx context.Context, req *http.Request) (*http.Response, error)
}
```

---

## Implementation

```go
// Local interface — does NOT import pkg/logger
type loggerAdapter interface {
    Debug(ctx context.Context, message string, metadata ...any)
    Error(ctx context.Context, err error, message string, metadata ...any)
}

type customHttp struct {
    client *http.Client
    logger loggerAdapter
}

func HttpClient(timeout time.Duration, logger loggerAdapter) Client {
    client := &http.Client{Timeout: timeout}
    return &customHttp{client: client, logger: logger}
}
```

---

## Key Features

### Request/Response Logging
Every request is logged with:
- URL, method, headers, request body, response body, status code
- Debug level for successful requests
- Error level for failed requests (error or status >= 400)

### Sensitive Data Masking
```go
// Headers: Authorization masked
filteredHeaders["Authorization"] = []string{"Bearer ******"}

// Payload: access_token masked
filteredPayload["access_token"] = "******"
```

### Tracing Integration
Injects DataDog trace context into outgoing HTTP headers when tracing is enabled:
```go
func (c *customHttp) withContext(ctx context.Context, req *http.Request) *http.Client {
    if span, found := tracer.SpanFromContext(ctx); found {
        _ = tracer.Inject(span.Context(), tracer.HTTPHeadersCarrier(req.Header))
    }
    return c.client
}
```

### Body Preservation
Request and response bodies are read for logging, then restored so they can be read again by the caller:
```go
req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
```

---

## Key Rules

- **No internal imports** — standalone package
- **Logger via local interface** — defines `loggerAdapter`, does NOT import `pkg/logger`
- **Exported interface, unexported struct** — `Client` interface, `customHttp` struct
- **Factory function** — `HttpClient(timeout, logger)` returns `Client`
- **Copy-ready** — can be dropped into any Go project

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
