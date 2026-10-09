---
name: Testing Setup (BDD Test Infrastructure)
description: This skill should be used when the user asks about "test setup", "test infrastructure", "how mocks are injected", "how to add a new factory to tests", "ResetAll pattern", "Before/After hooks", "test initialization", or when modifying the BDD test bootstrap in test/integration/steps/.
version: 1.0.0
---

# Testing Setup — BDD Test Infrastructure Patterns

This skill documents the BDD test infrastructure setup patterns, mock injection sequence, and how to extend the test setup when adding new dependencies.

---

## Factory Types: ALL Singletons, Two Categories

**ALL factories are singletons** using `sync.Once` + a cached instance. There are no non-singleton factories. The distinction is whether they have setters for mock injection.

### External Infrastructure Factories (Singletons WITH Setters)

These wrap external connections (DB, Redis, HTTP, broker, storage). They have **setters** for ONE-TIME mock injection in `InitializeScenario`. The setter is called ONCE — not per-scenario.

| Factory | Setter | What It Replaces |
|---|---|---|
| `database.SetRedis(client)` | Redis connection | Real Redis with miniredis |
| `database.SetMaestroDb(db)` | MySQL connection | Real MySQL with SQLite |
| `database.SetCatalystDb(db)` | MySQL connection | Real MySQL with SQLite |
| `storage.SetCloudStorageInstance(client)` | Cloud storage | Real S3 with in-memory mock |
| `broker.SetBrokerBuilder(builder)` | Message broker factory | Real RabbitMQ with in-memory mock |
| `timeutils.SetCurrentTimeProvider(fn)` | Time provider | Real time with controllable clock |

### Inner-Layer Factories (Singletons WITHOUT Setters)

These are wiring factories (persistence, webservice, publisher, service, handler). They are **singletons** (`sync.Once` + cached instance) but have **NO setters and NO reset functions**. They pull from external factories, so they automatically pick up whatever mocks were injected via the external factory setters.

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// Inner-layer factory — singleton, no setter, no reset
var userPersistenceOnce sync.Once
var userPersistenceInstance adapter.UserPersistence

func GetUserPersistence() adapter.UserPersistence {
    if userPersistenceInstance == nil {
        userPersistenceOnce.Do(func() {
            userPersistenceInstance = persistence.NewUserPersistence(
                database.GetMainDb(),  // pulls the mock that was injected once
            )
        })
    }
    return userPersistenceInstance
}
```

**No factory EVER has a `Reset*()` function.** Not external factories, not inner-layer factories. Between BDD scenarios, only mock STATE is reset (messages, files, keys, data) — not the mock instances, not the factories.

---

## Mock Injection Sequence in InitializeScenario

The `InitializeScenario` function in `test/integration/steps/step_definitions_test.go` follows this exact sequence:

### 1. Create mock instances

```go
test := &testUtils{
	uri:          "http://localhost:" + properties.Properties().Application.ServerPort,
	client:       &http.Client{},
	time:         mock.NewTime(),
	api:          mock.NewApiServer(),
	redis:        mock.NewRedis(),
	broker:       mock.NewBrokerClient[any](),
	cloudStorage: mock.NewCloudStorageClient(),
	maestroDb: mock.NewDb("maestro", map[string]any{
		"campaigns":             &model.Campaign{},
		"moments":               &model.Moment{},
		"creator_group_invites": &model.Invite{},
		// ... all tables with their model structs
	}),
	catalystDb: mock.NewDb("catalyst", map[string]any{
		"instagram_business_credentials": &model.Credential{},
	}),
}
```

### 2. Inject mocks into factories ONE TIME (before any scenario runs)

```go
timeutils.GetTimeConfig().SetCurrentTimeProvider(test.time.Now)
database.SetRedis(test.redis)
database.SetMaestroDb(test.maestroDb.DbConn)
database.SetCatalystDb(test.catalystDb.DbConn)
storage.SetCloudStorageInstance(test.cloudStorage)
broker.SetBrokerBuilder(test.broker.Builder)
```

This happens **ONCE per test suite run** and is NEVER repeated. The `Set{Name}()` functions mark the `sync.Once` as done and set the mock as the instance. These mock instances live forever — between scenarios, only their internal state is cleared (not the instance itself). The setters are NOT called again in the `before()` hook.

### 3. Register Before/After hooks

```go
ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	test.before()
	test.startApp()
	return ctx, nil
})
```

### 4. Register all step definitions

```go
ctx.Given(`^the tables are empty$`, test.theTablesAreEmpty)
ctx.When(`^I call "([^"]*)" "([^"]*)" with the following payload$`, test.iCallWithTheFollowingPayload)
// ... all other steps
```

---

## Before/After Hook Patterns

The `before()` function runs before EVERY scenario to ensure clean mock state. It does NOT re-inject mocks (that happened once in `InitializeScenario`) and does NOT call any factory reset functions (they don't exist).

```go
func (t *testUtils) before() {
	t.headers = make(map[string]string)
	t.receiveCount = 0

	// Clear mock STATE only (messages, files, keys, data — NOT the mock instances)
	mock.ClearRedis(t.redis)
	t.broker.Reset()
	t.cloudStorage.Reset()

	// Clear database tables (truncate, not replace the DB connection)
	_ = t.maestroDb.ClearDB()
	_ = t.catalystDb.ClearDB()

	// Reset API mock server (clear responses/request history)
	t.api.Start()
}
```

The sequence matters:

1. **Reset test state** (`headers`, `receiveCount`) — clears per-scenario request state
2. **Clear mock state** — flushes accumulated messages, files, keys from in-memory mocks via `Reset()` methods
3. **Clear databases** — truncates all tables in SQLite mock databases via `ClearDB()`
4. **Start API mock** — resets the HTTP mock server for a fresh set of responses

**What `before()` does NOT do:**
- Does NOT call factory setters (e.g., `database.SetRedis()`) — injection happened once in `InitializeScenario`
- Does NOT call factory reset functions — factories NEVER have `Reset*()` functions
- Does NOT create new mock instances — mock instances live forever, only their state is cleared

```
┌─────────────────────────────────────────────────────────┐
│ InitializeScenario (runs ONCE per test suite)            │
│   1. Create mock instances (NewRedis, NewDb, etc.)       │
│   2. Inject mocks via factory setters (ONE-TIME)         │
│   3. Register before/after hooks and step definitions    │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│ before() (runs BEFORE EVERY scenario)                    │
│   1. Clear test state (headers, counters)                │
│   2. Reset mock STATE: broker.Reset(), ClearRedis(), etc.│
│   3. Clear DB tables: ClearDB()                          │
│   4. Reset API mock: api.Start()                         │
│                                                          │
│   NEVER: re-inject mocks, call factory resets            │
└─────────────────────────────────────────────────────────┘
```

The `startApp()` function uses `sync.Once` to start the application server only on the first scenario. Subsequent scenarios reuse the running server:

```go
var serverInit sync.Once

func (t *testUtils) startApp() {
	serverInit.Do(func() {
		go func() {
			if !properties.Properties().Application.IsLambda {
				application.Start()
			}
		}()
	})
}
```

The `After` hook is currently empty but is available for teardown logic if needed.

---

## How Adding a New Factory Affects Test Setup

When you add a new dependency factory, follow these steps to integrate it with the test infrastructure:

### Step 1: Create the factory as a singleton (ALL factories are singletons)

In `internal/infra/dependency/{layer}/{name}.go`:

```go
var myNewThingOnce sync.Once
var myNewThingInstance adapter.MyNewThing

func GetMyNewThing() adapter.MyNewThing {
	if myNewThingInstance == nil {
		myNewThingOnce.Do(func() {
			myNewThingInstance = whatever.NewMyNewThing(
				// pull other factories as deps
			)
		})
	}
	return myNewThingInstance
}

// NO Reset function — factories NEVER have Reset*() functions
```

### Step 2: If it is an external infrastructure dependency, add a Setter

Only if the factory wraps an external dependency (database, cache, API client, storage, broker), add a setter for ONE-TIME mock injection:

```go
func SetMyNewThing(instance adapter.MyNewThing) {
	myNewThingOnce.Do(func() {})
	myNewThingInstance = instance
}
```

Then inject the mock ONCE in `InitializeScenario`:

```go
{package}.SetMyNewThing(test.myNewMock)
```

And add the mock's state reset to `before()`:

```go
func (t *testUtils) before() {
	// ... existing resets
	t.myNewMock.Reset()  // clear accumulated state, NOT re-inject
}
```

### Step 3: If it is an inner-layer factory, no setter needed

Inner-layer factories (services, persistence, webservices, publishers, controllers) do NOT need setters. They are singletons that pull from external factories. Since the external factory mocks were injected once and their state is cleared between scenarios, the inner-layer singletons work correctly without any reset or re-injection.

---

## Relationship Between Dependency Factories and Test Mocks

```
Test Layer (test/integration/)        Infrastructure Layer (infra/dependency/)    Integration/Application Layer
──────────────────────────────        ────────────────────────────────────────    ─────────────────────────────
mock.NewRedis()                  →    database.SetRedis(mock)  [ONE-TIME]    →   (auto) persistence uses mock DB
mock.NewDb("maestro")            →    database.SetMaestroDb(mock) [ONE-TIME] →   (auto) persistence uses mock DB
mock.NewCloudStorage()           →    storage.SetCloudStorage(mock) [ONE-TIME]→  (auto) post persistence uses mock storage
mock.NewBrokerClient()           →    broker.SetBrokerBuilder() [ONE-TIME]   →   (auto) publisher uses mock broker
mock.NewApiServer()              →    (env vars point to mock URL) [ONE-TIME]→   (auto) webservice calls mock server
mock.NewTime()                   →    timeutils.SetCurrentTimeProvider() [1x]→   (auto) all time calls use mock

InitializeScenario: injects mocks ONCE via setters (mock instances live forever)
before(): resets mock STATE only (ClearRedis, broker.Reset, ClearDB — NOT re-injection)
```

The key insight: you only mock the outermost infrastructure boundaries. All inner layers are real code that runs through the full application path. ALL factories are singletons. Mock instances are injected ONCE and live forever. Between scenarios, only mock state is cleared — not the mock instances, not the factories.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
