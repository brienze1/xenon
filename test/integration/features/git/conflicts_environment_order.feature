#language: en
#utf-8

@all @xen_011 @git @sync @conflicts @conflicts_environment_order
Feature: The environment order merges on pull like the collection order
  environments.order in workspace.yaml merges with the same rule as the collections' order: their
  sequence first, then the ids only I added; an id removed on either side stays removed. Two people
  reordering or adding environments on different machines never get a conflict about the order.

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
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXG000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - 01JXG00000000000000000000A
          - 01JXG00000000000000000000B
          - 01JXG00000000000000000000C
      variables: []
      """
    And the workspace file "acme/environments/a.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JXG00000000000000000000A
      name: a
      color: none
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "acme/environments/b.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JXG00000000000000000000B
      name: b
      color: none
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "acme/environments/c.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JXG00000000000000000000C
      name: c
      color: none
      confirmBeforeSend: false
      variables: []
      """
    And I call "SyncService.Link" with
      """
      { "local_id": "{{Acme.local_id}}", "repository": "{{repo}}", "branch": "main", "subfolder": "api/xenon", "account_id": "{{account}}", "clone_path": "{{workspace_root}}/clones/acme" }
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Add the Acme workspace", "item_ids": [ "01JXG000000000000000000000", "01JXG00000000000000000000A", "01JXG00000000000000000000B", "01JXG00000000000000000000C" ] }
      """
    And I call "SyncService.Push" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """

  @auto_merge @reorder
  Scenario: Both sides reorder and add an environment - the orders merge without a conflict
    Given the workspace file "clones/acme/api/xenon/environments/d.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JXG00000000000000000000D
      name: d
      color: none
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "clones/acme/api/xenon/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXG000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - 01JXG00000000000000000000C
          - 01JXG00000000000000000000A
          - 01JXG00000000000000000000B
          - 01JXG00000000000000000000D
      variables: []
      sync:
        repository: {{repo}}
        branch: main
        subfolder: api/xenon
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Add d", "item_ids": [ "01JXG000000000000000000000", "01JXG00000000000000000000D" ] }
      """
    And a teammate changes "api/xenon/environments/e.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: environment
      id: 01JXG00000000000000000000E
      name: e
      color: none
      confirmBeforeSend: false
      variables: []
      """
    And a teammate changes "api/xenon/workspace.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: workspace
      id: 01JXG000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - 01JXG00000000000000000000B
          - 01JXG00000000000000000000A
          - 01JXG00000000000000000000C
          - 01JXG00000000000000000000E
      variables: []
      sync:
        repository: {{repo}}
        branch: main
        subfolder: api/xenon
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "status.state" equal to "ahead"
    And the response field "session" should not exist
    And the workspace file "clones/acme/api/xenon/workspace.yaml" should contain the field "environments.order" with the items "01JXG00000000000000000000B, 01JXG00000000000000000000A, 01JXG00000000000000000000C, 01JXG00000000000000000000E, 01JXG00000000000000000000D"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments" array length equal to 5
    And the response should contain the field "environments.0.name" equal to "b"
    And the response should contain the field "environments.3.name" equal to "e"
    And the response should contain the field "environments.4.name" equal to "d"

  @auto_merge @add_and_remove
  Scenario: One side adds an environment while the other removes one - both changes are kept
    Given the workspace file "clones/acme/api/xenon/environments/d.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JXG00000000000000000000D
      name: d
      color: none
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "clones/acme/api/xenon/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXG000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - 01JXG00000000000000000000A
          - 01JXG00000000000000000000B
          - 01JXG00000000000000000000C
          - 01JXG00000000000000000000D
      variables: []
      sync:
        repository: {{repo}}
        branch: main
        subfolder: api/xenon
      """
    And I call "SyncService.Commit" with
      """
      { "local_id": "{{Acme.local_id}}", "message": "Add d", "item_ids": [ "01JXG000000000000000000000", "01JXG00000000000000000000D" ] }
      """
    And a teammate changes "api/xenon/workspace.yaml" on the remote "acme/backend" to
      """
      xenon: 1
      kind: workspace
      id: 01JXG000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - 01JXG00000000000000000000C
          - 01JXG00000000000000000000A
      variables: []
      sync:
        repository: {{repo}}
        branch: main
        subfolder: api/xenon
      """
    When I call "SyncService.Pull" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "status.state" equal to "ahead"
    And the response field "session" should not exist
    And the workspace file "clones/acme/api/xenon/workspace.yaml" should contain the field "environments.order" with the items "01JXG00000000000000000000C, 01JXG00000000000000000000A, 01JXG00000000000000000000D"
