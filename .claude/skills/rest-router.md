---
name: REST Router & Server (Clean Architecture)
description: This skill should be used when the user asks to "create a router", "add REST routes", "create an HTTP server", "add an endpoint", "create middleware", "create a context wrapper", "add route validation", or mentions anything about routers, HTTP servers, or Fiber/Gin setup in the internal/infra/server/ package.
version: 1.0.0
---

# REST Router & Server — Golang Clean Architecture

The REST router lives in the infrastructure layer and is responsible for framework-specific HTTP handling. Following Clean Architecture, the framework is fully decoupled from the integration layer through a custom context abstraction and middleware adapter pattern. If the framework is swapped, only the infra layer changes — the integration layer remains untouched.

**Framework choice:** Any HTTP framework can be used (Fiber, Gin, Echo, Chi, etc.). The examples below use Fiber, but the pattern is identical for any framework — only the `infra/server/context/` and `infra/server/middleware/` wrappers change. The integration layer is always framework-free.

---

## Core Rules

### Rule 1: Location & Structure

The router and server infrastructure MUST live in `internal/infra/server/` organized by concern.

```
internal/
└── infra/
    └── server/
        ├── adapter/                    ← Infra-level validator adapter interfaces
        │   ├── authorization.go
        │   ├── payload_validator.go
        │   ├── path_validator.go
        │   ├── query_validator.go
        │   ├── form_validator.go
        │   └── file_type_validator.go
        ├── context/                    ← Framework context wrapper (swap point)
        │   └── fiber.go               ← Replace with gin.go, echo.go, chi.go, etc.
        ├── middleware/                 ← Framework middleware adapter (swap point)
        │   └── fiber_middleware.go     ← Replace with gin_middleware.go, etc.
        └── router/                     ← Route definitions
            ├── user_routes.go          ← Routes for user controller
            ├── order_routes.go         ← Routes for order controller
            └── router.go              ← Main router setup & app creation
```

---

### Rule 2: Framework Decoupling — Custom Context

The framework's context MUST be wrapped in a custom context that implements the `adapter.MiddlewareContext` interface. This interface lives in the integration layer and is framework-agnostic.

**Integration adapter interface** — `internal/integration/adapter/middleware_context.go`:
> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package adapter

import (
    "context"
    "mime/multipart"
)

type MiddlewareContext interface {
    Get(key string, defaultValue ...string) string
    Locals(key any, value ...any) any
    Next() error
    Status(status int) MiddlewareContext
    JSON(data any, cType ...string) error
    Context() context.Context
    Path() string
    Body() []byte
    Method(override ...string) string
    SendString(s string) error
    Query(s string) string
    PathParam(s string) string
    FormParam(s string) string
    FormFile(file string) (*multipart.FileHeader, error)
    Set(header string, value string)
    SendStatus(status int) error
}
```

**Framework context wrapper** (Fiber example — replace with Gin, Echo, etc. as needed) — `internal/infra/server/context/fiber.go`:
```go
package context

import (
    "context"
    "mime/multipart"

    "myproject/internal/integration/adapter"
    "github.com/gofiber/fiber/v2"
)

type fiberCtx struct {
    ctx *fiber.Ctx
}

func Fiber(ctx *fiber.Ctx) adapter.MiddlewareContext {
    return &fiberCtx{ctx: ctx}
}

func (f *fiberCtx) Get(key string, defaultValue ...string) string {
    return f.ctx.Get(key, defaultValue...)
}

func (f *fiberCtx) Locals(key any, value ...any) any {
    return f.ctx.Locals(key, value...)
}

func (f *fiberCtx) Next() error {
    return f.ctx.Next()
}

func (f *fiberCtx) Status(status int) adapter.MiddlewareContext {
    f.ctx = f.ctx.Status(status)
    return f
}

func (f *fiberCtx) JSON(data any, cType ...string) error {
    return f.ctx.JSON(data, cType...)
}

func (f *fiberCtx) Context() context.Context {
    return f.ctx.Context()
}

func (f *fiberCtx) Path() string {
    path := f.ctx.Route().Path
    if path[len(path)-1:] == "/" {
        path = path[:len(path)-1]
    }
    return path
}

func (f *fiberCtx) Body() []byte           { return f.ctx.Body() }
func (f *fiberCtx) Query(s string) string   { return f.ctx.Query(s) }
func (f *fiberCtx) PathParam(s string) string { return f.ctx.Params(s) }
func (f *fiberCtx) FormParam(s string) string { return f.ctx.FormValue(s) }
func (f *fiberCtx) Method(override ...string) string { return f.ctx.Method(override...) }
func (f *fiberCtx) SendString(s string) error { return f.ctx.SendString(s) }
func (f *fiberCtx) Set(header string, value string) { f.ctx.Set(header, value) }
func (f *fiberCtx) SendStatus(status int) error { return f.ctx.SendStatus(status) }
func (f *fiberCtx) FormFile(file string) (*multipart.FileHeader, error) {
    return f.ctx.FormFile(file)
}
```

This way, swapping the framework means only replacing `infra/server/context/` and `infra/server/middleware/` — the entire integration layer (controllers, middlewares, validators) stays untouched. For example, a Gin wrapper would implement the same `MiddlewareContext` interface delegating to `*gin.Context`.

---

### Rule 3: Framework Decoupling — Middleware Adapter

Custom middleware (from the integration layer) must be converted to framework-specific handlers via a middleware adapter.

**Integration middleware interface** — `internal/integration/adapter/middleware.go`:
```go
package adapter

type Middleware interface {
    Handle(ctx MiddlewareContext) error
}
```

**Fiber middleware adapter** — `internal/infra/server/middleware/fiber_middleware.go`:
```go
package middleware

import (
    "myproject/internal/infra/server/context"
    "myproject/internal/integration/adapter"
    "github.com/gofiber/fiber/v2"
)

func FiberMiddleware(middleware adapter.Middleware) fiber.Handler {
    return func(fiberCtx *fiber.Ctx) error {
        ctx := context.Fiber(fiberCtx)
        return middleware.Handle(ctx)
    }
}
```

All integration-layer middlewares (validators, authorization, CORS, trace, error handler) implement `adapter.Middleware` and are converted to Fiber handlers via `FiberMiddleware()`.

---

### Rule 4: Router Must Be Broken Into Separate Route Files

The router MUST be broken into separate files per controller/domain. A single monolithic router file is forbidden — it becomes unmanageable at scale.

```
internal/infra/server/router/
├── router.go              ← Main setup: fiber.New(), global middleware, group creation
├── user_routes.go         ← func (r *router) registerUserRoutes(group fiber.Router)
├── order_routes.go        ← func (r *router) registerOrderRoutes(group fiber.Router)
├── payment_routes.go      ← func (r *router) registerPaymentRoutes(group fiber.Router)
└── health_routes.go       ← func (r *router) registerHealthRoutes(app *fiber.App)
```

**Main router file** — creates the app, applies global middleware, calls route registrars:
```go
func (r *router) Route() *fiber.App {
    app := fiber.New()
    app.Use(recover.New())
    app.Use(middleware.FiberMiddleware(middleware2.NewCorsMiddleware()))

    r.registerHealthRoutes(app)

    v1 := app.Group("/v1")
    for _, mw := range r.middlewares {
        v1.Use(mw)
    }

    r.registerUserRoutes(v1)
    r.registerOrderRoutes(v1)
    r.registerPaymentRoutes(v1)

    return app
}
```

**Individual route file** — focused on one controller's routes:
```go
// File: internal/infra/server/router/user_routes.go
func (r *router) registerUserRoutes(v1 fiber.Router) {
    users := v1.Group("/users")

    users.Get("/",
        r.validateScope("users:read"),
        r.validateQuery("USR-01400", vartype.String, "name"),
        r.validateQuery("USR-02400", vartype.Int, "page", 1),
        r.validateQuery("USR-03400", vartype.Int, "limit", 10),
        func(c *fiber.Ctx) error {
            name := c.Locals("name").(*string)
            page := c.Locals("page").(*int)
            limit := c.Locals("limit").(*int)

            res, err := r.usersController.List(c.Context(), name, *page, *limit)
            if err != nil {
                return err
            }
            return c.Status(fiber.StatusOK).JSON(res)
        })

    users.Post("/",
        r.validateScope("users:write"),
        r.validatePayload("USR-04400", dto.CreateUser{}),
        func(c *fiber.Ctx) error {
            user := c.Locals("CreateUser").(*dto.CreateUser)

            result, err := r.usersController.Create(c.Context(), *user)
            if err != nil {
                return err
            }
            return c.Status(fiber.StatusCreated).JSON(result)
        })

    users.Get("/:id",
        r.validateScope("users:read"),
        r.validatePath("USR-05400", vartype.UUID, "id"),
        func(c *fiber.Ctx) error {
            id := c.Locals("id").(*string)

            res, err := r.usersController.GetById(c.Context(), *id)
            if err != nil {
                return err
            }
            return c.Status(fiber.StatusOK).JSON(res)
        })
}
```

---

### Rule 5: Validation in Routes via Middleware Chain

Routes validate input by chaining middleware before the handler. Validators parse and validate input, storing the result in `ctx.Locals()` for the handler to use.

**Helper methods on the router** (delegate to `FiberMiddleware`):
```go
func (r *router) validateScope(scope string) fiber.Handler {
    return middleware.FiberMiddleware(r.authorizationMiddleware.Validate(scope))
}

func (r *router) validatePayload(errCode string, entity any) fiber.Handler {
    return middleware.FiberMiddleware(r.payloadValidator.Validate(errCode, entity))
}

func (r *router) validatePath(errCode string, varType enums.VarType, param string) fiber.Handler {
    return middleware.FiberMiddleware(r.pathValidator.Validate(errCode, varType, param))
}

func (r *router) validateQuery(errCode string, varType enums.VarType, param string, defaultVal ...any) fiber.Handler {
    return middleware.FiberMiddleware(r.queryValidator.Validate(errCode, varType, param, defaultVal...))
}
```

**Route chain**: scope → path/query validation → payload validation → handler
```go
users.Put("/:id",
    r.validateScope("users:write"),
    r.validatePath("USR-06400", vartype.UUID, "id"),
    r.validatePayload("USR-07400", dto.UpdateUser{}),
    func(c *fiber.Ctx) error {
        id := c.Locals("id").(*string)
        user := c.Locals("UpdateUser").(*dto.UpdateUser)
        // ...
    })
```

---

### Rule 6: Middlewares Live in Integration Layer

Custom middlewares (CORS, trace, authorization, error handler) MUST live in `internal/integration/entrypoint/middleware/`. They implement `adapter.Middleware` and use `adapter.MiddlewareContext` — no framework imports.

```
internal/
└── integration/
    └── entrypoint/
        └── middleware/
            ├── cors.go              ← CORS headers
            ├── trace.go             ← Correlation ID extraction/generation
            ├── authorization.go     ← JWT validation & scope checking
            └── error_handler.go     ← Panic recovery & error formatting
```

The infra layer converts them: `middleware.FiberMiddleware(middleware2.NewCorsMiddleware())`

---

### Rule 7: Validator Adapters Live in Infra Server

Validator adapter interfaces for the router live in `internal/infra/server/adapter/` because they are specific to the router's needs.

```go
// File: internal/infra/server/adapter/payload_validator.go
package adapter

import intAdapter "myproject/internal/integration/adapter"

type PayloadValidator interface {
    Validate(errCode string, entity any) intAdapter.Middleware
}
```

```go
// File: internal/infra/server/adapter/query_validator.go
package adapter

import (
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/entrypoint/enums"
)

type QueryValidator interface {
    Validate(errCode string, varType enums.VarType, param string, defaultVal ...any) adapter.Middleware
}
```

The implementations live in `internal/integration/entrypoint/middleware/` and return `adapter.Middleware`.

---

### Rule 8: Global Middleware Pipeline

Global middleware is applied to the API version group in order:

```go
// 1. Framework-level (infra)
app.Use(recover.New())                                       // Fiber panic recovery
app.Use(middleware.FiberMiddleware(middleware2.NewCorsMiddleware()))  // CORS

// 2. Application-level (integration, converted to Fiber handlers)
v1 := app.Group("/v1")
v1.Use(middleware.FiberMiddleware(traceMiddleware))        // Correlation IDs
v1.Use(middleware.FiberMiddleware(errorHandlerMiddleware)) // Error formatting

// 3. Route-level (per endpoint)
users.Get("/:id",
    r.validateScope("users:read"),                         // Authorization
    r.validatePath("USR-01400", vartype.UUID, "id"),       // Path validation
    handler,                                                // Business logic
)
```

---

## Pre-Built Validator Implementations

These validators are ready to copy into any new project. They live in `internal/integration/entrypoint/middleware/` and implement `adapter.Middleware`.

### VarType Enum (used by query, path, and form validators)

```go
// File: internal/integration/entrypoint/enums/var_type.go
package enums

type VarType string
```

```go
// File: internal/integration/entrypoint/enums/vartype/value.go
package vartype

import "myproject/internal/integration/entrypoint/enums"

const (
    Date                      enums.VarType = "date"
    String                    enums.VarType = "string"
    UUID                      enums.VarType = "uuid"
    CommaDelimitedStringArray enums.VarType = "comma_delimited_string_array"
    Int                       enums.VarType = "int"
    Bool                      enums.VarType = "bool"
    Float64                   enums.VarType = "float64"
    Any                       enums.VarType = "any"
)
```

---

### Infra Server Adapter Interfaces

```go
// File: internal/infra/server/adapter/authorization_validator.go
package adapter

import "myproject/internal/integration/adapter"

type AuthorizationValidator interface {
    Validate(scope string) adapter.Middleware
}
```

```go
// File: internal/infra/server/adapter/payload_validator.go
package adapter

import "myproject/internal/integration/adapter"

type PayloadValidator interface {
    Validate(errorCode string, entity any) adapter.Middleware
}
```

```go
// File: internal/infra/server/adapter/query_validator.go
package adapter

import (
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/entrypoint/enums"
)

type QueryValidator interface {
    Validate(errorCode string, variable enums.VarType, name string, defaultVal ...any) adapter.Middleware
}
```

```go
// File: internal/infra/server/adapter/path_validator.go
package adapter

import (
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/entrypoint/enums"
)

type PathValidator interface {
    Validate(errorCode string, variable enums.VarType, name string, defaultValOpt ...any) adapter.Middleware
}
```

```go
// File: internal/infra/server/adapter/form_validator.go
package adapter

import (
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/entrypoint/enums"
)

type FormValidator interface {
    Validate(errorCode string, variable enums.VarType, name string, defaultValOpt ...any) adapter.Middleware
}
```

```go
// File: internal/infra/server/adapter/file_type_validator.go
package adapter

import (
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/entrypoint/enums"
)

type FileTypeValidator interface {
    Validate(errorCode string, file string, fileType ...enums.FileType) adapter.Middleware
}
```

---

### Payload Validator (JSON body parsing + struct validation)

```go
// File: internal/integration/entrypoint/middleware/payload_validator.go
package middleware

import (
    "encoding/json"
    "errors"
    "reflect"

    infraAdapter "myproject/internal/infra/server/adapter"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/entrypoint/validator"
    "myproject/pkg/errs"
)

type payloadValidator struct {
    validator validator.Validate
}

type payloadValidatorHandler struct {
    validator validator.Validate
    entity    reflect.Type
    errorCode string
}

func PayloadValidator() infraAdapter.PayloadValidator {
    return &payloadValidator{
        validator: validator.CustomValidator(),
    }
}

func (p *payloadValidator) Validate(errorCode string, entity any) adapter.Middleware {
    var entityType reflect.Type
    if entity != nil {
        entityType = reflect.TypeOf(entity)
    }
    return &payloadValidatorHandler{
        validator: validator.CustomValidator(),
        entity:    entityType,
        errorCode: errorCode,
    }
}

func (p *payloadValidatorHandler) Handle(ctx adapter.MiddlewareContext) error {
    if p.entity == nil {
        return ctx.Next()
    }

    entity := reflect.New(p.entity).Interface()
    err := json.Unmarshal(ctx.Body(), entity)
    if err != nil {
        var mappedErr errs.BaseError
        if errors.As(err, &mappedErr) {
            return err
        }
        return errs.BadRequestError("Invalid request body", p.errorCode)
    }

    err = p.validator.Struct(p.errorCode, entity)
    if err != nil {
        return err
    }

    ctx.Locals(p.entity.Name(), entity)

    return ctx.Next()
}
```

---

### Query Validator (query param parsing with type conversion)

```go
// File: internal/integration/entrypoint/middleware/query_validator.go
package middleware

import (
    "encoding/json"
    "fmt"
    "reflect"
    "strconv"
    "strings"
    "time"

    infraAdapter "myproject/internal/infra/server/adapter"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/entrypoint/enums"
    "myproject/internal/integration/entrypoint/enums/vartype"
    "myproject/pkg/errs"
    "github.com/google/uuid"
)

type queryValidator struct{}

type queryValidatorHandler struct {
    defaultVal   any
    errorCode    string
    variableType enums.VarType
    name         string
}

func QueryValidator() infraAdapter.QueryValidator {
    return &queryValidator{}
}

func (q *queryValidator) Validate(errorCode string, variable enums.VarType, name string, defaultValOpt ...any) adapter.Middleware {
    var defaultVal any
    if len(defaultValOpt) > 0 {
        defaultVal = defaultValOpt[0]
    }

    return &queryValidatorHandler{
        errorCode:    errorCode,
        variableType: variable,
        name:         name,
        defaultVal:   defaultVal,
    }
}

func (q *queryValidatorHandler) Handle(ctx adapter.MiddlewareContext) error {
    if q.name == "" {
        return ctx.Next()
    }

    queryValue := ctx.Query(q.name)

    if queryValue == "" {
        ctx.Locals(q.name, typedValue(q.variableType, q.defaultVal))
        return ctx.Next()
    }

    entity, err := q.parseAndValidateQueryValue(queryValue)
    if err != nil {
        return err
    }

    if isNilOrZero(entity) {
        ctx.Locals(q.name, typedValue(q.variableType, q.defaultVal))
    } else {
        ctx.Locals(q.name, entity)
    }

    return ctx.Next()
}

func (q *queryValidatorHandler) parseAndValidateQueryValue(val string) (any, error) {
    errorDetails := errs.ErrorDetails{
        Attribute: q.name,
        Messages:  []string{"INVALID_VALUE"},
    }

    switch q.variableType {
    case vartype.String:
        return &val, nil
    case vartype.UUID:
        if _, err := uuid.Parse(val); err != nil {
            return nil, errs.BadRequestError(
                fmt.Sprintf("Invalid value for query parameter '%s': must be a valid UUID", q.name),
                q.errorCode, errorDetails,
            )
        }
        return &val, nil
    case vartype.CommaDelimitedStringArray:
        if val == "" {
            return []string{}, nil
        }
        return strings.Split(val, ","), nil
    case vartype.Int:
        parsed, err := strconv.Atoi(val)
        if err != nil {
            return nil, errs.BadRequestError(
                fmt.Sprintf("Invalid value for query parameter '%s': must be a valid integer", q.name),
                q.errorCode, errorDetails,
            )
        }
        if q.name == "page" || q.name == "limit" {
            if parsed < 1 {
                return nil, errs.BadRequestError(
                    fmt.Sprintf("Invalid value for query parameter '%s': must be greater than 0", q.name),
                    q.errorCode, errorDetails,
                )
            }
            if q.name == "limit" && parsed > 1000 {
                return nil, errs.BadRequestError(
                    fmt.Sprintf("Invalid value for query parameter '%s': must not exceed 1000", q.name),
                    q.errorCode, errorDetails,
                )
            }
        }
        return &parsed, nil
    case vartype.Bool:
        parsed, err := strconv.ParseBool(val)
        if err != nil {
            return nil, errs.BadRequestError(
                fmt.Sprintf("Invalid value for query parameter '%s': must be a valid boolean (true, false, 1, 0)", q.name),
                q.errorCode, errorDetails,
            )
        }
        return &parsed, nil
    case vartype.Float64:
        parsed, err := strconv.ParseFloat(val, 64)
        if err != nil {
            return nil, errs.BadRequestError(
                fmt.Sprintf("Invalid value for query parameter '%s': must be a valid number", q.name),
                q.errorCode, errorDetails,
            )
        }
        return &parsed, nil
    case vartype.Date:
        parsed, err := time.Parse("2006-01-02", val)
        if err != nil {
            return nil, errs.BadRequestError(
                fmt.Sprintf("Invalid value for query parameter '%s': must be a valid date (YYYY-MM-DD)", q.name),
                q.errorCode, errorDetails,
            )
        }
        return &parsed, nil
    default:
        var obj any
        if err := json.Unmarshal([]byte(val), &obj); err != nil {
            return nil, errs.BadRequestError(
                fmt.Sprintf("Invalid value for query parameter '%s': must be valid JSON", q.name),
                q.errorCode, errorDetails,
            )
        }
        return obj, nil
    }
}

func isNilOrZero(value any) bool {
    return value == nil || reflect.ValueOf(value).IsZero()
}

func typedValue(t enums.VarType, val any) any {
    if val != nil {
        v := reflect.ValueOf(val)
        ptr := reflect.New(v.Type())
        ptr.Elem().Set(v)
        return ptr.Interface()
    }

    switch t {
    case vartype.String, vartype.UUID:
        var v *string
        return v
    case vartype.Int:
        var v *int
        return v
    case vartype.Bool:
        var v *bool
        return v
    case vartype.Float64:
        var v *float64
        return v
    case vartype.Date:
        var v *time.Time
        return v
    default:
        return nil
    }
}
```

---

### Path Validator (path param parsing & validation)

```go
// File: internal/integration/entrypoint/middleware/path_validator.go
package middleware

import (
    "encoding/json"
    "reflect"
    "strconv"
    "time"

    infraAdapter "myproject/internal/infra/server/adapter"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/entrypoint/enums"
    "myproject/internal/integration/entrypoint/enums/vartype"
    "myproject/pkg/errs"
    "github.com/google/uuid"
)

type pathValidator struct{}

type pathValidatorHandler struct {
    defaultVal   any
    variableType enums.VarType
    name         string
    errorCode    string
}

func PathValidator() infraAdapter.PathValidator {
    return &pathValidator{}
}

func (p *pathValidator) Validate(errorCode string, variable enums.VarType, name string, defaultValOpt ...any) adapter.Middleware {
    var defaultVal any
    if len(defaultValOpt) > 0 {
        defaultVal = defaultValOpt[0]
    }

    return &pathValidatorHandler{
        errorCode:    errorCode,
        variableType: variable,
        name:         name,
        defaultVal:   defaultVal,
    }
}

func (p *pathValidatorHandler) Handle(ctx adapter.MiddlewareContext) error {
    if p.name == "" {
        return ctx.Next()
    }

    pathValue := ctx.PathParam(p.name)
    var entity any
    switch p.variableType {
    case vartype.Any:
        entity = pathValue
    case vartype.String, vartype.UUID:
        _, err := uuid.Parse(pathValue)
        if err != nil {
            return errs.ClientError("invalid path variable", p.errorCode, 400)
        }
        entity = pathValue
    case vartype.Int:
        var err error
        entity, err = strconv.Atoi(pathValue)
        if err != nil {
            entity = nil
        }
    case vartype.Bool:
        var err error
        entity, err = strconv.ParseBool(pathValue)
        if err != nil {
            entity = nil
        }
    case vartype.Float64:
        var err error
        entity, err = strconv.ParseFloat(pathValue, 64)
        if err != nil {
            entity = nil
        }
    case vartype.Date:
        parsed, err := time.Parse("2006-01-02", pathValue)
        if err != nil {
            return errs.ClientError("invalid path variable", p.errorCode, 400)
        }
        entity = &parsed
    default:
        err := json.Unmarshal([]byte(pathValue), entity)
        if err != nil {
            entity = nil
        }
    }

    if entity == nil || reflect.ValueOf(entity).IsZero() {
        if p.defaultVal == nil {
            return errs.ClientError("invalid path variable", p.errorCode, 400)
        } else {
            ctx.Locals(p.name, p.defaultVal)
        }
    } else {
        ctx.Locals(p.name, entity)
    }

    return ctx.Next()
}
```

---

### Form Validator (form param parsing)

```go
// File: internal/integration/entrypoint/middleware/form_validator.go
package middleware

import (
    "encoding/json"
    "errors"
    "reflect"
    "strconv"
    "strings"

    infraAdapter "myproject/internal/infra/server/adapter"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/entrypoint/enums"
    "myproject/internal/integration/entrypoint/enums/vartype"
)

type formValidator struct{}

type formValidatorHandler struct {
    defaultVal   any
    variableType enums.VarType
    name         string
    errorCode    string
}

func FormValidator() infraAdapter.FormValidator {
    return &formValidator{}
}

func (f *formValidator) Validate(errorCode string, variable enums.VarType, name string, defaultValOpt ...any) adapter.Middleware {
    var defaultVal any
    if len(defaultValOpt) > 0 {
        defaultVal = defaultValOpt[0]
    }

    return &formValidatorHandler{
        errorCode:    errorCode,
        variableType: variable,
        name:         name,
        defaultVal:   defaultVal,
    }
}

func (f *formValidatorHandler) Handle(ctx adapter.MiddlewareContext) error {
    if f.name == "" {
        return ctx.Next()
    }

    formValue := ctx.FormParam(f.name)
    var entity any
    switch f.variableType {
    case vartype.CommaDelimitedStringArray:
        if formValue != "" {
            entity = strings.Split(formValue, ",")
        } else {
            entity = []string{}
        }
    case vartype.String:
        entity = formValue
    case vartype.Int:
        var err error
        entity, err = strconv.Atoi(formValue)
        if err != nil {
            entity = nil
        }
    case vartype.Bool:
        var err error
        entity, err = strconv.ParseBool(formValue)
        if err != nil {
            entity = nil
        }
    case vartype.Float64:
        var err error
        entity, err = strconv.ParseFloat(formValue, 64)
        if err != nil {
            entity = nil
        }
    default:
        err := json.Unmarshal([]byte(formValue), entity)
        if err != nil {
            entity = nil
        }
    }

    if entity == nil || reflect.ValueOf(entity).IsZero() {
        if f.defaultVal == nil {
            return errors.New("invalid form variable")
        } else {
            ctx.Locals(f.name, f.defaultVal)
        }
    } else {
        ctx.Locals(f.name, entity)
    }

    return ctx.Next()
}
```

---

### Authorization Validator (JWT + scope checking)

```go
// File: internal/integration/entrypoint/middleware/authorization.go
package middleware

import (
    "strings"

    "myproject/internal/application/usecase"
    infraAdapter "myproject/internal/infra/server/adapter"
    "myproject/internal/integration/adapter"
    "myproject/pkg/errs"
    "github.com/lestrrat-go/jwx/v2/jwk"
    "github.com/lestrrat-go/jwx/v2/jwt"
)

type authorizationValidator struct {
    listJwkUseCase usecase.ListJwk
}

type authorizationValidatorHandler struct {
    listJwkUseCase usecase.ListJwk
    scope          string
}

func Authorization(listJwkUseCase usecase.ListJwk) infraAdapter.AuthorizationValidator {
    return &authorizationValidator{
        listJwkUseCase: listJwkUseCase,
    }
}

func (a *authorizationValidator) Validate(scope string) adapter.Middleware {
    return &authorizationValidatorHandler{
        listJwkUseCase: a.listJwkUseCase,
        scope:          scope,
    }
}

func (a *authorizationValidatorHandler) Handle(ctx adapter.MiddlewareContext) error {
    base64Jwt := extractToken(ctx.Get("Authorization"))

    keySet, err := a.getKeySet(ctx)
    if err != nil {
        return err
    }

    token, err := a.parseToken(base64Jwt, keySet)
    if err != nil {
        return err
    }

    if !a.isAuthorized(token) {
        return errs.ForbiddenError("forbidden", "FO-00403")
    }

    if clientId, ok := token.Get("client_id"); ok {
        ctx.Locals("client_id", clientId)
    } else {
        return errs.ForbiddenError("forbidden", "FO-02403")
    }
    if userId, ok := token.Get("sub"); ok {
        ctx.Locals("user_id", userId)
    } else {
        return errs.ForbiddenError("forbidden", "FO-03403")
    }

    if roleValue, ok := token.Get("role"); ok {
        ctx.Locals("role", roleValue)
    }

    return ctx.Next()
}

func extractToken(authorizationHeader string) string {
    return strings.Replace(authorizationHeader, "Bearer ", "", 1)
}

func (a *authorizationValidatorHandler) getKeySet(ctx adapter.MiddlewareContext) (jwk.Set, error) {
    jwkSet, err := a.listJwkUseCase.ListJwk(ctx.Context())
    if err != nil {
        return nil, err
    }
    return jwkSet, nil
}

func (a *authorizationValidatorHandler) parseToken(base64Jwt string, keySet jwk.Set) (jwt.Token, error) {
    token, err := jwt.Parse([]byte(base64Jwt), jwt.WithKeySet(keySet))
    if err != nil {
        return nil, errs.UnauthorizedError(err.Error(), "FO-01401")
    }
    return token, nil
}

func (a *authorizationValidatorHandler) isAuthorized(token jwt.Token) bool {
    if scopeObject, ok := token.Get("scope"); ok {
        if scopes, ok := scopeObject.([]any); ok {
            for _, scopeName := range scopes {
                if a.scope == scopeName.(string) {
                    return true
                }
            }
            return false
        }

        if scopeStr, ok := scopeObject.(string); ok {
            scopes := strings.Split(scopeStr, " ")
            for _, scope := range scopes {
                if a.scope == scope {
                    return true
                }
            }
            return false
        }

        return false
    }
    return false
}
```

---

## Code Examples

### Example 1 — Main router file (setup & delegation)

```go
// File: internal/infra/server/router/router.go
package router

import (
    "net/http"

    adapter2 "myproject/internal/infra/server/adapter"
    "myproject/internal/infra/server/middleware"
    "myproject/internal/infra/properties"
    "myproject/internal/integration/adapter"
    "myproject/internal/integration/entrypoint/enums"
    middleware2 "myproject/internal/integration/entrypoint/middleware"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/recover"
)

type router struct {
    authorizationMiddleware adapter2.AuthorizationValidator
    payloadValidator        adapter2.PayloadValidator
    pathValidator           adapter2.PathValidator
    queryValidator          adapter2.QueryValidator
    usersController         adapter.UsersController
    orderController         adapter.OrderController
    healthController        http.HandlerFunc
    middlewares             []fiber.Handler
}

func FiberRouter(
    healthController http.HandlerFunc,
    authorizationMiddleware adapter2.AuthorizationValidator,
    middlewares map[int]adapter.Middleware,
    payloadValidator adapter2.PayloadValidator,
    pathValidator adapter2.PathValidator,
    queryValidator adapter2.QueryValidator,
    usersController adapter.UsersController,
    orderController adapter.OrderController,
) *router {
    fiberMiddlewares := make([]fiber.Handler, 0)
    for i := 0; i < len(middlewares); i++ {
        fiberMiddlewares = append(fiberMiddlewares, middleware.FiberMiddleware(middlewares[i]))
    }

    return &router{
        healthController:        healthController,
        authorizationMiddleware: authorizationMiddleware,
        payloadValidator:        payloadValidator,
        pathValidator:           pathValidator,
        queryValidator:          queryValidator,
        middlewares:             fiberMiddlewares,
        usersController:         usersController,
        orderController:         orderController,
    }
}

func (r *router) Route() *fiber.App {
    app := fiber.New()
    app.Use(recover.New())
    app.Use(middleware.FiberMiddleware(middleware2.NewCorsMiddleware()))

    r.registerHealthRoutes(app)

    v1 := app.Group("/v1")
    for _, mw := range r.middlewares {
        v1.Use(mw)
    }

    r.registerUserRoutes(v1)
    r.registerOrderRoutes(v1)

    return app
}

// Validation helpers
func (r *router) validateScope(scope string) fiber.Handler {
    return middleware.FiberMiddleware(r.authorizationMiddleware.Validate(scope))
}

func (r *router) validatePayload(errCode string, entity any) fiber.Handler {
    return middleware.FiberMiddleware(r.payloadValidator.Validate(errCode, entity))
}

func (r *router) validatePath(errCode string, varType enums.VarType, param string) fiber.Handler {
    return middleware.FiberMiddleware(r.pathValidator.Validate(errCode, varType, param))
}

func (r *router) validateQuery(errCode string, varType enums.VarType, param string, defaultVal ...any) fiber.Handler {
    return middleware.FiberMiddleware(r.queryValidator.Validate(errCode, varType, param, defaultVal...))
}
```

---

### Example 2 — Separate route file

```go
// File: internal/infra/server/router/user_routes.go
package router

import (
    "myproject/internal/integration/entrypoint/dto"
    "myproject/internal/integration/entrypoint/enums/vartype"
    "github.com/gofiber/fiber/v2"
)

func (r *router) registerUserRoutes(v1 fiber.Router) {
    users := v1.Group("/users")

    users.Get("/",
        r.validateScope("users:read"),
        r.validateQuery("USR-01400", vartype.Int, "page", 1),
        r.validateQuery("USR-02400", vartype.Int, "limit", 10),
        func(c *fiber.Ctx) error {
            page := c.Locals("page").(*int)
            limit := c.Locals("limit").(*int)

            res, err := r.usersController.List(c.Context(), *page, *limit)
            if err != nil {
                return err
            }
            return c.Status(fiber.StatusOK).JSON(res)
        })

    users.Post("/",
        r.validateScope("users:write"),
        r.validatePayload("USR-03400", dto.CreateUser{}),
        func(c *fiber.Ctx) error {
            user := c.Locals("CreateUser").(*dto.CreateUser)

            result, err := r.usersController.Create(c.Context(), *user)
            if err != nil {
                return err
            }
            return c.Status(fiber.StatusCreated).JSON(result)
        })

    users.Get("/:id",
        r.validateScope("users:read"),
        r.validatePath("USR-04400", vartype.UUID, "id"),
        func(c *fiber.Ctx) error {
            id := c.Locals("id").(*string)

            res, err := r.usersController.GetById(c.Context(), *id)
            if err != nil {
                return err
            }
            return c.Status(fiber.StatusOK).JSON(res)
        })

    users.Put("/:id",
        r.validateScope("users:write"),
        r.validatePath("USR-05400", vartype.UUID, "id"),
        r.validatePayload("USR-06400", dto.UpdateUser{}),
        func(c *fiber.Ctx) error {
            id := c.Locals("id").(*string)
            user := c.Locals("UpdateUser").(*dto.UpdateUser)

            result, err := r.usersController.Update(c.Context(), *id, *user)
            if err != nil {
                return err
            }
            return c.Status(fiber.StatusOK).JSON(result)
        })

    users.Delete("/:id",
        r.validateScope("users:write"),
        r.validatePath("USR-07400", vartype.UUID, "id"),
        func(c *fiber.Ctx) error {
            id := c.Locals("id").(*string)

            err := r.usersController.Delete(c.Context(), *id)
            if err != nil {
                return err
            }
            return c.SendStatus(fiber.StatusNoContent)
        })
}
```

---

### Example 3 — Health route (separate file)

```go
// File: internal/infra/server/router/health_routes.go
package router

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/adaptor"
)

func (r *router) registerHealthRoutes(app *fiber.App) {
    app.Get("/health", adaptor.HTTPHandlerFunc(r.healthController))
}
```

---

## Architecture Diagram

```
┌──────────────────────────────────────────────────────────┐
│ Infrastructure Layer (infra/server/)                      │
│                                                           │
│  router/          → Fiber routes, validation chains       │
│  context/fiber.go → Wraps fiber.Ctx → MiddlewareContext   │
│  middleware/       → FiberMiddleware() adapter function    │
│  adapter/          → Validator interfaces (infra-specific)│
│                                                           │
│  ⬇ depends on (via FiberMiddleware adapter)               │
├──────────────────────────────────────────────────────────┤
│ Integration Layer (integration/entrypoint/)               │
│                                                           │
│  adapter/          → Middleware, MiddlewareContext ifaces  │
│  middleware/        → Authorization, CORS, Trace, etc.    │
│  validator/         → Custom validation rules             │
│  controller/        → Business endpoint handlers          │
│                                                           │
│  ⬇ depends on (via application adapters)                  │
├──────────────────────────────────────────────────────────┤
│ Application Layer                                         │
│  Service → UseCase → Domain                               │
└──────────────────────────────────────────────────────────┘

Key: Integration layer has ZERO framework imports.
     Only infra layer imports Fiber/Gin/etc.
```

---

## Anti-Patterns

### Anti-Pattern 1 — Framework types in integration layer (FORBIDDEN)

```go
// BAD — middleware importing Fiber
import "github.com/gofiber/fiber/v2"

func (m *corsMiddleware) Handle(c *fiber.Ctx) error { ... }  // WRONG

// GOOD — uses framework-agnostic MiddlewareContext
func (m *corsMiddleware) Handle(ctx adapter.MiddlewareContext) error { ... }
```

---

### Anti-Pattern 2 — Monolithic router file (FORBIDDEN)

```go
// BAD — 1000+ lines all in router.go
func (r *router) Route() *fiber.App {
    // ... 50 routes for users
    // ... 30 routes for orders
    // ... 40 routes for payments
    // THIS IS WRONG — split into separate files
}
```

---

### Anti-Pattern 3 — Validation in handler instead of middleware chain (FORBIDDEN)

```go
// BAD — inline validation in handler
users.Post("/", func(c *fiber.Ctx) error {
    var user dto.CreateUser
    if err := c.BodyParser(&user); err != nil { return err }  // WRONG
    if err := validate.Struct(&user); err != nil { return err } // WRONG
    // ...
})

// GOOD — validation via middleware chain
users.Post("/",
    r.validateScope("users:write"),
    r.validatePayload("USR-01400", dto.CreateUser{}),
    func(c *fiber.Ctx) error {
        user := c.Locals("CreateUser").(*dto.CreateUser)  // Already validated
        // ...
    })
```

---

### Anti-Pattern 4 — Controllers importing Fiber (FORBIDDEN)

```go
// BAD — controller depends on framework
import "github.com/gofiber/fiber/v2"

func (c *userController) Create(ctx *fiber.Ctx) error { ... }  // WRONG

// GOOD — controller uses context.Context, receives domain types
func (c *userController) Create(ctx context.Context, user dto.CreateUser) (*dto.UserResponse, error) { ... }
```

---

## Quick Reference Checklist

When creating or reviewing router infrastructure, verify every item:

- [ ] Framework context wrapped in custom context implementing `adapter.MiddlewareContext`
- [ ] `FiberMiddleware()` adapter converts `adapter.Middleware` → `fiber.Handler`
- [ ] Integration layer has ZERO framework imports (no fiber, gin, etc.)
- [ ] Router split into separate files per controller/domain
- [ ] Main router file only handles setup, global middleware, and delegation
- [ ] Validation done via middleware chain helpers (`validateScope`, `validatePayload`, `validatePath`, `validateQuery`)
- [ ] Validated data accessed via `c.Locals()` in handlers
- [ ] Custom middlewares live in `integration/entrypoint/middleware/`
- [ ] Validator adapter interfaces live in `infra/server/adapter/`
- [ ] Error codes passed to validators for consistent error responses
- [ ] Global middleware order: recover → CORS → trace → error handler → route-specific
- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
