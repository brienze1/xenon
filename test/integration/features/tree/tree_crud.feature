#language: en
#utf-8

@all @xen_001 @tree @tree_crud
Feature: Create, rename, duplicate and delete sidebar items
  Every collection and folder is a directory with its own settings file and every request is
  its own YAML file, so the folder on disk mirrors the sidebar. Each change in the sidebar is the
  matching change on disk, and the parent's order follows it.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    And the "billing" variable is set to "01JB0000000000000000000001"
    And the "invoices" variable is set to "01JB0000000000000000000002"
    And the "create" variable is set to "01JB0000000000000000000003"
    And the "ping" variable is set to "01JB0000000000000000000004"
    And the workspace file "globex/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JB0000000000000000000000
      name: Globex
      description: ""
      order:
        - billing-api
      """
    And the workspace file "globex/collections/billing-api/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing API
      order:
        - invoices
        - ping.yaml
      """
    And the workspace file "globex/collections/billing-api/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      order:
        - create-invoice.yaml
      """
    And the workspace file "globex/collections/billing-api/invoices/create-invoice.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create invoice
      method: POST
      url: https://api.example.test/invoices
      """
    And the workspace file "globex/collections/billing-api/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{ping}}
      name: Ping
      method: GET
      url: https://api.example.test/ping
      """

  @success @create
  Scenario: A collection, nested folders and a request become directories and files with every default written
    When I call "TreeService.CreateCollection" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "Billing API" }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "collection"
    And the response should contain the field "name" equal to "Billing API"
    And I save the response field "id" as "billing"
    When I call "TreeService.CreateFolder" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{billing}}", "name": "Invoices" }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "folder"
    And I save the response field "id" as "invoices"
    When I call "TreeService.CreateFolder" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{invoices}}", "name": "Drafts" }
      """
    Then the call should succeed
    And I save the response field "id" as "drafts"
    When I call "TreeService.CreateRequest" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{drafts}}", "name": "Create invoice" }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "request"
    And the response should contain the field "name" equal to "Create invoice"
    And the response should contain the field "method" equal to "GET"
    And I save the response field "id" as "create"
    And the workspace file "acme/workspace.yaml" should contain the field "order" with the items "billing-api"
    And the workspace file "acme/collections/billing-api/collection.yaml" should equal
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing API
      description: ""
      order:
        - invoices
      auth:
        type: inherit
      headers: []
      variables: []
      scripts:
        preRequest: ""
        postResponse: ""
      postman: {}
      """
    And the workspace file "acme/collections/billing-api/invoices/folder.yaml" should equal
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      description: ""
      order:
        - drafts
      auth:
        type: inherit
      headers: []
      scripts:
        preRequest: ""
        postResponse: ""
      inheritScripts: true
      postman: {}
      """
    And the workspace file "acme/collections/billing-api/invoices/drafts/folder.yaml" should contain the field "order" with the items "create-invoice.yaml"
    And the workspace file "acme/collections/billing-api/invoices/drafts/create-invoice.yaml" should equal
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create invoice
      description: ""
      method: GET
      url: ""
      params: []
      pathVariables: []
      headers: []
      autoHeaders:
        disabled: []
      auth:
        type: inherit
      body:
        mode: none
      settings:
        timeoutMs: 0
        followRedirects: true
        verifySsl: true
        encodeUrl: true
      scripts:
        preRequest: ""
        postResponse: ""
      inheritScripts: true
      postman: {}
      """

  @success @create
  Scenario: The tree comes back from disk in the same shape
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Globex.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 1
    And the response should contain the field "0.id" equal to "{{billing}}"
    And the response should contain the field "0.kind" equal to "collection"
    And the response should contain the field "0.name" equal to "Billing API"
    And the response should contain the field "0.children.0.id" equal to "{{invoices}}"
    And the response should contain the field "0.children.0.kind" equal to "folder"
    And the response should contain the field "0.children.0.children.0.id" equal to "{{create}}"
    And the response should contain the field "0.children.0.children.0.kind" equal to "request"
    And the response should contain the field "0.children.0.children.0.name" equal to "Create invoice"
    And the response should contain the field "0.children.0.children.0.method" equal to "POST"
    And the response should contain the field "0.children.0.children.0.url" equal to "https://api.example.test/invoices"

  @success @create
  Scenario: A request created from an unsaved tab keeps what was typed
    When I call "TreeService.CreateRequest" with
      """
      {
        "local_id": "{{Globex.local_id}}",
        "parent_id": "{{billing}}",
        "name": "List invoices",
        "request": { "method": "post", "url": "https://api.example.test/invoices?page=1", "params": [ { "key": "page", "value": "1" } ] }
      }
      """
    Then the call should succeed
    And the response should contain the field "method" equal to "POST"
    And the response should contain the field "url" equal to "https://api.example.test/invoices?page=1"
    And the workspace file "globex/collections/billing-api/list-invoices.yaml" should contain the field "name" equal to "List invoices"
    And the workspace file "globex/collections/billing-api/list-invoices.yaml" should contain the field "method" equal to "POST"
    And the workspace file "globex/collections/billing-api/list-invoices.yaml" should contain the field "params.0.key" equal to "page"
    And the workspace file "globex/collections/billing-api/collection.yaml" should contain the field "order" with the items "invoices, list-invoices.yaml"

  @success @rename
  Scenario: Renaming a request renames its file and updates the parent order
    When I call "TreeService.Rename" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{create}}", "name": "Create draft invoice" }
      """
    Then the call should succeed
    And the response should contain the field "id" equal to "{{create}}"
    And the response should contain the field "name" equal to "Create draft invoice"
    And the workspace file "globex/collections/billing-api/invoices/create-invoice.yaml" should not exist
    And the workspace file "globex/collections/billing-api/invoices/create-draft-invoice.yaml" should contain the field "name" equal to "Create draft invoice"
    And the workspace file "globex/collections/billing-api/invoices/create-draft-invoice.yaml" should contain the field "id" equal to "{{create}}"
    And the workspace file "globex/collections/billing-api/invoices/create-draft-invoice.yaml" should contain the field "url" equal to "https://api.example.test/invoices"
    And the workspace file "globex/collections/billing-api/invoices/folder.yaml" should contain the field "order" with the items "create-draft-invoice.yaml"

  @success @rename
  Scenario: Renaming a folder renames its directory and keeps its children
    When I call "TreeService.Rename" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{invoices}}", "name": "Bills" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Bills"
    And the workspace folder "globex/collections/billing-api/bills" should exist
    And the workspace file "globex/collections/billing-api/invoices/folder.yaml" should not exist
    And the workspace file "globex/collections/billing-api/bills/folder.yaml" should contain the field "name" equal to "Bills"
    And the workspace file "globex/collections/billing-api/bills/create-invoice.yaml" should contain the field "id" equal to "{{create}}"
    And the workspace file "globex/collections/billing-api/collection.yaml" should contain the field "order" with the items "bills, ping.yaml"

  @success @rename
  Scenario: Renaming a collection updates the workspace order
    When I call "TreeService.Rename" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{billing}}", "name": "Payments" }
      """
    Then the call should succeed
    And the workspace file "globex/collections/payments/collection.yaml" should contain the field "name" equal to "Payments"
    And the workspace file "globex/workspace.yaml" should contain the field "order" with the items "payments"

  @success @duplicate
  Scenario: Duplicating a folder copies everything below it with new ids, right after the original
    When I call "TreeService.Duplicate" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{invoices}}" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Invoices Copy"
    And the response should contain the field "kind" equal to "folder"
    And the response should contain the field "id" not equal to "{{invoices}}"
    And the response should contain the field "children.0.name" equal to "Create invoice"
    And the response should contain the field "children.0.id" not equal to "{{create}}"
    And the workspace file "globex/collections/billing-api/collection.yaml" should contain the field "order" with the items "invoices, invoices-copy, ping.yaml"
    And the workspace file "globex/collections/billing-api/invoices-copy/folder.yaml" should contain the field "name" equal to "Invoices Copy"
    And the workspace file "globex/collections/billing-api/invoices-copy/folder.yaml" should contain the field "id" not equal to "{{invoices}}"
    And the workspace file "globex/collections/billing-api/invoices-copy/create-invoice.yaml" should contain the field "name" equal to "Create invoice"
    And the workspace file "globex/collections/billing-api/invoices-copy/create-invoice.yaml" should contain the field "id" not equal to "{{create}}"
    And the workspace file "globex/collections/billing-api/invoices-copy/create-invoice.yaml" should contain the field "url" equal to "https://api.example.test/invoices"
    And the workspace file "globex/collections/billing-api/invoices/create-invoice.yaml" should contain the field "id" equal to "{{create}}"

  @success @duplicate
  Scenario: Duplicating a request places "<name> Copy" right after it
    When I call "TreeService.Duplicate" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{ping}}" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Ping Copy"
    And the response should contain the field "method" equal to "GET"
    And the workspace file "globex/collections/billing-api/ping-copy.yaml" should contain the field "name" equal to "Ping Copy"
    And the workspace file "globex/collections/billing-api/ping-copy.yaml" should contain the field "id" not equal to "{{ping}}"
    And the workspace file "globex/collections/billing-api/collection.yaml" should contain the field "order" with the items "invoices, ping.yaml, ping-copy.yaml"

  @success @duplicate
  Scenario: Duplicating a collection appends the copy right after it in the workspace order
    When I call "TreeService.Duplicate" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{billing}}" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Billing API Copy"
    And the workspace file "globex/collections/billing-api-copy/collection.yaml" should contain the field "id" not equal to "{{billing}}"
    And the workspace file "globex/collections/billing-api-copy/invoices/create-invoice.yaml" should contain the field "id" not equal to "{{create}}"
    And the workspace file "globex/workspace.yaml" should contain the field "order" with the items "billing-api, billing-api-copy"

  @success @delete
  Scenario: Deleting a folder removes everything below it and reports the deleted requests
    When I call "TreeService.Delete" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{invoices}}" }
      """
    Then the call should succeed
    And the response should contain the field "deleted_request_ids" with the items "{{create}}"
    And the workspace file "globex/collections/billing-api/invoices/folder.yaml" should not exist
    And the workspace file "globex/collections/billing-api/invoices/create-invoice.yaml" should not exist
    And the workspace file "globex/collections/billing-api/collection.yaml" should contain the field "order" with the items "ping.yaml"

  @success @delete
  Scenario: Deleting a collection removes it from the workspace order
    When I call "TreeService.Delete" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{billing}}" }
      """
    Then the call should succeed
    And the response should contain the field "deleted_request_ids" array length equal to 2
    And the workspace file "globex/collections/billing-api/collection.yaml" should not exist
    And the workspace file "globex/workspace.yaml" should contain the field "order" with the items ""
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Globex.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0

  @not_found
  Scenario Outline: A parent that does not exist or cannot hold children is refused - <case>
    When I call "TreeService.<method>" with
      """
      { "local_id": "{{Globex.local_id}}", "parent_id": "<parent>", "name": "Anything" }
      """
    Then the call should fail with code "TRE-01404"

    Examples:
      | case             | method        | parent                     |
      | unknown folder   | CreateFolder  | 01JA2Z6A1B2C3D4E5F6G7H8J9K |
      | request folder   | CreateFolder  | {{create}}                 |
      | unknown request  | CreateRequest | 01JA2Z6A1B2C3D4E5F6G7H8J9K |
      | request request  | CreateRequest | {{ping}}                   |

  @not_found
  Scenario Outline: An item that does not exist cannot be changed - <method>
    When I call "TreeService.<method>" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "01JA2Z6A1B2C3D4E5F6G7H8J9K", "name": "Anything" }
      """
    Then the call should fail with code "TRE-02404"

    Examples:
      | method    |
      | Rename    |
      | Duplicate |
      | Delete    |

  @not_found
  Scenario: Tree calls on an unknown workspace fail
    When I call "TreeService.GetTree" with
      """
      { "local_id": "01JA2Z5V3B6C7D8E9F0G1H2J3K" }
      """
    Then the call should fail with code "WKS-07404"

  @validation
  Scenario Outline: Invalid input is rejected - <method>
    When I call "TreeService.<method>" with
      """
      <payload>
      """
    Then the call should fail with code "<code>"

    Examples:
      | method           | payload                                                                   | code      |
      | GetTree          | { "local_id": "" }                                                        | TRE-04400 |
      | Search           | { "local_id": "" , "query": "x" }                                         | TRE-05400 |
      | CreateCollection | { "local_id": "{{Globex.local_id}}", "name": "   " }                        | TRE-06400 |
      | CreateFolder     | { "local_id": "{{Globex.local_id}}", "parent_id": "", "name": "Invoices" }  | TRE-07400 |
      | CreateRequest    | { "local_id": "{{Globex.local_id}}", "parent_id": "x", "name": "" }         | TRE-08400 |
      | Rename           | { "local_id": "{{Globex.local_id}}", "id": "x", "name": "" }                | TRE-09400 |
      | Duplicate        | { "local_id": "{{Globex.local_id}}", "id": "" }                             | TRE-10400 |
      | Delete           | { "local_id": "", "id": "x" }                                             | TRE-11400 |
      | GetSettings      | { "local_id": "{{Globex.local_id}}", "id": "" }                             | TRE-12400 |
      | SaveSettings     | { "local_id": "{{Globex.local_id}}", "id": "", "auth": { "type": "none" } } | TRE-13400 |
