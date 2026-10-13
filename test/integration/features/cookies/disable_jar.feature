#language: en
#utf-8

@all @xen_005 @cookies @disable_jar
Feature: A request can disable the cookie jar
  The request setting "Disable cookie jar" sends no jar cookies and stores none of the cookies
  the response sets; the response still lists them.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the requests to "/login" respond with the header "Set-Cookie" equal to "session=abc; Path=/"

  @store
  Scenario: Cookies of a request with the jar disabled are not stored
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login", "settings": { "disable_cookie_jar": true } } }
      """
    Then the call should succeed
    And the response should contain the field "cookies_set.0.name" equal to "session"
    And the event "cookies:changed" should have been emitted 0 times
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0

  @send
  Scenario: A request with the jar disabled sends no jar cookies and shows no Cookie header
    Given I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login" } }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/profile", "settings": { "disable_cookie_jar": true } } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/profile" should not have the header "Cookie"
    When I call "RequestService.AutoHeaders" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "{{api}}/profile", "settings": { "disable_cookie_jar": true } } }
      """
    Then the response field "" should contain 0 items with "key" equal to "Cookie"
