#language: en
#utf-8

@all @xen_006 @export @postman_export @postman_environment_globals
Feature: Export environments and globals as Postman files
  An environment exports as a Postman environment JSON and the workspace globals as a Postman
  globals JSON. Plain rows carry their initial values, or their current values when "Export current
  values as initial values" is ticked (the workspace itself does not change). Secret rows keep only
  their name: type secret and an empty value; no secret value ever reaches a Postman export.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the current time is "2026-10-11T10:00:00Z"
    And the "local" variable is set to "01JX0000000000000000000101"
    And the workspace file "acme/environments/local.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{local}}
      name: local
      color: green
      confirmBeforeSend: true
      variables:
        - key: base_url
          value: https://local.acme.test
          description: where the api runs
        - key: tenant
          value: acme
          enabled: false
        - key: api_key
          secret: true
      """
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JX0000000000000000000100
      name: Acme
      variables:
        - key: region
          value: eu-west-1
        - key: retries
          value: "3"
          enabled: false
        - key: signing_key
          secret: true
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{local}}", "key": "api_key", "value": "local-SECRET-key" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{local}}", "key": "base_url", "value": "http://127.0.0.1:8080" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "region", "value": "us-east-2" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "signing_key", "value": "globals-SECRET-key" }
      """

  @environment
  Scenario: An environment exports with initial values and secret names only
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "{{local}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/local.postman_environment.json"
      }
      """
    Then the call should succeed
    And the response should contain the field "counts.environments" equal to "1"
    And the response should contain the field "counts.variables" equal to "3"
    And the response should contain the field "secrets" equal to "names_only"
    And the file "local.postman_environment.json" should contain the field "id" equal to "{{local}}"
    And the file "local.postman_environment.json" should contain the field "name" equal to "local"
    And the file "local.postman_environment.json" should contain the field "_postman_variable_scope" equal to "environment"
    And the file "local.postman_environment.json" should contain the field "_postman_exported_at" equal to "2026-10-11T10:00:00.000Z"
    And the file "local.postman_environment.json" should contain the field "_postman_exported_using" matching "^Xenon/[0-9]"
    And the file "local.postman_environment.json" should contain the field "values" array length equal to 3
    And the file "local.postman_environment.json" should contain the field "values.0.key" equal to "base_url"
    And the file "local.postman_environment.json" should contain the field "values.0.value" equal to "https://local.acme.test"
    And the file "local.postman_environment.json" should contain the field "values.0.type" equal to "default"
    And the file "local.postman_environment.json" should contain the field "values.0.enabled" equal to "true"
    And the file "local.postman_environment.json" should contain the field "values.0.description" equal to "where the api runs"
    And the file "local.postman_environment.json" should contain the field "values.1.enabled" equal to "false"
    And the file "local.postman_environment.json" should contain the field "values.2.key" equal to "api_key"
    And the file "local.postman_environment.json" should contain the field "values.2.type" equal to "secret"
    And the file "local.postman_environment.json" should contain the field "values.2.value" equal to ""
    And the file "local.postman_environment.json" should not contain "local-SECRET-key"
    And the file "local.postman_environment.json" should not contain "127.0.0.1:8080"
    And the file "local.postman_environment.json" should not contain the field "color"

  @environment @current_values
  Scenario: Current values replace initial values only in the export when asked
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "{{local}}", "current_values_as_initial": true,
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/local.postman_environment.json"
      }
      """
    Then the call should succeed
    And the file "local.postman_environment.json" should contain the field "values.0.value" equal to "http://127.0.0.1:8080"
    And the file "local.postman_environment.json" should contain the field "values.1.value" equal to "acme"
    And the file "local.postman_environment.json" should contain the field "values.2.value" equal to ""
    And the file "local.postman_environment.json" should not contain "local-SECRET-key"
    And the workspace file "acme/environments/local.yaml" should be unchanged

  @globals
  Scenario: The globals export as a Postman globals file
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "globals",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/acme.postman_globals.json"
      }
      """
    Then the call should succeed
    And the response should contain the field "counts.variables" equal to "3"
    And the file "acme.postman_globals.json" should contain the field "name" equal to "Acme Globals"
    And the file "acme.postman_globals.json" should contain the field "_postman_variable_scope" equal to "globals"
    And the file "acme.postman_globals.json" should contain the field "values" array length equal to 3
    And the file "acme.postman_globals.json" should contain the field "values.0.key" equal to "region"
    And the file "acme.postman_globals.json" should contain the field "values.0.value" equal to "eu-west-1"
    And the file "acme.postman_globals.json" should contain the field "values.1.enabled" equal to "false"
    And the file "acme.postman_globals.json" should contain the field "values.2.type" equal to "secret"
    And the file "acme.postman_globals.json" should contain the field "values.2.value" equal to ""
    And the file "acme.postman_globals.json" should not contain "globals-SECRET-key"

  @globals @current_values
  Scenario: Current global values can be exported as initial values
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "globals", "current_values_as_initial": true,
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/acme.postman_globals.json"
      }
      """
    Then the call should succeed
    And the file "acme.postman_globals.json" should contain the field "values.0.value" equal to "us-east-2"
    And the file "acme.postman_globals.json" should not contain "globals-SECRET-key"
    And the workspace file "acme/workspace.yaml" should be unchanged

  @collection @current_values
  Scenario: Current collection variable values can be exported as initial values
    Given the resource folder "export/shop" is copied to the workspace folder "acme"
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "collection:01JX0000000000000000000001", "key": "page_size", "value": "50" }
      """
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "folder", "item_id": "01JX0000000000000000000002", "current_values_as_initial": true,
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/products.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "products.postman_collection.json" should contain the field "variable.1.key" equal to "page_size"
    And the file "products.postman_collection.json" should contain the field "variable.1.value" equal to "50"
    And the workspace file "acme/collections/shop-api/collection.yaml" should contain the field "variables.1.value" equal to "20"

  @dialog
  Scenario: The export dialog knows the environment has secrets and current values
    When I call "ExportService.Options" with
      """
      { "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "{{local}}" }
      """
    Then the call should succeed
    And the response should contain the field "suggested_file_name" equal to "local"
    And the response should contain the field "has_secrets" equal to "true"
    And the response should contain the field "has_current_values" equal to "true"
    And the response should contain the field "formats.1.suggested_file_name" equal to "local.postman_environment.json"
    When I call "ExportService.Options" with
      """
      { "local_id": "{{Acme.local_id}}", "item_kind": "globals" }
      """
    Then the call should succeed
    And the response should contain the field "formats" array length equal to 1
    And the response should contain the field "formats.0.format" equal to "postman"
    And the response should contain the field "formats.0.suggested_file_name" equal to "globals.postman_globals.json"
