#language: en
#utf-8

@all @xen_004 @scripts @run_request_skip
Feature: pm.execution and pm.info control and describe the running send
  pm.execution.runRequest(id) runs another saved request of the workspace with its own scripts
  (its collection and folder scripts included) and resolves to its response; the variable
  changes it makes merge back into the calling script; it never reaches History. Options
  "variables" override variables for that run only. A request that skips itself resolves to null
  and an unknown id rejects. pm.execution.skipRequest() exists in pre-request scripts only;
  pm.execution.setNextRequest() does nothing (the collection runner is not part of Xenon);
  pm.execution.location lists the path of the request and "current" names the item whose script
  is running. pm.info gives the event name, the request name and id, iteration 0 of 1.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK7000000000000000000001"
    And the "invoices" variable is set to "01JK7000000000000000000002"
    And the "create" variable is set to "01JK7000000000000000000004"
    And the "fetch" variable is set to "01JK7000000000000000000005"
    And the "dev" variable is set to "01JK7000000000000000000009"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK7000000000000000000000
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
      order:
        - fetch-token
        - invoices
      scripts:
        preRequest: console.log("collection pre", pm.info.requestName)
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      """
    And the workspace file "acme/collections/billing/fetch-token.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{fetch}}
      name: Fetch token
      method: POST
      url: "{{base_url}}/token?client={{client}}"
      scripts:
        postResponse: |
          pm.environment.set("token", pm.response.json().token)
          console.log("token post", pm.response.code)
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: base_url
          value: "{{api}}"
          enabled: true
          description: ""
        - key: client
          value: acme
          enabled: true
          description: ""
      """
    And the 1 "POST" request to "/token" returns status 200 with the following response
      """
      {"token":"t-1"}
      """
    And the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      {"id":"inv-1"}
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    And I call "ConsoleService.Clear"

  @run_request
  Scenario: runRequest runs the saved request with its scripts and merges its variable changes back
    Given the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/invoices"
      headers:
        - key: Authorization
          value: "Bearer {{token}}"
          enabled: true
          description: ""
      scripts:
        preRequest: |
          const res = await pm.execution.runRequest("{{fetch}}")
          console.log("after run", res.code, res.json().token, pm.environment.get("token"))
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/invoices", "headers": [{ "key": "Authorization", "value": "Bearer {{token}}", "enabled": true }] } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the response should contain the field "script_errors" array length equal to 0
    And the 1 "POST" request to "/token" should have the query "client" equal to "acme"
    And the 1 "GET" request to "/invoices" should have the header "Authorization" equal to "Bearer t-1"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.token" equal to "t-1"
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "0.summary" equal to "collection pre Create"
    And the response should contain the field "1.summary" equal to "collection pre Fetch token"
    And the response should contain the field "2.summary" equal to "token post 200"
    And the response should contain the field "3.summary" equal to "after run 200 t-1 t-1"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.url" equal to "{{base_url}}/invoices"

  @run_request @options
  Scenario: runRequest variables override the target's variables for that run only
    Given the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/invoices"
      scripts:
        preRequest: |
          await pm.execution.runRequest("{{fetch}}", { variables: { client: "globex" } })
          console.log(pm.environment.get("client"), pm.variables.get("client"))
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    And the 1 "POST" request to "/token" should have the query "client" equal to "globex"
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "3.summary" equal to "acme acme"

  @run_request @errors
  Scenario: runRequest rejects an id that is not a saved request and resolves null for a skipped one
    Given the workspace file "acme/collections/billing/fetch-token.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{fetch}}
      name: Fetch token
      method: POST
      url: "{{base_url}}/token"
      scripts:
        preRequest: pm.execution.skipRequest()
      """
    And the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/invoices"
      scripts:
        preRequest: |
          try { await pm.execution.runRequest("01JK7999999999999999999999") } catch (err) { console.log(err.message) }
          console.log(await pm.execution.runRequest("{{fetch}}"))
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the "POST" request to "/token" should have been received 0 times
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "1.summary" equal to "runRequest: no saved request with id 01JK7999999999999999999999"
    And the response should contain the field "3.summary" equal to "null"

  @skip_request
  Scenario: skipRequest is not available in post-response scripts
    Given the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/invoices"
      scripts:
        postResponse: |
          console.log(typeof pm.execution.skipRequest)
          pm.execution.skipRequest()
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the response should contain the field "script_errors.0.message" equal to "TypeError: pm.execution.skipRequest is not a function"
    And the response should contain the field "script_errors.0.line" equal to "2"

  @location @info
  Scenario: pm.execution.location, setNextRequest and pm.info describe the running script
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      order:
        - fetch-token
        - invoices
      scripts:
        postResponse: console.log(pm.execution.location.current, pm.info.eventName)
      """
    And the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/invoices"
      scripts:
        preRequest: |
          pm.execution.setNextRequest("Fetch token")
          console.log(pm.execution.location.join(" / "), pm.execution.location.current, pm.info.eventName, pm.info.iteration, pm.info.iterationCount, pm.info.requestName, pm.info.requestId)
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the "POST" request to "/token" should have been received 0 times
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "0.summary" equal to "Billing / Invoices / Create Create prerequest 0 1 Create 01JK7000000000000000000004"
    And the response should contain the field "1.summary" equal to "Billing test"
