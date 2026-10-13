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

    @xen_005 @code_view
    Examples:
      | view |
      | code |

  @xen_005 @closed_tabs
  Scenario: The closed-tab stack keeps the 20 most recently closed tabs with their drafts
    When I call "UIStateService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tabs": [],
        "closed_tabs": [
          { "closed_at": "2026-10-11T10:00:00Z", "tab": { "tab_id": "t1", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:01Z", "tab": { "tab_id": "t2", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:02Z", "tab": { "tab_id": "t3", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:03Z", "tab": { "tab_id": "t4", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:04Z", "tab": { "tab_id": "t5", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:05Z", "tab": { "tab_id": "t6", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:06Z", "tab": { "tab_id": "t7", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:07Z", "tab": { "tab_id": "t8", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:08Z", "tab": { "tab_id": "t9", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:09Z", "tab": { "tab_id": "t10", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:10Z", "tab": { "tab_id": "t11", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:11Z", "tab": { "tab_id": "t12", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:12Z", "tab": { "tab_id": "t13", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:13Z", "tab": { "tab_id": "t14", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:14Z", "tab": { "tab_id": "t15", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:15Z", "tab": { "tab_id": "t16", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:16Z", "tab": { "tab_id": "t17", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:17Z", "tab": { "tab_id": "t18", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:18Z", "tab": { "tab_id": "t19", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:19Z", "tab": { "tab_id": "t20", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:20Z", "tab": { "tab_id": "t21", "request_id": "", "pinned": true } },
          { "closed_at": "2026-10-11T10:00:21Z", "tab": { "tab_id": "t22", "request_id": "01JA2Z5V3B6C7D8E9F0G1H2J3K", "pinned": true, "base_revision": "r1", "draft": { "name": "Edited", "method": "PUT", "url": "https://api.example.test/edited" } } }
        ]
      }
      """
    Then the call should succeed
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "closed_tabs" array length equal to 20
    And the response should contain the field "closed_tabs.0.tab.tab_id" equal to "t3"
    And the response should contain the field "closed_tabs.19.tab.tab_id" equal to "t22"
    And the response should contain the field "closed_tabs.19.tab.draft.name" equal to "Edited"
    And the response should contain the field "closed_tabs.19.tab.draft.method" equal to "PUT"
    And the response should contain the field "closed_tabs.19.tab.base_revision" equal to "r1"
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should exist
    And no file under the workspace folder should contain "api.example.test/edited"

  @examples @xen_003
  Scenario: An example tab keeps its kind and example id
    When I call "UIStateService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tabs": [
          { "tab_id": "tab-request", "request_id": "01JB0000000000000000000004", "pinned": true, "base_revision": "abc123" },
          { "tab_id": "tab-example", "kind": "example", "request_id": "01JB0000000000000000000004", "example_id": "01JB0000000000000000000044", "pinned": false, "base_revision": "abc123" }
        ],
        "active_tab_id": "tab-example",
        "preview_tab_id": "tab-example",
        "sidebar": { "expanded": [], "selected_id": "", "view": "collections", "search": "" },
        "layout": {
          "sidebar_width": 280, "sidebar_visible": true, "right_panel_width": 320, "right_panel_visible": true,
          "right_panel_view": "info", "console_height": 200, "console_visible": false, "split": "stacked", "split_ratio": 0.5
        }
      }
      """
    Then the call should succeed
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "tabs.0.kind" equal to "request"
    And the response should contain the field "tabs.0.example_id" equal to ""
    And the response should contain the field "tabs.1.kind" equal to "example"
    And the response should contain the field "tabs.1.request_id" equal to "01JB0000000000000000000004"
    And the response should contain the field "tabs.1.example_id" equal to "01JB0000000000000000000044"
    And the response should contain the field "preview_tab_id" equal to "tab-example"

  @environment_tabs @xen_007
  Scenario: Environment and globals tabs keep their kind and environment id
    When I call "UIStateService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tabs": [
          { "tab_id": "tab-request", "request_id": "01JB0000000000000000000004", "pinned": true, "base_revision": "abc123" },
          { "tab_id": "tab-env", "kind": "environment", "request_id": "", "environment_id": "01JB0000000000000000000077", "pinned": false },
          { "tab_id": "tab-globals", "kind": "globals", "request_id": "", "pinned": true }
        ],
        "active_tab_id": "tab-env",
        "preview_tab_id": "tab-env",
        "closed_tabs": [
          { "closed_at": "2026-10-12T10:00:00Z", "tab": { "tab_id": "tab-closed-env", "kind": "environment", "request_id": "", "environment_id": "01JB0000000000000000000078", "pinned": true } }
        ],
        "sidebar": { "expanded": [], "selected_id": "", "view": "environments", "search": "" }
      }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "tabs.1.kind" equal to "environment"
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "tabs.1.environment_id" equal to "01JB0000000000000000000077"
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "tabs" array length equal to 3
    And the response should contain the field "tabs.0.kind" equal to "request"
    And the response should contain the field "tabs.0.environment_id" equal to ""
    And the response should contain the field "tabs.1.kind" equal to "environment"
    And the response should contain the field "tabs.1.environment_id" equal to "01JB0000000000000000000077"
    And the response should contain the field "tabs.1.pinned" equal to "false"
    And the response should contain the field "tabs.2.kind" equal to "globals"
    And the response should contain the field "tabs.2.environment_id" equal to ""
    And the response should contain the field "active_tab_id" equal to "tab-env"
    And the response should contain the field "preview_tab_id" equal to "tab-env"
    And the response should contain the field "closed_tabs.0.tab.kind" equal to "environment"
    And the response should contain the field "closed_tabs.0.tab.environment_id" equal to "01JB0000000000000000000078"
    And the response should contain the field "sidebar.view" equal to "environments"

  @environment_tabs @xen_007
  Scenario: A state file written before environment tabs existed still loads
    Given the app data file "workspaces/{{Acme.local_id}}/state.json" contains
      """
      {
        "version": 1,
        "tabs": [
          { "tab_id": "tab-old", "request_id": "01JB0000000000000000000004", "example_id": "", "base_revision": "abc123", "pinned": true },
          { "tab_id": "tab-old-example", "request_id": "01JB0000000000000000000004", "example_id": "01JB0000000000000000000044", "base_revision": "", "pinned": true }
        ],
        "active_tab_id": "tab-old",
        "preview_tab_id": "",
        "sidebar": { "view": "history", "selected_id": "", "search": "", "expanded": [] },
        "layout": { "sidebar_width": 280, "sidebar_visible": true, "right_panel_width": 320, "right_panel_visible": true, "right_panel_view": "info", "console_height": 200, "console_visible": false, "split": "stacked", "split_ratio": 0.5 }
      }
      """
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "tabs" array length equal to 2
    And the response should contain the field "tabs.0.kind" equal to "request"
    And the response should contain the field "tabs.0.environment_id" equal to ""
    And the response should contain the field "tabs.1.kind" equal to "example"
    And the response should contain the field "sidebar.view" equal to "history"

  @environment_tabs @xen_007
  Scenario: A kind that does not match its ids falls back to what the ids say
    When I call "UIStateService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tabs": [
          { "tab_id": "tab-a", "kind": "environment", "request_id": "01JB0000000000000000000004", "pinned": true },
          { "tab_id": "tab-b", "kind": "unheard-of", "request_id": "", "environment_id": "01JB0000000000000000000077", "pinned": true }
        ]
      }
      """
    Then the call should succeed
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "tabs.0.kind" equal to "request"
    And the response should contain the field "tabs.1.kind" equal to "environment"

  @item_tabs @xen_007
  Scenario: Collection and folder tabs keep their kind and item id
    When I call "UIStateService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tabs": [
          { "tab_id": "tab-request", "request_id": "01JB0000000000000000000004", "pinned": true, "base_revision": "abc123" },
          { "tab_id": "tab-collection", "kind": "collection", "request_id": "", "item_id": "01JB0000000000000000000001", "pinned": false },
          { "tab_id": "tab-folder", "kind": "folder", "request_id": "", "item_id": "01JB0000000000000000000002", "pinned": true }
        ],
        "active_tab_id": "tab-collection",
        "preview_tab_id": "tab-collection",
        "closed_tabs": [
          { "closed_at": "2026-10-12T10:00:00Z", "tab": { "tab_id": "tab-closed-folder", "kind": "folder", "request_id": "", "item_id": "01JB0000000000000000000003", "pinned": true } }
        ]
      }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "tabs.1.kind" equal to "collection"
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "tabs.1.item_id" equal to "01JB0000000000000000000001"
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "tabs.2.kind" equal to "folder"
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "tabs" array length equal to 3
    And the response should contain the field "tabs.0.kind" equal to "request"
    And the response should contain the field "tabs.0.item_id" equal to ""
    And the response should contain the field "tabs.1.kind" equal to "collection"
    And the response should contain the field "tabs.1.item_id" equal to "01JB0000000000000000000001"
    And the response should contain the field "tabs.1.pinned" equal to "false"
    And the response should contain the field "tabs.2.kind" equal to "folder"
    And the response should contain the field "tabs.2.item_id" equal to "01JB0000000000000000000002"
    And the response should contain the field "preview_tab_id" equal to "tab-collection"
    And the response should contain the field "closed_tabs.0.tab.kind" equal to "folder"
    And the response should contain the field "closed_tabs.0.tab.item_id" equal to "01JB0000000000000000000003"

  @item_tabs @xen_007
  Scenario: A state file written before collection and folder tabs existed still loads
    Given the app data file "workspaces/{{Acme.local_id}}/state.json" contains
      """
      {
        "version": 1,
        "tabs": [
          { "tab_id": "tab-old", "kind": "request", "request_id": "01JB0000000000000000000004", "example_id": "", "environment_id": "", "base_revision": "abc123", "pinned": true },
          { "tab_id": "tab-env", "kind": "environment", "request_id": "", "environment_id": "01JB0000000000000000000077", "pinned": true },
          { "tab_id": "tab-globals", "kind": "globals", "request_id": "", "pinned": false }
        ],
        "active_tab_id": "tab-old",
        "preview_tab_id": "tab-globals",
        "sidebar": { "view": "collections", "selected_id": "", "search": "", "expanded": [] },
        "layout": { "sidebar_width": 280, "sidebar_visible": true, "right_panel_width": 320, "right_panel_visible": true, "right_panel_view": "info", "console_height": 200, "console_visible": false, "split": "stacked", "split_ratio": 0.5 }
      }
      """
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "tabs" array length equal to 3
    And the response should contain the field "tabs.0.kind" equal to "request"
    And the response should contain the field "tabs.0.item_id" equal to ""
    And the response should contain the field "tabs.1.kind" equal to "environment"
    And the response should contain the field "tabs.1.item_id" equal to ""
    And the response should contain the field "tabs.2.kind" equal to "globals"
    And the response should contain the field "preview_tab_id" equal to "tab-globals"

  @item_tabs @xen_007
  Scenario: A collection or folder kind without an item id falls back to what the ids say
    When I call "UIStateService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tabs": [
          { "tab_id": "tab-a", "kind": "collection", "request_id": "01JB0000000000000000000004", "pinned": true },
          { "tab_id": "tab-b", "kind": "folder", "request_id": "", "environment_id": "01JB0000000000000000000077", "pinned": true },
          { "tab_id": "tab-c", "kind": "unheard-of", "request_id": "", "item_id": "01JB0000000000000000000001", "pinned": true },
          { "tab_id": "tab-d", "kind": "folder", "request_id": "", "item_id": "01JB0000000000000000000002", "pinned": true }
        ]
      }
      """
    Then the call should succeed
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "tabs.0.kind" equal to "request"
    And the response should contain the field "tabs.1.kind" equal to "environment"
    And the response should contain the field "tabs.2.kind" equal to "collection"
    And the response should contain the field "tabs.2.item_id" equal to "01JB0000000000000000000001"
    And the response should contain the field "tabs.3.kind" equal to "folder"
