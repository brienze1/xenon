#language: en
#utf-8

@all @xen_011 @import @import_environment_order
Feature: Imported environments take their place in the environment order
  An environment imported into a workspace without replacing another (no clash, or "Keep both")
  is placed last in environments.order; several imported at once land last in import order. An
  import that replaces an existing environment puts the replacing one in the replaced one's
  place. Importing a whole Xenon workspace copies its workspace.yaml, so its environment order is
  kept as it was.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the "gamma" variable is set to "01JXI000000000000000000003"
    And the "old_dev" variable is set to "01JXI000000000000000000001"
    And the workspace file "acme/environments/gamma.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{gamma}}
      name: Gamma
      color: none
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "acme/environments/zulu.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JXI000000000000000000004
      name: Zulu
      color: none
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXI000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - 01JXI000000000000000000004
          - {{gamma}}
      variables: []
      """

  @no_clash
  Scenario: An environment imported without a clash is listed last
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dev.postman_environment.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And I save the response field "items.0.id" as "dev"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "01JXI000000000000000000004, {{gamma}}, {{dev}}"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments.0.name" equal to "Zulu"
    And the response should contain the field "environments.1.name" equal to "Gamma"
    And the response should contain the field "environments.2.name" equal to "dev"

  @clash @keep_both
  Scenario: An environment kept alongside a clashing one is listed last
    Given the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{old_dev}}
      name: dev
      color: green
      variables: []
      """
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXI000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - {{old_dev}}
          - 01JXI000000000000000000004
          - {{gamma}}
      variables: []
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dev.postman_environment.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "resolution": "keep_both" } ] }
      """
    Then the call should succeed
    And I save the response field "items.0.id" as "imported"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "{{old_dev}}, 01JXI000000000000000000004, {{gamma}}, {{imported}}"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments" array length equal to 4
    And the response should contain the field "environments.0.name" equal to "dev"
    And the response should contain the field "environments.3.id" equal to "{{imported}}"

  @clash @replace
  Scenario: A replacing environment takes the replaced one's place
    Given the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{old_dev}}
      name: dev
      color: green
      variables: []
      """
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXI000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - 01JXI000000000000000000004
          - {{old_dev}}
          - {{gamma}}
      variables: []
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dev.postman_environment.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "resolution": "replace" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "replaced"
    And I save the response field "items.0.id" as "dev"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "01JXI000000000000000000004, {{dev}}, {{gamma}}"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments" array length equal to 3
    And the response should contain the field "environments.0.name" equal to "Zulu"
    And the response should contain the field "environments.1.id" equal to "{{dev}}"
    And the response should contain the field "environments.2.name" equal to "Gamma"

  @batch
  Scenario: Environments imported together are listed last in import order
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dump.zip"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      {
        "source_id": "{{source}}",
        "items": [
          { "key": "0", "include": false },
          { "key": "1", "include": false },
          { "key": "2", "include": false },
          { "key": "3", "include": true, "target_local_id": "{{Acme.local_id}}" },
          { "key": "4", "include": true, "target_local_id": "{{Acme.local_id}}" }
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 2
    And I save the response field "items.0.id" as "first"
    And I save the response field "items.1.id" as "second"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "01JXI000000000000000000004, {{gamma}}, {{first}}, {{second}}"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments.0.name" equal to "Zulu"
    And the response should contain the field "environments.1.name" equal to "Gamma"
    And the response should contain the field "environments.2.name" equal to "billing-dev"
    And the response should contain the field "environments.3.name" equal to "crm-prod"

  @workspace_folder
  Scenario: Importing a whole Xenon workspace keeps its environment order
    Given I call "ImportService.Detect" with
      """
      { "paths": ["{{Acme.path}}"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_folder": "{{temp_root}}/copy" } ] }
      """
    Then the call should succeed
    And the file "copy/workspace.yaml" should contain the field "environments.order" with the items "01JXI000000000000000000004, {{gamma}}"
