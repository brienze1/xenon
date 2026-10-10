#language: en
#utf-8

@all @xen_001 @ui_state @ui_state_roundtrip
Feature: Remember tabs, drafts, sidebar and layout per workspace
  Each workspace keeps its open tabs (order, active tab, preview tab, pinned flags, unsaved
  drafts and the revision each draft started from), its sidebar state and its layout in the
  local app data, under workspaces/<localId>/state.json. Nothing of it is ever written into the
  workspace folder, so it never ends up in the workspace's git repository.

  Background:
    Given the workspace "Acme" exists

  @defaults
  Scenario: A workspace without saved state gets the defaults
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "local_id" equal to "{{Acme.local_id}}"
    And the response should contain the field "version" equal to "1"
    And the response should contain the field "tabs" array length equal to 0
    And the response should contain the field "active_tab_id" equal to ""
    And the response should contain the field "preview_tab_id" equal to ""
    And the response should contain the field "sidebar.expanded" array length equal to 0
    And the response should contain the field "sidebar.selected_id" equal to ""
    And the response should contain the field "sidebar.view" equal to "collections"
    And the response should contain the field "sidebar.search" equal to ""
    And the response should contain the field "layout.sidebar_width" equal to "280"
    And the response should contain the field "layout.sidebar_visible" equal to "true"
    And the response should contain the field "layout.right_panel_width" equal to "320"
    And the response should contain the field "layout.right_panel_visible" equal to "true"
    And the response should contain the field "layout.right_panel_view" equal to "info"
    And the response should contain the field "layout.console_height" equal to "200"
    And the response should contain the field "layout.console_visible" equal to "false"
    And the response should contain the field "layout.split" equal to "stacked"
    And the response should contain the field "layout.split_ratio" equal to "0.5"
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should not exist

  @success
  Scenario: Tabs, drafts, sidebar and layout survive a round trip
    When I call "UIStateService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tabs": [
          { "tab_id": "tab-saved", "request_id": "01JB0000000000000000000004", "pinned": true, "base_revision": "abc123" },
          {
            "tab_id": "tab-new",
            "request_id": "",
            "pinned": true,
            "base_revision": "",
            "draft": {
              "id": "",
              "name": "Untitled Request",
              "method": "post",
              "url": "https://api.example.test/users?page=2",
              "params": [{ "key": "page", "value": "2", "enabled": true, "description": "" }],
              "headers": [{ "key": "X-Trace", "value": "on", "enabled": false, "description": "debug" }],
              "auth": { "type": "bearer", "bearer": { "token": "{{token}}" } },
              "body": { "mode": "raw", "raw": { "language": "json", "content": "hello body" } }
            }
          },
          { "tab_id": "tab-preview", "request_id": "01JB0000000000000000000005", "pinned": false, "base_revision": "def456" }
        ],
        "active_tab_id": "tab-new",
        "preview_tab_id": "tab-preview",
        "sidebar": { "expanded": ["01JB0000000000000000000001"], "selected_id": "01JB0000000000000000000004", "view": "history", "search": "user" },
        "layout": {
          "sidebar_width": 320, "sidebar_visible": false, "right_panel_width": 300, "right_panel_visible": false,
          "right_panel_view": "info", "console_height": 240, "console_visible": true, "split": "side-by-side", "split_ratio": 0.6
        }
      }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "version" equal to "1"
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "tabs" array length equal to 3
    And the app data path "workspaces/{{Acme.local_id}}/state.json" should have the permissions "0600"
    And the workspace folder "acme" should contain the entries ".gitignore, collections, environments, workspace.yaml"
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "tabs" array length equal to 3
    And the response should contain the field "tabs.0.tab_id" equal to "tab-saved"
    And the response should contain the field "tabs.0.request_id" equal to "01JB0000000000000000000004"
    And the response should contain the field "tabs.0.pinned" equal to "true"
    And the response should contain the field "tabs.0.base_revision" equal to "abc123"
    And the response field "tabs.0.draft" should not exist
    And the response should contain the field "tabs.1.tab_id" equal to "tab-new"
    And the response should contain the field "tabs.1.request_id" equal to ""
    And the response should contain the field "tabs.1.draft.name" equal to "Untitled Request"
    And the response should contain the field "tabs.1.draft.method" equal to "POST"
    And the response should contain the field "tabs.1.draft.url" equal to "https://api.example.test/users?page=2"
    And the response should contain the field "tabs.1.draft.params.0.key" equal to "page"
    And the response should contain the field "tabs.1.draft.headers.0.enabled" equal to "false"
    And the response should contain the field "tabs.1.draft.headers.0.description" equal to "debug"
    And the response should contain the field "tabs.1.draft.auth.type" equal to "bearer"
    And the response should contain the field "tabs.1.draft.auth.bearer.token" equal to "{{token}}"
    And the response should contain the field "tabs.1.draft.body.mode" equal to "raw"
    And the response should contain the field "tabs.1.draft.body.raw.content" equal to "hello body"
    And the response should contain the field "tabs.2.tab_id" equal to "tab-preview"
    And the response should contain the field "tabs.2.pinned" equal to "false"
    And the response should contain the field "active_tab_id" equal to "tab-new"
    And the response should contain the field "preview_tab_id" equal to "tab-preview"
    And the response should contain the field "sidebar.expanded" with the items "01JB0000000000000000000001"
    And the response should contain the field "sidebar.selected_id" equal to "01JB0000000000000000000004"
    And the response should contain the field "sidebar.view" equal to "history"
    And the response should contain the field "sidebar.search" equal to "user"
    And the response should contain the field "layout.sidebar_width" equal to "320"
    And the response should contain the field "layout.sidebar_visible" equal to "false"
    And the response should contain the field "layout.right_panel_width" equal to "300"
    And the response should contain the field "layout.right_panel_visible" equal to "false"
    And the response should contain the field "layout.console_height" equal to "240"
    And the response should contain the field "layout.console_visible" equal to "true"
    And the response should contain the field "layout.split" equal to "side-by-side"
    And the response should contain the field "layout.split_ratio" equal to "0.6"

  @activate
  Scenario: Activating a workspace returns its saved state with its tree
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JB0000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JB0000000000000000000001
      name: Billing
      order: []
      """
    And I call "UIStateService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "tabs": [{ "tab_id": "tab-1", "request_id": "", "pinned": true, "draft": { "name": "Scratch", "url": "https://x.test" } }], "active_tab_id": "tab-1" }
      """
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "summary.local_id" equal to "{{Acme.local_id}}"
    And the response should contain the field "tree" array length equal to 1
    And the response should contain the field "tree.0.name" equal to "Billing"
    And the response should contain the field "ui_state.active_tab_id" equal to "tab-1"
    And the response should contain the field "ui_state.tabs.0.draft.name" equal to "Scratch"

  @normalize
  Scenario: Inconsistent state is normalised before it is stored
    When I call "UIStateService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tabs": [
          { "tab_id": "tab-a", "request_id": "01JB0000000000000000000004", "pinned": true },
          { "tab_id": "tab-a", "request_id": "01JB0000000000000000000005", "pinned": false },
          { "tab_id": "tab-b", "request_id": "01JB0000000000000000000006", "pinned": true }
        ],
        "active_tab_id": "tab-gone",
        "preview_tab_id": "tab-b",
        "layout": { "sidebar_width": 50, "right_panel_width": 9000, "console_height": 10, "split_ratio": 0.95 }
      }
      """
    Then the call should succeed
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "tabs" array length equal to 2
    And the response should contain the field "tabs.0.request_id" equal to "01JB0000000000000000000004"
    And the response should contain the field "tabs.1.tab_id" equal to "tab-b"
    And the response should contain the field "active_tab_id" equal to "tab-a"
    And the response should contain the field "preview_tab_id" equal to ""
    And the response should contain the field "layout.sidebar_width" equal to "200"
    And the response should contain the field "layout.right_panel_width" equal to "600"
    And the response should contain the field "layout.console_height" equal to "120"
    And the response should contain the field "layout.split_ratio" equal to "0.8"
    And the response should contain the field "layout.split" equal to "stacked"
    And the response should contain the field "sidebar.view" equal to "collections"

  @normalize
  Scenario: A preview tab must be the only unpinned preview and an empty tab list has no active tab
    When I call "UIStateService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tabs": [],
        "active_tab_id": "tab-a",
        "preview_tab_id": "tab-a",
        "layout": { "sidebar_width": 600, "right_panel_width": 240, "console_height": 120, "split_ratio": 0.2 }
      }
      """
    Then the call should succeed
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "tabs" array length equal to 0
    And the response should contain the field "active_tab_id" equal to ""
    And the response should contain the field "preview_tab_id" equal to ""
    And the response should contain the field "layout.sidebar_width" equal to "600"
    And the response should contain the field "layout.right_panel_width" equal to "240"
    And the response should contain the field "layout.console_height" equal to "120"
    And the response should contain the field "layout.split_ratio" equal to "0.2"

  @isolation
  Scenario: Each workspace keeps its own state
    Given the workspace "Globex" exists
    And I call "UIStateService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "tabs": [{ "tab_id": "acme-tab", "request_id": "", "pinned": true, "draft": { "name": "Acme only" } }], "active_tab_id": "acme-tab" }
      """
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Globex.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "tabs" array length equal to 0
    And the app data file "workspaces/{{Globex.local_id}}/state.json" should not exist

  @corrupted
  Scenario: A state file that cannot be read falls back to the defaults
    Given the app data file "workspaces/{{Acme.local_id}}/state.json" contains
      """
      { "version": 1, "tabs": [ broken
      """
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "tabs" array length equal to 0
    And the response should contain the field "layout.split" equal to "stacked"

  @validation
  Scenario Outline: Invalid UI state requests are rejected - <case>
    When I call "UIStateService.<method>" with
      """
      <payload>
      """
    Then the call should fail with code "<code>"

    Examples:
      | case             | method | payload                                                                              | code      |
      | get without id   | Get    | { "local_id": "" }                                                                   | UIS-00400 |
      | save without id  | Save   | { "local_id": " " }                                                                  | UIS-01400 |
      | tab without id   | Save   | { "local_id": "{{Acme.local_id}}", "tabs": [{ "tab_id": "" }] }                      | UIS-01400 |
      | unknown view     | Save   | { "local_id": "{{Acme.local_id}}", "sidebar": { "view": "graph" } }                  | UIS-01400 |
      | unknown split    | Save   | { "local_id": "{{Acme.local_id}}", "layout": { "split": "diagonal" } }               | UIS-01400 |
      | unknown get      | Get    | { "local_id": "01JA2Z5V3B6C7D8E9F0G1H2J3K" }                                         | WKS-07404 |
      | unknown save     | Save   | { "local_id": "01JA2Z5V3B6C7D8E9F0G1H2J3K" }                                         | WKS-07404 |

  @xen_002 @right_panel
  Scenario Outline: The right panel remembers its <view> view
    When I call "UIStateService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "tabs": [], "layout": { "right_panel_view": "<view>" } }
      """
    Then the call should succeed
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "layout.right_panel_view" equal to "<view>"

    Examples:
      | view      |
      | info      |
      | variables |
