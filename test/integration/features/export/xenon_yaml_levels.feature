#language: en
#utf-8

@all @xen_006 @export @xenon_export @xenon_yaml_levels
Feature: Export as Xenon YAML at every level
  A workspace exports as a zip of its folder (root folder = the workspace folder name) without the
  local-only and staging files (.xenon/, .xenon-import-*/, *.xenon-tmp). A collection or folder
  exports as a zip of its directory or as one bundled YAML file holding the whole subtree in
  sidebar order. Zips carry an xenon-export.yaml manifest at their root. A request or an
  environment exports as its YAML file, byte for byte. No export carries a secret value, a current
  value, a token or a cookie unless asked to.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the current time is "2026-10-11T10:00:00Z"
    And the resource folder "export/shop" is copied to the workspace folder "acme"
    And the workspace file "acme/environments/local.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JX0000000000000000000101
      name: local
      variables:
        - key: base_url
          value: https://local.acme.test
        - key: api_key
          secret: true
      """
    And the workspace file "acme/.xenon/state.yaml" contains
      """
      local: only
      """
    And the workspace file "acme/.xenon-import-01JX/0/junk.yaml" contains
      """
      staged: true
      """
    And the workspace file "acme/.workspace.yaml.xenon-tmp" contains
      """
      half: written
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:01JX0000000000000000000101", "key": "api_key", "value": "levels-SECRET-key" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:01JX0000000000000000000101", "key": "base_url", "value": "http://127.0.0.1:9999" }
      """

  @workspace @zip
  Scenario: The workspace zip holds the workspace folder without local-only and staging files
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "workspace",
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/acme.zip"
      }
      """
    Then the call should succeed
    And the response should contain the field "secrets" equal to "omitted"
    And the response should contain the field "counts.collections" equal to "1"
    And the response should contain the field "counts.requests" equal to "9"
    And the response should contain the field "counts.environments" equal to "1"
    And the zip file "acme.zip" should contain exactly the entries "xenon-export.yaml, acme/.gitignore, acme/workspace.yaml, acme/environments/local.yaml, acme/files/invoice.txt, acme/files/payload.bin, acme/collections/shop-api/collection.yaml, acme/collections/shop-api/ping.yaml, acme/collections/shop-api/products/folder.yaml, acme/collections/shop-api/products/list-products.yaml, acme/collections/shop-api/products/get-product.yaml, acme/collections/shop-api/products/create-product.yaml, acme/collections/shop-api/orders/folder.yaml, acme/collections/shop-api/orders/upload-invoice.yaml, acme/collections/shop-api/orders/login.yaml, acme/collections/shop-api/orders/graphql.yaml, acme/collections/shop-api/orders/raw-file.yaml, acme/collections/shop-api/orders/signed.yaml"
    And the file "acme.zip" should not contain "levels-SECRET-key"
    And the file "acme.zip" should not contain "127.0.0.1:9999"
    When the zip file "acme.zip" is extracted to "unzipped"
    Then the file "unzipped/xenon-export.yaml" should contain the field "xenon" equal to "1"
    And the file "unzipped/xenon-export.yaml" should contain the field "kind" equal to "export"
    And the file "unzipped/xenon-export.yaml" should contain the field "root" equal to "workspace"
    And the file "unzipped/xenon-export.yaml" should contain the field "exportedAt" equal to "2026-10-11T10:00:00Z"
    And the file "unzipped/xenon-export.yaml" should contain the field "exportedBy" matching "^Xenon/[0-9]"
    And the file "unzipped/xenon-export.yaml" should contain the field "currentValuesAsInitial" equal to "false"
    And the file "unzipped/xenon-export.yaml" should contain the field "secrets" equal to "null"
    And the folder "unzipped/acme/collections" should equal the folder "workspaces/acme/collections"
    And the file "unzipped/acme/environments/local.yaml" should equal the file "workspaces/acme/environments/local.yaml"
    And the file "unzipped/acme/workspace.yaml" should equal the file "workspaces/acme/workspace.yaml"
    And the file "unzipped/acme/files/payload.bin" should equal the resource file "export/shop/files/payload.bin"

  @collection @zip
  Scenario: A collection zip holds its directory and a manifest
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "01JX0000000000000000000001",
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/shop-api.zip"
      }
      """
    Then the call should succeed
    And the response should contain the field "counts.collections" equal to "1"
    And the response should contain the field "counts.folders" equal to "2"
    And the response should contain the field "counts.requests" equal to "9"
    And the zip file "shop-api.zip" should contain exactly the entries "xenon-export.yaml, shop-api/collection.yaml, shop-api/ping.yaml, shop-api/products/folder.yaml, shop-api/products/list-products.yaml, shop-api/products/get-product.yaml, shop-api/products/create-product.yaml, shop-api/orders/folder.yaml, shop-api/orders/upload-invoice.yaml, shop-api/orders/login.yaml, shop-api/orders/graphql.yaml, shop-api/orders/raw-file.yaml, shop-api/orders/signed.yaml"
    When the zip file "shop-api.zip" is extracted to "unzipped"
    Then the file "unzipped/xenon-export.yaml" should contain the field "root" equal to "collection"
    And the folder "unzipped/shop-api" should equal the folder "workspaces/acme/collections/shop-api"

  @folder @zip
  Scenario: A folder zip holds only that folder
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "folder", "item_id": "01JX0000000000000000000002",
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/products.zip"
      }
      """
    Then the call should succeed
    And the response should contain the field "counts.folders" equal to "1"
    And the response should contain the field "counts.requests" equal to "3"
    And the zip file "products.zip" should contain exactly the entries "xenon-export.yaml, products/folder.yaml, products/list-products.yaml, products/get-product.yaml, products/create-product.yaml"
    When the zip file "products.zip" is extracted to "unzipped"
    Then the file "unzipped/xenon-export.yaml" should contain the field "root" equal to "folder"
    And the folder "unzipped/products" should equal the folder "workspaces/acme/collections/shop-api/products"

  @collection @bundle
  Scenario: A collection bundle holds the whole subtree in sidebar order
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "01JX0000000000000000000001",
        "format": "xenon_yaml", "shape": "bundle", "target_path": "{{temp_root}}/shop-api.yaml"
      }
      """
    Then the call should succeed
    And the response should contain the field "counts.requests" equal to "9"
    And the file "shop-api.yaml" should contain the field "xenon" equal to "1"
    And the file "shop-api.yaml" should contain the field "kind" equal to "bundle"
    And the file "shop-api.yaml" should contain the field "root" equal to "collection"
    And the file "shop-api.yaml" should contain the field "exportedAt" equal to "2026-10-11T10:00:00Z"
    And the file "shop-api.yaml" should contain the field "exportedBy" matching "^Xenon/[0-9]"
    And the file "shop-api.yaml" should contain the field "secrets" equal to "null"
    And the file "shop-api.yaml" should contain the field "item.kind" equal to "collection"
    And the file "shop-api.yaml" should contain the field "item.id" equal to "01JX0000000000000000000001"
    And the file "shop-api.yaml" should contain the field "item.name" equal to "Shop API"
    And the file "shop-api.yaml" should contain the field "item.variables.3.key" equal to "api_secret"
    And the file "shop-api.yaml" should not contain the field "item.variables.3.value"
    And the file "shop-api.yaml" should contain the field "item.children" array length equal to 3
    And the file "shop-api.yaml" should contain the field "item.children.0.kind" equal to "folder"
    And the file "shop-api.yaml" should contain the field "item.children.0.name" equal to "Products"
    And the file "shop-api.yaml" should contain the field "item.children.0.children" array length equal to 3
    And the file "shop-api.yaml" should contain the field "item.children.0.children.0.name" equal to "List products"
    And the file "shop-api.yaml" should contain the field "item.children.0.children.1.name" equal to "Get product"
    And the file "shop-api.yaml" should contain the field "item.children.0.children.2.name" equal to "Create product"
    And the file "shop-api.yaml" should contain the field "item.children.0.children.2.kind" equal to "request"
    And the file "shop-api.yaml" should contain the field "item.children.1.name" equal to "Orders"
    And the file "shop-api.yaml" should contain the field "item.children.1.children" array length equal to 5
    And the file "shop-api.yaml" should contain the field "item.children.1.children.4.name" equal to "Signed"
    And the file "shop-api.yaml" should contain the field "item.children.2.kind" equal to "request"
    And the file "shop-api.yaml" should contain the field "item.children.2.name" equal to "Ping"

  @folder @bundle
  Scenario: A folder bundle holds the folder and its requests
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "folder", "item_id": "01JX0000000000000000000003",
        "format": "xenon_yaml", "shape": "bundle", "target_path": "{{temp_root}}/orders.yaml"
      }
      """
    Then the call should succeed
    And the file "orders.yaml" should contain the field "root" equal to "folder"
    And the file "orders.yaml" should contain the field "item.kind" equal to "folder"
    And the file "orders.yaml" should contain the field "item.name" equal to "Orders"
    And the file "orders.yaml" should contain the field "item.children" array length equal to 5
    And the file "orders.yaml" should contain the field "item.children.0.name" equal to "Upload invoice"

  @request @file
  Scenario: A request exports as its YAML file, byte for byte
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "request", "item_id": "01JX0000000000000000000013",
        "format": "xenon_yaml", "shape": "file", "target_path": "{{temp_root}}/create-product.yaml"
      }
      """
    Then the call should succeed
    And the response should contain the field "counts.requests" equal to "1"
    And the file "create-product.yaml" should equal the resource file "export/shop/collections/shop-api/products/create-product.yaml"

  @environment @file
  Scenario: An environment exports as its YAML file without current or secret values
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "01JX0000000000000000000101",
        "format": "xenon_yaml", "shape": "file", "target_path": "{{temp_root}}/local.yaml"
      }
      """
    Then the call should succeed
    And the response should contain the field "counts.environments" equal to "1"
    And the response should contain the field "counts.variables" equal to "2"
    And the response should contain the field "secrets" equal to "omitted"
    And the file "local.yaml" should equal the file "workspaces/acme/environments/local.yaml"
    And the file "local.yaml" should not contain "levels-SECRET-key"
    And the file "local.yaml" should not contain "127.0.0.1:9999"

  @errors
  Scenario Outline: An item kind only exports in its own shapes (<kind> as <shape>)
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "<kind>", "item_id": "<id>",
        "format": "xenon_yaml", "shape": "<shape>", "target_path": "{{temp_root}}/wrong.out"
      }
      """
    Then the call should fail with code "EXP-00400"
    And the file "wrong.out" should not exist

    Examples:
      | kind        | id                         | shape  |
      | request     | 01JX0000000000000000000013 | zip    |
      | environment | 01JX0000000000000000000101 | bundle |
      | workspace   |                            | file   |
      | collection  | 01JX0000000000000000000001 | file   |
