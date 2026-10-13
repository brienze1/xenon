#language: en
#utf-8

@all @xen_001 @file_changes @external_changes
Feature: Pick up changes made outside Xenon
  While a workspace is active its folder is watched. Files added, changed or removed by another
  program refresh the sidebar; a changed or deleted request file is reported with its new
  revision so open tabs can reload or offer "Keep mine" / "Load from disk". Xenon's own writes
  and its temporary files are never reported.

  Background:
    Given the workspace "Acme" exists
    And the "ping" variable is set to "01JF0000000000000000000002"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JF0000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JF0000000000000000000001
      name: Billing
      order:
        - ping.yaml
      """
    And the workspace file "acme/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{ping}}
      name: Ping
      method: GET
      url: https://acme.test/ping
      """
    And I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{ping}}" }
      """
    And I save the response field "revision" as "original_revision"

  @watching
  Scenario: Activating a workspace starts watching its folder
    Then the file watcher should be watching the workspace path "acme"

  @added
  Scenario: A file added outside Xenon refreshes the tree
    Given the workspace file "acme/collections/billing/health.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JF0000000000000000000003
      name: Health
      """
    When the file watcher reports a change to the workspace path "acme/collections/billing/health.yaml"
    Then the event "workspace:tree-changed" should have been emitted 1 times
    And the 1 "workspace:tree-changed" event should contain the field "local_id" equal to "{{Acme.local_id}}"
    And the event "request:changed-on-disk" should have been emitted 0 times
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.children" array length equal to 2

  @changed
  Scenario: A request file edited outside Xenon is reported with its new revision
    Given the workspace file "acme/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{ping}}
      name: Ping v2
      method: POST
      url: https://acme.test/ping
      """
    When the file watcher reports a change to the workspace path "acme/collections/billing/ping.yaml"
    Then the event "request:changed-on-disk" should have been emitted 1 times
    And the 1 "request:changed-on-disk" event should contain the field "local_id" equal to "{{Acme.local_id}}"
    And the 1 "request:changed-on-disk" event should contain the field "request_id" equal to "{{ping}}"
    And the 1 "request:changed-on-disk" event should contain the field "deleted" equal to "false"
    And the event "workspace:tree-changed" should have been emitted 1 times
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{ping}}" }
      """
    Then the response should contain the field "name" equal to "Ping v2"
    And the response should contain the field "revision" not equal to "{{original_revision}}"
    And I save the response field "revision" as "new_revision"
    And the 1 "request:changed-on-disk" event should contain the field "revision" equal to "{{new_revision}}"

  @changed @conflict
  Scenario: Saving over an outside change needs "Keep mine"
    Given the workspace file "acme/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{ping}}
      name: Ping edited elsewhere
      method: GET
      url: https://acme.test/ping
      """
    And the file watcher reports a change to the workspace path "acme/collections/billing/ping.yaml"
    When I call "RequestService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "base_revision": "{{original_revision}}", "force": false, "request": { "id": "{{ping}}", "name": "Ping mine", "method": "GET", "url": "https://acme.test/ping" } }
      """
    Then the call should fail with code "REQ-02409"
    When I call "RequestService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "base_revision": "{{original_revision}}", "force": true, "request": { "id": "{{ping}}", "name": "Ping mine", "method": "GET", "url": "https://acme.test/ping" } }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "name" equal to "Ping mine"

  @deleted
  Scenario: A request file deleted outside Xenon is reported as deleted
    Given the workspace path "acme/collections/billing/ping.yaml" is removed
    When the file watcher reports the removal of the workspace path "acme/collections/billing/ping.yaml"
    Then the event "request:changed-on-disk" should have been emitted 1 times
    And the 1 "request:changed-on-disk" event should contain the field "request_id" equal to "{{ping}}"
    And the 1 "request:changed-on-disk" event should contain the field "deleted" equal to "true"
    And the 1 "request:changed-on-disk" event should contain the field "revision" equal to ""
    And the event "workspace:tree-changed" should have been emitted 1 times

  @deleted
  Scenario: A collection folder deleted outside Xenon reports its requests as deleted
    Given the workspace path "acme/collections/billing" is removed
    When the file watcher reports the removal of the workspace path "acme/collections/billing"
    Then the event "request:changed-on-disk" should have been emitted 1 times
    And the 1 "request:changed-on-disk" event should contain the field "request_id" equal to "{{ping}}"
    And the 1 "request:changed-on-disk" event should contain the field "deleted" equal to "true"
    And the event "workspace:tree-changed" should have been emitted 1 times

  @own_writes
  Scenario: Xenon's own saves are not reported
    Given I call "RequestService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "base_revision": "{{original_revision}}", "force": false, "request": { "id": "{{ping}}", "name": "Ping saved", "method": "GET", "url": "https://acme.test/ping" } }
      """
    When the file watcher reports a change to the workspace path "acme/collections/billing/ping.yaml"
    Then the event "request:changed-on-disk" should have been emitted 0 times
    And the event "workspace:tree-changed" should have been emitted 0 times

  @own_writes
  Scenario: Items created from the sidebar are not reported
    Given I call "TreeService.CreateRequest" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "01JF0000000000000000000001", "name": "Status" }
      """
    When the file watcher reports a change to the workspace path "acme/collections/billing/status.yaml"
    And the file watcher reports a change to the workspace path "acme/collections/billing/collection.yaml"
    Then the event "workspace:tree-changed" should have been emitted 0 times

  @own_writes
  Scenario: An own write that was changed again afterwards is reported
    Given I call "RequestService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "base_revision": "{{original_revision}}", "force": false, "request": { "id": "{{ping}}", "name": "Ping saved", "method": "GET", "url": "https://acme.test/ping" } }
      """
    And the workspace file "acme/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{ping}}
      name: Ping overwritten
      """
    When the file watcher reports a change to the workspace path "acme/collections/billing/ping.yaml"
    Then the event "request:changed-on-disk" should have been emitted 1 times

  @own_writes @expired
  Scenario: An own write older than two seconds is reported again
    Given the current time is "2026-10-09T10:00:00Z"
    And I call "RequestService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "base_revision": "{{original_revision}}", "force": false, "request": { "id": "{{ping}}", "name": "Ping saved", "method": "GET", "url": "https://acme.test/ping" } }
      """
    And the current time is "2026-10-09T10:00:03Z"
    When the file watcher reports a change to the workspace path "acme/collections/billing/ping.yaml"
    Then the event "request:changed-on-disk" should have been emitted 1 times

  @ignored
  Scenario: Temporary files and files outside the collections are ignored
    Given the workspace file "acme/collections/billing/.ping.yaml.xenon-tmp" contains
      """
      partial
      """
    And the workspace file "acme/.xenon/cache" contains
      """
      local
      """
    When the file watcher reports a change to the workspace path "acme/collections/billing/.ping.yaml.xenon-tmp"
    And the file watcher reports a change to the workspace path "acme/.xenon/cache"
    Then the event "workspace:tree-changed" should have been emitted 0 times
    And the event "request:changed-on-disk" should have been emitted 0 times

  @workspace_file
  Scenario: The workspace file edited outside Xenon refreshes the tree
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JF0000000000000000000000
      name: Acme
      order: []
      """
    When the file watcher reports a change to the workspace path "acme/workspace.yaml"
    Then the event "workspace:tree-changed" should have been emitted 1 times
    And the event "request:changed-on-disk" should have been emitted 0 times

  @missing
  Scenario: The active workspace folder disappearing is reported
    Given the workspace path "acme" is removed
    When the file watcher reports the removal of the workspace path "acme"
    Then the event "workspace:folder-missing" should have been emitted 1 times
    And the 1 "workspace:folder-missing" event should contain the field "local_id" equal to "{{Acme.local_id}}"
    And the event "workspace:tree-changed" should have been emitted 0 times
    And the file watcher should not be watching any folder

  @stop
  Scenario: Removing the active workspace stops watching it
    When I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Acme.local_id}}", "move_to_trash": false }
      """
    Then the call should succeed
    And the file watcher should not be watching any folder
