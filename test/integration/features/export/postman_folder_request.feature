#language: en
#utf-8

@all @xen_006 @export @postman_export @postman_folder_request
Feature: Export a folder or a single request as a Postman collection
  A folder exports as a Postman collection v2.1 named after the folder; a request exports as a
  collection holding only that request. What the item inherits in Xenon is placed on the exported
  root so it behaves the same in Postman: the nearest ancestor auth when the item inherits auth,
  the ancestor scripts that would run (respecting the inherit-parent-scripts toggles and disabled
  scripts) before the item's own, the collection variables, and the ancestor headers copied onto
  each request.

  Background:
    Given the workspace "Acme" exists
    And the resource folder "export/shop" is copied to the workspace folder "acme"
    And the "products" variable is set to "01JX0000000000000000000002"
    And the "orders" variable is set to "01JX0000000000000000000003"
    And the "list" variable is set to "01JX0000000000000000000011"
    And the "get" variable is set to "01JX0000000000000000000012"

  @folder
  Scenario: A folder becomes a collection carrying the inherited auth, scripts and variables
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "folder", "item_id": "{{products}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/products.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "products.postman_collection.json" should be a valid Postman v2.1 collection
    And the response should contain the field "counts.folders" equal to "1"
    And the response should contain the field "counts.requests" equal to "3"
    And the file "products.postman_collection.json" should contain the field "info.name" equal to "Products"
    And the file "products.postman_collection.json" should contain the field "info._postman_id" equal to "{{products}}"
    And the file "products.postman_collection.json" should contain the field "info.description" equal to "Catalogue endpoints"
    And the file "products.postman_collection.json" should contain the field "auth.type" equal to "bearer"
    And the file "products.postman_collection.json" should contain the field "auth.bearer.0.value" equal to "{{token}}"
    And the file "products.postman_collection.json" should contain the field "event" array length equal to 3
    And the file "products.postman_collection.json" should contain the field "event.0.listen" equal to "prerequest"
    And the file "products.postman_collection.json" should contain the field "event.0.script.exec.0" equal to
      """
      pm.request.headers.add({ key: "X-From-Collection", value: "yes" });
      """
    And the file "products.postman_collection.json" should contain the field "event.1.listen" equal to "test"
    And the file "products.postman_collection.json" should contain the field "event.2.listen" equal to "prerequest"
    And the file "products.postman_collection.json" should contain the field "event.2.script.exec.0" equal to
      """
      pm.request.headers.add({ key: "X-From-Folder", value: "products" });
      """
    And the file "products.postman_collection.json" should contain the field "variable" array length equal to 4
    And the file "products.postman_collection.json" should contain the field "variable.0.key" equal to "token"
    And the file "products.postman_collection.json" should contain the field "item" array length equal to 3
    And the file "products.postman_collection.json" should contain the field "item.0.name" equal to "List products"
    And the file "products.postman_collection.json" should contain the field "item.0.request.header.0.key" equal to "X-Tenant"
    And the file "products.postman_collection.json" should contain the field "item.1.request.auth.type" equal to "apikey"

  @folder @auth
  Scenario: A folder with its own auth keeps it
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "folder", "item_id": "{{orders}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/orders.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "orders.postman_collection.json" should be a valid Postman v2.1 collection
    And the file "orders.postman_collection.json" should contain the field "auth.type" equal to "noauth"
    And the file "orders.postman_collection.json" should contain the field "event" array length equal to 2
    And the file "orders.postman_collection.json" should contain the field "item.0.request.header.0.key" equal to "X-Tenant"
    And the file "orders.postman_collection.json" should contain the field "item.0.request.header.1.key" equal to "X-Area"

  @folder @scripts
  Scenario: A folder that does not inherit parent scripts exports only its own scripts
    Given the workspace file "acme/collections/shop-api/products/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{products}}
      name: Products
      inheritScripts: false
      order:
        - list-products.yaml
        - get-product.yaml
        - create-product.yaml
      scripts:
        preRequest: pm.variables.set("folder", "yes");
      """
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "folder", "item_id": "{{products}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/products.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "products.postman_collection.json" should contain the field "event" array length equal to 1
    And the file "products.postman_collection.json" should contain the field "event.0.script.exec.0" equal to
      """
      pm.variables.set("folder", "yes");
      """
    And the response field "dropped" should contain 0 items with "field" equal to "inheritScripts"

  @folder @scripts
  Scenario: Disabled ancestor scripts are not carried
    Given the workspace file "acme/collections/shop-api/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JX0000000000000000000001
      name: Shop API
      order:
        - products
        - orders
        - ping.yaml
      scripts:
        preRequest: pm.variables.set("collection", "yes");
        postResponse: pm.test("collection", function () {});
        disabled:
          - postResponse
      """
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "folder", "item_id": "{{orders}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/orders.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "orders.postman_collection.json" should contain the field "event" array length equal to 1
    And the file "orders.postman_collection.json" should contain the field "event.0.listen" equal to "prerequest"
    And the file "orders.postman_collection.json" should not contain the field "variable"

  @request
  Scenario: A request that inherits auth becomes a collection with that auth at the root
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "request", "item_id": "{{list}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/list.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "list.postman_collection.json" should be a valid Postman v2.1 collection
    And the response should contain the field "counts.requests" equal to "1"
    And the file "list.postman_collection.json" should contain the field "info.name" equal to "List products"
    And the file "list.postman_collection.json" should contain the field "auth.type" equal to "bearer"
    And the file "list.postman_collection.json" should contain the field "event" array length equal to 3
    And the file "list.postman_collection.json" should contain the field "event.2.script.exec.0" equal to
      """
      pm.request.headers.add({ key: "X-From-Folder", value: "products" });
      """
    And the file "list.postman_collection.json" should contain the field "variable" array length equal to 4
    And the file "list.postman_collection.json" should contain the field "item" array length equal to 1
    And the file "list.postman_collection.json" should contain the field "item.0.name" equal to "List products"
    And the file "list.postman_collection.json" should contain the field "item.0.id" equal to "{{list}}"
    And the file "list.postman_collection.json" should contain the field "item.0.event.0.listen" equal to "test"
    And the file "list.postman_collection.json" should contain the field "item.0.request.header.0.key" equal to "X-Tenant"
    And the file "list.postman_collection.json" should contain the field "item.0.response" array length equal to 2

  @request @auth
  Scenario: A request with its own auth keeps it and the root does not add one
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "request", "item_id": "{{get}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/get.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "get.postman_collection.json" should be a valid Postman v2.1 collection
    And the file "get.postman_collection.json" should not contain the field "auth"
    And the file "get.postman_collection.json" should contain the field "item.0.request.auth.type" equal to "apikey"
