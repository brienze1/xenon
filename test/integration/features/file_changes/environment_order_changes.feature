#language: en
#utf-8

@all @xen_011 @file_changes @environment_order_changes
Feature: An environment order changed outside Xenon refreshes the list
  When workspace.yaml changes on disk (a git pull or a hand edit), variables:changed is emitted
  and the environments are listed in the new order without a restart.

  Background:
    Given the workspace "Acme" exists
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXF000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - 01JXF000000000000000000001
          - 01JXF000000000000000000002
      variables: []
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JXF000000000000000000001
      name: dev
      color: none
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JXF000000000000000000002
      name: prod
      color: red
      confirmBeforeSend: true
      variables: []
      """
    And I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """

  @external_edit
  Scenario: A new order written outside Xenon is listed after the change is noticed
    Given the emitted events are forgotten
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXF000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - 01JXF000000000000000000002
          - 01JXF000000000000000000001
      variables: []
      """
    When the file watcher reports a change to the workspace path "acme/workspace.yaml"
    Then a "variables:changed" event with the field "local_id" equal to "{{Acme.local_id}}" should have been emitted within 1000 ms
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "environments.0.name" equal to "prod"
    And the response should contain the field "environments.1.name" equal to "dev"
