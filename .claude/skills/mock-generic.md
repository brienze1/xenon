---
name: Mock - Generic Guidelines (BDD Testing)
description: This skill should be used when the user asks to "create a mock", "add a test mock", "create test infrastructure", or mentions anything about creating mocks in test/integration/mock/ or test/mock/.
version: 1.0.0
---

# Mocks — Generic Guidelines

Mocks replace external dependencies in BDD tests. They must be generic, inspectable, thread-safe, and mock at the outermost layer only.

**Location:** `test/integration/mock/` (or `test/mock/`)

---

## Core Rules

### Rule 1: Mock the Outermost Layer
Mock external infrastructure (DB, Redis, HTTP APIs, queues, storage), NEVER inner layers (persistence, services, use cases).

### Rule 2: Implement the Same Interface
Mocks must implement the same interface that the real dependency uses, allowing injection via factory setters.

### Rule 3: Inspectable — Test Assertions
Every mock must expose methods for test verification:
- **Count**: How many times was it called? How many messages sent?
- **Retrieve**: Get the Nth message/call/request
- **Inspect**: What headers/body/params were sent?
- **Validate**: Does a specific field equal a value?

### Rule 4: Every Mock MUST Have a Reset() Method
Every mock MUST have a `Reset()` method (or equivalent like `ClearDB()`, `ClearRedis()`) that clears its internal state — accumulated messages, stored files, cached keys, database rows, request history, etc. The `Reset()` method clears STATE only; it does not destroy or recreate the mock instance.

### Rule 5: Mock Instances Are Injected ONCE and Reused Forever
Mock instances are created once in `InitializeScenario`, injected once via factory setters, and live for the entire test suite. They are NEVER recreated or re-injected between scenarios. Only their internal state is cleared via `Reset()`.

### Rule 6: BDD before() Hooks Reset Mock STATE, Not Mock Instances
The `before()` hook calls `mock.Reset()`, `ClearDB()`, `ClearRedis()` to clear accumulated data. It does NOT call factory setters to re-inject mocks. It does NOT call factory reset functions (they don't exist).

### Rule 7: Thread-Safe
Use `sync.RWMutex` for mocks accessed concurrently (HTTP API mock, message queues).

### Rule 8: Singleton Pattern
Mocks use singleton or controlled initialization — created once, injected via factory setters ONE TIME.

### Rule 9: No Internal Imports (When Possible)
Mocks should ideally implement library interfaces (e.g., `redis.UniversalClient`, `*gorm.DB`) without importing internal code. When they MUST implement internal adapters, keep the coupling minimal.

---

## Mock Constructor Pattern

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// Singleton with factory function
var instance *MockType
var once sync.Once

func NewMockType() *MockType {
    once.Do(func() {
        instance = &MockType{
            // initialize internal state
        }
    })
    return instance
}

// Reset for test isolation
func (m *MockType) Reset() {
    m.mu.Lock()
    defer m.mu.Unlock()
    // clear all stored state
}
```

---

## Available Mock Skills

| Mock | Purpose | Skill |
|---|---|---|
| Database (SQLite) | Replaces MySQL/PostgreSQL | Mock - Database |
| Redis (miniredis) | Replaces Redis | Mock - Redis |
| HTTP API | Mocks external REST APIs | Mock - API |
| SQS | Mocks AWS SQS | Mock - SQS |
| SNS | Mocks AWS SNS | Mock - SNS |
| S3 | Mocks AWS S3 | Mock - S3 |
| Message Broker | Mocks RabbitMQ/generic broker | Mock - Broker |
| Time | Controls time.Now() | Mock - Time |
| DynamoDB | Mocks AWS DynamoDB | Mock - DynamoDB |
| Cloud Storage | Abstracts S3/GCS | Mock - Cloud Storage |
| MongoDB | Mocks MongoDB | Mock - MongoDB |
| Secrets Manager | Mocks AWS Secrets Manager | Mock - Secrets Manager |
| SES | Mocks AWS SES | Mock - SES |
| Rekognition | Mocks AWS Rekognition | Mock - Rekognition |

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
