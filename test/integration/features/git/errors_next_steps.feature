#language: en
#utf-8

@all @xen_007 @git @sync @errors_next_steps
Feature: Sync errors with a next step, push retries and one operation at a time
  Every remote failure carries a code and a next step in error_details (attribute next_step):
  offline GIT-10503, credentials refused GIT-11401, organisation approval or SSO required
  GIT-12403, a push rejected by branch protection, signing rules or hooks GIT-14403 (Xenon runs no
  hooks and signs no commits), any other rejection GIT-13409 with the host's message. Status shows
  the same error as its last error. A push the remote refuses because the branch moved pulls
  (replaying local commits) and retries once. Only one sync operation runs per workspace: another
  one meanwhile is refused with GIT-05409 and Status reports syncing.

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


  @offline
  Scenario: Offline - commits still work, pull and status report it with a next step
    Given the git http server goes offline
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: HEAD
      url: https://api.acme.test/ping
      """
    When I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Ping with HEAD", "item_ids": [ "01JS0000000000000000000002" ] }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "ahead"
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should fail with code "GIT-10503"
    And the error should contain the field "error.error_details.0.attribute" equal to "next_step"
    And the error should contain the field "error.error_details.0.messages.0" equal to "Offline - check the connection and try again"
    When I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should fail with code "GIT-10503"
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "error"
    And the response should contain the field "ahead" equal to "1"
    And the response should contain the field "last_error.code" equal to "GIT-10503"
    And the response should contain the field "last_error.next_step" equal to "Offline - check the connection and try again"

  @auth
  Scenario: Credentials the host no longer accepts
    Given the git http server requires the username "oauth2" and the token "rotated-token"
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should fail with code "GIT-11401"
    And the error should contain the field "error.error_details.0.attribute" equal to "next_step"
    And the error should contain the field "error.error_details.0.messages.0" equal to "Sign in again, or replace the token or SSH key of this account"
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "last_error.code" equal to "GIT-11401"

  @organisation
  Scenario Outline: Organisation restrictions - <case>
    Given the git http server answers the remote "acme/backend" with status 403 and the body "<body>"
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should fail with code "GIT-12403"
    And the error should contain the field "error.error_details.0.attribute" equal to "next_step"
    And the error should contain the field "error.error_details.0.messages.0" containing "<next>"

    Examples:
      | case                  | body                                                                                                       | next                          |
      | oauth app restriction | Although you appear to have the correct authorization credentials, the acme organization has enabled OAuth App access restrictions | approve Xenon                 |
      | saml sso              | Resource protected by organization SAML enforcement. You must grant your Personal Access token access to this organization.     | authorise single sign-on (SSO) |

  @rejected
  Scenario Outline: A push the remote rejects - <case>
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
    And the remote "acme/backend" rejects pushes with "<message>"
    When I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should fail with code "<code>"
    And the error should contain the field "error.error_details.0.attribute" equal to "next_step"
    And the error should contain the field "error.description" containing "<message>"
    And the remote "acme/backend" branch "main" should have 2 commits

    Examples:
      | case                | message                                       | code      |
      | protected branch    | protected branch hook declined                | GIT-14403 |
      | signed commits only | Commits must have verified signatures         | GIT-14403 |
      | something else      | refusing to update: repository quota exceeded | GIT-13409 |

  @retry
  Scenario: A push behind the remote pulls, replays and retries once
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
    And a teammate changes "api/xenon/collections/api/users.yaml" on the remote "acme/backend" to
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
    Then the call should succeed
    And the response should contain the field "state" equal to "up_to_date"
    And the remote "acme/backend" branch "main" should have 4 commits
    And the last commit on the remote "acme/backend" branch "main" should have the message containing "Sync: Ping with HEAD"
    And the file "api/xenon/collections/api/users.yaml" on the remote "acme/backend" should equal
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000003
      name: Users
      method: POST
      url: https://api.acme.test/users
      """
    And the workspace file "clones/acme/api/xenon/collections/api/users.yaml" should contain the field "method" equal to "POST"

  @retry @conflict
  Scenario: A push behind the remote whose replay conflicts stops with the conflict and pushes nothing
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
    And a teammate changes "api/xenon/collections/api/ping.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Ping
      method: OPTIONS
      url: https://api.acme.test/ping
      """
    When I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "conflict"
    And the remote "acme/backend" branch "main" should have 3 commits
    And the workspace file "clones/acme/api/xenon/collections/api/ping.yaml" should be unchanged

  @busy
  Scenario: A second operation while one runs is refused and never interleaves
    Given the git http server delays its responses by 400 ms
    When I start calling "SyncService.Pull" in the background with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And a "git:progress" event with the field "local_id" equal to "{{Acme.local_id}}" should have been emitted within 2000 ms
    And I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should fail with code "GIT-05409"
    When I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Anything", "item_ids": [ "01JS0000000000000000000002" ] }
      """
    Then the call should fail with code "GIT-05409"
    When I call "SyncService.Status" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "syncing"
    And the background call should finish within 10000 ms
    And the call should succeed
    And the response should contain the field "status.state" equal to "up_to_date"
