#language: en
#utf-8

@all @xen_011 @environments @environment_order_placement
Feature: New environments go last and existing ones keep their place
  Creating or duplicating an environment places it last: environments.order becomes the current
  list with the new id appended, also in a workspace that was never reordered. Renaming keeps an
  environment where it is. Deleting removes its id from environments.order; a workspace with no
  stored order is not rewritten by a delete.

  Background:
    Given the workspace "Acme" exists
    And the "beta" variable is set to "01JXP000000000000000000002"
    And the "gamma" variable is set to "01JXP000000000000000000003"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXP000000000000000000000
      name: Acme
      order: []
      variables: []
      """
    And the workspace file "acme/environments/beta.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{beta}}
      name: Beta
      color: none
      confirmBeforeSend: false
      variables: []
      """
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

  @create @no_order
  Scenario: Creating an environment in a workspace never reordered lists it last and stores the order
    When I call "EnvironmentService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "Alpha" }
      """
    Then the call should succeed
    And I save the response field "id" as "alpha"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "{{beta}}, {{gamma}}, {{alpha}}"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments.0.name" equal to "Beta"
    And the response should contain the field "environments.1.name" equal to "Gamma"
    And the response should contain the field "environments.2.name" equal to "Alpha"

  @create @stored_order
  Scenario: Creating an environment appends it to a stored order
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXP000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - {{gamma}}
          - {{beta}}
      variables: []
      """
    When I call "EnvironmentService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "Alpha" }
      """
    Then the call should succeed
    And I save the response field "id" as "alpha"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "{{gamma}}, {{beta}}, {{alpha}}"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments.0.name" equal to "Gamma"
    And the response should contain the field "environments.1.name" equal to "Beta"
    And the response should contain the field "environments.2.name" equal to "Alpha"

  @duplicate
  Scenario: A duplicated environment is listed last
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXP000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - {{gamma}}
          - {{beta}}
      variables: []
      """
    When I call "EnvironmentService.Duplicate" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{gamma}}" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Gamma Copy"
    And I save the response field "id" as "copy"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "{{gamma}}, {{beta}}, {{copy}}"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments.2.name" equal to "Gamma Copy"

  @rename
  Scenario: Renaming an environment keeps its place and does not touch the stored order
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXP000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - {{gamma}}
          - {{beta}}
      variables: []
      """
    When I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{gamma}}", "name": "Zeta", "description": "", "color": "none", "confirm_before_send": false }
      """
    Then the call should succeed
    And the workspace file "acme/environments/zeta.yaml" should contain the field "id" equal to "{{gamma}}"
    And the workspace file "acme/workspace.yaml" should be unchanged
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments.0.name" equal to "Zeta"
    And the response should contain the field "environments.1.name" equal to "Beta"

  @delete @stored_order
  Scenario: Deleting an environment removes it from the stored order
    Given the "alpha" variable is set to "01JXP000000000000000000001"
    And the workspace file "acme/environments/alpha.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{alpha}}
      name: Alpha
      color: none
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXP000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - {{gamma}}
          - {{alpha}}
          - {{beta}}
      variables: []
      """
    When I call "EnvironmentService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{alpha}}" }
      """
    Then the call should succeed
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "{{gamma}}, {{beta}}"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments" array length equal to 2
    And the response should contain the field "environments.0.name" equal to "Gamma"
    And the response should contain the field "environments.1.name" equal to "Beta"

  @delete @no_order
  Scenario: Deleting an environment in a workspace never reordered leaves workspace.yaml as it was
    When I call "EnvironmentService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{beta}}" }
      """
    Then the call should succeed
    And the workspace file "acme/workspace.yaml" should be unchanged
    And the workspace file "acme/workspace.yaml" should not contain the field "environments"
