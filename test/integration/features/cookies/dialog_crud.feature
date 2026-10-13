#language: en
#utf-8

@all @xen_005 @cookies @dialog_crud
Feature: Manage cookies in the Cookies dialog
  The Cookies dialog lists the workspace's cookies grouped by domain. A cookie is added or edited
  as a Set-Cookie line for a domain (a Domain attribute makes it a domain cookie, otherwise it is
  host-only), single cookies and whole domains can be deleted and the jar cleared. Invalid cookies
  are refused with the reason. Every change is announced with cookies:changed and applies to the
  next request.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url

  @create @list
  Scenario: Cookies added in the dialog are listed by domain with their attributes
    When I call "CookieService.Upsert" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "example.test", "raw": "token=t1; Path=/api; Expires=Wed, 21 Oct 2037 07:28:00 GMT; Secure; HttpOnly; SameSite=Lax" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "token"
    And the response should contain the field "value" equal to "t1"
    And the response should contain the field "domain" equal to "example.test"
    And the response should contain the field "path" equal to "/api"
    And the response should contain the field "expires" equal to "2037-10-21T07:28:00Z"
    And the response should contain the field "secure" equal to "true"
    And the response should contain the field "http_only" equal to "true"
    And the response should contain the field "same_site" equal to "Lax"
    And the response should contain the field "host_only" equal to "true"
    And the response should contain the field "raw" equal to "token=t1; Path=/api; Expires=Wed, 21 Oct 2037 07:28:00 GMT; HttpOnly; Secure; SameSite=Lax"
    And the event "cookies:changed" should have been emitted 1 times
    When I call "CookieService.Upsert" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "api.example.test", "raw": "lang=en; Domain=example.test" }
      """
    Then the response should contain the field "domain" equal to "example.test"
    And the response should contain the field "host_only" equal to "false"
    And the response should contain the field "path" equal to "/"
    And the response should contain the field "expires" equal to ""
    When I call "CookieService.Upsert" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "another.test", "raw": "a=1" }
      """
    And I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "0.domain" equal to "another.test"
    And the response should contain the field "1.domain" equal to "example.test"
    And the response should contain the field "1.cookies" array length equal to 2
    And the response should contain the field "1.cookies.0.name" equal to "lang"
    And the response should contain the field "1.cookies.1.name" equal to "token"

  @edit
  Scenario: Editing a cookie replaces it, also when its name or path changes
    Given I call "CookieService.Upsert" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "example.test", "raw": "token=t1; Path=/" }
      """
    When I call "CookieService.Upsert" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "example.test", "raw": "token=t2; Path=/" }
      """
    Then the response should contain the field "value" equal to "t2"
    When I call "CookieService.Upsert" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "example.test", "raw": "auth=t3; Path=/v2", "replaces": { "name": "token", "path": "/" } }
      """
    Then the call should succeed
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.cookies" array length equal to 1
    And the response should contain the field "0.cookies.0.name" equal to "auth"
    And the response should contain the field "0.cookies.0.value" equal to "t3"
    And the response should contain the field "0.cookies.0.path" equal to "/v2"

  @send
  Scenario: A cookie added for the API's domain is sent with the next request
    Given I call "CookieService.Upsert" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "127.0.0.1", "raw": "manual=yes" }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/profile" } }
      """
    Then the 1 "GET" request to "/profile" should have the header "Cookie" equal to "manual=yes"

  @delete
  Scenario: Cookies, domains and the whole jar can be deleted
    Given I call "CookieService.Upsert" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "example.test", "raw": "a=1; Path=/" }
      """
    And I call "CookieService.Upsert" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "example.test", "raw": "b=2; Path=/docs" }
      """
    And I call "CookieService.Upsert" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "other.test", "raw": "c=3" }
      """
    And I call "CookieService.Upsert" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "third.test", "raw": "d=4" }
      """
    When I call "CookieService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "example.test", "path": "/docs", "name": "b" }
      """
    Then the call should succeed
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.cookies" array length equal to 1
    And the response should contain the field "0.cookies.0.name" equal to "a"
    When I call "CookieService.DeleteDomain" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "other.test" }
      """
    And I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "1.domain" equal to "third.test"
    When I call "CookieService.Clear" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0
    And the event "cookies:changed" should have been emitted 7 times

  @invalid
  Scenario Outline: Invalid cookies are refused - <case>
    When I call "CookieService.Upsert" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "<domain>", "raw": "<raw>" }
      """
    Then the call should fail with code "<code>"
    And the event "cookies:changed" should have been emitted 0 times

    Examples:
      | case                     | domain       | raw                                    | code      |
      | no name                  | example.test | =value                                 | CKE-00400 |
      | not a cookie             | example.test | just text                              | CKE-00400 |
      | public suffix domain     | example.test | a=1; Domain=com                        | CKE-00400 |
      | domain of another site   | example.test | a=1; Domain=other.test                 | CKE-00400 |
      | already expired          | example.test | a=1; Expires=Thu, 01 Jan 2015 00:00:00 GMT | CKE-00400 |
      | blank line               | example.test |                                        | CKE-02400 |
      | blank domain             |              | a=1                                    | CKE-02400 |

  @validation
  Scenario Outline: Invalid input is rejected - <method>
    When I call "CookieService.<method>" with
      """
      <payload>
      """
    Then the call should fail with code "<code>"

    Examples:
      | method        | payload                                                      | code      |
      | List          | { "local_id": "" }                                           | CKE-01400 |
      | Delete        | { "local_id": "{{Acme.local_id}}", "domain": "", "path": "/", "name": "a" } | CKE-03400 |
      | DeleteDomain  | { "local_id": "{{Acme.local_id}}", "domain": " " }            | CKE-04400 |
      | Clear         | { "local_id": "" }                                           | CKE-05400 |
      | GetAllowlist  | { "local_id": "" }                                           | CKE-06400 |
      | SaveAllowlist | { "local_id": "", "domains": [] }                            | CKE-07400 |
