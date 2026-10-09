---
name: BDD Testing (Clean Architecture)
description: This skill should be used when the user asks to "create BDD tests", "add integration tests", "set up BDD", "create test infrastructure", or mentions anything about BDD testing, Godog, Gherkin, or the test/ folder.
version: 1.0.0
---

# BDD Testing — Golang Clean Architecture

BDD (Behavior-Driven Development) tests are integration tests that verify the application end-to-end using Gherkin feature files and Godog step definitions. All external dependencies are mocked at the outermost layer (databases, APIs, queues, storage) — never mock inner layers (persistence, services, use cases).

---

## Directory Structure

```
test/
└── integration/
    ├── features/              ← Gherkin .feature files
    │   ├── create_user.feature
    │   ├── process_order.feature
    │   └── ...
    ├── steps/                 ← Step definition files (broken by concern)
    │   ├── step_definitions_test.go   ← Main test runner + setup
    │   ├── given_steps_test.go        ← Given step implementations
    │   ├── when_steps_test.go         ← When step implementations
    │   ├── then_steps_test.go         ← Then step implementations
    │   └── helpers_test.go            ← Shared test utilities
    ├── mock/                  ← Mock implementations
    │   ├── api.go             ← HTTP API mock
    │   ├── db.go              ← SQLite database mock
    │   ├── redis.go           ← Redis mock (miniredis)
    │   ├── sqs.go             ← AWS SQS mock
    │   ├── sns.go             ← AWS SNS mock
    │   ├── s3.go              ← AWS S3 mock
    │   ├── dynamodb.go        ← AWS DynamoDB mock
    │   ├── broker.go          ← Message broker mock
    │   ├── cloud_storage.go   ← Cloud storage mock
    │   ├── time.go            ← Time mock
    │   ├── ses.go             ← AWS SES mock
    │   ├── secrets_manager.go ← Secrets Manager mock
    │   ├── rekognition.go     ← AWS Rekognition mock
    │   └── mongo.go           ← MongoDB mock
    └── resources/             ← Test files (videos, images, CSVs, .env)
```

---

## Core Rules

### Rule 1: Mock at the Outermost Layer Only

NEVER mock inner layers (persistence, services, use cases). Always mock the actual external dependencies:

```
MOCK THESE (outermost):              NEVER MOCK THESE (inner):
├── Database (SQLite instead of MySQL/Postgres)  ├── Persistence files
├── Redis (miniredis)                            ├── Services
├── HTTP APIs (httptest mock server)             ├── Use cases
├── SQS/SNS (in-memory mock)                    ├── Controllers
├── S3 (in-memory mock)                         ├── Publishers
├── DynamoDB (in-memory mock)                   └── Webservices
├── MongoDB (in-memory mock)
├── Time (controllable clock)
└── Message brokers (in-memory mock)
```

This ensures the full application code path is tested — from controller through service through persistence — with only the infrastructure boundaries faked.

### Rule 2: Inject Mocks via Factory Setters (ONE TIME in InitializeScenario)

Use the dependency injection factory setters (see Dependency Injection skill) to inject mocks ONCE before any scenario runs. This happens in `InitializeScenario`, NOT in the `before()` hook:

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// InitializeScenario — called ONCE per test suite
database.SetMainDb(test.mainDb.DbConn)
database.SetRedis(test.redis)
storage.SetCloudStorage(test.cloudStorage)
```

Mock instances live forever. Between scenarios, only mock STATE is cleared (via `Reset()`, `ClearDB()`, `ClearRedis()`) — never re-inject mocks via setters.

### Rule 2b: Between Scenarios, Only Mock State Is Cleared

The `before()` hook resets accumulated mock data, NOT mock instances or factories:

```go
func (t *testUtils) before() {
    mock.ClearRedis(t.redis)       // clear keys — NOT re-inject
    t.broker.Reset()                // clear messages — NOT re-inject
    t.cloudStorage.Reset()          // clear files — NOT re-inject
    t.mainDb.ClearDB()              // truncate tables — NOT re-inject
}
```

### Rule 3: Step Definitions Must Be Reusable

Step definitions MUST be generic and reusable across ALL feature files. Never create steps that only serve one scenario.

See: **Step Definitions skill** for details.

### Rule 4: Feature Files Follow Gherkin Conventions

See: **Feature Files skill** for patterns and conventions.

### Rule 5: Mocks Must Be Inspectable

All mocks must provide methods for test assertions — counting messages, retrieving sent data, checking call history, etc.

See: **Mock skills** for each mock type.

---

## Test Runner Setup

```go
// File: test/integration/steps/step_definitions_test.go
package steps

import (
    "flag"
    "testing"
    "github.com/cucumber/godog"
)

var tags string

func init() {
    flag.StringVar(&tags, "scenarios", "", "tags to run")
}

func TestFeatures(t *testing.T) {
    flag.Parse()

    suite := godog.TestSuite{
        ScenarioInitializer: func(s *godog.ScenarioContext) {
            InitializeScenario(s)
        },
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"../features"},
            Tags:     tags,
            Strict:   true,
            TestingT: t,
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}
```

### Running Tests

```bash
# Run all tests
go test ./test/integration/steps/ -v

# Run specific tag
go test ./test/integration/steps/ -v -scenarios="@create_user"

# Run specific scenario
go test ./test/integration/steps/ -v -scenarios="@specific_test"
```

---

## Related Skills

| Skill | Purpose |
|---|---|
| **Step Definitions** | How to write reusable step definitions |
| **Feature Files** | How to write Gherkin feature files |
| **Mock - Generic** | General mock creation guidelines |
| **Mock - Database** | SQLite in-memory mock |
| **Mock - Redis** | Miniredis mock |
| **Mock - API** | HTTP API mock server |
| **Mock - SQS** | AWS SQS mock |
| **Mock - SNS** | AWS SNS mock |
| **Mock - S3** | AWS S3 mock |
| **Mock - Broker** | Message broker mock |
| **Mock - Time** | Controllable time mock |
| **Mock - DynamoDB** | AWS DynamoDB mock |
| **Mock - Cloud Storage** | Cloud storage abstraction mock |
| **Mock - MongoDB** | MongoDB mock |

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
