#language: en
#utf-8

@all @xen_001 @workspaces @create_workspace
Feature: Create a workspace
  A workspace is a plain folder chosen by the user. Creating one writes the workspace file,
  the collections/ and environments/ folders and Xenon's .gitignore block, registers the
  folder in the local app data and makes it the active workspace.

  @success
  Scenario: Creating a workspace in a folder that does not exist yet
    When I call "WorkspaceService.Create" with
      """
      { "name": "Acme", "path": "{{workspace_root}}/acme" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Acme"
    And the response should contain the path field "path" equal to "{{workspace_root}}/acme"
    And the response should contain the field "missing" equal to "false"
    And the response field "local_id" should exist
    And I save the workspace file "acme/workspace.yaml" field "id" as "workspace_id"
    And the workspace file "acme/workspace.yaml" should equal
      """
      xenon: 1
      kind: workspace
      id: {{workspace_id}}
      name: Acme
      description: ""
      order: []
      variables: []
      sync: {}
      """
    And the workspace folder "acme/collections" should exist
    And the workspace folder "acme/environments" should exist
    And the workspace file "acme/.gitignore" should equal
      """
      # Xenon — local-only files (never commit)
      .xenon/
      *.xenon-tmp
      .xenon-import-*/
      """

  @success
  Scenario: Creating a workspace in an existing empty folder
    Given the workspace folder "empty" exists
    When I call "WorkspaceService.Create" with
      """
      { "name": "  Globex  ", "path": "{{workspace_root}}/empty" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Globex"
    And the workspace file "empty/workspace.yaml" should contain the field "name" equal to "Globex"
    And the workspace file "empty/workspace.yaml" should contain the field "kind" equal to "workspace"

  @success @registry
  Scenario: The new workspace is registered in the app data and becomes active
    When I call "WorkspaceService.Create" with
      """
      { "name": "Acme", "path": "{{workspace_root}}/acme" }
      """
    And I save the response field "local_id" as "acme_id"
    Then the app data file "app.json" should contain the field "version" equal to "1"
    And the app data file "app.json" should contain the field "active_local_id" equal to "{{acme_id}}"
    And the app data file "app.json" should contain the field "workspaces" array length equal to 1
    And the app data file "app.json" should contain the field "workspaces.0.local_id" equal to "{{acme_id}}"
    And the app data file "app.json" should contain the field "workspaces.0.display_name_cache" equal to "Acme"
    When I call "WorkspaceService.List"
    Then the call should succeed
    And the response should contain the field "active_local_id" equal to "{{acme_id}}"
    And the response should contain the field "workspaces" array length equal to 1
    And the response should contain the field "workspaces.0.name" equal to "Acme"

  @success @registry
  Scenario: A second workspace is appended to the registry and becomes active
    Given the workspace "Acme" exists
    When I call "WorkspaceService.Create" with
      """
      { "name": "Globex", "path": "{{workspace_root}}/globex" }
      """
    And I save the response field "local_id" as "globex_id"
    Then the app data file "app.json" should contain the field "workspaces" array length equal to 2
    And the app data file "app.json" should contain the field "workspaces.0.local_id" equal to "{{Acme.local_id}}"
    And the app data file "app.json" should contain the field "active_local_id" equal to "{{globex_id}}"

  @success @permissions
  Scenario: App data is owner-only on macOS and Linux
    Given the workspace "Acme" exists
    Then the app data path "." should have the permissions "0700"
    And the app data path "app.json" should have the permissions "0600"

  @validation
  Scenario Outline: An invalid name is rejected - <case>
    When I call "WorkspaceService.Create" with
      """
      { "name": "<name>", "path": "{{workspace_root}}/acme" }
      """
    Then the call should fail with code "WKS-00400"
    And the error should contain the field "error.error_details.0.attribute" equal to "name"
    And the workspace file "acme/workspace.yaml" should not exist
    And the app data file "app.json" should not exist

    Examples:
      | case     | name |
      | empty    |      |
      | too long | aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa |

  @validation @xen_014 @xen_014_errors
  Scenario: A blank name is rejected
    When I call "WorkspaceService.Create" with
      """
      { "name": "   ", "path": "{{workspace_root}}/acme" }
      """
    Then the call should fail with code "WKS-00400"
    And the error should contain the field "error.error_details.0.attribute" equal to "name"
    And the error should contain the field "error.error_details.0.messages.0" equal to "must not be blank"
    And the error should contain the field "error.description" equal to "invalid request: name must not be blank"

  @validation
  Scenario: A name of exactly 200 characters is accepted
    When I call "WorkspaceService.Create" with
      """
      { "name": "éééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééé", "path": "{{workspace_root}}/acme" }
      """
    Then the call should succeed

  @validation @xen_014 @xen_014_errors
  Scenario: A relative folder path is rejected
    When I call "WorkspaceService.Create" with
      """
      { "name": "Acme", "path": "relative/acme" }
      """
    Then the call should fail with code "WKS-00400"
    And the error should contain the field "error.error_details.0.attribute" equal to "path"
    And the error should contain the field "error.error_details.0.messages.0" equal to "must be an absolute path"
    And the error should contain the field "error.description" equal to "invalid request: path must be an absolute path"

  @conflict
  Scenario: A folder that is not empty is refused
    Given the workspace file "taken/readme.md" contains
      """
      # someone else's files
      """
    When I call "WorkspaceService.Create" with
      """
      { "name": "Acme", "path": "{{workspace_root}}/taken" }
      """
    Then the call should fail with code "WKS-01409"
    And the workspace file "taken/workspace.yaml" should not exist
    And the workspace file "taken/.gitignore" should not exist

  @conflict
  Scenario: A path that is a file is refused
    Given the workspace file "notes.txt" contains
      """
      not a folder
      """
    When I call "WorkspaceService.Create" with
      """
      { "name": "Acme", "path": "{{workspace_root}}/notes.txt" }
      """
    Then the call should fail with code "WKS-01409"

  @conflict
  Scenario Outline: A folder overlapping a registered workspace is refused - <case>
    Given the workspace "Acme" exists
    When I call "WorkspaceService.Create" with
      """
      { "name": "Other", "path": "<path>" }
      """
    Then the call should fail with code "WKS-02409"
    And the app data file "app.json" should contain the field "workspaces" array length equal to 1

    Examples:
      | case             | path                              |
      | same folder      | {{workspace_root}}/acme           |
      | inside it        | {{workspace_root}}/acme/nested    |
      | containing it    | {{workspace_root}}                |
