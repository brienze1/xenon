#language: en
#utf-8

@all @xen_006 @export @postman_export @roundtrip_corpus
Feature: Postman collections and environments come back out of Xenon unchanged
  Every Postman file of the stage-3 import corpus is imported into a workspace and exported back as
  Postman. The export is a valid Postman collection v2.1 and equals the original under the FR-014
  normalisation rules (export-formats §4): key order ignored, Postman defaults equal to absent keys,
  urls compared by raw, query and variables, the exporter stamps and secret values left out. A v2.0
  collection comes back as the same collection in v2.1. OAuth tokens typed into a collection were
  moved to the vault on import and are never exported.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available

  @collection
  Scenario Outline: The Postman collection <fixture> round-trips
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/<fixture>"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    And I save the response field "items.0.id" as "collection"
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{collection}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/roundtrip.postman_collection.json"
      }
      """
    Then the call should succeed
    And the response should contain the field "path" equal to "{{temp_root}}/roundtrip.postman_collection.json"
    And the response should contain the field "secrets" equal to "names_only"
    And the file "roundtrip.postman_collection.json" should be a valid Postman v2.1 collection
    And the file "roundtrip.postman_collection.json" should equal the Postman resource file "import/postman/<expected>" after normalisation

    Examples:
      | fixture                                                     | expected                                                    |
      | full-v2.1.json                                              | full-v2.1.json                                              |
      | full-v2.0.json                                              | full-v2.1.json                                              |
      | oauth2-token.json                                           | oauth2-token.json                                           |
      | large-1000.json                                             | large-1000.json                                             |
      | dump/collection/11111111-aaaa-4000-8000-000000000001.json | dump/collection/11111111-aaaa-4000-8000-000000000001.json |
      | dump/collection/22222222-bbbb-4000-8000-000000000002.json | dump/collection/22222222-bbbb-4000-8000-000000000002.json |
      | dump/collection/33333333-cccc-4000-8000-000000000003.json | dump/collection/33333333-cccc-4000-8000-000000000003.json |

  @collection @tokens
  Scenario: A token typed into an OAuth 2.0 collection is not exported
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/oauth2-token.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    And I save the response field "items.0.id" as "collection"
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{collection}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/token.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "token.postman_collection.json" should not contain "imported-t0ken-value"
    And the file "token.postman_collection.json" should contain the field "auth.type" equal to "oauth2"
    And the file "token.postman_collection.json" should contain the field "item.0.request.auth.type" equal to "oauth2"

  @environment
  Scenario Outline: The Postman environment <fixture> round-trips without its secret values
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/<fixture>"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    And I save the response field "items.0.id" as "environment"
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "{{environment}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/roundtrip.postman_environment.json"
      }
      """
    Then the call should succeed
    And the file "roundtrip.postman_environment.json" should equal the Postman resource file "import/postman/<fixture>" after normalisation
    And the file "roundtrip.postman_environment.json" should not contain "<secret>"

    Examples:
      | fixture                                                      | secret                 |
      | dev.postman_environment.json                                 | dev-api-key-SECRET-123 |
      | dump/environment/44444444-dddd-4000-8000-000000000004.json | tok-billing-SECRET     |
      | dump/environment/55555555-eeee-4000-8000-000000000005.json | SECRET                 |

  @globals
  Scenario: Postman globals round-trip
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/globals.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "globals",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/roundtrip.postman_globals.json"
      }
      """
    Then the call should succeed
    And the file "roundtrip.postman_globals.json" should equal the Postman resource file "import/postman/globals.json" after normalisation
