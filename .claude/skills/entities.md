---
name: Entities (Clean Architecture)
description: This skill should be used when the user asks to "create a domain entity", "add an entity", "write a domain model", "implement a domain object", "add business logic to an entity", "refactor anemic entity", "move business rules to entity", "create a value object", "define a domain enum", or mentions anything about entities in the internal/domain/entity/ package.
version: 1.0.0
---

# Domain Entities — Golang DDD + Clean Architecture

This skill governs how domain entities must be written in this project. It covers location rules, structural constraints, composition rules, and — most critically — the rich domain model requirement that forbids anemic entities.

---

## Core Rules

### Rule 1: Location

All entities MUST live in the `internal/domain/entity/` package.

```
internal/
└── domain/
    └── entity/
        ├── credential.go
        ├── task.go
        ├── schedule.go
        └── assignment.go
```

Every file in this package must declare `package entity` at the top.

**One struct per file — always.** Each entity gets its own file. If an entity is composed of other entities or value objects that are also structs, each of those gets its own file too. Never put multiple struct definitions in a single file.

```
// GOOD — one struct per file
entity/
├── profile.go           ← type Profile struct
├── workout.go           ← type Workout struct
├── workout_step.go      ← type WorkoutStep struct
└── credential.go        ← type Credential struct

// BAD — multiple structs in one file (FORBIDDEN)
entity/
└── workout.go           ← type Workout struct + type WorkoutStep struct  ← WRONG
```

---

### Rule 2: No Integration Tags

Entities MUST NOT carry any struct tags from integration or infrastructure layers. The domain layer knows nothing about how data is serialized or persisted.

Forbidden tags: `json:`, `bson:`, `gorm:`, `xml:`, `yaml:`, `db:`, `validate:`, `mapstructure:`, `form:`, and any other tag that belongs to an external library.

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// BAD — DO NOT DO THIS
type User struct {
    ID   int64  `json:"id" bson:"_id" gorm:"primaryKey"`
    Name string `json:"name" bson:"name"`
}

// GOOD
type User struct {
    ID   int64
    Name string
}
```

Tag concerns belong in integration-layer DTOs, not in domain entities.

---

### Rule 3: Layer Decoupling — Allowed Imports Only

Entities MUST NOT import anything from the application, integration, or infrastructure layers. Violating this rule creates upward dependencies that break Clean Architecture.

Allowed imports:
- Go standard library (`time`, `strings`, `math`, etc.)
- `internal/domain/enums/` and its sub-packages
- Other entities within `internal/domain/entity/` (same package, same bounded context)
- Shared domain utilities (e.g., `pkg/timeutils`) that have no external dependencies

```go
// BAD — importing from outer layers
import (
    "myproject/internal/infra/properties"       // WRONG — infrastructure layer
    "myproject/internal/integration/http/dto"   // WRONG — integration layer
    "myproject/internal/application/service"    // WRONG — application layer
)

// GOOD — only domain and stdlib
import (
    "time"
    "myproject/internal/domain/enums"
    "myproject/internal/domain/enums/mediatype"
    "myproject/pkg/timeutils"
)
```

---

### Rule 4: Allowed Composition

An entity's fields may be composed of:

| Allowed type | Example |
|---|---|
| Primitive types | `string`, `int`, `int64`, `bool`, `float64` |
| Pointer variants | `*string`, `*int64`, `*time.Time` |
| Standard library types | `time.Time` |
| Domain enums | `enums.Channel`, `enums.MediaType` |
| Value objects (domain layer) | types defined in `internal/domain/` |
| Other entities (same bounded context) | `Schedule`, `Project`, `Credential` inside `Assignment` |

Entities MUST NOT embed structs from other layers (e.g., a GORM `Model`, a MongoDB document type, or an HTTP response struct).

---

### Rule 5: Rich Domain Model (NOT Anemic) — CRITICAL

This is the most important rule. Entities must OWN their business logic. An entity that is nothing but fields with getters and setters is an anemic model — it is forbidden.

**The rule of thumb:** If you find yourself writing `if entity.SomeField == "someValue"` inside a service or use case, that check belongs as a method on the entity (`entity.IsSomeState() bool`).

Business behavior that lives on entities:
- State checks and predicates (`IsActive`, `IsPending`, `IsValid`, `IsExpired`)
- Equality and identity comparisons (`Equals`)
- Calculations and derivations (`GetMediaType`, `TotalAmount`)
- Delegation to composed sub-entities (`IsAccountConnected` delegating to `Credential.IsValid`)
- Collection queries (`HasPendingTasksOfIdTypeAndChannel` iterating over `[]Task`)

Entities MUST NOT call repositories, external services, or anything that requires I/O. Side-effect-free methods only.

```go
// BAD — entity calling a repository (belongs in application layer)
func (o *Order) Save(repo OrderRepository) error {
    return repo.Save(o)
}

// BAD — logic leaking into a service instead of living on the entity
// OrderService:
if order.Status == "active" && order.Total > 0 { ... }

// GOOD — logic on the entity, service delegates
func (o *Order) IsReadyToFulfill() bool {
    return o.Status == "active" && o.Total > 0
}
// OrderService:
if order.IsReadyToFulfill() { ... }
```

---

## Receiver Style Guidelines

- Use **pointer receivers** (`*Entity`) for entities that represent mutable state or are large structs where copying is wasteful.
- Use **value receivers** (`Entity`) for small, immutable entities where a copy is acceptable and the method does not mutate state.

Be consistent within a type: if any method uses a pointer receiver, all methods should use pointer receivers.

---

## Code Examples

### Example 1 — Simple entity with validation logic

```go
package entity

import (
    "time"
    "myproject/pkg/timeutils"
)

type Credential struct {
    Token          *string
    ConnectionType *string
    ExpiresAt      *time.Time
    UserId         int64
}

func (c *Credential) IsValid() bool {
    if c.Token == nil || c.ConnectionType == nil {
        return false
    }
    if c.ExpiresAt != nil && c.ExpiresAt.Before(timeutils.Now()) {
        return false
    }
    return true
}
```

Key points:
- No struct tags.
- Pointer fields model optional/nullable data without a separate "null" sentinel.
- `IsValid` encapsulates all validity rules — callers never inspect fields directly.

---

### Example 2 — Entity with enum composition and equality check

```go
package entity

import (
    "time"
    "myproject/internal/domain/enums"
)

type Task struct {
    Schedule  Schedule
    MediaUrl  *string
    PostedAt  *time.Time
    MediaType enums.MediaType
    Channel   enums.Channel
    TaskId    int64
}

func (t Task) IsPending() bool {
    return t.PostedAt == nil && t.Schedule.IsActive() && t.MediaUrl != nil
}

func (t Task) Equals(taskId int64, mediaType enums.MediaType, channel enums.Channel) bool {
    return t.TaskId == taskId && t.MediaType == mediaType && t.Channel == channel
}
```

Key points:
- `Schedule` is composed directly — same bounded context, same package.
- `IsPending` delegates to `Schedule.IsActive()` rather than re-implementing the check.
- `Equals` uses domain enum types, not raw strings, keeping comparisons type-safe.

---

### Example 3 — Entity with time-based business logic

```go
package entity

import (
    "time"
    "myproject/pkg/timeutils"
)

type Schedule struct {
    StartsAt *time.Time
    EndsAt   *time.Time
    Id       int64
}

func (s *Schedule) IsActive() bool {
    if s.StartsAt == nil || s.EndsAt == nil {
        return false
    }
    now := timeutils.Now()
    return s.StartsAt.Before(now) && s.EndsAt.After(now)
}
```

Key points:
- `timeutils.Now()` is a shared utility with no external dependencies — acceptable import.
- Nil guards come first; the method never panics on incomplete data.
- The caller (`Task.IsPending`) trusts `IsActive` completely — no field access from outside.

---

### Example 4 — Aggregate root with rich business logic

An aggregate root composes multiple entities from the same bounded context and exposes higher-level business queries.

```go
package entity

import (
    "myproject/internal/domain/enums"
    "myproject/internal/domain/enums/assignmentstatus"
)

type Assignment struct {
    Credential *Credential
    Status     string
    Tasks      []Task
    Project    Project
    UserId     int64
}

func (a *Assignment) IsActive() bool {
    return assignmentstatus.Active.Equals(a.Status)
}

func (a *Assignment) IsAccountConnected() bool {
    return a.Credential != nil && a.Credential.IsValid()
}

func (a *Assignment) IsProjectActive() bool {
    return a.Project.IsActive()
}

func (a *Assignment) HasPendingTasksOfIdTypeAndChannel(taskId int64, mediaType enums.MediaType, channel enums.Channel) bool {
    for _, task := range a.Tasks {
        if task.IsPending() && task.Equals(taskId, mediaType, channel) {
            return true
        }
    }
    return false
}
```

Key points:
- `IsActive` delegates status comparison to the `assignmentstatus` enum sub-package — no magic string in the entity body.
- `IsAccountConnected` guards the nil check AND delegates validity to `Credential.IsValid()`.
- `IsProjectActive` is a one-liner that delegates entirely to the composed `Project` entity.
- `HasPendingTasksOfIdTypeAndChannel` is a collection query that belongs on the aggregate — it combines `IsPending` and `Equals` from the `Task` entity.
- No service file needs to iterate over `Tasks` or inspect `Credential` fields directly.

---

### Example 5 — Domain enum: 2-file pattern

Enums follow a strict 2-file convention.

**File 1: type + methods** at `internal/domain/enums/channel.go`

```go
package enums

import "strings"

type Channel string

func (c Channel) Equals(s string) bool {
    return strings.EqualFold(string(c), s)
}
```

**File 2: constants + factory** at `internal/domain/enums/channel/value.go`

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

Key points:
- The `enums` package owns the type and its behavior methods (e.g., `Equals`).
- The `enums/{enumname}` sub-package owns the named constants and the factory function.
- This avoids circular imports: entities import `enums` for the type; they import `enums/channel` for constants and `GetChannel`.
- `GetChannel` provides a safe, centralized conversion from raw strings (e.g., from event payloads) to typed enum values.

---

### Example 6 — Entity with domain-level type conversion

```go
package entity

import (
    "time"
    "myproject/internal/domain/enums"
    "myproject/internal/domain/enums/mediatype"
)

type EventMetadata struct {
    OccurredAt time.Time
    EventId    string
    Channel    string
    MediaType  string
    ProjectId  int64
    UserId     int64
    TaskId     int64
}

func (m EventMetadata) GetMediaType() enums.MediaType {
    return mediatype.GetMediaType(m.MediaType)
}
```

Key points:
- Raw string fields (`Channel`, `MediaType`) are acceptable when the entity is receiving data from an external event payload.
- Conversion to typed enum values happens through a method on the entity, not scattered in use-case code.
- Callers use `m.GetMediaType()` instead of calling `mediatype.GetMediaType(m.MediaType)` everywhere.

---

### Example 7 — Entity with switch-based business rules

```go
package entity

type MediaInfo struct {
    ID        string
    MediaType string
    MediaURL  string
}

func (m *MediaInfo) ShouldProcess() bool {
    switch m.MediaType {
    case "VIDEO", "REELS":
        return true
    default:
        return false
    }
}
```

Key points:
- The decision of whether to process a media item is business logic — it belongs on the entity.
- Switch-based rules are readable and easy to extend.
- Callers ask `mediaInfo.ShouldProcess()` — they do not switch on `mediaInfo.MediaType` themselves.

---

## Anti-Patterns

### Anti-Pattern 1 — Integration tags on domain entities (FORBIDDEN)

```go
// BAD — DO NOT DO THIS
type User struct {
    ID   int64  `json:"id" bson:"_id" gorm:"primaryKey"`
    Name string `json:"name" bson:"name"`
}
```

Tags couple the domain model to serialization and persistence frameworks. When the persistence layer changes, the domain model must change too — a direct violation of Clean Architecture. Put tags on integration-layer DTOs and map them to entities.

---

### Anti-Pattern 2 — Anemic entity (FORBIDDEN)

```go
// BAD — just a data bag with no behavior
type Order struct {
    ID     int64
    Status string
    Total  float64
}

// All logic leaks into OrderService instead of living on the entity
func (s *OrderService) ProcessOrder(o *Order) {
    if o.Status == "active" && o.Total > 0 {
        // do something
    }
}
```

The entity should own these predicates. The correct version:

```go
// GOOD
func (o *Order) IsReadyToFulfill() bool {
    return o.Status == "active" && o.Total > 0
}
```

---

### Anti-Pattern 3 — Importing from outer layers (FORBIDDEN)

```go
// BAD — entity importing from infrastructure/integration layers
import (
    "myproject/internal/infra/properties"      // WRONG — infrastructure layer
    "myproject/internal/integration/dto"        // WRONG — integration layer
    "myproject/internal/application/service"    // WRONG — application layer
)
```

Dependencies must flow inward only: infrastructure → application → domain. The domain layer imports nothing from outer layers.

---

### Anti-Pattern 4 — Entity calling external services or repositories (FORBIDDEN)

```go
// BAD — entity should not perform I/O or call repositories
func (o *Order) Save(repo OrderRepository) error {
    return repo.Save(o)
}
```

Persistence is an infrastructure concern. Use cases and services in the application layer call repositories, not entities. Entities are pure domain objects with no side effects.

---

## Quick Reference Checklist

When creating or reviewing a domain entity, verify every item:

- [ ] File is in `internal/domain/entity/` and declares `package entity`
- [ ] **One struct per file** — no multiple struct definitions in the same file
- [ ] No struct tags of any kind (`json`, `bson`, `gorm`, `xml`, `yaml`, `db`, etc.)
- [ ] Imports are limited to stdlib, `internal/domain/`, and domain-safe `pkg/` utilities
- [ ] Entity is NOT anemic — it has methods that encode business rules
- [ ] State checks like `status == "x"` live as predicate methods (`IsX() bool`) on the entity
- [ ] Composed entities belong to the same bounded context
- [ ] Methods are side-effect-free — no I/O, no repository calls, no external service calls
- [ ] Pointer vs value receiver choice is consistent within the type
- [ ] Enums use the 2-file pattern (`enums/type.go` + `enums/{name}/value.go`)
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
