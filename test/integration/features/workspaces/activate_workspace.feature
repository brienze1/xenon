#language: en
#utf-8

@all @xen_001 @workspaces @activate_workspace
Feature: List and activate workspaces
  The switcher lists every registered workspace, flags the ones whose folder is gone, and
  activating one makes it the workspace the app reopens on the next start.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Globex" exists

  @success
  Scenario: Listing returns every workspace in registration order with the active one
    When I call "WorkspaceService.List"
    Then the call should succeed
    And the response should contain the field "workspaces" array length equal to 2
    And the response should contain the field "workspaces.0.local_id" equal to "{{Acme.local_id}}"
    And the response should contain the field "workspaces.0.name" equal to "Acme"
    And the response should contain the path field "workspaces.0.path" equal to "{{Acme.path}}"
    And the response should contain the field "workspaces.0.missing" equal to "false"
    And the response should contain the field "workspaces.1.name" equal to "Globex"
    And the response should contain the field "active_local_id" equal to "{{Globex.local_id}}"

  @success
  Scenario: Activating a workspace makes it the active one
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "summary.local_id" equal to "{{Acme.local_id}}"
    And the response should contain the field "summary.name" equal to "Acme"
    And the response should contain the field "summary.missing" equal to "false"
    And the app data file "app.json" should contain the field "active_local_id" equal to "{{Acme.local_id}}"

  @success
  Scenario: Activating refreshes the cached name from the workspace file
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JA2Z5V3B6C7D8E9F0G1H2J3K
      name: Acme Corp
      description: ""
      order: []
      """
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "summary.name" equal to "Acme Corp"
    And the app data file "app.json" should contain the field "workspaces.0.display_name_cache" equal to "Acme Corp"

  @missing
  Scenario: A workspace whose folder was deleted is listed as missing and cannot be activated
    Given the workspace path "acme" is removed
    When I call "WorkspaceService.List"
    Then the response should contain the field "workspaces.0.missing" equal to "true"
    And the response should contain the field "workspaces.0.name" equal to "Acme"
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should fail with code "WKS-06404"
    And the app data file "app.json" should contain the field "active_local_id" equal to "{{Globex.local_id}}"

  @not_found
  Scenario: Activating an unknown workspace fails
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "01JA2Z5V3B6C7D8E9F0G1H2J3K" }
      """
    Then the call should fail with code "WKS-07404"

  @validation
  Scenario: Activating without an id is rejected
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "" }
      """
    Then the call should fail with code "WKS-08400"

  @malformed
  Scenario: A workspace file that cannot be read is reported
    Given the workspace file "acme/workspace.yaml" contains
      """
      name: [unclosed
      """
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should fail with code "WKS-15422"

  @version
  Scenario: A workspace file written by a newer Xenon is refused
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 2
      kind: workspace
      id: 01JA2Z5V3B6C7D8E9F0G1H2J3K
      name: Acme
      """
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should fail with code "WKS-04422"
