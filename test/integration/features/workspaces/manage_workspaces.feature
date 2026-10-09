#language: en
#utf-8

@all @xen_001 @workspaces @manage_workspaces
Feature: Open, rename, remove and locate workspaces
  An existing workspace folder can be added to Xenon; two registered workspaces may never be
  the same folder or nested in each other. Renaming changes only the name in workspace.yaml.
  Removing either keeps the files or moves the folder to the OS trash; when the trash fails the
  workspace stays registered. A workspace whose folder was moved can be located again.

  Background:
    Given the workspace "Acme" exists
    And the workspace file "shared/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JC0000000000000000000000
      name: Shared APIs
      description: kept in git
      order: []
      """

  @open @success
  Scenario: Opening an existing workspace folder registers it and makes it active
    When I call "WorkspaceService.Open" with
      """
      { "path": "{{workspace_root}}/shared" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Shared APIs"
    And the response should contain the path field "path" equal to "{{workspace_root}}/shared"
    And the response should contain the field "missing" equal to "false"
    And I save the response field "local_id" as "shared_id"
    And the app data file "app.json" should contain the field "workspaces" array length equal to 2
    And the app data file "app.json" should contain the field "workspaces.1.local_id" equal to "{{shared_id}}"
    And the app data file "app.json" should contain the field "active_local_id" equal to "{{shared_id}}"
    And the workspace file "shared/workspace.yaml" should be unchanged

  @open @overlap
  Scenario Outline: Opening a folder that overlaps a registered workspace is refused - <case>
    When I call "WorkspaceService.Open" with
      """
      { "path": "<path>" }
      """
    Then the call should fail with code "WKS-02409"
    And the app data file "app.json" should contain the field "workspaces" array length equal to 1

    Examples:
      | case             | path                                |
      | same folder      | {{workspace_root}}/acme             |
      | parent folder    | {{workspace_root}}                  |
      | child folder     | {{workspace_root}}/acme/collections |

  @open @not_found
  Scenario: Opening a folder without a workspace file is refused
    Given the workspace folder "plain" exists
    When I call "WorkspaceService.Open" with
      """
      { "path": "{{workspace_root}}/plain" }
      """
    Then the call should fail with code "WKS-03404"

  @open @not_found
  Scenario: Opening a folder that does not exist is refused
    When I call "WorkspaceService.Open" with
      """
      { "path": "{{workspace_root}}/nowhere" }
      """
    Then the call should fail with code "WKS-03404"

  @open @version
  Scenario: Opening a workspace written by a newer Xenon is refused
    Given the workspace file "future/workspace.yaml" contains
      """
      xenon: 2
      kind: workspace
      id: 01JC0000000000000000000009
      name: Future
      """
    When I call "WorkspaceService.Open" with
      """
      { "path": "{{workspace_root}}/future" }
      """
    Then the call should fail with code "WKS-04422"

  @open @validation
  Scenario: Opening needs an absolute path
    When I call "WorkspaceService.Open" with
      """
      { "path": "relative/folder" }
      """
    Then the call should fail with code "WKS-21400"

  @rename @success
  Scenario: Renaming changes only the name in workspace.yaml and the switcher
    Given the workspace file "acme/workspace.yaml" contains
      """
      # company workspace
      xenon: 1
      kind: workspace
      id: 01JC0000000000000000000001
      name: Acme
      description: ""
      order: []
      custom: keep me
      """
    When I call "WorkspaceService.Rename" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "  Acme Corp  " }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Acme Corp"
    And the response should contain the path field "path" equal to "{{workspace_root}}/acme"
    And the workspace file "acme/workspace.yaml" should equal
      """
      # company workspace
      xenon: 1
      kind: workspace
      id: 01JC0000000000000000000001
      name: Acme Corp
      description: ""
      order: []
      custom: keep me
      """
    And the app data file "app.json" should contain the field "workspaces.0.display_name_cache" equal to "Acme Corp"
    When I call "WorkspaceService.List"
    Then the response should contain the field "workspaces.0.name" equal to "Acme Corp"

  @rename @validation
  Scenario Outline: Invalid renames are rejected - <case>
    When I call "WorkspaceService.Rename" with
      """
      { "local_id": "<local_id>", "name": "<name>" }
      """
    Then the call should fail with code "<code>"

    Examples:
      | case           | local_id                   | name | code      |
      | blank name     | {{Acme.local_id}}          |      | WKS-22400 |
      | no id          |                            | New  | WKS-22400 |
      | unknown id     | 01JA2Z5V3B6C7D8E9F0G1H2J3K | New  | WKS-07404 |

  @rename @missing
  Scenario: Renaming a workspace whose folder is gone fails
    Given the workspace path "acme" is removed
    When I call "WorkspaceService.Rename" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "Acme Corp" }
      """
    Then the call should fail with code "WKS-06404"

  @remove @success
  Scenario: Removing from Xenon keeps the files and makes the next workspace active
    Given the workspace "Globex" exists
    And the workspace "Initech" exists
    And I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Globex.local_id}}" }
      """
    And I call "UIStateService.Save" with
      """
      { "local_id": "{{Globex.local_id}}", "tabs": [] }
      """
    When I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Globex.local_id}}", "move_to_trash": false }
      """
    Then the call should succeed
    And the response should contain the field "workspaces" array length equal to 2
    And the response should contain the field "workspaces.0.name" equal to "Acme"
    And the response should contain the field "workspaces.1.name" equal to "Initech"
    And the response should contain the field "active_local_id" equal to "{{Initech.local_id}}"
    And the workspace file "globex/workspace.yaml" should exist
    And the trash should contain 0 items
    And the app data file "workspaces/{{Globex.local_id}}/state.json" should not exist
    And the app data file "app.json" should contain the field "active_local_id" equal to "{{Initech.local_id}}"

  @remove @success
  Scenario: Removing the last workspace in the list falls back to the previous one
    Given the workspace "Globex" exists
    When I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Globex.local_id}}", "move_to_trash": false }
      """
    Then the call should succeed
    And the response should contain the field "active_local_id" equal to "{{Acme.local_id}}"

  @remove @success
  Scenario: Removing an inactive workspace keeps the active one
    Given the workspace "Globex" exists
    When I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Acme.local_id}}", "move_to_trash": false }
      """
    Then the call should succeed
    And the response should contain the field "workspaces" array length equal to 1
    And the response should contain the field "active_local_id" equal to "{{Globex.local_id}}"

  @remove @success
  Scenario: Removing the only workspace leaves no active workspace
    When I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Acme.local_id}}", "move_to_trash": false }
      """
    Then the call should succeed
    And the response should contain the field "workspaces" array length equal to 0
    And the response should contain the field "active_local_id" equal to ""

  @remove @trash
  Scenario: Removing and moving the folder to the trash
    When I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Acme.local_id}}", "move_to_trash": true }
      """
    Then the call should succeed
    And the trash should contain 1 items
    And the trash item 1 should be the workspace path "acme"
    And the app data file "app.json" should contain the field "workspaces" array length equal to 0

  @remove @trash @missing
  Scenario: Removing a workspace whose folder is gone does not touch the trash
    Given the workspace path "acme" is removed
    When I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Acme.local_id}}", "move_to_trash": true }
      """
    Then the call should succeed
    And the trash should contain 0 items
    And the app data file "app.json" should contain the field "workspaces" array length equal to 0

  @remove @trash @failure
  Scenario: A failing trash keeps the workspace registered
    Given the trash fails with "volume has no trash"
    When I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Acme.local_id}}", "move_to_trash": true }
      """
    Then the call should fail with code "WKS-05500"
    And the app data file "app.json" should contain the field "workspaces" array length equal to 1
    And the app data file "app.json" should contain the field "active_local_id" equal to "{{Acme.local_id}}"
    And the workspace file "acme/workspace.yaml" should exist

  @remove @validation
  Scenario Outline: Invalid removals are rejected - <case>
    When I call "WorkspaceService.Remove" with
      """
      { "local_id": "<local_id>", "move_to_trash": false }
      """
    Then the call should fail with code "<code>"

    Examples:
      | case       | local_id                   | code      |
      | no id      |                            | WKS-23400 |
      | unknown id | 01JA2Z5V3B6C7D8E9F0G1H2J3K | WKS-07404 |

  @locate @success
  Scenario: Locating a moved workspace folder
    Given the workspace path "acme" is moved to "moved/acme"
    When I call "WorkspaceService.List"
    Then the response should contain the field "workspaces.0.missing" equal to "true"
    When I call "WorkspaceService.Locate" with
      """
      { "local_id": "{{Acme.local_id}}", "path": "{{workspace_root}}/moved/acme" }
      """
    Then the call should succeed
    And the response should contain the field "local_id" equal to "{{Acme.local_id}}"
    And the response should contain the path field "path" equal to "{{workspace_root}}/moved/acme"
    And the response should contain the field "missing" equal to "false"
    And the app data file "app.json" should contain the field "workspaces" array length equal to 1
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed

  @locate @failure
  Scenario Outline: Locating applies the same checks as opening - <case>
    Given the workspace "Globex" exists
    And the workspace folder "plain" exists
    When I call "WorkspaceService.Locate" with
      """
      { "local_id": "<local_id>", "path": "<path>" }
      """
    Then the call should fail with code "<code>"

    Examples:
      | case                 | local_id                   | path                        | code      |
      | no workspace file    | {{Acme.local_id}}          | {{workspace_root}}/plain    | WKS-03404 |
      | overlaps another     | {{Acme.local_id}}          | {{workspace_root}}/globex   | WKS-02409 |
      | relative path        | {{Acme.local_id}}          | plain                       | WKS-24400 |
      | unknown workspace    | 01JA2Z5V3B6C7D8E9F0G1H2J3K | {{workspace_root}}/shared   | WKS-07404 |

  @locate @success
  Scenario: Locating a workspace at its own folder again is allowed
    When I call "WorkspaceService.Locate" with
      """
      { "local_id": "{{Acme.local_id}}", "path": "{{workspace_root}}/acme" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Acme"
