#language: en
#utf-8

@all @xen_001 @requests @auto_headers
Feature: Headers Xenon adds automatically
  The Headers tab lists the headers Xenon adds by itself, greyed out, each one switchable off.
  A user header with the same name (any case) replaces the automatic one.

  Background:
    Given the workspace "Acme" exists

  @success
  Scenario: A GET without body or auth gets the connection headers only
    When I call "RequestService.AutoHeaders" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "https://api.example.test:8443/users?page=2" } }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 5
    And the response field "" should contain an item with "key" equal to "Host" and "value" equal to "api.example.test:8443"
    And the response field "" should contain an item with "key" equal to "User-Agent" and "value" equal to "Xenon/0.1.0"
    And the response field "" should contain an item with "key" equal to "Accept" and "value" equal to "*/*"
    And the response field "" should contain an item with "key" equal to "Accept-Encoding" and "value" equal to "gzip, deflate, br"
    And the response field "" should contain an item with "key" equal to "Connection" and "value" equal to "keep-alive"
    And the response field "" should contain an item with "key" equal to "Connection" and "disabled" equal to "false"
    And the response field "" should contain 0 items with "key" equal to "Content-Type"
    And the response field "" should contain 0 items with "key" equal to "Authorization"

  @content_type
  Scenario Outline: Content-Type follows the raw body language - <language>
    When I call "RequestService.AutoHeaders" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request": {
          "method": "POST",
          "url": "https://api.example.test/items",
          "body": { "mode": "raw", "raw": { "language": "<language>", "content": "x" } }
        }
      }
      """
    Then the call should succeed
    And the response field "" should contain an item with "key" equal to "Content-Type" and "value" equal to "<content_type>"
    And the response field "" should contain an item with "key" equal to "Content-Length" and "value" equal to "<calculated when request is sent>"

    Examples:
      | language   | content_type           |
      | json       | application/json       |
      | xml        | application/xml        |
      | text       | text/plain             |
      | html       | text/html              |
      | javascript | application/javascript |

  @content_type
  Scenario Outline: Content-Type follows the body mode - <mode>
    When I call "RequestService.AutoHeaders" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request": { "method": "POST", "url": "https://api.example.test/items", "body": { "mode": "<mode>" } }
      }
      """
    Then the call should succeed
    And the response field "" should contain an item with "key" equal to "Content-Type" and "value" equal to "<content_type>"

    Examples:
      | mode       | content_type                                                  |
      | urlencoded | application/x-www-form-urlencoded                             |
      | formdata   | multipart/form-data; boundary=<calculated when request is sent> |
      | graphql    | application/json                                              |
      | binary     | application/octet-stream                                      |

  @content_type
  Scenario: No body means no Content-Type and no Content-Length
    When I call "RequestService.AutoHeaders" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request": { "method": "POST", "url": "https://api.example.test/items", "body": { "mode": "none" } }
      }
      """
    Then the call should succeed
    And the response field "" should contain 0 items with "key" equal to "Content-Type"
    And the response field "" should contain 0 items with "key" equal to "Content-Length"

  @auth
  Scenario: Bearer auth adds an Authorization header
    When I call "RequestService.AutoHeaders" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request": { "url": "https://api.example.test/me", "auth": { "type": "bearer", "bearer": { "token": "abc" } } }
      }
      """
    Then the call should succeed
    And the response field "" should contain an item with "key" equal to "Authorization" and "value" equal to "Bearer abc"

  @auth
  Scenario: Basic auth adds the encoded credentials
    When I call "RequestService.AutoHeaders" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request": { "url": "https://api.example.test/me", "auth": { "type": "basic", "basic": { "username": "user", "password": "pass" } } }
      }
      """
    Then the call should succeed
    And the response field "" should contain an item with "key" equal to "Authorization" and "value" equal to "Basic dXNlcjpwYXNz"

  @auth
  Scenario Outline: No Authorization header for auth "<type>" on an unsaved request
    When I call "RequestService.AutoHeaders" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request": { "url": "https://api.example.test/me", "auth": { "type": "<type>", "bearer": { "token": "abc" } } }
      }
      """
    Then the call should succeed
    And the response field "" should contain 0 items with "key" equal to "Authorization"

    Examples:
      | type    |
      | inherit |
      | none    |
      | apikey  |

  @disabled
  Scenario: An auto header switched off stays listed as disabled, matched in any case
    When I call "RequestService.AutoHeaders" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request": { "url": "https://api.example.test/me", "auto_headers": { "disabled": ["user-agent"] } }
      }
      """
    Then the call should succeed
    And the response field "" should contain an item with "key" equal to "User-Agent" and "disabled" equal to "true"
    And the response field "" should contain an item with "key" equal to "Accept" and "disabled" equal to "false"

  @override
  Scenario: A user header with the same name replaces the automatic one, a disabled one does not
    When I call "RequestService.AutoHeaders" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request": {
          "url": "https://api.example.test/me",
          "auth": { "type": "bearer", "bearer": { "token": "abc" } },
          "headers": [
            { "key": "accept", "value": "application/json" },
            { "key": "AUTHORIZATION", "value": "Token xyz" },
            { "key": "User-Agent", "value": "custom/1.0", "enabled": false }
          ]
        }
      }
      """
    Then the call should succeed
    And the response field "" should contain 0 items with "key" equal to "Accept"
    And the response field "" should contain 0 items with "key" equal to "Authorization"
    And the response field "" should contain an item with "key" equal to "User-Agent" and "value" equal to "Xenon/0.1.0"

  @host
  Scenario: The host is shown when the URL can be read and computed at send time otherwise
    When I call "RequestService.AutoHeaders" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "url": "{{host}}/users/:id" } }
      """
    Then the call should succeed
    And the response field "" should contain an item with "key" equal to "Host" and "value" equal to "<calculated when request is sent>"
    When I call "RequestService.AutoHeaders" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "url": "api.example.test/users" } }
      """
    Then the response field "" should contain an item with "key" equal to "Host" and "value" equal to "api.example.test"

  @validation
  Scenario: A workspace is required
    When I call "RequestService.AutoHeaders" with
      """
      { "local_id": "", "request": { "url": "https://api.example.test" } }
      """
    Then the call should fail with code "REQ-07400"
