#language: en
#utf-8

@all @xen_002 @variables @resolution_precedence
Feature: Variables resolve with Postman's scope precedence when sending
  Every "{{name}}" in a request is resolved by Go at send time, strongest scope first: the active
  environment, then the variables of the request's collection, then the workspace globals. A
  disabled row is ignored as if it were absent, so the next scope is used. A row's current value
  (local to this machine) overrides its initial value (the file). Names nobody defines are sent
  literally, as Postman does.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JV0000000000000000000001"
    And the "invoices" variable is set to "01JV0000000000000000000002"
    And the "list" variable is set to "01JV0000000000000000000003"
    And the "dev" variable is set to "01JV0000000000000000000010"
    And the "hml" variable is set to "01JV0000000000000000000011"
    And the "staging" variable is set to "01JV0000000000000000000012"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JV0000000000000000000000
      name: Acme
      order:
        - billing
      variables:
        - key: tenant
          value: from-globals
          enabled: true
          description: ""
        - key: region
          value: eu-from-globals
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      variables:
        - key: tenant
          value: from-collection
          enabled: true
          description: ""
        - key: page_size
          value: "50"
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      variables:
        - key: tenant
          value: from-folder
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/invoices/list.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{list}}
      name: List invoices
      method: GET
      url: "{{base_url}}/invoices"
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      color: none
      confirmBeforeSend: false
      variables:
        - key: base_url
          value: "{{api}}/dev"
          enabled: true
          description: ""
        - key: tenant
          value: from-dev
          enabled: true
          description: ""
      """
    And the workspace file "acme/environments/hml.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{hml}}
      name: hml
      color: yellow
      confirmBeforeSend: false
      variables:
        - key: base_url
          value: "{{api}}/hml"
          enabled: true
          description: ""
      """
    And the workspace file "acme/environments/staging.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{staging}}
      name: staging
      variables:
        - key: base_url
          value: "{{api}}/staging"
          enabled: true
          description: ""
        - key: tenant
          value: from-staging
          enabled: false
          description: "disabled, so the collection value is used"
      """

  @switch
  Scenario: The same request goes to the target of the active environment
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    Then the call should succeed
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{list}}", "method": "GET", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "200"
    And the "GET" request to "/dev/invoices" should have been received 1 times
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{list}}", "method": "GET", "url": "{{base_url}}/invoices" } }
      """
    Then the call should succeed
    And the "GET" request to "/hml/invoices" should have been received 1 times
    And the "GET" request to "/dev/invoices" should have been received 1 times

  @precedence
  Scenario Outline: The strongest scope that defines a name wins - <case>
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "<environment>" }
      """
    And I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "id": "{{list}}",
          "method": "GET",
          "url": "{{api}}/precedence",
          "headers": [
            { "key": "X-Tenant", "value": "{{tenant}}" },
            { "key": "X-Page-Size", "value": "{{page_size}}" },
            { "key": "X-Region", "value": "{{region}}" }
          ]
        }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/precedence" should have the header "X-Tenant" equal to "<tenant>"
    And the 1 "GET" request to "/precedence" should have the header "X-Page-Size" equal to "50"
    And the 1 "GET" request to "/precedence" should have the header "X-Region" equal to "eu-from-globals"

    Examples:
      | case                                          | environment | tenant          |
      | environment over collection and globals       | {{dev}}     | from-dev        |
      | collection when the environment lacks it      | {{hml}}     | from-collection |
      | a disabled environment row falls through      | {{staging}} | from-collection |
      | no environment uses the collection            |             | from-collection |

  @globals
  Scenario: A request outside any collection only sees the environment and the globals
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    And I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "unsaved",
        "request": {
          "method": "GET",
          "url": "{{base_url}}/unsaved",
          "headers": [ { "key": "X-Tenant", "value": "{{tenant}}" }, { "key": "X-Page-Size", "value": "{{page_size}}" } ]
        }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/hml/unsaved" should have the header "X-Tenant" equal to "from-globals"
    And the 1 "GET" request to "/hml/unsaved" should have the header "X-Page-Size" equal to "{{page_size}}"

  @globals
  Scenario: A disabled global is ignored and the name is sent literally
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JV0000000000000000000000
      name: Acme
      order:
        - billing
      variables:
        - key: tenant
          value: from-globals
          enabled: false
          description: ""
        - key: region
          value: enabled-global
          enabled: true
          description: ""
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "unsaved", "request": { "method": "GET", "url": "{{api}}/disabled", "headers": [ { "key": "X-Tenant", "value": "{{tenant}}" }, { "key": "X-Region", "value": "{{region}}" } ] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/disabled" should have the header "X-Tenant" equal to "{{tenant}}"
    And the 1 "GET" request to "/disabled" should have the header "X-Region" equal to "enabled-global"

  @no_environment @unresolved
  Scenario: With no environment an undefined name is sent literally and does not block the send
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "id": "{{list}}",
          "method": "GET",
          "url": "{{api}}/literal/{{user_id}}?base={{base_url}}",
          "headers": [ { "key": "X-Base", "value": "{{base_url}}" }, { "key": "X-Page-Size", "value": "{{page_size}}" } ]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "200"
    And the "GET" request to "/literal/{{user_id}}" should have been received 1 times
    And the 1 "GET" request to "/literal/{{user_id}}" should have the header "X-Base" equal to "{{base_url}}"
    And the 1 "GET" request to "/literal/{{user_id}}" should have the header "X-Page-Size" equal to "50"
    And the 1 "GET" request to "/literal/{{user_id}}" should have the query "base" equal to "{{base_url}}"

  @folders
  Scenario: Folders have no variables - a folder.yaml "variables" key is ignored
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{list}}", "method": "GET", "url": "{{api}}/folder", "headers": [ { "key": "X-Tenant", "value": "{{tenant}}" } ] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/folder" should have the header "X-Tenant" equal to "from-collection"

  @current_values
  Scenario: A current value kept on this machine overrides the initial value in the file
    Given the app data file "workspaces/{{Acme.local_id}}/variables.json" contains
      """
      {
        "version": 1,
        "activeEnvironmentId": "{{dev}}",
        "current": {
          "env:{{dev}}": { "tenant": "from-dev-current" },
          "collection:{{billing}}": { "page_size": "10" },
          "globals": { "region": "us-from-current" }
        }
      }
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "id": "{{list}}",
          "method": "GET",
          "url": "{{base_url}}/current",
          "headers": [
            { "key": "X-Tenant", "value": "{{tenant}}" },
            { "key": "X-Page-Size", "value": "{{page_size}}" },
            { "key": "X-Region", "value": "{{region}}" }
          ]
        }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/dev/current" should have the header "X-Tenant" equal to "from-dev-current"
    And the 1 "GET" request to "/dev/current" should have the header "X-Page-Size" equal to "10"
    And the 1 "GET" request to "/dev/current" should have the header "X-Region" equal to "us-from-current"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.value" equal to "from-dev"

  @current_values
  Scenario: A current value never revives a disabled row
    Given the app data file "workspaces/{{Acme.local_id}}/variables.json" contains
      """
      { "version": 1, "activeEnvironmentId": "{{staging}}", "current": { "env:{{staging}}": { "tenant": "staging-current" } } }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{list}}", "method": "GET", "url": "{{base_url}}/disabled-current", "headers": [ { "key": "X-Tenant", "value": "{{tenant}}" } ] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/staging/disabled-current" should have the header "X-Tenant" equal to "from-collection"

  @active_file_missing
  Scenario: An active environment whose file is gone is treated as no environment
    Given the app data file "workspaces/{{Acme.local_id}}/variables.json" contains
      """
      { "version": 1, "activeEnvironmentId": "01JV9999999999999999999999", "current": {} }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{list}}", "method": "GET", "url": "{{api}}/gone", "headers": [ { "key": "X-Tenant", "value": "{{tenant}}" } ] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/gone" should have the header "X-Tenant" equal to "from-collection"
