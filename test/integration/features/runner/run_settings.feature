#language: en
#utf-8

@all @xen_017 @runner @run_settings
Feature: Run settings: environment, stop on error, variables, cookies and logs
  A run can use another environment (or none) without changing the active one. "Stop run if an
  error occurs" ends the run after a request that cannot be sent or a script error, never after a
  failed assertion. "Keep variable values" on persists script changes as a single send does; off,
  later rows see them and they are discarded at the end. A run starts from a copy of the stored
  cookies (or an empty jar), and its cookies reach the workspace jar only when saving is on.
  "Turn off logs during run" keeps the run out of the console and out of its results.

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
    And the 1 "POST" request to "/orders" returns status 201 with the following response
      """
      {"id":"o-1"}
      """

  @environment_override
  Scenario: A run can use another environment without changing the active one
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
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "{{stage}}" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the 1 "POST" request to "/login" should have the header "Authorization" equal to "Bearer t-stage"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "activeEnvironmentId" equal to "{{dev}}"
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "environment_name" equal to "stage"

  @environment_override
  Scenario: A run can use no environment at all
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "none" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the 1 "POST" request to "/login" should have the header "Authorization" equal to "Bearer {{token}}"

  @stop_on_error
  Scenario: Stop on error ends the run after a request that cannot be sent
    Given the "closed" variable is set to a closed port url
    And the workspace file "acme/collections/shop/orders/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{closed}}/orders"
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": true, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the last "runner:finished" event should contain the field "status" equal to "stopped_on_error"
    And the "GET" request to "/orders/1" should have been received 0 times
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows" array length equal to 2
    And the response should contain the field "rows.1.outcome" equal to "error"

  @stop_on_error
  Scenario: Stop on error ends the run after a script error
    Given the workspace file "acme/collections/shop/orders/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{api}}/orders"
      scripts:
        preRequest: throw new Error("broken script")
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": true, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the last "runner:finished" event should contain the field "status" equal to "stopped_on_error"
    And the "POST" request to "/orders" should have been received 0 times
    And the "GET" request to "/orders/1" should have been received 0 times
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.1.outcome" equal to "error"
    And the response should contain the field "rows.1.script_errors.0.message" containing "broken script"

  @stop_on_error
  Scenario: A failed assertion never stops the run, even with stop on error
    Given the 1 "POST" request to "/orders" returns status 500 with the following response
      """
      {}
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": true, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the last "runner:finished" event should contain the field "status" equal to "completed"
    And the "POST" request to "/logout" should have been received 1 times

  @without_stop_on_error
  Scenario: Without stop on error the run goes on after a request that cannot be sent
    Given the "closed" variable is set to a closed port url
    And the workspace file "acme/collections/shop/orders/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{closed}}/orders"
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the last "runner:finished" event should contain the field "status" equal to "completed"
    And the "POST" request to "/logout" should have been received 1 times

  @keep_values
  Scenario: With keep variable values on, script changes persist exactly as after a single send
    Given the workspace file "acme/collections/shop/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: "{{api}}/login"
      scripts:
        postResponse: |
          pm.environment.set("session", "sess-abc")
          pm.collectionVariables.set("cart", "cart-42")
      """
    And the workspace file "acme/collections/shop/orders/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{api}}/orders"
      body:
        mode: raw
        raw:
          language: text
          content: "{{session}}|{{cart}}"
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the 1 "POST" request to "/orders" should have the body equal to
      """
      sess-abc|cart-42
      """
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain "sess-abc"
    And the event "variables:changed" should have been emitted at most 10 times
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-create", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{api}}/orders", "body": { "mode": "raw", "raw": { "language": "text", "content": "{{session}}|{{cart}}" } } } }
      """
    Then the call should succeed
    And the 2 "POST" request to "/orders" should have the body equal to
      """
      sess-abc|cart-42
      """

  @keep_values_off
  Scenario: With keep variable values off, later rows see the changes and nothing outlives the run
    Given the workspace file "acme/collections/shop/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: "{{api}}/login"
      scripts:
        postResponse: |
          pm.environment.set("session", "sess-abc")
          pm.collectionVariables.set("cart", "cart-42")
      """
    And the workspace file "acme/collections/shop/orders/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{api}}/orders"
      body:
        mode: raw
        raw:
          language: text
          content: "{{session}}|{{cart}}"
      """
    And the workspace file "acme/collections/shop/logout.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{logout}}
      name: Logout
      method: POST
      url: "{{api}}/logout"
      scripts:
        preRequest: pm.vault.set("vaulted", "vault-xyz")
      """
    And I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": false, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the 1 "POST" request to "/orders" should have the body equal to
      """
      sess-abc|cart-42
      """
    And the event "variables:changed" should have been emitted 0 times
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "sess-abc"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "cart-42"
    And no file under the workspace folder should contain "sess-abc"
    And no file under the workspace folder should contain "cart"
    And the workspace file "acme/environments/dev.yaml" should not contain the field "variables.1"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-create", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{api}}/orders", "body": { "mode": "raw", "raw": { "language": "text", "content": "{{session}}|{{cart}}" } } } }
      """
    Then the 2 "POST" request to "/orders" should have the body equal to
      """
      {{session}}|{{cart}}
      """

  @stored_cookies
  Scenario: A run starts from a copy of the stored cookies
    Given the requests to "/login" respond with the header "Set-Cookie" equal to "session=abc; Path=/"
    When I send the saved request "{{login}}" in the workspace "Acme"
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{get}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the 1 "GET" request to "/orders/1" should have the header "Cookie" equal to "session=abc"

  @skip_stored_cookies
  Scenario: Without stored cookies a run starts empty, yet cookies set during the run are sent by later rows
    Given the requests to "/login" respond with the header "Set-Cookie" equal to "session=abc; Path=/"
    When I send the saved request "{{login}}" in the workspace "Acme"
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{get}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": true, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "first"
    And a "runner:finished" event with the field "run_id" equal to "{{first}}" should have been emitted within 5000 ms
    And the 1 "GET" request to "/orders/1" should not have the header "Cookie"
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": true, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "second"
    And a "runner:finished" event with the field "run_id" equal to "{{second}}" should have been emitted within 5000 ms
    And the 2 "GET" request to "/orders/1" should have the header "Cookie" equal to "session=abc"

  @save_cookies
  Scenario Outline: Cookies set during a run reach the workspace jar only when saving is on - save <save>
    Given the requests to "/login" respond with the header "Set-Cookie" equal to "session=abc; Path=/"
    And I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": <save> }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to <domains>

    Examples:
      | save  | domains |
      | true  | 1       |
      | false | 0       |

  @logs_off
  Scenario: Turning off logs keeps the run out of the console and out of its results
    Given I call "ConsoleService.Clear"
    And the emitted events are forgotten
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": true, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the event "console:entries" should have been emitted 0 times
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.0.console" array length equal to 0
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "all" }
      """
    Then the response should contain the field "" array length equal to 0

  @logs_on
  Scenario: With logs on, a run's console output reaches the app console as for a single send
    Given I call "ConsoleService.Clear"
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "0.summary" equal to "shop pre"
