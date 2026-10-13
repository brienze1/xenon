#language: en
#utf-8

@all @xen_004 @tests @test_results
Feature: Test results of a send
  Every pm.test in the scripts of a send is listed in the send result, in run order, with its
  name, status (passed, failed or skipped), the assertion message when it failed and the script
  it came from (owner kind, name, phase and a label such as "Billing · Post-response"). Tests in
  pre-request scripts are listed too, also when a pre-request script stops the send, and legacy
  tests["…"] results are collected at the end of their script. The result counts the tests
  (total, passed, failed, skipped) for the Test Results tab title. Failure messages are
  chai-postman's, as postman-sandbox 6.7.4 words them.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK4000000000000000000001"
    And the "invoices" variable is set to "01JK4000000000000000000002"
    And the "create" variable is set to "01JK4000000000000000000003"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK4000000000000000000000
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
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      """
    And the 1 "GET" request to "/invoices" returns status 404 with content type "application/json" and body
      """
      {"error":{"code":"E1"}}
      """

  @assertions @failing
  Scenario Outline: <name> fails with chai-postman's message against a 404 JSON response
    Given the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: |
          pm.test("<name>", () => <assertion>)
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the response should contain the field "tests" array length equal to 1
    And the response should contain the field "tests.0.name" equal to "<name>"
    And the response should contain the field "tests.0.status" equal to "failed"
    And the response should contain the field "tests.0.error_message" equal to "<message>"

    Examples:
      | name            | assertion                                                | message                                                                               |
      | status 200      | pm.response.to.have.status(200)                          | expected response to have status code 200 but got 404                                 |
      | status Created  | pm.response.to.have.status("Created")                    | expected response to have status reason 'Created' but got 'Not Found'                 |
      | ok              | pm.response.to.be.ok                                     | expected response to have status reason 'OK' but got 'NOT FOUND'                      |
      | success         | pm.response.to.be.success                                | expected response code to be 2XX but found 404                                        |
      | serverError     | pm.response.to.be.serverError                            | expected response code to be 5XX but found 404                                        |
      | redirection     | pm.response.to.be.redirection                            | expected response code to be 3XX but found 404                                        |
      | info            | pm.response.to.be.info                                   | expected response code to be 1XX but found 404                                        |
      | badRequest      | pm.response.to.be.badRequest                             | expected response to have status reason 'BAD REQUEST' but got 'NOT FOUND'             |
      | unauthorized    | pm.response.to.be.unauthorized                           | expected response to have status reason 'UNAUTHORIZED' but got 'NOT FOUND'            |
      | forbidden       | pm.response.to.be.forbidden                              | expected response to have status reason 'FORBIDDEN' but got 'NOT FOUND'               |
      | accepted        | pm.response.to.be.accepted                               | expected response to have status reason 'ACCEPTED' but got 'NOT FOUND'                |
      | rateLimited     | pm.response.to.be.rateLimited                            | expected response to have status reason 'TOO MANY REQUESTS' but got 'NOT FOUND'       |
      | withoutContent  | pm.response.to.be.withoutContent                         | expected response to have status reason 'NO CONTENT' but got 'NOT FOUND'              |
      | notAcceptable   | pm.response.to.be.notAcceptable                          | expected response to have status reason 'ACCEPTABLE' but got 'NOT FOUND'              |
      | header          | pm.response.to.have.header("X-Trace")                    | expected response to have header with key 'X-Trace'                                   |
      | header value    | pm.response.to.have.header("Content-Type", "text/plain") | expected 'Content-Type' response header to be 'text/plain' but got 'application/json' |
      | statusCodeClass | pm.response.to.have.statusCodeClass(2)                   | expected response code to be 2XX but found 404                                        |
      | statusReason    | pm.response.to.have.statusReason("OK")                   | expected response to have status reason 'OK' but got 'Not Found'                      |
      | expect eql      | pm.expect(pm.response.json().error.code).to.eql("E2")    | expected 'E1' to deeply equal 'E2'                                                    |

  @assertions @passing
  Scenario Outline: <name> passes against a 404 JSON response
    Given the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: |
          pm.test("<name>", () => <assertion>)
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the response should contain the field "tests" array length equal to 1
    And the response should contain the field "tests.0.name" equal to "<name>"
    And the response should contain the field "tests.0.status" equal to "passed"
    And the response field "tests.0.error_message" should not exist

    Examples:
      | name         | assertion                                      |
      | status 404   | pm.response.to.have.status(404)                |
      | error        | pm.response.to.be.error                        |
      | clientError  | pm.response.to.be.clientError                  |
      | notFound     | pm.response.to.be.notFound                     |
      | json         | pm.response.to.be.json                         |
      | withBody     | pm.response.to.be.withBody                     |
      | responseTime | pm.response.to.have.responseTime.below(100000) |
      | not ok       | pm.response.to.not.be.ok                       |

  @quoted_messages
  Scenario: Body and JSON body assertions quote the response in their messages
    Given the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: |
          pm.test("body", () => pm.response.to.have.body("nope"))
          pm.test("jsonBody path", () => pm.response.to.have.jsonBody("error.code", "E2"))
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "tests.0.status" equal to "failed"
    And the response should contain the field "tests.0.error_message" equal to
      """
      expected response body to equal 'nope' but got '{"error":{"code":"E1"}}'
      """
    And the response should contain the field "tests.1.status" equal to "failed"
    And the response should contain the field "tests.1.error_message" equal to
      """
      expected response body json at "error.code" to contain 'E2' but got 'E1'
      """

  @json_schema
  Scenario: A JSON schema assertion lists the schema errors
    Given the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: |
          pm.test("schema", () => pm.response.to.have.jsonSchema({ type: "object", required: ["id"] }))
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "tests.0.status" equal to "failed"
    And the response should contain the field "tests.0.error_message" equal to
      """
      expected data to satisfy schema but found following errors: 
      data should have required property 'id'
      """

  @order @labels @count
  Scenario: Tests from collection, folder and request scripts are listed in run order with their script and counted
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      scripts:
        preRequest: pm.test("collection pre", () => {})
        postResponse: pm.test("collection post", () => pm.response.to.have.status(200))
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      scripts:
        postResponse: |
          pm.test("folder post", () => pm.response.to.be.notFound)
          pm.test.skip("folder later")
      """
    And the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: pm.test("request post", () => pm.expect(pm.response.json().error.code).to.eql("E1"))
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "tests" array length equal to 5
    And the response should contain the field "tests.0.name" equal to "collection pre"
    And the response should contain the field "tests.0.index" equal to "0"
    And the response should contain the field "tests.0.status" equal to "passed"
    And the response should contain the field "tests.0.script.owner_kind" equal to "collection"
    And the response should contain the field "tests.0.script.owner_name" equal to "Billing"
    And the response should contain the field "tests.0.script.phase" equal to "preRequest"
    And the response should contain the field "tests.0.script.label" equal to "Billing · Pre-request"
    And the response should contain the field "tests.1.name" equal to "collection post"
    And the response should contain the field "tests.1.index" equal to "1"
    And the response should contain the field "tests.1.status" equal to "failed"
    And the response should contain the field "tests.1.script.label" equal to "Billing · Post-response"
    And the response should contain the field "tests.2.name" equal to "folder post"
    And the response should contain the field "tests.2.status" equal to "passed"
    And the response should contain the field "tests.2.script.owner_kind" equal to "folder"
    And the response should contain the field "tests.2.script.owner_name" equal to "Invoices"
    And the response should contain the field "tests.2.script.label" equal to "Invoices · Post-response"
    And the response should contain the field "tests.3.name" equal to "folder later"
    And the response should contain the field "tests.3.status" equal to "skipped"
    And the response field "tests.3.error_message" should not exist
    And the response should contain the field "tests.4.name" equal to "request post"
    And the response should contain the field "tests.4.index" equal to "4"
    And the response should contain the field "tests.4.status" equal to "passed"
    And the response should contain the field "tests.4.script.owner_kind" equal to "request"
    And the response should contain the field "tests.4.script.label" equal to "Create · Post-response"
    And the response should contain the field "test_summary.total" equal to "5"
    And the response should contain the field "test_summary.passed" equal to "3"
    And the response should contain the field "test_summary.failed" equal to "1"
    And the response should contain the field "test_summary.skipped" equal to "1"

  @no_tests
  Scenario: A send without tests has an empty list and zero counts
    Given the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "tests" array length equal to 0
    And the response should contain the field "test_summary.total" equal to "0"
    And the response should contain the field "test_summary.passed" equal to "0"
    And the response should contain the field "test_summary.failed" equal to "0"
    And the response should contain the field "test_summary.skipped" equal to "0"

  @pre_request_stopped
  Scenario: Tests of a pre-request script that then stops the send are still listed
    Given the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: |
          pm.test("token present", () => pm.expect(pm.variables.get("token")).to.be.a("string"))
          throw new Error("no token")
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "stopped_by_script_error"
    And the response should contain the field "tests" array length equal to 1
    And the response should contain the field "tests.0.name" equal to "token present"
    And the response should contain the field "tests.0.status" equal to "failed"
    And the response should contain the field "tests.0.error_message" equal to "expected undefined to be a string"
    And the response should contain the field "tests.0.script.label" equal to "Create · Pre-request"
    And the response should contain the field "test_summary.failed" equal to "1"
    And the "GET" request to "/invoices" should have been received 0 times

  @skipped_send
  Scenario: Tests of a pre-request script that skips the request are listed
    Given the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: |
          pm.test("checked first", () => {})
          pm.execution.skipRequest()
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "skipped_by_script"
    And the response should contain the field "tests" array length equal to 1
    And the response should contain the field "tests.0.status" equal to "passed"
    And the response should contain the field "test_summary.passed" equal to "1"

  @legacy
  Scenario: Legacy tests are collected at the end of their script, after its pm.test results
    Given the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: |
          tests["has error"] = responseBody.includes("error")
          tests["is 200"] = responseCode.code === 200
          pm.test("modern", () => {})
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "tests" array length equal to 3
    And the response should contain the field "tests.0.name" equal to "modern"
    And the response should contain the field "tests.1.name" equal to "has error"
    And the response should contain the field "tests.1.status" equal to "passed"
    And the response should contain the field "tests.2.name" equal to "is 200"
    And the response should contain the field "tests.2.status" equal to "failed"
    And the response should contain the field "tests.2.error_message" equal to "expected false to be truthy"
    And the response should contain the field "tests.2.script.label" equal to "Create · Post-response"
    And the response should contain the field "test_summary.total" equal to "3"
    And the response should contain the field "test_summary.passed" equal to "2"
