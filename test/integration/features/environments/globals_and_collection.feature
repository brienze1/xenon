#language: en
#utf-8

@all @xen_002 @environments @globals_and_collection
Feature: Globals and collection variables use the same table
  The workspace's globals live in workspace.yaml and a collection's variables in its
  collection.yaml, both as variable rows with initial values only; current values stay local.
  Folders have no variables: a "variables" key found in a folder.yaml is ignored and preserved.
  Keys are case-sensitive, stored exactly as typed and must not contain "{", "}" or line breaks.
  Duplicate keys in one table are allowed (the last enabled row wins) and flagged.

  Background:
    Given the workspace "Acme" exists
    And the "billing" variable is set to "01JD0000000000000000000001"
    And the "invoices" variable is set to "01JD0000000000000000000002"
    And the "api" variable is set to the mocked api url
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JD0000000000000000000000
      name: Acme
      order:
        - billing
      variables:
        - key: user
          value: luis
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      order:
        - invoices
      variables:
        - key: page_size
          value: "50"
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      variables:
        - key: kept
          value: as-is
      """

  @globals
  Scenario: Globals are edited through the table and stored in workspace.yaml
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals" }
      """
    Then the call should succeed
    And the response should contain the field "scope" equal to "globals"
    And the response should contain the field "rows.0.key" equal to "user"
    And the response should contain the field "rows.0.current_value" equal to "luis"
    When I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "globals",
        "rows": [
          { "key": "user", "initial_value": "luis", "set_current": true, "current_value": "brienze", "enabled": true, "description": "" },
          { "key": "region", "initial_value": "eu", "enabled": true, "description": "where" }
        ]
      }
      """
    Then the call should succeed
    And the workspace file "acme/workspace.yaml" should contain the field "variables.1.key" equal to "region"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.1.value" equal to "eu"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.1.description" equal to "where"
    And the workspace file "acme/workspace.yaml" should contain the field "name" equal to "Acme"
    And the workspace file "acme/workspace.yaml" should contain the field "order" with the items "billing"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.globals.user" equal to "brienze"
    And the 1 "variables:changed" event should contain the field "scopes.0" equal to "globals"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/whoami", "headers": [ { "key": "X-User", "value": "{{user}}" }, { "key": "X-Region", "value": "{{region}}" } ] } }
      """
    Then the 1 "GET" request to "/whoami" should have the header "X-User" equal to "brienze"
    And the 1 "GET" request to "/whoami" should have the header "X-Region" equal to "eu"

  @collection
  Scenario: Collection variables are edited through the table and stored in collection.yaml
    When I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "collection:{{billing}}",
        "rows": [
          { "key": "page_size", "initial_value": "100", "enabled": true, "description": "" },
          { "key": "currency", "initial_value": "EUR", "enabled": true, "description": "" }
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "scope" equal to "collection:{{billing}}"
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "variables.0.value" equal to "100"
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "variables.1.key" equal to "currency"
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "order" with the items "invoices"
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "name" equal to "Billing"
    When I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "collection:{{billing}}", "key": "locale", "value": "pt-BR", "secret": false }
      """
    Then the call should succeed
    And the response should contain the field "key" equal to "locale"
    And the response should contain the field "initial_value" equal to "pt-BR"
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "variables.2.value" equal to "pt-BR"

  @collection @errors
  Scenario: Adding a variable whose name is already in the table is rejected
    When I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "collection:{{billing}}", "key": "page_size", "value": "1", "secret": false }
      """
    Then the call should fail with code "VAR-04409"

  @folders
  Scenario: Folders have no variables and keep a variables key found in their file
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "collection:{{invoices}}" }
      """
    Then the call should fail with code "VAR-06422"
    When I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "collection:{{invoices}}", "rows": [ { "key": "x", "initial_value": "y", "enabled": true } ] }
      """
    Then the call should fail with code "VAR-06422"
    And the workspace file "acme/collections/billing/invoices/folder.yaml" should be unchanged
    When I call "TreeService.SaveSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{invoices}}", "description": "the folder", "auth": { "type": "inherit" }, "headers": [] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/invoices/folder.yaml" should contain the field "variables.0.key" equal to "kept"
    And the workspace file "acme/collections/billing/invoices/folder.yaml" should contain the field "variables.0.value" equal to "as-is"

  @collection @errors
  Scenario: An unknown collection has no table
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "collection:01JD9999999999999999999999" }
      """
    Then the call should fail with code "TRE-02404"

  @keys @errors
  Scenario Outline: Invalid keys are rejected and nothing is written - <case>
    When I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "<key>", "initial_value": "v", "enabled": true } ] }
      """
    Then the call should fail with code "VAR-00400"
    And the workspace file "acme/workspace.yaml" should be unchanged

    Examples:
      | case          | key       |
      | empty         |           |
      | opening brace | a{b       |
      | closing brace | a}b       |
      | line break    | a\nb      |

  @keys
  Scenario: Keys are stored exactly as typed and duplicates are flagged while the last enabled row wins
    When I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "globals",
        "rows": [
          { "key": "tenant", "initial_value": "first", "enabled": true, "description": "" },
          { "key": " tenant ", "initial_value": "spaced", "enabled": true, "description": "" },
          { "key": "Tenant", "initial_value": "capital", "enabled": true, "description": "" },
          { "key": "tenant", "initial_value": "second", "enabled": true, "description": "" },
          { "key": "tenant", "initial_value": "disabled", "enabled": false, "description": "" }
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "rows.0.duplicate" equal to "true"
    And the response should contain the field "rows.1.duplicate" equal to "false"
    And the response should contain the field "rows.1.surrounding_spaces" equal to "true"
    And the response should contain the field "rows.2.duplicate" equal to "false"
    And the response should contain the field "rows.3.duplicate" equal to "true"
    And the response should contain the field "rows.0.surrounding_spaces" equal to "false"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.1.key" equal to " tenant "
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/keys", "headers": [ { "key": "X-A", "value": "{{tenant}}" }, { "key": "X-B", "value": "{{ tenant }}" }, { "key": "X-C", "value": "{{Tenant}}" } ] } }
      """
    Then the 1 "GET" request to "/keys" should have the header "X-A" equal to "second"
    And the 1 "GET" request to "/keys" should have the header "X-B" equal to "spaced"
    And the 1 "GET" request to "/keys" should have the header "X-C" equal to "capital"

  @outside_edit
  Scenario: Globals and collection variables changed outside Xenon are announced
    Given I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      order:
        - invoices
      variables:
        - key: page_size
          value: "25"
          enabled: true
          description: ""
      """
    When the file watcher reports a change to the workspace path "acme/collections/billing/collection.yaml"
    Then the event "variables:changed" should have been emitted 1 times
    And the 1 "variables:changed" event should contain the field "scopes.0" equal to "collection:{{billing}}"
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JD0000000000000000000000
      name: Acme
      order:
        - billing
      variables:
        - key: user
          value: someone-else
          enabled: true
          description: ""
      """
    When the file watcher reports a change to the workspace path "acme/workspace.yaml"
    Then the event "variables:changed" should have been emitted 2 times
    And the 2 "variables:changed" event should contain the field "scopes.0" equal to "globals"
    And the event "workspace:tree-changed" should have been emitted 2 times
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals" }
      """
    Then the response should contain the field "rows.0.initial_value" equal to "someone-else"

  @outside_edit
  Scenario: A request file changed outside Xenon does not announce variable changes
    Given I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And the workspace file "acme/collections/billing/invoices/list.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JD0000000000000000000003
      name: List
      """
    When the file watcher reports a change to the workspace path "acme/collections/billing/invoices/list.yaml"
    Then the event "workspace:tree-changed" should have been emitted 1 times
    And the event "variables:changed" should have been emitted 0 times
