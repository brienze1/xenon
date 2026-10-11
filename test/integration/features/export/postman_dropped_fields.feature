#language: en
#utf-8

@all @xen_006 @export @postman_export @postman_dropped_fields
Feature: Xenon-only data is dropped from Postman exports and listed in the summary
  Postman has no place for some Xenon data: inherit-parent-scripts off below the exported root,
  confirm-before-send, environment colours, Xenon's useToken and autoRefresh OAuth 2.0 settings,
  Bruno-only fields, binary example bodies, per-request timeouts, collection and folder headers
  (copied onto the requests below) and Xenon's {{vault:…}} lookups. Each is dropped (or changed)
  and listed in the summary with the item it applied to. Data Postman can hold is restored as
  Postman had it: auth types Xenon does not support yet, disabled scripts and disabled bodies.
  Credentials typed literally into auth fields are exported as typed and the summary warns.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the resource folder "export/shop" is copied to the workspace folder "acme"
    And the "shop" variable is set to "01JX0000000000000000000001"

  @summary
  Scenario: The summary lists every dropped Xenon-only field of a collection
    Given the workspace file "acme/collections/shop-api/orders/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JX0000000000000000000022
      name: Login
      method: POST
      url: "{{vault:login_url}}/login"
      inheritScripts: false
      settings:
        timeoutMs: 1500
      auth:
        type: oauth2
        oauth2:
          grant_type: client_credentials
          accessTokenUrl: https://auth.acme.test/token
          clientId: shop
          clientSecret: "{{client_secret}}"
          useToken: id_token
          autoRefresh: false
      bruno:
        seq: 3
      """
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{shop}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/shop.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "shop.postman_collection.json" should be a valid Postman v2.1 collection
    And the response field "dropped" should contain an item with "item_path" equal to "Shop API" and "field" equal to "headers"
    And the response field "dropped" should contain an item with "item_path" equal to "Shop API/Orders" and "field" equal to "headers"
    And the response field "dropped" should contain an item with "item_path" equal to "Shop API/Products/List products" and "field" equal to "examples.Logo.body"
    And the response field "dropped" should contain an item with "item_path" equal to "Shop API/Orders/Login" and "field" equal to "inheritScripts"
    And the response field "dropped" should contain an item with "item_path" equal to "Shop API/Orders/Login" and "field" equal to "settings.timeoutMs"
    And the response field "dropped" should contain an item with "item_path" equal to "Shop API/Orders/Login" and "field" equal to "auth.oauth2.useToken, auth.oauth2.autoRefresh"
    And the response field "dropped" should contain an item with "item_path" equal to "Shop API/Orders/Login" and "field" equal to "bruno"
    And the response field "dropped" should contain an item with "item_path" equal to "Shop API/Orders/Login" and "field" equal to "url"
    And the response should contain the field "dropped" array length equal to 8
    And the file "shop.postman_collection.json" should contain the field "item.1.item.1.request.auth.type" equal to "oauth2"
    And the file "shop.postman_collection.json" should not contain "useToken"
    And the file "shop.postman_collection.json" should not contain "autoRefresh"
    And the file "shop.postman_collection.json" should not contain "seq"
    And the file "shop.postman_collection.json" should not contain "timeoutMs"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.1.request.url.raw" equal to "{{vault:login_url}}/login"

  @environment
  Scenario: Environment colour and confirm-before-send are dropped and listed
    Given the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JX0000000000000000000102
      name: prod
      color: red
      confirmBeforeSend: true
      variables:
        - key: base_url
          value: https://acme.test
      """
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "01JX0000000000000000000102",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/prod.postman_environment.json"
      }
      """
    Then the call should succeed
    And the response should contain the field "dropped" array length equal to 2
    And the response field "dropped" should contain an item with "item_path" equal to "prod" and "field" equal to "color"
    And the response field "dropped" should contain an item with "item_path" equal to "prod" and "field" equal to "confirmBeforeSend"
    And the file "prod.postman_environment.json" should not contain "confirmBeforeSend"
    And the file "prod.postman_environment.json" should not contain "red"

  @auth @bruno
  Scenario: A Bruno auth type Postman cannot hold is exported as No Auth with a note
    Given the workspace file "acme/collections/shop-api/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JX0000000000000000000031
      name: Ping
      method: GET
      url: "{{base_url}}/ping"
      auth:
        type: unsupported
        unsupported:
          source: bruno
          type: wsse
          data:
            username: ana
            password: "{{wsse_password}}"
      """
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "request", "item_id": "01JX0000000000000000000031",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/ping.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "ping.postman_collection.json" should be a valid Postman v2.1 collection
    And the file "ping.postman_collection.json" should contain the field "item.0.request.auth.type" equal to "noauth"
    And the file "ping.postman_collection.json" should not contain "wsse"
    And the response field "dropped" should contain an item with "item_path" equal to "Shop API/Ping" and "field" equal to "auth"

  @postman_only
  Scenario: Postman-only auth types, disabled scripts and disabled bodies come back as Postman had them
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/full-v2.1.json"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    And I save the response field "items.0.id" as "acme"
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{acme}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/acme.postman_collection.json"
      }
      """
    Then the call should succeed
    And the response should contain the field "dropped" array length equal to 0
    And the file "acme.postman_collection.json" should contain the field "item.0.item.2.request.auth.type" equal to "hawk"
    And the file "acme.postman_collection.json" should contain the field "item.0.item.2.request.auth.hawk.0.key" equal to "authId"
    And the file "acme.postman_collection.json" should contain the field "item.0.item.2.request.auth.hawk.0.value" equal to "hawk-id"
    And the file "acme.postman_collection.json" should contain the field "item.0.event.0.disabled" equal to "true"
    And the file "acme.postman_collection.json" should contain the field "item.1.item.6.name" equal to "Disabled body"
    And the file "acme.postman_collection.json" should contain the field "item.1.item.6.request.body.disabled" equal to "true"
    And the file "acme.postman_collection.json" should contain the field "item.1.item.6.request.body.raw" equal to "ignored"

  @warnings
  Scenario: Credentials typed literally into auth fields are exported as typed with a warning
    Given the workspace file "acme/collections/shop-api/products/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: 01JX0000000000000000000002
      name: Products
      order:
        - list-products.yaml
        - get-product.yaml
        - create-product.yaml
      auth:
        type: basic
        basic:
          username: admin
          password: hunter2
      """
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{shop}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/shop.postman_collection.json"
      }
      """
    Then the call should succeed
    And the response should contain the field "warnings" array length equal to 1
    And the response field "warnings" should contain an item with "code" equal to "literal_credential" and "item_path" equal to "Shop API/Products"
    And the response should contain the field "warnings.0.message" containing "basic.password"
    And the file "shop.postman_collection.json" should contain the field "item.0.auth.basic.1.value" equal to "hunter2"
