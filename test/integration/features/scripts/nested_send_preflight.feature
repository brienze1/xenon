#language: en
#utf-8

@all @xen_004 @scripts @nested_send_preflight
Feature: Requests sent from scripts share the main send's preflight
  Confirming a send to an environment with "Confirm before send" covers every request its scripts
  send in the same execution: a script never triggers a separate confirmation. A request sent
  from a script that uses a secret without a value on this machine (or in a locked vault) is
  not sent; the script receives an error naming the secrets, and there is no prompt. Requests
  sent from scripts are not added to History. The {{variables}} of a script request resolve with
  the send's local scope.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK4000000000000000000001"
    And the "create" variable is set to "01JK4000000000000000000004"
    And the "prod" variable is set to "01JK4000000000000000000009"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK4000000000000000000000
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
    And the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{prod}}
      name: prod
      color: red
      confirmBeforeSend: true
      variables:
        - key: api_key
          secret: true
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
    And the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      {"id":"inv-1"}
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    And I call "ConsoleService.Clear"

  @confirmed
  Scenario: The main send's confirmation covers the requests its scripts send
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      headers:
        - key: Authorization
          value: "Bearer {{token}}"
          enabled: true
          description: ""
      scripts:
        preRequest: |
          pm.variables.set("path", "token")
          const res = await pm.sendRequest({ url: "{{api}}/{{path}}", method: "POST", header: { "X-Client": "{{client}}" }, body: { mode: "raw", raw: "grant=client" } })
          console.log(res.code, res.status, res.headers.get("Content-Type"), res.text())
          pm.variables.set("token", res.json().token)
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "confirmed_environment_id": "{{prod}}",
        "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices", "headers": [{ "key": "Authorization", "value": "Bearer {{token}}", "enabled": true, "description": "" }] }
      }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the response should contain the field "script_errors" array length equal to 0
    And the "POST" request to "/token" should have been received 1 times
    And the 1 "POST" request to "/token" should have the header "X-Client" equal to "acme-client"
    And the 1 "POST" request to "/token" should have the body equal to
      """
      grant=client
      """
    And the 1 "GET" request to "/invoices" should have the header "Authorization" equal to "Bearer t-1"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "0.summary" equal to
      """
      200 OK application/json {"token":"t-1"}
      """
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.url" equal to "{{api}}/invoices"

  @unconfirmed
  Scenario: Without the confirmation no script runs and nothing is sent
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
          console.log("should not run")
          pm.sendRequest("{{api}}/token", () => {})
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should fail with code "SND-05409"
    And the "POST" request to "/token" should have been received 0 times
    And the "GET" request to "/token" should have been received 0 times
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 0

  @missing_secret
  Scenario Outline: A script request that uses a missing secret is not sent and the script gets an error naming it - <form>
    Given the workspace file "acme/collections/billing/create.yaml" contains
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
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "confirmed_environment_id": "{{prod}}", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the "POST" request to "/token" should have been received 0 times
    And the "GET" request to "/invoices" should have been received 1 times
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.summary" equal to "The request uses secrets that have no value on this machine or are locked: api_key"

    Examples:
      | form     | script                                                                                                                                                         |
      | callback | "pm.sendRequest({ url: '{{api}}/token', method: 'POST', header: [{ key: 'X-Key', value: '{{api_key}}' }] }, (err, res) => { console.log(err.message) })"        |
      | promise  | "try { await pm.sendRequest({ url: '{{api}}/token?key={{api_key}}', method: 'POST' }) } catch (err) { console.log(err.message) }"                                 |
