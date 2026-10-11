#language: en
#utf-8

@all @xen_006 @export @postman_export @postman_workspace_zip
Feature: Export the whole workspace as a zip of Postman files
  File → Export as Postman writes one zip holding a collection JSON per collection under
  collections/, an environment JSON per environment under environments/ and the workspace globals
  as globals.postman_globals.json. Entry names are portable (forward slashes, no absolute paths,
  safe unique file names). No secret value is in the zip.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the resource folder "export/shop" is copied to the workspace folder "acme"
    And the workspace file "acme/collections/zz-shop/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JX0000000000000000000201
      name: Shop-API
      """
    And the workspace file "acme/environments/local.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JX0000000000000000000101
      name: local
      variables:
        - key: base_url
          value: https://local.acme.test
        - key: api_key
          secret: true
      """
    And the workspace file "acme/environments/broken.yaml" contains
      """
      xenon: 1
      kind: environment
      variables: [
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:01JX0000000000000000000101", "key": "api_key", "value": "zip-SECRET-key" }
      """

  @zip
  Scenario: The zip holds every collection, environment and the globals
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "workspace",
        "format": "postman", "shape": "zip", "target_path": "{{temp_root}}/acme-postman.zip"
      }
      """
    Then the call should succeed
    And the response should contain the field "counts.collections" equal to "2"
    And the response should contain the field "counts.requests" equal to "9"
    And the response should contain the field "counts.environments" equal to "1"
    And the zip file "acme-postman.zip" should contain exactly the entries "collections/shop-api.postman_collection.json, collections/shop-api-2.postman_collection.json, environments/local.postman_environment.json, globals.postman_globals.json"
    And the file "acme-postman.zip" should not contain "zip-SECRET-key"
    And the last "export:progress" event should contain the field "done" equal to "4"
    And the last "export:progress" event should contain the field "total" equal to "4"
    When the zip file "acme-postman.zip" is extracted to "unzipped"
    Then the file "unzipped/collections/shop-api.postman_collection.json" should be a valid Postman v2.1 collection
    And the file "unzipped/collections/shop-api.postman_collection.json" should contain the field "info.name" equal to "Shop API"
    And the file "unzipped/collections/shop-api-2.postman_collection.json" should be a valid Postman v2.1 collection
    And the file "unzipped/collections/shop-api-2.postman_collection.json" should contain the field "info.name" equal to "Shop-API"
    And the file "unzipped/environments/local.postman_environment.json" should contain the field "values.1.type" equal to "secret"
    And the file "unzipped/environments/local.postman_environment.json" should contain the field "values.1.value" equal to ""
    And the file "unzipped/globals.postman_globals.json" should contain the field "_postman_variable_scope" equal to "globals"

  @zip @dialog
  Scenario: The export dialog offers a zip for the workspace
    When I call "ExportService.Options" with
      """
      { "local_id": "{{Acme.local_id}}", "item_kind": "workspace" }
      """
    Then the call should succeed
    And the response should contain the field "suggested_file_name" equal to "acme"
    And the response should contain the field "has_secrets" equal to "true"
    And the response should contain the field "formats.0.shapes" with the items "zip"
    And the response should contain the field "formats.0.suggested_file_name" equal to "acme.zip"
    And the response should contain the field "formats.1.shapes" with the items "zip"
    And the response should contain the field "formats.1.suggested_file_name" equal to "acme-postman.zip"

  @zip @atomic
  Scenario: An existing file is replaced and no temp file is left beside it
    Given the file "acme-postman.zip" exists with content
      """
      old
      """
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "workspace",
        "format": "postman", "shape": "zip", "target_path": "{{temp_root}}/acme-postman.zip"
      }
      """
    Then the call should succeed
    And the zip file "acme-postman.zip" should contain exactly the entries "collections/shop-api.postman_collection.json, collections/shop-api-2.postman_collection.json, environments/local.postman_environment.json, globals.postman_globals.json"
    And the file ".acme-postman.zip.xenon-tmp" should not exist

  @errors
  Scenario: A target folder that does not exist fails without leaving a file
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "workspace",
        "format": "postman", "shape": "zip", "target_path": "{{temp_root}}/missing/acme-postman.zip"
      }
      """
    Then the call should fail with code "EXP-02500"
    And the file "missing/acme-postman.zip" should not exist
