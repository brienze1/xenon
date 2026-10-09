#language: en
#utf-8

@all @xen_001 @dialogs @pick_folder
Feature: Native folder picker
  The UI asks Go to show the native folder picker (for example to choose where a new
  workspace lives); an empty path means the user cancelled.

  @success
  Scenario: The chosen folder is returned
    Given the folder picker will return "{{workspace_root}}/acme"
    When I call "DialogService.PickFolder" with
      """
      { "title": "Choose the workspace folder" }
      """
    Then the call should succeed
    And the response should contain the path field "path" equal to "{{workspace_root}}/acme"
    And the folder picker should have been opened with the title "Choose the workspace folder"

  @cancelled
  Scenario: Cancelling the picker returns an empty path
    When I call "DialogService.PickFolder" with
      """
      { "title": "Choose the workspace folder" }
      """
    Then the call should succeed
    And the response should contain the field "path" equal to ""
