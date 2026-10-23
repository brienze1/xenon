#language: en
#utf-8

@all @xen_018 @aliases @alias_environments @alias_environment_send
Feature: An alias limited to environments can only be sent while one of them is active
  An alias linked to environments is available only while one of its environments is active; with
  no active environment (or a remembered one that was deleted) it is unavailable. Go refuses to send
  an unavailable alias with ALS-09403, naming the alias's environments, before any script runs, any
  confirmation is asked or anything is sent, and writes no history entry. Aliases set to All
  environments and the Default are never refused.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01KB0000000000000000000001"
    And the "login" variable is set to "01KB0000000000000000000002"
    And the "hml" variable is set to "01KB0000000000000000000004"
    And the "dev" variable is set to "01KB0000000000000000000005"
    And the "prod" variable is set to "01KB0000000000000000000006"
    And the "mx" variable is set to "01KB0000000000000000000010"
    And the "qa" variable is set to "01KB0000000000000000000011"
    And the "v2" variable is set to "01KB0000000000000000000012"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01KB0000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/environments/hml.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{hml}}
      name: HML
      color: green
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: DEV
      color: cyan
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{prod}}
      name: Prod
      color: red
      confirmBeforeSend: false
      variables: []
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      order:
        - login.yaml
      """
    And the workspace file "acme/collections/billing/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: "{{api}}/login"
      aliases:
        - id: {{mx}}
          name: mx-test-user
          environments:
            - {{hml}}
          method: POST
          url: "{{api}}/login"
          scripts:
            preRequest: console.log("mx pre-request ran")
            postResponse: ""
          inheritScripts: true
        - id: {{qa}}
          name: qa-smoke
          environments:
            - {{hml}}
            - {{dev}}
          method: POST
          url: "{{api}}/login"
        - id: {{v2}}
          name: v2
          method: POST
          url: "{{api}}/login"
      """

  @available
  Scenario: A linked alias is sent while its environment is active
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    And the 1 "POST" request to "/login" returns status 200 with the following response
      """
      {"ok":true}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{mx}}", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "200"
    And the "POST" request to "/login" should have been received 1 times

  @unavailable @errors
  Scenario Outline: A linked alias is refused and nothing happens when its environment is not active - <case>
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "<active>" }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{mx}}", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should fail with code "ALS-09403"
    And the error should contain the field "error.description" containing "mx-test-user"
    And the error should contain the field "error.description" containing "HML"
    And the "POST" request to "/login" should have been received 0 times
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should not contain "mx pre-request ran"

    Examples:
      | case                  | active    |
      | another environment   | {{prod}}  |
      | no active environment |           |

  @unavailable @errors
  Scenario: A remembered active environment that was deleted counts as no environment
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    And the workspace path "acme/environments/prod.yaml" is removed
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{mx}}", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should fail with code "ALS-09403"
    And the "POST" request to "/login" should have been received 0 times

  @multiple
  Scenario: An alias linked to several environments is sent in any of them and refused elsewhere
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    And the 1 "POST" request to "/login" returns status 200 with the following response
      """
      {"ok":true}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{qa}}", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{qa}}", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should fail with code "ALS-09403"
    And the error should contain the field "error.description" containing "HML"
    And the error should contain the field "error.description" containing "DEV"
    And the "POST" request to "/login" should have been received 1 times

  @scope_guard
  Scenario Outline: An All environments alias and the Default are sent whatever is active - <version> with <case>
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "<active>" }
      """
    And the 1 "POST" request to "/login" returns status 200 with the following response
      """
      {"ok":true}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "<alias>", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    And the "POST" request to "/login" should have been received 1 times

    Examples:
      | version | case                  | active   | alias   |
      | v2      | Prod                  | {{prod}} | {{v2}}  |
      | v2      | no active environment |          | {{v2}}  |
      | Default | Prod                  | {{prod}} |         |
      | Default | no active environment |          |         |

  @confirm_before_send
  Scenario: With a "Confirm before send" environment active, the alias is refused before any confirmation is asked
    Given the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{prod}}
      name: Prod
      color: red
      confirmBeforeSend: true
      variables: []
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{mx}}", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should fail with code "ALS-09403"
    And the "POST" request to "/login" should have been received 0 times

  @environment_lifecycle
  Scenario: An alias whose only environment was deleted is refused in every environment
    Given I call "EnvironmentService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{mx}}", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should fail with code "ALS-09403"
    And the error should contain the field "error.description" containing "Deleted environment"
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{mx}}", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should fail with code "ALS-09403"
    And the "POST" request to "/login" should have been received 0 times

  @environment_lifecycle
  Scenario: Renaming the linked environment keeps the alias sendable while it is active
    Given I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}", "name": "Staging", "description": "", "color": "green", "confirm_before_send": false }
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    And the 1 "POST" request to "/login" returns status 200 with the following response
      """
      {"ok":true}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{mx}}", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    And the "POST" request to "/login" should have been received 1 times
