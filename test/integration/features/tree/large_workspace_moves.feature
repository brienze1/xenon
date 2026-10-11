#language: en
#utf-8

@all @xen_005 @tree @performance @large_workspace_moves
Feature: Moves and snippets stay fast in a large workspace
  In a workspace with 1,000 requests across 4 collections, a drag-and-drop move answers in under
  200 ms, and generating a snippet for one of its requests answers in under 50 ms.

  Background:
    Given the workspace "Big" exists
    And the workspace folder "big" holds 1000 generated requests across 4 collections
    And I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Big.local_id}}" }
      """

  @move
  Scenario: A request moved to another collection in under 200 ms
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Big.local_id}}", "id": "01JG0000000000000000001000", "new_parent_id": "01JG0000000000000000000002", "index": 10 }
      """
    Then the call should succeed
    And the last call should have taken less than 200 ms
    And the response should contain the field "1.children.10.id" equal to "01JG0000000000000000001000"
    And the workspace file "big/collections/collection-2/request-0000.yaml" should contain the field "name" equal to "Request 0000"
    And the workspace file "big/collections/collection-1/request-0000.yaml" should not exist

  @move @reorder
  Scenario: A request reordered within its collection in under 200 ms
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Big.local_id}}", "id": "01JG0000000000000000001996", "new_parent_id": "01JG0000000000000000000001", "index": 0 }
      """
    Then the call should succeed
    And the last call should have taken less than 200 ms
    And the response should contain the field "0.children.0.id" equal to "01JG0000000000000000001996"

  @snippets
  Scenario Outline: A snippet of a saved request in under 50 ms - <language>
    Given I call "SnippetService.Generate" with
      """
      { "local_id": "{{Big.local_id}}", "language": "<language>", "request": { "id": "01JG0000000000000000001500", "method": "GET", "url": "https://big.test/items/500", "auth": { "type": "inherit" } } }
      """
    When I call "SnippetService.Generate" with
      """
      { "local_id": "{{Big.local_id}}", "language": "<language>", "request": { "id": "01JG0000000000000000001500", "method": "POST", "url": "https://big.test/items/500", "headers": [ { "key": "Accept", "value": "application/json", "enabled": true } ], "auth": { "type": "inherit" }, "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"name\":\"Ana\",\"tags\":[\"a\",\"b\"]}" } } } }
      """
    Then the call should succeed
    And the last call should have taken less than 50 ms
    And the response should contain the field "text" containing "big.test/items/500"

    Examples:
      | language        |
      | curl            |
      | js_fetch        |
      | go_native       |
      | python_requests |
