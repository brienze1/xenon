#language: en
#utf-8

@all @xen_003 @import @bruno @bruno_bru
Feature: Import a Bruno collection in the .bru format
  A Bruno collection folder (bruno.json, collection.bru, folder.bru, request .bru files and
  environments/*.bru) becomes a Xenon collection with the same tree in Bruno's sequence order and
  its environments. Requests keep method (custom ones too), URL, query and path params, headers
  (disabled rows included), body, auth (inherit and none included), docs as the description and
  the settings Xenon has; Bruno-only fields go to the bruno: block. Bruno keeps secret values out
  of the collection, so secret variables come in as missing secrets. gRPC and WebSocket requests
  are skipped and listed. A collection mixing .bru and YAML files is refused.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available

  @detect
  Scenario: Detecting the collection lists it with its environments and counts
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "bruno_bru"
    And the response should contain the field "items" array length equal to 3
    And the response should contain the field "items.0.kind" equal to "collection"
    And the response should contain the field "items.0.name" equal to "Bruno Shop"
    And the response should contain the field "items.0.counts.folders" equal to "2"
    And the response should contain the field "items.0.counts.requests" equal to "8"
    And the response should contain the field "items.0.counts.variables" equal to "2"
    And the response should contain the field "items.0.counts.scripts" equal to "6"
    And the response should contain the field "items.1.kind" equal to "environment"
    And the response should contain the field "items.1.name" equal to "local"
    And the response should contain the field "items.1.counts.variables" equal to "4"
    And the response should contain the field "items.2.kind" equal to "environment"
    And the response should contain the field "items.2.name" equal to "prod"

  @tree
  Scenario: The tree follows Bruno's sequence numbers and skips gRPC and WebSocket requests
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the response should contain the field "workspaces.0.created.collections" equal to "1"
    And the response should contain the field "workspaces.0.created.folders" equal to "2"
    And the response should contain the field "workspaces.0.created.requests" equal to "8"
    And the response field "warnings" should contain 2 items with "code" equal to "unsupported_request_type"
    And the response field "warnings" should contain an item with "code" equal to "unsupported_request_type" and "message" equal to "Stream is a grpc request, which Xenon does not support; it was skipped"
    And the response field "warnings" should contain an item with "code" equal to "unsupported_request_type" and "message" equal to "Socket is a ws request, which Xenon does not support; it was skipped"
    And the response field "warnings" should contain an item with "code" equal to "unsupported_request_type" and "path" equal to "Bruno Shop/Orders"
    And the response field "warnings" should contain 1 items with "code" equal to "unsupported_auth"
    And the response field "warnings" should contain 3 items with "code" equal to "missing_file"
    And the response field "warnings" should contain 3 items with "code" equal to "untranslated_script"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "order" with the items "products, orders, health.yaml"
    And the workspace file "acme/collections/bruno-shop/orders/folder.yaml" should contain the field "order" with the items "upload-invoice.yaml, raw-file.yaml, query.yaml, form.yaml, xml.yaml"
    And the workspace file "acme/collections/bruno-shop/products/folder.yaml" should contain the field "order" with the items "create-product.yaml, list-products.yaml"
    And the workspace file "acme/collections/bruno-shop/orders/stream.yaml" should not exist
    And the workspace file "acme/collections/bruno-shop/orders/socket.yaml" should not exist
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.name" equal to "Bruno Shop"
    And the response should contain the field "0.children.0.name" equal to "Products"
    And the response should contain the field "0.children.1.name" equal to "Orders"
    And the response should contain the field "0.children.2.name" equal to "Health"
    And the response should contain the field "0.children.2.method" equal to "PURGE"

  @collection
  Scenario: The collection keeps its headers, auth, variables, scripts, docs and bruno.json fields
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "name" equal to "Bruno Shop"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "id" matching "^[0-9A-HJKMNP-TV-Z]{26}$"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "description" equal to "The Bruno shop collection."
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "headers.0.key" equal to "X-Client"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "headers.0.value" equal to "xenon"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "headers.0.enabled" equal to "true"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "headers.1.key" equal to "X-Debug"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "headers.1.enabled" equal to "false"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "auth.type" equal to "bearer"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "auth.bearer.token" equal to "{{token}}"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "variables" array length equal to 2
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "variables.0.key" equal to "api_version"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "variables.0.value" equal to "v2"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "variables.0.enabled" equal to "true"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "variables.1.key" equal to "unused"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "variables.1.enabled" equal to "false"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "scripts.preRequest" equal to
      """
      pm.request.headers.upsert({ key: "X-Start", value: Date.now() });
      """
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "scripts.postResponse" equal to
      """
      pm.test("collection ok", function() {
        pm.expect(pm.response.code).to.equal(200);
      });
      """
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "bruno.version" equal to "1"
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "bruno.ignore" with the items "node_modules, .git"

  @folders
  Scenario: Folders keep their auth, headers, scripts and sequence
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/bruno-shop/orders/folder.yaml" should contain the field "name" equal to "Orders"
    And the workspace file "acme/collections/bruno-shop/orders/folder.yaml" should contain the field "kind" equal to "folder"
    And the workspace file "acme/collections/bruno-shop/orders/folder.yaml" should contain the field "auth.type" equal to "basic"
    And the workspace file "acme/collections/bruno-shop/orders/folder.yaml" should contain the field "auth.basic.username" equal to "shop"
    And the workspace file "acme/collections/bruno-shop/orders/folder.yaml" should contain the field "auth.basic.password" equal to "{{shop_pw}}"
    And the workspace file "acme/collections/bruno-shop/orders/folder.yaml" should contain the field "bruno.seq" equal to "2"
    And the workspace file "acme/collections/bruno-shop/products/folder.yaml" should contain the field "auth.type" equal to "inherit"
    And the workspace file "acme/collections/bruno-shop/products/folder.yaml" should contain the field "headers.0.key" equal to "X-Area"
    And the workspace file "acme/collections/bruno-shop/products/folder.yaml" should contain the field "headers.0.value" equal to "products"
    And the workspace file "acme/collections/bruno-shop/products/folder.yaml" should contain the field "scripts.postResponse" equal to
      """
      pm.variables.set("lastStatus", pm.response.code);
      """

  @requests
  Scenario: A request keeps method, URL, query params, headers, inherited auth, docs, settings and its Bruno-only fields
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "name" equal to "List products"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "id" matching "^[0-9A-HJKMNP-TV-Z]{26}$"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "method" equal to "GET"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "url" equal to "{{base_url}}/products?page=1&debug=true"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "params.0.key" equal to "page"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "params.0.value" equal to "1"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "params.0.enabled" equal to "true"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "params.1.key" equal to "debug"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "params.1.enabled" equal to "false"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "headers.0.key" equal to "Accept"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "headers.0.value" equal to "application/json"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "headers.1.key" equal to "X-Trace"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "headers.1.enabled" equal to "false"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "auth.type" equal to "inherit"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "body.mode" equal to "none"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "description" equal to "Lists all **products**."
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "settings.encodeUrl" equal to "true"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "settings.timeoutMs" equal to "5000"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "settings.followRedirects" equal to "false"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "bruno.settings.maxRedirects" equal to "3"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "bruno.tags" with the items "smoke"
    And the workspace file "acme/collections/bruno-shop/products/list-products.yaml" should contain the field "bruno.seq" equal to "2"
    And the workspace file "acme/collections/bruno-shop/health.yaml" should contain the field "method" equal to "PURGE"
    And the workspace file "acme/collections/bruno-shop/health.yaml" should contain the field "url" equal to "{{base_url}}/health"
    And the workspace file "acme/collections/bruno-shop/health.yaml" should contain the field "settings.encodeUrl" equal to "false"
    And the workspace file "acme/collections/bruno-shop/products/create-product.yaml" should contain the field "method" equal to "POST"
    And the workspace file "acme/collections/bruno-shop/products/create-product.yaml" should contain the field "url" equal to "{{base_url}}/products/:category"
    And the workspace file "acme/collections/bruno-shop/products/create-product.yaml" should contain the field "pathVariables.0.key" equal to "category"
    And the workspace file "acme/collections/bruno-shop/products/create-product.yaml" should contain the field "pathVariables.0.value" equal to "chairs"

  @bodies
  Scenario: Every body mode maps to the matching Xenon body and file references must be picked again
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/bruno-shop/products/create-product.yaml" should contain the field "body.mode" equal to "raw"
    And the workspace file "acme/collections/bruno-shop/products/create-product.yaml" should contain the field "body.raw.language" equal to "json"
    And the workspace file "acme/collections/bruno-shop/products/create-product.yaml" should contain the field "body.raw.content" equal to
      """
      {
        "name": "Chair",
        "price": 9.5
      }
      """
    And the workspace file "acme/collections/bruno-shop/health.yaml" should contain the field "body.mode" equal to "raw"
    And the workspace file "acme/collections/bruno-shop/health.yaml" should contain the field "body.raw.language" equal to "text"
    And the workspace file "acme/collections/bruno-shop/health.yaml" should contain the field "body.raw.content" equal to "ping"
    And the workspace file "acme/collections/bruno-shop/orders/xml.yaml" should contain the field "body.raw.language" equal to "xml"
    And the workspace file "acme/collections/bruno-shop/orders/xml.yaml" should contain the field "body.raw.content" equal to
      """
      <order id="1"/>
      """
    And the workspace file "acme/collections/bruno-shop/orders/form.yaml" should contain the field "body.mode" equal to "urlencoded"
    And the workspace file "acme/collections/bruno-shop/orders/form.yaml" should contain the field "body.urlencoded.0.key" equal to "a"
    And the workspace file "acme/collections/bruno-shop/orders/form.yaml" should contain the field "body.urlencoded.0.value" equal to "1"
    And the workspace file "acme/collections/bruno-shop/orders/form.yaml" should contain the field "body.urlencoded.1.key" equal to "b"
    And the workspace file "acme/collections/bruno-shop/orders/form.yaml" should contain the field "body.urlencoded.1.enabled" equal to "false"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "body.mode" equal to "formdata"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "body.formdata.0.key" equal to "note"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "body.formdata.0.type" equal to "text"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "body.formdata.0.value" equal to "hello"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "body.formdata.1.key" equal to "file"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "body.formdata.1.type" equal to "file"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "body.formdata.1.src" equal to "/Users/someone/invoice.pdf"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "body.formdata.1.srcState" equal to "repick"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "body.formdata.2.key" equal to "meta"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "body.formdata.2.enabled" equal to "false"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "body.formdata.2.src" equal to "/tmp/meta.json"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "body.formdata.2.srcState" equal to "repick"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "bruno.formdata.0.key" equal to "meta"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "bruno.formdata.0.contentType" equal to "application/json"
    And the workspace file "acme/collections/bruno-shop/orders/raw-file.yaml" should contain the field "body.mode" equal to "binary"
    And the workspace file "acme/collections/bruno-shop/orders/raw-file.yaml" should contain the field "body.binary.src" equal to "/tmp/payload.bin"
    And the workspace file "acme/collections/bruno-shop/orders/raw-file.yaml" should contain the field "body.binary.srcState" equal to "repick"
    And the workspace file "acme/collections/bruno-shop/orders/raw-file.yaml" should contain the field "bruno.file.contentType" equal to "application/octet-stream"
    And the workspace file "acme/collections/bruno-shop/orders/query.yaml" should contain the field "body.mode" equal to "graphql"
    And the workspace file "acme/collections/bruno-shop/orders/query.yaml" should contain the field "body.graphql.query" equal to
      """
      query {
        orders {
          id
        }
      }
      """
    And the workspace file "acme/collections/bruno-shop/orders/query.yaml" should contain the field "body.graphql.variables" equal to
      """
      {
        "limit": 10
      }
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    Then the response field "warnings" should contain an item with "code" equal to "missing_file" and "message" equal to "/Users/someone/invoice.pdf must be picked again on this machine before it is sent"
    And the response field "warnings" should contain an item with "code" equal to "missing_file" and "path" equal to "Bruno Shop/Orders/Raw file"

  @auth
  Scenario: Supported auth types map to Xenon auth and the others are kept as unsupported
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response field "warnings" should contain an item with "code" equal to "unsupported_auth" and "message" equal to "Auth type ntlm is not supported yet; it is kept for export and the request is sent without auth"
    And the response field "warnings" should contain an item with "code" equal to "unsupported_auth" and "path" equal to "Bruno Shop/Orders/Upload invoice"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "auth.type" equal to "unsupported"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "auth.unsupported.source" equal to "bruno"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "auth.unsupported.type" equal to "ntlm"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "auth.unsupported.data.username" equal to "ana"
    And the workspace file "acme/collections/bruno-shop/orders/upload-invoice.yaml" should contain the field "auth.unsupported.data.domain" equal to "ACME"
    And the workspace file "acme/collections/bruno-shop/products/create-product.yaml" should contain the field "auth.type" equal to "bearer"
    And the workspace file "acme/collections/bruno-shop/products/create-product.yaml" should contain the field "auth.bearer.token" equal to "{{token}}"
    And the workspace file "acme/collections/bruno-shop/orders/xml.yaml" should contain the field "auth.type" equal to "none"
    And the workspace file "acme/collections/bruno-shop/orders/query.yaml" should contain the field "auth.type" equal to "digest"
    And the workspace file "acme/collections/bruno-shop/orders/query.yaml" should contain the field "auth.digest.username" equal to "ana"
    And the workspace file "acme/collections/bruno-shop/orders/query.yaml" should contain the field "auth.digest.password" equal to "{{pw}}"
    And the workspace file "acme/collections/bruno-shop/orders/form.yaml" should contain the field "auth.type" equal to "apikey"
    And the workspace file "acme/collections/bruno-shop/orders/form.yaml" should contain the field "auth.apikey.key" equal to "X-Api-Key"
    And the workspace file "acme/collections/bruno-shop/orders/form.yaml" should contain the field "auth.apikey.value" equal to "{{api_key}}"
    And the workspace file "acme/collections/bruno-shop/orders/form.yaml" should contain the field "auth.apikey.in" equal to "header"
    And the workspace file "acme/collections/bruno-shop/health.yaml" should contain the field "auth.type" equal to "oauth2"
    And the workspace file "acme/collections/bruno-shop/health.yaml" should contain the field "auth.oauth2.grant_type" equal to "client_credentials"
    And the workspace file "acme/collections/bruno-shop/health.yaml" should contain the field "auth.oauth2.accessTokenUrl" equal to "https://auth.acme.test/token"
    And the workspace file "acme/collections/bruno-shop/health.yaml" should contain the field "auth.oauth2.clientId" equal to "shop"
    And the workspace file "acme/collections/bruno-shop/health.yaml" should contain the field "auth.oauth2.clientSecret" equal to "{{client_secret}}"
    And the workspace file "acme/collections/bruno-shop/health.yaml" should contain the field "auth.oauth2.scope" equal to "read"
    And the workspace file "acme/collections/bruno-shop/orders/raw-file.yaml" should contain the field "auth.type" equal to "awsv4"
    And the workspace file "acme/collections/bruno-shop/orders/raw-file.yaml" should contain the field "auth.awsv4.accessKey" equal to "AKIA"
    And the workspace file "acme/collections/bruno-shop/orders/raw-file.yaml" should contain the field "auth.awsv4.secretKey" equal to "{{aws_secret}}"
    And the workspace file "acme/collections/bruno-shop/orders/raw-file.yaml" should contain the field "auth.awsv4.service" equal to "s3"
    And the workspace file "acme/collections/bruno-shop/orders/raw-file.yaml" should contain the field "auth.awsv4.region" equal to "us-east-1"

  @environments
  Scenario: Environments come in with plain rows and missing secrets
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "1", "include": true, "target_local_id": "{{Acme.local_id}}" }, { "key": "2", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "workspaces.0.created.environments" equal to "2"
    And the response should contain the field "failed" array length equal to 0
    And I save the response field "items.0.id" as "local"
    And the workspace file "acme/environments/local.yaml" should contain the field "name" equal to "local"
    And the workspace file "acme/environments/local.yaml" should contain the field "color" equal to "green"
    And the workspace file "acme/environments/local.yaml" should contain the field "variables" array length equal to 4
    And the workspace file "acme/environments/local.yaml" should contain the field "variables.0.key" equal to "base_url"
    And the workspace file "acme/environments/local.yaml" should contain the field "variables.0.value" equal to "http://localhost:3000"
    And the workspace file "acme/environments/local.yaml" should contain the field "variables.0.enabled" equal to "true"
    And the workspace file "acme/environments/local.yaml" should contain the field "variables.1.key" equal to "debug"
    And the workspace file "acme/environments/local.yaml" should contain the field "variables.1.enabled" equal to "false"
    And the workspace file "acme/environments/local.yaml" should contain the field "variables.2.key" equal to "token"
    And the workspace file "acme/environments/local.yaml" should contain the field "variables.2.secret" equal to "true"
    And the workspace file "acme/environments/local.yaml" should not contain the field "variables.2.value"
    And the workspace file "acme/environments/local.yaml" should contain the field "variables.3.key" equal to "old_token"
    And the workspace file "acme/environments/local.yaml" should contain the field "variables.3.secret" equal to "true"
    And the workspace file "acme/environments/local.yaml" should contain the field "variables.3.enabled" equal to "false"
    And the workspace file "acme/environments/prod.yaml" should contain the field "color" equal to "red"
    And the workspace file "acme/environments/prod.yaml" should contain the field "confirmBeforeSend" equal to "true"
    And the workspace file "acme/environments/prod.yaml" should contain the field "variables.1.key" equal to "token"
    And the workspace file "acme/environments/prod.yaml" should contain the field "variables.1.secret" equal to "true"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{local}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.0.current_value" equal to "http://localhost:3000"
    And the response should contain the field "rows.2.secret_state" equal to "missing"
    And the response should contain the field "rows.3.secret_state" equal to "missing"

  @refused
  Scenario Outline: A collection mixing .bru and YAML request files is refused - <case>
    Given the file "<folder>/<marker>" exists with content
      """
      <marker_content>
      """
    And the file "<folder>/<stray>" exists with content
      """
      <stray_content>
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/<folder>"] }
      """
    Then the call should fail with code "IMP-00415"
    And the error should contain the field "error.description" equal to "<folder> mixes Bruno .bru and YAML files"
    And the workspace folder "acme" should contain the entries ".gitignore, collections, environments, workspace.yaml"

    Examples:
      | case                 | folder     | marker             | marker_content                              | stray     | stray_content                       |
      | .bru with a YAML one | bru-mixed  | bruno.json         | { "version": "1", "name": "Mixed" }         | B.yml     | info: { name: B, type: http }       |
      | YAML with a .bru one | yaml-mixed | opencollection.yml | opencollection: "1.0.0"                     | A.bru     | meta { name: A }                    |
