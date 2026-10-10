#language: en
#utf-8

@all @xen_004 @scripts @unsupported_apis
Feature: APIs Xenon cannot support fail the way Postman's sandbox fails, or with an error naming them
  pm.require (Postman's package library), pm.mock, pm.state and pm.message do not exist in
  postman-sandbox 6.7.4 either, so scripts see them as undefined and fail with the same
  TypeError. pm.datasets exists in Postman but needs Postman's cloud: calling it fails with
  "pm.datasets is not supported in Xenon". The runtime features postman-sandbox exposes but Xenon
  cannot run (SharedArrayBuffer, Atomics) fail with an error naming the feature. The script error
  is reported with its line like any other error.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JL0000000000000000000001"
    And the "create" variable is set to "01JL0000000000000000000004"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JL0000000000000000000000
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
      """
    And I call "ConsoleService.Clear"

  @pm
  Scenario Outline: <api> fails and stops the send
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: |
          console.log(typeof <api>)
          <call>
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "stopped_by_script_error"
    And the response should contain the field "script_errors.0.kind" equal to "<kind>"
    And the response should contain the field "script_errors.0.message" equal to "<message>"
    And the response should contain the field "script_errors.0.line" equal to "2"
    And the "GET" request to "/invoices" should have been received 0 times
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "0.summary" equal to "<type>"

    Examples:
      | api               | call                        | type      | kind                   | message                                                      |
      | pm.require        | pm.require("@team/lib")     | undefined | script_runtime         | TypeError: pm.require is not a function                      |
      | pm.mock           | pm.mock.get("x")            | undefined | script_runtime         | TypeError: Cannot read properties of undefined (reading 'get') |
      | pm.state          | pm.state.get("x")           | undefined | script_runtime         | TypeError: Cannot read properties of undefined (reading 'get') |
      | pm.message        | pm.message.get("x")         | undefined | script_runtime         | TypeError: Cannot read properties of undefined (reading 'get') |
      | pm.datasets       | pm.datasets("orders")       | function  | script_unsupported_api | Error: pm.datasets is not supported in Xenon                 |
      | SharedArrayBuffer | new SharedArrayBuffer(8)    | function  | script_unsupported_api | Error: SharedArrayBuffer is not supported in Xenon           |
      | Atomics           | Atomics.add(new Int32Array(1), 0, 1) | object | script_unsupported_api | Error: Atomics is not supported in Xenon                   |

  @caught
  Scenario: A script can catch the error and go on
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: |
          try { pm.datasets("orders").executeQuery("select 1") } catch (e) { console.log(e.name, e.message) }
      """
    And the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      {"id":"inv-1"}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "0.summary" equal to "Error pm.datasets is not supported in Xenon"
