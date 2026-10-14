#language: en
#utf-8

@all @xen_008 @mcp @mcp_stdio_bridge
Feature: The xenon mcp stdio bridge
  Clients that only speak stdio run "<xenon executable> mcp". The bridge reads the AI access settings from app.json,
  connects to the running Xenon's MCP server and pumps JSON-RPC messages between stdin/stdout and that server. It
  never starts the Xenon window. When AI access is off, when Xenon is not running, or when another program answers on
  the port, it prints one line naming the cause and exits with code 1, forwarding nothing.

  Background:
    Given the workspace "Acme" exists

  @success
  Scenario: A stdio client lists Xenon's tools through the bridge
    When I run the stdio bridge with the input
      """
      {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"stdio-client","version":"1.0.0"}}}
      {"jsonrpc":"2.0","method":"notifications/initialized"}
      {"jsonrpc":"2.0","id":2,"method":"tools/list"}
      """
    Then the bridge exit code should be 0
    And the bridge output should match "serverInfo.{1,4}name.{1,4}xenon"
    And the bridge output should match "name.{1,4}tree"
    And the bridge output should match "name.{1,4}send"

  @ai_access_off
  Scenario: AI access is off
    When I call "SettingsService.Save" with
      """
      { "ai_access": { "enabled": false, "port": {{mcp_port}}, "skip_approvals": false } }
      """
    And I run the stdio bridge with the input
      """
      {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"stdio-client","version":"1.0.0"}}}
      """
    Then the bridge exit code should be 1
    And the bridge error output should contain "Xenon AI access is off — turn it on in Xenon → Settings → AI access (MCP)."
    And the bridge output should not contain "jsonrpc"

  @not_running
  Scenario: Xenon is not running on the configured port
    Given the "closed_port" variable is set to a free local port
    And the "XENON_MCP_PORT" env var is set to "{{closed_port}}"
    When I run the stdio bridge with the input
      """
      {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"stdio-client","version":"1.0.0"}}}
      """
    Then the bridge exit code should be 1
    And the bridge error output should contain "Xenon is not running — start Xenon with AI access on, then retry."
    And the bridge output should not contain "jsonrpc"

  @foreign_server
  Scenario: Another program answers on the port, so nothing is forwarded to it
    Given the "foreign_port" variable is set to the port of the mocked api url
    And the "XENON_MCP_PORT" env var is set to "{{foreign_port}}"
    And the 1 "POST" request to "/mcp" returns status 200 with the following response
      """
      {"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-11-25","capabilities":{"tools":{}},"serverInfo":{"name":"not-xenon","version":"9.9.9"}}}
      """
    When I run the stdio bridge with the input
      """
      {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"stdio-client","version":"1.0.0"}}}
      {"jsonrpc":"2.0","method":"notifications/initialized"}
      {"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"send","arguments":{"request":{"method":"GET","url":"https://example.test"}}}}
      """
    Then the bridge exit code should be 1
    And the bridge error output should contain "is in use by another program — change the port in Xenon → Settings → AI access (MCP)."
    And the bridge output should not contain "not-xenon"
    And the "POST" request to "/mcp" should have been received 1 times
