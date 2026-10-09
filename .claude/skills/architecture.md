---
name: Architecture Overview (Golang Clean Architecture + DDD)
description: This skill should be used when the user asks about "project structure", "architecture", "where should this go", "layer organization", "folder structure", "how to organize", "create a new project", "scaffold a project", or needs guidance on which layer or folder a piece of code belongs in.
version: 1.0.0
---

# Architecture Overview — Golang Clean Architecture + DDD

This architecture is a custom adaptation of Clean Architecture combined with Domain-Driven Design. It uses a strict 4-layer structure where dependencies flow inward only. Every component has a defined place, and every cross-layer communication goes through adapter interfaces.

---

## The 4 Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    INFRASTRUCTURE (infra)                     │
│  Frameworks, drivers, config, dependency factories            │
│  Knows: everything (outermost layer)                         │
├─────────────────────────────────────────────────────────────┤
│                    INTEGRATION                                │
│  Entrypoints, persistence, webservices, publishers            │
│  Knows: application, domain                                  │
├─────────────────────────────────────────────────────────────┤
│                    APPLICATION                                │
│  Use cases, services, adapters, factories                     │
│  Knows: domain only                                          │
├─────────────────────────────────────────────────────────────┤
│                    DOMAIN                                      │
│  Entities, enums, value objects                               │
│  Knows: nothing (innermost layer)                            │
└─────────────────────────────────────────────────────────────┘

Dependency Rule: Dependencies ONLY flow inward (↓)
Domain ← Application ← Integration ← Infrastructure
```

---

## Complete Project Structure

```
project-root/
├── cmd/
│   ├── server/
│   │   └── main.go                          ← Production entry point
│   └── local/
│       └── main.go                          ← Local development entry point
│
├── internal/
│   ├── domain/                              ← LAYER 1: Domain (innermost)
│   │   ├── entity/                          ← Domain entities with business logic
│   │   │   ├── user.go                         → Skill: entities
│   │   │   ├── order.go
│   │   │   └── credential.go
│   │   └── enums/                           ← Domain enums (type + values)
│   │       ├── order_status.go                 → Skill: enums
│   │       ├── orderstatus/
│   │       │   └── value.go
│   │       ├── channel.go
│   │       └── channel/
│   │           └── value.go
│   │
│   ├── application/                         ← LAYER 2: Application
│   │   ├── usecase/                         ← Use case interfaces + pure functions
│   │   │   ├── find_user.go                    → Skill: usecases
│   │   │   ├── save_order.go
│   │   │   └── calculate_score.go
│   │   ├── adapter/                         ← Application adapter interfaces
│   │   │   ├── process_order.go                → Skill: adapters
│   │   │   ├── post_validator.go
│   │   │   ├── validator_factory.go
│   │   │   └── downloader_factory.go
│   │   ├── service/                         ← Service implementations
│   │   │   ├── process_order.go                → Skill: services
│   │   │   ├── reels_validator.go
│   │   │   └── story_validator.go
│   │   └── factory/                         ← Factory implementations
│   │       ├── validator_factory.go
│   │       └── downloader_factory.go
│   │
│   ├── integration/                         ← LAYER 3: Integration
│   │   ├── adapter/                         ← Integration adapter interfaces
│   │   │   ├── user_persistence.go             → Skill: adapters
│   │   │   ├── payment_webservice.go
│   │   │   ├── order_publisher.go
│   │   │   ├── handler.go
│   │   │   ├── error_handler.go
│   │   │   └── validator.go
│   │   ├── entrypoint/                      ← Entry points (controllers, consumers)
│   │   │   ├── controller/                     → Skill: controllers
│   │   │   │   ├── handler.go
│   │   │   │   └── error_handler.go            → Skill: error-handler
│   │   │   ├── dto/                            → Skill: dtos
│   │   │   │   ├── event.go
│   │   │   │   └── create_user.go
│   │   │   ├── errors/
│   │   │   │   └── validation_error.go
│   │   │   ├── validator/                      → Skill: rest-router (validators section)
│   │   │   │   └── custom.go
│   │   │   ├── middleware/                     → Skill: rest-router (middleware section)
│   │   │   │   ├── cors.go
│   │   │   │   ├── trace.go
│   │   │   │   ├── authorization.go
│   │   │   │   ├── payload_validator.go
│   │   │   │   ├── query_validator.go
│   │   │   │   ├── path_validator.go
│   │   │   │   └── form_validator.go
│   │   │   └── enums/                          → Skill: enums
│   │   │       ├── var_type.go
│   │   │       └── vartype/
│   │   │           └── value.go
│   │   ├── persistence/                     ← Database access
│   │   │   ├── model/                          → Skill: models
│   │   │   │   ├── user.go
│   │   │   │   └── order.go
│   │   │   ├── user.go                         → Skill: persistence
│   │   │   └── order.go
│   │   ├── webservice/                      ← External API clients
│   │   │   ├── dto/                            → Skill: dtos
│   │   │   │   └── payment_response.go
│   │   │   ├── enums/                          → Skill: enums
│   │   │   │   └── connection_type.go
│   │   │   └── payment.go                     → Skill: webservices
│   │   └── publisher/                       ← Message publishers
│   │       ├── dto/                            → Skill: dtos
│   │       │   └── order_event.go
│   │       └── order_processing.go            → Skill: publishers
│   │
│   └── infra/                               ← LAYER 4: Infrastructure (outermost)
│       ├── application.go                   ← App startup
│       ├── properties/                         → Skill: properties
│       │   └── properties.go
│       ├── dependency/                         → Skill: dependency-injection
│       │   ├── database/
│       │   │   ├── main_db.go
│       │   │   └── redis.go
│       │   ├── integration/
│       │   │   ├── user_persistence.go
│       │   │   ├── payment_webservice.go
│       │   │   └── handler.go
│       │   ├── application/
│       │   │   └── process_order.go
│       │   ├── http/
│       │   │   └── client.go
│       │   └── storage/
│       │       └── cloud_storage.go
│       ├── server/                             → Skill: rest-router
│       │   ├── router/
│       │   │   ├── router.go
│       │   │   └── user_routes.go
│       │   ├── context/
│       │   │   └── fiber.go
│       │   ├── middleware/
│       │   │   └── fiber_middleware.go
│       │   └── adapter/
│       │       ├── payload_validator.go
│       │       ├── query_validator.go
│       │       └── path_validator.go
│       └── adapter/                         ← Infra adapters (only when lib has no interface)
│           └── cloud_storage.go                → Skill: adapters
│
├── pkg/                                     ← Standalone reusable packages
│   ├── errs/                                   → Skill: pkg-errs
│   ├── logger/                                 → Skill: pkg-logger
│   ├── http/                                   → Skill: pkg-http
│   ├── redis/                                  → Skill: pkg-redis
│   ├── timeutils/                              → Skill: pkg-timeutils
│   ├── aws/                                    → Skill: pkg-aws
│   ├── dbutils/                                → Skill: pkg-dbutils
│   ├── encryption/                             → Skill: pkg-encryption
│   ├── geoutils/                               → Skill: pkg-geoutils
│   ├── tracer/                                 → Skill: pkg-tracer
│   ├── collection/                             → Skill: pkg-collection
│   └── utils/                                  → Skill: pkg-utils
│
├── test/                                    ← BDD integration tests
│   └── integration/                            → Skill: bdd-testing
│       ├── features/                           → Skill: bdd-feature-files
│       │   ├── create_user.feature
│       │   └── process_order.feature
│       ├── steps/                              → Skill: bdd-step-definitions
│       │   ├── step_definitions_test.go
│       │   ├── given_steps_test.go
│       │   ├── when_steps_test.go
│       │   └── then_steps_test.go
│       ├── mock/                               → Skill: mock-generic + individual mocks
│       │   ├── db.go
│       │   ├── redis.go
│       │   ├── api.go
│       │   ├── sqs.go
│       │   ├── sns.go
│       │   ├── s3.go
│       │   ├── broker.go
│       │   ├── time.go
│       │   └── cloud_storage.go
│       └── resources/
│
├── go.mod
├── go.sum
└── CLAUDE.md
```

---

## Dependency Rule — What Can Import What

```
Domain:
  ├── Can import: Go stdlib, pkg/ utilities (timeutils, etc.)
  └── Cannot import: application, integration, infra, external libs (except via pkg/)

Application:
  ├── Can import: domain (entities, enums), Go stdlib, pkg/
  └── Cannot import: integration, infra

Integration:
  ├── Can import: domain, application (use cases, adapters), Go stdlib, pkg/, external libs
  └── Cannot import: infra

Infrastructure:
  ├── Can import: everything (outermost layer)
  └── This is where frameworks live (Fiber, GORM, AWS SDK, etc.)

pkg/:
  ├── Can import: Go stdlib, external libs
  └── Cannot import: internal/ (anything), other pkg/ packages
```

---

## Cross-Layer Communication — Always Through Adapters

Every cross-layer call goes through an adapter interface. The implementing layer owns the adapter.

```
Entrypoint (controller)
    │
    ↓ calls via application adapter (adapter.ProcessOrderService)
    │
Service (service.processOrder)
    │
    ↓ calls via use case interface (usecase.FindUser)
    │
    ↓ implemented by integration adapter (adapter.UserPersistence)
    │
Persistence (persistence.userPersistence)
    │
    ↓ uses model (model.User) → queries DB → converts via ToEntity()
    │
Returns domain entity (entity.User)
```

→ Skill: **adapters**

---

## "Where Does This Go?" Decision Guide

### I need to create a new...

| What | Where | Skill |
|---|---|---|
| Business entity with logic | `internal/domain/entity/` | entities |
| Enum type + values | `internal/domain/enums/` (or innermost layer that uses it) | enums |
| Use case (I/O needed) | `internal/application/usecase/` (interface) | usecases |
| Use case (pure logic) | `internal/application/usecase/` (function) | usecases |
| Service (orchestrates use cases) | `internal/application/service/` | services |
| Application adapter | `internal/application/adapter/` | adapters |
| REST controller | `internal/integration/entrypoint/controller/` | controllers |
| Event/message controller | `internal/integration/entrypoint/controller/` | controllers |
| Error handler | `internal/integration/entrypoint/controller/` | error-handler |
| Request/response DTO | `internal/integration/entrypoint/dto/` | dtos |
| External API DTO | `internal/integration/webservice/dto/` | dtos |
| Publisher event DTO | `internal/integration/publisher/dto/` | dtos |
| Database model | `internal/integration/persistence/model/` | models |
| Database access | `internal/integration/persistence/` | persistence |
| External API client | `internal/integration/webservice/` | webservices |
| Message publisher | `internal/integration/publisher/` | publishers |
| Integration adapter | `internal/integration/adapter/` | adapters |
| Custom middleware | `internal/integration/entrypoint/middleware/` | rest-router |
| Input validator | `internal/integration/entrypoint/validator/` | rest-router |
| HTTP router/routes | `internal/infra/server/router/` | rest-router |
| Framework context wrapper | `internal/infra/server/context/` | rest-router |
| Dependency factory | `internal/infra/dependency/{type}/` | dependency-injection |
| Environment config | `internal/infra/properties/` | properties |
| App startup | `internal/infra/application.go` | dependency-injection |
| Reusable library | `pkg/{type}/` | pkg-generic |
| BDD feature file | `test/integration/features/` | bdd-feature-files |
| Step definitions | `test/integration/steps/` | bdd-step-definitions |
| Test mock | `test/integration/mock/` | mock-generic |

---

## Data Flow Through the Architecture

### REST API Request

```
HTTP Request
    ↓
[infra] Fiber Router → validates (scope, path, query, payload) via middleware chain
    ↓
[infra] FiberMiddleware converts fiber.Handler → adapter.Middleware
    ↓
[integration] Controller receives validated data via ctx.Locals()
    ↓ converts DTO → entity via dto.ToEntity()
    ↓
[application] Service orchestrates use cases, delegates business logic to entities
    ↓
[application] Use case interface → implemented by integration layer
    ↓
[integration] Persistence queries DB via model → converts via model.ToEntity()
    ↓
[domain] Entity methods execute business logic (IsActive, IsValid, etc.)
    ↓
Response flows back up: entity → DTO → JSON response
```

### Event/Message Processing

```
Message arrives (SQS, RabbitMQ, SNS)
    ↓
[infra] Broker consumer calls controller handler
    ↓
[integration] Controller validates DTO, converts to entity
    ↓
[application] Service orchestrates: fetch data → validate → process → publish
    ↓
[integration] Persistence reads, Publisher writes, Webservice calls external API
    ↓
[domain] Entities own all business decisions
    ↓
Error handler catches any errors at the edge
```

---

## Key Architecture Principles

### 1. Entities Own Business Logic (DDD)
Never put business logic in services or use cases. Entities are rich, not anemic.
→ Skill: **entities**

### 2. Dependencies Flow Inward Only
Domain knows nothing. Infrastructure knows everything. Each layer only knows inner layers.
→ Skill: **adapters**

### 3. Framework Decoupling
Frameworks (Fiber, GORM, AWS SDK) live in infra or are wrapped in pkg/. The integration layer never imports frameworks directly.
→ Skills: **rest-router**, **persistence**, **webservices**, **publishers**

### 4. Adapter Interfaces for Every Cross-Layer Call
No direct imports of concrete implementations across layers. Always through interfaces.
→ Skill: **adapters**

### 5. DTOs/Models Handle Conversion
Domain entities never know about serialization. DTOs and models own the ToEntity()/New{Name}() conversion.
→ Skills: **dtos**, **models**

### 6. Error Codes at Origin
Every error gets a unique code at the exact line where it's thrown. The error handler at the edge formats them consistently.
→ Skill: **error-handler**

### 7. Properties — Not Singleton, Not Injected
Properties re-read env vars every call (runtime swap). Used inline, never in constructors.
→ Skill: **properties**

### 8. Distributed Dependency Injection
No centralized injector. Distributed singleton factories, each pulling from other factories.
→ Skill: **dependency-injection**

### 9. BDD Tests Mock the Outermost Layer
Mock databases (SQLite), APIs (httptest), queues (in-memory) — never inner layers.
→ Skills: **bdd-testing**, **mock-generic**

### 10. Packages Are Self-Contained
Every `pkg/` folder works standalone. Zero internal imports, zero cross-pkg imports.
→ Skill: **pkg-generic**

### 11. Minimal Comments — No Inline Comments
No comments inside function bodies. Go doc comments only when truly necessary, one line max.
→ Skill: **comments**

---

## Logging Conventions by Layer

| Layer | Level | Pattern |
|---|---|---|
| Controller | `logger.Info` | Started/Finished with raw input/output |
| Service | `logger.Info` | Started/Finished with raw input/output |
| Persistence | `logger.Debug` | Started/Finished with raw input/output |
| Webservice | `logger.Debug` | Started/Finished with raw input/output |
| Publisher | `logger.Debug` | Started/Finished with raw input/output |
| Error Handler | `logger.Error` | Started/Finished with error details |

---

## Struct Export Conventions

| Type | Exported? | Reason |
|---|---|---|
| Domain entities | Yes | Used across layers as data carriers |
| DTOs | Yes | Used for serialization across layers |
| Models | Yes | Used for database operations |
| Enums (types + constants) | Yes | Used across layers |
| Services (structs) | No | Only exposed via constructor → adapter interface |
| Controllers (structs) | No | Only exposed via constructor → adapter interface |
| Persistence (structs) | No | Only exposed via constructor → adapter interface |
| Webservices (structs) | No | Only exposed via constructor → adapter interface |
| Publishers (structs) | No | Only exposed via constructor → adapter interface |
| Error handlers (structs) | No | Only exposed via constructor → adapter interface |
| Properties (struct) | No | Only exposed via `Properties()` function |

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
