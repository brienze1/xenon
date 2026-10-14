---
name: BDD Step Definitions (Clean Architecture)
description: This skill should be used when the user asks to "create step definitions", "add BDD steps", "write Given/When/Then steps", or mentions anything about step definitions in test/integration/steps/.
version: 1.0.0
---

# BDD Step Definitions — Golang Clean Architecture

Step definitions are the glue between Gherkin feature files and Go test code. They MUST be generic, reusable across all feature files, and broken into separate files by concern.

**Location:** `test/integration/steps/`

---

## Core Rules

### Rule 1: Break Into Separate Files (CRITICAL)

Step definitions MUST be split into logical files — never one monolithic file. A single step definitions file with hundreds of lines is FORBIDDEN. Each file should be focused on a specific concern and remain small and manageable.

```
test/integration/steps/
├── step_definitions_test.go   ← Test runner, testUtils struct, InitializeScenario, Before/After (ONLY setup, no step implementations)
├── given_db_steps_test.go     ← Given steps for database setup (table creation, clearing)
├── given_env_steps_test.go    ← Given steps for environment variables and headers
├── given_api_steps_test.go    ← Given steps for mock API response configuration
├── when_http_steps_test.go    ← When steps for HTTP calls (REST API tests)
├── when_event_steps_test.go   ← When steps for event processing (SQS, SNS, Lambda)
├── then_response_steps_test.go ← Then steps for HTTP response assertions
├── then_db_steps_test.go      ← Then steps for database state verification
├── then_queue_steps_test.go   ← Then steps for SQS/SNS message verification
├── then_api_steps_test.go     ← Then steps for mock API request verification
└── helpers_test.go            ← Shared utilities (execute HTTP, parse JSON, etc.)
```

The main `step_definitions_test.go` file should ONLY contain:
- `TestFeatures` function (test runner)
- `testUtils` struct definition
- `InitializeScenario` function (registers steps by calling registration functions from other files)
- `Before`/`After` hooks
- Mock initialization and injection

It should NOT contain any step implementation functions. Those live in the grouped files above.

### Rule 2: Steps Must Be Generic and Reusable

Every step MUST be reusable across multiple scenarios. Never create a step that only works for one test.

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// GOOD — generic, works for any table
ctx.Given(`^the "([^"]*)" exists$`, test.theExists)
ctx.Given(`^the tables are empty$`, test.theTablesAreEmpty)

// BAD — too specific, only works for one scenario
ctx.Given(`^a campaign with id 123 and title "Test" exists$`, test.specificCampaignSetup)  // WRONG
```

### Rule 3: Use Regex Parameters for Flexibility

Steps use regex groups `([^"]*)` to capture variable values from feature files:

```go
ctx.Given(`^the "([^"]*)" env var is set to "([^"]*)"$`, test.theEnvVarIsSetTo)
ctx.Then(`^the db should contain (\d+) objects in the "([^"]*)" table$`, test.theDbShouldContainObjectsInTheTable)
ctx.Then(`^the sqs queue "([^"]*)" should have (\d+) messages published$`, test.theSqsQueueShouldHaveMessagesPublished)
```

---

## Test Setup Pattern

```go
type testUtils struct {
    uri          string
    headers      map[string]string
    client       *http.Client
    response     *response
    time         *mock.Time
    api          *mock.ApiMock
    mainDb       *mock.Db
    redis        *redis.Client
    broker       *mock.BrokerClient[any]
    cloudStorage *mock.CloudStorageClient
}

type response struct {
    status int
    body   any
    err    error
}

func InitializeScenario(ctx *godog.ScenarioContext) {
    test := &testUtils{
        uri:          "http://localhost:" + properties.Properties().Application.ServerPort,
        client:       &http.Client{},
        time:         mock.NewTime(),
        api:          mock.NewApiServer(),
        redis:        mock.NewRedis(),
        broker:       mock.NewBrokerClient[any](),
        cloudStorage: mock.NewCloudStorageClient(),
        mainDb: mock.NewDb("main", map[string]any{
            "users":  &model.User{},
            "orders": &model.Order{},
        }),
    }

    // Inject mocks into factories ONE TIME (never repeated per-scenario)
    database.SetMainDb(test.mainDb.DbConn)
    database.SetRedis(test.redis)
    storage.SetCloudStorage(test.cloudStorage)
    timeutils.SetCurrentTimeProvider(test.time.Now)

    ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
        test.before()  // Resets mock STATE only — never re-injects mocks
        return ctx, nil
    })

    // Register all steps
    // Given
    ctx.Given(`^the tables are empty$`, test.theTablesAreEmpty)
    ctx.Given(`^the "([^"]*)" env var is set to "([^"]*)"$`, test.theEnvVarIsSetTo)
    ctx.Given(`^the "([^"]*)" exists$`, test.theExists)
    ctx.Given(`^the header is empty$`, test.theHeaderIsEmpty)
    ctx.Given(`^the header contains the key "([^"]*)" with "([^"]*)"$`, test.theHeaderContainsTheKeyWith)
    ctx.Given(`^the (\d+) "([^"]*)" request to "([^"]*)" returns status (\d+) with the following response$`, test.theRequestToReturnsStatusWithTheFollowingResponse)

    // When
    ctx.When(`^I call "([^"]*)" "([^"]*)" with the following payload$`, test.iCallWithTheFollowingPayload)
    ctx.When(`^the following event is received via sqs$`, test.theFollowingEventIsReceivedViaQueue)
    ctx.When(`^the following event is received via sns$`, test.theFollowingEventIsReceivedViaTopic)

    // Then
    ctx.Then(`^the status returned should be (\d+)$`, test.theStatusReturnedShouldBe)
    ctx.Then(`^the response should contain the field "([^"]*)" equal to "([^"]*)"$`, test.theResponseShouldContainTheFieldEqualTo)
    ctx.Then(`^the db should contain (\d+) objects in the "([^"]*)" table$`, test.theDbShouldContainObjectsInTheTable)
    ctx.Then(`^the sqs queue "([^"]*)" should have (\d+) messages published$`, test.theSqsQueueShouldHaveMessagesPublished)
    ctx.Then(`^the sns topic "([^"]*)" should have (\d+) messages published$`, test.theSnsTopicShouldHaveMessagesPublished)
}
```

---

## Standard Step Library

### Given Steps (Setup)

| Step pattern | Purpose |
|---|---|
| `the tables are empty` | Clears all database tables |
| `the "VAR" env var is set to "value"` | Sets environment variable |
| `the "VAR" env var is set to the mocked api url` | Sets env var to mock API base URL |
| `the header is empty` | Resets request headers |
| `the header contains the key "key" with "value"` | Adds header for next request |
| `the "table_name" exists` + JSON docstring | Inserts data into database table |
| `the N "METHOD" request to "path" returns status CODE with the following response` + JSON | Mocks external API response |
| `the N "METHOD" request to "path" returns status CODE with the "file" file and type "type"` | Mocks file response |

### When Steps (Actions)

| Step pattern | Purpose |
|---|---|
| `I call "METHOD" "path" with the following payload` + JSON | HTTP request with body |
| `I call "METHOD" "path" with multipart form data` + table | Multipart form request |
| `I call "METHOD" "path" with the following csv payload` + CSV | CSV upload |
| `the following event is received via sqs` + JSON | Simulates SQS message |
| `the following event is received via sns` + JSON | Simulates SNS message |

### Then Steps (Assertions)

| Step pattern | Purpose |
|---|---|
| `the status returned should be CODE` | Validates HTTP status |
| `the response should contain the field "path" equal to "value"` | Validates response JSON field |
| `the response should contain the field "path" array length equal to N` | Validates array length |
| `the lambda should finish without errors` | Validates no error returned |
| `the lambda should finish with "message" error` | Validates specific error |
| `the db should contain N objects in the "table" table` | Counts DB records |
| `the db should contain the "table" with the id "id" colum "col" equal to "val"` | Validates specific record |
| `the sqs queue "url" should have N messages published` | Counts SQS messages |
| `the sqs queue "url" should have the N message published with "field" field equal to "value"` | Validates SQS message content |
| `the sns topic "arn" should have N messages published` | Counts SNS messages |
| `the N "METHOD" request for "path" headers should have the variable "key" with value "val"` | Validates mock API received headers |
| `the N "METHOD" request for "path" json body should have the variable "field" with value "val"` | Validates mock API received body |

### MCP, stdio-bridge and echo steps (Xenon)

These live in `when_mcp_steps_test.go`, `then_mcp_steps_test.go` and `given_api_steps_test.go`. Every scenario gets
a free MCP port in `XENON_MCP_PORT` (`{{mcp_port}}`, `{{mcp_url}}`). An MCP result becomes "the response" (or "the
error", as `{"error": {...}}`), so the generic response and error steps apply to it.

| Step pattern | Purpose |
|---|---|
| `the MCP client is named "name"` | Picks the client (one session per name, kept for the scenario) |
| `I call the MCP tool "tool" with` + JSON / `I call the MCP tool "tool"` | Calls a tool; the text result is the response |
| `I call the MCP tool "tool" N times with` + JSON | Repeats a call (fails on the first error) |
| `I start calling the MCP tool "tool" in the background [as "name"] with` + JSON | Starts a call that blocks (for example, waiting for approval) |
| `the background call ["name"] should finish within N ms` | Waits for a background call and makes its result the response |
| `I list the MCP tools` | The response is the tool list |
| `the MCP session is closed` | Closes the current client's session |
| `I post to the MCP endpoint with the headers` + table / `I post a body of N bytes to the MCP endpoint` | Raw HTTP posts; assert with `the route response status should be N` |
| `the MCP endpoint should refuse connections` | Nothing listens on the MCP port |
| `I run the stdio bridge with the input` + lines | Runs `xenon mcp` with the given stdin |
| `the bridge exit code should be N` / `the bridge output should [not] contain\|match "x"` / `the bridge error output should contain "x"` | Bridge assertions |
| `the mock API echoes every request to "path" [with status N]` | Echo route: JSON `{method,path,query,headers,body}`, headers reflected as `X-Echo-<Header>` |
| `the "var" variable is set to the port of the mocked api url` | The mock API's port |
| `the error should not contain "x"` / `the response should have at most N characters` / `the response field "f" should have N characters` / `every item of the response field "f" should have the field "k" matching "re"` | Generic assertions |
| `the Xenon window should have been brought to the front N times` | Window-focuser mock |

Register steps that can open or continue a scenario with `ctx.Step`: godog binds `Given`/`When`/`Then` strictly.
Step arguments cannot contain `"`, so compare such values with a docstring step or a regex instead.

---

## Before Hook (Reset Mock STATE Only)

The `before()` hook clears accumulated mock data between scenarios. It does NOT re-inject mocks via factory setters (that happened once in `InitializeScenario`) and does NOT call factory reset functions (they don't exist).

```go
func (t *testUtils) before() {
    t.headers = make(map[string]string)

    // Clear mock STATE only — mock instances live forever
    mock.ClearRedis(t.redis)
    t.broker.Reset()
    t.cloudStorage.Reset()
    t.mainDb.ClearDB()
    t.api.Start()

    // NEVER do this in before():
    // database.SetRedis(t.redis)           // WRONG — setter is ONE-TIME in InitializeScenario
    // storage.SetCloudStorage(t.storage)   // WRONG — setter is ONE-TIME in InitializeScenario
    // someFactory.ResetSomething()          // WRONG — factories NEVER have Reset functions
}
```

Every scenario starts with clean mock state — accumulated messages, files, keys, and rows are cleared, but the mock instances and factory wiring remain intact.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
