#language: en
#utf-8

@all @xen_017 @runner @runner_plan
Feature: The Runner lists a collection's requests and remembers the selection, order and settings
  RunnerService.Plan lists every request under a collection or folder (subfolders included, saved
  examples excluded) in sidebar order, numbered, with the folder path relative to the target. The
  selection, order, chosen aliases and run settings are remembered per collection or folder in the
  workspace's local UI state (state.json in the app-data folder), never in the workspace files. The
  remembered rows are reconciled with the tree: removed requests drop out, new ones join selected
  after their preceding sidebar sibling, and an alias that no longer exists falls back to Default.

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

  @plan
  Scenario: The plan lists every request in sidebar order with folder path, method and name, all selected, with Postman's defaults
    Given the workspace file "acme/collections/shop/orders/get.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{get}}
      name: Get
      method: GET
      url: "{{api}}/orders/1"
      examples:
        - id: 01JR0000000000000000000050
          name: Found
          request:
            method: GET
            url: "{{api}}/orders/1"
          response:
            code: 200
            status: OK
            body: "{}"
      """
    When I call "RunnerService.Plan" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the call should succeed
    And the response should contain the field "target.item_id" equal to "{{shop}}"
    And the response should contain the field "target.kind" equal to "collection"
    And the response should contain the field "target.name" equal to "Shop"
    And the response should contain the field "rows" array length equal to 4
    And the response should contain the field "rows.0.number" equal to "1"
    And the response should contain the field "rows.0.request_id" equal to "{{login}}"
    And the response should contain the field "rows.0.name" equal to "Login"
    And the response should contain the field "rows.0.method" equal to "POST"
    And the response should contain the field "rows.0.folder_path" array length equal to 0
    And the response should contain the field "rows.0.selected" equal to "true"
    And the response should contain the field "rows.0.alias_id" equal to ""
    And the response should contain the field "rows.1.request_id" equal to "{{create}}"
    And the response should contain the field "rows.1.folder_path" with the items "Orders"
    And the response should contain the field "rows.2.request_id" equal to "{{get}}"
    And the response should contain the field "rows.2.number" equal to "3"
    And the response should contain the field "rows.3.request_id" equal to "{{logout}}"
    And the response should contain the field "config.type" equal to "functional"
    And the response should contain the field "config.functional.iterations" equal to "1"
    And the response should contain the field "config.functional.delay_ms" equal to "0"
    And the response should contain the field "config.functional.keep_variable_values" equal to "true"
    And the response should contain the field "config.functional.persist_responses" equal to "false"
    And the response should contain the field "config.functional.stop_on_error" equal to "false"
    And the response should contain the field "config.functional.save_cookies" equal to "false"
    And the response should contain the field "config.performance.virtual_users" equal to "10"
    And the response should contain the field "config.performance.duration_seconds" equal to "600"
    And the response should contain the field "config.performance.profile.kind" equal to "fixed"
    And the response should contain the field "active_environment_id" equal to "{{dev}}"
    And the response field "environments" should contain an item with "id" equal to "{{dev}}" and "name" equal to "dev"
    And the response should contain the field "data_file_missing" equal to "false"

  @folder_plan
  Scenario: A folder's plan lists only its requests, with paths relative to the folder
    When I call "RunnerService.Plan" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{orders}}" }
      """
    Then the call should succeed
    And the response should contain the field "target.kind" equal to "folder"
    And the response should contain the field "rows" array length equal to 2
    And the response should contain the field "rows.0.request_id" equal to "{{create}}"
    And the response should contain the field "rows.0.folder_path" array length equal to 0
    And the response should contain the field "rows.1.request_id" equal to "{{get}}"

  @preferences
  Scenario: Selection, order and settings are remembered locally and never written to the workspace
    When I call "RunnerService.SavePreferences" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{logout}}", "alias_id": "", "selected": true }, { "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": false }, { "request_id": "{{get}}", "alias_id": "", "selected": true }], "functional": { "iterations": 3, "data_file": "", "delay_ms": 250, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": false, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "none" } }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should exist
    And no file under the workspace folder should contain "keep_variable_values"
    When I call "RunnerService.Plan" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.0.request_id" equal to "{{logout}}"
    And the response should contain the field "rows.0.number" equal to "1"
    And the response should contain the field "rows.1.request_id" equal to "{{login}}"
    And the response should contain the field "rows.2.request_id" equal to "{{create}}"
    And the response should contain the field "rows.2.selected" equal to "false"
    And the response should contain the field "rows.3.request_id" equal to "{{get}}"
    And the response should contain the field "config.functional.iterations" equal to "3"
    And the response should contain the field "config.functional.delay_ms" equal to "250"
    And the response should contain the field "config.functional.keep_variable_values" equal to "false"
    And the response should contain the field "config.environment_id" equal to "none"
    When the app is restarted
    And I call "RunnerService.Plan" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the response should contain the field "rows.0.request_id" equal to "{{logout}}"

  @preferences
  Scenario: Saving the UI state keeps the runner preferences
    Given I call "RunnerService.SavePreferences" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    When I call "UIStateService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "tabs": [], "active_tab_id": "" }
      """
    Then the call should succeed
    When I call "RunnerService.Plan" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the response should contain the field "rows.0.request_id" equal to "{{logout}}"

  @reconcile
  Scenario: Removed requests drop out and new requests join, selected, after their preceding sidebar sibling
    Given the "refund" variable is set to "01JR0000000000000000000007"
    And I call "RunnerService.SavePreferences" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{logout}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": false }, { "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    And the workspace path "acme/collections/shop/logout.yaml" is removed
    And the workspace file "acme/collections/shop/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{shop}}
      name: Shop
      order:
        - login.yaml
        - orders
      """
    And the workspace file "acme/collections/shop/orders/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{orders}}
      name: Orders
      order:
        - create.yaml
        - refund.yaml
        - get.yaml
      """
    And the workspace file "acme/collections/shop/orders/refund.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{refund}}
      name: Refund
      method: POST
      url: "{{api}}/refund"
      """
    When I call "RunnerService.Plan" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows" array length equal to 4
    And the response should contain the field "rows.0.request_id" equal to "{{get}}"
    And the response should contain the field "rows.0.selected" equal to "false"
    And the response should contain the field "rows.1.request_id" equal to "{{login}}"
    And the response should contain the field "rows.2.request_id" equal to "{{create}}"
    And the response should contain the field "rows.3.request_id" equal to "{{refund}}"
    And the response should contain the field "rows.3.selected" equal to "true"
    And the response should contain the field "rows.3.number" equal to "4"

  @reconcile
  Scenario: A new request whose preceding siblings are all absent joins first
    Given the "first" variable is set to "01JR0000000000000000000008"
    And I call "RunnerService.SavePreferences" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{logout}}", "alias_id": "", "selected": true }, { "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    And the workspace file "acme/collections/shop/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{shop}}
      name: Shop
      order:
        - first.yaml
        - login.yaml
        - orders
        - logout.yaml
      """
    And the workspace file "acme/collections/shop/first.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{first}}
      name: First
      method: GET
      url: "{{api}}/first"
      """
    When I call "RunnerService.Plan" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the response should contain the field "rows" array length equal to 5
    And the response should contain the field "rows.0.request_id" equal to "{{first}}"
    And the response should contain the field "rows.1.request_id" equal to "{{logout}}"

  @alias_preference
  Scenario: A remembered alias is kept, and falls back to the Default once the alias is deleted
    Given the "mx" variable is set to "01JR0000000000000000000020"
    And the workspace file "acme/collections/shop/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: "{{api}}/login"
      aliases:
        - id: {{mx}}
          name: mx-test-user
          method: POST
          url: "{{api}}/login-mx"
      """
    And I call "RunnerService.SavePreferences" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "{{mx}}", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    When I call "RunnerService.Plan" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the response should contain the field "rows.0.alias_id" equal to "{{mx}}"
    And the response field "rows.0.aliases" should contain an item with "id" equal to "{{mx}}" and "name" equal to "mx-test-user"
    Given the workspace file "acme/collections/shop/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: "{{api}}/login"
      """
    When I call "RunnerService.Plan" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the response should contain the field "rows.0.alias_id" equal to ""

  @missing_data_file
  Scenario: A remembered data file that no longer exists is cleared and reported
    Given I call "RunnerService.SavePreferences" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [], "functional": { "iterations": 1, "data_file": "{{temp_root}}/gone/users.csv", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    When I call "RunnerService.Plan" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the call should succeed
    And the response should contain the field "data_file_missing" equal to "true"
    And the response should contain the field "config.functional.data_file" equal to ""
    When I call "RunnerService.Plan" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}" }
      """
    Then the response should contain the field "data_file_missing" equal to "false"

  @unknown_target
  Scenario: The plan of an unknown collection or folder is refused
    When I call "RunnerService.Plan" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "01JR0000000000000000009999" }
      """
    Then the call should fail with code "RUN-07404"
