#language: en
#utf-8

@all @xen_007 @git @sync @uncommitted_preserved
Feature: Uncommitted local work survives a pull
  Uncommitted changes in the workspace are kept through a fast-forward and through a replay: the
  incoming files are written around them. When an incoming change touches a file with uncommitted
  local edits, that item is a conflict: Pull stops with the conflict status and changes nothing.

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


  @fast_forward
  Scenario: An uncommitted edit is kept through a fast-forward
    Given the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: HEAD
      url: https://api.acme.test/ping?draft=1
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
    And the response should contain the field "status.state" equal to "local_changes"
    And the response should contain the field "status.local_changes" equal to "1"
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should be unchanged
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "method" equal to "POST"

  @replay
  Scenario: An uncommitted edit is kept through a replay
    Given the workspace file "clones/acme/api/xenon/collections/api/orders.yaml" contains
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
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: HEAD
      url: https://api.acme.test/ping?draft=1
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
    And the response should contain the field "status.local_changes" equal to "1"
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should be unchanged
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "method" equal to "POST"
    When I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the file "api/xenon/collections/api/ping.yaml" on the remote "acme/backend" should equal
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: GET
      url: https://api.acme.test/ping
      """

  @conflict
  Scenario: An incoming change to a file with uncommitted edits is a conflict
    Given the workspace file "clones/acme/api/xenon/collections/api/users.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: PATCH
      url: https://api.acme.test/users?draft=1
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
    And the response should contain the field "session.conflicts.0.kind" equal to "both_modified"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should be unchanged
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should contain the field "method" equal to "GET"
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "behind"
    And the response should contain the field "behind" equal to "2"
    And the response should contain the field "local_changes" equal to "1"

  @conflict
  Scenario: An incoming deletion of a file with uncommitted edits is a conflict
    Given the workspace file "clones/acme/api/xenon/collections/api/users.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: PATCH
      url: https://api.acme.test/users?draft=1
      """
    And a teammate deletes "api/xenon/collections/api/users.yaml" on the remote "acme/backend"
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
    And the response should contain the field "session.conflicts.0.kind" equal to "delete_modify"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should be unchanged
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should contain the field "method" equal to "GET"
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "behind"
    And the response should contain the field "behind" equal to "2"
    And the response should contain the field "local_changes" equal to "1"
