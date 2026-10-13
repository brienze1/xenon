#language: en
#utf-8

@all @xen_004 @console @console_filters
Feature: Filtering and searching the console
  List narrows the console to Requests (request entries), Logs (console calls) or Errors (script
  errors, failed tests and console.error calls, and requests that failed or got a 4xx/5xx
  status), and to entries whose text, URL or script name contains the search text, ignoring case.
  All (the default) lists everything. The filter must be one of all, requests, logs or errors.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK8200000000000000000001"
    And the "create" variable is set to "01JK8200000000000000000002"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK8200000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      scripts:
        preRequest: console.log("Signing the request")
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: |
          console.warn("slow invoice list")
          console.error("missing field")
          pm.test("has items", () => pm.expect([]).to.not.be.empty)
          await pm.sendRequest("{{api}}/missing")
      """
    And the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      []
      """
    And the 1 "GET" request to "/missing" returns status 404 with the following response
      """
      {}
      """
    And I call "ConsoleService.Clear"
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """

  Scenario Outline: The <filter> filter lists <count> entries
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "<filter>" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to <count>
    And the response should contain the field "0.summary" equal to "<first>"

    Examples:
      | filter   | count | first                                |
      | all      | 6     | Signing the request                  |
      | requests | 2     | GET {{api}}/invoices                 |
      | logs     | 3     | Signing the request                  |
      | errors   | 3     | missing field                        |

  Scenario: The errors filter includes failed tests and requests with an error status
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "errors" }
      """
    Then the response should contain the field "1.kind" equal to "request"
    And the response should contain the field "1.summary" equal to "GET {{api}}/missing"
    And the response should contain the field "2.kind" equal to "test_failure"
    And the response should contain the field "2.summary" equal to "has items: expected [] not to be empty"

  Scenario: Without a filter every entry is listed
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 6

  Scenario Outline: Searching "<search>" lists <count> entries
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "search": "<search>"<filter> }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to <count>

    Examples:
      | search    | filter                 | count |
      | SIGNING   |                        | 1     |
      | invoice   |                        | 2     |
      | /missing  |                        | 1     |
      | billing   |                        | 1     |
      | invoice   | , "filter": "requests" | 1     |
      | nothing   |                        | 0     |

  Scenario: An unknown filter fails validation
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "warnings" }
      """
    Then the call should fail with code "CON-00400"
