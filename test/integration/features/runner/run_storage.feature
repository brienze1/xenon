#language: en
#utf-8

@all @xen_017 @runner @run_storage
Feature: Finished runs are stored locally, capped, masked and deletable
  Every finished, stopped or failed run is stored as JSON under the workspace's app-data folder
  (runs/<collection or folder id>/<run id>.json), never in the workspace, never in History. The 25
  most recent runs are kept per collection or folder. Every secret value (vault secrets and secrets
  set during the run) is masked as ••••• in URLs, headers, bodies, console output, test names and
  messages and script errors; response headers and bodies are never stored. Runs can be deleted one
  by one or cleared, are removed with their workspace or when their folder is gone, and an
  unreadable file is skipped.

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

  @stored
  Scenario: A finished run is stored in the app-data folder, never in the workspace
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the app data file "workspaces/{{Acme.local_id}}/runs/{{shop}}/{{run}}.json" should exist
    And the app data file "workspaces/{{Acme.local_id}}/runs/{{shop}}/{{run}}.json" should contain the field "id" equal to "{{run}}"
    And the app data file "workspaces/{{Acme.local_id}}/runs/{{shop}}/{{run}}.json" should contain the field "status" equal to "completed"
    And no file under the workspace folder should contain "{{run}}"
    When I call "RunnerService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 1
    And the response should contain the field "0.id" equal to "{{run}}"
    And the response should contain the field "0.type" equal to "functional"
    And the response should contain the field "0.status" equal to "completed"
    And the response should contain the field "0.environment_name" equal to "dev"
    And the response should contain the field "0.summary.total_requests" equal to "4"
    And the response should contain the field "0.summary.tests_passed" equal to "2"
    And the response should contain the field "0.started_at" matching "^\d{4}-\d{2}-\d{2}T"
    And the response field "0.summary.duration_ms" should be between 0 and 10000

  @newest_first
  Scenario: Stored runs are listed newest first
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "older"
    And a "runner:finished" event with the field "run_id" equal to "{{older}}" should have been emitted within 5000 ms
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "newer"
    And a "runner:finished" event with the field "run_id" equal to "{{newer}}" should have been emitted within 5000 ms
    When I call "RunnerService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "0.id" equal to "{{newer}}"
    And the response should contain the field "1.id" equal to "{{older}}"

  @cap
  Scenario: Only the 25 most recent runs of a collection are kept, and a folder keeps its own runs
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "oldest"
    And a "runner:finished" event with the field "run_id" equal to "{{oldest}}" should have been emitted within 5000 ms
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{orders}}", "config": { "type": "functional", "rows": [{ "request_id": "{{create}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "folder"
    And a "runner:finished" event with the field "run_id" equal to "{{folder}}" should have been emitted within 5000 ms
    When I call "RunnerService.Start" 25 times waiting for a "runner:finished" event after each call with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    And I call "RunnerService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the response should contain the field "" array length equal to 25
    And the app data file "workspaces/{{Acme.local_id}}/runs/{{shop}}/{{oldest}}.json" should not exist
    When I call "RunnerService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{orders}}" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.id" equal to "{{folder}}"

  @masking
  Scenario: Every secret value is masked in the stored run, including secrets set during the run
    Given the keychain is available
    And I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "s3cr3t-key", "enabled": true } ] }
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
        - key: minted
          secret: true
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/shop/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: "{{api}}/login?k={{api_key}}"
      headers:
        - key: X-Api-Key
          value: "{{api_key}}"
      body:
        mode: raw
        raw:
          language: text
          content: "key={{api_key}}"
      scripts:
        preRequest: console.log("using", pm.variables.get("api_key"))
        postResponse: |
          pm.environment.set("minted", "mint-777")
          pm.test("checks " + pm.variables.get("api_key"), function () { throw new Error("bad " + pm.variables.get("api_key")); });
          throw new Error("script saw " + pm.variables.get("api_key"))
      """
    And the workspace file "acme/collections/shop/logout.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{logout}}
      name: Logout
      method: POST
      url: "{{api}}/logout"
      headers:
        - key: X-Minted
          value: "{{minted}}"
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "acknowledge_missing_secrets": true, "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the 1 "POST" request to "/login" should have the header "X-Api-Key" equal to "s3cr3t-key"
    And the 1 "POST" request to "/logout" should have the header "X-Minted" equal to "mint-777"
    And no file under the app data folder should contain "s3cr3t-key"
    And no file under the app data folder should contain "mint-777"
    And the app data file "workspaces/{{Acme.local_id}}/runs/{{shop}}/{{run}}.json" should contain "•••••"
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should not contain "s3cr3t-key"
    And the response should not contain "mint-777"
    And the response should contain the field "rows.0.url" containing "•••••"
    And the response should contain the field "rows.0.sent_request.body" equal to "key=•••••"
    And the response field "rows.0.sent_request.headers" should contain an item with "key" equal to "X-Api-Key" and "value" equal to "•••••"
    And the response field "rows.1.sent_request.headers" should contain an item with "key" equal to "X-Minted" and "value" equal to "•••••"
    And the response should contain the field "rows.0.tests.0.name" equal to "checks •••••"
    And the response should contain the field "rows.0.tests.0.error" containing "bad •••••"
    And the response should contain the field "rows.0.script_errors.0.message" containing "script saw •••••"
    And the response should contain the field "rows.0.console" containing "using •••••"

  @no_response_stored
  Scenario: Response headers and bodies are never stored
    Given the 1 "POST" request to "/login" returns status 200 with the following response
      """
      {"marker":"response-body-only"}
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": true, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the app data file "workspaces/{{Acme.local_id}}/runs/{{shop}}/{{run}}.json" should not contain "response-body-only"

  @delete
  Scenario: A stored run can be deleted; an unknown run cannot
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    When I call "RunnerService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/runs/{{shop}}/{{run}}.json" should not exist
    When I call "RunnerService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the response should contain the field "" array length equal to 0
    When I call "RunnerService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should fail with code "RUN-01404"
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should fail with code "RUN-01404"

  @delete_active
  Scenario: An active run cannot be deleted
    Given the "POST" requests to "/login" respond after 1000ms
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    When I call "RunnerService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should fail with code "RUN-02409"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms

  @clear
  Scenario: Clear runs deletes every stored run of the collection only
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "one"
    And a "runner:finished" event with the field "run_id" equal to "{{one}}" should have been emitted within 5000 ms
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "two"
    And a "runner:finished" event with the field "run_id" equal to "{{two}}" should have been emitted within 5000 ms
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{orders}}", "config": { "type": "functional", "rows": [{ "request_id": "{{create}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "folder"
    And a "runner:finished" event with the field "run_id" equal to "{{folder}}" should have been emitted within 5000 ms
    When I call "RunnerService.Clear" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/runs/{{shop}}/{{one}}.json" should not exist
    And the app data file "workspaces/{{Acme.local_id}}/runs/{{shop}}/{{two}}.json" should not exist
    And the app data file "workspaces/{{Acme.local_id}}/runs/{{orders}}/{{folder}}.json" should exist
    When I call "RunnerService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the response should contain the field "" array length equal to 0

  @workspace_isolation
  Scenario: Each workspace lists only its own runs
    Given the workspace "Globex" exists
    And the workspace file "globex/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JR0000000000000000000100
      name: Globex
      order:
        - shop
      """
    And the workspace file "globex/collections/shop/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{shop}}
      name: Shop
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    When I call "RunnerService.List" with
      """
      { "local_id": "{{Globex.local_id}}", "item_id": "{{shop}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 0

  @workspace_removed
  Scenario: Removing a workspace deletes its stored runs
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    When I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Acme.local_id}}", "move_to_trash": false }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/runs/{{shop}}/{{run}}.json" should not exist

  @orphans
  Scenario: Runs of a folder that no longer exists are deleted when the runner data is read
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{orders}}", "config": { "type": "functional", "rows": [{ "request_id": "{{create}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the app data file "workspaces/{{Acme.local_id}}/runs/{{orders}}/{{run}}.json" should exist
    Given the workspace path "acme/collections/shop/orders" is removed
    And the workspace file "acme/collections/shop/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{shop}}
      name: Shop
      order:
        - login.yaml
        - logout.yaml
      """
    And I call "RunnerService.Plan" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/runs/{{orders}}/{{run}}.json" should not exist

  @unreadable
  Scenario: An unreadable stored run file is left out of the list
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    Given the app data file "workspaces/{{Acme.local_id}}/runs/{{shop}}/01JR0000000000000000000099.json" contains
      """
      this is not a run
      """
    When I call "RunnerService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 1
    And the response should contain the field "0.id" equal to "{{run}}"
