#language: en
#utf-8

@all @xen_003 @import @postman_collection
Feature: Import a Postman collection (v2.1)
  A Postman collection export becomes a Xenon collection in the chosen workspace with the same tree,
  order, methods, URLs, params, path variables, headers, bodies, auth, scripts, variables,
  descriptions and saved examples. Every Postman field Xenon has no place for is kept in the
  item's postman: block so a later Postman export loses nothing. Every item gets a new Xenon id.
  Imported file references come from another machine and are never read or sent until the file
  is picked again. A summary lists counts and every warning with the item it concerns.

  Background:
    Given the workspace "Acme" exists
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/full-v2.1.json"] }
      """
    And I save the response field "source_id" as "source"
    And I save the response field "items.0.key" as "item"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "{{item}}", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """

  @summary
  Scenario: The summary counts what came in
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the response should contain the field "items.0.local_id" equal to "{{Acme.local_id}}"
    And the response should contain the field "items.0.name" equal to "Acme API"
    And the response should contain the field "items.0.id" matching "^[0-9A-HJKMNP-TV-Z]{26}$"
    And the response should contain the field "workspaces.0.local_id" equal to "{{Acme.local_id}}"
    And the response should contain the field "workspaces.0.created.collections" equal to "1"
    And the response should contain the field "workspaces.0.created.folders" equal to "3"
    And the response should contain the field "workspaces.0.created.requests" equal to "14"
    And the response should contain the field "workspaces.0.created.examples" equal to "3"
    And the response should contain the field "workspaces.0.created.variables" equal to "3"
    And the response should contain the field "workspaces.0.created.scripts" equal to "7"
    And the response should contain the field "failed" array length equal to 0
    And the 1 "import:progress" event should contain the field "done" equal to "0"
    And the last "import:progress" event should contain the field "done" equal to "1"
    And the last "import:progress" event should contain the field "total" equal to "1"
    And the last "import:progress" event should contain the field "source_id" equal to "{{source}}"

  @summary @warnings
  Scenario: The summary lists every warning with the item it concerns
    Then the call should succeed
    And the response field "warnings" should contain 6 items with "code" equal to "unsupported_auth"
    And the response field "warnings" should contain 3 items with "code" equal to "missing_file"
    And the response field "warnings" should contain 1 items with "code" equal to "folder_variables"
    And the response field "warnings" should contain 1 items with "code" equal to "request_variables"
    And the response field "warnings" should contain 1 items with "code" equal to "raw_header_string"
    And the response field "warnings" should contain 1 items with "code" equal to "disabled_body"
    And the response field "warnings" should contain 1 items with "code" equal to "pm_require"
    And the response field "warnings" should contain 3 items with "code" equal to "pm_api_outside_mvp"
    And the response should contain the field "warnings" array length equal to 17
    And the response field "warnings" should contain an item with "code" equal to "unsupported_auth" and "path" equal to "Acme API/Users/Propfind"
    And the response field "warnings" should contain an item with "code" equal to "missing_file" and "path" equal to "Acme API/Bodies/Binary"
    And the response field "warnings" should contain an item with "code" equal to "folder_variables" and "path" equal to "Acme API/Users"
    And the response field "warnings" should contain an item with "code" equal to "request_variables" and "path" equal to "Acme API/Ping"
    And the response field "warnings" should contain an item with "code" equal to "unsupported_auth" and "message" equal to "Auth type hawk is not supported yet; it is kept for export and the request is sent without auth"
    And the response field "warnings" should contain an item with "code" equal to "missing_file" and "message" equal to "/Users/someone/.ssh/id_rsa must be picked again on this machine before it is sent"

  @tree
  Scenario: The sidebar shows the same tree in the same order
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 1
    And the response should contain the field "0.name" equal to "Acme API"
    And the response should contain the field "0.kind" equal to "collection"
    And the response should contain the field "0.children" array length equal to 4
    And the response should contain the field "0.children.0.name" equal to "Users"
    And the response should contain the field "0.children.0.kind" equal to "folder"
    And the response should contain the field "0.children.0.children.0.name" equal to "List users"
    And the response should contain the field "0.children.0.children.1.name" equal to "Admin"
    And the response should contain the field "0.children.0.children.1.children.0.name" equal to "Get user"
    And the response should contain the field "0.children.0.children.1.children.1.name" equal to "Delete user"
    And the response should contain the field "0.children.0.children.2.name" equal to "Propfind"
    And the response should contain the field "0.children.0.children.2.method" equal to "PROPFIND"
    And the response should contain the field "0.children.1.name" equal to "Bodies"
    And the response should contain the field "0.children.1.children" array length equal to 8
    And the response should contain the field "0.children.1.children.0.name" equal to "Raw JSON"
    And the response should contain the field "0.children.1.children.7.name" equal to "Empty body"
    And the response should contain the field "0.children.1.children.7.method" equal to "GET"
    And the response should contain the field "0.children.2.name" equal to "Ping"
    And the response should contain the field "0.children.3.name" equal to "Rebuilt URL"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "order" with the items "users, bodies, ping.yaml, rebuilt-url.yaml"
    And the workspace file "acme/collections/acme-api/users/folder.yaml" should contain the field "order" with the items "list-users.yaml, admin, propfind.yaml"
    And the workspace file "acme/workspace.yaml" should contain the field "order" with the items "acme-api"

  @ids
  Scenario: Every imported item gets a new Xenon id and keeps its Postman id for export
    Then the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "id" matching "^[0-9A-HJKMNP-TV-Z]{26}$"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "postman.info._postman_id" equal to "c0ffee00-0000-4000-8000-000000000001"
    And the workspace file "acme/collections/acme-api/users/folder.yaml" should contain the field "id" matching "^[0-9A-HJKMNP-TV-Z]{26}$"
    And the workspace file "acme/collections/acme-api/users/folder.yaml" should contain the field "postman.id" equal to "fld-users"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "id" matching "^[0-9A-HJKMNP-TV-Z]{26}$"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "postman.id" equal to "req-list-users"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "examples.0.id" matching "^[0-9A-HJKMNP-TV-Z]{26}$"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "examples.0.postman.id" equal to "ex-list-ok"

  @collection
  Scenario: The collection keeps its description, auth, scripts, variables and Postman-only fields
    Then the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "name" equal to "Acme API"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "description" equal to "The **Acme** API"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "auth.type" equal to "bearer"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "auth.bearer.token" equal to "{{token}}"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "scripts.preRequest" equal to
      """
      pm.environment.set("ts", Date.now());
      """
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "scripts.postResponse" equal to
      """
      pm.test("status ok", () => pm.response.to.have.status(200));
      """
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "variables.0.key" equal to "base_url"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "variables.0.value" equal to "https://api.acme.test"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "variables.0.enabled" equal to "true"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "variables.0.postman.type" equal to "string"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "variables.1.value" equal to "50"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "variables.1.description" equal to "Default page"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "variables.1.postman.type" equal to "number"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "variables.2.key" equal to "legacy"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "variables.2.enabled" equal to "false"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "postman.info.schema" equal to "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "postman.info._exporter_id" equal to "1234567"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "postman.info.description.type" equal to "text/markdown"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "postman.protocolProfileBehavior.disableBodyPruning" equal to "true"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "postman.event.prerequest.id" equal to "col-pre"
    And the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "postman.auth.bearer.token" equal to "string"

  @folders
  Scenario: Folders keep auth (explicit none and basic), scripts including disabled ones, and their variables for export
    Then the workspace file "acme/collections/acme-api/users/folder.yaml" should contain the field "name" equal to "Users"
    And the workspace file "acme/collections/acme-api/users/folder.yaml" should contain the field "description" equal to "User endpoints"
    And the workspace file "acme/collections/acme-api/users/folder.yaml" should contain the field "auth.type" equal to "none"
    And the workspace file "acme/collections/acme-api/users/folder.yaml" should contain the field "scripts.preRequest" equal to
      """
      pm.variables.set("x", 1);
      """
    And the workspace file "acme/collections/acme-api/users/folder.yaml" should contain the field "scripts.disabled" with the items "preRequest"
    And the workspace file "acme/collections/acme-api/users/folder.yaml" should contain the field "postman.variable.0.key" equal to "page"
    And the workspace file "acme/collections/acme-api/users/admin/folder.yaml" should contain the field "auth.type" equal to "basic"
    And the workspace file "acme/collections/acme-api/users/admin/folder.yaml" should contain the field "auth.basic.username" equal to "admin"
    And the workspace file "acme/collections/acme-api/users/admin/folder.yaml" should contain the field "auth.basic.password" equal to "{{admin_pw}}"
    And the workspace file "acme/collections/acme-api/users/admin/folder.yaml" should contain the field "scripts.postResponse" equal to
      """
      pm.test("admin", () => pm.expect(pm.response.code).to.eql(200));
      """
    And the workspace file "acme/collections/acme-api/users/admin/folder.yaml" should not contain the field "scripts.disabled"

  @requests
  Scenario: A request keeps method, URL, query rows, headers, inherited auth and its test script
    Then the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "method" equal to "GET"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "url" equal to "{{base_url}}/users?page=1&size=10&debug=true"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "params.0.key" equal to "page"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "params.1.description" equal to "page size"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "params.2.key" equal to "debug"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "params.2.enabled" equal to "false"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "headers.0.key" equal to "Accept"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "headers.1.key" equal to "X-Trace"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "headers.1.enabled" equal to "false"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "headers.1.description" equal to "tracing"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "auth.type" equal to "inherit"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "scripts.postResponse" equal to
      """
      pm.test("list", function () {
          var data = pm.response.json();
          pm.expect(data.length).to.be.above(0);
      });
      """
    And the workspace file "acme/collections/acme-api/users/admin/get-user.yaml" should contain the field "url" equal to "{{base_url}}/users/:id"
    And the workspace file "acme/collections/acme-api/users/admin/get-user.yaml" should contain the field "pathVariables.0.key" equal to "id"
    And the workspace file "acme/collections/acme-api/users/admin/get-user.yaml" should contain the field "pathVariables.0.value" equal to "42"
    And the workspace file "acme/collections/acme-api/users/admin/get-user.yaml" should contain the field "pathVariables.0.description" equal to "user id"
    And the workspace file "acme/collections/acme-api/users/propfind.yaml" should contain the field "method" equal to "PROPFIND"
    And the workspace file "acme/collections/acme-api/users/propfind.yaml" should contain the field "url" equal to "{{base_url}}/files?depth=1"
    And the workspace file "acme/collections/acme-api/users/propfind.yaml" should contain the field "params.0.key" equal to "depth"
    And the workspace file "acme/collections/acme-api/users/propfind.yaml" should contain the field "params.0.value" equal to "1"
    And the workspace file "acme/collections/acme-api/bodies/empty-body.yaml" should contain the field "method" equal to "GET"
    And the workspace file "acme/collections/acme-api/bodies/empty-body.yaml" should contain the field "auth.type" equal to "none"
    And the workspace file "acme/collections/acme-api/bodies/empty-body.yaml" should contain the field "body.mode" equal to "none"

  @requests @strings
  Scenario: String requests, object descriptions and URLs without raw are accepted
    Then the workspace file "acme/collections/acme-api/ping.yaml" should contain the field "method" equal to "GET"
    And the workspace file "acme/collections/acme-api/ping.yaml" should contain the field "url" equal to "https://api.acme.test/ping"
    And the workspace file "acme/collections/acme-api/ping.yaml" should contain the field "description" equal to "Health check"
    And the workspace file "acme/collections/acme-api/ping.yaml" should contain the field "postman.description.type" equal to "text/plain"
    And the workspace file "acme/collections/acme-api/ping.yaml" should contain the field "postman.variable.0.key" equal to "x"
    And the workspace file "acme/collections/acme-api/rebuilt-url.yaml" should contain the field "url" equal to "https://api.acme.test:8443/v1/rebuilt?q=1#frag"
    And the workspace file "acme/collections/acme-api/rebuilt-url.yaml" should contain the field "params.0.key" equal to "q"
    And the workspace file "acme/collections/acme-api/rebuilt-url.yaml" should contain the field "postman.request.url.port" equal to "8443"
    And the workspace file "acme/collections/acme-api/rebuilt-url.yaml" should contain the field "postman.request.url.hash" equal to "frag"
    And the workspace file "acme/collections/acme-api/rebuilt-url.yaml" should contain the field "postman.request.proxy.host" equal to "proxy.acme.test"
    And the workspace file "acme/collections/acme-api/rebuilt-url.yaml" should contain the field "postman.request.certificate.name" equal to "client"

  @bodies
  Scenario: Every body mode maps to the matching Xenon body
    Then the workspace file "acme/collections/acme-api/bodies/raw-json.yaml" should contain the field "body.mode" equal to "raw"
    And the workspace file "acme/collections/acme-api/bodies/raw-json.yaml" should contain the field "body.raw.language" equal to "json"
    And the workspace file "acme/collections/acme-api/bodies/raw-json.yaml" should contain the field "body.raw.content" equal to
      """
      {
        "amount": 1000
      }
      """
    And the workspace file "acme/collections/acme-api/bodies/raw-json.yaml" should contain the field "headers.0.key" equal to "Content-Type"
    And the workspace file "acme/collections/acme-api/bodies/raw-json.yaml" should contain the field "headers.0.value" equal to "application/json"
    And the workspace file "acme/collections/acme-api/bodies/raw-json.yaml" should contain the field "headers.1.key" equal to "X-Raw"
    And the workspace file "acme/collections/acme-api/bodies/raw-json.yaml" should contain the field "postman.request.header.raw" equal to "true"
    And the workspace file "acme/collections/acme-api/bodies/raw-xml.yaml" should contain the field "body.raw.language" equal to "xml"
    And the workspace file "acme/collections/acme-api/bodies/raw-xml.yaml" should contain the field "body.raw.content" equal to "<a/>"
    And the workspace file "acme/collections/acme-api/bodies/urlencoded.yaml" should contain the field "body.mode" equal to "urlencoded"
    And the workspace file "acme/collections/acme-api/bodies/urlencoded.yaml" should contain the field "body.urlencoded.1.key" equal to "b"
    And the workspace file "acme/collections/acme-api/bodies/urlencoded.yaml" should contain the field "body.urlencoded.1.enabled" equal to "false"
    And the workspace file "acme/collections/acme-api/bodies/urlencoded.yaml" should contain the field "body.urlencoded.1.description" equal to "off"
    And the workspace file "acme/collections/acme-api/bodies/form-data.yaml" should contain the field "body.mode" equal to "formdata"
    And the workspace file "acme/collections/acme-api/bodies/form-data.yaml" should contain the field "body.formdata.0.type" equal to "text"
    And the workspace file "acme/collections/acme-api/bodies/form-data.yaml" should contain the field "body.formdata.0.value" equal to "hi"
    And the workspace file "acme/collections/acme-api/bodies/form-data.yaml" should contain the field "body.formdata.0.description" equal to "a note"
    And the workspace file "acme/collections/acme-api/bodies/form-data.yaml" should contain the field "body.formdata.1.type" equal to "file"
    And the workspace file "acme/collections/acme-api/bodies/form-data.yaml" should contain the field "body.formdata.1.src" equal to "/Users/someone/.ssh/id_rsa"
    And the workspace file "acme/collections/acme-api/bodies/form-data.yaml" should contain the field "body.formdata.1.srcState" equal to "repick"
    And the workspace file "acme/collections/acme-api/bodies/form-data.yaml" should contain the field "body.formdata.2.src" equal to "/tmp/a.txt"
    And the workspace file "acme/collections/acme-api/bodies/form-data.yaml" should contain the field "body.formdata.2.srcState" equal to "repick"
    And the workspace file "acme/collections/acme-api/bodies/form-data.yaml" should contain the field "body.formdata.2.enabled" equal to "false"
    And the workspace file "acme/collections/acme-api/bodies/form-data.yaml" should contain the field "postman.request.body.formdata.2.src" with the items "/tmp/a.txt, /tmp/b.txt"
    And the workspace file "acme/collections/acme-api/bodies/form-data.yaml" should contain the field "postman.request.body.formdata.2.contentType" equal to "text/plain"
    And the workspace file "acme/collections/acme-api/bodies/binary.yaml" should contain the field "body.mode" equal to "binary"
    And the workspace file "acme/collections/acme-api/bodies/binary.yaml" should contain the field "body.binary.src" equal to "/tmp/payload.bin"
    And the workspace file "acme/collections/acme-api/bodies/binary.yaml" should contain the field "body.binary.srcState" equal to "repick"
    And the workspace file "acme/collections/acme-api/bodies/graphql.yaml" should contain the field "body.mode" equal to "graphql"
    And the workspace file "acme/collections/acme-api/bodies/graphql.yaml" should contain the field "body.graphql.query" equal to "query { me { id } }"
    And the workspace file "acme/collections/acme-api/bodies/graphql.yaml" should contain the field "body.graphql.variables" equal to
      """
      {"a":1}
      """
    And the workspace file "acme/collections/acme-api/bodies/disabled-body.yaml" should contain the field "body.mode" equal to "none"
    And the workspace file "acme/collections/acme-api/bodies/disabled-body.yaml" should contain the field "postman.request.body.mode" equal to "raw"
    And the workspace file "acme/collections/acme-api/bodies/disabled-body.yaml" should contain the field "postman.request.body.raw" equal to "ignored"
    And the workspace file "acme/collections/acme-api/bodies/disabled-body.yaml" should contain the field "postman.request.body.disabled" equal to "true"

  @auth
  Scenario Outline: Supported auth types map to Xenon's auth format - <file>
    Then the workspace file "acme/collections/acme-api/<file>" should contain the field "auth.type" equal to "<type>"
    And the workspace file "acme/collections/acme-api/<file>" should contain the field "<field>" equal to "<value>"

    Examples:
      | file                       | type   | field                     | value                        |
      | users/admin/get-user.yaml  | apikey | auth.apikey.key           | X-Api-Key                    |
      | users/admin/get-user.yaml  | apikey | auth.apikey.in            | header                       |
      | users/admin/delete-user.yaml | digest | auth.digest.realm       | acme                         |
      | users/admin/delete-user.yaml | digest | auth.digest.password    | {{pw}}                       |
      | bodies/form-data.yaml      | oauth2 | auth.oauth2.grant_type    | client_credentials           |
      | bodies/form-data.yaml      | oauth2 | auth.oauth2.useBrowser    | false                        |
      | bodies/binary.yaml         | awsv4  | auth.awsv4.region         | us-east-1                    |
      | bodies/binary.yaml         | awsv4  | auth.awsv4.secretKey      | {{aws_secret}}               |

  @auth @unsupported
  Scenario Outline: Unsupported auth types are kept unchanged for export - <original>
    Then the workspace file "acme/collections/acme-api/<file>" should contain the field "auth.type" equal to "unsupported"
    And the workspace file "acme/collections/acme-api/<file>" should contain the field "auth.unsupported.source" equal to "postman"
    And the workspace file "acme/collections/acme-api/<file>" should contain the field "auth.unsupported.type" equal to "<original>"
    And the workspace file "acme/collections/acme-api/<file>" should contain the field "auth.unsupported.data.0.key" equal to "<key>"

    Examples:
      | file                       | original | key         |
      | users/propfind.yaml        | hawk     | authId      |
      | bodies/raw-json.yaml       | ntlm     | username    |
      | bodies/raw-xml.yaml        | jwt      | algorithm   |
      | bodies/urlencoded.yaml     | asap     | kid         |
      | bodies/graphql.yaml        | edgegrid | accessToken |
      | bodies/disabled-body.yaml  | oauth1   | consumerKey |

  @settings
  Scenario: protocolProfileBehavior maps to request settings and is kept for export
    Then the workspace file "acme/collections/acme-api/users/admin/delete-user.yaml" should contain the field "settings.followRedirects" equal to "false"
    And the workspace file "acme/collections/acme-api/users/admin/delete-user.yaml" should contain the field "settings.verifySsl" equal to "false"
    And the workspace file "acme/collections/acme-api/users/admin/delete-user.yaml" should contain the field "settings.encodeUrl" equal to "false"
    And the workspace file "acme/collections/acme-api/users/admin/delete-user.yaml" should contain the field "postman.protocolProfileBehavior.strictSSL" equal to "false"
    And the workspace file "acme/collections/acme-api/users/admin/delete-user.yaml" should contain the field "postman.protocolProfileBehavior.disableCookies" equal to "true"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "settings.followRedirects" equal to "true"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "settings.encodeUrl" equal to "true"

  @scripts
  Scenario: Scripts become pre-request and post-response scripts at every level
    Then the workspace file "acme/collections/acme-api/bodies/raw-json.yaml" should contain the field "scripts.preRequest" equal to
      """
      // pm.visualizer.set is only mentioned in a comment
      const jar = pm.cookies.jar();
      console.log(pm.info.requestName, "pm.iterationData in a string");
      """
    And the workspace file "acme/collections/acme-api/bodies/raw-json.yaml" should contain the field "postman.event.prerequest.id" equal to "raw-pre"
    And the workspace file "acme/collections/acme-api/bodies/form-data.yaml" should contain the field "scripts.preRequest" equal to
      """
      const moment = pm.require('npm:moment@2.29.4');
      pm.environment.set("now", moment().format());
      """
    And the workspace file "acme/collections/acme-api/bodies/form-data.yaml" should contain the field "scripts.postResponse" equal to ""

  @examples
  Scenario: Saved examples are stored with their request
    Then the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "examples.0.name" equal to "200 OK"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "examples.0.request.method" equal to "GET"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "examples.0.request.url" equal to "{{base_url}}/users?page=1&size=10&debug=true"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "examples.0.request.params.2.enabled" equal to "false"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "examples.0.response.code" equal to "200"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "examples.0.response.status" equal to "OK"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "examples.0.response.contentType" equal to "application/json"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "examples.0.response.headers.0.key" equal to "Content-Type"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "examples.0.response.bodyEncoding" equal to "text"
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "examples.0.response.body" equal to
      """
      [
        { "id": 1, "name": "Ana" }
      ]
      """
    And the workspace file "acme/collections/acme-api/users/list-users.yaml" should contain the field "examples.0.postman._postman_previewlanguage" equal to "json"
    And the workspace file "acme/collections/acme-api/users/admin/get-user.yaml" should contain the field "examples.0.name" equal to "Found"
    And the workspace file "acme/collections/acme-api/users/admin/get-user.yaml" should contain the field "examples.0.postman.responseTime" equal to "120"
    And the workspace file "acme/collections/acme-api/users/admin/get-user.yaml" should contain the field "examples.1.name" equal to "Not found"
    And the workspace file "acme/collections/acme-api/users/admin/get-user.yaml" should contain the field "examples.1.response.code" equal to "404"
    And the workspace file "acme/collections/acme-api/users/admin/get-user.yaml" should contain the field "examples.1.response.status" equal to "Not Found"
    And the workspace file "acme/collections/acme-api/users/admin/get-user.yaml" should contain the field "examples.1.response.contentType" equal to "text/plain"
    And the workspace file "acme/collections/acme-api/users/admin/get-user.yaml" should contain the field "examples.1.response.body" equal to "missing"

  @open
  Scenario: An imported request opens with its unsupported auth and re-pick markers
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "0.children.0.children.2.id" as "propfind"
    And I save the response field "0.children.1.children.3.id" as "form"
    And I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{propfind}}" }
      """
    Then the call should succeed
    And the response should contain the field "auth.type" equal to "unsupported"
    And the response should contain the field "auth.unsupported.type" equal to "hawk"
    And the response should contain the field "auth.unsupported.source" equal to "postman"
    And the response should not contain "hawk_key"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{form}}" }
      """
    Then the response should contain the field "body.formdata.0.src_state" equal to "ok"
    And the response should contain the field "body.formdata.1.src_state" equal to "repick"

  @save
  Scenario: Saving an imported request from the editor keeps its scripts, examples, auth details and Postman fields
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "0.children.0.children.1.children.0.id" as "get_user"
    And I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{get_user}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "RequestService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "base_revision": "{{revision}}",
        "request": { "id": "{{get_user}}", "name": "Get user", "method": "GET", "url": "{{base_url}}/users/:id", "auth": { "type": "apikey" } }
      }
      """
    Then the call should succeed
    And the workspace file "acme/collections/acme-api/users/admin/get-user.yaml" should contain the field "auth.apikey.key" equal to "X-Api-Key"
    And the workspace file "acme/collections/acme-api/users/admin/get-user.yaml" should contain the field "examples" array length equal to 2
    And the workspace file "acme/collections/acme-api/users/admin/get-user.yaml" should contain the field "postman.id" equal to "req-get-user"

  @files @security
  Scenario: Imported file references are never read or sent until the file is picked again
    Given the file "secret.txt" exists with content
      """
      top secret
      """
    And the "api" variable is set to the mocked api url
    And the file "upload.postman_collection.json" exists with content
      """
      {
        "info": { "name": "Uploads", "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json" },
        "item": [
          { "name": "Upload", "request": { "method": "POST", "url": "{{api}}/upload",
            "body": { "mode": "formdata", "formdata": [ { "key": "f", "type": "file", "src": "{{temp_root}}/secret.txt" } ] } } }
        ]
      }
      """
    And the 1 "POST" request to "/upload" returns status 200 with content type "text/plain" and body
      """
      ok
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/upload.postman_collection.json"] }
      """
    And I save the response field "source_id" as "upload_source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{upload_source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/uploads/upload.yaml" should contain the field "body.formdata.0.srcState" equal to "repick"
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "POST", "url": "{{api}}/upload", "body": { "mode": "formdata", "formdata": [ { "key": "f", "type": "file", "src": "{{temp_root}}/secret.txt", "src_state": "repick" } ] } }
      }
      """
    Then the call should succeed
    And the response should contain the field "error.kind" equal to "missing_file"
    And the "POST" request to "/upload" should have been received 0 times
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "POST", "url": "{{api}}/upload", "body": { "mode": "formdata", "formdata": [ { "key": "f", "type": "file", "src": "{{temp_root}}/secret.txt", "src_state": "ok" } ] } }
      }
      """
    Then the response should contain the field "status_code" equal to "200"
    And the 1 "POST" request to "/upload" should have the form file "f" named "secret.txt" with content "top secret"

  @auth @oauth2_token @xen_005
  Scenario: A literal OAuth 2.0 access token moves to the vault; a variable reference stays in the file
    Given the keychain is available
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/oauth2-token.json"] }
      """
    And I save the response field "source_id" as "token_source"
    And I save the response field "items.0.key" as "token_item"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{token_source}}", "items": [ { "key": "{{token_item}}", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And I save the response field "items.0.id" as "token_collection"
    And no file under the workspace folder should contain "imported-t0ken-value"
    And no file under the app data folder should contain "imported-t0ken-value"
    And the workspace file "acme/collections/token-api/collection.yaml" should contain the field "auth.oauth2.clientId" equal to "xenon"
    And the workspace file "acme/collections/token-api/collection.yaml" should not contain the field "auth.oauth2.accessToken"
    And the workspace file "acme/collections/token-api/me.yaml" should contain the field "auth.oauth2.accessToken" equal to "{{token_var}}"
    When I call "OAuthService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{token_collection}}", "which": "access" }
      """
    Then the response should contain the field "value" equal to "imported-t0ken-value"

  @variables @secrets @xen_009
  Scenario: A collection variable Postman marks as secret goes to the vault, never to the collection file
    Given the keychain is available
    And the log output is captured
    And the file "secret-vars.json" exists with content
      """
      {
        "info": { "_postman_id": "5ec4e700-0000-4000-8000-000000000001", "name": "Secret Vars", "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json" },
        "variable": [
          { "key": "plain", "value": "visible" },
          { "key": "apiSecret", "value": "s3cr3t-fixture-v21", "type": "secret" },
          { "key": "emptySecret", "value": "", "type": "secret" }
        ],
        "item": [ { "name": "Ping", "request": { "method": "GET", "url": "https://secret-vars.test/ping" } } ]
      }
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/secret-vars.json"] }
      """
    And I save the response field "source_id" as "secret_source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{secret_source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should not contain "s3cr3t-fixture-v21"
    And I save the response field "items.0.id" as "secret_vars"
    And the workspace file "acme/collections/secret-vars/collection.yaml" should contain the field "variables.0.value" equal to "visible"
    And the workspace file "acme/collections/secret-vars/collection.yaml" should contain the field "variables.1.key" equal to "apiSecret"
    And the workspace file "acme/collections/secret-vars/collection.yaml" should contain the field "variables.1.secret" equal to "true"
    And the workspace file "acme/collections/secret-vars/collection.yaml" should not contain the field "variables.1.value"
    And the workspace file "acme/collections/secret-vars/collection.yaml" should contain the field "variables.2.secret" equal to "true"
    And no file under the workspace folder should contain "s3cr3t-fixture-v21"
    And the log output should not contain "s3cr3t-fixture-v21"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "collection:{{secret_vars}}", "key": "apiSecret" }
      """
    Then the call should succeed
    And the response should contain the field "value" equal to "s3cr3t-fixture-v21"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "collection:{{secret_vars}}" }
      """
    Then the response should contain the field "rows.1.secret_state" equal to "set"
    And the response should contain the field "rows.2.secret_state" equal to "missing"
