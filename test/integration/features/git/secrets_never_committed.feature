#language: en
#utf-8

@all @xen_007 @git @sync @secrets_never_committed
Feature: Secrets and local state never reach the repository
  Only workspace files are committed. Secret values, current values, open tabs and drafts, history,
  cookies and Git credentials live outside the workspace folder, so after a full commit and push
  none of their values appears anywhere in the remote's history (SC-003).

  Background:
    Given the keychain is available
    And the git http server requires the username "oauth2" and the token "tok-never-committed"
    And the remote "acme/backend" has the file "README.md" with
      """
      Backend service
      """
    And the "repo" variable is set to the http url of the remote "acme/backend"
    And the "host" variable is set to the git http server host
    And the "api" variable is set to the mocked api url
    And I call "GitAccountService.AddToken" with
      """
      { "provider": "gitlab", "host": "{{host}}", "token": "tok-never-committed", "repository": "{{repo}}", "commit_name": "Luis", "commit_email": "luis@acme.test" }
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
    And the workspace file "acme/collections/api/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JS0000000000000000000002
      name: Login
      method: GET
      url: "{{base_url}}/login"
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JS0000000000000000000004
      name: dev
      variables:
        - key: base_url
          value: "{{api}}"
          enabled: true
          description: ""
      """
    And I save the workspace file "acme/workspace.yaml" field "id" as "workspace_id"
    And I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "env:01JS0000000000000000000004",
        "rows": [
          { "key": "base_url", "initial_value": "{{api}}", "enabled": true, "description": "" },
          { "key": "api_key", "secret": true, "secret_value": "env-s3cr3t-value", "enabled": true, "description": "vendor" },
          { "key": "user", "initial_value": "shared-user", "set_current": true, "current_value": "current-only-value", "enabled": true, "description": "" }
        ]
      }
      """
    And I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "globals",
        "rows": [ { "key": "password", "secret": true, "secret_value": "globals-s3cr3t-value", "enabled": true, "description": "" } ]
      }
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JS0000000000000000000004" }
      """
    And I call "UIStateService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "tabs": [{ "tab_id": "draft-tab", "request_id": "", "pinned": true, "draft": { "name": "tab-draft-only-value" } }], "active_tab_id": "draft-tab" }
      """
    And the requests to "/login" respond with the header "Set-Cookie" equal to "session=cookie-only-value; Path=/"
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "draft-tab", "request": { "id": "01JS0000000000000000000002", "method": "GET", "url": "{{base_url}}/login", "headers": [ { "key": "X-Trace", "value": "history-only-value" }, { "key": "X-Key", "value": "{{api_key}}" } ] } }
      """

  @sc003
  Scenario: A full commit and push carries no secret, current value, tab, history, cookie or token
    When I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme" }
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Add the Acme workspace", "item_ids": [ "{{workspace_id}}", "01JS0000000000000000000001", "01JS0000000000000000000002", "01JS0000000000000000000004" ] }
      """
    And I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "up_to_date"
    And the file "api/xenon/environments/dev.yaml" on the remote "acme/backend" should exist
    And the file "api/xenon/collections/api/login.yaml" on the remote "acme/backend" should exist
    And the remote "acme/backend" history should not contain "env-s3cr3t-value"
    And the remote "acme/backend" history should not contain "globals-s3cr3t-value"
    And the remote "acme/backend" history should not contain "current-only-value"
    And the remote "acme/backend" history should not contain "tab-draft-only-value"
    And the remote "acme/backend" history should not contain "history-only-value"
    And the remote "acme/backend" history should not contain "cookie-only-value"
    And the remote "acme/backend" history should not contain "tok-never-committed"
    And no file under the workspace folder should contain "tok-never-committed"
    And no file under the app data folder should contain "tok-never-committed"
