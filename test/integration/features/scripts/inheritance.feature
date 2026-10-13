#language: en
#utf-8

@all @xen_004 @scripts @inheritance
Feature: "Inherit parent scripts" cuts the script chain
  Folders and requests have "Inherit parent scripts" (inheritScripts, on by default). When it is
  off on a folder, the scripts of everything above that folder are skipped for requests inside
  it, in both phases; the folder's own scripts and the ones below it still run. When it is off
  on a request, only the request's own scripts run. A collection has no parent, so the toggle
  does not apply to it.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK1000000000000000000001"
    And the "invoices" variable is set to "01JK1000000000000000000002"
    And the "drafts" variable is set to "01JK1000000000000000000003"
    And the "create" variable is set to "01JK1000000000000000000004"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK1000000000000000000000
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
    And the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      {"id":"inv-1"}
      """
    And I call "ConsoleService.Clear"

  @folder_off
  Scenario: Inherit off on the nested folder skips the collection and the outer folder in both phases
    Given the workspace file "acme/collections/billing/invoices/drafts/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{drafts}}
      name: Drafts
      scripts:
        preRequest: console.log("drafts pre")
        postResponse: console.log("drafts post")
      inheritScripts: false
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
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "" array length equal to 4
    And the response should contain the field "0.summary" equal to "drafts pre"
    And the response should contain the field "1.summary" equal to "request pre"
    And the response should contain the field "2.summary" equal to "drafts post"
    And the response should contain the field "3.summary" equal to "request post"

  @request_off
  Scenario: Inherit off on the request runs only the request's own scripts
    Given the workspace file "acme/collections/billing/invoices/drafts/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{drafts}}
      name: Drafts
      scripts:
        preRequest: console.log("drafts pre")
        postResponse: console.log("drafts post")
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
      inheritScripts: false
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "0.summary" equal to "request pre"
    And the response should contain the field "1.summary" equal to "request post"

  @outer_folder_off
  Scenario: Inherit off on the outer folder keeps that folder and everything below it
    Given the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      scripts:
        preRequest: console.log("invoices pre")
        postResponse: console.log("invoices post")
      inheritScripts: false
      """
    And the workspace file "acme/collections/billing/invoices/drafts/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{drafts}}
      name: Drafts
      scripts:
        preRequest: console.log("drafts pre")
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
        postResponse: console.log("request post")
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "" array length equal to 4
    And the response should contain the field "0.summary" equal to "invoices pre"
    And the response should contain the field "1.summary" equal to "drafts pre"
    And the response should contain the field "2.summary" equal to "invoices post"
    And the response should contain the field "3.summary" equal to "request post"

  @default_on
  Scenario: A request file without the toggle inherits every parent script
    Given the workspace file "acme/collections/billing/invoices/drafts/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      """
    And the workspace file "acme/collections/billing/invoices/drafts/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{drafts}}
      name: Drafts
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "" array length equal to 4
    And the response should contain the field "0.summary" equal to "collection pre"
    And the response should contain the field "1.summary" equal to "invoices pre"
    And the response should contain the field "2.summary" equal to "collection post"
    And the response should contain the field "3.summary" equal to "invoices post"
