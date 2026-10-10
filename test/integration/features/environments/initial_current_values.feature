#language: en
#utf-8

@all @xen_002 @environments @initial_current_values
Feature: Initial values in the files, current values on this machine
  A variable row's initial value lives in the workspace file and is shared through git; its
  current value is a local override kept only in app data workspaces/<localId>/variables.json.
  Without a current value the row uses (and shows) its initial value. Reset drops the override,
  Persist writes it to the file as the new initial value; Reset all / Persist all do it for every
  row. Every change is announced with "variables:changed" so open views refresh. Files edited
  outside Xenon (for example by git) are picked up live and the local overrides survive.

  Background:
    Given the workspace "Acme" exists
    And the "dev" variable is set to "01JC0000000000000000000010"
    And the "api" variable is set to the mocked api url
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
          value: "{{api}}/initial"
          enabled: true
          description: ""
        - key: tenant
          value: acme
          enabled: true
          description: ""
      """
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" contains
      """
      { "version": 1, "activeEnvironmentId": "{{dev}}", "current": {} }
      """

  @current
  Scenario: A current value overrides the initial value without touching the file
    When I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "base_url", "value": "{{api}}/current" }
      """
    Then the call should succeed
    And the response should contain the field "key" equal to "base_url"
    And the response should contain the field "initial_value" equal to "{{api}}/initial"
    And the response should contain the field "current_value" equal to "{{api}}/current"
    And the response should contain the field "has_current_override" equal to "true"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.value" equal to "{{api}}/initial"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.base_url" equal to "{{api}}/current"
    And no file under the workspace folder should contain "/current"
    And the event "variables:changed" should have been emitted 1 times
    And the 1 "variables:changed" event should contain the field "local_id" equal to "{{Acme.local_id}}"
    And the 1 "variables:changed" event should contain the field "scopes.0" equal to "env:{{dev}}"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{base_url}}/users" } }
      """
    Then the call should succeed
    And the "GET" request to "/current/users" should have been received 1 times

  @current
  Scenario: An empty current value clears the override so the initial value applies
    Given I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "base_url", "value": "{{api}}/current" }
      """
    When I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "base_url", "value": "" }
      """
    Then the call should succeed
    And the response should contain the field "current_value" equal to "{{api}}/initial"
    And the response should contain the field "has_current_override" equal to "false"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "/current"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{base_url}}/users" } }
      """
    Then the "GET" request to "/initial/users" should have been received 1 times

  @current @errors
  Scenario: A current value for an unknown row is rejected
    When I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "nope", "value": "x" }
      """
    Then the call should fail with code "VAR-03404"

  @reset
  Scenario: Reset puts the current value back to the initial value
    Given I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "base_url", "value": "{{api}}/current" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "tenant", "value": "globex" }
      """
    When I call "VariableService.Reset" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "keys": ["base_url"] }
      """
    Then the call should succeed
    And the response should contain the field "scope" equal to "env:{{dev}}"
    And the response should contain the field "rows.0.current_value" equal to "{{api}}/initial"
    And the response should contain the field "rows.0.has_current_override" equal to "false"
    And the response should contain the field "rows.1.current_value" equal to "globex"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.tenant" equal to "globex"
    And the workspace file "acme/environments/dev.yaml" should be unchanged

  @reset
  Scenario: Reset all drops every override of the table
    Given I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "base_url", "value": "{{api}}/current" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "tenant", "value": "globex" }
      """
    When I call "VariableService.Reset" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "keys": [] }
      """
    Then the call should succeed
    And the response should contain the field "rows.0.has_current_override" equal to "false"
    And the response should contain the field "rows.1.has_current_override" equal to "false"
    And the response should contain the field "rows.1.current_value" equal to "acme"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "globex"

  @persist
  Scenario: Persist writes the current value to the file as the new initial value
    Given I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "tenant", "value": "globex" }
      """
    When I call "VariableService.Persist" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "keys": ["tenant"] }
      """
    Then the call should succeed
    And the response should contain the field "rows.1.initial_value" equal to "globex"
    And the response should contain the field "rows.1.current_value" equal to "globex"
    And the response should contain the field "rows.1.has_current_override" equal to "false"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.value" equal to "globex"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.value" equal to "{{api}}/initial"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "globex"

  @persist
  Scenario: Persist all writes every override to the file
    Given I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "base_url", "value": "https://persisted.acme.test" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "tenant", "value": "globex" }
      """
    When I call "VariableService.Persist" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "keys": [] }
      """
    Then the call should succeed
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.value" equal to "https://persisted.acme.test"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.value" equal to "globex"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "globex"

  @table
  Scenario: Saving the table writes initial values to the file and current values locally
    When I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "env:{{dev}}",
        "rows": [
          { "key": "base_url", "initial_value": "https://dev.acme.test", "set_current": true, "current_value": "http://localhost:8080", "enabled": true, "description": "the API" },
          { "key": "tenant", "initial_value": "acme", "enabled": false, "description": "" },
          { "key": "page_size", "initial_value": "50", "enabled": true, "description": "" }
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "rows" array length equal to 3
    And the response should contain the field "rows.0.current_value" equal to "http://localhost:8080"
    And the response should contain the field "rows.2.current_value" equal to "50"
    And the response should contain the field "warnings" array length equal to 0
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.value" equal to "https://dev.acme.test"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.description" equal to "the API"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.enabled" equal to "false"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.2.key" equal to "page_size"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.base_url" equal to "http://localhost:8080"
    And no file under the workspace folder should contain "localhost:8080"

  @table @rename
  Scenario: Renaming a row moves its current value to the new name
    Given I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "tenant", "value": "globex" }
      """
    When I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "env:{{dev}}",
        "rows": [
          { "key": "base_url", "initial_value": "{{api}}/initial", "enabled": true, "description": "" },
          { "key": "company", "previous_key": "tenant", "initial_value": "acme", "enabled": true, "description": "" }
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "rows.1.key" equal to "company"
    And the response should contain the field "rows.1.current_value" equal to "globex"
    And the response should contain the field "rows.1.has_current_override" equal to "true"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.company" equal to "globex"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "tenant"

  @table
  Scenario: Removing a row from the table drops its current value
    Given I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "tenant", "value": "globex" }
      """
    When I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "env:{{dev}}",
        "rows": [ { "key": "base_url", "initial_value": "{{api}}/initial", "enabled": true, "description": "" } ]
      }
      """
    Then the call should succeed
    And the response should contain the field "rows" array length equal to 1
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "globex"

  @table @rollback
  Scenario: A save that cannot write the file leaves the file and the local values as before
    Given I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "tenant", "value": "globex" }
      """
    And the workspace path "acme/environments" is read-only
    When I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "env:{{dev}}",
        "rows": [
          { "key": "base_url", "initial_value": "https://changed.acme.test", "enabled": true, "description": "" },
          { "key": "tenant", "initial_value": "acme", "set_current": true, "current_value": "initech", "enabled": true, "description": "" }
        ]
      }
      """
    Then the call should fail with code "ENV-08500"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.value" equal to "{{api}}/initial"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.tenant" equal to "globex"

  @restart
  Scenario: Current values survive a restart
    Given I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "base_url", "value": "{{api}}/current" }
      """
    When the app is restarted
    And I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.0.current_value" equal to "{{api}}/current"

  @outside_edit
  Scenario: An initial value changed outside Xenon is picked up and the local override is kept
    Given I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "base_url", "value": "{{api}}/current" }
      """
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
          value: "{{api}}/from-git"
          enabled: true
          description: ""
        - key: tenant
          value: from-git
          enabled: true
          description: ""
      """
    When the file watcher reports a change to the workspace path "acme/environments/dev.yaml"
    Then the event "variables:changed" should have been emitted 2 times
    And the 2 "variables:changed" event should contain the field "scopes.0" equal to "env:{{dev}}"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}" }
      """
    Then the response should contain the field "rows.0.initial_value" equal to "{{api}}/from-git"
    And the response should contain the field "rows.0.current_value" equal to "{{api}}/current"
    And the response should contain the field "rows.1.current_value" equal to "from-git"

  @outside_edit
  Scenario: Xenon's own writes to an environment file are not reported as outside edits
    Given I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I call "VariableService.Persist" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "keys": [] }
      """
    When the file watcher reports a change to the workspace path "acme/environments/dev.yaml"
    Then the event "variables:changed" should have been emitted 1 times

  @outside_edit
  Scenario: The active environment's file deleted outside Xenon makes "No environment" active
    Given I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And the workspace path "acme/environments/dev.yaml" is removed
    When the file watcher reports the removal of the workspace path "acme/environments/dev.yaml"
    Then the event "variables:changed" should have been emitted 1 times
    And the 1 "variables:changed" event should contain the field "reason" equal to "active_environment_removed"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "activeEnvironmentId" equal to ""
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "active_id" equal to ""

  @outside_edit
  Scenario: Another environment file deleted outside Xenon keeps the active environment
    Given I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And the workspace file "acme/environments/other.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JC0000000000000000000011
      name: other
      """
    And the workspace path "acme/environments/other.yaml" is removed
    When the file watcher reports the removal of the workspace path "acme/environments/other.yaml"
    Then the event "variables:changed" should have been emitted 1 times
    And the 1 "variables:changed" event should contain the field "scopes" equal to "[]"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "activeEnvironmentId" equal to "{{dev}}"

  @errors @validation
  Scenario Outline: Invalid input is rejected - <method>
    When I call "VariableService.<method>" with
      """
      <payload>
      """
    Then the call should fail with code "<code>"

    Examples:
      | method     | payload                                                        | code      |
      | GetTable   | { "local_id": "", "scope": "globals" }                         | VAR-05400 |
      | GetTable   | { "local_id": "{{Acme.local_id}}", "scope": "nope" }           | VAR-05400 |
      | SaveTable  | { "local_id": "x", "scope": "" }                               | VAR-08400 |
      | SetCurrent | { "local_id": "x", "scope": "globals", "key": "" }             | VAR-09400 |
      | Reset      | { "local_id": "", "scope": "globals" }                         | VAR-10400 |
      | Persist    | { "local_id": "", "scope": "globals" }                         | VAR-11400 |
      | AddVariable | { "local_id": "x", "scope": "globals", "key": "" }            | VAR-12400 |
      | GetTable   | { "local_id": "{{Acme.local_id}}", "scope": "env:01JC9999999999999999999999" } | ENV-01404 |
