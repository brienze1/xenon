---
name: Package - utils (General Utilities)
description: This skill should be used when the user asks to "create utility functions", "add helper functions", "create parsing utilities", or mentions anything about the pkg/utils/ package.
version: 1.0.0
---

# Package: utils — General Utilities

The `utils` package provides small, general-purpose utility functions that don't belong to any specific domain. Keep this package minimal — only truly generic functions that are used across multiple layers.

**Location:** `pkg/utils/`

---

## Functions

### ParseFloat

Safely parses a string to float32 with a default fallback:

```go
package utils

import "strconv"

func ParseFloat(value string, defaultValue float32) float32 {
    parsed, err := strconv.ParseFloat(value, 32)
    if err != nil {
        return defaultValue
    }
    return float32(parsed)
}
```

### Ptr (Generic Pointer Helper)

Returns a pointer to any value — useful for constructing structs with pointer fields:

```go
func Ptr[T any](v T) *T {
    return &v
}
```

Usage:
```go
user := entity.User{
    Name:  utils.Ptr("John"),
    Email: utils.Ptr("john@example.com"),
    Age:   utils.Ptr(30),
}
```

---

## Guidelines for Adding New Utils

Only add functions here if they are:
1. **Truly generic** — not specific to any domain or layer
2. **Used in multiple places** — don't add single-use helpers
3. **Pure functions** — no side effects, no state, no I/O
4. **Small** — if it grows beyond a few functions, consider a dedicated package

---

## Key Rules

- **No internal imports** — standalone package
- **No cross-pkg imports** — fully self-contained
- **Minimal surface area** — resist the urge to dump everything here
- **Pure functions only** — no state, no side effects
- **Copy-ready** — no external dependencies

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
