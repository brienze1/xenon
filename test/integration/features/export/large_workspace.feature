#language: en
#utf-8

@all @xen_006 @export @performance @large_export
Feature: A large workspace exports quickly
  Exporting a workspace of 1,000 requests across 4 collections takes under 5 seconds in every
  format (SC-005), and the export holds every request.

  Background:
    Given the workspace "Big" exists
    And the workspace folder "big" holds 1000 generated requests across 4 collections

  @workspace
  Scenario: A 1,000-request workspace exports as Xenon YAML in under 5 seconds
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Big.local_id}}", "item_kind": "workspace",
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/big.zip"
      }
      """
    Then the call should succeed
    And the last call should have taken less than 5000 ms
    And the response should contain the field "counts.collections" equal to "4"
    And the response should contain the field "counts.requests" equal to "1000"
    And the zip file "big.zip" is extracted to "unzipped"
    And the file "unzipped/big/collections/collection-4/request-0999.yaml" should contain the field "name" equal to "Request 0999"

  @workspace @postman
  Scenario: A 1,000-request workspace exports as Postman in under 5 seconds
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Big.local_id}}", "item_kind": "workspace",
        "format": "postman", "shape": "zip", "target_path": "{{temp_root}}/big-postman.zip"
      }
      """
    Then the call should succeed
    And the last call should have taken less than 5000 ms
    And the response should contain the field "counts.requests" equal to "1000"
    And the zip file "big-postman.zip" is extracted to "unzipped"
    And the file "unzipped/collections/collection-4.postman_collection.json" should contain the field "item" array length equal to 250
    And the file "unzipped/collections/collection-4.postman_collection.json" should be a valid Postman v2.1 collection
