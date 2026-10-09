#language: en
#utf-8

@all @xen_001 @tree @tree_robustness
Feature: Hand-edited and unusual workspace files
  Workspaces live in git and get edited by hand. A broken file shows up as an error item and is
  never overwritten, fields and comments Xenon does not own survive a save, copied files with the
  same id are told apart, and nothing outside the workspace folder is ever read.

  Background:
    Given the workspace "Acme" exists
    And the "billing" variable is set to "01JB0000000000000000000001"
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
        - broken.yaml
        - ping.yaml
      """
    And the workspace file "acme/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB0000000000000000000004
      name: Ping
      method: GET
      url: https://api.example.test/ping
      """

  @malformed
  Scenario: A file that cannot be parsed becomes an error item and the rest still loads
    Given the workspace file "acme/collections/billing/broken.yaml" contains
      """
      xenon: 1
      kind: request
      name: [unclosed
      """
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "0.children" array length equal to 2
    And the response should contain the field "0.children.0.kind" equal to "error"
    And the response should contain the field "0.children.0.name" equal to "broken.yaml"
    And the response should contain the field "0.children.0.error" containing "cannot be read"
    And the response should contain the field "0.children.1.name" equal to "Ping"
    When I call "TreeService.CreateRequest" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{billing}}", "name": "Pong" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/broken.yaml" should be unchanged
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "order" with the items "broken.yaml, ping.yaml, pong.yaml"

  @malformed
  Scenario: A broken collection file becomes an error item without hiding the other collections
    Given the workspace file "acme/collections/archive/collection.yaml" contains
      """
      kind: [collection
      """
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 2
    And the response should contain the field "0.name" equal to "Billing"
    And the response should contain the field "1.kind" equal to "error"
    And the response should contain the field "1.name" equal to "archive"

  @version
  Scenario: A request written by a newer Xenon is an error item
    Given the workspace file "acme/collections/billing/broken.yaml" contains
      """
      xenon: 2
      kind: request
      id: 01JB0000000000000000000009
      name: Future
      """
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.children.0.kind" equal to "error"
    And the response should contain the field "0.children.0.error" containing "newer Xenon"

  @version
  Scenario: A workspace written by a newer Xenon is refused
    Given the workspace file "acme/workspace.yaml" contains
      """
      xenon: 2
      kind: workspace
      id: 01JB0000000000000000000000
      name: Acme
      """
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should fail with code "WKS-04422"

  @preservation
  Scenario: Unknown fields and comments in a request survive a save
    Given the workspace file "acme/collections/billing/ping.yaml" contains
      """
      # maintained by the platform team
      xenon: 1
      kind: request
      id: 01JB0000000000000000000004
      name: Ping
      x-owner: platform # who to ask
      method: GET
      url: https://api.example.test/ping # old host
      auth:
        type: apikey
        apikey:
          key: X-Key
          value: abc
          in: header
      """
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JB0000000000000000000004" }
      """
    And I save the response field "revision" as "revision"
    And I call "RequestService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "base_revision": "{{revision}}",
        "request": {
          "id": "01JB0000000000000000000004",
          "name": "Ping",
          "method": "GET",
          "url": "https://api.example.test/v2/ping",
          "auth": { "type": "bearer", "bearer": { "token": "abc" } }
        }
      }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/ping.yaml" should equal
      """
      # maintained by the platform team
      xenon: 1
      kind: request
      id: 01JB0000000000000000000004
      name: Ping
      description: ""
      x-owner: platform # who to ask
      method: GET
      url: https://api.example.test/v2/ping # old host
      params: []
      pathVariables: []
      headers: []
      autoHeaders:
        disabled: []
      auth:
        type: bearer
        bearer:
          token: abc
        apikey:
          key: X-Key
          value: abc
          in: header
      body:
        mode: none
      settings:
        timeoutMs: 0
        followRedirects: true
        verifySsl: true
        encodeUrl: true
      scripts:
        preRequest: ""
        postResponse: ""
      inheritScripts: true
      postman: {}
      """

  @preservation
  Scenario: Unknown fields in a collection survive a settings save
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing # shown in the sidebar
      x-team: payments
      order:
        - ping.yaml
      """
    When I call "TreeService.SaveSettings" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "id": "{{billing}}",
        "description": "Billing endpoints",
        "auth": { "type": "none" },
        "headers": [ { "key": "X-Tenant", "value": "acme" } ]
      }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/collection.yaml" should equal
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing # shown in the sidebar
      description: Billing endpoints
      x-team: payments
      order:
        - ping.yaml
      auth:
        type: none
      headers:
        - key: X-Tenant
          value: acme
          enabled: true
          description: ""
      variables: []
      scripts:
        preRequest: ""
        postResponse: ""
      postman: {}
      """

  @duplicate_ids
  Scenario: A copied file with the same id gets a new id and a warning until it is saved
    Given the workspace file "acme/collections/billing/broken.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB0000000000000000000004
      name: Ping (copied by hand)
      method: GET
      url: https://api.example.test/copy
      """
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "0.children.0.id" equal to "01JB0000000000000000000004"
    And the response field "0.children.0.warning" should not exist
    And the response should contain the field "0.children.1.name" equal to "Ping"
    And the response should contain the field "0.children.1.id" not equal to "01JB0000000000000000000004"
    And the response should contain the field "0.children.1.warning" containing "same id"
    And I save the response field "0.children.1.id" as "reidentified"
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.children.1.id" equal to "{{reidentified}}"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{reidentified}}" }
      """
    Then the response should contain the field "url" equal to "https://api.example.test/ping"
    And I save the response field "revision" as "revision"
    When I call "RequestService.Save" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "base_revision": "{{revision}}",
        "request": { "id": "{{reidentified}}", "name": "Ping", "method": "GET", "url": "https://api.example.test/ping" }
      }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/ping.yaml" should contain the field "id" equal to "{{reidentified}}"
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response field "0.children.1.warning" should not exist
    And the response should contain the field "0.children.1.id" equal to "{{reidentified}}"

  @ignored
  Scenario: Directories without a marker file are listed as a warning and their files are ignored
    Given the workspace file "acme/collections/billing/notes/todo.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB0000000000000000000007
      name: Hidden
      """
    And the workspace file "acme/collections/scratch/hidden.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB0000000000000000000008
      name: Hidden too
      """
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 2
    And the response should contain the field "1.kind" equal to "error"
    And the response should contain the field "1.name" equal to "scratch"
    And the response should contain the field "1.error" containing "collection.yaml"
    And the response should contain the field "0.children" array length equal to 2
    And the response should contain the field "0.children.1.kind" equal to "error"
    And the response should contain the field "0.children.1.name" equal to "notes"
    And the response should contain the field "0.children.1.error" containing "folder.yaml"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JB0000000000000000000007" }
      """
    Then the call should fail with code "REQ-01404"

  @symlinks
  Scenario: Symbolic links inside the workspace are ignored
    Given the workspace file "outside/secret.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB0000000000000000000005
      name: Secret
      """
    And the workspace file "outside/folder/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: 01JB0000000000000000000006
      name: Linked folder
      """
    And the workspace path "acme/collections/billing/linked.yaml" is a symbolic link to "outside/secret.yaml"
    And the workspace path "acme/collections/billing/linked" is a symbolic link to "outside/folder"
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "0.children" array length equal to 1
    And the response should contain the field "0.children.0.name" equal to "Ping"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JB0000000000000000000005" }
      """
    Then the call should fail with code "REQ-01404"

  @line_endings
  Scenario: Files with Windows line endings and a byte order mark are read without being rewritten
    Given the workspace file "acme/collections/billing/ping.yaml" contains with Windows line endings and a byte order mark
      """
      xenon: 1
      kind: request
      id: 01JB0000000000000000000004
      name: Ping
      method: POST
      url: https://api.example.test/ping
      """
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.children.0.name" equal to "Ping"
    And the response should contain the field "0.children.0.method" equal to "POST"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JB0000000000000000000004" }
      """
    Then the response should contain the field "url" equal to "https://api.example.test/ping"
    And the workspace file "acme/collections/billing/ping.yaml" should still have Windows line endings and a byte order mark
