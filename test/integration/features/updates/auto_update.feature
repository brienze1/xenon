#language: en
#utf-8

@all @xen_013 @updates @auto_update
Feature: Updates arrive by themselves
  The UI calls CheckOnStartup once the window is usable. It removes leftovers and checks GitHub
  once per app run, whatever the setting. With "Update automatically" on (the default, stored as
  settings.autoUpdate in app.json, absent = on) an offered update is downloaded and verified in the
  background and an "update:status" event reports "ready" with automatic true (the UI shows the
  Restart now / Later toast). With it off nothing is downloaded and the status stays "available"
  (the UI shows the footer notice). Restart installs and restarts; a normal shutdown with a ready
  update installs it without starting anything. Failures stay quiet.

  Background:
    Given the "api" variable is set to the mocked api url
    And the "XENON_UPDATE_API_URL" env var is set to "{{api}}"
    And the "XENON_VERSION" env var is set to "0.1.4"
    And the "XENON_UPDATE_PLATFORM" env var is set to "linux-amd64"
    And the "XENON_INSTALL_PATH" env var is set to "{{temp_root}}/apps/xenon"
    And the file "apps/xenon" exists with content
      """
      xenon 0.1.4
      """
    And the 1 "GET" request to "/releases/latest" returns status 200 with the following response
      """
      {
        "tag_name": "v0.1.7",
        "html_url": "https://github.com/brienze1/xenon/releases/tag/v0.1.7",
        "draft": false,
        "prerelease": false,
        "assets": [
          { "name": "xenon-0.1.7-linux-amd64.tar.gz", "size": 112, "browser_download_url": "{{api}}/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" },
          { "name": "checksums.txt", "size": 488, "browser_download_url": "{{api}}/download/v0.1.7/checksums.txt" }
        ]
      }
      """
    And the 1 "GET" request to "/download/v0.1.7/checksums.txt" returns status 200 with content type "text/plain" and the resource file "updates/checksums.txt"
    And the 1 "GET" request to "/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" returns status 200 with content type "application/gzip" and the resource file "updates/xenon-0.1.7-linux-amd64.tar.gz"

  @setting @defaults
  Scenario: Update automatically is on when nothing was stored
    When I call "SettingsService.Get"
    Then the call should succeed
    And the response should contain the field "auto_update" equal to "true"

  @setting @old_app_json
  Scenario: An app.json saved before this feature reads Update automatically as on
    Given the app data file "app.json" contains
      """
      { "settings": { "scriptTimeoutMs": 2500, "historyUserCap": 1000, "historyAiCap": 1000 } }
      """
    When I call "SettingsService.Get"
    Then the response should contain the field "auto_update" equal to "true"
    And the response should contain the field "script_timeout_ms" equal to "2500"

  @setting @save
  Scenario: Turning it off is stored per machine and survives a restart
    When I call "SettingsService.Save" with
      """
      { "auto_update": false }
      """
    Then the call should succeed
    And the response should contain the field "auto_update" equal to "false"
    And the app data file "app.json" should contain the field "settings.autoUpdate" equal to "false"
    Given the app is restarted
    When I call "SettingsService.Get"
    Then the response should contain the field "auto_update" equal to "false"
    When I call "SettingsService.Save" with
      """
      { "script_timeout_ms": 3000 }
      """
    Then the response should contain the field "auto_update" equal to "false"

  @automatic @success
  Scenario: With the setting on, the startup check downloads and verifies in the background
    When I call "UpdateService.CheckOnStartup"
    Then the call should succeed
    And the response should contain the field "state" equal to "downloading"
    And the response should contain the field "latest_version" equal to "0.1.7"
    And an "update:status" event with the field "state" equal to "ready" should have been emitted within 5000 ms
    And the last "update:status" event should contain the field "automatic" equal to "true"
    And the last "update:status" event should contain the field "latest_version" equal to "0.1.7"
    And the app data file "updates/xenon-0.1.7-linux-amd64.tar.gz" should exist
    And the file "apps/xenon" should equal the resource file "updates/expected/xenon-0.1.4.txt"
    And the app should have spawned 0 processes
    And the app should have quit 0 times

  @automatic @restart_now
  Scenario: Restart now installs the downloaded update and restarts
    Given I call "UpdateService.CheckOnStartup"
    And an "update:status" event with the field "state" equal to "ready" should have been emitted within 5000 ms
    When I call "UpdateService.Restart"
    Then the call should succeed
    And the file "apps/xenon" should equal the resource file "updates/expected/xenon-0.1.7.txt"
    And the app should have spawned 1 process
    And the 1 spawned process should run "{{temp_root}}/apps/xenon"
    And the app should have quit 1 time

  @automatic @later
  Scenario: After Later, a normal shutdown installs the update without starting anything
    Given I call "UpdateService.CheckOnStartup"
    And an "update:status" event with the field "state" equal to "ready" should have been emitted within 5000 ms
    When the app shuts down
    Then the file "apps/xenon" should equal the resource file "updates/expected/xenon-0.1.7.txt"
    And the app should have spawned 0 processes
    And the app should have quit 0 times

  @shutdown_without_update
  Scenario: A shutdown without a ready update changes nothing
    Given I call "SettingsService.Save" with
      """
      { "auto_update": false }
      """
    And I call "UpdateService.CheckOnStartup"
    When the app shuts down
    Then the file "apps/xenon" should equal the resource file "updates/expected/xenon-0.1.4.txt"
    And the app should have spawned 0 processes

  @manual_only
  Scenario: With the setting off, the startup check only reports the update
    Given I call "SettingsService.Save" with
      """
      { "auto_update": false }
      """
    When I call "UpdateService.CheckOnStartup"
    Then the call should succeed
    And the response should contain the field "state" equal to "available"
    And the response should contain the field "latest_version" equal to "0.1.7"
    And the "GET" request to "/releases/latest" should have been received 1 times
    And the "GET" request to "/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" should have been received 0 times
    And the app data file "updates/xenon-0.1.7-linux-amd64.tar.gz" should not exist

  @up_to_date
  Scenario: Nothing is downloaded when Xenon is up to date
    Given the "XENON_VERSION" env var is set to "0.1.7"
    When I call "UpdateService.CheckOnStartup"
    Then the response should contain the field "state" equal to "up_to_date"
    And the "GET" request to "/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" should have been received 0 times

  @quiet_failure
  Scenario: A startup check that fails stays quiet
    Given the 1 "GET" request to "/releases/latest" returns status 429 with the following response
      """
      { "message": "API rate limit exceeded" }
      """
    When I call "UpdateService.CheckOnStartup"
    Then the call should succeed
    And the response should contain the field "state" equal to "check_failed"
    And the response should contain the field "reason" equal to "unreachable"
    And the "GET" request to "/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" should have been received 0 times

  @quiet_failure
  Scenario: A background download that cannot be verified ends as a failure, never as ready
    Given the 1 "GET" request to "/download/v0.1.7/checksums.txt" returns status 200 with content type "text/plain" and body
      """
      0000000000000000000000000000000000000000000000000000000000000000  xenon-0.1.7-linux-amd64.tar.gz
      """
    When I call "UpdateService.CheckOnStartup"
    Then an "update:status" event with the field "state" equal to "failed" should have been emitted within 5000 ms
    And the last "update:status" event should contain the field "reason" equal to "unverified"
    And the last "update:status" event should contain the field "automatic" equal to "true"
    And the app data file "updates/xenon-0.1.7-linux-amd64.tar.gz" should not exist
    And the file "apps/xenon" should equal the resource file "updates/expected/xenon-0.1.4.txt"

  @once_per_run
  Scenario: The startup check runs once per app run
    Given I call "SettingsService.Save" with
      """
      { "auto_update": false }
      """
    And I call "UpdateService.CheckOnStartup"
    When I call "UpdateService.CheckOnStartup"
    Then the call should succeed
    And the response should contain the field "state" equal to "available"
    And the "GET" request to "/releases/latest" should have been received 1 times
    Given the 2 "GET" request to "/releases/latest" returns status 200 with the following response
      """
      { "tag_name": "v0.1.8", "html_url": "https://github.com/brienze1/xenon/releases/tag/v0.1.8", "draft": false, "prerelease": false, "assets": [] }
      """
    And the app is restarted
    When I call "UpdateService.CheckOnStartup"
    Then the response should contain the field "latest_version" equal to "0.1.8"
    And the "GET" request to "/releases/latest" should have been received 2 times
