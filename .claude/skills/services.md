---
name: Services (Clean Architecture)
description: This skill should be used when the user asks to "create a service", "add a service", "implement a service", "orchestrate usecases", "create an application service", or mentions anything about services in the internal/application/service/ package or adapter interfaces in internal/application/adapter/.
version: 1.0.0
---

# Services — Golang Clean Architecture

This skill governs how application services must be structured, how they orchestrate use cases, and the strict rules they must follow. Services are the glue between use cases — they orchestrate, they do NOT own business logic.

---

## Core Rules

### Rule 1: Location

Services MUST live in `internal/application/service/` and declare `package service`.

```
internal/
└── application/
    ├── adapter/
    │   └── process_order.go        ← adapter.ProcessOrderService interface
    ├── service/
    │   └── process_order.go        ← service.processOrder struct + constructor
    └── usecase/
        └── ...
```

---

### Rule 2: Must Implement an Adapter Interface (Always)

Services MUST always implement an interface defined in `internal/application/adapter/`. They are NEVER used directly — always through the adapter interface.

**Adapter interface** — `internal/application/adapter/process_order.go`:
> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
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

**Service implementation** — `internal/application/service/process_order.go`:
```go
package service

type processOrder struct {
    findOrder  usecase.FindOrder
    saveOrder  usecase.SaveOrder
}

func NewProcessOrderService(
    findOrder usecase.FindOrder,
    saveOrder usecase.SaveOrder,
) adapter.ProcessOrderService {
    return &processOrder{
        findOrder: findOrder,
        saveOrder: saveOrder,
    }
}
```

---

### Rule 3: Naming Conventions

Services and their adapter interfaces follow a strict naming pattern:

| Element | Pattern | Example |
|---|---|---|
| Adapter interface | `adapter.{ProcessName}Service` | `adapter.ProcessOrderService` |
| Service struct | `service.{processName}` (unexported) | `service.processOrder` |
| Constructor | `service.New{ProcessName}Service(...)` | `service.NewProcessOrderService(...)` |
| Constructor return type | `adapter.{ProcessName}Service` | `adapter.ProcessOrderService` |

The word "Service" appears in the **adapter interface name** and the **constructor name**, but NOT in the struct name.

```go
// Adapter: adapter.FilterAndProcessPostsService
// Struct:  service.filterAndProcessPosts (no "Service")
// Constructor: service.NewFilterAndProcessPostsService(...) adapter.FilterAndProcessPostsService
```

---

### Rule 4: Domain Types Only

Service method signatures MUST only use domain types. Same rule as use cases — no DTOs, no models, no integration types.

```go
// GOOD — domain entity and enum
func (f *filterPosts) FilterAndProcess(ctx context.Context, source string, eventType enums.EventType, metadata entity.PostMetadata) error

// BAD — DTO from integration layer
func (f *filterPosts) FilterAndProcess(ctx context.Context, request dto.FilterRequest) error  // WRONG
```

---

### Rule 5: Orchestrate Use Cases — Do NOT Own Business Logic (CRITICAL)

This is the most important rule. Services orchestrate use cases and **delegate business decisions to domain entities**. A service MUST NOT contain business logic itself.

```go
// GOOD — delegates business logic to entity methods
if !invite.IsAccountConnected() {
    logger.Info(ctx, "Post ignored - account not connected", metadata)
    return nil
}

if !invite.IsCampaignActive() {
    logger.Info(ctx, "Campaign not active", metadata)
    return nil
}

if !invite.HasPendingTasksOfType(taskId, mediaType, channel) {
    logger.Info(ctx, "No pending tasks", metadata)
    return nil
}
```

```go
// BAD — business logic in the service (should be on the entity)
if invite.Status != "participating" {  // WRONG — should be invite.IsParticipating()
    return nil
}

if invite.Credential == nil || invite.Credential.Token == nil {  // WRONG — should be invite.IsAccountConnected()
    return nil
}
```

**The rule of thumb:** If you see a field access with a comparison (`entity.Field == "value"`) in a service, that check belongs as a method on the entity. The service should only call entity methods, never inspect entity internals.

---

### Rule 6: Logger — Started and Finished

Every service method MUST begin with a `Started` log and end with a `Finished` log, both including all raw input parameters.

```go
func (f *processOrder) Process(ctx context.Context, source string, eventType enums.EventType, metadata entity.OrderMetadata) error {
    logger.Info(ctx, "Started", metadata, map[string]any{"source": source, "event": eventType})

    // ... orchestration logic ...

    logger.Info(ctx, "Finished", metadata, map[string]any{"source": source, "event": eventType})
    return nil
}
```

Key points:
- `Started` is the first statement in the method.
- `Finished` is the last statement before the final `return nil`.
- Both include all input parameters for full traceability.
- If the method returns early due to an error, `Finished` is NOT logged (the error propagates up).

---

### Rule 7: Error Handling — Never Eat, Never Wrap

Services MUST pass through errors from use case calls without altering them. Error creation with codes, HTTP status, etc. belongs in the integration layer.

```go
// GOOD — pass through the error as-is
result, err := f.findOrder.FindById(ctx, orderId)
if err != nil {
    return err  // Pass through untouched
}

// GOOD — special handling for known error types (but still no wrapping)
result, err := f.findOrder.FindById(ctx, orderId)
if err != nil {
    if errs.IsNotFoundError(err) {
        logger.Info(ctx, "Order not found, ignoring", metadata)
        return nil  // Graceful handling of expected case
    }
    return err  // Unknown errors pass through
}

// BAD — wrapping/altering the error
result, err := f.findOrder.FindById(ctx, orderId)
if err != nil {
    return fmt.Errorf("failed to find order: %w", err)  // WRONG — don't wrap
}

// BAD — creating a new error from an existing one
result, err := f.findOrder.FindById(ctx, orderId)
if err != nil {
    return errors.New("order lookup failed")  // WRONG — lost the original error
}
```

Services CAN create new errors when domain logic demands it (e.g., a validation result from an entity), but never from a pre-existing error returned by another function.

---

### Rule 8: No Service Calling Another Service (FORBIDDEN)

Services MUST NOT call other services. If a service needs behavior from another service, either:
1. Extract the shared logic into a use case that both can use
2. Let the entrypoint (controller/consumer) orchestrate multiple service calls

```go
// BAD — service calling another service
type processOrder struct {
    validatePaymentService adapter.ValidatePaymentService  // WRONG
}

// GOOD — service uses use cases, entrypoint orchestrates services
type processOrder struct {
    findOrder usecase.FindOrder   // Use case
    saveOrder usecase.SaveOrder   // Use case
}
```

---

### Rule 9: Single Responsibility — One Process

Services have a single responsibility: orchestrating **one process**. A process may involve multiple use case calls, but it represents one cohesive workflow.

If a service exceeds ~200-250 lines, it's a signal that either:
- It contains business logic that should be on domain entities
- It could be broken into smaller use cases with focused logic

---

### Rule 10: Services Can Call Use Cases, Never Other Services

The dependency direction is clear:

```
Entrypoint → Service → Use Cases → (implemented by Integration Layer)
                  ↘ Entity methods (business logic delegation)
```

A service can depend on:
- Use case interfaces (`usecase.FindOrder`, `usecase.SaveOrder`)
- Adapter interfaces for factories or other application-level contracts (`adapter.ValidatorFactory`)
- Domain entities and enums (for method calls and parameters)
- Logger (`pkg/logger`)

A service MUST NOT depend on:
- Other services
- Integration layer types (DTOs, models, repositories directly)
- Infrastructure layer types (database clients, HTTP clients, etc.)

---

## Code Examples

### Example 1 — Full service with orchestration

**Adapter interface** — `internal/application/adapter/process_order.go`:
```go
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

**Service implementation** — `internal/application/service/process_order.go`:
```go
package service

import (
    "context"

    "myproject/internal/application/adapter"
    "myproject/internal/application/usecase"
    "myproject/internal/domain/entity"
    "myproject/internal/domain/enums"
    "myproject/pkg/errs"
    "myproject/pkg/logger"
)

type processOrder struct {
    getAssignment   usecase.GetAssignment
    validatorFactory adapter.ValidatorFactory
    downloadMedia   usecase.DownloadMedia
    saveMedia       usecase.SaveMedia
    processEvent    usecase.ProcessEvent
}

func NewProcessOrderService(
    getAssignment usecase.GetAssignment,
    validatorFactory adapter.ValidatorFactory,
    downloadMedia usecase.DownloadMedia,
    saveMedia usecase.SaveMedia,
    processEvent usecase.ProcessEvent,
) adapter.ProcessOrderService {
    return &processOrder{
        getAssignment:   getAssignment,
        validatorFactory: validatorFactory,
        downloadMedia:   downloadMedia,
        saveMedia:       saveMedia,
        processEvent:    processEvent,
    }
}

func (p *processOrder) Process(ctx context.Context, source string, eventType enums.EventType, metadata entity.OrderMetadata) error {
    logger.Info(ctx, "Started", metadata, map[string]any{"source": source, "event": eventType})

    assignment, err := p.getAssignment.GetByProjectAndUser(ctx, metadata.ProjectId, metadata.UserId)
    if err != nil {
        if errs.IsNotFoundError(err) {
            logger.Info(ctx, "Order ignored", metadata)
            return nil
        }
        return err
    }

    if !assignment.IsAccountConnected() {
        logger.Info(ctx, "Order ignored - account not connected", metadata)
        return nil
    }

    mediaType := metadata.GetMediaType()
    if !p.validatorFactory.Build(ctx, mediaType).IsValid(ctx, eventType, metadata, *assignment) {
        logger.Info(ctx, "Order ignored", metadata, assignment)
        return nil
    }

    data, fileName, mediaInfo, err := p.downloadMedia.Download(ctx, metadata.EventId, *assignment.Credential)
    if err != nil {
        return err
    }

    if mediaInfo != nil && !mediaInfo.ShouldProcess() {
        logger.Info(ctx, "Order ignored - non-processable media type", metadata, mediaInfo)
        return nil
    }

    savedPath, err := p.saveMedia.Save(ctx, data, *fileName, mediaType, metadata.ProjectId, metadata.UserId)
    if err != nil {
        return err
    }

    err = p.processEvent.Execute(ctx, *savedPath, assignment.GetMediaRawUrl(metadata.TaskId), source, metadata, eventType)
    if err != nil {
        return err
    }

    logger.Info(ctx, "Finished", metadata, map[string]any{"source": source, "event": eventType})
    return nil
}
```

Key points:
- Constructor returns the adapter interface, not the struct.
- Struct is unexported (`processOrder`, not `ProcessOrder`).
- `Started` and `Finished` logs bracket the method.
- All business checks delegate to entity methods (`assignment.IsAccountConnected()`, `mediaInfo.ShouldProcess()`).
- Errors from use cases pass through untouched.
- Not-found errors are handled gracefully (known domain case).

---

### Example 2 — Smaller service implementing a validator

**Adapter interface** — `internal/application/adapter/post_validator.go`:
```go
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

**Service implementation** — `internal/application/service/reels_validator.go`:
```go
package service

import (
    "context"

    "myproject/internal/application/adapter"
    "myproject/internal/domain/entity"
    "myproject/internal/domain/enums"
    "myproject/internal/domain/enums/channel"
    "myproject/internal/domain/enums/eventtype"
    "myproject/internal/domain/enums/mediatype"
    "myproject/pkg/logger"
)

type reelsValidator struct{}

func NewReelsValidator() adapter.PostValidator {
    return &reelsValidator{}
}

func (r *reelsValidator) IsValid(ctx context.Context, eventType enums.EventType, metadata entity.PostMetadata, assignment entity.Assignment) bool {
    if !channel.Instagram.Equals(metadata.Channel) {
        logger.Debug(ctx, "invalid channel", metadata, assignment)
        return false
    }
    if !mediatype.Reels.Equals(metadata.MediaType) {
        logger.Debug(ctx, "invalid media type", metadata, assignment)
        return false
    }
    if eventtype.Recognized != eventType {
        logger.Debug(ctx, "invalid event type", metadata, assignment)
        return false
    }

    return true
}
```

Key points:
- Stateless service — no dependencies injected, no fields.
- Still implements an adapter interface.
- Uses enum constants for comparisons, not raw strings.
- Logs each rejection reason for debuggability.

---

### Example 3 — Validator using entity business logic

```go
package service

import (
    "context"

    "myproject/internal/application/adapter"
    "myproject/internal/domain/entity"
    "myproject/internal/domain/enums"
    "myproject/internal/domain/enums/channel"
    "myproject/internal/domain/enums/eventtype"
    "myproject/internal/domain/enums/mediatype"
    "myproject/pkg/logger"
)

type storyValidator struct{}

func NewStoryValidator() adapter.PostValidator {
    return &storyValidator{}
}

func (s *storyValidator) IsValid(ctx context.Context, eventType enums.EventType, metadata entity.PostMetadata, assignment entity.Assignment) bool {
    if !channel.Instagram.Equals(metadata.Channel) {
        logger.Debug(ctx, "invalid channel", metadata, assignment)
        return false
    }
    if !mediatype.Story.Equals(metadata.MediaType) {
        logger.Debug(ctx, "invalid media type", metadata, assignment)
        return false
    }
    if eventtype.NotRecognized != eventType {
        logger.Debug(ctx, "invalid event type", metadata, assignment)
        return false
    }

    if !assignment.IsCampaignActive() {
        logger.Debug(ctx, "campaign is not active", metadata, assignment)
        return false
    }

    if !assignment.IsParticipating() {
        logger.Debug(ctx, "user is not participating", metadata, assignment)
        return false
    }

    if !assignment.HasPendingTasksOfIdTypeAndChannel(metadata.TaskId, mediatype.Story, channel.Instagram) {
        logger.Debug(ctx, "user has no pending tasks", metadata, assignment)
        return false
    }

    return true
}
```

Key points:
- All business decisions delegate to entity methods: `assignment.IsCampaignActive()`, `assignment.IsParticipating()`, `assignment.HasPendingTasksOfIdTypeAndChannel(...)`.
- The service does NOT inspect entity fields directly — it calls entity methods.
- Each check is logged for observability.

---

## Anti-Patterns

### Anti-Pattern 1 — Business logic in the service (FORBIDDEN)

```go
// BAD — inspecting entity internals instead of calling entity methods
func (s *processOrder) Process(ctx context.Context, ...) error {
    if assignment.Credential == nil || assignment.Credential.Token == nil {  // WRONG
        return nil
    }
    if assignment.Status != "active" {  // WRONG
        return nil
    }
    for _, task := range assignment.Tasks {  // WRONG — iterating entity collections
        if task.PostedAt == nil && task.TaskId == metadata.TaskId {
            // ...
        }
    }
}

// GOOD — delegate to entity methods
func (s *processOrder) Process(ctx context.Context, ...) error {
    if !assignment.IsAccountConnected() { return nil }
    if !assignment.IsActive() { return nil }
    if !assignment.HasPendingTasksOfIdTypeAndChannel(metadata.TaskId, mediaType, ch) { return nil }
}
```

---

### Anti-Pattern 2 — Service calling another service (FORBIDDEN)

```go
// BAD
type processOrder struct {
    validateService adapter.ValidateOrderService  // WRONG — service depending on service
}
```

Extract shared logic into a use case or let the entrypoint orchestrate.

---

### Anti-Pattern 3 — Wrapping errors (FORBIDDEN)

```go
// BAD — wrapping an error from a use case call
result, err := s.findOrder.FindById(ctx, orderId)
if err != nil {
    return fmt.Errorf("failed to find order: %w", err)  // WRONG
}

// BAD — replacing an error
if err != nil {
    return errors.New("order lookup failed")  // WRONG — original error lost
}

// GOOD — pass through
if err != nil {
    return err
}
```

---

### Anti-Pattern 4 — Service used directly without adapter interface (FORBIDDEN)

```go
// BAD — injecting the concrete struct
type handler struct {
    service *service.processOrder  // WRONG — must use adapter interface
}

// GOOD — injecting through adapter interface
type handler struct {
    service adapter.ProcessOrderService
}
```

---

### Anti-Pattern 5 — Missing Started/Finished logs (FORBIDDEN)

```go
// BAD — no observability
func (s *processOrder) Process(ctx context.Context, source string, metadata entity.OrderMetadata) error {
    // jumps straight into logic...
    order, err := s.findOrder.FindById(ctx, metadata.OrderId)
    // ...
}

// GOOD
func (s *processOrder) Process(ctx context.Context, source string, metadata entity.OrderMetadata) error {
    logger.Info(ctx, "Started", metadata, map[string]any{"source": source})
    // ... logic ...
    logger.Info(ctx, "Finished", metadata, map[string]any{"source": source})
    return nil
}
```

---

### Anti-Pattern 6 — Service too large (CODE SMELL)

If a service exceeds ~200-250 lines, investigate:
1. Does it contain business logic that should be on entities?
2. Can some orchestration steps be extracted into focused use cases?
3. Is it doing two processes instead of one?

---

## Service vs Use Case Decision

```
Does the logic need to orchestrate multiple use case calls?
├── YES → Service
│         • Orchestrates use cases
│         • Delegates business checks to entities
│         • Has Started/Finished logging
│         • Implements adapter interface
└── NO  → Use Case
          ├── Needs I/O? → Interface use case
          └── Pure logic? → Plain function use case
```

---

## Quick Reference Checklist

When creating or reviewing a service, verify every item:

- [ ] File is in `internal/application/service/` and declares `package service`
- [ ] Implements an adapter interface from `internal/application/adapter/`
- [ ] Adapter interface name follows `adapter.{ProcessName}Service`
- [ ] Struct is unexported (`processName`, not `ProcessName`)
- [ ] Constructor is `New{ProcessName}Service(...)` returning `adapter.{ProcessName}Service`
- [ ] Method signatures use only domain types (entities, enums, primitives)
- [ ] Begins with `logger.Info(ctx, "Started", ...)` including all raw inputs
- [ ] Ends with `logger.Info(ctx, "Finished", ...)` including all raw inputs
- [ ] All business checks delegate to entity methods — no field access with comparisons
- [ ] Errors from use case calls pass through untouched (no wrapping, no replacing)
- [ ] Does NOT call other services
- [ ] Does NOT own business logic — only orchestrates
- [ ] Under ~200-250 lines
- [ ] Single process responsibility
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
