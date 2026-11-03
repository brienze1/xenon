#language: en
#utf-8

@all @xen_023 @environments @environment_revision
Feature: The window's environment saves are refused when the environment changed since it was opened
  An environment carries a revision (a fingerprint of its file) that the window gets with the environment and its
  variable table and sends back when it saves either. A save based on an older revision is refused with ENV-19409
  and writes nothing, unless it is forced ("Keep mine"). A save based on the current revision writes and returns the
  new revision.

  Background:
    Given the workspace "Acme" exists
    And the "dev" variable is set to "01JW8000000000000000000510"
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      color: green
      variables:
        - key: base_url
          value: https://dev.acme.test
          enabled: true
          description: ""
      """

  @revision_returned
  Scenario: The environment and its variable table carry the same revision
    When I call "EnvironmentService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    Then the call should succeed
    And the response should contain the field "revision" matching "^[0-9a-f]{64}$"
    And I save the response field "revision" as "opened"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}" }
      """
    Then the call should succeed
    And the response should contain the field "revision" equal to "{{opened}}"

  @fresh_save
  Scenario: A save based on the current revision writes and returns the new revision
    Given I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}" }
      """
    And I save the response field "revision" as "opened"
    When I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "base_revision": "{{opened}}", "rows": [ { "key": "base_url", "initial_value": "https://dev2.acme.test", "description": "", "enabled": true } ] }
      """
    Then the call should succeed
    And the response should contain the field "revision" matching "^[0-9a-f]{64}$"
    And the response should contain the field "revision" not equal to "{{opened}}"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.value" equal to "https://dev2.acme.test"

  @stale_table_save
  Scenario: A variable table saved from an outdated copy is refused and the newer change is kept
    Given I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}" }
      """
    And I save the response field "revision" as "opened"
    And I call the MCP tool "env_edit" with
      """
      { "action": "add_var", "scope": "env:{{dev}}", "key": "ai_key", "value": "from-ai" }
      """
    When I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "base_revision": "{{opened}}", "rows": [ { "key": "base_url", "initial_value": "https://dev2.acme.test", "description": "", "enabled": true } ] }
      """
    Then the call should fail with code "ENV-19409"
    And the error should contain the field "error.message" containing "the environment changed on disk since it was opened"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables" array length equal to 2
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.value" equal to "https://dev.acme.test"

  @stale_settings_save
  Scenario: Environment settings saved from an outdated copy are refused and the newer change is kept
    Given I call "EnvironmentService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    And I save the response field "revision" as "opened"
    And I call the MCP tool "env_edit" with
      """
      { "action": "update", "id": "{{dev}}", "description": "Edited by AI" }
      """
    When I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}", "name": "dev", "description": "", "color": "red", "confirm_before_send": false, "base_revision": "{{opened}}" }
      """
    Then the call should fail with code "ENV-19409"
    And the workspace file "acme/environments/dev.yaml" should contain the field "description" equal to "Edited by AI"
    And the workspace file "acme/environments/dev.yaml" should contain the field "color" equal to "green"

  @forced_save
  Scenario: Keeping my version forces the save over the newer change
    Given I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}" }
      """
    And I save the response field "revision" as "opened"
    And I call the MCP tool "env_edit" with
      """
      { "action": "add_var", "scope": "env:{{dev}}", "key": "ai_key", "value": "from-ai" }
      """
    When I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "base_revision": "{{opened}}", "force": true, "rows": [ { "key": "base_url", "initial_value": "https://dev2.acme.test", "description": "", "enabled": true } ] }
      """
    Then the call should succeed
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables" array length equal to 1
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.value" equal to "https://dev2.acme.test"
