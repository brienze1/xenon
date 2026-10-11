#language: en
#utf-8

@all @xen_005 @auth @awsv4
Feature: AWS Signature Version 4
  AWS Signature signs the request with the AWS SDK's SigV4 signer after every header is final,
  as the last auth step: stale Authorization and X-Amz-* values are removed first, keys may come
  from variables, the region defaults to us-east-1 and the service to execute-api. Header mode adds
  X-Amz-Date, X-Amz-Security-Token (with a session token) and Authorization (plus
  X-Amz-Content-Sha256 for S3); query mode presigns the URL. Form-data bodies are signed as an empty
  payload, as Postman does. The signatures match AWS's SigV4 test suite for every vector that
  applies to an HTTP client (path normalisation, single path encoding, line folding, repeated
  header names - Xenon sends one header per name - and the header-after-signing token case do not
  apply; see the spec's assumptions).

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the current time is "2015-08-30T12:36:00Z"

  @vectors
  Scenario Outline: The signature of test-suite vector <vector> matches AWS's
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": <request> }
      """
    Then the call should succeed
    And the 1 "<method>" request to "<path>" should have the header "Authorization" equal to the resource file "<authz>"
    And the 1 "<method>" request to "<path>" should have the header "X-Amz-Date" equal to "20150830T123600Z"

    Examples:
      | vector | method | path | authz | request |
      | get-header-value-trim | GET | / | aws-sigv4-test-suite/get-header-value-trim/get-header-value-trim.authz | {"method": "GET", "url": "{{api}}/", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "My-Header1", "value": " value1", "enabled": true}, {"key": "My-Header2", "value": " \"a   b   c\"", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | get-unreserved | GET | /-._~0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz | aws-sigv4-test-suite/get-unreserved/get-unreserved.authz | {"method": "GET", "url": "{{api}}/-._~0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | get-vanilla | GET | / | aws-sigv4-test-suite/get-vanilla/get-vanilla.authz | {"method": "GET", "url": "{{api}}/", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | get-vanilla-empty-query-key | GET | / | aws-sigv4-test-suite/get-vanilla-empty-query-key/get-vanilla-empty-query-key.authz | {"method": "GET", "url": "{{api}}/?Param1=value1", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | get-vanilla-query | GET | / | aws-sigv4-test-suite/get-vanilla-query/get-vanilla-query.authz | {"method": "GET", "url": "{{api}}/", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | get-vanilla-query-order-key | GET | / | aws-sigv4-test-suite/get-vanilla-query-order-key/get-vanilla-query-order-key.authz | {"method": "GET", "url": "{{api}}/?Param1=value2&Param1=Value1", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | get-vanilla-query-order-key-case | GET | / | aws-sigv4-test-suite/get-vanilla-query-order-key-case/get-vanilla-query-order-key-case.authz | {"method": "GET", "url": "{{api}}/?Param2=value2&Param1=value1", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | get-vanilla-query-order-value | GET | / | aws-sigv4-test-suite/get-vanilla-query-order-value/get-vanilla-query-order-value.authz | {"method": "GET", "url": "{{api}}/?Param1=value2&Param1=value1", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | get-vanilla-query-unreserved | GET | / | aws-sigv4-test-suite/get-vanilla-query-unreserved/get-vanilla-query-unreserved.authz | {"method": "GET", "url": "{{api}}/?-._~0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz=-._~0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | get-vanilla-utf8-query | GET | / | aws-sigv4-test-suite/get-vanilla-utf8-query/get-vanilla-utf8-query.authz | {"method": "GET", "url": "{{api}}/?ሴ=bar", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | get-vanilla-with-session-token | GET | / | aws-sigv4-test-suite/get-vanilla-with-session-token/get-vanilla-with-session-token.authz | {"method": "GET", "url": "{{api}}/", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "6e86291e8372ff2a2260956d9b8aae1d763fbf315fa00fa31553b73ebf194267", "region": "us-east-1", "service": "service"}}} |
      | post-header-key-case | POST | / | aws-sigv4-test-suite/post-header-key-case/post-header-key-case.authz | {"method": "POST", "url": "{{api}}/", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | post-header-key-sort | POST | / | aws-sigv4-test-suite/post-header-key-sort/post-header-key-sort.authz | {"method": "POST", "url": "{{api}}/", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "My-Header1", "value": "value1", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | post-header-value-case | POST | / | aws-sigv4-test-suite/post-header-value-case/post-header-value-case.authz | {"method": "POST", "url": "{{api}}/", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "My-Header1", "value": "VALUE1", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | post-vanilla | POST | / | aws-sigv4-test-suite/post-vanilla/post-vanilla.authz | {"method": "POST", "url": "{{api}}/", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | post-vanilla-empty-query-value | POST | / | aws-sigv4-test-suite/post-vanilla-empty-query-value/post-vanilla-empty-query-value.authz | {"method": "POST", "url": "{{api}}/?Param1=value1", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | post-vanilla-query | POST | / | aws-sigv4-test-suite/post-vanilla-query/post-vanilla-query.authz | {"method": "POST", "url": "{{api}}/?Param1=value1", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}} |
      | post-x-www-form-urlencoded | POST | / | aws-sigv4-test-suite/post-x-www-form-urlencoded/post-x-www-form-urlencoded.authz | {"method": "POST", "url": "{{api}}/", "headers": [{"key": "Content-Type", "value": "application/x-www-form-urlencoded", "enabled": true}, {"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}, "body": {"mode": "raw", "raw": {"language": "text", "content": "Param1=value1"}}} |
      | post-x-www-form-urlencoded-parameters | POST | / | aws-sigv4-test-suite/post-x-www-form-urlencoded-parameters/post-x-www-form-urlencoded-parameters.authz | {"method": "POST", "url": "{{api}}/", "headers": [{"key": "Content-Type", "value": "application/x-www-form-urlencoded; charset=utf8", "enabled": true}, {"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "", "region": "us-east-1", "service": "service"}}, "body": {"mode": "raw", "raw": {"language": "text", "content": "Param1=value1"}}} |
      | post-sts-token/post-sts-header-before | POST | / | aws-sigv4-test-suite/post-sts-token/post-sts-header-before/post-sts-header-before.authz | {"method": "POST", "url": "{{api}}/", "headers": [{"key": "Host", "value": "example.amazonaws.com", "enabled": true}, {"key": "X-Amz-Date", "value": "20150830T123600Z", "enabled": true}], "auto_headers": {"disabled": ["User-Agent", "Accept", "Accept-Encoding", "Connection", "Content-Type"]}, "settings": {"timeout_ms": 0}, "auth": {"type": "awsv4", "awsv4": {"access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "session_token": "AQoDYXdzEPT//////////wEXAMPLEtc764bNrC9SAPBSM22wDOk4x4HIZ8j4FZTwdQWLWsKWHGBuFqwAeMicRXmxfpSPfIeoIYRqTflfKD8YUuwthAx7mSEI/qkPpKPi/kMcGdQrmGdeehM4IC1NtBmUpp2wUE8phUZampKsburEDy0KPkyQDYwT7WZ0wq5VSXDvp75YU9HFvlRd8Tx6q6fE8YQcHNVXAkiY9q6d+xo0rKwT38xVqr7ZD0u0iPPkUL64lIZbqBAz+scqKmlzm8FDrypNC9Yjc8fPOLn9FX9KSYvKTr4rvx3iSIlTJabIQwj2ICCR/oLxBA==", "region": "us-east-1", "service": "service"}}} |

  @defaults @variables @session_token
  Scenario: Keys from variables, default region and service, a session token and stale values removed
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JH0000000000000000000000
      name: Acme
      variables:
        - key: aws_key
          value: AKIDEXAMPLE
          enabled: true
          description: ""
        - key: aws_secret
          value: wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY
          enabled: true
          description: ""
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/prod/items?X-Amz-Signature=stale&page=1",
          "headers": [ { "key": "Authorization", "value": "Bearer stale", "enabled": true }, { "key": "X-Amz-Date", "value": "19990101T000000Z", "enabled": true } ],
          "auth": { "type": "awsv4", "awsv4": { "access_key": "{{aws_key}}", "secret_key": "{{aws_secret}}", "session_token": "SESSION" } }
        }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/prod/items" should have the header "Authorization" matching "^AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20150830/us-east-1/execute-api/aws4_request, SignedHeaders=[a-z0-9;-]*host;[a-z0-9;-]*x-amz-date;x-amz-security-token, Signature=[0-9a-f]{64}$"
    And the 1 "GET" request to "/prod/items" should have the header "X-Amz-Date" equal to "20150830T123600Z"
    And the 1 "GET" request to "/prod/items" should have the header "X-Amz-Security-Token" equal to "SESSION"
    And the 1 "GET" request to "/prod/items" should have the raw query "page=1"
    And the 1 "GET" request to "/prod/items" should not have the header "X-Amz-Content-Sha256"

  @query
  Scenario: Query mode presigns the URL
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/prod/items?page=1",
          "auth": { "type": "awsv4", "awsv4": { "access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "region": "eu-west-1", "service": "execute-api", "add_auth_data_to_query": true } }
        }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/prod/items" should have the query "X-Amz-Algorithm" equal to "AWS4-HMAC-SHA256"
    And the 1 "GET" request to "/prod/items" should have the query "X-Amz-Credential" equal to "AKIDEXAMPLE/20150830/eu-west-1/execute-api/aws4_request"
    And the 1 "GET" request to "/prod/items" should have the query "X-Amz-Date" equal to "20150830T123600Z"
    And the 1 "GET" request to "/prod/items" should have the query "X-Amz-Expires" equal to "86400"
    And the 1 "GET" request to "/prod/items" should have the query "X-Amz-Signature" matching "^[0-9a-f]{64}$"
    And the 1 "GET" request to "/prod/items" should have the query "page" equal to "1"
    And the 1 "GET" request to "/prod/items" should not have the header "Authorization"

  @s3 @payload
  Scenario: S3 requests carry the payload hash header
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "PUT",
          "url": "{{api}}/bucket/key.txt",
          "body": { "mode": "raw", "raw": { "language": "text", "content": "hello" } },
          "auth": { "type": "awsv4", "awsv4": { "access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "service": "s3" } }
        }
      }
      """
    Then the call should succeed
    And the 1 "PUT" request to "/bucket/key.txt" should have the header "X-Amz-Content-Sha256" equal to "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
    And the 1 "PUT" request to "/bucket/key.txt" should have the header "Authorization" matching "x-amz-content-sha256"

  @payload @formdata
  Scenario: A form-data body is signed as an empty payload
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "POST",
          "url": "{{api}}/bucket/upload",
          "body": { "mode": "formdata", "formdata": [ { "key": "note", "value": "hello", "type": "text", "enabled": true } ] },
          "auth": { "type": "awsv4", "awsv4": { "access_key": "AKIDEXAMPLE", "secret_key": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "service": "s3" } }
        }
      }
      """
    Then the call should succeed
    And the 1 "POST" request to "/bucket/upload" should have the header "X-Amz-Content-Sha256" equal to "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    And the 1 "POST" request to "/bucket/upload" should have the form field "note" equal to "hello"
