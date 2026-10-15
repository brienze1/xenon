#language: en
#utf-8

@all @xen_011 @environments @reorder_environments
Feature: Reorder environments
  The Environments view lists environments in the order the user drags them into. The order lives
  in workspace.yaml under environments.order as a list of environment ids, next to the
  collections' order, so it survives a restart and travels through git. The list shows the stored
  order first (first occurrence wins, ids with no readable environment are ignored), then every
  other readable environment alphabetically ignoring case, then broken environment files by file
  name. A workspace that was never reordered lists alphabetically and its workspace.yaml is never
  rewritten just by listing. Reordering never changes the active environment or any environment
  file.

  Background:
    Given the workspace "Acme" exists
    And the "alpha" variable is set to "01JXE000000000000000000001"
    And the "beta" variable is set to "01JXE000000000000000000002"
    And the "gamma" variable is set to "01JXE000000000000000000003"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXE000000000000000000000
      name: Acme
      order: []
      variables: []
      """
    And the workspace file "acme/environments/alpha.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{alpha}}
      name: alpha
      color: none
      confirmBeforeSend: false
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
      color: yellow
      confirmBeforeSend: false
      variables:
        - key: base_url
          value: https://gamma.acme.test
          enabled: true
          description: ""
      """

  @reorder
  Scenario: Moving an environment returns the new order and stores it in workspace.yaml
    When I call "EnvironmentService.Reorder" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{gamma}}", "index": 0 }
      """
    Then the call should succeed
    And the response should contain the field "environments" array length equal to 3
    And the response should contain the field "environments.0.id" equal to "{{gamma}}"
    And the response should contain the field "environments.1.id" equal to "{{alpha}}"
    And the response should contain the field "environments.2.id" equal to "{{beta}}"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "{{gamma}}, {{alpha}}, {{beta}}"
    And the workspace file "acme/workspace.yaml" should contain the field "order" with the items ""
    And the workspace file "acme/workspace.yaml" should contain the field "id" equal to "01JXE000000000000000000000"
    And the event "variables:changed" should have been emitted 1 times
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments.0.name" equal to "Gamma"
    And the response should contain the field "environments.1.name" equal to "alpha"
    And the response should contain the field "environments.2.name" equal to "Beta"

  @reorder @index
  Scenario Outline: Moving <name> to index <index> places it at that position, or last past the end
    When I call "EnvironmentService.Reorder" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{<variable>}}", "index": <index> }
      """
    Then the call should succeed
    And the response should contain the field "environments.0.name" equal to "<first>"
    And the response should contain the field "environments.1.name" equal to "<second>"
    And the response should contain the field "environments.2.name" equal to "<third>"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "<order>"

    Examples:
      | name  | variable | index | first | second | third | order                           |
      | alpha | alpha    | 1     | Beta  | alpha  | Gamma | {{beta}}, {{alpha}}, {{gamma}}  |
      | alpha | alpha    | 2     | Beta  | Gamma  | alpha | {{beta}}, {{gamma}}, {{alpha}}  |
      | alpha | alpha    | 99    | Beta  | Gamma  | alpha | {{beta}}, {{gamma}}, {{alpha}}  |
      | Gamma | gamma    | 1     | alpha | Gamma  | Beta  | {{alpha}}, {{gamma}}, {{beta}}  |

  @reorder @no_op
  Scenario: Moving an environment to the place it already has writes nothing
    When I call "EnvironmentService.Reorder" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{alpha}}", "index": 0 }
      """
    Then the call should succeed
    And the response should contain the field "environments.0.id" equal to "{{alpha}}"
    And the response should contain the field "environments.1.id" equal to "{{beta}}"
    And the response should contain the field "environments.2.id" equal to "{{gamma}}"
    And the workspace file "acme/workspace.yaml" should be unchanged

  @reorder @no_op @stored_order
  Scenario: Moving an environment to its place in a stored order writes nothing
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXE000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - {{gamma}}
          - {{beta}}
          - {{alpha}}
      variables: []
      """
    When I call "EnvironmentService.Reorder" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{beta}}", "index": 1 }
      """
    Then the call should succeed
    And the response should contain the field "environments.0.id" equal to "{{gamma}}"
    And the response should contain the field "environments.1.id" equal to "{{beta}}"
    And the response should contain the field "environments.2.id" equal to "{{alpha}}"
    And the workspace file "acme/workspace.yaml" should be unchanged

  @list @no_order
  Scenario: A workspace that was never reordered lists alphabetically, ignoring case, and is not rewritten
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "environments.0.name" equal to "alpha"
    And the response should contain the field "environments.1.name" equal to "Beta"
    And the response should contain the field "environments.2.name" equal to "Gamma"
    And the workspace file "acme/workspace.yaml" should be unchanged
    And the workspace file "acme/workspace.yaml" should not contain the field "environments"

  @list @stored_order
  Scenario: The stored order comes first, then unlisted environments alphabetically, then broken files
    Given the "delta" variable is set to "01JXE000000000000000000004"
    And the workspace file "acme/environments/delta.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{delta}}
      name: delta
      color: none
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "acme/environments/aaa-broken.yaml" contains
      """
      xenon: 1
      kind: environment
      name: Broken
      variables: []
      """
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXE000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - 01JXE00000000000000000009Z
          - {{gamma}}
          - {{alpha}}
          - {{gamma}}
      variables: []
      """
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "environments" array length equal to 5
    And the response should contain the field "environments.0.id" equal to "{{gamma}}"
    And the response should contain the field "environments.1.id" equal to "{{alpha}}"
    And the response should contain the field "environments.2.id" equal to "{{beta}}"
    And the response should contain the field "environments.3.id" equal to "{{delta}}"
    And the response should contain the field "environments.4.name" equal to "aaa-broken.yaml"
    And the workspace file "acme/workspace.yaml" should be unchanged

  @reorder @stored_order
  Scenario: Reordering rewrites the order from the readable environments, dropping stale and duplicate ids
    Given the workspace file "acme/environments/aaa-broken.yaml" contains
      """
      xenon: 1
      kind: environment
      name: Broken
      variables: []
      """
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXE000000000000000000000
      name: Acme
      order: []
      environments:
        order:
          - 01JXE00000000000000000009Z
          - {{gamma}}
          - {{alpha}}
          - {{gamma}}
      variables: []
      """
    When I call "EnvironmentService.Reorder" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{beta}}", "index": 0 }
      """
    Then the call should succeed
    And the response should contain the field "environments" array length equal to 4
    And the response should contain the field "environments.0.id" equal to "{{beta}}"
    And the response should contain the field "environments.1.id" equal to "{{gamma}}"
    And the response should contain the field "environments.2.id" equal to "{{alpha}}"
    And the response should contain the field "environments.3.name" equal to "aaa-broken.yaml"
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "{{beta}}, {{gamma}}, {{alpha}}"

  @list @malformed
  Scenario Outline: A malformed environments value is treated as no order and the next reorder replaces it - <case>
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JXE000000000000000000000
      name: Acme
      order: []
      environments: <value>
      variables: []
      """
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "environments.0.name" equal to "alpha"
    And the response should contain the field "environments.1.name" equal to "Beta"
    And the response should contain the field "environments.2.name" equal to "Gamma"
    When I call "EnvironmentService.Reorder" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{gamma}}", "index": 0 }
      """
    Then the call should succeed
    And the workspace file "acme/workspace.yaml" should contain the field "environments.order" with the items "{{gamma}}, {{alpha}}, {{beta}}"

    Examples:
      | case                   | value                     |
      | a text value           | oops                      |
      | order is not a list    | { order: oops }           |
      | order holds a mapping  | { order: [ { a: b } ] }   |

  @reorder @scope
  Scenario: Reordering keeps the active environment and leaves every environment file untouched
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{beta}}" }
      """
    When I call "EnvironmentService.Reorder" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{gamma}}", "index": 0 }
      """
    Then the call should succeed
    And the response should contain the field "active_id" equal to "{{beta}}"
    And the workspace file "acme/environments/alpha.yaml" should be unchanged
    And the workspace file "acme/environments/beta.yaml" should be unchanged
    And the workspace file "acme/environments/gamma.yaml" should be unchanged
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "active_id" equal to "{{beta}}"

  @validation
  Scenario Outline: A reorder request must name a workspace, an environment and a position from 0 - <case>
    When I call "EnvironmentService.Reorder" with
      """
      <payload>
      """
    Then the call should fail with code "ENV-17400"
    And the workspace file "acme/workspace.yaml" should be unchanged

    Examples:
      | case             | payload                                                              |
      | negative index   | { "local_id": "{{Acme.local_id}}", "id": "{{gamma}}", "index": -1 }  |
      | blank id         | { "local_id": "{{Acme.local_id}}", "id": " ", "index": 0 }           |
      | missing local id | { "id": "{{gamma}}", "index": 0 }                                    |

  @not_found
  Scenario Outline: Reordering an environment that does not exist or cannot be read fails - <case>
    Given the workspace file "acme/environments/aaa-broken.yaml" contains
      """
      xenon: 1
      kind: environment
      name: Broken
      variables: []
      """
    When I call "EnvironmentService.Reorder" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "<id>", "index": 0 }
      """
    Then the call should fail with code "ENV-01404"
    And the workspace file "acme/workspace.yaml" should be unchanged

    Examples:
      | case        | id                          |
      | unknown id  | 01JXE00000000000000000009Z  |
      | broken file | error:aaa-broken.yaml       |

  @read_only
  Scenario: A reorder that cannot write workspace.yaml fails and leaves the order as it was
    Given the workspace path "acme" is read-only
    When I call "EnvironmentService.Reorder" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{gamma}}", "index": 0 }
      """
    Then the call should fail with code "WKS-26500"
    And the workspace file "acme/workspace.yaml" should be unchanged
