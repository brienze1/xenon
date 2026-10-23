#language: en
#utf-8

@all @xen_001 @dialogs @pick_file
Feature: Native file pickers
  The request builder asks Go for a file to send (form-data and binary bodies) and the
  response viewer asks where to save a body; an empty path means the user cancelled.

  @success
  Scenario: The chosen file is returned
    Given the file picker will return "{{temp_root}}/files/payload.bin"
    When I call "DialogService.PickFile" with
      """
      { "title": "Choose a file" }
      """
    Then the call should succeed
    And the response should contain the path field "path" equal to "{{temp_root}}/files/payload.bin"
    And the file picker should have been opened with the title "Choose a file"

  @success @xen_017 @file_filters
  Scenario: File filters reach the native picker
    Given the file picker will return "{{temp_root}}/data/users.csv"
    When I call "DialogService.PickFile" with
      """
      { "title": "Choose iteration data", "filters": [{ "display_name": "Data files", "pattern": "*.csv;*.json" }] }
      """
    Then the call should succeed
    And the response should contain the path field "path" equal to "{{temp_root}}/data/users.csv"
    And the file picker should have been opened with the title "Choose iteration data"
    And the file picker should have been opened with the filters "Data files (*.csv;*.json)"

  @success @xen_017 @file_filters
  Scenario: Without filters the picker opens unfiltered, as before
    When I call "DialogService.PickFile" with
      """
      { "title": "Choose a file" }
      """
    Then the call should succeed
    And the file picker should have been opened with the filters ""

  @success
  Scenario: The save dialog suggests a name and returns the chosen path
    Given the save dialog will return "{{temp_root}}/out/response.json"
    When I call "DialogService.PickSaveFile" with
      """
      { "title": "Save response body", "suggested_name": "response.json" }
      """
    Then the call should succeed
    And the response should contain the path field "path" equal to "{{temp_root}}/out/response.json"
    And the save dialog should have been opened with the title "Save response body"
    And the save dialog should have suggested the name "response.json"

  @cancelled
  Scenario Outline: Cancelling a picker returns an empty path - <method>
    When I call "DialogService.<method>" with
      """
      { "title": "Choose", "suggested_name": "x.bin" }
      """
    Then the call should succeed
    And the response should contain the field "path" equal to ""

    Examples:
      | method       |
      | PickFile     |
      | PickSaveFile |
