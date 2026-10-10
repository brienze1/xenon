#language: en
#utf-8

@all @xen_002 @secrets @secret_sweep
Feature: A full session never leaves a secret value anywhere
  After a whole session that creates secrets, sends with them (successfully, to a host that does
  not resolve and into a timeout), lists History, duplicates the environment, turns a plain row
  into a secret, inspects and restarts, a search of the workspace folder, the app data folder
  (except the encrypted *.age files), the logs, History and every bridge response for the secret
  values finds nothing (SC-002, FR-023). Only VaultService.Reveal returns a value, on request.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the log output is captured
    And the "api" variable is set to the mocked api url
    And the "dev" variable is set to "01JW0000000000000000000010"
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

  @sweep
  Scenario: Create, send, fail, list, duplicate, switch type, inspect and restart leave no trace
    When I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "env:{{dev}}",
        "rows": [
          { "key": "base_url", "initial_value": "{{api}}", "enabled": true, "description": "" },
          { "key": "api_key", "secret": true, "secret_value": "sweep-s3cr3t-api-key", "enabled": true, "description": "" },
          { "key": "secret_host", "secret": true, "secret_value": "sweep-s3cr3t-host.invalid", "enabled": true, "description": "" },
          { "key": "plain_token", "initial_value": "sweep-plain-value", "enabled": true, "description": "" }
        ]
      }
      """
    Then the call should succeed
    And the response should not contain "sweep-s3cr3t-api-key"
    And the response should not contain "sweep-s3cr3t-host"
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    And I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "env:{{dev}}",
        "rows": [
          { "key": "base_url", "initial_value": "{{api}}", "enabled": true, "description": "" },
          { "key": "api_key", "secret": true, "enabled": true, "description": "" },
          { "key": "secret_host", "secret": true, "enabled": true, "description": "" },
          { "key": "plain_token", "secret": true, "secret_value": "sweep-plain-value", "enabled": true, "description": "" }
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "warnings" with the items "git_history"
    And the response should not contain "sweep-s3cr3t"
    And the response should not contain "sweep-plain-value"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{base_url}}/orders?key={{api_key}}", "headers": [ { "key": "Authorization", "value": "Bearer {{plain_token}}", "enabled": true } ], "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"key\": \"{{api_key}}\"}" } } } }
      """
    Then the call should succeed
    And the response should contain the field "ok" equal to "true"
    And the 1 "POST" request to "/orders" should have the query "key" equal to "sweep-s3cr3t-api-key"
    And the 1 "POST" request to "/orders" should have the header "Authorization" equal to "Bearer sweep-plain-value"
    And the response should not contain "sweep-s3cr3t"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "https://{{secret_host}}/" } }
      """
    Then the call should succeed
    And the response should contain the field "error.kind" equal to "dns"
    And the response should contain the field "error.message" equal to "Could not resolve host •••••"
    And the response should not contain "sweep-s3cr3t"
    Given the "GET" requests to "/slow" respond after 1000ms
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{base_url}}/slow?key={{api_key}}", "settings": { "timeout_ms": 100 } } }
      """
    Then the call should succeed
    And the response should contain the field "error.kind" equal to "timeout"
    And the response should not contain "sweep-s3cr3t"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 3
    And the response should not contain "sweep-s3cr3t"
    When I call "EnvironmentService.Duplicate" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    Then the call should succeed
    And the response should not contain "sweep-s3cr3t"
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "{{base_url}}/orders?key={{api_key}}", "headers": [ { "key": "Host", "value": "{{secret_host}}", "enabled": true } ] } }
      """
    Then the call should succeed
    And the response should not contain "sweep-s3cr3t"
    When I call "VariableService.Suggest" with
      """
      { "local_id": "{{Acme.local_id}}", "prefix": "" }
      """
    Then the call should succeed
    And the response should not contain "sweep-s3cr3t"
    Given the app is restarted
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{base_url}}/orders?key={{api_key}}" } }
      """
    Then the call should succeed
    And the 2 "POST" request to "/orders" should have the query "key" equal to "sweep-s3cr3t-api-key"
    And the response should not contain "sweep-s3cr3t"
    When I call "EnvironmentService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    Then the response should not contain "sweep-s3cr3t"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}" }
      """
    Then the response should not contain "sweep-s3cr3t"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should not contain "sweep-s3cr3t"
    And no file under the workspace folder should contain "sweep-s3cr3t"
    And no file under the app data folder should contain "sweep-s3cr3t"
    And no file under the workspace folder should contain "sweep-plain-value"
    And no file under the app data folder should contain "sweep-plain-value"
    And the log output should contain "Finished"
    And the log output should not contain "sweep-s3cr3t"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "plain_token" }
      """
    Then the response should contain the field "value" equal to "sweep-plain-value"
