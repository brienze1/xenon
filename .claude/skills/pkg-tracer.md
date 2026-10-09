---
name: Package - tracer (Distributed Tracing)
description: This skill should be used when the user asks to "add tracing", "create DataDog tracing", "add distributed tracing", "trace HTTP requests", "trace database queries", "trace Redis operations", or mentions anything about the pkg/tracer/ package.
version: 1.0.0
---

# Package: tracer — Distributed Tracing Integration

The `tracer` package provides DataDog distributed tracing integration for HTTP (Fiber), database (GORM/PostgreSQL), and Redis. Each integration wraps the underlying client with trace context propagation.

**Location:** `pkg/tracer/`

---

## Architecture

```
pkg/tracer/
├── fiber.go    ← HTTP request tracing middleware for Fiber
├── gorm.go     ← GORM PostgreSQL connection with SQL tracing
└── redis.go    ← Redis client with tracing support
```

---

## Fiber Tracing Middleware

Wraps Fiber with DataDog APM tracing — extracts incoming trace context and creates spans for each request:

```go
package tracer

import (
    fibertrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gofiber/fiber.v2"
    "github.com/gofiber/fiber/v2"
)

func FiberTraceMiddleware(app *fiber.App, serviceName string) {
    app.Use(fibertrace.Middleware(
        fibertrace.WithServiceName(serviceName),
    ))
}
```

---

## GORM Tracing (PostgreSQL)

Creates a GORM PostgreSQL connection with automatic SQL query tracing:

```go
package tracer

import (
    sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
    gormtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorm.io/gorm.v1"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func GormPostgres(dsn, serviceName string, config *gorm.Config) (*gorm.DB, error) {
    sqltrace.Register("pgx", &stdlib.Driver{}, sqltrace.WithServiceName(serviceName))
    sqlDb, err := sqltrace.Open("pgx", dsn)
    if err != nil {
        return nil, err
    }

    return gormtrace.Open(
        postgres.New(postgres.Config{Conn: sqlDb}),
        config,
    )
}
```

---

## Redis Tracing

Creates a Redis client with automatic command tracing:

```go
package tracer

import (
    redistrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/redis/go-redis.v9"
    "github.com/redis/go-redis/v9"
)

func TracedRedisClient(opts *redis.Options, serviceName string) *redis.Client {
    client := redis.NewClient(opts)
    redistrace.WrapClient(client, redistrace.WithServiceName(serviceName))
    return client
}
```

---

## Key Rules

- **No internal imports** — standalone package
- **DataDog APM** — uses `dd-trace-go` library
- **Optional** — tracing is enabled via `DD_TRACE_ENABLED` env var
- **Wraps existing clients** — doesn't replace, just adds tracing layer
- **Service name** — each traced component gets a descriptive service name
- **Copy-ready** — depends on `dd-trace-go` and the relevant client libraries

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
