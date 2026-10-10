#language: en
#utf-8

@all @xen_002 @secrets @keychain_mode
Feature: Vaults unlock through the OS keychain
  The first secret saved in the app probes the OS keychain (write, read and delete of a test
  entry). When it works the app uses keychain mode: each workspace vault gets its own age
  identity stored in the keychain under service "dev.brienze.xenon.vault" with the workspace's
  local id as the account, so after a restart secrets resolve without typing anything. Reveal is
  the only call that returns a secret value; CopySecret puts it on the clipboard from Go.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    And the keychain is available
    And the "api" variable is set to the mocked api url

  @mode
  Scenario: The first secret chooses keychain mode and stores the vault identity in the keychain
    When I call "VaultService.Status"
    Then the response should contain the field "mode" equal to "none"
    And the response should contain the field "needs_passphrase_setup" equal to "false"
    When I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "acme-key", "enabled": true } ] }
      """
    Then the call should succeed
    And the keychain should contain an entry for the account "{{Acme.local_id}}"
    And the keychain should not contain an entry for the account "probe"
    And the keychain should not contain an entry for the account "{{Globex.local_id}}"
    And the app data file "app.json" should contain the field "vault.mode" equal to "keychain"
    And the app data file "workspaces/{{Acme.local_id}}/vault-key.age" should not exist
    And the event "vault:status-changed" should have been emitted 1 times
    And the 1 "vault:status-changed" event should contain the field "mode" equal to "keychain"
    When I call "VaultService.Status"
    Then the response should contain the field "mode" equal to "keychain"
    And the response should contain the field "locked_local_ids" equal to "[]"
    And the response should contain the field "broken_local_ids" equal to "[]"

  @restart
  Scenario: After a restart secrets resolve with no input
    Given I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "acme-key", "enabled": true } ] }
      """
    When the app is restarted
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/me", "headers": [ { "key": "X-Api-Key", "value": "{{api_key}}" } ] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/me" should have the header "X-Api-Key" equal to "acme-key"
    When I call "VaultService.Status"
    Then the response should contain the field "mode" equal to "keychain"
    And the response should contain the field "locked_local_ids" equal to "[]"

  @reveal
  Scenario: Reveal returns the value and CopySecret puts it on the clipboard
    Given I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "acme-key", "enabled": true } ] }
      """
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "api_key" }
      """
    Then the call should succeed
    And the response should contain the field "value" equal to "acme-key"
    When I call "VaultService.CopySecret" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "api_key" }
      """
    Then the call should succeed
    And the clipboard should contain "acme-key"

  @reveal @errors
  Scenario Outline: Reveal and CopySecret only work for stored secrets - <case>
    Given I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "globals",
        "rows": [
          { "key": "api_key", "secret": true, "secret_value": "acme-key", "enabled": true },
          { "key": "empty", "secret": true, "enabled": true },
          { "key": "plain", "initial_value": "visible", "enabled": true }
        ]
      }
      """
    When I call "VaultService.<method>" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "<key>" }
      """
    Then the call should fail with code "VLT-03404"
    And the clipboard should contain ""

    Examples:
      | case             | method     | key     |
      | no value         | Reveal     | empty   |
      | plain row        | Reveal     | plain   |
      | unknown row      | Reveal     | nope    |
      | no value         | CopySecret | empty   |
      | plain row        | CopySecret | plain   |

  @isolation
  Scenario: Each workspace has its own vault and identity
    Given I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "acme-key", "enabled": true } ] }
      """
    And I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Globex.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "globex-key", "enabled": true } ] }
      """
    Then the keychain should contain an entry for the account "{{Acme.local_id}}"
    And the keychain should contain an entry for the account "{{Globex.local_id}}"
    And the app data file "workspaces/{{Acme.local_id}}/vault.age" should exist
    And the app data file "workspaces/{{Globex.local_id}}/vault.age" should exist
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "api_key" }
      """
    Then the response should contain the field "value" equal to "acme-key"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Globex.local_id}}", "scope": "globals", "key": "api_key" }
      """
    Then the response should contain the field "value" equal to "globex-key"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Globex.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/who", "headers": [ { "key": "X-Api-Key", "value": "{{api_key}}" } ] } }
      """
    Then the 1 "GET" request to "/who" should have the header "X-Api-Key" equal to "globex-key"

  @probe
  Scenario: A keychain that fails the probe leads to passphrase mode
    Given the keychain fails with "access denied"
    When I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "acme-key", "enabled": true } ] }
      """
    Then the call should fail with code "VLT-05409"
    And the app data file "app.json" should contain the field "vault.mode" equal to "passphrase"

  @probe
  Scenario: XENON_FORCE_PASSPHRASE skips the keychain even when it works
    Given the environment variable "XENON_FORCE_PASSPHRASE" is set to "true"
    When I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "acme-key", "enabled": true } ] }
      """
    Then the call should fail with code "VLT-05409"
    And the app data file "app.json" should contain the field "vault.mode" equal to "passphrase"
    And the keychain should contain 0 entries

  @app_json
  Scenario: The vault settings survive other writes to app.json
    Given I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "acme-key", "enabled": true } ] }
      """
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Globex.local_id}}" }
      """
    Then the call should succeed
    And the app data file "app.json" should contain the field "vault.mode" equal to "keychain"
    And the app data file "app.json" should contain the field "active_local_id" equal to "{{Globex.local_id}}"
