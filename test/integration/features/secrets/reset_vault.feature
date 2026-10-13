#language: en
#utf-8

@all @xen_002 @secrets @reset_vault
Feature: A vault that cannot be opened is reported broken and can be reset
  When a workspace's vault cannot be decrypted (its keychain entry or key file is gone, the vault
  file is corrupted, or it was deleted while its identity remains) Xenon never guesses: the
  workspace is listed in broken_local_ids, its secrets are missing, and new secret values are
  refused until "Reset vault" is confirmed, which discards the stored values and starts empty.
  Removing a workspace deletes its vault, key file, keychain entry and local current values.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    And the keychain is available
    And the "api" variable is set to the mocked api url
    And I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "acme-key", "enabled": true } ] }
      """
    And I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Globex.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "globex-key", "enabled": true } ] }
      """

  @broken
  Scenario Outline: A vault that cannot be opened is broken - <case>
    Given <damage>
    When I call "VaultService.Status"
    Then the response should contain the field "broken_local_ids" with the items "{{Acme.local_id}}"
    And the response should contain the field "locked_local_ids" equal to "[]"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals" }
      """
    Then the response should contain the field "rows.0.secret_state" equal to "missing"
    When I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "api_key", "value": "new-key" }
      """
    Then the call should fail with code "VLT-06409"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "api_key" }
      """
    Then the call should fail with code "VLT-06409"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/me", "headers": [ { "key": "X-Api-Key", "value": "{{api_key}}" } ] } }
      """
    Then the call should fail with code "SND-06409"
    And the error should contain the field "error.error_details.0.attribute" equal to "missing_secrets"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Globex.local_id}}", "scope": "globals", "key": "api_key" }
      """
    Then the response should contain the field "value" equal to "globex-key"

    Examples:
      | case                  | damage                                                                                       |
      | keychain entry lost   | the keychain entry for the account "{{Acme.local_id}}" is removed                            |
      | vault file deleted    | the app data path "workspaces/{{Acme.local_id}}/vault.age" is removed                        |
      | vault file corrupted  | the app data file "workspaces/{{Acme.local_id}}/vault.age" contains the text "not an age file" |

  @reset
  Scenario: Resetting a broken vault needs confirmation and starts an empty vault
    Given the keychain entry for the account "{{Acme.local_id}}" is removed
    When I call "VaultService.Reset" with
      """
      { "local_id": "{{Acme.local_id}}", "confirm": false }
      """
    Then the call should fail with code "VLT-07400"
    When I call "VaultService.Reset" with
      """
      { "local_id": "{{Acme.local_id}}", "confirm": true }
      """
    Then the call should succeed
    And the response should contain the field "broken_local_ids" equal to "[]"
    And the event "vault:status-changed" should have been emitted 2 times
    And the app data file "workspaces/{{Acme.local_id}}/vault.age" should not exist
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals" }
      """
    Then the response should contain the field "rows.0.secret_state" equal to "missing"
    When I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "api_key", "value": "new-key" }
      """
    Then the call should succeed
    And the response should contain the field "secret_state" equal to "set"
    And the keychain should contain an entry for the account "{{Acme.local_id}}"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "api_key" }
      """
    Then the response should contain the field "value" equal to "new-key"

  @remove_workspace
  Scenario: Removing a workspace deletes its vault, key, keychain entry and local values
    Given I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "enabled": true }, { "key": "user", "initial_value": "luis", "set_current": true, "current_value": "brienze", "enabled": true } ] }
      """
    When I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Acme.local_id}}", "move_to_trash": false }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/vault.age" should not exist
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not exist
    And the keychain should not contain an entry for the account "{{Acme.local_id}}"
    And the keychain should contain an entry for the account "{{Globex.local_id}}"
    And the workspace file "acme/workspace.yaml" should exist
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Globex.local_id}}", "scope": "globals", "key": "api_key" }
      """
    Then the response should contain the field "value" equal to "globex-key"

  @remove_workspace
  Scenario: Removing a workspace continues when its keychain entry cannot be deleted
    Given the keychain fails with "access denied"
    When I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Acme.local_id}}", "move_to_trash": false }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/vault.age" should not exist

  @errors
  Scenario: Resetting the vault of an unknown workspace fails
    When I call "VaultService.Reset" with
      """
      { "local_id": "01JH0000000000000000000000", "confirm": true }
      """
    Then the call should fail with code "WKS-07404"
