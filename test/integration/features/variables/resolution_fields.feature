#language: en
#utf-8

@all @xen_002 @variables @resolution_fields
Feature: Variables are resolved in every part of the request that is sent
  Go resolves "{{name}}" on a copy of the request: URL, query, path variable values, header keys
  and values, every body mode (raw, form-data keys/values/files, urlencoded, GraphQL query and
  variables, binary file path), auth fields, and the auth and headers inherited from the
  collection and folders. The stored request and its History entry keep the text as written.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JF0000000000000000000001"
    And the "invoices" variable is set to "01JF0000000000000000000002"
    And the "create" variable is set to "01JF0000000000000000000003"
    And the "dev" variable is set to "01JF0000000000000000000010"
    And the file "uploads/payload.bin" exists with content
      """
      binary payload
      """
    And the file "uploads/invoice.txt" exists with content
      """
      invoice file
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
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      auth:
        type: bearer
        bearer:
          token: "{{collection_token}}"
      headers:
        - key: "X-{{header_suffix}}"
          value: "{{team}}"
          enabled: true
          description: ""
      variables:
        - key: collection_token
          value: "{{token_prefix}}-collection"
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      headers:
        - key: X-Folder
          value: "{{folder_value}}"
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create invoice
      method: POST
      url: "{{base_url}}/invoices"
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: base_url
          value: "{{api}}"
          enabled: true
          description: ""
        - key: org
          value: acme-org
          enabled: true
          description: ""
        - key: query_key
          value: page
          enabled: true
          description: ""
        - key: query_value
          value: "2"
          enabled: true
          description: ""
        - key: header_key
          value: X-Trace
          enabled: true
          description: ""
        - key: header_value
          value: trace-42
          enabled: true
          description: ""
        - key: header_suffix
          value: Team
          enabled: true
          description: ""
        - key: team
          value: core
          enabled: true
          description: ""
        - key: folder_value
          value: from-folder
          enabled: true
          description: ""
        - key: token
          value: tok-123
          enabled: true
          description: ""
        - key: token_prefix
          value: inherited
          enabled: true
          description: ""
        - key: user
          value: luis
          enabled: true
          description: ""
        - key: password
          value: pa55word
          enabled: true
          description: ""
        - key: amount
          value: "1000"
          enabled: true
          description: ""
        - key: field_name
          value: customer
          enabled: true
          description: ""
        - key: customer
          value: Globex
          enabled: true
          description: ""
        - key: upload_dir
          value: "{{temp_root}}/uploads"
          enabled: true
          description: ""
        - key: limit
          value: "5"
          enabled: true
          description: ""
        - key: entity
          value: invoices
          enabled: true
          description: ""
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """

  @url @headers @auth @raw
  Scenario: URL, query, path variables, headers, Bearer token and raw body are resolved
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "POST",
          "url": "{{base_url}}/orgs/:org/invoices?{{query_key}}={{query_value}}",
          "params": [ { "key": "{{query_key}}", "value": "{{query_value}}" } ],
          "path_variables": [ { "key": "org", "value": "{{org}}" } ],
          "headers": [ { "key": "{{header_key}}", "value": "{{header_value}}" } ],
          "auth": { "type": "bearer", "bearer": { "token": "{{token}}" } },
          "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"amount\":{{amount}},\"customer\":\"{{customer}}\"}" } }
        }
      }
      """
    Then the call should succeed
    And the "POST" request to "/orgs/acme-org/invoices" should have been received 1 times
    And the 1 "POST" request to "/orgs/acme-org/invoices" should have the query "page" equal to "2"
    And the 1 "POST" request to "/orgs/acme-org/invoices" should have the header "X-Trace" equal to "trace-42"
    And the 1 "POST" request to "/orgs/acme-org/invoices" should have the header "Authorization" equal to "Bearer tok-123"
    And the 1 "POST" request to "/orgs/acme-org/invoices" should have the body equal to
      """
      {"amount":1000,"customer":"Globex"}
      """
    And the 1 "POST" request to "/orgs/acme-org/invoices" should have the header "Content-Length" equal to "35"

  @auth
  Scenario: Basic auth username and password are resolved before encoding
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{base_url}}/basic", "auth": { "type": "basic", "basic": { "username": "{{user}}", "password": "{{password}}" } } } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/basic" should have the header "Authorization" equal to "Basic bHVpczpwYTU1d29yZA=="

  @inherited
  Scenario: Inherited collection auth and collection/folder headers are resolved, nested values included
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "method": "POST", "url": "{{base_url}}/invoices", "auth": { "type": "inherit" } } }
      """
    Then the call should succeed
    And the 1 "POST" request to "/invoices" should have the header "Authorization" equal to "Bearer inherited-collection"
    And the 1 "POST" request to "/invoices" should have the header "X-Team" equal to "core"
    And the 1 "POST" request to "/invoices" should have the header "X-Folder" equal to "from-folder"

  @form_data
  Scenario: Form-data keys, text values and file paths are resolved
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "POST",
          "url": "{{base_url}}/upload",
          "body": {
            "mode": "formdata",
            "formdata": [
              { "key": "{{field_name}}", "value": "{{customer}}", "type": "text" },
              { "key": "file", "type": "file", "src": "{{upload_dir}}/invoice.txt" }
            ]
          }
        }
      }
      """
    Then the call should succeed
    And the 1 "POST" request to "/upload" should have the form field "customer" equal to "Globex"
    And the 1 "POST" request to "/upload" should have the form file "file" named "invoice.txt" with content "invoice file"

  @urlencoded
  Scenario: Urlencoded keys and values are resolved
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "POST",
          "url": "{{base_url}}/form",
          "body": { "mode": "urlencoded", "urlencoded": [ { "key": "{{field_name}}", "value": "{{customer}}" }, { "key": "amount", "value": "{{amount}}" } ] }
        }
      }
      """
    Then the call should succeed
    And the 1 "POST" request to "/form" should have the body equal to
      """
      customer=Globex&amount=1000
      """

  @graphql
  Scenario: The GraphQL query and variables are resolved
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "POST",
          "url": "{{base_url}}/graphql",
          "body": { "mode": "graphql", "graphql": { "query": "query { {{entity}} { id } }", "variables": "{ \"limit\": {{limit}} }" } }
        }
      }
      """
    Then the call should succeed
    And the 1 "POST" request to "/graphql" should have the body equal to
      """
      {"query":"query { invoices { id } }","variables":{"limit":5}}
      """

  @binary
  Scenario: The binary body file path is resolved before the file is read
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "PUT", "url": "{{base_url}}/blob", "body": { "mode": "binary", "binary": { "src": "{{upload_dir}}/payload.bin" } } } }
      """
    Then the call should succeed
    And the response field "error" should not exist
    And the 1 "PUT" request to "/blob" should have the body equal to
      """
      binary payload
      """

  @history
  Scenario: History and the stored request keep the unresolved text
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "id": "{{create}}",
          "method": "POST",
          "url": "{{base_url}}/invoices?{{query_key}}={{query_value}}",
          "headers": [ { "key": "{{header_key}}", "value": "{{header_value}}" } ],
          "auth": { "type": "bearer", "bearer": { "token": "{{token}}" } },
          "body": { "mode": "raw", "raw": { "language": "text", "content": "amount={{amount}}" } }
        }
      }
      """
    Then the call should succeed
    And the 1 "POST" request to "/invoices" should have the body equal to
      """
      amount=1000
      """
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.url" equal to "{{base_url}}/invoices?{{query_key}}={{query_value}}"
    When I save the response field "0.id" as "entry"
    And I call "HistoryService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{entry}}" }
      """
    Then the call should succeed
    And the response should contain the field "request.headers.0.key" equal to "{{header_key}}"
    And the response should contain the field "request.headers.0.value" equal to "{{header_value}}"
    And the response should contain the field "request.auth.bearer.token" equal to "{{token}}"
    And the response should contain the field "request.body.raw.content" equal to "amount={{amount}}"
    And the workspace file "acme/collections/billing/invoices/create.yaml" should contain the field "url" equal to "{{base_url}}/invoices"
