#language: en
#utf-8

@all @xen_017 @runner @run_functional
Feature: Functional runs send the selected requests of a collection or folder in order
  A functional run sends the selected rows of a collection (or folder) in the shown order, each request
  as saved on disk, through the same pipeline as a single send: inherited auth and headers, the
  collection → folder → request script chain, then the post-response chain. Every pm.test result is
  collected per row. Runs are not recorded in History. The run continues in the background; the
  "runner:finished" event reports its end and RunnerService.Get returns its results.

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

  @run_order @success
  Scenario: A functional run sends every selected request in sidebar order with inherited auth, headers and scripts
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "POST /login, POST /orders, GET /orders/1, POST /logout" in order
    And the 1 "POST" request to "/login" should have the header "Authorization" equal to "Bearer t-123"
    And the 1 "GET" request to "/orders/1" should have the header "Authorization" equal to "Bearer t-123"
    And the 1 "POST" request to "/logout" should have the header "X-Suite" equal to "runner"
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "completed"
    And the response should contain the field "target.name" equal to "Shop"
    And the response should contain the field "rows" array length equal to 4
    And the response should contain the field "rows.0.name" equal to "Login"
    And the response should contain the field "rows.0.method" equal to "POST"
    And the response should contain the field "rows.0.url" equal to "{{api}}/login"
    And the response should contain the field "rows.0.status" equal to "200"
    And the response should contain the field "rows.0.outcome" equal to "passed"
    And the response should contain the field "rows.0.iteration" equal to "0"
    And the response should contain the field "rows.0.tests.0.name" equal to "login ok"
    And the response should contain the field "rows.0.tests.0.status" equal to "passed"
    And the response should contain the field "rows.0.console.0.summary" equal to "shop pre"
    And the response should contain the field "rows.0.sent_request.method" equal to "POST"
    And the response should contain the field "rows.0.sent_request.url" equal to "{{api}}/login"
    And the response should contain the field "rows.1.name" equal to "Create"
    And the response should contain the field "rows.1.folder_path" with the items "Orders"
    And the response should contain the field "rows.1.status" equal to "201"
    And the response should contain the field "rows.1.outcome" equal to "passed"
    And the response should contain the field "rows.2.name" equal to "Get"
    And the response should contain the field "rows.2.outcome" equal to "passed"
    And the response should contain the field "rows.3.name" equal to "Logout"
    And the response should contain the field "summary.iterations" equal to "1"
    And the response should contain the field "summary.total_requests" equal to "4"
    And the response should contain the field "summary.tests_passed" equal to "2"
    And the response should contain the field "summary.tests_failed" equal to "0"
    And the response should contain the field "summary.passed" equal to "4"
    And the response field "summary.avg_response_ms" should be between 0 and 5000
    And the response field "summary.duration_ms" should be between 0 and 5000

  @run_order @selection
  Scenario: Only selected rows run, in the order given
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{logout}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": false }, { "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": false }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "POST /logout, POST /login" in order
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows" array length equal to 2
    And the response should contain the field "rows.0.name" equal to "Logout"
    And the response should contain the field "rows.1.name" equal to "Login"
    And the response should contain the field "summary.total_requests" equal to "2"

  @saved_not_draft
  Scenario: A run sends the saved request, never a tab's unsaved draft
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-login", "request": { "id": "{{login}}", "name": "Login", "method": "PUT", "url": "{{api}}/draft-login" } }
      """
    Then the call should succeed
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the "PUT" request to "/draft-login" should have been received 1 times
    And the "POST" request to "/login" should have been received 1 times
    And the workspace file "acme/collections/shop/login.yaml" should contain the field "url" equal to "{{api}}/login"

  @failed_test
  Scenario: A failed test marks the row failed and the run goes on
    Given the 1 "POST" request to "/orders" returns status 500 with the following response
      """
      {"error":"boom"}
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the "POST" request to "/logout" should have been received 1 times
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "completed"
    And the response should contain the field "rows.1.outcome" equal to "failed"
    And the response should contain the field "rows.1.tests.0.name" equal to "created"
    And the response should contain the field "rows.1.tests.0.status" equal to "failed"
    And the response should contain the field "rows.1.tests.0.error" containing "500"
    And the response should contain the field "summary.tests_failed" equal to "1"
    And the response should contain the field "summary.failed" equal to "1"

  @skip_request
  Scenario: pm.execution.skipRequest skips the row and reports it as skipped
    Given the workspace file "acme/collections/shop/orders/get.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{get}}
      name: Get
      method: GET
      url: "{{api}}/orders/1"
      auth:
        type: inherit
      scripts:
        preRequest: pm.execution.skipRequest()
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the "GET" request to "/orders/1" should have been received 0 times
    And the "POST" request to "/logout" should have been received 1 times
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.2.outcome" equal to "skipped"
    And the response should contain the field "rows.2.status" equal to "0"
    And the response should contain the field "summary.skipped" equal to "1"

  @no_history
  Scenario: Runner sends are never recorded in History
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the event "history:changed" should have been emitted 0 times
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 0

  @alias_row
  Scenario: A row with an alias chosen sends that alias
    Given the "mx" variable is set to "01JR0000000000000000000020"
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
      aliases:
        - id: {{mx}}
          name: mx-test-user
          method: POST
          url: "{{api}}/login-mx"
          auth:
            type: inherit
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "{{mx}}", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "POST /login-mx, POST /logout" in order
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.0.alias_id" equal to "{{mx}}"
    And the response should contain the field "rows.0.alias_name" equal to "mx-test-user"
    And the response should contain the field "rows.1.alias_name" equal to ""

  @missing_request
  Scenario: A request deleted before its turn is reported as a row error and the run goes on
    Given the "POST" requests to "/orders" respond after 800ms
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    Given the workspace path "acme/collections/shop/orders/get.yaml" is removed
    Then a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the "GET" request to "/orders/1" should have been received 0 times
    And the "POST" request to "/logout" should have been received 1 times
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "completed"
    And the response should contain the field "rows.2.outcome" equal to "error"
    And the response should contain the field "rows.2.error" containing "not found"
    And the response should contain the field "summary.errors" equal to "1"

  @folder_target
  Scenario: A folder runs only its own requests, still inheriting the collection's auth, headers and scripts
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{orders}}", "config": { "type": "functional", "rows": [{ "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "POST /orders, GET /orders/1" in order
    And the 1 "POST" request to "/orders" should have the header "Authorization" equal to "Bearer t-123"
    And the 1 "POST" request to "/orders" should have the header "X-Suite" equal to "runner"
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "target.kind" equal to "folder"
    And the response should contain the field "target.name" equal to "Orders"
    And the response should contain the field "rows.0.console.0.summary" equal to "shop pre"
