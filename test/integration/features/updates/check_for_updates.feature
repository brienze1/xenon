#language: en
#utf-8

@all @xen_013 @updates @check_for_updates
Feature: Check for updates
  Settings → About & Updates shows the running version and, after Check for updates, the latest
  published release of brienze1/xenon from GitHub's releases/latest. An update is offered only
  when that release is strictly newer and carries the package for this system
  (xenon-<version>-<platform>.<tar.gz|zip>). Any check that cannot complete — no network, an error
  status, a rate limit, an unreadable answer, a version that is not N.N.N, a draft or prerelease —
  ends quietly as "check_failed" with the reason "unreachable", never as an error. One operation
  runs at a time, and every state change is published as an "update:status" event.

  Background:
    Given the "api" variable is set to the mocked api url
    And the "XENON_UPDATE_API_URL" env var is set to "{{api}}"
    And the "XENON_VERSION" env var is set to "0.1.4"
    And the "XENON_UPDATE_PLATFORM" env var is set to "linux-amd64"
    And the system browser ignores the pages it opens
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
          { "name": "checksums.txt", "size": 488, "browser_download_url": "{{api}}/download/v0.1.7/checksums.txt" }
        ]
      }
      """

  @idle
  Scenario: Before any check the status is idle with the running version
    When I call "UpdateService.GetStatus"
    Then the call should succeed
    And the response should contain the field "state" equal to "idle"
    And the response should contain the field "current_version" equal to "0.1.4"
    And the response should contain the field "latest_version" equal to ""
    And the "GET" request to "/releases/latest" should have been received 0 times

  @available
  Scenario: A newer release with a package for this system offers an update
    When I call "UpdateService.Check"
    Then the call should succeed
    And the response should contain the field "state" equal to "available"
    And the response should contain the field "current_version" equal to "0.1.4"
    And the response should contain the field "latest_version" equal to "0.1.7"
    And the response should contain the field "release_url" equal to "https://github.com/brienze1/xenon/releases/tag/v0.1.7"
    And the response should contain the field "reason" equal to ""
    And the 1 "GET" request to "/releases/latest" should have the header "Accept" equal to "application/vnd.github+json"
    And the 1 "update:status" event should contain the field "state" equal to "checking"
    And the last "update:status" event should contain the field "state" equal to "available"
    And the last "update:status" event should contain the field "latest_version" equal to "0.1.7"
    And the "GET" request to "/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" should have been received 0 times
    When I call "UpdateService.GetStatus"
    Then the response should contain the field "state" equal to "available"

  @up_to_date
  Scenario Outline: A release that is not newer is up to date (never a downgrade) - running <running>
    Given the "XENON_VERSION" env var is set to "<running>"
    When I call "UpdateService.Check"
    Then the call should succeed
    And the response should contain the field "state" equal to "up_to_date"
    And the response should contain the field "latest_version" equal to "0.1.7"
    And the response should contain the field "current_version" equal to "<running>"

    Examples:
      | running |
      | 0.1.7   |
      | 0.1.10  |
      | 0.2.0   |
      | 1.0.0   |

  @numeric_order
  Scenario: Versions compare numerically, not as text
    Given the "XENON_VERSION" env var is set to "0.1.9"
    And the 1 "GET" request to "/releases/latest" returns status 200 with the following response
      """
      { "tag_name": "v0.1.10", "html_url": "https://github.com/brienze1/xenon/releases/tag/v0.1.10", "draft": false, "prerelease": false,
        "assets": [ { "name": "xenon-0.1.10-linux-amd64.tar.gz", "size": 112, "browser_download_url": "{{api}}/download/v0.1.10/xenon-0.1.10-linux-amd64.tar.gz" } ] }
      """
    When I call "UpdateService.Check"
    Then the response should contain the field "state" equal to "available"
    And the response should contain the field "latest_version" equal to "0.1.10"

  @no_package
  Scenario: A newer release without a package for this system offers no update
    Given the "XENON_UPDATE_PLATFORM" env var is set to "linux-arm64"
    When I call "UpdateService.Check"
    Then the call should succeed
    And the response should contain the field "state" equal to "no_package"
    And the response should contain the field "latest_version" equal to "0.1.7"
    And the response should contain the field "release_url" equal to "https://github.com/brienze1/xenon/releases/tag/v0.1.7"

  @platforms
  Scenario Outline: The package is picked by platform - <platform>
    Given the "XENON_UPDATE_PLATFORM" env var is set to "<platform>"
    When I call "UpdateService.Check"
    Then the response should contain the field "state" equal to "available"

    Examples:
      | platform         |
      | darwin-universal |
      | windows-amd64    |
      | linux-amd64      |

  @check_failed
  Scenario Outline: A check that cannot complete fails quietly - status <status>
    Given the 1 "GET" request to "/releases/latest" returns status <status> with the following response
      """
      { "message": "API rate limit exceeded for 203.0.113.7." }
      """
    When I call "UpdateService.Check"
    Then the call should succeed
    And the response should contain the field "state" equal to "check_failed"
    And the response should contain the field "reason" equal to "unreachable"
    And the response should contain the field "current_version" equal to "0.1.4"
    And the last "update:status" event should contain the field "state" equal to "check_failed"

    Examples:
      | status |
      | 403    |
      | 404    |
      | 429    |
      | 500    |
      | 503    |

  @check_failed
  Scenario Outline: An unusable answer fails quietly - <case>
    Given the 1 "GET" request to "/releases/latest" returns status 200 with content type "application/json" and body
      """
      <body>
      """
    When I call "UpdateService.Check"
    Then the call should succeed
    And the response should contain the field "state" equal to "check_failed"
    And the response should contain the field "reason" equal to "unreachable"

    Examples:
      | case               | body                                                                                                           |
      | not json           | <html>GitHub is down</html>                                                                                    |
      | no tag             | { "html_url": "https://github.com/brienze1/xenon/releases/tag/v0.1.7", "assets": [] }                          |
      | prerelease suffix  | { "tag_name": "v0.1.7-beta", "html_url": "https://github.com/brienze1/xenon/releases", "assets": [] }          |
      | not a version      | { "tag_name": "nightly", "html_url": "https://github.com/brienze1/xenon/releases", "assets": [] }              |
      | prerelease release | { "tag_name": "v0.1.7", "html_url": "https://github.com/brienze1/xenon/releases", "prerelease": true, "assets": [] } |
      | draft release      | { "tag_name": "v0.1.7", "html_url": "https://github.com/brienze1/xenon/releases", "draft": true, "assets": [] }      |

  @offline
  Scenario: No network fails quietly
    Given the "closed" variable is set to a closed port url
    And the "XENON_UPDATE_API_URL" env var is set to "{{closed}}"
    When I call "UpdateService.Check"
    Then the call should succeed
    And the response should contain the field "state" equal to "check_failed"
    And the response should contain the field "reason" equal to "unreachable"

  @recheck
  Scenario: A failed check can be retried
    Given the 1 "GET" request to "/releases/latest" returns status 503 with the following response
      """
      {}
      """
    And the 2 "GET" request to "/releases/latest" returns status 200 with the following response
      """
      { "tag_name": "v0.1.7", "html_url": "https://github.com/brienze1/xenon/releases/tag/v0.1.7", "draft": false, "prerelease": false,
        "assets": [ { "name": "xenon-0.1.7-linux-amd64.tar.gz", "size": 112, "browser_download_url": "{{api}}/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" } ] }
      """
    When I call "UpdateService.Check"
    Then the response should contain the field "state" equal to "check_failed"
    When I call "UpdateService.Check"
    Then the response should contain the field "state" equal to "available"
    And the "GET" request to "/releases/latest" should have been received 2 times

  @busy
  Scenario: A check while another runs returns the current status without a second request
    Given the "GET" requests to "/releases/latest" respond after 700ms
    When I start calling "UpdateService.Check" in the background with
      """
      {}
      """
    Then an "update:status" event with the field "state" equal to "checking" should have been emitted within 1000 ms
    When I call "UpdateService.GetStatus"
    Then the response should contain the field "state" equal to "checking"
    When I call "UpdateService.Check"
    Then the call should succeed
    And the response should contain the field "state" equal to "checking"
    And the background call should finish within 3000 ms
    And the "GET" request to "/releases/latest" should have been received 1 times

  @release_page
  Scenario: What's new opens the latest release page in the system browser
    Given I call "UpdateService.Check"
    When I call "UpdateService.OpenReleasePage"
    Then the call should succeed
    And the system browser should have opened 1 page
    And the 1 page opened in the system browser should start with "https://github.com/brienze1/xenon/releases/tag/v0.1.7"

  @release_page
  Scenario: Without a known release nothing is opened
    When I call "UpdateService.OpenReleasePage"
    Then the call should succeed
    And the response should contain the field "state" equal to "idle"
    And the system browser should have opened 0 pages

  @logs
  Scenario: The check result is logged with versions only
    Given the log output is captured
    When I call "UpdateService.Check"
    Then the log output should contain "0.1.7"
