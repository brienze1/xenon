#language: en
#utf-8

@all @xen_001 @requests @bulk_edit
Feature: Bulk edit key/value tables as text
  Every key/value table can switch to a text mode with one "key:value" row per line; a "//"
  prefix marks a disabled row. Blank lines are ignored and only the first ":" separates the key
  from the value.

  @parse
  Scenario: Lines become enabled rows in order
    When I call "RequestService.ParseBulk" with
      """
      { "text": "page:2\nlimit:10" }
      """
    Then the call should succeed
    And the response should contain the field "rows" array length equal to 2
    And the response should contain the field "rows.0.key" equal to "page"
    And the response should contain the field "rows.0.value" equal to "2"
    And the response should contain the field "rows.0.enabled" equal to "true"
    And the response should contain the field "rows.1.key" equal to "limit"
    And the response should contain the field "rows.1.value" equal to "10"

  @parse
  Scenario: A "//" prefix marks a disabled row
    When I call "RequestService.ParseBulk" with
      """
      { "text": "//debug:1\n// verbose:yes" }
      """
    Then the call should succeed
    And the response should contain the field "rows" array length equal to 2
    And the response should contain the field "rows.0.key" equal to "debug"
    And the response should contain the field "rows.0.value" equal to "1"
    And the response should contain the field "rows.0.enabled" equal to "false"
    And the response should contain the field "rows.1.key" equal to "verbose"
    And the response should contain the field "rows.1.enabled" equal to "false"

  @parse
  Scenario: Only the first colon separates key and value
    When I call "RequestService.ParseBulk" with
      """
      { "text": "url:https://api.example.test:8443/a\nAuthorization:Bearer {{token}}" }
      """
    Then the call should succeed
    And the response should contain the field "rows.0.key" equal to "url"
    And the response should contain the field "rows.0.value" equal to "https://api.example.test:8443/a"
    And the response should contain the field "rows.1.key" equal to "Authorization"
    And the response should contain the field "rows.1.value" equal to "Bearer {{token}}"

  @parse
  Scenario: Blank lines are skipped, spaces around keys and values are trimmed and a line without a colon is a key
    When I call "RequestService.ParseBulk" with
      """
      { "text": "\n  page : 2  \r\n\n   \nflag\n" }
      """
    Then the call should succeed
    And the response should contain the field "rows" array length equal to 2
    And the response should contain the field "rows.0.key" equal to "page"
    And the response should contain the field "rows.0.value" equal to "2"
    And the response should contain the field "rows.1.key" equal to "flag"
    And the response should contain the field "rows.1.value" equal to ""
    And the response should contain the field "rows.1.enabled" equal to "true"

  @parse
  Scenario: Empty text gives no rows
    When I call "RequestService.ParseBulk" with
      """
      { "text": "" }
      """
    Then the call should succeed
    And the response should contain the field "rows" array length equal to 0

  @format
  Scenario: Rows are written one per line with disabled rows prefixed
    When I call "RequestService.FormatBulk" with
      """
      {
        "rows": [
          { "key": "page", "value": "2", "description": "dropped in bulk mode" },
          { "key": "debug", "value": "1", "enabled": false },
          { "key": "url", "value": "a:b" },
          { "key": "", "value": "" },
          { "key": "flag", "value": "" }
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "text" equal to
      """
      page:2
      //debug:1
      url:a:b
      flag:
      """

  @format
  Scenario: No rows give empty text
    When I call "RequestService.FormatBulk" with
      """
      { "rows": [] }
      """
    Then the call should succeed
    And the response should contain the field "text" equal to ""
