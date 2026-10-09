#language: en
#utf-8

@all @xen_001 @tree @performance @large_workspace
Feature: A large workspace stays fast
  A workspace with 1,000 requests across 4 collections opens and shows its tree in under
  2 seconds, and searching it answers in under 200 ms per keystroke (SC-002).

  Background:
    Given the workspace "Big" exists
    And the workspace folder "big" holds 1000 generated requests across 4 collections

  @open
  Scenario: Opening the workspace returns the whole tree in under 2 seconds
    When I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Big.local_id}}" }
      """
    Then the call should succeed
    And the last call should have taken less than 2000 ms
    And the response should contain the field "tree" array length equal to 4
    And the response should contain the field "tree.0.children" array length equal to 250
    And the response should contain the field "tree.3.children.249.name" equal to "Request 0999"
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Big.local_id}}" }
      """
    Then the call should succeed
    And the last call should have taken less than 2000 ms
    And the response should contain the field "" array length equal to 4

  @search
  Scenario Outline: Every keystroke of a search answers in under 200 ms - <query>
    Given I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Big.local_id}}" }
      """
    When I call "TreeService.Search" with
      """
      { "local_id": "{{Big.local_id}}", "query": "<query>" }
      """
    Then the call should succeed
    And the last call should have taken less than 200 ms
    And the response should contain the field "match_ids" array length equal to <matches>

    Examples:
      | query           | matches |
      | r               | 1000    |
      | request 09      | 100     |
      | request 099     | 10      |
      | items/999       | 1       |
      | nothing-matches | 0       |
