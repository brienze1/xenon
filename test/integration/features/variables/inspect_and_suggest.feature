#language: en
#utf-8

@all @xen_002 @variables @inspect_and_suggest
Feature: Inspecting, suggesting and adding variables from a request
  VariableService.Inspect walks the same fields the send-time resolver walks (URL, params, path
  variables, headers, the active auth type, the active body mode, plus the auth and headers the
  request inherits from its collection and folders) and reports every "{{name}}" with its field,
  its start and end (JavaScript string offsets, so the editor can decorate them), its status and
  the scope it comes from. Nothing is generated: dynamic variables are only marked "dynamic", and
  secret values are always masked. Inspect also lists each variable once, with the fields that use
  it, following nested references. VariableService.Suggest offers names for "{{" autocomplete
  (every scope plus Postman's dynamic variables, at most 50), and VariableService.AddVariable
  creates a variable in a scope in one step — as a secret too, for "Move to secret variable".

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the "billing" variable is set to "01JI0000000000000000000001"
    And the "invoices" variable is set to "01JI0000000000000000000002"
    And the "list" variable is set to "01JI0000000000000000000003"
    And the "dev" variable is set to "01JI0000000000000000000010"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JI0000000000000000000000
      name: Acme
      order:
        - billing
      variables:
        - key: user
          value: luis
          enabled: true
          description: "who is calling"
        - key: base_url
          value: "http://globals.acme.test"
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      auth:
        type: bearer
        bearer:
          token: "{{token}}"
      headers:
        - key: X-Tenant
          value: "{{tenant}}"
          enabled: true
      variables:
        - key: tenant
          value: acme-co
          enabled: true
          description: ""
        - key: token
          value: collection-token
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      headers:
        - key: X-Page
          value: "{{page_size}}"
          enabled: true
      """
    And the workspace file "acme/collections/billing/invoices/list.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{list}}
      name: List invoices
      method: GET
      url: "{{base_url}}/invoices"
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
          value: "http://dev.acme.test"
          enabled: true
          description: "the dev gateway"
        - key: api_key
          secret: true
          enabled: true
          description: ""
        - key: auth_header
          value: "Key {{api_key}}"
          enabled: true
          description: ""
        - key: loop_a
          value: "{{loop_b}}"
          enabled: true
          description: ""
        - key: loop_b
          value: "{{loop_a}}"
          enabled: true
          description: ""
        - key: bare
          value: "{{nobody}}"
          enabled: true
          description: ""
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """

  @inspect @url
  Scenario: References in the URL carry their field, offsets, status and scope
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "{{base_url}}/users/{{user_id}}" } }
      """
    Then the call should succeed
    And the response should contain the field "references" array length equal to 2
    And the response should contain the field "references.0.field" equal to "url"
    And the response should contain the field "references.0.name" equal to "base_url"
    And the response should contain the field "references.0.start" equal to "0"
    And the response should contain the field "references.0.end" equal to "12"
    And the response should contain the field "references.0.status" equal to "resolved"
    And the response should contain the field "references.0.scope" equal to "environment"
    And the response should contain the field "references.0.display_value" equal to "http://dev.acme.test"
    And the response should contain the field "references.0.secret" equal to "false"
    And the response should contain the field "references.1.name" equal to "user_id"
    And the response should contain the field "references.1.start" equal to "19"
    And the response should contain the field "references.1.end" equal to "30"
    And the response should contain the field "references.1.status" equal to "unresolved"
    And the response should contain the field "references.1.scope" equal to "none"
    And the response should contain the field "references.1.display_value" equal to ""

  @inspect @url @offsets
  Scenario: Offsets count JavaScript string units, not bytes
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "https://é.test/😀/{{user}}" } }
      """
    Then the response should contain the field "references.0.name" equal to "user"
    And the response should contain the field "references.0.start" equal to "18"
    And the response should contain the field "references.0.end" equal to "26"
    And the response should contain the field "references.0.scope" equal to "globals"

  @inspect @tables @body
  Scenario: Key/value tables, auth and body are inspected field by field
    When I call "VariableService.Inspect" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "request": {
          "method": "POST",
          "url": "/x",
          "params": [ { "key": "page", "value": "{{page_size}}", "enabled": true } ],
          "path_variables": [ { "key": "id", "value": "{{user}}" } ],
          "headers": [ { "key": "{{tenant}}-Key", "value": "v", "enabled": false } ],
          "auth": { "type": "basic", "basic": { "username": "{{user}}", "password": "p" }, "bearer": { "token": "{{ignored_bearer}}" } },
          "body": {
            "mode": "formdata",
            "raw": { "language": "json", "content": "{{ignored_raw}}" },
            "formdata": [ { "key": "who", "value": "{{user}}", "type": "text", "enabled": true } ]
          }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "references" array length equal to 5
    And the response should contain the field "references.0.field" equal to "params.0.value"
    And the response should contain the field "references.0.status" equal to "unresolved"
    And the response should contain the field "references.1.field" equal to "path_variables.0.value"
    And the response should contain the field "references.2.field" equal to "headers.0.key"
    And the response should contain the field "references.2.start" equal to "0"
    And the response should contain the field "references.2.end" equal to "10"
    And the response should contain the field "references.2.status" equal to "unresolved"
    And the response should contain the field "references.3.field" equal to "auth.basic.username"
    And the response should contain the field "references.3.display_value" equal to "luis"
    And the response should contain the field "references.4.field" equal to "body.formdata.0.value"
    And the response should not contain "ignored_bearer"
    And the response should not contain "ignored_raw"

  @inspect @body
  Scenario: Raw and GraphQL bodies are inspected in their active mode
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "POST", "url": "/x", "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"who\": \"{{user}}\"}" } } } }
      """
    Then the response should contain the field "references.0.field" equal to "body.raw"
    And the response should contain the field "references.0.start" equal to "9"
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "POST", "url": "/x", "body": { "mode": "graphql", "graphql": { "query": "{ me(id: \"{{user}}\") }", "variables": "{\"t\": \"{{tenant}}\"}" } } } }
      """
    Then the response should contain the field "references.0.field" equal to "body.graphql.query"
    And the response should contain the field "references.1.field" equal to "body.graphql.variables"

  @inspect @inherited
  Scenario: Inherited auth and headers are reported with the ancestor they come from
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "id": "{{list}}", "method": "GET", "url": "{{base_url}}/invoices", "auth": { "type": "inherit" } } }
      """
    Then the call should succeed
    And the response should contain the field "references" array length equal to 4
    And the response should contain the field "references.1.field" equal to "inherited:{{billing}}.auth.bearer.token"
    And the response should contain the field "references.1.name" equal to "token"
    And the response should contain the field "references.1.scope" equal to "collection"
    And the response should contain the field "references.1.display_value" equal to "collection-token"
    And the response should contain the field "references.2.field" equal to "inherited:{{billing}}.headers.0.value"
    And the response should contain the field "references.2.name" equal to "tenant"
    And the response should contain the field "references.3.field" equal to "inherited:{{invoices}}.headers.0.value"
    And the response should contain the field "references.3.name" equal to "page_size"
    And the response should contain the field "references.3.status" equal to "unresolved"
    And the response should contain the field "variables" array length equal to 4
    And the response field "variables" should contain an item with "name" equal to "token" and "scope" equal to "collection"
    And the response field "variables" should contain an item with "name" equal to "page_size" and "status" equal to "unresolved"

  @inspect @inherited
  Scenario: An inherited header the request overrides is not reported
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "id": "{{list}}", "method": "GET", "url": "/x", "auth": { "type": "none" }, "headers": [ { "key": "x-tenant", "value": "own", "enabled": true } ] } }
      """
    Then the response should contain the field "references" array length equal to 1
    And the response should contain the field "references.0.name" equal to "page_size"

  @inspect @secrets
  Scenario: Secret values are masked and missing secrets reported
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "/x", "headers": [ { "key": "X-Key", "value": "{{api_key}}", "enabled": true } ] } }
      """
    Then the response should contain the field "references.0.status" equal to "missing_secret"
    And the response should contain the field "references.0.secret" equal to "true"
    And the response should contain the field "references.0.scope" equal to "environment"
    When I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key", "value": "inspect-s3cr3t" }
      """
    And I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "/x", "headers": [ { "key": "X-Key", "value": "{{api_key}}", "enabled": true }, { "key": "Authorization", "value": "{{auth_header}}", "enabled": true } ] } }
      """
    Then the call should succeed
    And the response should contain the field "references.0.status" equal to "resolved"
    And the response should contain the field "references.0.display_value" equal to "•••••"
    And the response should contain the field "references.1.name" equal to "auth_header"
    And the response should contain the field "references.1.display_value" equal to "Key •••••"
    And the response should not contain "inspect-s3cr3t"

  @inspect @nested
  Scenario: Nested names are listed once with every field that uses them, cycles are unresolved
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "{{bare}}/{{loop_a}}", "headers": [ { "key": "Authorization", "value": "{{auth_header}}", "enabled": true }, { "key": "X-Key", "value": "{{api_key}}", "enabled": true } ] } }
      """
    Then the response should contain the field "references.0.name" equal to "bare"
    And the response should contain the field "references.0.status" equal to "resolved"
    And the response should contain the field "references.0.display_value" equal to "{{nobody}}"
    And the response should contain the field "references.1.name" equal to "loop_a"
    And the response should contain the field "references.1.status" equal to "unresolved"
    And the response field "variables" should contain an item with "name" equal to "nobody" and "status" equal to "unresolved"
    And the response field "variables" should contain an item with "name" equal to "api_key" and "status" equal to "missing_secret"
    And the response field "variables" should contain 1 items with "name" equal to "api_key"
    And the response field "variables" should contain an item with "name" equal to "auth_header" and "scope" equal to "environment"

  @inspect @used_in
  Scenario: The variables list carries every field a name is used in
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "{{user}}/{{user}}", "params": [ { "key": "u", "value": "{{user}}", "enabled": true } ] } }
      """
    Then the response should contain the field "references" array length equal to 3
    And the response should contain the field "variables" array length equal to 1
    And the response should contain the field "variables.0.name" equal to "user"
    And the response should contain the field "variables.0.display_value" equal to "luis"
    And the response should contain the field "variables.0.used_in" with the items "url, params.0.value"

  @inspect @dynamic
  Scenario: Dynamic variables are marked dynamic without generating a value
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "/{{$guid}}/{{$randomEmail}}/{{$notARealOne}}" } }
      """
    Then the response should contain the field "references.0.status" equal to "dynamic"
    And the response should contain the field "references.0.scope" equal to "dynamic"
    And the response should contain the field "references.0.display_value" equal to ""
    And the response should contain the field "references.1.status" equal to "dynamic"
    And the response should contain the field "references.2.status" equal to "unresolved"

  @inspect @dynamic
  Scenario: A user variable named like a dynamic one wins
    Given I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "$guid", "value": "fixed-guid", "secret": false }
      """
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "/{{$guid}}" } }
      """
    Then the response should contain the field "references.0.status" equal to "resolved"
    And the response should contain the field "references.0.scope" equal to "globals"
    And the response should contain the field "references.0.display_value" equal to "fixed-guid"

  @inspect @no_environment
  Scenario: With no active environment the collection and globals apply
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "" }
      """
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "id": "{{list}}", "method": "GET", "url": "{{base_url}}/{{tenant}}", "auth": { "type": "none" } } }
      """
    Then the response should contain the field "references.0.scope" equal to "globals"
    And the response should contain the field "references.0.display_value" equal to "http://globals.acme.test"
    And the response should contain the field "references.1.scope" equal to "collection"

  @inspect @scopes
  Scenario: Inspect names the variable tables the request can use
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "id": "{{list}}", "method": "GET", "url": "/x" } }
      """
    Then the response should contain the field "scopes" array length equal to 3
    And the response should contain the field "scopes.0.scope" equal to "environment"
    And the response should contain the field "scopes.0.ref" equal to "env:{{dev}}"
    And the response should contain the field "scopes.1.scope" equal to "collection"
    And the response should contain the field "scopes.1.ref" equal to "collection:{{billing}}"
    And the response should contain the field "scopes.2.scope" equal to "globals"
    And the response should contain the field "scopes.2.ref" equal to "globals"
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "" }
      """
    And I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "/x" } }
      """
    Then the response should contain the field "scopes" array length equal to 1
    And the response should contain the field "scopes.0.ref" equal to "globals"

  @inspect @validation
  Scenario: Inspect needs a workspace
    When I call "VariableService.Inspect" with
      """
      { "local_id": "", "request": { "method": "GET", "url": "/x" } }
      """
    Then the call should fail with code "VAR-15400"

  @suggest
  Scenario: Suggestions match a prefix across scopes, case-insensitively
    When I call "VariableService.Suggest" with
      """
      { "local_id": "{{Acme.local_id}}", "prefix": "BA" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 2
    And the response should contain the field "0.name" equal to "bare"
    And the response should contain the field "1.name" equal to "base_url"
    And the response should contain the field "1.scope" equal to "environment"
    And the response should contain the field "1.description" equal to "the dev gateway"
    And the response should contain the field "1.secret" equal to "false"
    And the response should contain the field "1.dynamic" equal to "false"

  @suggest
  Scenario: Collection variables are suggested for the request's collection only
    When I call "VariableService.Suggest" with
      """
      { "local_id": "{{Acme.local_id}}", "prefix": "t" }
      """
    Then the response should contain the field "" array length equal to 0
    When I call "VariableService.Suggest" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{list}}", "prefix": "t" }
      """
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "0.name" equal to "tenant"
    And the response should contain the field "0.scope" equal to "collection"
    And the response should contain the field "1.name" equal to "token"

  @suggest @secrets
  Scenario: Secret variables are suggested without values
    When I call "VariableService.Suggest" with
      """
      { "local_id": "{{Acme.local_id}}", "prefix": "api" }
      """
    Then the response should contain the field "0.name" equal to "api_key"
    And the response should contain the field "0.secret" equal to "true"

  @suggest @dynamic
  Scenario: Dynamic variables are suggested after user variables, at most 50 in total
    When I call "VariableService.Suggest" with
      """
      { "local_id": "{{Acme.local_id}}", "prefix": "$" }
      """
    Then the response should contain the field "" array length equal to 50
    And the response should contain the field "0.name" equal to "$guid"
    And the response should contain the field "0.scope" equal to "dynamic"
    And the response should contain the field "0.dynamic" equal to "true"
    When I call "VariableService.Suggest" with
      """
      { "local_id": "{{Acme.local_id}}", "prefix": "$randomEm" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.name" equal to "$randomEmail"

  @suggest @dynamic
  Scenario: A user variable shadows the dynamic variable of the same name
    Given I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "$guid", "value": "fixed-guid", "secret": false }
      """
    When I call "VariableService.Suggest" with
      """
      { "local_id": "{{Acme.local_id}}", "prefix": "$gu" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.scope" equal to "globals"
    And the response should contain the field "0.dynamic" equal to "false"

  @suggest @precedence
  Scenario: A name defined in several scopes is suggested once, from the strongest scope
    When I call "VariableService.Suggest" with
      """
      { "local_id": "{{Acme.local_id}}", "prefix": "base" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.scope" equal to "environment"

  @suggest @validation
  Scenario: Suggest needs a workspace
    When I call "VariableService.Suggest" with
      """
      { "local_id": "", "prefix": "a" }
      """
    Then the call should fail with code "VAR-16400"

  @add
  Scenario Outline: Add to <target> creates the variable there and it resolves at once
    When I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "<scope>", "key": "user_id", "value": "42", "secret": false }
      """
    Then the call should succeed
    And the response should contain the field "key" equal to "user_id"
    And the response should contain the field "initial_value" equal to "42"
    And the workspace file "<file>" should contain the field "<field>" equal to "user_id"
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "id": "{{list}}", "method": "GET", "url": "/users/{{user_id}}", "auth": { "type": "none" } } }
      """
    Then the response should contain the field "references.0.status" equal to "resolved"
    And the response should contain the field "references.0.scope" equal to "<target>"

    Examples:
      | target      | scope                 | file                                     | field           |
      | environment | env:{{dev}}           | acme/environments/dev.yaml               | variables.6.key |
      | collection  | collection:{{billing}} | acme/collections/billing/collection.yaml | variables.2.key |
      | globals     | globals               | acme/workspace.yaml                      | variables.2.key |

  @add @secrets
  Scenario: Move to secret variable stores the selected value in the vault only
    When I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "pasted_token", "value": "pasted-s3cr3t-token", "secret": true }
      """
    Then the call should succeed
    And the response should contain the field "secret" equal to "true"
    And the response should contain the field "secret_state" equal to "set"
    And the response should not contain "pasted-s3cr3t-token"
    And the workspace file "acme/environments/dev.yaml" should not contain the field "variables.6.value"
    And no file under the workspace folder should contain "pasted-s3cr3t-token"
    And no file under the app data folder should contain "pasted-s3cr3t-token"
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "GET", "url": "/x", "headers": [ { "key": "Authorization", "value": "Bearer {{pasted_token}}", "enabled": true } ] } }
      """
    Then the response should contain the field "references.0.status" equal to "resolved"
    And the response should contain the field "references.0.display_value" equal to "•••••"
    And the response should not contain "pasted-s3cr3t-token"
