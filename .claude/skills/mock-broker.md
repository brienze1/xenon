---
name: Mock - Broker (Message Broker)
description: Use when creating or referencing the message broker mock (RabbitMQ, generic) for BDD tests.
version: 1.0.0
---

# Mock: Broker (Message Broker)

Generic message broker mock that tracks published messages. Works with RabbitMQ, SQS-based brokers, or any publish/subscribe pattern.

**Location:** `test/integration/mock/broker.go`

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package mock

import (
    "context"
    "encoding/json"
    "sync"
)

type BrokerClient[T any] struct {
    mu       sync.RWMutex
    messages []T
    Builder  func() BrokerProducer[T]  // Factory for DI injection
}

type BrokerProducer[T any] interface {
    Publish(ctx context.Context, message T) error
}

func NewBrokerClient[T any]() *BrokerClient[T] {
    b := &BrokerClient[T]{
        messages: make([]T, 0),
    }
    b.Builder = func() BrokerProducer[T] { return b }
    return b
}

func (b *BrokerClient[T]) Publish(ctx context.Context, message T) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.messages = append(b.messages, message)
    return nil
}

// Inspection methods
func (b *BrokerClient[T]) GetMessages() []T {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return b.messages
}

func (b *BrokerClient[T]) GetMessageCount() int {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return len(b.messages)
}

func (b *BrokerClient[T]) GetMessage(index int) *T {
    b.mu.RLock()
    defer b.mu.RUnlock()
    if index < 0 || index >= len(b.messages) { return nil }
    return &b.messages[index]
}

func (b *BrokerClient[T]) GetMessageAsJSON(index int) map[string]any {
    msg := b.GetMessage(index)
    if msg == nil { return nil }
    data, _ := json.Marshal(msg)
    var result map[string]any
    json.Unmarshal(data, &result)
    return result
}

func (b *BrokerClient[T]) Reset() {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.messages = make([]T, 0)
}
```

**Usage:**
```go
test.broker = mock.NewBrokerClient[any]()
broker.SetBrokerBuilder(test.broker.Builder)  // ONE-TIME injection in InitializeScenario
```

**Lifecycle:** The broker mock is created once and injected once via `broker.SetBrokerBuilder()` in `InitializeScenario`. The `Reset()` method clears all accumulated messages between BDD scenarios. The mock instance is never recreated or re-injected — only its internal state is cleared.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
