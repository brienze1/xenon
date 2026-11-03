#language: en
#utf-8

@all @xen_023 @mcp @mcp_session_context
Feature: Every MCP session has its own workspace and environment
  An MCP session starts in the workspace the Xenon window shows and that workspace's active environment, at the moment
  it connects. From then on "workspace_edit switch|open|create" and "env_edit set_active" change only that session:
  other sessions and the window keep their own workspace and environment. "send", "snippet", "env inspect" and "run"
  resolve variables with the session's environment; "send" takes an optional per-call "environment". A session that
  ends forgets its context, and the next one starts from the window again.

  Background:
    Given the workspace "Globex" exists
    And the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "dev" variable is set to "01JW8000000000000000000310"
    And the "prod" variable is set to "01JW8000000000000000000311"
    And the "qa" variable is set to "01JW8000000000000000000312"
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: base_url
          value: "{{api}}/dev"
          enabled: true
          description: ""
        - key: tag
          value: dev
          enabled: true
          description: ""
      """
    And the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{prod}}
      name: prod
      variables:
        - key: base_url
          value: "{{api}}/prod"
          enabled: true
          description: ""
        - key: tag
          value: prod
          enabled: true
          description: ""
      """
    And the workspace file "globex/environments/qa.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{qa}}
      name: qa
      variables:
        - key: base_url
          value: "{{api}}/qa"
          enabled: true
          description: ""
      """
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" contains
      """
      { "version": 1, "activeEnvironmentId": "{{dev}}", "current": {} }
      """

  @seeded
  Scenario: A new session starts in the window's workspace and environment
    Given the MCP client is named "one"
    When I call the MCP tool "workspace" with
      """
      { "action": "active" }
      """
    Then the call should succeed
    And the response should contain the field "id" equal to "{{Acme.local_id}}"
    And the response should contain the field "environment_id" equal to "{{dev}}"
    When I call the MCP tool "workspace" with
      """
      { "action": "list" }
      """
    Then the response field "items" should contain an item with "id" equal to "{{Acme.local_id}}" and "active" equal to "true"
    And the response field "items" should contain 1 items with "active" equal to "true"
    Given I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Globex.local_id}}" }
      """
    When I call the MCP tool "workspace" with
      """
      { "action": "active" }
      """
    Then the response should contain the field "id" equal to "{{Acme.local_id}}"
    And the response should contain the field "environment_id" equal to "{{dev}}"

  @isolated_workspaces
  Scenario: Switching the workspace in one session leaves the other session and the window where they are
    Given the MCP client is named "one"
    And I call the MCP tool "workspace" with
      """
      { "action": "active" }
      """
    And the MCP client is named "two"
    When I call the MCP tool "workspace_edit" with
      """
      { "action": "switch", "id": "{{Globex.local_id}}" }
      """
    Then the call should succeed
    When I call the MCP tool "workspace" with
      """
      { "action": "active" }
      """
    Then the response should contain the field "id" equal to "{{Globex.local_id}}"
    When I call the MCP tool "env" with
      """
      { "action": "list" }
      """
    Then the response field "items" should contain 1 items with "id" equal to "{{qa}}"
    Given the MCP client is named "one"
    When I call the MCP tool "workspace" with
      """
      { "action": "active" }
      """
    Then the response should contain the field "id" equal to "{{Acme.local_id}}"
    When I call the MCP tool "env" with
      """
      { "action": "list" }
      """
    Then the response field "items" should contain an item with "id" equal to "{{dev}}" and "active" equal to "true"
    And the app data file "app.json" should contain the field "active_local_id" equal to "{{Acme.local_id}}"

  @isolated_environments
  Scenario: Choosing an environment in one session changes neither the other session nor the window
    Given the MCP client is named "one"
    And I call the MCP tool "workspace" with
      """
      { "action": "active" }
      """
    And the MCP client is named "two"
    When I call the MCP tool "env_edit" with
      """
      { "action": "set_active", "id": "{{prod}}" }
      """
    Then the call should succeed
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/ping" } }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "200"
    And the "GET" request to "/prod/ping" should have been received 1 times
    When I call the MCP tool "env" with
      """
      { "action": "list" }
      """
    Then the response field "items" should contain an item with "id" equal to "{{prod}}" and "active" equal to "true"
    Given the MCP client is named "one"
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/ping" } }
      """
    Then the call should succeed
    And the "GET" request to "/dev/ping" should have been received 1 times
    And the "GET" request to "/prod/ping" should have been received 1 times
    When I call the MCP tool "env" with
      """
      { "action": "list" }
      """
    Then the response field "items" should contain an item with "id" equal to "{{dev}}" and "active" equal to "true"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "activeEnvironmentId" equal to "{{dev}}"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "active_id" equal to "{{dev}}"

  @send_environment_override
  Scenario: A send can use another environment, or none, for that call only
    Given the MCP client is named "one"
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/ping" }, "environment": "prod" }
      """
    Then the call should succeed
    And the "GET" request to "/prod/ping" should have been received 1 times
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{api}}/tag", "headers": [ { "key": "X-Env", "value": "{{tag}}" } ] }, "environment": "{{prod}}" }
      """
    Then the call should succeed
    And the 1 "GET" request to "/tag" should have the header "X-Env" equal to "prod"
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{api}}/tag", "headers": [ { "key": "X-Env", "value": "{{tag}}" } ] }, "environment": "none", "send_unresolved": true }
      """
    Then the call should succeed
    And the 2 "GET" request to "/tag" should have the header "X-Env" matching "^[{][{]tag[}][}]$"
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/ping" }, "environment": "staging" }
      """
    Then the call should fail with code "MCP-07400"
    And the error should contain the field "error.message" containing "environment must be an environment name or id, or none"
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/ping" } }
      """
    Then the call should succeed
    And the "GET" request to "/dev/ping" should have been received 1 times
    When I call the MCP tool "workspace" with
      """
      { "action": "active" }
      """
    Then the response should contain the field "environment_id" equal to "{{dev}}"

  @snippet_and_inspect
  Scenario: Snippets and variable inspection use the session's environment
    Given the MCP client is named "one"
    And I call the MCP tool "workspace" with
      """
      { "action": "active" }
      """
    And the MCP client is named "two"
    And I call the MCP tool "env_edit" with
      """
      { "action": "set_active", "id": "{{prod}}" }
      """
    When I call the MCP tool "env" with
      """
      { "action": "inspect", "text": "{{base_url}}/ping" }
      """
    Then the call should succeed
    And the response field "refs" should contain an item with "name" equal to "base_url" and "value" equal to "{{api}}/prod"
    When I call the MCP tool "snippet" with
      """
      { "action": "generate", "language": "curl", "request": { "method": "GET", "url": "{{base_url}}/ping" } }
      """
    Then the call should succeed
    And the response should contain the field "code" containing "{{api}}/prod/ping"
    Given the MCP client is named "one"
    When I call the MCP tool "env" with
      """
      { "action": "inspect", "text": "{{base_url}}/ping" }
      """
    Then the response field "refs" should contain an item with "name" equal to "base_url" and "value" equal to "{{api}}/dev"
    When I call the MCP tool "snippet" with
      """
      { "action": "generate", "language": "curl", "request": { "method": "GET", "url": "{{base_url}}/ping" } }
      """
    Then the response should contain the field "code" containing "{{api}}/dev/ping"

  @reenter_keeps_choice
  Scenario: Coming back to a workspace restores the environment the session chose there
    Given the MCP client is named "one"
    And I call the MCP tool "env_edit" with
      """
      { "action": "set_active", "id": "{{prod}}" }
      """
    And I call the MCP tool "workspace_edit" with
      """
      { "action": "switch", "id": "{{Globex.local_id}}" }
      """
    When I call the MCP tool "env_edit" with
      """
      { "action": "set_active", "id": "{{qa}}" }
      """
    Then the call should succeed
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/ping" } }
      """
    Then the "GET" request to "/qa/ping" should have been received 1 times
    When I call the MCP tool "workspace_edit" with
      """
      { "action": "switch", "id": "{{Acme.local_id}}" }
      """
    And I call the MCP tool "workspace" with
      """
      { "action": "active" }
      """
    Then the response should contain the field "id" equal to "{{Acme.local_id}}"
    And the response should contain the field "environment_id" equal to "{{prod}}"
    And the app data file "app.json" should contain the field "active_local_id" equal to "{{Acme.local_id}}"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "activeEnvironmentId" equal to "{{dev}}"

  @closed_session
  Scenario: A session keeps its snapshot while connected, and a new session starts from the window again
    Given the MCP client is named "one"
    And I call the MCP tool "workspace" with
      """
      { "action": "active" }
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    When I call the MCP tool "workspace" with
      """
      { "action": "active" }
      """
    Then the response should contain the field "environment_id" equal to "{{dev}}"
    Given the MCP session is closed
    When I call the MCP tool "workspace" with
      """
      { "action": "active" }
      """
    Then the call should succeed
    And the response should contain the field "environment_id" equal to "{{prod}}"

  @workspace_removed
  Scenario: A session whose workspace is removed has no workspace until it picks one
    Given the MCP client is named "one"
    And I call the MCP tool "workspace_edit" with
      """
      { "action": "switch", "id": "{{Globex.local_id}}" }
      """
    And I call "WorkspaceService.Remove" with
      """
      { "local_id": "{{Globex.local_id}}", "move_to_trash": false }
      """
    When I call the MCP tool "tree" with
      """
      { "action": "list" }
      """
    Then the call should fail with code "MCP-06404"
    When I call the MCP tool "workspace_edit" with
      """
      { "action": "switch", "id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    When I call the MCP tool "tree" with
      """
      { "action": "list" }
      """
    Then the call should succeed

  @environment_deleted
  Scenario: A session whose environment is deleted resolves with no environment
    Given the MCP client is named "one"
    And I call the MCP tool "env_edit" with
      """
      { "action": "set_active", "id": "{{prod}}" }
      """
    And I call "EnvironmentService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    When I call the MCP tool "env" with
      """
      { "action": "list" }
      """
    Then the call should succeed
    And the response field "items" should contain 0 items with "active" equal to "true"
    When I call the MCP tool "env" with
      """
      { "action": "inspect", "text": "{{tag}}" }
      """
    Then the call should succeed
    And the response field "refs" should contain 0 items with "value" equal to "dev"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "active_id" equal to "{{dev}}"
