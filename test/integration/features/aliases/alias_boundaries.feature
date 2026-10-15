#language: en
#utf-8

@all @xen_016 @aliases @alias_boundaries
Feature: Aliases travel with export, import, duplicate, move and Save As, and stay out of the documentation
  Postman has no aliases, so a Postman export writes each alias as its own request right after its
  Default, named "<request> (<alias>)", with the alias's content and no examples; importing that file
  back gives separate requests. A Xenon YAML export carries aliases unchanged. Duplicating or moving
  a request carries its aliases; Save As from an alias creates a request without aliases. A
  collection's OpenAPI document is generated from Defaults only and applying it never touches
  aliases.

  Background:
    Given the workspace "Beta" exists
    And the workspace "Acme" exists
    And the "billing" variable is set to "01JU0000000000000000000001"
    And the "login" variable is set to "01JU0000000000000000000002"
    And the "orders" variable is set to "01JU0000000000000000000003"
    And the "auth_folder" variable is set to "01JU0000000000000000000004"
    And the "test_user" variable is set to "01JU0000000000000000000010"
    And the "mx" variable is set to "01JU0000000000000000000011"
    And the "ok" variable is set to "01JU0000000000000000000020"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JU0000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      order:
        - auth
        - login.yaml
        - orders.yaml
      """
    And the workspace file "acme/collections/billing/auth/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{auth_folder}}
      name: Auth
      """
    And the workspace file "acme/collections/billing/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: https://api.acme.test/v1/login
      body:
        mode: raw
        raw:
          language: json
          content: '{"username":"luis"}'
      examples:
        - id: {{ok}}
          name: OK
          request:
            method: POST
            url: https://api.acme.test/v1/login
          response:
            code: 200
            status: OK
            bodyEncoding: text
            body: '{"token":"t"}'
      aliases:
        - id: {{test_user}}
          name: test-user
          method: POST
          url: https://api.acme.test/v2/login
          headers:
            - key: X-Canary
              value: "true"
              enabled: true
              description: ""
          auth:
            type: bearer
            bearer:
              token: typed-alias-token
          body:
            mode: raw
            raw:
              language: json
              content: '{"username":"test"}'
        - id: {{mx}}
          name: mx-test-user
          method: POST
          url: https://api.acme.test/v3/login
          body:
            mode: raw
            raw:
              language: json
              content: '{"username":"mx-test"}'
      """
    And the workspace file "acme/collections/billing/orders.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{orders}}
      name: Orders
      method: GET
      url: https://api.acme.test/v1/orders
      """

  @postman_export
  Scenario: A Postman export writes each alias as a request right after its Default
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{billing}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/billing.postman_collection.json"
      }
      """
    Then the call should succeed
    And the response should contain the field "counts.requests" equal to "4"
    And the response field "warnings" should contain an item with "code" equal to "literal_credential" and "item_path" equal to "Billing/Login (test-user)"
    And the file "billing.postman_collection.json" should be a valid Postman v2.1 collection
    And the file "billing.postman_collection.json" should contain the field "item" array length equal to 5
    And the file "billing.postman_collection.json" should contain the field "item.0.name" equal to "Auth"
    And the file "billing.postman_collection.json" should contain the field "item.1.name" equal to "Login"
    And the file "billing.postman_collection.json" should contain the field "item.1.request.url.raw" equal to "https://api.acme.test/v1/login"
    And the file "billing.postman_collection.json" should contain the field "item.1.response" array length equal to 1
    And the file "billing.postman_collection.json" should contain the field "item.2.name" equal to "Login (test-user)"
    And the file "billing.postman_collection.json" should contain the field "item.2.request.method" equal to "POST"
    And the file "billing.postman_collection.json" should contain the field "item.2.request.url.raw" equal to "https://api.acme.test/v2/login"
    And the file "billing.postman_collection.json" should contain the field "item.2.request.header.0.key" equal to "X-Canary"
    And the file "billing.postman_collection.json" should contain the field "item.2.request.body.raw" equal to
      """
      {"username":"test"}
      """
    And the file "billing.postman_collection.json" should contain the field "item.2.request.auth.type" equal to "bearer"
    And the file "billing.postman_collection.json" should not contain the field "item.2.response.0"
    And the file "billing.postman_collection.json" should contain the field "item.3.name" equal to "Login (mx-test-user)"
    And the file "billing.postman_collection.json" should contain the field "item.3.request.url.raw" equal to "https://api.acme.test/v3/login"
    And the file "billing.postman_collection.json" should contain the field "item.4.name" equal to "Orders"

  @postman_export @roundtrip
  Scenario: Importing the Postman export back gives separate requests, not aliases
    Given I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{billing}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/billing.postman_collection.json"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/billing.postman_collection.json"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "workspaces.0.created.requests" equal to "4"
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Beta.local_id}}" }
      """
    Then the response should contain the field "0.children.1.name" equal to "Login"
    And the response should contain the field "0.children.2.name" equal to "Login (test-user)"
    And the response should contain the field "0.children.3.name" equal to "Login (mx-test-user)"
    And I save the response field "0.children.1.id" as "imported_login"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Beta.local_id}}", "id": "{{imported_login}}" }
      """
    Then the response should contain the field "aliases" array length equal to 0

  @xenon_export @roundtrip
  Scenario: A Xenon YAML export imported into another workspace keeps the aliases
    Given I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{billing}}",
        "format": "xenon_yaml", "shape": "bundle", "target_path": "{{temp_root}}/billing.yaml"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/billing.yaml"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "workspaces.0.created.requests" equal to "2"
    And the workspace file "beta/collections/billing/login.yaml" should contain the field "aliases" array length equal to 2
    And the workspace file "beta/collections/billing/login.yaml" should contain the field "aliases.0.name" equal to "test-user"
    And the workspace file "beta/collections/billing/login.yaml" should contain the field "aliases.0.url" equal to "https://api.acme.test/v2/login"
    And the workspace file "beta/collections/billing/login.yaml" should contain the field "aliases.0.headers.0.key" equal to "X-Canary"
    And the workspace file "beta/collections/billing/login.yaml" should contain the field "aliases.1.name" equal to "mx-test-user"
    And the workspace file "beta/collections/billing/login.yaml" should contain the field "aliases.1.body.raw.content" equal to
      """
      {"username":"mx-test"}
      """
    And the workspace file "beta/collections/billing/login.yaml" should contain the field "url" equal to "https://api.acme.test/v1/login"
    And the workspace file "beta/collections/billing/orders.yaml" should not contain the field "aliases"

  @duplicate @move
  Scenario: Duplicating a request copies its aliases and moving it keeps them
    When I call "TreeService.Duplicate" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Login Copy"
    And I save the response field "id" as "copy"
    And the workspace file "acme/collections/billing/login-copy.yaml" should contain the field "aliases" array length equal to 2
    And the workspace file "acme/collections/billing/login-copy.yaml" should contain the field "aliases.0.name" equal to "test-user"
    And the workspace file "acme/collections/billing/login-copy.yaml" should contain the field "aliases.0.url" equal to "https://api.acme.test/v2/login"
    And the workspace file "acme/collections/billing/login-copy.yaml" should contain the field "aliases.1.name" equal to "mx-test-user"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases" array length equal to 2
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{copy}}", "new_parent_id": "{{auth_folder}}", "index": 0 }
      """
    Then the call should succeed
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{copy}}" }
      """
    Then the response should contain the field "aliases" array length equal to 2
    And the response should contain the field "aliases.1.request.url" equal to "https://api.acme.test/v3/login"

  @save_as
  Scenario: Save As with an alias's content creates a request without aliases and leaves the original as it was
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    Then the response should contain the field "aliases.1.name" equal to "mx-test-user"
    When I call "RequestService.SaveAs" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "parent_id": "{{billing}}",
        "name": "Login MX",
        "request": {
          "id": "{{login}}", "name": "Login", "method": "POST", "url": "https://api.acme.test/v3/login",
          "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"username\":\"mx-test\"}" } }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "url" equal to "https://api.acme.test/v3/login"
    And the response should contain the field "aliases" array length equal to 0
    And the workspace file "acme/collections/billing/login-mx.yaml" should not contain the field "aliases"
    And the workspace file "acme/collections/billing/login-mx.yaml" should not contain the field "examples"
    And the workspace file "acme/collections/billing/login.yaml" should be unchanged

  @documentation
  Scenario: The OpenAPI document shows Defaults only and applying it keeps the aliases
    When I call "TreeService.GetDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}" }
      """
    Then the call should succeed
    And the response should contain the field "yaml" containing "/login"
    And the response should not contain "/v2"
    And the response should not contain "/v3"
    And the response should not contain "test-user"
    And I save the response field "yaml" as "generated"
    When I call "TreeService.ApplyDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "yaml": "{{generated}}" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases" array length equal to 2
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.url" equal to "https://api.acme.test/v2/login"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    Then the response should contain the field "aliases" array length equal to 2
    And the response should contain the field "aliases.0.request.auth.type" equal to "bearer"
