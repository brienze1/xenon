#language: en
#utf-8

@all @xen_008 @history @ai_activity
Feature: Everything an AI client does is recorded in History
  Every MCP tool call made while a workspace is active is recorded in that workspace's History. Sends become ordinary
  send entries with source "ai" and the client's name; every other call (reads included) becomes an action entry with
  the tool, action, target, outcome, duration and its arguments, with secret-carrying arguments redacted by name and
  the leak guard applied, capped at 2 KB. Human sends have source "user"; older entries without a source count as
  "user". History is filtered by source (all, user, ai) and Clear History clears only what the filter shows. Each
  source has its own cap (Settings → History, 1,000 by default, 100 to 20,000), oldest dropped first.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the MCP client is named "Claude Code"
    And the current time is "2026-10-13T10:00:00Z"

  @action_rows
  Scenario: A read, a write and a failing call each leave one action entry
    When I call the MCP tool "tree" with
      """
      { "action": "list" }
      """
    And I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "collection", "name": "Payments" }
      """
    And I save the response field "id" as "payments"
    And I call the MCP tool "tree" with
      """
      { "action": "explode" }
      """
    And I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 3
    And the response should contain the field "0.kind" equal to "action"
    And the response should contain the field "0.source" equal to "ai"
    And the response should contain the field "0.client" equal to "Claude Code"
    And the response should contain the field "0.activity.tool" equal to "tree"
    And the response should contain the field "0.activity.action" equal to "explode"
    And the response should contain the field "0.activity.outcome" equal to "error"
    And the response should contain the field "0.activity.code" equal to "MCP-07400"
    And the response should contain the field "1.activity.tool" equal to "tree_edit"
    And the response should contain the field "1.activity.action" equal to "create"
    And the response should contain the field "1.activity.outcome" equal to "ok"
    And the response should contain the field "1.activity.target.id" equal to "{{payments}}"
    And the response should contain the field "1.activity.target.name" equal to "Payments"
    And the response should contain the field "1.activity.duration_ms" matching "^[0-9]+$"
    And the response should contain the field "2.activity.tool" equal to "tree"
    And the response should contain the field "2.activity.action" equal to "list"
    And the response should contain the field "2.sent_at" matching "^2026-10-13T10:00:"
    And the response field "0.activity.args" should not exist
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" should have 3 lines
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" line 2 should contain the field "kind" equal to "action"
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" line 2 should contain the field "source" equal to "ai"
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" line 2 should contain the field "activity.tool" equal to "tree_edit"
    And the workspace folder "acme" should contain the entries ".gitignore, collections, environments, workspace.yaml"

  @action_detail
  Scenario: An action entry's detail shows its arguments, capped at 2 KB
    Given the "long" variable is set to "x" repeated 3000 times separated by ""
    When I call the MCP tool "tree" with
      """
      { "action": "search", "query": "{{long}}" }
      """
    And I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    And I save the response field "0.id" as "entry"
    And I call "HistoryService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{entry}}" }
      """
    Then the call should succeed
    And the response should contain the field "activity.args" matching "action.{1,4}search"
    And the response should contain the field "activity.args" matching "…$"
    And the response field "activity.args" should have 2049 characters

  @sends
  Scenario: AI sends are send entries tagged with the client, never duplicated as actions
    Given the 1 "GET" request to "/ping" returns status 200 with the following response
      """
      {}
      """
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{api}}/ping" } }
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/ping" } }
      """
    And I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "0.kind" equal to "send"
    And the response should contain the field "0.source" equal to "user"
    And the response field "0.client" should not exist
    And the response should contain the field "1.kind" equal to "send"
    And the response should contain the field "1.source" equal to "ai"
    And the response should contain the field "1.client" equal to "Claude Code"
    And the response should contain the field "1.status_code" equal to "200"

  @sends
  Scenario: A denied send never went out and is recorded as an action entry
    Given the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JT0000000000000000000011
      name: Prod
      confirmBeforeSend: true
      variables: []
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JT0000000000000000000011" }
      """
    When I start calling the MCP tool "send" in the background with
      """
      { "request": { "method": "GET", "url": "{{api}}/ping" } }
      """
    Then an "ai:approval-requested" event with the field "client" equal to "Claude Code" should have been emitted within 2000 ms
    When I call "AIAccessService.PendingApprovals"
    And I save the response field "0.id" as "approval"
    And I call "AIAccessService.AnswerApproval" with
      """
      { "id": "{{approval}}", "approved": false }
      """
    Then the background call should finish within 5000 ms
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.kind" equal to "action"
    And the response should contain the field "0.activity.tool" equal to "send"
    And the response should contain the field "0.activity.outcome" equal to "denied"
    And the "GET" request to "/ping" should have been received 0 times

  @sends
  Scenario: A timed-out send never went out and is recorded as an action entry
    Given the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JT0000000000000000000011
      name: Prod
      confirmBeforeSend: true
      variables: []
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JT0000000000000000000011" }
      """
    When I start calling the MCP tool "send" in the background with
      """
      { "request": { "method": "GET", "url": "{{api}}/ping" } }
      """
    Then an "ai:approval-requested" event with the field "client" equal to "Claude Code" should have been emitted within 2000 ms
    When 121 seconds pass
    Then the background call should finish within 5000 ms
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the response should contain the field "0.kind" equal to "action"
    And the response should contain the field "0.activity.outcome" equal to "timed_out"
    And the "GET" request to "/ping" should have been received 0 times

  @sends
  Scenario: A send with invalid arguments is recorded as an action entry with its error code
    When I call the MCP tool "send" with
      """
      { "request": { "method": 42 } }
      """
    Then the call should fail with code "MCP-07400"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.kind" equal to "action"
    And the response should contain the field "0.activity.outcome" equal to "error"
    And the response should contain the field "0.activity.code" equal to "MCP-07400"

  @filter
  Scenario: History is filtered by source and old entries count as user sends
    Given the app data file "workspaces/{{Acme.local_id}}/history.jsonl" contains
      """
      {"version":1,"id":"01JT0000000000000000000099","sent_at":"2026-10-01T09:00:00Z","request":{"name":"Old","method":"GET","url":"https://old.test"},"result":{"status_code":200,"status_text":"OK"}}
      """
    When I call the MCP tool "tree" with
      """
      { "action": "list" }
      """
    And I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "user" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.id" equal to "01JT0000000000000000000099"
    And the response should contain the field "0.source" equal to "user"
    And the response should contain the field "0.kind" equal to "send"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the response should contain the field "" array length equal to 1
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "all" }
      """
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "0.source" equal to "ai"

  @clear
  Scenario Outline: Clear History clears only the filtered source - <source>
    Given the 1 "GET" request to "/ping" returns status 200 with the following response
      """
      {}
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/ping" } }
      """
    And I call the MCP tool "tree" with
      """
      { "action": "list" }
      """
    When I call "HistoryService.Clear" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "<source>" }
      """
    Then the call should succeed
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "user" }
      """
    Then the response should contain the field "" array length equal to <user_left>
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the response should contain the field "" array length equal to <ai_left>

    Examples:
      | source | user_left | ai_left |
      | user   | 0         | 2       |
      | ai     | 1         | 0       |
      | all    | 0         | 0       |

  @clear
  Scenario: The AI clears History through the destructive delete tool
    Given I call the MCP tool "tree" with
      """
      { "action": "list" }
      """
    When I call the MCP tool "delete" with
      """
      { "target": "history", "source": "ai" }
      """
    Then the call should succeed
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.activity.tool" equal to "delete"

  @caps
  Scenario: Each source has its own cap and one source never pushes the other out
    Given I call "SettingsService.Save" with
      """
      { "history_user_cap": 10000, "history_ai_cap": 100 }
      """
    And the 1 "GET" request to "/ping" returns status 200 with the following response
      """
      {}
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/ping" } }
      """
    When I call the MCP tool "tree" 101 times with
      """
      { "action": "list" }
      """
    And I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the response should contain the field "" array length equal to 100
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "user" }
      """
    Then the response should contain the field "" array length equal to 1

  @caps
  Scenario: Lowering a cap shows the effect at once; raising it drops nothing
    Given I call the MCP tool "tree" 150 times with
      """
      { "action": "list" }
      """
    When I call "SettingsService.Save" with
      """
      { "history_ai_cap": 100 }
      """
    And I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the response should contain the field "" array length equal to 100
    When I call "SettingsService.Save" with
      """
      { "history_ai_cap": 1000 }
      """
    And I call the MCP tool "tree" with
      """
      { "action": "list" }
      """
    And I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the response should contain the field "" array length equal to 101

  @no_workspace
  Scenario: A call made with no active workspace is only logged
    Given the log output is captured
    And I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Acme.local_id}}", "move_to_trash": false }
      """
    When I call the MCP tool "workspace" with
      """
      { "action": "list" }
      """
    Then the call should succeed
    And the log output should contain "workspace"
    And the log output should contain "Claude Code"
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" should not exist

  @switch
  Scenario: A workspace switch is recorded in the workspace that becomes active
    Given the workspace "Globex" exists
    When I call the MCP tool "workspace_edit" with
      """
      { "action": "switch", "id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the response should contain the field "0.activity.tool" equal to "workspace_edit"
    And the response should contain the field "0.activity.action" equal to "switch"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Globex.local_id}}", "source": "ai" }
      """
    Then the response should contain the field "" array length equal to 0

  @event
  Scenario: Recording an entry tells the window
    Given the emitted events are forgotten
    When I call the MCP tool "tree" with
      """
      { "action": "list" }
      """
    Then a "history:changed" event with the field "local_id" equal to "{{Acme.local_id}}" should have been emitted within 1000 ms

  @mcp_history_tool
  Scenario: Through MCP, History lists compact rows and gives one entry's detail
    Given I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "collection", "name": "Payments" }
      """
    When I call the MCP tool "history" with
      """
      { "action": "list", "source": "ai" }
      """
    Then the call should succeed
    And the response field "items" should contain an item with "tool" equal to "tree_edit" and "action" equal to "create"
    And the response field "items" should contain an item with "tool" equal to "tree_edit" and "target" equal to "Payments"
    And the response field "items.0.args" should not exist
    When I save the response field "items.1.id" as "entry"
    And I call the MCP tool "history" with
      """
      { "action": "get", "id": "{{entry}}" }
      """
    Then the call should succeed
    And the response should contain the field "tool" equal to "tree_edit"
    And the response should contain the field "args" containing "Payments"

  @args_on_disk
  Scenario: Secret values in AI call arguments never reach History on disk, even when the call fails
    Given the keychain is unavailable
    And the "dev" variable is set to "01JT0000000000000000000010"
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables: []
      """
    And I call "VaultService.SetupPassphrase" with
      """
      { "passphrase": "correct horse", "confirmation": "correct horse" }
      """
    And I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "locked_one", "secret": true, "value": "first-value" }
      """
    And the log output is captured
    When the app is restarted
    And I call the MCP tool "env_edit" with
      """
      { "action": "add_var", "scope": "env:{{dev}}", "key": "new_secret", "secret": true, "value": "arg-secret-add-1" }
      """
    Then the call should fail with code "VLT-04409"
    When I call the MCP tool "env_edit" with
      """
      { "action": "set_current", "scope": "env:{{dev}}", "key": "locked_one", "value": "arg-secret-current-2" }
      """
    Then the call should fail with code "VLT-04409"
    When I call the MCP tool "import" with
      """
      { "action": "run", "paths": ["{{temp_root}}/missing.xenon.yaml"], "passphrase": "arg-passphrase-3" }
      """
    And I call the MCP tool "tree_edit" with
      """
      { "action": "save_settings", "id": "01JT0000000000000000000001", "variables": [ { "key": "k", "secret": true, "secret_value": "arg-secret-settings-4" } ] }
      """
    Then the app data file "workspaces/{{Acme.local_id}}/history.jsonl" should exist
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" should not contain "arg-secret-add-1"
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" should not contain "arg-secret-current-2"
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" should not contain "arg-passphrase-3"
    And the app data file "workspaces/{{Acme.local_id}}/history.jsonl" should not contain "arg-secret-settings-4"
    And the log output should not contain "arg-secret-add-1"
    And the log output should not contain "arg-passphrase-3"
