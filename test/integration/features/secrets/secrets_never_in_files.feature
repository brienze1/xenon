#language: en
#utf-8

@all @xen_002 @secrets @secrets_never_in_files
Feature: Secret values never reach the workspace folder, local files, responses or logs
  A secret row stores only its key, "secret: true", enabled and description in the workspace
  files. Its value lives in the workspace's age-encrypted vault (app data
  workspaces/<localId>/vault.age). No other file, no bridge response other than
  VaultService.Reveal, no History entry and no log line ever carries the value; text Xenon
  produces from a resolved request shows "•••••" instead.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the "billing" variable is set to "01JE0000000000000000000001"
    And the "dev" variable is set to "01JE0000000000000000000010"
    And the "api" variable is set to the mocked api url
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JE0000000000000000000000
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
      """
    And I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "env:{{dev}}",
        "rows": [
          { "key": "base_url", "initial_value": "{{api}}", "enabled": true, "description": "" },
          { "key": "api_key", "secret": true, "secret_value": "env-s3cr3t-value", "enabled": true, "description": "vendor" }
        ]
      }
      """
    And I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "collection:{{billing}}",
        "rows": [ { "key": "token", "secret": true, "secret_value": "collection-s3cr3t-value", "enabled": true, "description": "" } ]
      }
      """
    And I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "globals",
        "rows": [ { "key": "password", "secret": true, "secret_value": "globals-s3cr3t-value", "enabled": true, "description": "" } ]
      }
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """

  @files
  Scenario: The workspace files list secret rows without values
    Then the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.key" equal to "api_key"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.secret" equal to "true"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.description" equal to "vendor"
    And the workspace file "acme/environments/dev.yaml" should not contain the field "variables.1.value"
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "variables.0.secret" equal to "true"
    And the workspace file "acme/collections/billing/collection.yaml" should not contain the field "variables.0.value"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.0.secret" equal to "true"
    And the workspace file "acme/workspace.yaml" should not contain the field "variables.0.value"
    And no file under the workspace folder should contain "env-s3cr3t-value"
    And no file under the workspace folder should contain "collection-s3cr3t-value"
    And no file under the workspace folder should contain "globals-s3cr3t-value"

  @files
  Scenario: Only the encrypted vault holds the values in app data
    Then the app data file "workspaces/{{Acme.local_id}}/vault.age" should exist
    And the app data path "workspaces/{{Acme.local_id}}/vault.age" should have the permissions "600"
    And the app data file "workspaces/{{Acme.local_id}}/vault.age" should not contain "env-s3cr3t-value"
    And the app data file "workspaces/{{Acme.local_id}}/vault.age" should not contain "api_key"
    And no file under the app data folder should contain "env-s3cr3t-value"
    And no file under the app data folder should contain "collection-s3cr3t-value"
    And no file under the app data folder should contain "globals-s3cr3t-value"

  @responses
  Scenario: Tables, environments and lists never return the values
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}" }
      """
    Then the response should contain the field "rows.1.secret_state" equal to "set"
    And the response should contain the field "rows.1.current_value" equal to ""
    And the response should contain the field "rows.1.initial_value" equal to ""
    And the response should contain the field "rows.1.has_current_override" equal to "true"
    And the response should not contain "env-s3cr3t-value"
    When I call "EnvironmentService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    Then the response should not contain "env-s3cr3t-value"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should not contain "env-s3cr3t-value"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "collection:{{billing}}" }
      """
    Then the response should not contain "collection-s3cr3t-value"
    When I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "password", "value": "rotated-s3cr3t-value" }
      """
    Then the call should succeed
    And the response should contain the field "secret_state" equal to "set"
    And the response should not contain "rotated-s3cr3t-value"
    When I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "pin", "value": "added-s3cr3t-value", "secret": true }
      """
    Then the call should succeed
    And the response should contain the field "secret" equal to "true"
    And the response should not contain "added-s3cr3t-value"
    And no file under the workspace folder should contain "added-s3cr3t-value"
    And no file under the app data folder should contain "added-s3cr3t-value"

  @send
  Scenario: Secrets resolve when sending while History and logs keep them out
    Given the log output is captured
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{base_url}}/secure/{{api_key}}?token={{token}}",
          "headers": [ { "key": "X-Password", "value": "{{password}}" } ]
        }
      }
      """
    Then the call should succeed
    And the "GET" request to "/secure/env-s3cr3t-value" should have been received 1 times
    And the 1 "GET" request to "/secure/env-s3cr3t-value" should have the query "token" equal to "collection-s3cr3t-value"
    And the 1 "GET" request to "/secure/env-s3cr3t-value" should have the header "X-Password" equal to "globals-s3cr3t-value"
    And the response should not contain "s3cr3t-value"
    And the log output should contain "Finished"
    And the log output should not contain "env-s3cr3t-value"
    And the log output should not contain "collection-s3cr3t-value"
    And the log output should not contain "globals-s3cr3t-value"
    And no file under the app data folder should contain "s3cr3t-value"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.url" equal to "{{base_url}}/secure/{{api_key}}?token={{token}}"
    And the response should not contain "s3cr3t-value"

  @send @masking
  Scenario: A failed send shows the secret masked in the error message and the History summary
    Given the "closed" variable is set to a closed port url
    And I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "env:{{dev}}",
        "rows": [
          { "key": "base_url", "initial_value": "{{api}}", "enabled": true, "description": "" },
          { "key": "api_key", "secret": true, "enabled": true, "description": "vendor" },
          { "key": "secret_host", "secret": true, "secret_value": "unreachable-secret.invalid", "enabled": true, "description": "" }
        ]
      }
      """
    And the log output is captured
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "http://{{secret_host}}/ping" } }
      """
    Then the call should succeed
    And the response should contain the field "ok" equal to "false"
    And the response should contain the field "error.kind" equal to "dns"
    And the response should contain the field "error.message" equal to "Could not resolve host •••••"
    And the response should not contain "unreachable-secret"
    And the log output should not contain "unreachable-secret"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should not contain "unreachable-secret"
    And no file under the app data folder should contain "unreachable-secret"
