#language: en
#utf-8

@all @xen_018 @aliases @alias_environments @alias_environment_storage
Feature: An alias can be limited to a set of environments, stored with the alias in its request's file
  Each alias has an Environments setting: All environments (the default, written as no key at all)
  or the ids of one or more of the workspace's environments, written as "environments:" under that
  alias. The link is set when creating the alias or with AliasService.SetEnvironments, a write of
  its own that never touches the alias's content. Every id must be an environment of the workspace;
  duplicates are stored once. Ids are stable, so renaming an environment keeps the link, and deleting
  an environment never rewrites request files: the id stays and is reported as it is.

  Background:
    Given the workspace "Acme" exists
    And the "billing" variable is set to "01KA0000000000000000000001"
    And the "login" variable is set to "01KA0000000000000000000002"
    And the "orders" variable is set to "01KA0000000000000000000003"
    And the "hml" variable is set to "01KA0000000000000000000004"
    And the "dev" variable is set to "01KA0000000000000000000005"
    And the "prod" variable is set to "01KA0000000000000000000006"
    And the "mx" variable is set to "01KA0000000000000000000010"
    And the "smoke" variable is set to "01KA0000000000000000000011"
    And the "unknown" variable is set to "01KA0000000000000000000099"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01KA0000000000000000000000
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
        - orders.yaml
      """
    And the workspace file "acme/collections/billing/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: https://api.acme.test/v1/login
      aliases:
        - id: {{mx}}
          name: mx-test-user
          environments:
            - {{hml}}
          method: POST
          url: https://api.acme.test/v2/login
        - id: {{smoke}}
          name: prod-smoke-user
          method: POST
          url: https://api.acme.test/v3/login
      """
    And the workspace file "acme/collections/billing/orders.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{orders}}
      name: Orders
      description: ""
      method: GET
      url: https://api.acme.test/v1/orders
      params: []
      pathVariables: []
      headers: []
      autoHeaders:
        disabled: []
      auth:
        type: inherit
      body:
        mode: none
      settings:
        timeoutMs: 0
        followRedirects: true
        verifySsl: true
        encodeUrl: true
        disableCookieJar: false
      scripts:
        preRequest: ""
        postResponse: ""
      inheritScripts: true
      aliases:
        - id: {{smoke}}
          name: canary
          method: GET
          url: https://api.acme.test/v2/orders
          params: []
          pathVariables: []
          headers: []
          autoHeaders:
            disabled: []
          auth:
            type: inherit
          body:
            mode: none
          settings:
            timeoutMs: 0
            followRedirects: true
            verifySsl: true
            encodeUrl: true
            disableCookieJar: false
          scripts:
            preRequest: ""
            postResponse: ""
          inheritScripts: true
      postman: {}
      """

  @get
  Scenario: Opening a request returns each alias's environment ids, empty for All environments
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    Then the call should succeed
    And the response should contain the field "aliases.0.environment_ids" array length equal to 1
    And the response should contain the field "aliases.0.environment_ids.0" equal to "{{hml}}"
    And the response should contain the field "aliases.1.environment_ids" array length equal to 0

  @set
  Scenario: Setting an alias's environments writes them under that alias only and keeps its content
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.SetEnvironments" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "alias_id": "{{smoke}}", "environment_ids": ["{{prod}}", "{{dev}}"], "base_revision": "{{revision}}" }
      """
    Then the call should succeed
    And the response should contain the field "revision" not equal to "{{revision}}"
    And the response should contain the field "aliases.1.environment_ids" array length equal to 2
    And the response should contain the field "aliases.1.environment_ids.0" equal to "{{prod}}"
    And the response should contain the field "aliases.1.environment_ids.1" equal to "{{dev}}"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.environments" array length equal to 2
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.environments.0" equal to "{{prod}}"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.environments.1" equal to "{{dev}}"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.url" equal to "https://api.acme.test/v3/login"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.name" equal to "prod-smoke-user"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.environments.0" equal to "{{hml}}"
    And the workspace file "acme/collections/billing/login.yaml" should not contain the field "environments"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "url" equal to "https://api.acme.test/v1/login"

  @set
  Scenario: Repeated and blank ids are stored once, in the order first given
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.SetEnvironments" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "alias_id": "{{smoke}}", "environment_ids": ["{{dev}}", " ", "{{hml}}", "{{dev}}"], "base_revision": "{{revision}}" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.environments" array length equal to 2
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.environments.0" equal to "{{dev}}"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.1.environments.1" equal to "{{hml}}"

  @set @clear
  Scenario: Setting All environments removes the key from the alias
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.SetEnvironments" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "alias_id": "{{mx}}", "environment_ids": [], "base_revision": "{{revision}}" }
      """
    Then the call should succeed
    And the response should contain the field "aliases.0.environment_ids" array length equal to 0
    And the workspace file "acme/collections/billing/login.yaml" should not contain the field "aliases.0.environments"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.name" equal to "mx-test-user"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.url" equal to "https://api.acme.test/v2/login"

  @set @byte_identical
  Scenario: Linking an alias and setting it back to All environments leaves the file exactly as it was
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{orders}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.SetEnvironments" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "alias_id": "{{smoke}}", "environment_ids": ["{{hml}}"], "base_revision": "{{revision}}" }
      """
    Then the call should succeed
    And I save the response field "revision" as "revision"
    And the workspace file "acme/collections/billing/orders.yaml" should contain the field "aliases.0.environments.0" equal to "{{hml}}"
    When I call "AliasService.SetEnvironments" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{orders}}", "alias_id": "{{smoke}}", "environment_ids": [], "base_revision": "{{revision}}" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/orders.yaml" should equal
      """
      xenon: 1
      kind: request
      id: {{orders}}
      name: Orders
      description: ""
      method: GET
      url: https://api.acme.test/v1/orders
      params: []
      pathVariables: []
      headers: []
      autoHeaders:
        disabled: []
      auth:
        type: inherit
      body:
        mode: none
      settings:
        timeoutMs: 0
        followRedirects: true
        verifySsl: true
        encodeUrl: true
        disableCookieJar: false
      scripts:
        preRequest: ""
        postResponse: ""
      inheritScripts: true
      aliases:
        - id: {{smoke}}
          name: canary
          method: GET
          url: https://api.acme.test/v2/orders
          params: []
          pathVariables: []
          headers: []
          autoHeaders:
            disabled: []
          auth:
            type: inherit
          body:
            mode: none
          settings:
            timeoutMs: 0
            followRedirects: true
            verifySsl: true
            encodeUrl: true
            disableCookieJar: false
          scripts:
            preRequest: ""
            postResponse: ""
          inheritScripts: true
      postman: {}
      """

  @save
  Scenario: Saving an alias's content keeps its environments
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "base_revision": "{{revision}}",
        "alias": { "id": "{{mx}}", "name": "mx-test-user", "environment_ids": [], "request": { "method": "POST", "url": "https://api.acme.test/v9/login" } }
      }
      """
    Then the call should succeed
    And the response should contain the field "aliases.0.environment_ids.0" equal to "{{hml}}"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.url" equal to "https://api.acme.test/v9/login"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.environments" array length equal to 1
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.environments.0" equal to "{{hml}}"

  @create
  Scenario: Creating an alias with environments links it from the start
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "base_revision": "{{revision}}", "name": "canary", "environment_ids": ["{{hml}}"] }
      """
    Then the call should succeed
    And the response should contain the field "aliases" array length equal to 3
    And the response should contain the field "aliases.2.name" equal to "canary"
    And the response should contain the field "aliases.2.environment_ids" array length equal to 1
    And the response should contain the field "aliases.2.environment_ids.0" equal to "{{hml}}"
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.2.environments.0" equal to "{{hml}}"

  @create
  Scenario Outline: Creating an alias without environments makes it available in all environments - <case>
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "base_revision": "{{revision}}", "name": "canary"<extra> }
      """
    Then the call should succeed
    And the response should contain the field "aliases.2.environment_ids" array length equal to 0
    And the workspace file "acme/collections/billing/login.yaml" should not contain the field "aliases.2.environments"

    Examples:
      | case        | extra                    |
      | not given   |                          |
      | empty list  | , "environment_ids": []  |

  @errors
  Scenario Outline: An id that is not an environment of the workspace is refused and nothing is written - <call>
    Given the workspace file "acme/environments/broken.yaml" contains
      """
      xenon: 1
      kind: environment
      id: [not, a, valid, id
      """
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.<call>" with
      """
      <payload>
      """
    Then the call should fail with code "ALS-10400"
    And the workspace file "acme/collections/billing/login.yaml" should be unchanged

    Examples:
      | call            | payload                                                                                                                                                              |
      | SetEnvironments | { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "alias_id": "{{smoke}}", "environment_ids": ["{{hml}}", "{{unknown}}"], "base_revision": "{{revision}}" } |
      | SetEnvironments | { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "alias_id": "{{smoke}}", "environment_ids": ["error:broken.yaml"], "base_revision": "{{revision}}" }            |
      | Create          | { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "base_revision": "{{revision}}", "name": "canary", "environment_ids": ["{{unknown}}"] }                    |

  @errors
  Scenario: Setting the environments of an unknown alias fails and writes nothing
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.SetEnvironments" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "alias_id": "{{unknown}}", "environment_ids": ["{{hml}}"], "base_revision": "{{revision}}" }
      """
    Then the call should fail with code "ALS-01404"
    And the workspace file "acme/collections/billing/login.yaml" should be unchanged

  @errors @conflict
  Scenario: Setting environments on a file that changed on disk is refused
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    Given the workspace file "acme/collections/billing/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: https://api.acme.test/v1/login
      aliases:
        - id: {{mx}}
          name: mx-renamed-by-git
          method: POST
          url: https://api.acme.test/v2/login
      """
    When I call "AliasService.SetEnvironments" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "alias_id": "{{mx}}", "environment_ids": ["{{prod}}"], "base_revision": "{{revision}}" }
      """
    Then the call should fail with code "REQ-02409"
    And the workspace file "acme/collections/billing/login.yaml" should be unchanged

  @validation @errors
  Scenario Outline: SetEnvironments validates its payload - <case>
    When I call "AliasService.SetEnvironments" with
      """
      <payload>
      """
    Then the call should fail with code "ALS-11400"

    Examples:
      | case         | payload                                                                                                  |
      | no workspace | { "local_id": "", "request_id": "{{login}}", "alias_id": "{{mx}}", "environment_ids": [] }               |
      | no request   | { "local_id": "{{Acme.local_id}}", "request_id": "", "alias_id": "{{mx}}", "environment_ids": [] }       |
      | no alias id  | { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "alias_id": " ", "environment_ids": [] }   |

  @environment_lifecycle @byte_identical
  Scenario: Deleting a linked environment leaves the request file untouched and the id still linked
    When I call "EnvironmentService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/login.yaml" should be unchanged
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    Then the response should contain the field "aliases.0.environment_ids" array length equal to 1
    And the response should contain the field "aliases.0.environment_ids.0" equal to "{{hml}}"

  @environment_lifecycle
  Scenario: A deleted environment's id is dropped by saving a new choice without it
    Given I call "EnvironmentService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.SetEnvironments" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "alias_id": "{{mx}}", "environment_ids": ["{{dev}}"], "base_revision": "{{revision}}" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.environments" array length equal to 1
    And the workspace file "acme/collections/billing/login.yaml" should contain the field "aliases.0.environments.0" equal to "{{dev}}"

  @environment_lifecycle @byte_identical
  Scenario: Renaming and recolouring a linked environment keeps the link and leaves the request file untouched
    When I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}", "name": "Staging", "description": "", "color": "yellow", "confirm_before_send": false }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/login.yaml" should be unchanged
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    Then the response should contain the field "aliases.0.environment_ids.0" equal to "{{hml}}"

  @environment_lifecycle
  Scenario: Duplicating a linked environment links nothing to the copy
    When I call "EnvironmentService.Duplicate" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{hml}}" }
      """
    Then the call should succeed
    And the response should contain the field "id" not equal to "{{hml}}"
    And the workspace file "acme/collections/billing/login.yaml" should be unchanged
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    Then the response should contain the field "aliases.0.environment_ids" array length equal to 1
    And the response should contain the field "aliases.0.environment_ids.0" equal to "{{hml}}"

  @duplicate
  Scenario: Duplicating a request keeps its aliases' environments
    When I call "TreeService.Duplicate" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/login-copy.yaml" should contain the field "aliases.0.environments" array length equal to 1
    And the workspace file "acme/collections/billing/login-copy.yaml" should contain the field "aliases.0.environments.0" equal to "{{hml}}"
    And the workspace file "acme/collections/billing/login-copy.yaml" should not contain the field "aliases.1.environments"
    And I save the response field "id" as "copy"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{copy}}" }
      """
    Then the response should contain the field "aliases.0.environment_ids.0" equal to "{{hml}}"
    And the response should contain the field "aliases.1.environment_ids" array length equal to 0

  @save_as
  Scenario: Save As from a linked alias's content creates a request with no aliases and no environments
    When I call "RequestService.SaveAs" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "parent_id": "{{billing}}",
        "name": "Login MX",
        "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "https://api.acme.test/v2/login" }
      }
      """
    Then the call should succeed
    And the response should contain the field "aliases" array length equal to 0
    And the workspace file "acme/collections/billing/login-mx.yaml" should not contain the field "aliases"
    And the workspace file "acme/collections/billing/login-mx.yaml" should not contain the field "environments"
