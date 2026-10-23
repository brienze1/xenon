#language: en
#utf-8

@all @xen_017 @runner @run_lifecycle
Feature: A run reports progress and can be paused, resumed and stopped
  While a run is active the Runner receives "runner:progress" after every row and "runner:finished"
  once at the end. Pause holds the run once the request in flight completes; Resume continues with
  the next row; Stop cancels the request in flight (that row is "cancelled") and ends the run as
  "stopped". Only one run may be active per collection or folder; runs on different targets and
  single sends from tabs go on in parallel. Switching workspace stops its active runs.

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

  @progress
  Scenario: Progress is published after every row
    Given the "POST" requests to "/login" respond after 300ms
    And the "POST" requests to "/orders" respond after 300ms
    And the "GET" requests to "/orders/1" respond after 300ms
    And the "POST" requests to "/logout" respond after 300ms
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And the response should contain the field "status" equal to "running"
    And a "runner:progress" event with the field "progress.completed" equal to "1" should have been emitted within 1000 ms
    And a "runner:progress" event with the field "progress.completed" equal to "2" should have been emitted within 1000 ms
    And a "runner:progress" event with the field "progress.completed" equal to "3" should have been emitted within 1000 ms
    And a "runner:progress" event with the field "progress.completed" equal to "4" should have been emitted within 1000 ms
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 2000 ms
    And the last "runner:progress" event should contain the field "run_id" equal to "{{run}}"
    And the last "runner:progress" event should contain the field "item_id" equal to "{{shop}}"
    And the last "runner:progress" event should contain the field "progress.planned" equal to "4"
    And the last "runner:finished" event should contain the field "item_id" equal to "{{shop}}"
    And the last "runner:finished" event should contain the field "target_name" equal to "Shop"
    And the last "runner:finished" event should contain the field "type" equal to "functional"
    And the last "runner:finished" event should contain the field "status" equal to "completed"
    And the last "runner:finished" event should contain the field "summary.total_requests" equal to "4"

  @pause_resume
  Scenario: Pause holds the run after the request in flight, resume continues with the next row
    Given the "POST" requests to "/orders" respond after 800ms
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    When I wait 300 ms
    And I call "RunnerService.Pause" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And a "runner:progress" event with the field "status" equal to "paused" should have been emitted within 2000 ms
    And the "POST" request to "/orders" should have been received 1 times
    When I wait 500 ms
    Then the "GET" request to "/orders/1" should have been received 0 times
    When I call "RunnerService.Resume" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the "GET" request to "/orders/1" should have been received 1 times
    And the last "runner:finished" event should contain the field "status" equal to "completed"

  @stop
  Scenario: Stop cancels the request in flight and ends the run as stopped
    Given the "POST" requests to "/orders" respond after 3000ms
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    When I wait 300 ms
    And I call "RunnerService.Stop" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 2000 ms
    And the last "runner:finished" event should contain the field "status" equal to "stopped"
    And the "GET" request to "/orders/1" should have been received 0 times
    And the "POST" request to "/logout" should have been received 0 times
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "stopped"
    And the response should contain the field "rows" array length equal to 2
    And the response should contain the field "rows.1.outcome" equal to "cancelled"

  @one_run_per_target
  Scenario: A second run on the same collection is refused while one is active
    Given the "POST" requests to "/login" respond after 1000ms
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should fail with code "RUN-02409"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the "POST" request to "/login" should have been received 1 times

  @overlap
  Scenario: Runs on different collections overlap, and tabs keep sending during a run
    Given the "admin" variable is set to "01JR0000000000000000000030"
    And the "ping" variable is set to "01JR0000000000000000000031"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JR0000000000000000000000
      name: Acme
      order:
        - shop
        - admin
      """
    And the workspace file "acme/collections/admin/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{admin}}
      name: Admin
      order:
        - ping.yaml
      """
    And the workspace file "acme/collections/admin/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{ping}}
      name: Ping
      method: GET
      url: "{{api}}/ping"
      """
    And the "POST" requests to "/login" respond after 1000ms
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{admin}}", "config": { "type": "functional", "rows": [{ "request_id": "{{ping}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "other"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/health" } }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "200"
    And a "runner:finished" event with the field "run_id" equal to "{{other}}" should have been emitted within 2000 ms
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the "GET" request to "/ping" should have been received 1 times
    And the "POST" request to "/logout" should have been received 1 times

  @workspace_switch
  Scenario: Activating another workspace stops the active run and stores it as stopped
    Given the workspace "Globex" exists
    And the "POST" requests to "/orders" respond after 3000ms
    And I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    When I wait 300 ms
    And I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Globex.local_id}}" }
      """
    Then a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 2000 ms
    And the last "runner:finished" event should contain the field "status" equal to "stopped"
    And the "POST" request to "/logout" should have been received 0 times
    And the app data file "workspaces/{{Acme.local_id}}/runs/{{shop}}/{{run}}.json" should contain the field "status" equal to "stopped"

  @active
  Scenario: Active runs of a workspace are listed until they finish
    Given the "POST" requests to "/login" respond after 800ms
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    When I call "RunnerService.Active" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 1
    And the response should contain the field "0.id" equal to "{{run}}"
    And the response should contain the field "0.status" equal to "running"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    When I call "RunnerService.Active" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0
