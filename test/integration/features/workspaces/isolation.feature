#language: en
#utf-8

@all @xen_001 @workspaces @isolation
Feature: Workspaces are a hard boundary
  Everything is scoped by the workspace's local id: the tree, search, UI state and file events
  of one workspace never show up in another, however often the user switches between them.
  Only the active workspace is watched for changes made outside Xenon.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JD0000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JD0000000000000000000001
      name: Billing
      order:
        - invoices.yaml
      """
    And the workspace file "acme/collections/billing/invoices.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JD0000000000000000000002
      name: List invoices
      method: GET
      url: https://acme.test/invoices
      """
    And the workspace file "globex/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JE0000000000000000000000
      name: Globex
      order:
        - shipping
      """
    And the workspace file "globex/collections/shipping/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JE0000000000000000000001
      name: Shipping
      order:
        - invoices.yaml
      """
    And the workspace file "globex/collections/shipping/invoices.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JE0000000000000000000002
      name: Shipping invoices
      method: POST
      url: https://globex.test/invoices
      """
    And I call "UIStateService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "tabs": [{ "tab_id": "acme-tab", "request_id": "", "pinned": true, "draft": { "name": "Acme draft" } }], "active_tab_id": "acme-tab" }
      """
    And I call "UIStateService.Save" with
      """
      { "local_id": "{{Globex.local_id}}", "tabs": [{ "tab_id": "globex-tab", "request_id": "01JE0000000000000000000002", "pinned": false }], "active_tab_id": "globex-tab", "preview_tab_id": "globex-tab" }
      """

  @switching
  Scenario: Each workspace keeps exactly its own tree and state across 100 switches
    When I call "WorkspaceService.Activate" 100 times cycling through the payloads
      """
      [
        { "local_id": "{{Acme.local_id}}" },
        { "local_id": "{{Globex.local_id}}" }
      ]
      """
    Then the call should succeed
    And the response should contain the field "summary.name" equal to "Globex"
    And the response should contain the field "tree" array length equal to 1
    And the response should contain the field "tree.0.name" equal to "Shipping"
    And the response should contain the field "tree.0.children.0.name" equal to "Shipping invoices"
    And the response should contain the field "ui_state.tabs" array length equal to 1
    And the response should contain the field "ui_state.tabs.0.tab_id" equal to "globex-tab"
    And the response should contain the field "ui_state.preview_tab_id" equal to "globex-tab"
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "summary.name" equal to "Acme"
    And the response should contain the field "tree" array length equal to 1
    And the response should contain the field "tree.0.name" equal to "Billing"
    And the response should contain the field "tree.0.children.0.name" equal to "List invoices"
    And the response should contain the field "ui_state.tabs" array length equal to 1
    And the response should contain the field "ui_state.tabs.0.draft.name" equal to "Acme draft"
    And the response should contain the field "ui_state.preview_tab_id" equal to ""
    And the file watcher should be watching the workspace path "acme"

  @tree
  Scenario: Tree and search only return the requested workspace's items
    When I call "TreeService.Search" with
      """
      { "local_id": "{{Acme.local_id}}", "query": "invoices" }
      """
    Then the call should succeed
    And the response should contain the field "match_ids" with the items "01JD0000000000000000000002"
    When I call "TreeService.Search" with
      """
      { "local_id": "{{Globex.local_id}}", "query": "invoices" }
      """
    Then the response should contain the field "match_ids" with the items "01JE0000000000000000000002"
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Globex.local_id}}" }
      """
    Then the response should contain the field "0.name" equal to "Shipping"
    And the response should contain the field "" array length equal to 1

  @requests
  Scenario: A request id from another workspace is not reachable
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "01JD0000000000000000000002" }
      """
    Then the call should fail with code "REQ-01404"

  @watcher
  Scenario: Only the active workspace is watched and reports outside changes
    Given I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Globex.local_id}}" }
      """
    And the workspace file "acme/collections/billing/refunds.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JD0000000000000000000003
      name: Refunds
      """
    When the file watcher reports a change to the workspace path "acme/collections/billing/refunds.yaml"
    Then the file watcher should be watching the workspace path "globex"
    And the event "workspace:tree-changed" should have been emitted 0 times
    And the event "request:changed-on-disk" should have been emitted 0 times

  @missing
  Scenario: Activating a workspace whose folder was deleted reports it as missing
    Given the workspace path "globex" is removed
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Globex.local_id}}" }
      """
    Then the call should fail with code "WKS-06404"
    And the event "workspace:folder-missing" should have been emitted 1 times
    And the 1 "workspace:folder-missing" event should contain the field "local_id" equal to "{{Globex.local_id}}"
