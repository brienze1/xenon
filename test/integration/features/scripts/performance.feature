#language: en
#utf-8

@all @xen_004 @scripts @performance
Feature: Script performance
  Scripts must not slow sends down: a send with four typical scripts (signing with CryptoJS,
  storing a value, a test and a log, one per level) takes under 100 ms on top of the network once
  the libraries are loaded (SC-002). A script stuck in an endless loop is stopped within its time
  limit plus half a second, and the app keeps answering other calls meanwhile (SC-003).

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JKA000000000000000000001"
    And the "invoices" variable is set to "01JKA000000000000000000002"
    And the "create" variable is set to "01JKA000000000000000000003"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JKA000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      {"id":"inv-1","total":20}
      """
    And the 2 "GET" request to "/invoices" returns status 200 with the following response
      """
      {"id":"inv-1","total":20}
      """

  @typical_scripts
  Scenario: Four typical scripts add less than 100 ms to a send
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      scripts:
        preRequest: |
          const timestamp = Date.now().toString()
          const signature = CryptoJS.HmacSHA256("GET" + timestamp, "key").toString(CryptoJS.enc.Base64)
          pm.request.headers.upsert({ key: "X-Signature", value: signature })
          pm.request.headers.upsert({ key: "X-Timestamp", value: timestamp })
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      scripts:
        preRequest: pm.variables.set("tenant", "acme")
        postResponse: pm.collectionVariables.set("last_invoice", pm.response.json().id)
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
        postResponse: |
          pm.test("status is 200", () => pm.response.to.have.status(200))
          pm.test("has a total", () => pm.expect(pm.response.json().total).to.be.above(0))
          console.log("invoice", pm.response.json().id)
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "test_summary.passed" equal to "2"
    And the response should contain the field "script_errors" array length equal to 0
    And the 2 "GET" request to "/invoices" should have the header "X-Timestamp" matching "^[0-9]+$"
    And the last call should have taken less than 100 ms

  @endless_loop
  Scenario: An endless loop is stopped within its limit plus half a second while the app keeps answering
    Given I call "SettingsService.Save" with
      """
      { "script_timeout_ms": 1000 }
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      """
    And the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: while (true) {}
      """
    When I start calling "RequestService.Send" in the background with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    And I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the last call should have taken less than 500 ms
    And the background call should finish within 1500 ms
    And the response should contain the field "outcome" equal to "stopped_by_script_error"
    And the response should contain the field "script_errors.0.kind" equal to "script_timeout"
    And the "GET" request to "/invoices" should have been received 0 times
