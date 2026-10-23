#language: en
#utf-8

@all @xen_017 @runner @run_iterations
Feature: Iterations, iteration data and the delay between requests
  A run repeats the selection once per iteration. pm.info.iteration (0-based) and
  pm.info.iterationCount describe it in every script. An iteration data file (CSV with a header row,
  or a JSON array of objects, read as UTF-8) gives each iteration its row through pm.iterationData
  and {{field}}, which ranks above the environment and below local variables. Fewer iterations use
  the first rows; more reuse the last one. The delay elapses between rows, across iterations, but
  not after the last row.

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

  @csv
  Scenario: Each iteration uses its CSV row through {{field}}, pm.iterationData and pm.info
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
          content: "{{email}}|{{password}}"
      scripts:
        preRequest: console.log(pm.info.iteration, pm.info.iterationCount, pm.iterationData.get("email"))
      """
    And the file "data/users.csv" exists with content
      """
      email,password
      a@x.test,pa
      b@x.test,pb
      c@x.test,pc
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 3, "data_file": "{{temp_root}}/data/users.csv", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the "POST" request to "/login" should have been received 3 times
    And the 1 "POST" request to "/login" should have the body equal to
      """
      a@x.test|pa
      """
    And the 2 "POST" request to "/login" should have the body equal to
      """
      b@x.test|pb
      """
    And the 3 "POST" request to "/login" should have the body equal to
      """
      c@x.test|pc
      """
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "data_file_name" equal to "users.csv"
    And the response should contain the field "summary.iterations" equal to "3"
    And the response should contain the field "rows" array length equal to 3
    And the response should contain the field "rows.0.iteration" equal to "0"
    And the response should contain the field "rows.2.iteration" equal to "2"
    And the response should contain the field "rows.0.console.1.summary" equal to "0 3 a@x.test"
    And the response should contain the field "rows.2.console.1.summary" equal to "2 3 c@x.test"

  @precedence
  Scenario: Iteration data overrides the environment and loses to local variables
    Given the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: email
          value: env@x.test
          enabled: true
          description: ""
        - key: password
          value: env-pass
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
      url: "{{api}}/login"
      body:
        mode: raw
        raw:
          language: text
          content: "{{email}}|{{password}}"
      scripts:
        preRequest: pm.variables.set("password", "local-pass")
      """
    And the file "data/users.csv" exists with content
      """
      email,password
      data@x.test,data-pass
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "{{temp_root}}/data/users.csv", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the 1 "POST" request to "/login" should have the body equal to
      """
      data@x.test|local-pass
      """

  @csv_shape
  Scenario: A byte-order mark is ignored, extra cells are dropped and missing cells are empty
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
          content: "{{email}}|{{password}}"
      scripts:
        preRequest: console.log(pm.info.iteration, pm.info.iterationCount, pm.iterationData.get("email"))
      """
    And the file "data/users.csv" exists with content
      """
      ﻿email,password
      a@x.test,pa,extra
      b@x.test
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 2, "data_file": "{{temp_root}}/data/users.csv", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the 1 "POST" request to "/login" should have the body equal to
      """
      a@x.test|pa
      """
    And the 2 "POST" request to "/login" should have the body equal to
      """
      b@x.test|
      """

  @json_data
  Scenario: JSON values keep their types in pm.iterationData and their text form in {{field}}
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
          content: "{{n}}|{{ok}}|{{none}}|{{obj}}"
      scripts:
        preRequest: console.log(typeof pm.iterationData.get("n"), typeof pm.iterationData.get("ok"), pm.iterationData.get("obj").a)
      """
    And the file "data/rows.json" exists with content
      """
      [{ "n": 1.5, "ok": true, "none": null, "obj": { "a": 1 } }]
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "{{temp_root}}/data/rows.json", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the 1 "POST" request to "/login" should have the body equal to
      """
      1.5|true||{"a":1}
      """
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.0.console.1.summary" equal to "number boolean 1"

  @rows_vs_iterations
  Scenario Outline: Fewer iterations use the first rows, more iterations reuse the last row - <iterations> iterations
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
          content: "{{email}}|{{password}}"
      scripts:
        preRequest: console.log(pm.info.iteration, pm.info.iterationCount, pm.iterationData.get("email"))
      """
    And the file "data/users.csv" exists with content
      """
      email,password
      a@x.test,pa
      b@x.test,pb
      c@x.test,pc
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": <iterations>, "data_file": "{{temp_root}}/data/users.csv", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the "POST" request to "/login" should have been received <iterations> times
    And the <iterations> "POST" request to "/login" should have the body equal to
      """
      <last>
      """

    Examples:
      | iterations | last        |
      | 2          | b@x.test\|pb |
      | 5          | c@x.test\|pc |

  @invalid_data_file
  Scenario Outline: A data file that cannot be used refuses the start - <case>
    Given the file "data/<file>" exists with content
      """
      <content>
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "{{temp_root}}/data/<target>", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should fail with code "RUN-05400"
    And the "POST" request to "/login" should have been received 0 times

    Examples:
      | case                 | file       | target       | content           |
      | empty CSV            | empty.csv  | empty.csv    |                   |
      | JSON object          | obj.json   | obj.json     | {"email":"a"}     |
      | JSON array of values | vals.json  | vals.json    | [1, 2]            |
      | missing file         | users.csv  | missing.csv  | email             |

  @delay
  Scenario: The delay separates every row, across iterations, and never follows the last row
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{shop}}", "config": { "type": "functional", "rows": [{ "request_id": "{{login}}", "alias_id": "", "selected": true }, { "request_id": "{{logout}}", "alias_id": "", "selected": true }], "functional": { "iterations": 2, "data_file": "", "delay_ms": 300, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "POST /login, POST /logout, POST /login, POST /logout" in order
    And the 1 "POST" request to "/logout" should have been received at least 300 ms after the 1 "POST" request to "/login"
    And the 2 "POST" request to "/login" should have been received at least 300 ms after the 1 "POST" request to "/logout"
    And the 2 "POST" request to "/logout" should have been received at least 300 ms after the 2 "POST" request to "/login"
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response field "summary.duration_ms" should be between 900 and 1180
