#language: en
#utf-8

@all @xen_008 @mcp @mcp_browse_and_send
Feature: AI clients browse the workspace and send requests
  The read tools answer compactly: lists return ids and names (plus kind, method and child counts), one level at a
  time unless the full depth is asked for, 100 rows per page with a cursor. "send" sends a saved request or an ad-hoc
  one with the active environment exactly like the UI does (scripts, history, cookies) and returns status, headers,
  timing, size, test results and the body. Text bodies are capped at 16 KB with "truncated" and the total size, and
  "response body" reads further with offset/length; binary bodies are never inlined and can be saved to a path.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JX0000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JX0000000000000000000001
      name: Billing
      order:
        - invoices
        - ping.yaml
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: 01JX0000000000000000000002
      name: Invoices
      order:
        - list-invoices.yaml
      """
    And the workspace file "acme/collections/billing/invoices/list-invoices.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JX0000000000000000000003
      name: List invoices
      method: GET
      url: {{api}}/invoices
      headers:
        - key: Accept
          value: application/json
          enabled: true
          description: ""
      scripts:
        postResponse: |
          pm.test("status is 200", function () { pm.response.to.have.status(200); });
      """
    And the workspace file "acme/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JX0000000000000000000004
      name: Ping
      method: HEAD
      url: {{api}}/ping
      """

  @workspace
  Scenario: The workspace tool lists workspaces and shows the active one
    When I call the MCP tool "workspace" with
      """
      { "action": "list" }
      """
    Then the call should succeed
    And the response field "items" should contain an item with "id" equal to "{{Acme.local_id}}" and "name" equal to "Acme"
    And the response field "items" should contain an item with "name" equal to "Acme" and "active" equal to "true"
    When I call the MCP tool "workspace" with
      """
      { "action": "active" }
      """
    Then the call should succeed
    And the response should contain the field "id" equal to "{{Acme.local_id}}"
    And the response should contain the field "name" equal to "Acme"
    And the response should contain the field "path" equal to "{{Acme.path}}"

  @tree_list
  Scenario: Listing the tree returns only the top level by default
    When I call the MCP tool "tree" with
      """
      { "action": "list" }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 1
    And the response should contain the field "items.0.id" equal to "01JX0000000000000000000001"
    And the response should contain the field "items.0.kind" equal to "collection"
    And the response should contain the field "items.0.name" equal to "Billing"
    And the response should contain the field "items.0.children" equal to "2"
    And the response field "items.0.items" should not exist
    And the response field "next" should not exist

  @tree_list
  Scenario: Listing one parent returns its direct children with their kind and method
    When I call the MCP tool "tree" with
      """
      { "action": "list", "parent_id": "01JX0000000000000000000001" }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 2
    And the response should contain the field "items.0.id" equal to "01JX0000000000000000000002"
    And the response should contain the field "items.0.kind" equal to "folder"
    And the response should contain the field "items.0.name" equal to "Invoices"
    And the response should contain the field "items.0.children" equal to "1"
    And the response should contain the field "items.1.id" equal to "01JX0000000000000000000004"
    And the response should contain the field "items.1.kind" equal to "request"
    And the response should contain the field "items.1.method" equal to "HEAD"
    And the response field "items.1.url" should not exist

  @tree_list
  Scenario: The full tree is returned only when asked for
    When I call the MCP tool "tree" with
      """
      { "action": "list", "depth": "all" }
      """
    Then the call should succeed
    And the response should contain the field "items.0.items.0.name" equal to "Invoices"
    And the response should contain the field "items.0.items.0.items.0.name" equal to "List invoices"
    And the response should contain the field "items.0.items.0.items.0.method" equal to "GET"
    And the response should contain the field "items.0.items.1.name" equal to "Ping"

  @tree_get
  Scenario: Reading a request returns it whole, with its path
    When I call the MCP tool "tree" with
      """
      { "action": "get", "id": "01JX0000000000000000000003" }
      """
    Then the call should succeed
    And the response should contain the field "id" equal to "01JX0000000000000000000003"
    And the response should contain the field "name" equal to "List invoices"
    And the response should contain the field "method" equal to "GET"
    And the response should contain the field "url" equal to "{{api}}/invoices"
    And the response should contain the field "headers.0.key" equal to "Accept"
    And the response should contain the field "path" equal to "Billing / Invoices / List invoices"
    And the response should contain the field "scripts.post_response" containing "status is 200"

  @tree_get
  Scenario: Reading a folder returns its summary and child count
    When I call the MCP tool "tree" with
      """
      { "action": "get", "id": "01JX0000000000000000000002" }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "folder"
    And the response should contain the field "name" equal to "Invoices"
    And the response should contain the field "children" equal to "1"
    And the response should contain the field "path" equal to "Billing / Invoices"

  @tree_get
  Scenario: An unknown id is not found
    When I call the MCP tool "tree" with
      """
      { "action": "get", "id": "01JX9999999999999999999999" }
      """
    Then the call should fail with code "MCP-08404"

  @tree_search
  Scenario: Searching returns compact matches with their path
    When I call the MCP tool "tree" with
      """
      { "action": "search", "query": "invoice" }
      """
    Then the call should succeed
    And the response field "items" should contain an item with "id" equal to "01JX0000000000000000000003" and "path" equal to "Billing / Invoices / List invoices"
    And the response field "items" should contain an item with "id" equal to "01JX0000000000000000000002" and "kind" equal to "folder"
    And the response field "items" should contain 0 items with "name" equal to "Ping"

  @send_saved
  Scenario: Sending a saved request returns the response and records an AI send in History
    Given the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      [{"id":"inv-1"}]
      """
    And the MCP client is named "Claude Code"
    When I call the MCP tool "send" with
      """
      { "request_id": "01JX0000000000000000000003" }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "200"
    And the response should contain the field "status_text" equal to "OK"
    And the response should contain the field "content_type" equal to "application/json"
    And the response should contain the field "size" equal to "16"
    And the response should contain the field "body" equal to
      """
      [{"id":"inv-1"}]
      """
    And the response should contain the field "ms" matching "^[0-9]+$"
    And the response should contain the field "tests.0.name" equal to "status is 200"
    And the response should contain the field "tests.0.passed" equal to "true"
    And the response field "truncated" should not exist
    And the 1 "GET" request to "/invoices" should have the header "Accept" equal to "application/json"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.kind" equal to "send"
    And the response should contain the field "0.source" equal to "ai"
    And the response should contain the field "0.client" equal to "Claude Code"
    And the response should contain the field "0.url" equal to "{{api}}/invoices"
    And the response should contain the field "0.status_code" equal to "200"

  @send_ad_hoc
  Scenario: An ad-hoc request is sent with the active environment and nothing is saved to the tree
    Given the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JX0000000000000000000010
      name: dev
      variables:
        - key: base_url
          value: "{{api}}"
          enabled: true
          description: ""
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JX0000000000000000000010" }
      """
    And the 1 "POST" request to "/orders" returns status 201 with the following response
      """
      {"id":"o-1"}
      """
    When I call the MCP tool "send" with
      """
      { "request": { "method": "POST", "url": "{{base_url}}/orders", "headers": [ { "key": "Content-Type", "value": "application/json" } ], "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"qty\":2}" } } } }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "201"
    And the 1 "POST" request to "/orders" should have the body field "qty" equal to "2"
    And the workspace folder "acme/collections/billing" should contain the entries "collection.yaml, invoices, ping.yaml"
    And the workspace folder "acme/collections" should contain the entries "billing"

  @four_calls
  Scenario: Browsing to a saved request, sending it and reading its body takes 4 calls
    Given the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      [{"id":"inv-1"}]
      """
    When I call the MCP tool "tree" with
      """
      { "action": "list" }
      """
    And I save the response field "items.0.id" as "collection"
    And I call the MCP tool "tree" with
      """
      { "action": "list", "parent_id": "{{collection}}", "depth": 2 }
      """
    And I save the response field "items.0.items.0.id" as "request"
    And I call the MCP tool "send" with
      """
      { "request_id": "{{request}}" }
      """
    And I call the MCP tool "response" with
      """
      { "action": "body" }
      """
    Then the call should succeed
    And the response should contain the field "text" equal to
      """
      [{"id":"inv-1"}]
      """

  @truncation
  Scenario: A large text body is cut at 16 KB and read further with offset and length
    Given the 1 "GET" request to "/big" returns status 200 with content type "text/plain" and a generated body of 40000 bytes
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{api}}/big" } }
      """
    Then the call should succeed
    And the response should contain the field "truncated" equal to "true"
    And the response should contain the field "size" equal to "40000"
    And the response field "body" should have 16384 characters
    When I call the MCP tool "response" with
      """
      { "action": "body", "offset": 16384, "length": 16384 }
      """
    Then the call should succeed
    And the response should contain the field "offset" equal to "16384"
    And the response should contain the field "size" equal to "40000"
    And the response should contain the field "truncated" equal to "true"
    And the response field "text" should have 16384 characters
    When I call the MCP tool "response" with
      """
      { "action": "body", "offset": 32768 }
      """
    Then the response field "text" should have 7232 characters
    And the response field "truncated" should not exist

  @truncation
  Scenario: A body read is capped at 64 KB
    Given the 1 "GET" request to "/huge" returns status 200 with content type "text/plain" and a generated body of 100000 bytes
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{api}}/huge" } }
      """
    And I call the MCP tool "response" with
      """
      { "action": "body", "offset": 0, "length": 999999 }
      """
    Then the call should succeed
    And the response field "text" should have 65536 characters
    And the response should contain the field "truncated" equal to "true"

  @binary
  Scenario: A binary body is never inlined and can be saved to a path
    Given the 1 "GET" request to "/logo" returns status 200 with content type "application/octet-stream" and a generated body of 2048 bytes
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{api}}/logo" } }
      """
    Then the call should succeed
    And the response should contain the field "binary" equal to "true"
    And the response should contain the field "content_type" equal to "application/octet-stream"
    And the response should contain the field "size" equal to "2048"
    And the response field "body" should not exist
    When I call the MCP tool "response" with
      """
      { "action": "body" }
      """
    Then the response should contain the field "binary" equal to "true"
    And the response field "text" should not exist
    When I call the MCP tool "export" with
      """
      { "action": "response_body", "path": "{{temp_root}}/logo.bin" }
      """
    Then the call should succeed
    And the file "logo.bin" should be the generated body of 2048 bytes

  @paging
  Scenario: Lists are paged at 100 rows with a cursor
    Given the workspace "Big" exists
    And the workspace folder "big" holds 150 generated requests across 1 collections
    When I call the MCP tool "tree" with
      """
      { "action": "list", "parent_id": "01JG0000000000000000000001" }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 100
    And the response field "next" should exist
    When I save the response field "next" as "cursor"
    And I call the MCP tool "tree" with
      """
      { "action": "list", "parent_id": "01JG0000000000000000000001", "cursor": "{{cursor}}" }
      """
    Then the response should contain the field "items" array length equal to 50
    And the response should contain the field "items.0.name" equal to "Request 0100"
    And the response field "next" should not exist

  @performance
  Scenario: A large workspace answers fast and its top level stays small
    Given the workspace "Huge" exists
    And the workspace folder "huge" holds 1000 generated requests across 10 collections
    When I call the MCP tool "tree" with
      """
      { "action": "list" }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 10
    And the response should have at most 8000 characters
    When I call the MCP tool "tree" with
      """
      { "action": "search", "query": "Request 0999" }
      """
    Then the call should succeed
    And the last call should have taken less than 200 ms
    When I call the MCP tool "tree" with
      """
      { "action": "list", "parent_id": "01JG0000000000000000000001" }
      """
    Then the last call should have taken less than 200 ms

  @performance
  Scenario: A send to a target that answers at once completes within 100 ms
    Given the 1 "GET" request to "/fast" returns status 204 with the following response
      """
      """
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{api}}/fast" } }
      """
    And I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{api}}/fast" } }
      """
    Then the call should succeed
    And the last call should have taken less than 100 ms
