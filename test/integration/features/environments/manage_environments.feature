#language: en
#utf-8

@all @xen_002 @environments @manage_environments
Feature: Create, rename, duplicate and delete environments
  Each environment is one readable YAML file in the workspace's environments/ folder holding its
  id, name, colour, "Confirm before send" flag and variable rows (initial values only). Names are
  unique per workspace, ignoring case. Renaming renames the file. Duplicating copies the rows,
  initial values, settings, this machine's current values and the secret values into
  "<name> Copy". Deleting removes the file, the local current values and the secret values; if the
  environment was active, "No environment" becomes active.

  Background:
    Given the workspace "Acme" exists
    And the "hml" variable is set to "01JB0000000000000000000010"
    And the workspace file "acme/environments/hml.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{hml}}
      name: hml
      description: homologation
      color: yellow
      confirmBeforeSend: false
      variables:
        - key: base_url
          value: https://hml.acme.test
          enabled: true
          description: the API
        - key: tenant
          value: acme
          enabled: false
          description: ""
        - key: api_key
          secret: true
          enabled: true
          description: from the vendor portal
      """

  @create
  Scenario: Creating an environment writes its file and returns it
    When I call "EnvironmentService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "dev" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "dev"
    And the response should contain the field "color" equal to "none"
    And the response should contain the field "confirm_before_send" equal to "false"
    And the response should contain the field "variables" array length equal to 0
    And the response should contain the field "id" matching "^[0-9A-HJKMNP-TV-Z]{26}$"
    And I save the response field "id" as "dev"
    And the workspace file "acme/environments/dev.yaml" should contain the field "xenon" equal to "1"
    And the workspace file "acme/environments/dev.yaml" should contain the field "kind" equal to "environment"
    And the workspace file "acme/environments/dev.yaml" should contain the field "id" equal to "{{dev}}"
    And the workspace file "acme/environments/dev.yaml" should contain the field "name" equal to "dev"
    And the workspace file "acme/environments/dev.yaml" should contain the field "color" equal to "none"
    And the workspace file "acme/environments/dev.yaml" should contain the field "confirmBeforeSend" equal to "false"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables" with the items ""
    And the event "variables:changed" should have been emitted 1 times
    And the 1 "variables:changed" event should contain the field "scopes.0" equal to "env:{{dev}}"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments" array length equal to 2
    And the response should contain the field "environments.0.name" equal to "hml"
    And the response should contain the field "environments.1.name" equal to "dev"

  @create @defaults
  Scenario Outline: A new environment named like production defaults to red with confirmation - <name>
    When I call "EnvironmentService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "<name>" }
      """
    Then the call should succeed
    And the response should contain the field "color" equal to "<color>"
    And the response should contain the field "confirm_before_send" equal to "<confirm>"
    And the workspace file "acme/environments/<file>" should contain the field "color" equal to "<color>"
    And the workspace file "acme/environments/<file>" should contain the field "confirmBeforeSend" equal to "<confirm>"

    Examples:
      | name          | file              | color | confirm |
      | prod          | prod.yaml         | red   | true    |
      | Production EU | production-eu.yaml | red   | true    |
      | pre-PROD      | pre-prod.yaml     | red   | true    |
      | staging       | staging.yaml      | none  | false   |

  @create @errors
  Scenario Outline: An environment name must be unique ignoring case and not blank - <case>
    When I call "EnvironmentService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "<name>" }
      """
    Then the call should fail with code "<code>"
    And the workspace folder "acme/environments" should contain the entries "hml.yaml"

    Examples:
      | case             | name | code      |
      | same name        | hml  | ENV-02409 |
      | different case   | HML  | ENV-02409 |
      | blank            |      | ENV-11400 |

  @get
  Scenario: Getting an environment returns its settings and rows with current values
    Given the app data file "workspaces/{{Acme.local_id}}/variables.json" contains
      """
      { "version": 1, "activeEnvironmentId": "", "current": { "env:{{hml}}": { "base_url": "http://localhost:8080" } } }
      """
    When I call "EnvironmentService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "hml"
    And the response should contain the field "description" equal to "homologation"
    And the response should contain the field "color" equal to "yellow"
    And the response should contain the field "variables" array length equal to 3
    And the response should contain the field "variables.0.key" equal to "base_url"
    And the response should contain the field "variables.0.initial_value" equal to "https://hml.acme.test"
    And the response should contain the field "variables.0.current_value" equal to "http://localhost:8080"
    And the response should contain the field "variables.0.has_current_override" equal to "true"
    And the response should contain the field "variables.0.secret" equal to "false"
    And the response should contain the field "variables.0.description" equal to "the API"
    And the response field "variables.0.secret_state" should not exist
    And the response should contain the field "variables.1.current_value" equal to "acme"
    And the response should contain the field "variables.1.has_current_override" equal to "false"
    And the response should contain the field "variables.1.enabled" equal to "false"
    And the response should contain the field "variables.2.key" equal to "api_key"
    And the response should contain the field "variables.2.secret" equal to "true"
    And the response should contain the field "variables.2.initial_value" equal to ""
    And the response should contain the field "variables.2.current_value" equal to ""
    And the response should contain the field "variables.2.secret_state" equal to "missing"

  @get @errors
  Scenario: Getting an unknown environment fails
    When I call "EnvironmentService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JB9999999999999999999999" }
      """
    Then the call should fail with code "ENV-01404"

  @update
  Scenario: Updating an environment changes its settings and renaming renames its file
    When I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}", "name": "Homologation", "description": "shared", "color": "cyan", "confirm_before_send": true }
      """
    Then the call should succeed
    And the response should contain the field "id" equal to "{{hml}}"
    And the response should contain the field "name" equal to "Homologation"
    And the response should contain the field "color" equal to "cyan"
    And the response should contain the field "confirm_before_send" equal to "true"
    And the response should contain the field "variables" array length equal to 3
    And the workspace file "acme/environments/hml.yaml" should not exist
    And the workspace file "acme/environments/homologation.yaml" should contain the field "id" equal to "{{hml}}"
    And the workspace file "acme/environments/homologation.yaml" should contain the field "name" equal to "Homologation"
    And the workspace file "acme/environments/homologation.yaml" should contain the field "description" equal to "shared"
    And the workspace file "acme/environments/homologation.yaml" should contain the field "color" equal to "cyan"
    And the workspace file "acme/environments/homologation.yaml" should contain the field "confirmBeforeSend" equal to "true"
    And the workspace file "acme/environments/homologation.yaml" should contain the field "variables.0.value" equal to "https://hml.acme.test"
    And the workspace file "acme/environments/homologation.yaml" should not contain the field "variables.2.value"
    And the event "variables:changed" should have been emitted 1 times

  @update @errors
  Scenario Outline: Invalid updates are rejected - <case>
    Given the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JB0000000000000000000011
      name: dev
      """
    When I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "<id>", "name": "<name>", "description": "", "color": "<color>", "confirm_before_send": false }
      """
    Then the call should fail with code "<code>"
    And the workspace file "acme/environments/hml.yaml" should contain the field "name" equal to "hml"

    Examples:
      | case                 | id          | name | color   | code      |
      | name taken           | {{hml}}     | DEV  | none    | ENV-02409 |
      | unknown colour       | {{hml}}     | hml  | purple  | ENV-12400 |
      | blank name           | {{hml}}     |      | none    | ENV-12400 |
      | unknown environment  | 01JB0000000000000000000099 | x | none | ENV-01404 |

  @update
  Scenario: Renaming an environment to a different case of its own name keeps it
    When I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}", "name": "HML", "description": "", "color": "yellow", "confirm_before_send": false }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "HML"

  @duplicate
  Scenario: Duplicating copies rows, settings, current values and secret values
    Given the app data file "workspaces/{{Acme.local_id}}/variables.json" contains
      """
      { "version": 1, "activeEnvironmentId": "", "current": { "env:{{hml}}": { "base_url": "http://localhost:9000" } } }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{hml}}", "key": "api_key", "value": "hml-s3cr3t" }
      """
    When I call "EnvironmentService.Duplicate" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "hml Copy"
    And the response should contain the field "id" not equal to "{{hml}}"
    And the response should contain the field "description" equal to "homologation"
    And the response should contain the field "color" equal to "yellow"
    And the response should contain the field "variables" array length equal to 3
    And the response should contain the field "variables.0.current_value" equal to "http://localhost:9000"
    And the response should contain the field "variables.1.enabled" equal to "false"
    And the response should contain the field "variables.2.secret_state" equal to "set"
    And the response should contain the field "variables.2.current_value" equal to ""
    And I save the response field "id" as "copy"
    And the workspace file "acme/environments/hml-copy.yaml" should contain the field "id" equal to "{{copy}}"
    And the workspace file "acme/environments/hml-copy.yaml" should contain the field "name" equal to "hml Copy"
    And the workspace file "acme/environments/hml-copy.yaml" should contain the field "variables.0.value" equal to "https://hml.acme.test"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{copy}}.base_url" equal to "http://localhost:9000"
    And no file under the workspace folder should contain "hml-s3cr3t"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{copy}}", "key": "api_key" }
      """
    Then the response should contain the field "value" equal to "hml-s3cr3t"
    When I call "EnvironmentService.Duplicate" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    Then the response should contain the field "name" equal to "hml Copy 2"
    And the workspace file "acme/environments/hml-copy-2.yaml" should exist
    When I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{copy}}", "name": "prod", "description": "", "color": "red", "confirm_before_send": true }
      """
    Then the call should succeed
    And the workspace file "acme/environments/prod.yaml" should contain the field "id" equal to "{{copy}}"

  @delete
  Scenario: Deleting the active environment removes its file, local values and secrets
    Given I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{hml}}", "key": "api_key", "value": "doomed-s3cr3t" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{hml}}", "key": "base_url", "value": "http://localhost:7000" }
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    When I call "EnvironmentService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    Then the call should succeed
    And the response should contain the field "active_id" equal to ""
    And the workspace file "acme/environments/hml.yaml" should not exist
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "activeEnvironmentId" equal to ""
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "localhost:7000"
    Given the workspace file "acme/environments/restored.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{hml}}
      name: restored from git
      variables:
        - key: api_key
          secret: true
          enabled: true
          description: ""
      """
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{hml}}", "key": "api_key" }
      """
    Then the call should fail with code "VLT-03404"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments" array length equal to 1
    And the response should contain the field "active_id" equal to ""

  @delete
  Scenario: Deleting an inactive environment keeps the active one
    Given the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JB0000000000000000000011
      name: dev
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JB0000000000000000000011" }
      """
    When I call "EnvironmentService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    Then the call should succeed
    And the response should contain the field "active_id" equal to "01JB0000000000000000000011"
    And the event "variables:changed" should have been emitted 2 times
    And the 1 "variables:changed" event should contain the field "scopes" equal to "[]"
    And the 2 "variables:changed" event should contain the field "scopes.0" equal to "env:{{hml}}"

  @delete @errors
  Scenario: Deleting an unknown environment fails
    When I call "EnvironmentService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JB9999999999999999999999" }
      """
    Then the call should fail with code "ENV-01404"

  @malformed
  Scenario: A malformed environment file is listed with an error and cannot be opened
    Given the workspace file "acme/environments/broken.yaml" contains
      """
      xenon: 1
      kind: environment
      id: [unclosed
      """
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response field "environments" should contain an item with "name" equal to "broken.yaml" and "id" equal to "error:broken.yaml"
    And the response should contain the field "environments.1.error" containing "broken.yaml"
    When I call "EnvironmentService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "error:broken.yaml" }
      """
    Then the call should fail with code "ENV-01404"

  @errors @validation
  Scenario Outline: Invalid input is rejected - <method>
    When I call "EnvironmentService.<method>" with
      """
      <payload>
      """
    Then the call should fail with code "<code>"

    Examples:
      | method    | payload                               | code      |
      | Get       | { "local_id": "", "id": "x" }         | ENV-13400 |
      | Get       | { "local_id": "x", "id": "" }         | ENV-13400 |
      | Create    | { "local_id": "", "name": "x" }       | ENV-11400 |
      | Update    | { "local_id": "x", "id": "", "name": "x", "color": "none" } | ENV-12400 |
      | Duplicate | { "local_id": "x", "id": "" }         | ENV-14400 |
      | Delete    | { "local_id": "x", "id": " " }        | ENV-15400 |
