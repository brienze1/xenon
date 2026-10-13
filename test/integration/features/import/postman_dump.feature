#language: en
#utf-8

@all @xen_003 @import @postman_dump
Feature: Sort a Postman data dump into workspaces
  A Postman data dump (a zip or the unzipped folder with archive.json) carries collections and
  environments from all of Luis's Postman workspaces with no workspace membership. Detecting it
  lists every collection and environment as its own item; parts that are neither (mocks,
  monitors, …) are skipped and listed. The plan gives each item its own target workspace — one
  can be created inline while planning — and one import writes each item into its workspace and
  returns one combined summary with the results per workspace.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    And the keychain is available

  @detect
  Scenario Outline: Every collection and environment of the dump is listed - <form>
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/<path>"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "postman_dump"
    And the response should contain the field "items" array length equal to 5
    And the response should contain the field "items.0.kind" equal to "collection"
    And the response should contain the field "items.0.name" equal to "Billing API"
    And the response should contain the field "items.0.counts.requests" equal to "2"
    And the response should contain the field "items.1.kind" equal to "collection"
    And the response should contain the field "items.1.name" equal to "CRM API"
    And the response should contain the field "items.2.kind" equal to "collection"
    And the response should contain the field "items.2.name" equal to "Ops API"
    And the response should contain the field "items.2.counts.requests" equal to "3"
    And the response should contain the field "items.3.kind" equal to "environment"
    And the response should contain the field "items.3.name" equal to "billing-dev"
    And the response should contain the field "items.3.counts.variables" equal to "2"
    And the response should contain the field "items.4.kind" equal to "environment"
    And the response should contain the field "items.4.name" equal to "crm-prod"
    And the response field "warnings" should contain an item with "code" equal to "skipped_part" and "path" equal to "mock"
    And the response should not contain "tok-billing-SECRET"

    Examples:
      | form   | path     |
      | zip    | dump.zip |
      | folder | dump     |

  @assign
  Scenario: Each item lands in the workspace chosen for it, including one created while planning
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dump.zip"] }
      """
    And I save the response field "source_id" as "source"
    And I call "WorkspaceService.Create" with
      """
      { "name": "Initech", "path": "{{workspace_root}}/initech" }
      """
    And I save the response field "local_id" as "initech"
    And I call "ImportService.Import" with
      """
      {
        "source_id": "{{source}}",
        "items": [
          { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" },
          { "key": "1", "include": true, "target_local_id": "{{Globex.local_id}}" },
          { "key": "2", "include": true, "target_local_id": "{{initech}}" },
          { "key": "3", "include": true, "target_local_id": "{{Acme.local_id}}" },
          { "key": "4", "include": true, "target_local_id": "{{Globex.local_id}}" }
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "failed" array length equal to 0
    And the response should contain the field "items" array length equal to 5
    And the response should contain the field "items.0.local_id" equal to "{{Acme.local_id}}"
    And the response should contain the field "items.3.status" equal to "imported"
    And the response should contain the field "items.3.local_id" equal to "{{Acme.local_id}}"
    And the response should contain the field "items.4.local_id" equal to "{{Globex.local_id}}"
    And the response should contain the field "workspaces" array length equal to 3
    And the response should contain the field "workspaces.0.local_id" equal to "{{Acme.local_id}}"
    And the response should contain the field "workspaces.0.created.collections" equal to "1"
    And the response should contain the field "workspaces.0.created.requests" equal to "2"
    And the response should contain the field "workspaces.0.created.environments" equal to "1"
    And the response should contain the field "workspaces.0.created.variables" equal to "2"
    And the response should contain the field "workspaces.1.local_id" equal to "{{Globex.local_id}}"
    And the response should contain the field "workspaces.1.created.collections" equal to "1"
    And the response should contain the field "workspaces.1.created.environments" equal to "1"
    And the response should contain the field "workspaces.2.local_id" equal to "{{initech}}"
    And the response should contain the field "workspaces.2.created.collections" equal to "1"
    And the response should contain the field "workspaces.2.created.requests" equal to "3"
    And the response should contain the field "workspaces.2.created.environments" equal to "0"
    And the response field "warnings" should contain an item with "code" equal to "skipped_part" and "path" equal to "mock"
    And the response should not contain "tok-billing-SECRET"
    And the workspace folder "acme/collections" should contain the entries "billing-api"
    And the workspace folder "acme/environments" should contain the entries "billing-dev.yaml"
    And the workspace folder "globex/collections" should contain the entries "crm-api"
    And the workspace folder "globex/environments" should contain the entries "crm-prod.yaml"
    And the workspace folder "initech/collections" should contain the entries "ops-api"
    And the workspace folder "initech/environments" should contain the entries ""
    And the workspace file "acme/collections/billing-api/collection.yaml" should contain the field "order" with the items "list-invoices.yaml, create-invoice.yaml"
    And the workspace file "acme/collections/billing-api/collection.yaml" should contain the field "postman.info._postman_id" equal to "11111111-aaaa-4000-8000-000000000001"
    And the workspace file "globex/environments/crm-prod.yaml" should contain the field "color" equal to "red"
    And the workspace file "globex/environments/crm-prod.yaml" should contain the field "confirmBeforeSend" equal to "true"
    And the workspace file "acme/environments/billing-dev.yaml" should contain the field "variables.1.secret" equal to "true"
    And no file under the workspace folder should contain "tok-billing-SECRET"
    When I call "WorkspaceService.List"
    Then the response should contain the field "workspaces" array length equal to 3

  @include
  Scenario: Items left out of the plan are not imported
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dump"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      {
        "source_id": "{{source}}",
        "items": [
          { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" },
          { "key": "1", "include": false },
          { "key": "2", "include": false },
          { "key": "3", "include": false },
          { "key": "4", "include": true, "target_local_id": "{{Acme.local_id}}" }
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 2
    And the response should contain the field "workspaces" array length equal to 1
    And the response should contain the field "workspaces.0.created.collections" equal to "1"
    And the response should contain the field "workspaces.0.created.environments" equal to "1"
    And the workspace folder "acme/collections" should contain the entries "billing-api"
    And the workspace folder "acme/environments" should contain the entries "crm-prod.yaml"
    And the workspace folder "globex/collections" should contain the entries ""

  @clash
  Scenario: Clashes are found per target workspace and resolved once for the whole dump
    Given the workspace file "globex/environments/crm-prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JF0000000000000000000002
      name: crm-prod
      color: none
      variables: []
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dump.zip"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Clashes" with
      """
      {
        "source_id": "{{source}}",
        "items": [
          { "key": "1", "include": true, "target_local_id": "{{Globex.local_id}}" },
          { "key": "4", "include": true, "target_local_id": "{{Globex.local_id}}" },
          { "key": "3", "include": true, "target_local_id": "{{Acme.local_id}}" }
        ]
      }
      """
    Then the call should succeed
    And the response should contain the field "clashes" array length equal to 1
    And the response should contain the field "clashes.0.item_key" equal to "4"
    And the response should contain the field "clashes.0.kind" equal to "environment"
    When I call "ImportService.Import" with
      """
      {
        "source_id": "{{source}}",
        "items": [
          { "key": "1", "include": true, "target_local_id": "{{Globex.local_id}}" },
          { "key": "4", "include": true, "target_local_id": "{{Globex.local_id}}" },
          { "key": "3", "include": true, "target_local_id": "{{Acme.local_id}}" }
        ],
        "apply_to_all": "keep_both"
      }
      """
    Then the call should succeed
    And the workspace folder "globex/environments" should contain the entries "crm-prod-(imported).yaml, crm-prod.yaml"
    And the workspace folder "acme/environments" should contain the entries "billing-dev.yaml"
