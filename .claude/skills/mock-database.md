---
name: Mock - Database (SQLite In-Memory)
description: Use when creating or referencing the SQLite in-memory database mock for BDD tests.
version: 1.0.0
---

# Mock: Database (SQLite In-Memory)

Replaces MySQL/PostgreSQL with SQLite in-memory for BDD tests. Uses GORM so the same queries work in both production and tests.

**Location:** `test/integration/mock/db.go`

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package mock

import (
    "fmt"
    "sync"
    "sync/atomic"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

type Db struct {
    DbConn *gorm.DB
    models map[string]any
}

var dbInstances = make(map[string]*Db)
var dbMu sync.Mutex

func NewDb(name string, models map[string]any) *Db {
    dbMu.Lock()
    defer dbMu.Unlock()

    if db, exists := dbInstances[name]; exists {
        return db
    }

    conn, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", name)), &gorm.Config{})
    if err != nil {
        panic("failed to connect to test database: " + err.Error())
    }

    for _, model := range models {
        if err := conn.AutoMigrate(model); err != nil {
            panic("failed to migrate model: " + err.Error())
        }
    }

    db := &Db{DbConn: conn, models: models}
    dbInstances[name] = db
    return db
}

func (d *Db) ClearDB() error {
    for tableName, model := range d.models {
        if err := d.DbConn.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(model).Error; err != nil {
            return fmt.Errorf("failed to clear table %s: %w", tableName, err)
        }
        // Reset auto-increment
        d.DbConn.Exec(fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name='%s'", tableName))
    }
    return nil
}

func (d *Db) GetModel(tableName string) any {
    return d.models[tableName]
}
```

**Usage in step definitions:**
```go
test.mainDb = mock.NewDb("main", map[string]any{
    "users":  &model.User{},
    "orders": &model.Order{},
})
database.SetMainDb(test.mainDb.DbConn)
```

**Supports multiple databases** (multiton pattern):
```go
test.maestroDb = mock.NewDb("maestro", maestroModels)
test.catalystDb = mock.NewDb("catalyst", catalystModels)
```

**Lifecycle:** The database mock is created once and injected once via `database.SetMainDb(test.mainDb.DbConn)` in `InitializeScenario`. Between BDD scenarios, `ClearDB()` truncates all tables (clearing state) — the mock instance and DB connection are never recreated or re-injected.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
