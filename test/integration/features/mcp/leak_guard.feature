#language: en
#utf-8

@all @xen_008 @mcp @mcp_leak_guard
Feature: No MCP result ever carries a secret's value
  Every tool result passes through one final leak guard that replaces each known secret value with "•••••" in its raw,
  URL-encoded, JSON-escaped and base64 forms (and basic-auth user:password). Known values are the active workspace's
  vault secrets and OAuth tokens plus every value seen this session. Bodies are masked whole before they are cut, so a
  secret on a cut point never comes back in halves. Files written on the AI's behalf from a response are masked too.
  The guard never changes what is sent, stored by the human's own actions, or shown in the UI.

  The secret used here is S3cr3t/Val+ue&ok. Its forms: raw S3cr3t/Val+ue&ok, query S3cr3t%2FVal%2Bue%26ok, path
  S3cr3t%2FVal+ue&ok, JSON-escaped S3cr3t/Val+ue&ok, base64 UzNjcjN0L1ZhbCt1ZSZvaw==, basic auth
  luis:S3cr3t/Val+ue&ok → bHVpczpTM2NyM3QvVmFsK3VlJm9r.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the log output is captured
    And the "api" variable is set to the mocked api url
    And the "secret" variable is set to "S3cr3t/Val+ue&ok"
    And the mock API echoes every request to "/echo"
    And the "dev" variable is set to "01JZ0000000000000000000010"
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
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    And I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_token", "secret": true, "value": "{{secret}}" }
      """

  @echo
  Scenario Outline: An echoed secret is masked in every form in the send result and the stored response - <form>
    When I call the MCP tool "send" with
      """
      {
        "request": {
          "method": "POST",
          "url": "{{base_url}}/echo?t={{api_token}}",
          "headers": [ { "key": "X-Token", "value": "{{api_token}}" } ],
          "auth": { "type": "basic", "basic": { "username": "luis", "password": "{{api_token}}" } },
          "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"t\":\"{{api_token}}\",\"b64\":\"UzNjcjN0L1ZhbCt1ZSZvaw==\"}" } }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "200"
    And the response should not contain "<value>"
    And the response should contain the field "body" containing "•••••"
    When I call the MCP tool "response" with
      """
      { "action": "last" }
      """
    Then the call should succeed
    And the response should not contain "<value>"
    When I call the MCP tool "response" with
      """
      { "action": "body" }
      """
    Then the call should succeed
    And the response should not contain "<value>"
    And the response should contain the field "text" containing "•••••"

    Examples:
      | form         | value                        |
      | raw          | S3cr3t/Val+ue&ok             |
      | query        | S3cr3t%2FVal%2Bue%26ok       |
      | path         | S3cr3t%2FVal+ue&ok           |
      | json-escaped | S3cr3t/Val+ue&ok        |
      | base64       | UzNjcjN0L1ZhbCt1ZSZvaw==     |
      | basic auth   | bHVpczpTM2NyM3QvVmFsK3VlJm9r |

  @target_unchanged
  Scenario: The target receives the real value and nothing on disk or in the log holds it unmasked
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/echo", "headers": [ { "key": "X-Token", "value": "{{api_token}}" } ] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/echo" should have the header "X-Token" equal to "{{secret}}"
    And the log output should not contain "{{secret}}"
    And no file under the workspace folder should contain "{{secret}}"
    And no file under the app data folder should contain "{{secret}}"

  @response_headers
  Scenario: A secret echoed back in a response header is masked
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/echo", "headers": [ { "key": "X-Token", "value": "{{api_token}}" } ] } }
      """
    Then the call should succeed
    And the response should not contain "S3cr3t"
    And the response should contain the field "headers" containing "X-Echo-X-Token"

  @scripts
  Scenario: Values a script copies into a test name, the console or a plain variable are masked
    When I call the MCP tool "send" with
      """
      {
        "request": {
          "method": "GET",
          "url": "{{base_url}}/echo",
          "scripts": {
            "pre_request": "",
            "post_response": "const t = pm.environment.get('api_token'); pm.test('token is ' + t, function () {}); console.log('token', t); pm.environment.set('copied', t);",
            "disabled": []
          }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "tests.0.name" equal to "token is •••••"
    And the response should not contain "S3cr3t"
    When I call the MCP tool "console" with
      """
      { "action": "list" }
      """
    Then the call should succeed
    And the response should not contain "S3cr3t"
    When I save the response field "items.0.id" as "entry"
    And I call the MCP tool "console" with
      """
      { "action": "get", "id": "{{entry}}" }
      """
    Then the response should not contain "S3cr3t"
    When I call the MCP tool "env" with
      """
      { "action": "get", "scope": "env:{{dev}}" }
      """
    Then the response field "vars" should contain an item with "key" equal to "copied" and "current" equal to "•••••"
    And the response should not contain "S3cr3t"
    When I call the MCP tool "snippet" with
      """
      { "action": "generate", "language": "curl", "request": { "method": "GET", "url": "{{base_url}}/x", "headers": [ { "key": "X-Copied", "value": "{{copied}}" } ] } }
      """
    Then the call should succeed
    And the response should not contain "S3cr3t"

  @history
  Scenario: A secret a human typed literally into a request is masked in History results
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/echo?leak={{secret}}" } }
      """
    And I call the MCP tool "history" with
      """
      { "action": "list" }
      """
    Then the call should succeed
    And the response should not contain "S3cr3t"
    When I save the response field "items.0.id" as "entry"
    And I call the MCP tool "history" with
      """
      { "action": "get", "id": "{{entry}}" }
      """
    Then the call should succeed
    And the response should not contain "S3cr3t"

  @tree
  Scenario: A secret typed literally into descriptions is masked in the tree, settings and documentation
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JZ0000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JZ0000000000000000000001
      name: Billing
      description: "Token for staging is {{secret}}"
      order:
        - me.yaml
      """
    And the workspace file "acme/collections/billing/me.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JZ0000000000000000000002
      name: Me
      description: "Call with {{secret}}"
      method: GET
      url: "{{api}}/me"
      """
    When I call the MCP tool "tree" with
      """
      { "action": "get", "id": "01JZ0000000000000000000002" }
      """
    Then the call should succeed
    And the response should not contain "S3cr3t"
    And the response should contain the field "description" equal to "Call with •••••"
    When I call the MCP tool "tree" with
      """
      { "action": "settings", "id": "01JZ0000000000000000000001" }
      """
    Then the response should not contain "S3cr3t"
    When I call the MCP tool "docs" with
      """
      { "action": "get", "collection_id": "01JZ0000000000000000000001" }
      """
    Then the call should succeed
    And the response should not contain "S3cr3t"

  @cookies
  Scenario: A cookie whose value equals a secret is masked
    Given the requests to "/echo" respond with the header "Set-Cookie" equal to "session={{secret}}; Path=/"
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/echo" } }
      """
    And I call the MCP tool "cookies" with
      """
      { "action": "list" }
      """
    Then the call should succeed
    And the response field "items" should contain an item with "name" equal to "session" and "value" equal to "•••••"
    And the response should not contain "S3cr3t"

  @errors
  Scenario: An error message that would contain a secret is masked
    Given I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "secret_host", "secret": true, "value": "leaky-internal-host-77" }
      """
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "http://{{secret_host}}.invalid/" } }
      """
    Then the call should succeed
    And the response should contain the field "error.message" equal to "Could not resolve host •••••.invalid"
    And the response should not contain "leaky-internal-host-77"
    When I call the MCP tool "env_edit" with
      """
      { "action": "set_current", "scope": "missing-scope", "key": "{{secret}}", "value": "x" }
      """
    Then the error should not contain "S3cr3t"

  @slicing
  Scenario: A secret across the 16 KB cut point never comes back in two halves
    Given the "filler" variable is set to "a" repeated 16380 times separated by ""
    And the 1 "GET" request to "/boundary" returns status 200 with content type "text/plain" and body
      """
      {{filler}}{{secret}} and the rest of the body
      """
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/boundary" } }
      """
    Then the call should succeed
    And the response should contain the field "truncated" equal to "true"
    And the response should not contain "S3cr"
    And the response should not contain "Val+ue"
    When I call the MCP tool "response" with
      """
      { "action": "body", "offset": 16384, "length": 1000 }
      """
    Then the call should succeed
    And the response should not contain "Val+ue"
    And the response should not contain "ue&ok"
    And the response should not contain "ue&ok"

  @slicing
  Scenario: A secret across a body read boundary never comes back in two halves
    Given the "filler" variable is set to "b" repeated 100 times separated by ""
    And the 1 "GET" request to "/page" returns status 200 with content type "text/plain" and body
      """
      {{filler}}{{secret}}{{filler}}
      """
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/page" } }
      """
    And I call the MCP tool "response" with
      """
      { "action": "body", "offset": 0, "length": 104 }
      """
    Then the call should succeed
    And the response should not contain "S3cr"
    When I call the MCP tool "response" with
      """
      { "action": "body", "offset": 104, "length": 50 }
      """
    Then the response should not contain "Val+ue"
    And the response should not contain "ue&ok"

  @replaced
  Scenario: A value replaced during the session stays masked in an older response
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/echo", "headers": [ { "key": "X-Token", "value": "{{api_token}}" } ] } }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_token", "value": "brand-new-token-9" }
      """
    And I call the MCP tool "response" with
      """
      { "action": "last" }
      """
    Then the call should succeed
    And the response should not contain "S3cr3t"

  @oauth
  Scenario: An OAuth 2.0 access token used by a send is masked
    Given the "oauth" variable is set to the mocked oauth provider url
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JZ0000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JZ0000000000000000000001
      name: Billing
      auth:
        type: oauth2
        oauth2:
          grant_type: client_credentials
          accessTokenUrl: "{{oauth}}/token"
          clientId: xenon
          clientSecret: s3cret
          scope: api
      order:
        - echo.yaml
      """
    And the workspace file "acme/collections/billing/echo.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JZ0000000000000000000003
      name: Echo
      method: GET
      url: "{{api}}/echo"
      auth:
        type: inherit
      """
    And I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "01JZ0000000000000000000001" }
      """
    When I call the MCP tool "send" with
      """
      { "request_id": "01JZ0000000000000000000003" }
      """
    Then the call should succeed
    And the 1 "GET" request to "/echo" should have the header "Authorization" equal to "Bearer access-1"
    And the response should not contain "access-1"
    When I call the MCP tool "tree" with
      """
      { "action": "token", "owner_id": "01JZ0000000000000000000001" }
      """
    Then the call should succeed
    And the response should contain the field "type" equal to "Bearer"
    And the response should not contain "access-1"

  @ai_written_files
  Scenario: A response body saved by the AI is written masked
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/echo", "headers": [ { "key": "X-Token", "value": "{{api_token}}" } ] } }
      """
    And I call the MCP tool "export" with
      """
      { "action": "response_body", "path": "{{temp_root}}/echo.json" }
      """
    Then the call should succeed
    And the file "echo.json" should exist
    And the file "echo.json" should not contain "S3cr3t"

  @ai_written_files
  Scenario: An example the AI creates from a response is written masked
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JZ0000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JZ0000000000000000000001
      name: Billing
      order:
        - echo.yaml
      """
    And the workspace file "acme/collections/billing/echo.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JZ0000000000000000000003
      name: Echo
      method: GET
      url: "{{api}}/echo"
      headers:
        - key: X-Token
          value: "{{api_token}}"
          enabled: true
          description: ""
      """
    When I call the MCP tool "send" with
      """
      { "request_id": "01JZ0000000000000000000003" }
      """
    And I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "example", "request_id": "01JZ0000000000000000000003", "name": "Echoed", "from_last_response": true }
      """
    Then the call should succeed
    And no file under the workspace folder should contain "S3cr3t"

  @ai_written_files
  Scenario: A binary body that contains a secret is not written for the AI
    Given the 1 "GET" request to "/blob" returns status 200 with content type "application/octet-stream" and body
      """
      binary-prefix {{secret}} binary-suffix
      """
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/blob" } }
      """
    And I call the MCP tool "export" with
      """
      { "action": "response_body", "path": "{{temp_root}}/blob.bin" }
      """
    Then the call should fail with code "MCP-02403"
    And the file "blob.bin" should not exist

  @export_files
  Scenario Outline: Exported files never hold a secret value - <format>
    When I call the MCP tool "export" with
      """
      { "action": "run", "target": { "kind": "environment", "id": "{{dev}}" }, "format": "<format>", "path": "{{temp_root}}/dev.<extension>" }
      """
    Then the call should succeed
    And the file "dev.<extension>" should exist
    And the file "dev.<extension>" should not contain "S3cr3t"
    And the file "dev.<extension>" should not contain "UzNjcjN0L1ZhbCt1ZSZvaw=="

    Examples:
      | format  | extension |
      | xenon   | yaml      |
      | postman | json      |
