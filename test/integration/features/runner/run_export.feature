#language: en
#utf-8

@all @xen_017 @runner @run_export
Feature: A run's results export as JSON or JUnit XML
  JSON exports the stored run as is (secrets masked). JUnit XML has a <testsuites> root, one
  <testsuite> per executed row and iteration (named "<folder path>/<request> [iteration N]", with
  tests, failures, errors, skipped and time), one <testcase> per pm.test, <failure> for failed tests,
  <skipped/> for skipped ones and an <error> case for a row that could not be sent.

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

  @json
  Scenario: The JSON export is the stored run
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    When I call "RunnerService.Export" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}", "format": "json", "path": "{{temp_root}}/out/run.json" }
      """
    Then the call should succeed
    And the file "out/run.json" should contain the field "id" equal to "{{run}}"
    And the file "out/run.json" should contain the field "status" equal to "completed"
    And the file "out/run.json" should contain the field "rows" array length equal to 4
    And the file "out/run.json" should contain the field "summary.total_requests" equal to "4"

  @json_masked
  Scenario: The JSON export never contains a secret value
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
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    When I call "RunnerService.Export" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}", "format": "json", "path": "{{temp_root}}/out/run.json" }
      """
    Then the call should succeed
    And the file "out/run.json" should not contain "s3cr3t-key"
    And the file "out/run.json" should contain "•••••"

  @junit
  Scenario: The JUnit export has a suite per row with a case per test, failures, skips and errors
    Given the "closed" variable is set to a closed port url
    And the workspace file "acme/collections/shop/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: "{{api}}/login"
      scripts:
        postResponse: |
          pm.test("login ok", function () { pm.response.to.have.status(200); });
          pm.test.skip("later", function () {});
      """
    And the 1 "POST" request to "/orders" returns status 500 with the following response
      """
      {}
      """
    And the workspace file "acme/collections/shop/logout.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{logout}}
      name: Logout
      method: POST
      url: "{{closed}}/logout"
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{create}}", "alias_id": "", "selected": true }, { "request_id": "{{get}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    When I call "RunnerService.Export" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}", "format": "junit", "path": "{{temp_root}}/out/run.xml" }
      """
    Then the call should succeed
    And the file "out/run.xml" should contain "<?xml"
    And the file "out/run.xml" should contain "<testsuites"
    And the file "out/run.xml" should contain "<testsuite name="
    And the file "out/run.xml" should contain "Login [iteration 1]"
    And the file "out/run.xml" should contain "Orders/Create [iteration 1]"
    And the file "out/run.xml" should contain "<testcase name="
    And the file "out/run.xml" should contain "classname="
    And the file "out/run.xml" should contain "<failure message="
    And the file "out/run.xml" should contain "<skipped"
    And the file "out/run.xml" should contain "<error message="
    And the file "out/run.xml" should contain "failures="
    And the file "out/run.xml" should contain "errors="
    And the file "out/run.xml" should contain "skipped="
    And the file "out/run.xml" should contain "time="

  @unknown_run
  Scenario: Exporting an unknown run is refused
    When I call "RunnerService.Export" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "01JR0000000000000000009999", "format": "json", "path": "{{temp_root}}/out/run.json" }
      """
    Then the call should fail with code "RUN-01404"
    And the file "out/run.json" should not exist
