#language: en
#utf-8

@all @xen_003 @import @atomicity
Feature: Imports are atomic per item, cancellable before writing and point at real items
  Each imported collection is written to a staging folder inside the workspace and moved into
  place in one step. If writing fails, nothing of that collection remains and the other items
  still import. An import cancelled before writing starts writes nothing. Every warning in the
  summary points at an item that exists after the import.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Globex" exists

  @failure
  Scenario: A write failure leaves no trace of that collection while the other items succeed
    Given the workspace path "acme/collections" is read-only
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/full-v2.1.json", "{{resources}}/import/postman/dump/collection/22222222-bbbb-4000-8000-000000000002.json"] }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 2
    And the response should contain the field "items.1.name" equal to "CRM API"
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      {
        "source_id": "{{source}}",
        "items": [
          { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" },
          { "key": "1", "include": true, "target_local_id": "{{Globex.local_id}}" }
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "failed" array length equal to 1
    And the response should contain the field "failed.0.item_key" equal to "0"
    And the response should contain the field "failed.0.name" equal to "Acme API"
    And the response should contain the field "items.0.status" equal to "failed"
    And the response should contain the field "items.1.status" equal to "imported"
    And the workspace folder "acme/collections" should contain the entries ""
    And the workspace folder "acme" should contain the entries ".gitignore, collections, environments, workspace.yaml"
    And the workspace file "acme/workspace.yaml" should contain the field "order" with the items ""
    And the workspace file "globex/collections/crm-api/collection.yaml" should contain the field "name" equal to "CRM API"
    And the workspace folder "globex" should contain the entries ".gitignore, collections, environments, workspace.yaml"

  @cancel
  Scenario: An import cancelled before writing starts writes nothing
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/full-v2.1.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" already cancelled with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should fail with code "IMP-10499"
    And the workspace folder "acme/collections" should contain the entries ""
    And the event "import:progress" should have been emitted 0 times
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace folder "acme/collections" should contain the entries "acme-api"

  @excluded
  Scenario: Items that are not included are not written
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/full-v2.1.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": false, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 0
    And the workspace folder "acme/collections" should contain the entries ""

  @links
  Scenario: Every warning in the summary points at an existing item
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/full-v2.1.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    And I save the response field "warnings" as "warnings"
    And I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save every "id" found in the response as "tree_ids"
    Then every item of the saved "warnings" should have the field "item_id" in "tree_ids"
    And every item of the saved "warnings" should have the field "local_id" in "Acme.local_id"

  @unknown_target
  Scenario: An unknown target workspace is refused before anything is written
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/full-v2.1.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "01JA2Z5V3B6C7D8E9F0G1H2J3K" } ] }
      """
    Then the call should fail with code "WKS-07404"
    And the workspace folder "acme/collections" should contain the entries ""
