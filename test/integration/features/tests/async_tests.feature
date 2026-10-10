#language: en
#utf-8

@all @xen_004 @tests @async_tests
Feature: Asynchronous tests
  A pm.test whose function takes a done callback is asynchronous: it is recorded when done is
  called, failed when done gets an error or the function throws before calling it, and listed
  with async true. The script waits for its timers, so a done called from a timer still counts.
  A test written as an async function without done is not awaited (Postman behaviour): it is
  recorded as passed and a failure after an await is reported as a script error, not a test
  failure. An assertion that throws inside a timer is a script error too, and its test is not
  listed. A done that is never called leaves the test out.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK4100000000000000000001"
    And the "create" variable is set to "01JK4100000000000000000002"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK4100000000000000000000
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
    And the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      {"id":"inv-1"}
      """

  @done
  Scenario: A test that calls done from a timer passes and is listed after the synchronous tests that finished first
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: |
          pm.test("eventually", (done) => {
            setTimeout(() => {
              pm.expect(pm.response.json().id).to.eql("inv-1")
              done()
            }, 10)
          })
          pm.test("right away", () => {})
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    And the response should contain the field "tests" array length equal to 2
    And the response should contain the field "tests.0.name" equal to "right away"
    And the response should contain the field "tests.0.async" equal to "false"
    And the response should contain the field "tests.1.name" equal to "eventually"
    And the response should contain the field "tests.1.status" equal to "passed"
    And the response should contain the field "tests.1.async" equal to "true"
    And the response should contain the field "test_summary.passed" equal to "2"

  @done_error
  Scenario: done(err) fails the test with the error's message
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: |
          pm.test("late failure", (done) => {
            setTimeout(() => done(new Error("token expired")), 10)
          })
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    And the response should contain the field "tests" array length equal to 1
    And the response should contain the field "tests.0.status" equal to "failed"
    And the response should contain the field "tests.0.error_message" equal to "token expired"
    And the response should contain the field "tests.0.async" equal to "true"
    And the response should contain the field "test_summary.failed" equal to "1"

  @throws_before_done
  Scenario: An asynchronous test that throws before calling done fails with the thrown message
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: |
          pm.test("early", (done) => { throw new Error("early") })
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "tests" array length equal to 1
    And the response should contain the field "tests.0.status" equal to "failed"
    And the response should contain the field "tests.0.error_message" equal to "early"

  @never_done
  Scenario: A test whose done is never called is left out
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: |
          pm.test("forgotten", (done) => {})
          pm.test("sync", () => {})
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    And the response should contain the field "tests" array length equal to 1
    And the response should contain the field "tests.0.name" equal to "sync"

  @async_function
  Scenario: A failure after an await in an async test function is a script error and the test is listed as passed
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: |
          pm.test("async function", async () => {
            await new Promise((resolve) => setTimeout(resolve, 5))
            pm.expect(1).to.eql(2)
          })
          console.log("after")
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the response should contain the field "status_code" equal to "200"
    And the response should contain the field "tests" array length equal to 1
    And the response should contain the field "tests.0.name" equal to "async function"
    And the response should contain the field "tests.0.status" equal to "passed"
    And the response should contain the field "script_errors" array length equal to 1
    And the response should contain the field "script_errors.0.kind" equal to "script_unhandled_rejection"
    And the response should contain the field "script_errors.0.message" equal to "Unhandled promise rejection: AssertionError: expected 1 to deeply equal 2"
    And the response should contain the field "script_errors.0.script.label" equal to "Create · Post-response"

  @timer_assertion
  Scenario: An assertion that throws inside a timer is a script error and its test is not listed
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: |
          pm.test("timer assert", (done) => {
            setTimeout(() => {
              pm.expect(1).to.eql(2)
              done()
            }, 10)
          })
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "tests" array length equal to 0
    And the response should contain the field "script_errors" array length equal to 1
    And the response should contain the field "script_errors.0.kind" equal to "script_runtime"
    And the response should contain the field "script_errors.0.message" equal to "AssertionError: expected 1 to deeply equal 2"

  @await_send_request
  Scenario: A done test that waits for pm.sendRequest is recorded once the response arrives
    Given the 1 "GET" request to "/token" returns status 200 with the following response
      """
      {"token":"t-1"}
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
          pm.test("token endpoint answers", (done) => {
            pm.sendRequest(pm.variables.replaceIn("{{api}}/token"), (err, res) => {
              pm.expect(err).to.equal(null)
              pm.expect(res.json().token).to.eql("t-1")
              done()
            })
          })
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      variables:
        - key: api
          value: "{{api}}"
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    And the response should contain the field "tests" array length equal to 1
    And the response should contain the field "tests.0.status" equal to "passed"
    And the response should contain the field "tests.0.async" equal to "true"
