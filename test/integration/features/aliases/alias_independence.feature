#language: en
#utf-8

@all @xen_016 @aliases @alias_independence
Feature: The Default and every alias of a request are saved independently
  An alias is a full copy of the request's sendable content, not a set of overrides. Saving an alias
  changes only that alias; saving the Default never changes an alias; a later change to the Default
  never reaches an alias. Scripts follow the request rule: an alias saved without scripts edited in
  the tab keeps its stored scripts.

  Background:
    Given the workspace "Acme" exists
    And the "billing" variable is set to "01JR0000000000000000000001"
    And the "customers" variable is set to "01JR0000000000000000000002"
    And the "v2" variable is set to "01JR0000000000000000000010"
    And the "v3" variable is set to "01JR0000000000000000000011"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JR0000000000000000000000
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
        - customers.yaml
      """
    And the workspace file "acme/collections/billing/customers.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{customers}}
      name: Customers
      method: GET
      url: https://api.acme.test/v1/customers
      headers:
        - key: Accept
          value: application/json
          enabled: true
          description: ""
      scripts:
        preRequest: console.log("default pre")
        postResponse: ""
      aliases:
        - id: {{v2}}
          name: v2
          method: GET
          url: https://api.acme.test/v2/customers
          headers:
            - key: Accept
              value: application/json
              enabled: true
              description: ""
          scripts:
            preRequest: console.log("v2 pre")
            postResponse: console.log("v2 post")
          inheritScripts: true
        - id: {{v3}}
          name: v3
          method: GET
          url: https://api.acme.test/v3/customers
          headers:
            - key: Accept
              value: application/json
              enabled: true
              description: ""
      """

  @alias_save
  Scenario: Saving an alias that changes every sendable field leaves the Default and the other alias unchanged
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{customers}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}", "request_id": "{{customers}}", "base_revision": "{{revision}}", "force": false,
        "alias": {
          "id": "{{v2}}", "name": "v2",
          "request": {
            "method": "PUT",
            "url": "https://api.acme.test/v2/customers/:customerId?expand=orders",
            "params": [ { "key": "expand", "value": "orders", "enabled": true } ],
            "path_variables": [ { "key": "customerId", "value": "42" } ],
            "headers": [ { "key": "X-Canary", "value": "true", "enabled": true } ],
            "auto_headers": { "disabled": [ "User-Agent" ] },
            "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"tier\":\"gold\"}" } },
            "auth": { "type": "bearer", "bearer": { "token": "{{v2_token}}" } },
            "scripts": { "pre_request": "console.log(\"v2 edited\")", "post_response": "" },
            "inherit_scripts": false,
            "settings": { "timeout_ms": 1500, "follow_redirects": false, "verify_ssl": false, "encode_url": false, "disable_cookie_jar": true }
          }
        }
      }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.method" equal to "PUT"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.url" equal to "https://api.acme.test/v2/customers/:customerId?expand=orders"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.params.0.key" equal to "expand"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.pathVariables.0.value" equal to "42"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.headers.0.key" equal to "X-Canary"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.headers" array length equal to 1
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.autoHeaders.disabled.0" equal to "User-Agent"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.body.raw.content" equal to
      """
      {"tier":"gold"}
      """
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.auth.type" equal to "bearer"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.auth.bearer.token" equal to "{{v2_token}}"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.scripts.preRequest" equal to
      """
      console.log("v2 edited")
      """
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.inheritScripts" equal to "false"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.settings.timeoutMs" equal to "1500"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.settings.followRedirects" equal to "false"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.settings.verifySsl" equal to "false"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.settings.encodeUrl" equal to "false"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.settings.disableCookieJar" equal to "true"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "method" equal to "GET"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "url" equal to "https://api.acme.test/v1/customers"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "headers.0.key" equal to "Accept"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "scripts.preRequest" equal to
      """
      console.log("default pre")
      """
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "auth.type" equal to "inherit"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.1.url" equal to "https://api.acme.test/v3/customers"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.1.headers.0.key" equal to "Accept"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{customers}}" }
      """
    Then the response should contain the field "method" equal to "GET"
    And the response should contain the field "aliases.0.request.method" equal to "PUT"
    And the response should contain the field "aliases.0.request.settings.timeout_ms" equal to "1500"
    And the response should contain the field "aliases.1.request.method" equal to "GET"

  @default_save
  Scenario: Saving the Default after adding a header does not reach the aliases
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{customers}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "RequestService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}", "base_revision": "{{revision}}", "force": false,
        "request": {
          "id": "{{customers}}", "name": "Customers", "method": "GET", "url": "https://api.acme.test/v1/customers",
          "headers": [
            { "key": "Accept", "value": "application/json", "enabled": true },
            { "key": "Accept-Language", "value": "pt-BR", "enabled": true }
          ]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "headers" array length equal to 2
    And the response should contain the field "aliases" array length equal to 2
    And the response should contain the field "aliases.0.request.headers" array length equal to 1
    And the response should contain the field "aliases.1.request.headers" array length equal to 1
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "headers" array length equal to 2
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.headers" array length equal to 1
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.1.headers" array length equal to 1
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.url" equal to "https://api.acme.test/v2/customers"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{customers}}" }
      """
    Then the response should contain the field "aliases.0.request.url" equal to "https://api.acme.test/v2/customers"
    And the response should contain the field "aliases.0.request.scripts.pre_request" equal to
      """
      console.log("v2 pre")
      """

  @alias_save @scripts
  Scenario: Saving an alias without scripts edited in the tab keeps its stored scripts
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{customers}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}", "request_id": "{{customers}}", "base_revision": "{{revision}}", "force": false,
        "alias": { "id": "{{v2}}", "name": "v2", "request": { "method": "GET", "url": "https://api.acme.test/v2/customers?page=2" } }
      }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.url" equal to "https://api.acme.test/v2/customers?page=2"
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.scripts.preRequest" equal to
      """
      console.log("v2 pre")
      """
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "aliases.0.scripts.postResponse" equal to
      """
      console.log("v2 post")
      """
    And the workspace file "acme/collections/billing/customers.yaml" should contain the field "scripts.preRequest" equal to
      """
      console.log("default pre")
      """
