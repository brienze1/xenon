#language: en
#utf-8

@all @xen_007 @git @sync @link_and_clone
Feature: Link a workspace to a Git repository
  Linking clones the repository into Xenon's own clone (a location the user picks), checks out
  only the workspace's subfolder and moves the workspace there: an empty or missing subfolder
  receives a copy of the workspace files (first commit offered), a subfolder that already holds a
  Xenon workspace needs a choice - adopt the repository's workspace or replace it with mine. The
  old folder is left untouched and the registry points to the subfolder. workspace.yaml gets the
  shareable sync section (repository, branch, subfolder); the account and the clone location stay
  local. Repositories that use Large File Storage in the subfolder and missing branches are refused.
  "New workspace from Git repository" (CloneAsWorkspace) clones into a new Xenon clone and registers
  the subfolder as a new, active workspace: an existing workspace there is opened as it is (its
  secrets are missing - values never travel in Git), an empty subfolder becomes a new workspace
  (named as given, or after the repository) whose first commit is offered. Without a branch the
  remote's default branch is used.

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

  @not_linked
  Scenario: A workspace without a repository is not linked and works as before
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "not_linked"

  @empty_subfolder
  Scenario: Linking to a missing subfolder places the workspace files there and offers the first commit
    When I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "local_changes"
    And the response should contain the field "local_changes" equal to "3"
    And the response should contain the field "ahead" equal to "0"
    And the response should contain the field "behind" equal to "0"
    And the response should contain the field "settings.repository" equal to "{{repo}}"
    And the response should contain the field "settings.branch" equal to "main"
    And the response should contain the field "settings.subfolder" equal to "api/xenon"
    And the response should contain the field "settings.account_id" equal to "{{account}}"
    And the response should contain the path field "settings.clone_path" equal to "{{workspace_root}}/clones/acme"
    And the workspace file "clones/acme/api/xenon/workspace.yaml" should contain the field "name" equal to "Acme"
    And the workspace file "clones/acme/api/xenon/workspace.yaml" should contain the field "sync.repository" equal to "{{repo}}"
    And the workspace file "clones/acme/api/xenon/workspace.yaml" should contain the field "sync.branch" equal to "main"
    And the workspace file "clones/acme/api/xenon/workspace.yaml" should contain the field "sync.subfolder" equal to "api/xenon"
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should contain the field "name" equal to "Ping"
    And the workspace file "clones/acme/api/xenon/.gitignore" should exist
    And the workspace file "clones/acme/README.md" should not exist
    And the workspace file "acme/collections/api/ping.yaml" should exist
    And the workspace file "acme/workspace.yaml" should not contain the field "sync.repository"
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "sync.accountId" equal to "{{account}}"
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "sync.repository" equal to "{{repo}}"
    And no file under the app data folder should contain "tok-sync"
    And no file under the workspace folder should contain "tok-sync"
    And the file watcher should be watching the workspace path "clones/acme/api/xenon"
    When I call "WorkspaceService.List"
    Then the response should contain the path field "workspaces.0.path" equal to "{{workspace_root}}/clones/acme/api/xenon"
    When I call "SyncService.Changes" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 3
    And the response field "" should contain an item with "item_name" equal to "Acme" and "change" equal to "added"
    And the response field "" should contain an item with "item_name" equal to "API" and "item_kind" equal to "collection"
    And the response field "" should contain an item with "item_name" equal to "Ping" and "item_kind" equal to "request"
    When I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Add the Acme workspace", "item_ids": [ "{{workspace_id}}", "01JS0000000000000000000001", "01JS0000000000000000000002" ] }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "ahead"
    And the response should contain the field "ahead" equal to "1"
    When I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "up_to_date"
    And the remote "acme/backend" branch "main" should have 2 commits
    And the file "api/xenon/collections/api/ping.yaml" on the remote "acme/backend" should exist
    And the file "api/xenon/.gitignore" on the remote "acme/backend" should exist
    And the file "README.md" on the remote "acme/backend" should equal
      """
      Backend service
      """
    And the last commit on the remote "acme/backend" branch "main" should have the message containing "Add the Acme workspace"
    And the last commit on the remote "acme/backend" branch "main" should have the author containing "Luis <luis@acme.test>"
    And the remote "acme/backend" history should not contain "tok-sync"

  @existing_workspace
  Scenario: A subfolder that already holds a workspace needs a choice and nothing is left behind
    Given the remote "acme/backend" has the file "api/xenon/workspace.yaml" with
      """
      xenon: 1
      kind: workspace
      id: 01JS0000000000000000000100
      name: Backend API
      """
    When I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme" }
      """
    Then the call should fail with code "GIT-01409"
    And the workspace file "clones/acme" should not exist
    When I call "WorkspaceService.List"
    Then the response should contain the path field "workspaces.0.path" equal to "{{workspace_root}}/acme"
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "not_linked"

  @existing_workspace @adopt
  Scenario: Use the repository's workspace
    Given the remote "acme/backend" has the file "api/xenon/workspace.yaml" with
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
    And the remote "acme/backend" has the file "api/xenon/collections/shop/collection.yaml" with
      """
      xenon: 1
      kind: collection
      id: 01JS0000000000000000000101
      name: Shop
      """
    When I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme", "mode": "adopt_repo" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "up_to_date"
    And the workspace file "clones/acme/api/xenon/workspace.yaml" should contain the field "name" equal to "Backend API"
    And the workspace file "clones/acme/api/xenon/collections/shop/collection.yaml" should exist
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should not exist
    And the workspace file "acme/collections/api/ping.yaml" should exist
    When I call "WorkspaceService.List"
    Then the response should contain the field "workspaces.0.name" equal to "Backend API"
    And the response should contain the path field "workspaces.0.path" equal to "{{workspace_root}}/clones/acme/api/xenon"

  @existing_workspace @replace
  Scenario: Replace the repository's workspace with mine
    Given the remote "acme/backend" has the file "api/xenon/workspace.yaml" with
      """
      xenon: 1
      kind: workspace
      id: 01JS0000000000000000000100
      name: Backend API
      """
    And the remote "acme/backend" has the file "api/xenon/collections/shop/collection.yaml" with
      """
      xenon: 1
      kind: collection
      id: 01JS0000000000000000000101
      name: Shop
      """
    When I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme", "mode": "replace_with_mine" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "local_changes"
    And the workspace file "clones/acme/api/xenon/workspace.yaml" should contain the field "name" equal to "Acme"
    And the workspace file "clones/acme/api/xenon/collections/shop/collection.yaml" should not exist
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should exist
    When I call "SyncService.Changes" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response field "" should contain an item with "item_name" equal to "Shop" and "change" equal to "deleted"
    And the response field "" should contain an item with "item_name" equal to "Acme" and "change" equal to "modified"
    And the response field "" should contain an item with "item_name" equal to "Ping" and "change" equal to "added"

  @lfs
  Scenario: A subfolder tracked with Large File Storage is refused
    Given the remote "acme/backend" has the file ".gitattributes" with
      """
      api/xenon/** filter=lfs diff=lfs merge=lfs -text
      """
    When I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme" }
      """
    Then the call should fail with code "GIT-02415"
    And the error should contain the field "error.error_details.0.attribute" equal to "next_step"
    And the workspace file "clones/acme" should not exist
    When I call "WorkspaceService.List"
    Then the response should contain the path field "workspaces.0.path" equal to "{{workspace_root}}/acme"

  @branch_missing
  Scenario: A branch that does not exist on the remote is refused with the next step
    When I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{repo}}", "branch": "release", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme" }
      """
    Then the call should fail with code "GIT-03404"
    And the error should contain the field "error.error_details.0.attribute" equal to "next_step"
    And the workspace file "clones/acme" should not exist

  @root
  Scenario: The repository root as the workspace folder
    Given the remote "web/tools" exists
    And the "tools" variable is set to the http url of the remote "web/tools"
    When I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{tools}}", "branch": "main", "subfolder": "", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/tools" }
      """
    Then the call should succeed
    And the workspace file "clones/tools/workspace.yaml" should contain the field "sync.subfolder" equal to ""
    When I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "First", "item_ids": [ "{{workspace_id}}", "01JS0000000000000000000001", "01JS0000000000000000000002" ] }
      """
    And I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the remote "web/tools" branch "main" should have 1 commit
    And the file "collections/api/ping.yaml" on the remote "web/tools" should exist

  @validation
  Scenario Outline: Link validation - <case>
    Given the workspace folder "clones/busy" exists
    And the workspace file "clones/busy/notes.txt" contains
      """
      someone else's file
      """
    When I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "<repository>", "branch": "<branch>", "subfolder": "<subfolder>", "account_id": "<account>", "clone_path": "<clone>", "mode": "<mode>" }
      """
    Then the call should fail with code "<code>"

    Examples:
      | case                | repository | branch | subfolder | account                    | clone                              | mode      | code      |
      | no repository       |            | main   | api       | {{account}}                | {{workspace_root}}/clones/a        |           | GIT-63400 |
      | no branch           | {{repo}}   |        | api       | {{account}}                | {{workspace_root}}/clones/a        |           | GIT-63400 |
      | no account          | {{repo}}   | main   | api       |                            | {{workspace_root}}/clones/a        |           | GIT-63400 |
      | relative clone path | {{repo}}   | main   | api       | {{account}}                | clones/a                           |           | GIT-63400 |
      | unknown mode        | {{repo}}   | main   | api       | {{account}}                | {{workspace_root}}/clones/a        | merge     | GIT-63400 |
      | subfolder escapes   | {{repo}}   | main   | ../up     | {{account}}                | {{workspace_root}}/clones/a        |           | GIT-63400 |
      | clone path in use   | {{repo}}   | main   | api       | {{account}}                | {{workspace_root}}/clones/busy     |           | GIT-65409 |
      | inside workspace    | {{repo}}   | main   | api       | {{account}}                | {{workspace_root}}/acme/clone      |           | GIT-66409 |
      | unknown account     | {{repo}}   | main   | api       | 01JZ0000000000000000000000 | {{workspace_root}}/clones/a        |           | ACC-04404 |

  @clone @clone_existing
  Scenario: A repository whose subfolder holds a workspace is cloned as a new, active workspace with its secrets missing
    Given the remote "acme/backend" has the file "api/xenon/workspace.yaml" with
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
    And the remote "acme/backend" has the file "api/xenon/collections/shop/collection.yaml" with
      """
      xenon: 1
      kind: collection
      id: 01JS0000000000000000000101
      name: Shop
      """
    And the remote "acme/backend" has the file "api/xenon/environments/hml.yaml" with
      """
      xenon: 1
      kind: environment
      id: 01JS0000000000000000000102
      name: hml
      color: yellow
      confirmBeforeSend: false
      variables:
        - key: base_url
          value: https://hml.acme.test
          enabled: true
        - key: api_key
          secret: true
          enabled: true
      """
    When I call "SyncService.CloneAsWorkspace" with
      """
      { "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/backend" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Backend API"
    And the response should contain the path field "path" equal to "{{workspace_root}}/clones/backend/api/xenon"
    And the response should contain the field "missing" equal to "false"
    And I save the response field "local_id" as "backend"
    And the workspace file "clones/backend/api/xenon/collections/shop/collection.yaml" should exist
    And the workspace file "clones/backend/README.md" should not exist
    And the app data file "workspaces/{{backend}}/state.json" should contain the field "sync.accountId" equal to "{{account}}"
    And the app data file "workspaces/{{backend}}/state.json" should contain the field "sync.branch" equal to "main"
    And no file under the app data folder should contain "tok-sync"
    And no file under the workspace folder should contain "tok-sync"
    When I call "WorkspaceService.List"
    Then the response should contain the field "workspaces" array length equal to 2
    And the response should contain the field "active_local_id" equal to "{{backend}}"
    And the response should contain the path field "workspaces.0.path" equal to "{{workspace_root}}/acme"
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{backend}}" }
      """
    Then the call should succeed
    And the file watcher should be watching the workspace path "clones/backend/api/xenon"
    When I call "EnvironmentService.Get" with
      """
      { "local_id": "{{backend}}", "id": "01JS0000000000000000000102" }
      """
    Then the call should succeed
    And the response should contain the field "variables.0.current_value" equal to "https://hml.acme.test"
    And the response should contain the field "variables.1.key" equal to "api_key"
    And the response should contain the field "variables.1.secret" equal to "true"
    And the response should contain the field "variables.1.secret_state" equal to "missing"
    When I call "SyncService.Status" with
      """
      { "local_id": "{{backend}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "up_to_date"
    And the response should contain the field "settings.repository" equal to "{{repo}}"
    And the response should contain the field "settings.branch" equal to "main"
    And the response should contain the field "settings.subfolder" equal to "api/xenon"
    And the response should contain the field "settings.account_id" equal to "{{account}}"
    And the response should contain the path field "settings.clone_path" equal to "{{workspace_root}}/clones/backend"

  @clone @clone_empty
  Scenario: An empty subfolder becomes a new workspace with the shared sync settings and the first commit offered
    When I call "SyncService.CloneAsWorkspace" with
      """
      { "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/backend", "name": "Backend" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Backend"
    And the response should contain the path field "path" equal to "{{workspace_root}}/clones/backend/api/xenon"
    And I save the response field "local_id" as "backend"
    And I save the workspace file "clones/backend/api/xenon/workspace.yaml" field "id" as "backend_id"
    And the workspace file "clones/backend/api/xenon/workspace.yaml" should contain the field "name" equal to "Backend"
    And the workspace file "clones/backend/api/xenon/workspace.yaml" should contain the field "sync.repository" equal to "{{repo}}"
    And the workspace file "clones/backend/api/xenon/workspace.yaml" should contain the field "sync.branch" equal to "main"
    And the workspace file "clones/backend/api/xenon/workspace.yaml" should contain the field "sync.subfolder" equal to "api/xenon"
    And the workspace file "clones/backend/api/xenon/.gitignore" should exist
    And the workspace file "clones/backend/README.md" should not exist
    When I call "WorkspaceService.List"
    Then the response should contain the field "active_local_id" equal to "{{backend}}"
    When I call "SyncService.Status" with
      """
      { "local_id": "{{backend}}" }
      """
    Then the response should contain the field "state" equal to "local_changes"
    And the response should contain the field "local_changes" equal to "1"
    When I call "SyncService.Changes" with
      """
      { "local_id": "{{backend}}" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response field "" should contain an item with "item_name" equal to "Backend" and "change" equal to "added"
    When I call "SyncService.Commit" with
      """
      { "local_id": "{{backend}}", "message": "Add the Backend workspace", "item_ids": [ "{{backend_id}}" ] }
      """
    And I call "SyncService.Push" with
      """
      { "local_id": "{{backend}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "up_to_date"
    And the file "api/xenon/workspace.yaml" on the remote "acme/backend" should exist
    And the file "README.md" on the remote "acme/backend" should exist

  @clone @clone_default_branch
  Scenario: Without a branch the remote's default branch is cloned and the workspace is named after the repository
    Given the remote "web/tools" has the file "notes.md" with
      """
      Tools
      """
    And the default branch of the remote "web/tools" is renamed to "trunk"
    And the "tools" variable is set to the http url of the remote "web/tools"
    When I call "SyncService.CloneAsWorkspace" with
      """
      { "repository": "{{tools}}", "subfolder": "", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/tools" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "tools"
    And the response should contain the path field "path" equal to "{{workspace_root}}/clones/tools"
    And I save the response field "local_id" as "tools_ws"
    And the workspace file "clones/tools/workspace.yaml" should contain the field "sync.branch" equal to "trunk"
    And the workspace file "clones/tools/notes.md" should exist
    When I call "SyncService.Status" with
      """
      { "local_id": "{{tools_ws}}" }
      """
    Then the response should contain the field "settings.branch" equal to "trunk"
    And the response should contain the field "state" equal to "local_changes"

  @clone @clone_refused
  Scenario Outline: Cloning as a workspace is refused and leaves nothing behind - <case>
    Given the remote "acme/backend" has the file "<file>" with
      """
      <content>
      """
    When I call "SyncService.CloneAsWorkspace" with
      """
      { "repository": "{{repo}}", "branch": "<branch>", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/backend" }
      """
    Then the call should fail with code "<code>"
    And the error should contain the field "error.error_details.0.attribute" equal to "next_step"
    And the workspace file "clones/backend" should not exist
    When I call "WorkspaceService.List"
    Then the response should contain the field "workspaces" array length equal to 1
    And the response should contain the field "active_local_id" equal to "{{Acme.local_id}}"

    Examples:
      | case           | file           | content                                          | branch  | code      |
      | lfs            | .gitattributes | api/xenon/** filter=lfs diff=lfs merge=lfs -text | main    | GIT-02415 |
      | branch missing | notes.md       | notes                                            | release | GIT-03404 |

  @clone @validation
  Scenario Outline: Clone as workspace validation - <case>
    Given the workspace folder "clones/busy" exists
    And the workspace file "clones/busy/notes.txt" contains
      """
      someone else's file
      """
    When I call "SyncService.CloneAsWorkspace" with
      """
      { "repository": "<repository>", "branch": "main", "subfolder": "<subfolder>", "account_id": "<account>", "clone_path": "<clone>" }
      """
    Then the call should fail with code "<code>"

    Examples:
      | case                | repository | subfolder | account                    | clone                          | code      |
      | no repository       |            | api       | {{account}}                | {{workspace_root}}/clones/a    | GIT-39400 |
      | no account          | {{repo}}   | api       |                            | {{workspace_root}}/clones/a    | GIT-39400 |
      | relative clone path | {{repo}}   | api       | {{account}}                | clones/a                       | GIT-39400 |
      | subfolder escapes   | {{repo}}   | ../up     | {{account}}                | {{workspace_root}}/clones/a    | GIT-39400 |
      | clone path in use   | {{repo}}   | api       | {{account}}                | {{workspace_root}}/clones/busy | GIT-65409 |
      | inside workspace    | {{repo}}   | api       | {{account}}                | {{workspace_root}}/acme/clone  | GIT-66409 |
      | unknown account     | {{repo}}   | api       | 01JZ0000000000000000000000 | {{workspace_root}}/clones/a    | ACC-04404 |
