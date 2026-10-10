#language: en
#utf-8

@all @xen_003 @examples @examples_crud
Feature: Saved examples are created, edited, renamed, duplicated and deleted with their request
  A request can hold any number of saved examples: a name, a snapshot of the request (method, URL,
  params, path variables, headers, body, auth) and a response (status code and text, content
  type, headers, body). They live in the request's own file under "examples:", in sidebar order,
  and names need not be unique. The tree lists them under their request with their status code.
  Saving an example follows the request rule: it is refused when the request file changed on disk
  since it was opened, unless the user keeps their version.

  Background:
    Given the workspace "Acme" exists
    And the "billing" variable is set to "01JG0000000000000000000001"
    And the "invoice" variable is set to "01JG0000000000000000000002"
    And the "created" variable is set to "01JG0000000000000000000010"
    And the "missing" variable is set to "01JG0000000000000000000011"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JG0000000000000000000000
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
      headers:
        - key: Content-Type
          value: application/json
          enabled: true
          description: ""
      body:
        mode: raw
        raw:
          language: json
          content: '{"amount":1000}'
      examples:
        - id: {{created}}
          name: Created
          request:
            method: POST
            url: "{{base_url}}/invoices"
            params: []
            pathVariables: []
            headers:
              - key: Content-Type
                value: application/json
            body:
              mode: raw
              raw:
                language: json
                content: '{"amount":1000}'
            auth:
              type: inherit
          response:
            code: 201
            status: Created
            contentType: application/json
            headers:
              - key: Content-Type
                value: application/json
              - key: Set-Cookie
                value: a=1
              - key: Set-Cookie
                value: b=2
            bodyEncoding: text
            body: |
              { "id": 42 }
          postman:
            id: ex-created
            responseTime: 12
        - id: {{missing}}
          name: Not found
          request:
            method: POST
            url: "{{base_url}}/invoices/9"
          response:
            code: 404
            status: Not Found
            bodyEncoding: text
            body: ""
      """

  @get
  Scenario: Getting an example returns its request snapshot, its response and the request revision
    When I call "ExampleService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "example_id": "{{created}}" }
      """
    Then the call should succeed
    And the response should contain the field "id" equal to "{{created}}"
    And the response should contain the field "request_id" equal to "{{invoice}}"
    And the response should contain the field "name" equal to "Created"
    And the response should contain the field "revision" matching "^[0-9a-f]{64}$"
    And the response should contain the field "request.method" equal to "POST"
    And the response should contain the field "request.url" equal to "{{base_url}}/invoices"
    And the response should contain the field "request.headers.0.key" equal to "Content-Type"
    And the response should contain the field "request.body.mode" equal to "raw"
    And the response should contain the field "request.body.raw.content" equal to
      """
      {"amount":1000}
      """
    And the response should contain the field "request.auth.type" equal to "inherit"
    And the response should contain the field "response.code" equal to "201"
    And the response should contain the field "response.status" equal to "Created"
    And the response should contain the field "response.content_type" equal to "application/json"
    And the response should contain the field "response.headers" array length equal to 3
    And the response should contain the field "response.headers.2.value" equal to "b=2"
    And the response should contain the field "response.body_encoding" equal to "text"
    And the response should contain the field "response.body" matching "^\{ .id.: 42 \}\n$"
    And the response should contain the field "response.is_binary" equal to "false"
    And the response should contain the field "response.size_bytes" equal to "13"
    And the response should contain the field "response.body_url" equal to ""

  @get @errors
  Scenario Outline: Getting an unknown example or request fails - <case>
    When I call "ExampleService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "<request>", "example_id": "<example>" }
      """
    Then the call should fail with code "<code>"

    Examples:
      | case            | request                    | example                    | code      |
      | unknown example | {{invoice}}                | 01JG9999999999999999999999 | EXM-01404 |
      | unknown request | 01JG9999999999999999999998 | {{created}}                | REQ-01404 |

  @tree
  Scenario: The tree lists the examples under their request with their status code
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "0.children.0.name" equal to "Create invoice"
    And the response should contain the field "0.children.0.examples" array length equal to 2
    And the response should contain the field "0.children.0.examples.0.id" equal to "{{created}}"
    And the response should contain the field "0.children.0.examples.0.kind" equal to "example"
    And the response should contain the field "0.children.0.examples.0.name" equal to "Created"
    And the response should contain the field "0.children.0.examples.0.code" equal to "201"
    And the response should contain the field "0.children.0.examples.1.name" equal to "Not found"
    And the response should contain the field "0.children.0.examples.1.code" equal to "404"

  @create
  Scenario: Creating an example copies the request and starts with an empty 200 OK response
    When I call "ExampleService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "name": "Draft" }
      """
    Then the call should succeed
    And the response should contain the field "id" matching "^[0-9A-HJKMNP-TV-Z]{26}$"
    And the response should contain the field "name" equal to "Draft"
    And the response should contain the field "request.method" equal to "POST"
    And the response should contain the field "request.url" equal to "{{base_url}}/invoices"
    And the response should contain the field "request.body.raw.content" equal to
      """
      {"amount":1000}
      """
    And the response should contain the field "response.code" equal to "200"
    And the response should contain the field "response.status" equal to "OK"
    And the response should contain the field "response.body" equal to ""
    And I save the response field "id" as "draft"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples" array length equal to 3
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.2.id" equal to "{{draft}}"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.2.name" equal to "Draft"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.2.request.url" equal to "{{base_url}}/invoices"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.2.response.code" equal to "200"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.2.response.bodyEncoding" equal to "text"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.postman.id" equal to "ex-created"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "method" equal to "POST"

  @create @errors
  Scenario: An example needs a name
    When I call "ExampleService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "name": "  " }
      """
    Then the call should fail with code "EXM-03400"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should be unchanged

  @save
  Scenario: Saving an example writes its request and response into the request file
    When I call "ExampleService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "example_id": "{{created}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "ExampleService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request_id": "{{invoice}}",
        "base_revision": "{{revision}}",
        "example": {
          "id": "{{created}}",
          "name": "Accepted",
          "request": {
            "method": "PUT",
            "url": "{{base_url}}/invoices/1",
            "params": [ { "key": "dry", "value": "1", "enabled": true, "description": "" } ],
            "path_variables": [],
            "headers": [],
            "body": { "mode": "none" },
            "auth": { "type": "bearer", "bearer": { "token": "{{token}}" } }
          },
          "response": {
            "code": 202,
            "status": "Accepted",
            "content_type": "text/plain",
            "headers": [ { "key": "X-Trace", "value": "t-1", "enabled": true, "description": "" } ],
            "body": "queued",
            "body_encoding": "text"
          }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Accepted"
    And the response should contain the field "response.code" equal to "202"
    And the response should contain the field "revision" not equal to "{{revision}}"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples" array length equal to 2
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.id" equal to "{{created}}"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.name" equal to "Accepted"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.request.method" equal to "PUT"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.request.params.0.key" equal to "dry"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.request.auth.type" equal to "bearer"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.response.code" equal to "202"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.response.contentType" equal to "text/plain"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.response.headers.0.key" equal to "X-Trace"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.response.body" equal to "queued"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.postman.id" equal to "ex-created"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.1.name" equal to "Not found"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "method" equal to "POST"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "url" equal to "{{base_url}}/invoices"
    When the app is restarted
    And I call "ExampleService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "example_id": "{{created}}" }
      """
    Then the response should contain the field "name" equal to "Accepted"
    And the response should contain the field "response.body" equal to "queued"

  @save @conflict
  Scenario: Saving is refused when the request file changed on disk, unless forced
    When I call "ExampleService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "example_id": "{{missing}}" }
      """
    And I save the response field "revision" as "revision"
    Given the workspace file "acme/collections/billing/create-invoice.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{invoice}}
      name: Create invoice
      method: PATCH
      url: "{{base_url}}/invoices"
      examples:
        - id: {{missing}}
          name: Not found
          response:
            code: 404
            status: Not Found
      """
    When I call "ExampleService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request_id": "{{invoice}}",
        "base_revision": "{{revision}}",
        "example": { "id": "{{missing}}", "name": "Gone", "request": { "method": "GET", "url": "x" }, "response": { "code": 410, "status": "Gone", "body": "", "body_encoding": "text" } }
      }
      """
    Then the call should fail with code "REQ-02409"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.name" equal to "Not found"
    When I call "ExampleService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request_id": "{{invoice}}",
        "base_revision": "{{revision}}",
        "force": true,
        "example": { "id": "{{missing}}", "name": "Gone", "request": { "method": "GET", "url": "x" }, "response": { "code": 410, "status": "Gone", "body": "", "body_encoding": "text" } }
      }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.name" equal to "Gone"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.response.code" equal to "410"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "method" equal to "PATCH"

  @save @errors
  Scenario: Saving an unknown example fails
    When I call "ExampleService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request_id": "{{invoice}}",
        "force": true,
        "example": { "id": "01JG9999999999999999999999", "name": "Ghost", "request": { "method": "GET", "url": "x" }, "response": { "code": 200, "status": "OK", "body": "", "body_encoding": "text" } }
      }
      """
    Then the call should fail with code "EXM-01404"

  @save @errors
  Scenario Outline: An example's response code must be between 100 and 599 - <code>
    When I call "ExampleService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request_id": "{{invoice}}",
        "force": true,
        "example": { "id": "{{created}}", "name": "Created", "request": { "method": "GET", "url": "x" }, "response": { "code": <code>, "status": "", "body": "", "body_encoding": "text" } }
      }
      """
    Then the call should fail with code "EXM-04400"

    Examples:
      | code |
      | 99   |
      | 600  |

  @rename
  Scenario: Renaming an example changes only its name
    When I call "ExampleService.Rename" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "example_id": "{{created}}", "name": "Invoice created" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Invoice created"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.name" equal to "Invoice created"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.response.code" equal to "201"

  @rename @errors
  Scenario: Renaming to a blank name fails
    When I call "ExampleService.Rename" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "example_id": "{{created}}", "name": "" }
      """
    Then the call should fail with code "EXM-03400"

  @duplicate
  Scenario: Duplicating an example places "<name> Copy" right after the original
    When I call "ExampleService.Duplicate" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "example_id": "{{created}}" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Created Copy"
    And the response should contain the field "id" not equal to "{{created}}"
    And the response should contain the field "response.code" equal to "201"
    And I save the response field "id" as "copy"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples" array length equal to 3
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.name" equal to "Created"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.1.id" equal to "{{copy}}"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.1.name" equal to "Created Copy"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.1.response.body" matching "^\{ .id.: 42 \}\n$"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.2.name" equal to "Not found"

  @delete
  Scenario: Deleting examples removes them from the request file
    When I call "ExampleService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "example_id": "{{created}}" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples" array length equal to 1
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.name" equal to "Not found"
    When I call "ExampleService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "example_id": "{{missing}}" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/create-invoice.yaml" should not contain the field "examples"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "name" equal to "Create invoice"
    When I call "ExampleService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{invoice}}", "example_id": "{{missing}}" }
      """
    Then the call should fail with code "EXM-01404"

  @request_save
  Scenario: Saving the request keeps its examples
    When I call "RequestService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "force": true,
        "request": { "id": "{{invoice}}", "name": "Create invoice", "method": "PUT", "url": "{{base_url}}/invoices" }
      }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "method" equal to "PUT"
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples" array length equal to 2
    And the workspace file "acme/collections/billing/create-invoice.yaml" should contain the field "examples.0.response.code" equal to "201"
