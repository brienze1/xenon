#language: en
#utf-8

@all @xen_005 @snippets @generation_golden
Feature: Code snippets match Postman's code generators
  The Code panel turns the current request into cURL, JavaScript fetch, Go net/http or Python
  requests following postman-code-generators 2.1.1: option names and defaults, flag order, body
  rendering per mode and quoting. Each request of the golden corpus (every body mode and auth
  type, auth materialised as headers or query parameters) is generated for every language and
  option set of resources/snippets/options.json and compared with the output of the reference
  generators (make snippet-golden). File paths stay as they are written in the request, relative
  to the workspace folder.

  Background:
    Given the workspace "Acme" exists

  @golden
  Scenario Outline: <case> as <language> with the <options> options
    When I call "SnippetService.Generate" with resource fields
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "<language>",
        "request": {{resource:snippets/xenon.yaml#requests.<case>}},
        "options": {{resource:snippets/options.json#<language>.<options>}}
      }
      """
    Then the call should succeed
    And the response field "text" should equal the resource file "snippets/expected/<case>.<language>.<options>.txt"

    Examples:
      | case              | language        | options    |
      | auth-apikey-query | curl            | default    |
      | auth-apikey-query | curl            | compact    |
      | auth-apikey-query | curl            | windows    |
      | auth-apikey-query | curl            | powershell |
      | auth-apikey-query | js_fetch        | default    |
      | auth-apikey-query | js_fetch        | options    |
      | auth-apikey-query | go_native       | default    |
      | auth-apikey-query | go_native       | options    |
      | auth-apikey-query | python_requests | default    |
      | auth-apikey-query | python_requests | options    |
      | auth-awsv4        | curl            | default    |
      | auth-awsv4        | curl            | compact    |
      | auth-awsv4        | curl            | windows    |
      | auth-awsv4        | curl            | powershell |
      | auth-awsv4        | js_fetch        | default    |
      | auth-awsv4        | js_fetch        | options    |
      | auth-awsv4        | go_native       | default    |
      | auth-awsv4        | go_native       | options    |
      | auth-awsv4        | python_requests | default    |
      | auth-awsv4        | python_requests | options    |
      | auth-basic        | curl            | default    |
      | auth-basic        | curl            | compact    |
      | auth-basic        | curl            | windows    |
      | auth-basic        | curl            | powershell |
      | auth-basic        | js_fetch        | default    |
      | auth-basic        | js_fetch        | options    |
      | auth-basic        | go_native       | default    |
      | auth-basic        | go_native       | options    |
      | auth-basic        | python_requests | default    |
      | auth-basic        | python_requests | options    |
      | auth-bearer       | curl            | default    |
      | auth-bearer       | curl            | compact    |
      | auth-bearer       | curl            | windows    |
      | auth-bearer       | curl            | powershell |
      | auth-bearer       | js_fetch        | default    |
      | auth-bearer       | js_fetch        | options    |
      | auth-bearer       | go_native       | default    |
      | auth-bearer       | go_native       | options    |
      | auth-bearer       | python_requests | default    |
      | auth-bearer       | python_requests | options    |
      | auth-digest       | curl            | default    |
      | auth-digest       | curl            | compact    |
      | auth-digest       | curl            | windows    |
      | auth-digest       | curl            | powershell |
      | auth-digest       | js_fetch        | default    |
      | auth-digest       | js_fetch        | options    |
      | auth-digest       | go_native       | default    |
      | auth-digest       | go_native       | options    |
      | auth-digest       | python_requests | default    |
      | auth-digest       | python_requests | options    |
      | auth-oauth2       | curl            | default    |
      | auth-oauth2       | curl            | compact    |
      | auth-oauth2       | curl            | windows    |
      | auth-oauth2       | curl            | powershell |
      | auth-oauth2       | js_fetch        | default    |
      | auth-oauth2       | js_fetch        | options    |
      | auth-oauth2       | go_native       | default    |
      | auth-oauth2       | go_native       | options    |
      | auth-oauth2       | python_requests | default    |
      | auth-oauth2       | python_requests | options    |
      | get-query         | curl            | default    |
      | get-query         | curl            | compact    |
      | get-query         | curl            | windows    |
      | get-query         | curl            | powershell |
      | get-query         | js_fetch        | default    |
      | get-query         | js_fetch        | options    |
      | get-query         | go_native       | default    |
      | get-query         | go_native       | options    |
      | get-query         | python_requests | default    |
      | get-query         | python_requests | options    |
      | post-formdata     | curl            | default    |
      | post-formdata     | curl            | compact    |
      | post-formdata     | curl            | windows    |
      | post-formdata     | curl            | powershell |
      | post-formdata     | js_fetch        | default    |
      | post-formdata     | js_fetch        | options    |
      | post-formdata     | go_native       | default    |
      | post-formdata     | go_native       | options    |
      | post-formdata     | python_requests | default    |
      | post-formdata     | python_requests | options    |
      | post-graphql      | curl            | default    |
      | post-graphql      | curl            | compact    |
      | post-graphql      | curl            | windows    |
      | post-graphql      | curl            | powershell |
      | post-graphql      | js_fetch        | default    |
      | post-graphql      | js_fetch        | options    |
      | post-graphql      | go_native       | default    |
      | post-graphql      | go_native       | options    |
      | post-graphql      | python_requests | default    |
      | post-graphql      | python_requests | options    |
      | post-raw-json     | curl            | default    |
      | post-raw-json     | curl            | compact    |
      | post-raw-json     | curl            | windows    |
      | post-raw-json     | curl            | powershell |
      | post-raw-json     | js_fetch        | default    |
      | post-raw-json     | js_fetch        | options    |
      | post-raw-json     | go_native       | default    |
      | post-raw-json     | go_native       | options    |
      | post-raw-json     | python_requests | default    |
      | post-raw-json     | python_requests | options    |
      | post-raw-text-at  | curl            | default    |
      | post-raw-text-at  | curl            | compact    |
      | post-raw-text-at  | curl            | windows    |
      | post-raw-text-at  | curl            | powershell |
      | post-raw-text-at  | js_fetch        | default    |
      | post-raw-text-at  | js_fetch        | options    |
      | post-raw-text-at  | go_native       | default    |
      | post-raw-text-at  | go_native       | options    |
      | post-raw-text-at  | python_requests | default    |
      | post-raw-text-at  | python_requests | options    |
      | post-urlencoded   | curl            | default    |
      | post-urlencoded   | curl            | compact    |
      | post-urlencoded   | curl            | windows    |
      | post-urlencoded   | curl            | powershell |
      | post-urlencoded   | js_fetch        | default    |
      | post-urlencoded   | js_fetch        | options    |
      | post-urlencoded   | go_native       | default    |
      | post-urlencoded   | go_native       | options    |
      | post-urlencoded   | python_requests | default    |
      | post-urlencoded   | python_requests | options    |
      | put-binary        | curl            | default    |
      | put-binary        | curl            | compact    |
      | put-binary        | curl            | windows    |
      | put-binary        | curl            | powershell |
      | put-binary        | js_fetch        | default    |
      | put-binary        | js_fetch        | options    |
      | put-binary        | go_native       | default    |
      | put-binary        | go_native       | options    |
      | put-binary        | python_requests | default    |
      | put-binary        | python_requests | options    |

  @languages
  Scenario: The four languages are offered with Postman's options and defaults
    When I call "SnippetService.Languages"
    Then the call should succeed
    And the response should contain the field "" array length equal to 4
    And the response should contain the field "0.id" equal to "curl"
    And the response should contain the field "1.id" equal to "js_fetch"
    And the response should contain the field "2.id" equal to "go_native"
    And the response should contain the field "3.id" equal to "python_requests"
    And the response field "0.options" should contain an item with "id" equal to "multiLine" and "default" equal to "true"
    And the response field "0.options" should contain an item with "id" equal to "longFormat" and "default" equal to "true"
    And the response field "0.options" should contain an item with "id" equal to "lineContinuationCharacter" and "default" equal to "\"
    And the response field "0.options" should contain an item with "id" equal to "quoteType" and "default" equal to "single"
    And the response field "0.options" should contain an item with "id" equal to "requestTimeoutInSeconds" and "default" equal to "0"
    And the response field "0.options" should contain an item with "id" equal to "followRedirect" and "default" equal to "true"
    And the response field "0.options" should contain an item with "id" equal to "trimRequestBody" and "default" equal to "false"
    And the response field "1.options" should contain an item with "id" equal to "requestTimeout" and "default" equal to "0"
    And the response field "1.options" should contain an item with "id" equal to "indentCount" and "default" equal to "2"
    And the response field "2.options" should contain an item with "id" equal to "indentType" and "default" equal to "Space"
    And the response field "3.options" should contain an item with "id" equal to "followRedirect" and "default" equal to "true"
    And the response field "0.options" should contain an item with "id" equal to "multiLine" and "description" equal to "Split cURL command across multiple lines"
    And the response field "0.options" should contain an item with "id" equal to "silent" and "description" equal to "Display the requested data without showing the cURL progress meter or error messages"
    And the response field "1.options" should contain an item with "id" equal to "requestTimeout" and "description" equal to "Set number of milliseconds the request should wait for a response before timing out (use 0 for infinity)"
    And the response field "2.options" should contain an item with "id" equal to "indentType" and "description" equal to "Select the character used to indent lines of code"
    And the response field "3.options" should contain an item with "id" equal to "trimRequestBody" and "description" equal to "Remove white space and additional lines that may affect the server's response"

  @options
  Scenario: Options that are not valid for the language fall back to their defaults
    When I call "SnippetService.Generate" with resource fields
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "curl",
        "request": {{resource:snippets/xenon.yaml#requests.auth-basic}},
        "options": { "lineContinuationCharacter": "|", "quoteType": "backtick", "requestTimeoutInSeconds": -4, "multiLine": "yes", "unknown": true }
      }
      """
    Then the call should succeed
    And the response field "text" should equal the resource file "snippets/expected/auth-basic.curl.default.txt"

  @validation
  Scenario Outline: An invalid snippet request is refused - <case>
    When I call "SnippetService.Generate" with
      """
      <payload>
      """
    Then the call should fail with code "<code>"

    Examples:
      | case              | payload                                                                                                    | code      |
      | no workspace      | { "local_id": "", "language": "curl", "request": { "method": "GET", "url": "https://x.test" } }             | SNP-00400 |
      | unknown language  | { "local_id": "{{Acme.local_id}}", "language": "cobol", "request": { "method": "GET", "url": "https://x.test" } } | SNP-00400 |
      | unknown workspace | { "local_id": "01JQ0000000000000000000098", "language": "curl", "request": { "method": "GET", "url": "https://x.test" } } | WKS-07404 |
