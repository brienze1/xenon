#language: en
#utf-8

@all @xen_002 @sending @preflight_missing_secrets
Feature: Sending with missing or locked secrets asks first
  Before anything is sent Go checks every secret the request uses (directly, nested in another
  variable's value, or through inherited auth and headers). If any has no value on this machine
  or sits in a locked vault, Send is refused with SND-06409 listing them: nothing goes out and
  nothing is recorded in History. The UI then offers "Send anyway", which re-sends with
  acknowledge_missing_secrets and the placeholders go out literally. Unused secrets never block.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "dev" variable is set to "01JG0000000000000000000010"
    And the "billing" variable is set to "01JG0000000000000000000001"
    And the "list" variable is set to "01JG0000000000000000000002"
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
        type: bearer
        bearer:
          token: "{{token}}"
      """
    And the workspace file "acme/collections/billing/list.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{list}}
      name: List
      method: GET
      url: "{{api}}/list"
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
        - key: api_key
          secret: true
          enabled: true
          description: ""
        - key: token
          secret: true
          enabled: true
          description: ""
        - key: auth_header
          value: "Key {{api_key}}"
          enabled: true
          description: ""
        - key: unused
          secret: true
          enabled: true
          description: ""
      """
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" contains
      """
      { "version": 1, "activeEnvironmentId": "{{dev}}", "current": {} }
      """

  @missing
  Scenario: A request using missing secrets is refused and nothing is sent
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/direct", "headers": [ { "key": "X-Api-Key", "value": "{{api_key}}" } ] } }
      """
    Then the call should fail with code "SND-06409"
    And the error should contain the field "error.error_details.0.attribute" equal to "missing_secrets"
    And the error should contain the field "error.error_details.0.messages.0" equal to "api_key"
    And the "GET" request to "/direct" should have been received 0 times
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0

  @missing
  Scenario: Send anyway sends the placeholders literally
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "acknowledge_missing_secrets": true, "request": { "method": "GET", "url": "{{api}}/direct", "headers": [ { "key": "X-Api-Key", "value": "{{api_key}}" } ] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/direct" should have the header "X-Api-Key" equal to "{{api_key}}"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 1

  @missing @nested
  Scenario: Secrets used through other variables and inherited auth are found
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{list}}", "method": "GET", "url": "{{api}}/list", "auth": { "type": "inherit" }, "headers": [ { "key": "Authorization-Extra", "value": "{{auth_header}}" } ] } }
      """
    Then the call should fail with code "SND-06409"
    And the error should contain the field "error.error_details.0.messages.0" equal to "api_key"
    And the error should contain the field "error.error_details.0.messages.1" equal to "token"
    And the "GET" request to "/list" should have been received 0 times

  @missing
  Scenario: Unused missing secrets do not block a send
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/plain" } }
      """
    Then the call should succeed
    And the "GET" request to "/plain" should have been received 1 times

  @missing
  Scenario: Once the secret has a value the request is sent with it
    Given the keychain is available
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key", "value": "real-key" }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/direct", "headers": [ { "key": "X-Api-Key", "value": "{{auth_header}}" } ] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/direct" should have the header "X-Api-Key" equal to "Key real-key"

  @locked
  Scenario: Secrets of a locked vault block the send the same way
    Given the keychain is unavailable
    And I call "VaultService.SetupPassphrase" with
      """
      { "passphrase": "correct horse", "confirmation": "correct horse" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key", "value": "real-key" }
      """
    And the app is restarted
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/direct", "headers": [ { "key": "X-Api-Key", "value": "{{api_key}}" } ] } }
      """
    Then the call should fail with code "SND-06409"
    And the error should contain the field "error.error_details.0.attribute" equal to "locked_secrets"
    And the error should contain the field "error.error_details.0.messages.0" equal to "api_key"
    And the "GET" request to "/direct" should have been received 0 times
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "acknowledge_missing_secrets": true, "request": { "method": "GET", "url": "{{api}}/direct", "headers": [ { "key": "X-Api-Key", "value": "{{api_key}}" } ] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/direct" should have the header "X-Api-Key" equal to "{{api_key}}"
    When I call "VaultService.Unlock" with
      """
      { "passphrase": "correct horse" }
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/direct", "headers": [ { "key": "X-Api-Key", "value": "{{api_key}}" } ] } }
      """
    Then the call should succeed
    And the 2 "GET" request to "/direct" should have the header "X-Api-Key" equal to "real-key"
