---
name: Package - collection (Generic Collections)
description: This skill should be used when the user asks to "create a queue", "add a stack", "create a generic collection", or mentions anything about the pkg/collection/ package.
version: 1.0.0
---

# Package: collection — Generic Collections

The `collection` package provides generic data structure implementations using Go generics.

**Location:** `pkg/collection/`

---

## Queue[T]

A generic FIFO queue with Push and Pop operations:

```go
package collection

type Queue[T any] interface {
    Push(item T)
    Pop() (T, bool)
}

type queue[T any] struct {
    items []T
}

func NewQueue[T any]() Queue[T] {
    return &queue[T]{items: make([]T, 0)}
}

func (q *queue[T]) Push(item T) {
    q.items = append(q.items, item)
}

func (q *queue[T]) Pop() (T, bool) {
    if len(q.items) == 0 {
        var zero T
        return zero, false
    }
    item := q.items[0]
    q.items = q.items[1:]
    return item, true
}
```

---

## Key Rules

- **No internal imports** — standalone package
- **Uses Go generics** — `[T any]` for type safety
- **Exported interface, unexported struct**
- **Copy-ready** — no dependencies beyond stdlib

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
