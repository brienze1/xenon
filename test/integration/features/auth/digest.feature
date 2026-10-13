#language: en
#utf-8

@all @xen_005 @auth @digest
Feature: Digest auth
  Digest auth sends the request, receives the 401 (or 403) challenge with
  "WWW-Authenticate: Digest", answers it once and shows the final response; the console shows both
  exchanges. Algorithms MD5, MD5-sess, SHA-256, SHA-256-sess, SHA-512-256 and SHA-512-256-sess and
  qop auth, auth-int or none are supported, with the RFC 7616 hexadecimal nonce count. The
  server's realm, nonce and opaque are kept locally per auth owner (never in the workspace files)
  and reused by the next send. "Disable retry" sends without answering the challenge.

  Background:
    Given the workspace "Acme" exists
    And the "digest" variable is set to the mocked digest server url
    And the "billing" variable is set to "01JK0000000000000000000001"
    And the "secure" variable is set to "01JK0000000000000000000002"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK0000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      auth:
        type: digest
        digest:
          username: ana
          password: "{{digest_password}}"
      variables:
        - key: digest_password
          value: s3cret
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/secure.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{secure}}
      name: Secure
      method: GET
      url: "{{digest}}/secure?page=1"
      auth:
        type: inherit
      """

  @algorithms
  Scenario Outline: The challenge is answered with <algorithm> and qop "<qop>"
    Given the digest server is configured with
      """
      { "algorithm": "<algorithm>", "qop": "<qop>" }
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "POST",
          "url": "{{digest}}/protected?x=1",
          "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"amount\": 10}" } },
          "auth": { "type": "digest", "digest": { "username": "ana", "password": "s3cret", "algorithm": "<algorithm>" } }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "200"
    And the digest server should have received 2 requests and accepted 1
    And the 2 request to the digest server should have the header "Authorization" matching "^Digest username=.ana., realm=.xenon@test., nonce=.[^,]+, uri=./protected\?x=1."
    And the 2 request to the digest server should have the header "Authorization" matching "algorithm=.?<algorithm>.?(,|$)"

    Examples:
      | algorithm | qop |
      | MD5 | auth |
      | MD5 | auth-int |
      | MD5 |  |
      | MD5-sess | auth |
      | MD5-sess | auth-int |
      | SHA-256 | auth |
      | SHA-256 | auth-int |
      | SHA-256 |  |
      | SHA-256-sess | auth |
      | SHA-256-sess | auth-int |
      | SHA-512-256 | auth |
      | SHA-512-256 | auth-int |
      | SHA-512-256 |  |
      | SHA-512-256-sess | auth |
      | SHA-512-256-sess | auth-int |

  @nonce_count
  Scenario: The nonce count is hexadecimal and the challenge is reused by the next send
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{secure}}", "name": "Secure", "method": "GET", "url": "{{digest}}/secure?page=1", "auth": { "type": "inherit" } } }
      """
    Then the response should contain the field "status_code" equal to "200"
    And the 2 request to the digest server should have the header "Authorization" matching "nc=00000001"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{secure}}", "name": "Secure", "method": "GET", "url": "{{digest}}/secure?page=1", "auth": { "type": "inherit" } } }
      """
    Then the response should contain the field "status_code" equal to "200"
    And the digest server should have received 3 requests and accepted 2
    And the 3 request to the digest server should have the header "Authorization" matching "nc=00000002"
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "digest_state.{{billing}}.realm" equal to "xenon@test"
    And no file under the workspace folder should contain "7ypf/xlj9XXwfDPEoM4URrv"

  @retry_once
  Scenario: A wrong password is answered once, never in a loop
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{digest}}/secure", "auth": { "type": "digest", "digest": { "username": "ana", "password": "wrong" } } } }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "401"
    And the digest server should have received 2 requests and accepted 0

  @disable_retry
  Scenario: "Disable retry" sends without answering the challenge
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{digest}}/secure", "auth": { "type": "digest", "digest": { "username": "ana", "password": "s3cret", "disable_retry_request": true } } } }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "401"
    And the digest server should have received 1 request and accepted 0

  @typed_challenge
  Scenario: Advanced fields typed by the user answer on the first request
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{digest}}/secure",
          "auth": { "type": "digest", "digest": { "username": "ana", "password": "s3cret", "realm": "xenon@test", "nonce": "7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v", "opaque": "FQhe/qaU925kfnzjCev0ciny7QMkPqMAFRtzCUYo5tdS", "qop": "auth", "algorithm": "MD5", "nonce_count": "00000005", "client_nonce": "0a4f113b" } }
        }
      }
      """
    Then the response should contain the field "status_code" equal to "200"
    And the digest server should have received 1 request and accepted 1
    And the 1 request to the digest server should have the header "Authorization" matching "cnonce=.0a4f113b."
    And the 1 request to the digest server should have the header "Authorization" matching "nc=00000006"

  @console
  Scenario: The console shows both exchanges
    Given I call "ConsoleService.Clear"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{secure}}", "name": "Secure", "method": "GET", "url": "{{digest}}/secure?page=1", "auth": { "type": "inherit" } } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "0.kind" equal to "request"
    And the response should contain the field "0.response.status_code" equal to "401"
    And the response should contain the field "1.kind" equal to "request"
    And the response should contain the field "1.response.status_code" equal to "200"
    And the response should not contain "s3cret"
