#language: en
#utf-8

@all @xen_018 @aliases @alias_environments @alias_environment_boundaries
Feature: History records the active environment, and alias environments travel with exports
  Every send's history entry records the environment that was active, as it was at send time (id,
  name, colour), or none. A Postman export writes a linked alias's request with a note naming its
  environments in the description; aliases for all environments export unchanged. A Xenon YAML
  export carries the link as it is.

  Background:
    Given the workspace "Beta" exists
    And the workspace "Acme" exists
    And the keychain is available
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01KC0000000000000000000001"
    And the "login" variable is set to "01KC0000000000000000000002"
    And the "hml" variable is set to "01KC0000000000000000000004"
    And the "dev" variable is set to "01KC0000000000000000000005"
    And the "gone" variable is set to "01KC0000000000000000000006"
    And the "mx" variable is set to "01KC0000000000000000000010"
    And the "qa" variable is set to "01KC0000000000000000000011"
    And the "v2" variable is set to "01KC0000000000000000000012"
    And the "old" variable is set to "01KC0000000000000000000013"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01KC0000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/environments/hml.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{hml}}
      name: HML
      color: green
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: DEV
      color: cyan
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      order:
        - login.yaml
      """
    And the workspace file "acme/collections/billing/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      description: Logs a user in.
      method: POST
      url: "{{api}}/login"
      aliases:
        - id: {{mx}}
          name: mx-test-user
          environments:
            - {{hml}}
          method: POST
          url: "{{api}}/login"
        - id: {{qa}}
          name: qa-smoke
          environments:
            - {{hml}}
            - {{dev}}
            - {{gone}}
          method: POST
          url: "{{api}}/login"
        - id: {{v2}}
          name: v2
          method: POST
          url: "{{api}}/login"
      """

  @history
  Scenario: Alias and Default sends record the active environment as it was at send time
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    And the 2 "POST" request to "/login" returns status 200 with the following response
      """
      {"ok":true}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{mx}}", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    When I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}", "name": "Staging", "description": "", "color": "yellow", "confirm_before_send": false }
      """
    And I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 2
    And the response field "" should contain 2 items with "environment.id" equal to "{{hml}}"
    And the response field "" should contain 2 items with "environment.name" equal to "HML"
    And the response field "" should contain 2 items with "environment.color" equal to "green"
    And the response field "" should contain an item with "alias_name" equal to "mx-test-user" and "environment.name" equal to "HML"

  @history
  Scenario: A send with no active environment records none
    Given the 1 "POST" request to "/login" returns status 200 with the following response
      """
      {"ok":true}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{v2}}", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.alias_name" equal to "v2"
    And the response field "0.environment" should not exist

  @history @backwards_compatible
  Scenario: Entries written before environments were recorded load with none next to new entries that have one
    Given the app data file "workspaces/{{Acme.local_id}}/history.jsonl" contains
      """
      {"version":1,"id":"{{old}}","sent_at":"2026-10-01T09:00:00Z","source_request_id":"{{login}}","alias_id":"{{mx}}","alias_name":"mx-test-user","request":{"name":"Login","method":"POST","url":"https://old.test/login"},"result":{"status_code":200,"status_text":"OK"}}
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    And the 1 "POST" request to "/login" returns status 200 with the following response
      """
      {"ok":true}
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 2
    And the response field "" should contain 1 items with "environment.name" equal to "HML"
    And the response field "" should contain an item with "id" equal to "{{old}}" and "alias_name" equal to "mx-test-user"
    When I call "HistoryService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{old}}" }
      """
    Then the call should succeed
    And the response field "entry.environment" should not exist
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" line 2 should contain the field "environment.id" equal to "{{hml}}"
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" line 2 should contain the field "environment.color" equal to "green"

  @postman_export
  Scenario: A Postman export notes a linked alias's environments in its request description
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{billing}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/billing.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "billing.postman_collection.json" should be a valid Postman v2.1 collection
    And the file "billing.postman_collection.json" should contain the field "item.0.name" equal to "Login"
    And the file "billing.postman_collection.json" should contain the field "item.0.description" equal to "Logs a user in."
    And the file "billing.postman_collection.json" should contain the field "item.1.name" equal to "Login (mx-test-user)"
    And the file "billing.postman_collection.json" should contain the field "item.1.description" equal to
      """
      Logs a user in.

      Xenon alias for environments: HML
      """
    And the file "billing.postman_collection.json" should contain the field "item.2.name" equal to "Login (qa-smoke)"
    And the file "billing.postman_collection.json" should contain the field "item.2.description" equal to
      """
      Logs a user in.

      Xenon alias for environments: HML, DEV, Deleted environment
      """
    And the file "billing.postman_collection.json" should contain the field "item.3.name" equal to "Login (v2)"
    And the file "billing.postman_collection.json" should contain the field "item.3.description" equal to "Logs a user in."

  @postman_export
  Scenario: A linked alias of a request without a description gets only the note
    Given the workspace file "acme/collections/billing/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: "{{api}}/login"
      aliases:
        - id: {{mx}}
          name: mx-test-user
          environments:
            - {{hml}}
          method: POST
          url: "{{api}}/login"
      """
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{billing}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/billing.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "billing.postman_collection.json" should contain the field "item.1.description" equal to "Xenon alias for environments: HML"
    And the file "billing.postman_collection.json" should not contain the field "item.0.description"

  @xenon_export @roundtrip
  Scenario: A Xenon YAML export imported into another workspace keeps the alias environments
    Given I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{billing}}",
        "format": "xenon_yaml", "shape": "bundle", "target_path": "{{temp_root}}/billing.yaml"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/billing.yaml"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace file "beta/collections/billing/login.yaml" should contain the field "aliases.0.environments" array length equal to 1
    And the workspace file "beta/collections/billing/login.yaml" should contain the field "aliases.0.environments.0" equal to "{{hml}}"
    And the workspace file "beta/collections/billing/login.yaml" should contain the field "aliases.1.environments" array length equal to 3
    And the workspace file "beta/collections/billing/login.yaml" should not contain the field "aliases.2.environments"
