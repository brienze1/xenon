#language: en
#utf-8

@all @xen_007 @documentation @documentation_apply
Feature: Edits to a collection's OpenAPI document are saved back onto its requests
  The Documentation tab edits the OpenAPI document both ways. Saving it first shows a plan — the
  requests it updates, creates and deletes, by name — and then writes every change together, one
  atomic write per request file. Operations are matched to requests by "x-xenon-id"; an operation
  without one becomes a new request in the folder named by its first tag (created when missing),
  and a request whose operation was removed is deleted only once the deletions are confirmed.
  Only what OpenAPI expresses changes (name, description, method, URL, parameters, headers, body
  and examples, and the collection's title and description); scripts, auth values and settings
  stay as they are. A document that is not valid YAML or OpenAPI is refused with its line and
  column, and nothing is written.

  Background:
    Given the workspace "Acme" exists
    And the "billing" variable is set to "01JK8000000000000000000001"
    And the "invoices" variable is set to "01JK8000000000000000000002"
    And the "list" variable is set to "01JK8000000000000000000010"
    And the "create" variable is set to "01JK8000000000000000000011"
    And the "ping" variable is set to "01JK8000000000000000000012"
    And the "created" variable is set to "01JK8000000000000000000020"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK8000000000000000000000
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
      description: Invoices and payments.
      order:
        - invoices
        - ping.yaml
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      order:
        - list-invoices.yaml
        - create-invoice.yaml
      """
    And the workspace file "acme/collections/billing/invoices/list-invoices.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{list}}
      name: List invoices
      method: GET
      url: "{{host}}/invoices?status=open"
      params:
        - key: status
          value: open
          enabled: true
          description: Invoice status
      headers:
        - key: X-Tenant
          value: acme
          enabled: true
          description: ""
      auth:
        type: inherit
      scripts:
        preRequest: console.log('keep')
      """
    And the workspace file "acme/collections/billing/invoices/create-invoice.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create invoice
      method: POST
      url: "{{host}}/invoices"
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
      auth:
        type: bearer
        bearer:
          token: s3cr3t-token
      examples:
        - id: {{created}}
          name: Created
          request:
            method: POST
            url: "{{host}}/invoices"
          response:
            code: 201
            status: Created
            contentType: application/json
            bodyEncoding: text
            body: '{"id":42}'
      """
    And the workspace file "acme/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{ping}}
      name: Ping
      method: GET
      url: "{{host}}/ping"
      auth:
        type: none
      """
    And the temp folder "workspaces/acme" is copied to "before"

  @roundtrip
  Scenario: Saving the generated document without edits changes nothing
    When I call "TreeService.GetDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}" }
      """
    Then the call should succeed
    And I save the response field "yaml" as "generated"
    When I call "TreeService.PlanDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "yaml": "{{generated}}" }
      """
    Then the call should succeed
    And the response should contain the field "collection_changed" equal to "false"
    And the response should contain the field "updated" array length equal to 0
    And the response should contain the field "created" array length equal to 0
    And the response should contain the field "deleted" array length equal to 0
    When I call "TreeService.ApplyDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "yaml": "{{generated}}" }
      """
    Then the call should succeed
    And the response should contain the field "yaml" equal to "{{generated}}"
    And the folder "workspaces/acme" should equal the folder "before"

  @update
  Scenario: Descriptions, headers, bodies, examples and names flow back onto their requests
    When I call "TreeService.PlanDocumentation" with resource fields
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "yaml": {{resource:documentation/billing-edits.yaml#update}} }
      """
    Then the call should succeed
    And the response should contain the field "collection_changed" equal to "true"
    And the response should contain the field "updated" array length equal to 3
    And the response should contain the field "updated.0.name" equal to "List invoices"
    And the response should contain the field "updated.1.name" equal to "Create invoice"
    And the response should contain the field "updated.2.name" equal to "Health check"
    And the response should contain the field "updated.2.id" equal to "{{ping}}"
    And the response should contain the field "created" array length equal to 0
    And the response should contain the field "deleted" array length equal to 0
    And the folder "workspaces/acme" should equal the folder "before"
    When I call "TreeService.ApplyDocumentation" with resource fields
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "yaml": {{resource:documentation/billing-edits.yaml#update}} }
      """
    Then the call should succeed
    And the response should contain the field "updated" array length equal to 3
    And the response field "updated.0.revision" should exist
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "description" equal to "All billing calls."
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "name" equal to "Billing"
    And the workspace file "acme/collections/billing/invoices/list-invoices.yaml" should contain the field "description" equal to "Lists open invoices."
    And the workspace file "acme/collections/billing/invoices/list-invoices.yaml" should contain the field "url" equal to "{{host}}/invoices?status=open"
    And the workspace file "acme/collections/billing/invoices/list-invoices.yaml" should contain the field "params.0.value" equal to "open"
    And the workspace file "acme/collections/billing/invoices/list-invoices.yaml" should contain the field "headers" array length equal to 2
    And the workspace file "acme/collections/billing/invoices/list-invoices.yaml" should contain the field "headers.0.key" equal to "X-Tenant"
    And the workspace file "acme/collections/billing/invoices/list-invoices.yaml" should contain the field "headers.0.value" equal to "globex"
    And the workspace file "acme/collections/billing/invoices/list-invoices.yaml" should contain the field "headers.1.key" equal to "X-Request-Id"
    And the workspace file "acme/collections/billing/invoices/list-invoices.yaml" should contain the field "headers.1.value" equal to "{{requestId}}"
    And the workspace file "acme/collections/billing/invoices/list-invoices.yaml" should contain the field "headers.1.description" equal to "Request id"
    And the workspace file "acme/collections/billing/invoices/list-invoices.yaml" should contain the field "scripts.preRequest" equal to "console.log('keep')"
    And the workspace file "acme/collections/billing/invoices/create-invoice.yaml" should contain the field "body.mode" equal to "raw"
    And the workspace file "acme/collections/billing/invoices/create-invoice.yaml" should contain the field "body.raw.language" equal to "json"
    And the workspace file "acme/collections/billing/invoices/create-invoice.yaml" should contain the field "body.raw.content" equal to
      """
      {
        "amount": 2500,
        "currency": "EUR"
      }
      """
    And the workspace file "acme/collections/billing/invoices/create-invoice.yaml" should contain the field "headers.0.key" equal to "Content-Type"
    And the workspace file "acme/collections/billing/invoices/create-invoice.yaml" should contain the field "auth.bearer.token" equal to "s3cr3t-token"
    And the workspace file "acme/collections/billing/invoices/create-invoice.yaml" should contain the field "examples.0.id" equal to "{{created}}"
    And the workspace file "acme/collections/billing/invoices/create-invoice.yaml" should contain the field "examples.0.response.code" equal to "201"
    And the workspace file "acme/collections/billing/invoices/create-invoice.yaml" should contain the field "examples.0.response.body" equal to
      """
      {
        "id": 43
      }
      """
    And the workspace file "acme/collections/billing/health-check.yaml" should contain the field "name" equal to "Health check"
    And the workspace file "acme/collections/billing/health-check.yaml" should contain the field "url" equal to "{{host}}/ping"
    And the workspace file "acme/collections/billing/ping.yaml" should not exist

  @create
  Scenario: Operations without an x-xenon-id become new requests in the folder of their first tag
    When I call "TreeService.PlanDocumentation" with resource fields
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "yaml": {{resource:documentation/billing-edits.yaml#create}} }
      """
    Then the call should succeed
    And the response should contain the field "updated" array length equal to 0
    And the response should contain the field "deleted" array length equal to 0
    And the response should contain the field "created" array length equal to 2
    And the response should contain the field "created.0.name" equal to "Archive invoice"
    And the response should contain the field "created.0.folder" equal to "Invoices / Archive"
    And the response should contain the field "created.0.method" equal to "DELETE"
    And the response should contain the field "created.0.path" equal to "/invoices/{id}"
    And the response should contain the field "created.1.name" equal to "Version"
    And the response should contain the field "created.1.folder" equal to ""
    When I call "TreeService.ApplyDocumentation" with resource fields
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "yaml": {{resource:documentation/billing-edits.yaml#create}} }
      """
    Then the call should succeed
    And the response should contain the field "created" array length equal to 2
    And the response field "created.0.id" should exist
    And the workspace file "acme/collections/billing/invoices/archive/folder.yaml" should contain the field "name" equal to "Archive"
    And the workspace file "acme/collections/billing/invoices/archive/archive-invoice.yaml" should contain the field "name" equal to "Archive invoice"
    And the workspace file "acme/collections/billing/invoices/archive/archive-invoice.yaml" should contain the field "method" equal to "DELETE"
    And the workspace file "acme/collections/billing/invoices/archive/archive-invoice.yaml" should contain the field "url" equal to "{{host}}/invoices/:id"
    And the workspace file "acme/collections/billing/invoices/archive/archive-invoice.yaml" should contain the field "pathVariables.0.key" equal to "id"
    And the workspace file "acme/collections/billing/invoices/archive/archive-invoice.yaml" should contain the field "pathVariables.0.value" equal to "7"
    And the workspace file "acme/collections/billing/invoices/archive/archive-invoice.yaml" should contain the field "pathVariables.0.description" equal to "Invoice id"
    And the workspace file "acme/collections/billing/invoices/archive/archive-invoice.yaml" should contain the field "examples.0.name" equal to "Archived"
    And the workspace file "acme/collections/billing/invoices/archive/archive-invoice.yaml" should contain the field "examples.0.response.code" equal to "200"
    And the workspace file "acme/collections/billing/version.yaml" should contain the field "method" equal to "GET"
    And the workspace file "acme/collections/billing/version.yaml" should contain the field "url" equal to "{{host}}/version"
    And the workspace file "acme/collections/billing/version.yaml" should contain the field "description" equal to "Returns the API version."
    And the folder "workspaces/acme/collections/billing/invoices/list-invoices.yaml" should equal the folder "before/collections/billing/invoices/list-invoices.yaml"
    And the folder "workspaces/acme/collections/billing/ping.yaml" should equal the folder "before/collections/billing/ping.yaml"

  @delete
  Scenario: Removed operations are listed as deletions and deleted only once confirmed
    When I call "TreeService.PlanDocumentation" with resource fields
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "yaml": {{resource:documentation/billing-edits.yaml#delete}} }
      """
    Then the call should succeed
    And the response should contain the field "deleted" array length equal to 1
    And the response should contain the field "deleted.0.id" equal to "{{ping}}"
    And the response should contain the field "deleted.0.name" equal to "Ping"
    When I call "TreeService.ApplyDocumentation" with resource fields
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "yaml": {{resource:documentation/billing-edits.yaml#delete}} }
      """
    Then the call should fail with code "DOC-04400"
    And the folder "workspaces/acme" should equal the folder "before"
    When I call "TreeService.ApplyDocumentation" with resource fields
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "yaml": {{resource:documentation/billing-edits.yaml#delete}}, "confirm_deletions": true }
      """
    Then the call should succeed
    And the response should contain the field "deleted.0.id" equal to "{{ping}}"
    And the workspace file "acme/collections/billing/ping.yaml" should not exist
    And the workspace file "acme/collections/billing/invoices/list-invoices.yaml" should exist

  @rename
  Scenario Outline: A new summary renames the request like any other rename and keeps its id and tabs
    Given I call "UIStateService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "tabs": [{ "tab_id": "tab-ping", "request_id": "{{ping}}", "pinned": true }], "active_tab_id": "tab-ping" }
      """
    When I call "TreeService.ApplyDocumentation" with resource fields
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "yaml": {{resource:documentation/billing-edits.yaml#<edit>}} }
      """
    Then the call should succeed
    And the response should contain the field "updated.0.id" equal to "{{ping}}"
    And the response should contain the field "updated.0.name" equal to "<name>"
    And the workspace folder "acme/collections/billing" should contain the entries "<entries>"
    And the workspace file "acme/collections/billing/<file>" should contain the field "id" equal to "{{ping}}"
    And the workspace file "acme/collections/billing/<file>" should contain the field "name" equal to "<name>"
    And the workspace file "acme/collections/billing/<file>" should contain the field "url" equal to "{{host}}/ping"
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "order" with the items "invoices, <file>"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{ping}}" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "<name>"
    When I call "UIStateService.Get" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "tabs.0.request_id" equal to "{{ping}}"

    Examples:
      | edit        | name         | file             | entries                                       |
      | rename      | Health check | health-check.yaml | collection.yaml, health-check.yaml, invoices |
      | rename_case | PING         | ping.yaml        | collection.yaml, invoices, ping.yaml          |

  @invalid
  Scenario Outline: A document that is not valid is refused with its line and column and nothing is written
    When I call "TreeService.<method>" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "yaml": "<yaml>", "confirm_deletions": true }
      """
    Then the call should fail with code "<code>"
    And the error should contain the field "error.error_details.0.attribute" equal to "line"
    And the error should contain the field "error.error_details.0.messages.0" equal to "<line>"
    And the error should contain the field "error.error_details.1.attribute" equal to "column"
    And the error should contain the field "error.error_details.1.messages.0" equal to "<column>"
    And the error should contain the field "error.description" containing "<reason>"
    And the folder "workspaces/acme" should equal the folder "before"

    Examples:
      | method               | yaml                                                                          | code      | line | column | reason                       |
      | ApplyDocumentation   | openapi: 3.1.0\\ninfo:\\n  title: Billing\\n   bad: x\\n                         | DOC-01400 | 4    | 4      | mapping values               |
      | PlanDocumentation    | openapi: 3.1.0\\ninfo:\\n  title: Billing\\n   bad: x\\n                         | DOC-01400 | 4    | 4      | mapping values               |
      | PreviewDocumentation | openapi: 3.1.0\\ninfo:\\n  title: Billing\\n   bad: x\\n                         | DOC-01400 | 4    | 4      | mapping values               |
      | ApplyDocumentation   | openapi: 3.1.0\\ninfo:\\n  version: 1.0.0\\npaths: {}\\n                          | DOC-02400 | 3    | 3      | info.title                   |
      | ApplyDocumentation   | swagger: '2.0'\\ninfo:\\n  title: Billing\\n  version: 1.0.0\\n                   | DOC-02400 | 1    | 1      | openapi                      |
      | ApplyDocumentation   | openapi: 3.1.0\\ninfo:\\n  title: Billing\\n  version: 1.0.0\\npaths:\\n  invoices: {}\\n | DOC-02400 | 6    | 3      | must start with /            |
      | ApplyDocumentation   | - just\\n- a list\\n                                                            | DOC-02400 | 1    | 1      | mapping                      |

  @invalid
  Scenario: Two operations with the same x-xenon-id are refused at the second one
    When I call "TreeService.ApplyDocumentation" with resource fields
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "yaml": {{resource:documentation/billing-edits.yaml#duplicate}}, "confirm_deletions": true }
      """
    Then the call should fail with code "DOC-03400"
    And the error should contain the field "error.error_details.0.messages.0" equal to "13"
    And the error should contain the field "error.error_details.1.messages.0" equal to "19"
    And the folder "workspaces/acme" should equal the folder "before"

  @preview
  Scenario: Previewing an edited document shows it without writing anything
    When I call "TreeService.PreviewDocumentation" with resource fields
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "yaml": {{resource:documentation/billing-edits.yaml#create}} }
      """
    Then the call should succeed
    And the response should contain the field "document.operations" array length equal to 5
    And the response should contain the field "document.operations.2.summary" equal to "Archive invoice"
    And the response should contain the field "document.operations.2.request_id" equal to ""
    And the response should contain the field "document.operations.2.tag" equal to "Invoices / Archive"
    And the response should contain the field "document.operations.2.parameters.0.example" equal to "7"
    And the folder "workspaces/acme" should equal the folder "before"
