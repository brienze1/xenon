#language: en
#utf-8

@all @xen_002 @secrets @passphrase_mode
Feature: Vaults unlock with a passphrase when the keychain is unavailable
  Without a usable OS keychain the app uses passphrase mode: the first vault asks the user to
  choose a passphrase (entered twice, at least 8 characters). Each workspace's vault identity is
  then wrapped with it in workspaces/<localId>/vault-key.age (scrypt), and app.json keeps a
  check token so a wrong passphrase is refused before any vault is touched. The passphrase is
  never stored: after a restart every vault is locked until it is typed once, which unlocks all
  of them.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    And the keychain is unavailable
    And the "api" variable is set to the mocked api url

  @setup
  Scenario: The first secret asks for a passphrase to be chosen
    When I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "acme-key", "enabled": true } ] }
      """
    Then the call should fail with code "VLT-05409"
    And the workspace file "acme/workspace.yaml" should contain the field "variables" with the items ""
    And the app data file "workspaces/{{Acme.local_id}}/vault.age" should not exist
    When I call "VaultService.Status"
    Then the response should contain the field "mode" equal to "passphrase"
    And the response should contain the field "needs_passphrase_setup" equal to "true"
    When I call "VaultService.SetupPassphrase" with
      """
      { "passphrase": "correct horse", "confirmation": "correct horse" }
      """
    Then the call should succeed
    And the response should contain the field "mode" equal to "passphrase"
    And the response should contain the field "needs_passphrase_setup" equal to "false"
    And the response should contain the field "locked_local_ids" equal to "[]"
    And the event "vault:status-changed" should have been emitted 2 times
    And the app data file "app.json" should contain the field "vault.mode" equal to "passphrase"
    And the app data file "app.json" should not contain "correct horse"
    When I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "acme-key", "enabled": true } ] }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/vault-key.age" should exist
    And the app data path "workspaces/{{Acme.local_id}}/vault-key.age" should have the permissions "600"
    And the app data file "workspaces/{{Acme.local_id}}/vault.age" should exist
    And the keychain should contain 0 entries
    And no file under the app data folder should contain "correct horse"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "api_key" }
      """
    Then the response should contain the field "value" equal to "acme-key"

  @setup @errors
  Scenario Outline: A passphrase must be confirmed and long enough - <case>
    When I call "VaultService.SetupPassphrase" with
      """
      { "passphrase": "<passphrase>", "confirmation": "<confirmation>" }
      """
    Then the call should fail with code "<code>"
    And the app data file "app.json" should not contain "passphraseCheck"

    Examples:
      | case      | passphrase    | confirmation  | code      |
      | mismatch  | correct horse | correct horsE | VLT-00400 |
      | too short | short12       | short12       | VLT-01400 |

  @setup @errors
  Scenario: A passphrase can only be chosen once
    Given the vault passphrase is "correct horse"
    When I call "VaultService.SetupPassphrase" with
      """
      { "passphrase": "another one", "confirmation": "another one" }
      """
    Then the call should fail with code "VLT-09409"

  @unlock
  Scenario: After a restart every vault is locked until the passphrase is typed once
    Given the vault passphrase is "correct horse"
    And I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "acme-key", "enabled": true } ] }
      """
    And I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Globex.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "globex-key", "enabled": true } ] }
      """
    When the app is restarted
    And I call "VaultService.Status"
    Then the response should contain the field "mode" equal to "passphrase"
    And the response should contain the field "needs_passphrase_setup" equal to "false"
    And the response should contain the field "locked_local_ids" with the items "{{Acme.local_id}}, {{Globex.local_id}}"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals" }
      """
    Then the response should contain the field "rows.0.secret_state" equal to "locked"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "api_key" }
      """
    Then the call should fail with code "VLT-04409"
    When I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "api_key", "value": "new" }
      """
    Then the call should fail with code "VLT-04409"
    When I call "VaultService.Unlock" with
      """
      { "passphrase": "wrong horse" }
      """
    Then the call should fail with code "VLT-02401"
    When I call "VaultService.Unlock" with
      """
      { "passphrase": "correct horse" }
      """
    Then the call should succeed
    And the response should contain the field "locked_local_ids" equal to "[]"
    And the event "vault:status-changed" should have been emitted 2 times
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Globex.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/who", "headers": [ { "key": "X-Api-Key", "value": "{{api_key}}" } ] } }
      """
    Then the 1 "GET" request to "/who" should have the header "X-Api-Key" equal to "globex-key"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "api_key" }
      """
    Then the response should contain the field "value" equal to "acme-key"

  @unlock @errors
  Scenario: Unlocking needs a passphrase to have been chosen
    When I call "VaultService.Unlock" with
      """
      { "passphrase": "correct horse" }
      """
    Then the call should fail with code "VLT-08409"

  @unlock
  Scenario: A workspace without a vault is never reported as locked
    Given the vault passphrase is "correct horse"
    And I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "acme-key", "enabled": true } ] }
      """
    When the app is restarted
    And I call "VaultService.Status"
    Then the response should contain the field "locked_local_ids" with the items "{{Acme.local_id}}"

  @broken
  Scenario: A lost key file makes the vault broken and Reset starts an empty one
    Given the vault passphrase is "correct horse"
    And I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "token", "value": "acme-token", "secret": true }
      """
    And the app data path "workspaces/{{Acme.local_id}}/vault-key.age" is removed
    When I call "VaultService.Status"
    Then the response should contain the field "broken_local_ids" with the items "{{Acme.local_id}}"
    When I call "VaultService.Reset" with
      """
      { "local_id": "{{Acme.local_id}}", "confirm": true }
      """
    Then the call should succeed
    And the response should contain the field "broken_local_ids" equal to "[]"
    When I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "token", "value": "fresh-token" }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/vault-key.age" should exist
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "token" }
      """
    Then the response should contain the field "value" equal to "fresh-token"

  @errors @validation
  Scenario Outline: Invalid input is rejected - <method>
    When I call "VaultService.<method>" with
      """
      <payload>
      """
    Then the call should fail with code "<code>"

    Examples:
      | method          | payload                                              | code      |
      | SetupPassphrase | { "passphrase": "", "confirmation": "" }             | VLT-10400 |
      | Unlock          | { "passphrase": "" }                                 | VLT-11400 |
      | Reveal          | { "local_id": "", "scope": "globals", "key": "k" }   | VLT-12400 |
      | CopySecret      | { "local_id": "x", "scope": "", "key": "k" }         | VLT-13400 |
      | Reset           | { "local_id": "" }                                   | VLT-14400 |
      | Reveal          | { "local_id": "x", "scope": "bad", "key": "k" }      | VLT-12400 |
