---
name: Properties (Clean Architecture)
description: This skill should be used when the user asks to "create properties", "add environment variables", "add config", "read env vars", "add a property", "create a properties file", "access environment configuration", or mentions anything about properties in the internal/infra/properties/ package.
version: 1.0.0
---

# Properties — Golang Clean Architecture

Properties is the centralized environment variable access layer. It reads environment variables and exposes them as typed, structured fields. It is NOT a singleton for non-secret values — every call re-reads from the environment, allowing runtime swaps. It lives in the infrastructure layer and is used inline throughout the integration layer.

---

## Core Rules

### Rule 1: Location

Properties MUST live in `internal/infra/properties/` and declare `package properties`.

```
internal/
└── infra/
    └── properties/
        └── properties.go
```

---

### Rule 2: Not a Singleton — Re-read Every Call

The `Properties()` function MUST create a new struct every time it is called, re-reading all non-secret environment variables from `os.Getenv`. This allows env vars to be swapped during runtime without restarting the application.

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
func Properties() *properties {
    return &properties{
        Application: &application{
            ServerPort:  os.Getenv("SERVER_PORT"),
            ServiceName: os.Getenv("SERVICE_NAME"),
        },
        Services: &services{
            PaymentApiUrl: os.Getenv("SERVICES_PAYMENT_API_URL"),
            Timeout:       time.Duration(parseInt(os.Getenv("SERVICES_TIMEOUT_IN_SECONDS"), 15)) * time.Second,
        },
        // ... re-read everything from os.Getenv
    }
}
```

```go
// BAD — singleton caching env vars (prevents runtime swap)
var instance *properties
var once sync.Once

func Properties() *properties {
    once.Do(func() {
        instance = &properties{...}  // WRONG — only reads once
    })
    return instance
}
```

---

### Rule 3: Secrets — Initialize Once with sync.Once

Secret values (from AWS Secrets Manager, Vault, etc.) are an exception — they MUST be initialized once and stored for the rest of the execution using `sync.Once`.

```go
var (
    initPropertiesOnce sync.Once
    secretsInitialized map[string]string
)

func (p *properties) Init(ctx context.Context, secretsManager SecretsManagerAdapter) {
    initPropertiesOnce.Do(
        func() {
            if secretsManager != nil {
                err := secretsManager.GetSecret(ctx, p.Application.Secret, &secretsInitialized)
                if err != nil {
                    panic("failed to get secret from secrets manager. err: " + err.Error())
                }
            }
        },
    )
}
```

The `Init` function is called by the dependency injector at startup. After initialization, secrets are available via the `secretsInitialized` map.

---

### Rule 4: Used Inline — Never Injected, Never Constructed in Constructors

Properties MUST be called inline wherever env variables are needed. They are NEVER injected via constructors and NEVER stored as struct fields.

```go
// GOOD — used inline in the function
func (w *paymentWebservice) Process(ctx context.Context, ...) error {
    url := properties.Properties().Services.PaymentApiUrl
    timeout := properties.Properties().Services.Timeout
    // ...
}

// BAD — injected in constructor
type paymentWebservice struct {
    client  httpClient.Client
    baseURL string  // WRONG — don't store properties in struct
}

func NewPaymentWebservice(client httpClient.Client, baseURL string) adapter.PaymentWebservice {
    return &paymentWebservice{client: client, baseURL: baseURL}  // WRONG
}

// BAD — stored as field from properties
type paymentWebservice struct {
    client httpClient.Client
    props  *properties.Properties  // WRONG — don't store properties
}
```

---

### Rule 5: Package and Function Naming

The package is `properties` and the function to access it is `Properties()`. Usage reads as `properties.Properties()`.

```go
package properties

func Properties() *properties {
    return &properties{...}
}
```

Usage:
```go
import "myproject/internal/infra/properties"

url := properties.Properties().Services.PaymentApiUrl
bucket := properties.Properties().Aws.S3.MediaBucket
dbHost := properties.Properties().Database.Main.Host
```

---

### Rule 6: Struct Visibility — Unexported Structs, Exported Fields

The properties struct and all its nested grouping structs are **unexported**. However, the fields inside those structs are **exported** so they can be accessed by callers.

```go
// Unexported structs
type properties struct {
    Application *application
    Services    *services
    Database    *database
}

type application struct {
    ServerPort  string    // Exported field
    ServiceName string    // Exported field
}

type services struct {
    PaymentApiUrl string        // Exported field
    Timeout       time.Duration // Exported field
}
```

Exception: If a nested struct needs to be used as a type outside the package (e.g., for database connection config passed to a DB factory), it can be exported.

```go
// Exported because it's used as a parameter type by infra/database
type DatabaseConnection struct {
    Host     string
    User     string
    Password string
    Name     string
    Port     string
}
```

---

### Rule 7: Organize by Concern

Group related env vars into nested structs by concern (application, database, broker, services, cloud providers, etc.).

```go
type properties struct {
    Application *application
    Broker      *broker
    Database    *database
    Redis       *redis
    Aws         *aws
    Services    *services
}
```

This provides a clean, navigable API: `properties.Properties().Aws.S3.MediaBucket`.

---

## Code Examples

### Example 1 — Full properties file

```go
package properties

import (
    "os"
    "strconv"
    "time"
)

type properties struct {
    Application *application
    Broker      *broker
    Database    *database
    Redis       *redis
    Aws         *aws
    Services    *services
}

type application struct {
    ServerPort  string
    ServiceName string
    IsLambda    bool
}

type broker struct {
    Host          string
    Username      string
    Password      string
    ConsumerQueue string
    ProducerQueue string
}

type database struct {
    Main     *DatabaseConnection
    Readonly *DatabaseConnection
    LogLevel string
}

type DatabaseConnection struct {
    Host     string
    User     string
    Password string
    Name     string
    Port     string
}

type redis struct {
    Host         string
    Port         string
    Username     string
    Password     string
    GlobalPrefix string
}

type aws struct {
    S3  *s3
    Sqs *sqs
}

type s3 struct {
    Region      string
    MediaBucket string
}

type sqs struct {
    Region string
}

type services struct {
    PaymentApiUrl      string
    NotificationApiUrl string
    Timeout            time.Duration
}

func Properties() *properties {
    return &properties{
        Application: &application{
            ServerPort:  os.Getenv("SERVER_PORT"),
            ServiceName: os.Getenv("SERVICE_NAME"),
            IsLambda:    os.Getenv("IS_LAMBDA") == "true",
        },
        Broker: &broker{
            Host:          os.Getenv("BROKER_HOST"),
            Username:      os.Getenv("BROKER_USERNAME"),
            Password:      os.Getenv("BROKER_PASSWORD"),
            ConsumerQueue: os.Getenv("BROKER_CONSUMER_QUEUE"),
            ProducerQueue: os.Getenv("BROKER_PRODUCER_QUEUE"),
        },
        Database: &database{
            LogLevel: os.Getenv("DB_LOG_LEVEL"),
            Main: &DatabaseConnection{
                Host:     os.Getenv("MAIN_DB_HOST"),
                User:     os.Getenv("MAIN_DB_USERNAME"),
                Password: os.Getenv("MAIN_DB_PASSWORD"),
                Name:     os.Getenv("MAIN_DB_NAME"),
                Port:     os.Getenv("MAIN_DB_PORT"),
            },
            Readonly: &DatabaseConnection{
                Host:     os.Getenv("READONLY_DB_HOST"),
                User:     os.Getenv("READONLY_DB_USERNAME"),
                Password: os.Getenv("READONLY_DB_PASSWORD"),
                Name:     os.Getenv("READONLY_DB_NAME"),
                Port:     os.Getenv("READONLY_DB_PORT"),
            },
        },
        Redis: &redis{
            Host:         os.Getenv("REDIS_HOST"),
            Port:         os.Getenv("REDIS_PORT"),
            Username:     os.Getenv("REDIS_USERNAME"),
            Password:     os.Getenv("REDIS_PASSWORD"),
            GlobalPrefix: os.Getenv("REDIS_GLOBAL_PREFIX"),
        },
        Aws: &aws{
            S3: &s3{
                Region:      os.Getenv("AWS_S3_REGION"),
                MediaBucket: os.Getenv("AWS_S3_MEDIA_BUCKET"),
            },
            Sqs: &sqs{
                Region: os.Getenv("AWS_SQS_REGION"),
            },
        },
        Services: &services{
            PaymentApiUrl:      os.Getenv("SERVICES_PAYMENT_API_URL"),
            NotificationApiUrl: os.Getenv("SERVICES_NOTIFICATION_API_URL"),
            Timeout:            time.Duration(parseInt(os.Getenv("SERVICES_TIMEOUT_IN_SECONDS"), 15)) * time.Second,
        },
    }
}

func parseInt(s string, defaultVal int) int {
    i, err := strconv.Atoi(s)
    if err != nil {
        return defaultVal
    }
    return i
}
```

---

### Example 2 — Properties with secrets initialization

```go
package properties

import (
    "context"
    "os"
    "sync"
)

var (
    initPropertiesOnce sync.Once
    secretsInitialized map[string]string
)

func (p *properties) Init(ctx context.Context, secretsManager SecretsManagerAdapter) {
    initPropertiesOnce.Do(
        func() {
            if secretsManager != nil {
                err := secretsManager.GetSecret(ctx, p.Database.Secret, &secretsInitialized)
                if err != nil {
                    panic("failed to get secret from secrets manager. err: " + err.Error())
                }
            }
        },
    )
}

// After Init, secrets can be used to override database credentials
func (p *properties) GetDbPassword(key string) string {
    if val, ok := secretsInitialized[key]; ok {
        return val
    }
    return os.Getenv(key)
}
```

---

### Example 3 — Usage in integration layer (inline)

```go
// In a webservice — inline usage
func (w *paymentWebservice) Process(ctx context.Context, ...) error {
    url := properties.Properties().Services.PaymentApiUrl
    req, _ := http.NewRequest(http.MethodPost, url+"/payments", body)
    // ...
}

// In a persistence — inline usage
func (p *mediaPersistence) Save(ctx context.Context, ...) error {
    bucket := properties.Properties().Aws.S3.MediaBucket
    err := p.cloudStorage.Put(ctx, bucket, objectKey, contentType, reader)
    // ...
}

// In infra/database — used to build connections
func NewMainDatabase() *gorm.DB {
    config := properties.Properties().Database.Main
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", config.User, config.Password, config.Host, config.Port, config.Name)
    // ...
}
```

---

## Anti-Patterns

### Anti-Pattern 1 — Singleton for non-secret values (FORBIDDEN)

```go
// BAD — caches env vars, prevents runtime swap
var instance *properties
var once sync.Once

func Properties() *properties {
    once.Do(func() { instance = &properties{...} })
    return instance
}
```

---

### Anti-Pattern 2 — Injecting properties in constructors (FORBIDDEN)

```go
// BAD — properties passed as constructor parameter
func NewPaymentWebservice(client httpClient.Client, paymentUrl string) adapter.PaymentWebservice {
    return &paymentWebservice{client: client, paymentUrl: paymentUrl}  // WRONG
}

// BAD — properties struct stored as field
type paymentWebservice struct {
    props *properties  // WRONG
}
```

---

### Anti-Pattern 3 — Exported properties struct (FORBIDDEN)

```go
// BAD — properties struct is exported
type Properties struct {  // WRONG — must be unexported
    Application *Application
}

// GOOD — unexported struct, exported fields
type properties struct {
    Application *application
}

type application struct {
    ServerPort string  // Exported field — accessible via properties.Properties().Application.ServerPort
}
```

---

### Anti-Pattern 4 — Using os.Getenv directly in integration layer (FORBIDDEN)

```go
// BAD — direct os.Getenv in webservice
func (w *paymentWebservice) Process(ctx context.Context, ...) error {
    url := os.Getenv("SERVICES_PAYMENT_API_URL")  // WRONG — use properties
    // ...
}

// GOOD
func (w *paymentWebservice) Process(ctx context.Context, ...) error {
    url := properties.Properties().Services.PaymentApiUrl
    // ...
}
```

---

## Quick Reference Checklist

When creating or reviewing properties, verify every item:

- [ ] Lives in `internal/infra/properties/` and declares `package properties`
- [ ] Function is `Properties()` returning `*properties` (unexported struct)
- [ ] Creates a new struct every call — NOT a singleton (allows runtime env swap)
- [ ] All fields read from `os.Getenv` inside `Properties()` function
- [ ] Nested structs are unexported, but their fields are exported
- [ ] Secrets use `sync.Once` for one-time initialization
- [ ] `Init()` is called by the injector at startup (not by the properties itself)
- [ ] Properties are used inline (`properties.Properties().X.Y`) — never injected or stored
- [ ] No direct `os.Getenv` calls in integration layer — always go through properties
- [ ] Helper functions (e.g., `parseInt`) handle type conversions with defaults
- [ ] Organized by concern (Application, Database, Broker, Services, Aws, etc.)
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
