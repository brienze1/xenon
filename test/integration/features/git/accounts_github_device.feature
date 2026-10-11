#language: en
#utf-8

@all @xen_007 @git @accounts @accounts_github_device
Feature: GitHub browser sign-in (device flow)
  "Add GitHub account" asks GitHub for a device code with Xenon's OAuth App client id (a build-time
  value; when it is empty sign-in is "not configured" and token accounts remain available), opens
  the verification page in the system browser and shows the user code. The UI polls; Go never asks
  GitHub sooner than the interval (raised by 5 s on slow_down) and stops on expired_token or
  access_denied. On success the account takes the GitHub login, name, avatar and primary email and
  the token goes to the keychain only.

  Background:
    Given the keychain is available
    And the system browser ignores the pages it opens
    And the "api" variable is set to the mocked api url
    And the "XENON_GITHUB_LOGIN_URL" env var is set to "{{api}}"
    And the "XENON_GITHUB_API_URL" env var is set to "{{api}}"
    And the "XENON_GITHUB_CLIENT_ID" env var is set to "Iv1.xenontest"
    And the current time is "2026-10-11T12:00:00Z"
    And the 1 "POST" request to "/login/device/code" returns status 200 with the following response
      """
      { "device_code": "dc-secret-device-code", "user_code": "WDJB-MJHT", "verification_uri": "https://github.com/login/device", "expires_in": 900, "interval": 5 }
      """

  @success
  Scenario: Sign in, honouring pending and slow_down
    Given the 1 "POST" request to "/login/oauth/access_token" returns status 200 with the following response
      """
      { "error": "authorization_pending" }
      """
    And the 2 "POST" request to "/login/oauth/access_token" returns status 200 with the following response
      """
      { "error": "slow_down", "interval": 10 }
      """
    And the 3 "POST" request to "/login/oauth/access_token" returns status 200 with the following response
      """
      { "access_token": "gho_devicetoken", "token_type": "bearer", "scope": "repo,read:user,user:email,write:public_key" }
      """
    And the 1 "GET" request to "/user" returns status 200 with the following response
      """
      { "login": "luis", "id": 7, "name": "Luis Brienze", "avatar_url": "https://avatars.example.test/7" }
      """
    And the 1 "GET" request to "/user/emails" returns status 200 with the following response
      """
      [ { "email": "old@acme.test", "primary": false, "verified": true }, { "email": "luis@acme.test", "primary": true, "verified": true } ]
      """
    When I call "GitAccountService.StartGitHubSignIn"
    Then the call should succeed
    And the response should contain the field "user_code" equal to "WDJB-MJHT"
    And the response should contain the field "verification_uri" equal to "https://github.com/login/device"
    And the response should contain the field "expires_in" equal to "900"
    And the response should contain the field "interval" equal to "5"
    And the response should not contain "dc-secret-device-code"
    And I save the response field "session_id" as "session"
    And the system browser should have opened 1 page
    And the 1 page opened in the system browser should start with "https://github.com/login/device"
    And the 1 "POST" request to "/login/device/code" should have the body matching "client_id=Iv1\.xenontest"
    And the 1 "POST" request to "/login/device/code" should have the body matching "scope=repo"
    When I call "GitAccountService.PollGitHubSignIn" with
      """
      { "session_id": "{{session}}" }
      """
    Then the response should contain the field "status" equal to "pending"
    And the "POST" request to "/login/oauth/access_token" should have been received 0 times
    Given the current time is "2026-10-11T12:00:06Z"
    When I call "GitAccountService.PollGitHubSignIn" with
      """
      { "session_id": "{{session}}" }
      """
    Then the response should contain the field "status" equal to "pending"
    And the 1 "POST" request to "/login/oauth/access_token" should have the body matching "device_code=dc-secret-device-code"
    And the 1 "POST" request to "/login/oauth/access_token" should have the body matching "grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code"
    Given the current time is "2026-10-11T12:00:12Z"
    When I call "GitAccountService.PollGitHubSignIn" with
      """
      { "session_id": "{{session}}" }
      """
    Then the response should contain the field "status" equal to "pending"
    And the response should contain the field "interval" equal to "10"
    Given the current time is "2026-10-11T12:00:18Z"
    When I call "GitAccountService.PollGitHubSignIn" with
      """
      { "session_id": "{{session}}" }
      """
    Then the response should contain the field "status" equal to "pending"
    And the "POST" request to "/login/oauth/access_token" should have been received 2 times
    Given the current time is "2026-10-11T12:00:23Z"
    When I call "GitAccountService.PollGitHubSignIn" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "done"
    And the response should contain the field "account.kind" equal to "github"
    And the response should contain the field "account.username" equal to "luis"
    And the response should contain the field "account.host" equal to "github.com"
    And the response should contain the field "account.commit_name" equal to "Luis Brienze"
    And the response should contain the field "account.commit_email" equal to "luis@acme.test"
    And the response should contain the field "account.avatar_url" equal to "https://avatars.example.test/7"
    And the response should not contain "gho_devicetoken"
    And I save the response field "account.id" as "account"
    And the 1 "GET" request to "/user" should have the header "Authorization" equal to "Bearer gho_devicetoken"
    And the 1 "GET" request to "/user/emails" should have the header "Authorization" equal to "Bearer gho_devicetoken"
    And the keychain should contain an entry for the account "{{account}}"
    And no file under the app data folder should contain "gho_devicetoken"
    When I call "GitAccountService.PollGitHubSignIn" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should fail with code "ACC-14404"

  @refused
  Scenario Outline: The sign-in ends when GitHub answers <error>
    Given the 1 "POST" request to "/login/oauth/access_token" returns status 200 with the following response
      """
      { "error": "<error>" }
      """
    And I call "GitAccountService.StartGitHubSignIn"
    And I save the response field "session_id" as "session"
    And the current time is "2026-10-11T12:00:06Z"
    When I call "GitAccountService.PollGitHubSignIn" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "<status>"
    And the response field "account" should not exist
    And the keychain should contain 0 entries

    Examples:
      | error         | status  |
      | expired_token | expired |
      | access_denied | denied  |

  @refused
  Scenario: A code past its lifetime expires without asking GitHub
    Given I call "GitAccountService.StartGitHubSignIn"
    And I save the response field "session_id" as "session"
    And the current time is "2026-10-11T12:15:00Z"
    When I call "GitAccountService.PollGitHubSignIn" with
      """
      { "session_id": "{{session}}" }
      """
    Then the response should contain the field "status" equal to "expired"
    And the "POST" request to "/login/oauth/access_token" should have been received 0 times

  @not_configured
  Scenario: Without a client id GitHub sign-in is not configured
    Given the "XENON_GITHUB_CLIENT_ID" env var is set to ""
    When I call "GitAccountService.StartGitHubSignIn"
    Then the call should fail with code "ACC-01503"
    And the "POST" request to "/login/device/code" should have been received 0 times
    And the system browser should have opened 0 pages

  @unknown_session
  Scenario: Polling an unknown session
    When I call "GitAccountService.PollGitHubSignIn" with
      """
      { "session_id": "nope" }
      """
    Then the call should fail with code "ACC-14404"
