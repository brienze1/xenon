#language: en
#utf-8

@all @xen_005 @auth @oauth2 @oauth2_refresh
Feature: OAuth 2.0 token expiry and refresh
  A token is expired 30 seconds before its expiry. Before a send whose effective auth is OAuth 2.0
  with an expired token, Xenon refreshes it (refresh URL, or the access token URL, with the
  refresh parameters) when a refresh token exists and auto-refresh is on, and sends with the new
  token. Otherwise the send is refused with SND-07409 until it is acknowledged ("Send anyway").
  Concurrent sends sharing one expired token cause a single refresh: the others wait for it.

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

    And the current time is "2026-10-11T10:00:00Z"
    And the oauth provider is configured with
      """
      { "expires_in": 120 }
      """
    And I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """

  @auto_refresh
  Scenario: An expired token is refreshed before the send
    Given the current time is "2026-10-11T10:05:00Z"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the call should succeed
    And the oauth provider should have received 1 "refresh_token" token request
    And the 1 "refresh_token" token request should have the form field "refresh_token" equal to "refresh-1"
    And the 1 "refresh_token" token request should have the header "Authorization" equal to "Basic eGVub246czNjcmV0"
    And the 1 "GET" request to "/orders" should have the header "Authorization" equal to "Bearer access-2"
    When I call "OAuthService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    Then the response should contain the field "expired" equal to "false"
    And the response should contain the field "has_refresh_token" equal to "true"

  @auto_refresh
  Scenario: A token that has not expired is used as it is
    Given the current time is "2026-10-11T10:01:00Z"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the call should succeed
    And the oauth provider should have received 0 "refresh_token" token requests
    And the 1 "GET" request to "/orders" should have the header "Authorization" equal to "Bearer access-1"

  @auto_refresh @refresh_url
  Scenario: The refresh goes to the refresh token URL with the refresh parameters
    Given the workspace file "acme/collections/billing/collection.yaml" contains
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
          refreshTokenUrl: "{{oauth}}/token?via=refresh"
          clientId: xenon
          clientSecret: s3cret
          scope: api
          refreshRequestParams:
            - key: tenant
              value: acme
              enabled: true
              send_as: request_body
      """
    And the current time is "2026-10-11T10:05:00Z"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the call should succeed
    And the 1 "refresh_token" token request should have the query "via" equal to "refresh"
    And the 1 "refresh_token" token request should have the form field "tenant" equal to "acme"

  @expired @send_anyway
  Scenario: With auto-refresh off an expired token needs "Send anyway"
    Given the workspace file "acme/collections/billing/collection.yaml" contains
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
          autoRefresh: false
      """
    And the current time is "2026-10-11T10:05:00Z"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the call should fail with code "SND-07409"
    And the error should contain the field "error.error_details.0.attribute" equal to "owner_id"
    And the error should contain the field "error.error_details.0.messages.0" equal to "{{billing}}"
    And the "GET" request to "/orders" should have been received 0 times
    And the oauth provider should have received 0 "refresh_token" token requests
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "send_with_expired_token": true, "request": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/orders" should have the header "Authorization" equal to "Bearer access-1"

  @expired
  Scenario: Without a refresh token an expired token needs "Send anyway"
    Given the oauth provider is configured with
      """
      { "no_refresh_token": true }
      """
    And I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    And the current time is "2026-10-11T10:05:00Z"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the call should fail with code "SND-07409"

  @expired
  Scenario: A failed refresh asks for "Send anyway" and keeps the old token
    Given the oauth provider is configured with
      """
      { "token_error": "invalid_grant", "error_description": "refresh token revoked" }
      """
    And the current time is "2026-10-11T10:05:00Z"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the call should fail with code "SND-07409"
    And the oauth provider should have received 1 "refresh_token" token request
    When I call "OAuthService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}", "which": "access" }
      """
    Then the response should contain the field "value" equal to "access-1"

  @manual_refresh
  Scenario: Refresh obtains a new token on demand
    When I call "OAuthService.Refresh" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    Then the call should succeed
    And the response should contain the field "has_refresh_token" equal to "true"
    When I call "OAuthService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}", "which": "refresh" }
      """
    Then the response should contain the field "value" equal to "refresh-2"

  @manual_refresh
  Scenario: Refresh without a refresh token is refused
    Given the oauth provider is configured with
      """
      { "no_refresh_token": true }
      """
    And I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    When I call "OAuthService.Refresh" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    Then the call should fail with code "OAU-06404"

  @single_flight
  Scenario: Two concurrent sends with one expired token refresh it once
    Given the oauth provider is configured with
      """
      { "delay_ms": 400 }
      """
    And the current time is "2026-10-11T10:05:00Z"
    When I start calling "RequestService.Send" in the background with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-2", "request": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "{{api}}/orders", "auth": { "type": "inherit" } } }
      """
    Then the call should succeed
    And the background call should finish within 5000 ms
    And the call should succeed
    And the oauth provider should have received 1 "refresh_token" token request
    And the 1 "GET" request to "/orders" should have the header "Authorization" equal to "Bearer access-2"
    And the 2 "GET" request to "/orders" should have the header "Authorization" equal to "Bearer access-2"
