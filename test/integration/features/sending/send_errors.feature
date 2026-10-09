#language: en
#utf-8

@all @xen_001 @sending @send_errors
Feature: Errors that prevent a response
  When no response arrives the result says why - unresolvable host, refused connection,
  invalid URL, missing body file or cancellation - with a hint when a setting changes it.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url

  @dns
  Scenario: A host that cannot be resolved
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "http://xenon-unresolvable.invalid/ping" } }
      """
    Then the call should succeed
    And the response should contain the field "ok" equal to "false"
    And the response should contain the field "error.kind" equal to "dns"
    And the response should contain the field "error.message" equal to "Could not resolve host xenon-unresolvable.invalid"
    And the response should contain the field "error.hint" equal to ""

  @connection_refused
  Scenario: Nothing listens on the port
    Given the "closed" variable is set to a closed port url
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{closed}}/ping" } }
      """
    Then the call should succeed
    And the response should contain the field "ok" equal to "false"
    And the response should contain the field "error.kind" equal to "connection_refused"
    And the response should contain the field "error.hint" equal to ""

  @invalid_url
  Scenario Outline: An invalid URL is not sent - <case>
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "<url>" } }
      """
    Then the call should succeed
    And the response should contain the field "ok" equal to "false"
    And the response should contain the field "error.kind" equal to "invalid_url"
    And the response should contain the field "error.message" equal to "<message>"
    And the response should contain the field "error.hint" equal to "Check the URL"

    Examples:
      | case             | url                | message                         |
      | empty            |                    | Enter a URL to send the request |
      | broken host      | http://[::1/users  | Invalid URL: http://[::1/users  |
      | unknown scheme   | ftp://example.test | Invalid URL: ftp://example.test |

  @missing_file
  Scenario: A body file that does not exist blocks the send
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "POST", "url": "{{api}}/blob", "body": { "mode": "binary", "binary": { "src": "{{temp_root}}/missing.bin" } } }
      }
      """
    Then the call should succeed
    And the response should contain the field "ok" equal to "false"
    And the response should contain the field "error.kind" equal to "missing_file"
    And the response should contain the field "error.message" equal to "File not found: {{temp_root}}/missing.bin"
    And the response should contain the field "error.hint" equal to "Choose the file again in the Body tab"
    And the "POST" request to "/blob" should have been received 0 times

  @missing_file
  Scenario: A form-data file that does not exist blocks the send
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "POST",
          "url": "{{api}}/upload",
          "body": { "mode": "formdata", "formdata": [ { "key": "f", "type": "file", "src": "{{temp_root}}/gone.txt" } ] }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "error.kind" equal to "missing_file"
    And the "POST" request to "/upload" should have been received 0 times

  @cancel
  Scenario: Cancelling an in-flight request aborts it
    Given the "GET" requests to "/slow" respond after 3000ms
    When I call "RequestService.Send" and cancel it after 100ms with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/slow" } }
      """
    Then the call should succeed
    And the response should contain the field "ok" equal to "false"
    And the response should contain the field "error.kind" equal to "cancelled"
    And the response should contain the field "error.message" equal to "Request cancelled"
    When I call "RequestService.LastResult" with
      """
      { "tab_id": "tab-1" }
      """
    Then the response should contain the field "error.kind" equal to "cancelled"
