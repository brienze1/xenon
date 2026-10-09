#language: en
#utf-8

@all @xen_001 @tree @settings_inheritance
Feature: Collection and folder settings are inherited by the requests below them
  Collections and folders have their own Authorization and headers. A request whose auth is
  "Inherit from parent" uses the nearest ancestor that defines auth (a collection's "inherit"
  means no auth), and ancestor headers are sent unless a nearer level defines the same name.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JB0000000000000000000001"
    And the "invoices" variable is set to "01JB0000000000000000000002"
    And the "create" variable is set to "01JB0000000000000000000003"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JB0000000000000000000000
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
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      """
    And the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      auth:
        type: inherit
      """

  @settings
  Scenario: Collection settings are read and written
    When I call "TreeService.SaveSettings" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "id": "{{billing}}",
        "description": "All billing calls",
        "auth": { "type": "bearer", "bearer": { "token": "collection-token" } },
        "headers": [ { "key": "X-Tenant", "value": "acme", "description": "tenant" } ]
      }
      """
    Then the call should succeed
    And the response should contain the field "id" equal to "{{billing}}"
    And the response should contain the field "kind" equal to "collection"
    And the response should contain the field "name" equal to "Billing"
    And the response should contain the field "auth.type" equal to "bearer"
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "auth.bearer.token" equal to "collection-token"
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "headers.0.key" equal to "X-Tenant"
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "description" equal to "All billing calls"
    When I call "TreeService.GetSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}" }
      """
    Then the call should succeed
    And the response should contain the field "description" equal to "All billing calls"
    And the response should contain the field "auth.bearer.token" equal to "collection-token"
    And the response should contain the field "headers.0.value" equal to "acme"
    And the response should contain the field "headers.0.description" equal to "tenant"

  @settings
  Scenario: Requests have no item settings
    When I call "TreeService.GetSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{create}}" }
      """
    Then the call should fail with code "TRE-03404"

  @auth
  Scenario: A request that inherits uses the collection's Bearer token
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      auth:
        type: bearer
        bearer:
          token: collection-token
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices", "auth": { "type": "inherit" } }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/invoices" should have the header "Authorization" equal to "Bearer collection-token"

  @auth
  Scenario: The nearest folder overrides the collection
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      auth:
        type: bearer
        bearer:
          token: collection-token
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      auth:
        type: basic
        basic:
          username: user
          password: pass
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices", "auth": { "type": "inherit" } }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/invoices" should have the header "Authorization" equal to "Basic dXNlcjpwYXNz"

  @auth
  Scenario: A folder set to "No auth" stops the collection's auth
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      auth:
        type: bearer
        bearer:
          token: collection-token
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      auth:
        type: none
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices", "auth": { "type": "inherit" } }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/invoices" should not have the header "Authorization"

  @auth
  Scenario: A collection set to inherit means no auth
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices", "auth": { "type": "inherit" } }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/invoices" should not have the header "Authorization"

  @auth
  Scenario: The request's own auth wins over its ancestors
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      auth:
        type: bearer
        bearer:
          token: collection-token
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices", "auth": { "type": "bearer", "bearer": { "token": "mine" } } }
      }
      """
    Then the 1 "GET" request to "/invoices" should have the header "Authorization" equal to "Bearer mine"

  @auth
  Scenario: An unsaved request has no ancestors
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      auth:
        type: bearer
        bearer:
          token: collection-token
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "GET", "url": "{{api}}/invoices", "auth": { "type": "inherit" } }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/invoices" should not have the header "Authorization"

  @headers
  Scenario: Headers are overridden by name from the collection down to the request, ignoring case
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      headers:
        - key: X-Tenant
          value: acme
        - key: X-Env
          value: collection
        - key: X-Level
          value: collection
        - key: X-Disabled
          value: never
          enabled: false
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      headers:
        - key: x-env
          value: folder
        - key: X-Level
          value: folder
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "id": "{{create}}",
          "name": "Create",
          "method": "GET",
          "url": "{{api}}/invoices",
          "headers": [ { "key": "X-ENV", "value": "request" } ]
        }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/invoices" should have the header "X-Tenant" equal to "acme"
    And the 1 "GET" request to "/invoices" should have the header "X-Env" equal to "request"
    And the 1 "GET" request to "/invoices" should have the header "X-Level" equal to "folder"
    And the 1 "GET" request to "/invoices" should not have the header "X-Disabled"

  @auto_headers
  Scenario: The auto header preview shows the inherited Authorization
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      auth:
        type: bearer
        bearer:
          token: collection-token
      """
    When I call "RequestService.AutoHeaders" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices", "auth": { "type": "inherit" } }
      }
      """
    Then the call should succeed
    And the response field "" should contain an item with "key" equal to "Authorization" and "value" equal to "Bearer collection-token"
