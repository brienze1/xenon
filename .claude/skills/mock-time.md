---
name: Mock - Time (Controllable Clock)
description: Use when creating or referencing the time mock for BDD tests.
version: 1.0.0
---

# Mock: Time (Controllable Clock)

Replaces `time.Now()` with a controllable clock for deterministic testing of time-dependent logic (schedules, expiration, etc.).

**Location:** `test/integration/mock/time.go`

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package mock

import "time"

type Time struct {
    currentTime time.Time
    updatedAt   time.Time
}

func NewTime() *Time {
    return &Time{
        currentTime: time.Now().UTC(),
        updatedAt:   time.Now().UTC(),
    }
}

func (t *Time) SetCurrentTime(currentTime time.Time) {
    t.currentTime = currentTime.UTC()
    t.updatedAt = time.Now().UTC()
}

func (t *Time) Now() time.Time {
    elapsed := time.Since(t.updatedAt)
    return t.currentTime.Add(elapsed).UTC()
}
```

**Usage:**
```go
test.time = mock.NewTime()
timeutils.SetCurrentTimeProvider(test.time.Now)  // ONE-TIME injection in InitializeScenario

// In a feature file step:
// Given the current time is "2024-06-15T10:30:00Z"
test.time.SetCurrentTime(parsedTime)
```

**Lifecycle:** The time mock is created once and injected once via `timeutils.SetCurrentTimeProvider()` in `InitializeScenario`. The time mock resets naturally since `SetCurrentTime()` overwrites the previous value. The mock instance is never recreated or re-injected — it lives for the entire test suite.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
