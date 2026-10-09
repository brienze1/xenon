---
name: Naming Conventions (Cross-Layer Reference)
description: This skill should be used when the user asks about "naming conventions", "naming rules", "how to name", "naming patterns", "what should I call", "struct naming", "file naming", "package naming", or when reviewing code for naming consistency across layers.
version: 1.0.0
---

# Naming Conventions — Golang Clean Architecture

Consolidated cross-layer naming reference. All naming rules extracted from individual layer skills into one lookup table.

---

## Struct Naming

| Layer | Element | Exported? | Pattern | Example | Source Skill |
|---|---|---|---|---|---|
| Domain | Entity | YES | `{Name}` | `MediaInfo`, `Campaign`, `User` | entities.md |
| Domain | Enum type | YES | `{Name}` | `OrderStatus`, `Channel` | enums.md |
| Application | Use case interface | YES | `{Action}{Subject}` | `FindUser`, `SaveOrder` | usecases.md |
| Application | Service struct | NO | `{action}{Subject}` | `processOrder`, `filterFallbackPosts` | services.md |
| Application | Service adapter | YES | `{Action}{Subject}Service` | `ProcessOrderService`, `FilterAndProcessFallbackPostsService` | services.md |
| Integration | Adapter interface | YES | `{Subject}{Type}` | `UserPersistence`, `PaymentWebservice`, `OrderPublisher` | adapters.md |
| Integration | Controller struct | NO | `{name}` | `handler`, `errorHandler` | controllers.md |
| Integration | Persistence struct | NO | `{name}Persistence` | `userPersistence`, `postPersistence` | persistence.md |
| Integration | Webservice struct | NO | `{name}Webservice` | `instagramWebservice`, `paymentWebservice` | webservices.md |
| Integration | Publisher struct | NO | `{name}Publisher` | `videoProcessingPublisher`, `orderProcessingPublisher` | publishers.md |
| Integration | Model struct | YES | `{Name}` (no "Model" suffix) | `User`, `Campaign`, `Credential` | models.md |
| Integration | DTO struct | YES | `{Name}` (no "DTO" suffix) | `Event`, `PostMetadata`, `MediaInfo` | dtos.md |
| Integration | Error handler struct | NO | `errorHandler` | `errorHandler` | error-handler.md |
| Infrastructure | Properties struct | NO | `properties` | `properties` | properties.md |

---

## Constructor Naming

| Layer | Element | Pattern | Return Type | Example | Source Skill |
|---|---|---|---|---|---|
| Application | Service | `New{Name}(deps...)` | Adapter interface | `NewFilterAndProcessFallbackPosts(...) adapter.FilterAndProcessFallbackPostsService` | services.md |
| Integration | Persistence | `New{Name}Persistence(deps...)` | Adapter interface | `NewUserPersistence(db, cache) adapter.UserPersistence` | persistence.md |
| Integration | Webservice | `New{Name}(deps...)` | Adapter interface | `NewInstagramWebservice(client) adapter.InstagramWebservice` | webservices.md |
| Integration | Publisher | `New{Name}(deps...)` | Adapter interface | `NewVideoProcessing(broker) adapter.VideoProcessingPublisher` | publishers.md |
| Integration | Controller | `NewHandler(deps...)` | Adapter interface | `NewHandler(errorHandler, service) adapter.Handler` | controllers.md |
| Integration | Error handler | `NewErrorHandler()` | Adapter interface | `NewErrorHandler() adapter.ErrorHandler` | error-handler.md |
| Integration | Model | No constructor | N/A | Models are structs with tags, no constructor | models.md |

---

## Adapter Interface Naming

| Owning Layer | Pattern | Example | Source Skill |
|---|---|---|---|
| Application | `{Action}{Subject}Service` | `ProcessOrderService`, `FilterAndProcessFallbackPostsService` | services.md |
| Integration | `{Subject}{Type}` | `UserPersistence`, `PaymentWebservice`, `VideoProcessingPublisher` | adapters.md |
| Integration | Embeds use case interfaces | `type UserPersistence interface { usecase.FindUser; usecase.SaveUser }` | adapters.md |
| Integration | Handler | `Handler` | controllers.md |
| Integration | Error handler | `ErrorHandler` | error-handler.md |

---

## Factory Naming (Dependency Injection)

| Element | Pattern | Example | Source Skill |
|---|---|---|---|
| Package | Type-based sub-folder | `database`, `integration`, `application`, `http`, `storage`, `broker` | dependency-injection.md |
| Getter | `Get{DependencyName}()` | `database.GetRedis()`, `integration.GetUserPersistence()` | dependency-injection.md |
| Setter | `Set{DependencyName}(instance)` | `database.SetRedis(mockRedis)` | dependency-injection.md |
| Reset | `Reset{DependencyName}()` | `integration.ResetHandler()` | dependency-injection.md |
| Singleton vars | `{name}Once` + `{name}Instance` | `handlerOnce sync.Once`, `handlerInstance adapter.Handler` | dependency-injection.md |

---

## File Naming

| Layer | Pattern | Example | Source Skill |
|---|---|---|---|
| Domain entities | `{entity_name}.go` | `media_info.go`, `campaign.go` | entities.md |
| Domain enums (type) | `{enum_name}.go` | `order_status.go`, `channel.go` | enums.md |
| Domain enums (values) | `{enumname}/value.go` | `orderstatus/value.go` | enums.md |
| Use cases | `{action}_{subject}.go` | `find_user.go`, `save_order.go` | usecases.md |
| Services | `{action}_{subject}.go` | `process_order.go`, `filter_fallback_posts.go` | services.md |
| Service adapters | `{action}_{subject}.go` | `process_order.go` (in `adapter/`) | services.md |
| Persistence | `{subject}.go` | `user.go`, `credential.go` | persistence.md |
| Models | `{subject}.go` | `user.go`, `campaign.go` | models.md |
| Webservices | `{subject}.go` | `instagram.go`, `payment.go` | webservices.md |
| Publishers | `{subject}.go` | `video_processing.go` | publishers.md |
| DTOs | `{subject}.go` | `event.go`, `post_metadata.go` | dtos.md |
| Controllers | `handler.go`, `error_handler.go` | `handler.go` | controllers.md |
| Factories | `{subject}.go` | `handler.go`, `redis.go`, `client.go` | dependency-injection.md |
| Properties | `properties.go` | `properties.go` | properties.md |
| Step definitions | `{given/when/then}_{concern}_steps_test.go` | `given_db_steps_test.go` | bdd-step-definitions.md |

---

## Package Naming

| Layer | Package | Location | Source Skill |
|---|---|---|---|
| Domain | `entity` | `internal/domain/entity/` | entities.md |
| Domain | `enums` | `internal/domain/enums/` | enums.md |
| Domain | `{enumname}` | `internal/domain/enums/{enumname}/` | enums.md |
| Application | `usecase` | `internal/application/usecase/` | usecases.md |
| Application | `service` | `internal/application/service/` | services.md |
| Application | `adapter` | `internal/application/adapter/` | adapters.md |
| Integration | `adapter` | `internal/integration/adapter/` | adapters.md |
| Integration | `controller` | `internal/integration/entrypoint/controller/` | controllers.md |
| Integration | `dto` | `internal/integration/entrypoint/dto/` | dtos.md |
| Integration | `dto` | `internal/integration/webservice/dto/` | dtos.md |
| Integration | `dto` | `internal/integration/publisher/dto/` | dtos.md |
| Integration | `model` | `internal/integration/persistence/model/` | models.md |
| Integration | `persistence` | `internal/integration/persistence/` | persistence.md |
| Integration | `webservice` | `internal/integration/webservice/` | webservices.md |
| Integration | `publisher` | `internal/integration/publisher/` | publishers.md |
| Infrastructure | `properties` | `internal/infra/properties/` | properties.md |
| Infrastructure | `database` | `internal/infra/dependency/database/` | dependency-injection.md |
| Infrastructure | `integration` | `internal/infra/dependency/integration/` | dependency-injection.md |
| Infrastructure | `application` | `internal/infra/dependency/application/` | dependency-injection.md |
| Test | `steps` | `test/integration/steps/` | bdd-step-definitions.md |
| Test | `mock` | `test/mock/` | bdd-testing.md |

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
