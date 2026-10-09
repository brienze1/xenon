---
name: Mock - Redis (Miniredis)
description: Use when creating or referencing the Redis mock for BDD tests.
version: 1.0.0
---

# Mock: Redis (Miniredis)

Replaces Redis with miniredis in-memory implementation. Compatible with `redis.UniversalClient`.

**Location:** `test/integration/mock/redis.go`

```go
package mock

import (
    "github.com/alicebob/miniredis/v2"
    "github.com/redis/go-redis/v9"
)

var miniRedis *miniredis.Miniredis

func NewRedis() *redis.Client {
    if miniRedis == nil {
        miniRedis = miniredis.RunT(&testing.T{})
    }
    return redis.NewClient(&redis.Options{Addr: miniRedis.Addr()})
}

func ClearRedis(client *redis.Client) error {
    return client.FlushAll(context.Background()).Err()
}
```

**Usage:**
```go
test.redis = mock.NewRedis()
database.SetRedis(test.redis)
```

**Lifecycle:** The Redis mock is created once and injected once via `database.SetRedis()` in `InitializeScenario`. Between BDD scenarios, `ClearRedis(client)` flushes all keys (clearing state). The `ClearRedis()` function is the `Reset()` equivalent for this mock. The mock instance is never recreated or re-injected.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
