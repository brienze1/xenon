#language: en
#utf-8

@all @xen_004 @scripts @errors_and_limits
Feature: Script errors, limits and skipRequest
  A pre-request script that throws, has a syntax error, runs past its time limit, overflows the
  stack, uses too much memory or leaves a promise rejection unhandled stops the send: nothing goes
  out, the result has outcome stopped_by_script_error with the error, the script that failed
  (owner kind, name, phase) and the line and column, and History records the attempt. A
  post-response script error keeps the response and the remaining post-response scripts still
  run. pm.execution.skipRequest() in a pre-request script skips the send without an error. The
  time limit is the script timeout setting (default 10 s) and includes waiting for timers.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK3000000000000000000001"
    And the "create" variable is set to "01JK3000000000000000000004"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK3000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      scripts:
        postResponse: console.log("collection post")
      """
    And the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      {"id":"inv-1"}
      """
    And I call "ConsoleService.Clear"

  @pre_throw
  Scenario: A pre-request script that throws stops the send and reports where it failed
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: |
          console.log("before")
          throw new Error("boom")
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "stopped_by_script_error"
    And the response should contain the field "ok" equal to "false"
    And the response should contain the field "status_code" equal to "0"
    And the response field "error" should not exist
    And the response should contain the field "script_errors" array length equal to 1
    And the response should contain the field "script_errors.0.kind" equal to "script_runtime"
    And the response should contain the field "script_errors.0.message" equal to "Error: boom"
    And the response should contain the field "script_errors.0.line" equal to "2"
    And the response should contain the field "script_errors.0.column" equal to "7"
    And the response should contain the field "script_errors.0.script.owner_kind" equal to "request"
    And the response should contain the field "script_errors.0.script.owner_name" equal to "Create"
    And the response should contain the field "script_errors.0.script.phase" equal to "preRequest"
    And the "GET" request to "/invoices" should have been received 0 times
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "0.summary" equal to "before"
    And the response should contain the field "1.kind" equal to "error"
    And the response should contain the field "1.summary" equal to "Error: boom"
    And the response should contain the field "1.script.owner_name" equal to "Create"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.outcome" equal to "stopped_by_script_error"
    And the response should contain the field "0.url" equal to "{{api}}/invoices"
    And the response field "0.status_code" should not exist

  @syntax
  Scenario: A syntax error in the collection script stops the send before later scripts run
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      scripts:
        preRequest: const a = ;
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: console.log("request pre")
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "stopped_by_script_error"
    And the response should contain the field "script_errors.0.kind" equal to "script_syntax"
    And the response should contain the field "script_errors.0.message" equal to "SyntaxError: Unexpected token ;"
    And the response should contain the field "script_errors.0.line" equal to "1"
    And the response should contain the field "script_errors.0.column" equal to "11"
    And the response should contain the field "script_errors.0.script.owner_kind" equal to "collection"
    And the response should contain the field "script_errors.0.script.owner_name" equal to "Billing"
    And the "GET" request to "/invoices" should have been received 0 times
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.kind" equal to "error"

  @post_throw
  Scenario: A post-response error keeps the response and the remaining post-response scripts run
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      scripts:
        postResponse: |
          null.missing
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: console.log("request post")
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the response should contain the field "status_code" equal to "200"
    And the response should contain the field "inline_body" equal to
      """
      {"id":"inv-1"}
      """
    And the response should contain the field "script_errors" array length equal to 1
    And the response should contain the field "script_errors.0.kind" equal to "script_runtime"
    And the response should contain the field "script_errors.0.script.owner_kind" equal to "collection"
    And the response should contain the field "script_errors.0.script.phase" equal to "postResponse"
    And the response should contain the field "script_errors.0.line" equal to "1"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "0.kind" equal to "error"
    And the response should contain the field "1.summary" equal to "request post"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.outcome" equal to "sent"
    And the response should contain the field "0.status_code" equal to "200"

  @timeout
  Scenario Outline: A script running past the time limit is stopped and the send does not go out - <case>
    Given I call "SettingsService.Save" with
      """
      { "script_timeout_ms": 1000 }
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: <script>
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the last call should have taken less than 4000 ms
    And the response should contain the field "outcome" equal to "stopped_by_script_error"
    And the response should contain the field "script_errors.0.kind" equal to "script_timeout"
    And the response should contain the field "script_errors.0.message" equal to "Script execution timed out"
    And the response should contain the field "script_errors.0.script.owner_name" equal to "Create"
    And the "GET" request to "/invoices" should have been received 0 times

    Examples:
      | case           | script                                           |
      | endless loop   | "while (true) {}"                                |
      | endless timers | "function tick() { setTimeout(tick, 10) }; tick()" |

  @stack_overflow
  Scenario: Deep recursion is stopped with a stack overflow error naming the script
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: |
          function down(depth) { return down(depth + 1) + 1 }
          down(0)
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "stopped_by_script_error"
    And the response should contain the field "script_errors.0.kind" equal to "script_stack_overflow"
    And the response should contain the field "script_errors.0.message" equal to "RangeError: Maximum call stack size exceeded"
    And the response should contain the field "script_errors.0.script.owner_name" equal to "Create"
    And the "GET" request to "/invoices" should have been received 0 times

  @memory
  Scenario: A script that keeps allocating is stopped before the app runs out of memory
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: |
          const hoard = []
          while (true) { hoard.push(new Array(100000).fill("x")) }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "stopped_by_script_error"
    And the response should contain the field "script_errors.0.kind" equal to "script_memory"
    And the response should contain the field "script_errors.0.message" equal to "Script stopped: it used too much memory"
    And the "GET" request to "/invoices" should have been received 0 times

  @unhandled_rejection
  Scenario: An unhandled promise rejection is a script error
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: Promise.reject(new Error("nope"))
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "stopped_by_script_error"
    And the response should contain the field "script_errors.0.kind" equal to "script_unhandled_rejection"
    And the response should contain the field "script_errors.0.message" equal to "Unhandled promise rejection: Error: nope"
    And the "GET" request to "/invoices" should have been received 0 times

  @skip_request
  Scenario: skipRequest skips the send without an error; later pre-request scripts do not run
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      scripts:
        preRequest: |
          pm.execution.skipRequest()
          console.log("after skip")
        postResponse: console.log("collection post")
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: console.log("request pre")
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "skipped_by_script"
    And the response should contain the field "script_errors" array length equal to 0
    And the "GET" request to "/invoices" should have been received 0 times
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.summary" equal to "after skip"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.outcome" equal to "skipped_by_script"

  @post_skip
  Scenario: skipRequest in a post-response script does nothing
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: |
          pm.execution.skipRequest()
          console.log("still here")
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the "GET" request to "/invoices" should have been received 1 times
