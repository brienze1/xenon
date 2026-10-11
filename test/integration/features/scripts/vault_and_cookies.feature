#language: en
#utf-8

@all @xen_004 @scripts @vault_and_cookies
Feature: pm.vault uses the workspace secrets and pm.cookies the response cookies
  Postman's vault maps to the workspace-level secrets (the secret rows of the globals):
  pm.vault.get, set and unset return promises and read and write them in the XEN-002 vault
  (set on a new name adds a secret row to the globals), and {{vault:name}} resolves to the same
  value anywhere a variable can be used. pm.cookies lists the cookies the response set (has, get,
  toObject and the rest of Postman's list API); pm.cookies.jar() is the workspace cookie jar of
  stage 5 (features/cookies/script_jar_allowlist.feature).

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JL1000000000000000000001"
    And the "create" variable is set to "01JL1000000000000000000004"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JL1000000000000000000000
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
      """
    And I call "VariableService.SaveTable" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "scope": "globals",
        "rows": [
          { "key": "api_key", "secret": true, "secret_value": "vault-key-1", "enabled": true },
          { "key": "old", "secret": true, "secret_value": "vault-old-1", "enabled": true }
        ]
      }
      """
    And the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      {"id":"inv-1"}
      """
    And the requests to "/invoices" respond with the header "Set-Cookie" equal to "session=s-1; Path=/; HttpOnly"
    And the requests to "/invoices" respond with the header "Set-Cookie" equal to "theme=dark"
    And I call "ConsoleService.Clear"

  @vault
  Scenario: pm.vault reads and writes the workspace secrets and {{vault:name}} resolves them
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      headers:
        - key: X-Api-Key
          value: "{{vault:api_key}}"
          enabled: true
          description: ""
      scripts:
        preRequest: |
          const key = await pm.vault.get("api_key")
          await pm.vault.set("signing_key", ["vault", "new", "1"].join("-"))
          await pm.vault.unset("old")
          console.log(key === "vault-key-1", (await pm.vault.get("signing_key")).length === 11, await pm.vault.get("old"), await pm.vault.get("missing"))
          pm.request.headers.upsert({ key: "X-Signing", value: "{{vault:signing_key}}" })
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices", "headers": [{ "key": "X-Api-Key", "value": "{{vault:api_key}}", "enabled": true }] } }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    And the 1 "GET" request to "/invoices" should have the header "X-Api-Key" equal to "vault-key-1"
    And the 1 "GET" request to "/invoices" should have the header "X-Signing" equal to "vault-new-1"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.2.key" equal to "signing_key"
    And the workspace file "acme/workspace.yaml" should contain the field "variables.2.secret" equal to "true"
    And no file under the workspace folder should contain "vault-new-1"
    And no file under the app data folder should contain "vault-new-1"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "signing_key" }
      """
    Then the response should contain the field "value" equal to "vault-new-1"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "old" }
      """
    Then the call should fail with code "VLT-03404"
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "0.summary" equal to "true true null null"

  @cookies
  Scenario: pm.cookies lists the response cookies
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        postResponse: |
          console.log(pm.cookies.count(), pm.cookies.has("session"), pm.cookies.get("session"), pm.cookies.get("nope"), JSON.stringify(pm.cookies.toObject()))
          console.log(pm.cookies.one("session").httpOnly, pm.cookies.one("session").path, pm.response.cookies.count())
          console.log(typeof pm.cookies.jar().get, typeof pm.cookies.jar().clear)
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "0.summary" equal to
      """
      2 true s-1 undefined {"session":"s-1","theme":"dark"}
      """
    And the response should contain the field "1.summary" equal to "true / 2"
    And the response should contain the field "2.summary" equal to "function function"
