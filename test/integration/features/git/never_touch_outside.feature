#language: en
#utf-8

@all @xen_007 @git @sync @never_touch_outside
Feature: Xenon never touches checkouts other than its own clones
  A workspace opened from a developer's own checkout and then linked is cloned again into Xenon's
  own clone; the developer's checkout - working files and its .git folder - stays byte-identical
  through a whole sync session (SC-004).

  Background:
    Given the keychain is available
    And the git http server requires the username "oauth2" and the token "tok-sync"
    And the "repo" variable is set to the http url of the remote "acme/backend"
    And the "host" variable is set to the git http server host
    And the remote "acme/backend" has the file "README.md" with
      """
      Backend service
      """
    And the remote "acme/backend" has the file "api/xenon/workspace.yaml" with
      """
      xenon: 1
      kind: workspace
      id: 01JS0000000000000000000100
      name: Backend API
      sync:
        repository: {{repo}}
        branch: main
        subfolder: api/xenon
      """
    And the remote "acme/backend" has the file "api/xenon/.gitignore" with
      """
      # Xenon — local-only files (never commit)
      .xenon/
      *.xenon-tmp
      .xenon-import-*/
      """
    And the remote "acme/backend" has the file "api/xenon/collections/api/collection.yaml" with
      """
      xenon: 1
      kind: collection
      id: 01JS0000000000000000000001
      name: API
      """
    And the remote "acme/backend" has the file "api/xenon/collections/api/ping.yaml" with
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: GET
      url: https://api.acme.test/ping
      """
    And I call "GitAccountService.AddToken" with
      """
      { "provider": "gitlab", "host": "{{host}}", "token": "tok-sync", "repository": "{{repo}}", "commit_name": "Luis", "commit_email": "luis@acme.test" }
      """
    And I save the response field "id" as "account"

  @sc004
  Scenario: A developer checkout of the same repository is byte-identical after a sync session
    Given the remote "acme/backend" is cloned to the temp path "workspaces/dev"
    And I call "WorkspaceService.Open" with
      """
      { "path": "{{workspace_root}}/dev/api/xenon" }
      """
    And I save the response field "local_id" as "dev_id"
    And the temp folder "workspaces/dev" is copied to "snapshots/dev"
    When I call "SyncService.Link" with
      """
      { "local_id": "{{dev_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme", "mode": "adopt_repo" }
      """
    Then the call should succeed
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
      { "local_id": "{{dev_id}}", "message": "Ping with HEAD", "item_ids": [ "01JS0000000000000000000002" ] }
      """
    And I call "SyncService.Push" with
      """
      { "local_id": "{{dev_id}}" }
      """
    And a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: GET
      url: https://api.acme.test/users
      """
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{dev_id}}" }
      """
    And I call "SyncService.Status" with
      """
      { "local_id": "{{dev_id}}" }
      """
    Then the response should contain the field "state" equal to "up_to_date"
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should exist
    And the folder "workspaces/dev" should equal the folder "snapshots/dev"
    And the remote "acme/backend" branch "main" should have 7 commits
