---
name: Packages - Generic (Clean Architecture)
description: This skill should be used when the user asks to "create a package", "add a pkg", "create a shared library", "create a utility package", or mentions anything about creating new packages in the pkg/ folder.
version: 1.0.0
---

# Packages (pkg/) — Golang Clean Architecture

The `pkg/` folder contains self-contained, reusable library packages. They are fully decoupled from the application's internal code and from each other. Any single `pkg/<type>` folder can be copied to another repository and work out of the box.

---

## Core Rules

### Rule 1: Location

All packages MUST live in `pkg/<type>/` where `<type>` describes the concern.

```
pkg/
├── errs/          ← Custom error types
├── logger/        ← Structured logging
├── http/          ← HTTP client wrapper
├── redis/         ← Redis helper
└── timeutils/     ← Time utilities with mocking support
```

---

### Rule 2: Zero Internal Imports (CRITICAL)

Packages MUST NEVER import anything from `internal/`. They are completely independent of the application code.

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// BAD — pkg importing from internal
import "myproject/internal/domain/entity"      // FORBIDDEN
import "myproject/internal/infra/properties"   // FORBIDDEN

// GOOD — pkg only imports stdlib, other pkgs, or external libs
import (
    "context"
    "encoding/json"
    "time"
    "github.com/redis/go-redis/v9"
)
```

---

### Rule 3: Zero Cross-Package Imports

Packages MUST NOT import other `pkg/` packages. Each package is a standalone unit.

```go
// BAD — pkg importing another pkg
// File: pkg/http/client.go
import "myproject/pkg/logger"  // FORBIDDEN — creates coupling

// GOOD — accept logger as interface parameter
type loggerAdapter interface {
    Debug(ctx context.Context, message string, metadata ...any)
    Error(ctx context.Context, err error, message string, metadata ...any)
}

func HttpClient(timeout time.Duration, logger loggerAdapter) Client {
    // ...
}
```

---

### Rule 4: Expose Interfaces, Not Concrete Types

Packages should expose interfaces for their main functionality. Concrete implementations are unexported and returned via factory functions.

```go
// GOOD — exported interface, unexported struct
type Client interface {
    Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

type customHttp struct {  // unexported
    client *http.Client
    logger loggerAdapter
}

func HttpClient(timeout time.Duration, logger loggerAdapter) Client {
    return &customHttp{...}
}
```

---

### Rule 5: Accept Dependencies via Interfaces

When a package needs functionality from another package (e.g., logging), it MUST define its own interface for that dependency and accept it via constructor parameters. This maintains full decoupling.

```go
// Package defines its OWN interface for what it needs
type loggerAdapter interface {
    Debug(ctx context.Context, message string, metadata ...any)
    Error(ctx context.Context, err error, message string, metadata ...any)
}

// Constructor accepts the interface
func NewRedisHelper(client redis.UniversalClient, prefix string, logger loggerAdapter) RedisHelper {
    return &redisHelper{client: client, prefix: prefix, logger: logger}
}
```

The caller passes whatever implements that interface — typically the logger wrapper from `pkg/logger`, but it could be anything.

---

### Rule 6: Self-Contained and Copy-Ready

Every `pkg/<type>/` folder must be independently usable. If you copy it to another repo:
- It compiles without changes (no internal imports)
- It works without other pkg folders (no cross-pkg imports)
- Dependencies are injected, not hardcoded

---

### Rule 7: Package Naming

The package name matches the folder name. Keep it short and descriptive.

| Folder | Package | Purpose |
|---|---|---|
| `pkg/errs` | `errs` | Custom error types |
| `pkg/logger` | `logger` | Structured logging |
| `pkg/http` | `http` | HTTP client wrapper |
| `pkg/redis` | `redis` (or `redishelper`) | Redis operations helper |
| `pkg/timeutils` | `timeutils` | Time utilities |

---

## Anti-Patterns

### Anti-Pattern 1 — Importing from internal (FORBIDDEN)

```go
// File: pkg/logger/logger.go
import "myproject/internal/infra/properties"  // WRONG
```

### Anti-Pattern 2 — Cross-package import (FORBIDDEN)

```go
// File: pkg/http/client.go
import "myproject/pkg/logger"  // WRONG — define local interface instead
```

### Anti-Pattern 3 — Hardcoded dependencies (FORBIDDEN)

```go
// BAD — creating its own logger
func HttpClient() Client {
    logger := log.New(os.Stdout, "", log.LstdFlags)  // WRONG — inject instead
}
```

---

## Quick Reference Checklist

When creating a new package, verify every item:

- [ ] Lives in `pkg/<type>/` with descriptive package name
- [ ] Zero imports from `internal/`
- [ ] Zero imports from other `pkg/` packages
- [ ] Dependencies accepted via locally-defined interfaces
- [ ] Main functionality exposed via exported interface
- [ ] Concrete implementations are unexported
- [ ] Factory function creates and returns the interface
- [ ] Can be copied to another repo and compile without changes
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
