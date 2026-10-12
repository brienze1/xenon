#language: en
#utf-8

@all @xen_007 @git @sync @conflicts @conflicts_requests
Feature: Resolve request conflicts after a pull
  When the same request changed on both sides, Pull keeps a sync session open with the conflicting
  requests, each rendered as base, mine and theirs. Each one is resolved on its own, in any order:
  Keep mine, Keep theirs or Keep both (theirs saved next to mine as "<name> (theirs)" with a new id);
  a request deleted on one side and modified on the other is Keep or Delete. Finish commits the
  result on top of the remote (one replayed commit, or no commit at all when the result is the
  remote) and refuses while anything is unresolved (GIT-04409). Cancel leaves the workspace exactly
  as it was before the pull. Uncommitted local work survives Finish.

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
    And a teammate changes "api/xenon/collections/api/ping.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: OPTIONS
      url: https://api.acme.test/ping
      """


  @both_modified @render
  Scenario: The conflicting request is listed with base, mine and theirs rendered as requests
    Given a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
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
    And the response should contain the field "status.state" equal to "conflict"
    And the response field "session.id" should exist
    And the response should contain the field "session.conflicts" array length equal to 1
    And the response should contain the field "session.conflicts.0.item_id" equal to "01JS0000000000000000000003"
    And the response should contain the field "session.conflicts.0.item_name" equal to "Users"
    And the response should contain the field "session.conflicts.0.item_kind" equal to "request"
    And the response should contain the field "session.conflicts.0.kind" equal to "both_modified"
    And the response should contain the field "session.conflicts.0.resolution" equal to "unresolved"
    And the response should contain the field "session.conflicts.0.base.request.method" equal to "GET"
    And the response should contain the field "session.conflicts.0.mine.request.method" equal to "PATCH"
    And the response should contain the field "session.conflicts.0.theirs.request.method" equal to "POST"
    And the response should contain the field "session.conflicts.0.theirs.request.url" equal to "https://api.acme.test/users"
    And the response should contain the field "session.conflicts.0.mine.text" containing "method: PATCH"
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "conflict"
    And the response should contain the field "ahead" equal to "1"
    And the response should contain the field "behind" equal to "2"
    When I call "SyncService.Session" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "conflicts" array length equal to 1
    And the response should contain the field "conflicts.0.item_name" equal to "Users"

  @both_modified @keep_mine
  Scenario: Keep mine commits my version on top of the remote and keeps uncommitted work
    Given a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
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
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000003", "resolution": "mine" }
      """
    Then the call should succeed
    And the response should contain the field "item_id" equal to "01JS0000000000000000000003"
    And the response should contain the field "resolution" equal to "mine"
    When I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "ahead"
    And the response should contain the field "ahead" equal to "1"
    And the response should contain the field "behind" equal to "0"
    And the response should contain the field "local_changes" equal to "1"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "method" equal to "PATCH"
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should contain the field "method" equal to "OPTIONS"
    And the workspace file "clones/acme/api/xenon/collections/api/orders.yaml" should be unchanged
    And the remote "acme/backend" branch "main" should have 4 commits
    When I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the remote "acme/backend" branch "main" should have 5 commits
    And the last commit on the remote "acme/backend" branch "main" should have the message containing "Sync: Users with PATCH"
    And the last commit on the remote "acme/backend" branch "main" should have the author containing "Luis <luis@acme.test>"
    And the file "api/xenon/collections/api/users.yaml" on the remote "acme/backend" should equal
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: PATCH
      url: https://api.acme.test/users
      """
    And the file "api/xenon/collections/api/orders.yaml" on the remote "acme/backend" should not exist

  @both_modified @keep_theirs
  Scenario: Keep theirs takes the remote version and leaves nothing to push
    Given a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
      """
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000003", "resolution": "theirs" }
      """
    And I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "up_to_date"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "method" equal to "POST"
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should contain the field "method" equal to "OPTIONS"
    And the remote "acme/backend" branch "main" should have 4 commits

  @both_modified @keep_both
  Scenario: Keep both saves theirs as "<name> (theirs)" with a new id next to mine
    Given a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
      """
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000003", "resolution": "both" }
      """
    And I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "ahead"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "method" equal to "PATCH"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "id" equal to "01JS0000000000000000000003"
    And the workspace file "clones/acme/api/xenon/collections/api/users-(theirs).yaml" should contain the field "name" equal to "Users (theirs)"
    And the workspace file "clones/acme/api/xenon/collections/api/users-(theirs).yaml" should contain the field "method" equal to "POST"
    And the workspace file "clones/acme/api/xenon/collections/api/users-(theirs).yaml" should contain the field "id" not equal to "01JS0000000000000000000003"
    And the workspace file "clones/acme/api/xenon/collections/api/users-(theirs).yaml" should contain the field "id" matching "^[0-9A-Z]{26}$"
    And the workspace file "clones/acme/api/xenon/collections/api/collection.yaml" should contain the field "order" with the items "users.yaml, users-(theirs).yaml, ping.yaml"
    When I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "up_to_date"
    And the file "api/xenon/collections/api/users-(theirs).yaml" on the remote "acme/backend" should exist

  @delete_modify
  Scenario Outline: A request deleted on the remote and modified here - <resolution>
    Given a teammate deletes "api/xenon/collections/api/users.yaml" on the remote "acme/backend"
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "session.conflicts.0.kind" equal to "delete_modify"
    And the response should contain the field "session.conflicts.0.mine.request.method" equal to "PATCH"
    And the response field "session.conflicts.0.theirs" should not exist
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000003", "resolution": "<resolution>" }
      """
    And I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "<state>"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should <presence>
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should contain the field "method" equal to "OPTIONS"

    Examples:
      | resolution | state      | presence  |
      | keep       | ahead      | exist     |
      | delete     | up_to_date | not exist |

  @validation
  Scenario Outline: A resolution that does not fit the conflict is refused - <case>
    Given a teammate deletes "api/xenon/collections/api/users.yaml" on the remote "acme/backend"
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "<item>", "resolution": "<resolution>" }
      """
    Then the call should fail with code "<code>"

    Examples:
      | case                    | item                       | resolution | code      |
      | keep both on a deletion | 01JS0000000000000000000003 | both       | GIT-19400 |
      | keep mine on a deletion | 01JS0000000000000000000003 | mine       | GIT-19400 |
      | unknown resolution      | 01JS0000000000000000000003 | maybe      | GIT-29400 |
      | item without a conflict | 01JS0000000000000000000002 | keep       | GIT-35404 |

  @validation
  Scenario: Delete is not a choice for a request changed on both sides
    Given a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
      """
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000003", "resolution": "delete" }
      """
    Then the call should fail with code "GIT-19400"

  @unresolved @any_order
  Scenario: Finish is refused while a conflict is unresolved; conflicts resolve in any order
    Given a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
      """
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" contains
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
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "session.conflicts" array length equal to 2
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000003", "resolution": "theirs" }
      """
    And I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should fail with code "GIT-04409"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "method" equal to "PATCH"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000002", "resolution": "mine" }
      """
    And I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000003", "resolution": "mine" }
      """
    And I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "ahead"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "method" equal to "PATCH"
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should contain the field "method" equal to "HEAD"
    When I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "up_to_date"
    And the last commit on the remote "acme/backend" branch "main" should have the message containing "Sync: Users with PATCH (+1 more)"

  @cancel
  Scenario: Cancel leaves the workspace exactly as before the pull
    Given a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
      """
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "session.id" as "session"
    And I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000003", "resolution": "theirs" }
      """
    When I call "SyncService.CancelPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "diverged"
    And the response should contain the field "ahead" equal to "1"
    And the response should contain the field "behind" equal to "2"
    And the response should contain the field "local_changes" equal to "0"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should be unchanged
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should contain the field "method" equal to "GET"
    And the remote "acme/backend" branch "main" should have 4 commits
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "diverged"
    When I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should fail with code "GIT-16404"
    When I call "SyncService.Session" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response field "id" should not exist

  @stale
  Scenario: Finish is refused when the branch moved since the pull
    Given a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
      """
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "session.id" as "session"
    And I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000003", "resolution": "mine" }
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
    When I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should fail with code "GIT-28409"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "method" equal to "PATCH"
    And the workspace file "clones/acme/api/xenon/collections/api/orders.yaml" should be unchanged
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "diverged"
    And the response should contain the field "ahead" equal to "2"

  @uncommitted
  Scenario Outline: A conflict with uncommitted edits - <resolution>
    Given a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
      """
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "session.id" as "session"
    And I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000003", "resolution": "theirs" }
      """
    And I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: PUT
      url: https://api.acme.test/users?draft=1
      """
    And a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: DELETE
      url: https://api.acme.test/users
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "session.conflicts.0.mine.request.method" equal to "PUT"
    And the response should contain the field "session.conflicts.0.theirs.request.method" equal to "DELETE"
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000003", "resolution": "<resolution>" }
      """
    And I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "<state>"
    And the response should contain the field "behind" equal to "0"
    And the response should contain the field "local_changes" equal to "<local changes>"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "method" equal to "<method>"
    And the remote "acme/backend" branch "main" should have 5 commits

    Examples:
      | resolution | state         | local changes | method |
      | mine       | local_changes | 1             | PUT    |
      | theirs     | up_to_date    | 0             | DELETE |

  @uncommitted @keep_both
  Scenario: Keep both with uncommitted edits keeps mine and adds their copy as a local change
    Given the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: TRACE
      url: https://api.acme.test/ping
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "session.conflicts" array length equal to 1
    And the response should contain the field "session.conflicts.0.item_id" equal to "01JS0000000000000000000002"
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000002", "resolution": "both" }
      """
    And I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "ahead"
    And the response should contain the field "local_changes" equal to "2"
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should be unchanged
    And the workspace file "clones/acme/api/xenon/collections/api/ping-(theirs).yaml" should contain the field "method" equal to "OPTIONS"
    And the workspace file "clones/acme/api/xenon/collections/api/ping-(theirs).yaml" should contain the field "name" equal to "Ping (theirs)"

  @push
  Scenario: A push whose replay conflicts opens a session that resolves and pushes
    Given a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
      """
    When I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "conflict"
    When I call "SyncService.Session" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "conflicts" array length equal to 1
    And I save the response field "id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000003", "resolution": "mine" }
      """
    And I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    And I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "up_to_date"
    And the remote "acme/backend" branch "main" should have 5 commits
    And the file "api/xenon/collections/api/ping.yaml" on the remote "acme/backend" should exist

  @busy
  Scenario: Finish and Cancel wait their turn; a new pull replaces the open session
    Given a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
      """
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "session.id" as "session"
    And the git http server delays its responses by 400 ms
    And the emitted events are forgotten
    When I start calling "SyncService.Pull" in the background with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And a "git:progress" event with the field "phase" equal to "fetch" should have been emitted within 2000 ms
    And I call "SyncService.CancelPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should fail with code "GIT-05409"
    When I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should fail with code "GIT-05409"
    And the background call should finish within 10000 ms
    And the call should succeed
    And the response should contain the field "status.state" equal to "conflict"
    And the response should contain the field "session.id" not equal to "{{session}}"
    When I call "SyncService.CancelPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should fail with code "GIT-16404"
