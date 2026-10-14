#language: en
#utf-8

@all @xen_009 @import @postman_v3_environment_globals
Feature: Import Postman v3 environments and globals
  A v3 environment (postman/environments/<name>.environment.yaml, or the JSON `postman environment get --json`
  prints) becomes a Xenon environment, and v3 globals (postman/globals/workspace.globals.yaml, or `postman globals get
  --json`) merge into the workspace globals, exactly like the v2.1 exports. A row is secret when it has `type: secret`
  or `secret: true`; its value goes to the vault and never to a workspace file, a response or a log line. The CLI's
  `********` mask and a link to a Postman vault secret are never stored as values: those rows import as secrets with
  no value, and a vault link is reported so the user can enter the value in Xenon.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available

  @detect
  Scenario Outline: v3 environments and globals are detected - <source>
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/env/<source>"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "<kind>"
    And the response should contain the field "version" equal to "3"
    And the response should contain the field "items" array length equal to 1
    And the response should contain the field "items.0.kind" equal to "<item_kind>"
    And the response should contain the field "items.0.name" equal to "<name>"
    And the response should contain the field "items.0.counts.variables" equal to "<variables>"
    And the response should not contain "s3cr3t-fixture"

    Examples:
      | source                  | kind                | item_kind   | name      | variables |
      | staging.environment.yaml | postman_environment | environment | staging   | 5         |
      | noname.environment.yaml | postman_environment | environment | noname    | 1         |
      | cli-environment.json    | postman_environment | environment | Cloud dev | 2         |
      | workspace.globals.yaml  | postman_globals     | globals     | Globals   | 2         |
      | cloud-globals.json      | postman_globals     | globals     | Globals   | 1         |
      | local-globals.json      | postman_globals     | globals     | Globals   | 1         |

  @rows @secrets
  Scenario: Plain, disabled, secret and vault-linked rows become the environment's rows
    Given the log output is captured
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/env/staging.environment.yaml"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the response should contain the field "workspaces.0.created.environments" equal to "1"
    And the response should contain the field "workspaces.0.created.variables" equal to "5"
    And the response field "warnings" should contain 1 items with "code" equal to "linked_secret"
    And the response field "warnings" should contain an item with "code" equal to "linked_secret" and "message" equal to "linked is linked to a Postman vault secret; enter its value in Xenon"
    And the response should not contain "s3cr3t-fixture-env"
    And the response should not contain "s3cr3t-fixture-token"
    And I save the response field "items.0.id" as "staging"
    And the workspace file "acme/environments/staging.yaml" should contain the field "name" equal to "staging"
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables" array length equal to 5
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables.0.key" equal to "host"
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables.0.value" equal to "https://staging.acme.test"
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables.0.enabled" equal to "true"
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables.1.key" equal to "debug"
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables.1.enabled" equal to "false"
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables.2.key" equal to "apiKey"
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables.2.secret" equal to "true"
    And the workspace file "acme/environments/staging.yaml" should not contain the field "variables.2.value"
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables.3.key" equal to "token"
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables.3.secret" equal to "true"
    And the workspace file "acme/environments/staging.yaml" should not contain the field "variables.3.value"
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables.4.key" equal to "linked"
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables.4.secret" equal to "true"
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables.4.postman.source.provider" equal to "postman"
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables.4.postman.source.postman.secretId" equal to "11111111-2222-3333-4444-555555555555"
    And no file under the workspace folder should contain "s3cr3t-fixture-env"
    And no file under the workspace folder should contain "s3cr3t-fixture-token"
    And the log output should not contain "s3cr3t-fixture-env"
    And the log output should not contain "s3cr3t-fixture-token"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{staging}}", "key": "apiKey" }
      """
    Then the call should succeed
    And the response should contain the field "value" equal to "s3cr3t-fixture-env"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{staging}}", "key": "token" }
      """
    Then the response should contain the field "value" equal to "s3cr3t-fixture-token"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{staging}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.2.secret_state" equal to "set"
    And the response should contain the field "rows.3.secret_state" equal to "set"
    And the response should contain the field "rows.4.secret_state" equal to "missing"

  @cli_json @secrets
  Scenario: The CLI's masked secret is imported as a secret with no value and cloud fields are kept
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/env/cli-environment.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And I save the response field "items.0.id" as "cloud"
    And the workspace file "acme/environments/cloud-dev.yaml" should contain the field "name" equal to "Cloud dev"
    And the workspace file "acme/environments/cloud-dev.yaml" should contain the field "variables.0.value" equal to "https://dev.cloud.test"
    And the workspace file "acme/environments/cloud-dev.yaml" should contain the field "variables.1.key" equal to "cloud_token"
    And the workspace file "acme/environments/cloud-dev.yaml" should contain the field "variables.1.secret" equal to "true"
    And the workspace file "acme/environments/cloud-dev.yaml" should not contain the field "variables.1.value"
    And the workspace file "acme/environments/cloud-dev.yaml" should contain the field "postman.id" equal to "8454478-2a54fdfa-03d8-40ea-b14c-614a46ff7246"
    And the workspace file "acme/environments/cloud-dev.yaml" should contain the field "postman.owner" equal to "8454478"
    And no file under the workspace folder should contain "********"
    And no file under the app data folder should contain "********"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{cloud}}" }
      """
    Then the response should contain the field "rows.1.secret_state" equal to "missing"

  @name
  Scenario: An environment file without a name is named after the file
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/env/noname.environment.yaml"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/environments/noname.yaml" should contain the field "name" equal to "noname"
    And the workspace file "acme/environments/noname.yaml" should contain the field "variables.0.value" equal to "eu-west-1"

  @globals @secrets
  Scenario: v3 globals merge into the workspace globals with their secret in the vault
    Given the log output is captured
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/env/workspace.globals.yaml"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the response should contain the field "workspaces.0.created.variables" equal to "2"
    And the workspace file "acme/workspace.yaml" should contain the field "variables" array length equal to 2
    And the workspace file "acme/workspace.yaml" should contain the field "variables.0.key" equal to "g1"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.0.value" equal to "v1"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.1.key" equal to "gsecret"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.1.secret" equal to "true"
    And the workspace file "acme/workspace.yaml" should not contain the field "variables.1.value"
    And no file under the workspace folder should contain "s3cr3t-fixture-global"
    And the log output should not contain "s3cr3t-fixture-global"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "gsecret" }
      """
    Then the response should contain the field "value" equal to "s3cr3t-fixture-global"

  @globals @clash
  Scenario Outline: A v3 global whose key already exists is resolved - <resolution>
    Given I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "globals",
        "rows": [ { "key": "cloud_region", "initial_value": "eu-central-1", "enabled": true, "description": "" } ]
      }
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/env/cloud-globals.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Clashes" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the response should contain the field "clashes" array length equal to 1
    And the response field "clashes" should contain an item with "kind" equal to "variable" and "existing_name" equal to "cloud_region"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "resolution": "<resolution>" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/workspace.yaml" should contain the field "variables" array length equal to 1
    And the workspace file "acme/workspace.yaml" should contain the field "variables.0.value" equal to "<value>"

    Examples:
      | resolution | value        |
      | replace    | us-east-1    |
      | keep_both  | eu-central-1 |
      | skip       | eu-central-1 |
