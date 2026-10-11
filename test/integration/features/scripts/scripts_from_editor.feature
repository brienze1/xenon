#language: en
#utf-8

@all @xen_004 @scripts @scripts_from_editor
Feature: Scripts edited in the app
  A request opened in the app carries its scripts and its "Inherit parent scripts" toggle, so the
  Scripts panel can show and edit them. A send runs the scripts the request carries, saved or not
  (an unsaved or edited request runs what is in the editor); a request that does not carry scripts
  runs the ones in its saved file. Saving a request that carries scripts writes them to its file.
  Collection and folder settings carry their scripts too (and folders the inherit toggle), and
  saving them writes the scripts.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK9000000000000000000001"
    And the "invoices" variable is set to "01JK9000000000000000000002"
    And the "create" variable is set to "01JK9000000000000000000003"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK9000000000000000000000
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
        preRequest: console.log('collection pre')
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      scripts:
        postResponse: console.log('folder post')
      inheritScripts: true
      """
    And the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: console.log('saved pre')
        disabled:
          - postResponse
        postResponse: console.log('saved post')
      inheritScripts: true
      """
    And the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      {"id":"inv-1"}
      """
    And I call "ConsoleService.Clear"

  @get
  Scenario: An opened request carries its scripts and its inherit toggle
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{create}}" }
      """
    Then the call should succeed
    And the response should contain the field "scripts.pre_request" equal to "console.log('saved pre')"
    And the response should contain the field "scripts.post_response" equal to "console.log('saved post')"
    And the response should contain the field "scripts.disabled" with the items "postResponse"
    And the response should contain the field "inherit_scripts" equal to "true"

  @send_edited
  Scenario: A send runs the scripts the request carries, not the saved ones
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices", "scripts": { "pre_request": "console.log('edited pre')", "post_response": "console.log('edited post')", "disabled": [] }, "inherit_scripts": true } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "" array length equal to 4
    And the response should contain the field "0.summary" equal to "collection pre"
    And the response should contain the field "1.summary" equal to "edited pre"
    And the response should contain the field "2.summary" equal to "folder post"
    And the response should contain the field "3.summary" equal to "edited post"
    And the workspace file "acme/collections/billing/invoices/create.yaml" should contain the field "scripts.preRequest" equal to "console.log('saved pre')"

  @send_inherit_off
  Scenario: Turning inherit off in the editor skips the parent scripts for that send
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices", "scripts": { "pre_request": "console.log('own pre')", "post_response": "", "disabled": [] }, "inherit_scripts": false } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.summary" equal to "own pre"

  @send_disabled
  Scenario: A phase disabled in the editor does not run
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices", "scripts": { "pre_request": "console.log('own pre')", "post_response": "console.log('own post')", "disabled": ["preRequest"] } } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "" array length equal to 3
    And the response should contain the field "0.summary" equal to "collection pre"
    And the response should contain the field "2.summary" equal to "own post"

  @send_unsaved
  Scenario: An unsaved request runs the scripts in its editor
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-2", "request": { "name": "Untitled", "method": "GET", "url": "{{api}}/invoices", "scripts": { "pre_request": "pm.request.headers.add({ key: 'X-Draft', value: 'yes' })", "post_response": "pm.test('draft ok', () => pm.response.to.have.status(200))", "disabled": [] } } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/invoices" should have the header "X-Draft" equal to "yes"
    And the response should contain the field "tests.0.name" equal to "draft ok"
    And the response should contain the field "tests.0.status" equal to "passed"

  @send_without_scripts
  Scenario: A request that does not carry scripts runs its saved scripts
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "" array length equal to 3
    And the response should contain the field "1.summary" equal to "saved pre"

  @save
  Scenario: Saving a request writes the scripts and the inherit toggle it carries
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{create}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "RequestService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "base_revision": "{{revision}}",
        "force": false,
        "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices", "scripts": { "pre_request": "pm.variables.set('a', 1)", "post_response": "pm.test('ok', () => {})", "disabled": [] }, "inherit_scripts": false }
      }
      """
    Then the call should succeed
    And the response should contain the field "scripts.pre_request" equal to "pm.variables.set('a', 1)"
    And the response should contain the field "inherit_scripts" equal to "false"
    And the workspace file "acme/collections/billing/invoices/create.yaml" should contain the field "scripts.preRequest" equal to "pm.variables.set('a', 1)"
    And the workspace file "acme/collections/billing/invoices/create.yaml" should contain the field "scripts.postResponse" equal to "pm.test('ok', () => {})"
    And the workspace file "acme/collections/billing/invoices/create.yaml" should not contain the field "scripts.disabled"
    And the workspace file "acme/collections/billing/invoices/create.yaml" should contain the field "inheritScripts" equal to "false"

  @save_without_scripts
  Scenario: Saving a request that does not carry scripts keeps the saved ones
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{create}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "RequestService.Save" with
      """
      { "local_id": "{{Acme.local_id}}", "base_revision": "{{revision}}", "force": false, "request": { "id": "{{create}}", "name": "Create", "method": "POST", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/invoices/create.yaml" should contain the field "scripts.preRequest" equal to "console.log('saved pre')"
    And the workspace file "acme/collections/billing/invoices/create.yaml" should contain the field "method" equal to "POST"

  @item_settings
  Scenario: Folder settings carry and save scripts and the inherit toggle
    When I call "TreeService.GetSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{invoices}}" }
      """
    Then the call should succeed
    And the response should contain the field "scripts.post_response" equal to "console.log('folder post')"
    And the response should contain the field "scripts.pre_request" equal to ""
    And the response should contain the field "inherit_scripts" equal to "true"
    When I call "TreeService.SaveSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{invoices}}", "description": "", "auth": { "type": "inherit" }, "headers": [], "scripts": { "pre_request": "console.log('folder pre')", "post_response": "", "disabled": [] }, "inherit_scripts": false }
      """
    Then the call should succeed
    And the response should contain the field "scripts.pre_request" equal to "console.log('folder pre')"
    And the response should contain the field "inherit_scripts" equal to "false"
    And the workspace file "acme/collections/billing/invoices/folder.yaml" should contain the field "scripts.preRequest" equal to "console.log('folder pre')"
    And the workspace file "acme/collections/billing/invoices/folder.yaml" should contain the field "inheritScripts" equal to "false"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    And I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "0.summary" equal to "folder pre"
    And the response should contain the field "1.summary" equal to "saved pre"

  @collection_settings
  Scenario: Collection settings carry scripts and have no inherit toggle
    When I call "TreeService.SaveSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{billing}}", "description": "", "auth": { "type": "none" }, "headers": [], "scripts": { "pre_request": "", "post_response": "console.log('collection post')", "disabled": [] } }
      """
    Then the call should succeed
    And the response should contain the field "scripts.post_response" equal to "console.log('collection post')"
    And the response field "inherit_scripts" should not exist
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "scripts.postResponse" equal to "console.log('collection post')"
    And the workspace file "acme/collections/billing/collection.yaml" should not contain the field "inheritScripts"

  @settings_without_scripts
  Scenario: Saving settings that do not carry scripts keeps the saved ones
    When I call "TreeService.SaveSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{invoices}}", "description": "Invoice calls", "auth": { "type": "inherit" }, "headers": [] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/invoices/folder.yaml" should contain the field "scripts.postResponse" equal to "console.log('folder post')"
    And the workspace file "acme/collections/billing/invoices/folder.yaml" should contain the field "description" equal to "Invoice calls"
