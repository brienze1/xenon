#language: en
#utf-8

@all @xen_001 @sending @send_basics
Feature: Send a request
  Go sends the request built in a tab - method, URL with query and path variables, headers,
  body and auth - and returns the status, headers, size, body and timings. In stage 1 every
  "{{variable}}" is sent literally.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url

  @success
  Scenario: A GET with a query string returns the full result
    Given the 1 "GET" request to "/users" returns status 200 with the following response
      """
      {"users":[{"id":1}]}
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/users?page=2&limit=10",
          "params": [ { "key": "page", "value": "2" }, { "key": "limit", "value": "10" } ]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "tab_id" equal to "tab-1"
    And the response should contain the field "ok" equal to "true"
    And the response should contain the field "status_code" equal to "200"
    And the response should contain the field "status_text" equal to "OK"
    And the response should contain the field "proto" equal to "HTTP/1.1"
    And the response should contain the field "content_type" equal to "application/json"
    And the response should contain the field "is_binary" equal to "false"
    And the response should contain the field "size_bytes" equal to "20"
    And the response should contain the field "inline_body" equal to
      """
      {"users":[{"id":1}]}
      """
    And the response should contain the field "body_url" equal to "/xenon/response/tab-1"
    And the response field "headers" should contain an item with "key" equal to "Content-Type" and "value" equal to "application/json"
    And the response field "timings.dns_ms" should exist
    And the response field "timings.connect_ms" should exist
    And the response field "timings.tls_ms" should exist
    And the response field "timings.first_byte_ms" should exist
    And the response field "timings.download_ms" should exist
    And the response field "timings.total_ms" should exist
    And the response field "error" should not exist
    And the 1 "GET" request to "/users" should have the query "page" equal to "2"
    And the 1 "GET" request to "/users" should have the query "limit" equal to "10"
    And the 1 "GET" request to "/users" should have the header "User-Agent" equal to "Xenon/0.1.0"
    And the 1 "GET" request to "/users" should have the header "Accept" equal to "*/*"
    And the 1 "GET" request to "/users" should have the header "Accept-Encoding" equal to "gzip, deflate, br"

  @success
  Scenario: A 4xx response is a result, not a failure
    Given the 1 "DELETE" request to "/users/1" returns status 404 with the following response
      """
      {"error":"not found"}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "DELETE", "url": "{{api}}/users/1" } }
      """
    Then the call should succeed
    And the response should contain the field "ok" equal to "true"
    And the response should contain the field "status_code" equal to "404"
    And the response should contain the field "status_text" equal to "Not Found"

  @path_variables
  Scenario: Path variable values replace their segments
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/customers/:customerId/invoices/:invoiceId",
          "path_variables": [ { "key": "customerId", "value": "42" }, { "key": "invoiceId", "value": "7" } ]
        }
      }
      """
    Then the call should succeed
    And the "GET" request to "/customers/42/invoices/7" should have been received 1 times

  @method
  Scenario: A custom method is sent upper-case
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "purge", "url": "{{api}}/cache" } }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "200"
    And the "PURGE" request to "/cache" should have been received 1 times

  @body @raw
  Scenario Outline: A raw <language> body is sent with its Content-Type
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "POST",
          "url": "{{api}}/items",
          "body": { "mode": "raw", "raw": { "language": "<language>", "content": "<content>" } }
        }
      }
      """
    Then the call should succeed
    And the 1 "POST" request to "/items" should have the header "Content-Type" equal to "<content_type>"
    And the 1 "POST" request to "/items" should have the body equal to
      """
      <content>
      """

    Examples:
      | language   | content_type           | content                  |
      | json       | application/json       | [1, 2, 3]                |
      | xml        | application/xml        | <amount>1000</amount>    |
      | text       | text/plain             | plain words              |
      | html       | text/html              | <p>hello</p>             |
      | javascript | application/javascript | console.log(1)           |

  @body @formdata
  Scenario: Form-data sends text fields and files
    Given the file "files/note.txt" exists with content
      """
      file body
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "POST",
          "url": "{{api}}/upload",
          "body": {
            "mode": "formdata",
            "formdata": [
              { "key": "note", "type": "text", "value": "hello" },
              { "key": "skipped", "type": "text", "value": "no", "enabled": false },
              { "key": "attachment", "type": "file", "src": "{{temp_root}}/files/note.txt" }
            ]
          }
        }
      }
      """
    Then the call should succeed
    And the 1 "POST" request to "/upload" should have the form field "note" equal to "hello"
    And the 1 "POST" request to "/upload" should not have the form field "skipped"
    And the 1 "POST" request to "/upload" should have the form file "attachment" named "note.txt" with content "file body"

  @body @urlencoded
  Scenario: Url-encoded rows are encoded in order and disabled rows are skipped
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "POST",
          "url": "{{api}}/login",
          "body": {
            "mode": "urlencoded",
            "urlencoded": [
              { "key": "user", "value": "luis" },
              { "key": "skip", "value": "x", "enabled": false },
              { "key": "msg", "value": "two words&more" }
            ]
          }
        }
      }
      """
    Then the call should succeed
    And the 1 "POST" request to "/login" should have the header "Content-Type" equal to "application/x-www-form-urlencoded"
    And the 1 "POST" request to "/login" should have the body equal to
      """
      user=luis&msg=two+words%26more
      """

  @body @binary
  Scenario: A binary body sends the file bytes
    Given the file "files/payload.bin" exists with content
      """
      raw-bytes
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "PUT", "url": "{{api}}/blob", "body": { "mode": "binary", "binary": { "src": "{{temp_root}}/files/payload.bin" } } }
      }
      """
    Then the call should succeed
    And the 1 "PUT" request to "/blob" should have the header "Content-Type" equal to "application/octet-stream"
    And the 1 "PUT" request to "/blob" should have the body equal to
      """
      raw-bytes
      """

  @body @binary
  Scenario: A relative file path is read from the workspace folder
    Given the workspace file "acme/fixtures/data.bin" contains
      """
      inside the workspace
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "POST", "url": "{{api}}/blob", "body": { "mode": "binary", "binary": { "src": "fixtures/data.bin" } } }
      }
      """
    Then the call should succeed
    And the 1 "POST" request to "/blob" should have the body equal to
      """
      inside the workspace
      """

  @body @graphql
  Scenario: GraphQL is sent as a JSON body with the query and the variables
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "POST",
          "url": "{{api}}/graphql",
          "body": { "mode": "graphql", "graphql": { "query": "query { me { id } }", "variables": "{ \"limit\": 5 }" } }
        }
      }
      """
    Then the call should succeed
    And the 1 "POST" request to "/graphql" should have the header "Content-Type" equal to "application/json"
    And the 1 "POST" request to "/graphql" should have the body equal to
      """
      {"query":"query { me { id } }","variables":{"limit":5}}
      """

  @body @graphql
  Scenario: Empty GraphQL variables are sent as an empty object
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "POST", "url": "{{api}}/graphql", "body": { "mode": "graphql", "graphql": { "query": "{ me }", "variables": "  " } } }
      }
      """
    Then the call should succeed
    And the 1 "POST" request to "/graphql" should have the body equal to
      """
      {"query":"{ me }","variables":{}}
      """

  @auth
  Scenario: Bearer auth reaches the server
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "GET", "url": "{{api}}/me", "auth": { "type": "bearer", "bearer": { "token": "abc" } } }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/me" should have the header "Authorization" equal to "Bearer abc"

  @auth
  Scenario: Basic auth reaches the server
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "GET", "url": "{{api}}/me", "auth": { "type": "basic", "basic": { "username": "user", "password": "pass" } } }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/me" should have the header "Authorization" equal to "Basic dXNlcjpwYXNz"

  @headers
  Scenario: User headers are sent, variables literally, and they replace automatic ones
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/me",
          "headers": [
            { "key": "X-Token", "value": "{{token}}" },
            { "key": "user-agent", "value": "custom/1.0" },
            { "key": "X-Off", "value": "no", "enabled": false }
          ]
        }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/me" should have the header "X-Token" equal to "{{token}}"
    And the 1 "GET" request to "/me" should have the header "User-Agent" equal to "custom/1.0"
    And the 1 "GET" request to "/me" should not have the header "X-Off"

  @headers
  Scenario: Automatic headers switched off are not sent
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": { "method": "GET", "url": "{{api}}/me", "auto_headers": { "disabled": ["User-Agent", "accept"] } }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/me" should not have the header "User-Agent"
    And the 1 "GET" request to "/me" should not have the header "Accept"

  @response_headers
  Scenario: Repeated response headers are all kept
    Given the requests to "/cookies" respond with the header "Set-Cookie" equal to "a=1"
    And the requests to "/cookies" respond with the header "Set-Cookie" equal to "b=2"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/cookies" } }
      """
    Then the call should succeed
    And the response field "headers" should contain 2 items with "key" equal to "Set-Cookie"
    And the response field "headers" should contain an item with "key" equal to "Set-Cookie" and "value" equal to "a=1"
    And the response field "headers" should contain an item with "key" equal to "Set-Cookie" and "value" equal to "b=2"

  @validation
  Scenario Outline: Sending needs a workspace and a tab - <case>
    When I call "RequestService.Send" with
      """
      { "local_id": "<local_id>", "tab_id": "<tab_id>", "request": { "url": "{{api}}/me" } }
      """
    Then the call should fail with code "SND-00400"

    Examples:
      | case         | local_id           | tab_id |
      | no workspace |                    | tab-1  |
      | no tab       | {{Acme.local_id}}  |        |

  @not_found
  Scenario: Sending in an unknown workspace fails
    When I call "RequestService.Send" with
      """
      { "local_id": "01JA2Z5V3B6C7D8E9F0G1H2J3K", "tab_id": "tab-1", "request": { "url": "{{api}}/me" } }
      """
    Then the call should fail with code "WKS-07404"
