#language: en
#utf-8

@all @xen_007 @git @accounts @accounts_token_hosts
Feature: Git token accounts
  A token account works with any host. Xenon verifies the token on save with a read of a repository
  on that host (or, for GitHub, the user API), using each host's BasicAuth username convention:
  GitHub the login (or x-access-token), GitLab "oauth2", Bitbucket API tokens
  "x-bitbucket-api-token-auth", Azure DevOps an empty username, other hosts the given username.
  The token lives only in the OS keychain (service dev.brienze.xenon.git, account = the account id),
  never in app.json, workspace files or remote URLs. Removing an account deletes its keychain entry
  and unlinks it from every workspace that used it (FR-016).

  Background:
    Given the keychain is available
    And the remote "acme/backend" has the file "api/xenon/workspace.yaml" with
      """
      name: Acme
      """
    And the "repo" variable is set to the http url of the remote "acme/backend"
    And the "host" variable is set to the git http server host

  @conventions
  Scenario Outline: A <provider> token is verified with the host's username convention and kept in the keychain only
    Given the git http server requires the username "<username>" and the token "tok-<provider>-secret"
    When I call "GitAccountService.AddToken" with
      """
      { "provider": "<provider>", "host": "{{host}}", "username": "luis", "token": "tok-<provider>-secret", "repository": "{{repo}}", "commit_name": "Luis", "commit_email": "luis@acme.test" }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "token"
    And the response should contain the field "provider" equal to "<provider>"
    And the response should contain the field "host" equal to "{{host}}"
    And the response should contain the field "commit_name" equal to "Luis"
    And the response should contain the field "commit_email" equal to "luis@acme.test"
    And the response should not contain "tok-<provider>-secret"
    And I save the response field "id" as "account"
    And the git http server should have received 1 request with the username "<username>"
    And the keychain should contain an entry for the account "{{account}}"
    And the keychain should contain 1 entries
    And the app data file "app.json" should contain the field "gitAccounts.0.id" equal to "{{account}}"
    And no file under the app data folder should contain "tok-<provider>-secret"

    Examples:
      | provider     | username                   |
      | github       | luis                       |
      | gitlab       | oauth2                     |
      | bitbucket    | x-bitbucket-api-token-auth |
      | azure_devops |                            |
      | other        | luis                       |

  @invalid
  Scenario: An invalid token is refused and nothing is stored
    Given the git http server requires the username "oauth2" and the token "the-right-one"
    When I call "GitAccountService.AddToken" with
      """
      { "provider": "gitlab", "host": "{{host}}", "token": "a-wrong-token", "repository": "{{repo}}", "commit_name": "Luis", "commit_email": "luis@acme.test" }
      """
    Then the call should fail with code "ACC-02401"
    And the keychain should contain 0 entries
    And the app data file "app.json" should not contain "gitAccounts"

  @invalid
  Scenario: A token for another host than the repository is refused before it is sent
    When I call "GitAccountService.AddToken" with
      """
      { "provider": "gitlab", "host": "gitlab.acme.test", "token": "never-sent-token", "repository": "{{repo}}", "commit_name": "Luis", "commit_email": "luis@acme.test" }
      """
    Then the call should fail with code "ACC-18400"
    And the git http server should have received 0 requests with the username "oauth2"
    And the keychain should contain 0 entries

  @invalid
  Scenario: A token for a host Xenon cannot verify on its own needs a repository
    When I call "GitAccountService.AddToken" with
      """
      { "provider": "other", "host": "git.acme.test", "username": "deploy", "token": "some-token", "commit_name": "Luis", "commit_email": "luis@acme.test" }
      """
    Then the call should fail with code "ACC-03400"

  @validation
  Scenario Outline: Token form validation - <case>
    When I call "GitAccountService.AddToken" with
      """
      { "provider": "<provider>", "host": "<host>", "token": "<token>", "repository": "{{repo}}", "commit_name": "Luis", "commit_email": "<email>" }
      """
    Then the call should fail with code "ACC-20400"

    Examples:
      | case          | provider | host     | token | email          |
      | no host       | gitlab   |          | t     | luis@acme.test |
      | no token      | gitlab   | {{host}} |       | luis@acme.test |
      | bad provider  | svn      | {{host}} | t     | luis@acme.test |
      | invalid email | gitlab   | {{host}} | t     | not-an-email   |

  @github_api
  Scenario: A GitHub token without a repository is verified with the user API and takes the login
    Given the "api" variable is set to the mocked api url
    And the "XENON_GITHUB_API_URL" env var is set to "{{api}}"
    And the 1 "GET" request to "/user" returns status 200 with the following response
      """
      { "login": "octo", "id": 42, "name": "Octo Cat", "avatar_url": "https://avatars.example.test/42" }
      """
    When I call "GitAccountService.AddToken" with
      """
      { "provider": "github", "host": "github.com", "token": "ghp-test-token", "commit_name": "", "commit_email": "" }
      """
    Then the call should succeed
    And the response should contain the field "username" equal to "octo"
    And the response should contain the field "display_name" equal to "octo @ github.com"
    And the response should contain the field "commit_name" equal to "Octo Cat"
    And the response should contain the field "commit_email" equal to "42+octo@users.noreply.github.com"
    And the response should contain the field "avatar_url" equal to "https://avatars.example.test/42"
    And the 1 "GET" request to "/user" should have the header "Authorization" equal to "Bearer ghp-test-token"

  @github_api @invalid
  Scenario: A GitHub token the user API refuses is invalid
    Given the "api" variable is set to the mocked api url
    And the "XENON_GITHUB_API_URL" env var is set to "{{api}}"
    And the 1 "GET" request to "/user" returns status 401 with the following response
      """
      { "message": "Bad credentials" }
      """
    When I call "GitAccountService.AddToken" with
      """
      { "provider": "github", "host": "github.com", "token": "ghp-bad", "commit_name": "", "commit_email": "" }
      """
    Then the call should fail with code "ACC-17401"
    And the keychain should contain 0 entries

  @keychain
  Scenario: Without a keychain the token is kept encrypted with the vault passphrase
    Given the keychain is unavailable
    And the git http server requires the username "oauth2" and the token "tok-passphrase-mode"
    And I call "VaultService.SetupPassphrase" with
      """
      { "passphrase": "correct horse", "confirmation": "correct horse" }
      """
    When I call "GitAccountService.AddToken" with
      """
      { "provider": "gitlab", "host": "{{host}}", "token": "tok-passphrase-mode", "repository": "{{repo}}", "commit_name": "Luis", "commit_email": "luis@acme.test" }
      """
    Then the call should succeed
    And I save the response field "id" as "account"
    And the app data file "git/credentials/{{account}}.age" should exist
    And no file under the app data folder should contain "tok-passphrase-mode"
    And the keychain should contain 0 entries
    When I call "GitAccountService.TestConnection" with
      """
      { "account_id": "{{account}}", "repository": "{{repo}}" }
      """
    Then the call should succeed
    And the response should contain the field "default_branch" equal to "main"

  @manage
  Scenario: Accounts are listed without secrets, edited and removed; workspaces using a removed account are unlinked from it
    Given the git http server requires the username "oauth2" and the token "tok-manage"
    And the workspace "Acme" exists
    And I call "GitAccountService.AddToken" with
      """
      { "provider": "gitlab", "host": "{{host}}", "token": "tok-manage", "repository": "{{repo}}", "commit_name": "Luis", "commit_email": "luis@acme.test" }
      """
    And I save the response field "id" as "account"
    And the app data file "workspaces/{{Acme.local_id}}/state.json" contains
      """
      { "sync": { "accountId": "{{account}}", "clonePath": "/clones/acme", "autoPull": true, "autoPush": false } }
      """
    When I call "GitAccountService.Update" with
      """
      { "id": "{{account}}", "display_name": "Work GitLab", "commit_name": "Luis B", "commit_email": "lb@acme.test" }
      """
    Then the call should succeed
    When I call "GitAccountService.List"
    Then the response should contain the field "0.display_name" equal to "Work GitLab"
    And the response should contain the field "0.commit_name" equal to "Luis B"
    And the response should contain the field "0.commit_email" equal to "lb@acme.test"
    And the response should not contain "tok-manage"
    When I call "GitAccountService.Remove" with
      """
      { "id": "{{account}}" }
      """
    Then the call should succeed
    And the keychain should contain 0 entries
    And the app data file "app.json" should contain the field "gitAccounts" array length equal to 0
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "sync.accountId" equal to ""
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "sync.clonePath" equal to "/clones/acme"
    When I call "GitAccountService.List"
    Then the response should contain the field "" array length equal to 0

  @keychain
  Scenario: Without a keychain and before a passphrase is chosen the token cannot be stored
    Given the keychain is unavailable
    And the git http server requires the username "oauth2" and the token "tok-locked"
    When I call "GitAccountService.AddToken" with
      """
      { "provider": "gitlab", "host": "{{host}}", "token": "tok-locked", "repository": "{{repo}}", "commit_name": "Luis", "commit_email": "luis@acme.test" }
      """
    Then the call should fail with code "ACC-16409"
    And the app data file "app.json" should not contain "gitAccounts"

  @manage
  Scenario: Unknown accounts
    When I call "GitAccountService.Remove" with
      """
      { "id": "01JZ0000000000000000000000" }
      """
    Then the call should fail with code "ACC-04404"
