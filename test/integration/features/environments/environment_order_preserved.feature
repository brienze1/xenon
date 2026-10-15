#language: en
#utf-8

@all @xen_011 @environments @environment_order_preserved
Feature: Other changes to workspace.yaml keep the environment order
  workspace.yaml is also rewritten when collections are reordered, when globals change and when the
  workspace is renamed. Each of those keeps environments.order, and any other key under
  environments, exactly as it was.

  Background:
    Given the workspace "Acme" exists
    And the "billing" variable is set to "01JXK000000000000000000011"
    And the "crm" variable is set to "01JXK000000000000000000012"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXK000000000000000000000
      name: Acme
      order:
        - billing
        - crm
      environments:
        order:
          - 01JXK000000000000000000003
          - 01JXK000000000000000000001
        pinned: 01JXK000000000000000000003
      variables: []
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      order: []
      """
    And the workspace file "acme/collections/crm/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{crm}}
      name: CRM
      order: []
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JXK000000000000000000001
      name: dev
      color: none
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JXK000000000000000000003
      name: prod
      color: red
      confirmBeforeSend: true
      variables: []
      """

  @collection_move
  Scenario: Reordering collections keeps environments.order
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{crm}}", "new_parent_id": "", "index": 0 }
      """
    Then the call should succeed
    And the workspace file "acme/workspace.yaml" should contain the field "order" with the items "crm, billing"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "01JXK000000000000000000003, 01JXK000000000000000000001"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.pinned" equal to "01JXK000000000000000000003"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments.0.name" equal to "prod"
    And the response should contain the field "environments.1.name" equal to "dev"

  @globals
  Scenario: Editing the globals keeps environments.order
    When I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "tenant", "value": "acme", "secret": false }
      """
    Then the call should succeed
    And the workspace file "acme/workspace.yaml" should contain the field "variables.0.key" equal to "tenant"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "01JXK000000000000000000003, 01JXK000000000000000000001"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.pinned" equal to "01JXK000000000000000000003"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments.0.name" equal to "prod"
    And the response should contain the field "environments.1.name" equal to "dev"

  @workspace_rename
  Scenario: Renaming the workspace keeps environments.order
    When I call "WorkspaceService.Rename" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "Acme Corp" }
      """
    Then the call should succeed
    And the workspace file "acme/workspace.yaml" should contain the field "name" equal to "Acme Corp"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "01JXK000000000000000000003, 01JXK000000000000000000001"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.pinned" equal to "01JXK000000000000000000003"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments.0.name" equal to "prod"
    And the response should contain the field "environments.1.name" equal to "dev"
