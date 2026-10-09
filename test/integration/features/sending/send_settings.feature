#language: en
#utf-8

@all @xen_001 @sending @send_settings
Feature: Per-request send settings
  Each request has a timeout (0 = none), follow redirects, SSL certificate verification and
  automatic URL encoding. A setting that causes a failure is named in the error hint.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url

  @timeout
  Scenario: A request slower than its timeout fails with a timeout error
    Given the "GET" requests to "/slow" respond after 1000ms
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "GET", "url": "{{api}}/slow", "settings": { "timeout_ms": 100 } }
      }
      """
    Then the call should succeed
    And the response should contain the field "ok" equal to "false"
    And the response should contain the field "status_code" equal to "0"
    And the response should contain the field "error.kind" equal to "timeout"
    And the response should contain the field "error.message" equal to "Timed out after 100 ms"
    And the response should contain the field "error.hint" equal to "Increase the timeout in the request Settings (0 = no timeout)"

  @timeout
  Scenario: Without a timeout a slow response is awaited
    Given the "GET" requests to "/slow" respond after 300ms
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "GET", "url": "{{api}}/slow", "settings": { "timeout_ms": 0 } }
      }
      """
    Then the call should succeed
    And the response should contain the field "ok" equal to "true"
    And the response should contain the field "status_code" equal to "200"

  @redirects
  Scenario: Redirects are followed by default
    Given the requests to "/old" redirect to "{{api}}/new" with status 302
    And the 1 "GET" request to "/new" returns status 200 with the following response
      """
      {"moved":true}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/old" } }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "200"
    And the response should contain the field "inline_body" equal to
      """
      {"moved":true}
      """
    And the "GET" request to "/new" should have been received 1 times

  @redirects
  Scenario: With redirects off the 3xx response is shown
    Given the requests to "/old" redirect to "{{api}}/new" with status 302
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "GET", "url": "{{api}}/old", "settings": { "follow_redirects": false } }
      }
      """
    Then the call should succeed
    And the response should contain the field "ok" equal to "true"
    And the response should contain the field "status_code" equal to "302"
    And the response should contain the field "status_text" equal to "Found"
    And the response field "headers" should contain an item with "key" equal to "Location" and "value" equal to "{{api}}/new"
    And the "GET" request to "/new" should have been received 0 times

  @ssl
  Scenario: A self-signed certificate is refused while SSL verification is on
    Given the "tls_api" variable is set to the mocked tls api url
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{tls_api}}/secure" } }
      """
    Then the call should succeed
    And the response should contain the field "ok" equal to "false"
    And the response should contain the field "error.kind" equal to "tls"
    And the response should contain the field "error.message" equal to "Certificate not trusted"
    And the response should contain the field "error.hint" equal to "Turn off SSL certificate verification in Settings to send anyway"
    And the "GET" request to "/secure" should have been received 0 times

  @ssl
  Scenario: With SSL verification off the self-signed server answers
    Given the "tls_api" variable is set to the mocked tls api url
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "GET", "url": "{{tls_api}}/secure", "settings": { "verify_ssl": false } }
      }
      """
    Then the call should succeed
    And the response should contain the field "ok" equal to "true"
    And the response should contain the field "status_code" equal to "200"
    And the "GET" request to "/secure" should have been received 1 times

  @encode_url
  Scenario: With URL encoding on, unsafe characters are percent-encoded and existing escapes kept
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "GET", "url": "{{api}}/files/my file.txt?q=a|b c&done=a%20b&json={\"a\":1}" }
      }
      """
    Then the call should succeed
    And the "GET" request to "/files/my file.txt" should have been received 1 times
    And the 1 "GET" request to "/files/my file.txt" should have the raw query "q=a%7Cb%20c&done=a%20b&json=%7B%22a%22:1%7D"
    And the 1 "GET" request to "/files/my file.txt" should have the query "q" equal to "a|b c"

  @encode_url
  Scenario: With URL encoding off the query is sent exactly as typed
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "GET", "url": "{{api}}/search?q=a|b&x=%41", "settings": { "encode_url": false } }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/search" should have the raw query "q=a|b&x=%41"
