#language: en
#utf-8

@all @xen_007 @git @sync @sync_performance
Feature: Sync performance
  A typical change - 1 to 10 requests - is committed and pushed in under 5 seconds, and a
  conflict-free pull of 10 changed requests finishes in under 5 seconds, against the local
  smart-HTTP Git server (SC-001).

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
    And the workspace folder "perf" holds 10 generated requests across 2 collections
    And I call "WorkspaceService.Open" with
      """
      { "path": "{{workspace_root}}/perf" }
      """
    And I save the response field "local_id" as "perf"
    And I call "SyncService.Link" with
      """
      { "local_id": "{{perf}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/perf" }
      """
    And I call "SyncService.Commit" with
      """
      {
        "local_id": "{{perf}}", "message": "Add the perf workspace",
        "item_ids": [
          "01JG0000000000000000000000", "01JG0000000000000000000001", "01JG0000000000000000000002",
          "01JG0000000000000000001000", "01JG0000000000000000001001", "01JG0000000000000000001002", "01JG0000000000000000001003", "01JG0000000000000000001004",
          "01JG0000000000000000001005", "01JG0000000000000000001006", "01JG0000000000000000001007", "01JG0000000000000000001008", "01JG0000000000000000001009"
        ]
      }
      """
    And I call "SyncService.Push" with
      """
      { "local_id": "{{perf}}" }
      """

  @commit_push
  Scenario: Committing and pushing 10 changed requests takes under 5 seconds
    Given the workspace folder "clones/perf/api/xenon" holds 10 generated requests across 2 collections
    When I call "SyncService.Commit" with
      """
      {
        "local_id": "{{perf}}", "message": "Point the requests at the new host",
        "item_ids": [
          "01JG0000000000000000001000", "01JG0000000000000000001001", "01JG0000000000000000001002", "01JG0000000000000000001003", "01JG0000000000000000001004",
          "01JG0000000000000000001005", "01JG0000000000000000001006", "01JG0000000000000000001007", "01JG0000000000000000001008", "01JG0000000000000000001009"
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "ahead" equal to "1"
    And the last call should have taken less than 2000 ms
    When I call "SyncService.Push" with
      """
      { "local_id": "{{perf}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "local_changes"
    And the last call should have taken less than 3000 ms
    And the remote "acme/backend" branch "main" should have 3 commits
    And the last commit on the remote "acme/backend" branch "main" should have the message containing "Point the requests at the new host"

  @pull
  Scenario: A conflict-free pull of 10 changed requests takes under 5 seconds
    Given the workspace folder "teammate" holds 10 generated requests across 2 collections
    And a teammate commits the workspace folder "teammate" as "api/xenon" on the remote "acme/backend"
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{perf}}" }
      """
    Then the call should succeed
    And the response should contain the field "status.state" equal to "up_to_date"
    And the last call should have taken less than 5000 ms
    And the workspace file "clones/perf/api/xenon/collections/collection-1/request-0000.yaml" should contain the field "url" equal to "https://teammate.test/items/0"
    And the workspace file "clones/perf/api/xenon/collections/collection-2/request-0009.yaml" should contain the field "url" equal to "https://teammate.test/items/9"
