---
name: Use Cases (Clean Architecture)
description: This skill should be used when the user asks to "create a usecase", "add a usecase", "define a usecase", "create a use case interface", "add application logic", "implement a usecase function", or mentions anything about usecases or use cases in the internal/application/usecase/ package.
version: 1.0.0
---

# Use Cases — Golang Clean Architecture

This skill governs how use cases must be structured, where they live, what they can depend on, and when to use interfaces vs plain functions. Use cases are the application layer's contract for behavior.

---

## Core Rules

### Rule 1: Location

All use cases MUST live in `internal/application/usecase/` and declare `package usecase`.

```
internal/
└── application/
    └── usecase/
        ├── find_user.go
        ├── save_order.go
        ├── process_payment.go
        └── calculate_score.go
```

---

### Rule 2: Single Responsibility (SOLID)

Each use case MUST have a **single responsibility**. This does NOT mean one method — it means one cohesive purpose that may have multiple methods expressing different ways to fulfill that purpose.

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// GOOD — single responsibility (find user), multiple methods
type FindUser interface {
    FindById(ctx context.Context, userId string) (*entity.User, error)
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
}

// GOOD — single responsibility (save order), one method
type SaveOrder interface {
    Save(ctx context.Context, order entity.Order) error
}
```

```go
// BAD — mixed responsibilities (find + delete = two different concerns)
type UserManager interface {
    FindById(ctx context.Context, userId string) (*entity.User, error)
    Delete(ctx context.Context, userId string) error
}
```

The test: if you can describe the use case with one verb/action (find, save, process, download, validate), it has a single responsibility. If you need "and" to describe it, split it.

---

### Rule 3: Naming — No "UseCase" Suffix

Use case names MUST NOT include "UseCase" in the name. The package already communicates that.

```go
// BAD
type FindUserUseCase interface { ... }
// Usage: usecase.FindUserUseCase — redundant

// GOOD
type FindUser interface { ... }
// Usage: usecase.FindUser — clean and readable
```

The naming pattern is: `usecase.{ActionNoun}` — e.g., `usecase.FindUser`, `usecase.SavePost`, `usecase.ProcessPayment`.

---

### Rule 4: Domain Types Only

Use cases MUST only use domain types in their signatures. They cannot reference DTOs, models, persistence structs, or any type from integration/infrastructure layers.

Allowed types in use case signatures:
- Go standard library (`context.Context`, `string`, `int64`, `[]byte`, `error`, etc.)
- Domain entities (`entity.User`, `entity.Order`, etc.)
- Domain enums (`enums.Channel`, `enums.MediaType`, etc.)
- Domain value objects

```go
// GOOD — only domain types
type GetCredential interface {
    GetByCreatorId(ctx context.Context, creatorId int64) (*entity.Credential, error)
}

// GOOD — domain entity and enum
type ProcessEvent interface {
    Execute(ctx context.Context, metadata entity.EventMetadata, eventType enums.EventType) error
}

// BAD — integration layer DTO in signature
type GetCredential interface {
    GetByCreatorId(ctx context.Context, creatorId int64) (*dto.CredentialResponse, error)  // WRONG
}

// BAD — persistence model in signature
type FindUser interface {
    FindById(ctx context.Context, userId string) (*model.UserDocument, error)  // WRONG
}
```

---

### Rule 5: Interface vs Plain Function

Use cases come in two forms depending on whether they need an external implementation.

**Interface use case** — when the behavior requires integration layer implementation (database, API calls, message broker, file storage, etc.). The interface is defined in the application layer and implemented in the integration layer.

```go
// File: internal/application/usecase/find_user.go
package usecase

import (
    "context"
    "myproject/internal/domain/entity"
)

type FindUser interface {
    FindById(ctx context.Context, userId string) (*entity.User, error)
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
}
```

**Plain function use case** — when the behavior is pure logic that operates on domain types without needing I/O or external dependencies. No interface needed.

```go
// File: internal/application/usecase/calculate_score.go
package usecase

func CalculateScore(visualSimilarity, audioSimilarity, visualWeight, audioWeight float64) float64 {
    if visualWeight+audioWeight != 1 {
        panic("visualWeight and audioWeight must sum to 1")
    }
    if visualWeight < 0 || visualWeight > 1 {
        panic("visualWeight must be between 0 and 1")
    }
    if audioWeight < 0 || audioWeight > 1 {
        panic("audioWeight must be between 0 and 1")
    }

    return (visualSimilarity*visualWeight + audioSimilarity*audioWeight) / ((visualWeight + audioWeight) * 100)
}
```

**Decision logic:**
- Needs a database, API, file system, or message broker? → **Interface**
- Pure domain logic, calculations, or validation between domain types? → **Plain function**

---

### Rule 6: No Use Case Calling Another Use Case (FORBIDDEN)

Use cases MUST NOT call other use cases. Each use case is an independent, atomic unit of behavior. Orchestration of multiple use cases is the responsibility of a **service** in the application layer.

```go
// BAD — use case calling another use case
type ProcessOrder interface {
    Process(ctx context.Context, orderId string) error
}

// Inside the implementation:
func (p *processOrder) Process(ctx context.Context, orderId string) error {
    user, _ := p.findUser.FindById(ctx, order.UserId)  // WRONG — calling another use case
    // ...
}
```

If you need to combine multiple use cases, that orchestration belongs in a service (see the services skill).

---

### Rule 7: No Use Case Calling Services (FORBIDDEN)

Use cases MUST NOT call services. Services orchestrate use cases, not the other way around.

```go
// BAD — use case depending on a service
type ProcessPayment interface {
    Process(ctx context.Context, payment entity.Payment) error
}

// Inside the implementation:
func (p *processPayment) Process(ctx context.Context, payment entity.Payment) error {
    p.orderService.ValidateOrder(ctx, payment.OrderId)  // WRONG — calling a service
    // ...
}
```

---

## Code Examples

### Example 1 — Simple interface use case (single method)

```go
package usecase

import (
    "context"
    "myproject/internal/domain/entity"
)

type GetCredential interface {
    GetByCreatorId(ctx context.Context, creatorId int64) (*entity.Credential, error)
}
```

Key points:
- One responsibility: get a credential.
- Returns a domain entity pointer (nullable).
- Implementation will be provided by a persistence adapter in the integration layer.

---

### Example 2 — Interface use case with multiple methods (same responsibility)

```go
package usecase

import (
    "context"
    "myproject/internal/domain/entity"
)

type FindUser interface {
    FindById(ctx context.Context, userId string) (*entity.User, error)
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
}
```

Key points:
- Single responsibility: find a user.
- Two methods: different lookup strategies for the same purpose.
- Both return the same domain entity — confirms it's one cohesive concern.

---

### Example 3 — Interface use case with mixed domain types

```go
package usecase

import (
    "context"
    "myproject/internal/domain/entity"
)

type DownloadMedia interface {
    Download(ctx context.Context, postId string, credential entity.Credential) ([]byte, *string, *entity.MediaInfo, error)
}
```

Key points:
- Accepts a domain entity (`entity.Credential`) as input.
- Returns raw bytes, a pointer string, and a domain entity (`entity.MediaInfo`).
- All types are either primitives or domain types — no integration types leak in.

---

### Example 4 — Interface use case with entity and enum

```go
package usecase

import (
    "context"
    "myproject/internal/domain/entity"
    "myproject/internal/domain/enums"
)

type ProcessEvent interface {
    Execute(ctx context.Context, savedUrl, approvedUrl, source string, metadata entity.EventMetadata, eventType enums.EventType) error
}
```

Key points:
- Mixes primitives (`string`), domain entity, and domain enum in the signature.
- All domain types — no leakage from other layers.

---

### Example 5 — Interface use case with enum parameter

```go
package usecase

import (
    "context"
    "myproject/internal/domain/enums"
)

type SaveMedia interface {
    Save(ctx context.Context, data []byte, fileName string, mediaType enums.MediaType, projectId, userId int64) (*string, error)
}
```

Key points:
- Uses a domain enum (`enums.MediaType`) alongside primitives.
- Returns a pointer string (nullable result, e.g., a URL or path).

---

### Example 6 — Plain function use case (no interface needed)

```go
package usecase

import "myproject/internal/domain/enums"

func ShouldAlert(channel enums.Channel, mediaType enums.MediaType, minScore, maxScore, visualWeight, audioWeight float64) bool {
    if minScore < 0 || minScore > 1 {
        panic("minScore must be between 0 and 1")
    }
    if maxScore < 0 || maxScore > 1 {
        panic("maxScore must be between 0 and 1")
    }
    if minScore > maxScore {
        panic("minScore cannot be greater than maxScore")
    }
    if visualWeight+audioWeight != 1 {
        panic("visualWeight and audioWeight must sum to 1")
    }

    score := (visualWeight + audioWeight) / 2
    return score < minScore || score > maxScore
}
```

Key points:
- Pure logic — no I/O, no external dependencies.
- Only uses domain enums and primitives.
- No interface needed — it's a standalone function.
- Panics on invalid invariants (developer errors, not user errors).

---

### Example 7 — Plain function use case (calculation)

```go
package usecase

func CalculateWeightedScore(visualSimilarity, audioSimilarity, visualWeight, audioWeight float64) float64 {
    if visualWeight+audioWeight != 1 {
        panic("weights must sum to 1")
    }

    return (visualSimilarity*visualWeight + audioSimilarity*audioWeight) / ((visualWeight + audioWeight) * 100)
}
```

Key points:
- Pure math — operates only on primitives.
- No interface, no struct, no dependencies.
- Self-contained business rule expressed as a function.

---

## Anti-Patterns

### Anti-Pattern 1 — "UseCase" suffix in name (FORBIDDEN)

```go
// BAD
type FindUserUseCase interface { ... }
type SaveOrderUseCase interface { ... }

// GOOD
type FindUser interface { ... }
type SaveOrder interface { ... }
```

The package name (`usecase`) already provides the context.

---

### Anti-Pattern 2 — Mixed responsibilities (FORBIDDEN)

```go
// BAD — find + delete = two different responsibilities
type UserRepository interface {
    FindById(ctx context.Context, userId string) (*entity.User, error)
    Delete(ctx context.Context, userId string) error
    Update(ctx context.Context, user entity.User) error
}
```

Split into separate use cases:
```go
// GOOD
type FindUser interface {
    FindById(ctx context.Context, userId string) (*entity.User, error)
}

type DeleteUser interface {
    Delete(ctx context.Context, userId string) error
}

type UpdateUser interface {
    Update(ctx context.Context, user entity.User) error
}
```

---

### Anti-Pattern 3 — Integration types in signature (FORBIDDEN)

```go
// BAD — DTO from integration layer
import "myproject/internal/integration/http/dto"

type FindUser interface {
    FindById(ctx context.Context, userId string) (*dto.UserResponse, error)  // WRONG
}

// BAD — persistence model
import "myproject/internal/integration/persistence/model"

type FindUser interface {
    FindById(ctx context.Context, userId string) (*model.UserDocument, error)  // WRONG
}
```

Use cases speak the domain language only.

---

### Anti-Pattern 4 — Use case calling another use case (FORBIDDEN)

```go
// BAD — use case depends on another use case
type ProcessOrder struct {
    findUser   usecase.FindUser    // WRONG
    findOrder  usecase.FindOrder   // WRONG
}

func (p *ProcessOrder) Process(ctx context.Context, orderId string) error {
    order, _ := p.findOrder.FindById(ctx, orderId)
    user, _ := p.findUser.FindById(ctx, order.UserId)
    // ...
}
```

This orchestration belongs in a service, not a use case.

---

### Anti-Pattern 5 — Use case calling a service (FORBIDDEN)

```go
// BAD — use case depends on a service
type ValidatePayment struct {
    orderService service.OrderService  // WRONG — reverse dependency
}
```

Services call use cases, never the other way around.

---

### Anti-Pattern 6 — Interface when a plain function suffices (UNNECESSARY)

```go
// BAD — unnecessary interface for pure logic
type ScoreCalculator interface {
    Calculate(visual, audio float64) float64
}

// GOOD — just a function
func CalculateScore(visual, audio float64) float64 {
    return (visual + audio) / 2
}
```

If there's no I/O and no need for polymorphism, use a plain function.

---

## Decision Flowchart

```
Does the use case need I/O (DB, API, filesystem, message broker)?
├── YES → Interface use case
│         • Define interface in application/usecase/
│         • Implementation lives in integration layer
│         • Injected via dependency injection
└── NO  → Does it operate purely on domain types with logic?
          ├── YES → Plain function use case
          │         • Define as exported function in application/usecase/
          │         • No interface, no struct
          └── NO  → It's probably not a use case
                    • Consider if it belongs on an entity method instead
```

---

## Quick Reference Checklist

When creating or reviewing a use case, verify every item:

- [ ] File is in `internal/application/usecase/` and declares `package usecase`
- [ ] Name does NOT include "UseCase" suffix
- [ ] Has a single responsibility (one verb/action describes it)
- [ ] Signature uses only domain types (entities, enums, value objects) and primitives
- [ ] No imports from integration or infrastructure layers
- [ ] Does NOT call other use cases
- [ ] Does NOT call services
- [ ] Interface form: used when I/O or external implementation is needed
- [ ] Function form: used when behavior is pure logic with no external dependencies
- [ ] Multiple methods on an interface share the same responsibility (different strategies, same purpose)
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
