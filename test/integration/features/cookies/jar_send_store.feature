#language: en
#utf-8

@all @xen_005 @cookies @jar_send_store
Feature: The workspace cookie jar stores and sends cookies like a browser
  Every workspace has a cookie jar. Cookies set by responses are stored in it and sent with later
  requests whose domain and path match, as an auto-generated Cookie header that can be switched
  off. A manual Cookie header is merged with the jar cookies, as in Postman. The send result lists
  the cookies the response set (the response Cookies tab), and every change to the jar is
  announced with cookies:changed.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the requests to "/login" respond with the header "Set-Cookie" equal to "session=abc; Path=/; HttpOnly"
    And the requests to "/login" respond with the header "Set-Cookie" equal to "admin=yes; Path=/admin; Max-Age=3600; Secure"

  @store
  Scenario: Cookies set by a response are stored and listed in the send result
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    And the response should contain the field "cookies_set" array length equal to 2
    And the response field "cookies_set" should contain an item with "name" equal to "session" and "value" equal to "abc"
    And the response field "cookies_set" should contain an item with "name" equal to "session" and "http_only" equal to "true"
    And the response field "cookies_set" should contain an item with "name" equal to "session" and "path" equal to "/"
    And the response field "cookies_set" should contain an item with "name" equal to "session" and "domain" equal to "127.0.0.1"
    And the response field "cookies_set" should contain an item with "name" equal to "session" and "expires" equal to ""
    And the response field "cookies_set" should contain an item with "name" equal to "admin" and "secure" equal to "true"
    And the response field "cookies_set" should contain an item with "name" equal to "admin" and "path" equal to "/admin"
    And the event "cookies:changed" should have been emitted 1 times
    And the 1 "cookies:changed" event should contain the field "local_id" equal to "{{Acme.local_id}}"
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 1
    And the response should contain the field "0.domain" equal to "127.0.0.1"
    And the response should contain the field "0.cookies" array length equal to 2
    And the response field "0.cookies" should contain an item with "name" equal to "session" and "host_only" equal to "true"
    And the response field "0.cookies" should contain an item with "name" equal to "admin" and "value" equal to "yes"

  @send
  Scenario: Stored cookies are sent to matching domains and paths only
    Given I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login" } }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/profile" } }
      """
    Then the call should succeed
    And the response should contain the field "cookies_set" array length equal to 0
    And the 1 "GET" request to "/profile" should have the header "Cookie" equal to "session=abc"
    And the event "cookies:changed" should have been emitted 1 times
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/admin/users" } }
      """
    Then the 1 "GET" request to "/admin/users" should not have the header "Cookie"

  @send @variables
  Scenario: The URL is resolved before the jar is asked for cookies
    Given I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login" } }
      """
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JN0000000000000000000000
      name: Acme
      variables:
        - key: base
          value: "{{api}}"
          enabled: true
          description: ""
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{base}}/profile" } }
      """
    Then the 1 "GET" request to "/profile" should have the header "Cookie" equal to "session=abc"

  @auto_header
  Scenario: The jar cookies show as an auto-generated Cookie header that can be switched off
    Given I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login" } }
      """
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JN0000000000000000000000
      name: Acme
      variables:
        - key: base
          value: "{{api}}"
          enabled: true
          description: ""
      """
    When I call "RequestService.AutoHeaders" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "{{base}}/profile" } }
      """
    Then the call should succeed
    And the response field "" should contain an item with "key" equal to "Cookie" and "value" equal to "session=abc"
    And the response field "" should contain an item with "key" equal to "Cookie" and "disabled" equal to "false"
    When I call "RequestService.AutoHeaders" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "{{api}}/profile", "auto_headers": { "disabled": ["cookie"] } } }
      """
    Then the response field "" should contain an item with "key" equal to "Cookie" and "disabled" equal to "true"
    When I call "RequestService.AutoHeaders" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "https://elsewhere.test/profile" } }
      """
    Then the response field "" should contain 0 items with "key" equal to "Cookie"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/profile", "auto_headers": { "disabled": ["Cookie"] } } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/profile" should not have the header "Cookie"

  @merge
  Scenario: A manual Cookie header is merged with the jar cookies
    Given I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login" } }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/profile", "headers": [ { "key": "Cookie", "value": "theme=dark", "enabled": true } ] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/profile" should have the header "Cookie" equal to "theme=dark; session=abc"

  @expire
  Scenario: A response that expires a cookie removes it from the jar
    Given I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login" } }
      """
    And the requests to "/logout" respond with the header "Set-Cookie" equal to "session=; Path=/; Max-Age=0"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/logout" } }
      """
    Then the call should succeed
    And the event "cookies:changed" should have been emitted 2 times
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/profile" } }
      """
    Then the 1 "GET" request to "/profile" should not have the header "Cookie"
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.cookies" array length equal to 1
    And the response should contain the field "0.cookies.0.name" equal to "admin"

  @nested
  Scenario: Requests sent from scripts use the jar too
    Given I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login" } }
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/profile",
          "scripts": { "pre_request": "await pm.sendRequest(pm.variables.replaceIn('{{api}}/nested'))" }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    And the 1 "GET" request to "/nested" should have the header "Cookie" equal to "session=abc"
