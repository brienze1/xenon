#language: en
#utf-8

@all @xen_009 @import @postman_v3_collection
Feature: Import a Postman v3 collection (JSON and git-native folders)
  Postman's v3 format comes in two shapes: the JSON `postman collection get --json` prints, and the git-native folder
  `postman workspace pull` writes (one *.request.yaml per request, .resources/definition.yaml per collection or
  folder, examples as *.example.yaml files). Both import with the same fidelity as the v2.1 export of the same
  collection: the same tree and order, requests, bodies, auth, scripts, variables, descriptions and saved examples,
  and the same Postman-only fields kept for export. v3-only cases are reported in the summary: extra multi-auth
  entries, GraphQL/gRPC-only collection scripts, and request protocols Xenon does not support. Secret collection
  variables go to the vault. A single *.request.yaml goes into an existing collection or folder.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available

  @fidelity
  Scenario Outline: A v3 collection imports exactly like its v2.1 export - <form>
    Given the workspace "Globex" exists
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/rich-v2.1.json"] }
      """
    And I save the response field "source_id" as "v21"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{v21}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/<source>"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "postman_collection"
    And the response should contain the field "version" equal to "3"
    And the response should contain the field "items" array length equal to 1
    And the response should contain the field "items.0.kind" equal to "collection"
    And the response should contain the field "items.0.name" equal to "Rich API"
    And the response should contain the field "items.0.counts.folders" equal to "2"
    And the response should contain the field "items.0.counts.requests" equal to "16"
    And the response should contain the field "items.0.counts.examples" equal to "1"
    And the response should contain the field "items.0.counts.variables" equal to "3"
    And I save the response field "source_id" as "v3"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{v3}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Globex.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the response should contain the field "failed" array length equal to 0
    And the folder "workspaces/globex/collections/rich-api" should equal the folder "workspaces/acme/collections/rich-api" apart from ids

    Examples:
      | form   | source    |
      | json   | rich.json |
      | folder | rich      |

  @tree
  Scenario Outline: The tree, ids and order come through - <form>
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/<source>"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "workspaces.0.created.collections" equal to "1"
    And the response should contain the field "workspaces.0.created.folders" equal to "2"
    And the response should contain the field "workspaces.0.created.requests" equal to "16"
    And the response should contain the field "workspaces.0.created.examples" equal to "1"
    And the workspace file "acme/collections/rich-api/collection.yaml" should contain the field "order" with the items "users, bodies"
    And the workspace file "acme/collections/rich-api/users/folder.yaml" should contain the field "order" with the items "get-user.yaml, inherit-auth.yaml, no-auth.yaml"
    And the workspace file "acme/collections/rich-api/bodies/folder.yaml" should contain the field "order" with the items "json.yaml, text.yaml, xml.yaml, html.yaml, javascript.yaml, form.yaml, urlenc.yaml, graphql.yaml, binary.yaml, apikey.yaml, digest.yaml, hawk.yaml, redirects.yaml"
    And the workspace file "acme/collections/rich-api/collection.yaml" should contain the field "id" matching "^[0-9A-HJKMNP-TV-Z]{26}$"
    And the workspace file "acme/collections/rich-api/collection.yaml" should contain the field "postman.info._postman_id" equal to "a0000000-0000-4000-8000-000000000001"
    And the workspace file "acme/collections/rich-api/users/folder.yaml" should contain the field "postman.id" equal to "b0000000-0000-4000-8000-000000000001"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "postman.id" equal to "c0000000-0000-4000-8000-000000000001"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "examples.0.postman.id" equal to "d0000000-0000-4000-8000-000000000001"

    Examples:
      | form   | source    |
      | json   | rich.json |
      | folder | rich      |

  @mapping
  Scenario Outline: Requests keep method, URL, rows, auth, scripts, settings, descriptions and examples - <form>
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/<source>"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/rich-api/collection.yaml" should contain the field "description" equal to "The **Rich** API"
    And the workspace file "acme/collections/rich-api/collection.yaml" should contain the field "auth.type" equal to "bearer"
    And the workspace file "acme/collections/rich-api/collection.yaml" should contain the field "auth.bearer.token" equal to "{{token}}"
    And the workspace file "acme/collections/rich-api/collection.yaml" should contain the field "scripts.preRequest" equal to
      """
      pm.environment.set("ts", Date.now());
      """
    And the workspace file "acme/collections/rich-api/collection.yaml" should contain the field "scripts.postResponse" equal to
      """
      pm.test("status ok", () => pm.response.to.have.status(200));
      """
    And the workspace file "acme/collections/rich-api/collection.yaml" should contain the field "variables.0.key" equal to "baseUrl"
    And the workspace file "acme/collections/rich-api/collection.yaml" should contain the field "variables.0.value" equal to "https://api.rich.test"
    And the workspace file "acme/collections/rich-api/collection.yaml" should contain the field "variables.1.enabled" equal to "false"
    And the workspace file "acme/collections/rich-api/collection.yaml" should contain the field "variables.2.description" equal to "Default page"
    And the workspace file "acme/collections/rich-api/collection.yaml" should not contain the field "postman.variable"
    And the workspace file "acme/collections/rich-api/users/folder.yaml" should contain the field "description" equal to "Users docs"
    And the workspace file "acme/collections/rich-api/users/folder.yaml" should contain the field "auth.type" equal to "basic"
    And the workspace file "acme/collections/rich-api/users/folder.yaml" should contain the field "auth.basic.username" equal to "{{user}}"
    And the workspace file "acme/collections/rich-api/users/folder.yaml" should contain the field "scripts.preRequest" equal to
      """
      console.log("users folder");
      """
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "method" equal to "GET"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "url" equal to "{{baseUrl}}/users/:id?page=1&skip=2"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "description" equal to "Request docs"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "params.0.key" equal to "page"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "params.0.description" equal to "Page"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "params.1.key" equal to "skip"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "params.1.enabled" equal to "false"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "pathVariables.0.key" equal to "id"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "pathVariables.0.value" equal to "42"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "pathVariables.0.description" equal to "User id"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "headers.0.key" equal to "Accept"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "headers.1.key" equal to "X-Off"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "headers.1.enabled" equal to "false"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "headers.1.description" equal to "off header"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "auth.type" equal to "inherit"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "scripts.postResponse" equal to
      """
      pm.test("200", () => pm.response.to.have.status(200));
      """
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "examples.0.name" equal to "200 OK"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "examples.0.request.method" equal to "GET"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "examples.0.request.url" equal to "{{baseUrl}}/users/42"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "examples.0.response.code" equal to "200"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "examples.0.response.status" equal to "OK"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "examples.0.response.contentType" equal to "application/json"
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "examples.0.response.body" equal to
      """
      {"id":42}
      """
    And the workspace file "acme/collections/rich-api/users/get-user.yaml" should contain the field "examples.0.postman._postman_previewlanguage" equal to "json"
    And the workspace file "acme/collections/rich-api/users/inherit-auth.yaml" should contain the field "auth.type" equal to "inherit"
    And the workspace file "acme/collections/rich-api/users/no-auth.yaml" should contain the field "auth.type" equal to "none"
    And the workspace file "acme/collections/rich-api/bodies/redirects.yaml" should contain the field "settings.followRedirects" equal to "false"
    And the workspace file "acme/collections/rich-api/bodies/redirects.yaml" should contain the field "settings.verifySsl" equal to "false"
    And the workspace file "acme/collections/rich-api/bodies/redirects.yaml" should contain the field "postman.protocolProfileBehavior.maxRedirects" equal to "3"
    And the workspace file "acme/collections/rich-api/bodies/json.yaml" should not contain the field "postman.protocolProfileBehavior"

    Examples:
      | form   | source    |
      | json   | rich.json |
      | folder | rich      |

  @bodies
  Scenario Outline: Every v3 body type maps to the matching Xenon body - <form>
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/<source>"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response field "warnings" should contain 2 items with "code" equal to "missing_file"
    And the response field "warnings" should contain an item with "code" equal to "missing_file" and "path" equal to "Rich API/Bodies/Form"
    And the response field "warnings" should contain an item with "code" equal to "missing_file" and "path" equal to "Rich API/Bodies/Binary"
    And the workspace file "acme/collections/rich-api/bodies/<file>" should contain the field "body.mode" equal to "<mode>"
    And the workspace file "acme/collections/rich-api/bodies/<file>" should contain the field "<field>" equal to "<value>"

    Examples:
      | form   | source    | file            | mode       | field                       | value                              |
      | json   | rich.json | json.yaml       | raw        | body.raw.language           | json                               |
      | json   | rich.json | text.yaml       | raw        | body.raw.content            | hello                              |
      | json   | rich.json | xml.yaml        | raw        | body.raw.language           | xml                                |
      | json   | rich.json | html.yaml       | raw        | body.raw.language           | html                               |
      | json   | rich.json | javascript.yaml | raw        | body.raw.language           | javascript                         |
      | json   | rich.json | form.yaml       | formdata   | body.formdata.1.enabled     | false                              |
      | json   | rich.json | form.yaml       | formdata   | body.formdata.2.srcState    | repick                             |
      | json   | rich.json | urlenc.yaml     | urlencoded | body.urlencoded.1.enabled   | false                              |
      | json   | rich.json | graphql.yaml    | graphql    | body.graphql.query          | query { me { id } }                |
      | json   | rich.json | binary.yaml     | binary     | body.binary.src             | /nonexistent/xenon-fixture/blob.bin |
      | folder | rich      | json.yaml       | raw        | body.raw.language           | json                               |
      | folder | rich      | text.yaml       | raw        | body.raw.language           | text                               |
      | folder | rich      | form.yaml       | formdata   | body.formdata.2.src         | /nonexistent/xenon-fixture/a.png   |
      | folder | rich      | urlenc.yaml     | urlencoded | body.urlencoded.0.key       | a                                  |
      | folder | rich      | graphql.yaml    | graphql    | body.graphql.query          | query { me { id } }                |
      | folder | rich      | binary.yaml     | binary     | body.binary.srcState        | repick                             |

  @auth
  Scenario Outline: Auth maps like v2.1 and unsupported types are kept for export - <file>
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/rich"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response field "warnings" should contain an item with "code" equal to "unsupported_auth" and "path" equal to "Rich API/Bodies/Hawk"
    And the workspace file "acme/collections/rich-api/bodies/<file>" should contain the field "auth.type" equal to "<type>"
    And the workspace file "acme/collections/rich-api/bodies/<file>" should contain the field "<field>" equal to "<value>"

    Examples:
      | file        | type        | field                   | value      |
      | apikey.yaml | apikey      | auth.apikey.key         | X-Key      |
      | apikey.yaml | apikey      | auth.apikey.in          | header     |
      | digest.yaml | digest      | auth.digest.password    | {{pass}}   |
      | hawk.yaml   | unsupported | auth.unsupported.type   | hawk       |
      | hawk.yaml   | unsupported | auth.unsupported.source | postman    |

  @paste
  Scenario: Pasted v3 JSON is detected like the same JSON from a file
    When I call "ImportService.Detect" with
      """
      { "text": "{\"$kind\":\"collection\",\"name\":\"Pasted v3\",\"items\":[{\"$kind\":\"http-request\",\"name\":\"Ping\",\"method\":\"GET\",\"url\":\"https://pasted.test/ping\"}]}" }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "postman_collection"
    And the response should contain the field "version" equal to "3"
    And the response should contain the field "items.0.name" equal to "Pasted v3"
    And the response should contain the field "items.0.counts.requests" equal to "1"

  @extras @order
  Scenario: Order, names, map and list rows, GraphQL requests and settings in a git-native folder
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/extras"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/extras/collection.yaml" should contain the field "name" equal to "Extras"
    And the workspace file "acme/collections/extras/collection.yaml" should contain the field "order" with the items "nested, list-headers.yaml, map-headers.yaml, health-check.yaml, query.yaml, redirects.yaml, no-order.yaml"
    And the workspace file "acme/collections/extras/health-check.yaml" should contain the field "name" equal to "Health/check"
    And the workspace file "acme/collections/extras/map-headers.yaml" should contain the field "headers.0.key" equal to "Content-Type"
    And the workspace file "acme/collections/extras/map-headers.yaml" should contain the field "headers.0.value" equal to "application/json"
    And the workspace file "acme/collections/extras/map-headers.yaml" should contain the field "headers.1.key" equal to "X-Trace"
    And the workspace file "acme/collections/extras/map-headers.yaml" should contain the field "auth.bearer.token" equal to "{{mapToken}}"
    And the workspace file "acme/collections/extras/map-headers.yaml" should not contain the field "postman.protocolProfileBehavior"
    And the workspace file "acme/collections/extras/map-headers.yaml" should contain the field "examples.0.name" equal to "First"
    And the workspace file "acme/collections/extras/map-headers.yaml" should contain the field "examples.0.response.code" equal to "201"
    And the workspace file "acme/collections/extras/map-headers.yaml" should contain the field "examples.0.response.headers.0.key" equal to "Content-Type"
    And the workspace file "acme/collections/extras/map-headers.yaml" should contain the field "examples.1.name" equal to "Second"
    And the workspace file "acme/collections/extras/map-headers.yaml" should contain the field "examples.1.response.code" equal to "400"
    And the workspace file "acme/collections/extras/list-headers.yaml" should contain the field "headers.0.key" equal to "Content-Type"
    And the workspace file "acme/collections/extras/list-headers.yaml" should contain the field "headers.1.value" equal to "1"
    And the workspace file "acme/collections/extras/list-headers.yaml" should contain the field "auth.bearer.token" equal to "{{listToken}}"
    And the workspace file "acme/collections/extras/query.yaml" should contain the field "method" equal to "POST"
    And the workspace file "acme/collections/extras/query.yaml" should contain the field "url" equal to "{{host}}/graphql"
    And the workspace file "acme/collections/extras/query.yaml" should contain the field "headers.0.key" equal to "X-Gql"
    And the workspace file "acme/collections/extras/query.yaml" should contain the field "body.mode" equal to "graphql"
    And the workspace file "acme/collections/extras/query.yaml" should contain the field "body.graphql.query" equal to "query { me { id } }"
    And the workspace file "acme/collections/extras/query.yaml" should contain the field "body.graphql.variables" equal to
      """
      {"x":1}
      """
    And the workspace file "acme/collections/extras/query.yaml" should contain the field "scripts.preRequest" equal to
      """
      console.log("before query");
      """
    And the workspace file "acme/collections/extras/redirects.yaml" should contain the field "settings.followRedirects" equal to "false"
    And the workspace file "acme/collections/extras/redirects.yaml" should contain the field "postman.protocolProfileBehavior.maxRedirects" equal to "3"
    And the workspace file "acme/collections/extras/redirects.yaml" should not contain the field "postman.protocolProfileBehavior.disabledSystemHeaders"
    And the workspace file "acme/collections/extras/nested/folder.yaml" should contain the field "postman.variable.0.key" equal to "folderVar"
    And the workspace folder "acme/collections/extras" should contain the entries "collection.yaml, health-check.yaml, list-headers.yaml, map-headers.yaml, nested, no-order.yaml, query.yaml, redirects.yaml"

  @extras @warnings
  Scenario: Every v3 part Xenon cannot represent is reported once
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/extras"] }
      """
    Then the call should succeed
    And the response field "warnings" should contain 1 items with "code" equal to "unsupported_auth"
    And the response field "warnings" should contain an item with "code" equal to "unsupported_auth" and "message" equal to "Postman multi-auth on Extras lists 2 auths; Xenon imported the first (bearer) and kept basic for export only"
    And the response field "warnings" should contain an item with "code" equal to "unsupported_auth" and "path" equal to "Extras"
    And the response field "warnings" should contain 1 items with "code" equal to "protocol_script"
    And the response field "warnings" should contain an item with "code" equal to "protocol_script" and "message" equal to "The graphql:beforeQuery script on Extras runs only for GraphQL requests in Postman; it is kept for export and Xenon does not run it"
    And the response field "warnings" should contain 3 items with "code" equal to "unsupported_request_type"
    And the response field "warnings" should contain an item with "code" equal to "unsupported_request_type" and "message" equal to "Stream is a grpc-request request, which Xenon does not support; it was skipped"
    And the response field "warnings" should contain an item with "code" equal to "unsupported_request_type" and "message" equal to "Socket is a websocket-request request, which Xenon does not support; it was skipped"
    And the response field "warnings" should contain an item with "code" equal to "unsupported_request_type" and "message" equal to "Future is a carrier-pigeon-request request, which Xenon does not support; it was skipped"
    And the response field "warnings" should contain 1 items with "code" equal to "folder_variables"
    And the response field "warnings" should contain an item with "code" equal to "folder_variables" and "path" equal to "Extras/Nested"
    When I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/extras/collection.yaml" should contain the field "auth.type" equal to "bearer"
    And the workspace file "acme/collections/extras/collection.yaml" should contain the field "auth.bearer.token" equal to "{{primaryToken}}"
    And the workspace file "acme/collections/extras/collection.yaml" should contain the field "postman.postmanV3.auth.1.type" equal to "basic"
    And the workspace file "acme/collections/extras/collection.yaml" should contain the field "scripts.preRequest" equal to
      """
      console.log("extras pre");
      """
    And the workspace file "acme/collections/extras/collection.yaml" should contain the field "postman.postmanV3.scripts.0.type" equal to "graphql:beforeQuery"

  @extras @secrets
  Scenario: Secret collection variables go to the vault and never to a file or a log
    Given the log output is captured
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/extras"] }
      """
    Then the call should succeed
    And the response should not contain "s3cr3t-fixture-collection"
    When I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should not contain "s3cr3t-fixture-collection"
    And I save the response field "items.0.id" as "extras"
    And the workspace file "acme/collections/extras/collection.yaml" should contain the field "variables.0.key" equal to "host"
    And the workspace file "acme/collections/extras/collection.yaml" should contain the field "variables.0.value" equal to "https://extras.test"
    And the workspace file "acme/collections/extras/collection.yaml" should contain the field "variables.1.key" equal to "apiSecret"
    And the workspace file "acme/collections/extras/collection.yaml" should contain the field "variables.1.secret" equal to "true"
    And the workspace file "acme/collections/extras/collection.yaml" should not contain the field "variables.1.value"
    And the workspace file "acme/collections/extras/collection.yaml" should contain the field "variables.2.key" equal to "masked"
    And the workspace file "acme/collections/extras/collection.yaml" should contain the field "variables.2.secret" equal to "true"
    And the workspace file "acme/collections/extras/collection.yaml" should not contain the field "variables.2.value"
    And no file under the workspace folder should contain "s3cr3t-fixture-collection"
    And no file under the workspace folder should contain "********"
    And no file under the app data folder should contain "s3cr3t-fixture-collection"
    And the log output should not contain "s3cr3t-fixture-collection"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "collection:{{extras}}", "key": "apiSecret" }
      """
    Then the call should succeed
    And the response should contain the field "value" equal to "s3cr3t-fixture-collection"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "collection:{{extras}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.1.secret_state" equal to "set"
    And the response should contain the field "rows.2.secret_state" equal to "missing"

  @request
  Scenario: A single request file goes into an existing folder with its examples
    Given I call "TreeService.CreateCollection" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "Main" }
      """
    And I save the response field "id" as "main"
    And I call "TreeService.CreateFolder" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{main}}", "name": "Target" }
      """
    And I save the response field "id" as "target"
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/extras/Map headers.request.yaml"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "postman_request"
    And the response should contain the field "version" equal to "3"
    And the response should contain the field "items" array length equal to 1
    And the response should contain the field "items.0.kind" equal to "request"
    And the response should contain the field "items.0.name" equal to "Map headers"
    And the response should contain the field "items.0.counts.examples" equal to "2"
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should fail with code "IMP-35400"
    And the error should contain the field "error.description" equal to "A request must be imported into a collection or folder"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{target}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the workspace file "acme/collections/main/target/map-headers.yaml" should contain the field "method" equal to "POST"
    And the workspace file "acme/collections/main/target/map-headers.yaml" should contain the field "headers.0.key" equal to "Content-Type"
    And the workspace file "acme/collections/main/target/map-headers.yaml" should contain the field "examples" array length equal to 2
    And the workspace file "acme/collections/main/target/map-headers.yaml" should contain the field "examples.0.name" equal to "First"

  @malformed
  Scenario: A broken v3 file is refused naming the file and line, never quoting its content
    Given the log output is captured
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/malformed"] }
      """
    Then the call should fail with code "IMP-11400"
    And the error should contain the field "error.description" equal to "Broken.request.yaml is not a valid Postman v3 file: line 4: a value has the wrong type"
    And the log output should not contain "s3cr3t-fixture-broken"
    And the workspace folder "acme" should contain the entries ".gitignore, collections, environments, workspace.yaml"

  @clash
  Scenario Outline: A v3 collection clash is resolved with <resolution>
    Given I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/rich.json"] }
      """
    And I save the response field "source_id" as "first"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{first}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/rich"] }
      """
    And I save the response field "source_id" as "second"
    And I call "ImportService.Clashes" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "clashes" array length equal to 1
    And the response should contain the field "clashes.0.kind" equal to "collection"
    And the response should contain the field "clashes.0.existing_name" equal to "Rich API"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "resolution": "<resolution>" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "<status>"
    And the workspace file "acme/workspace.yaml" should contain the field "order" with the items "<order>"

    Examples:
      | resolution | status   | order                            |
      | replace    | replaced | rich-api                         |
      | keep_both  | imported | rich-api, rich-api-(imported)    |
      | skip       | skipped  | rich-api                         |
