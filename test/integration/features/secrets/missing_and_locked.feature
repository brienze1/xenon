#language: en
#utf-8

@all @xen_002 @secrets @missing_and_locked
Feature: Secrets without a value on this machine are missing, secrets of a locked vault are locked
  A secret row that exists in the workspace files but has no value in this machine's vault (a
  cloned repository, a new machine) is shown with secret_state "missing". When the vault is locked
  (passphrase mode, not unlocked yet) its secrets are shown as "locked". Neither ever returns a
  value; setting a value on a missing secret stores it and makes it "set".

  Background:
    Given the workspace "Acme" exists
    And the "dev" variable is set to "01JF0000000000000000000010"
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
          description: cloned from git
        - key: stray
          secret: true
          value: leaked-in-git
          enabled: true
          description: ""
      """

  @missing
  Scenario: A secret that has no value here is missing until one is set
    Given the keychain is available
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.0.secret_state" equal to "missing"
    And the response should contain the field "rows.0.has_current_override" equal to "false"
    And the response should contain the field "rows.1.secret_state" equal to "missing"
    And the response should not contain "leaked-in-git"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key" }
      """
    Then the call should fail with code "VLT-03404"
    When I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key", "value": "now-set" }
      """
    Then the call should succeed
    And the response should contain the field "secret_state" equal to "set"
    And the response should contain the field "current_value" equal to ""
    When I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key", "value": "" }
      """
    Then the call should succeed
    And the response should contain the field "secret_state" equal to "missing"

  @missing
  Scenario: A value found next to a secret row in a file is ignored and removed on the next save
    Given the keychain is available
    When I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "env:{{dev}}",
        "rows": [
          { "key": "api_key", "secret": true, "enabled": true, "description": "cloned from git" },
          { "key": "stray", "secret": true, "enabled": true, "description": "" }
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "warnings" with the items "git_history"
    And no file under the workspace folder should contain "leaked-in-git"
    And the workspace file "acme/environments/dev.yaml" should not contain the field "variables.1.value"

  @locked
  Scenario: Secrets of a locked vault are locked and cannot be revealed
    Given the keychain is unavailable
    And the vault passphrase is "correct horse"
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key", "value": "locked-away" }
      """
    When the app is restarted
    And I call "EnvironmentService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    Then the response should contain the field "variables.0.secret_state" equal to "locked"
    And the response should contain the field "variables.1.secret_state" equal to "locked"
    And the response should not contain "locked-away"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key" }
      """
    Then the call should fail with code "VLT-04409"
    When I call "VaultService.CopySecret" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key" }
      """
    Then the call should fail with code "VLT-04409"

  @locked
  Scenario: A locked vault without a passphrase chosen yet asks for one
    Given the keychain is unavailable
    When I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key", "value": "x" }
      """
    Then the call should fail with code "VLT-05409"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}" }
      """
    Then the response should contain the field "rows.0.secret_state" equal to "missing"
