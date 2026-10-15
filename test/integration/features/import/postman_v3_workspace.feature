#language: en
#utf-8

@all @xen_009 @import @postman_v3_workspace
Feature: Import a whole `postman workspace pull` folder
  `postman workspace pull` writes a postman/ folder with collections/, environments/ and globals/ (plus parts Xenon
  does not import, such as specs/ and flows/). Picking that folder, the repository folder that holds it, its
  collections/ folder, or a zip of any of them lists every collection, environment and the globals as separate items.
  Like a Postman data dump, each item is assigned its own target workspace and parts that are not imported are listed
  as skipped. Only postman/ is read when the repository folder is picked, so nothing else in the repository can affect
  the import. One broken file refuses the whole source before anything is written.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    And the keychain is available

  @detect
  Scenario Outline: Every collection, environment and the globals are listed - <form>
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/<path>"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "postman_workspace"
    And the response should contain the field "version" equal to "3"
    And the response should contain the field "items" array length equal to 5
    And the response should contain the field "items.0.kind" equal to "collection"
    And the response should contain the field "items.0.name" equal to "Billing API"
    And the response should contain the field "items.0.counts.requests" equal to "2"
    And the response should contain the field "items.1.kind" equal to "collection"
    And the response should contain the field "items.1.name" equal to "Users"
    And the response should contain the field "items.2.kind" equal to "environment"
    And the response should contain the field "items.2.name" equal to "prod"
    And the response should contain the field "items.3.kind" equal to "environment"
    And the response should contain the field "items.3.name" equal to "staging"
    And the response should contain the field "items.3.counts.variables" equal to "2"
    And the response should contain the field "items.4.kind" equal to "globals"
    And the response should contain the field "items.4.name" equal to "Globals"
    And the response field "warnings" should contain 2 items with "code" equal to "skipped_part"
    And the response field "warnings" should contain an item with "code" equal to "skipped_part" and "message" equal to "The workspace's flows part (1 file) is not a collection, an environment or globals and was skipped"
    And the response field "warnings" should contain an item with "code" equal to "skipped_part" and "message" equal to "The workspace's specs part (1 file) is not a collection, an environment or globals and was skipped"
    And the response field "warnings" should contain an item with "code" equal to "skipped_part" and "path" equal to "specs"
    And the response should not contain "s3cr3t-fixture-workspace"

    Examples:
      | form                    | path                       |
      | repository folder       | workspace                  |
      | postman folder          | workspace/postman          |
      | zip of the repository   | workspace.zip              |

  @detect @collections
  Scenario: Picking postman/collections lists only the collections
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/workspace/postman/collections"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "postman_workspace"
    And the response should contain the field "version" equal to "3"
    And the response should contain the field "items" array length equal to 2
    And the response should contain the field "items.0.name" equal to "Billing API"
    And the response should contain the field "items.1.name" equal to "Users"
    And the response field "warnings" should contain 0 items with "code" equal to "skipped_part"

  @assign
  Scenario: Each item lands in the workspace chosen for it, with secrets in that workspace's vault
    Given the log output is captured
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/v3/workspace"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      {
        "source_id": "{{source}}",
        "items": [
          { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" },
          { "key": "1", "include": true, "target_local_id": "{{Globex.local_id}}" },
          { "key": "2", "include": true, "target_local_id": "{{Globex.local_id}}" },
          { "key": "3", "include": true, "target_local_id": "{{Acme.local_id}}" },
          { "key": "4", "include": true, "target_local_id": "{{Globex.local_id}}" }
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "failed" array length equal to 0
    And the response should contain the field "items" array length equal to 5
    And the response should contain the field "items.0.local_id" equal to "{{Acme.local_id}}"
    And the response should contain the field "items.1.local_id" equal to "{{Globex.local_id}}"
    And the response should contain the field "items.3.local_id" equal to "{{Acme.local_id}}"
    And the response should contain the field "workspaces" array length equal to 2
    And I save the response field "items.1.id" as "users"
    And I save the response field "items.3.id" as "staging"
    And the workspace file "acme/collections/billing-api/list-invoices.yaml" should contain the field "method" equal to "GET"
    And the workspace file "acme/collections/billing-api/create-invoice.yaml" should contain the field "body.raw.language" equal to "json"
    And the workspace file "acme/collections/billing-api/collection.yaml" should contain the field "order" with the items "list-invoices.yaml, create-invoice.yaml"
    And the workspace file "globex/collections/users/me.yaml" should contain the field "url" equal to "https://users.acme.test/me"
    And the workspace file "globex/environments/prod.yaml" should contain the field "variables.0.value" equal to "https://acme.test"
    And the workspace file "acme/environments/staging.yaml" should contain the field "variables.1.secret" equal to "true"
    And the workspace file "globex/workspace.yaml" should contain the field "variables.0.key" equal to "tenant"
    And the workspace file "acme/workspace.yaml" should contain the field "order" with the items "billing-api"
    And the workspace file "globex/workspace.yaml" should contain the field "order" with the items "users"
    And no file under the workspace folder should contain "s3cr3t-fixture-workspace-env"
    And no file under the workspace folder should contain "s3cr3t-fixture-workspace-collection"
    And the log output should not contain "s3cr3t-fixture-workspace"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{staging}}", "key": "stagingKey" }
      """
    Then the response should contain the field "value" equal to "s3cr3t-fixture-workspace-env"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Globex.local_id}}", "scope": "collection:{{users}}", "key": "usersSecret" }
      """
    Then the response should contain the field "value" equal to "s3cr3t-fixture-workspace-collection"

  @broken
  Scenario: One broken file refuses the whole workspace before anything is written
    Given the file "broken/postman/collections/Mini/Ping.request.yaml" exists with content
      """
      $kind: http-request
      url: https://mini.test/ping
      method: GET
      """
    And the file "broken/postman/environments/bad.environment.yaml" exists with content
      """
      name: bad
      values: 'oops'
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/broken"] }
      """
    Then the call should fail with code "IMP-11400"
    And the error should contain the field "error.description" equal to "environments/bad.environment.yaml is not a valid Postman v3 file: line 2: a value has the wrong type"
    And the workspace folder "acme" should contain the entries ".gitignore, collections, environments, workspace.yaml"

  @nothing @xen_014 @xen_014_loose
  Scenario: A postman folder with nothing importable is not a Postman workspace
    Given the file "specs-only/postman/specs/openapi.yaml" exists with content
      """
      openapi: 3.0.0
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/specs-only"] }
      """
    Then the call should fail with code "IMP-00415"
    And the error should contain the field "error.description" equal to "specs-only has no file Xenon can import; pass a Postman v3 workspace or collection folder, a Postman data dump (archive.json), a Bruno collection (bruno.json or opencollection.yml), a Xenon workspace folder, or a folder of standalone export files"
