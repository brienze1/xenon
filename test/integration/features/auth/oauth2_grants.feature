#language: en
#utf-8

@all @xen_005 @auth @oauth2 @oauth2_grants
Feature: OAuth 2.0 grants and token use
  "Get new access token" obtains a token with the owner's OAuth 2.0 configuration (Postman's
  field names): client credentials and password grants call the token endpoint directly;
  authorization code (with or without PKCE) and implicit open the system browser on the auth URL
  and catch the redirect on a loopback listener at the callback URL. The token lives only in the
  workspace vault, keyed by the auth owner, and is never returned except by Reveal. Requests whose
  effective auth is OAuth 2.0 carry it as "Authorization: <prefix> <token>" or as the access_token
  query parameter; a "mac" token is not sent.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the "api" variable is set to the mocked api url
    And the "oauth" variable is set to the mocked oauth provider url
    And the "port" variable is set to a free local port
    And the "billing" variable is set to "01JG0000000000000000000001"
    And the "orders" variable is set to "01JG0000000000000000000002"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JG0000000000000000000000
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
          clientSecret: s3cret
          scope: api
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

  @client_credentials @client_authentication
  Scenario: Client credentials with the client sent as a Basic header
    When I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    Then the call should succeed
    And the response should contain the field "owner_id" equal to "{{billing}}"
    And the response should contain the field "token_type" equal to "Bearer"
    And the response should contain the field "scope" equal to "api"
    And the response should contain the field "has_refresh_token" equal to "true"
    And the response should contain the field "has_id_token" equal to "false"
    And the response should contain the field "use_token" equal to "access_token"
    And the response should not contain "access-1"
    And the oauth provider should have received 1 "client_credentials" token request
    And the 1 "client_credentials" token request should have the header "Authorization" equal to "Basic eGVub246czNjcmV0"
    And the 1 "client_credentials" token request should have the form field "scope" equal to "api"
    And the 1 "client_credentials" token request should not have the form field "client_secret"
    And the system browser should have opened 0 pages

  @client_credentials @client_authentication
  Scenario: Client credentials with the client sent in the body, with audience, resource and custom parameters
    When I call "OAuthService.GetNewToken" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "owner_id": "{{billing}}",
        "auth": {
          "grant_type": "client_credentials",
          "access_token_url": "{{oauth}}/token",
          "client_id": "xenon",
          "client_secret": "s3cret",
          "client_authentication": "body",
          "audience": "https://api.acme.test",
          "resource": "urn:acme:billing",
          "token_request_params": [
            { "key": "tenant", "value": "acme", "enabled": true, "send_as": "request_body" },
            { "key": "trace", "value": "t-1", "enabled": true, "send_as": "request_url" },
            { "key": "X-Region", "value": "eu", "enabled": true, "send_as": "request_header" },
            { "key": "skipped", "value": "x", "enabled": false, "send_as": "request_body" }
          ]
        }
      }
      """
    Then the call should succeed
    And the 1 "client_credentials" token request should have the form field "client_id" equal to "xenon"
    And the 1 "client_credentials" token request should have the form field "client_secret" equal to "s3cret"
    And the 1 "client_credentials" token request should not have the header "Authorization"
    And the 1 "client_credentials" token request should have the form field "audience" equal to "https://api.acme.test"
    And the 1 "client_credentials" token request should have the form field "resource" equal to "urn:acme:billing"
    And the 1 "client_credentials" token request should have the form field "tenant" equal to "acme"
    And the 1 "client_credentials" token request should have the query "trace" equal to "t-1"
    And the 1 "client_credentials" token request should have the header "X-Region" equal to "eu"
    And the 1 "client_credentials" token request should not have the form field "skipped"

  @password
  Scenario: Password grant sends the username and password
    When I call "OAuthService.GetNewToken" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "owner_id": "{{billing}}",
        "auth": { "grant_type": "password_credentials", "access_token_url": "{{oauth}}/token", "client_id": "xenon", "client_secret": "s3cret", "username": "ana", "password": "pw", "scope": "api" }
      }
      """
    Then the call should succeed
    And the 1 "password" token request should have the form field "username" equal to "ana"
    And the 1 "password" token request should have the form field "password" equal to "pw"
    And the 1 "password" token request should have the form field "scope" equal to "api"

  @vault
  Scenario: The token lives only in the vault and an inherited auth shares it
    When I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/orders" should have the header "Authorization" equal to "Bearer access-1"
    And no file under the workspace folder should contain "access-1"
    And no file under the workspace folder should contain "refresh-1"
    And no file under the app data folder should contain "access-1"
    And no file under the app data folder should contain "refresh-1"
    When I call "OAuthService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    Then the response should contain the field "token_type" equal to "Bearer"
    And the response should not contain "access-1"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should not contain "access-1"

  @reveal @copy @clear
  Scenario: Reveal and Copy hand out the token; Clear removes it
    Given I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    When I call "OAuthService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}", "which": "access" }
      """
    Then the response should contain the field "value" equal to "access-1"
    When I call "OAuthService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}", "which": "refresh" }
      """
    Then the response should contain the field "value" equal to "refresh-1"
    When I call "OAuthService.Copy" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}", "which": "access" }
      """
    Then the call should succeed
    And the clipboard should contain "access-1"
    When I call "OAuthService.Clear" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    Then the call should succeed
    When I call "OAuthService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    Then the response should contain the field "" equal to "null"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the 1 "GET" request to "/orders" should not have the header "Authorization"

  @authorization_code @pkce
  Scenario: Authorization code with PKCE (S256) and a generated verifier
    When I call "OAuthService.GetNewToken" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "owner_id": "{{billing}}",
        "auth": {
          "grant_type": "authorization_code_with_pkce",
          "auth_url": "{{oauth}}/authorize?prompt=login",
          "access_token_url": "{{oauth}}/token",
          "redirect_uri": "http://127.0.0.1:{{port}}/callback",
          "client_id": "xenon-desktop",
          "scope": "openid api",
          "auth_request_params": [ { "key": "login_hint", "value": "ana", "enabled": true, "send_as": "request_url" } ]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "has_id_token" equal to "true"
    And the system browser should have opened 1 page
    And the 1 page opened in the system browser should start with "{{oauth}}/authorize?prompt=login&"
    And the 1 oauth provider authorization request should have the query "response_type" equal to "code"
    And the 1 oauth provider authorization request should have the query "client_id" equal to "xenon-desktop"
    And the 1 oauth provider authorization request should have the query "redirect_uri" equal to "http://127.0.0.1:{{port}}/callback"
    And the 1 oauth provider authorization request should have the query "scope" equal to "openid api"
    And the 1 oauth provider authorization request should have the query "code_challenge_method" equal to "S256"
    And the 1 oauth provider authorization request should have the query "login_hint" equal to "ana"
    And the 1 "authorization_code" token request should have the form field "code" equal to "code-1"
    And the 1 "authorization_code" token request should have the form field "redirect_uri" equal to "http://127.0.0.1:{{port}}/callback"
    And the 1 "authorization_code" token request should have the form field "code_verifier" matching "^[A-Za-z0-9._~-]{43,128}$"

  @authorization_code @pkce
  Scenario: Authorization code with PKCE (plain) and a typed verifier
    When I call "OAuthService.GetNewToken" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "owner_id": "{{billing}}",
        "auth": {
          "grant_type": "authorization_code_with_pkce",
          "auth_url": "{{oauth}}/authorize",
          "access_token_url": "{{oauth}}/token",
          "redirect_uri": "http://localhost:{{port}}/cb",
          "client_id": "xenon-desktop",
          "challenge_algorithm": "plain",
          "code_verifier": "the-typed-verifier-0123456789-0123456789-0123456789"
        }
      }
      """
    Then the call should succeed
    And the 1 oauth provider authorization request should have the query "code_challenge" equal to "the-typed-verifier-0123456789-0123456789-0123456789"
    And the 1 oauth provider authorization request should have the query "code_challenge_method" equal to "plain"
    And the 1 "authorization_code" token request should have the form field "code_verifier" equal to "the-typed-verifier-0123456789-0123456789-0123456789"

  @authorization_code
  Scenario: Authorization code without PKCE sends no challenge, with a typed state
    When I call "OAuthService.GetNewToken" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "owner_id": "{{billing}}",
        "auth": {
          "grant_type": "authorization_code",
          "auth_url": "{{oauth}}/authorize",
          "access_token_url": "{{oauth}}/token",
          "redirect_uri": "http://127.0.0.1:{{port}}/callback",
          "client_id": "xenon-desktop",
          "client_secret": "s3cret",
          "state": "my-state"
        }
      }
      """
    Then the call should succeed
    And the 1 oauth provider authorization request should have the query "state" equal to "my-state"
    And the 1 oauth provider authorization request should not have the query "code_challenge"
    And the 1 "authorization_code" token request should not have the form field "code_verifier"
    And the 1 "authorization_code" token request should have the header "Authorization" equal to "Basic eGVub24tZGVza3RvcDpzM2NyZXQ="

  @implicit
  Scenario: Implicit grant takes the token from the redirect fragment
    When I call "OAuthService.GetNewToken" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "owner_id": "{{billing}}",
        "auth": { "grant_type": "implicit", "auth_url": "{{oauth}}/authorize", "redirect_uri": "http://127.0.0.1:{{port}}/callback", "client_id": "xenon-desktop", "scope": "api" }
      }
      """
    Then the call should succeed
    And the response should contain the field "has_refresh_token" equal to "false"
    And the 1 oauth provider authorization request should have the query "response_type" equal to "token"
    And the oauth provider should have received 0 "authorization_code" token requests
    When I call "OAuthService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}", "which": "access" }
      """
    Then the response should contain the field "value" equal to "access-1"

  @injection
  Scenario: The token replaces an Authorization header, with the configured prefix
    Given I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "tab-1", "auth": { "grant_type": "client_credentials", "access_token_url": "{{oauth}}/token", "client_id": "xenon", "client_secret": "s3cret" } }
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/things",
          "headers": [ { "key": "authorization", "value": "Basic old", "enabled": true } ],
          "auth": { "type": "oauth2", "oauth2": { "grant_type": "client_credentials", "access_token_url": "{{oauth}}/token", "header_prefix": "Token", "add_token_to": "header", "auto_refresh": true } }
        }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/things" should have the header "Authorization" equal to "Token access-1"

  @injection
  Scenario: The token goes to the access_token query parameter, replacing one
    Given I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "tab-1", "auth": { "grant_type": "client_credentials", "access_token_url": "{{oauth}}/token", "client_id": "xenon", "client_secret": "s3cret" } }
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/things?access_token=old&page=1",
          "params": [ { "key": "access_token", "value": "old", "enabled": true }, { "key": "page", "value": "1", "enabled": true } ],
          "auth": { "type": "oauth2", "oauth2": { "grant_type": "client_credentials", "access_token_url": "{{oauth}}/token", "header_prefix": "Bearer", "add_token_to": "queryParams", "auto_refresh": true } }
        }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/things" should have the raw query "page=1&access_token=access-1"
    And the 1 "GET" request to "/things" should not have the header "Authorization"

  @id_token
  Scenario: "Use token" switches to the ID token
    Given I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}", "auth": { "grant_type": "client_credentials", "access_token_url": "{{oauth}}/token", "client_id": "xenon", "client_secret": "s3cret", "scope": "openid" } }
      """
    When I call "OAuthService.Use" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}", "use_token": "id_token" }
      """
    Then the response should contain the field "use_token" equal to "id_token"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the 1 "GET" request to "/orders" should have the header "Authorization" equal to "Bearer id-1"

  @mac
  Scenario: A mac token is not sent
    Given the oauth provider is configured with
      """
      { "token_type": "mac" }
      """
    And I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/orders" should not have the header "Authorization"

  @postman_callback
  Scenario: A configuration with Postman's web callback explains the loopback URL to register
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      """
    And the workspace file "acme/collections/billing/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JG0000000000000000000003
      name: Login
      method: GET
      url: https://api.acme.test/me
      auth:
        type: oauth2
        oauth2:
          grant_type: authorization_code
          redirect_uri: https://oauth.pstmn.io/v1/callback
          authUrl: https://sso.acme.test/authorize
          accessTokenUrl: https://sso.acme.test/token
          clientId: xenon
      """
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JG0000000000000000000003" }
      """
    Then the response should contain the field "auth.oauth2.uses_postman_callback" equal to "true"
    And the response should contain the field "auth.oauth2.suggested_callback" equal to "http://127.0.0.1:53682/callback"
    When I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "01JG0000000000000000000003" }
      """
    Then the call should fail with code "OAU-01400"
    And the system browser should have opened 0 pages

  @legacy
  Scenario Outline: The <grant> grant is flagged as legacy in the configuration
    Given the workspace file "acme/collections/billing/legacy.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JG0000000000000000000004
      name: Legacy
      method: GET
      url: https://api.acme.test/me
      auth:
        type: oauth2
        oauth2:
          grant_type: <grant>
          accessTokenUrl: https://sso.acme.test/token
      """
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JG0000000000000000000004" }
      """
    Then the response should contain the field "auth.oauth2.grant_type" equal to "<grant>"
    And the response should contain the field "auth.oauth2.legacy" equal to "true"

    Examples:
      | grant                |
      | implicit             |
      | password_credentials |

  @safe_browser
  Scenario Outline: An auth URL that is not https (or http on a loopback host) is never opened
    When I call "OAuthService.GetNewToken" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "owner_id": "{{billing}}",
        "auth": { "grant_type": "authorization_code", "auth_url": "<auth_url>", "access_token_url": "{{oauth}}/token", "redirect_uri": "http://127.0.0.1:{{port}}/callback", "client_id": "xenon" }
      }
      """
    Then the call should fail with code "OAU-01400"
    And the system browser should have opened 0 pages

    Examples:
      | auth_url                        |
      | http://sso.acme.test/authorize  |
      | file:///etc/passwd              |
      | javascript:alert(1)             |
      | ftp://sso.acme.test/authorize   |
      | ms-settings:privacy             |

  @callback_url
  Scenario Outline: A callback URL that is not loopback http with a port is refused
    When I call "OAuthService.GetNewToken" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "owner_id": "{{billing}}",
        "auth": { "grant_type": "authorization_code", "auth_url": "{{oauth}}/authorize", "access_token_url": "{{oauth}}/token", "redirect_uri": "<callback>", "client_id": "xenon" }
      }
      """
    Then the call should fail with code "OAU-01400"
    And the system browser should have opened 0 pages

    Examples:
      | callback                         |
      | https://127.0.0.1:53682/callback |
      | http://192.168.1.10:53682/cb     |
      | http://127.0.0.1/callback        |
      | http://evil.test:53682/callback  |
