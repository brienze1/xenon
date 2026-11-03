#language: en
#utf-8

@all @xen_023 @mcp @mcp_concurrent_writes
Feature: Parallel MCP sessions editing the same workspace never lose a write
  Writes to one workspace are serialised, and an MCP edit always lands on top of the latest stored version: two
  sessions editing different fields of the same request, or adding different variables to the same environment, at
  the same time, both keep every change and neither gets a conflict.

  Background:
    Given the workspace "Acme" exists
    And the "dev" variable is set to "01JW8000000000000000000410"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JW8000000000000000000400
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JW8000000000000000000401
      name: Billing
      order:
        - ping.yaml
      """
    And the workspace file "acme/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JW8000000000000000000402
      name: Ping
      method: GET
      url: https://acme.test/ping
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: base_url
          value: https://dev.acme.test
          enabled: true
          description: ""
      """

  @parallel_request_edits
  Scenario: Two sessions editing different fields of one request both keep their changes
    Given the MCP client is named "one"
    And I start calling the MCP tool "tree_edit" 40 times in the background as "one" with
      """
      { "action": "update", "id": "01JW8000000000000000000402", "description": "one-{{call_number}}" }
      """
    And the MCP client is named "two"
    And I start calling the MCP tool "tree_edit" 40 times in the background as "two" with
      """
      { "action": "update", "id": "01JW8000000000000000000402", "url": "https://acme.test/two-{{call_number}}" }
      """
    Then the background call "one" should finish within 20000 ms
    And the call should succeed
    And the background call "two" should finish within 20000 ms
    And the call should succeed
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "description" equal to "one-40"
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "url" equal to "https://acme.test/two-40"

  @parallel_env_edits
  Scenario: Two sessions adding variables to one environment both keep every variable
    Given the MCP client is named "one"
    And I start calling the MCP tool "env_edit" 15 times in the background as "one" with
      """
      { "action": "add_var", "scope": "env:{{dev}}", "key": "one_{{call_number}}", "value": "1" }
      """
    And the MCP client is named "two"
    And I start calling the MCP tool "env_edit" 15 times in the background as "two" with
      """
      { "action": "add_var", "scope": "env:{{dev}}", "key": "two_{{call_number}}", "value": "2" }
      """
    Then the background call "one" should finish within 20000 ms
    And the call should succeed
    And the background call "two" should finish within 20000 ms
    And the call should succeed
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables" array length equal to 31
