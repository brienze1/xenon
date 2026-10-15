#language: en
#utf-8

@all @xen_013 @updates @install_update
Feature: Install an update
  Update downloads the package for this system from the URL the release lists, streams it into
  <app-data>/updates while hashing it, and accepts it only when its SHA-256 matches the release's
  checksums.txt; it then reports "ready". Restart replaces the installed app as a whole (the .app
  bundle, the .exe or the executable) with two renames in its own folder, restores the previous
  app on any failure, starts the new executable with XENON_RESTART_AFTER_PID and quits. Nothing is
  installed from an unverified package, an unsafe archive, an unwritable folder or a build that is
  not packaged, and nothing is ever downloaded from a URL the release does not list.

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
          { "name": "xenon-0.1.7-darwin-universal.tar.gz", "size": 261, "browser_download_url": "{{api}}/download/v0.1.7/xenon-0.1.7-darwin-universal.tar.gz" },
          { "name": "xenon-0.1.7-linux-amd64.tar.gz", "size": 112, "browser_download_url": "{{api}}/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" },
          { "name": "xenon-0.1.7-windows-amd64.zip", "size": 128, "browser_download_url": "{{api}}/download/v0.1.7/xenon-0.1.7-windows-amd64.zip" },
          { "name": "xenon-0.1.7-evil-amd64.tar.gz", "size": 109, "browser_download_url": "{{api}}/download/v0.1.7/xenon-0.1.7-evil-amd64.tar.gz" },
          { "name": "xenon-0.1.7-multi-amd64.tar.gz", "size": 140, "browser_download_url": "{{api}}/download/v0.1.7/xenon-0.1.7-multi-amd64.tar.gz" },
          { "name": "checksums.txt", "size": 488, "browser_download_url": "{{api}}/download/v0.1.7/checksums.txt" }
        ]
      }
      """
    And the 1 "GET" request to "/download/v0.1.7/checksums.txt" returns status 200 with content type "text/plain" and the resource file "updates/checksums.txt"
    And the 1 "GET" request to "/download/v0.1.7/xenon-0.1.7-darwin-universal.tar.gz" returns status 200 with content type "application/gzip" and the resource file "updates/xenon-0.1.7-darwin-universal.tar.gz"
    And the 1 "GET" request to "/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" returns status 200 with content type "application/gzip" and the resource file "updates/xenon-0.1.7-linux-amd64.tar.gz"
    And the 1 "GET" request to "/download/v0.1.7/xenon-0.1.7-windows-amd64.zip" returns status 200 with content type "application/zip" and the resource file "updates/xenon-0.1.7-windows-amd64.zip"
    And the 1 "GET" request to "/download/v0.1.7/xenon-0.1.7-evil-amd64.tar.gz" returns status 200 with content type "application/gzip" and the resource file "updates/xenon-0.1.7-evil-amd64.tar.gz"
    And the 1 "GET" request to "/download/v0.1.7/xenon-0.1.7-multi-amd64.tar.gz" returns status 200 with content type "application/gzip" and the resource file "updates/xenon-0.1.7-multi-amd64.tar.gz"

  @download @success
  Scenario: Update downloads and verifies the package, then reports ready without restarting
    Given I call "UpdateService.Check"
    When I call "UpdateService.Update"
    Then the call should succeed
    And the response should contain the field "state" equal to "ready"
    And the response should contain the field "automatic" equal to "false"
    And the response should contain the field "latest_version" equal to "0.1.7"
    And the app data file "updates/xenon-0.1.7-linux-amd64.tar.gz" should exist
    And the app data file "updates/xenon-0.1.7-linux-amd64.tar.gz.part" should not exist
    And an "update:status" event with the field "state" equal to "downloading" should have been emitted within 100 ms
    And an "update:status" event with the field "received" equal to "112" should have been emitted within 100 ms
    And the last "update:status" event should contain the field "state" equal to "ready"
    And the last "update:status" event should contain the field "total" equal to "112"
    And the "GET" request to "/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" should have been received 1 times
    And the "GET" request to "/download/v0.1.7/checksums.txt" should have been received 1 times
    And the "GET" request to "/download/v0.1.7/xenon-0.1.7-windows-amd64.zip" should have been received 0 times
    And the file "apps/xenon" should equal the resource file "updates/expected/xenon-0.1.4.txt"
    And the app should have spawned 0 processes
    And the app should have quit 0 times

  @install @success
  Scenario Outline: Restart swaps a single-file app and starts the new one - <platform>
    Given the "XENON_UPDATE_PLATFORM" env var is set to "<platform>"
    And the "XENON_INSTALL_PATH" env var is set to "{{temp_root}}/apps/<executable>"
    And the file "apps/<executable>" exists with content
      """
      xenon 0.1.4
      """
    And I call "UpdateService.Check"
    And I call "UpdateService.Update"
    When I call "UpdateService.Restart"
    Then the call should succeed
    And the file "apps/<executable>" should equal the resource file "updates/expected/xenon-0.1.7.txt"
    And the file "apps/.<executable>.xenon-update" should not exist
    And the app should have spawned 1 process
    And the 1 spawned process should run "{{temp_root}}/apps/<executable>"
    And the 1 spawned process should have the env var "XENON_RESTART_AFTER_PID" matching "^[0-9]+$"
    And the app should have quit 1 time
    And an "update:status" event with the field "state" equal to "installing" should have been emitted within 100 ms

    Examples:
      | platform      | executable |
      | linux-amd64   | xenon      |
      | windows-amd64 | xenon.exe  |

  @install @success @bundle
  Scenario: Restart replaces a macOS bundle as a whole and starts the executable inside it
    Given the "XENON_UPDATE_PLATFORM" env var is set to "darwin-universal"
    And the "XENON_INSTALL_PATH" env var is set to "{{temp_root}}/apps/Xenon.app"
    And the file "apps/Xenon.app/Contents/MacOS/xenon" exists with content
      """
      xenon 0.1.4
      """
    And the file "apps/Xenon.app/Contents/Resources/only-in-0.1.4.txt" exists with content
      """
      removed by the update
      """
    And I call "UpdateService.Check"
    And I call "UpdateService.Update"
    When I call "UpdateService.Restart"
    Then the call should succeed
    And the file "apps/Xenon.app/Contents/MacOS/xenon" should equal the resource file "updates/expected/xenon-0.1.7.txt"
    And the file "apps/Xenon.app/Contents/Info.plist" should exist
    And the file "apps/Xenon.app/Contents/Resources/only-in-0.1.4.txt" should not exist
    And the file "apps/.Xenon.app.xenon-old" should not exist
    And the file "apps/.Xenon.app.xenon-update" should not exist
    And the app should have spawned 1 process
    And the 1 spawned process should run "{{temp_root}}/apps/Xenon.app/Contents/MacOS/xenon"
    And the app should have quit 1 time

  @install @renamed_bundle
  Scenario: A bundle installed under another name keeps its name
    Given the "XENON_UPDATE_PLATFORM" env var is set to "darwin-universal"
    And the "XENON_INSTALL_PATH" env var is set to "{{temp_root}}/apps/My Xenon.app"
    And the file "apps/My Xenon.app/Contents/MacOS/xenon" exists with content
      """
      xenon 0.1.4
      """
    And I call "UpdateService.Check"
    And I call "UpdateService.Update"
    When I call "UpdateService.Restart"
    Then the file "apps/My Xenon.app/Contents/MacOS/xenon" should equal the resource file "updates/expected/xenon-0.1.7.txt"
    And the file "apps/Xenon.app" should not exist
    And the 1 spawned process should run "{{temp_root}}/apps/My Xenon.app/Contents/MacOS/xenon"

  @listed_urls
  Scenario: Packages are downloaded only from the URLs the release lists, following redirects
    Given the 1 "GET" request to "/releases/latest" returns status 200 with the following response
      """
      { "tag_name": "v0.1.7", "html_url": "https://github.com/brienze1/xenon/releases/tag/v0.1.7", "draft": false, "prerelease": false,
        "assets": [
          { "name": "xenon-0.1.7-linux-amd64.tar.gz", "size": 112, "browser_download_url": "{{api}}/cdn/a1/xenon-0.1.7-linux-amd64.tar.gz" },
          { "name": "checksums.txt", "size": 488, "browser_download_url": "{{api}}/cdn/a2/checksums.txt" }
        ] }
      """
    And the requests to "/cdn/a1/xenon-0.1.7-linux-amd64.tar.gz" redirect to "{{api}}/objects/linux" with status 302
    And the 1 "GET" request to "/objects/linux" returns status 200 with content type "application/octet-stream" and the resource file "updates/xenon-0.1.7-linux-amd64.tar.gz"
    And the 1 "GET" request to "/cdn/a2/checksums.txt" returns status 200 with content type "text/plain" and the resource file "updates/checksums.txt"
    And I call "UpdateService.Check"
    When I call "UpdateService.Update"
    Then the response should contain the field "state" equal to "ready"
    And the "GET" request to "/objects/linux" should have been received 1 times
    And the "GET" request to "/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" should have been received 0 times
    And the "GET" request to "/download/v0.1.7/checksums.txt" should have been received 0 times

  @unverified
  Scenario Outline: A package that cannot be verified is discarded and nothing changes - <case>
    Given the 1 "GET" request to "/download/v0.1.7/checksums.txt" returns status 200 with content type "text/plain" and body
      """
      <checksums>
      """
    And I call "UpdateService.Check"
    When I call "UpdateService.Update"
    Then the call should succeed
    And the response should contain the field "state" equal to "failed"
    And the response should contain the field "reason" equal to "unverified"
    And the app data file "updates/xenon-0.1.7-linux-amd64.tar.gz" should not exist
    And the app data file "updates/xenon-0.1.7-linux-amd64.tar.gz.part" should not exist
    And the file "apps/xenon" should equal the resource file "updates/expected/xenon-0.1.4.txt"
    When I call "UpdateService.Restart"
    Then the response should contain the field "state" equal to "failed"
    And the app should have spawned 0 processes
    And the app should have quit 0 times

    Examples:
      | case          | checksums                                                                                                     |
      | wrong digest  | 0000000000000000000000000000000000000000000000000000000000000000  xenon-0.1.7-linux-amd64.tar.gz                |
      | no line       | 32f96a8c6c131cb8645431888fc676e51ae9dc951434f40615cdfab9f35f090f  xenon-0.1.7-windows-amd64.zip                 |
      | empty file    |                                                                                                               |

  @unverified
  Scenario: A release without a checksums file is never installed
    Given the 1 "GET" request to "/releases/latest" returns status 200 with the following response
      """
      { "tag_name": "v0.1.7", "html_url": "https://github.com/brienze1/xenon/releases/tag/v0.1.7", "draft": false, "prerelease": false,
        "assets": [ { "name": "xenon-0.1.7-linux-amd64.tar.gz", "size": 112, "browser_download_url": "{{api}}/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" } ] }
      """
    And I call "UpdateService.Check"
    When I call "UpdateService.Update"
    Then the response should contain the field "state" equal to "failed"
    And the response should contain the field "reason" equal to "unverified"
    And the file "apps/xenon" should equal the resource file "updates/expected/xenon-0.1.4.txt"

  @download_failed
  Scenario: A failed download leaves the app untouched and can be retried
    Given the 1 "GET" request to "/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" returns status 500 with content type "text/plain" and body
      """
      upstream error
      """
    And the 2 "GET" request to "/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" returns status 200 with content type "application/gzip" and the resource file "updates/xenon-0.1.7-linux-amd64.tar.gz"
    And the 2 "GET" request to "/download/v0.1.7/checksums.txt" returns status 200 with content type "text/plain" and the resource file "updates/checksums.txt"
    And I call "UpdateService.Check"
    When I call "UpdateService.Update"
    Then the response should contain the field "state" equal to "failed"
    And the response should contain the field "reason" equal to "download_failed"
    And the app data file "updates/xenon-0.1.7-linux-amd64.tar.gz.part" should not exist
    And the file "apps/xenon" should equal the resource file "updates/expected/xenon-0.1.4.txt"
    When I call "UpdateService.Update"
    Then the response should contain the field "state" equal to "ready"

  @download_failed
  Scenario: An unreachable download host fails the download
    Given the "closed" variable is set to a closed port url
    And the 1 "GET" request to "/releases/latest" returns status 200 with the following response
      """
      { "tag_name": "v0.1.7", "html_url": "https://github.com/brienze1/xenon/releases/tag/v0.1.7", "draft": false, "prerelease": false,
        "assets": [
          { "name": "xenon-0.1.7-linux-amd64.tar.gz", "size": 112, "browser_download_url": "{{closed}}/xenon-0.1.7-linux-amd64.tar.gz" },
          { "name": "checksums.txt", "size": 488, "browser_download_url": "{{api}}/download/v0.1.7/checksums.txt" }
        ] }
      """
    And I call "UpdateService.Check"
    When I call "UpdateService.Update"
    Then the response should contain the field "state" equal to "failed"
    And the response should contain the field "reason" equal to "download_failed"

  @install_failed
  Scenario Outline: An archive with unexpected content is refused and the app stays - <case>
    Given the "XENON_UPDATE_PLATFORM" env var is set to "<platform>"
    And I call "UpdateService.Check"
    And I call "UpdateService.Update"
    When I call "UpdateService.Restart"
    Then the call should succeed
    And the response should contain the field "state" equal to "failed"
    And the response should contain the field "reason" equal to "install_failed"
    And the file "apps/xenon" should equal the resource file "updates/expected/xenon-0.1.4.txt"
    And the file "escape" should not exist
    And the file "apps/escape" should not exist
    And the file "apps/.xenon.xenon-update" should not exist
    And the file "apps/.xenon.xenon-old" should not exist
    And the app should have spawned 0 processes
    And the app should have quit 0 times

    Examples:
      | case                 | platform    |
      | path traversal       | evil-amd64  |
      | two top-level entries | multi-amd64 |

  @not_writable
  Scenario: An install folder Xenon cannot write is refused without asking for rights
    Given the file "blocker.txt" exists with content
      """
      a file where the install folder should be
      """
    And the "XENON_INSTALL_PATH" env var is set to "{{temp_root}}/blocker.txt/xenon"
    And I call "UpdateService.Check"
    And I call "UpdateService.Update"
    When I call "UpdateService.Restart"
    Then the call should succeed
    And the response should contain the field "state" equal to "failed"
    And the response should contain the field "reason" equal to "not_writable"
    And the response should contain the path field "folder" equal to "{{temp_root}}/blocker.txt"
    And the file "blocker.txt" should exist
    And the app should have spawned 0 processes
    And the app should have quit 0 times

  @not_packaged
  Scenario: A build that is not packaged refuses to install
    Given the "XENON_INSTALL_PATH" env var is set to ""
    And I call "UpdateService.Check"
    And I call "UpdateService.Update"
    When I call "UpdateService.Restart"
    Then the call should succeed
    And the response should contain the field "state" equal to "failed"
    And the response should contain the field "reason" equal to "not_packaged"
    And the file "apps/xenon" should equal the resource file "updates/expected/xenon-0.1.4.txt"
    And the app should have spawned 0 processes

  @state_guard
  Scenario: Update and Restart before an update is available change nothing
    When I call "UpdateService.Update"
    Then the call should succeed
    And the response should contain the field "state" equal to "idle"
    When I call "UpdateService.Restart"
    Then the response should contain the field "state" equal to "idle"
    And the "GET" request to "/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" should have been received 0 times
    And the app should have spawned 0 processes

  @busy
  Scenario: A second Update while downloading returns the current status without a second download
    Given the "GET" requests to "/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" respond after 700ms
    And I call "UpdateService.Check"
    When I start calling "UpdateService.Update" in the background with
      """
      {}
      """
    Then an "update:status" event with the field "state" equal to "downloading" should have been emitted within 1000 ms
    When I call "UpdateService.Update"
    Then the response should contain the field "state" equal to "downloading"
    And the background call should finish within 3000 ms
    And the "GET" request to "/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" should have been received 1 times

  @leftovers
  Scenario: Leftovers of an interrupted update are removed at the next start
    Given I call "SettingsService.Save" with
      """
      { "auto_update": false }
      """
    And the file "apps/.xenon.xenon-old" exists with content
      """
      previous app
      """
    And the file "apps/.xenon.xenon-update/xenon" exists with content
      """
      half extracted
      """
    And the app data file "updates/xenon-0.1.6-linux-amd64.tar.gz" contains the text "verified but never installed"
    And the app data file "updates/xenon-0.1.7-linux-amd64.tar.gz.part" contains the text "partial"
    When I call "UpdateService.CheckOnStartup"
    Then the call should succeed
    And the file "apps/.xenon.xenon-old" should not exist
    And the file "apps/.xenon.xenon-update" should not exist
    And the app data file "updates/xenon-0.1.6-linux-amd64.tar.gz" should not exist
    And the app data file "updates/xenon-0.1.7-linux-amd64.tar.gz.part" should not exist
    And the file "apps/xenon" should equal the resource file "updates/expected/xenon-0.1.4.txt"

  @logs
  Scenario: Download, verification and install are logged with file names and versions only
    Given the log output is captured
    And I call "UpdateService.Check"
    And I call "UpdateService.Update"
    When I call "UpdateService.Restart"
    Then the log output should contain "xenon-0.1.7-linux-amd64.tar.gz"
    And the log output should contain "0.1.7"
