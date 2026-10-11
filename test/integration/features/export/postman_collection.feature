#language: en
#utf-8

@all @xen_006 @export @postman_export @postman_collection
Feature: Export a Xenon collection as a Postman collection v2.1
  A collection created in Xenon exports to one Postman collection v2.1 JSON file that the official
  schema accepts, with the same folders, requests and order, methods, urls (raw plus host, path,
  query and path variables), headers, bodies, auth, scripts, variables, descriptions and saved
  examples. Collection and folder headers have no Postman place and are copied onto every request
  below them. Secret variables keep their name with type secret and an empty value. The export
  dialog asks Go for the formats and shapes of an item, opens the native save dialog and can show
  the written file in the OS file manager.

  Background:
    Given the workspace "Acme" exists
    And the resource folder "export/shop" is copied to the workspace folder "acme"
    And the "shop" variable is set to "01JX0000000000000000000001"

  @summary @schema
  Scenario: The collection export is a valid Postman v2.1 file with a summary
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{shop}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/shop.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "shop.postman_collection.json" should be a valid Postman v2.1 collection
    And the response should contain the field "path" equal to "{{temp_root}}/shop.postman_collection.json"
    And the response should contain the field "secrets" equal to "names_only"
    And the response should contain the field "counts.collections" equal to "1"
    And the response should contain the field "counts.folders" equal to "2"
    And the response should contain the field "counts.requests" equal to "9"
    And the response should contain the field "counts.examples" equal to "2"
    And the response should contain the field "counts.variables" equal to "4"
    And the 1 "export:progress" event should contain the field "done" equal to "0"
    And the last "export:progress" event should contain the field "done" equal to "1"
    And the last "export:progress" event should contain the field "total" equal to "1"

  @collection
  Scenario: The collection keeps its info, auth, scripts and variables
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{shop}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/shop.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "shop.postman_collection.json" should contain the field "info.name" equal to "Shop API"
    And the file "shop.postman_collection.json" should contain the field "info._postman_id" equal to "{{shop}}"
    And the file "shop.postman_collection.json" should contain the field "info.description" equal to "Every request feature Xenon has, created in Xenon."
    And the file "shop.postman_collection.json" should contain the field "info.schema" equal to "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
    And the file "shop.postman_collection.json" should contain the field "auth.type" equal to "bearer"
    And the file "shop.postman_collection.json" should contain the field "auth.bearer.0.key" equal to "token"
    And the file "shop.postman_collection.json" should contain the field "auth.bearer.0.value" equal to "{{token}}"
    And the file "shop.postman_collection.json" should contain the field "event" array length equal to 2
    And the file "shop.postman_collection.json" should contain the field "event.0.listen" equal to "prerequest"
    And the file "shop.postman_collection.json" should contain the field "event.0.script.type" equal to "text/javascript"
    And the file "shop.postman_collection.json" should contain the field "event.0.script.exec.0" equal to
      """
      pm.request.headers.add({ key: "X-From-Collection", value: "yes" });
      """
    And the file "shop.postman_collection.json" should contain the field "event.1.listen" equal to "test"
    And the file "shop.postman_collection.json" should contain the field "event.1.script.exec" array length equal to 3
    And the file "shop.postman_collection.json" should contain the field "event.1.script.exec.1" equal to "    pm.response.to.have.status(200);"
    And the file "shop.postman_collection.json" should contain the field "variable" array length equal to 4
    And the file "shop.postman_collection.json" should contain the field "variable.0.key" equal to "token"
    And the file "shop.postman_collection.json" should contain the field "variable.0.value" equal to "t-123"
    And the file "shop.postman_collection.json" should contain the field "variable.1.description" equal to "Default page"
    And the file "shop.postman_collection.json" should contain the field "variable.2.disabled" equal to "true"
    And the file "shop.postman_collection.json" should contain the field "variable.3.key" equal to "api_secret"
    And the file "shop.postman_collection.json" should contain the field "variable.3.type" equal to "secret"
    And the file "shop.postman_collection.json" should contain the field "variable.3.value" equal to ""
    And the file "shop.postman_collection.json" should not contain the field "headers"

  @tree
  Scenario: Folders and requests keep their order and descriptions
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{shop}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/shop.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "shop.postman_collection.json" should contain the field "item" array length equal to 3
    And the file "shop.postman_collection.json" should contain the field "item.0.name" equal to "Products"
    And the file "shop.postman_collection.json" should contain the field "item.0.id" equal to "01JX0000000000000000000002"
    And the file "shop.postman_collection.json" should contain the field "item.0.description" equal to "Catalogue endpoints"
    And the file "shop.postman_collection.json" should contain the field "item.0.event.0.listen" equal to "prerequest"
    And the file "shop.postman_collection.json" should not contain the field "item.0.auth"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.name" equal to "List products"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.description" equal to "Lists the catalogue"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.1.name" equal to "Get product"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.2.name" equal to "Create product"
    And the file "shop.postman_collection.json" should contain the field "item.1.name" equal to "Orders"
    And the file "shop.postman_collection.json" should contain the field "item.1.auth.type" equal to "noauth"
    And the file "shop.postman_collection.json" should contain the field "item.1.item" array length equal to 5
    And the file "shop.postman_collection.json" should contain the field "item.2.name" equal to "Ping"

  @url @headers
  Scenario: Urls, query parameters, path variables and headers
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{shop}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/shop.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.method" equal to "GET"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.url.raw" equal to "{{base_url}}/products?page=1&size={{page_size}}"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.url.host" with the items "{{base_url}}"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.url.path" with the items "products"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.url.query" array length equal to 3
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.url.query.1.key" equal to "size"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.url.query.1.value" equal to "{{page_size}}"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.url.query.1.description" equal to "page size"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.url.query.2.disabled" equal to "true"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.header" array length equal to 3
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.header.0.key" equal to "X-Tenant"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.header.0.value" equal to "shop"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.header.1.key" equal to "Accept"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.header.2.disabled" equal to "true"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.request.header.2.description" equal to "tracing"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.1.request.url.raw" equal to "{{base_url}}/products/:id"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.1.request.url.path" with the items "products, :id"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.1.request.url.variable.0.key" equal to "id"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.1.request.url.variable.0.value" equal to "42"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.1.request.url.variable.0.description" equal to "product id"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.0.request.header.0.key" equal to "X-Tenant"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.0.request.header.1.key" equal to "X-Area"
    And the file "shop.postman_collection.json" should contain the field "item.2.request.method" equal to "HEAD"

  @bodies
  Scenario: Every body mode
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{shop}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/shop.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "shop.postman_collection.json" should not contain the field "item.0.item.0.request.body"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.2.request.body.mode" equal to "raw"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.2.request.body.raw" equal to
      """
      {
        "name": "Pen",
        "price": 1.5
      }
      """
    And the file "shop.postman_collection.json" should contain the field "item.0.item.2.request.body.options.raw.language" equal to "json"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.0.request.body.mode" equal to "formdata"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.0.request.body.formdata.0.key" equal to "kind"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.0.request.body.formdata.0.type" equal to "text"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.0.request.body.formdata.0.value" equal to "invoice"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.0.request.body.formdata.1.disabled" equal to "true"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.0.request.body.formdata.2.type" equal to "file"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.0.request.body.formdata.2.src" equal to "files/invoice.txt"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.0.request.body.formdata.2.description" equal to "the invoice"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.1.request.body.mode" equal to "urlencoded"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.1.request.body.urlencoded.1.value" equal to "s3 cret&x"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.1.request.body.urlencoded.2.disabled" equal to "true"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.2.request.body.mode" equal to "graphql"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.2.request.body.graphql.query" equal to "query { order(id: 7) { id total } }"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.2.request.body.graphql.variables" equal to
      """
      {"verbose":true}
      """
    And the file "shop.postman_collection.json" should contain the field "item.1.item.3.request.body.mode" equal to "file"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.3.request.body.file.src" equal to "files/payload.bin"

  @auth
  Scenario: Every supported auth type
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{shop}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/shop.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "shop.postman_collection.json" should not contain the field "item.0.item.0.request.auth"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.1.request.auth.type" equal to "apikey"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.1.request.auth.apikey" array length equal to 2
    And the file "shop.postman_collection.json" should contain the field "item.0.item.1.request.auth.apikey.0.key" equal to "key"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.1.request.auth.apikey.0.value" equal to "X-Api-Key"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.1.request.auth.apikey.1.key" equal to "value"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.1.request.auth.apikey.1.value" equal to "{{api_key}}"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.2.request.auth.type" equal to "basic"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.2.request.auth.basic.0.key" equal to "username"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.2.request.auth.basic.0.value" equal to "admin"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.2.request.auth.basic.1.key" equal to "password"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.4.request.auth.type" equal to "awsv4"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.4.request.auth.awsv4.0.key" equal to "accessKey"
    And the file "shop.postman_collection.json" should contain the field "item.1.item.4.request.auth.awsv4.0.value" equal to "AKIDEXAMPLE"
    And the file "shop.postman_collection.json" should contain the field "item.2.request.auth.type" equal to "noauth"

  @scripts @settings
  Scenario: Request scripts, disabled scripts and settings
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{shop}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/shop.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.event.0.listen" equal to "test"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.event.0.script.exec.0" equal to
      """
      pm.test("listed", function () {
      """
    And the file "shop.postman_collection.json" should contain the field "item.2.event.0.listen" equal to "prerequest"
    And the file "shop.postman_collection.json" should contain the field "item.2.event.0.disabled" equal to "true"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.2.protocolProfileBehavior.followRedirects" equal to "false"
    And the file "shop.postman_collection.json" should not contain the field "item.0.item.0.protocolProfileBehavior"

  @examples
  Scenario: Saved examples become Postman responses
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{shop}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/shop.postman_collection.json"
      }
      """
    Then the call should succeed
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.response" array length equal to 2
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.response.0.name" equal to "Listed"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.response.0.code" equal to "200"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.response.0.status" equal to "OK"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.response.0.originalRequest.method" equal to "GET"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.response.0.originalRequest.url.raw" equal to "{{base_url}}/products?page=1"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.response.0.header.0.key" equal to "Content-Type"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.response.0.body" equal to
      """
      [{"id":1,"name":"Pen"}]
      """
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.response.1.name" equal to "Logo"
    And the file "shop.postman_collection.json" should contain the field "item.0.item.0.response.1.body" equal to ""

  @dialog
  Scenario: The export dialog gets the formats and shapes of a collection
    When I call "ExportService.Options" with
      """
      { "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{shop}}" }
      """
    Then the call should succeed
    And the response should contain the field "suggested_file_name" equal to "shop-api"
    And the response should contain the field "has_secrets" equal to "true"
    And the response should contain the field "has_current_values" equal to "false"
    And the response should contain the field "formats" array length equal to 2
    And the response should contain the field "formats.0.format" equal to "xenon_yaml"
    And the response should contain the field "formats.0.shapes" with the items "zip, bundle"
    And the response should contain the field "formats.0.include_secrets_allowed" equal to "true"
    And the response should contain the field "formats.1.format" equal to "postman"
    And the response should contain the field "formats.1.shapes" with the items "file"
    And the response should contain the field "formats.1.include_secrets_allowed" equal to "false"
    And the response should contain the field "formats.1.suggested_file_name" equal to "shop-api.postman_collection.json"

  @dialog
  Scenario: The save dialog suggests the file name and returns the chosen path
    Given the save dialog will return "{{temp_root}}/chosen.postman_collection.json"
    When I call "ExportService.PickTarget" with
      """
      { "suggested_file_name": "shop-api.postman_collection.json", "format": "postman", "shape": "file" }
      """
    Then the call should succeed
    And the response should contain the field "path" equal to "{{temp_root}}/chosen.postman_collection.json"
    And the save dialog should have suggested the name "shop-api.postman_collection.json"
    And the save dialog should have been opened with the title "Export as Postman"

  @dialog
  Scenario: The summary's Show in folder opens the file manager on the exported file
    When I call "ExportService.ShowInFolder" with
      """
      { "path": "{{temp_root}}/shop.postman_collection.json" }
      """
    Then the call should succeed
    And the file manager should have shown "{{temp_root}}/shop.postman_collection.json"

  @dialog @errors
  Scenario: The file manager failing is reported
    Given the file manager fails with "no file manager"
    When I call "ExportService.ShowInFolder" with
      """
      { "path": "{{temp_root}}/shop.postman_collection.json" }
      """
    Then the call should fail with code "EXP-09500"

  @errors
  Scenario Outline: Invalid export requests are refused - <case>
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "<kind>", "item_id": "{{shop}}", "format": "<format>", "shape": "<shape>",
        "include_secrets": <secrets>, "passphrase": "<passphrase>", "passphrase_confirmation": "<confirmation>",
        "target_path": "{{temp_root}}/refused.json"
      }
      """
    Then the call should fail with code "<code>"
    And the file "refused.json" should not exist

    Examples:
      | case                             | kind       | format     | shape  | secrets | passphrase | confirmation | code      |
      | a Postman zip of a collection    | collection | postman    | zip    | false   |            |              | EXP-00400 |
      | secrets in a Postman export      | collection | postman    | file   | true    | 12345678   | 12345678     | EXP-00400 |
      | a short passphrase               | collection | xenon_yaml | bundle | true    | 1234567    | 1234567      | EXP-01400 |
      | passphrases that do not match    | collection | xenon_yaml | bundle | true    | 12345678   | 87654321     | EXP-01400 |
      | an unknown shape                 | collection | postman    | tar    | false   |            |              | EXP-12400 |
      | a Postman zip of a request       | request    | postman    | zip    | false   |            |              | EXP-00400 |
