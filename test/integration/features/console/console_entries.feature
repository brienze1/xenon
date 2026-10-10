#language: en
#utf-8

@all @xen_004 @console @console_entries
Feature: Console entries
  The console is a session log of everything a send does, in order: every request that goes out
  (the main send and the requests its scripts send with pm.sendRequest or pm.execution.runRequest)
  with method, URL, status, time and size, every console call of a script with its level and its
  arguments as text and as object trees, script errors with their location and failed tests.
  Entries are listed oldest first (List pages with after_id and limit) and Get returns one entry
  with its full detail: for a request, the request as it was sent (variables resolved, script
  changes applied) and the response. Bodies up to 1 MB are inline; larger bodies are opened on
  demand from body_url. The console keeps the newest 5,000 entries. New entries reach the UI as
  console:entries events, batched at most every 50 ms; Clear and console.clear() in a script empty
  the console and tell the UI with a cleared event.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK8000000000000000000001"
    And the "create" variable is set to "01JK8000000000000000000002"
    And the "fetch" variable is set to "01JK8000000000000000000003"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK8000000000000000000000
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
      variables:
        - key: base_url
          value: "{{api}}"
          enabled: true
          description: ""
      """
    And the 1 "POST" request to "/invoices" returns status 201 with the following response
      """
      {"id":"inv-1"}
      """
    And the requests to "/invoices" respond with the header "X-Trace" equal to "trace-1"
    And I call "ConsoleService.Clear"

  @request_entry
  Scenario: The main send is a request entry between the pre-request and post-response logs, with its resolved request and response in the detail
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{base_url}}/invoices?draft=true"
      headers:
        - key: X-Client
          value: "{{client}}"
          enabled: true
          description: ""
      scripts:
        preRequest: |
          pm.variables.set("client", "acme")
          pm.request.headers.upsert({ key: "X-Signed", value: "yes" })
          console.log("pre")
        postResponse: console.log("post")
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{base_url}}/invoices?draft=true", "headers": [{ "key": "X-Client", "value": "{{client}}", "enabled": true, "description": "" }], "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"amount\": 10}" } } } }
      """
    Then the call should succeed
    And I save the response field "send_id" as "send"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 3
    And the response should contain the field "0.summary" equal to "pre"
    And the response should contain the field "1.kind" equal to "request"
    And the response should contain the field "1.level" equal to "info"
    And the response should contain the field "1.send_id" equal to "{{send}}"
    And the response should contain the field "1.summary" equal to "POST {{api}}/invoices?draft=true"
    And the response should contain the field "1.request.method" equal to "POST"
    And the response should contain the field "1.request.url" equal to "{{api}}/invoices?draft=true"
    And the response should contain the field "1.request.nested" equal to "false"
    And the response field "1.request.headers" should not exist
    And the response should contain the field "1.response.status_code" equal to "201"
    And the response should contain the field "1.response.status_text" equal to "Created"
    And the response should contain the field "1.response.size_bytes" equal to "14"
    And the response field "1.response.time_ms" should exist
    And the response field "1.response.body" should not exist
    And the response field "1.script" should not exist
    And the response should contain the field "2.summary" equal to "post"
    When I save the response field "1.id" as "entry"
    And I call "ConsoleService.Get" with
      """
      { "id": "{{entry}}" }
      """
    Then the call should succeed
    And the response should contain the field "id" equal to "{{entry}}"
    And the response should contain the field "kind" equal to "request"
    And the response field "request.headers" should contain an item with "key" equal to "X-Client" and "value" equal to "acme"
    And the response field "request.headers" should contain an item with "key" equal to "X-Signed" and "value" equal to "yes"
    And the response should contain the field "request.body" equal to
      """
      {"amount": 10}
      """
    And the response field "response.headers" should contain an item with "key" equal to "X-Trace" and "value" equal to "trace-1"
    And the response should contain the field "response.body" equal to
      """
      {"id":"inv-1"}
      """
    And the response should contain the field "response.content_type" equal to "application/json"
    And the response field "response.body_url" should not exist
    And the response should contain the field "masked" equal to "false"

  @nested
  Scenario: Requests sent by pm.sendRequest and runRequest are nested request entries in the order they happen
    Given the 1 "GET" request to "/token" returns status 200 with the following response
      """
      {"token":"t-1"}
      """
    And the 1 "GET" request to "/fetch" returns status 200 with the following response
      """
      {"ok":true}
      """
    And the workspace file "acme/collections/billing/fetch.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{fetch}}
      name: Fetch
      method: GET
      url: "{{base_url}}/fetch"
      scripts:
        preRequest: console.log("fetch pre")
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{base_url}}/invoices"
      scripts:
        preRequest: |
          const token = await pm.sendRequest("{{base_url}}/token")
          console.log("token", token.json().token)
          await pm.execution.runRequest("{{fetch}}")
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    And I save the response field "send_id" as "send"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 5
    And the response should contain the field "0.kind" equal to "request"
    And the response should contain the field "0.summary" equal to "GET {{api}}/token"
    And the response should contain the field "0.request.nested" equal to "true"
    And the response should contain the field "0.response.status_code" equal to "200"
    And the response should contain the field "0.send_id" equal to "{{send}}"
    And the response should contain the field "0.script.owner_name" equal to "Create"
    And the response should contain the field "0.script.phase" equal to "preRequest"
    And the response should contain the field "1.summary" equal to "token t-1"
    And the response should contain the field "2.summary" equal to "fetch pre"
    And the response should contain the field "2.script.owner_name" equal to "Fetch"
    And the response should contain the field "3.kind" equal to "request"
    And the response should contain the field "3.summary" equal to "GET {{api}}/fetch"
    And the response should contain the field "3.request.nested" equal to "true"
    And the response should contain the field "4.kind" equal to "request"
    And the response should contain the field "4.summary" equal to "POST {{api}}/invoices"
    And the response should contain the field "4.request.nested" equal to "false"

  @log_levels
  Scenario: Console calls keep their level, their text and their arguments as object trees
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{base_url}}/invoices"
      scripts:
        postResponse: |
          console.log("plain", 1)
          console.info({ id: pm.response.json().id, tags: ["a", "b"] })
          console.warn("careful")
          console.error("broken")
          console.debug("details")
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 6
    And the response should contain the field "1.kind" equal to "log"
    And the response should contain the field "1.level" equal to "log"
    And the response should contain the field "1.summary" equal to "plain 1"
    And the response should contain the field "1.arguments" with the items "plain,1"
    And the response field "1.values" should not exist
    And the response should contain the field "2.level" equal to "info"
    And the response should contain the field "3.level" equal to "warn"
    And the response should contain the field "4.level" equal to "error"
    And the response should contain the field "4.kind" equal to "log"
    And the response should contain the field "5.level" equal to "debug"
    And the response should contain the field "5.script.owner_kind" equal to "request"
    And the response should contain the field "5.script.phase" equal to "postResponse"
    And the response should contain the field "5.script.label" equal to "Create · Post-response"
    When I save the response field "2.id" as "entry"
    And I call "ConsoleService.Get" with
      """
      { "id": "{{entry}}" }
      """
    Then the call should succeed
    And the response should contain the field "values" array length equal to 1
    And the response should contain the field "values.0.id" equal to "inv-1"
    And the response should contain the field "values.0.tags" with the items "a,b"

  @errors
  Scenario: A script error and failed tests are error entries with the script and the error's location
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{base_url}}/invoices"
      scripts:
        postResponse: |
          pm.test("is 200", () => pm.response.to.have.status(200))
          pm.test("is created", () => pm.response.to.have.status(201))
          null.missing
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 3
    And the response should contain the field "0.kind" equal to "request"
    And the response should contain the field "1.kind" equal to "test_failure"
    And the response should contain the field "1.level" equal to "error"
    And the response should contain the field "1.summary" equal to "is 200: expected response to have status code 200 but got 201"
    And the response should contain the field "1.script.label" equal to "Create · Post-response"
    And the response should contain the field "2.kind" equal to "error"
    And the response should contain the field "2.summary" equal to "TypeError: Cannot read properties of null (reading 'missing')"
    And the response should contain the field "2.line" equal to "3"
    And the response should contain the field "2.column" equal to "6"
    And the response should contain the field "2.script.owner_name" equal to "Create"

  @send_failure
  Scenario: A request that cannot be sent is a request entry with the error
    Given the "closed" variable is set to a closed port url
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{closed}}/down"
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{closed}}/down" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.kind" equal to "request"
    And the response should contain the field "0.level" equal to "error"
    And the response should contain the field "0.summary" equal to "GET {{closed}}/down"
    And the response should contain the field "0.response.status_code" equal to "0"
    And the response field "0.response.error" should exist

  @get_unknown
  Scenario: Get of an entry that is no longer kept fails with CON-01404
    When I call "ConsoleService.Get" with
      """
      { "id": "01JK8999999999999999999999" }
      """
    Then the call should fail with code "CON-01404"

  @get_invalid
  Scenario: Get without an id fails validation
    When I call "ConsoleService.Get" with
      """
      { "id": "" }
      """
    Then the call should fail with code "CON-02400"

  @large_body
  Scenario: A response body over 1 MB is not inline in the detail and is opened from its body_url
    Given the 1 "GET" request to "/export" returns status 200 with content type "text/plain" and a generated body of 1500000 bytes
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/export"
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/export" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    And I save the response field "0.id" as "entry"
    And I call "ConsoleService.Get" with
      """
      { "id": "{{entry}}" }
      """
    Then the response should contain the field "response.size_bytes" equal to "1500000"
    And the response field "response.body" should not exist
    And the response should contain the field "response.body_url" equal to "/xenon/response/console:{{entry}}"
    When I request "/xenon/response/console:{{entry}}" from the "ResponseBodyRoute" route
    Then the route response status should be 200
    And the route response body should be the generated body of 1500000 bytes
    When I call "ConsoleService.Clear"
    And I request "/xenon/response/console:{{entry}}" from the "ResponseBodyRoute" route
    Then the route response status should be 404

  @pagination
  Scenario: List pages through the entries with after_id and limit
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{base_url}}/invoices"
      scripts:
        preRequest: |
          for (let i = 0; i < 5; i++) console.log("line " + i)
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{base_url}}/invoices" } }
      """
    When I call "ConsoleService.List" with
      """
      { "limit": 2 }
      """
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "1.summary" equal to "line 1"
    When I save the response field "1.id" as "cursor"
    And I call "ConsoleService.List" with
      """
      { "after_id": "{{cursor}}", "limit": 10 }
      """
    Then the response should contain the field "" array length equal to 4
    And the response should contain the field "0.summary" equal to "line 2"
    And the response should contain the field "3.kind" equal to "request"

  @list_invalid
  Scenario: List without a limit fails validation
    When I call "ConsoleService.List" with
      """
      { "limit": 0 }
      """
    Then the call should fail with code "CON-00400"

  @ring_buffer
  Scenario: The console keeps the newest 5,000 entries
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{base_url}}/invoices"
      scripts:
        postResponse: |
          for (let i = 0; i < 5010; i++) console.log("line " + i)
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 5000 }
      """
    Then the response should contain the field "" array length equal to 5000
    And the response should contain the field "0.summary" equal to "line 10"
    And the response should contain the field "4999.summary" equal to "line 5009"

  @events
  Scenario: New entries reach the UI as console:entries events, batched rather than one event per entry
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{base_url}}/invoices"
      scripts:
        postResponse: |
          for (let i = 0; i < 20; i++) console.log("line " + i)
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    And the "console:entries" events should have 21 items in the field "entries" within 500 ms
    And the event "console:entries" should have been emitted at most 3 times
    And the 1 "console:entries" event should contain the field "entries.0.kind" equal to "request"
    And the 1 "console:entries" event should contain the field "cleared" equal to "false"
    And the last "console:entries" event should contain the field "entries.-1.summary" equal to "line 19"

  @clear
  Scenario: Clear empties the console and tells the UI
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{base_url}}/invoices"
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{base_url}}/invoices" } }
      """
    And the "console:entries" events should have 1 items in the field "entries" within 500 ms
    When I call "ConsoleService.Clear"
    Then the call should succeed
    And the last "console:entries" event should contain the field "cleared" equal to "true"
    And the last "console:entries" event should contain the field "entries" equal to "[]"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 0

  @script_clear
  Scenario: console.clear() in a script clears everything logged before it
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{base_url}}/invoices"
      scripts:
        preRequest: console.log("before clear")
        postResponse: |
          console.log("still before")
          console.clear()
          console.log("after clear")
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.summary" equal to "after clear"
    And a "console:entries" event with the field "cleared" equal to "true" should have been emitted within 500 ms
    And a "console:entries" event with the field "entries.0.summary" equal to "after clear" should have been emitted within 500 ms
