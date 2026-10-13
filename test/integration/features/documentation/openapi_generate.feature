#language: en
#utf-8

@all @xen_007 @documentation @documentation_generate
Feature: A collection is documented as an OpenAPI 3.1 document
  The Documentation tab of a collection shows the collection as an OpenAPI 3.1 (Swagger) document,
  built from the requests themselves: folders become tags ("Products / Admin" for nested folders),
  each request becomes an operation under its path and method with its name, description, path,
  query and header parameters, body and saved examples, and the auth type it uses (never its
  values) becomes a security scheme. "{{variables}}" stay as written, the server is the common
  start of the request URLs, and every operation carries "x-xenon-id" so edits map back onto its
  request. A second request with the same method and path is listed under "x-xenon-variants" of
  the first. Top-level folders group their own tag and every nested folder's tag under
  "x-tagGroups", the way Redoc shows sections. Secrets never appear: nothing is resolved, and auth values are never written.

  Background:
    Given the workspace "Acme" exists
    And the "shop" variable is set to "01JK7000000000000000000001"
    And the "products" variable is set to "01JK7000000000000000000002"
    And the "admin" variable is set to "01JK7000000000000000000003"
    And the "auth" variable is set to "01JK7000000000000000000004"
    And the "list" variable is set to "01JK7000000000000000000010"
    And the "get" variable is set to "01JK7000000000000000000011"
    And the "create" variable is set to "01JK7000000000000000000012"
    And the "login" variable is set to "01JK7000000000000000000013"
    And the "upload" variable is set to "01JK7000000000000000000014"
    And the "getV2" variable is set to "01JK7000000000000000000015"
    And the "page" variable is set to "01JK7000000000000000000020"
    And the "found" variable is set to "01JK7000000000000000000021"
    And the "missing" variable is set to "01JK7000000000000000000022"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK7000000000000000000000
      name: Acme
      order:
        - shop
      """
    And the workspace file "acme/collections/shop/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{shop}}
      name: Shop API
      description: The shop's public API.
      order:
        - products
        - auth
        - upload-avatar.yaml
        - get-product-v2.yaml
      auth:
        type: bearer
        bearer:
          token: s3cr3t-collection-token
      variables:
        - key: baseUrl
          value: https://shop.example.com
          enabled: true
          description: ""
        - key: apiSecret
          value: s3cr3t-variable-value
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/shop/products/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{products}}
      name: Products
      order:
        - list-products.yaml
        - get-product.yaml
        - admin
      """
    And the workspace file "acme/collections/shop/products/admin/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{admin}}
      name: Admin
      order:
        - create-product.yaml
      auth:
        type: apikey
        apikey:
          key: X-API-Key
          value: s3cr3t-admin-key
          in: header
      """
    And the workspace file "acme/collections/shop/auth/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{auth}}
      name: Auth
      order:
        - login.yaml
      """
    And the workspace file "acme/collections/shop/products/list-products.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{list}}
      name: List products
      description: Lists the products, one page at a time.
      method: GET
      url: "{{baseUrl}}/products?limit=10&sort=name"
      params:
        - key: limit
          value: "10"
          enabled: true
          description: Page size
        - key: sort
          value: name
          enabled: true
          description: ""
        - key: debug
          value: "true"
          enabled: false
          description: ""
      headers:
        - key: Accept
          value: application/json
          enabled: true
          description: ""
        - key: X-Trace
          value: "{{traceId}}"
          enabled: true
          description: Correlation id
      auth:
        type: inherit
      examples:
        - id: {{page}}
          name: A page
          request:
            method: GET
            url: "{{baseUrl}}/products?limit=10&sort=name"
          response:
            code: 200
            status: OK
            contentType: application/json; charset=utf-8
            bodyEncoding: text
            body: '[{"id":1,"name":"Mug"}]'
      """
    And the workspace file "acme/collections/shop/products/get-product.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{get}}
      name: Get product
      method: GET
      url: "{{baseUrl}}/products/:id"
      pathVariables:
        - key: id
          value: "42"
          description: Product id
      auth:
        type: inherit
      examples:
        - id: {{found}}
          name: Found
          request:
            method: GET
            url: "{{baseUrl}}/products/42"
          response:
            code: 200
            status: OK
            contentType: application/json
            bodyEncoding: text
            body: '{"id":42,"name":"Mug"}'
        - id: {{missing}}
          name: Missing
          request:
            method: GET
            url: "{{baseUrl}}/products/0"
          response:
            code: 404
            status: Not Found
            contentType: text/plain
            bodyEncoding: text
            body: no such product
      """
    And the workspace file "acme/collections/shop/products/admin/create-product.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create product
      method: POST
      url: "{{baseUrl}}/products"
      headers:
        - key: Content-Type
          value: application/json
          enabled: true
          description: ""
      body:
        mode: raw
        raw:
          language: json
          content: '{"name": "Mug", "price": {{price}}}'
      auth:
        type: inherit
      """
    And the workspace file "acme/collections/shop/auth/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: "{{baseUrl}}/login"
      body:
        mode: urlencoded
        urlencoded:
          - key: username
            value: alice
            enabled: true
            description: Account name
          - key: password
            value: "{{password}}"
            enabled: true
            description: ""
      auth:
        type: none
      """
    And the workspace file "acme/collections/shop/upload-avatar.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{upload}}
      name: Upload avatar
      method: POST
      url: "{{baseUrl}}/users/{{userId}}/avatar"
      body:
        mode: formdata
        formdata:
          - key: avatar
            type: file
            src: avatar.png
            enabled: true
            description: ""
          - key: caption
            type: text
            value: Me
            enabled: true
            description: ""
      auth:
        type: oauth2
        oauth2:
          grant_type: client_credentials
          accessTokenUrl: "{{baseUrl}}/oauth/token"
          clientId: shop-app
          clientSecret: s3cr3t-client-secret
          scope: avatars:write profile
      """
    And the workspace file "acme/collections/shop/get-product-v2.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{getV2}}
      name: Get product (v2)
      method: GET
      url: "{{baseUrl}}/products/:id"
      pathVariables:
        - key: id
          value: "7"
          description: ""
      headers:
        - key: X-Api-Version
          value: "2"
          enabled: true
          description: ""
      auth:
        type: basic
        basic:
          username: alice
          password: s3cr3t-password
      """

  @golden
  Scenario: The documentation of a representative collection matches the expected OpenAPI document
    When I call "TreeService.GetDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{shop}}" }
      """
    Then the call should succeed
    And the response field "yaml" should equal the resource file "documentation/shop-api.openapi.yaml"

  @secrets
  Scenario: No auth value, variable value or secret appears in the documentation
    When I call "TreeService.GetDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{shop}}" }
      """
    Then the call should succeed
    And the response should not contain "s3cr3t"
    And the response should not contain "https://shop.example.com"
    And the response should not contain "shop-app"

  @preview
  Scenario: The rendered view lists every operation in document order with its folder, method, path and request
    When I call "TreeService.GetDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{shop}}" }
      """
    Then the call should succeed
    And the response should contain the field "document.title" equal to "Shop API"
    And the response should contain the field "document.description" equal to "The shop's public API."
    And the response should contain the field "document.servers.0" equal to "{{baseUrl}}"
    And the response should contain the field "document.operations" array length equal to 6
    And the response should contain the field "document.operations.0.request_id" equal to "{{list}}"
    And the response should contain the field "document.operations.0.tag" equal to "Products"
    And the response should contain the field "document.operations.0.method" equal to "GET"
    And the response should contain the field "document.operations.0.path" equal to "/products"
    And the response should contain the field "document.operations.0.security.0" equal to "bearerAuth"
    And the response should contain the field "document.operations.1.request_id" equal to "{{create}}"
    And the response should contain the field "document.operations.1.tag" equal to "Products / Admin"
    And the response should contain the field "document.operations.1.request_body.media_type" equal to "application/json"
    And the response should contain the field "document.operations.2.responses.1.code" equal to "404"
    And the response should contain the field "document.operations.3.request_id" equal to "{{getV2}}"
    And the response should contain the field "document.operations.3.method" equal to "GET"
    And the response should contain the field "document.operations.3.path" equal to "/products/{id}"
    And the response should contain the field "document.operations.3.variant" equal to "true"
    And the response should contain the field "document.security_schemes.1.name" equal to "apiKeyAuth"
    And the response should contain the field "document.security_schemes.1.type" equal to "apiKey"

  @tag_groups
  Scenario: Every folder is a tag named by its path, grouped under its top-level folder in x-tagGroups
    When I call "TreeService.GetDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{shop}}" }
      """
    Then the call should succeed
    And the response should contain the field "document.tags" array length equal to 3
    And the response should contain the field "document.tags.0.name" equal to "Products"
    And the response should contain the field "document.tags.1.name" equal to "Products / Admin"
    And the response should contain the field "document.tags.2.name" equal to "Auth"
    And the response should contain the field "document.tag_groups" array length equal to 2
    And the response should contain the field "document.tag_groups.0.name" equal to "Products"
    And the response should contain the field "document.tag_groups.0.tags" with the items "Products, Products / Admin"
    And the response should contain the field "document.tag_groups.1.name" equal to "Auth"
    And the response should contain the field "document.tag_groups.1.tags" with the items "Auth"

  @folders
  Scenario: Requests have no documentation of their own
    When I call "TreeService.GetDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{get}}" }
      """
    Then the call should fail with code "DOC-05400"

  @empty
  Scenario: A collection without requests is documented with no paths
    Given the workspace file "acme/collections/empty/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JK7000000000000000000099
      name: Empty
      """
    When I call "TreeService.GetDocumentation" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JK7000000000000000000099" }
      """
    Then the call should succeed
    And the response should contain the field "yaml" equal to
      """
      openapi: 3.1.0
      info:
        title: Empty
        version: 1.0.0
      paths: {}

      """
