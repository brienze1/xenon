#language: en
#utf-8

@all @xen_007 @git @sync @commit_push_pull
Feature: Commit, push and pull
  A commit takes only the checked items, only under the subfolder, authored with the account's
  commit name and email; a message is required. Push sends the branch. Pull fetches and
  fast-forwards: the teammate's changes are written into the workspace, open clean tabs reload
  (request:changed-on-disk) and the sidebar refreshes once (workspace:tree-changed); the files a
  pull writes are Xenon's own writes, so the file watcher does not report them again.

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

  @commit
  Scenario: Only the checked items are committed, with the account's identity
    Given the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: HEAD
      url: https://api.acme.test/ping
      """
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: DELETE
      url: https://api.acme.test/users
      """
    When I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Ping with HEAD", "item_ids": [ "01JS0000000000000000000002" ] }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "ahead"
    And the response should contain the field "ahead" equal to "1"
    And the response should contain the field "local_changes" equal to "1"
    And the last "git:status-changed" event should contain the field "status.state" equal to "ahead"
    When I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "local_changes"
    And the response should contain the field "ahead" equal to "0"
    And the remote "acme/backend" branch "main" should have 3 commits
    And the file "api/xenon/collections/api/ping.yaml" on the remote "acme/backend" should equal
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: HEAD
      url: https://api.acme.test/ping
      """
    And the file "api/xenon/collections/api/users.yaml" on the remote "acme/backend" should equal
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: GET
      url: https://api.acme.test/users
      """
    And the last commit on the remote "acme/backend" branch "main" should have the message containing "Ping with HEAD"
    And the last commit on the remote "acme/backend" branch "main" should have the author containing "Luis <luis@acme.test>"
    And the file "README.md" on the remote "acme/backend" should equal
      """
      Backend service
      """

  @commit
  Scenario: The commit uses the account's edited identity and a deletion is committed
    Given I call "GitAccountService.Update" with
      """
      { "id": "{{account}}", "display_name": "Work", "commit_name": "Luis Brienze", "commit_email": "lb@acme.test" }
      """
    And the workspace path "clones/acme/api/xenon/collections/api/users.yaml" is removed
    When I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Drop users", "item_ids": [ "01JS0000000000000000000003" ] }
      """
    And I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "up_to_date"
    And the file "api/xenon/collections/api/users.yaml" on the remote "acme/backend" should not exist
    And the file "api/xenon/collections/api/ping.yaml" on the remote "acme/backend" should exist
    And the last commit on the remote "acme/backend" branch "main" should have the author containing "Luis Brienze <lb@acme.test>"

  @commit @validation
  Scenario Outline: A commit needs a message and items - <case>
    When I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "<message>", "item_ids": <items> }
      """
    Then the call should fail with code "<code>"

    Examples:
      | case            | message | items                            | code      |
      | empty message   |         | [ "01JS0000000000000000000002" ] | GIT-00400 |
      | blank message   |    | [ "01JS0000000000000000000002" ] | GIT-00400 |
      | no items        | Update  | [ ]                              | GIT-00400 |
      | nothing changed | Update  | [ "01JS0000000000000000000002" ] | GIT-08400 |

  @pull
  Scenario: Pull fast-forwards the teammate's change, reloads clean tabs and refreshes the tree once
    Given a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
      """
    And a teammate changes "api/xenon/collections/api/orders.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000005
      name: Orders
      method: GET
      url: https://api.acme.test/orders
      """
    And a teammate deletes "api/xenon/collections/api/ping.yaml" on the remote "acme/backend"
    And a teammate changes "docs/outside.md" on the remote "acme/backend" to
      """
      not a workspace file
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "status.state" equal to "up_to_date"
    And the response field "session" should not exist
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "method" equal to "POST"
    And the workspace file "clones/acme/api/xenon/collections/api/orders.yaml" should contain the field "name" equal to "Orders"
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should not exist
    And the workspace file "clones/acme/docs/outside.md" should not exist
    And the event "workspace:tree-changed" should have been emitted 1 times
    And a "request:changed-on-disk" event with the field "request_id" equal to "01JS0000000000000000000003" should have been emitted within 100 ms
    And a "request:changed-on-disk" event with the field "deleted" equal to "true" should have been emitted within 100 ms
    When the file watcher reports a change to the workspace path "clones/acme/api/xenon/collections/api/users.yaml"
    And the file watcher reports the removal of the workspace path "clones/acme/api/xenon/collections/api/ping.yaml"
    Then the event "workspace:tree-changed" should have been emitted 1 times
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "up_to_date"

  @pull
  Scenario: Pull with nothing new
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "status.state" equal to "up_to_date"
    And the event "workspace:tree-changed" should have been emitted 0 times

  @round_trip
  Scenario: Two Xenon clones exchange changes through the remote
    Given the workspace "Globex" exists
    And I call "SyncService.Link" with
      """
      { "local_id": "{{Globex.local_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/second", "mode": "adopt_repo" }
      """
    And the workspace file "clones/second/api/xenon/collections/api/users.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: PUT
      url: https://api.acme.test/users
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Globex.local_id}}", "message": "Users with PUT", "item_ids": [ "01JS0000000000000000000003" ] }
      """
    And I call "SyncService.Push" with
      """
      { "local_id": "{{Globex.local_id}}" }
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "method" equal to "PUT"
