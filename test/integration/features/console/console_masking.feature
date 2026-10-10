#language: en
#utf-8

@all @xen_004 @console @console_masking
Feature: Secrets are masked in the console
  Every occurrence of a secret value in a console entry is shown as ••••• and the entry is marked
  masked: the URL, headers and body of a request entry (as sent and as received), log text and
  logged objects, error messages and failed tests. A secret value a script sets during a send is
  masked in the entries that follow. Values a script transforms (substrings, base64) are not
  recognised.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK8100000000000000000001"
    And the "create" variable is set to "01JK8100000000000000000002"
    And the "dev" variable is set to "01JK8100000000000000000009"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK8100000000000000000000
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
        - key: api_key
          secret: true
          enabled: true
          description: ""
        - key: session
          secret: true
          enabled: true
          description: ""
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key", "value": "k3y-s3cr3t-value" }
      """
    And I call "ConsoleService.Clear"

  @everywhere
  Scenario: A secret sent, echoed back and logged never appears in the console
    Given the 1 "POST" request to "/echo" returns status 200 with the following response
      """
      {"echo":"k3y-s3cr3t-value"}
      """
    And the requests to "/echo" respond with the header "X-Echo" equal to "k3y-s3cr3t-value"
    And the 1 "GET" request to "/token" returns status 200 with the following response
      """
      {"ok":true}
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{base_url}}/echo?key={{api_key}}"
      headers:
        - key: X-Api-Key
          value: "{{api_key}}"
          enabled: true
          description: ""
      scripts:
        preRequest: |
          await pm.sendRequest({ url: "{{base_url}}/token", header: { Authorization: "Bearer {{api_key}}" } })
          console.log("key is", pm.environment.get("api_key"))
          console.info({ nested: { key: pm.environment.get("api_key") } })
        postResponse: |
          pm.test("echo", () => pm.expect(pm.response.json().echo).to.eql("other"))
          throw new Error("leaked " + pm.environment.get("api_key"))
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{base_url}}/echo?key={{api_key}}", "headers": [{ "key": "X-Api-Key", "value": "{{api_key}}", "enabled": true, "description": "" }], "body": { "mode": "raw", "raw": { "language": "text", "content": "key={{api_key}}" } } } }
      """
    Then the call should succeed
    And the 1 "POST" request to "/echo" should have the header "X-Api-Key" equal to "k3y-s3cr3t-value"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 6
    And the response should not contain "k3y-s3cr3t-value"
    And the response should contain the field "1.summary" equal to "key is •••••"
    And the response should contain the field "1.masked" equal to "true"
    And the response should contain the field "3.summary" equal to "POST {{api}}/echo?key=•••••"
    And the response should contain the field "3.masked" equal to "true"
    And the response should contain the field "4.summary" equal to "echo: expected '•••••' to deeply equal 'other'"
    And the response should contain the field "5.summary" equal to "Error: leaked •••••"
    When I save the response field "0.id" as "nested"
    And I save the response field "2.id" as "logged"
    And I save the response field "3.id" as "main"
    And I call "ConsoleService.Get" with
      """
      { "id": "{{nested}}" }
      """
    Then the response should not contain "k3y-s3cr3t-value"
    And the response field "request.headers" should contain an item with "key" equal to "Authorization" and "value" equal to "Bearer •••••"
    When I call "ConsoleService.Get" with
      """
      { "id": "{{logged}}" }
      """
    Then the response should not contain "k3y-s3cr3t-value"
    And the response should contain the field "values.0.nested.key" equal to "•••••"
    When I call "ConsoleService.Get" with
      """
      { "id": "{{main}}" }
      """
    Then the response should not contain "k3y-s3cr3t-value"
    And the response should contain the field "masked" equal to "true"
    And the response should contain the field "request.url" equal to "{{api}}/echo?key=•••••"
    And the response field "request.headers" should contain an item with "key" equal to "X-Api-Key" and "value" equal to "•••••"
    And the response should contain the field "request.body" equal to "key=•••••"
    And the response field "response.headers" should contain an item with "key" equal to "X-Echo" and "value" equal to "•••••"
    And the response should contain the field "response.body" equal to
      """
      {"echo":"•••••"}
      """

  @set_mid_run
  Scenario: A secret value set by a script is masked in the entries that follow
    Given the 1 "GET" request to "/me" returns status 200 with the following response
      """
      {"session":"s3ss-from-login"}
      """
    And the 1 "GET" request to "/check" returns status 200 with the following response
      """
      {"ok":true}
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      scripts:
        postResponse: pm.environment.set("session", pm.response.json().session)
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/me"
      scripts:
        postResponse: |
          console.log("request sees", pm.environment.get("session"))
          await pm.sendRequest(pm.environment.replaceIn("{{base_url}}/check?session={{session}}"))
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/me" } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/check" should have the query "session" equal to "s3ss-from-login"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 3
    And the response should not contain "s3ss-from-login"
    And the response should contain the field "1.summary" equal to "request sees •••••"
    And the response should contain the field "1.masked" equal to "true"
    And the response should contain the field "2.summary" equal to "GET {{api}}/check?session=•••••"
    And the response should contain the field "2.masked" equal to "true"

  @set_mid_run_next_send
  Scenario: A secret set during one send is masked in the console of the next send
    Given the 1 "GET" request to "/me" returns status 200 with the following response
      """
      {"session":"s3ss-from-login"}
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/me"
      scripts:
        preRequest: console.log("before", pm.environment.get("session") || "none")
        postResponse: |
          pm.environment.set("session", pm.response.json().session)
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/me" } }
      """
    And I call "ConsoleService.Clear"
    And the 1 "GET" request to "/me" returns status 200 with the following response
      """
      {"session":"s3ss-from-login"}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/me" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should not contain "s3ss-from-login"
    And the response should contain the field "0.summary" equal to "before •••••"
