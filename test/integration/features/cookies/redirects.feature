#language: en
#utf-8

@all @xen_005 @cookies @redirects
Feature: Cookies set along a redirect chain are all stored
  When Xenon follows redirects, every hop's Set-Cookie headers go into the workspace jar and the
  following hops already carry them, as in a browser. The send result lists the cookies the final
  response set.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the requests to "/start" redirect to "/middle" with status 302
    And the requests to "/start" respond with the header "Set-Cookie" equal to "step=1; Path=/"
    And the requests to "/middle" redirect to "/end" with status 303
    And the requests to "/middle" respond with the header "Set-Cookie" equal to "hop=2; Path=/"
    And the requests to "/end" respond with the header "Set-Cookie" equal to "final=3; Path=/"

  @follow
  Scenario: Every hop's cookies are stored and sent to the next hops
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/start" } }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "200"
    And the 1 "GET" request to "/middle" should have the header "Cookie" equal to "step=1"
    And the 1 "GET" request to "/end" should have the header "Cookie" equal to "step=1; hop=2"
    And the response should contain the field "cookies_set" array length equal to 1
    And the response should contain the field "cookies_set.0.name" equal to "final"
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.cookies" array length equal to 3
    And the response field "0.cookies" should contain an item with "name" equal to "step" and "value" equal to "1"
    And the response field "0.cookies" should contain an item with "name" equal to "hop" and "value" equal to "2"
    And the response field "0.cookies" should contain an item with "name" equal to "final" and "value" equal to "3"

  @no_follow
  Scenario: Without following redirects only the first response's cookies are stored
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/start", "settings": { "follow_redirects": false } } }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "302"
    And the response should contain the field "cookies_set.0.name" equal to "step"
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.cookies" array length equal to 1
    And the response should contain the field "0.cookies.0.name" equal to "step"
