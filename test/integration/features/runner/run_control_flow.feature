#language: en
#utf-8

@all @xen_017 @runner @run_control_flow
Feature: Scripts steer a run with pm.execution.setNextRequest
  Inside a functional run pm.execution.setNextRequest (and the legacy postman.setNextRequest) chooses
  the next row when the current one finishes: by request id, else by the first selected row with
  that name. Calls from pre-request and post-response scripts count, including in a row skipped
  with skipRequest, and the last call wins. null, or a target that is not a selected row, ends the
  iteration (the latter with a console warning); the next iteration starts at row 1. Outside a run
  the call stays a no-op.

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
    And the "a" variable is set to "01JR0000000000000000000041"
    And the "b" variable is set to "01JR0000000000000000000042"
    And the "c" variable is set to "01JR0000000000000000000043"
    And the "d" variable is set to "01JR0000000000000000000044"
    And the workspace file "acme/collections/flow/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JR0000000000000000000040
      name: Flow
      order:
        - a.yaml
        - b.yaml
        - c.yaml
        - d.yaml
      """
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JR0000000000000000000000
      name: Acme
      order:
        - shop
        - flow
      """
    And the workspace file "acme/collections/flow/b.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{b}}
      name: B
      method: GET
      url: "{{api}}/b"
      """
    And the workspace file "acme/collections/flow/c.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{c}}
      name: C
      method: GET
      url: "{{api}}/c"
      """
    And the workspace file "acme/collections/flow/d.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{d}}
      name: D
      method: GET
      url: "{{api}}/d"
      """

  @by_name
  Scenario: setNextRequest by name jumps to that row
    Given the workspace file "acme/collections/flow/a.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{a}}
      name: A
      method: GET
      url: "{{api}}/a"
      scripts:
        preRequest: ""
        postResponse: pm.execution.setNextRequest('C')
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "01JR0000000000000000000040", "config": { "type": "functional", "rows": [{ "request_id": "{{a}}", "alias_id": "", "selected": true }, { "request_id": "{{b}}", "alias_id": "", "selected": true }, { "request_id": "{{c}}", "alias_id": "", "selected": true }, { "request_id": "{{d}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "GET /a, GET /c, GET /d" in order

  @by_id
  Scenario: setNextRequest by request id jumps to that row
    Given the workspace file "acme/collections/flow/a.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{a}}
      name: A
      method: GET
      url: "{{api}}/a"
      scripts:
        preRequest: ""
        postResponse: pm.execution.setNextRequest('{{d}}')
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "01JR0000000000000000000040", "config": { "type": "functional", "rows": [{ "request_id": "{{a}}", "alias_id": "", "selected": true }, { "request_id": "{{b}}", "alias_id": "", "selected": true }, { "request_id": "{{c}}", "alias_id": "", "selected": true }, { "request_id": "{{d}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "GET /a, GET /d" in order

  @backwards
  Scenario: setNextRequest can go back to an earlier row
    Given the workspace file "acme/collections/flow/a.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{a}}
      name: A
      method: GET
      url: "{{api}}/a"
      scripts:
        preRequest: ""
        postResponse: ""
      """
    And the workspace file "acme/collections/flow/c.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{c}}
      name: C
      method: GET
      url: "{{api}}/c"
      scripts:
        postResponse: |
          if (!pm.collectionVariables.get("looped")) {
            pm.collectionVariables.set("looped", "yes");
            pm.execution.setNextRequest("A");
          }
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "01JR0000000000000000000040", "config": { "type": "functional", "rows": [{ "request_id": "{{a}}", "alias_id": "", "selected": true }, { "request_id": "{{b}}", "alias_id": "", "selected": true }, { "request_id": "{{c}}", "alias_id": "", "selected": true }, { "request_id": "{{d}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": false, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "GET /a, GET /b, GET /c, GET /a, GET /b, GET /c, GET /d" in order

  @last_call_wins
  Scenario Outline: The last setNextRequest call of a row wins - <case>
    Given the workspace file "acme/collections/flow/a.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{a}}
      name: A
      method: GET
      url: "{{api}}/a"
      scripts:
        preRequest: <pre>
        postResponse: <post>
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "01JR0000000000000000000040", "config": { "type": "functional", "rows": [{ "request_id": "{{a}}", "alias_id": "", "selected": true }, { "request_id": "{{b}}", "alias_id": "", "selected": true }, { "request_id": "{{c}}", "alias_id": "", "selected": true }, { "request_id": "{{d}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "<sent>" in order

    Examples:
      | case                         | pre                                   | post                                                              | sent          |
      | two calls in one script      |                                       | pm.execution.setNextRequest('B'); pm.execution.setNextRequest('D') | GET /a, GET /d |
      | a pre-request call counts    | pm.execution.setNextRequest('D')      |                                                                   | GET /a, GET /d |
      | post-response call overrides | pm.execution.setNextRequest('B')      | pm.execution.setNextRequest('D')                                  | GET /a, GET /d |

  @skipped_row
  Scenario: A call in a skipped row applies right after the skip
    Given the workspace file "acme/collections/flow/a.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{a}}
      name: A
      method: GET
      url: "{{api}}/a"
      scripts:
        preRequest: ""
        postResponse: ""
      """
    And the workspace file "acme/collections/flow/b.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{b}}
      name: B
      method: GET
      url: "{{api}}/b"
      scripts:
        preRequest: pm.execution.setNextRequest('D'); pm.execution.skipRequest()
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "01JR0000000000000000000040", "config": { "type": "functional", "rows": [{ "request_id": "{{a}}", "alias_id": "", "selected": true }, { "request_id": "{{b}}", "alias_id": "", "selected": true }, { "request_id": "{{c}}", "alias_id": "", "selected": true }, { "request_id": "{{d}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "GET /a, GET /d" in order
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.1.name" equal to "B"
    And the response should contain the field "rows.1.outcome" equal to "skipped"

  @stop_iteration
  Scenario: setNextRequest(null) ends the iteration and the next one starts at row 1
    Given the workspace file "acme/collections/flow/a.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{a}}
      name: A
      method: GET
      url: "{{api}}/a"
      scripts:
        preRequest: ""
        postResponse: pm.execution.setNextRequest(null)
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "01JR0000000000000000000040", "config": { "type": "functional", "rows": [{ "request_id": "{{a}}", "alias_id": "", "selected": true }, { "request_id": "{{b}}", "alias_id": "", "selected": true }, { "request_id": "{{c}}", "alias_id": "", "selected": true }, { "request_id": "{{d}}", "alias_id": "", "selected": true }], "functional": { "iterations": 2, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "GET /a, GET /a" in order

  @not_in_run
  Scenario Outline: A target that is not a selected row ends the iteration with a warning - <case>
    Given the workspace file "acme/collections/flow/a.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{a}}
      name: A
      method: GET
      url: "{{api}}/a"
      scripts:
        preRequest: ""
        postResponse: pm.execution.setNextRequest('<target>')
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "01JR0000000000000000000040", "config": { "type": "functional", "rows": [{ "request_id": "{{a}}", "alias_id": "", "selected": true }, { "request_id": "{{b}}", "alias_id": "", "selected": true }, { "request_id": "{{c}}", "alias_id": "", "selected": false }, { "request_id": "{{d}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "GET /a" in order
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.0.console" containing "Couldn't find request"
    And the response should contain the field "rows.0.console" containing "<target>"
    And the response should contain the field "rows.0.console" containing "in this run; the iteration ends."

    Examples:
      | case           | target |
      | unchecked row  | C      |
      | unknown name   | Nope   |

  @legacy
  Scenario: The legacy postman.setNextRequest works the same way
    Given the workspace file "acme/collections/flow/a.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{a}}
      name: A
      method: GET
      url: "{{api}}/a"
      scripts:
        preRequest: ""
        postResponse: postman.setNextRequest('D')
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "01JR0000000000000000000040", "config": { "type": "functional", "rows": [{ "request_id": "{{a}}", "alias_id": "", "selected": true }, { "request_id": "{{b}}", "alias_id": "", "selected": true }, { "request_id": "{{c}}", "alias_id": "", "selected": true }, { "request_id": "{{d}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 5000 ms
    And the mock API should have received the requests "GET /a, GET /d" in order

  @single_send
  Scenario: Outside a run setNextRequest stays a no-op
    Given the workspace file "acme/collections/flow/a.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{a}}
      name: A
      method: GET
      url: "{{api}}/a"
      scripts:
        preRequest: ""
        postResponse: pm.execution.setNextRequest('D')
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-a", "request": { "id": "{{a}}", "name": "A", "method": "GET", "url": "{{api}}/a" } }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    And the mock API should have received the requests "GET /a" in order
