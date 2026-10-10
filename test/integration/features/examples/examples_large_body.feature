#language: en
#utf-8

@all @xen_003 @examples @examples_large_body
Feature: Examples keep large and binary response bodies
  An example's body is stored in the request file however large it is: text as a literal block,
  anything that is not UTF-8 text as base64 ("bodyEncoding: base64"). Bodies up to 1 MB of text
  come back inline; larger or binary bodies are served with their exact bytes by the in-app
  response route under "example:<id>", like live responses.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JJ0000000000000000000001"
    And the "report" variable is set to "01JJ0000000000000000000002"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JJ0000000000000000000000
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
      order:
        - report.yaml
      """
    And the workspace file "acme/collections/billing/report.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{report}}
      name: Report
      method: GET
      url: "{{api}}/report"
      """

  @large
  Scenario: An 8 MB body is stored and served through the response route
    Given the 1 "GET" request to "/report" returns status 200 with content type "text/plain" and a generated body of 8388608 bytes
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{report}}", "method": "GET", "url": "{{api}}/report" } }
      """
    When I call "ExampleService.CreateFromResponse" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{report}}", "tab_id": "tab-1", "name": "Big" }
      """
    Then the call should succeed
    And I save the response field "id" as "big"
    And the response field "response.body" should not exist
    And the response should contain the field "response.size_bytes" equal to "8388608"
    And the response should contain the field "response.body_url" equal to "/xenon/response/example:{{big}}"
    And the workspace file "acme/collections/billing/report.yaml" should contain the field "examples.0.response.bodyEncoding" equal to "text"
    When I request "/xenon/response/example:{{big}}" from the "ResponseBodyRoute" route
    Then the route response status should be 200
    And the route response header "Content-Type" should be "text/plain"
    And the route response body should be the generated body of 8388608 bytes
    When the app is restarted
    And I call "ExampleService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{report}}", "example_id": "{{big}}" }
      """
    Then the call should succeed
    And the response field "response.body" should not exist
    And the response should contain the field "response.body_url" equal to "/xenon/response/example:{{big}}"
    When I request "/xenon/response/example:{{big}}" from the "ResponseBodyRoute" route
    Then the route response status should be 200
    And the route response body should be the generated body of 8388608 bytes

  @binary
  Scenario: A binary body round-trips through base64
    Given the 1 "GET" request to "/report" returns status 200 with content type "image/png" and the resource file "pixel.png"
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{report}}", "method": "GET", "url": "{{api}}/report" } }
      """
    When I call "ExampleService.CreateFromResponse" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{report}}", "tab_id": "tab-1", "name": "Pixel" }
      """
    Then the call should succeed
    And I save the response field "id" as "pixel"
    And the response should contain the field "response.is_binary" equal to "true"
    And the response should contain the field "response.body_encoding" equal to "base64"
    And the response field "response.body" should not exist
    And the workspace file "acme/collections/billing/report.yaml" should contain the field "examples.0.response.bodyEncoding" equal to "base64"
    And the workspace file "acme/collections/billing/report.yaml" should contain the field "examples.0.response.body" matching "^iVBORw0KGgo"
    When the app is restarted
    And I call "ExampleService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{report}}", "example_id": "{{pixel}}" }
      """
    Then the response should contain the field "response.body_url" equal to "/xenon/response/example:{{pixel}}"
    When I request "/xenon/response/example:{{pixel}}" from the "ResponseBodyRoute" route
    Then the route response status should be 200
    And the route response header "Content-Type" should be "image/png"
    And the route response body should be the resource file "pixel.png"

  @inline
  Scenario: A small text body comes back inline
    Given the 1 "GET" request to "/report" returns status 200 with content type "text/csv" and body
      """
      month,total
      jan,10
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{report}}", "method": "GET", "url": "{{api}}/report" } }
      """
    When I call "ExampleService.CreateFromResponse" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{report}}", "tab_id": "tab-1", "name": "Csv" }
      """
    Then the call should succeed
    And the response should contain the field "response.body" equal to
      """
      month,total
      jan,10
      """
    And the response should contain the field "response.body_url" equal to ""
    And the workspace file "acme/collections/billing/report.yaml" should contain the field "examples.0.response.body" equal to
      """
      month,total
      jan,10
      """
