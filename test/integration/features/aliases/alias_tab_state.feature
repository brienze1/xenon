#language: en
#utf-8

@all @xen_016 @aliases @alias_tab_state
Feature: A request tab remembers the alias it shows and the unsaved draft of every version
  The tab state kept in workspaces/<localId>/state.json gains the alias the tab shows ("alias_id",
  empty for the Default) and the unsaved drafts of the versions not on screen ("version_drafts",
  keyed by alias id, "" for the Default). "draft" keeps meaning the draft of the version on screen,
  so state written before aliases existed loads unchanged.

  Background:
    Given the workspace "Acme" exists
    And the "orders" variable is set to "01JT0000000000000000000002"
    And the "canary" variable is set to "01JT0000000000000000000010"
    And the "v2" variable is set to "01JT0000000000000000000011"

  @roundtrip
  Scenario: The shown alias and the drafts of the other versions survive a restart
    When I call "UIStateService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tabs": [
          {
            "tab_id": "tab-orders",
            "request_id": "{{orders}}",
            "pinned": true,
            "base_revision": "rev-1",
            "alias_id": "{{canary}}",
            "draft": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "https://api.acme.test/v1/orders?canary=edited" },
            "version_drafts": {
              "": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "https://api.acme.test/v1/orders?default=edited" },
              "{{v2}}": { "id": "{{orders}}", "name": "Orders", "method": "GET", "url": "https://api.acme.test/v2/orders?edited=1" }
            }
          }
        ],
        "active_tab_id": "tab-orders"
      }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "tabs.0.alias_id" equal to "{{canary}}"
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "tabs.0.tab_id" equal to "tab-orders"
    And the response should contain the field "tabs.0.alias_id" equal to "{{canary}}"
    And the response should contain the field "tabs.0.draft.url" equal to "https://api.acme.test/v1/orders?canary=edited"
    And the response should contain the field "tabs.0.version_drafts..url" equal to "https://api.acme.test/v1/orders?default=edited"
    And the response should contain the field "tabs.0.version_drafts.01JT0000000000000000000011.url" equal to "https://api.acme.test/v2/orders?edited=1"
    And the workspace folder "acme" should contain the entries ".gitignore, collections, environments, workspace.yaml"

  @backward_compatible
  Scenario: Tab state without aliases loads with the Default shown and no version drafts
    Given the app data file "workspaces/{{Acme.local_id}}/state.json" contains
      """
      {
        "version": 1,
        "tabs": [ { "tab_id": "tab-orders", "request_id": "{{orders}}", "pinned": true, "base_revision": "rev-1" } ],
        "active_tab_id": "tab-orders"
      }
      """
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "tabs.0.request_id" equal to "{{orders}}"
    And the response field "tabs.0.alias_id" should not exist
    And the response field "tabs.0.version_drafts" should not exist
    When I call "UIStateService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tabs": [ { "tab_id": "tab-orders", "request_id": "{{orders}}", "pinned": true, "base_revision": "rev-1", "alias_id": "{{canary}}" } ],
        "active_tab_id": "tab-orders"
      }
      """
    Then the call should succeed
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "tabs.0.alias_id" equal to "{{canary}}"
