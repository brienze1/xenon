#language: en
#utf-8

@all @xen_007 @git @sync @conflicts @conflicts_structural
Feature: Resolve field conflicts in collection, folder and workspace settings files
  collection.yaml, folder.yaml and workspace.yaml merge field by field: order lists combine (theirs
  first, then my additions) and fields changed on one side only are taken from that side. Only a
  field changed differently on both sides is a conflict, resolved with Keep mine or Keep theirs for
  that field; the file is resolved once every field is.

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
      description: The API
      order:
        - ping.yaml
        - admin
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
    And the workspace file "acme/collections/api/admin/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: 01JS0000000000000000000007
      name: Admin
      description: Admin endpoints
      """
    And I save the workspace file "acme/workspace.yaml" field "id" as "workspace_id"
    And I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme" }
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Add the Acme workspace", "item_ids": [ "{{workspace_id}}", "01JS0000000000000000000001", "01JS0000000000000000000002", "01JS0000000000000000000007" ] }
      """
    And I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """


  @field
  Scenario Outline: A collection field changed differently on both sides - keep <resolution>
    Given the workspace file "clones/acme/api/xenon/collections/api/orders.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000005
      name: Orders
      method: GET
      url: https://api.acme.test/orders
      """
    And the workspace file "clones/acme/api/xenon/collections/api/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JS0000000000000000000001
      name: API
      description: Mine
      order:
        - ping.yaml
        - admin
        - orders.yaml
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Add orders", "item_ids": [ "01JS0000000000000000000001", "01JS0000000000000000000005" ] }
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
    And a teammate changes "api/xenon/collections/api/collection.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: collection
      id: 01JS0000000000000000000001
      name: API
      description: Theirs
      order:
        - ping.yaml
        - admin
        - items.yaml
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "status.state" equal to "conflict"
    And the response should contain the field "session.conflicts" array length equal to 1
    And the response should contain the field "session.conflicts.0.item_id" equal to "01JS0000000000000000000001"
    And the response should contain the field "session.conflicts.0.item_kind" equal to "collection"
    And the response should contain the field "session.conflicts.0.kind" equal to "field"
    And the response should contain the field "session.conflicts.0.fields" array length equal to 1
    And the response should contain the field "session.conflicts.0.fields.0.path" equal to "description"
    And the response should contain the field "session.conflicts.0.fields.0.mine" equal to "Mine"
    And the response should contain the field "session.conflicts.0.fields.0.theirs" equal to "Theirs"
    And the response should contain the field "session.conflicts.0.fields.0.resolution" equal to "unresolved"
    And the response should contain the field "session.conflicts.0.base.text" containing "description: The API"
    And the response should contain the field "session.conflicts.0.theirs.text" containing "description: Theirs"
    And the workspace file "clones/acme/api/xenon/collections/api/items.yaml" should not exist
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000001", "resolution": "<resolution>", "field_path": "description" }
      """
    Then the call should succeed
    And the response should contain the field "resolution" equal to "merged"
    And the response should contain the field "fields.0.resolution" equal to "<resolution>"
    When I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "ahead"
    And the workspace file "clones/acme/api/xenon/collections/api/collection.yaml" should contain the field "description" equal to "<description>"
    And the workspace file "clones/acme/api/xenon/collections/api/collection.yaml" should contain the field "order" with the items "ping.yaml, admin, items.yaml, orders.yaml"
    And the workspace file "clones/acme/api/xenon/collections/api/items.yaml" should exist
    And the workspace file "clones/acme/api/xenon/collections/api/orders.yaml" should exist
    When I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "up_to_date"
    And the file "api/xenon/collections/api/collection.yaml" on the remote "acme/backend" should equal
      """
      xenon: 1
      kind: collection
      id: 01JS0000000000000000000001
      name: API
      description: <description>
      order:
        - ping.yaml
        - admin
        - items.yaml
        - orders.yaml
      """

    Examples:
      | resolution | description |
      | mine       | Mine        |
      | theirs     | Theirs      |

  @field @any_order
  Scenario: Every field of a folder is resolved on its own before Finish
    Given the workspace file "clones/acme/api/xenon/collections/api/admin/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: 01JS0000000000000000000007
      name: Administration
      description: Mine
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Rename admin", "item_ids": [ "01JS0000000000000000000007" ] }
      """
    And a teammate changes "api/xenon/collections/api/admin/folder.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: folder
      id: 01JS0000000000000000000007
      name: Admin tools
      description: Theirs
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "session.conflicts.0.item_kind" equal to "folder"
    And the response should contain the field "session.conflicts.0.item_name" equal to "Administration"
    And the response should contain the field "session.conflicts.0.fields" array length equal to 2
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000007", "resolution": "theirs", "field_path": "description" }
      """
    Then the response should contain the field "resolution" equal to "unresolved"
    When I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should fail with code "GIT-04409"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000007", "resolution": "mine", "field_path": "name" }
      """
    Then the response should contain the field "resolution" equal to "merged"
    When I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "ahead"
    And the workspace file "clones/acme/api/xenon/collections/api/admin/folder.yaml" should contain the field "name" equal to "Administration"
    And the workspace file "clones/acme/api/xenon/collections/api/admin/folder.yaml" should contain the field "description" equal to "Theirs"

  @auto_merge
  Scenario: Different fields changed on each side merge without a conflict
    Given the workspace file "clones/acme/api/xenon/collections/api/admin/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: 01JS0000000000000000000007
      name: Admin
      description: Mine
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Describe admin", "item_ids": [ "01JS0000000000000000000007" ] }
      """
    And a teammate changes "api/xenon/collections/api/admin/folder.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: folder
      id: 01JS0000000000000000000007
      name: Admin tools
      description: Admin endpoints
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "status.state" equal to "ahead"
    And the response field "session" should not exist
    And the workspace file "clones/acme/api/xenon/collections/api/admin/folder.yaml" should contain the field "name" equal to "Admin tools"
    And the workspace file "clones/acme/api/xenon/collections/api/admin/folder.yaml" should contain the field "description" equal to "Mine"

  @validation
  Scenario Outline: A field resolution must name a conflicting field and keep one side - <case>
    Given the workspace file "clones/acme/api/xenon/collections/api/admin/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: 01JS0000000000000000000007
      name: Admin
      description: Mine
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Describe admin", "item_ids": [ "01JS0000000000000000000007" ] }
      """
    And a teammate changes "api/xenon/collections/api/admin/folder.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: folder
      id: 01JS0000000000000000000007
      name: Admin
      description: Theirs
      """
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000007", "resolution": "<resolution>", "field_path": "<field>" }
      """
    Then the call should fail with code "GIT-19400"

    Examples:
      | case                | resolution | field       |
      | no field            | mine       |             |
      | unknown field       | theirs     | name        |
      | keep both           | both       | description |
      | keep on a field     | keep       | description |

  @cancel
  Scenario: Cancel after a field conflict changes nothing
    Given the workspace file "clones/acme/api/xenon/collections/api/admin/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: 01JS0000000000000000000007
      name: Admin
      description: Mine
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Describe admin", "item_ids": [ "01JS0000000000000000000007" ] }
      """
    And a teammate changes "api/xenon/collections/api/admin/folder.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: folder
      id: 01JS0000000000000000000007
      name: Admin
      description: Theirs
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
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "session.id" as "session"
    When I call "SyncService.CancelPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "diverged"
    And the workspace file "clones/acme/api/xenon/collections/api/admin/folder.yaml" should be unchanged
    And the workspace file "clones/acme/api/xenon/collections/api/items.yaml" should not exist
