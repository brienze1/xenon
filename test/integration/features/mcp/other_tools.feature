#language: en
#utf-8

@all @xen_008 @mcp @mcp_other_tools
Feature: AI clients use import, export, snippets, cookies, the console, scripts and settings
  Each tool behaves like the matching Xenon feature in the window: import reads files by path, export writes a path
  (never with secrets), snippets are generated with secrets masked, cookies and the allowlist are read and edited,
  the console is read and cleared, scripts are checked and the pm reference read, and app settings other than AI
  access are read and changed. Changes emit the same events the window listens to.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url

  @import
  Scenario: Importing a Postman collection by path adds it to the active workspace
    Given the emitted events are forgotten
    When I call the MCP tool "import" with
      """
      { "action": "detect", "paths": ["{{resources}}/import/postman/full-v2.1.json"] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.name" equal to "Acme API"
    When I call the MCP tool "import" with
      """
      { "action": "run", "paths": ["{{resources}}/import/postman/full-v2.1.json"] }
      """
    Then the call should succeed
    And the workspace folder "acme/collections" should contain the entries "acme-api"
    And a "workspace:tree-changed" event with the field "local_id" equal to "{{Acme.local_id}}" should have been emitted within 1000 ms

  @import @clashes
  Scenario Outline: An import clash is resolved as asked - <on_clash>
    Given I call the MCP tool "import" with
      """
      { "action": "run", "paths": ["{{resources}}/import/postman/full-v2.1.json"] }
      """
    When I call the MCP tool "import" with
      """
      { "action": "clashes", "paths": ["{{resources}}/import/postman/full-v2.1.json"] }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 1
    When I call the MCP tool "import" with
      """
      { "action": "run", "paths": ["{{resources}}/import/postman/full-v2.1.json"], "on_clash": "<on_clash>" }
      """
    Then the call should succeed
    And the workspace folder "acme/collections" should contain the entries "<entries>"

    Examples:
      | on_clash  | entries                          |
      | replace   | acme-api                         |
      | keep_both | acme-api, acme-api-(imported)    |
      | skip      | acme-api                         |

  @import @xen_009 @mcp_import_v3
  Scenario: Importing a Postman v3 workspace folder by path gives the dialog's preview and imports every item
    When I list the MCP tools
    Then the response field "" should contain an item with "name" equal to "import" and "description" equal to "Import Postman v2 or v3 files and folders, Bruno or Xenon files from paths into the workspace."
    When I call the MCP tool "import" with
      """
      { "action": "detect", "paths": ["{{resources}}/import/postman/v3/workspace"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "postman_workspace"
    And the response should contain the field "version" equal to "3"
    And the response should contain the field "items" array length equal to 5
    And the response field "warnings" should contain 2 items with "code" equal to "skipped_part"
    And the response should not contain "s3cr3t-fixture-workspace"
    When I call the MCP tool "import" with
      """
      { "action": "run", "paths": ["{{resources}}/import/postman/v3/workspace"] }
      """
    Then the call should succeed
    And the response should not contain "s3cr3t-fixture-workspace"
    And the workspace folder "acme/collections" should contain the entries "billing-api, users"
    And the workspace folder "acme/environments" should contain the entries "prod.yaml, staging.yaml"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.0.key" equal to "tenant"
    And no file under the workspace folder should contain "s3cr3t-fixture-workspace"

  @import @xen_009 @mcp_import_v3
  Scenario: Importing a v3 collection JSON by path reports the same kind and version as the dialog
    When I call the MCP tool "import" with
      """
      { "action": "detect", "paths": ["{{resources}}/import/postman/v3/rich.json"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "postman_collection"
    And the response should contain the field "version" equal to "3"
    And the response should contain the field "items.0.name" equal to "Rich API"
    And the response field "warnings" should contain 2 items with "code" equal to "missing_file"

  @export
  Scenario Outline: Exporting a collection to a path, as <format>
    Given I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "collection", "name": "Payments" }
      """
    And I save the response field "id" as "payments"
    When I call the MCP tool "export" with
      """
      { "action": "options", "target": { "kind": "collection", "id": "{{payments}}" } }
      """
    Then the call should succeed
    When I call the MCP tool "export" with
      """
      { "action": "run", "target": { "kind": "collection", "id": "{{payments}}" }, "format": "<format>", "path": "{{temp_root}}/<file>" }
      """
    Then the call should succeed
    And the file "<file>" should exist

    Examples:
      | format  | file                  |
      | xenon   | payments.xenon.yaml   |
      | postman | payments.postman.json |

  @export
  Scenario: A Postman export is a valid Postman collection
    Given I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "collection", "name": "Payments" }
      """
    And I save the response field "id" as "payments"
    When I call the MCP tool "export" with
      """
      { "action": "run", "target": { "kind": "collection", "id": "{{payments}}" }, "format": "postman", "path": "{{temp_root}}/payments.json" }
      """
    Then the file "payments.json" should be a valid Postman v2.1 collection

  @snippets
  Scenario: Snippet languages are listed and a snippet is generated
    When I call the MCP tool "snippet" with
      """
      { "action": "languages" }
      """
    Then the call should succeed
    And the response field "items" should contain 1 items with "id" equal to "curl"
    When I call the MCP tool "snippet" with
      """
      { "action": "generate", "language": "curl", "request": { "method": "GET", "url": "{{api}}/users", "headers": [ { "key": "Accept", "value": "application/json" } ] } }
      """
    Then the call should succeed
    And the response should contain the field "code" containing "curl"
    And the response should contain the field "code" containing "{{api}}/users"

  @cookies
  Scenario: Cookies are listed, added, deleted and cleared
    Given the emitted events are forgotten
    When I call the MCP tool "cookies_edit" with
      """
      { "action": "upsert", "domain": "example.test", "raw": "token=t1; Path=/api; Secure; HttpOnly" }
      """
    Then the call should succeed
    And a "cookies:changed" event with the field "local_id" equal to "{{Acme.local_id}}" should have been emitted within 1000 ms
    When I call the MCP tool "cookies" with
      """
      { "action": "list" }
      """
    Then the response field "items" should contain an item with "domain" equal to "example.test" and "name" equal to "token"
    When I call the MCP tool "cookies_edit" with
      """
      { "action": "upsert", "domain": "other.test", "raw": "a=1; Path=/" }
      """
    And I call the MCP tool "delete" with
      """
      { "target": "cookie", "domain": "example.test", "name": "token", "path": "/api" }
      """
    Then the call should succeed
    When I call the MCP tool "cookies" with
      """
      { "action": "list" }
      """
    Then the response should contain the field "items" array length equal to 1
    When I call the MCP tool "delete" with
      """
      { "target": "cookies" }
      """
    And I call the MCP tool "cookies" with
      """
      { "action": "list" }
      """
    Then the response should contain the field "items" array length equal to 0

  @cookies
  Scenario: The cookie allowlist is read and saved
    When I call the MCP tool "cookies_edit" with
      """
      { "action": "allowlist", "domains": ["api.acme.test", "auth.acme.test"] }
      """
    Then the call should succeed
    When I call the MCP tool "cookies" with
      """
      { "action": "allowlist" }
      """
    Then the response should contain the field "domains" with the items "api.acme.test, auth.acme.test"

  @console
  Scenario: The console is read and cleared
    Given I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{api}}/ping", "scripts": { "pre_request": "console.log('hello from script')", "post_response": "", "disabled": [] } } }
      """
    When I call the MCP tool "console" with
      """
      { "action": "list", "type": "log" }
      """
    Then the call should succeed
    And the response field "items" should contain an item with "type" equal to "log" and "summary" equal to "hello from script"
    When I call the MCP tool "delete" with
      """
      { "target": "console" }
      """
    Then the call should succeed
    When I call the MCP tool "console" with
      """
      { "action": "list" }
      """
    Then the response should contain the field "items" array length equal to 0

  @scripts
  Scenario: Scripts are checked and the pm reference is read
    When I call the MCP tool "scripts" with
      """
      { "action": "check", "text": "pm.test('x', function () {" }
      """
    Then the call should succeed
    And the response should contain the field "errors" array length equal to 1
    When I call the MCP tool "scripts" with
      """
      { "action": "check", "text": "pm.test('x', function () {});" }
      """
    Then the response should contain the field "errors" array length equal to 0
    When I call the MCP tool "scripts" with
      """
      { "action": "reference" }
      """
    Then the call should succeed
    And the response field "members" should contain an item with "path" equal to "pm.environment.set" and "kind" equal to "function"
    When I call the MCP tool "scripts" with
      """
      { "action": "snippets" }
      """
    Then the call should succeed
    And the response field "items.0.code" should exist

  @settings
  Scenario: App settings are read and changed, except AI access
    When I call the MCP tool "settings" with
      """
      { "action": "get" }
      """
    Then the call should succeed
    And the response should contain the field "script_timeout_ms" equal to "10000"
    And the response should contain the field "history_user_cap" equal to "1000"
    And the response should contain the field "history_ai_cap" equal to "1000"
    And the response field "ai_access" should not exist
    When I call the MCP tool "settings_edit" with
      """
      { "action": "save", "history_ai_cap": 500 }
      """
    Then the call should succeed
    And the app data file "app.json" should contain the field "settings.historyAiCap" equal to "500"
    And the app data file "app.json" should contain the field "settings.scriptTimeoutMs" equal to "10000"

  @settings
  Scenario Outline: AI access settings cannot be changed through MCP - <field>
    When I call the MCP tool "settings_edit" with
      """
      { "action": "save", "ai_access": { "<field>": <value> } }
      """
    Then the call should fail with code "MCP-02403"
    When I call "SettingsService.Get"
    Then the response should contain the field "ai_access.enabled" equal to "true"
    And the response should contain the field "ai_access.skip_approvals" equal to "false"

    Examples:
      | field          | value |
      | enabled        | false |
      | skip_approvals | true  |
      | port           | 50000 |

  @settings
  Scenario: Settings out of range are refused with the same codes as the UI
    When I call the MCP tool "settings_edit" with
      """
      { "action": "save", "history_user_cap": 99 }
      """
    Then the call should fail with code "SET-02400"
