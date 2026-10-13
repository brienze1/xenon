#language: en
#utf-8

@all @xen_008 @mcp @mcp_environments
Feature: AI clients manage environments, globals and variables
  "env" lists environments and reads a scope's variables; "env_edit" creates, updates and duplicates environments,
  switches the active one and edits variables (initial and current values, enabled, description, add and remove);
  "delete" removes an environment. Current values are local and never touch the workspace files; initial values do.
  Every change emits variables:changed so the open window follows.

  Background:
    Given the workspace "Acme" exists
    And the "dev" variable is set to "01JW8000000000000000000010"
    And the "prod" variable is set to "01JW8000000000000000000011"
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      color: green
      confirmBeforeSend: false
      variables:
        - key: base_url
          value: https://dev.acme.test
          enabled: true
          description: "Dev host"
      """
    And the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{prod}}
      name: Prod
      color: red
      confirmBeforeSend: true
      variables: []
      """

  @list
  Scenario: Listing environments is compact and marks the active and protected ones
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    When I call the MCP tool "env" with
      """
      { "action": "list" }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 2
    And the response field "items" should contain an item with "id" equal to "{{dev}}" and "active" equal to "true"
    And the response field "items" should contain an item with "id" equal to "{{prod}}" and "confirm_before_send" equal to "true"
    And the response field "items" should contain an item with "name" equal to "Prod" and "color" equal to "red"
    And the response field "items.0.variables" should not exist

  @set_active
  Scenario: Switching the active environment updates the switcher in the window
    Given the emitted events are forgotten
    When I call the MCP tool "env_edit" with
      """
      { "action": "set_active", "id": "{{prod}}" }
      """
    Then the call should succeed
    And a "variables:changed" event with the field "local_id" equal to "{{Acme.local_id}}" should have been emitted within 1000 ms
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "active_id" equal to "{{prod}}"

  @set_active
  Scenario: No environment can be made active
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    When I call the MCP tool "env_edit" with
      """
      { "action": "set_active", "id": null }
      """
    Then the call should succeed
    When I call the MCP tool "env" with
      """
      { "action": "list" }
      """
    Then the response field "items" should contain 0 items with "active" equal to "true"

  @manage
  Scenario: Creating, updating, duplicating and deleting environments
    When I call the MCP tool "env_edit" with
      """
      { "action": "create", "name": "staging", "color": "yellow", "confirm_before_send": false }
      """
    Then the call should succeed
    And I save the response field "id" as "staging"
    And the workspace file "acme/environments/staging.yaml" should contain the field "name" equal to "staging"
    When I call the MCP tool "env_edit" with
      """
      { "action": "update", "id": "{{staging}}", "name": "stage", "confirm_before_send": true }
      """
    Then the call should succeed
    And the workspace file "acme/environments/stage.yaml" should contain the field "confirmBeforeSend" equal to "true"
    When I call the MCP tool "env_edit" with
      """
      { "action": "duplicate", "id": "{{dev}}" }
      """
    Then the call should succeed
    And I save the response field "id" as "dev_copy"
    When I call the MCP tool "env" with
      """
      { "action": "list" }
      """
    Then the response should contain the field "items" array length equal to 4
    When I call the MCP tool "delete" with
      """
      { "target": "environment", "id": "{{dev_copy}}" }
      """
    Then the call should succeed
    When I call the MCP tool "env" with
      """
      { "action": "list" }
      """
    Then the response should contain the field "items" array length equal to 3

  @current_value
  Scenario: Setting a current value keeps the workspace file unchanged
    Given the emitted events are forgotten
    When I call the MCP tool "env_edit" with
      """
      { "action": "set_current", "scope": "env:{{dev}}", "key": "base_url", "value": "http://localhost:8080" }
      """
    Then the call should succeed
    And the workspace file "acme/environments/dev.yaml" should be unchanged
    And the 1 "variables:changed" event should contain the field "scopes.0" equal to "env:{{dev}}"
    When I call the MCP tool "env" with
      """
      { "action": "get", "scope": "env:{{dev}}" }
      """
    Then the response field "vars" should contain an item with "key" equal to "base_url" and "current" equal to "http://localhost:8080"
    And the response field "vars" should contain an item with "key" equal to "base_url" and "initial" equal to "https://dev.acme.test"
    When I call the MCP tool "env_edit" with
      """
      { "action": "reset", "scope": "env:{{dev}}", "key": "base_url" }
      """
    Then the call should succeed
    When I call the MCP tool "env" with
      """
      { "action": "get", "scope": "env:{{dev}}" }
      """
    Then the response field "vars.0.current" should not exist

  @initial_value
  Scenario: Saving the variable table changes the initial values in the workspace file
    When I call the MCP tool "env_edit" with
      """
      { "action": "set_vars", "scope": "env:{{dev}}", "vars": [ { "key": "base_url", "initial": "https://dev2.acme.test", "description": "Dev host" }, { "key": "timeout", "initial": "30", "enabled": false } ] }
      """
    Then the call should succeed
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.value" equal to "https://dev2.acme.test"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.key" equal to "timeout"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.enabled" equal to "false"
    When I call the MCP tool "env" with
      """
      { "action": "get", "scope": "env:{{dev}}" }
      """
    Then the response field "vars" should contain an item with "key" equal to "timeout" and "enabled" equal to "false"
    And the response field "vars" should contain an item with "key" equal to "base_url" and "description" equal to "Dev host"

  @persist
  Scenario: Persisting copies the current value into the initial value
    Given I call the MCP tool "env_edit" with
      """
      { "action": "set_current", "scope": "env:{{dev}}", "key": "base_url", "value": "https://kept.acme.test" }
      """
    When I call the MCP tool "env_edit" with
      """
      { "action": "persist", "scope": "env:{{dev}}", "key": "base_url" }
      """
    Then the call should succeed
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.value" equal to "https://kept.acme.test"

  @globals
  Scenario: Globals are read and edited like any other scope
    When I call the MCP tool "env_edit" with
      """
      { "action": "add_var", "scope": "globals", "key": "tenant", "value": "acme" }
      """
    Then the call should succeed
    And the workspace file "acme/workspace.yaml" should contain the field "variables.0.key" equal to "tenant"
    When I call the MCP tool "env" with
      """
      { "action": "get", "scope": "globals" }
      """
    Then the response field "vars" should contain an item with "key" equal to "tenant" and "initial" equal to "acme"
