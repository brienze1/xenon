#language: en
#utf-8

@all @xen_007 @git @accounts @host_keys
Feature: SSH host keys
  Xenon keeps its own known_hosts in app data (ssh/known_hosts), pre-seeded with the published keys
  of github.com, gitlab.com and bitbucket.org. The first connection to any other host stops with a
  "git:host-key-unknown" event carrying the fingerprint; once the user confirms that fingerprint the
  key is remembered. A host whose key changed is blocked with a warning and nothing is remembered.

  Background:
    Given the keychain is available
    And the remote "acme/backend" has the file "api/xenon/workspace.yaml" with
      """
      name: Acme
      """
    And the "repo" variable is set to the ssh url of the remote "acme/backend"
    And the "ssh_host" variable is set to the git ssh server host
    And the "fingerprint" variable is set to the git ssh server host key fingerprint
    And an ssh key "agent" is saved at the temp path "keys/agent"
    And the ssh agent holds the ssh key "agent"
    And the git ssh server authorizes the ssh key "agent"
    And I call "GitAccountService.AddSSH" with
      """
      { "mode": "agent", "host": "{{ssh_host}}", "commit_name": "Luis", "commit_email": "luis@acme.test" }
      """
    And I save the response field "account.id" as "account"

  @first_use
  Scenario: An unknown host asks for confirmation, then is remembered
    When I call "GitAccountService.TestConnection" with
      """
      { "account_id": "{{account}}", "repository": "{{repo}}" }
      """
    Then the call should fail with code "ACC-05409"
    And the event "git:host-key-unknown" should have been emitted 1 times
    And the 1 "git:host-key-unknown" event should contain the field "host" equal to "{{ssh_host}}"
    And the 1 "git:host-key-unknown" event should contain the field "fingerprint" equal to "{{fingerprint}}"
    And the 1 "git:host-key-unknown" event should contain the field "key_type" equal to "ssh-ed25519"
    And the 1 "git:host-key-unknown" event should contain the field "repository" equal to "{{repo}}"
    When I call "GitAccountService.ConfirmHostKey" with
      """
      { "repository": "{{repo}}", "fingerprint": "{{fingerprint}}", "accept": true }
      """
    Then the call should succeed
    And the app data file "ssh/known_hosts" should exist
    When I call "GitAccountService.TestConnection" with
      """
      { "account_id": "{{account}}", "repository": "{{repo}}" }
      """
    Then the call should succeed
    And the event "git:host-key-unknown" should have been emitted 1 times

  @first_use
  Scenario: Declining remembers nothing
    When I call "GitAccountService.ConfirmHostKey" with
      """
      { "repository": "{{repo}}", "fingerprint": "{{fingerprint}}", "accept": false }
      """
    Then the call should succeed
    When I call "GitAccountService.TestConnection" with
      """
      { "account_id": "{{account}}", "repository": "{{repo}}" }
      """
    Then the call should fail with code "ACC-05409"

  @first_use
  Scenario: A confirmation for another fingerprint than the host presents is refused
    When I call "GitAccountService.ConfirmHostKey" with
      """
      { "repository": "{{repo}}", "fingerprint": "SHA256:not-the-presented-key", "accept": true }
      """
    Then the call should fail with code "ACC-15400"
    And the app data file "ssh/known_hosts" should not exist

  @changed
  Scenario: A changed host key blocks the connection
    Given the app data file "ssh/known_hosts" contains
      """
      {{ssh_host}} ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
      """
    When I call "GitAccountService.TestConnection" with
      """
      { "account_id": "{{account}}", "repository": "{{repo}}" }
      """
    Then the call should fail with code "ACC-06409"
    And the error should contain the field "error.error_details.0.messages.0" equal to "{{fingerprint}}"
    And the event "git:host-key-unknown" should have been emitted 0 times
    When I call "GitAccountService.ConfirmHostKey" with
      """
      { "repository": "{{repo}}", "fingerprint": "{{fingerprint}}", "accept": true }
      """
    Then the call should fail with code "ACC-06409"
