#language: en
#utf-8

@all @xen_006 @export @atomic_cancel
Feature: Exports are atomic, cancellable and report their progress
  An export is written to a temp file beside the target and renamed into place only when it is
  complete (FR-018). Cancelling it part-way, or a failure while writing, leaves no file behind and
  an existing file at the target untouched; a write failure reports EXP-02500. An export of more
  than 200 files announces export:progress every 50 files while it writes, so the dialog's
  progress bar moves; a small export only announces its start and its end.

  Background:
    Given the workspace "Big" exists
    And the workspace folder "big" holds 1000 generated requests across 4 collections

  @progress
  Scenario Outline: A large export announces its progress every 50 files (<format>)
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Big.local_id}}", "item_kind": "workspace",
        "format": "<format>", "shape": "zip", "target_path": "{{temp_root}}/big.zip"
      }
      """
    Then the call should succeed
    And the file "big.zip" should exist
    And the event "export:progress" should have been emitted <events> times
    And the 1 "export:progress" event should contain the field "done" equal to "0"
    And the 1 "export:progress" event should contain the field "total" equal to "<total>"
    And the 2 "export:progress" event should contain the field "done" equal to "50"
    And the 3 "export:progress" event should contain the field "done" equal to "100"
    And the last "export:progress" event should contain the field "done" equal to "<total>"
    And the last "export:progress" event should contain the field "total" equal to "<total>"

    Examples:
      | format     | total | events |
      | xenon_yaml | 1007  | 22     |

  @progress @small
  Scenario: A small export only announces its start and its end
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Big.local_id}}", "item_kind": "workspace",
        "format": "postman", "shape": "zip", "target_path": "{{temp_root}}/big-postman.zip"
      }
      """
    Then the call should succeed
    And the event "export:progress" should have been emitted 2 times
    And the 1 "export:progress" event should contain the field "done" equal to "0"
    And the last "export:progress" event should contain the field "done" equal to "5"
    And the last "export:progress" event should contain the field "total" equal to "5"

  @cancel
  Scenario: Cancelling an export part-way leaves no file behind
    When I call "ExportService.Export" and cancel it on the 3 "export:progress" event with
      """
      {
        "local_id": "{{Big.local_id}}", "item_kind": "workspace",
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/big.zip"
      }
      """
    Then the call should fail with code "EXP-05499"
    And the file "big.zip" should not exist
    And the file ".big.zip.xenon-tmp" should not exist
    And the event "export:progress" should have been emitted 3 times
    And the last "export:progress" event should contain the field "done" equal to "100"

  @cancel @overwrite
  Scenario: Cancelling an export part-way keeps the file it would have replaced
    Given the file "big.zip" exists with content
      """
      the previous export
      """
    And the file "previous.zip" exists with content
      """
      the previous export
      """
    When I call "ExportService.Export" and cancel it on the 2 "export:progress" event with
      """
      {
        "local_id": "{{Big.local_id}}", "item_kind": "workspace",
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/big.zip"
      }
      """
    Then the call should fail with code "EXP-05499"
    And the file "big.zip" should equal the file "previous.zip"
    And the file ".big.zip.xenon-tmp" should not exist

  @cancel @before
  Scenario: An export cancelled before it starts writes nothing
    When I call "ExportService.Export" already cancelled with
      """
      {
        "local_id": "{{Big.local_id}}", "item_kind": "collection", "item_id": "01JG0000000000000000000001",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/collection-1.postman_collection.json"
      }
      """
    Then the call should fail with code "EXP-05499"
    And the file "collection-1.postman_collection.json" should not exist
    And the file ".collection-1.postman_collection.json.xenon-tmp" should not exist

  @failure
  Scenario Outline: A write failure leaves no file behind and reports EXP-02500 (<format>)
    Given the file "taken.zip/inside.txt" exists with content
      """
      a folder where the export should go
      """
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Big.local_id}}", "item_kind": "workspace",
        "format": "<format>", "shape": "zip", "target_path": "{{temp_root}}/taken.zip"
      }
      """
    Then the call should fail with code "EXP-02500"
    And the file ".taken.zip.xenon-tmp" should not exist
    And the file "taken.zip/inside.txt" should exist

    Examples:
      | format     |
      | xenon_yaml |
      | postman    |
