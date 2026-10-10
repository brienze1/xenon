#language: en
#utf-8

@all @xen_004 @scripts @execution_order
Feature: Scripts run around every send in Postman's order
  A send runs the pre-request scripts of the request's collection, its folders (outer to inner)
  and the request itself, then sends, then runs the post-response scripts in the same order.
  Every script runs in its own scope: a top-level declaration in one script is not visible in
  another, and values pass only through pm.variables (the local scope of this one send).
  Scripts imported as disabled are not run. Script console output is recorded in the session
  console, grouped by the send's id.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK0000000000000000000001"
    And the "invoices" variable is set to "01JK0000000000000000000002"
    And the "drafts" variable is set to "01JK0000000000000000000003"
    And the "create" variable is set to "01JK0000000000000000000004"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK0000000000000000000000
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
      scripts:
        preRequest: console.log("collection pre")
        postResponse: console.log("collection post")
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      scripts:
        preRequest: console.log("invoices pre")
        postResponse: console.log("invoices post")
      inheritScripts: true
      """
    And the workspace file "acme/collections/billing/invoices/drafts/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{drafts}}
      name: Drafts
      scripts:
        preRequest: console.log("drafts pre")
        postResponse: console.log("drafts post")
      inheritScripts: true
      """
    And the workspace file "acme/collections/billing/invoices/drafts/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: console.log("request pre")
        postResponse: console.log("request post")
      inheritScripts: true
      """
    And the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      {"id":"inv-1"}
      """
    And I call "ConsoleService.Clear"

  @order
  Scenario: Pre-request scripts run outer to inner, then the request is sent, then post-response scripts run outer to inner
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the response should contain the field "status_code" equal to "200"
    And the response should contain the field "script_errors" array length equal to 0
    And the response should contain the field "send_id" matching "^[0-9A-Z]{26}$"
    And the "GET" request to "/invoices" should have been received 1 times
    When I save the response field "send_id" as "send"
    And I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 8
    And the response should contain the field "0.summary" equal to "collection pre"
    And the response should contain the field "1.summary" equal to "invoices pre"
    And the response should contain the field "2.summary" equal to "drafts pre"
    And the response should contain the field "3.summary" equal to "request pre"
    And the response should contain the field "4.summary" equal to "collection post"
    And the response should contain the field "5.summary" equal to "invoices post"
    And the response should contain the field "6.summary" equal to "drafts post"
    And the response should contain the field "7.summary" equal to "request post"
    And the response should contain the field "0.kind" equal to "log"
    And the response should contain the field "0.level" equal to "log"
    And the response should contain the field "0.send_id" equal to "{{send}}"
    And the response should contain the field "7.send_id" equal to "{{send}}"
    And the response should contain the field "0.script.owner_kind" equal to "collection"
    And the response should contain the field "0.script.owner_name" equal to "Billing"
    And the response should contain the field "0.script.phase" equal to "preRequest"
    And the response should contain the field "2.script.owner_kind" equal to "folder"
    And the response should contain the field "2.script.owner_name" equal to "Drafts"
    And the response should contain the field "7.script.owner_kind" equal to "request"
    And the response should contain the field "7.script.owner_name" equal to "Create"
    And the response should contain the field "7.script.phase" equal to "postResponse"

  @levels
  Scenario: Every console level and argument list is recorded in order
    Given the workspace file "acme/collections/billing/invoices/drafts/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: |
          console.info("info", 1, true)
          console.warn("warn")
          console.error("error", { "a": 1 })
          console.debug("debug", [1, 2])
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 10
    And the response should contain the field "3.level" equal to "info"
    And the response should contain the field "3.summary" equal to "info 1 true"
    And the response should contain the field "3.arguments" with the items "info, 1, true"
    And the response should contain the field "4.level" equal to "warn"
    And the response should contain the field "5.level" equal to "error"
    And the response should contain the field "5.summary" equal to
      """
      error {"a":1}
      """
    And the response should contain the field "6.level" equal to "debug"
    And the response should contain the field "6.summary" equal to "debug [1,2]"

  @scope
  Scenario: Each script has its own scope and values pass through pm.variables for this send only
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      scripts:
        preRequest: |
          const secretLocal = "collection only"
          pm.variables.set("shared", "from collection")
          pm.variables.set("count", 1)
      """
    And the workspace file "acme/collections/billing/invoices/drafts/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: |
          console.log(typeof secretLocal)
          console.log(pm.variables.get("shared"))
          console.log(typeof pm.variables.get("count"), pm.variables.has("count"), pm.variables.has("nope"))
        postResponse: console.log("post sees", pm.variables.get("shared"))
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 8
    And the response should contain the field "2.summary" equal to "undefined"
    And the response should contain the field "3.summary" equal to "from collection"
    And the response should contain the field "4.summary" equal to "number true false"
    And the response should contain the field "7.summary" equal to "post sees from collection"
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      """
    And I call "ConsoleService.Clear"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 8
    And the response should contain the field "3.summary" equal to "undefined"
    And the response should contain the field "4.summary" equal to "undefined false false"
    And the response should contain the field "7.summary" equal to "post sees undefined"

  @disabled
  Scenario: Disabled scripts are not run
    Given the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      scripts:
        preRequest: console.log("invoices pre")
        postResponse: console.log("invoices post")
        disabled:
          - preRequest
      inheritScripts: true
      """
    And the workspace file "acme/collections/billing/invoices/drafts/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{drafts}}
      name: Drafts
      scripts:
        preRequest: console.log("drafts pre")
        postResponse: console.log("drafts post")
        disabled:
          - preRequest
          - postResponse
      inheritScripts: true
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 5
    And the response should contain the field "0.summary" equal to "collection pre"
    And the response should contain the field "1.summary" equal to "request pre"
    And the response should contain the field "2.summary" equal to "collection post"
    And the response should contain the field "3.summary" equal to "invoices post"
    And the response should contain the field "4.summary" equal to "request post"

  @unsaved
  Scenario: An unsaved request has no scripts and is sent as before
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the response should contain the field "tests" array length equal to 0
    And the response should contain the field "script_errors" array length equal to 0
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 0

  @limit
  Scenario: The console list returns at most the requested number of entries, oldest first
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 3 }
      """
    Then the response should contain the field "" array length equal to 3
    And the response should contain the field "2.summary" equal to "drafts pre"
    When I save the response field "2.id" as "third"
    And I call "ConsoleService.List" with
      """
      { "after_id": "{{third}}", "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 5
    And the response should contain the field "0.summary" equal to "request pre"
