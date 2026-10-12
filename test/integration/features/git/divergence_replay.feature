#language: en
#utf-8

@all @xen_007 @git @sync @divergence_replay
Feature: Pull when both sides have commits
  When the branch has local commits and the remote moved, Pull replays: the unpushed local commits
  become one commit on top of the remote (linear history, no merge commits) whose subject starts
  with "Sync:" and whose body keeps every original message. Requests changed on one side only are
  taken from that side; collection, folder and workspace files changed on both sides merge field by
  field (order lists: theirs first, then my additions). The same request changed differently on
  both sides is a conflict: Pull stops with the conflict status, changes nothing and keeps a sync
  session open until the conflicts are resolved (conflicts_*.feature).

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


  @replay
  Scenario: Non-overlapping changes become one commit on top of the remote, pushed after
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
    And the workspace file "clones/acme/api/xenon/collections/api/orders.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000005
      name: Orders
      method: GET
      url: https://api.acme.test/orders
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Add orders", "item_ids": [ "01JS0000000000000000000005" ] }
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
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "status.state" equal to "ahead"
    And the response should contain the field "status.ahead" equal to "1"
    And the response should contain the field "status.behind" equal to "0"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "method" equal to "POST"
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should contain the field "method" equal to "HEAD"
    And the workspace file "clones/acme/api/xenon/collections/api/orders.yaml" should exist
    And the remote "acme/backend" branch "main" should have 3 commits
    When I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "up_to_date"
    And the remote "acme/backend" branch "main" should have 4 commits
    And the last commit on the remote "acme/backend" branch "main" should have the message containing "Sync: Ping with HEAD (+1 more)"
    And the last commit on the remote "acme/backend" branch "main" should have the message containing "- Ping with HEAD"
    And the last commit on the remote "acme/backend" branch "main" should have the message containing "- Add orders"
    And the last commit on the remote "acme/backend" branch "main" should have the author containing "Luis <luis@acme.test>"
    And the file "api/xenon/collections/api/users.yaml" on the remote "acme/backend" should equal
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
      """
    And the file "api/xenon/collections/api/ping.yaml" on the remote "acme/backend" should equal
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: HEAD
      url: https://api.acme.test/ping
      """
    And the file "api/xenon/collections/api/orders.yaml" on the remote "acme/backend" should exist

  @replay @structural
  Scenario: Both sides added a request to the same collection - the order lists merge
    Given the workspace file "clones/acme/api/xenon/collections/api/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JS0000000000000000000001
      name: API
      order:
        - ping.yaml
        - users.yaml
        - orders.yaml
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
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Add orders", "item_ids": [ "01JS0000000000000000000001", "01JS0000000000000000000005" ] }
      """
    And a teammate changes "api/xenon/collections/api/collection.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: collection
      id: 01JS0000000000000000000001
      name: API
      description: The public API
      order:
        - ping.yaml
        - users.yaml
        - items.yaml
      """
    And a teammate changes "api/xenon/collections/api/items.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000006
      name: Items
      method: GET
      url: https://api.acme.test/items
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "status.state" equal to "ahead"
    And the workspace file "clones/acme/api/xenon/collections/api/collection.yaml" should contain the field "order" with the items "ping.yaml, users.yaml, items.yaml, orders.yaml"
    And the workspace file "clones/acme/api/xenon/collections/api/collection.yaml" should contain the field "description" equal to "The public API"
    And the workspace file "clones/acme/api/xenon/collections/api/items.yaml" should exist
    And the workspace file "clones/acme/api/xenon/collections/api/orders.yaml" should exist
    When I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "up_to_date"

  @conflict
  Scenario: The same request changed on both sides stops the pull without changing anything
    Given the workspace file "clones/acme/api/xenon/collections/api/users.yaml" contains
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
    And a teammate changes "api/xenon/collections/api/ping.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: OPTIONS
      url: https://api.acme.test/ping
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "status.state" equal to "conflict"
    And the response should contain the field "session.conflicts" array length equal to 1
    And the response should contain the field "session.conflicts.0.item_id" equal to "01JS0000000000000000000003"
    And the response should contain the field "session.conflicts.0.item_name" equal to "Users"
    And the response should contain the field "session.conflicts.0.kind" equal to "both_modified"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should be unchanged
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should contain the field "method" equal to "GET"
    And the remote "acme/backend" branch "main" should have 4 commits
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "conflict"
    And the response should contain the field "ahead" equal to "1"
    And the response should contain the field "behind" equal to "2"
