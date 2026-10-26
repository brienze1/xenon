#language: en
#utf-8

@all @xen_008 @mcp @mcp_secrets_write_only
Feature: Secrets are write-only for AI clients
  An AI can create a secret variable, set or replace its value (straight into the vault), turn a plain variable into a
  secret and delete a secret, in the globals, an environment or a collection's variables. It can send requests that
  use secrets. It can never read a secret back: listings and inspections show "•••••", turning a secret plain is
  refused, snippets and exports with secrets included are refused, and there is no tool to reveal, copy, set up,
  unlock or reset the vault. Those stay human-only in Xenon (MCP-02403).

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the "api" variable is set to the mocked api url
    And the "dev" variable is set to "01JY0000000000000000000010"
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
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """

  @create_secret
  Scenario: Adding a secret variable stores the value in the vault only and shows it masked
    When I call the MCP tool "env_edit" with
      """
      { "action": "add_var", "scope": "env:{{dev}}", "key": "api_token", "value": "tok-ai-written-123", "secret": true }
      """
    Then the call should succeed
    And the response should not contain "tok-ai-written-123"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.key" equal to "api_token"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.secret" equal to "true"
    And the workspace file "acme/environments/dev.yaml" should not contain the field "variables.1.value"
    And no file under the workspace folder should contain "tok-ai-written-123"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_token" }
      """
    Then the response should contain the field "value" equal to "tok-ai-written-123"

  @replace_secret
  Scenario: Replacing a secret's value updates the vault and never echoes either value
    Given I call the MCP tool "env_edit" with
      """
      { "action": "add_var", "scope": "env:{{dev}}", "key": "api_token", "value": "first-value-aaaa", "secret": true }
      """
    When I call the MCP tool "env_edit" with
      """
      { "action": "set_current", "scope": "env:{{dev}}", "key": "api_token", "value": "second-value-bbbb" }
      """
    Then the call should succeed
    And the response should not contain "first-value-aaaa"
    And the response should not contain "second-value-bbbb"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_token" }
      """
    Then the response should contain the field "value" equal to "second-value-bbbb"

  @scopes
  Scenario Outline: Secrets can be written in every scope that holds variables - <scope_name>
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JY0000000000000000000001
      name: Billing
      """
    When I call the MCP tool "env_edit" with
      """
      { "action": "add_var", "scope": "<scope>", "key": "scoped_secret", "value": "scoped-secret-value-1", "secret": true }
      """
    Then the call should succeed
    And the response should not contain "scoped-secret-value-1"
    When I call the MCP tool "env" with
      """
      { "action": "get", "scope": "<scope>" }
      """
    Then the response field "vars" should contain an item with "key" equal to "scoped_secret" and "state" equal to "set"
    And the response should not contain "scoped-secret-value-1"
    And no file under the workspace folder should contain "scoped-secret-value-1"

    Examples:
      | scope_name  | scope                                 |
      | globals     | globals                               |
      | environment | env:{{dev}}                           |
      | collection  | collection:01JY0000000000000000000001 |

  @listing_masked
  Scenario: Listings and inspections show plain values in full and secrets masked
    Given I call the MCP tool "env_edit" with
      """
      { "action": "add_var", "scope": "env:{{dev}}", "key": "api_token", "value": "listing-secret-777", "secret": true }
      """
    When I call the MCP tool "env" with
      """
      { "action": "get", "scope": "env:{{dev}}" }
      """
    Then the call should succeed
    And the response field "vars" should contain an item with "key" equal to "base_url" and "initial" equal to "{{api}}"
    And the response field "vars" should contain an item with "key" equal to "api_token" and "secret" equal to "true"
    And the response field "vars" should contain an item with "key" equal to "api_token" and "state" equal to "set"
    And the response should not contain "listing-secret-777"
    When I call the MCP tool "env" with
      """
      { "action": "inspect", "text": "{{base_url}}/me?t={{api_token}}" }
      """
    Then the call should succeed
    And the response field "refs" should contain an item with "name" equal to "base_url" and "value" equal to "{{api}}"
    And the response field "refs" should contain an item with "name" equal to "api_token" and "value" equal to "•••••"
    And the response should not contain "listing-secret-777"

  @plain_to_secret
  Scenario: A plain variable can be turned into a secret, which moves its value into the vault
    When I call the MCP tool "env_edit" with
      """
      { "action": "set_vars", "scope": "env:{{dev}}", "vars": [ { "key": "base_url", "secret": true, "secret_value": "https://plain-to-secret.example.test" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.0.secret" equal to "true"
    And the workspace file "acme/environments/dev.yaml" should not contain the field "variables.0.value"
    And the response should not contain "plain-to-secret.example.test"
    And no file under the workspace folder should contain "plain-to-secret.example.test"

  @secret_to_plain
  Scenario: Turning a secret plain would reveal it, so it is refused and nothing changes
    Given I call the MCP tool "env_edit" with
      """
      { "action": "add_var", "scope": "env:{{dev}}", "key": "api_token", "value": "keep-me-secret-9", "secret": true }
      """
    When I call the MCP tool "env_edit" with
      """
      { "action": "set_vars", "scope": "env:{{dev}}", "vars": [ { "key": "base_url", "initial": "{{api}}" }, { "key": "api_token", "secret": false } ] }
      """
    Then the call should fail with code "MCP-02403"
    And the error should contain the field "error.message" containing "human-only"
    And the error should not contain "keep-me-secret-9"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.secret" equal to "true"
    And no file under the workspace folder should contain "keep-me-secret-9"

  @delete_secret
  Scenario: A secret variable can be deleted
    Given I call the MCP tool "env_edit" with
      """
      { "action": "add_var", "scope": "env:{{dev}}", "key": "api_token", "value": "to-be-deleted-42", "secret": true }
      """
    When I call the MCP tool "env_edit" with
      """
      { "action": "set_vars", "scope": "env:{{dev}}", "vars": [ { "key": "base_url", "initial": "{{api}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables" array length equal to 1
    When I call the MCP tool "env" with
      """
      { "action": "get", "scope": "env:{{dev}}" }
      """
    Then the response field "vars" should contain 0 items with "key" equal to "api_token"

  @send_with_secret
  Scenario: A request using a secret is sent with the real value
    Given I call the MCP tool "env_edit" with
      """
      { "action": "add_var", "scope": "env:{{dev}}", "key": "api_token", "value": "real-token-for-send", "secret": true }
      """
    And the 1 "GET" request to "/me" returns status 200 with the following response
      """
      {"user":"luis"}
      """
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/me", "headers": [ { "key": "Authorization", "value": "Bearer {{api_token}}" } ] } }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "200"
    And the 1 "GET" request to "/me" should have the header "Authorization" equal to "Bearer real-token-for-send"
    And the response should not contain "real-token-for-send"

  @refused
  Scenario: A snippet with secrets included is human-only
    When I call the MCP tool "snippet" with
      """
      { "action": "generate", "language": "curl", "include_secrets": true, "request": { "method": "GET", "url": "{{base_url}}/me" } }
      """
    Then the call should fail with code "MCP-02403"
    And the error should contain the field "error.message" containing "human-only"

  @refused
  Scenario: An export with secrets included is human-only and writes nothing
    When I call the MCP tool "export" with
      """
      { "action": "run", "target": { "kind": "environment", "id": "{{dev}}" }, "format": "xenon", "path": "{{temp_root}}/dev-export.yaml", "include_secrets": true }
      """
    Then the call should fail with code "MCP-02403"
    And the file "dev-export.yaml" should not exist

  @locked_vault
  Scenario: With a locked vault an AI gets the same "secrets unavailable" result as the UI and cannot unlock it
    Given the keychain is unavailable
    And I call "VaultService.SetupPassphrase" with
      """
      { "passphrase": "correct horse", "confirmation": "correct horse" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "base_url", "value": "{{api}}" }
      """
    And I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_token", "secret": true, "value": "locked-token-1" }
      """
    When the app is restarted
    And I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{api}}/me", "headers": [ { "key": "Authorization", "value": "Bearer {{api_token}}" } ] } }
      """
    Then the call should fail with code "SND-06409"
    And the "GET" request to "/me" should have been received 0 times
    When I call the MCP tool "env_edit" with
      """
      { "action": "add_var", "scope": "env:{{dev}}", "key": "another", "value": "while-locked-55", "secret": true }
      """
    Then the call should fail with code "VLT-04409"
    And the error should not contain "while-locked-55"

  @no_vault_tools
  Scenario Outline: No tool or action reveals, copies or manages the vault - <tool> <action>
    When I call the MCP tool "<tool>" with
      """
      { "action": "<action>" }
      """
    Then the call should fail with code "MCP-07400"

    Examples:
      | tool     | action           |
      | env      | reveal           |
      | env      | copy             |
      | env_edit | unlock           |
      | env_edit | reset_vault      |
      | env_edit | setup_passphrase |
      | tree     | reveal_token     |
      | tree     | copy_token       |
