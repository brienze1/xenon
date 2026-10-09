#language: en
#utf-8

@all @xen_001 @history
Feature: Send history per workspace
  Every send - successful or not, from a saved request or an unsaved tab - is appended to the
  workspace's history with the request exactly as the user wrote it (variables unresolved) and a
  summary of the result. History lives in the local app data under
  workspaces/<localId>/history.jsonl, one JSON object per line, never in the workspace folder.
  It keeps the 1,000 newest entries; the file is compacted to those once it grows past 1,200
  lines.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the current time is "2026-10-09T10:00:00Z"

  @success
  Scenario: A successful send of an unsaved request is recorded as written
    Given the 1 "POST" request to "/users" returns status 201 with the following response
      """
      {"id":"u-1"}
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "name": "Untitled Request",
          "method": "post",
          "url": "{{api}}/users?team={{team}}",
          "params": [{ "key": "team", "value": "{{team}}", "enabled": true, "description": "" }],
          "headers": [{ "key": "X-Trace", "value": "{{trace}}", "enabled": true, "description": "trace id" }],
          "auth": { "type": "bearer", "bearer": { "token": "{{token}}" } },
          "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"name\":\"{{name}}\"}" } },
          "settings": { "timeout_ms": 5000, "follow_redirects": false, "verify_ssl": true, "encode_url": true }
        }
      }
      """
    Then the call should succeed
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 1
    And the response should contain the field "0.method" equal to "POST"
    And the response should contain the field "0.url" equal to "{{api}}/users?team={{team}}"
    And the response should contain the field "0.status_code" equal to "201"
    And the response should contain the field "0.size_bytes" equal to "12"
    And the response should contain the field "0.sent_at" matching "^2026-10-09T10:00:"
    And the response should contain the field "0.total_ms" matching "^[0-9]+$"
    And the response field "0.error_kind" should not exist
    When I save the response field "0.id" as "entry"
    And I call "HistoryService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{entry}}" }
      """
    Then the call should succeed
    And the response should contain the field "entry.id" equal to "{{entry}}"
    And the response should contain the field "entry.status_code" equal to "201"
    And the response should contain the field "request.name" equal to "Untitled Request"
    And the response should contain the field "request.method" equal to "POST"
    And the response should contain the field "request.url" equal to "{{api}}/users?team={{team}}"
    And the response should contain the field "request.params.0.value" equal to "{{team}}"
    And the response should contain the field "request.headers.0.key" equal to "X-Trace"
    And the response should contain the field "request.headers.0.value" equal to "{{trace}}"
    And the response should contain the field "request.headers.0.description" equal to "trace id"
    And the response should contain the field "request.auth.type" equal to "bearer"
    And the response should contain the field "request.auth.bearer.token" equal to "{{token}}"
    And the response should contain the field "request.body.mode" equal to "raw"
    And the response should contain the field "request.body.raw.content" equal to
      """
      {"name":"{{name}}"}
      """
    And the response should contain the field "request.settings.timeout_ms" equal to "5000"
    And the response should contain the field "request.settings.follow_redirects" equal to "false"
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" should have 1 lines
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" line 1 should contain the field "version" equal to "1"
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" line 1 should contain the field "id" equal to "{{entry}}"
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" line 1 should contain the field "request.auth.bearer.token" equal to "{{token}}"
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" line 1 should contain the field "result.status_code" equal to "201"
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" line 1 should contain the field "result.status_text" equal to "Created"
    And the app data path "workspaces/{{Acme.local_id}}/history.jsonl" should have the permissions "0600"
    And the workspace folder "acme" should contain the entries ".gitignore, collections, environments, workspace.yaml"

  @error
  Scenario Outline: A send that fails is recorded with its error - <case>
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "<url>" } }
      """
    Then the call should succeed
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.url" equal to "<url>"
    And the response should contain the field "0.error_kind" equal to "<kind>"
    And the response should contain the field "0.size_bytes" equal to "0"
    And the response field "0.status_code" should not exist
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" line 1 should contain the field "result.error_kind" equal to "<kind>"
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" line 1 should contain the field "result.error_message" equal to "<message>"

    Examples:
      | case        | url                                    | kind        | message                                          |
      | empty url   |                                        | invalid_url | Enter a URL to send the request                  |
      | unknown dns | http://xenon-unresolvable.invalid/ping | dns         | Could not resolve host xenon-unresolvable.invalid |

  @saved
  Scenario: A send of a saved request remembers which request it came from
    Given the workspace file "acme/workspace.yaml" contains
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
      id: 01JF0000000000000000000001
      name: Billing
      order:
        - invoices.yaml
      """
    And the workspace file "acme/collections/billing/invoices.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JF0000000000000000000002
      name: List invoices
      method: GET
      url: {{api}}/invoices
      """
    And the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      []
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "01JF0000000000000000000002", "name": "List invoices", "method": "GET", "url": "{{api}}/invoices" } }
      """
    And I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    And I save the response field "0.id" as "entry"
    And I call "HistoryService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{entry}}" }
      """
    Then the call should succeed
    And the response should contain the field "entry.source_request_id" equal to "01JF0000000000000000000002"
    And the response should contain the field "request.id" equal to "01JF0000000000000000000002"
    And the response should contain the field "request.name" equal to "List invoices"
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" line 1 should contain the field "source_request_id" equal to "01JF0000000000000000000002"

  @order @paging
  Scenario: History is listed newest first and pages with a cursor
    When I call "RequestService.Send" 3 times cycling through the payloads
      """
      [
        { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/alpha" } },
        { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/beta" } },
        { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "DELETE", "url": "{{api}}/gamma" } }
      ]
      """
    And I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 3
    And the response should contain the field "0.url" equal to "{{api}}/gamma"
    And the response should contain the field "1.url" equal to "{{api}}/beta"
    And the response should contain the field "2.url" equal to "{{api}}/alpha"
    And the response should contain the field "2.status_code" equal to "200"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "limit": 2 }
      """
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "1.url" equal to "{{api}}/beta"
    When I save the response field "1.id" as "cursor"
    And I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "limit": 2, "before_id": "{{cursor}}" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.url" equal to "{{api}}/alpha"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "before_id": "01JZZZZZZZZZZZZZZZZZZZZZZZ" }
      """
    Then the response should contain the field "" array length equal to 0

  @search
  Scenario Outline: History is searchable by URL or method, ignoring case - <case>
    When I call "RequestService.Send" 3 times cycling through the payloads
      """
      [
        { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/Orders" } },
        { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/orders/7/items" } },
        { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "purge", "url": "{{api}}/cache" } }
      ]
      """
    And I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "query": "<query>" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to <count>

    Examples:
      | case               | query     | count |
      | url, other case    | ORDERS    | 2     |
      | url segment        | /items    | 1     |
      | method             | post      | 1     |
      | custom method      | PURGE     | 1     |
      | blank query        |           | 3     |
      | nothing matches    | invoices  | 0     |

  @clear
  Scenario: Clearing empties the workspace's history
    When I call "RequestService.Send" 2 times cycling through the payloads
      """
      [{ "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/ping" } }]
      """
    And I call "HistoryService.Clear" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" should have 0 lines
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/after" } }
      """
    And I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.url" equal to "{{api}}/after"

  @clear
  Scenario: Clearing a workspace that never sent anything succeeds
    When I call "HistoryService.Clear" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 0

  @cap
  Scenario: Only the 1,000 newest entries are kept and the file is compacted past 1,200 lines
    When I call "RequestService.Send" 1200 times cycling through the payloads
      """
      [{ "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "" } }]
      """
    Then the app data file "workspaces/{{Acme.local_id}}/history.jsonl" should have 1200 lines
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "limit": 1000 }
      """
    Then the response should contain the field "" array length equal to 1000
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/newest" } }
      """
    Then the app data file "workspaces/{{Acme.local_id}}/history.jsonl" should have 1000 lines
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "limit": 1000 }
      """
    Then the response should contain the field "" array length equal to 1000
    And the response should contain the field "0.url" equal to "{{api}}/newest"
    And the response should contain the field "1.url" equal to ""

  @stored
  Scenario: History written earlier is read back and unreadable lines are skipped
    Given the app data file "workspaces/{{Acme.local_id}}/history.jsonl" contains
      """
      {"version":1,"id":"01JF1000000000000000000001","sent_at":"2026-10-08T09:00:00Z","source_request_id":"","request":{"name":"Old","method":"GET","url":"https://old.test/a"},"result":{"status_code":200,"status_text":"OK","error_kind":"","error_message":"","total_ms":12,"size_bytes":34}}
      not json at all
      {"version":1,"id":"01JF1000000000000000000002","sent_at":"2026-10-08T09:05:00Z","source_request_id":"","request":{"name":"Older","method":"PUT","url":"https://old.test/b"},"result":{"status_code":0,"status_text":"","error_kind":"timeout","error_message":"Timed out after 10 ms","total_ms":10,"size_bytes":0}}
      """
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 2
    And the response should contain the field "0.id" equal to "01JF1000000000000000000002"
    And the response should contain the field "0.method" equal to "PUT"
    And the response should contain the field "0.error_kind" equal to "timeout"
    And the response should contain the field "0.total_ms" equal to "10"
    And the response should contain the field "0.sent_at" equal to "2026-10-08T09:05:00Z"
    And the response should contain the field "1.status_code" equal to "200"
    And the response should contain the field "1.size_bytes" equal to "34"
    When I call "HistoryService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JF1000000000000000000001" }
      """
    Then the call should succeed
    And the response should contain the field "request.url" equal to "https://old.test/a"
    And the response should contain the field "request.method" equal to "GET"
    And the response should contain the field "entry.source_request_id" equal to ""
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "" } }
      """
    Then the app data file "workspaces/{{Acme.local_id}}/history.jsonl" should have 4 lines

  @validation
  Scenario Outline: Invalid history requests are rejected - <case>
    When I call "HistoryService.<method>" with
      """
      <payload>
      """
    Then the call should fail with code "<code>"

    Examples:
      | case               | method | payload                                                                     | code      |
      | list without id    | List   | { "local_id": " " }                                                         | HIS-00400 |
      | negative limit     | List   | { "local_id": "{{Acme.local_id}}", "limit": -1 }                            | HIS-00400 |
      | limit over the cap | List   | { "local_id": "{{Acme.local_id}}", "limit": 1001 }                          | HIS-00400 |
      | get without id     | Get    | { "local_id": "{{Acme.local_id}}", "id": "" }                               | HIS-01400 |
      | clear without id   | Clear  | { "local_id": "" }                                                          | HIS-02400 |
      | unknown entry      | Get    | { "local_id": "{{Acme.local_id}}", "id": "01JF9999999999999999999999" }     | HIS-03404 |
      | unknown list       | List   | { "local_id": "01JA2Z5V3B6C7D8E9F0G1H2J3K" }                                | WKS-07404 |
      | unknown get        | Get    | { "local_id": "01JA2Z5V3B6C7D8E9F0G1H2J3K", "id": "01JF9999999999999999999999" } | WKS-07404 |
      | unknown clear      | Clear  | { "local_id": "01JA2Z5V3B6C7D8E9F0G1H2J3K" }                                | WKS-07404 |
