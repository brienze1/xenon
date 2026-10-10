#language: en
#utf-8

@all @xen_003 @examples @save_response_as_example
Feature: Save the response of a tab as an example of its request
  "Save response as example" creates a named example from the request as it was sent from the tab
  (with its {{variables}} unresolved, as in Postman) and the response the tab received: status
  code and text, content type, headers and body. A tab without a received response has nothing to
  save.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JH0000000000000000000001"
    And the "invoice" variable is set to "01JH0000000000000000000002"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JH0000000000000000000000
      name: Acme
      order:
        - billing
      variables:
        - key: base_url
          value: "{{api}}"
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      order:
        - create-invoice.yaml
      """
    And the workspace file "acme/collections/billing/create-invoice.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{invoice}}
      name: Create invoice
      method: POST
      url: "{{base_url}}/invoices"
      """
    And the 1 "POST" request to "/invoices" returns status 201 with content type "application/json" and body
      """
      {"id":42,"status":"draft"}
      """
    And the requests to "/invoices" respond with the header "X-Trace" equal to "t-9"

  @create
  Scenario: The example keeps the unresolved request as sent and the received response
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "id": "{{invoice}}",
          "method": "POST",
          "url": "{{base_url}}/invoices?draft={{draft}}",
          "params": [ { "key": "draft", "value": "{{draft}}" } ],
          "headers": [ { "key": "Content-Type", "value": "application/json" } ],
          "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"amount\":{{amount}}}" } },
          "auth": { "type": "bearer", "bearer": { "token": "{{token}}" } }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "201"
    When I call "ExampleService.CreateFromResponse" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "tab_id": "tab-1", "name": "Created" }
      """
    Then the call should succeed
    And the response should contain the field "id" matching "^[0-9A-HJKMNP-TV-Z]{26}$"
    And the response should contain the field "name" equal to "Created"
    And the response should contain the field "request.method" equal to "POST"
    And the response should contain the field "request.url" equal to "{{base_url}}/invoices?draft={{draft}}"
    And the response should contain the field "request.params.0.value" equal to "{{draft}}"
    And the response should contain the field "request.headers.0.value" equal to "application/json"
    And the response should contain the field "request.body.raw.content" equal to
      """
      {"amount":{{amount}}}
      """
    And the response should contain the field "request.auth.bearer.token" equal to "{{token}}"
    And the response should contain the field "response.code" equal to "201"
    And the response should contain the field "response.status" equal to "Created"
    And the response should contain the field "response.content_type" equal to "application/json"
    And the response field "response.headers" should contain an item with "key" equal to "X-Trace" and "value" equal to "t-9"
    And the response should contain the field "response.body" equal to
      """
      {"id":42,"status":"draft"}
      """
    And I save the response field "id" as "example"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples" array length equal to 1
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.id" equal to "{{example}}"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.name" equal to "Created"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.request.url" equal to "{{base_url}}/invoices?draft={{draft}}"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.response.code" equal to "201"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.response.status" equal to "Created"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.response.bodyEncoding" equal to "text"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.response.body" equal to
      """
      {"id":42,"status":"draft"}
      """
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "url" equal to "{{base_url}}/invoices"

  @errors
  Scenario: A tab without a response has nothing to save
    When I call "ExampleService.CreateFromResponse" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "tab_id": "tab-none", "name": "Created" }
      """
    Then the call should fail with code "EXM-02404"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should not contain the field "examples"

  @errors
  Scenario: A send that failed has no response to save
    Given the "closed" variable is set to a closed port url
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-2", "request": { "id": "{{invoice}}", "method": "GET", "url": "{{closed}}/x" } }
      """
    Then the call should succeed
    And the response should contain the field "ok" equal to "false"
    When I call "ExampleService.CreateFromResponse" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "tab_id": "tab-2", "name": "Broken" }
      """
    Then the call should fail with code "EXM-02404"

  @errors
  Scenario: The example needs a name
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{invoice}}", "method": "POST", "url": "{{base_url}}/invoices" } }
      """
    And I call "ExampleService.CreateFromResponse" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "tab_id": "tab-1", "name": "" }
      """
    Then the call should fail with code "EXM-03400"
