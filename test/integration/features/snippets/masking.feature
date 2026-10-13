#language: en
#utf-8

@all @xen_005 @snippets @masking
Feature: Snippets materialise auth and mask secrets
  A snippet shows the current request as it would be sent: the tab's unsaved edits, variables
  resolved, the headers and auth inherited from the collection and folders, and the auth turned
  into headers or query parameters (Basic, Bearer, API key, the OAuth 2.0 token from the vault,
  AWS Signature v4, Digest with the challenge Xenon learned). Secret values show as ••••• unless
  "Include secrets" is on, and so do the credentials computed from them (a Basic credential, an
  AWS signature, a Digest response). Copy puts exactly the text shown on the clipboard. Auth that
  cannot be shown is left out with a comment line and a notice.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the "oauth" variable is set to the mocked oauth provider url
    And the "billing" variable is set to "01JR0000000000000000000001"
    And the "orders" variable is set to "01JR0000000000000000000002"
    And the "dev" variable is set to "01JR0000000000000000000009"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JR0000000000000000000000
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
      headers:
        - key: X-Tenant
          value: acme
          enabled: true
      auth:
        type: bearer
        bearer:
          token: "{{token}}"
      """
    And the workspace file "acme/collections/billing/orders.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{orders}}
      name: Orders
      method: GET
      url: "{{base_url}}/orders?key={{api_key}}"
      auth:
        type: inherit
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: base_url
          value: https://api.example.test
          enabled: true
          description: ""
        - key: token
          secret: true
          enabled: true
          description: ""
        - key: api_key
          secret: true
          enabled: true
          description: ""
        - key: password
          secret: true
          enabled: true
          description: ""
        - key: aws_secret
          secret: true
          enabled: true
          description: ""
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "token", "value": "t0k3n-s3cr3t" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key", "value": "k3y s3cr3t" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "password", "value": "p4ss-s3cr3t" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "aws_secret", "value": "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY" }
      """

  @masked
  Scenario: Secrets are masked by default; variables, inherited headers and auth are applied
    Given the log output is captured
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "curl",
        "request": { "id": "{{orders}}", "method": "GET", "url": "{{base_url}}/orders?key={{api_key}}", "auth": { "type": "inherit" } }
      }
      """
    Then the call should succeed
    And the response should contain the field "text" equal to
      """
      curl --location 'https://api.example.test/orders?key=•••••' \
      --header 'X-Tenant: acme' \
      --header 'Authorization: Bearer •••••'
      """
    And the response should contain the field "notices" array length equal to 0
    And the response should not contain "t0k3n-s3cr3t"
    And the response should not contain "k3y"
    And the log output should not contain "t0k3n-s3cr3t"
    And the log output should not contain "k3y s3cr3t"

  @include_secrets
  Scenario Outline: Include secrets shows the real values - <language>
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "<language>",
        "include_secrets": true,
        "request": { "id": "{{orders}}", "method": "GET", "url": "{{base_url}}/orders?key={{api_key}}", "auth": { "type": "inherit" } }
      }
      """
    Then the call should succeed
    And the response should contain the field "text" containing "Bearer t0k3n-s3cr3t"
    And the response should contain the field "text" containing "<query>"
    And the response should contain the field "text" containing "acme"

    Examples:
      | language        | query                  |
      | curl            | key=k3y%20s3cr3t   |
      | js_fetch        | key=k3y%20s3cr3t   |
      | go_native       | key=k3y%20s3cr3t   |
      | python_requests | key=k3y%20s3cr3t   |

  @masked
  Scenario Outline: Every language masks the secrets - <language>
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "<language>",
        "request": { "id": "{{orders}}", "method": "GET", "url": "{{base_url}}/orders?key={{api_key}}", "auth": { "type": "inherit" } }
      }
      """
    Then the call should succeed
    And the response should contain the field "text" containing "Bearer •••••"
    And the response should contain the field "text" containing "key=•••••"
    And the response should not contain "t0k3n"
    And the response should not contain "k3y"

    Examples:
      | language        |
      | curl            |
      | js_fetch        |
      | go_native       |
      | python_requests |

  @draft
  Scenario: The tab's unsaved edits are what the snippet shows
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "curl",
        "options": { "multiLine": false },
        "request": {
          "id": "{{orders}}",
          "method": "POST",
          "url": "{{base_url}}/orders/draft",
          "headers": [ { "key": "X-Tenant", "value": "globex", "enabled": true }, { "key": "X-Off", "value": "no", "enabled": false } ],
          "auth": { "type": "none" },
          "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"a\":1}" } }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "text" equal to
      """
      curl --location 'https://api.example.test/orders/draft' --header 'X-Tenant: globex' --header 'Content-Type: application/json' --data '{"a":1}'
      """

  @basic
  Scenario: A Basic credential built from a secret is masked as a whole
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "curl",
        "options": { "multiLine": false },
        "request": { "method": "GET", "url": "{{base_url}}/me", "auth": { "type": "basic", "basic": { "username": "ana", "password": "{{password}}" } } }
      }
      """
    Then the call should succeed
    And the response should contain the field "text" equal to "curl --location 'https://api.example.test/me' --header 'Authorization: Basic •••••'"
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "curl",
        "include_secrets": true,
        "options": { "multiLine": false },
        "request": { "method": "GET", "url": "{{base_url}}/me", "auth": { "type": "basic", "basic": { "username": "ana", "password": "{{password}}" } } }
      }
      """
    Then the response should contain the field "text" equal to "curl --location 'https://api.example.test/me' --header 'Authorization: Basic YW5hOnA0c3MtczNjcjN0'"

  @basic @literal
  Scenario: Auth typed literally in the request is not a secret
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "curl",
        "options": { "multiLine": false },
        "request": { "method": "GET", "url": "{{base_url}}/me", "auth": { "type": "apikey", "apikey": { "key": "X-Api-Key", "value": "plain-key", "in": "header" } } }
      }
      """
    Then the response should contain the field "text" equal to "curl --location 'https://api.example.test/me' --header 'X-Api-Key: plain-key'"

  @copy
  Scenario: Copy puts exactly the shown text on the clipboard
    When I call "SnippetService.Copy" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "curl",
        "options": { "multiLine": false },
        "request": { "id": "{{orders}}", "method": "GET", "url": "{{base_url}}/orders?key={{api_key}}", "auth": { "type": "inherit" } }
      }
      """
    Then the call should succeed
    And the clipboard should contain "curl --location 'https://api.example.test/orders?key=•••••' --header 'X-Tenant: acme' --header 'Authorization: Bearer •••••'"
    When I call "SnippetService.Copy" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "curl",
        "include_secrets": true,
        "options": { "multiLine": false },
        "request": { "id": "{{orders}}", "method": "GET", "url": "{{base_url}}/orders?key={{api_key}}", "auth": { "type": "inherit" } }
      }
      """
    Then the clipboard should contain "curl --location 'https://api.example.test/orders?key=k3y%20s3cr3t' --header 'X-Tenant: acme' --header 'Authorization: Bearer t0k3n-s3cr3t'"

  @oauth2
  Scenario: The OAuth 2.0 token comes from the vault and is masked
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      auth:
        type: oauth2
        oauth2:
          grant_type: client_credentials
          accessTokenUrl: "{{oauth}}/token"
          clientId: xenon
          clientSecret: s3cret
          headerPrefix: Token
      """
    And I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{billing}}" }
      """
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "python_requests",
        "request": { "id": "{{orders}}", "method": "GET", "url": "{{base_url}}/orders", "auth": { "type": "inherit" } }
      }
      """
    Then the call should succeed
    And the response should contain the field "text" containing "'Authorization': 'Token •••••'"
    And the response should not contain "access-1"
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "python_requests",
        "include_secrets": true,
        "request": { "id": "{{orders}}", "method": "GET", "url": "{{base_url}}/orders", "auth": { "type": "inherit" } }
      }
      """
    Then the response should contain the field "text" containing "'Authorization': 'Token access-1'"

  @oauth2 @notice
  Scenario: Without a token the OAuth 2.0 auth is left out with a notice
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      auth:
        type: oauth2
        oauth2:
          grant_type: client_credentials
          accessTokenUrl: "{{oauth}}/token"
          clientId: xenon
      """
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "js_fetch",
        "request": { "id": "{{orders}}", "method": "GET", "url": "{{base_url}}/orders", "auth": { "type": "inherit" } }
      }
      """
    Then the call should succeed
    And the response should contain the field "notices" with the items "No OAuth 2.0 access token yet: get one in the Authorization tab."
    And the response should contain the field "text" matching "^// No OAuth 2.0 access token yet: get one in the Authorization tab.\nconst requestOptions"

  @awsv4
  Scenario: An AWS signature computed from a secret key is masked
    Given the current time is "2026-10-11T10:00:00Z"
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "curl",
        "request": { "method": "GET", "url": "https://abc.execute-api.us-east-1.amazonaws.com/prod/items", "auth": { "type": "awsv4", "awsv4": { "access_key": "AKIDEXAMPLE", "secret_key": "{{aws_secret}}", "region": "us-east-1", "service": "execute-api" } } }
      }
      """
    Then the call should succeed
    And the response should contain the field "text" matching "--header 'X-Amz-Date: 20261011T1000[0-9]{2}Z'"
    And the response should contain the field "text" matching "--header 'Authorization: AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20261011/us-east-1/execute-api/aws4_request, SignedHeaders=host;x-amz-date, Signature=•••••'"
    And the response should not contain "wJalrXUtnFEMI"
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "curl",
        "include_secrets": true,
        "request": { "method": "GET", "url": "https://abc.execute-api.us-east-1.amazonaws.com/prod/items", "auth": { "type": "awsv4", "awsv4": { "access_key": "AKIDEXAMPLE", "secret_key": "{{aws_secret}}", "region": "us-east-1", "service": "execute-api" } } }
      }
      """
    Then the response should contain the field "text" matching "Signature=[0-9a-f]{64}'"

  @digest @notice
  Scenario: Digest without a known challenge is left out with a notice
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "curl",
        "options": { "multiLine": false },
        "request": { "method": "GET", "url": "{{base_url}}/digest", "auth": { "type": "digest", "digest": { "username": "ana", "password": "{{password}}" } } }
      }
      """
    Then the call should succeed
    And the response should contain the field "notices" with the items "Digest: send the request once so Xenon learns the server's challenge."
    And the response should contain the field "text" equal to
      """
      # Digest: send the request once so Xenon learns the server's challenge.
      curl --location 'https://api.example.test/digest'
      """

  @digest
  Scenario: Digest answers the typed challenge and masks a response computed from a secret
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "curl",
        "options": { "multiLine": false },
        "request": { "method": "GET", "url": "{{base_url}}/digest", "auth": { "type": "digest", "digest": { "username": "ana", "password": "{{password}}", "realm": "xenon@test", "nonce": "n1", "algorithm": "MD5" } } }
      }
      """
    Then the call should succeed
    And the response should contain the field "notices" array length equal to 0
    And the response should contain the field "text" matching "^curl --location 'https://api.example.test/digest' --header 'Authorization: Digest username=.ana., realm=.xenon@test., nonce=.n1., uri=./digest., .*response=.•••••."
    And the response should not contain "p4ss"

  @unsupported @notice
  Scenario Outline: Auth types Xenon cannot send are left out with a comment line - <language>
    When I call "SnippetService.Generate" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "language": "<language>",
        "request": { "method": "GET", "url": "{{base_url}}/hawk", "auth": { "type": "unsupported", "unsupported": { "source": "postman", "type": "hawk" } } }
      }
      """
    Then the call should succeed
    And the response should contain the field "notices" with the items "Auth type hawk is not supported, so the snippet has no auth."
    And the response should contain the field "text" matching "^<comment> Auth type hawk is not supported, so the snippet has no auth.\n<next>"

    Examples:
      | language        | comment | next            |
      | curl            | #       | curl --location |
      | js_fetch        | //      | const           |
      | go_native       | //      | package main    |
      | python_requests | #       | import requests |
