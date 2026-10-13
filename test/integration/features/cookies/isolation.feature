#language: en
#utf-8

@all @xen_005 @cookies @isolation
Feature: Each workspace has its own cookie jar
  Cookies never cross workspaces: a cookie stored while working in one workspace is neither sent
  nor listed in another.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    And the "api" variable is set to the mocked api url
    And the requests to "/login" respond with the header "Set-Cookie" equal to "session=acme; Path=/"

  @isolation
  Scenario: A cookie stored in one workspace stays there
    Given I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login" } }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Globex.local_id}}", "tab_id": "tab-2", "request": { "method": "GET", "url": "{{api}}/profile" } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/profile" should not have the header "Cookie"
    When I call "CookieService.List" with
      """
      { "local_id": "{{Globex.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.cookies.0.value" equal to "acme"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/profile" } }
      """
    Then the 2 "GET" request to "/profile" should have the header "Cookie" equal to "session=acme"

  @unknown
  Scenario: Cookie calls on an unknown workspace fail
    When I call "CookieService.List" with
      """
      { "local_id": "01JN0000000000000000000099" }
      """
    Then the call should fail with code "WKS-07404"
