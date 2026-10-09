---
name: Models (Clean Architecture)
description: This skill should be used when the user asks to "create a model", "add a database model", "create a persistence model", "add a GORM model", "create a MongoDB model", "create a DynamoDB model", "add a table mapping", or mentions anything about models in the internal/integration/persistence/model/ package.
version: 1.0.0
---

# Models — Golang Clean Architecture

Models are the database abstraction layer. They map between domain entities and database representations (SQL, MongoDB, DynamoDB, etc.). Models are to databases what DTOs are to API contracts — they own the tags, conversion logic, and database-specific behavior.

---

## Core Rules

### Rule 1: Location

Models MUST live in `internal/integration/persistence/model/` and declare `package model`.

**One struct per file — always.** Each model gets its own file. If a model has related sub-structs (e.g., embedded types for GORM relationships), each gets its own file.

```
internal/
└── integration/
    └── persistence/
        ├── model/
        │   ├── user.go              ← type User struct
        │   ├── order.go             ← type Order struct
        │   ├── order_item.go        ← type OrderItem struct
        │   ├── credential.go        ← type Credential struct
        │   └── payment.go           ← type Payment struct
        ├── user.go              ← persistence implementation (uses models)
        └── order.go
```

---

### Rule 2: Naming — No "Model" in the Name (NEVER)

Model names MUST NOT contain "Model" anywhere. The package name `model` already communicates the purpose.

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// BAD
type UserModel struct { ... }
type OrderModel struct { ... }

// GOOD
type User struct { ... }
type Order struct { ... }

// Usage reads naturally:
// model.User, model.Order
```

---

### Rule 3: Conversion Methods — Same as DTOs

Models MUST have conversion methods to/from domain entities. The domain never imports models.

**Model-to-Entity** — method on the model:
```go
func (u *User) ToEntity() entity.User {
    return entity.User{
        Id:    u.ID,
        Name:  u.Name,
        Email: u.Email,
    }
}
```

**Entity-to-Model** — constructor function:
```go
func NewUser(user entity.User) User {
    return User{
        ID:    user.Id,
        Name:  user.Name,
        Email: user.Email,
    }
}
```

For read-only models, only `ToEntity()` is needed. For write operations, also provide the `New{Name}(entity)` constructor.

**CRITICAL: Models MUST ALWAYS be created via constructors when converting from entities. NEVER create model structs inline in persistence files.** All field mapping belongs inside the model constructor.

```go
// BAD — model created inline in persistence (FORBIDDEN)
userModel := model.User{ID: user.Id, Name: user.Name, Email: user.Email}

// GOOD — model created via constructor
userModel := model.NewUser(user)
```

---

### Rule 4: Models Can Have Helper Methods

Unlike DTOs, models often need helper methods to navigate relationships, extract derived values, or aggregate data from nested models. These are database-level helpers, NOT business logic.

```go
func (c *Campaign) GetAdIds() []int64 {
    adIds := make([]int64, 0)
    for _, moment := range c.Moments {
        adIds = append(adIds, moment.GetAdIds()...)
    }
    return adIds
}

func (f *Format) GetMediaType() enums.MediaType {
    return mediatype.GetMediaType(f.Name)
}

func (f *Format) GetChannel() enums.Channel {
    return f.FormatSource.GetChannel()
}
```

These helpers bridge the gap between complex database relationships and flat domain entities.

---

## Model Types

### Type 1: SQL Models (GORM)

SQL models use `gorm:` tags for column mapping and relationships. They also carry `json:` tags for BDD testing mappings.

**Key characteristics:**
- `gorm:` tags for column names, primary keys, foreign keys, relationships
- `json:` tags alongside for BDD test support
- `TableName()` method to specify the database table
- Support for `gorm.DeletedAt` soft delete
- GORM hooks (`BeforeCreate`, `AfterFind`, etc.) when needed

```go
package model

import (
    "time"
    "myproject/internal/domain/entity"
    "gorm.io/gorm"
)

type User struct {
    DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at"`
    Name      string         `json:"name" gorm:"column:name"`
    Email     string         `json:"email" gorm:"column:email"`
    Status    string         `json:"status" gorm:"column:status"`
    CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
    ID        int64          `json:"id" gorm:"primaryKey;column:id"`
}

func (User) TableName() string {
    return "users"
}

func (u *User) ToEntity() entity.User {
    return entity.User{
        Id:     u.ID,
        Name:   u.Name,
        Email:  u.Email,
        Status: u.Status,
    }
}

func NewUser(user entity.User) User {
    return User{
        ID:     user.Id,
        Name:   user.Name,
        Email:  user.Email,
        Status: user.Status,
    }
}
```

#### GORM Relationships

```go
type Order struct {
    DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at"`
    Items     []OrderItem    `json:"items" gorm:"foreignKey:OrderId;references:Id"`
    Status    string         `json:"status" gorm:"column:status"`
    ID        int64          `json:"id" gorm:"primaryKey;column:id"`
    UserId    int64          `json:"user_id" gorm:"column:user_id"`
}

func (Order) TableName() string {
    return "orders"
}
```

#### GORM Hooks

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    u.CreatedAt = time.Now().UTC()
    return nil
}

func (u *User) AfterFind(tx *gorm.DB) error {
    // Post-query transformations if needed
    return nil
}
```

---

### Type 2: MongoDB Models (BSON)

MongoDB models use `bson:` tags for document field mapping.

**Key characteristics:**
- `bson:` tags for field names
- `bson:"_id"` for the document ID
- `bson:",inline"` for embedding
- Can have aggregation pipeline helpers

```go
package model

import (
    "time"
    "myproject/internal/domain/entity"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Name      string             `bson:"name"`
    Email     string             `bson:"email"`
    Status    string             `bson:"status"`
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
}

func (u *User) ToEntity() entity.User {
    return entity.User{
        Id:     u.ID.Hex(),
        Name:   u.Name,
        Email:  u.Email,
        Status: u.Status,
    }
}

func NewUser(user entity.User) User {
    return User{
        Name:      user.Name,
        Email:     user.Email,
        Status:    user.Status,
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
    }
}
```

MongoDB models may also need inner methods for complex document operations:

```go
// Collection name helper
func (User) CollectionName() string {
    return "users"
}

// Index definitions
func (User) Indexes() []mongo.IndexModel {
    return []mongo.IndexModel{
        {Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
    }
}
```

---

### Type 3: Key-Value NoSQL Models (DynamoDB)

DynamoDB models use `dynamodbav:` tags for attribute mapping.

**Key characteristics:**
- `dynamodbav:` tags for attribute names
- Partition key and sort key attributes
- Simple flat structures (DynamoDB favors denormalization)

```go
package model

import (
    "myproject/internal/domain/entity"
)

type Session struct {
    UserId    string `dynamodbav:"user_id"`
    SessionId string `dynamodbav:"session_id"`
    Token     string `dynamodbav:"token"`
    ExpiresAt string `dynamodbav:"expires_at"`
    CreatedAt string `dynamodbav:"created_at"`
}

func (s *Session) ToEntity() entity.Session {
    return entity.Session{
        UserId:    s.UserId,
        SessionId: s.SessionId,
        Token:     s.Token,
    }
}

func NewSession(session entity.Session) Session {
    return Session{
        UserId:    session.UserId,
        SessionId: session.SessionId,
        Token:     session.Token,
        ExpiresAt: session.ExpiresAt.UTC().Format("2006-01-02T15:04:05.000Z"),
        CreatedAt: session.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
    }
}
```

---

## Code Examples

### Example 1 — Simple GORM model with ToEntity

```go
package model

import (
    "time"
    "myproject/internal/domain/entity"
)

type Credential struct {
    Token          *string    `json:"token" gorm:"column:long_term_access_token"`
    ConnectionType *string    `json:"connection_type" gorm:"column:connection_type"`
    ExpiresAt      *time.Time `json:"expires_at" gorm:"column:expires_at"`
    UserId         int64      `json:"user_id" gorm:"primaryKey;column:fk_id_user"`
}

func (Credential) TableName() string {
    return "credentials"
}

func (c *Credential) ToEntity() entity.Credential {
    return entity.Credential{
        CreatorId:      c.UserId,
        Token:          c.Token,
        ConnectionType: c.ConnectionType,
        ExpiresAt:      c.ExpiresAt,
    }
}
```

Key points:
- Column names can differ from field names (database naming vs Go naming).
- `TableName()` maps to the actual database table.
- `ToEntity()` maps database columns to domain entity fields.

---

### Example 2 — GORM model with relationships and soft delete

```go
package model

import (
    "myproject/internal/domain/entity"
    "gorm.io/gorm"
)

type Assignment struct {
    DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at"`
    Status     string         `json:"status" gorm:"column:status"`
    Project    Project        `json:"project" gorm:"foreignKey:ProjectId;references:Id"`
    Id         int64          `json:"id" gorm:"primaryKey;column:id"`
    UserId     int64          `json:"user_id" gorm:"column:user_id"`
    ProjectId  int64          `json:"project_id" gorm:"column:project_id"`
}

func (Assignment) TableName() string {
    return "assignments"
}

func (i *Assignment) ToEntity() entity.Assignment {
    return entity.Assignment{
        UserId:   i.UserId,
        Status:   i.Status,
        Project:  i.Project.ToEntity(),
    }
}

func (i *Assignment) GetTaskIds() []int64 {
    return i.Project.GetTaskIds()
}
```

Key points:
- `gorm.DeletedAt` enables GORM soft delete.
- Foreign key relationships load nested models.
- `ToEntity()` delegates to nested model's `ToEntity()`.
- Helper methods navigate the relationship tree.

---

### Example 3 — GORM model with nested collection conversion

```go
package model

import (
    "myproject/internal/domain/entity"
    "gorm.io/gorm"
)

type Project struct {
    DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at"`
    Title     string         `json:"title" gorm:"column:title"`
    Schedules []Schedule     `json:"schedules" gorm:"foreignKey:ProjectId;references:Id"`
    Id        int64          `json:"id" gorm:"primaryKey;column:id"`
    StatusId  int            `json:"status_id" gorm:"column:status_id"`
}

func (Project) TableName() string {
    return "projects"
}

func (p *Project) ToEntity() entity.Project {
    schedules := make([]entity.Schedule, 0)
    for _, schedule := range p.Schedules {
        schedules = append(schedules, schedule.ToEntity())
    }

    return entity.Project{
        Id:        p.Id,
        Title:     p.Title,
        StatusId:  p.StatusId,
        Schedules: schedules,
    }
}

func (p *Project) GetTaskIds() []int64 {
    taskIds := make([]int64, 0)
    for _, schedule := range p.Schedules {
        taskIds = append(taskIds, schedule.GetTaskIds()...)
    }
    return taskIds
}
```

Key points:
- Slice conversion iterates and calls `ToEntity()` on each nested model.
- Helper methods aggregate data from nested collections.

---

### Example 4 — GORM model with enum conversion helpers

```go
package model

import (
    "myproject/internal/domain/enums"
    "myproject/internal/domain/enums/mediatype"
)

type Format struct {
    Name           string       `json:"name" gorm:"column:name"`
    FormatSource   FormatSource `json:"format_source" gorm:"foreignKey:FormatSourceId;references:Id"`
    Id             int64        `json:"id" gorm:"primaryKey;column:id"`
    FormatSourceId int64        `json:"format_source_id" gorm:"column:format_source_id"`
}

func (Format) TableName() string {
    return "formats"
}

func (f *Format) GetMediaType() enums.MediaType {
    return mediatype.GetMediaType(f.Name)
}

func (f *Format) GetChannel() enums.Channel {
    return f.FormatSource.GetChannel()
}
```

Key points:
- No `ToEntity()` — this model is a helper used by other models.
- Converts raw database strings to domain enums via parser functions.
- Delegates through relationships (`FormatSource.GetChannel()`).

---

### Example 5 — GORM model with complex ToEntity (cross-model mapping)

```go
package model

import (
    "time"
    "myproject/internal/domain/entity"
    "gorm.io/gorm"
)

type Task struct {
    Content   *Content       `json:"content" gorm:"foreignKey:TaskId;references:Id"`
    PostedAt  *time.Time     `json:"posted_at" gorm:"column:posted_at"`
    DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at"`
    Item      Item           `json:"item" gorm:"foreignKey:ItemId;references:Id"`
    Id        int64          `json:"id" gorm:"primaryKey;column:id"`
    ItemId    int64          `json:"item_id" gorm:"column:item_id"`
    UserId    int64          `json:"user_id" gorm:"column:user_id"`
}

func (Task) TableName() string {
    return "tasks"
}

func (t *Task) ToEntity() entity.Task {
    var mediaUrl *string
    if t.Content != nil {
        mediaUrl = &t.Content.MediaUrl
    }

    return entity.Task{
        TaskId:    t.ItemId,
        MediaUrl:  mediaUrl,
        Schedule:  t.Item.Schedule.ToEntity(),
        PostedAt:  t.PostedAt,
        MediaType: t.Item.GetMediaType(),
        Channel:   t.Item.GetChannel(),
    }
}
```

Key points:
- Maps across multiple nested models to produce a flat entity.
- Handles nullable relationships (`Content` pointer check).
- Uses helper methods on related models for enum conversion.

---

### Example 6 — Simple supporting model (no ToEntity needed)

```go
package model

import "gorm.io/gorm"

type Content struct {
    DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at"`
    MediaUrl  string         `json:"media_url" gorm:"column:media_url"`
    Id        int64          `json:"id" gorm:"primaryKey;column:id"`
    TaskId    int64          `json:"task_id" gorm:"column:task_id"`
}

func (Content) TableName() string {
    return "content"
}
```

Key points:
- Not every model needs `ToEntity()` — some are purely supporting models used by others.
- Still needs `TableName()` for GORM mapping.

---

## Anti-Patterns

### Anti-Pattern 1 — "Model" in the name (FORBIDDEN)

```go
// BAD
type UserModel struct { ... }
type OrderModel struct { ... }

// GOOD
type User struct { ... }
type Order struct { ... }
```

---

### Anti-Pattern 2 — Model in the domain layer (FORBIDDEN)

```go
// BAD
// File: internal/domain/model/user.go
type User struct {
    ID int64 `gorm:"primaryKey"`  // WRONG — domain has no database tags
}
```

Models live exclusively in `integration/persistence/model/`.

---

### Anti-Pattern 3 — Domain entity importing model (FORBIDDEN)

```go
// BAD — entity knows about models
// File: internal/domain/entity/user.go
import "myproject/internal/integration/persistence/model"

func (u *User) ToModel() model.User { ... }  // WRONG
```

Conversion lives on the model, not the entity.

---

### Anti-Pattern 4 — Business logic on models (FORBIDDEN)

```go
// BAD — business logic belongs on the entity
func (u *User) IsActive() bool {
    return u.Status == "active" && u.DeletedAt.Time.IsZero()  // WRONG
}
```

Models can have database-level helpers (navigation, aggregation, enum conversion), but business logic belongs on domain entities.

---

### Anti-Pattern 5 — Missing TableName for GORM models (FORBIDDEN)

```go
// BAD — GORM will infer table name, which may not match the actual table
type User struct {
    ID int64 `gorm:"primaryKey"`
}
// GORM infers "users" — but what if the table is "app_users"?

// GOOD — always explicit
func (User) TableName() string {
    return "app_users"
}
```

---

## Quick Reference Checklist

When creating or reviewing a model, verify every item:

- [ ] Lives in `internal/integration/persistence/model/` and declares `package model`
- [ ] **One struct per file** — no multiple struct definitions in the same file
- [ ] Name does NOT contain "Model"
- [ ] Has appropriate database tags (`gorm:`, `bson:`, `dynamodbav:`, etc.)
- [ ] GORM models have `json:` tags alongside for BDD test support
- [ ] GORM models have `TableName()` method (always explicit)
- [ ] Has `ToEntity()` method for model → entity conversion
- [ ] Has `New{Name}(entity)` constructor for entity → model conversion (if writes are needed)
- [ ] Domain layer does NOT import the model — conversion is on the model side
- [ ] Helper methods are database-level (navigation, aggregation, enum conversion) — NOT business logic
- [ ] Nested model collections use loop + `ToEntity()` for conversion
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
