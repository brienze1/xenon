#language: en
#utf-8

@all @xen_002 @environments @active_environment
Feature: Each workspace remembers its active environment
  EnvironmentService.List returns the workspace's environments (one file each under
  environments/) sorted by name, with their colour and "Confirm before send" flag, plus the active
  environment. SetActive picks one ("" is "No environment"). The choice is local to this machine:
  it is kept per workspace in app data workspaces/<localId>/variables.json (owner-only), never in
  the workspace folder, and is still there after a restart.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    And the "dev" variable is set to "01JA0000000000000000000010"
    And the "prod" variable is set to "01JA0000000000000000000011"
    And the "local" variable is set to "01JA0000000000000000000012"
    And the "globex_prod" variable is set to "01JA0000000000000000000020"
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      color: green
      confirmBeforeSend: false
      variables:
        - key: base_url
          value: https://dev.acme.test
          enabled: true
          description: ""
      """
    And the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{prod}}
      name: Prod
      color: red
      confirmBeforeSend: true
      variables: []
      """
    And the workspace file "acme/environments/local.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{local}}
      name: local
      """
    And the workspace file "globex/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{globex_prod}}
      name: prod
      color: magenta
      confirmBeforeSend: true
      """

  @list
  Scenario: Environments are listed by name with their colour and flag, none active at first
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "active_id" equal to ""
    And the response should contain the field "environments" array length equal to 3
    And the response should contain the field "environments.0.id" equal to "{{dev}}"
    And the response should contain the field "environments.0.name" equal to "dev"
    And the response should contain the field "environments.0.color" equal to "green"
    And the response should contain the field "environments.0.confirm_before_send" equal to "false"
    And the response field "environments.0.error" should not exist
    And the response should contain the field "environments.1.name" equal to "local"
    And the response should contain the field "environments.1.color" equal to "none"
    And the response should contain the field "environments.1.confirm_before_send" equal to "false"
    And the response should contain the field "environments.2.id" equal to "{{prod}}"
    And the response should contain the field "environments.2.name" equal to "Prod"
    And the response should contain the field "environments.2.color" equal to "red"
    And the response should contain the field "environments.2.confirm_before_send" equal to "true"

  @list
  Scenario: A workspace without environments lists none
    Given the workspace path "acme/environments" is removed
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "environments" array length equal to 0
    And the response should contain the field "active_id" equal to ""

  @list @malformed
  Scenario: A malformed environment file is listed with an error and the others still work
    Given the workspace file "acme/environments/broken.yaml" contains
      """
      xenon: 1
      kind: environment
      id: [not, a, string
      """
    And the workspace file "acme/environments/notes.txt" contains
      """
      not an environment
      """
    And the workspace file "acme/environments/request.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JA0000000000000000000099
      name: Wrong kind
      """
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "environments" array length equal to 5
    And the response field "environments" should contain an item with "name" equal to "broken.yaml" and "id" equal to "error:broken.yaml"
    And the response field "environments" should contain an item with "name" equal to "request.yaml" and "id" equal to "error:request.yaml"
    And the response field "environments" should contain an item with "name" equal to "dev" and "color" equal to "green"
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "error:broken.yaml" }
      """
    Then the call should fail with code "ENV-01404"

  @set_active
  Scenario: Setting the active environment is saved locally and listed
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    Then the call should succeed
    And the response should contain the field "active_id" equal to "{{prod}}"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "version" equal to "1"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "activeEnvironmentId" equal to "{{prod}}"
    And the app data path "workspaces/{{Acme.local_id}}/variables.json" should have the permissions "600"
    And no file under the workspace folder should contain "activeEnvironmentId"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "active_id" equal to "{{prod}}"

  @set_active
  Scenario: "No environment" clears the active environment and keeps local current values
    Given the app data file "workspaces/{{Acme.local_id}}/variables.json" contains
      """
      { "version": 1, "activeEnvironmentId": "{{dev}}", "current": { "env:{{dev}}": { "base_url": "http://localhost:8080" } } }
      """
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "" }
      """
    Then the call should succeed
    And the response should contain the field "active_id" equal to ""
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "activeEnvironmentId" equal to ""
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.base_url" equal to "http://localhost:8080"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "active_id" equal to ""

  @restart
  Scenario: The active environment saved on this machine is restored on the next start
    Given the app data file "workspaces/{{Acme.local_id}}/variables.json" contains
      """
      { "version": 1, "activeEnvironmentId": "{{local}}", "current": {} }
      """
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "active_id" equal to "{{local}}"

  @restart
  Scenario: An active environment whose file no longer exists is listed as "No environment"
    Given the app data file "workspaces/{{Acme.local_id}}/variables.json" contains
      """
      { "version": 1, "activeEnvironmentId": "01JA9999999999999999999999", "current": {} }
      """
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "active_id" equal to ""

  @restart
  Scenario: An unreadable variables.json is treated as no local state
    Given the app data file "workspaces/{{Acme.local_id}}/variables.json" contains
      """
      { not json
      """
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "active_id" equal to ""
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "activeEnvironmentId" equal to "{{dev}}"

  @isolation
  Scenario: Each workspace keeps its own active environment
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{globex_prod}}" }
      """
    And I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "active_id" equal to "{{dev}}"
    And the response should contain the field "environments" array length equal to 3
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Globex.local_id}}" }
      """
    Then the response should contain the field "active_id" equal to "{{globex_prod}}"
    And the response should contain the field "environments" array length equal to 1
    And the response should contain the field "environments.0.color" equal to "magenta"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "activeEnvironmentId" equal to "{{dev}}"
    And the app data file "workspaces/{{Globex.local_id}}/variables.json" should contain the field "activeEnvironmentId" equal to "{{globex_prod}}"

  @isolation @errors
  Scenario: An environment of another workspace cannot be activated
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{globex_prod}}" }
      """
    Then the call should fail with code "ENV-01404"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not exist

  @errors
  Scenario: Calls for an unknown workspace fail
    When I call "EnvironmentService.List" with
      """
      { "local_id": "01JA0000000000000000000000" }
      """
    Then the call should fail with code "WKS-07404"
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "01JA0000000000000000000000", "id": "" }
      """
    Then the call should fail with code "WKS-07404"

  @errors @validation
  Scenario Outline: Invalid input is rejected - <method>
    When I call "EnvironmentService.<method>" with
      """
      <payload>
      """
    Then the call should fail with code "<code>"

    Examples:
      | method    | payload                    | code      |
      | List      | { "local_id": "" }         | ENV-03400 |
      | SetActive | { "local_id": " ", "id": "" } | ENV-04400 |
