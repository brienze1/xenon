#language: en
#utf-8

@all @xen_007 @git @sync @unlink
Feature: Unlink a workspace from its Git repository
  "Unlink repository" moves away from Git without losing anything: the workspace files as they
  are now (committed or not, local-only files excluded) are copied to a folder the user picks, the
  workspace is re-pointed there, the local sync settings are cleared and Xenon's own clone is
  deleted (the UI confirms first). The remote is not touched, and neither is the folder the
  workspace lived in before it was linked. The target must be a new or empty folder outside every
  workspace and outside the clone.

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
    And I save the workspace file "acme/workspace.yaml" field "id" as "workspace_id"
    And I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme" }
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Add the Acme workspace", "item_ids": [ "{{workspace_id}}", "01JS0000000000000000000001", "01JS0000000000000000000002" ] }
      """
    And I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """

  @success
  Scenario: Unlinking copies the current files to the chosen folder, re-points the workspace and deletes the clone
    Given the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: HEAD
      url: https://api.acme.test/ping
      """
    When I call "SyncService.Unlink" with
      """
      { "local_id": "{{Acme.local_id}}", "target_folder": "{{workspace_root}}/acme-plain" }
      """
    Then the call should succeed
    And the response should contain the field "local_id" equal to "{{Acme.local_id}}"
    And the response should contain the field "name" equal to "Acme"
    And the response should contain the path field "path" equal to "{{workspace_root}}/acme-plain"
    And the workspace file "acme-plain/workspace.yaml" should contain the field "name" equal to "Acme"
    And the workspace file "acme-plain/collections/api/ping.yaml" should contain the field "method" equal to "HEAD"
    And the workspace file "acme-plain/collections/api/collection.yaml" should exist
    And the workspace file "acme-plain/.gitignore" should exist
    And the workspace file "acme-plain/.git" should not exist
    And the workspace file "acme-plain/README.md" should not exist
    And the workspace file "clones/acme" should not exist
    And the workspace file "acme/collections/api/ping.yaml" should contain the field "method" equal to "GET"
    And the remote "acme/backend" branch "main" should have 2 commits
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should not contain "{{repo}}"
    And the file watcher should be watching the workspace path "acme-plain"
    When I call "WorkspaceService.List"
    Then the response should contain the path field "workspaces.0.path" equal to "{{workspace_root}}/acme-plain"
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "not_linked"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JS0000000000000000000002" }
      """
    Then the call should succeed
    And the response should contain the field "method" equal to "HEAD"

  @conflict
  Scenario: Unlinking drops a pull waiting for conflict resolution
    Given the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: PATCH
      url: https://api.acme.test/ping
      """
    And a teammate changes "api/xenon/collections/api/ping.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: POST
      url: https://api.acme.test/ping
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "status.state" equal to "conflict"
    When I call "SyncService.Unlink" with
      """
      { "local_id": "{{Acme.local_id}}", "target_folder": "{{workspace_root}}/acme-plain" }
      """
    Then the call should succeed
    And the workspace file "acme-plain/collections/api/ping.yaml" should contain the field "method" equal to "PATCH"
    When I call "SyncService.Session" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response field "id" should not exist

  @validation
  Scenario Outline: Unlinking is refused and nothing changes - <case>
    Given the workspace folder "busy" exists
    And the workspace file "busy/notes.txt" contains
      """
      someone else's file
      """
    And the workspace "Globex" exists
    When I call "SyncService.Unlink" with
      """
      { "local_id": "<local_id>", "target_folder": "<target>" }
      """
    Then the call should fail with code "<code>"
    And the workspace file "clones/acme/api/xenon/workspace.yaml" should exist
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "up_to_date"

    Examples:
      | case                  | local_id           | target                                  | code      |
      | no target             | {{Acme.local_id}}  |                                         | GIT-40400 |
      | relative target       | {{Acme.local_id}}  | acme-plain                              | GIT-40400 |
      | target not empty      | {{Acme.local_id}}  | {{workspace_root}}/busy                 | GIT-43409 |
      | target inside clone   | {{Acme.local_id}}  | {{workspace_root}}/clones/acme/plain    | GIT-44409 |
      | target in a workspace | {{Acme.local_id}}  | {{workspace_root}}/globex/plain         | GIT-44409 |
      | not linked            | {{Globex.local_id}} | {{workspace_root}}/globex-plain        | GIT-09409 |
