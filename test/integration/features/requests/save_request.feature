#language: en
#utf-8

@all @xen_001 @requests @save_request
Feature: Open and save requests
  A saved request opens with a revision (the hash of its file). Saving writes the file only when
  it has not changed on disk since it was opened, unless the user chose to keep their version.
  Save as creates a copy with a new id. Body files inside the workspace are stored relative to it.

  Background:
    Given the workspace "Acme" exists
    And the "billing" variable is set to "01JB0000000000000000000001"
    And the "ping" variable is set to "01JB0000000000000000000004"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JB0000000000000000000000
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
      order:
        - ping.yaml
      """
    And the workspace file "acme/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{ping}}
      name: Ping
      description: Health check
      method: GET
      url: https://api.example.test/ping?verbose=1
      params:
        - key: verbose
          value: "1"
          enabled: true
          description: ""
      headers:
        - key: Accept
          value: application/json
          enabled: false
          description: off for now
      """

  @get
  Scenario: Opening a saved request returns every field and its revision
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{ping}}" }
      """
    Then the call should succeed
    And the response should contain the field "id" equal to "{{ping}}"
    And the response should contain the field "name" equal to "Ping"
    And the response should contain the field "description" equal to "Health check"
    And the response should contain the field "method" equal to "GET"
    And the response should contain the field "url" equal to "https://api.example.test/ping?verbose=1"
    And the response should contain the field "params.0.key" equal to "verbose"
    And the response should contain the field "headers.0.enabled" equal to "false"
    And the response should contain the field "headers.0.description" equal to "off for now"
    And the response should contain the field "auth.type" equal to "inherit"
    And the response should contain the field "settings.follow_redirects" equal to "true"
    And the response should contain the field "revision" matching "^[0-9a-f]{64}$"

  @get
  Scenario: Opening an unknown request fails
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JA2Z6A1B2C3D4E5F6G7H8J9K" }
      """
    Then the call should fail with code "REQ-01404"

  @save
  Scenario: Saving writes the file and returns the new revision
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{ping}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "RequestService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "base_revision": "{{revision}}",
        "force": false,
        "request": { "id": "{{ping}}", "name": "Ping", "description": "Health check", "method": "HEAD", "url": "https://api.example.test/health" }
      }
      """
    Then the call should succeed
    And the response should contain the field "method" equal to "HEAD"
    And the response should contain the field "revision" not equal to "{{revision}}"
    And I save the response field "revision" as "saved"
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "method" equal to "HEAD"
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "url" equal to "https://api.example.test/health"
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "id" equal to "{{ping}}"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{ping}}" }
      """
    Then the response should contain the field "url" equal to "https://api.example.test/health"
    And the response should contain the field "revision" equal to "{{saved}}"

  @conflict
  Scenario: Saving over a file that changed on disk is refused
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{ping}}" }
      """
    And I save the response field "revision" as "revision"
    Given the workspace file "acme/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{ping}}
      name: Ping
      method: GET
      url: https://api.example.test/changed-by-git
      """
    When I call "RequestService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "base_revision": "{{revision}}",
        "request": { "id": "{{ping}}", "name": "Ping", "method": "GET", "url": "https://api.example.test/mine" }
      }
      """
    Then the call should fail with code "REQ-02409"
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "url" equal to "https://api.example.test/changed-by-git"

  @conflict
  Scenario: "Keep mine" saves over the changed file
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{ping}}" }
      """
    And I save the response field "revision" as "revision"
    Given the workspace file "acme/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{ping}}
      name: Ping
      method: GET
      url: https://api.example.test/changed-by-git
      """
    When I call "RequestService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "base_revision": "{{revision}}",
        "force": true,
        "request": { "id": "{{ping}}", "name": "Ping", "method": "GET", "url": "https://api.example.test/mine" }
      }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "url" equal to "https://api.example.test/mine"

  @conflict
  Scenario: Saving a request whose file was deleted fails
    Given the workspace path "acme/collections/billing/ping.yaml" is removed
    When I call "RequestService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "base_revision": "x",
        "request": { "id": "{{ping}}", "name": "Ping", "method": "GET", "url": "https://api.example.test/mine" }
      }
      """
    Then the call should fail with code "REQ-01404"

  @save_as
  Scenario: Save as creates a copy with a new id in the chosen folder
    When I call "RequestService.SaveAs" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "parent_id": "{{billing}}",
        "name": "Ping v2",
        "request": { "id": "{{ping}}", "name": "Ping", "method": "GET", "url": "https://api.example.test/v2/ping" }
      }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "Ping v2"
    And the response should contain the field "id" not equal to "{{ping}}"
    And the response should contain the field "url" equal to "https://api.example.test/v2/ping"
    And the response should contain the field "revision" matching "^[0-9a-f]{64}$"
    And I save the response field "id" as "copy"
    And the workspace file "acme/collections/billing/ping-v2.yaml" should contain the field "id" equal to "{{copy}}"
    And the workspace file "acme/collections/billing/ping-v2.yaml" should contain the field "name" equal to "Ping v2"
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "url" equal to "https://api.example.test/ping?verbose=1"
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "order" with the items "ping.yaml, ping-v2.yaml"

  @save_as
  Scenario: Save as into an unknown folder fails
    When I call "RequestService.SaveAs" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "parent_id": "01JA2Z6A1B2C3D4E5F6G7H8J9K",
        "name": "Ping v2",
        "request": { "name": "Ping", "method": "GET", "url": "https://api.example.test/v2/ping" }
      }
      """
    Then the call should fail with code "TRE-01404"

  @file_refs
  Scenario: Body files inside the workspace are stored relative to it and files outside stay absolute
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{ping}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "RequestService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "base_revision": "{{revision}}",
        "request": {
          "id": "{{ping}}",
          "name": "Ping",
          "method": "POST",
          "url": "https://api.example.test/upload",
          "body": {
            "mode": "formdata",
            "formdata": [
              { "key": "inside", "type": "file", "src": "{{Acme.path}}/fixtures/invoice.pdf" },
              { "key": "outside", "type": "file", "src": "{{temp_root}}/outside/photo.png" },
              { "key": "relative", "type": "file", "src": "fixtures/already-relative.txt" }
            ],
            "binary": { "src": "{{Acme.path}}/fixtures/payload.bin" }
          }
        }
      }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "body.formdata.0.src" equal to "fixtures/invoice.pdf"
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "body.formdata.1.src" equal to "{{temp_root}}/outside/photo.png"
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "body.formdata.2.src" equal to "fixtures/already-relative.txt"
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "body.binary.src" equal to "fixtures/payload.bin"
    And the response should contain the field "body.formdata.0.src" equal to "fixtures/invoice.pdf"

  @validation
  Scenario Outline: Invalid input is rejected - <method>
    When I call "RequestService.<method>" with
      """
      <payload>
      """
    Then the call should fail with code "<code>"

    Examples:
      | method | payload                                                                                     | code      |
      | Get    | { "local_id": "{{Acme.local_id}}", "id": "" }                                               | REQ-08400 |
      | Save   | { "local_id": "", "base_revision": "x", "request": { "id": "{{ping}}", "name": "Ping" } }    | REQ-09400 |
      | Save   | { "local_id": "{{Acme.local_id}}", "base_revision": "x", "request": { "name": "Ping" } }     | REQ-09400 |
      | SaveAs | { "local_id": "{{Acme.local_id}}", "parent_id": "{{billing}}", "name": " ", "request": {} } | REQ-10400 |
