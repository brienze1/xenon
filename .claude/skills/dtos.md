---
name: DTOs (Clean Architecture)
description: This skill should be used when the user asks to "create a DTO", "add a DTO", "create a contract", "create a request/response struct", "create an event payload", "add json tags", "create a data transfer object", or mentions anything about DTOs in any integration layer dto/ package.
version: 1.0.0
---

# DTOs — Golang Clean Architecture

DTOs (Data Transfer Objects) are the integration layer's contract structs. They carry serialization tags, validation tags, and conversion logic between external contracts (JSON, BSON, Protobuf, etc.) and domain entities. DTOs live exclusively in the integration layer.

---

## Core Rules

### Rule 1: Location — Inside the Integration Scope

DTOs MUST live in a `dto/` folder inside the specific integration scope that uses them.

```
internal/
└── integration/
    ├── entrypoint/
    │   └── dto/                    ← Request/event input contracts
    │       ├── event.go
    │       └── user.go
    ├── webservice/
    │   └── dto/                    ← External API response contracts
    │       └── media_info.go
    ├── publisher/
    │   └── dto/                    ← Message broker event contracts
    │       └── video_processing_event.go
    └── persistence/
        └── model/                  ← Database models (see Models skill)
            └── user.go
```

Each `dto/` folder declares `package dto`. When multiple scopes have DTOs, they are separate packages — they do NOT share types across scopes.

**Note:** Database objects use **models**, not DTOs. See the Models skill for database-specific conventions.

**One struct per file — always.** Each DTO gets its own file. If a DTO is composed of sub-structs, each sub-struct gets its own file too. Never put multiple struct definitions in a single file.

```
// GOOD — one struct per file
dto/
├── event.go                        ← type Event struct
├── post_metadata.go                ← type PostMetadata struct
├── workout_detail_response.go      ← type WorkoutDetailResponse struct
├── workout_detail_workout.go       ← type WorkoutDetailWorkout struct
└── workout_detail_step.go          ← type WorkoutDetailStep struct

// BAD — multiple structs in one file (FORBIDDEN)
dto/
└── workout_detail_response.go      ← type WorkoutDetailResponse + WorkoutDetailWorkout + WorkoutDetailStep  ← WRONG
```

---

### Rule 2: Contract Tags (Always Present)

DTOs almost always carry serialization tags for their contract format. The tag type depends on the integration:

| Integration scope | Common tags | Example |
|---|---|---|
| Entrypoint (REST/events) | `json:`, `validate:` | `json:"user_id" validate:"required,gt=0"` |
| Webservice (external APIs) | `json:` | `json:"media_url"` |
| Publisher (message broker) | `json:` | `json:"campaign_id"` |
| gRPC / Protobuf | `protobuf:` | `protobuf:"bytes,1,opt,name=user_id"` |
| XML contracts | `xml:` | `xml:"user_id"` |

**Note:** Database tags (`bson:`, `gorm:`, `dynamodbav:`) belong on **models**, not DTOs. See the Models skill.

**Tag naming convention:** For new implementations without an existing pattern, use `snake_case` for all contract tags.

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// GOOD — snake_case tags
type User struct {
    UserId    int64  `json:"user_id"`
    FirstName string `json:"first_name"`
    Email     string `json:"email"`
    CreatedAt string `json:"created_at"`
}

// BAD — camelCase tags (unless the existing contract requires it)
type User struct {
    UserId    int64  `json:"userId"`      // WRONG for new implementations
    FirstName string `json:"firstName"`   // WRONG for new implementations
}
```

If the existing contract already uses a different convention (e.g., an external API returns camelCase), match the existing contract.

---

### Rule 3: Naming — No "DTO" in the Name (NEVER)

DTO names MUST NOT contain "DTO" anywhere. The package name `dto` already communicates the purpose.

```go
// BAD
type UserDTO struct { ... }
type OrderResponseDTO struct { ... }

// GOOD
type User struct { ... }
type OrderResponse struct { ... }

// Usage reads naturally:
// dto.User, dto.OrderResponse
```

---

### Rule 4: Constructors — Entity-to-DTO and DTO-to-Entity

DTOs MUST have conversion methods to/from domain entities. This keeps the domain layer ignorant of DTOs.

**DTO-to-Entity** — method on the DTO:
```go
func (u *User) ToEntity() entity.User {
    return entity.User{
        Id:    u.UserId,
        Name:  u.FirstName + " " + u.LastName,
        Email: u.Email,
    }
}
```

**Entity-to-DTO** — constructor function:
```go
func NewUser(user entity.User) User {
    return User{
        UserId:    user.Id,
        FirstName: user.Name,
        Email:     user.Email,
    }
}
```

The direction of conversion determines which pattern to use:
- **Inbound** (request/event → entity): DTO has a `ToEntity()` method
- **Outbound** (entity → response/event): DTO has a `New{Name}(...)` constructor

Some DTOs need both (e.g., persistence models that read and write).

**CRITICAL: DTOs MUST ALWAYS be created via constructors. NEVER create DTO structs inline in the caller.** All field mapping, date formatting, enum conversion, and any other transformation logic belongs INSIDE the DTO constructor — not in the publisher, webservice, controller, or any other file that uses the DTO.

```go
// BAD — DTO created inline in the caller (FORBIDDEN)
event := dto.VideoProcessingEvent{
    OrderId:   metadata.OrderId,
    PostedAt:  metadata.PostedAt.Format("2006-01-02T15:04:05Z"),  // formatting scattered in caller
    EventType: strings.ToUpper(eventType.String()),                // conversion scattered in caller
    MediaType: strings.ToUpper(metadata.MediaType),                // conversion scattered in caller
}

// GOOD — DTO created via constructor (all logic encapsulated)
event := dto.NewVideoProcessingEvent(savedUrl, approvedUrl, source, metadata, eventType)
```

This rule applies to ALL DTOs in ALL scopes (entrypoint, webservice, publisher). No exceptions.

---

### Rule 5: Domain Knows Nothing About DTOs

The domain layer MUST NEVER import DTOs. Conversion always happens in the integration layer.

```go
// GOOD — conversion in the integration layer
func (h *handler) Handle(ctx context.Context, event dto.Event) error {
    entity := event.EventData.ToEntity()  // DTO → entity at the boundary
    return h.service.Process(ctx, entity)
}

// BAD — entity importing DTO
// File: internal/domain/entity/user.go
import "myproject/internal/integration/entrypoint/dto"  // WRONG — domain can't know DTOs
```

---

### Rule 6: Date/Time Formatting

When DTOs serialize dates as strings, they MUST follow these formats:

| Type | Format (Go layout) | Example |
|---|---|---|
| DateTime (UTC) | `2006-01-02T15:04:05.000Z` | `2024-03-15T14:30:45.123Z` |
| Date only | `2006-01-02` | `2024-03-15` |

DTOs handle the conversion between `time.Time` (domain) and formatted strings (contract).

```go
// DateTime formatting in a DTO constructor
func NewOrder(order entity.Order) Order {
    return Order{
        OrderId:   order.Id,
        CreatedAt: order.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
    }
}

// DateTime parsing in ToEntity
func (o *Order) ToEntity() entity.Order {
    createdAt, _ := time.Parse("2006-01-02T15:04:05.000Z", o.CreatedAt)
    return entity.Order{
        Id:        o.OrderId,
        CreatedAt: createdAt,
    }
}
```

---

### Rule 7: Validation Tags (go-playground/validator)

Entrypoint DTOs can use `validate:` tags from the `go-playground/validator` package for input validation. The controller validates the DTO before processing.

```go
type CreateOrder struct {
    ProductId  int64  `json:"product_id" validate:"required,gt=0"`
    Quantity   int    `json:"quantity" validate:"required,gt=0,lte=100"`
    Channel    string `json:"channel" validate:"required,oneof_ignore_case=instagram youtube tiktok"`
    ScheduleAt string `json:"schedule_at" validate:"required,datetime_format=2006-01-02T15:04:05.000Z"`
    TimeSlot   string `json:"time_slot" validate:"required,hour_minute"`
}
```

**Custom validations** can be created in `internal/integration/entrypoint/validator/` and registered on the custom validator instance:

```
internal/
└── integration/
    └── entrypoint/
        └── validator/
            └── custom.go           ← Custom validation rules registration
```

```go
// File: internal/integration/entrypoint/validator/custom.go
package validator

import "github.com/go-playground/validator/v10"

type Validate interface {
    Struct(code string, s interface{}) error
}

type customValidator struct {
    validator *validator.Validate
}

func CustomValidator() Validate {
    validate := validator.New()
    _ = validate.RegisterValidation("not_nil", notNil)
    _ = validate.RegisterValidation("oneof_ignore_case", oneOfIgnoreCase)
    _ = validate.RegisterValidation("datetime_format", dateTimeFormat)
    _ = validate.RegisterValidation("hour_minute", hourMinute)
    // Register json tag names for error messages
    validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
        name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
        if name == "-" {
            return ""
        }
        return name
    })
    return &customValidator{validator: validate}
}
```

Custom validation functions are just `func(validator.FieldLevel) bool` — they receive the field and return whether it's valid.

---

### Rule 8: JSON Parsing Logic

DTOs can contain custom JSON marshaling/unmarshaling logic when the contract format differs from Go's default behavior. Common cases include date parsing and enum conversion.

```go
func (e Event) GetEventType() enums.EventType {
    return eventtype.GetEventType(e.EventType)
}

func (e Event) GetCorrelationId() string {
    if e.EventCorrelationId == "" {
        e.EventCorrelationId = uuid.NewString()
    }
    return e.EventCorrelationId
}
```

---

## Code Examples

### Example 1 — Entrypoint DTO (inbound request with validation)

```go
// File: internal/integration/entrypoint/dto/create_order.go
package dto

import (
    "time"
    "myproject/internal/domain/entity"
)

type CreateOrder struct {
    ProductId  int64  `json:"product_id" validate:"required,gt=0"`
    Quantity   int    `json:"quantity" validate:"required,gt=0"`
    Channel    string `json:"channel" validate:"required"`
    ScheduleAt string `json:"schedule_at" validate:"required"`
}

func (c *CreateOrder) ToEntity() entity.Order {
    scheduleAt, _ := time.Parse("2006-01-02T15:04:05.000Z", c.ScheduleAt)
    return entity.Order{
        ProductId:  c.ProductId,
        Quantity:   c.Quantity,
        Channel:    c.Channel,
        ScheduleAt: &scheduleAt,
    }
}
```

Key points:
- `json:` tags for the contract, `validate:` tags for input validation.
- `ToEntity()` converts the DTO to a domain entity, parsing the date string.
- No "DTO" in the name.

---

### Example 2 — Entrypoint DTO (outbound response)

```go
// File: internal/integration/entrypoint/dto/order_response.go
package dto

import "myproject/internal/domain/entity"

type OrderResponse struct {
    OrderId   int64  `json:"order_id"`
    Status    string `json:"status"`
    Total     float64 `json:"total"`
    CreatedAt string `json:"created_at"`
}

func NewOrderResponse(order entity.Order) OrderResponse {
    return OrderResponse{
        OrderId:   order.Id,
        Status:    order.Status,
        Total:     order.Total,
        CreatedAt: order.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
    }
}
```

Key points:
- Constructor `NewOrderResponse(entity)` converts entity → DTO.
- Date is formatted to the contract's string format.

---

### Example 3 — Entrypoint DTO (event envelope with helper methods)

```go
// File: internal/integration/entrypoint/dto/event.go
package dto

import (
    "time"
    "myproject/internal/domain/enums"
    "myproject/internal/domain/enums/eventtype"
    "github.com/google/uuid"
)

type Event struct {
    EventDate          time.Time    `json:"event_date"`
    EventId            string       `json:"event_id"`
    EventCorrelationId string       `json:"event_correlation_id"`
    BusinessUniqueId   string       `json:"business_unique_id"`
    Source             string       `json:"source"`
    EventType          string       `json:"event_type"`
    EventData          OrderPayload `json:"event_data" validate:"required"`
    SpecVersion        int          `json:"spec_version"`
}

func (e Event) GetCorrelationId() string {
    if e.EventCorrelationId == "" {
        e.EventCorrelationId = uuid.NewString()
    }
    return e.EventCorrelationId
}

func (e Event) GetEventType() enums.EventType {
    return eventtype.GetEventType(e.EventType)
}
```

Key points:
- Nested DTO (`OrderPayload`) for the event data.
- Helper methods for enum conversion and correlation ID generation.
- `validate:"required"` on nested DTO ensures the payload is present.

---

### Example 4 — Entrypoint DTO (nested payload with ToEntity)

```go
// File: internal/integration/entrypoint/dto/order_payload.go
package dto

import (
    "time"
    "myproject/internal/domain/entity"
)

type OrderPayload struct {
    PostedAt   time.Time `json:"posted_at" validate:"required"`
    OrderId    string    `json:"order_id" validate:"required,gt=0"`
    Channel    string    `json:"channel" validate:"required"`
    MediaType  string    `json:"media_type" validate:"required"`
    ProjectId  int64     `json:"project_id" validate:"required,gt=0"`
    UserId     int64     `json:"user_id" validate:"required,gt=0"`
    TaskId     int64     `json:"task_id" validate:"required,gt=0"`
}

func (o *OrderPayload) ToEntity() entity.OrderMetadata {
    return entity.OrderMetadata{
        ProjectId: o.ProjectId,
        UserId:    o.UserId,
        TaskId:    o.TaskId,
        OrderId:   o.OrderId,
        Channel:   o.Channel,
        MediaType: o.MediaType,
        PostedAt:  o.PostedAt,
    }
}
```

---

### Example 5 — Webservice DTO (external API response)

```go
// File: internal/integration/webservice/dto/media_info.go
package dto

import "myproject/internal/domain/entity"

type MediaInfo struct {
    ID        string `json:"id"`
    MediaURL  string `json:"media_url"`
    MediaType string `json:"media_type"`
}

func (m *MediaInfo) ToEntity() *entity.MediaInfo {
    return &entity.MediaInfo{
        ID:        m.ID,
        MediaType: m.MediaType,
        MediaURL:  m.MediaURL,
    }
}
```

Key points:
- `json:` tags match the external API's response format.
- `ToEntity()` converts the external response to a domain entity.
- Returns a pointer (nullable — the response might not have this data).

---

### Example 6 — Publisher DTO (outbound event)

```go
// File: internal/integration/publisher/dto/order_processing_event.go
package dto

import (
    "strings"
    "myproject/internal/domain/entity"
    "myproject/internal/domain/enums"
)

type OrderProcessingEvent struct {
    OrderId           string `json:"order_id"`
    PostedAt          string `json:"posted_at"`
    ApprovedMediaUrl  string `json:"approved_media_url"`
    DownloadedUrl     string `json:"downloaded_url"`
    OriginalUrl       string `json:"original_url"`
    Channel           string `json:"channel"`
    Source            string `json:"source"`
    EventType         string `json:"event_type"`
    MediaType         string `json:"media_type"`
    ProjectTitle      string `json:"project_title"`
    ProjectId         int64  `json:"project_id"`
    UserId            int64  `json:"user_id"`
    TaskId            int64  `json:"task_id"`
}

func NewOrderProcessingEvent(
    savedUrl, approvedUrl, source string,
    metadata entity.OrderMetadata,
    eventType enums.EventType,
) OrderProcessingEvent {
    return OrderProcessingEvent{
        ProjectId:        metadata.ProjectId,
        ProjectTitle:     metadata.ProjectTitle,
        UserId:           metadata.UserId,
        TaskId:           metadata.TaskId,
        OrderId:          metadata.OrderId,
        PostedAt:         metadata.PostedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
        ApprovedMediaUrl: approvedUrl,
        DownloadedUrl:    savedUrl,
        OriginalUrl:      metadata.PostUrl,
        Channel:          metadata.Channel,
        Source:           source,
        EventType:        strings.ToUpper(eventType.String()),
        MediaType:        strings.ToUpper(metadata.MediaType),
    }
}
```

Key points:
- Constructor receives domain entity + primitives and builds the DTO.
- Date formatted to UTC string for the contract.
- Enum converted to uppercase string for the external contract.

---

### Example 7 — Entrypoint DTO with custom validation tags

```go
// File: internal/integration/entrypoint/dto/schedule_task.go
package dto

import (
    "time"
    "myproject/internal/domain/entity"
)

type ScheduleTask struct {
    TaskId     int64  `json:"task_id" validate:"required,gt=0"`
    Channel    string `json:"channel" validate:"required,oneof_ignore_case=instagram youtube tiktok"`
    ScheduleAt string `json:"schedule_at" validate:"required,datetime_format=2006-01-02T15:04:05.000Z"`
    TimeSlot   string `json:"time_slot" validate:"required,hour_minute"`
    Priority   int    `json:"priority" validate:"required,gte=1,lte=10"`
}

func (s *ScheduleTask) ToEntity() entity.Task {
    scheduleAt, _ := time.Parse("2006-01-02T15:04:05.000Z", s.ScheduleAt)
    return entity.Task{
        TaskId:     s.TaskId,
        Channel:    s.Channel,
        ScheduleAt: &scheduleAt,
        Priority:   s.Priority,
    }
}
```

Key points:
- Uses custom validation tags: `oneof_ignore_case`, `datetime_format`, `hour_minute`.
- Custom validations are registered in `entrypoint/validator/custom.go`.
- Standard validators like `required`, `gt`, `gte`, `lte` from go-playground/validator.

---

## Anti-Patterns

### Anti-Pattern 1 — "DTO" in the name (FORBIDDEN)

```go
// BAD
type UserDTO struct { ... }
type OrderResponseDTO struct { ... }

// GOOD
type User struct { ... }
type OrderResponse struct { ... }
```

---

### Anti-Pattern 2 — DTO in the domain layer (FORBIDDEN)

```go
// BAD — DTO in domain
// File: internal/domain/dto/user.go
type User struct {
    Name string `json:"name"`  // WRONG — domain has no serialization tags
}
```

DTOs live exclusively in `integration/<scope>/dto/`.

---

### Anti-Pattern 7 — Database tags on DTOs (FORBIDDEN)

```go
// BAD — database tags belong on models, not DTOs
// File: internal/integration/persistence/dto/user.go
type User struct {
    ID   int64  `gorm:"primaryKey" bson:"_id"`  // WRONG — use models for database
    Name string `gorm:"column:name"`
}
```

Database objects use **models** (see Models skill), not DTOs.

---

### Anti-Pattern 3 — Domain entity importing DTO (FORBIDDEN)

```go
// BAD — entity knows about DTOs
// File: internal/domain/entity/user.go
import "myproject/internal/integration/entrypoint/dto"

func (u *User) ToDTO() dto.User { ... }  // WRONG
```

Conversion lives on the DTO, not the entity. The domain never imports DTOs.

---

### Anti-Pattern 4 — Missing conversion methods (FORBIDDEN)

```go
// BAD — no ToEntity or constructor, forcing manual mapping in services
type User struct {
    UserId int64 `json:"user_id"`
    Name   string `json:"name"`
}

// Then in service/controller:
entity := entity.User{Id: userDto.UserId, Name: userDto.Name}  // WRONG — scattered mapping
```

Always provide `ToEntity()` and/or `New{Name}(entity)` methods.

---

### Anti-Pattern 5 — Wrong date format (FORBIDDEN)

```go
// BAD — inconsistent date formats
PostedAt: order.PostedAt.Format("2006-01-02 15:04:05")       // WRONG — missing T and Z
PostedAt: order.PostedAt.Format(time.RFC3339)                 // WRONG — not the project standard
PostedAt: order.PostedAt.Format("01/02/2006")                 // WRONG — not ISO format

// GOOD
PostedAt: order.PostedAt.UTC().Format("2006-01-02T15:04:05.000Z")  // DateTime
DueDate:  order.DueDate.Format("2006-01-02")                       // Date only
```

---

### Anti-Pattern 6 — Sharing DTOs across scopes (FORBIDDEN)

```go
// BAD — entrypoint importing publisher DTO
// File: internal/integration/entrypoint/controller/handler.go
import "myproject/internal/integration/publisher/dto"  // WRONG — cross-scope import

// Each scope has its own DTO package, even if the shapes are similar
```

---

### Anti-Pattern 8 — Inline DTO creation in caller (FORBIDDEN)

```go
// BAD — DTO struct built inline in publisher/controller/webservice
event := dto.VideoProcessingEvent{
    OrderId:   metadata.OrderId,
    PostedAt:  metadata.PostedAt.Format("2006-01-02T15:04:05Z"),
    EventType: strings.ToUpper(eventType.String()),
    MediaType: strings.ToUpper(metadata.MediaType),
}

// GOOD — DTO created via constructor
event := dto.NewVideoProcessingEvent(savedUrl, approvedUrl, source, metadata, eventType)
```

All field mapping, formatting, and conversion logic belongs inside the DTO constructor. The caller passes domain types in, the constructor handles everything else. No exceptions.

---

## Quick Reference Checklist

When creating or reviewing a DTO, verify every item:

- [ ] Lives in `internal/integration/<scope>/dto/` and declares `package dto`
- [ ] **One struct per file** — no multiple struct definitions in the same file
- [ ] Name does NOT contain "DTO"
- [ ] Has appropriate contract tags (`json:`, `xml:`, `protobuf:`, etc.) — NOT database tags
- [ ] **DTOs are ALWAYS created via constructors** — never inline struct creation in callers
- [ ] Entrypoint DTOs have `validate:` tags from go-playground/validator for input validation
- [ ] Custom validations live in `entrypoint/validator/` and are registered on the custom validator
- [ ] Tags use `snake_case` for new implementations
- [ ] Has `ToEntity()` method for inbound conversion (DTO → entity)
- [ ] Has `New{Name}(entity)` constructor for outbound conversion (entity → DTO)
- [ ] Domain layer does NOT import the DTO — conversion is on the DTO side
- [ ] DateTime strings use `2006-01-02T15:04:05.000Z` format (UTC)
- [ ] Date-only strings use `2006-01-02` format
- [ ] Not shared across integration scopes
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
