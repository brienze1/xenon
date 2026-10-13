#language: en
#utf-8

@all @xen_008 @settings @app_settings
Feature: History caps and AI access settings
  App settings gain History caps per source (history_user_cap and history_ai_cap, 1,000 each by default, 100 to
  20,000) and AI access (enabled by default, port 47821 by default from 1024 to 65535, "Dangerously skip approvals"
  off by default). They are stored in app.json under settings, never in a workspace. Save leaves omitted fields
  unchanged and validates every field before storing anything; AI access changes apply at once.

  @defaults
  Scenario: Defaults before anything is saved
    Given the "XENON_MCP_PORT" env var is set to ""
    When I call "SettingsService.Get"
    Then the call should succeed
    And the response should contain the field "script_timeout_ms" equal to "10000"
    And the response should contain the field "history_user_cap" equal to "1000"
    And the response should contain the field "history_ai_cap" equal to "1000"
    And the response should contain the field "ai_access.enabled" equal to "true"
    And the response should contain the field "ai_access.port" equal to "47821"
    And the response should contain the field "ai_access.skip_approvals" equal to "false"

  @save
  Scenario: Saved values are stored in app.json and survive a restart
    Given the workspace "Acme" exists
    When I call "SettingsService.Save" with
      """
      { "history_user_cap": 10000, "history_ai_cap": 100, "ai_access": { "enabled": true, "port": {{mcp_port}}, "skip_approvals": true } }
      """
    Then the call should succeed
    And the response should contain the field "history_user_cap" equal to "10000"
    And the app data file "app.json" should contain the field "settings.historyUserCap" equal to "10000"
    And the app data file "app.json" should contain the field "settings.historyAiCap" equal to "100"
    And the app data file "app.json" should contain the field "settings.aiAccess.skipApprovals" equal to "true"
    And the app data file "app.json" should contain the field "settings.aiAccess.port" equal to "{{mcp_port}}"
    And no file under the workspace folder should contain "historyAiCap"
    Given the app is restarted
    When I call "SettingsService.Get"
    Then the response should contain the field "history_ai_cap" equal to "100"
    And the response should contain the field "ai_access.skip_approvals" equal to "true"

  @partial
  Scenario: Omitted fields keep their values
    Given I call "SettingsService.Save" with
      """
      { "history_ai_cap": 250, "ai_access": { "enabled": true, "port": {{mcp_port}}, "skip_approvals": true } }
      """
    When I call "SettingsService.Save" with
      """
      { "script_timeout_ms": 2500 }
      """
    Then the call should succeed
    And the response should contain the field "script_timeout_ms" equal to "2500"
    And the response should contain the field "history_ai_cap" equal to "250"
    And the response should contain the field "history_user_cap" equal to "1000"
    And the response should contain the field "ai_access.skip_approvals" equal to "true"

  @old_app_json
  Scenario: An app.json from before this version reads with the defaults
    Given the app data file "app.json" contains
      """
      { "version": 1, "settings": { "scriptTimeoutMs": 3000 }, "workspaces": [] }
      """
    And the app is restarted
    When I call "SettingsService.Get"
    Then the call should succeed
    And the response should contain the field "script_timeout_ms" equal to "3000"
    And the response should contain the field "history_user_cap" equal to "1000"
    And the response should contain the field "history_ai_cap" equal to "1000"
    And the response should contain the field "ai_access.enabled" equal to "true"
    And the response should contain the field "ai_access.skip_approvals" equal to "false"

  @bounds
  Scenario Outline: Limits are accepted - <field> <value>
    When I call "SettingsService.Save" with
      """
      { "<field>": <value> }
      """
    Then the call should succeed
    And the response should contain the field "<field>" equal to "<value>"

    Examples:
      | field            | value |
      | history_user_cap | 100   |
      | history_user_cap | 20000 |
      | history_ai_cap   | 100   |
      | history_ai_cap   | 20000 |

  @out_of_range
  Scenario Outline: Out-of-range values are refused and nothing is stored - <case>
    When I call "SettingsService.Save" with
      """
      <payload>
      """
    Then the call should fail with code "<code>"
    When I call "SettingsService.Get"
    Then the response should contain the field "history_user_cap" equal to "1000"
    And the response should contain the field "history_ai_cap" equal to "1000"
    And the response should contain the field "script_timeout_ms" equal to "10000"

    Examples:
      | case             | payload                                                                           | code      |
      | user cap low     | { "history_user_cap": 99 }                                                        | SET-02400 |
      | user cap high    | { "history_user_cap": 20001 }                                                     | SET-02400 |
      | ai cap low       | { "history_ai_cap": 99 }                                                          | SET-03400 |
      | ai cap high      | { "history_ai_cap": 20001 }                                                       | SET-03400 |
      | port low         | { "ai_access": { "enabled": true, "port": 1023, "skip_approvals": false } }       | SET-04400 |
      | port high        | { "ai_access": { "enabled": true, "port": 65536, "skip_approvals": false } }      | SET-04400 |
      | first bad wins   | { "script_timeout_ms": 999, "history_user_cap": 99 }                              | SCR-01400 |
      | valid with a bad | { "script_timeout_ms": 2500, "history_ai_cap": 5 }                                | SET-03400 |

  @apply_now
  Scenario: Saving AI access applies it at once and tells the window
    Given the emitted events are forgotten
    When I call "SettingsService.Save" with
      """
      { "ai_access": { "enabled": false, "port": {{mcp_port}}, "skip_approvals": false } }
      """
    Then the call should succeed
    And an "ai:access-status-changed" event with the field "state" equal to "off" should have been emitted within 2000 ms
    And the MCP endpoint should refuse connections
