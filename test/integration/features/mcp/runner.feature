#language: en
#utf-8

@all @xen_017 @mcp @mcp_runner
Feature: AI clients run collections and read the results
  The "run" tool starts a run on a collection or folder (by path or id) with optional rows, settings
  and environment, defaulting to the remembered or default configuration, and stops a run. The
  "runs" tool reads a run (status, progress, summary and rows, secrets masked) and lists a target's
  stored runs. A run started over MCP follows the AI send policy: an environment marked "Confirm
  before send" asks the user once for the whole run. Runs started from either side are visible from
  the other, and runs never write History.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "shop" variable is set to "01JR0000000000000000000001"
    And the "orders" variable is set to "01JR0000000000000000000002"
    And the "login" variable is set to "01JR0000000000000000000003"
    And the "create" variable is set to "01JR0000000000000000000004"
    And the "get" variable is set to "01JR0000000000000000000005"
    And the "logout" variable is set to "01JR0000000000000000000006"
    And the "dev" variable is set to "01JR0000000000000000000010"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JR0000000000000000000000
      name: Acme
      order:
        - shop
      """
    And the workspace file "acme/collections/shop/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{shop}}
      name: Shop
      order:
        - login.yaml
        - orders
        - logout.yaml
      auth:
        type: bearer
        bearer:
          token: "{{token}}"
      headers:
        - key: X-Suite
          value: runner
          enabled: true
          description: ""
      scripts:
        preRequest: console.log("shop pre")
      """
    And the workspace file "acme/collections/shop/orders/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{orders}}
      name: Orders
      order:
        - create.yaml
        - get.yaml
      """
    And the workspace file "acme/collections/shop/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: "{{api}}/login"
      auth:
        type: inherit
      scripts:
        postResponse: pm.test("login ok", function () { pm.response.to.have.status(200); });
      """
    And the workspace file "acme/collections/shop/orders/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{api}}/orders"
      auth:
        type: inherit
      scripts:
        postResponse: pm.test("created", function () { pm.response.to.have.status(201); });
      """
    And the workspace file "acme/collections/shop/orders/get.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{get}}
      name: Get
      method: GET
      url: "{{api}}/orders/1"
      auth:
        type: inherit
      """
    And the workspace file "acme/collections/shop/logout.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{logout}}
      name: Logout
      method: POST
      url: "{{api}}/logout"
      auth:
        type: inherit
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: token
          value: t-123
          enabled: true
          description: ""
      """
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" contains
      """
      { "version": 1, "activeEnvironmentId": "{{dev}}", "current": {} }
      """
    And the MCP client is named "Claude Code"
    And the 1 "POST" request to "/orders" returns status 201 with the following response
      """
      {"id":"o-1"}
      """

  @start_defaults
  Scenario: An AI client starts a run with defaults and reads it once finished
    When I call the MCP tool "run" with
      """
      { "action": "start", "target": "Shop" }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "running"
    And the response should contain the field "planned" equal to "4"
    And I save the response field "run_id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "POST /login, POST /orders, GET /orders/1, POST /logout" in order
    When I call the MCP tool "runs" with
      """
      { "action": "get", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "completed"
    And the response should contain the field "summary.total_requests" equal to "4"
    And the response should contain the field "rows.0.tests.0.name" equal to "login ok"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0

  @start_explicit
  Scenario: An AI client chooses the rows, settings and environment by name
    Given the "stage" variable is set to "01JR0000000000000000000011"
    And the workspace file "acme/environments/stage.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{stage}}
      name: stage
      variables:
        - key: token
          value: t-stage
          enabled: true
          description: ""
      """
    When I call the MCP tool "run" with
      """
      { "action": "start", "item_id": "{{shop}}", "rows": [{ "request_id": "{{logout}}" }, { "request_id": "{{login}}" }], "functional": { "iterations": 2 }, "environment": "stage" }
      """
    Then the call should succeed
    And I save the response field "run_id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "POST /logout, POST /login, POST /logout, POST /login" in order
    And the 1 "POST" request to "/login" should have the header "Authorization" equal to "Bearer t-stage"

  @data_default_iterations
  Scenario: Without an iteration count, a data file sets one iteration per row
    Given the workspace file "acme/collections/shop/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: "{{api}}/login"
      body:
        mode: raw
        raw:
          language: text
          content: "{{email}}"
      """
    And the file "data/users.csv" exists with content
      """
      email
      a@x.test
      b@x.test
      c@x.test
      """
    When I call the MCP tool "run" with
      """
      { "action": "start", "item_id": "{{shop}}", "rows": [{ "request_id": "{{login}}" }], "functional": { "data_file": "{{temp_root}}/data/users.csv" } }
      """
    Then the call should succeed
    And I save the response field "run_id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the "POST" request to "/login" should have been received 3 times
    And the 3 "POST" request to "/login" should have the body equal to
      """
      c@x.test
      """

  @list_and_stop
  Scenario: An AI client lists stored runs and stops an active one
    Given the "POST" requests to "/orders" respond after 3000ms
    When I call the MCP tool "run" with
      """
      { "action": "start", "target": "Shop" }
      """
    Then the call should succeed
    And I save the response field "run_id" as "run"
    When I call the MCP tool "run" with
      """
      { "action": "start", "target": "Shop" }
      """
    Then the call should fail with code "RUN-02409"
    When I wait 300 ms
    And I call the MCP tool "run" with
      """
      { "action": "stop", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 2000 ms
    When I call the MCP tool "runs" with
      """
      { "action": "list", "target": "Shop" }
      """
    Then the call should succeed
    And the response should contain the field "0.id" equal to "{{run}}"
    And the response should contain the field "0.status" equal to "stopped"

  @errors
  Scenario Outline: Invalid run requests are refused - <case>
    When I call the MCP tool "run" with
      """
      <arguments>
      """
    Then the call should fail with code "<code>"
    And the "POST" request to "/login" should have been received 0 times

    Examples:
      | case             | arguments                                                         | code      |
      | no action        | { "target": "Shop" }                                              | MCP-07400 |
      | stop without id  | { "action": "stop" }                                              | MCP-07400 |
      | unknown target   | { "action": "start", "target": "Nope" }                           | RUN-07404 |
      | empty selection  | { "action": "start", "target": "Shop", "rows": [] }               | RUN-03400 |

  @approval
  Scenario: A confirm-before-send environment asks the user once for the whole run
    Given the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      color: red
      confirmBeforeSend: true
      variables:
        - key: token
          value: t-123
          enabled: true
          description: ""
      """
    When I start calling the MCP tool "run" in the background with
      """
      { "action": "start", "target": "Shop" }
      """
    Then an "ai:approval-requested" event with the field "client" equal to "Claude Code" should have been emitted within 2000 ms
    And the last "ai:approval-requested" event should contain the field "summary" equal to "Run Shop · 4 requests"
    And the "POST" request to "/login" should have been received 0 times
    When I call "AIAccessService.PendingApprovals"
    And I save the response field "0.id" as "approval"
    And I call "AIAccessService.AnswerApproval" with
      """
      { "id": "{{approval}}", "approved": true }
      """
    Then the background call should finish within 5000 ms
    And the call should succeed
    And I save the response field "run_id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the event "ai:approval-requested" should have been emitted 1 times
    And the "POST" request to "/logout" should have been received 1 times

  @approval
  Scenario: A denied approval starts nothing
    Given the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      color: red
      confirmBeforeSend: true
      variables:
        - key: token
          value: t-123
          enabled: true
          description: ""
      """
    When I start calling the MCP tool "run" in the background with
      """
      { "action": "start", "target": "Shop" }
      """
    Then an "ai:approval-requested" event with the field "client" equal to "Claude Code" should have been emitted within 2000 ms
    When I call "AIAccessService.PendingApprovals"
    And I save the response field "0.id" as "approval"
    And I call "AIAccessService.AnswerApproval" with
      """
      { "id": "{{approval}}", "approved": false }
      """
    Then the background call should finish within 5000 ms
    And the error should contain the field "error.message" containing "denied"
    And the "POST" request to "/login" should have been received 0 times
    And the event "runner:finished" should have been emitted 0 times

  @shared_view
  Scenario: A run started in the UI is readable over MCP, and an MCP run shows in the UI
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "ui"
    And a "runner:finished" event with the field "run_id" equal to "{{ui}}" should have been emitted within 5000 ms
    When I call the MCP tool "runs" with
      """
      { "action": "get", "run_id": "{{ui}}" }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "completed"
    Given the "POST" requests to "/login" respond after 1000ms
    When I call the MCP tool "run" with
      """
      { "action": "start", "target": "Shop" }
      """
    Then the call should succeed
    And I save the response field "run_id" as "ai"
    When I call "RunnerService.Active" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.id" equal to "{{ai}}"
    And a "runner:finished" event with the field "run_id" equal to "{{ai}}" should have been emitted within 5000 ms
