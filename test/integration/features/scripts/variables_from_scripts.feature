#language: en
#utf-8

@all @xen_004 @scripts @variables_from_scripts
Feature: Scripts read and write every variable scope with Postman's methods
  pm.environment (the active environment), pm.collectionVariables (the request's collection),
  pm.globals (the workspace globals) and pm.variables (the send's local scope) offer Postman's
  has, get, set, unset, clear, replaceIn and toObject. Writes follow XEN-002: they become current
  values kept on this machine (variables.json), never written to the workspace files; a secret
  row keeps its value in the vault; a name the scope does not have yet becomes a new row with an
  empty initial value and the written value as its current value. unset and clear drop the
  current values (the rows stay in the files). Values that are not strings stay as they are for
  the rest of the send and are stored as strings (objects JSON-encoded). With no active
  environment pm.environment writes do nothing and log a console warning. Every send whose
  scripts change variables announces "variables:changed".

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK5000000000000000000001"
    And the "create" variable is set to "01JK5000000000000000000004"
    And the "dev" variable is set to "01JK5000000000000000000009"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK5000000000000000000000
      name: Acme
      order:
        - billing
      variables:
        - key: user
          value: luis
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
        - key: page_size
          value: "50"
          enabled: true
          description: ""
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: base_url
          value: "{{api}}"
          enabled: true
          description: ""
        - key: tenant
          value: acme
          enabled: true
          description: ""
        - key: token
          value: ""
          enabled: true
          description: ""
        - key: api_key
          secret: true
          enabled: true
          description: ""
      """
    And the 1 "POST" request to "/token" returns status 200 with the following response
      """
      {"access_token":"tok-1"}
      """
    And the 1 "GET" request to "/me" returns status 200 with the following response
      """
      {"id":"me"}
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    And I call "ConsoleService.Clear"

  @environment
  Scenario: A post-response script stores a token as the environment's current value and the next send uses it
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{base_url}}/token"
      scripts:
        postResponse: pm.environment.set("token", pm.response.json().access_token)
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{base_url}}/token" } }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.token" equal to "tok-1"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.2.value" equal to ""
    And no file under the workspace folder should contain "tok-1"
    And the event "variables:changed" should have been emitted 2 times
    And the 2 "variables:changed" event should contain the field "scopes.0" equal to "env:{{dev}}"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{base_url}}/me", "headers": [ { "key": "Authorization", "value": "Bearer {{token}}", "enabled": true } ] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/me" should have the header "Authorization" equal to "Bearer tok-1"

  @methods
  Scenario: has, get, set, replaceIn and toObject behave as in Postman on every scope
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/me"
      scripts:
        preRequest: |
          console.log(pm.environment.name, pm.environment.has("tenant"), pm.environment.has("user"), pm.environment.get("tenant"))
          console.log(pm.collectionVariables.get("page_size"), pm.globals.get("user"), pm.variables.get("user"), pm.variables.get("missing"))
          pm.environment.set("tenant", "globex")
          pm.collectionVariables.set("page_size", "10")
          pm.globals.set("user", "ada")
          pm.variables.set("local_only", "l")
          console.log(pm.variables.replaceIn("{{tenant}}/{{page_size}}/{{user}}/{{local_only}}/{{unknown}}"))
          console.log(JSON.stringify(pm.collectionVariables.toObject()), JSON.stringify(pm.globals.toObject()))
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/me" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "0.summary" equal to "dev true false acme"
    And the response should contain the field "1.summary" equal to "50 luis luis undefined"
    And the response should contain the field "2.summary" equal to
      """
      globex/10/ada/l/{{unknown}}
      """
    And the response should contain the field "3.summary" equal to
      """
      {"page_size":"10"} {"user":"ada"}
      """
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.tenant" equal to "globex"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.collection:{{billing}}.page_size" equal to "10"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.globals.user" equal to "ada"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "local_only"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.0.value" equal to "luis"
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "variables.0.value" equal to "50"

  @new_row
  Scenario: Setting a name the scope does not have adds a row with an empty initial value
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/me"
      scripts:
        postResponse: |
          pm.environment.set("session", "s" + "-1")
          pm.collectionVariables.set("cursor", "c-1")
          pm.globals.set("last_run", "r-1")
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/me" } }
      """
    Then the call should succeed
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables" array length equal to 5
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.4.key" equal to "session"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.4.value" equal to ""
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "variables.1.key" equal to "cursor"
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "variables.1.value" equal to ""
    And the workspace file "acme/workspace.yaml" should contain the field "variables.1.key" equal to "last_run"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.1.value" equal to ""
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.session" equal to "s-1"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.collection:{{billing}}.cursor" equal to "c-1"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.globals.last_run" equal to "r-1"
    And no file under the workspace folder should contain "s-1"

  @secret
  Scenario: Writing a secret row keeps the value in the vault, never in a file
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/me"
      scripts:
        postResponse: |
          pm.environment.set("api_key", ["script", "secret", "1"].join("-"))
          console.log(pm.environment.get("api_key").length)
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/me" } }
      """
    Then the call should succeed
    And no file under the workspace folder should contain "script-secret-1"
    And no file under the app data folder should contain "script-secret-1"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "api_key" }
      """
    Then the call should succeed
    And the response should contain the field "value" equal to "script-secret-1"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "0.summary" equal to "15"

  @non_string
  Scenario: Values that are not strings keep their type during the send and are stored as strings
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      variables:
        - key: page_size
          value: "50"
          enabled: true
          description: ""
      scripts:
        preRequest: |
          pm.environment.set("count", 42)
          pm.globals.set("flags", { "beta": true, "ids": [1, 2] })
          pm.collectionVariables.set("enabled", false)
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/me?count={{count}}"
      scripts:
        preRequest: |
          console.log(typeof pm.environment.get("count"), pm.globals.get("flags").ids.length, pm.collectionVariables.get("enabled") === false)
      """
    And the 1 "GET" request to "/me" returns status 200 with the following response
      """
      {"id":"me"}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/me?count={{count}}" } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/me" should have the raw query "count=42"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.count" equal to "42"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.globals.flags" equal to
      """
      {"beta":true,"ids":[1,2]}
      """
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.collection:{{billing}}.enabled" equal to "false"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "0.summary" equal to "number 2 true"

  @unset @clear
  Scenario: unset and clear drop the current values for the rest of the send and on this machine
    Given I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "tenant", "value": "globex" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "user", "value": "ada" }
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/me"
      scripts:
        preRequest: |
          console.log(pm.environment.get("tenant"), pm.globals.get("user"))
          pm.environment.unset("tenant")
          pm.globals.clear()
          console.log(pm.environment.has("tenant"), pm.environment.get("tenant"), JSON.stringify(pm.globals.toObject()))
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/me" } }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "globex"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "ada"
    And the workspace file "acme/environments/dev.yaml" should contain the field "variables.1.key" equal to "tenant"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.0.key" equal to "user"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "0.summary" equal to "globex ada"
    And the response should contain the field "1.summary" equal to "false undefined {}"

  @local
  Scenario: pm.variables lives for one send only
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{base_url}}/me"
      scripts:
        preRequest: |
          console.log(pm.variables.get("run"))
          pm.variables.set("run", "first")
        postResponse: console.log(pm.variables.get("run"))
      """
    And the 2 "GET" request to "/me" returns status 200 with the following response
      """
      {"id":"me"}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/me" } }
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{base_url}}/me" } }
      """
    And I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "0.summary" equal to "undefined"
    And the response should contain the field "1.summary" equal to "first"
    And the response should contain the field "2.summary" equal to "undefined"
    And the response should contain the field "3.summary" equal to "first"
    And the event "variables:changed" should have been emitted 1 times

  @no_environment
  Scenario: With no active environment pm.environment writes do nothing and log a warning
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "" }
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/me"
      scripts:
        preRequest: |
          pm.environment.set("token", "t" + "-9")
          console.log(pm.environment.name, pm.environment.get("token"), JSON.stringify(pm.environment.toObject()))
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/me" } }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "t-9"
    And no file under the workspace folder should contain "t-9"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "0.level" equal to "warn"
    And the response should contain the field "0.summary" equal to
      """
      pm.environment.set("token") was ignored: no environment is active
      """
    And the response should contain the field "1.summary" equal to "undefined undefined {}"
