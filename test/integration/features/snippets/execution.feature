#language: en
#utf-8

@all @xen_005 @snippets @snippet_execution
Feature: Generated snippets reproduce the request when run
  A snippet run with its tool (curl through bash, node, go run, python with requests) from the
  workspace folder sends the same method, URL, headers and body as Xenon would (SC-006). CI
  installs the four runtimes on macOS, Windows and Linux; on a machine without one of them the
  scenarios for that runtime are skipped outside CI. As in Postman's generators, a binary-file
  body is a "<file contents here>" placeholder in fetch, Go and Python, and fetch reads form-data
  files from a browser file input, so those combinations are not run.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the workspace file "acme/files/invoice.txt" contains
      """
      invoice #7
      """

  @query @headers
  Scenario Outline: Query parameters and headers - <language>
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "<language>",
        "request": {
          "method": "GET",
          "url": "{{api}}/users?page=2&q=ana%20maria",
          "headers": [ { "key": "Accept", "value": "application/json", "enabled": true }, { "key": "X-Trace", "value": "it's \"me\"", "enabled": true } ]
        }
      }
      """
    And I run the response field "text" with "<runtime>" in the workspace folder "acme"
    Then the "GET" request to "/users" should have been received 1 times
    And the 1 "GET" request to "/users" should have the query "page" equal to "2"
    And the 1 "GET" request to "/users" should have the query "q" equal to "ana maria"
    And the 1 "GET" request to "/users" should have the header "Accept" equal to "application/json"
    And the 1 "GET" request to "/users" should have the header "X-Trace" matching "^it's .me.$"

    Examples:
      | language        | runtime |
      | curl            | curl    |
      | js_fetch        | node    |
      | go_native       | go      |
      | python_requests | python  |

  @raw
  Scenario Outline: A raw JSON body - <language>
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "<language>",
        "request": {
          "method": "POST",
          "url": "{{api}}/users",
          "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"name\":\"Ana\",\"tags\":[\"a\",\"b\"],\"admin\":false,\"boss\":null}" } }
        }
      }
      """
    And I run the response field "text" with "<runtime>" in the workspace folder "acme"
    Then the "POST" request to "/users" should have been received 1 times
    And the 1 "POST" request to "/users" should have the header "Content-Type" equal to "application/json"
    And the 1 "POST" request to "/users" should have the body field "name" equal to "Ana"
    And the 1 "POST" request to "/users" should have the body field "tags.1" equal to "b"
    And the 1 "POST" request to "/users" should have the body field "admin" equal to "false"
    And the 1 "POST" request to "/users" should have the body field "boss" equal to "null"

    Examples:
      | language        | runtime |
      | curl            | curl    |
      | js_fetch        | node    |
      | go_native       | go      |
      | python_requests | python  |

  @raw @text
  Scenario Outline: A raw text body with an @ and quotes - <language>
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "<language>",
        "request": {
          "method": "PUT",
          "url": "{{api}}/notes",
          "body": { "mode": "raw", "raw": { "language": "text", "content": "@team it's \"done\" $HOME" } }
        }
      }
      """
    And I run the response field "text" with "<runtime>" in the workspace folder "acme"
    Then the "PUT" request to "/notes" should have been received 1 times
    And the 1 "PUT" request to "/notes" should have the header "Content-Type" equal to "text/plain"
    And the 1 "PUT" request to "/notes" should have the body equal to
      """
      @team it's "done" $HOME
      """

    Examples:
      | language        | runtime |
      | curl            | curl    |
      | js_fetch        | node    |
      | go_native       | go      |
      | python_requests | python  |

  @urlencoded
  Scenario Outline: A url-encoded body - <language>
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "<language>",
        "request": {
          "method": "POST",
          "url": "{{api}}/login",
          "body": { "mode": "urlencoded", "urlencoded": [ { "key": "user", "value": "ana", "enabled": true }, { "key": "pass", "value": "s3 cret&x", "enabled": true }, { "key": "skip", "value": "1", "enabled": false } ] }
        }
      }
      """
    And I run the response field "text" with "<runtime>" in the workspace folder "acme"
    Then the "POST" request to "/login" should have been received 1 times
    And the 1 "POST" request to "/login" should have the header "Content-Type" matching "^application/x-www-form-urlencoded"
    And the 1 "POST" request to "/login" should have the body matching "^user=ana&pass=s3(%20|\+)cret%26x$"

    Examples:
      | language        | runtime |
      | curl            | curl    |
      | js_fetch        | node    |
      | go_native       | go      |
      | python_requests | python  |

  @formdata
  Scenario Outline: A form-data body with a text field and a file - <language>
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "<language>",
        "request": {
          "method": "POST",
          "url": "{{api}}/upload",
          "body": { "mode": "formdata", "formdata": [ { "key": "note", "value": "hello", "type": "text", "enabled": true }, { "key": "file", "src": "files/invoice.txt", "type": "file", "enabled": true } ] }
        }
      }
      """
    And I run the response field "text" with "<runtime>" in the workspace folder "acme"
    Then the "POST" request to "/upload" should have been received 1 times
    And the 1 "POST" request to "/upload" should have the form field "note" equal to "hello"
    And the 1 "POST" request to "/upload" should have the form file "file" named "invoice.txt" with content "invoice #7"

    Examples:
      | language        | runtime |
      | curl            | curl    |
      | go_native       | go      |
      | python_requests | python  |

  @graphql
  Scenario Outline: A GraphQL body - <language>
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "<language>",
        "request": {
          "method": "POST",
          "url": "{{api}}/graphql",
          "body": { "mode": "graphql", "graphql": { "query": "query User($id: ID!) { user(id: $id) { name } }", "variables": "{\"id\": \"42\"}" } }
        }
      }
      """
    And I run the response field "text" with "<runtime>" in the workspace folder "acme"
    Then the "POST" request to "/graphql" should have been received 1 times
    And the 1 "POST" request to "/graphql" should have the header "Content-Type" equal to "application/json"
    And the 1 "POST" request to "/graphql" should have the body field "query" equal to "query User($id: ID!) { user(id: $id) { name } }"
    And the 1 "POST" request to "/graphql" should have the body field "variables.id" equal to "42"

    Examples:
      | language        | runtime |
      | curl            | curl    |
      | js_fetch        | node    |
      | go_native       | go      |
      | python_requests | python  |

  @binary
  Scenario: A binary-file body - curl
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "curl",
        "request": { "method": "PUT", "url": "{{api}}/blob", "body": { "mode": "binary", "binary": { "src": "files/invoice.txt" } } }
      }
      """
    And I run the response field "text" with "curl" in the workspace folder "acme"
    Then the "PUT" request to "/blob" should have been received 1 times
    And the 1 "PUT" request to "/blob" should have the header "Content-Type" equal to "application/octet-stream"
    And the 1 "PUT" request to "/blob" should have the body equal to
      """
      invoice #7
      """

  @auth
  Scenario Outline: Basic, API key and AWS Signature auth - <language>
    Given the current time is "2026-10-11T10:00:00Z"
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "<language>",
        "request": { "method": "DELETE", "url": "{{api}}/orders/7", "auth": { "type": "basic", "basic": { "username": "ana", "password": "s3cret" } } }
      }
      """
    And I run the response field "text" with "<runtime>" in the workspace folder "acme"
    And I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "<language>",
        "request": { "method": "GET", "url": "{{api}}/search?q=x", "auth": { "type": "apikey", "apikey": { "key": "api_key", "value": "k3y", "in": "query" } } }
      }
      """
    And I run the response field "text" with "<runtime>" in the workspace folder "acme"
    And I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "<language>",
        "request": { "method": "GET", "url": "{{api}}/items", "auth": { "type": "awsv4", "awsv4": { "access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI", "region": "eu-west-1", "service": "execute-api" } } }
      }
      """
    And I run the response field "text" with "<runtime>" in the workspace folder "acme"
    Then the 1 "DELETE" request to "/orders/7" should have the header "Authorization" equal to "Basic YW5hOnMzY3JldA=="
    And the 1 "GET" request to "/search" should have the query "q" equal to "x"
    And the 1 "GET" request to "/search" should have the query "api_key" equal to "k3y"
    And the 1 "GET" request to "/items" should have the header "X-Amz-Date" matching "^20261011T1000[0-9]{2}Z$"
    And the 1 "GET" request to "/items" should have the header "Authorization" matching "^AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20261011/eu-west-1/execute-api/aws4_request, SignedHeaders=host;x-amz-date, Signature=[0-9a-f]{64}$"

    Examples:
      | language        | runtime |
      | curl            | curl    |
      | js_fetch        | node    |
      | go_native       | go      |
      | python_requests | python  |
