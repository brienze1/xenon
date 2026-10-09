---
name: Publishers (Clean Architecture)
description: This skill should be used when the user asks to "create a publisher", "add a message publisher", "publish to a queue", "publish to a topic", "send a message to SQS", "send a message to RabbitMQ", "create an event publisher", or mentions anything about publishers in the internal/integration/publisher/ package.
version: 1.0.0
---

# Publishers — Golang Clean Architecture

Publishers send messages to queues, topics, and message brokers (SQS, SNS, RabbitMQ, Kafka, etc.). They are integration layer implementations that convert domain entities to DTOs and publish them through an injected broker interface. They are simple, one-way operations.

---

## Core Rules

### Rule 1: Location

Publishers MUST live in `internal/integration/publisher/` and declare `package publisher`.

```
internal/
└── integration/
    └── publisher/
        ├── dto/                          ← Message payload DTOs
        │   └── order_processing_event.go
        └── order_processing.go           ← Publisher implementation
```

---

### Rule 2: Implements Integration Adapters (Which Aggregate Use Cases)

Same pattern as persistence and webservices — publishers implement integration adapters that embed use case interfaces.

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// Adapter: internal/integration/adapter/order_processing_publisher.go
package adapter

import "myproject/internal/application/usecase"

type OrderProcessingPublisher interface {
    usecase.ProcessOrder
}
```

```go
// Implementation: internal/integration/publisher/order_processing.go
type orderProcessingPublisher struct {
    messagePublisher broker.MessageProducer[any]
}
```

---

### Rule 3: Naming Conventions

| Element | Pattern | Example |
|---|---|---|
| Struct | Unexported, `{name}Publisher` | `orderProcessingPublisher`, `emailNotificationPublisher` |
| Constructor | `New{Name}(...)` | `NewOrderProcessing(...)` |
| Constructor return type | Integration adapter | `adapter.OrderProcessingPublisher` |
| Package | `publisher` | `package publisher` |

The struct is **always unexported** — exposed only via the constructor returning the adapter interface.

```go
func NewOrderProcessing(messagePublisher broker.MessageProducer[any]) adapter.OrderProcessingPublisher {
    return &orderProcessingPublisher{
        messagePublisher: messagePublisher,
    }
}
```

---

### Rule 4: Broker Injected via Interface

The broker (SQS, SNS, RabbitMQ, Kafka, etc.) MUST be injected via an interface adapter in the constructor. The broker implementation lives in the `infra` layer or `pkg`, fully decoupled from the publisher.

```go
// GOOD — broker injected as interface
type emailPublisher struct {
    messagePublisher broker.MessageProducer[any]  // Interface from pkg or infra
}

func NewEmailNotification(messagePublisher broker.MessageProducer[any]) adapter.EmailNotificationPublisher {
    return &emailPublisher{messagePublisher: messagePublisher}
}
```

```go
// BAD — creating broker inside publisher
func NewEmailNotification() adapter.EmailNotificationPublisher {
    conn, _ := amqp.Dial("amqp://localhost")  // WRONG — broker setup belongs in infra
    return &emailPublisher{conn: conn}
}
```

---

### Rule 5: Logger — Debug Level (Started/Finished)

Same pattern as persistence and webservices — every method logs `Started` and `Finished` at **Debug** level with all raw inputs/outputs.

```go
func (p *orderProcessingPublisher) Execute(ctx context.Context, ...) error {
    logger.Debug(ctx, "Started", metadata, map[string]any{"source": source, "eventType": eventType})

    // ... build DTO, publish ...

    logger.Debug(ctx, "Finished", metadata, map[string]any{"source": source, "eventType": eventType})
    return nil
}
```

---

### Rule 6: DTOs Build the Message Payload (Always Via Constructor)

Publisher DTOs live in `internal/integration/publisher/dto/` and represent the message payload contract. The DTO MUST always be created via its constructor — never inline struct creation. All field mapping, formatting, and conversion logic belongs inside the DTO constructor.

```go
// GOOD — DTO created via constructor (all conversion logic encapsulated in the DTO)
event := dto.NewOrderProcessingEvent(savedUrl, approvedUrl, source, metadata, eventType)

// BAD — DTO created inline (FORBIDDEN — formatting/conversion scattered in publisher)
event := dto.OrderProcessingEvent{
    OrderId:   metadata.OrderId,
    PostedAt:  metadata.PostedAt.Format("2006-01-02T15:04:05Z"),  // WRONG — belongs in DTO constructor
    EventType: strings.ToUpper(eventType.String()),                // WRONG — belongs in DTO constructor
}

// Wrap in message envelope and publish
message := domain.QueueMessage[any]{
    Type:    "order-processing",
    Payload: event,
}
err := p.messagePublisher.Publish(ctx, message)
```

---

### Rule 7: One-Way — Return Only Error

Publishers are one-way operations. They almost always return only `error` (nil on success). The use case interface dictates the return type, and since publishing is fire-and-forget, a single `error` return is the standard.

```go
// Typical use case for a publisher
type ProcessOrder interface {
    Execute(ctx context.Context, savedUrl, approvedUrl, source string, metadata entity.OrderMetadata, eventType enums.EventType) error
}
```

---

## Code Examples

### Example 1 — Full publisher with DTO construction

```go
package publisher

import (
    "context"

    domain "myproject/pkg/domain/value-objects"
    broker "myproject/pkg/contracts/message-broker"

    "myproject/internal/domain/entity"
    "myproject/internal/domain/enums"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/publisher/dto"
    "myproject/pkg/logger"
)

type orderProcessingPublisher struct {
    messagePublisher broker.MessageProducer[any]
}

func NewOrderProcessing(messagePublisher broker.MessageProducer[any]) adapter.OrderProcessingPublisher {
    return &orderProcessingPublisher{
        messagePublisher: messagePublisher,
    }
}

func (p *orderProcessingPublisher) Execute(ctx context.Context, savedUrl, approvedUrl, source string, metadata entity.OrderMetadata, eventType enums.EventType) error {
    logger.Debug(ctx, "Started", metadata, map[string]any{
        "savedUrl":    savedUrl,
        "approvedUrl": approvedUrl,
        "source":      source,
        "eventType":   eventType,
    })

    message := domain.QueueMessage[any]{
        Type:    "order-processing",
        Payload: dto.NewOrderProcessingEvent(savedUrl, approvedUrl, source, metadata, eventType),
    }

    err := p.messagePublisher.Publish(ctx, message)
    if err != nil {
        return err
    }

    logger.Debug(ctx, "Finished", metadata, map[string]any{
        "savedUrl":    savedUrl,
        "approvedUrl": approvedUrl,
        "source":      source,
        "eventType":   eventType,
    })
    return nil
}
```

---

### Example 2 — Simple notification publisher

```go
package publisher

import (
    "context"

    broker "myproject/pkg/contracts/message-broker"

    "myproject/internal/integration/adapter"
    "myproject/internal/integration/publisher/dto"
    "myproject/pkg/logger"
)

type emailNotificationPublisher struct {
    messagePublisher broker.MessageProducer[any]
}

func NewEmailNotification(messagePublisher broker.MessageProducer[any]) adapter.EmailNotificationPublisher {
    return &emailNotificationPublisher{
        messagePublisher: messagePublisher,
    }
}

func (p *emailNotificationPublisher) Send(ctx context.Context, recipient, subject, body string) error {
    logger.Debug(ctx, "Started", map[string]any{
        "recipient": recipient,
        "subject":   subject,
    })

    message := dto.NewEmailNotification(recipient, subject, body)

    err := p.messagePublisher.Publish(ctx, message)
    if err != nil {
        return err
    }

    logger.Debug(ctx, "Finished", map[string]any{
        "recipient": recipient,
        "subject":   subject,
    })
    return nil
}
```

---

### Example 3 — Publisher DTO

```go
// File: internal/integration/publisher/dto/order_processing_event.go
package dto

import (
    "strings"
    "myproject/internal/domain/entity"
    "myproject/internal/domain/enums"
)

type OrderProcessingEvent struct {
    OrderId         string `json:"order_id"`
    PostedAt        string `json:"posted_at"`
    ApprovedUrl     string `json:"approved_url"`
    DownloadedUrl   string `json:"downloaded_url"`
    OriginalUrl     string `json:"original_url"`
    Channel         string `json:"channel"`
    Source          string `json:"source"`
    EventType       string `json:"event_type"`
    MediaType       string `json:"media_type"`
    ProjectTitle    string `json:"project_title"`
    ProjectId       int64  `json:"project_id"`
    UserId          int64  `json:"user_id"`
    TaskId          int64  `json:"task_id"`
}

func NewOrderProcessingEvent(
    savedUrl, approvedUrl, source string,
    metadata entity.OrderMetadata,
    eventType enums.EventType,
) OrderProcessingEvent {
    return OrderProcessingEvent{
        ProjectId:     metadata.ProjectId,
        ProjectTitle:  metadata.ProjectTitle,
        UserId:        metadata.UserId,
        TaskId:        metadata.TaskId,
        OrderId:       metadata.OrderId,
        PostedAt:      metadata.PostedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
        ApprovedUrl:   approvedUrl,
        DownloadedUrl: savedUrl,
        OriginalUrl:   metadata.PostUrl,
        Channel:       metadata.Channel,
        Source:        source,
        EventType:     strings.ToUpper(eventType.String()),
        MediaType:     strings.ToUpper(metadata.MediaType),
    }
}
```

---

## Anti-Patterns

### Anti-Pattern 1 — Creating broker inside publisher (FORBIDDEN)

```go
// BAD — broker setup belongs in infra layer
func NewOrderProcessing() adapter.OrderProcessingPublisher {
    conn, _ := amqp.Dial("amqp://localhost")  // WRONG
    ch, _ := conn.Channel()
    return &orderProcessingPublisher{channel: ch}
}

// GOOD — broker injected as interface
func NewOrderProcessing(broker broker.MessageProducer[any]) adapter.OrderProcessingPublisher {
    return &orderProcessingPublisher{messagePublisher: broker}
}
```

---

### Anti-Pattern 2 — Exported struct (FORBIDDEN)

```go
// BAD
type OrderProcessingPublisher struct { ... }  // WRONG — must be unexported

// GOOD
type orderProcessingPublisher struct { ... }
```

---

### Anti-Pattern 3 — Building message inline without DTO (FORBIDDEN)

```go
// BAD — building payload inline, no DTO
func (p *publisher) Execute(ctx context.Context, metadata entity.OrderMetadata) error {
    payload := map[string]any{
        "order_id": metadata.OrderId,
        "user_id":  metadata.UserId,
    }
    return p.broker.Publish(ctx, payload)  // WRONG — use a DTO
}

// GOOD — DTO handles the contract
event := dto.NewOrderProcessingEvent(savedUrl, approvedUrl, source, metadata, eventType)
return p.broker.Publish(ctx, event)
```

---

### Anti-Pattern 4 — Business logic in publisher (FORBIDDEN)

```go
// BAD — filtering or deciding in publisher
func (p *publisher) Execute(ctx context.Context, metadata entity.OrderMetadata) error {
    if !metadata.IsValid() {  // WRONG — business logic belongs in service/entity
        return nil
    }
    // ...
}
```

---

## Quick Reference Checklist

When creating or reviewing a publisher, verify every item:

- [ ] Lives in `internal/integration/publisher/` and declares `package publisher`
- [ ] Struct is **unexported** (`orderProcessingPublisher`, not `OrderProcessingPublisher`)
- [ ] Constructor is `New{Name}(...)` returning `adapter.{Name}Publisher`
- [ ] Broker is injected via interface (never created inside)
- [ ] Broker implementation lives in `infra` or `pkg`, decoupled from this code
- [ ] Implements integration adapter that aggregates use case interfaces
- [ ] DTOs live in `internal/integration/publisher/dto/` and handle entity → payload conversion
- [ ] Logs `Started` and `Finished` at **Debug** level with all raw inputs/outputs
- [ ] Returns only `error` (nil on success) — one-way operation
- [ ] Contains no business logic
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
