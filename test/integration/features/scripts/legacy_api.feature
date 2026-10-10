#language: en
#utf-8

@all @xen_004 @scripts @legacy_api
Feature: Old collections written with Postman's legacy API run unchanged
  The legacy globals map to the pm API: tests[...] become test results when the script ends,
  postman.setEnvironmentVariable/setGlobalVariable and their getters and clearers write the same
  scopes as pm.environment/pm.globals (current values, XEN-002), responseBody, responseCode,
  responseHeaders, responseTime, responseCookies, request, globals, environment, data and
  iteration describe the send, and _ (lodash 3), CryptoJS, tv4, xml2Json, cheerio and Backbone
  are there. postman.setNextRequest does nothing. The first use of each legacy name in a script
  logs Postman's deprecation notice as a console warning.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK9000000000000000000001"
    And the "create" variable is set to "01JK9000000000000000000004"
    And the "dev" variable is set to "01JK9000000000000000000009"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK9000000000000000000000
      name: Acme
      order:
        - billing
      variables:
        - key: region
          value: eu
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: base_url
          value: "{{api}}"
          enabled: true
          description: ""
      """
    And the 1 "POST" request to "/invoices" returns status 201 with the following response
      """
      {"id":"inv-1","total":10}
      """
    And the requests to "/invoices" respond with the header "Set-Cookie" equal to "session=s-1; Path=/"
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    And I call "ConsoleService.Clear"

  @tests
  Scenario: tests[...] become test results and the response globals describe the response
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{base_url}}/invoices"
      scripts:
        postResponse: |
          var data = JSON.parse(responseBody);
          tests["status is 201"] = responseCode.code === 201;
          tests["has id"] = data.id === "inv-1";
          tests["total is 20"] = data.total === 20;
          console.log(responseCode.name, typeof responseTime, request.method, iteration, postman.getResponseHeader("content-type"), responseCookies.count(), postman.getResponseCookie("session").value)
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    And the response should contain the field "tests" array length equal to 3
    And the response should contain the field "tests.0.name" equal to "status is 201"
    And the response should contain the field "tests.0.status" equal to "passed"
    And the response should contain the field "tests.1.name" equal to "has id"
    And the response should contain the field "tests.1.status" equal to "passed"
    And the response should contain the field "tests.2.name" equal to "total is 20"
    And the response should contain the field "tests.2.status" equal to "failed"
    And the response should contain the field "tests.2.error_message" equal to "expected false to be truthy"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "0.level" equal to "warn"
    And the response should contain the field "0.summary" equal to
      """
      Using "responseCode" is deprecated. Use "pm.response.code" instead.
      """
    And the response should contain the field "1.summary" equal to
      """
      Using "tests" is deprecated. Use "pm.test()" instead.
      """
    And the response should contain the field "5.summary" equal to
      """
      Using "postman.getResponseCookie" is deprecated. Use "pm.cookies.get()" instead.
      """
    And the response should contain the field "6.summary" equal to "Created number POST 0 application/json 1 s-1"

  @variables
  Scenario: postman.* variable functions write the same scopes as pm and setNextRequest does nothing
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{base_url}}/invoices"
      scripts:
        preRequest: |
          postman.setEnvironmentVariable("invoice", 7);
          postman.setGlobalVariable("region", "us");
          postman.setNextRequest("Other");
          console.log(postman.getEnvironmentVariable("invoice"), typeof pm.environment.get("invoice"), postman.getGlobalVariable("region"), environment.base_url !== undefined, globals.region);
        postResponse: |
          postman.clearGlobalVariable("region");
          console.log(pm.globals.has("region"), pm.environment.get("invoice"));
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.invoice" equal to "7"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "us"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "7.summary" equal to "7 string us true us"
    And the response should contain the field "9.summary" equal to "false 7"

  @libraries
  Scenario: Legacy library globals and xml2Json work like in Postman
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{base_url}}/invoices"
      scripts:
        preRequest: |
          console.log(_.VERSION, typeof _.pluck, CryptoJS.MD5("a").toString(), tv4.validate(1, { type: "number" }), xml2Json("<a><b>1</b></a>").a.b, cheerio.load("<i>x</i>")("i").text(), typeof Backbone.Model)
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 7
    And the response should contain the field "0.summary" equal to
      """
      Using "_" is deprecated. Use "require('lodash')" instead.
      """
    And the response should contain the field "6.summary" equal to "3.10.2 function 0cc175b9c0f1b6a831c399e269772661 true 1 x function"
