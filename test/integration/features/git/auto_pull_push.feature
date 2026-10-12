#language: en
#utf-8

@all @xen_007 @git @sync @auto_pull_push
Feature: Pull when the workspace is opened and push after each save
  Per workspace, "Pull when opened" pulls in the background every time the workspace is opened or
  switched to: a conflict only sets the status to Conflict (the sync session waits for the user)
  and changes nothing. "Push after each save" waits until 3 seconds pass without another save,
  then commits every local change with the message "Update <item names>" and pushes. Background
  failures never fail the save or the activation; they are surfaced in the sync status (the
  git:status-changed event and Status) with their next step. Both options are off by default.

  Background:
    Given the keychain is available
    And the git http server requires the username "oauth2" and the token "tok-sync"
    And the remote "acme/backend" has the file "README.md" with
      """
      Backend service
      """
    And the "repo" variable is set to the http url of the remote "acme/backend"
    And the "host" variable is set to the git http server host
    And I call "GitAccountService.AddToken" with
      """
      { "provider": "gitlab", "host": "{{host}}", "token": "tok-sync", "repository": "{{repo}}", "commit_name": "Luis", "commit_email": "luis@acme.test" }
      """
    And I save the response field "id" as "account"
    And the workspace "Acme" exists
    And the workspace file "acme/collections/api/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JS0000000000000000000001
      name: API
      order:
        - users.yaml
        - ping.yaml
      """
    And the workspace file "acme/collections/api/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: GET
      url: https://api.acme.test/ping
      """
    And the workspace file "acme/collections/api/users.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: GET
      url: https://api.acme.test/users
      """
    And I save the workspace file "acme/workspace.yaml" field "id" as "workspace_id"
    And I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme" }
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Add the Acme workspace", "item_ids": [ "{{workspace_id}}", "01JS0000000000000000000001", "01JS0000000000000000000002", "01JS0000000000000000000003" ] }
      """
    And I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """

  @defaults
  Scenario: Both options are off until they are turned on
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "settings.auto_pull" equal to "false"
    And the response should contain the field "settings.auto_push" equal to "false"
    When I call "SyncService.SaveSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "account_id": "{{account}}", "auto_pull": true, "auto_push": true }
      """
    And I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "settings.auto_pull" equal to "true"
    And the response should contain the field "settings.auto_push" equal to "true"
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "sync.autoPull" equal to "true"
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "sync.autoPush" equal to "true"

  @auto_pull
  Scenario: Opening the workspace pulls the teammates' changes in the background
    Given I call "SyncService.SaveSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "account_id": "{{account}}", "auto_pull": true, "auto_push": false }
      """
    And a teammate changes "api/xenon/collections/api/ping.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: OPTIONS
      url: https://api.acme.test/ping
      """
    And the emitted events are forgotten
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should contain the field "method" equal to "GET"
    When 1 second passes
    Then the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should contain the field "method" equal to "OPTIONS"
    And the last "git:status-changed" event should contain the field "local_id" equal to "{{Acme.local_id}}"
    And the last "git:status-changed" event should contain the field "status.state" equal to "up_to_date"
    And the event "workspace:tree-changed" should have been emitted 1 times
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "up_to_date"
    And the response should contain the field "behind" equal to "0"

  @auto_pull @conflict
  Scenario: A conflicting pull on opening only reports the conflict and changes nothing
    Given I call "SyncService.SaveSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "account_id": "{{account}}", "auto_pull": true, "auto_push": false }
      """
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: PATCH
      url: https://api.acme.test/users
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Users with PATCH", "item_ids": [ "01JS0000000000000000000003" ] }
      """
    And a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
      """
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And 1 second passes
    Then the last "git:status-changed" event should contain the field "status.state" equal to "conflict"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "method" equal to "PATCH"
    And the remote "acme/backend" branch "main" should have 3 commits
    When I call "SyncService.Session" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "conflicts" array length equal to 1
    And the response should contain the field "conflicts.0.item_name" equal to "Users"
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "conflict"

  @auto_pull @off
  Scenario: Without "Pull when opened" opening the workspace does not pull
    Given a teammate changes "api/xenon/collections/api/ping.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: OPTIONS
      url: https://api.acme.test/ping
      """
    And the emitted events are forgotten
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And 5 seconds pass
    Then the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should contain the field "method" equal to "GET"
    And the event "git:status-changed" should have been emitted 0 times

  @auto_pull @failure
  Scenario: A failing background pull is surfaced in the status and the workspace still opens
    Given I call "SyncService.SaveSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "account_id": "{{account}}", "auto_pull": true, "auto_push": false }
      """
    And the git http server goes offline
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    When 1 second passes
    Then the last "git:status-changed" event should contain the field "status.state" equal to "error"
    And the last "git:status-changed" event should contain the field "status.last_error.code" equal to "GIT-10503"
    And the last "git:status-changed" event should contain the field "status.last_error.next_step" equal to "Offline - check the connection and try again"

  @auto_push
  Scenario: Saves are pushed once 3 seconds pass without another save, as "Update <names>"
    Given I call "SyncService.SaveSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "account_id": "{{account}}", "auto_pull": false, "auto_push": true }
      """
    And I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JS0000000000000000000002" }
      """
    And I save the response field "revision" as "ping_revision"
    And I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JS0000000000000000000003" }
      """
    And I save the response field "revision" as "users_revision"
    When I call "RequestService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "base_revision": "{{ping_revision}}", "force": false, "request": { "id": "01JS0000000000000000000002", "name": "Ping", "method": "HEAD", "url": "https://api.acme.test/ping" } }
      """
    Then the call should succeed
    When 2 seconds pass
    Then the remote "acme/backend" branch "main" should have 2 commits
    When I call "RequestService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "base_revision": "{{users_revision}}", "force": false, "request": { "id": "01JS0000000000000000000003", "name": "Users", "method": "PUT", "url": "https://api.acme.test/users" } }
      """
    And 2 seconds pass
    Then the remote "acme/backend" branch "main" should have 2 commits
    When 1 second passes
    Then the remote "acme/backend" branch "main" should have 3 commits
    And the last commit on the remote "acme/backend" branch "main" should have the message containing "Update Ping, Users"
    And the last commit on the remote "acme/backend" branch "main" should have the author containing "Luis <luis@acme.test>"
    And the file "api/xenon/collections/api/ping.yaml" on the remote "acme/backend" should exist
    And the remote "acme/backend" history should not contain "tok-sync"
    And the last "git:status-changed" event should contain the field "status.state" equal to "up_to_date"
    When 5 seconds pass
    Then the remote "acme/backend" branch "main" should have 3 commits

  @auto_push @off
  Scenario: Without "Push after each save" saves stay local changes
    Given I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JS0000000000000000000002" }
      """
    And I save the response field "revision" as "ping_revision"
    When I call "RequestService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "base_revision": "{{ping_revision}}", "force": false, "request": { "id": "01JS0000000000000000000002", "name": "Ping", "method": "HEAD", "url": "https://api.acme.test/ping" } }
      """
    And 5 seconds pass
    Then the remote "acme/backend" branch "main" should have 2 commits
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "local_changes"

  @auto_push @environment
  Scenario: Saving an environment is pushed too
    Given I call "SyncService.SaveSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "account_id": "{{account}}", "auto_pull": false, "auto_push": true }
      """
    And I call "EnvironmentService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "hml" }
      """
    And I save the response field "id" as "hml"
    When I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}", "name": "hml", "description": "homologation", "color": "yellow", "confirm_before_send": true }
      """
    Then the call should succeed
    When 3 seconds pass
    Then the remote "acme/backend" branch "main" should have 3 commits
    And the last commit on the remote "acme/backend" branch "main" should have the message containing "Update hml"
    And the file "api/xenon/environments/hml.yaml" on the remote "acme/backend" should exist

  @auto_push @failure
  Scenario: A rejected background push is surfaced in the status without failing the save
    Given I call "SyncService.SaveSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "account_id": "{{account}}", "auto_pull": false, "auto_push": true }
      """
    And the remote "acme/backend" rejects pushes with "protected branch hook declined"
    And I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JS0000000000000000000002" }
      """
    And I save the response field "revision" as "ping_revision"
    When I call "RequestService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "base_revision": "{{ping_revision}}", "force": false, "request": { "id": "01JS0000000000000000000002", "name": "Ping", "method": "HEAD", "url": "https://api.acme.test/ping" } }
      """
    Then the call should succeed
    When 3 seconds pass
    Then the last "git:status-changed" event should contain the field "status.state" equal to "error"
    And the last "git:status-changed" event should contain the field "status.last_error.code" equal to "GIT-14403"
    And the last "git:status-changed" event should contain the field "status.ahead" equal to "1"
    And the remote "acme/backend" branch "main" should have 2 commits
