#language: en
#utf-8

@all @xen_008 @mcp @mcp_authoring
Feature: AI clients build and edit collections
  "tree_edit" creates, updates, renames, moves and duplicates collections, folders, requests and examples, and saves
  collection and folder settings, writing exactly the files a UI action writes. "delete" removes items. "docs" and
  "docs_edit" read and apply a collection's OpenAPI document; deletes are applied only when the call confirms them.
  Every change made through MCP shows up in the open window at once: the tree reloads on workspace:tree-changed and
  an updated request also emits request:changed-on-disk, which the UI handles like any external change.

  Background:
    Given the workspace "Reference" exists
    And the workspace "Acme" exists
    And the "api" variable is set to the mocked api url

  @create
  Scenario: Creating a collection, a folder and a request writes the same files as the UI and refreshes the tree
    Given the emitted events are forgotten
    When I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "collection", "name": "Payments" }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "collection"
    And the response should contain the field "name" equal to "Payments"
    And I save the response field "id" as "payments"
    And a "workspace:tree-changed" event with the field "local_id" equal to "{{Acme.local_id}}" should have been emitted within 1000 ms
    When I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "folder", "parent_id": "{{payments}}", "name": "Orders" }
      """
    Then the call should succeed
    And I save the response field "id" as "orders"
    When I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "request", "parent_id": "{{orders}}", "name": "Create order", "method": "POST", "url": "{{api}}/orders" }
      """
    Then the call should succeed
    And the response should contain the field "method" equal to "POST"
    And the event "workspace:tree-changed" should have been emitted 3 times
    When I call "TreeService.CreateCollection" with
      """
      { "local_id": "{{Reference.local_id}}", "name": "Payments" }
      """
    And I save the response field "id" as "ref_payments"
    And I call "TreeService.CreateFolder" with
      """
      { "local_id": "{{Reference.local_id}}", "parent_id": "{{ref_payments}}", "name": "Orders" }
      """
    And I save the response field "id" as "ref_orders"
    And I call "TreeService.CreateRequest" with
      """
      { "local_id": "{{Reference.local_id}}", "parent_id": "{{ref_orders}}", "name": "Create order" }
      """
    And I save the response field "id" as "ref_request"
    And I call "RequestService.Get" with
      """
      { "local_id": "{{Reference.local_id}}", "id": "{{ref_request}}" }
      """
    And I save the response field "revision" as "ref_revision"
    And I call "RequestService.Save" with
      """
      { "local_id": "{{Reference.local_id}}", "base_revision": "{{ref_revision}}", "force": false, "request": { "id": "{{ref_request}}", "name": "Create order", "method": "POST", "url": "{{api}}/orders" } }
      """
    Then the call should succeed
    Then the folder "workspaces/reference/collections" should equal the folder "workspaces/acme/collections" apart from ids
    When I call the MCP tool "tree" with
      """
      { "action": "list", "depth": "all" }
      """
    Then the response should contain the field "items.0.items.0.items.0.name" equal to "Create order"

  @update
  Scenario: Updating a request merges the given fields, keeps the rest and tells an open tab
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JV0000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JV0000000000000000000001
      name: Billing
      order:
        - invoices.yaml
      """
    And the workspace file "acme/collections/billing/invoices.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JV0000000000000000000002
      name: Invoices
      method: GET
      url: https://api.acme.test/invoices
      headers:
        - key: Accept
          value: application/json
          enabled: true
          description: ""
      """
    And the emitted events are forgotten
    When I call the MCP tool "tree_edit" with
      """
      { "action": "update", "id": "01JV0000000000000000000002", "url": "https://api.acme.test/v2/invoices", "description": "Lists invoices" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/invoices.yaml" should contain the field "url" equal to "https://api.acme.test/v2/invoices"
    And the workspace file "acme/collections/billing/invoices.yaml" should contain the field "description" equal to "Lists invoices"
    And the workspace file "acme/collections/billing/invoices.yaml" should contain the field "method" equal to "GET"
    And the workspace file "acme/collections/billing/invoices.yaml" should contain the field "headers.0.key" equal to "Accept"
    And a "request:changed-on-disk" event with the field "request_id" equal to "01JV0000000000000000000002" should have been emitted within 1000 ms
    And a "workspace:tree-changed" event with the field "local_id" equal to "{{Acme.local_id}}" should have been emitted within 1000 ms

  @rename_move_duplicate_delete
  Scenario: Renaming, moving, duplicating and deleting items
    Given I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "collection", "name": "Payments" }
      """
    And I save the response field "id" as "payments"
    And I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "folder", "parent_id": "{{payments}}", "name": "Orders" }
      """
    And I save the response field "id" as "orders"
    And I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "request", "parent_id": "{{payments}}", "name": "Refund" }
      """
    And I save the response field "id" as "refund"
    When I call the MCP tool "tree_edit" with
      """
      { "action": "rename", "id": "{{refund}}", "name": "Refund order" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/payments/refund-order.yaml" should contain the field "name" equal to "Refund order"
    When I call the MCP tool "tree_edit" with
      """
      { "action": "move", "id": "{{refund}}", "parent_id": "{{orders}}" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/payments/orders/refund-order.yaml" should exist
    And the workspace file "acme/collections/payments/refund-order.yaml" should not exist
    When I call the MCP tool "tree_edit" with
      """
      { "action": "duplicate", "id": "{{refund}}" }
      """
    Then the call should succeed
    And I save the response field "id" as "copy"
    When I call the MCP tool "tree" with
      """
      { "action": "list", "parent_id": "{{orders}}" }
      """
    Then the response should contain the field "items" array length equal to 2
    When I call the MCP tool "delete" with
      """
      { "target": "item", "id": "{{copy}}" }
      """
    Then the call should succeed
    When I call the MCP tool "tree" with
      """
      { "action": "list", "parent_id": "{{orders}}" }
      """
    Then the response should contain the field "items" array length equal to 1
    And the response should contain the field "items.0.id" equal to "{{refund}}"

  @settings
  Scenario: Collection and folder settings are saved like the UI saves them
    Given I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "collection", "name": "Payments" }
      """
    And I save the response field "id" as "payments"
    When I call the MCP tool "tree_edit" with
      """
      {
        "action": "save_settings",
        "id": "{{payments}}",
        "auth": { "type": "bearer", "bearer": { "token": "{{payments_token}}" } },
        "headers": [ { "key": "X-Team", "value": "payments", "enabled": true } ],
        "scripts": { "pre_request": "console.log('pre')", "post_response": "", "disabled": [] },
        "variables": [ { "key": "region", "initial": "eu" } ],
        "description": "Payments API"
      }
      """
    Then the call should succeed
    And the workspace file "acme/collections/payments/collection.yaml" should contain the field "auth.type" equal to "bearer"
    And the workspace file "acme/collections/payments/collection.yaml" should contain the field "headers.0.key" equal to "X-Team"
    And the workspace file "acme/collections/payments/collection.yaml" should contain the field "variables.0.key" equal to "region"
    And the workspace file "acme/collections/payments/collection.yaml" should contain the field "description" equal to "Payments API"
    When I call the MCP tool "tree" with
      """
      { "action": "settings", "id": "{{payments}}" }
      """
    Then the call should succeed
    And the response should contain the field "auth.type" equal to "bearer"
    And the response should contain the field "headers.0.value" equal to "payments"
    And the response should contain the field "scripts.pre_request" equal to "console.log('pre')"

  @examples
  Scenario: Examples are created from the last response, renamed and deleted
    Given the 1 "GET" request to "/status" returns status 200 with the following response
      """
      {"up":true}
      """
    And I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "collection", "name": "Ops" }
      """
    And I save the response field "id" as "ops"
    And I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "request", "parent_id": "{{ops}}", "name": "Status", "method": "GET", "url": "{{api}}/status" }
      """
    And I save the response field "id" as "status"
    And I call the MCP tool "send" with
      """
      { "request_id": "{{status}}" }
      """
    When I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "example", "request_id": "{{status}}", "name": "Healthy", "from_last_response": true }
      """
    Then the call should succeed
    And I save the response field "id" as "example"
    When I call the MCP tool "tree" with
      """
      { "action": "get", "id": "{{example}}" }
      """
    Then the response should contain the field "name" equal to "Healthy"
    And the response should contain the field "response.status" equal to "200"
    When I call the MCP tool "tree_edit" with
      """
      { "action": "rename", "id": "{{example}}", "name": "All good" }
      """
    Then the call should succeed
    When I call the MCP tool "delete" with
      """
      { "target": "item", "id": "{{example}}" }
      """
    Then the call should succeed
    When I call the MCP tool "tree" with
      """
      { "action": "get", "id": "{{example}}" }
      """
    Then the call should fail with code "MCP-08404"

  @docs
  Scenario: Documentation is read, planned and applied, and deletes need confirmation
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JV0000000000000000000000
      name: Acme
      order:
        - shop
      """
    And the workspace file "acme/collections/shop/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JV0000000000000000000010
      name: Shop
      order:
        - list-products.yaml
        - old-endpoint.yaml
      """
    And the workspace file "acme/collections/shop/list-products.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JV0000000000000000000011
      name: List products
      method: GET
      url: https://shop.test/products
      """
    And the workspace file "acme/collections/shop/old-endpoint.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JV0000000000000000000012
      name: Old endpoint
      method: GET
      url: https://shop.test/old
      """
    When I call the MCP tool "docs" with
      """
      { "action": "get", "collection_id": "01JV0000000000000000000010" }
      """
    Then the call should succeed
    And the response should contain the field "yaml" containing "openapi: 3.1"
    And the response should contain the field "yaml" containing "/products"
    And the response should contain the field "yaml" containing "/old"
    Given the "edited" variable is set to the field "edited" of the resource file "mcp/shop-edits.yaml"
    When I call the MCP tool "docs" with
      """
      { "action": "plan", "collection_id": "01JV0000000000000000000010", "yaml": "{{edited}}" }
      """
    Then the call should succeed
    And the response should contain the field "deleted" array length equal to 1
    And the response should contain the field "deleted.0" containing "Old endpoint"
    When I call the MCP tool "docs_edit" with
      """
      { "action": "apply", "collection_id": "01JV0000000000000000000010", "yaml": "{{edited}}" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/shop/old-endpoint.yaml" should exist
    And the workspace file "acme/collections/shop/list-products.yaml" should contain the field "description" equal to "Every product in the shop"
    When I call the MCP tool "docs_edit" with
      """
      { "action": "apply", "collection_id": "01JV0000000000000000000010", "yaml": "{{edited}}", "confirm_deletes": true }
      """
    Then the call should succeed
    And the workspace file "acme/collections/shop/old-endpoint.yaml" should not exist
    And a "workspace:tree-changed" event with the field "local_id" equal to "{{Acme.local_id}}" should have been emitted within 1000 ms
