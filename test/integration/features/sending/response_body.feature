#language: en
#utf-8

@all @xen_001 @sending @response_body
Feature: Response bodies
  Small text bodies come back inline; any body is served with its exact bytes and content type
  by the in-app response route. Each tab keeps its last result while the app runs, until the
  tab releases it.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url

  @inline
  Scenario: A small text body is returned inline and served by the route
    Given the 1 "GET" request to "/doc" returns status 200 with content type "text/plain; charset=utf-8" and body
      """
      hello, xenon
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/doc" } }
      """
    Then the call should succeed
    And the response should contain the field "inline_body" equal to "hello, xenon"
    And the response should contain the field "size_bytes" equal to "12"
    And the response should contain the field "is_binary" equal to "false"
    And the response should contain the field "content_type" equal to "text/plain; charset=utf-8"
    When I save the response field "body_url" as "body_url"
    And I request "{{body_url}}" from the "ResponseBodyRoute" route
    Then the route response status should be 200
    And the route response header "Content-Type" should be "text/plain; charset=utf-8"
    And the route response body should be
      """
      hello, xenon
      """

  @large
  Scenario: A body over 1 MB is not inlined and the route serves the exact bytes
    Given the 1 "GET" request to "/big" returns status 200 with content type "text/plain" and a generated body of 8388608 bytes
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-big", "request": { "method": "GET", "url": "{{api}}/big" } }
      """
    Then the call should succeed
    And the response should contain the field "size_bytes" equal to "8388608"
    And the response should contain the field "is_binary" equal to "false"
    And the response field "inline_body" should not exist
    And the response should contain the field "body_url" equal to "/xenon/response/tab-big"
    When I request "/xenon/response/tab-big" from the "ResponseBodyRoute" route
    Then the route response status should be 200
    And the route response header "Content-Type" should be "text/plain"
    And the route response body should be the generated body of 8388608 bytes

  @large
  Scenario: A body between 1 MB and 5 MB is not inlined and still served
    Given the 1 "GET" request to "/medium" returns status 200 with content type "application/json" and a generated body of 2097152 bytes
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/medium" } }
      """
    Then the call should succeed
    And the response field "inline_body" should not exist
    When I request "/xenon/response/tab-1" from the "ResponseBodyRoute" route
    Then the route response status should be 200
    And the route response body should be the generated body of 2097152 bytes

  @binary
  Scenario: A binary body is flagged and never inlined
    Given the 1 "GET" request to "/logo" returns status 200 with content type "image/png" and the resource file "pixel.png"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/logo" } }
      """
    Then the call should succeed
    And the response should contain the field "is_binary" equal to "true"
    And the response should contain the field "content_type" equal to "image/png"
    And the response field "inline_body" should not exist
    When I request "/xenon/response/tab-1" from the "ResponseBodyRoute" route
    Then the route response header "Content-Type" should be "image/png"
    And the route response body should be the resource file "pixel.png"

  @compressed
  Scenario: A gzip body is decoded while the size is the size on the wire
    Given the 1 "GET" request to "/zipped" returns status 200 with content type "text/plain" and the resource file "hello.txt.gz"
    And the requests to "/zipped" respond with the header "Content-Encoding" equal to "gzip"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/zipped" } }
      """
    Then the call should succeed
    And the response should contain the field "inline_body" equal to "hello, xenon"
    And the response should contain the field "size_bytes" equal to "32"
    And the response field "headers" should contain an item with "key" equal to "Content-Encoding" and "value" equal to "gzip"

  @save
  Scenario: Saving the body writes the exact bytes
    Given the 1 "GET" request to "/logo" returns status 200 with content type "image/png" and the resource file "pixel.png"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/logo" } }
      """
    And I call "RequestService.SaveResponseBody" with
      """
      { "tab_id": "tab-1", "path": "{{temp_root}}/out/logo.png" }
      """
    Then the call should succeed
    And the file "out/logo.png" should equal the resource file "pixel.png"

  @save @large
  Scenario: Saving a large body writes the exact bytes
    Given the 1 "GET" request to "/big" returns status 200 with content type "text/plain" and a generated body of 6291456 bytes
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/big" } }
      """
    And I call "RequestService.SaveResponseBody" with
      """
      { "tab_id": "tab-1", "path": "{{temp_root}}/out/big.txt" }
      """
    Then the call should succeed
    And the file "out/big.txt" should be the generated body of 6291456 bytes

  @save @validation
  Scenario Outline: Saving needs a tab and an absolute path - <case>
    When I call "RequestService.SaveResponseBody" with
      """
      { "tab_id": "<tab_id>", "path": "<path>" }
      """
    Then the call should fail with code "SND-04400"

    Examples:
      | case          | tab_id | path                 |
      | no tab        |        | {{temp_root}}/x.bin  |
      | relative path | tab-1  | out/x.bin            |

  @save @not_found
  Scenario: Saving without a response fails
    When I call "RequestService.SaveResponseBody" with
      """
      { "tab_id": "tab-none", "path": "{{temp_root}}/out/x.bin" }
      """
    Then the call should fail with code "SND-03404"

  @last_result
  Scenario: The last result of a tab is returned again
    Given the 1 "GET" request to "/doc" returns status 201 with content type "text/plain" and body
      """
      first
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/doc" } }
      """
    And I call "RequestService.LastResult" with
      """
      { "tab_id": "tab-1" }
      """
    Then the call should succeed
    And the response should contain the field "tab_id" equal to "tab-1"
    And the response should contain the field "ok" equal to "true"
    And the response should contain the field "status_code" equal to "201"
    And the response should contain the field "status_text" equal to "Created"
    And the response should contain the field "inline_body" equal to "first"
    And the response should contain the field "size_bytes" equal to "5"
    And the response field "timings.total_ms" should exist

  @last_result
  Scenario: Each tab keeps its own result and a new send replaces the old one
    Given the 1 "GET" request to "/doc" returns status 200 with content type "text/plain" and body
      """
      first
      """
    And the 2 "GET" request to "/doc" returns status 200 with content type "text/plain" and body
      """
      second
      """
    And the 3 "GET" request to "/doc" returns status 200 with content type "text/plain" and body
      """
      third
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/doc" } }
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-2", "request": { "method": "GET", "url": "{{api}}/doc" } }
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/doc" } }
      """
    And I call "RequestService.LastResult" with
      """
      { "tab_id": "tab-1" }
      """
    Then the response should contain the field "inline_body" equal to "third"
    When I call "RequestService.LastResult" with
      """
      { "tab_id": "tab-2" }
      """
    Then the response should contain the field "inline_body" equal to "second"

  @last_result @not_found
  Scenario: A tab without a result
    When I call "RequestService.LastResult" with
      """
      { "tab_id": "tab-none" }
      """
    Then the call should fail with code "SND-03404"
    When I request "/xenon/response/tab-none" from the "ResponseBodyRoute" route
    Then the route response status should be 404

  @last_result @validation
  Scenario: Asking for a result needs a tab
    When I call "RequestService.LastResult" with
      """
      { "tab_id": "" }
      """
    Then the call should fail with code "SND-01400"

  @release
  Scenario Outline: Releasing a response frees it - <case>
    Given the 1 "GET" request to "/doc" returns status 200 with content type "text/plain" and a generated body of <size> bytes
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/doc" } }
      """
    And I call "RequestService.ReleaseResponse" with
      """
      { "tab_id": "tab-1" }
      """
    Then the call should succeed
    When I call "RequestService.LastResult" with
      """
      { "tab_id": "tab-1" }
      """
    Then the call should fail with code "SND-03404"
    When I request "/xenon/response/tab-1" from the "ResponseBodyRoute" route
    Then the route response status should be 404

    Examples:
      | case             | size    |
      | body in memory   | 1024    |
      | body in a file   | 6291456 |

  @release
  Scenario: Releasing a tab without a response is harmless, without a tab id it is rejected
    When I call "RequestService.ReleaseResponse" with
      """
      { "tab_id": "tab-none" }
      """
    Then the call should succeed
    When I call "RequestService.ReleaseResponse" with
      """
      { "tab_id": "" }
      """
    Then the call should fail with code "SND-02400"
