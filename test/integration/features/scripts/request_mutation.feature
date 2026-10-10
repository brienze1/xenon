#language: en
#utf-8

@all @xen_004 @scripts @request_mutation
Feature: Pre-request scripts change the request that is sent, never the stored one
  pm.request in a pre-request script is a copy of the request as written ({{variables}} not yet
  replaced). Changes to its method, URL and query, headers and body reach the network; the saved
  request file and the History entry keep the request exactly as written. Variables are resolved
  after the pre-request scripts, so pm.variables set by a script resolve in the same send and win
  over every other scope.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK2000000000000000000001"
    And the "create" variable is set to "01JK2000000000000000000004"
    And the "dev" variable is set to "01JK2000000000000000000009"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK2000000000000000000000
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

  @mutation
  Scenario: Method, query, headers and body changes reach the API; the file and History keep the request as written
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices?keep=1"
      headers:
        - key: X-Old
          value: old
          enabled: true
          description: ""
        - key: X-Keep
          value: keep
          enabled: true
          description: ""
      body:
        mode: raw
        raw:
          language: json
          content: '{"original":true}'
      scripts:
        preRequest: |
          pm.request.method = "POST"
          pm.request.headers.upsert({ key: "X-Signature", value: "sig-123" })
          pm.request.headers.upsert({ key: "X-Keep", value: "changed" })
          pm.request.headers.remove("X-Old")
          pm.request.url.query.add({ key: "page", value: "2" })
          pm.request.addQueryParams("size=10")
          pm.request.body.update('{"changed":true}')
          console.log(pm.request.method, pm.request.headers.get("X-Keep"), pm.request.url.query.count(), pm.request.body.raw)
      """
    And the 1 "POST" request to "/invoices" returns status 201 with the following response
      """
      {"id":"inv-1"}
      """
    And I call "ConsoleService.Clear"
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices?keep=1",
          "params": [{ "key": "keep", "value": "1", "enabled": true, "description": "" }],
          "headers": [
            { "key": "X-Old", "value": "old", "enabled": true, "description": "" },
            { "key": "X-Keep", "value": "keep", "enabled": true, "description": "" }
          ],
          "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"original\":true}" } }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the response should contain the field "status_code" equal to "201"
    And the "POST" request to "/invoices" should have been received 1 times
    And the "GET" request to "/invoices" should have been received 0 times
    And the 1 "POST" request to "/invoices" should have the header "X-Signature" equal to "sig-123"
    And the 1 "POST" request to "/invoices" should have the header "X-Keep" equal to "changed"
    And the 1 "POST" request to "/invoices" should not have the header "X-Old"
    And the 1 "POST" request to "/invoices" should have the raw query "keep=1&page=2&size=10"
    And the 1 "POST" request to "/invoices" should have the body equal to
      """
      {"changed":true}
      """
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "0.summary" equal to
      """
      POST changed 3 {"changed":true}
      """
    And the workspace file "acme/collections/billing/create.yaml" should be unchanged
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.method" equal to "GET"
    And the response should contain the field "0.url" equal to "{{api}}/invoices?keep=1"
    And the response should contain the field "0.outcome" equal to "sent"
    And the response should contain the field "0.status_code" equal to "201"
    When I save the response field "0.id" as "entry"
    And I call "HistoryService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{entry}}" }
      """
    Then the response should contain the field "request.headers" array length equal to 2
    And the response should contain the field "request.headers.0.key" equal to "X-Old"
    And the response should contain the field "request.headers.1.value" equal to "keep"
    And the response should contain the field "request.body.raw.content" equal to
      """
      {"original":true}
      """

  @variables
  Scenario: Variables set by a pre-request script resolve in the same send and win over the environment
    Given the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: token
          value: env-token
          enabled: true
          description: ""
        - key: tenant
          value: acme
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      scripts:
        preRequest: |
          pm.variables.set("token", "script-token")
          pm.variables.set("invoice", 42)
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices/{{invoice}}"
      headers:
        - key: Authorization
          value: "Bearer {{token}}"
          enabled: true
          description: ""
      scripts:
        preRequest: |
          pm.request.headers.upsert({ key: "X-Tenant", value: "{{tenant}}" })
          console.log(pm.variables.get("tenant"), pm.variables.get("token"), pm.variables.replaceIn("{{tenant}}-{{invoice}}"))
      """
    And the 1 "GET" request to "/invoices/42" returns status 200 with the following response
      """
      {"id":"42"}
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    And I call "ConsoleService.Clear"
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices/{{invoice}}",
          "headers": [{ "key": "Authorization", "value": "Bearer {{token}}", "enabled": true, "description": "" }]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the 1 "GET" request to "/invoices/42" should have the header "Authorization" equal to "Bearer script-token"
    And the 1 "GET" request to "/invoices/42" should have the header "X-Tenant" equal to "acme"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "0.summary" equal to "acme script-token acme-42"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.url" equal to "{{api}}/invoices/{{invoice}}"
