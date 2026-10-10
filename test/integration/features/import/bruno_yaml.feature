#language: en
#utf-8

@all @xen_003 @import @bruno @bruno_yaml
Feature: Import a Bruno collection in the OpenCollection YAML format
  Bruno's YAML format (opencollection.yml, folder.yml, request .yml files and environments/*.yml)
  imports exactly like the .bru format: the same tree in sequence order, the same request fields,
  auth, docs and settings, scripts translated to pm.*, request variables and assertions turned into
  script lines, environments with missing secrets, and gRPC/WebSocket requests skipped and listed.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/yaml-collection"] }
      """
    And I save the response field "source_id" as "source"

  @detect
  Scenario: Detecting the collection lists it with its environment and counts
    Then the call should succeed
    And the response should contain the field "kind" equal to "bruno_yaml"
    And the response should contain the field "items" array length equal to 2
    And the response should contain the field "items.0.kind" equal to "collection"
    And the response should contain the field "items.0.name" equal to "Bruno Shop YAML"
    And the response should contain the field "items.0.counts.folders" equal to "2"
    And the response should contain the field "items.0.counts.requests" equal to "5"
    And the response should contain the field "items.0.counts.variables" equal to "1"
    And the response should contain the field "items.1.kind" equal to "environment"
    And the response should contain the field "items.1.name" equal to "local"
    And the response should contain the field "items.1.counts.variables" equal to "4"

  @tree
  Scenario: The tree follows the sequence numbers and skips gRPC and WebSocket requests
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "workspaces.0.created.requests" equal to "5"
    And the response field "warnings" should contain 2 items with "code" equal to "unsupported_request_type"
    And the response field "warnings" should contain an item with "code" equal to "unsupported_request_type" and "message" equal to "Socket is a websocket request, which Xenon does not support; it was skipped"
    And the response field "warnings" should contain 1 items with "code" equal to "unsupported_auth"
    And the response field "warnings" should contain 2 items with "code" equal to "missing_file"
    And the workspace file "acme/collections/bruno-shop-yaml/collection.yaml" should contain the field "order" with the items "products, orders, health.yaml"
    And the workspace file "acme/collections/bruno-shop-yaml/orders/folder.yaml" should contain the field "order" with the items "upload-invoice.yaml, query.yaml"
    And the workspace file "acme/collections/bruno-shop-yaml/products/folder.yaml" should contain the field "order" with the items "create-product.yaml, list-products.yaml"

  @collection @folders
  Scenario: The collection and folders keep headers, auth, variables, scripts and docs
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/bruno-shop-yaml/collection.yaml" should contain the field "description" equal to "The Bruno shop collection."
    And the workspace file "acme/collections/bruno-shop-yaml/collection.yaml" should contain the field "headers.1.key" equal to "X-Debug"
    And the workspace file "acme/collections/bruno-shop-yaml/collection.yaml" should contain the field "headers.1.enabled" equal to "false"
    And the workspace file "acme/collections/bruno-shop-yaml/collection.yaml" should contain the field "auth.type" equal to "bearer"
    And the workspace file "acme/collections/bruno-shop-yaml/collection.yaml" should contain the field "auth.bearer.token" equal to "{{token}}"
    And the workspace file "acme/collections/bruno-shop-yaml/collection.yaml" should contain the field "variables.0.key" equal to "api_version"
    And the workspace file "acme/collections/bruno-shop-yaml/collection.yaml" should contain the field "variables.0.value" equal to "v2"
    And the workspace file "acme/collections/bruno-shop-yaml/collection.yaml" should contain the field "scripts.preRequest" equal to
      """
      pm.request.headers.upsert({ key: "X-Start", value: Date.now() });
      """
    And the workspace file "acme/collections/bruno-shop-yaml/collection.yaml" should contain the field "scripts.postResponse" equal to
      """
      pm.test("collection ok", function() {
        pm.expect(pm.response.code).to.equal(200);
      });
      """
    And the workspace file "acme/collections/bruno-shop-yaml/collection.yaml" should contain the field "bruno.opencollection" equal to "1.0.0"
    And the workspace file "acme/collections/bruno-shop-yaml/orders/folder.yaml" should contain the field "auth.type" equal to "basic"
    And the workspace file "acme/collections/bruno-shop-yaml/orders/folder.yaml" should contain the field "auth.basic.password" equal to "{{shop_pw}}"
    And the workspace file "acme/collections/bruno-shop-yaml/products/folder.yaml" should contain the field "headers.0.key" equal to "X-Area"
    And the workspace file "acme/collections/bruno-shop-yaml/products/folder.yaml" should contain the field "scripts.postResponse" equal to
      """
      pm.variables.set("lastStatus", pm.response.code);
      """

  @requests
  Scenario: Requests keep method, URL, params, headers, body, auth, docs, settings and scripts
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/bruno-shop-yaml/products/list-products.yaml" should contain the field "method" equal to "GET"
    And the workspace file "acme/collections/bruno-shop-yaml/products/list-products.yaml" should contain the field "url" equal to "{{base_url}}/products?page=1&debug=true"
    And the workspace file "acme/collections/bruno-shop-yaml/products/list-products.yaml" should contain the field "params.1.key" equal to "debug"
    And the workspace file "acme/collections/bruno-shop-yaml/products/list-products.yaml" should contain the field "params.1.enabled" equal to "false"
    And the workspace file "acme/collections/bruno-shop-yaml/products/list-products.yaml" should contain the field "headers.1.key" equal to "X-Trace"
    And the workspace file "acme/collections/bruno-shop-yaml/products/list-products.yaml" should contain the field "headers.1.enabled" equal to "false"
    And the workspace file "acme/collections/bruno-shop-yaml/products/list-products.yaml" should contain the field "auth.type" equal to "inherit"
    And the workspace file "acme/collections/bruno-shop-yaml/products/list-products.yaml" should contain the field "description" equal to "Lists all **products**."
    And the workspace file "acme/collections/bruno-shop-yaml/products/list-products.yaml" should contain the field "settings.timeoutMs" equal to "5000"
    And the workspace file "acme/collections/bruno-shop-yaml/products/list-products.yaml" should contain the field "settings.followRedirects" equal to "false"
    And the workspace file "acme/collections/bruno-shop-yaml/products/list-products.yaml" should contain the field "bruno.settings.maxRedirects" equal to "3"
    And the workspace file "acme/collections/bruno-shop-yaml/products/list-products.yaml" should contain the field "bruno.tags" with the items "smoke"
    And the workspace file "acme/collections/bruno-shop-yaml/products/list-products.yaml" should contain the field "scripts.preRequest" equal to
      """
      pm.variables.set("category", "chairs");
      const start = pm.environment.get("start");
      pm.variables.set("start", Date.now());
      """
    And the workspace file "acme/collections/bruno-shop-yaml/products/list-products.yaml" should contain the field "scripts.postResponse" equal to
      """
      pm.environment.set("lastId", pm.response.json().items[0].id);
      // XENON: untranslated bru.setNextRequest("Create product");
      pm.variables.set("firstId", pm.response.json().items[0].id);
      pm.test("res.status eq 200", () => pm.expect(pm.response.code).to.eql(200));
      pm.test("res.body.items isArray", () => pm.expect(pm.response.json().items).to.be.a("array"));
      pm.test("has items", function() {
        pm.expect(pm.response.code).to.equal(200);
      });
      """
    And the workspace file "acme/collections/bruno-shop-yaml/products/create-product.yaml" should contain the field "pathVariables.0.key" equal to "category"
    And the workspace file "acme/collections/bruno-shop-yaml/products/create-product.yaml" should contain the field "body.raw.language" equal to "json"
    And the workspace file "acme/collections/bruno-shop-yaml/products/create-product.yaml" should contain the field "body.raw.content" equal to
      """
      {
        "name": "Chair",
        "price": 9.5
      }
      """
    And the workspace file "acme/collections/bruno-shop-yaml/health.yaml" should contain the field "method" equal to "PURGE"
    And the workspace file "acme/collections/bruno-shop-yaml/health.yaml" should contain the field "body.raw.language" equal to "text"
    And the workspace file "acme/collections/bruno-shop-yaml/health.yaml" should contain the field "body.raw.content" equal to "ping"
    And the workspace file "acme/collections/bruno-shop-yaml/health.yaml" should contain the field "settings.encodeUrl" equal to "false"
    And the workspace file "acme/collections/bruno-shop-yaml/health.yaml" should contain the field "auth.type" equal to "oauth2"
    And the workspace file "acme/collections/bruno-shop-yaml/health.yaml" should contain the field "auth.oauth2.grant_type" equal to "client_credentials"
    And the workspace file "acme/collections/bruno-shop-yaml/health.yaml" should contain the field "auth.oauth2.accessTokenUrl" equal to "https://auth.acme.test/token"
    And the workspace file "acme/collections/bruno-shop-yaml/health.yaml" should contain the field "auth.oauth2.clientId" equal to "shop"
    And the workspace file "acme/collections/bruno-shop-yaml/orders/query.yaml" should contain the field "body.mode" equal to "graphql"
    And the workspace file "acme/collections/bruno-shop-yaml/orders/query.yaml" should contain the field "body.graphql.variables" equal to
      """
      {
        "limit": 10
      }
      """
    And the workspace file "acme/collections/bruno-shop-yaml/orders/query.yaml" should contain the field "auth.type" equal to "digest"
    And the workspace file "acme/collections/bruno-shop-yaml/orders/query.yaml" should contain the field "auth.digest.username" equal to "ana"
    And the workspace file "acme/collections/bruno-shop-yaml/orders/upload-invoice.yaml" should contain the field "auth.unsupported.type" equal to "ntlm"
    And the workspace file "acme/collections/bruno-shop-yaml/orders/upload-invoice.yaml" should contain the field "body.formdata.1.src" equal to "/Users/someone/invoice.pdf"
    And the workspace file "acme/collections/bruno-shop-yaml/orders/upload-invoice.yaml" should contain the field "body.formdata.1.srcState" equal to "repick"
    And the workspace file "acme/collections/bruno-shop-yaml/orders/upload-invoice.yaml" should contain the field "body.formdata.2.enabled" equal to "false"
    And the workspace file "acme/collections/bruno-shop-yaml/orders/upload-invoice.yaml" should contain the field "bruno.formdata.0.contentType" equal to "application/json"

  @scripts @untranslated
  Scenario: Untranslated lines are listed with their file and line
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response field "warnings" should contain 1 items with "code" equal to "untranslated_script"
    And the response field "warnings" should contain an item with "code" equal to "untranslated_script" and "location" equal to "Products/List products.yml:48"
    And the response field "warnings" should contain an item with "code" equal to "untranslated_script" and "path" equal to "Bruno Shop YAML/Products/List products"

  @environments
  Scenario: The environment comes in with plain rows, its colour and missing secrets
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "1", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And I save the response field "items.0.id" as "local"
    And the workspace file "acme/environments/local.yaml" should contain the field "color" equal to "green"
    And the workspace file "acme/environments/local.yaml" should contain the field "variables" array length equal to 4
    And the workspace file "acme/environments/local.yaml" should contain the field "variables.1.key" equal to "debug"
    And the workspace file "acme/environments/local.yaml" should contain the field "variables.1.value" equal to "true"
    And the workspace file "acme/environments/local.yaml" should contain the field "variables.1.enabled" equal to "false"
    And the workspace file "acme/environments/local.yaml" should contain the field "variables.2.secret" equal to "true"
    And the workspace file "acme/environments/local.yaml" should not contain the field "variables.2.value"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{local}}" }
      """
    Then the response should contain the field "rows.2.secret_state" equal to "missing"
