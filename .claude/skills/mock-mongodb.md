---
name: Mock - MongoDB
description: Use when creating or referencing the MongoDB mock for BDD tests.
version: 1.0.0
---

# Mock: MongoDB

In-memory MongoDB mock with BSON-compatible operations, aggregation pipeline support, and dot-notation field resolution.

**Location:** `test/integration/mock/mongo.go`

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package mock

import (
    "context"
    "sync"
    "go.mongodb.org/mongo-driver/bson"
)

type MongoCollection struct {
    mu   sync.RWMutex
    data []bson.M
}

type MongoClient struct {
    mu          sync.RWMutex
    collections map[string]*MongoCollection
}

func NewMongoClient() *MongoClient {
    return &MongoClient{collections: make(map[string]*MongoCollection)}
}

func (m *MongoClient) GetCollection(name string) *MongoCollection {
    m.mu.Lock()
    defer m.mu.Unlock()
    if m.collections[name] == nil {
        m.collections[name] = &MongoCollection{data: make([]bson.M, 0)}
    }
    return m.collections[name]
}

// CRUD operations
func (c *MongoCollection) InsertOne(ctx context.Context, document bson.M) error { ... }
func (c *MongoCollection) InsertMany(ctx context.Context, documents []bson.M) error { ... }
func (c *MongoCollection) FindOneByFilter(ctx context.Context, filter bson.M, dest any) error { ... }
func (c *MongoCollection) FindDataByFilter(ctx context.Context, filter bson.M, dest any) error { ... }
func (c *MongoCollection) UpdateOne(ctx context.Context, filter, update bson.M, upsert bool) error { ... }
func (c *MongoCollection) UpdateMany(ctx context.Context, filter, update bson.M) error { ... }
func (c *MongoCollection) RemoveDataByFilter(ctx context.Context, filter bson.M, multi bool) error { ... }

// Query operators supported: $in, $regex, $gt, $gte, $lt, $lte, $ne, $exists
// Aggregation pipeline: $match, $group, $sum, $project

// Inspection methods
func (c *MongoCollection) GetAllData() []bson.M { ... }
func (c *MongoCollection) GetCount() int { ... }
func (c *MongoCollection) HasDocumentMatchingField(field string, value any) bool { ... }

func (m *MongoClient) Reset() {
    m.mu.Lock()
    defer m.mu.Unlock()
    for _, col := range m.collections {
        col.data = make([]bson.M, 0)
    }
}
```

**Lifecycle:** The MongoDB mock is created once and injected once in `InitializeScenario`. The `Reset()` method clears all documents from all collections between BDD scenarios (preserving collection definitions). The mock instance is never recreated or re-injected — only its internal state is cleared.

Key features:
- **Dot-notation** for nested fields: `"address.city"` resolves to `doc["address"]["city"]`
- **Type coercion** for numeric comparisons (string ↔ int ↔ float)
- **Aggregation pipeline** with $match, $group, $sum stages
- **Upsert support** on UpdateOne

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
