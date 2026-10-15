#language: en
#utf-8

@all @xen_014 @mcp @mcp_tool_arguments
Feature: MCP tools declare their arguments and validation errors name them
  Every tool's input schema declares each argument its actions read, with its JSON type, a one-sentence description
  that says which actions use it and which require it, and the enum of a closed value set. Per-action requirements
  live in the descriptions; the schema never uses oneOf, anyOf or allOf. A failed call answers code, message and,
  when the failure has them, details of {field, problem}: a validation failure reads "invalid request: <field>
  <problem>" and names the argument the client sent, even when the action validates it under another name.

  Background:
    Given the workspace "Acme" exists

  @xen_014_schema @schema
  Scenario Outline: workspace_edit declares its arguments - <path>
    When I list the MCP tools
    Then the response field "" should contain an item with "name" equal to "workspace_edit" and "<path>" equal to "<value>"

    Examples:
      | path                                          | value                                                                                   |
      | inputSchema.properties.id.type                | string                                                                                  |
      | inputSchema.properties.id.description         | Workspace id, from workspace list. Used by switch and rename. Required for switch and rename. |
      | inputSchema.properties.name.type              | string                                                                                  |
      | inputSchema.properties.name.description       | Workspace name, at most 200 characters. Used by create and rename. Required for create and rename. |
      | inputSchema.properties.path.type              | string                                                                                  |
      | inputSchema.properties.path.description       | Absolute path of the workspace folder. Used by create and open. Required for create and open. |
      | inputSchema.properties.action.enum.0          | create                                                                                  |
      | inputSchema.required.0                        | action                                                                                  |

  @xen_014_schema @schema
  Scenario Outline: import declares paths as an array of strings and its other arguments - <path>
    When I list the MCP tools
    Then the response field "" should contain an item with "name" equal to "import" and "<path>" equal to "<value>"

    Examples:
      | path                                           | value                                                                                                      |
      | inputSchema.properties.paths.type              | array                                                                                                      |
      | inputSchema.properties.paths.items.type        | string                                                                                                     |
      | inputSchema.properties.paths.description       | Absolute paths of the files or folders to import. Used by detect, clashes and run. Required for detect, clashes and run. |
      | inputSchema.properties.source_id.type          | string                                                                                                     |
      | inputSchema.properties.on_clash.type           | string                                                                                                     |
      | inputSchema.properties.on_clash.enum.0         | keep_both                                                                                                  |
      | inputSchema.properties.on_clash.enum.1         | replace                                                                                                    |
      | inputSchema.properties.on_clash.enum.2         | skip                                                                                                       |
      | inputSchema.properties.passphrase.type         | string                                                                                                     |
      | inputSchema.required.0                         | action                                                                                                     |

  @xen_014_schema @schema
  Scenario Outline: Other tools declare typed arguments - <tool> <path>
    When I list the MCP tools
    Then the response field "" should contain an item with "name" equal to "<tool>" and "<path>" equal to "<value>"

    Examples:
      | tool          | path                                              | value   |
      | tree          | inputSchema.properties.depth.type.0               | integer |
      | tree          | inputSchema.properties.depth.type.1               | string  |
      | tree          | inputSchema.properties.limit.type                 | integer |
      | delete        | inputSchema.properties.target.enum.0              | item    |
      | delete        | inputSchema.properties.move_to_trash.type         | boolean |
      | export        | inputSchema.properties.target.type                | object  |
      | export        | inputSchema.properties.target.properties.kind.type | string |
      | git_edit      | inputSchema.properties.items.type                 | array   |
      | git_edit      | inputSchema.properties.items.items.type           | object  |
      | env_edit      | inputSchema.properties.index.type                 | integer |
      | env_edit      | inputSchema.properties.vars.items.type            | object  |
      | send          | inputSchema.properties.request.type               | object  |
      | send          | inputSchema.properties.request_id.type            | string  |
      | snippet       | inputSchema.properties.language.enum.0            | curl    |
      | settings_edit | inputSchema.properties.script_timeout_ms.type     | integer |
      | cookies_edit  | inputSchema.properties.domains.items.type         | string  |

  @xen_014_schema @schema @described
  Scenario: Every tool's input schema is an object and every tool declares more than its selector
    When I list the MCP tools
    Then every item of the response field "" should have the field "inputSchema.type" matching "^object$"
    And the response field "" should contain an item with "name" equal to "console" and "inputSchema.properties.query.type" equal to "string"
    And the response field "" should contain an item with "name" equal to "history" and "inputSchema.properties.source.enum.0" equal to "all"
    And the response field "" should contain an item with "name" equal to "docs_edit" and "inputSchema.properties.confirm_deletes.type" equal to "boolean"
    And the response field "" should contain an item with "name" equal to "scripts" and "inputSchema.properties.text.type" equal to "string"
    And the response field "" should contain an item with "name" equal to "response" and "inputSchema.properties.offset.type" equal to "integer"
    And the response should not contain "ai_access"
    And the response should not contain "oneOf"
    And the response should not contain "anyOf"
    And the response should not contain "allOf"

  @xen_014_errors @validation
  Scenario: Creating a workspace without a path names the missing argument
    When I call the MCP tool "workspace_edit" with
      """
      { "action": "create", "name": "Payments" }
      """
    Then the call should fail with code "WKS-00400"
    And the error should contain the field "error.message" equal to "invalid request: path is required"
    And the error should contain the field "error.details.0.field" equal to "path"
    And the error should contain the field "error.details.0.problem" equal to "is required"

  @xen_014_errors @validation
  Scenario: Creating a workspace with a relative path names the broken rule
    When I call the MCP tool "workspace_edit" with
      """
      { "action": "create", "name": "Payments", "path": "relative/dir" }
      """
    Then the call should fail with code "WKS-00400"
    And the error should contain the field "error.message" equal to "invalid request: path must be an absolute path"
    And the error should contain the field "error.details.0.field" equal to "path"
    And the error should contain the field "error.details.0.problem" equal to "must be an absolute path"

  @xen_014_errors @validation
  Scenario: Every failing argument is listed in field order
    When I call the MCP tool "workspace_edit" with
      """
      { "action": "create" }
      """
    Then the call should fail with code "WKS-00400"
    And the error should contain the field "error.message" equal to "invalid request: name is required; path is required"
    And the error should contain the field "error.details.0.field" equal to "name"
    And the error should contain the field "error.details.1.field" equal to "path"
    And the error should contain the field "error.details.1.problem" equal to "is required"

  @xen_014_errors @validation @mapped_name
  Scenario: A missing id is reported under the argument name, not the internal one
    When I call the MCP tool "workspace_edit" with
      """
      { "action": "switch" }
      """
    Then the call should fail with code "WKS-08400"
    And the error should contain the field "error.message" equal to "invalid request: id is required"
    And the error should contain the field "error.details.0.field" equal to "id"
    And the error should contain the field "error.details.0.problem" equal to "is required"
    And the error should not contain "local_id"

  @xen_014_errors @validation @mapped_name
  Scenario: An unknown clash choice is reported on on_clash
    When I call the MCP tool "import" with
      """
      { "action": "run", "paths": ["{{resources}}/import/postman/full-v2.1.json"], "on_clash": "merge" }
      """
    Then the call should fail with code "IMP-13400"
    And the error should contain the field "error.details.0.field" equal to "on_clash"
    And the error should contain the field "error.details.0.problem" equal to "must be one of replace, keep_both, skip"
    And the error should not contain "apply_to_all"
    And the workspace folder "acme/collections" should contain the entries ""

  @xen_014_errors @no_details
  Scenario: A failure without details answers no details list
    When I call the MCP tool "tree" with
      """
      { "action": "get", "id": "01HZZZZZZZZZZZZZZZZZZZZZZZ" }
      """
    Then the call should fail with code "MCP-08404"
    And the error should contain the field "error.message" containing "01HZZZZZZZZZZZZZZZZZZZZZZZ"
    And the error should not contain "details"
