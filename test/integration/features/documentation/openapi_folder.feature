#language: en
#utf-8

@all @xen_007 @documentation @documentation_folder
Feature: A folder is documented, and edited, as an OpenAPI document of what it holds
  Every folder, at any depth, has its own Documentation tab: the same OpenAPI 3.1 document a
  collection has, covering only the requests and sub-folders inside that folder. Its info title
  and description are the folder's name and description, its sub-folders become tags relative to
  it ("Admin / Audit") grouped under "x-tagGroups", and the auth its requests really use is still
  resolved through the folders and the collection above it. Saving it works both ways like a
  collection's, but only inside the folder: operations update its requests, new operations are
  created in it (in the sub-folder their tag names), and only its requests can be deleted.

  Background:
    Given the workspace "Acme" exists
    And the "shop" variable is set to "01JKF000000000000000000001"
    And the "products" variable is set to "01JKF000000000000000000002"
    And the "admin" variable is set to "01JKF000000000000000000003"
    And the "audit" variable is set to "01JKF000000000000000000004"
    And the "orders" variable is set to "01JKF000000000000000000005"
    And the "list" variable is set to "01JKF000000000000000000010"
    And the "create" variable is set to "01JKF000000000000000000011"
    And the "log" variable is set to "01JKF000000000000000000012"
    And the "listOrders" variable is set to "01JKF000000000000000000013"
    And the "ping" variable is set to "01JKF000000000000000000014"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JKF000000000000000000000
      name: Acme
      order:
        - shop
      """
    And the workspace file "acme/collections/shop/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{shop}}
      name: Shop
      description: The whole shop.
      order:
        - products
        - orders
        - ping.yaml
      auth:
        type: bearer
        bearer:
          token: s3cr3t-collection-token
      """
    And the workspace file "acme/collections/shop/products/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{products}}
      name: Products
      description: The product catalog.
      order:
        - list-products.yaml
        - admin
      """
    And the workspace file "acme/collections/shop/products/admin/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{admin}}
      name: Admin
      description: Back-office calls.
      order:
        - create-product.yaml
        - audit
      """
    And the workspace file "acme/collections/shop/products/admin/audit/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{audit}}
      name: Audit
      order:
        - audit-log.yaml
      """
    And the workspace file "acme/collections/shop/orders/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{orders}}
      name: Orders
      order:
        - list-orders.yaml
      """
    And the workspace file "acme/collections/shop/products/list-products.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{list}}
      name: List products
      method: GET
      url: "{{baseUrl}}/products"
      auth:
        type: inherit
      """
    And the workspace file "acme/collections/shop/products/admin/create-product.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create product
      method: POST
      url: "{{baseUrl}}/products"
      auth:
        type: inherit
      """
    And the workspace file "acme/collections/shop/products/admin/audit/audit-log.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{log}}
      name: Audit log
      method: GET
      url: "{{baseUrl}}/audit"
      auth:
        type: inherit
      """
    And the workspace file "acme/collections/shop/orders/list-orders.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{listOrders}}
      name: List orders
      method: GET
      url: "{{baseUrl}}/orders"
      auth:
        type: inherit
      """
    And the workspace file "acme/collections/shop/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{ping}}
      name: Ping
      method: GET
      url: "{{baseUrl}}/ping"
      auth:
        type: none
      """
    And the temp folder "workspaces/acme" is copied to "before"

  @golden
  Scenario: A folder's documentation covers only what is inside it, with its sub-folders as grouped tags
    When I call "TreeService.GetDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{products}}" }
      """
    Then the call should succeed
    And the response field "yaml" should equal the resource file "documentation/products-folder.openapi.yaml"
    And the response should not contain "s3cr3t"
    And the response should contain the field "document.title" equal to "Products"
    And the response should contain the field "document.description" equal to "The product catalog."
    And the response should contain the field "document.operations" array length equal to 3
    And the response should contain the field "document.operations.0.request_id" equal to "{{list}}"
    And the response should contain the field "document.operations.0.tag" equal to ""
    And the response should contain the field "document.operations.0.security.0" equal to "bearerAuth"
    And the response should contain the field "document.operations.1.request_id" equal to "{{create}}"
    And the response should contain the field "document.operations.1.tag" equal to "Admin"
    And the response should contain the field "document.operations.2.request_id" equal to "{{log}}"
    And the response should contain the field "document.operations.2.tag" equal to "Admin / Audit"
    And the response should contain the field "document.tags.0.name" equal to "Admin"
    And the response should contain the field "document.tags.0.description" equal to "Back-office calls."
    And the response should contain the field "document.tag_groups" array length equal to 1
    And the response should contain the field "document.tag_groups.0.name" equal to "Admin"
    And the response should contain the field "document.tag_groups.0.tags" with the items "Admin, Admin / Audit"

  @depth
  Scenario: A folder at any depth has its own documentation
    When I call "TreeService.GetDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{audit}}" }
      """
    Then the call should succeed
    And the response should contain the field "document.title" equal to "Audit"
    And the response should contain the field "document.operations" array length equal to 1
    And the response should contain the field "document.operations.0.request_id" equal to "{{log}}"
    And the response should contain the field "document.operations.0.tag" equal to ""
    And the response should contain the field "document.operations.0.security.0" equal to "bearerAuth"
    And the response should contain the field "document.tags" array length equal to 0
    And the response should contain the field "document.tag_groups" array length equal to 0

  @requests
  Scenario: Requests have no documentation of their own
    When I call "TreeService.GetDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{list}}" }
      """
    Then the call should fail with code "DOC-05400"

  @roundtrip
  Scenario: Saving a folder's generated document without edits changes nothing
    When I call "TreeService.GetDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{products}}" }
      """
    Then the call should succeed
    And I save the response field "yaml" as "generated"
    When I call "TreeService.PlanDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{products}}", "yaml": "{{generated}}" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Products"
    And the response should contain the field "collection_changed" equal to "false"
    And the response should contain the field "updated" array length equal to 0
    And the response should contain the field "created" array length equal to 0
    And the response should contain the field "deleted" array length equal to 0
    When I call "TreeService.ApplyDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{products}}", "yaml": "{{generated}}" }
      """
    Then the call should succeed
    And the response should contain the field "yaml" equal to "{{generated}}"
    And the folder "workspaces/acme" should equal the folder "before"

  @scoped
  Scenario: Saving a folder's document updates, creates and deletes requests only inside that folder
    When I call "TreeService.PlanDocumentation" with resource fields
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{products}}", "yaml": {{resource:documentation/products-edits.yaml#edits}} }
      """
    Then the call should succeed
    And the response should contain the field "collection_changed" equal to "true"
    And the response should contain the field "updated" array length equal to 1
    And the response should contain the field "updated.0.id" equal to "{{list}}"
    And the response should contain the field "created" array length equal to 2
    And the response should contain the field "created.0.name" equal to "Sales report"
    And the response should contain the field "created.0.folder" equal to "Admin / Reports"
    And the response should contain the field "created.1.name" equal to "Search products"
    And the response should contain the field "created.1.folder" equal to ""
    And the response should contain the field "deleted" array length equal to 1
    And the response should contain the field "deleted.0.id" equal to "{{log}}"
    When I call "TreeService.ApplyDocumentation" with resource fields
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{products}}", "yaml": {{resource:documentation/products-edits.yaml#edits}}, "confirm_deletions": true }
      """
    Then the call should succeed
    And the workspace file "acme/collections/shop/products/folder.yaml" should contain the field "description" equal to "Everything we sell."
    And the workspace file "acme/collections/shop/products/folder.yaml" should contain the field "name" equal to "Products"
    And the workspace file "acme/collections/shop/products/list-products.yaml" should contain the field "description" equal to "Lists every product."
    And the workspace file "acme/collections/shop/products/admin/reports/folder.yaml" should contain the field "name" equal to "Reports"
    And the workspace file "acme/collections/shop/products/admin/reports/sales-report.yaml" should contain the field "url" equal to "{{baseUrl}}/reports/sales"
    And the workspace file "acme/collections/shop/products/search-products.yaml" should contain the field "method" equal to "GET"
    And the workspace file "acme/collections/shop/products/search-products.yaml" should contain the field "url" equal to "{{baseUrl}}/products/search"
    And the workspace file "acme/collections/shop/products/admin/audit/audit-log.yaml" should not exist
    And the folder "workspaces/acme/collections/shop/collection.yaml" should equal the folder "before/collections/shop/collection.yaml"
    And the folder "workspaces/acme/collections/shop/orders" should equal the folder "before/collections/shop/orders"
    And the folder "workspaces/acme/collections/shop/ping.yaml" should equal the folder "before/collections/shop/ping.yaml"
    And the folder "workspaces/acme/collections/shop/products/admin/create-product.yaml" should equal the folder "before/collections/shop/products/admin/create-product.yaml"

  @rename
  Scenario: A new info title renames the folder like any other rename
    When I call "TreeService.ApplyDocumentation" with resource fields
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{products}}", "yaml": {{resource:documentation/products-edits.yaml#rename}} }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Catalog"
    And the response should contain the field "collection_changed" equal to "true"
    And the response should contain the field "document.title" equal to "Catalog"
    And the workspace folder "acme/collections/shop" should contain the entries "catalog, collection.yaml, orders, ping.yaml"
    And the workspace file "acme/collections/shop/catalog/folder.yaml" should contain the field "id" equal to "{{products}}"
    And the workspace file "acme/collections/shop/catalog/folder.yaml" should contain the field "name" equal to "Catalog"
    And the workspace file "acme/collections/shop/collection.yaml" should contain the field "order" with the items "catalog, orders, ping.yaml"
