---
name: Package - timeutils (Time Utilities)
description: This skill should be used when the user asks to "create time utilities", "mock time in tests", "create a time provider", or mentions anything about the pkg/timeutils/ package.
version: 1.0.0
---

# Package: timeutils — Time Utilities with Mocking Support

The `timeutils` package provides a configurable time provider that returns UTC time. It uses a singleton pattern with a swappable provider function, enabling tests to control time without modifying application code.

**Location:** `pkg/timeutils/`

---

## Architecture

```
pkg/timeutils/
├── config.go    ← Singleton config with provider pattern
└── time.go      ← Now() convenience function
```

---

## Usage

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// Application code — always use timeutils.Now() instead of time.Now()
now := timeutils.Now()  // Returns UTC time from configured provider
```

---

## Implementation

### Config (Singleton with Swappable Provider)

```go
package timeutils

import (
    "sync"
    "time"
)

type CurrentTimeProvider func() time.Time

type timeConfig struct {
    currentTimeProvider CurrentTimeProvider
}

var instance *timeConfig
var once sync.Once

func GetTimeConfig() *timeConfig {
    if instance == nil {
        once.Do(func() {
            instance = &timeConfig{
                currentTimeProvider: time.Now,  // Default: real time
            }
        })
    }
    return instance
}

func SetCurrentTimeProvider(provider CurrentTimeProvider) {
    GetTimeConfig().currentTimeProvider = provider
}
```

### Now() Function

```go
package timeutils

import "time"

func Now() time.Time {
    return GetTimeConfig().currentTimeProvider().UTC()
}
```

---

## Testing Support

In BDD tests, swap the time provider to control time:

```go
// In test setup
timeutils.SetCurrentTimeProvider(func() time.Time {
    return time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
})

// In test teardown — restore real time
timeutils.SetCurrentTimeProvider(time.Now)
```

This allows testing time-dependent logic (schedules, expiration, etc.) without flaky tests.

---

## Usage in Domain Entities

Domain entities use `timeutils.Now()` for time-based business logic:

```go
// File: internal/domain/entity/schedule.go
func (s *Schedule) IsActive() bool {
    if s.StartsAt == nil || s.EndsAt == nil {
        return false
    }
    now := timeutils.Now()
    return s.StartsAt.Before(now) && s.EndsAt.After(now)
}
```

---

## Key Rules

- **No internal imports** — standalone package
- **No cross-pkg imports** — fully self-contained
- **Always returns UTC** — `Now()` normalizes to UTC
- **Singleton pattern** — `sync.Once` for initialization
- **Swappable provider** — `SetCurrentTimeProvider()` for testing
- **Use `timeutils.Now()` everywhere** — never call `time.Now()` directly in application code
- **Copy-ready** — can be dropped into any Go project

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
