---
name: Webservices (Clean Architecture)
description: This skill should be used when the user asks to "create a webservice", "add an external API client", "integrate with an external API", "create an HTTP client wrapper", "add a REST client", or mentions anything about webservices in the internal/integration/webservice/ package.
version: 1.0.0
---

# Webservices — Golang Clean Architecture

Webservices are the integration layer's external API clients. They talk to external REST APIs, handle HTTP request/response logic, convert DTOs to/from entities, and return appropriate errors. They implement use case interfaces via integration adapters — same pattern as persistence files.

---

## Core Rules

### Rule 1: Location

Webservices MUST live in `internal/integration/webservice/` and declare `package webservice`.

```
internal/
└── integration/
    └── webservice/
        ├── dto/                    ← External API contract DTOs
        │   └── media_info.go
        ├── enums/                  ← Webservice-specific enums (if needed)
        │   └── connection_type.go
        ├── payment.go              ← webservice implementation
        └── notification.go
```

---

### Rule 2: HTTP Client Must Be Injected (Never Created)

Webservices MUST receive the HTTP client (or an interface implementing the same methods) via the constructor. Never create `http.Client` inside the webservice.

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// GOOD — HTTP client injected
type paymentWebservice struct {
    client httpClient.Client
}

func NewPaymentWebservice(client httpClient.Client) adapter.PaymentWebservice {
    return &paymentWebservice{client: client}
}
```

```go
// BAD — creating HTTP client inside webservice
type paymentWebservice struct {
    client *http.Client
}

func NewPaymentWebservice() adapter.PaymentWebservice {
    return &paymentWebservice{
        client: &http.Client{Timeout: 10 * time.Second},  // WRONG — create outside
    }
}
```

The HTTP client interface typically looks like:
```go
type Client interface {
    Do(ctx context.Context, req *http.Request) (*http.Response, error)
}
```

**URLs and environment variables** MUST be read from `properties` directly inside the webservice methods — never passed via the constructor. The constructor should only receive the HTTP client and optionally a cache client.

```go
// GOOD — URLs from properties, only client + cache in constructor
func NewPaymentWebservice(client httpClient.Client, cache *redis.Client) adapter.PaymentWebservice {
    return &paymentWebservice{client: client, cache: cache}
}

func (w *paymentWebservice) Process(ctx context.Context, ...) error {
    url := properties.Properties().Services.PaymentApiUrl  // URL from properties
    // ...
}

// BAD — URL passed in constructor
func NewPaymentWebservice(client httpClient.Client, baseURL string) adapter.PaymentWebservice {  // WRONG
    return &paymentWebservice{client: client, baseURL: baseURL}
}
```

---

### Rule 3: Naming Conventions

| Element | Pattern | Example |
|---|---|---|
| Struct | Unexported, `{name}Webservice` | `paymentWebservice`, `notificationWebservice` |
| Constructor | `New{Name}Webservice(...)` | `NewPaymentWebservice(...)` |
| Constructor return type | Integration adapter | `adapter.PaymentWebservice` |
| Package | `webservice` | `package webservice` |

The struct is **always unexported** — exposed only via the constructor returning the adapter interface.

---

### Rule 4: Implements Integration Adapters (Which Aggregate Use Cases)

Same pattern as persistence — webservices implement integration adapters that embed use case interfaces.

```go
// Adapter: internal/integration/adapter/payment_webservice.go
type PaymentWebservice interface {
    usecase.ProcessPayment
}

// Implementation: internal/integration/webservice/payment.go
type paymentWebservice struct {
    client httpClient.Client
}
```

---

### Rule 5: No Business Logic — Contract Logic Only

Webservices MUST NOT contain business logic. However, they CAN contain logic specific to the external service's contract:

**Allowed (contract logic):**
- Checking HTTP status codes to determine success/failure
- Parsing response bodies to detect API-specific errors
- Building request URLs, headers, and query parameters
- Determining file extensions from response content types
- Choosing the correct API endpoint based on a contract parameter

**Forbidden (business logic):**
- Deciding whether to process the response based on domain rules
- Filtering results based on domain criteria
- Validating domain invariants

```go
// GOOD — contract logic: checking external API response status
if response.StatusCode >= 400 && response.StatusCode < 500 {
    return errs.ClientError("Payment rejected", "PAY-00400", response.StatusCode, response.Body)
}
if response.StatusCode >= 500 {
    return errs.ServerError("Payment service unavailable", "PAY-01500", response.StatusCode)
}

// GOOD — contract logic: choosing endpoint based on connection type
switch connectionType {
case "facebook":
    baseURL = facebookApiUrl
case "instagram":
    baseURL = instagramApiUrl
default:
    return errs.ServerError("Invalid connection type", "WS-03500", 500)
}

// BAD — business logic in webservice
if user.IsActive() && order.Total > 100 {  // WRONG — belongs on entity/service
    // ...
}
```

---

### Rule 6: Logger — Debug Level (Started/Finished)

Same as persistence — every method logs `Started` and `Finished` at **Debug** level with raw inputs/outputs.

```go
func (w *paymentWebservice) Process(ctx context.Context, orderId string, amount float64) error {
    logger.Debug(ctx, "Started", map[string]any{"orderId": orderId, "amount": amount})

    // ... HTTP request ...

    logger.Debug(ctx, "Finished", map[string]any{"orderId": orderId, "status": response.StatusCode})
    return nil
}
```

---

### Rule 7: DTOs Handle Entity Conversion (Always Via Constructor)

Webservice DTOs live in `internal/integration/webservice/dto/` and handle conversion between external API contracts and domain entities. DTOs MUST always be created via constructors — never inline struct creation.

**Inbound** (API response → entity): decode into DTO, then call `ToEntity()`:
```go
var mediaInfoDTO dto.MediaInfo
if err = json.NewDecoder(response.Body).Decode(&mediaInfoDTO); err != nil {
    return errs.InternalServerError(err, "WS-03500", "error parsing response")
}

mediaInfo := mediaInfoDTO.ToEntity()
```

**Outbound** (entity → API request): create DTO via constructor, then marshal:
```go
// GOOD — constructor encapsulates all conversion
requestDTO := dto.NewPaymentRequest(order)
body, _ := json.Marshal(requestDTO)
req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))

// BAD — inline struct creation (FORBIDDEN)
requestDTO := dto.PaymentRequest{
    Amount: order.Total,                                    // WRONG — belongs in constructor
    Currency: strings.ToUpper(order.Currency.String()),     // WRONG — belongs in constructor
}
```

---

### Rule 8: Error Codes at Origin

Webservices create custom errors with unique codes at the exact line where the error occurs. These errors propagate up to the error handler.

```go
// Each error has a unique code — traceable to this exact line
return errs.InternalServerError(err, "PAY-00500", "Failed to create payment request")
return errs.InternalServerError(err, "PAY-01500", "Failed to call payment API")
return errs.ClientError("Payment rejected", "PAY-00400", response.StatusCode, response.Body)
return errs.ServerError("Payment service unavailable", "PAY-02500", response.StatusCode)
return errs.BadRequestError("Invalid payment method", "PAY-03400")
return errs.InternalServerError(err, "PAY-04500", "Error parsing payment response")
```

---

### Rule 9: Caching When Possible

Webservices can cache responses when the external API data is suitable for caching. Same pattern as persistence caching.

```go
type configWebservice struct {
    client httpClient.Client
    cache  *redis.Client
}

func NewConfigWebservice(client httpClient.Client, cache *redis.Client) adapter.ConfigWebservice {
    return &configWebservice{client: client, cache: cache}
}

func (w *configWebservice) GetSettings(ctx context.Context, appId string) (*entity.Settings, error) {
    logger.Debug(ctx, "Started", map[string]any{"appId": appId})

    // Try cache
    cached, err := w.cache.Get(ctx, fmt.Sprintf("settings:%s", appId)).Result()
    if err == nil {
        var settingsDTO dto.Settings
        _ = json.Unmarshal([]byte(cached), &settingsDTO)
        settings := settingsDTO.ToEntity()
        logger.Debug(ctx, "Finished (cache hit)", map[string]any{"appId": appId})
        return &settings, nil
    }

    // Fallback to external API call
    // ...

    // Cache for next time
    data, _ := json.Marshal(settingsDTO)
    w.cache.Set(ctx, fmt.Sprintf("settings:%s", appId), data, 5*time.Minute)

    settings := settingsDTO.ToEntity()
    logger.Debug(ctx, "Finished (cache miss)", map[string]any{"appId": appId})
    return &settings, nil
}
```

---

## Code Examples

### Example 1 — External API client (GET request with response parsing)

```go
package webservice

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "myproject/internal/domain/entity"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/webservice/dto"
    "myproject/pkg/errs"
    httpClient "myproject/pkg/http"
    "myproject/pkg/logger"
)

type mediaWebservice struct {
    client httpClient.Client
}

func NewMediaWebservice(client httpClient.Client) adapter.MediaWebservice {
    return &mediaWebservice{client: client}
}

func (w *mediaWebservice) GetInfo(ctx context.Context, postId string, credential entity.Credential) (*entity.MediaInfo, error) {
    logger.Debug(ctx, "Started", map[string]any{"postId": postId})

    req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", baseURL, postId), nil)
    if err != nil {
        return nil, errs.InternalServerError(err, "MED-00500", "Failed to create request")
    }

    q := req.URL.Query()
    q.Set("fields", "id,media_type,media_url")
    q.Set("access_token", *credential.Token)
    req.URL.RawQuery = q.Encode()

    response, err := w.client.Do(ctx, req)
    if err != nil {
        return nil, errs.InternalServerError(err, "MED-01500", "Failed to call media API")
    }
    defer response.Body.Close()

    if response.StatusCode >= 400 && response.StatusCode < 500 {
        return nil, errs.ClientError("Media API client error", "MED-00400", response.StatusCode, response.Body)
    }
    if response.StatusCode >= 500 {
        return nil, errs.ServerError("Media API server error", "MED-02500", response.StatusCode)
    }

    var mediaInfoDTO dto.MediaInfo
    if err = json.NewDecoder(response.Body).Decode(&mediaInfoDTO); err != nil {
        return nil, errs.InternalServerError(err, "MED-03500", "Error parsing media info")
    }

    mediaInfo := mediaInfoDTO.ToEntity()

    logger.Debug(ctx, "Finished", mediaInfo)
    return mediaInfo, nil
}
```

---

### Example 2 — External API client (POST request with request body)

```go
package webservice

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"

    "myproject/internal/domain/entity"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/webservice/dto"
    "myproject/pkg/errs"
    httpClient "myproject/pkg/http"
    "myproject/pkg/logger"
)

type notificationWebservice struct {
    client httpClient.Client
}

func NewNotificationWebservice(client httpClient.Client) adapter.NotificationWebservice {
    return &notificationWebservice{client: client}
}

func (w *notificationWebservice) Send(ctx context.Context, message string, event any) error {
    logger.Debug(ctx, "Started", map[string]any{"message": message})

    requestDTO := dto.NewNotification(message, event)
    body, err := json.Marshal(requestDTO)
    if err != nil {
        return errs.InternalServerError(err, "NOT-00500", "Failed to marshal notification")
    }

    req, err := http.NewRequest(http.MethodPost, notificationURL, bytes.NewReader(body))
    if err != nil {
        return errs.InternalServerError(err, "NOT-01500", "Failed to create notification request")
    }
    req.Header.Set("Content-Type", "application/json")

    response, err := w.client.Do(ctx, req)
    if err != nil {
        return errs.InternalServerError(err, "NOT-02500", "Failed to send notification")
    }
    defer response.Body.Close()

    if response.StatusCode >= 400 {
        return errs.ServerError("Notification service error", "NOT-03500", response.StatusCode)
    }

    logger.Debug(ctx, "Finished", map[string]any{"status": response.StatusCode})
    return nil
}
```

---

### Example 3 — Contract-specific helper function

```go
// Private helper for contract-specific logic (file extension from media type)
func getFileName(postId string, mediaInfo dto.MediaInfo) string {
    switch mediaInfo.MediaType {
    case "VIDEO", "REELS":
        return postId + ".mp4"
    case "IMAGE", "CAROUSEL_ALBUM":
        return postId + ".jpg"
    default:
        return postId + ".mp4"
    }
}
```

Key points:
- Helper is unexported — internal to the webservice package.
- Logic is contract-specific (external API's media type format), not business logic.
- Uses the DTO type, not the entity, since it's about the external contract.

---

## Anti-Patterns

### Anti-Pattern 1 — Creating HTTP client inside webservice (FORBIDDEN)

```go
// BAD — creates its own HTTP client
func NewPaymentWebservice() adapter.PaymentWebservice {
    return &paymentWebservice{
        client: &http.Client{Timeout: 10 * time.Second},  // WRONG
    }
}

// GOOD — HTTP client injected
func NewPaymentWebservice(client httpClient.Client) adapter.PaymentWebservice {
    return &paymentWebservice{client: client}
}
```

---

### Anti-Pattern 2 — Business logic in webservice (FORBIDDEN)

```go
// BAD — domain decision in webservice
func (w *mediaWebservice) Download(ctx context.Context, ...) error {
    mediaInfo := mediaInfoDTO.ToEntity()
    if !mediaInfo.ShouldProcess() {  // WRONG — business logic belongs in service
        return nil
    }
}
```

The webservice returns the entity. The service/entity decides what to do with it.

---

### Anti-Pattern 3 — Exported struct (FORBIDDEN)

```go
// BAD
type PaymentWebservice struct { ... }  // WRONG — must be unexported

// GOOD
type paymentWebservice struct { ... }
```

---

### Anti-Pattern 4 — Returning DTOs instead of entities (FORBIDDEN)

```go
// BAD — leaking DTO to application layer
func (w *mediaWebservice) GetInfo(ctx context.Context, ...) (*dto.MediaInfo, error) {  // WRONG

// GOOD — always return domain entities
func (w *mediaWebservice) GetInfo(ctx context.Context, ...) (*entity.MediaInfo, error) {
    // ... parse response into DTO, then convert
    mediaInfo := mediaInfoDTO.ToEntity()
    return mediaInfo, nil
}
```

---

### Anti-Pattern 5 — URLs or env vars in constructor (FORBIDDEN)

```go
// BAD — URL passed in constructor
type paymentWebservice struct {
    client  httpClient.Client
    baseURL string              // WRONG — read from properties inside methods
}

func NewPaymentWebservice(client httpClient.Client, baseURL string) adapter.PaymentWebservice {
    return &paymentWebservice{client: client, baseURL: baseURL}  // WRONG
}

// GOOD — constructor only receives client + cache
type paymentWebservice struct {
    client httpClient.Client
    cache  *redis.Client
}

func NewPaymentWebservice(client httpClient.Client, cache *redis.Client) adapter.PaymentWebservice {
    return &paymentWebservice{client: client, cache: cache}
}

func (w *paymentWebservice) Process(ctx context.Context, ...) error {
    url := properties.Properties().Services.PaymentApiUrl  // Read from properties here
    // ...
}
```

---

### Anti-Pattern 6 — Missing error codes (FORBIDDEN)

```go
// BAD — generic error without unique code
if err != nil {
    return fmt.Errorf("failed to call API: %w", err)  // WRONG — no error code
}

// GOOD — unique error code at the origin
if err != nil {
    return errs.InternalServerError(err, "PAY-01500", "Failed to call payment API")
}
```

---

## Quick Reference Checklist

When creating or reviewing a webservice, verify every item:

- [ ] Lives in `internal/integration/webservice/` and declares `package webservice`
- [ ] Struct is **unexported** (`paymentWebservice`, not `PaymentWebservice`)
- [ ] Constructor is `New{Name}Webservice(...)` returning `adapter.{Name}Webservice`
- [ ] HTTP client is injected via constructor (never created inside)
- [ ] Constructor only receives HTTP client and optionally cache — no URLs, no env vars
- [ ] URLs and env variables are read from `properties` inside methods, not stored as fields
- [ ] Implements integration adapter that aggregates use case interfaces
- [ ] Returns domain entities, never DTOs
- [ ] DTOs handle entity ↔ contract conversion (`ToEntity()`, `New{Name}(entity)`)
- [ ] DTOs live in `internal/integration/webservice/dto/`
- [ ] Logs `Started` and `Finished` at **Debug** level with raw inputs/outputs
- [ ] Contains only contract logic (status codes, API-specific parsing) — no business logic
- [ ] Every error has a unique code created at the exact line of failure
- [ ] Caching (when present) is transparent to upper layers
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
