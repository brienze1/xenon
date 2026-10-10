#language: en
#utf-8

@all @xen_002 @secrets @plain_secret_switch
Feature: Switching a row between plain and secret
  Turning a plain row into a secret moves its current value into the vault, removes the initial
  value from the file and warns that git history may still hold it. Turning a secret into a plain
  row writes the value to the workspace files, so it needs confirm_reveal; the current value
  becomes the initial value. A row turned secret by an outside edit (git) gets its plain local
  current value moved into the vault, so no plain copy stays on disk. Vault entries that no longer
  match a secret row are pruned on the next write of their table.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the "dev" variable is set to "01JH0000000000000000000010"
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: api_key
          value: shared-initial-key
          enabled: true
          description: ""
        - key: tenant
          value: acme
          enabled: true
          description: ""
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key", "value": "my-local-key" }
      """

  @to_secret
  Scenario: A plain row turned secret moves its current value into the vault and warns about git history
    When I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "env:{{dev}}",
        "rows": [
          { "key": "api_key", "initial_value": "shared-initial-key", "secret": true, "enabled": true, "description": "" },
          { "key": "tenant", "initial_value": "acme", "enabled": true, "description": "" }
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "warnings" with the items "git_history"
    And the response should contain the field "rows.0.secret" equal to "true"
    And the response should contain the field "rows.0.secret_state" equal to "set"
    And the response should contain the field "rows.0.initial_value" equal to ""
    And no file under the workspace folder should contain "shared-initial-key"
    And no file under the app data folder should contain "my-local-key"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key" }
      """
    Then the response should contain the field "value" equal to "my-local-key"

  @to_secret
  Scenario: A plain row without a current value takes its initial value into the vault
    Given I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key", "value": "" }
      """
    When I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "env:{{dev}}",
        "rows": [
          { "key": "api_key", "secret": true, "enabled": true, "description": "" },
          { "key": "tenant", "initial_value": "acme", "enabled": true, "description": "" }
        ]
      }
      """
    Then the call should succeed
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key" }
      """
    Then the response should contain the field "value" equal to "shared-initial-key"

  @to_plain
  Scenario: A secret turned plain needs confirmation and writes its value as the initial value
    Given I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "rows": [ { "key": "api_key", "secret": true, "enabled": true }, { "key": "tenant", "initial_value": "acme", "enabled": true } ] }
      """
    When I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "rows": [ { "key": "api_key", "initial_value": "", "enabled": true }, { "key": "tenant", "initial_value": "acme", "enabled": true } ] }
      """
    Then the call should fail with code "VAR-01409"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.secret" equal to "true"
    When I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "confirm_reveal": true, "rows": [ { "key": "api_key", "initial_value": "", "enabled": true }, { "key": "tenant", "initial_value": "acme", "enabled": true } ] }
      """
    Then the call should succeed
    And the response should contain the field "rows.0.secret" equal to "false"
    And the response should contain the field "rows.0.initial_value" equal to "my-local-key"
    And the response should contain the field "rows.0.has_current_override" equal to "false"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.value" equal to "my-local-key"
    And the workspace file "acme/environments/dev.yaml" should not contain the field "variables.0.secret"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key" }
      """
    Then the call should fail with code "VLT-03404"

  @persist @errors
  Scenario: Persist is not available for secret rows
    Given I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "rows": [ { "key": "api_key", "secret": true, "enabled": true }, { "key": "tenant", "initial_value": "acme", "enabled": true } ] }
      """
    When I call "VariableService.Persist" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "keys": ["api_key"] }
      """
    Then the call should fail with code "VAR-02422"
    When I call "VariableService.Persist" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "keys": [] }
      """
    Then the call should succeed
    And no file under the workspace folder should contain "my-local-key"

  @rename
  Scenario: Renaming a secret row moves its vault value
    Given I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "rows": [ { "key": "api_key", "secret": true, "enabled": true }, { "key": "tenant", "initial_value": "acme", "enabled": true } ] }
      """
    When I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "rows": [ { "key": "vendor_key", "previous_key": "api_key", "secret": true, "enabled": true }, { "key": "tenant", "initial_value": "acme", "enabled": true } ] }
      """
    Then the call should succeed
    And the response should contain the field "rows.0.secret_state" equal to "set"
    And the response should contain the field "warnings" array length equal to 0
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "vendor_key" }
      """
    Then the response should contain the field "value" equal to "my-local-key"

  @outside_edit
  Scenario: A row turned secret by an outside edit gets its plain current value moved into the vault
    Given I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "rows": [ { "key": "api_key", "initial_value": "shared-initial-key", "enabled": true }, { "key": "tenant", "initial_value": "acme", "enabled": true }, { "key": "other", "secret": true, "secret_value": "other-value", "enabled": true } ] }
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
        - key: tenant
          value: acme
          enabled: true
          description: ""
        - key: other
          secret: true
          enabled: true
          description: ""
      """
    When the file watcher reports a change to the workspace path "acme/environments/dev.yaml"
    Then no file under the app data folder should contain "my-local-key"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key" }
      """
    Then the response should contain the field "value" equal to "my-local-key"

  @outside_edit @locked
  Scenario: A row turned secret by an outside edit while the vault is locked loses its plain current value
    Given the keychain is unavailable
    And the vault passphrase is "correct horse"
    And I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "other", "value": "other-value", "secret": true }
      """
    And I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And the app is restarted
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
        - key: other
          secret: true
          enabled: true
          description: ""
      """
    When the file watcher reports a change to the workspace path "acme/environments/dev.yaml"
    Then no file under the app data folder should contain "my-local-key"

  @orphans
  Scenario: A vault value whose row turned plain outside Xenon is pruned on the next write of its table
    Given I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "rows": [ { "key": "api_key", "secret": true, "enabled": true }, { "key": "tenant", "initial_value": "acme", "enabled": true }, { "key": "other", "secret": true, "secret_value": "v1", "enabled": true } ] }
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: api_key
          value: ""
          enabled: true
          description: ""
        - key: other
          secret: true
          enabled: true
          description: ""
      """
    When I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "other", "value": "v2" }
      """
    Then the call should succeed
    Given the workspace file "acme/environments/dev.yaml" contains
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
        - key: other
          secret: true
          enabled: true
          description: ""
      """
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}" }
      """
    Then the response should contain the field "rows.0.secret_state" equal to "missing"
    And the response should contain the field "rows.1.secret_state" equal to "set"
