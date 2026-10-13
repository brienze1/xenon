#language: en
#utf-8

@all @xen_003 @import @large_collection @performance
Feature: Import a large collection quickly with steady progress
  A Postman collection of 1,000 requests imports in under 10 seconds in total (detect 2 s + import
  8 s), and the import:progress event keeps coming at least every 250 ms while it is written, also
  while one long item (the whole collection) is being written, so the dialog's progress bar never
  looks stuck (SC-002, FR-019).

  Background:
    Given the workspace "Acme" exists

  @success
  Scenario: A 1,000-request collection is imported in under 10 seconds with progress at most 250 ms apart
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/large-1000.json"] }
      """
    Then the call should succeed
    And the last call should have taken less than 2000 ms
    And the response should contain the field "items.0.counts.requests" equal to "1000"
    And the response should contain the field "items.0.counts.folders" equal to "10"
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the last call should have taken less than 8000 ms
    And the response should contain the field "workspaces.0.created.requests" equal to "1000"
    And the response should contain the field "failed" array length equal to 0
    And the "import:progress" events should have been emitted at most 250 ms apart
    And the 1 "import:progress" event should contain the field "current_item" equal to "Large API"
    And the workspace file "acme/collections/large-api/area-09/request-0999.yaml" should contain the field "method" equal to "POST"
    And the workspace file "acme/collections/large-api/area-09/request-0999.yaml" should contain the field "url" equal to "{{base_url}}/area/9/items/999?page=1"
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.children" array length equal to 10
    And the response should contain the field "0.children.9.children" array length equal to 100
