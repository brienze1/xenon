---
name: Package - redis (Redis Helper)
description: This skill should be used when the user asks to "create a Redis helper", "add Redis caching", "create Redis operations", or mentions anything about the pkg/redis/ package.
version: 1.0.0
---

# Package: redis — Redis Helper

The `redis` package provides a high-level helper that wraps `redis.UniversalClient` with structured operations (Put, Get, Delete, Clear), key prefixing, serialization/deserialization, TTL management, and debug logging.

**Location:** `pkg/redis/`

---

## Interface

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
type RedisHelper interface {
    Put(ctx context.Context, key string, value any, ttl ...time.Duration) error
    Get(ctx context.Context, key string, dest any) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context, prefix string) error
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

type redisHelper struct {
    client       redis.UniversalClient
    globalPrefix string
    logger       loggerAdapter
}

func Redis(client redis.UniversalClient, globalPrefix string, logger loggerAdapter) RedisHelper {
    return &redisHelper{
        client:       client,
        globalPrefix: globalPrefix,
        logger:       logger,
    }
}
```

---

## Key Features

### Key Prefixing
All keys are automatically prefixed with the global prefix for namespace isolation:
```go
key := r.globalPrefix + ":" + userKey
```

### Smart Serialization (Put)
Handles different value types:
- `string` → stored directly
- `[]byte` → stored directly
- Everything else → JSON marshaled

### Smart Deserialization (Get)
Handles different destination types:
- `*string` → stored directly
- `*[]byte` → stored directly
- Everything else → JSON unmarshaled

### Default TTL
Default TTL is 10 hours if none specified. Configurable via the `ttl` variadic parameter.

### Clear by Prefix
`Clear()` uses Redis SCAN to find and delete all keys matching a prefix pattern:
```go
func (r *redisHelper) Clear(ctx context.Context, prefix string) error {
    // Scans for globalPrefix:prefix:* and deletes all matches
}
```

### Debug Logging
All operations log Started/Finished at Debug level with key and metadata.

---

## Key Rules

- **No internal imports** — standalone package
- **Logger via local interface** — defines `loggerAdapter`, does NOT import `pkg/logger`
- **Exported interface, unexported struct** — `RedisHelper` interface, `redisHelper` struct
- **Factory function** — `Redis(client, prefix, logger)` returns `RedisHelper`
- **Accepts `redis.UniversalClient`** — works with real Redis and miniredis for tests
- **Copy-ready** — can be dropped into any Go project

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
