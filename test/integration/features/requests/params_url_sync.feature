#language: en
#utf-8

@all @xen_001 @requests @params_url_sync
Feature: Keep the Params tables and the URL in sync
  The URL is the source of truth for enabled query parameters and for path variables. Editing
  the URL re-derives the tables (keeping disabled rows and descriptions); editing the tables
  rebuilds the URL. Text is kept exactly as typed - encoding only happens when sending.

  @apply_url
  Scenario: The query string becomes enabled params in order
    When I call "RequestService.ApplyUrl" with
      """
      { "request": { "method": "GET", "url": "https://api.example.test/users?page=2&limit=10" } }
      """
    Then the call should succeed
    And the response should contain the field "url" equal to "https://api.example.test/users?page=2&limit=10"
    And the response should contain the field "method" equal to "GET"
    And the response should contain the field "params" array length equal to 2
    And the response should contain the field "params.0.key" equal to "page"
    And the response should contain the field "params.0.value" equal to "2"
    And the response should contain the field "params.0.enabled" equal to "true"
    And the response should contain the field "params.1.key" equal to "limit"
    And the response should contain the field "params.1.value" equal to "10"
    And the response should contain the field "path_variables" array length equal to 0

  @apply_url
  Scenario: Disabled rows and descriptions are kept when the URL changes
    When I call "RequestService.ApplyUrl" with
      """
      {
        "request": {
          "url": "https://api.example.test/users?page=3&limit=10",
          "params": [
            { "key": "page", "value": "2", "description": "Page number" },
            { "key": "debug", "value": "1", "enabled": false, "description": "only in dev" },
            { "key": "limit", "value": "10" }
          ]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "params" array length equal to 3
    And the response should contain the field "params.0.key" equal to "page"
    And the response should contain the field "params.0.value" equal to "3"
    And the response should contain the field "params.0.description" equal to "Page number"
    And the response should contain the field "params.1.key" equal to "debug"
    And the response should contain the field "params.1.enabled" equal to "false"
    And the response should contain the field "params.1.description" equal to "only in dev"
    And the response should contain the field "params.2.key" equal to "limit"
    And the response should contain the field "params.2.enabled" equal to "true"

  @apply_url
  Scenario: Removing a parameter from the URL removes its enabled row only
    When I call "RequestService.ApplyUrl" with
      """
      {
        "request": {
          "url": "https://api.example.test/users?page=2",
          "params": [
            { "key": "page", "value": "2" },
            { "key": "debug", "value": "1", "enabled": false },
            { "key": "limit", "value": "10" }
          ]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "params" array length equal to 2
    And the response should contain the field "params.0.key" equal to "page"
    And the response should contain the field "params.1.key" equal to "debug"
    And the response should contain the field "params.1.enabled" equal to "false"

  @apply_url
  Scenario: New parameters typed in the URL are appended after the existing rows
    When I call "RequestService.ApplyUrl" with
      """
      {
        "request": {
          "url": "https://api.example.test/users?a=1&b=2&c=3",
          "params": [
            { "key": "a", "value": "1" },
            { "key": "x", "value": "9", "enabled": false }
          ]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "params" array length equal to 4
    And the response should contain the field "params.0.key" equal to "a"
    And the response should contain the field "params.1.key" equal to "x"
    And the response should contain the field "params.2.key" equal to "b"
    And the response should contain the field "params.3.key" equal to "c"
    And the response should contain the field "params.3.value" equal to "3"

  @apply_url
  Scenario: Keys without values, values containing "=" and encoded text are read as typed
    When I call "RequestService.ApplyUrl" with
      """
      { "request": { "url": "https://api.example.test/search?flag&expr=a=b&empty=&q=a%20b&&t={{token}}#top" } }
      """
    Then the call should succeed
    And the response should contain the field "params" array length equal to 5
    And the response should contain the field "params.0.key" equal to "flag"
    And the response should contain the field "params.0.value" equal to ""
    And the response should contain the field "params.1.key" equal to "expr"
    And the response should contain the field "params.1.value" equal to "a=b"
    And the response should contain the field "params.2.key" equal to "empty"
    And the response should contain the field "params.2.value" equal to ""
    And the response should contain the field "params.3.value" equal to "a%20b"
    And the response should contain the field "params.4.key" equal to "t"
    And the response should contain the field "params.4.value" equal to "{{token}}"

  @apply_url
  Scenario: A URL without a query string leaves only the disabled rows
    When I call "RequestService.ApplyUrl" with
      """
      {
        "request": {
          "url": "https://api.example.test/users",
          "params": [
            { "key": "page", "value": "2" },
            { "key": "debug", "value": "1", "enabled": false }
          ]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "params" array length equal to 1
    And the response should contain the field "params.0.key" equal to "debug"

  @path_variables
  Scenario: Path variables are derived from ":name" segments in order and keep their values
    When I call "RequestService.ApplyUrl" with
      """
      {
        "request": {
          "url": "https://api.example.test/customers/:customerId/invoices/:invoiceId?draft=true",
          "path_variables": [
            { "key": "invoiceId", "value": "7", "description": "the invoice" },
            { "key": "gone", "value": "1" }
          ]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "path_variables" array length equal to 2
    And the response should contain the field "path_variables.0.key" equal to "customerId"
    And the response should contain the field "path_variables.0.value" equal to ""
    And the response should contain the field "path_variables.1.key" equal to "invoiceId"
    And the response should contain the field "path_variables.1.value" equal to "7"
    And the response should contain the field "path_variables.1.description" equal to "the invoice"
    And the response should contain the field "params.0.key" equal to "draft"

  @path_variables
  Scenario: Schemes, ports, variables and repeated names do not create extra path variables
    When I call "RequestService.ApplyUrl" with
      """
      { "request": { "url": "http://localhost:8080/{{base}}/users/:id/copies/:id/a:b?x=:y" } }
      """
    Then the call should succeed
    And the response should contain the field "path_variables" array length equal to 1
    And the response should contain the field "path_variables.0.key" equal to "id"

  @apply_params
  Scenario: The query string is rebuilt from the enabled params in order
    When I call "RequestService.ApplyParams" with
      """
      {
        "request": {
          "url": "https://api.example.test/users?old=1",
          "params": [
            { "key": "page", "value": "2" },
            { "key": "debug", "value": "1", "enabled": false },
            { "key": "limit", "value": "10" }
          ]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "url" equal to "https://api.example.test/users?page=2&limit=10"
    And the response should contain the field "params" array length equal to 3
    And the response should contain the field "params.1.enabled" equal to "false"

  @apply_params
  Scenario: Unticking every param removes the question mark
    When I call "RequestService.ApplyParams" with
      """
      {
        "request": {
          "url": "https://api.example.test/users?page=2",
          "params": [ { "key": "page", "value": "2", "enabled": false } ]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "url" equal to "https://api.example.test/users"

  @apply_params @encoding
  Scenario: Values are written as typed, empty values as bare keys, blank rows skipped and the fragment kept
    When I call "RequestService.ApplyParams" with
      """
      {
        "request": {
          "url": "https://api.example.test/search?x=1#top",
          "params": [
            { "key": "flag", "value": "" },
            { "key": "q", "value": "a b" },
            { "key": "", "value": "" },
            { "key": "t", "value": "{{token}}" }
          ]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "url" equal to "https://api.example.test/search?flag&q=a b&t={{token}}#top"

  @apply_params @path_variables
  Scenario: Renaming a path variable in the table renames its URL segment
    When I call "RequestService.ApplyParams" with
      """
      {
        "request": {
          "url": "https://api.example.test/users/:id/posts/:postId?draft=true",
          "params": [ { "key": "draft", "value": "true" } ],
          "path_variables": [ { "key": "userId", "value": "1" }, { "key": "postId", "value": "2" } ]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "url" equal to "https://api.example.test/users/:userId/posts/:postId?draft=true"
    And the response should contain the field "path_variables.0.value" equal to "1"

  @round_trip
  Scenario: The rest of the request is returned unchanged
    When I call "RequestService.ApplyUrl" with
      """
      {
        "request": {
          "id": "01JA2Z7C2D3E4F5G6H7J8K9M0N",
          "name": "Create invoice",
          "method": "post",
          "url": "https://api.example.test/invoices",
          "headers": [ { "key": "Accept", "value": "application/json", "description": "json please" } ],
          "auto_headers": { "disabled": ["User-Agent"] },
          "auth": { "type": "bearer", "bearer": { "token": "abc" } },
          "body": { "mode": "raw", "raw": { "language": "json", "content": "{ \"amount\": 1000 }" } },
          "settings": { "timeout_ms": 5000, "follow_redirects": false }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "id" equal to "01JA2Z7C2D3E4F5G6H7J8K9M0N"
    And the response should contain the field "name" equal to "Create invoice"
    And the response should contain the field "method" equal to "POST"
    And the response should contain the field "headers.0.key" equal to "Accept"
    And the response should contain the field "headers.0.enabled" equal to "true"
    And the response should contain the field "headers.0.description" equal to "json please"
    And the response should contain the field "auto_headers.disabled.0" equal to "User-Agent"
    And the response should contain the field "auth.type" equal to "bearer"
    And the response should contain the field "auth.bearer.token" equal to "abc"
    And the response should contain the field "body.mode" equal to "raw"
    And the response should contain the field "body.raw.language" equal to "json"
    And the response should contain the field "body.raw.content" equal to
      """
      { "amount": 1000 }
      """
    And the response should contain the field "settings.timeout_ms" equal to "5000"
    And the response should contain the field "settings.follow_redirects" equal to "false"
    And the response should contain the field "settings.verify_ssl" equal to "true"
    And the response should contain the field "settings.encode_url" equal to "true"
