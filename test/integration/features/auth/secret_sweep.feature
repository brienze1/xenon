#language: en
#utf-8

@all @xen_005 @auth @stage_five_secret_sweep
Feature: OAuth, cookies, Digest and snippets never leave a secret value behind
  After a session that gets an OAuth 2.0 token and sends with it, stores persistent cookies set by
  a response, answers a Digest challenge with a secret password, lists the console, opens the
  Cookies dialog and generates snippets (masked and, on request, with the secrets), a search of
  the workspace folder, the unencrypted app data files, the logs, the console and the masked
  snippets finds no token, cookie or secret value (SC-004). Tokens and cookies live only in the
  encrypted vault.age and cookies.age.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the log output is captured
    And the "api" variable is set to the mocked api url
    And the "oauth" variable is set to the mocked oauth provider url
    And the "digest" variable is set to the mocked digest server url
    And the digest server is configured with
      """
      { "password": "dig3st-pa55word" }
      """
    And the requests to "/login" respond with the header "Set-Cookie" equal to "remember=c00kie-r3m3mb3r; Path=/; Max-Age=86400"
    And the "billing" variable is set to "01JS0000000000000000000001"
    And the "orders" variable is set to "01JS0000000000000000000002"
    And the "dev" variable is set to "01JS0000000000000000000009"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JS0000000000000000000000
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
      auth:
        type: oauth2
        oauth2:
          grant_type: client_credentials
          accessTokenUrl: "{{oauth}}/token"
          clientId: xenon
          clientSecret: "{{client_secret}}"
      """
    And the workspace file "acme/collections/billing/orders.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{orders}}
      name: Orders
      method: GET
      url: "{{api}}/orders"
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
        - key: client_secret
          secret: true
          enabled: true
          description: ""
        - key: digest_password
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
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "client_secret", "value": "cl13nt-s3cr3t" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "digest_password", "value": "dig3st-pa55word" }
      """

  @sweep
  Scenario: Tokens, cookies, Digest passwords and snippet secrets stay out of every plain place
    When I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    Then the call should succeed
    And the response should not contain "access-1"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/orders" should have the header "Authorization" equal to "Bearer access-1"
    And the response should not contain "access-1"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-2", "request": { "method": "POST", "url": "{{api}}/login", "auth": { "type": "none" } } }
      """
    Then the call should succeed
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-3", "request": { "method": "GET", "url": "{{digest}}/secure", "auth": { "type": "digest", "digest": { "username": "ana", "password": "{{digest_password}}" } } } }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "200"
    And the response should not contain "dig3st-pa55word"
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "0.cookies.0.value" equal to "c00kie-r3m3mb3r"
    When I call "SnippetService.Generate" with
      """
      { "local_id": "{{Acme.local_id}}", "language": "curl", "request": { "id": "{{orders}}", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the call should succeed
    And the response should contain the field "text" containing "Bearer •••••"
    And the response should not contain "access-1"
    When I call "SnippetService.Generate" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-3", "language": "python_requests", "include_secrets": true, "request": { "method": "GET", "url": "{{digest}}/secure", "auth": { "type": "digest", "digest": { "username": "ana", "password": "{{digest_password}}" } } } }
      """
    Then the call should succeed
    And the response should contain the field "text" matching "'Authorization': 'Digest username=.ana."
    When I call "SnippetService.Copy" with
      """
      { "local_id": "{{Acme.local_id}}", "language": "js_fetch", "request": { "id": "{{orders}}", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the call should succeed
    And the response should not contain "access-1"
    And the response should not contain "refresh-1"
    And the response should not contain "c00kie-r3m3mb3r"
    And the response should not contain "dig3st-pa55word"
    And the response should not contain "cl13nt-s3cr3t"
    And no file under the workspace folder should contain "access-1"
    And no file under the workspace folder should contain "refresh-1"
    And no file under the workspace folder should contain "c00kie-r3m3mb3r"
    And no file under the workspace folder should contain "dig3st-pa55word"
    And no file under the workspace folder should contain "cl13nt-s3cr3t"
    And no file under the app data folder should contain "access-1"
    And no file under the app data folder should contain "refresh-1"
    And no file under the app data folder should contain "c00kie-r3m3mb3r"
    And no file under the app data folder should contain "dig3st-pa55word"
    And no file under the app data folder should contain "cl13nt-s3cr3t"
    And the log output should not contain "access-1"
    And the log output should not contain "refresh-1"
    And the log output should not contain "c00kie-r3m3mb3r"
    And the log output should not contain "dig3st-pa55word"
    And the log output should not contain "cl13nt-s3cr3t"
    And the app data file "workspaces/{{Acme.local_id}}/cookies.age" should exist
