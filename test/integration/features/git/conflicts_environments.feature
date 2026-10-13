#language: en
#utf-8

@all @xen_007 @git @sync @conflicts @conflicts_environments
Feature: Resolve environment conflicts after a pull
  Environment files merge row by row: rows changed on one side only, or added on either side, are
  combined without a prompt. A row changed differently on both sides makes the environment a
  conflict with the same choices as a request: Keep mine, Keep theirs or Keep both (theirs saved as
  "<name> (theirs)" with a new id). An environment deleted on one side and changed on the other is
  Keep or Delete. Secret values never appear in a conflict - only the variable names travel.

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
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JS0000000000000000000004
      name: Dev
      variables:
        - key: base_url
          value: https://dev.acme.test
          enabled: true
          description: ""
      """
    And I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "env:01JS0000000000000000000004",
        "rows": [
          { "key": "base_url", "initial_value": "https://dev.acme.test", "enabled": true, "description": "" },
          { "key": "api_key", "secret": true, "secret_value": "env-s3cr3t-value", "enabled": true, "description": "" }
        ]
      }
      """
    And I save the workspace file "acme/workspace.yaml" field "id" as "workspace_id"
    And I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme" }
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Add the Acme workspace", "item_ids": [ "{{workspace_id}}", "01JS0000000000000000000004" ] }
      """
    And I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """


  @rows_merge
  Scenario: Rows changed on different keys merge without a prompt
    Given the workspace file "clones/acme/api/xenon/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JS0000000000000000000004
      name: Dev
      variables:
        - key: base_url
          value: https://mine.acme.test
          enabled: true
          description: ""
        - key: api_key
          secret: true
          enabled: true
          description: ""
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Point dev at mine", "item_ids": [ "01JS0000000000000000000004" ] }
      """
    And a teammate changes "api/xenon/environments/dev.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: environment
      id: 01JS0000000000000000000004
      name: Dev
      variables:
        - key: base_url
          value: https://dev.acme.test
          enabled: true
          description: ""
        - key: api_key
          secret: true
          enabled: true
          description: ""
        - key: timeout
          value: "30"
          enabled: true
          description: ""
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "status.state" equal to "ahead"
    And the response field "session" should not exist
    And the workspace file "clones/acme/api/xenon/environments/dev.yaml" should contain the field "variables" array length equal to 3
    And the workspace file "clones/acme/api/xenon/environments/dev.yaml" should contain the field "variables.0.value" equal to "https://mine.acme.test"
    And the workspace file "clones/acme/api/xenon/environments/dev.yaml" should contain the field "variables.2.key" equal to "timeout"

  @item_conflict @render
  Scenario: A row changed differently on both sides makes the environment a conflict
    Given the workspace file "clones/acme/api/xenon/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JS0000000000000000000004
      name: Dev
      variables:
        - key: base_url
          value: https://mine.acme.test
          enabled: true
          description: ""
        - key: api_key
          secret: true
          enabled: true
          description: ""
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Point dev at mine", "item_ids": [ "01JS0000000000000000000004" ] }
      """
    And a teammate changes "api/xenon/environments/dev.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: environment
      id: 01JS0000000000000000000004
      name: Dev
      variables:
        - key: base_url
          value: https://theirs.acme.test
          enabled: true
          description: ""
        - key: api_key
          secret: true
          enabled: true
          description: ""
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "status.state" equal to "conflict"
    And the response should contain the field "session.conflicts" array length equal to 1
    And the response should contain the field "session.conflicts.0.item_id" equal to "01JS0000000000000000000004"
    And the response should contain the field "session.conflicts.0.item_name" equal to "Dev"
    And the response should contain the field "session.conflicts.0.item_kind" equal to "environment"
    And the response should contain the field "session.conflicts.0.kind" equal to "both_modified"
    And the response should contain the field "session.conflicts.0.mine.environment.name" equal to "Dev"
    And the response field "session.conflicts.0.mine.environment.variables" should contain an item with "key" equal to "base_url" and "value" equal to "https://mine.acme.test"
    And the response field "session.conflicts.0.theirs.environment.variables" should contain an item with "key" equal to "base_url" and "value" equal to "https://theirs.acme.test"
    And the response field "session.conflicts.0.theirs.environment.variables" should contain an item with "key" equal to "api_key" and "secret" equal to "true"
    And the response field "session.conflicts.0.base.environment.variables" should contain an item with "key" equal to "base_url" and "value" equal to "https://dev.acme.test"
    And the response should not contain "env-s3cr3t-value"

  @item_conflict
  Scenario Outline: Keep <resolution> for a conflicting environment
    Given the workspace file "clones/acme/api/xenon/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JS0000000000000000000004
      name: Dev
      variables:
        - key: base_url
          value: https://mine.acme.test
          enabled: true
          description: ""
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Point dev at mine", "item_ids": [ "01JS0000000000000000000004" ] }
      """
    And a teammate changes "api/xenon/environments/dev.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: environment
      id: 01JS0000000000000000000004
      name: Dev
      variables:
        - key: base_url
          value: https://theirs.acme.test
          enabled: true
          description: ""
      """
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000004", "resolution": "<resolution>" }
      """
    And I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "<state>"
    And the workspace file "clones/acme/api/xenon/environments/dev.yaml" should contain the field "variables.0.value" equal to "<value>"
    And the workspace file "clones/acme/api/xenon/environments/dev.yaml" should contain the field "id" equal to "01JS0000000000000000000004"

    Examples:
      | resolution | state      | value                    |
      | mine       | ahead      | https://mine.acme.test   |
      | theirs     | up_to_date | https://theirs.acme.test |

  @item_conflict @keep_both
  Scenario: Keep both saves their environment as "<name> (theirs)" with a new id
    Given the workspace file "clones/acme/api/xenon/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JS0000000000000000000004
      name: Dev
      variables:
        - key: base_url
          value: https://mine.acme.test
          enabled: true
          description: ""
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Point dev at mine", "item_ids": [ "01JS0000000000000000000004" ] }
      """
    And a teammate changes "api/xenon/environments/dev.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: environment
      id: 01JS0000000000000000000004
      name: Dev
      variables:
        - key: base_url
          value: https://theirs.acme.test
          enabled: true
          description: ""
      """
    And I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000004", "resolution": "both" }
      """
    And I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "ahead"
    And the workspace file "clones/acme/api/xenon/environments/dev.yaml" should contain the field "variables.0.value" equal to "https://mine.acme.test"
    And the workspace file "clones/acme/api/xenon/environments/dev-(theirs).yaml" should contain the field "name" equal to "Dev (theirs)"
    And the workspace file "clones/acme/api/xenon/environments/dev-(theirs).yaml" should contain the field "variables.0.value" equal to "https://theirs.acme.test"
    And the workspace file "clones/acme/api/xenon/environments/dev-(theirs).yaml" should contain the field "id" not equal to "01JS0000000000000000000004"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response field "environments" should contain 1 items with "name" equal to "Dev (theirs)"
    When I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "state" equal to "up_to_date"
    And the file "api/xenon/environments/dev-(theirs).yaml" on the remote "acme/backend" should exist

  @delete_modify
  Scenario Outline: An environment deleted here and changed on the remote - <resolution>
    Given the workspace path "clones/acme/api/xenon/environments/dev.yaml" is removed
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Drop dev", "item_ids": [ "01JS0000000000000000000004" ] }
      """
    And a teammate changes "api/xenon/environments/dev.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: environment
      id: 01JS0000000000000000000004
      name: Dev
      variables:
        - key: base_url
          value: https://theirs.acme.test
          enabled: true
          description: ""
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "session.conflicts.0.kind" equal to "delete_modify"
    And the response should contain the field "session.conflicts.0.item_name" equal to "Dev"
    And the response field "session.conflicts.0.mine" should not exist
    And I save the response field "session.id" as "session"
    When I call "SyncService.Resolve" with
      """
      { "session_id": "{{session}}", "item_id": "01JS0000000000000000000004", "resolution": "<resolution>" }
      """
    And I call "SyncService.FinishPull" with
      """
      { "session_id": "{{session}}" }
      """
    Then the call should succeed
    And the response should contain the field "state" equal to "<state>"
    And the workspace file "clones/acme/api/xenon/environments/dev.yaml" should <presence>

    Examples:
      | resolution | state      | presence  |
      | keep       | up_to_date | exist     |
      | delete     | ahead      | not exist |
