#language: en
#utf-8

@all @xen_005 @auth @oauth2 @oauth2_errors
Feature: OAuth 2.0 errors
  Getting a token fails with a clear code and stores nothing: an invalid configuration
  (OAU-01400), a callback port already in use (OAU-02409), no redirect within the time limit
  (OAU-03408, 5 minutes; shortened here), an error from the provider with its error and
  description (OAU-04401), a state that does not match (OAU-05400) and a cancelled wait
  (OAU-08499). The loopback listener stops in every case.

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


  @provider_error
  Scenario: The provider refuses the authorization
    Given the oauth provider is configured with
      """
      { "authorize_error": "access_denied", "error_description": "The user said no" }
      """
    When I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}", "auth": { "grant_type": "authorization_code", "auth_url": "{{oauth}}/authorize", "access_token_url": "{{oauth}}/token", "redirect_uri": "http://127.0.0.1:{{port}}/callback", "client_id": "xenon" } }
      """
    Then the call should fail with code "OAU-04401"
    And the error should contain the field "error.error_details.0.attribute" equal to "error"
    And the error should contain the field "error.error_details.0.messages.0" equal to "access_denied"
    And the error should contain the field "error.error_details.1.attribute" equal to "error_description"
    And the error should contain the field "error.error_details.1.messages.0" equal to "The user said no"
    When I call "OAuthService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    Then the response should contain the field "" equal to "null"

  @provider_error
  Scenario: The token endpoint refuses the client
    Given the oauth provider is configured with
      """
      { "token_error": "invalid_client", "error_description": "Unknown client" }
      """
    When I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    Then the call should fail with code "OAU-04401"
    And the error should contain the field "error.error_details.0.messages.0" equal to "invalid_client"

  @state
  Scenario: A redirect with another state is rejected
    Given the oauth provider is configured with
      """
      { "wrong_state": true }
      """
    When I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}", "auth": { "grant_type": "authorization_code", "auth_url": "{{oauth}}/authorize", "access_token_url": "{{oauth}}/token", "redirect_uri": "http://127.0.0.1:{{port}}/callback", "client_id": "xenon" } }
      """
    Then the call should fail with code "OAU-05400"
    And the oauth provider should have received 0 "authorization_code" token requests
    When I call "OAuthService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    Then the response should contain the field "" equal to "null"

  @port
  Scenario: A callback port in use is reported before the browser opens
    Given the local port "{{port}}" is in use
    When I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}", "auth": { "grant_type": "authorization_code", "auth_url": "{{oauth}}/authorize", "access_token_url": "{{oauth}}/token", "redirect_uri": "http://127.0.0.1:{{port}}/callback", "client_id": "xenon" } }
      """
    Then the call should fail with code "OAU-02409"
    And the system browser should have opened 0 pages

  @timeout
  Scenario: A login abandoned in the browser times out and frees the port
    Given the "XENON_OAUTH_CALLBACK_TIMEOUT_MS" env var is set to "300"
    And the system browser ignores the pages it opens
    When I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}", "auth": { "grant_type": "authorization_code", "auth_url": "{{oauth}}/authorize", "access_token_url": "{{oauth}}/token", "redirect_uri": "http://127.0.0.1:{{port}}/callback", "client_id": "xenon" } }
      """
    Then the call should fail with code "OAU-03408"
    And the last call should have taken less than 3000 ms
    And the system browser should have opened 1 page
    And the local port "{{port}}" is in use

  @cancel
  Scenario: Cancelling the wait stops it at once
    Given the system browser ignores the pages it opens
    When I call "OAuthService.GetNewToken" and cancel it after 200ms with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}", "auth": { "grant_type": "authorization_code", "auth_url": "{{oauth}}/authorize", "access_token_url": "{{oauth}}/token", "redirect_uri": "http://127.0.0.1:{{port}}/callback", "client_id": "xenon" } }
      """
    Then the call should fail with code "OAU-08499"
    And the last call should have taken less than 3000 ms
    And the local port "{{port}}" is in use

  @config
  Scenario Outline: An incomplete configuration is refused: <case>
    When I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}", "auth": <auth> }
      """
    Then the call should fail with code "OAU-01400"
    And the oauth provider should have received 0 "client_credentials" token requests

    Examples:
      | case                     | auth                                                                                         |
      | no access token URL      | { "grant_type": "client_credentials", "client_id": "xenon" }                                 |
      | no auth URL for the code | { "grant_type": "authorization_code", "access_token_url": "{{oauth}}/token" }                |

  @config
  Scenario: An owner whose auth is not OAuth 2.0 has no token to get
    When I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{orders}}" }
      """
    Then the call should fail with code "OAU-01400"
