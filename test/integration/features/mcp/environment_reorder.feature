#language: en
#utf-8

@all @xen_011 @mcp @environment_reorder
Feature: AI clients reorder environments
  "env_edit" has a "reorder" action that moves one environment to a position, exactly like a drag in
  the Environments view: it stores environments.order in workspace.yaml, returns the environments
  in their new order and emits variables:changed so the open window follows.

  Background:
    Given the workspace "Acme" exists
    And the "dev" variable is set to "01JXM000000000000000000010"
    And the "prod" variable is set to "01JXM000000000000000000011"
    And the "stage" variable is set to "01JXM000000000000000000012"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXM000000000000000000000
      name: Acme
      order: []
      variables: []
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      color: green
      confirmBeforeSend: false
      variables: []
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
    And the workspace file "acme/environments/stage.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{stage}}
      name: stage
      color: yellow
      confirmBeforeSend: false
      variables: []
      """

  @listing
  Scenario: env_edit offers the reorder action
    When I list the MCP tools
    Then the response field "" should contain an item with "name" equal to "env_edit" and "inputSchema.properties.action.enum.3" equal to "reorder"

  @reorder
  Scenario: Reordering moves the environment, stores the order and updates the window
    Given the emitted events are forgotten
    When I call the MCP tool "env_edit" with
      """
      { "action": "reorder", "id": "{{stage}}", "index": 0 }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 3
    And the response should contain the field "items.0.id" equal to "{{stage}}"
    And the response should contain the field "items.1.id" equal to "{{dev}}"
    And the response should contain the field "items.2.id" equal to "{{prod}}"
    And a "variables:changed" event with the field "local_id" equal to "{{Acme.local_id}}" should have been emitted within 1000 ms
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "{{stage}}, {{dev}}, {{prod}}"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments.0.name" equal to "stage"
    And the response should contain the field "environments.1.name" equal to "dev"
    And the response should contain the field "environments.2.name" equal to "Prod"
    When I call the MCP tool "env" with
      """
      { "action": "list" }
      """
    Then the response should contain the field "items.0.id" equal to "{{stage}}"

  @reorder @last
  Scenario: An index past the end puts the environment last
    When I call the MCP tool "env_edit" with
      """
      { "action": "reorder", "id": "{{dev}}", "index": 10 }
      """
    Then the call should succeed
    And the response should contain the field "items.2.id" equal to "{{dev}}"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "{{prod}}, {{stage}}, {{dev}}"

  @not_found
  Scenario: Reordering an environment that does not exist fails and changes nothing
    When I call the MCP tool "env_edit" with
      """
      { "action": "reorder", "id": "01JXM00000000000000000009Z", "index": 0 }
      """
    Then the call should fail with code "ENV-01404"
    And the workspace file "acme/workspace.yaml" should be unchanged

  @validation
  Scenario Outline: A reorder needs an environment id and a position from 0 - <case>
    When I call the MCP tool "env_edit" with
      """
      <arguments>
      """
    Then the call should fail with code "ENV-18400"
    And the workspace file "acme/workspace.yaml" should be unchanged

    Examples:
      | case           | arguments                                              |
      | missing index  | { "action": "reorder", "id": "{{dev}}" }               |
      | negative index | { "action": "reorder", "id": "{{dev}}", "index": -2 }  |
      | missing id     | { "action": "reorder", "index": 0 }                    |
