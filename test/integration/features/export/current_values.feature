#language: en
#utf-8

@all @xen_006 @export @xenon_export @current_values
Feature: Export current values as initial values in a Xenon YAML export
  "Export current values as initial values" (asked each time, off by default) writes the plain
  variables' current values as their initial values in the exported files only; secret rows never
  get a value and the workspace files do not change. Without the option the exported variable files
  are the workspace files byte for byte.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the resource folder "export/shop" is copied to the workspace folder "acme"
    And the workspace file "acme/environments/local.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JX0000000000000000000101
      name: local
      variables:
        - key: base_url
          value: https://local.acme.test
        - key: tenant
          value: acme
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
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:01JX0000000000000000000101", "key": "base_url", "value": "http://127.0.0.1:8080" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:01JX0000000000000000000101", "key": "api_key", "value": "current-SECRET-key" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "collection:01JX0000000000000000000001", "key": "page_size", "value": "50" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "region", "value": "us-east-2" }
      """

  @environment @file
  Scenario: An environment file carries the current values as initial values
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "01JX0000000000000000000101",
        "format": "xenon_yaml", "shape": "file", "current_values_as_initial": true,
        "target_path": "{{temp_root}}/local.yaml"
      }
      """
    Then the call should succeed
    And the file "local.yaml" should contain the field "variables.0.value" equal to "http://127.0.0.1:8080"
    And the file "local.yaml" should contain the field "variables.1.value" equal to "acme"
    And the file "local.yaml" should contain the field "variables.2.key" equal to "api_key"
    And the file "local.yaml" should not contain the field "variables.2.value"
    And the file "local.yaml" should not contain "current-SECRET-key"
    And the workspace file "acme/environments/local.yaml" should be unchanged

  @workspace @zip
  Scenario: A workspace zip carries current values for environments, collections and globals
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "workspace", "current_values_as_initial": true,
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/acme.zip"
      }
      """
    Then the call should succeed
    And the file "acme.zip" should not contain "current-SECRET-key"
    When the zip file "acme.zip" is extracted to "unzipped"
    Then the file "unzipped/xenon-export.yaml" should contain the field "currentValuesAsInitial" equal to "true"
    And the file "unzipped/acme/environments/local.yaml" should contain the field "variables.0.value" equal to "http://127.0.0.1:8080"
    And the file "unzipped/acme/workspace.yaml" should contain the field "variables.0.value" equal to "us-east-2"
    And the file "unzipped/acme/collections/shop-api/collection.yaml" should contain the field "variables.1.key" equal to "page_size"
    And the file "unzipped/acme/collections/shop-api/collection.yaml" should contain the field "variables.1.value" equal to "50"
    And the file "unzipped/acme/collections/shop-api/collection.yaml" should contain the field "variables.0.value" equal to "t-123"
    And the file "unzipped/acme/collections/shop-api/ping.yaml" should equal the file "workspaces/acme/collections/shop-api/ping.yaml"
    And the workspace file "acme/environments/local.yaml" should be unchanged
    And the workspace file "acme/workspace.yaml" should be unchanged
    And the workspace file "acme/collections/shop-api/collection.yaml" should contain the field "variables.1.value" equal to "20"

  @collection @bundle
  Scenario: A collection bundle carries the collection's current values
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "01JX0000000000000000000001", "current_values_as_initial": true,
        "format": "xenon_yaml", "shape": "bundle", "target_path": "{{temp_root}}/shop-api.yaml"
      }
      """
    Then the call should succeed
    And the file "shop-api.yaml" should contain the field "item.variables.1.value" equal to "50"
    And the file "shop-api.yaml" should not contain the field "item.variables.3.value"
    And the workspace file "acme/collections/shop-api/collection.yaml" should contain the field "variables.1.value" equal to "20"

  @without_option
  Scenario: Without the option the variable files are exported unchanged
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "workspace",
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/acme.zip"
      }
      """
    Then the call should succeed
    And the file "acme.zip" should not contain "127.0.0.1:8080"
    And the file "acme.zip" should not contain "us-east-2"
    When the zip file "acme.zip" is extracted to "unzipped"
    Then the file "unzipped/acme/environments/local.yaml" should equal the file "workspaces/acme/environments/local.yaml"
    And the file "unzipped/acme/workspace.yaml" should equal the file "workspaces/acme/workspace.yaml"
    And the file "unzipped/acme/collections/shop-api/collection.yaml" should equal the file "workspaces/acme/collections/shop-api/collection.yaml"
