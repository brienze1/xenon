#language: en
#utf-8

@all @xen_018 @mcp @aliases @alias_environments @mcp_alias_environments
Feature: AI clients read and set an alias's environments, and cannot send it in another environment
  Reading a request through "tree" lists, for each alias, its environments by id and name (absent for
  all environments, as MCP output leaves out every empty list; "Deleted environment" for an id with
  no environment). "tree_edit" with kind
  "alias" takes environment_ids on create and update: omitted keeps the link, [] allows all
  environments. "send" with alias_id is refused with ALS-09403 when the active environment is not
  one of the alias's, with no approval asked and nothing sent. History reads show the environment
  that was active for each send.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the "api" variable is set to the mocked api url
    And the MCP client is named "Claude Code"
    And the "billing" variable is set to "01KD0000000000000000000001"
    And the "login" variable is set to "01KD0000000000000000000002"
    And the "hml" variable is set to "01KD0000000000000000000004"
    And the "prod" variable is set to "01KD0000000000000000000006"
    And the "gone" variable is set to "01KD0000000000000000000007"
    And the "mx" variable is set to "01KD0000000000000000000010"
    And the "v2" variable is set to "01KD0000000000000000000012"
    And the "unknown" variable is set to "01KD0000000000000000000099"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01KD0000000000000000000000
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
    And the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{prod}}
      name: Prod
      color: red
      confirmBeforeSend: true
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
            - {{gone}}
          method: POST
          url: "{{api}}/login"
        - id: {{v2}}
          name: v2
          method: POST
          url: "{{api}}/v2/login"
      """

  @read
  Scenario: Reading a request lists each alias's environments by id and name
    When I call the MCP tool "tree" with
      """
      { "action": "get", "id": "{{login}}" }
      """
    Then the call should succeed
    And the response should contain the field "aliases.0.environments" array length equal to 2
    And the response should contain the field "aliases.0.environments.0.id" equal to "{{hml}}"
    And the response should contain the field "aliases.0.environments.0.name" equal to "HML"
    And the response should contain the field "aliases.0.environments.1.id" equal to "{{gone}}"
    And the response should contain the field "aliases.0.environments.1.name" equal to "Deleted environment"
    And the response field "aliases.1.environments" should not exist

  @create
  Scenario: Creating an alias with environment_ids links it
    When I call the MCP tool "tree_edit" with
      """
      { "action": "create", "kind": "alias", "request_id": "{{login}}", "name": "canary", "environment_ids": ["{{prod}}"] }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "canary"
    And the response should contain the field "environments" array length equal to 1
    And the response should contain the field "environments.0.id" equal to "{{prod}}"
    And the response should contain the field "environments.0.name" equal to "Prod"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.2.environments.0" equal to "{{prod}}"

  @update
  Scenario: Updating an alias sets, keeps and clears its environments
    When I call the MCP tool "tree_edit" with
      """
      { "action": "update", "kind": "alias", "request_id": "{{login}}", "id": "{{v2}}", "environment_ids": ["{{hml}}"] }
      """
    Then the call should succeed
    And the response should contain the field "environments.0.name" equal to "HML"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.environments" array length equal to 1
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.environments.0" equal to "{{hml}}"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.url" equal to "{{api}}/v2/login"
    When I call the MCP tool "tree_edit" with
      """
      { "action": "update", "kind": "alias", "request_id": "{{login}}", "id": "{{v2}}", "url": "{{api}}/v3/login" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.url" equal to "{{api}}/v3/login"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.environments.0" equal to "{{hml}}"
    When I call the MCP tool "tree_edit" with
      """
      { "action": "update", "kind": "alias", "request_id": "{{login}}", "id": "{{v2}}", "environment_ids": [] }
      """
    Then the call should succeed
    And the response field "environments" should not exist
    And the workspace file "acme/collections/billing/login.yaml" should not contain the field "aliases.1.environments"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.url" equal to "{{api}}/v3/login"

  @errors
  Scenario Outline: An environment id that does not exist is refused - <action>
    When I call the MCP tool "tree_edit" with
      """
      <arguments>
      """
    Then the call should fail with code "ALS-10400"
    And the workspace file "acme/collections/billing/login.yaml" should be unchanged

    Examples:
      | action | arguments                                                                                                         |
      | create | { "action": "create", "kind": "alias", "request_id": "{{login}}", "name": "canary", "environment_ids": ["{{unknown}}"] } |
      | update | { "action": "update", "kind": "alias", "request_id": "{{login}}", "id": "{{v2}}", "environment_ids": ["{{unknown}}"] }  |

  @send @errors
  Scenario: Sending a linked alias while another environment is active is refused without an approval prompt
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    When I call the MCP tool "send" with
      """
      { "request_id": "{{login}}", "alias_id": "{{mx}}" }
      """
    Then the call should fail with code "ALS-09403"
    And the "POST" request to "/login" should have been received 0 times
    And the event "ai:approval-requested" should have been emitted 0 times
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}", "source": "ai" }
      """
    Then the response field "" should contain 0 items with "kind" equal to "send"

  @send
  Scenario: Sending a linked alias in its environment works and history reads show the environment
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    And the 1 "POST" request to "/login" returns status 200 with the following response
      """
      {"ok":true}
      """
    When I call the MCP tool "send" with
      """
      { "request_id": "{{login}}", "alias_id": "{{mx}}" }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "200"
    When I call the MCP tool "history" with
      """
      { "action": "list" }
      """
    Then the call should succeed
    And the response field "items" should contain an item with "kind" equal to "send" and "environment" equal to "HML"
    And I save the response field "items.0.id" as "entry"
    When I call the MCP tool "history" with
      """
      { "action": "get", "id": "{{entry}}" }
      """
    Then the call should succeed
    And the response should contain the field "environment.id" equal to "{{hml}}"
    And the response should contain the field "environment.name" equal to "HML"
