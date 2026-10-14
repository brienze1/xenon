#language: en
#utf-8

@all @xen_008 @mcp @mcp_workspaces
Feature: AI clients list, create, open, switch, rename and remove workspaces
  There is one shared active workspace: every workspace-scoped tool acts on it, and switching it through MCP switches
  the Xenon window too (workspaces:changed). With no active workspace, workspace-scoped tools answer MCP-06404 while
  listing, creating and opening workspaces still work. An id from another workspace is never reachable (MCP-08404).

  @list
  Scenario: Listing workspaces marks the active one
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    When I call the MCP tool "workspace" with
      """
      { "action": "list" }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 2
    And the response field "items" should contain an item with "id" equal to "{{Globex.local_id}}" and "active" equal to "true"
    And the response field "items" should contain an item with "id" equal to "{{Acme.local_id}}" and "name" equal to "Acme"
    And the response field "items" should contain 1 items with "active" equal to "true"

  @switch
  Scenario: Switching the workspace moves the window to it
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    And the emitted events are forgotten
    When I call the MCP tool "workspace_edit" with
      """
      { "action": "switch", "id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And a "workspaces:changed" event with the field "active_local_id" equal to "{{Acme.local_id}}" should have been emitted within 1000 ms
    And the app data file "app.json" should contain the field "active_local_id" equal to "{{Acme.local_id}}"
    When I call the MCP tool "workspace" with
      """
      { "action": "active" }
      """
    Then the response should contain the field "id" equal to "{{Acme.local_id}}"

  @create_open_rename
  Scenario: Creating, renaming and opening workspaces
    Given the workspace "Acme" exists
    When I call the MCP tool "workspace_edit" with
      """
      { "action": "create", "name": "Payments", "path": "{{workspace_root}}/payments" }
      """
    Then the call should succeed
    And I save the response field "id" as "payments"
    And the workspace folder "payments" should exist
    And the 1 "workspaces:changed" event should contain the field "active_local_id" equal to "{{payments}}"
    When I call the MCP tool "workspace_edit" with
      """
      { "action": "rename", "id": "{{payments}}", "name": "Payments API" }
      """
    Then the call should succeed
    And the workspace file "payments/workspace.yaml" should contain the field "name" equal to "Payments API"
    Given the workspace file "existing/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JW9000000000000000000000
      name: Existing
      """
    When I call the MCP tool "workspace_edit" with
      """
      { "action": "open", "path": "{{workspace_root}}/existing" }
      """
    Then the call should succeed
    When I call the MCP tool "workspace" with
      """
      { "action": "list" }
      """
    Then the response field "items" should contain an item with "name" equal to "Existing" and "active" equal to "true"

  @remove
  Scenario: Removing a workspace moves its folder to the trash
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    When I call the MCP tool "delete" with
      """
      { "target": "workspace", "id": "{{Acme.local_id}}", "move_to_trash": true }
      """
    Then the call should succeed
    And the trash should contain 1 items
    And the trash item 1 should be the workspace path "acme"
    When I call the MCP tool "workspace" with
      """
      { "action": "list" }
      """
    Then the response should contain the field "items" array length equal to 1

  @no_active_workspace
  Scenario Outline: Workspace-scoped tools need an active workspace - <tool>
    Given the workspace "Acme" exists
    And I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Acme.local_id}}", "move_to_trash": false }
      """
    When I call the MCP tool "<tool>" with
      """
      <arguments>
      """
    Then the call should fail with code "MCP-06404"
    And the error should contain the field "error.message" containing "No active workspace"

    Examples:
      | tool     | arguments                                                   |
      | tree     | { "action": "list" }                                        |
      | env      | { "action": "list" }                                        |
      | history  | { "action": "list" }                                        |
      | send     | { "request": { "method": "GET", "url": "https://x.test" } } |
      | workspace | { "action": "active" }                                     |

  @no_active_workspace
  Scenario: Listing and creating workspaces work with no active workspace
    Given the workspace "Acme" exists
    And I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Acme.local_id}}", "move_to_trash": false }
      """
    When I call the MCP tool "workspace" with
      """
      { "action": "list" }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 0
    When I call the MCP tool "workspace_edit" with
      """
      { "action": "create", "name": "Fresh", "path": "{{workspace_root}}/fresh" }
      """
    Then the call should succeed
    When I call the MCP tool "tree" with
      """
      { "action": "list" }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 0

  @boundary
  Scenario: An item from another workspace is not reachable
    Given the workspace "Acme" exists
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JW9100000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JW9100000000000000000001
      name: Billing
      order:
        - ping.yaml
      """
    And the workspace file "acme/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JW9100000000000000000002
      name: Ping
      method: GET
      url: https://acme.test/ping
      """
    And the workspace "Globex" exists
    When I call the MCP tool "tree" with
      """
      { "action": "get", "id": "01JW9100000000000000000002" }
      """
    Then the call should fail with code "MCP-08404"
    When I call the MCP tool "send" with
      """
      { "request_id": "01JW9100000000000000000002" }
      """
    Then the call should fail with code "MCP-08404"
    When I call the MCP tool "delete" with
      """
      { "target": "item", "id": "01JW9100000000000000000002" }
      """
    Then the call should fail with code "MCP-08404"
    And the workspace file "acme/collections/billing/ping.yaml" should exist
