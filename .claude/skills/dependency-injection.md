---
name: Dependency Injection (Clean Architecture)
description: This skill should be used when the user asks to "create a factory", "add dependency injection", "wire dependencies", "create an injector", "add a singleton", "create infrastructure dependency", "create application startup", or mentions anything about factories in the internal/infra/dependency/ folder or application startup in internal/infra/application.go.
version: 1.0.0
---

# Dependency Injection — Golang Clean Architecture

Dependencies are wired using distributed singleton factories. Each factory creates and caches a single dependency, pulling other factories as needed. There is NO centralized injector file — the dependency graph builds naturally as factories call other factories, forming a pyramid that resolves itself at startup.

---

## Core Rules

### Rule 1: Distributed Factory Pattern (No Centralized Injector)

There is NO `injector.go` file. Instead, each dependency has its own factory file that:
1. Creates the dependency as a singleton (`sync.Once`)
2. Pulls other factories it depends on
3. Exposes a getter (and a setter only for external infrastructure dependencies — see Rule 5)

The dependency graph resolves naturally — when the top-level factory (e.g., Handler) is requested, it pulls its dependencies, which pull their dependencies, and so on until everything is wired.

```
GetHandler()
  → GetErrorHandler()
  → GetFilterService()
      → GetInvitePersistence()
          → GetMaestroDb()
      → GetPostValidatorFactory()
          → GetStoryValidator()
          → GetReelsValidator()
      → GetPostDownloaderFactory()
          → GetInstagramWebservice()
              → GetHttpClient()
      → GetPostPersistence()
          → GetCloudStorage()
      → GetVideoProcessingPublisher()
          → GetBroker()
```

---

### Rule 2: Location

Factories MUST live in `internal/infra/dependency/` organized by type in sub-folders.

```
internal/
└── infra/
    └── dependency/
        ├── database/
        │   ├── redis.go              ← database.GetRedis()
        │   ├── main_db.go            ← database.GetMainDb()
        │   └── readonly_db.go        ← database.GetReadonlyDb()
        ├── integration/
        │   ├── user_persistence.go   ← integration.GetUserPersistence()
        │   ├── payment_webservice.go ← integration.GetPaymentWebservice()
        │   ├── order_publisher.go    ← integration.GetOrderPublisher()
        │   └── handler.go           ← integration.GetHandler()
        ├── application/
        │   ├── process_order.go      ← application.GetProcessOrderService()
        │   └── validator_factory.go  ← application.GetValidatorFactory()
        ├── http/
        │   └── client.go            ← httpclient.GetHttpClient()
        └── storage/
            └── cloud_storage.go     ← storage.GetCloudStorage()
```

---

### Rule 3: Naming Convention

| Element | Pattern | Example |
|---|---|---|
| Package | Type-based sub-folder | `database`, `integration`, `application`, `http` |
| Getter | `Get{DependencyName}()` | `database.GetRedis()`, `integration.GetUserPersistence()` |
| Setter | `Set{DependencyName}(instance)` | `database.SetRedis(mockRedis)` |
| Return type | Adapter interface or common lib interface | `adapter.UserPersistence`, `redis.UniversalClient`, `*gorm.DB` |

---

### Rule 4: Singleton with sync.Once (ALL Factories — No Exceptions)

**Every factory MUST be a singleton** using `sync.Once` + a cached instance variable. This applies to ALL factories without exception — external infrastructure factories (DB, Redis, HTTP, broker, storage) AND inner-layer factories (persistence, webservice, publisher, service, handler). The instance is created once and reused for the lifetime of the application. There are NO non-singleton factories in this architecture.

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package database

import (
    "sync"
    "github.com/redis/go-redis/v9"
)

var redisConnOnce sync.Once
var redisConn redis.UniversalClient

func GetRedis() redis.UniversalClient {
    if redisConn == nil {
        redisConnOnce.Do(func() {
            redisConn = openRedisConn()
        })
    }
    return redisConn
}
```

---

### Rule 5: Setter for ONE-TIME Mock Injection — ONLY for External Dependencies

Setters are ONLY for factories that wrap **external infrastructure dependencies** — things that are mocked in BDD tests (databases, Redis, HTTP clients, message brokers, cloud storage, etc.). These are the outermost layer dependencies.

**CRITICAL: Setters are for ONE-TIME injection only.** Mock instances are created once in `InitializeScenario` and injected once via setters. Between BDD scenarios, only the mock's internal STATE is reset (via `mock.Reset()`, `ClearDB()`, `ClearRedis()`) — the setter is NOT called again. The mock instance lives forever; only its accumulated data is cleared.

**Factories that MUST have setters** (external/injectable):
- Database connections (`SetMainDb`, `SetRedis`)
- HTTP clients (`SetHttpClient`)
- Cloud storage (`SetCloudStorage`)
- Message broker producers (`SetBrokerProducer`)
- Time providers (via `timeutils.SetCurrentTimeProvider`)

**Factories that MUST NOT have setters** (inner layers — never mocked):
- Services (e.g., `GetProcessOrderService` — NO `SetProcessOrderService`)
- Use case orchestrations
- Application factories (validator factory, downloader factory)
- Controllers and error handlers
- Persistence implementations
- Webservice implementations
- Publisher implementations

The rule: **if it's in the middle of the codebase (application/integration layer implementations), it must NOT be injectable.** Only the outermost infrastructure boundaries get setters. We mock the DB, not the persistence. We mock the HTTP client, not the webservice.

```go
// GOOD — external dependency, has setter for ONE-TIME injection
func SetRedis(client redis.UniversalClient) {
    redisConnOnce.Do(func() {})
    redisConn = client
}

func SetMainDb(db *gorm.DB) {
    mainDbOnce.Do(func() {})
    mainDbConn = db
}

// BAD — Reset function on factory (FORBIDDEN — factories NEVER have reset functions)
func ResetBrokerProducer() {        // WRONG — FORBIDDEN
    brokerProducerOnce = sync.Once{}
    brokerProducerInstance = nil
}

// BAD — service factory should NOT have a setter
func SetProcessOrderService(instance adapter.ProcessOrderService) {  // WRONG — never mock inner layers
    // ...
}
```

In BDD tests, mocks are injected ONCE in `InitializeScenario`, and only mock STATE is reset in `before()`:
```go
// InitializeScenario — ONE-TIME injection (runs once per test suite)
database.SetRedis(test.redis)
database.SetMainDb(test.mainDb.DbConn)
storage.SetCloudStorage(test.cloudStorage)
broker.SetBrokerBuilder(test.broker.Builder)

// before() — resets mock STATE only (runs before every scenario)
func before() {
    mock.ClearRedis(t.redis)       // clear accumulated keys
    t.broker.Reset()                // clear accumulated messages
    t.cloudStorage.Reset()          // clear accumulated files
    t.mainDb.ClearDB()              // truncate tables

    // BAD — never re-inject mocks in before() (that happens once in InitializeScenario)
    // database.SetRedis(test.redis)  // WRONG — do not call setters per-scenario
    // BAD — never inject service/persistence/controller mocks
    // dependency.SetProcessOrderService(mockService)  // WRONG
}
```

---

### Rule 6: Adapter Usage — Only When Needed

Adapters for infra dependencies (in `internal/infra/adapter/`) are only needed when a library does NOT provide a suitable interface.

**No adapter needed** (library already provides an interface):
- `*gorm.DB` — used as-is, SQLite works in tests with the same interface
- `redis.UniversalClient` — used as-is, miniredis implements the same interface
- `http.Client` interface — if already abstracted in `pkg/http`

**Adapter needed** (library has no interface):
- A custom cloud storage wrapper with no common interface
- A third-party SDK that only exposes concrete types

```go
// No adapter needed — library interface is sufficient
func GetRedis() redis.UniversalClient { ... }
func GetMainDb() *gorm.DB { ... }

// Adapter needed — no common interface exists
func GetCloudStorage() adapter.CloudStorage { ... }
```

---

### Rule 7: Factories Pull Other Factories

Factories get their dependencies by calling other factories — never by receiving them as parameters.

```go
// File: internal/infra/dependency/integration/user_persistence.go
package integration

import (
    "sync"
    "myproject/internal/infra/dependency/database"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/persistence"
)

var userPersistenceOnce sync.Once
var userPersistenceInstance adapter.UserPersistence

func GetUserPersistence() adapter.UserPersistence {
    if userPersistenceInstance == nil {
        userPersistenceOnce.Do(func() {
            userPersistenceInstance = persistence.NewUserPersistence(
                database.GetMainDb(),  // Pulls database factory
            )
        })
    }
    return userPersistenceInstance
}
```

---

### Rule 8: Application Startup (application.go)

The `internal/infra/application.go` file is the entry point called by `cmd/` main files. It retrieves the top-level factory (which triggers the full dependency cascade) and starts the application.

```go
// File: internal/infra/application.go
package application

import (
    "context"
    "log"

    "myproject/internal/infra/dependency/integration"
    "myproject/pkg/logger"
)

func Start() {
    logger.Info(context.TODO(), "Starting application")

    handler := integration.GetHandler()
    err := handler.Start(context.Background())
    if err != nil {
        log.Fatalf("Error starting application: %v", err)
    }
}
```

Multiple `cmd/` files share the same `application.Start()`:

```go
// File: cmd/server/main.go
package main

import application "myproject/internal/infra"

func main() {
    application.Start()
}
```

```go
// File: cmd/local/main.go
package main

import (
    "myproject/internal/infra/dependency/database"
    application "myproject/internal/infra"
)

func main() {
    // Local overrides for development
    database.SetRedis(localRedisClient)
    application.Start()
}
```

---

## Code Examples

### Example 1 — Infrastructure factory (database connection)

```go
// File: internal/infra/dependency/database/main_db.go
package database

import (
    "fmt"
    "sync"

    "myproject/internal/infra/properties"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

var mainDbOnce sync.Once
var mainDbConn *gorm.DB

func GetMainDb() *gorm.DB {
    if mainDbConn == nil {
        mainDbOnce.Do(func() {
            config := properties.Properties().Database.Main
            dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
                config.User, config.Password, config.Host, config.Port, config.Name)
            db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
            if err != nil {
                panic("failed to connect to main database: " + err.Error())
            }
            mainDbConn = db
        })
    }
    return mainDbConn
}

func SetMainDb(db *gorm.DB) {
    mainDbOnce.Do(func() {})
    mainDbConn = db
}
```

---

### Example 2 — Infrastructure factory (Redis)

```go
// File: internal/infra/dependency/database/redis.go
package database

import (
    "context"
    "sync"
    "time"

    "myproject/internal/infra/properties"
    "github.com/redis/go-redis/v9"
)

var redisConnOnce sync.Once
var redisConn redis.UniversalClient

func GetRedis() redis.UniversalClient {
    if redisConn == nil {
        redisConnOnce.Do(func() {
            redisConn = openRedisConn()
        })
    }
    return redisConn
}

func SetRedis(client redis.UniversalClient) {
    redisConnOnce.Do(func() {})
    redisConn = client
}

func openRedisConn() redis.UniversalClient {
    props := properties.Properties().Redis
    opts := &redis.Options{
        Addr:         props.Host + ":" + props.Port,
        Password:     props.Password,
        Username:     props.Username,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        MaxRetries:   3,
        PoolSize:     10,
    }
    client := redis.NewClient(opts)
    if _, err := client.Ping(context.TODO()).Result(); err != nil {
        panic("failed to connect to redis: " + err.Error())
    }
    return client
}
```

---

### Example 3 — Infrastructure factory (HTTP client)

```go
// File: internal/infra/dependency/http/client.go
package httpclient

import (
    "sync"
    "time"

    httpPkg "myproject/pkg/http"
    "myproject/pkg/logger"
)

var httpClientOnce sync.Once
var httpClientInstance httpPkg.Client

func GetHttpClient() httpPkg.Client {
    if httpClientInstance == nil {
        httpClientOnce.Do(func() {
            httpClientInstance = httpPkg.HttpClient(20*time.Second, logger.LoggerWrapper())
        })
    }
    return httpClientInstance
}

func SetHttpClient(client httpPkg.Client) {
    httpClientOnce.Do(func() {})
    httpClientInstance = client
}
```

---

### Example 4 — Integration factory (persistence, pulling database factory)

```go
// File: internal/infra/dependency/integration/user_persistence.go
package integration

import (
    "sync"
    "myproject/internal/infra/dependency/database"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/persistence"
)

var userPersistenceOnce sync.Once
var userPersistenceInstance adapter.UserPersistence

func GetUserPersistence() adapter.UserPersistence {
    if userPersistenceInstance == nil {
        userPersistenceOnce.Do(func() {
            userPersistenceInstance = persistence.NewUserPersistence(
                database.GetMainDb(),
            )
        })
    }
    return userPersistenceInstance
}
```

---

### Example 5 — Application factory (service, pulling integration factories)

```go
// File: internal/infra/dependency/application/process_order.go
package application

import (
    "sync"
    appAdapter "myproject/internal/application/adapter"
    "myproject/internal/application/service"
    "myproject/internal/application/usecase"
    "myproject/internal/infra/dependency/integration"
)

var processOrderOnce sync.Once
var processOrderInstance appAdapter.ProcessOrderService

func GetProcessOrderService() appAdapter.ProcessOrderService {
    if processOrderInstance == nil {
        processOrderOnce.Do(func() {
            processOrderInstance = service.NewProcessOrderService(
                usecase.GetAssignment(integration.GetAssignmentPersistence()),
                GetValidatorFactory(),
                GetDownloaderFactory(),
                usecase.SaveMedia(integration.GetMediaPersistence()),
                integration.GetOrderPublisher(),
            )
        })
    }
    return processOrderInstance
}
```

---

### Example 6 — Top-level factory (handler, pulling everything)

```go
// File: internal/infra/dependency/integration/handler.go
package integration

import (
    "sync"
    intAdapter "myproject/internal/integration/adapter"
    "myproject/internal/integration/entrypoint/controller"
    "myproject/internal/infra/dependency/application"
)

var handlerOnce sync.Once
var handlerInstance intAdapter.Handler

func GetHandler() intAdapter.Handler {
    if handlerInstance == nil {
        handlerOnce.Do(func() {
            handlerInstance = controller.NewHandler(
                GetErrorHandler(),
                application.GetProcessOrderService(),
            )
        })
    }
    return handlerInstance
}
```

---

### Example 7 — Application startup file

```go
// File: internal/infra/application.go
package application

import (
    "context"
    "log"

    "myproject/internal/infra/dependency/integration"
    "myproject/pkg/logger"
)

func Start() {
    logger.Info(context.TODO(), "Starting application")

    handler := integration.GetHandler()  // This triggers the full cascade
    err := handler.Start(context.Background())
    if err != nil {
        log.Fatalf("Error starting application: %v", err)
    }
}
```

---

### Example 8 — cmd/main.go (minimal)

```go
// File: cmd/server/main.go
package main

import application "myproject/internal/infra"

func main() {
    application.Start()
}
```

---

## The Dependency Cascade

When `integration.GetHandler()` is called, the full tree resolves:

```
integration.GetHandler()
├── integration.GetErrorHandler()
│   └── (no deps or usecase for notifications)
└── application.GetProcessOrderService()
    ├── integration.GetAssignmentPersistence()
    │   └── database.GetMainDb()
    ├── application.GetValidatorFactory()
    │   ├── service.NewStoryValidator()
    │   └── service.NewReelsValidator()
    ├── application.GetDownloaderFactory()
    │   └── integration.GetMediaWebservice()
    │       └── httpclient.GetHttpClient()
    ├── integration.GetMediaPersistence()
    │   └── storage.GetCloudStorage()
    └── integration.GetOrderPublisher()
        └── broker.GetMessageProducer()
```

Each node is created once (singleton), cached, and reused. The tree is lazy — nothing is created until first requested.

---

## Anti-Patterns

### Anti-Pattern 1 — Centralized injector file (OLD WAY — FORBIDDEN)

```go
// BAD — monolithic injector that wires everything in one place
type injector struct {
    db      *gorm.DB
    redis   redis.UniversalClient
    service adapter.ProcessOrderService
    handler adapter.Handler
}

func (i *injector) Wire(ctx context.Context) *injector {
    if i.db == nil { i.db = database.GetMainDb() }
    if i.service == nil { i.service = service.NewProcessOrder(i.db) }
    if i.handler == nil { i.handler = controller.NewHandler(i.service) }
    return i
}
```

Use distributed factories instead — each dependency owns its own creation.

---

### Anti-Pattern 2 — Factory receiving dependencies as parameters (FORBIDDEN)

```go
// BAD — factory receives deps as params
func GetUserPersistence(db *gorm.DB) adapter.UserPersistence {  // WRONG
    return persistence.NewUserPersistence(db)
}

// GOOD — factory pulls from other factories
func GetUserPersistence() adapter.UserPersistence {
    // ...
    userPersistenceInstance = persistence.NewUserPersistence(
        database.GetMainDb(),  // Pull from factory
    )
}
```

---

### Anti-Pattern 3 — Missing setter on external dependency factory (FORBIDDEN)

```go
// BAD — external dependency factory without setter (tests can't inject mock)
var dbOnce sync.Once
var dbConn *gorm.DB

func GetMainDb() *gorm.DB {
    dbOnce.Do(func() { dbConn = openConnection() })
    return dbConn
}
// Missing SetMainDb — tests can't inject SQLite mock
```

---

### Anti-Pattern 4 — Setter on inner layer factory (FORBIDDEN)

```go
// BAD — service factory should NOT have a setter (inner layer, never mocked)
func SetProcessOrderService(instance adapter.ProcessOrderService) {  // WRONG
    processOrderOnce.Do(func() {})
    processOrderInstance = instance
}

// BAD — persistence factory should NOT have a setter
func SetUserPersistence(instance adapter.UserPersistence) {  // WRONG
    userPersistenceOnce.Do(func() {})
    userPersistenceInstance = instance
}

// BAD — controller/handler factory should NOT have a setter
func SetHandler(instance adapter.Handler) {  // WRONG
    handlerOnce.Do(func() {})
    handlerInstance = instance
}
```

Inner layer factories (services, persistence, webservices, publishers, controllers) must NOT have setters. They are tested through the real code path — we mock their external dependencies (DB, Redis, HTTP, brokers), not the implementations themselves.

---

### Anti-Pattern 5 — Facade/injector.go file (FORBIDDEN)

There must be NO centralized `injector.go` or facade file that re-exports `Get*/Set*` from sub-packages. Consumers import the sub-packages directly. The factories naturally wire everything up — that is the whole point of the distributed pattern.

```go
// BAD — facade re-exporting from sub-packages
package dependency

func GetHandler() adapter.Handler {
    return integrationFactory.GetHandler()  // WRONG — unnecessary indirection
}

// GOOD — consumer imports sub-package directly
import integrationFactory "myproject/internal/infra/dependency/integration"

handler := integrationFactory.GetHandler()
```

---

### Anti-Pattern 6 — Test-only code in internal/ (FORBIDDEN)

Files like `reset.go` that contain test-only logic MUST NOT live in `internal/`. There are no factory reset functions to call anyway (see Anti-Pattern 7). Between BDD scenarios, only mock STATE is reset in the `before()` hook.

```go
// BAD — internal/infra/dependency/reset.go (test-only code in production tree)
package dependency

func ResetAll() { ... }  // WRONG — factories don't have reset functions

// BAD — calling factory resets from test helpers
package steps_test

func ResetAll() {
    brokerFactory.ResetBrokerProducer()   // WRONG — this function should not exist
    httpFactory.ResetHttpClient()          // WRONG — this function should not exist
    integrationFactory.ResetHandler()      // WRONG — this function should not exist
}

// GOOD — before() hook resets mock STATE only
func (t *testUtils) before() {
    mock.ClearRedis(t.redis)
    t.broker.Reset()
    t.cloudStorage.Reset()
    t.mainDb.ClearDB()
}
```

---

### Anti-Pattern 7 — Reset functions on factories (ABSOLUTELY FORBIDDEN)

```go
// BAD — Reset function on ANY factory (FORBIDDEN — factories NEVER have reset functions, period)
func ResetBrokerProducer() {        // WRONG — FORBIDDEN
    brokerProducerOnce = sync.Once{}
    brokerProducerInstance = nil
}

func ResetHttpClient() {            // WRONG — FORBIDDEN
    httpClientOnce = sync.Once{}
    httpClientInstance = nil
}

func ResetHandler() {               // WRONG — FORBIDDEN
    handlerOnce = sync.Once{}
    handlerInstance = nil
}

func ResetUserPersistence() {       // WRONG — FORBIDDEN
    userPersistenceOnce = sync.Once{}
    userPersistenceInstance = nil
}
```

**Factories NEVER have `Reset*()` functions — no external factory, no inner-layer factory, no factory of any kind.** Mock instances are injected ONCE via setters and live forever. Between BDD scenarios, only mock STATE is reset (via `mock.Reset()`, `ClearDB()`, `ClearRedis()`). The factories themselves are never touched after initial setup.

---

### Anti-Pattern 8 — Non-singleton factory (FORBIDDEN)

```go
// BAD — creates a new instance every call
func GetMainDb() *gorm.DB {
    return openConnection()  // WRONG — no caching, creates new conn each time
}
```

---

## Quick Reference Checklist

When creating or reviewing a factory, verify every item:

- [ ] Lives in `internal/infra/dependency/{type}/` organized by concern
- [ ] Uses `sync.Once` + cached instance for singleton initialization (ALL factories, no exceptions)
- [ ] Has a `Get{Name}()` getter that returns the cached instance
- [ ] **External dependency factories** (DB, Redis, HTTP, broker, storage) have `Set{Name}()` setter for ONE-TIME mock injection
- [ ] **Inner layer factories** (services, persistence, webservices, publishers, controllers) do NOT have setters
- [ ] **No `Reset*()` functions on ANY factory** — factories NEVER have reset functions, period
- [ ] Setter consumes `sync.Once` to prevent real initialization
- [ ] Setters are called ONCE in `InitializeScenario`, never per-scenario in `before()`
- [ ] Pulls dependencies from other factories (never receives them as parameters)
- [ ] Returns an adapter interface (or common lib interface if available)
- [ ] Infra adapters only created when the library has no suitable interface
- [ ] `application.go` calls the top-level factory to trigger the cascade
- [ ] No centralized injector/facade file — each dependency owns its own factory
- [ ] No test-only code in `internal/` — mock state reset belongs in `before()` hook
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
