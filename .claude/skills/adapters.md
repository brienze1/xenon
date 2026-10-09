---
name: Adapters (Clean Architecture)
description: This skill should be used when the user asks to "create an adapter", "add an adapter interface", "connect layers", "create a persistence adapter", "create a publisher adapter", "create a webservice adapter", "create a repository adapter", or mentions anything about adapter interfaces in any adapter/ package across layers.
version: 1.0.0
---

# Adapters — Golang Clean Architecture

Adapters are interfaces that connect layers. They are the contracts that allow layers to communicate without creating direct dependencies. The implementing layer always owns the adapter.

---

## Core Rules

### Rule 1: Adapters Connect Layers

Every cross-layer communication MUST go through an adapter interface. No layer calls another layer's concrete implementation directly.

```
Entrypoint → (adapter) → Service
Service    → (adapter) → Use Case (implemented by Integration)
Persistence → (adapter) → Repository (within Integration)
```

---

### Rule 2: The Implementing Layer Owns the Adapter

The adapter interface MUST live in the layer that **implements** it, not the layer that calls it.

| Caller | Implementer (owns adapter) | Adapter location |
|---|---|---|
| Entrypoint | Application service | `internal/application/adapter/` |
| Application service | Integration (persistence, webservice, publisher) | `internal/integration/adapter/` ... but see Rule 3 |
| Persistence | Repository (infra) | `internal/integration/adapter/` |

**Exception — Use case interfaces**: When the integration layer implements a use case, the use case interface itself (in `application/usecase/`) already serves as the contract. The integration adapter simply aggregates use case interfaces (see Rule 3).

---

### Rule 3: Integration Adapters Aggregate Use Case Interfaces

When the integration layer implements use cases, the adapter MUST be an interface that embeds the use case interface(s). This avoids duplicating method signatures.

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// File: internal/integration/adapter/user_persistence.go
package adapter

import "myproject/internal/application/usecase"

type UserPersistence interface {
    usecase.FindUser
    usecase.SaveUser
}
```

This means `UserPersistence` automatically has all methods from `FindUser` and `SaveUser`. The concrete persistence struct implements these methods, and the adapter groups them under one type for dependency injection.

```go
// File: internal/integration/adapter/email_publisher.go
package adapter

import "myproject/internal/application/usecase"

type EmailPublisher interface {
    usecase.SendNotification
}
```

```go
// File: internal/integration/adapter/payment_webservice.go
package adapter

import "myproject/internal/application/usecase"

type PaymentWebservice interface {
    usecase.ProcessPayment
}
```

---

### Rule 4: Naming Convention

All adapters follow the pattern: `adapter.{Name}{Type}`

| Type suffix | When to use | Example |
|---|---|---|
| `Service` | Application service adapter | `adapter.ProcessOrderService` |
| `Persistence` | Database persistence adapter | `adapter.UserPersistence` |
| `Repository` | Infrastructure repository adapter | `adapter.UserRepository` |
| `Webservice` | External API client adapter | `adapter.PaymentWebservice` |
| `Publisher` | Message broker publisher adapter | `adapter.EventPublisher` |
| `Factory` | Factory adapter | `adapter.ValidatorFactory` |
| `Handler` | Entrypoint handler adapter | `adapter.Handler` |
| `Validator` | Validation adapter | `adapter.Validator` |
| (infra) Various | Infrastructure dependency (only when lib has no interface) | `adapter.CloudStorage` |

The package is always `adapter`, so usage reads as: `adapter.UserPersistence`, `adapter.ProcessOrderService`, etc.

**Note on infrastructure adapters** (`internal/infra/adapter/`): These are only created when a library does NOT provide a suitable interface for mocking/testing. If the library already provides a common interface (e.g., `*gorm.DB`, `redis.UniversalClient`, `http.Client`), no infra adapter is needed — use the library's interface directly. See the Dependency Injection skill for details.

---

### Rule 5: Allowed Types Per Layer

What types an adapter can use depends on which layers it connects.

**Application layer adapters** (`internal/application/adapter/`):
- Domain entities, enums, value objects
- Primitives (`string`, `int64`, `context.Context`, etc.)
- Other application-layer types (use case interfaces, other adapter interfaces)
- NEVER DTOs, models, or integration types

```go
// GOOD — application adapter using domain types only
package adapter

import (
    "context"
    "myproject/internal/domain/entity"
    "myproject/internal/domain/enums"
)

type ProcessOrderService interface {
    Process(ctx context.Context, source string, eventType enums.EventType, metadata entity.OrderMetadata) error
}
```

**Integration adapters implementing use cases** (`internal/integration/adapter/`):
- Only embed use case interfaces — no custom method signatures needed
- Domain entities, enums, value objects (inherited from use case signatures)

```go
// GOOD — integration adapter aggregating use cases
package adapter

import "myproject/internal/application/usecase"

type OrderPersistence interface {
    usecase.FindOrder
    usecase.SaveOrder
}
```

**Integration-to-integration adapters** (`internal/integration/adapter/`):
- Domain entities, enums, value objects
- Primitives
- DTOs, models, and integration-specific types ARE allowed (they stay within the integration layer, so no leaking)

```go
// GOOD — integration-to-integration adapter can use DTOs
package adapter

import (
    "context"
    "myproject/internal/integration/entrypoint/dto"
    domain "myproject/pkg/domain/value-objects"
)

type Handler interface {
    Handle(ctx context.Context, message domain.QueueMessage[dto.Event]) error
}
```

```go
// GOOD — generic validator within integration layer
package adapter

import "context"

type Validator[T any] interface {
    Validate(ctx context.Context, request *T) error
}
```

---

## Code Examples

### Example 1 — Application adapter for a service

```go
// File: internal/application/adapter/process_order.go
package adapter

import (
    "context"
    "myproject/internal/domain/entity"
    "myproject/internal/domain/enums"
)

type ProcessOrderService interface {
    Process(ctx context.Context, source string, eventType enums.EventType, metadata entity.OrderMetadata) error
}
```

Implemented by: `internal/application/service/process_order.go`

---

### Example 2 — Application adapter for a validator

```go
// File: internal/application/adapter/post_validator.go
package adapter

import (
    "context"
    "myproject/internal/domain/entity"
    "myproject/internal/domain/enums"
)

type PostValidator interface {
    IsValid(ctx context.Context, eventType enums.EventType, metadata entity.PostMetadata, assignment entity.Assignment) bool
}
```

Implemented by: services like `reelsValidator`, `storyValidator` in `internal/application/service/`

---

### Example 3 — Application adapter for a factory

```go
// File: internal/application/adapter/validator_factory.go
package adapter

import (
    "context"
    "myproject/internal/domain/enums"
)

type ValidatorFactory interface {
    Build(ctx context.Context, mediaType enums.MediaType) PostValidator
}
```

```go
// File: internal/application/adapter/downloader_factory.go
package adapter

import (
    "context"
    "myproject/internal/application/usecase"
    "myproject/internal/domain/enums"
)

type DownloaderFactory interface {
    Build(ctx context.Context, mediaType enums.MediaType) usecase.DownloadMedia
}
```

Key points:
- Factory returns another adapter interface or a use case interface.
- Only uses domain types in signatures.

---

### Example 4 — Integration adapter aggregating use cases (persistence)

```go
// File: internal/integration/adapter/user_persistence.go
package adapter

import "myproject/internal/application/usecase"

type UserPersistence interface {
    usecase.FindUser
    usecase.SaveUser
}
```

```go
// File: internal/integration/adapter/credential_persistence.go
package adapter

import "myproject/internal/application/usecase"

type CredentialPersistence interface {
    usecase.GetCredential
}
```

Key points:
- No method signatures — they come from the embedded use case interfaces.
- One adapter can aggregate multiple use cases if the same persistence layer implements them.

---

### Example 5 — Integration adapter aggregating use cases (webservice)

```go
// File: internal/integration/adapter/payment_webservice.go
package adapter

import "myproject/internal/application/usecase"

type PaymentWebservice interface {
    usecase.ProcessPayment
}
```

---

### Example 6 — Integration adapter aggregating use cases (publisher)

```go
// File: internal/integration/adapter/event_publisher.go
package adapter

import "myproject/internal/application/usecase"

type EventPublisher interface {
    usecase.ProcessEvent
}
```

---

### Example 7 — Integration-to-integration adapter (handler with DTOs)

```go
// File: internal/integration/adapter/handler.go
package adapter

import (
    "context"
    domain "myproject/pkg/domain/value-objects"
    "myproject/internal/integration/entrypoint/dto"
)

type Handler interface {
    Handle(ctx context.Context, message domain.QueueMessage[dto.Event]) error
}
```

Key points:
- Uses DTOs and integration-specific types — allowed because both caller and implementer are in the integration layer.
- No domain type leakage — DTOs stay within the integration layer.

---

### Example 8 — Integration-to-integration adapter (generic validator)

```go
// File: internal/integration/adapter/event_validator.go
package adapter

import "context"

type Validator[T any] interface {
    Validate(ctx context.Context, request *T) error
}
```

Key points:
- Generic adapter — reusable across different validation targets within the integration layer.

---

### Example 9 — Integration-to-integration adapter (error handler)

```go
// File: internal/integration/adapter/error_handler.go
package adapter

import "context"

type ErrorHandler interface {
    Handle(ctx context.Context, err error, event any) error
}
```

---

## Adapter Map — Where Adapters Live

```
internal/
├── application/
│   └── adapter/                         ← Application-level adapters
│       ├── process_order.go             ← adapter.ProcessOrderService (implemented by service)
│       ├── post_validator.go            ← adapter.PostValidator (implemented by service)
│       ├── validator_factory.go         ← adapter.ValidatorFactory (implemented by factory)
│       └── downloader_factory.go        ← adapter.DownloaderFactory (implemented by factory)
│
└── integration/
    └── adapter/                         ← Integration-level adapters
        ├── user_persistence.go          ← adapter.UserPersistence (aggregates use cases)
        ├── credential_persistence.go    ← adapter.CredentialPersistence (aggregates use cases)
        ├── payment_webservice.go        ← adapter.PaymentWebservice (aggregates use cases)
        ├── event_publisher.go           ← adapter.EventPublisher (aggregates use cases)
        ├── handler.go                   ← adapter.Handler (integration-to-integration)
        ├── event_validator.go           ← adapter.Validator (integration-to-integration)
        └── error_handler.go            ← adapter.ErrorHandler (integration-to-integration)
│
└── infra/
    └── adapter/                             ← Infrastructure-level adapters (only when lib has no interface)
        └── cloud_storage.go             ← adapter.CloudStorage (wraps SDK with no common interface)
```

**Note:** Infra adapters are rare. Most infrastructure dependencies (GORM, Redis, HTTP) already provide suitable interfaces. Only create an infra adapter when you need to wrap a library that has no testable interface.

---

## Anti-Patterns

### Anti-Pattern 1 — Direct layer call without adapter (FORBIDDEN)

```go
// BAD — entrypoint importing service struct directly
import "myproject/internal/application/service"

type handler struct {
    svc *service.processOrder  // WRONG — must use adapter interface
}

// GOOD
import "myproject/internal/application/adapter"

type handler struct {
    svc adapter.ProcessOrderService
}
```

---

### Anti-Pattern 2 — Adapter in the wrong layer (FORBIDDEN)

```go
// BAD — adapter for a service living in the integration layer
// File: internal/integration/adapter/process_order_service.go  ← WRONG LOCATION
type ProcessOrderService interface { ... }

// GOOD — service adapter lives in the application layer (the implementing layer)
// File: internal/application/adapter/process_order.go
type ProcessOrderService interface { ... }
```

---

### Anti-Pattern 3 — Duplicating use case methods in integration adapter (FORBIDDEN)

```go
// BAD — rewriting methods instead of embedding
package adapter

import (
    "context"
    "myproject/internal/domain/entity"
)

type UserPersistence interface {
    FindById(ctx context.Context, userId string) (*entity.User, error)  // WRONG — duplicate
    FindByEmail(ctx context.Context, email string) (*entity.User, error)  // WRONG — duplicate
}

// GOOD — embed the use case interface
type UserPersistence interface {
    usecase.FindUser  // Already has FindById and FindByEmail
}
```

---

### Anti-Pattern 4 — Application adapter using DTOs (FORBIDDEN)

```go
// BAD — application adapter leaking integration types
package adapter

import "myproject/internal/integration/persistence/model"

type FindUserService interface {
    Find(ctx context.Context, userId string) (*model.UserDocument, error)  // WRONG
}
```

Application adapters can only use domain types and primitives.

---

### Anti-Pattern 5 — Missing adapter for cross-layer call (FORBIDDEN)

```go
// BAD — service calling persistence directly without adapter
type processOrder struct {
    db *gorm.DB  // WRONG — direct infrastructure dependency
}

// GOOD — service depends on use case interface, persistence implements it
type processOrder struct {
    findOrder usecase.FindOrder  // Implemented by persistence via adapter
}
```

---

## Quick Reference Checklist

When creating or reviewing an adapter, verify every item:

- [ ] Adapter is an interface (never a struct)
- [ ] Lives in the `adapter/` package of the **implementing** layer
- [ ] Naming follows `adapter.{Name}{Type}` (e.g., `UserPersistence`, `ProcessOrderService`)
- [ ] Integration adapters implementing use cases embed use case interfaces (no method duplication)
- [ ] Application adapters use only domain types and primitives
- [ ] Integration-to-integration adapters may use DTOs/models (stays within the layer)
- [ ] Every cross-layer call goes through an adapter — no direct imports of concrete types
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
