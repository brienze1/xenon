#language: en
#utf-8

@all @xen_016 @aliases @alias_storage
Feature: Request aliases are created, renamed and deleted inside their request's file
  A saved request can hold named aliases: complete, independent versions of its sendable content
  (method, URL, params, path variables, headers, auto-headers, body, auth, scripts, settings). They
  live in the request's own file under "aliases:", in creation order, and a request without aliases
  is written exactly as before. A new alias is a copy of the saved Default. Alias names are trimmed,
  1 to 200 characters, unique within the request ignoring case, and "Default" is reserved. Every
  alias write is refused when the file changed on disk since it was opened, unless the user keeps
  their version.

  Background:
    Given the workspace "Acme" exists
    And the "billing" variable is set to "01JQ0000000000000000000001"
    And the "login" variable is set to "01JQ0000000000000000000002"
    And the "orders" variable is set to "01JQ0000000000000000000003"
    And the "mx" variable is set to "01JQ0000000000000000000010"
    And the "us" variable is set to "01JQ0000000000000000000011"
    And the "unknown" variable is set to "01JQ0000000000000000000099"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JQ0000000000000000000000
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
        - login.yaml
        - orders.yaml
      """
    And the workspace file "acme/collections/billing/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      description: ""
      method: POST
      url: https://api.acme.test/v1/login
      params: []
      pathVariables: []
      headers:
        - key: Content-Type
          value: application/json
          enabled: true
          description: ""
      autoHeaders:
        disabled: []
      auth:
        type: inherit
      body:
        mode: raw
        raw:
          language: json
          content: '{"username":"luis","password":"pw"}'
      settings:
        timeoutMs: 0
        followRedirects: true
        verifySsl: true
        encodeUrl: true
        disableCookieJar: false
      scripts:
        preRequest: ""
        postResponse: ""
      inheritScripts: true
      postman: {}
      """
    And the workspace file "acme/collections/billing/orders.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{orders}}
      name: Orders
      method: GET
      url: https://api.acme.test/v1/orders
      aliases:
        - id: {{mx}}
          name: mx-test-user
          method: GET
          url: https://api.acme.test/v2/orders
          params: []
          pathVariables: []
          headers:
            - key: X-Canary
              value: "true"
              enabled: true
              description: ""
          autoHeaders:
            disabled: []
          auth:
            type: inherit
          body:
            mode: none
          settings:
            timeoutMs: 0
            followRedirects: true
            verifySsl: true
            encodeUrl: true
            disableCookieJar: false
          scripts:
            preRequest: ""
            postResponse: ""
          inheritScripts: true
        - id: {{us}}
          name: us-test-user
          method: GET
          url: https://api.acme.test/v3/orders
      """

  @get
  Scenario: Opening a request returns its aliases in file order with their content
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{orders}}" }
      """
    Then the call should succeed
    And the response should contain the field "url" equal to "https://api.acme.test/v1/orders"
    And the response should contain the field "aliases" array length equal to 2
    And the response should contain the field "aliases.0.id" equal to "{{mx}}"
    And the response should contain the field "aliases.0.name" equal to "mx-test-user"
    And the response should contain the field "aliases.0.request.method" equal to "GET"
    And the response should contain the field "aliases.0.request.url" equal to "https://api.acme.test/v2/orders"
    And the response should contain the field "aliases.0.request.headers.0.key" equal to "X-Canary"
    And the response should contain the field "aliases.0.request.headers.0.value" equal to "true"
    And the response should contain the field "aliases.1.id" equal to "{{us}}"
    And the response should contain the field "aliases.1.name" equal to "us-test-user"
    And the response should contain the field "aliases.1.request.url" equal to "https://api.acme.test/v3/orders"
    And the response should contain the field "headers" array length equal to 0

  @get
  Scenario: A request without aliases returns an empty alias list
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    Then the call should succeed
    And the response should contain the field "aliases" array length equal to 0

  @create
  Scenario: Creating an alias copies the saved Default, appends it and returns the new revision
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "base_revision": "{{revision}}", "name": "  mx-test-user  " }
      """
    Then the call should succeed
    And the response should contain the field "id" equal to "{{login}}"
    And the response should contain the field "revision" not equal to "{{revision}}"
    And the response should contain the field "aliases" array length equal to 1
    And the response should contain the field "aliases.0.id" matching "^[0-9A-Z]{26}$"
    And the response should contain the field "aliases.0.name" equal to "mx-test-user"
    And the response should contain the field "aliases.0.request.method" equal to "POST"
    And the response should contain the field "aliases.0.request.url" equal to "https://api.acme.test/v1/login"
    And the response should contain the field "aliases.0.request.body.raw.content" equal to
      """
      {"username":"luis","password":"pw"}
      """
    And I save the response field "aliases.0.id" as "created"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases" array length equal to 1
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.id" equal to "{{created}}"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.name" equal to "mx-test-user"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.method" equal to "POST"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.url" equal to "https://api.acme.test/v1/login"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.headers.0.key" equal to "Content-Type"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.body.raw.content" equal to
      """
      {"username":"luis","password":"pw"}
      """
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "url" equal to "https://api.acme.test/v1/login"
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.children" array length equal to 2

  @create
  Scenario: A new alias is appended after the existing ones and copies the Default, not another alias
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{orders}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "base_revision": "{{revision}}", "name": "canary" }
      """
    Then the call should succeed
    And the response should contain the field "aliases" array length equal to 3
    And the response should contain the field "aliases.2.name" equal to "canary"
    And the response should contain the field "aliases.2.request.url" equal to "https://api.acme.test/v1/orders"
    And the response should contain the field "aliases.2.request.headers" array length equal to 0
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "aliases.0.id" equal to "{{mx}}"
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "aliases.0.url" equal to "https://api.acme.test/v2/orders"
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "aliases.1.id" equal to "{{us}}"
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "aliases.2.name" equal to "canary"

  @create @names @errors
  Scenario Outline: A new alias needs a valid name that is not taken - <case>
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{orders}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "base_revision": "{{revision}}", "name": "<name>" }
      """
    Then the call should fail with code "<code>"
    And the workspace file "acme/collections/billing/orders.yaml" should be unchanged

    Examples:
      | case                      | name           | code      |
      | blank                     |                | ALS-03400 |
      | reserved                  | Default        | ALS-04400 |
      | reserved in another case  | dEfAuLt        | ALS-04400 |
      | taken                     | mx-test-user   | ALS-05409 |
      | taken in another case     | MX-TEST-USER   | ALS-05409 |

  @create @names @errors
  Scenario: Names are compared after trimming surrounding spaces
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{orders}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "base_revision": "{{revision}}", "name": "   " }
      """
    Then the call should fail with code "ALS-03400"
    When I call "AliasService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "base_revision": "{{revision}}", "name": "  Us-Test-User  " }
      """
    Then the call should fail with code "ALS-05409"
    When I call "AliasService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "base_revision": "{{revision}}", "name": " default " }
      """
    Then the call should fail with code "ALS-04400"
    And the workspace file "acme/collections/billing/orders.yaml" should be unchanged

  @create @names @errors
  Scenario: A name longer than 200 characters is refused and 200 characters are accepted
    Given the "long" variable is set to "a" repeated 201 times separated by ""
    And the "limit" variable is set to "b" repeated 200 times separated by ""
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "base_revision": "{{revision}}", "name": "{{long}}" }
      """
    Then the call should fail with code "ALS-03400"
    And the workspace file "acme/collections/billing/login.yaml" should not contain the field "aliases"
    When I call "AliasService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "base_revision": "{{revision}}", "name": "{{limit}}" }
      """
    Then the call should succeed
    And the response should contain the field "aliases.0.name" equal to "{{limit}}"

  @create @errors
  Scenario: Creating an alias on an unknown request fails
    When I call "AliasService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{unknown}}", "base_revision": "", "name": "canary" }
      """
    Then the call should fail with code "REQ-01404"

  @create @conflict
  Scenario: Creating an alias on a file that changed on disk is refused
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    Given the workspace file "acme/collections/billing/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: https://api.acme.test/v1/changed-by-git
      """
    When I call "AliasService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "base_revision": "{{revision}}", "name": "canary" }
      """
    Then the call should fail with code "REQ-02409"
    And the workspace file "acme/collections/billing/login.yaml" should be unchanged

  @rename
  Scenario: Saving an alias with a new name renames it and keeps the others and the Default
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{orders}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "base_revision": "{{revision}}", "force": false,
        "alias": {
          "id": "{{mx}}", "name": "  qa-user ",
          "request": { "method": "GET", "url": "https://api.acme.test/v2/orders", "headers": [ { "key": "X-Canary", "value": "true", "enabled": true } ] }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "aliases.0.name" equal to "qa-user"
    And the response should contain the field "revision" not equal to "{{revision}}"
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "aliases.0.id" equal to "{{mx}}"
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "aliases.0.name" equal to "qa-user"
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "aliases.1.name" equal to "us-test-user"
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "aliases.1.url" equal to "https://api.acme.test/v3/orders"
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "url" equal to "https://api.acme.test/v1/orders"
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "name" equal to "Orders"

  @rename @names @errors
  Scenario Outline: Renaming an alias follows the name rules and writes nothing on error - <case>
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{orders}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "base_revision": "{{revision}}", "force": false,
        "alias": { "id": "{{mx}}", "name": "<name>", "request": { "method": "GET", "url": "https://api.acme.test/v9/orders" } }
      }
      """
    Then the call should fail with code "<code>"
    And the workspace file "acme/collections/billing/orders.yaml" should be unchanged

    Examples:
      | case                  | name          | code      |
      | blank                 |               | ALS-03400 |
      | reserved              | DEFAULT       | ALS-04400 |
      | taken by another      | Us-Test-User  | ALS-05409 |

  @rename
  Scenario: Saving an alias under its own name in another case is not a clash
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{orders}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "base_revision": "{{revision}}", "force": false,
        "alias": { "id": "{{mx}}", "name": "MX-Test-User", "request": { "method": "GET", "url": "https://api.acme.test/v2/orders" } }
      }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "aliases.0.name" equal to "MX-Test-User"

  @save @errors
  Scenario: Saving an unknown alias fails and writes nothing
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{orders}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "base_revision": "{{revision}}", "force": false,
        "alias": { "id": "{{unknown}}", "name": "ghost", "request": { "method": "GET", "url": "https://api.acme.test/ghost" } }
      }
      """
    Then the call should fail with code "ALS-01404"
    And the workspace file "acme/collections/billing/orders.yaml" should be unchanged

  @save @conflict
  Scenario: Saving an alias over a file that changed on disk is refused unless the user keeps their version
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{orders}}" }
      """
    And I save the response field "revision" as "revision"
    Given the workspace file "acme/collections/billing/orders.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{orders}}
      name: Orders
      method: GET
      url: https://api.acme.test/v1/orders-changed-by-git
      aliases:
        - id: {{mx}}
          name: mx-test-user
          method: GET
          url: https://api.acme.test/v2/orders
      """
    When I call "AliasService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "base_revision": "{{revision}}", "force": false,
        "alias": { "id": "{{mx}}", "name": "mx-test-user", "request": { "method": "GET", "url": "https://api.acme.test/v2/mine" } }
      }
      """
    Then the call should fail with code "REQ-02409"
    And the workspace file "acme/collections/billing/orders.yaml" should be unchanged
    When I call "AliasService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "base_revision": "{{revision}}", "force": true,
        "alias": { "id": "{{mx}}", "name": "mx-test-user", "request": { "method": "GET", "url": "https://api.acme.test/v2/mine" } }
      }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "aliases.0.url" equal to "https://api.acme.test/v2/mine"
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "url" equal to "https://api.acme.test/v1/orders-changed-by-git"

  @delete
  Scenario: Deleting an alias removes only that alias
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{orders}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "alias_id": "{{mx}}", "base_revision": "{{revision}}" }
      """
    Then the call should succeed
    And the response should contain the field "aliases" array length equal to 1
    And the response should contain the field "aliases.0.id" equal to "{{us}}"
    And the response should contain the field "revision" not equal to "{{revision}}"
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "aliases" array length equal to 1
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "aliases.0.name" equal to "us-test-user"
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "url" equal to "https://api.acme.test/v1/orders"

  @delete @byte_identical
  Scenario: Deleting the last alias leaves the file exactly as it was before any alias existed
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "base_revision": "{{revision}}", "name": "temporary" }
      """
    Then the call should succeed
    And I save the response field "revision" as "revision"
    And I save the response field "aliases.0.id" as "temporary"
    When I call "AliasService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "alias_id": "{{temporary}}", "base_revision": "{{revision}}" }
      """
    Then the call should succeed
    And the response should contain the field "aliases" array length equal to 0
    And the workspace file "acme/collections/billing/login.yaml" should equal
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      description: ""
      method: POST
      url: https://api.acme.test/v1/login
      params: []
      pathVariables: []
      headers:
        - key: Content-Type
          value: application/json
          enabled: true
          description: ""
      autoHeaders:
        disabled: []
      auth:
        type: inherit
      body:
        mode: raw
        raw:
          language: json
          content: '{"username":"luis","password":"pw"}'
      settings:
        timeoutMs: 0
        followRedirects: true
        verifySsl: true
        encodeUrl: true
        disableCookieJar: false
      scripts:
        preRequest: ""
        postResponse: ""
      inheritScripts: true
      postman: {}
      """

  @delete @errors
  Scenario: Deleting an unknown alias fails and writes nothing
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{orders}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "alias_id": "{{unknown}}", "base_revision": "{{revision}}" }
      """
    Then the call should fail with code "ALS-01404"
    And the workspace file "acme/collections/billing/orders.yaml" should be unchanged

  @delete @conflict
  Scenario: Deleting an alias from a file that changed on disk is refused
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{orders}}" }
      """
    And I save the response field "revision" as "revision"
    Given the workspace file "acme/collections/billing/orders.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{orders}}
      name: Orders
      method: GET
      url: https://api.acme.test/v1/orders
      aliases:
        - id: {{mx}}
          name: mx-renamed-by-git
          method: GET
          url: https://api.acme.test/v2/orders
      """
    When I call "AliasService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "alias_id": "{{mx}}", "base_revision": "{{revision}}" }
      """
    Then the call should fail with code "REQ-02409"
    And the workspace file "acme/collections/billing/orders.yaml" should be unchanged

  @validation @errors
  Scenario Outline: Alias calls validate their payload - <method> <case>
    When I call "AliasService.<method>" with
      """
      <payload>
      """
    Then the call should fail with code "<code>"

    Examples:
      | method | case             | payload                                                                                                                  | code      |
      | Create | no workspace     | { "local_id": "", "request_id": "{{orders}}", "name": "canary" }                                                         | ALS-06400 |
      | Create | no request       | { "local_id": "{{Acme.local_id}}", "request_id": "", "name": "canary" }                                                  | ALS-06400 |
      | Save   | no request       | { "local_id": "{{Acme.local_id}}", "request_id": "", "alias": { "id": "{{mx}}", "name": "x", "request": {} } }           | ALS-07400 |
      | Save   | no alias id      | { "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "alias": { "id": "", "name": "x", "request": {} } }       | ALS-07400 |
      | Delete | no alias id      | { "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "alias_id": "" }                                          | ALS-08400 |
