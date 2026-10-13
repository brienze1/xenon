#language: en
#utf-8

@all @xen_008 @mcp @mcp_git
Feature: AI clients use git sync on a linked workspace
  "git" shows the sync status and pending changes, "git_edit" commits, pulls and resolves, "git_push" pushes and
  "delete" with target git_link unlinks. Accounts are human-only: the workspace uses the account the human linked it
  with. Every change emits git:status-changed so the footer follows.

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
      id: 01JS8000000000000000000001
      name: API
      """
    And the workspace file "acme/collections/api/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS8000000000000000000002
      name: Ping
      method: GET
      url: https://api.acme.test/ping
      """
    And I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme" }
      """

  @status_commit_push
  Scenario: The AI sees the changes, commits them and pushes
    When I call the MCP tool "git" with
      """
      { "action": "status" }
      """
    Then the call should succeed
    And the response should contain the field "linked" equal to "true"
    And the response should contain the field "state" equal to "local_changes"
    And the response should contain the field "branch" equal to "main"
    And the response should not contain "tok-sync"
    When I call the MCP tool "git" with
      """
      { "action": "changes" }
      """
    Then the call should succeed
    And the response field "items" should contain an item with "name" equal to "Ping" and "change" equal to "added"
    Given the emitted events are forgotten
    When I call the MCP tool "git_edit" with
      """
      { "action": "commit", "message": "Add the API collection" }
      """
    Then the call should succeed
    And a "git:status-changed" event with the field "local_id" equal to "{{Acme.local_id}}" should have been emitted within 1000 ms
    When I call the MCP tool "git_push"
    Then the call should succeed
    And the remote "acme/backend" branch "main" should have 2 commits
    And the last commit on the remote "acme/backend" branch "main" should have the message containing "Add the API collection"
    And the last commit on the remote "acme/backend" branch "main" should have the author containing "Luis"
    When I call the MCP tool "git" with
      """
      { "action": "status" }
      """
    Then the response should contain the field "state" equal to "up_to_date"

  @pull
  Scenario: Pulling brings a teammate's change into the workspace and refreshes the tree
    Given I call the MCP tool "git_edit" with
      """
      { "action": "commit", "message": "Add the API collection" }
      """
    And I call the MCP tool "git_push"
    And a teammate changes "api/xenon/collections/api/ping.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS8000000000000000000002
      name: Ping
      method: HEAD
      url: https://api.acme.test/ping
      """
    And the emitted events are forgotten
    When I call the MCP tool "git_edit" with
      """
      { "action": "pull" }
      """
    Then the call should succeed
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should contain the field "method" equal to "HEAD"
    And a "workspace:tree-changed" event with the field "local_id" equal to "{{Acme.local_id}}" should have been emitted within 1000 ms
    And a "git:status-changed" event with the field "local_id" equal to "{{Acme.local_id}}" should have been emitted within 1000 ms

  @unlink
  Scenario: Unlinking goes through the destructive delete tool
    When I call the MCP tool "delete" with
      """
      { "target": "git_link" }
      """
    Then the call should succeed
    When I call the MCP tool "git" with
      """
      { "action": "status" }
      """
    Then the response should contain the field "linked" equal to "false"

  @accounts_human_only
  Scenario: Git accounts cannot be read or managed through MCP
    When I list the MCP tools
    Then the response should not contain "AddToken"
    And the response should not contain "git_accounts"
    When I call the MCP tool "git_edit" with
      """
      { "action": "add_account", "token": "tok-new" }
      """
    Then the call should fail with code "MCP-07400"
