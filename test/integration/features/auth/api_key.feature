#language: en
#utf-8

@all @xen_005 @auth @api_key
Feature: API key auth
  API key auth adds a key and value to the request, as a header or as a query parameter, as
  Postman does: a header with the same name (in any case) or a query parameter with the same key
  is replaced, an empty key and value send nothing, and folders and collections pass it on to the
  requests that inherit their auth.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url

  @header
  Scenario: The key is sent as a header
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "GET", "url": "{{api}}/things", "auth": { "type": "apikey", "apikey": { "key": "X-Api-Key", "value": "k3y", "in": "header" } } }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/things" should have the header "X-Api-Key" equal to "k3y"

  @header @replace
  Scenario: A header with the same name in another case is replaced
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/things",
          "headers": [ { "key": "x-api-key", "value": "old", "enabled": true }, { "key": "X-Other", "value": "kept", "enabled": true } ],
          "auth": { "type": "apikey", "apikey": { "key": "X-Api-Key", "value": "k3y", "in": "header" } }
        }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/things" should have the header "X-Api-Key" equal to "k3y"
    And the 1 "GET" request to "/things" should have the header "X-Other" equal to "kept"

  @query @replace
  Scenario: The key is sent as a query parameter replacing one with the same key
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/things?api_key=old&page=2",
          "params": [ { "key": "api_key", "value": "old", "enabled": true }, { "key": "page", "value": "2", "enabled": true } ],
          "auth": { "type": "apikey", "apikey": { "key": "api_key", "value": "k3y", "in": "query" } }
        }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/things" should have the raw query "page=2&api_key=k3y"
    And the 1 "GET" request to "/things" should not have the header "api_key"

  @empty
  Scenario Outline: An empty key and value send nothing in the <in>
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "GET", "url": "{{api}}/things", "auth": { "type": "apikey", "apikey": { "key": "", "value": "", "in": "<in>" } } }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/things" should have the raw query ""
    And the response field "error" should not exist

    Examples:
      | in     |
      | header |
      | query  |

  @inheritance @variables
  Scenario: A request inherits the API key of its collection, with variables resolved
    Given the "billing" variable is set to "01JF0000000000000000000001"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JF0000000000000000000000
      name: Acme
      order:
        - billing
      variables:
        - key: billing_key
          value: from-workspace
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      auth:
        type: apikey
        apikey:
          key: X-Billing-Key
          value: "{{billing_key}}"
          in: header
      """
    And the workspace file "acme/collections/billing/invoices.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JF0000000000000000000002
      name: Invoices
      method: GET
      url: "{{api}}/invoices"
      auth:
        type: inherit
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "id": "01JF0000000000000000000002", "name": "Invoices", "method": "GET", "url": "{{api}}/invoices", "auth": { "type": "inherit" } }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/invoices" should have the header "X-Billing-Key" equal to "from-workspace"

  @persistence
  Scenario: The API key settings are saved in the request file
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JF0000000000000000000001
      name: Billing
      """
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JF0000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JF0000000000000000000003
      name: Ping
      method: GET
      url: https://api.example.test/ping
      """
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JF0000000000000000000003" }
      """
    And I save the response field "revision" as "revision"
    And I call "RequestService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "base_revision": "{{revision}}",
        "request": { "id": "01JF0000000000000000000003", "name": "Ping", "method": "GET", "url": "https://api.example.test/ping", "auth": { "type": "apikey", "apikey": { "key": "X-Key", "value": "{{key}}", "in": "query" } } }
      }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "auth.apikey.key" equal to "X-Key"
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "auth.apikey.value" equal to "{{key}}"
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "auth.apikey.in" equal to "query"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JF0000000000000000000003" }
      """
    Then the response should contain the field "auth.apikey.in" equal to "query"
