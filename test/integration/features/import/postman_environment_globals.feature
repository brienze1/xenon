#language: en
#utf-8

@all @xen_003 @import @postman_environment_globals
Feature: Import Postman environments and globals
  A Postman environment export becomes a Xenon environment with the same name and rows: plain
  values become initial values and the enabled state is kept. Rows Postman marks as secret become
  secret variables whose values go to the workspace's vault, never to a workspace file, a summary,
  a warning or a log line; secret rows exported without a value are marked missing. Postman fields
  Xenon has no place for are kept in the postman: blocks (never a secret value). A globals export
  is merged into the workspace globals. An environment whose name already exists is replaced, kept
  as "<name> (imported)" or skipped; a global whose key already exists is replaced, or keeps its
  value with the imported value recorded in the summary (masked for secrets), or is skipped.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available

  @environment @preview
  Scenario: Detecting an environment lists it with its variable count and no secret value
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dev.postman_environment.json"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "postman_environment"
    And the response should contain the field "items" array length equal to 1
    And the response should contain the field "items.0.kind" equal to "environment"
    And the response should contain the field "items.0.name" equal to "dev"
    And the response should contain the field "items.0.counts.environments" equal to "1"
    And the response should contain the field "items.0.counts.variables" equal to "5"
    And the response should not contain "dev-api-key-SECRET-123"

  @environment @rows
  Scenario: Plain, disabled and secret rows become the environment's rows
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dev.postman_environment.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the response should contain the field "items.0.name" equal to "dev"
    And the response should contain the field "items.0.id" matching "^[0-9A-HJKMNP-TV-Z]{26}$"
    And the response should contain the field "workspaces.0.created.environments" equal to "1"
    And the response should contain the field "workspaces.0.created.variables" equal to "5"
    And the response should contain the field "failed" array length equal to 0
    And I save the response field "items.0.id" as "dev"
    And the workspace file "acme/environments/dev.yaml" should contain the field "kind" equal to "environment"
    And the workspace file "acme/environments/dev.yaml" should contain the field "id" equal to "{{dev}}"
    And the workspace file "acme/environments/dev.yaml" should contain the field "name" equal to "dev"
    And the workspace file "acme/environments/dev.yaml" should contain the field "color" equal to "none"
    And the workspace file "acme/environments/dev.yaml" should contain the field "confirmBeforeSend" equal to "false"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables" array length equal to 5
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.key" equal to "base_url"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.value" equal to "https://dev.acme.test"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.enabled" equal to "true"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.postman.type" equal to "default"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.key" equal to "tenant"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.value" equal to "acme"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.enabled" equal to "false"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.2.key" equal to "api_key"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.2.secret" equal to "true"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.2.enabled" equal to "true"
    And the workspace file "acme/environments/dev.yaml" should not contain the field "variables.2.value"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.2.postman.type" equal to "secret"
    And the workspace file "acme/environments/dev.yaml" should not contain the field "variables.2.postman.value"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.3.key" equal to "empty_secret"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.3.secret" equal to "true"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.4.key" equal to "any_value"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.4.value" equal to "x"
    And the workspace file "acme/environments/dev.yaml" should contain the field "postman.id" equal to "env-dev-0001"
    And the workspace file "acme/environments/dev.yaml" should contain the field "postman._postman_variable_scope" equal to "environment"
    And the workspace file "acme/environments/dev.yaml" should contain the field "postman._postman_exported_using" equal to "Postman/11.20.0"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.0.current_value" equal to "https://dev.acme.test"
    And the response should contain the field "rows.2.secret_state" equal to "set"
    And the response should contain the field "rows.3.secret_state" equal to "missing"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key" }
      """
    Then the call should succeed
    And the response should contain the field "value" equal to "dev-api-key-SECRET-123"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments" array length equal to 1
    And the response should contain the field "environments.0.name" equal to "dev"

  @environment @secrets
  Scenario: Secret values never reach a workspace file, the summary or the logs
    Given the log output is captured
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dev.postman_environment.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should not contain "dev-api-key-SECRET-123"
    And no file under the workspace folder should contain "dev-api-key-SECRET-123"
    And no file under the app data folder should contain "dev-api-key-SECRET-123"
    And the log output should contain "Finished"
    And the log output should not contain "dev-api-key-SECRET-123"

  @environment @defaults
  Scenario: A production-like environment gets the production defaults
    Given the file "prod.postman_environment.json" exists with content
      """
      {
        "id": "env-prod-1",
        "name": "Billing prod",
        "values": [ { "key": "base_url", "value": "https://billing.acme.test", "type": "default", "enabled": true } ],
        "_postman_variable_scope": "environment"
      }
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/prod.postman_environment.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/environments/billing-prod.yaml" should contain the field "color" equal to "red"
    And the workspace file "acme/environments/billing-prod.yaml" should contain the field "confirmBeforeSend" equal to "true"

  @environment @clash
  Scenario: An environment name that already exists is reported as a clash and must be resolved
    Given the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JF0000000000000000000001
      name: DEV
      color: green
      variables:
        - key: base_url
          value: https://old.acme.test
          enabled: true
          description: ""
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dev.postman_environment.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Clashes" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "clashes" array length equal to 1
    And the response should contain the field "clashes.0.kind" equal to "environment"
    And the response should contain the field "clashes.0.item_key" equal to "0"
    And the response should contain the field "clashes.0.existing_id" equal to "01JF0000000000000000000001"
    And the response should contain the field "clashes.0.existing_name" equal to "DEV"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should fail with code "IMP-04409"
    And the workspace file "acme/environments/dev.yaml" should contain the field "name" equal to "DEV"
    And the workspace folder "acme/environments" should contain the entries "dev.yaml"

  @environment @clash @replace
  Scenario: Replacing an environment swaps the existing one for the imported one
    Given the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JF0000000000000000000001
      name: dev
      color: green
      variables:
        - key: base_url
          value: https://old.acme.test
          enabled: true
          description: ""
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dev.postman_environment.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "resolution": "replace" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "replaced"
    And the response should contain the field "workspaces.0.replaced" equal to "1"
    And I save the response field "items.0.id" as "dev"
    And the workspace folder "acme/environments" should contain the entries "dev.yaml"
    And the workspace file "acme/environments/dev.yaml" should contain the field "id" equal to "{{dev}}"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables" array length equal to 5
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.value" equal to "https://dev.acme.test"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments" array length equal to 1
    And the response should contain the field "environments.0.id" equal to "{{dev}}"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key" }
      """
    Then the response should contain the field "value" equal to "dev-api-key-SECRET-123"

  @environment @clash @keep_both
  Scenario: Keeping both imports the environment as "<name> (imported)"
    Given the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JF0000000000000000000001
      name: dev
      color: green
      variables: []
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dev.postman_environment.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ], "apply_to_all": "keep_both" }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the response should contain the field "items.0.name" equal to "dev (imported)"
    And the workspace folder "acme/environments" should contain the entries "dev-(imported).yaml, dev.yaml"
    And the workspace file "acme/environments/dev.yaml" should contain the field "color" equal to "green"
    And the workspace file "acme/environments/dev-(imported).yaml" should contain the field "name" equal to "dev (imported)"
    And the workspace file "acme/environments/dev-(imported).yaml" should contain the field "variables" array length equal to 5

  @environment @clash @skip
  Scenario: Skipping leaves the existing environment untouched
    Given the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JF0000000000000000000001
      name: dev
      color: green
      variables: []
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dev.postman_environment.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "resolution": "skip" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "skipped"
    And the response should contain the field "workspaces.0.skipped" equal to "1"
    And the workspace file "acme/environments/dev.yaml" should be unchanged
    And no file under the app data folder should contain "dev-api-key-SECRET-123"

  @globals
  Scenario: A globals export is merged into the workspace globals
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/globals.json"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "postman_globals"
    And the response should contain the field "items.0.kind" equal to "globals"
    And the response should contain the field "items.0.counts.variables" equal to "2"
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the response should contain the field "workspaces.0.created.variables" equal to "2"
    And the response should contain the field "workspaces.0.created.environments" equal to "0"
    And the workspace file "acme/workspace.yaml" should contain the field "variables" array length equal to 2
    And the workspace file "acme/workspace.yaml" should contain the field "variables.0.key" equal to "region"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.0.value" equal to "eu-west-1"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.0.enabled" equal to "true"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.1.key" equal to "retries"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.1.enabled" equal to "false"
    And the workspace file "acme/workspace.yaml" should contain the field "postman.globals.id" equal to "globals-0001"
    And the workspace file "acme/workspace.yaml" should contain the field "postman.globals.name" equal to "My Workspace Globals"

  @globals @clash
  Scenario Outline: A global key that already exists is resolved - <resolution>
    Given I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "globals",
        "rows": [
          { "key": "region", "initial_value": "us-east-1", "enabled": true, "description": "" },
          { "key": "token", "secret": true, "secret_value": "old-token-value", "enabled": true, "description": "" }
        ]
      }
      """
    And the file "globals.json" exists with content
      """
      {
        "id": "globals-2",
        "name": "Globals",
        "values": [
          { "key": "region", "value": "eu-west-1", "type": "default", "enabled": true },
          { "key": "token", "value": "glob-SECRET-9", "type": "secret", "enabled": true },
          { "key": "fresh", "value": "new", "type": "default", "enabled": true }
        ],
        "_postman_variable_scope": "globals"
      }
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/globals.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Clashes" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the response should contain the field "clashes" array length equal to 2
    And the response field "clashes" should contain an item with "kind" equal to "variable" and "existing_name" equal to "region"
    And the response field "clashes" should contain an item with "kind" equal to "variable" and "existing_name" equal to "token"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "resolution": "<resolution>" } ] }
      """
    Then the call should succeed
    And the response should not contain "glob-SECRET-9"
    And the response should not contain "old-token-value"
    And the response field "warnings" should contain <kept> items with "code" equal to "variable_kept"
    And the workspace file "acme/workspace.yaml" should contain the field "variables" array length equal to 3
    And the workspace file "acme/workspace.yaml" should contain the field "variables.0.value" equal to "<region>"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.2.key" equal to "fresh"
    And no file under the workspace folder should contain "glob-SECRET-9"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "token" }
      """
    Then the response should contain the field "value" equal to "<token>"

    Examples:
      | resolution | kept | region    | token           |
      | replace    | 0    | eu-west-1 | glob-SECRET-9   |
      | keep_both  | 2    | us-east-1 | old-token-value |
      | skip       | 0    | us-east-1 | old-token-value |

  @globals @clash @keep_both
  Scenario: Keeping both records the imported values in the summary, masked for secrets
    Given I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "globals",
        "rows": [
          { "key": "region", "initial_value": "us-east-1", "enabled": true, "description": "" },
          { "key": "token", "secret": true, "secret_value": "old-token-value", "enabled": true, "description": "" }
        ]
      }
      """
    And the file "globals.json" exists with content
      """
      {
        "id": "globals-2",
        "name": "Globals",
        "values": [
          { "key": "region", "value": "eu-west-1", "type": "default", "enabled": true },
          { "key": "token", "value": "glob-SECRET-9", "type": "secret", "enabled": true }
        ],
        "_postman_variable_scope": "globals"
      }
      """
    And the log output is captured
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/globals.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ], "apply_to_all": "keep_both" }
      """
    Then the call should succeed
    And the response field "warnings" should contain an item with "code" equal to "variable_kept" and "message" equal to "region keeps its value; the imported value was eu-west-1"
    And the response field "warnings" should contain an item with "code" equal to "variable_kept" and "message" equal to "token keeps its value; the imported value was •••••"
    And the response field "warnings" should contain an item with "code" equal to "variable_kept" and "local_id" equal to "{{Acme.local_id}}"
    And the log output should not contain "glob-SECRET-9"
