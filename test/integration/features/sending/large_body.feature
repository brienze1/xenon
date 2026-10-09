#language: en
#utf-8

@all @xen_001 @sending @performance @large_body
Feature: A 50 MB response is ready to show quickly
  A 50 MB body is kept out of the JSON bridge and served with its exact bytes by the in-app
  response route in under 2 seconds once it was received (SC-003).

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the 1 "GET" request to "/huge" returns status 200 with content type "text/plain" and a generated body of 52428800 bytes

  @route
  Scenario: The route serves a 50 MB body in under 2 seconds
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-huge", "request": { "method": "GET", "url": "{{api}}/huge" } }
      """
    Then the call should succeed
    And the response should contain the field "size_bytes" equal to "52428800"
    And the response field "inline_body" should not exist
    When I request "/xenon/response/tab-huge" from the "ResponseBodyRoute" route
    Then the last call should have taken less than 2000 ms
    And the route response status should be 200
    And the route response body should be the generated body of 52428800 bytes
