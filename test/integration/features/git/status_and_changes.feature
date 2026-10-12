#language: en
#utf-8

@all @xen_007 @git @sync @status_and_changes
Feature: Sync status and the list of changes
  The status of a linked workspace is one of up to date, local changes (n), ahead (n to push),
  behind (n to pull), diverged, conflict, error or syncing; Status fetches the branch to know what
  is behind. Changes lists the changed workspace items by their Xenon names and kinds (request,
  folder, collection, environment, workspace) as added, modified, deleted or renamed - only files
  under the subfolder count. A workspace whose account was removed shows "No account - choose
  one" until another account is chosen in its sync settings.

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
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JS0000000000000000000004
      name: dev
      variables: []
      """
    And I save the workspace file "acme/workspace.yaml" field "id" as "workspace_id"
    And I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme" }
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Add the Acme workspace", "item_ids": [ "{{workspace_id}}", "01JS0000000000000000000001", "01JS0000000000000000000002", "01JS0000000000000000000003", "01JS0000000000000000000004" ] }
      """
    And I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """

  @up_to_date
  Scenario: Nothing changed anywhere
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "up_to_date"
    And the response should contain the field "local_changes" equal to "0"
    And the response should contain the field "ahead" equal to "0"
    And the response should contain the field "behind" equal to "0"
    And the response field "last_sync_at" should exist
    And the response field "last_error" should not exist
    And the last "git:status-changed" event should contain the field "local_id" equal to "{{Acme.local_id}}"
    And the last "git:status-changed" event should contain the field "status.state" equal to "up_to_date"
    When I call "SyncService.Changes" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0

  @changes
  Scenario: Added, modified, deleted and renamed items are listed by name
    Given the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: HEAD
      url: https://api.acme.test/ping
      """
    And the workspace file "clones/acme/api/xenon/collections/api/orders.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000005
      name: Orders
      method: GET
      url: https://api.acme.test/orders
      """
    And the workspace path "clones/acme/api/xenon/environments/dev.yaml" is removed
    And the workspace path "clones/acme/api/xenon/collections/api/users.yaml" is moved to "clones/acme/api/xenon/collections/api/people.yaml"
    When I call "SyncService.Changes" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 4
    And the response field "" should contain an item with "item_name" equal to "Ping" and "change" equal to "modified"
    And the response field "" should contain an item with "item_name" equal to "Orders" and "change" equal to "added"
    And the response field "" should contain an item with "item_name" equal to "dev" and "change" equal to "deleted"
    And the response field "" should contain an item with "item_name" equal to "Users" and "change" equal to "renamed"
    And the response field "" should contain an item with "item_id" equal to "01JS0000000000000000000004" and "item_kind" equal to "environment"
    And the response field "" should contain an item with "item_id" equal to "01JS0000000000000000000003" and "item_kind" equal to "request"
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "local_changes"
    And the response should contain the field "local_changes" equal to "4"

  @local_only
  Scenario: Local-only files in the workspace folder are not changes
    Given the workspace file "clones/acme/api/xenon/.xenon/cache.json" contains
      """
      { "local": true }
      """
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml.xenon-tmp" contains
      """
      partial write
      """
    When I call "SyncService.Changes" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0

  @behind
  Scenario: A teammate's push shows as behind after a status refresh
    Given a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
      """
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "behind"
    And the response should contain the field "behind" equal to "1"
    And the response should contain the field "ahead" equal to "0"

  @diverged
  Scenario: Local commits and a teammate's push make the branch diverged
    Given the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: HEAD
      url: https://api.acme.test/ping
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Ping with HEAD", "item_ids": [ "01JS0000000000000000000002" ] }
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
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "diverged"
    And the response should contain the field "ahead" equal to "1"
    And the response should contain the field "behind" equal to "1"

  @no_account
  Scenario: A workspace whose account was removed asks for another one
    Given I call "GitAccountService.Remove" with
      """
      { "id": "{{account}}" }
      """
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "error"
    And the response should contain the field "last_error.code" equal to "GIT-06428"
    And the response should contain the field "last_error.message" equal to "No account - choose one"
    And the response should contain the field "settings.account_id" equal to ""
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should fail with code "GIT-07428"
    Given I call "GitAccountService.AddToken" with
      """
      { "provider": "gitlab", "host": "{{host}}", "token": "tok-sync", "repository": "{{repo}}", "commit_name": "Luis", "commit_email": "luis@acme.test" }
      """
    And I save the response field "id" as "new_account"
    When I call "SyncService.SaveSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "account_id": "{{new_account}}", "auto_pull": false, "auto_push": false }
      """
    Then the call should succeed
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "up_to_date"
    And the response should contain the field "settings.account_id" equal to "{{new_account}}"

  @shared_settings
  Scenario: A teammate changing the shared sync settings is shown as a notice and not applied
    Given a teammate changes "api/xenon/workspace.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: workspace
      id: {{workspace_id}}
      name: Acme
      sync:
        repository: {{repo}}
        branch: main
        subfolder: services/xenon
      """
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "up_to_date"
    And the response should contain the field "settings.subfolder" equal to "api/xenon"
    And the response should contain the field "notice" containing "services/xenon"

  @unknown
  Scenario: Unknown workspaces
    When I call "SyncService.Status" with
      """
      { "local_id": "01JZ0000000000000000000000" }
      """
    Then the call should fail with code "WKS-07404"
