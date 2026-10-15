#language: en
#utf-8

@all @xen_016 @mcp @aliases @mcp_aliases
Feature: AI clients list, create, edit, rename, delete and send request aliases
  Aliases are not tree items, so the MCP tools address them through their request: reading a
  request through "tree" lists its aliases with their content; "tree_edit" with kind "alias" and a
  request_id creates (a copy of the saved Default), updates (only the given fields) and renames an
  alias; "delete" with target "alias" removes one by id and owner_id; "send" takes an optional
  alias_id next to request_id. The rules are the UI's - name rules, independence of every version -
  and every call is recorded as AI activity.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the MCP client is named "Claude Code"
    And the "billing" variable is set to "01JW0000000000000000000001"
    And the "login" variable is set to "01JW0000000000000000000002"
    And the "mx" variable is set to "01JW0000000000000000000010"
    And the "unknown" variable is set to "01JW0000000000000000000099"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JW0000000000000000000000
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
      order:
        - login.yaml
      """
    And the workspace file "acme/collections/billing/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: "{{api}}/login"
      headers:
        - key: Content-Type
          value: application/json
          enabled: true
          description: ""
      body:
        mode: raw
        raw:
          language: json
          content: '{"username":"luis"}'
      aliases:
        - id: {{mx}}
          name: mx-test-user
          method: POST
          url: "{{api}}/login"
          headers:
            - key: Content-Type
              value: application/json
              enabled: true
              description: ""
          body:
            mode: raw
            raw:
              language: json
              content: '{"username":"mx-test"}'
      """

  @read
  Scenario: Reading a request lists its aliases with their content
    When I call the MCP tool "tree" with
      """
      { "action": "get", "id": "{{login}}" }
      """
    Then the call should succeed
    And the response should contain the field "url" equal to "{{api}}/login"
    And the response should contain the field "aliases" array length equal to 1
    And the response should contain the field "aliases.0.id" equal to "{{mx}}"
    And the response should contain the field "aliases.0.name" equal to "mx-test-user"
    And the response should contain the field "aliases.0.method" equal to "POST"
    And the response should contain the field "aliases.0.url" equal to "{{api}}/login"
    And the response should contain the field "aliases.0.body.raw.content" equal to
      """
      {"username":"mx-test"}
      """

  @create @update @rename
  Scenario: An AI client creates, updates and renames an alias, leaving the Default untouched
    When I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "alias", "request_id": "{{login}}", "name": "v2" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "v2"
    And the response should contain the field "url" equal to "{{api}}/login"
    And I save the response field "id" as "v2"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases" array length equal to 2
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.name" equal to "v2"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.body.raw.content" equal to
      """
      {"username":"luis"}
      """
    When I call the MCP tool "tree_edit" with
      """
      { "action": "update", "kind": "alias", "request_id": "{{login}}", "id": "{{v2}}", "url": "{{api}}/v2/login", "headers": [ { "key": "X-Canary", "value": "true" } ] }
      """
    Then the call should succeed
    And the response should contain the field "url" equal to "{{api}}/v2/login"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.url" equal to "{{api}}/v2/login"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.headers.0.key" equal to "X-Canary"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.method" equal to "POST"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.body.raw.content" equal to
      """
      {"username":"luis"}
      """
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "url" equal to "{{api}}/login"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "headers.0.key" equal to "Content-Type"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.url" equal to "{{api}}/login"
    When I call the MCP tool "tree_edit" with
      """
      { "action": "rename", "kind": "alias", "request_id": "{{login}}", "id": "{{v2}}", "name": "canary" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.name" equal to "canary"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the response should contain the field "0.activity.tool" equal to "tree_edit"
    And the response should contain the field "0.activity.action" equal to "rename"
    And the response should contain the field "0.activity.outcome" equal to "ok"
    And the response should contain the field "1.activity.action" equal to "update"
    And the response should contain the field "2.activity.action" equal to "create"
    And the response should contain the field "2.activity.target.id" equal to "{{login}}"

  @errors
  Scenario Outline: Alias edits follow the UI rules - <case>
    When I call the MCP tool "tree_edit" with
      """
      <arguments>
      """
    Then the call should fail with code "<code>"
    And the workspace file "acme/collections/billing/login.yaml" should be unchanged

    Examples:
      | case                     | arguments                                                                                              | code      |
      | reserved name            | { "action": "create", "kind": "alias", "request_id": "{{login}}", "name": "Default" }                  | ALS-04400 |
      | taken name               | { "action": "create", "kind": "alias", "request_id": "{{login}}", "name": "MX-test-user" }             | ALS-05409 |
      | blank name               | { "action": "create", "kind": "alias", "request_id": "{{login}}", "name": " " }                        | ALS-03400 |
      | unknown alias            | { "action": "rename", "kind": "alias", "request_id": "{{login}}", "id": "{{unknown}}", "name": "x" }   | ALS-01404 |
      | unknown request          | { "action": "create", "kind": "alias", "request_id": "{{unknown}}", "name": "x" }                      | MCP-08404 |
      | unsupported action       | { "action": "duplicate", "kind": "alias", "request_id": "{{login}}", "id": "{{mx}}" }                  | MCP-07400 |

  @delete
  Scenario: Deleting an alias removes it and keeps the Default
    When I call the MCP tool "delete" with
      """
      { "target": "alias", "id": "{{mx}}", "owner_id": "{{login}}" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/login.yaml" should not contain the field "aliases"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "url" equal to "{{api}}/login"
    When I call the MCP tool "delete" with
      """
      { "target": "alias", "id": "{{mx}}", "owner_id": "{{login}}" }
      """
    Then the call should fail with code "ALS-01404"

  @send
  Scenario: Sending an alias sends its stored version and History records the alias and the AI source
    Given the 1 "POST" request to "/login" returns status 200 with the following response
      """
      {"token":"t"}
      """
    When I call the MCP tool "send" with
      """
      { "request_id": "{{login}}", "alias_id": "{{mx}}" }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "200"
    And the 1 "POST" request to "/login" should have the body field "username" equal to "mx-test"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the response field "" should contain 1 items with "kind" equal to "send"
    And the response field "" should contain an item with "alias_id" equal to "{{mx}}" and "alias_name" equal to "mx-test-user"

  @send @errors
  Scenario Outline: An alias send that cannot be resolved sends nothing - <case>
    When I call the MCP tool "send" with
      """
      <arguments>
      """
    Then the call should fail with code "<code>"
    And the "POST" request to "/login" should have been received 0 times

    Examples:
      | case                     | arguments                                                                              | code      |
      | unknown alias            | { "request_id": "{{login}}", "alias_id": "{{unknown}}" }                               | ALS-01404 |
      | alias with ad-hoc request | { "alias_id": "{{mx}}", "request": { "method": "POST", "url": "{{api}}/login" } }     | MCP-07400 |
