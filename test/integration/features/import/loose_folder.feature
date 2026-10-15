#language: en
#utf-8

@all @xen_014 @import @xen_014_loose
Feature: Import a folder of standalone export files
  A folder that is none of the recognised layouts (Postman data dump, Bruno collection, git-native Postman v3
  workspace or collection folder, Xenon workspace folder) is read as a folder of standalone exports: each regular,
  non-hidden file at its top level is detected as if its path had been passed on its own, in name order. Subfolders
  and files that are not importable on their own are skipped and named in one skipped_part warning; hidden entries
  are ignored silently. A folder with nothing importable is refused with the folder layouts Xenon accepts. A detected
  batch reports the kind and version all its inputs share, or "mixed" with no version.

  Background:
    Given the workspace "Acme" exists

  @expand
  Scenario: A folder of a collection, an environment and globals is detected as three items of mixed kind
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/loose/mixed"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "mixed"
    And the response should contain the field "version" equal to ""
    And the response should contain the field "items" array length equal to 3
    And the response should contain the field "items.0.kind" equal to "collection"
    And the response should contain the field "items.0.name" equal to "Loose Mixed"
    And the response should contain the field "items.1.kind" equal to "environment"
    And the response should contain the field "items.1.name" equal to "dev"
    And the response should contain the field "items.2.kind" equal to "globals"
    And the response should contain the field "warnings" array length equal to 0

  @expand @run
  Scenario: The detected folder imports like its files would one by one
    Given I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/loose/mixed"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [
        { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" },
        { "key": "1", "include": true, "target_local_id": "{{Acme.local_id}}" },
        { "key": "2", "include": true, "target_local_id": "{{Acme.local_id}}" }
      ] }
      """
    Then the call should succeed
    And the response should contain the field "failed" array length equal to 0
    And the workspace folder "acme/collections" should contain the entries "loose-mixed"
    And the workspace file "acme/environments/dev.yaml" should contain the field "name" equal to "dev"

  @skipped
  Scenario: Subfolders and files that are not exports are skipped and named; hidden files are ignored
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/loose/with-extras"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "mixed"
    And the response should contain the field "items" array length equal to 3
    And the response should contain the field "warnings" array length equal to 1
    And the response should contain the field "warnings.0.code" equal to "skipped_part"
    And the response should contain the field "warnings.0.severity" equal to "warning"
    And the response should contain the field "warnings.0.message" containing "Skipped in with-extras: "
    And the response should contain the field "warnings.0.message" containing "notes.txt (not a Xenon, Postman, Bruno or cURL export)"
    And the response should contain the field "warnings.0.message" containing "sub (folder)"
    And the response should not contain "hidden-copy"
    And the response should not contain "Hidden Copy"

  @nothing
  Scenario: A folder with nothing importable lists the layouts Xenon accepts
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/loose/nothing"] }
      """
    Then the call should fail with code "IMP-55415"
    And the error should contain the field "error.description" equal to "nothing has no file Xenon can import; pass a Postman v3 workspace or collection folder, a Postman data dump (archive.json), a Bruno collection (bruno.json or opencollection.yml), a Xenon workspace folder, or a folder of standalone export files"
    And the workspace folder "acme/collections" should contain the entries ""

  @kind
  Scenario: A folder whose files share one kind and version reports that kind
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/loose/collections-only"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "postman_collection"
    And the response should contain the field "version" equal to "2.1"
    And the response should contain the field "items" array length equal to 2
    And the response should contain the field "items.0.name" equal to "Loose A"
    And the response should contain the field "items.1.name" equal to "Loose B"

  @kind @mixed_paths
  Scenario Outline: The kind of a batch of paths - <case>
    When I call "ImportService.Detect" with
      """
      { "paths": [<paths>] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "<kind>"
    And the response should contain the field "version" equal to "<version>"

    Examples:
      | case                         | paths                                                                                                                 | kind               | version |
      | globals before a collection  | "{{resources}}/import/postman/globals.json", "{{resources}}/import/postman/full-v2.1.json"                          | mixed              |         |
      | v2.1 and v2.0 collections    | "{{resources}}/import/postman/full-v2.1.json", "{{resources}}/import/postman/full-v2.0.json"                        | mixed              |         |
      | two v2.1 collections         | "{{resources}}/import/loose/collections-only/a.json", "{{resources}}/import/loose/collections-only/b.json"          | postman_collection | 2.1     |
      | a loose folder and its kin   | "{{resources}}/import/loose/collections-only", "{{resources}}/import/loose/mixed/collection.json"                   | postman_collection | 2.1     |

  @mcp
  Scenario: AI clients detect a loose folder the same way
    When I call the MCP tool "import" with
      """
      { "action": "detect", "paths": ["{{resources}}/import/loose/mixed"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "mixed"
    And the response should contain the field "items" array length equal to 3
