---
name: Enums (Clean Architecture)
description: This skill should be used when the user asks to "create an enum", "add an enum", "define an enum type", "create enum values", "add enum constants", "create a parser for an enum", "move an enum to the correct layer", or mentions anything about enums in any enums/ package across any layer.
version: 1.0.0
---

# Enums — Golang DDD + Clean Architecture

This skill governs how enums must be structured, where they must live, and what logic they are allowed to contain. Enums follow a strict 2-file pattern and are placed in the innermost layer that uses them.

---

## Core Rules

### Rule 1: 2-File Declaration Pattern

Every enum MUST be declared in two files to keep the type definition separate from its values. This improves readability and avoids circular imports.

**File 1 — Type definition** at `enums/{enum_name}.go` (package `enums`)
Declares the type and its enum-specific methods (comparison, string conversion).

**File 2 — Constants and parser** at `enums/{enumname}/value.go` (package `{enumname}`)
Declares the constant values and optional parser/factory functions.

```
enums/
├── order_status.go            ← type OrderStatus string + methods (package enums)
└── orderstatus/
    └── value.go               ← const values + GetOrderStatus() parser (package orderstatus)
```

This results in ergonomic usage:
- **Type references** use `enums.OrderStatus`
- **Value references** use `orderstatus.Pending`, `orderstatus.Completed`
- **Parsing** uses `orderstatus.GetOrderStatus("pending")`

---

### Rule 2: Layer Placement — Innermost Layer Rule

Enums MUST live in the **innermost layer** that uses them. The dependency rule is: inner layers cannot import from outer layers, but outer layers can import from inner layers.

Layer order (innermost → outermost):
```
domain → application → integration (webservices, entrypoints, repositories) → infrastructure
```

**Decision logic:**
- If the domain layer uses the enum → it MUST live in `internal/domain/enums/`
- If only the integration layer uses the enum (e.g., a webservice contract detail) → it lives in `internal/integration/webservice/enums/`
- If only the entrypoint layer uses the enum → it lives in `internal/integration/entrypoint/enums/`
- If a domain entity AND an entrypoint both use it → it MUST live in `internal/domain/enums/` (because domain cannot import from integration, but integration can import from domain)

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// Entity in domain uses Channel enum → enum MUST be in domain/enums/
type Task struct {
    Channel enums.Channel  // enums = internal/domain/enums
}

// Entrypoint also uses it → imports from domain/enums (allowed direction)
import "myproject/internal/domain/enums"
```

```go
// ConnectionType is only used by a webservice client → lives in webservice/enums/
// File: internal/integration/webservice/enums/connection_type.go
type ConnectionType string
```

---

### Rule 3: Enum Logic — Not Business Logic

Enums own their OWN logic, not business logic. Enum methods are strictly about:
- **Comparison**: `Equals()` — type-safe comparison against raw values
- **String conversion**: `String()` — returns the string representation
- **Parsing/Factory**: `GetXxx()` / `Parse()` — converts raw input to the typed enum

Enums MUST NOT contain business rules, domain calculations, or behavior that belongs on an entity.

```go
// GOOD — enum-specific logic
func (c Channel) Equals(s string) bool {
    return strings.EqualFold(string(c), s)
}

// BAD — business logic on an enum (belongs on an entity)
func (s OrderStatus) CanBeRefunded() bool {
    return s == "completed" && time.Since(s.CompletedAt) < 30*24*time.Hour
}
```

---

### Rule 4: No Integration Tags

Like entities, enum types MUST NOT carry any struct tags (`json:`, `bson:`, etc.). Serialization concerns belong in integration-layer DTOs.

---

## Enum Patterns by Base Type

### String-based enum (most common)

**File 1 — Type definition**: `enums/channel.go`

```go
package enums

import "strings"

type Channel string

func (c Channel) Equals(s string) bool {
    return strings.EqualFold(string(c), s)
}
```

**File 2 — Values and parser**: `enums/channel/value.go`

```go
package channel

import (
    "strings"
    "myproject/internal/domain/enums"
)

const (
    Instagram enums.Channel = "instagram"
    Youtube   enums.Channel = "youtube"
    Tiktok    enums.Channel = "tiktok"
    Unknown   enums.Channel = "unknown"
)

func GetChannel(s string) enums.Channel {
    switch strings.ToLower(s) {
    case "instagram":
        return Instagram
    case "youtube":
        return Youtube
    case "tiktok":
        return Tiktok
    default:
        return Unknown
    }
}
```

Usage:
```go
import (
    "myproject/internal/domain/enums"
    "myproject/internal/domain/enums/channel"
)

// Type in struct field
type Task struct {
    Channel enums.Channel
}

// Using a constant value
task.Channel = channel.Instagram

// Parsing from raw string
ch := channel.GetChannel("instagram") // returns channel.Instagram

// Comparing
channel.Instagram.Equals("INSTAGRAM") // true (case-insensitive)
```

---

### Int-based enum

**File 1 — Type definition**: `enums/order_status.go`

```go
package enums

type OrderStatus int

func (o OrderStatus) Equals(i int) bool {
    return i == int(o)
}
```

**File 2 — Values**: `enums/orderstatus/value.go`

```go
package orderstatus

import "myproject/internal/domain/enums"

const (
    Pending   enums.OrderStatus = 1
    Active    enums.OrderStatus = 2
    Completed enums.OrderStatus = 3
    Cancelled enums.OrderStatus = 4
)
```

Usage:
```go
import (
    "myproject/internal/domain/enums"
    "myproject/internal/domain/enums/orderstatus"
)

type Order struct {
    StatusId int
}

func (o *Order) IsActive() bool {
    return orderstatus.Active.Equals(o.StatusId)
}
```

---

### String-based enum with String() method

When an enum needs to be converted back to a string representation (e.g., for logging or event publishing):

**File 1 — Type definition**: `enums/event_type.go`

```go
package enums

import "strings"

type EventType string

func (e EventType) Equals(s string) bool {
    return strings.EqualFold(string(e), s)
}

func (e EventType) String() string {
    return string(e)
}
```

**File 2 — Values and parser**: `enums/eventtype/value.go`

```go
package eventtype

import (
    "strings"
    "myproject/internal/domain/enums"
)

const (
    Recognized    enums.EventType = "post_recognized"
    NotRecognized enums.EventType = "post_not_recognized"
    Unknown       enums.EventType = "unknown"
)

func GetEventType(s string) enums.EventType {
    switch strings.ToLower(s) {
    case "post_recognized":
        return Recognized
    case "post_not_recognized":
        return NotRecognized
    default:
        return Unknown
    }
}
```

---

### Minimal enum (no methods needed)

When an enum is only used for constant grouping in a specific layer and does not need comparison or parsing methods:

**File 1 — Type definition**: `internal/integration/webservice/enums/connection_type.go`

```go
package enums

type ConnectionType string
```

**File 2 — Values only**: `internal/integration/webservice/enums/connectiontype/value.go`

```go
package connectiontype

const (
    Instagram = "instagram"
    Facebook  = "facebook"
)
```

Note: Even minimal enums follow the 2-file pattern. The constants here are untyped strings because no typed comparison is needed — this is acceptable for integration-layer enums that are only used as simple string constants.

---

## Allowed Enum Methods

| Method | Purpose | When to add |
|---|---|---|
| `Equals(raw) bool` | Case-insensitive (string) or direct (int) comparison | When entities or other code compare against raw values |
| `String() string` | Returns string representation | When the enum value needs to be serialized, logged, or published |
| `GetXxx(s string) EnumType` | Parser/factory — converts raw input to typed enum | When raw strings from external sources need to be converted to the enum |

These are the ONLY kinds of methods allowed on enums. Anything beyond this is business logic and belongs on an entity.

---

## Layer Placement Examples

### Domain enum — used by entities
```
internal/domain/enums/
├── channel.go
├── channel/
│   └── value.go
├── media_type.go
└── mediatype/
    └── value.go
```
Used by: `entity.Task`, `entity.Assignment`, entrypoints, services — everyone can import from domain.

### Integration/webservice enum — used only by a webservice client
```
internal/integration/webservice/enums/
├── connection_type.go
└── connectiontype/
    └── value.go
```
Used by: only the webservice client in the integration layer. Domain does NOT use it, so it stays in the integration layer.

### Integration/entrypoint enum — used only by controllers/consumers
```
internal/integration/entrypoint/enums/
├── sort_order.go
└── sortorder/
    └── value.go
```
Used by: only entrypoint controllers for request parsing. Not a domain concept.

---

## Anti-Patterns

### Anti-Pattern 1 — Single-file enum (FORBIDDEN)

```go
// BAD — type and constants in the same file/package
package enums

type Channel string

const (
    Instagram Channel = "instagram"
    Youtube   Channel = "youtube"
)
```

This causes import pollution and potential circular imports. Always split into 2 files.

---

### Anti-Pattern 2 — Business logic on an enum (FORBIDDEN)

```go
// BAD — this is business logic, not enum logic
func (s OrderStatus) CalculateRefundWindow() time.Duration {
    if s == "completed" {
        return 30 * 24 * time.Hour
    }
    return 0
}

// BAD — domain rules don't belong on enums
func (c Channel) GetMaxPostsPerDay() int {
    switch c {
    case "instagram": return 10
    case "tiktok": return 5
    default: return 3
    }
}
```

These belong as methods on the entity that owns the business rule.

---

### Anti-Pattern 3 — Enum in the wrong layer (FORBIDDEN)

```go
// BAD — Channel is used by domain entities, but placed in integration layer
// File: internal/integration/entrypoint/enums/channel.go
type Channel string

// This forces domain to import from integration → VIOLATES dependency rule
```

If domain uses it, it MUST live in `domain/enums/`.

---

### Anti-Pattern 4 — Using raw strings instead of enums (FORBIDDEN)

```go
// BAD — magic strings scattered across the codebase
if task.Channel == "instagram" { ... }
if task.Channel == "Instagram" { ... }  // inconsistent casing = bugs

// GOOD — typed enum with case-insensitive comparison
if channel.Instagram.Equals(task.Channel) { ... }
```

---

### Anti-Pattern 5 — Enum with struct tags (FORBIDDEN)

```go
// BAD — enum values should never carry serialization tags
type Status struct {
    Value string `json:"value"`
}
```

Enums are simple type aliases (`string` or `int`), not structs. Serialization is an integration concern.

---

## Naming Conventions

| Element | Convention | Example |
|---|---|---|
| Type file | `snake_case.go` in `enums/` package | `enums/order_status.go` |
| Sub-package directory | `lowercase` (no underscores) | `enums/orderstatus/` |
| Value file | always `value.go` | `enums/orderstatus/value.go` |
| Type name | `PascalCase` | `OrderStatus` |
| Constant names | `PascalCase` | `Pending`, `Active`, `Unknown` |
| Parser function | `Get{TypeName}(input) enums.TypeName` | `GetOrderStatus(s string) enums.OrderStatus` |

---

## Quick Reference Checklist

When creating or reviewing an enum, verify every item:

- [ ] Follows the 2-file pattern: type in `enums/{name}.go`, values in `enums/{name}/value.go`
- [ ] Type file is in the `enums` package, value file is in its own sub-package
- [ ] Placed in the innermost layer that uses it
- [ ] Contains only enum logic (Equals, String, parser) — no business logic
- [ ] String-based enums use `strings.EqualFold` for case-insensitive comparison
- [ ] Parser functions return an `Unknown` fallback value for unrecognized input
- [ ] No struct tags on enum types
- [ ] Constants use `PascalCase` naming
- [ ] Parser function follows `Get{TypeName}` naming convention
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
