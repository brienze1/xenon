#language: en
#utf-8

@all @xen_004 @scripts @settings_timeout
Feature: The script timeout setting
  The time a single script may run (including waiting for its timers and requests) is an app
  setting stored in app.json as settings.scriptTimeoutMs: 10000 ms by default, 1000 to 60000 ms.

  @default
  Scenario: The timeout is 10 seconds until it is changed
    When I call "SettingsService.Get"
    Then the call should succeed
    And the response should contain the field "script_timeout_ms" equal to "10000"

  @save
  Scenario: A saved timeout is stored in app.json and survives a restart
    Given the workspace "Acme" exists
    When I call "SettingsService.Save" with
      """
      { "script_timeout_ms": 2500 }
      """
    Then the call should succeed
    And the response should contain the field "script_timeout_ms" equal to "2500"
    And the app data file "app.json" should contain the field "settings.scriptTimeoutMs" equal to "2500"
    Given the app is restarted
    When I call "SettingsService.Get"
    Then the response should contain the field "script_timeout_ms" equal to "2500"
    Given the workspace "Beta" exists
    When I call "WorkspaceService.List"
    Then the call should succeed
    And the response should contain the field "workspaces" array length equal to 2
    And the app data file "app.json" should contain the field "settings.scriptTimeoutMs" equal to "2500"

  @bounds
  Scenario Outline: The limits are accepted - <value>
    When I call "SettingsService.Save" with
      """
      { "script_timeout_ms": <value> }
      """
    Then the call should succeed
    And the response should contain the field "script_timeout_ms" equal to "<value>"

    Examples:
      | value |
      | 1000  |
      | 60000 |

  @out_of_range
  Scenario Outline: Values outside 1-60 seconds are refused and nothing is stored - <value>
    When I call "SettingsService.Save" with
      """
      { "script_timeout_ms": <value> }
      """
    Then the call should fail with code "SCR-01400"
    When I call "SettingsService.Get"
    Then the response should contain the field "script_timeout_ms" equal to "10000"

    Examples:
      | value |
      | 999   |
      | 60001 |
      | 0     |

  @invalid_file
  Scenario: An app.json timeout outside the range falls back to the default
    Given the app data file "app.json" contains
      """
      { "version": 1, "workspaces": [], "settings": { "scriptTimeoutMs": 5 } }
      """
    When I call "SettingsService.Get"
    Then the call should succeed
    And the response should contain the field "script_timeout_ms" equal to "10000"
