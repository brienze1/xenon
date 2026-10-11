#language: en
#utf-8

@all @xen_004 @scripts @send_request
Feature: pm.sendRequest sends extra requests from a script
  pm.sendRequest takes a URL or a Postman request definition, resolves its {{variables}} (the
  send's local scope included) and either calls back (err, response) or returns a promise of the
  response; the script waits for it. The response offers Postman's Response API (code, status,
  headers, text(), json(), responseTime, responseSize). Requests sent from scripts reach the
  network but never History. A request that cannot be sent gives the script an error.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK6000000000000000000001"
    And the "create" variable is set to "01JK6000000000000000000004"
    And the "dev" variable is set to "01JK6000000000000000000009"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK6000000000000000000000
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
          value: acme-client
          enabled: true
          description: ""
      """
    And the 1 "POST" request to "/token" returns status 200 with the following response
      """
      {"token":"t-1"}
      """
    And the requests to "/token" respond with the header "X-Trace" equal to "trace-1"
    And the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      {"id":"inv-1"}
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    And I call "ConsoleService.Clear"

  @callback
  Scenario: The callback form gets the error and the response
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/invoices"
      scripts:
        preRequest: |
          pm.sendRequest({ url: "{{base_url}}/token", method: "POST", header: [{ key: "X-Client", value: "{{client}}" }], body: { mode: "urlencoded", urlencoded: [{ key: "grant_type", value: "client_credentials" }] } }, (err, res) => {
            console.log(err, res.code, res.status, res.headers.get("X-Trace"), res.json().token, typeof res.responseTime, res.responseSize > 0)
            pm.variables.set("token", res.json().token)
          })
        postResponse: console.log(pm.variables.get("token"))
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the response should contain the field "script_errors" array length equal to 0
    And the "POST" request to "/token" should have been received 1 times
    And the 1 "POST" request to "/token" should have the header "X-Client" equal to "acme-client"
    And the 1 "POST" request to "/token" should have the body equal to
      """
      grant_type=client_credentials
      """
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "0.summary" equal to "null 200 OK trace-1 t-1 number true"
    And the response should contain the field "1.summary" equal to "t-1"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.url" equal to "{{base_url}}/invoices"

  @promise
  Scenario: The promise form is awaited at the top level and a URL string is enough
    Given the 1 "GET" request to "/token" returns status 200 with the following response
      """
      {"token":"t-get"}
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
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
          const res = await pm.sendRequest("{{base_url}}/token?client={{client}}")
          pm.variables.set("token", res.json().token)
          console.log(res.code, res.text())
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/invoices", "headers": [{ "key": "Authorization", "value": "Bearer {{token}}", "enabled": true }] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/token" should have the query "client" equal to "acme-client"
    And the 1 "GET" request to "/invoices" should have the header "Authorization" equal to "Bearer t-get"
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "0.summary" equal to
      """
      200 {"token":"t-get"}
      """

  @parallel
  Scenario: Several requests in flight are all awaited before the script ends
    Given the 1 "GET" request to "/a" returns status 200 with the following response
      """
      {"n":1}
      """
    And the 1 "GET" request to "/b" returns status 201 with the following response
      """
      {"n":2}
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/invoices"
      scripts:
        preRequest: |
          const [a, b] = await Promise.all([pm.sendRequest("{{base_url}}/a"), pm.sendRequest("{{base_url}}/b")])
          pm.sendRequest("{{base_url}}/a", (err, res) => console.log("late", err === null, res.code))
          console.log(a.json().n + b.json().n, b.code)
      """
    And the 2 "GET" request to "/a" returns status 202 with the following response
      """
      {"n":3}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    And the "GET" request to "/a" should have been received 2 times
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "0.summary" equal to "3 201"
    And the response should contain the field "1.summary" equal to "late true 202"

  @error
  Scenario Outline: A request that cannot be sent gives the script an error - <form>
    Given the "closed" variable is set to a closed port url
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/invoices"
      scripts:
        preRequest: <script>
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "0.summary" equal to "Error true"

    Examples:
      | form     | script                                                                                                                     |
      | callback | "pm.sendRequest('{{closed}}/x', (err, res) => console.log(err.name, res === undefined))"                                    |
      | promise  | "try { await pm.sendRequest('{{closed}}/x') } catch (err) { console.log(err.name, err.message.length > 0) }"                |
