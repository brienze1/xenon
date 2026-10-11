#language: en
#utf-8

@all @xen_005 @cookies @persistence_encrypted
Feature: Cookies survive restarts encrypted with the workspace vault key
  Persistent cookies are written to workspaces/<localId>/cookies.age in the app data folder,
  encrypted to the workspace's vault identity, so no cookie value is ever on disk in clear text.
  Session cookies (no expiry) last until the app quits. While the vault is locked the cookies stay
  in memory and are written once it is unlocked.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the requests to "/login" respond with the header "Set-Cookie" equal to "remember=r3m3mb3r-value; Path=/; Max-Age=86400"
    And the requests to "/login" respond with the header "Set-Cookie" equal to "session=s3ss10n-value; Path=/"

  @restart
  Scenario: Persistent cookies survive a restart and session cookies do not
    Given the keychain is available
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/cookies.age" should exist
    And the app data path "workspaces/{{Acme.local_id}}/cookies.age" should have the permissions "600"
    And no file under the app data folder should contain "r3m3mb3r-value"
    And no file under the app data folder should contain "s3ss10n-value"
    And no file under the workspace folder should contain "r3m3mb3r-value"
    When the app is restarted
    And I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.cookies" array length equal to 1
    And the response should contain the field "0.cookies.0.name" equal to "remember"
    And the response should contain the field "0.cookies.0.value" equal to "r3m3mb3r-value"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/profile" } }
      """
    Then the 1 "GET" request to "/profile" should have the header "Cookie" equal to "remember=r3m3mb3r-value"

  @restart @dialog
  Scenario: Changes made in the dialog are written too
    Given the keychain is available
    And I call "CookieService.Upsert" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "example.test", "raw": "kept=k3pt-value; Max-Age=3600" }
      """
    And I call "CookieService.Upsert" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "example.test", "raw": "dropped=d-value; Max-Age=3600" }
      """
    And I call "CookieService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "domain": "example.test", "path": "/", "name": "dropped" }
      """
    When the app is restarted
    And I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.cookies" array length equal to 1
    And the response should contain the field "0.cookies.0.name" equal to "kept"
    And no file under the app data folder should contain "k3pt-value"

  @expiry
  Scenario: Cookies that expired while the app was closed are not restored
    Given the keychain is available
    And the current time is "2026-10-11T10:00:00Z"
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login" } }
      """
    And the current time is "2026-10-13T10:00:00Z"
    When the app is restarted
    And I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0

  @session_only
  Scenario: Only session cookies write no cookie file and no vault
    Given the keychain is available
    And the requests to "/temp" respond with the header "Set-Cookie" equal to "temp=1"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/temp" } }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/cookies.age" should not exist
    And the app data file "workspaces/{{Acme.local_id}}/vault.age" should not exist
    And the keychain should contain 0 entries

  @locked
  Scenario: While the vault is locked cookies stay in memory and are written after unlock
    Given the keychain is unavailable
    And I call "VaultService.SetupPassphrase" with
      """
      { "passphrase": "correct horse", "confirmation": "correct horse" }
      """
    And I call "VariableService.SaveTable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "rows": [ { "key": "api_key", "secret": true, "secret_value": "acme-key", "enabled": true } ] }
      """
    And the app is restarted
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/cookies.age" should not exist
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/profile" } }
      """
    Then the 1 "GET" request to "/profile" should have the header "Cookie" equal to "remember=r3m3mb3r-value; session=s3ss10n-value"
    When I call "VaultService.Unlock" with
      """
      { "passphrase": "correct horse" }
      """
    And I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.cookies" array length equal to 2
    And the app data file "workspaces/{{Acme.local_id}}/cookies.age" should exist
    And no file under the app data folder should contain "r3m3mb3r-value"
    When the app is restarted
    And I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0
    When I call "VaultService.Unlock" with
      """
      { "passphrase": "correct horse" }
      """
    And I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.cookies" array length equal to 1
    And the response should contain the field "0.cookies.0.name" equal to "remember"

  @undecided
  Scenario: Without a usable keychain or passphrase the cookies stay in memory
    Given the keychain is unavailable
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/cookies.age" should not exist
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.cookies" array length equal to 2
