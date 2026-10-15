#language: en
#utf-8

@all @xen_008 @mcp @mcp_connect
Feature: AI clients connect to Xenon's MCP server
  Xenon serves MCP over Streamable HTTP at http://127.0.0.1:<port>/mcp while "AI access (MCP)" is on (the default).
  Only loopback requests whose Host and Origin are loopback are handled, there is no token, and bodies over 4 MiB are
  refused. The server exposes 24 resource tools with an action argument; read-only tools are hinted read-only, the
  tools that can remove data are hinted destructive, and every description is one sentence. Turning AI access off or
  changing the port in Settings takes effect at once.

  Background:
    Given the workspace "Acme" exists

  @tool_list
  Scenario: A client connects and lists the 24 tools quickly and compactly
    When I list the MCP tools
    Then the last call should have taken less than 1000 ms
    And the response should contain the field "" array length equal to 24
    And the response should have at most 36000 characters
    And every item of the response field "" should have the field "description" matching "^[A-Z][^.]*\.$"
    And every item of the response field "" should have the field "name" matching "^[a-z]+(_[a-z]+)?$"

  @tool_hints
  Scenario Outline: Read-only tools are hinted read-only - <tool>
    When I list the MCP tools
    Then the response field "" should contain an item with "name" equal to "<tool>" and "annotations.readOnlyHint" equal to "true"
    And the response field "" should contain an item with "name" equal to "<tool>" and "annotations.idempotentHint" equal to "true"

    Examples:
      | tool      |
      | workspace |
      | tree      |
      | env       |
      | history   |
      | docs      |
      | response  |
      | snippet   |
      | cookies   |
      | console   |
      | git       |
      | scripts   |
      | settings  |

  @tool_hints
  Scenario Outline: Tools that change data carry the destructive hint they deserve - <tool>
    When I list the MCP tools
    Then the response field "" should contain an item with "name" equal to "<tool>" and "annotations.readOnlyHint" equal to "false"
    And the response field "" should contain an item with "name" equal to "<tool>" and "annotations.destructiveHint" equal to "<destructive>"
    And the response field "" should contain an item with "name" equal to "<tool>" and "annotations.idempotentHint" equal to "<idempotent>"

    Examples:
      | tool           | destructive | idempotent |
      | workspace_edit | false       | false      |
      | settings_edit  | false       | true       |
      | tree_edit      | true        | false      |
      | env_edit       | true        | false      |
      | docs_edit      | true        | false      |
      | cookies_edit   | true        | true       |
      | git_edit       | true        | false      |
      | import         | true        | false      |
      | export         | true        | false      |
      | send           | true        | false      |
      | delete         | true        | true       |
      | git_push       | true        | false      |

  @tool_hints
  Scenario Outline: Only tools that reach outside Xenon are open-world - <tool>
    When I list the MCP tools
    Then the response field "" should contain an item with "name" equal to "<tool>" and "annotations.openWorldHint" equal to "<open_world>"

    Examples:
      | tool      | open_world |
      | send      | true       |
      | git_push  | true       |
      | git_edit  | true       |
      | tree_edit | true       |
      | tree      | false      |
      | env_edit  | false      |
      | delete    | false      |

  @never_exposed
  Scenario Outline: Human-only features have no tool - <name>
    When I list the MCP tools
    Then the response field "" should contain 0 items with "name" equal to "<name>"

    Examples:
      | name         |
      | vault        |
      | reveal       |
      | copy         |
      | unlock       |
      | git_accounts |
      | dialog       |
      | window       |
      | ui_state     |
      | clipboard    |

  @errors
  Scenario: An unknown action is a tool error with a Xenon code, not a protocol failure
    When I call the MCP tool "tree" with
      """
      { "action": "explode" }
      """
    Then the call should fail with code "MCP-07400"

  @errors
  Scenario: A missing action is a tool error with a Xenon code
    When I call the MCP tool "tree"
    Then the call should fail with code "MCP-07400"

  @status
  Scenario: Xenon reports the running server, the stdio command and ready-to-paste configs
    When I call "AIAccessService.Status"
    Then the call should succeed
    And the response should contain the field "state" equal to "running"
    And the response should contain the field "address" equal to "{{mcp_url}}"
    And the response should contain the field "stdio_command" matching " mcp$"
    And the response should contain the field "http_config" containing "{{mcp_url}}"
    And the response should contain the field "stdio_config" matching "args.{1,4}\\[.{0,2}mcp"

  @on_off
  Scenario: Turning AI access off stops the server and turning it on starts it again, without a restart
    When I call "SettingsService.Save" with
      """
      { "ai_access": { "enabled": false, "port": {{mcp_port}}, "skip_approvals": false } }
      """
    Then the call should succeed
    And the last call should have taken less than 2000 ms
    And the MCP endpoint should refuse connections
    And an "ai:access-status-changed" event with the field "state" equal to "off" should have been emitted within 2000 ms
    When I call "AIAccessService.Status"
    Then the response should contain the field "state" equal to "off"
    When I call "SettingsService.Save" with
      """
      { "ai_access": { "enabled": true, "port": {{mcp_port}}, "skip_approvals": false } }
      """
    Then the call should succeed
    When I list the MCP tools
    Then the response should contain the field "" array length equal to 24

  @port_change
  Scenario: Changing the port moves the server at once
    Given the "new_port" variable is set to a free local port
    And the "old_port" variable is set to "{{mcp_port}}"
    When I call "SettingsService.Save" with
      """
      { "ai_access": { "enabled": true, "port": {{new_port}}, "skip_approvals": false } }
      """
    Then the call should succeed
    And the MCP endpoint should refuse connections
    Given the "mcp_port" variable is set to "{{new_port}}"
    When I call "AIAccessService.Status"
    Then the response should contain the field "address" equal to "http://127.0.0.1:{{new_port}}/mcp"
    When I list the MCP tools
    Then the response should contain the field "" array length equal to 24

  @port_busy
  Scenario: A port in use is reported in Settings and another port can be picked
    Given the "busy_port" variable is set to a free local port
    And the local port "{{busy_port}}" is in use
    When I call "SettingsService.Save" with
      """
      { "ai_access": { "enabled": true, "port": {{busy_port}}, "skip_approvals": false } }
      """
    Then the call should succeed
    When I call "AIAccessService.Status"
    Then the response should contain the field "state" equal to "error"
    And the response should contain the field "reason" equal to "port {{busy_port}} is in use"
    Given the "free_port" variable is set to a free local port
    When I call "SettingsService.Save" with
      """
      { "ai_access": { "enabled": true, "port": {{free_port}}, "skip_approvals": false } }
      """
    And I call "AIAccessService.Status"
    Then the response should contain the field "state" equal to "running"
    And the response should contain the field "address" equal to "http://127.0.0.1:{{free_port}}/mcp"

  @loopback_only
  Scenario Outline: Requests whose Host or Origin is not loopback are rejected and logged - <case>
    Given the log output is captured
    When I post to the MCP endpoint with the headers
      | header | value    |
      | <name> | <value>  |
    Then the route response status should be 403
    And the log output should contain "MCP request rejected"
    And the log output should contain "<reason>"

    Examples:
      | case             | name   | value                          | reason |
      | foreign host     | Host   | evil.example:{{mcp_port}}      | host   |
      | rebinding host   | Host   | attacker.test                  | host   |
      | foreign origin   | Origin | https://evil.example           | origin |
      | null origin      | Origin | null                           | origin |

  @loopback_only
  Scenario Outline: Loopback requests are handled - <case>
    When I post to the MCP endpoint with the headers
      | header | value   |
      | <name> | <value> |
    Then the route response status should be 200

    Examples:
      | case               | name   | value                             |
      | loopback host      | Host   | 127.0.0.1:{{mcp_port}}            |
      | localhost host     | Host   | localhost:{{mcp_port}}            |
      | loopback origin    | Origin | http://127.0.0.1:{{mcp_port}}     |
      | localhost origin   | Origin | http://localhost:{{mcp_port}}     |
      | no origin (CLI)    | Accept | application/json, text/event-stream |

  @size_limit
  Scenario: A request body over 4 MiB is refused
    When I post a body of 4194305 bytes to the MCP endpoint
    Then the route response status should be 413

  @concurrent_clients
  Scenario: Two clients work at the same time and both see each other's changes
    Given the MCP client is named "Claude Code"
    When I start calling the MCP tool "tree_edit" in the background with
      """
      { "action": "create", "kind": "collection", "name": "From Claude" }
      """
    And the MCP client is named "Cursor"
    And I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "collection", "name": "From Cursor" }
      """
    Then the call should succeed
    And the background call should finish within 5000 ms
    And the call should succeed
    When I call the MCP tool "tree" with
      """
      { "action": "list" }
      """
    Then the response field "items" should contain an item with "kind" equal to "collection" and "name" equal to "From Claude"
    And the response field "items" should contain an item with "kind" equal to "collection" and "name" equal to "From Cursor"
