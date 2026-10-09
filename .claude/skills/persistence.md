---
name: Persistence (Clean Architecture)
description: This skill should be used when the user asks to "create a persistence", "add a persistence layer", "implement a repository", "create database access", "add data retrieval", "implement caching", or mentions anything about persistence in the internal/integration/persistence/ package.
version: 1.0.0
---

# Persistence — Golang Clean Architecture

Persistence files are the integration layer's data access implementations. They implement use case interfaces (via integration adapters), interact with databases through models, and return domain entities. They are the bridge between the application layer's contracts and the actual data stores.

---

## Core Rules

### Rule 1: Location

Persistence files MUST live in `internal/integration/persistence/` and declare `package persistence`.

```
internal/
└── integration/
    └── persistence/
        ├── model/               ← Database models (see Models skill)
        │   ├── user.go
        │   └── order.go
        ├── user.go              ← persistence implementation
        ├── order.go
        └── credential.go
```

---

### Rule 2: Implements Integration Adapters (Which Aggregate Use Cases)

Persistence files implement adapter interfaces from `internal/integration/adapter/`. These adapters aggregate use case interfaces, so the persistence effectively implements the use case contracts.

```
usecase.FindUser (application layer)
        ↓ embedded in
adapter.UserPersistence (integration layer)
        ↓ implemented by
persistence.userPersistence (integration layer)
```

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// Adapter: internal/integration/adapter/user_persistence.go
type UserPersistence interface {
    usecase.FindUser
    usecase.SaveUser
}

// Implementation: internal/integration/persistence/user.go
type userPersistence struct {
    db    *gorm.DB
    cache redis.Client
}

func NewUserPersistence(db *gorm.DB, cache redis.Client) adapter.UserPersistence {
    return &userPersistence{db: db, cache: cache}
}
```

---

### Rule 3: Naming Conventions

| Element | Pattern | Example |
|---|---|---|
| Struct | Unexported, `{name}Persistence` | `userPersistence`, `orderPersistence` |
| Constructor | `New{Name}Persistence(...)` | `NewUserPersistence(...)` |
| Constructor return type | Integration adapter | `adapter.UserPersistence` |
| Package | `persistence` | `package persistence` |

The struct is **always unexported** — it lives inside the package only, exposed via the constructor that returns the adapter interface.

```go
// GOOD — unexported struct, exported constructor returning adapter
type credentialPersistence struct {
    db *gorm.DB
}

func NewCredentialPersistence(db *gorm.DB) adapter.CredentialPersistence {
    return &credentialPersistence{db: db}
}
```

---

### Rule 4: The Data Retrieval Pattern

Persistence follows a consistent pattern:
1. Receive domain-typed parameters (from use case contract)
2. Query the database using models
3. Convert model to entity via `model.ToEntity()`
4. Return domain entity

```go
func (p *userPersistence) FindById(ctx context.Context, userId int64) (*entity.User, error) {
    logger.Debug(ctx, "Started", map[string]any{"userId": userId})

    var userModel model.User
    err := p.db.WithContext(ctx).
        Model(&userModel).
        Where("id = ?", userId).
        First(&userModel).Error
    if err != nil {
        return nil, errs.DbError(err)
    }

    user := userModel.ToEntity()

    logger.Debug(ctx, "Finished", map[string]any{"userId": userId})
    return &user, nil
}
```

---

### Rule 5: Logger — Debug Level (Started/Finished)

Every persistence method MUST log `Started` and `Finished` at **Debug** level (not Info like controllers/services), including the raw input/output parameters.

```go
func (p *userPersistence) FindById(ctx context.Context, userId int64) (*entity.User, error) {
    logger.Debug(ctx, "Started", map[string]any{"userId": userId})

    // ... database query ...

    logger.Debug(ctx, "Finished", map[string]any{"userId": userId})
    return &user, nil
}
```

Debug level because persistence calls happen frequently and would be too noisy at Info level. They are invaluable for troubleshooting but should not fill production logs.

---

### Rule 6: Error Handling — Use errs.DbError

Database errors MUST be wrapped using `errs.DbError(err)` which auto-detects the error type:
- `gorm.ErrRecordNotFound` → `NotFoundError`
- Unique constraint violation → `BadRequestError`
- Everything else → `InternalServerError`

```go
err := p.db.WithContext(ctx).First(&userModel).Error
if err != nil {
    return nil, errs.DbError(err)  // Auto-detects and creates appropriate error with code
}
```

Error codes are generated at the origin (see Error Handler skill). For database errors, `errs.DbError` handles this automatically.

---

### Rule 7: Can Implement Multiple Use Cases

A single persistence file can implement multiple use case methods. This is acceptable because the underlying data source is the same.

```go
// adapter.UserPersistence embeds both usecase.FindUser and usecase.SaveUser
type userPersistence struct {
    db *gorm.DB
}

func (p *userPersistence) FindById(ctx context.Context, userId int64) (*entity.User, error) { ... }
func (p *userPersistence) FindByEmail(ctx context.Context, email string) (*entity.User, error) { ... }
func (p *userPersistence) Save(ctx context.Context, user entity.User) error { ... }
```

---

### Rule 8: Caching Belongs in Persistence (Not Domain)

When caching is needed, it MUST live in the persistence layer. The domain should not know something is cached.

```go
type userPersistence struct {
    db    *gorm.DB
    cache redis.Client
}

func (p *userPersistence) FindById(ctx context.Context, userId int64) (*entity.User, error) {
    logger.Debug(ctx, "Started", map[string]any{"userId": userId})

    // Try cache first
    cached, err := p.cache.Get(ctx, fmt.Sprintf("user:%d", userId)).Result()
    if err == nil {
        var userModel model.User
        json.Unmarshal([]byte(cached), &userModel)
        user := userModel.ToEntity()
        logger.Debug(ctx, "Finished (cache hit)", map[string]any{"userId": userId})
        return &user, nil
    }

    // Fallback to database
    var userModel model.User
    err = p.db.WithContext(ctx).Where("id = ?", userId).First(&userModel).Error
    if err != nil {
        return nil, errs.DbError(err)
    }

    // Store in cache
    data, _ := json.Marshal(userModel)
    p.cache.Set(ctx, fmt.Sprintf("user:%d", userId), data, 10*time.Minute)

    user := userModel.ToEntity()

    logger.Debug(ctx, "Finished (cache miss)", map[string]any{"userId": userId})
    return &user, nil
}
```

---

### Rule 9: No Calling Other Persistence Files (Use Preloads Instead)

Persistence files MUST NOT call other persistence files — that breaks single responsibility. Instead, use GORM's `Preload` to load related data through properly defined model relationships (one-to-many, many-to-many, etc.).

```go
// BAD — persistence calling another persistence
type orderPersistence struct {
    db              *gorm.DB
    userPersistence adapter.UserPersistence  // WRONG — breaks SRP
}

func (p *orderPersistence) FindById(ctx context.Context, orderId int64) (*entity.Order, error) {
    // ...
    user, err := p.userPersistence.FindById(ctx, order.UserId)  // WRONG
}

// GOOD — use Preload to load relationships
func (p *orderPersistence) FindById(ctx context.Context, orderId int64) (*entity.Order, error) {
    var orderModel model.Order
    err := p.db.WithContext(ctx).
        Model(&orderModel).
        Preload("User").
        Preload("Items.Product").
        Where("id = ?", orderId).
        First(&orderModel).Error
    // ...
}
```

If the related data comes from a **different database** (not the same GORM connection), then injecting another persistence adapter is acceptable as an exception.

---

### Rule 10: GORM Queries — Never Raw, Never Scan (CRITICAL)

All GORM queries MUST use GORM's fluent API methods. Raw SQL and `Scan` are FORBIDDEN.

**Allowed GORM methods:**
- `Where()`, `Or()`, `Not()`
- `Select()`, `Omit()`
- `First()`, `Find()`, `Last()`, `Take()`
- `Create()`, `Save()`, `Update()`, `Updates()`, `Delete()`
- `Preload()`, `Joins()`
- `Limit()`, `Offset()`, `Order()`, `Group()`, `Having()`
- `Count()`, `Pluck()`, `Distinct()`
- `Model()`, `Table()`
- `Scopes()` for reusable query conditions
- Subqueries via `p.db.Where("column IN (?)", p.db.Model(...).Select(...))`

```go
// BAD — raw SQL (FORBIDDEN)
err := p.db.Raw("SELECT * FROM users WHERE id = ?", userId).Scan(&userModel).Error

// BAD — Scan (FORBIDDEN)
err := p.db.Table("users").Where("id = ?", userId).Scan(&userModel).Error

// GOOD — GORM fluent API
err := p.db.WithContext(ctx).
    Model(&userModel).
    Where("id = ?", userId).
    First(&userModel).Error

// GOOD — complex query with multiple conditions
err := p.db.WithContext(ctx).
    Model(&orderModel).
    Preload("Items").
    Where("user_id = ? AND status = ?", userId, "active").
    Order("created_at DESC").
    Limit(10).
    Find(&orders).Error

// GOOD — subquery using GORM (not raw SQL)
err := p.db.WithContext(ctx).
    Model(&orderModel).
    Where("user_id IN (?)",
        p.db.Model(&model.Assignment{}).
            Select("user_id").
            Where("project_id = ?", projectId),
    ).
    Find(&orders).Error

// GOOD — GORM scopes for reusable conditions
func ActiveScope(db *gorm.DB) *gorm.DB {
    return db.Where("deleted_at IS NULL AND status = ?", "active")
}

err := p.db.WithContext(ctx).
    Scopes(ActiveScope).
    Where("user_id = ?", userId).
    Find(&orders).Error
```

---

## Code Examples

### Example 1 — Simple database query

```go
package persistence

import (
    "context"

    "myproject/internal/domain/entity"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/persistence/model"
    "myproject/pkg/errs"
    "myproject/pkg/logger"
    "gorm.io/gorm"
)

type credentialPersistence struct {
    db *gorm.DB
}

func NewCredentialPersistence(db *gorm.DB) adapter.CredentialPersistence {
    return &credentialPersistence{db: db}
}

func (c *credentialPersistence) GetByUserId(ctx context.Context, userId int64) (*entity.Credential, error) {
    logger.Debug(ctx, "Started", map[string]any{"userId": userId})

    var credentialModel model.Credential
    err := c.db.WithContext(ctx).
        Model(&credentialModel).
        Where("fk_id_user = ?", userId).
        First(&credentialModel).Error
    if err != nil {
        return nil, errs.DbError(err)
    }

    credential := credentialModel.ToEntity()

    logger.Debug(ctx, "Finished", map[string]any{"userId": userId})
    return &credential, nil
}
```

---

### Example 2 — Complex query with deep preloads

```go
package persistence

import (
    "context"

    "myproject/internal/domain/entity"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/persistence/model"
    "myproject/pkg/errs"
    "myproject/pkg/logger"
    "gorm.io/gorm"
)

type assignmentPersistence struct {
    db *gorm.DB
}

func NewAssignmentPersistence(db *gorm.DB) adapter.AssignmentPersistence {
    return &assignmentPersistence{db: db}
}

func (p *assignmentPersistence) GetByProjectAndUser(ctx context.Context, projectId, userId int64) (*entity.Assignment, error) {
    logger.Debug(ctx, "Started", map[string]any{"projectId": projectId, "userId": userId})

    var assignmentModel model.Assignment
    err := p.db.WithContext(ctx).
        Model(&assignmentModel).
        Preload("Project.Schedules.Tasks").
        Preload("Credential").
        Where("project_id = ? AND user_id = ?", projectId, userId).
        First(&assignmentModel).Error
    if err != nil {
        return nil, errs.DbError(err)
    }

    var tasksModel []model.Task
    err = p.db.WithContext(ctx).
        Model(&tasksModel).
        Preload("Item.Schedule").
        Preload("Item.Format.FormatSource").
        Preload("Content").
        Where("item_id IN ? AND user_id = ?", assignmentModel.GetTaskIds(), userId).
        Find(&tasksModel).Error
    if err != nil {
        return nil, errs.DbError(err)
    }

    tasks := make([]entity.Task, 0)
    for _, taskModel := range tasksModel {
        tasks = append(tasks, taskModel.ToEntity())
    }

    assignment := assignmentModel.ToEntity()
    assignment.Tasks = tasks

    logger.Debug(ctx, "Finished", map[string]any{"projectId": projectId, "userId": userId})
    return &assignment, nil
}
```

Key points:
- Uses deep `Preload` chains to load related data in one query.
- Multiple queries within the same persistence for different parts of the aggregate.
- Converts each model to entity individually, then assembles the aggregate.
- No calls to other persistence files — all data comes from the same DB via Preload.

---

### Example 3 — Cloud storage persistence (non-database)

```go
package persistence

import (
    "bytes"
    "context"
    "fmt"
    "strconv"

    storage "myproject/pkg/contracts/storage"

    "myproject/internal/domain/enums"
    "myproject/internal/integration/adapter"
    "myproject/pkg/logger"
)

type mediaPersistence struct {
    cloudStorage storage.CloudStorage
}

func NewMediaPersistence(cloudStorage storage.CloudStorage) adapter.MediaPersistence {
    return &mediaPersistence{cloudStorage: cloudStorage}
}

func (p *mediaPersistence) Save(ctx context.Context, data []byte, fileName string, mediaType enums.MediaType, projectId, userId int64) (*string, error) {
    logger.Debug(ctx, "Started", map[string]any{
        "projectId": projectId,
        "userId":    userId,
        "fileName":  fileName,
    })

    filePath := strconv.FormatInt(projectId, 10) + "/" + strconv.FormatInt(userId, 10)
    objectKey := filePath + "/" + fileName

    err := p.cloudStorage.Put(ctx, "media-bucket", objectKey, mediaType.String(), bytes.NewReader(data))
    if err != nil {
        return nil, err
    }

    result := fmt.Sprintf("/media-bucket/%s", objectKey)

    logger.Debug(ctx, "Finished", map[string]any{
        "projectId": projectId,
        "userId":    userId,
        "filePath":  result,
    })
    return &result, nil
}
```

Key points:
- Persistence is not just databases — it includes any external data storage (S3, GCS, etc.).
- Same pattern: debug logs, error passthrough, returns domain types.

---

### Example 4 — Persistence with caching

```go
package persistence

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "myproject/internal/domain/entity"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/persistence/model"
    "myproject/pkg/errs"
    "myproject/pkg/logger"
    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
)

type configPersistence struct {
    db    *gorm.DB
    cache *redis.Client
}

func NewConfigPersistence(db *gorm.DB, cache *redis.Client) adapter.ConfigPersistence {
    return &configPersistence{db: db, cache: cache}
}

func (p *configPersistence) GetByKey(ctx context.Context, key string) (*entity.Config, error) {
    logger.Debug(ctx, "Started", map[string]any{"key": key})

    // Try cache
    cached, err := p.cache.Get(ctx, fmt.Sprintf("config:%s", key)).Result()
    if err == nil {
        var configModel model.Config
        _ = json.Unmarshal([]byte(cached), &configModel)
        config := configModel.ToEntity()
        logger.Debug(ctx, "Finished (cache hit)", map[string]any{"key": key})
        return &config, nil
    }

    // Fallback to DB
    var configModel model.Config
    err = p.db.WithContext(ctx).Where("key = ?", key).First(&configModel).Error
    if err != nil {
        return nil, errs.DbError(err)
    }

    // Cache for next time
    data, _ := json.Marshal(configModel)
    p.cache.Set(ctx, fmt.Sprintf("config:%s", key), data, 30*time.Minute)

    config := configModel.ToEntity()
    logger.Debug(ctx, "Finished (cache miss)", map[string]any{"key": key})
    return &config, nil
}
```

---

## Anti-Patterns

### Anti-Pattern 1 — Exported struct (FORBIDDEN)

```go
// BAD — struct is exported
type UserPersistence struct {  // WRONG — must be unexported
    db *gorm.DB
}

// GOOD — unexported struct, exported only via constructor
type userPersistence struct {
    db *gorm.DB
}
```

---

### Anti-Pattern 2 — Returning models instead of entities (FORBIDDEN)

```go
// BAD — returning model to the application layer
func (p *userPersistence) FindById(ctx context.Context, userId int64) (*model.User, error) {  // WRONG
    // ...
}

// GOOD — always return domain entities
func (p *userPersistence) FindById(ctx context.Context, userId int64) (*entity.User, error) {
    // ... query model, then convert
    user := userModel.ToEntity()
    return &user, nil
}
```

---

### Anti-Pattern 3 — Business logic in persistence (FORBIDDEN)

```go
// BAD — business logic in persistence
func (p *userPersistence) FindById(ctx context.Context, userId int64) (*entity.User, error) {
    // ...
    user := userModel.ToEntity()
    if !user.IsActive() {  // WRONG — business decision
        return nil, nil
    }
    return &user, nil
}
```

Persistence retrieves data. Business logic belongs on entities, orchestrated by services.

---

### Anti-Pattern 4 — Caching in domain or application layer (FORBIDDEN)

```go
// BAD — cache in service
func (s *processOrder) Process(ctx context.Context, ...) error {
    cached := s.cache.Get(ctx, "user:1")  // WRONG — service shouldn't know about cache
    // ...
}

// GOOD — cache is hidden inside persistence
// The service just calls: s.findUser.FindById(ctx, 1)
// Persistence handles caching transparently
```

---

### Anti-Pattern 5 — Info-level logs in persistence (WRONG LEVEL)

```go
// BAD — too noisy for production
logger.Info(ctx, "Started", ...)   // WRONG — use Debug for persistence

// GOOD
logger.Debug(ctx, "Started", ...)
```

---

### Anti-Pattern 6 — Raw SQL queries (FORBIDDEN)

```go
// BAD — raw SQL
err := p.db.Raw("SELECT * FROM users WHERE id = ? AND status = 'active'", userId).Scan(&user).Error

// BAD — Scan instead of First/Find
err := p.db.Table("users").Where("id = ?", userId).Scan(&user).Error

// GOOD — GORM fluent API
err := p.db.WithContext(ctx).
    Model(&user).
    Where("id = ? AND status = ?", userId, "active").
    First(&user).Error
```

---

### Anti-Pattern 7 — Persistence calling another persistence (FORBIDDEN)

```go
// BAD — breaks single responsibility
type orderPersistence struct {
    db              *gorm.DB
    userPersistence adapter.UserPersistence  // WRONG
}

func (p *orderPersistence) FindById(ctx context.Context, orderId int64) (*entity.Order, error) {
    // ...
    user, _ := p.userPersistence.FindById(ctx, order.UserId)  // WRONG
}

// GOOD — use Preload for related data
func (p *orderPersistence) FindById(ctx context.Context, orderId int64) (*entity.Order, error) {
    err := p.db.WithContext(ctx).
        Preload("User").
        Where("id = ?", orderId).
        First(&orderModel).Error
}
```

Exception: if the data lives in a **different database** (different GORM connection), injecting another persistence adapter is acceptable.

---

## Quick Reference Checklist

When creating or reviewing a persistence file, verify every item:

- [ ] Lives in `internal/integration/persistence/` and declares `package persistence`
- [ ] Struct is **unexported** (`userPersistence`, not `UserPersistence`)
- [ ] Constructor is `New{Name}Persistence(...)` returning `adapter.{Name}Persistence`
- [ ] Implements integration adapter that aggregates use case interfaces
- [ ] Uses models for database interaction, converts to entities via `ToEntity()`
- [ ] Returns domain entities, never models
- [ ] Logs `Started` and `Finished` at **Debug** level with raw inputs/outputs
- [ ] Database errors use `errs.DbError(err)` for auto-detection
- [ ] Caching (when present) is transparent — domain/application layers don't know
- [ ] Does NOT call other persistence files (use Preload instead, unless different DB)
- [ ] GORM queries use fluent API only — no `Raw()`, no `Scan()`, no raw SQL strings
- [ ] Subqueries use GORM's fluent API (not raw SQL embedded in strings)
- [ ] No business logic — only data retrieval, conversion, and caching
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
