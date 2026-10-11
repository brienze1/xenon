#language: en
#utf-8

@all @xen_006 @import @xenon_import
Feature: Import Xenon YAML exports
  The Import dialog accepts every Xenon YAML export: a workspace zip or folder (imported as a new
  workspace in an empty or new folder, its files and ids kept), a collection or folder zip or
  bundle, a request file and an environment file (imported into an existing workspace with new ids
  and the XEN-003 clash handling). Encrypted secrets need the passphrase: the right one puts the
  values in the target workspace's vault, a missing one is refused with IMP-50401 and a wrong one
  with IMP-08401 before anything is written, and "import without secrets" leaves them missing. A file from a newer
  format version is refused. Exporting what was imported gives the same bytes, ids aside.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Beta" exists
    And the keychain is available
    And the current time is "2026-10-11T10:00:00Z"
    And the resource folder "export/shop" is copied to the workspace folder "acme"
    And the workspace file "acme/environments/local.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JX0000000000000000000101
      name: local
      variables:
        - key: base_url
          value: https://local.acme.test
        - key: api_key
          secret: true
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:01JX0000000000000000000101", "key": "api_key", "value": "import-SECRET-key" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "collection:01JX0000000000000000000001", "key": "api_secret", "value": "collection-SECRET-key" }
      """

  @detect
  Scenario Outline: Every Xenon YAML export kind is detected (<item_kind> as <shape>)
    Given I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "<item_kind>", "item_id": "<item_id>",
        "format": "xenon_yaml", "shape": "<shape>", "target_path": "{{temp_root}}/<file>"
      }
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/<file>"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "<kind>"
    And the response should contain the field "encrypted_secrets" equal to "false"
    And the response should contain the field "items" array length equal to 1
    And the response should contain the field "items.0.kind" equal to "<kind_of_item>"
    And the response should contain the field "items.0.name" equal to "<name>"

    Examples:
      | item_kind   | item_id                    | shape  | file                | kind                | kind_of_item | name           |
      | workspace   |                            | zip    | acme.zip            | xenon_workspace_zip | workspace    | Acme           |
      | collection  | 01JX0000000000000000000001 | zip    | shop-api.zip        | xenon_zip           | collection   | Shop API       |
      | folder      | 01JX0000000000000000000002 | zip    | products.zip        | xenon_zip           | folder       | Products       |
      | collection  | 01JX0000000000000000000001 | bundle | shop-api.yaml       | xenon_bundle        | collection   | Shop API       |
      | folder      | 01JX0000000000000000000003 | bundle | orders.yaml         | xenon_bundle        | folder       | Orders         |
      | request     | 01JX0000000000000000000013 | file   | create-product.yaml | xenon_request       | request      | Create product |
      | environment | 01JX0000000000000000000101 | file   | local.yaml          | xenon_environment   | environment  | local          |

  @detect @workspace_folder
  Scenario: A workspace folder is detected as a Xenon workspace
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{Acme.path}}"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "xenon_workspace_folder"
    And the response should contain the field "items.0.kind" equal to "workspace"
    And the response should contain the field "items.0.name" equal to "Acme"
    And the response should contain the field "items.0.counts.collections" equal to "1"
    And the response should contain the field "items.0.counts.requests" equal to "9"
    And the response should contain the field "items.0.counts.environments" equal to "1"

  @collection @bundle
  Scenario: A collection bundle is imported with new ids
    Given I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "01JX0000000000000000000001",
        "format": "xenon_yaml", "shape": "bundle", "target_path": "{{temp_root}}/shop-api.yaml"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/shop-api.yaml"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the response should contain the field "items.0.id" not equal to "01JX0000000000000000000001"
    And the response should contain the field "workspaces.0.created.collections" equal to "1"
    And the response should contain the field "workspaces.0.created.folders" equal to "2"
    And the response should contain the field "workspaces.0.created.requests" equal to "9"
    And the response should contain the field "failed" array length equal to 0
    And I save the response field "items.0.id" as "imported"
    And the workspace file "beta/collections/shop-api/collection.yaml" should contain the field "id" equal to "{{imported}}"
    And the workspace file "beta/collections/shop-api/collection.yaml" should contain the field "name" equal to "Shop API"
    And the workspace file "beta/collections/shop-api/collection.yaml" should contain the field "order" with the items "products, orders, ping.yaml"
    And the workspace file "beta/collections/shop-api/products/folder.yaml" should contain the field "id" not equal to "01JX0000000000000000000002"
    And the workspace file "beta/collections/shop-api/products/folder.yaml" should contain the field "order" with the items "list-products.yaml, get-product.yaml, create-product.yaml"
    And the workspace file "beta/collections/shop-api/products/create-product.yaml" should contain the field "id" not equal to "01JX0000000000000000000013"
    And the workspace file "beta/collections/shop-api/products/create-product.yaml" should contain the field "name" equal to "Create product"
    And the workspace file "beta/collections/shop-api/collection.yaml" should contain the field "variables.1.value" equal to "20"
    And the workspace file "beta/collections/shop-api/collection.yaml" should not contain the field "secrets"

  @collection @zip @clash
  Scenario: A collection zip imported twice reports the clash and keeps both when asked
    Given I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "01JX0000000000000000000001",
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/shop-api.zip"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/shop-api.zip"] }
      """
    And I save the response field "source_id" as "first"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{first}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/shop-api.zip"] }
      """
    And I save the response field "source_id" as "second"
    When I call "ImportService.Clashes" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "clashes" array length equal to 1
    And the response should contain the field "clashes.0.kind" equal to "collection"
    And the response should contain the field "clashes.0.existing_name" equal to "Shop API"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}", "resolution": "keep_both" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the workspace file "beta/workspace.yaml" should contain the field "order" with the items "shop-api, shop-api-(imported)"
    And the workspace file "beta/collections/shop-api-(imported)/products/get-product.yaml" should contain the field "name" equal to "Get product"

  @folder @zip
  Scenario: A folder zip is imported as a folder into a collection
    Given I call "TreeService.CreateCollection" with
      """
      { "local_id": "{{Beta.local_id}}", "name": "Main" }
      """
    And I save the response field "id" as "main"
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "folder", "item_id": "01JX0000000000000000000002",
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/products.zip"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/products.zip"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}", "target_parent_id": "{{main}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "workspaces.0.created.folders" equal to "1"
    And the response should contain the field "workspaces.0.created.requests" equal to "3"
    And the workspace file "beta/collections/main/collection.yaml" should contain the field "order" with the items "products"
    And the workspace file "beta/collections/main/products/folder.yaml" should contain the field "kind" equal to "folder"
    And the workspace file "beta/collections/main/products/folder.yaml" should contain the field "id" not equal to "01JX0000000000000000000002"
    And the workspace file "beta/collections/main/products/list-products.yaml" should contain the field "name" equal to "List products"

  @folder @bundle
  Scenario: A folder bundle imported at the top level becomes a collection
    Given I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "folder", "item_id": "01JX0000000000000000000003",
        "format": "xenon_yaml", "shape": "bundle", "target_path": "{{temp_root}}/orders.yaml"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/orders.yaml"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "workspaces.0.created.collections" equal to "1"
    And the response should contain the field "workspaces.0.created.requests" equal to "5"
    And the workspace file "beta/collections/orders/collection.yaml" should contain the field "kind" equal to "collection"
    And the workspace file "beta/collections/orders/collection.yaml" should contain the field "name" equal to "Orders"
    And the workspace file "beta/collections/orders/signed.yaml" should contain the field "name" equal to "Signed"

  @request @file
  Scenario: A request file is imported into a collection with a new id
    Given I call "TreeService.CreateCollection" with
      """
      { "local_id": "{{Beta.local_id}}", "name": "Main" }
      """
    And I save the response field "id" as "main"
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "request", "item_id": "01JX0000000000000000000013",
        "format": "xenon_yaml", "shape": "file", "target_path": "{{temp_root}}/create-product.yaml"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/create-product.yaml"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}", "target_parent_id": "{{main}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.id" not equal to "01JX0000000000000000000013"
    And I save the response field "items.0.id" as "imported"
    And the workspace file "beta/collections/main/create-product.yaml" should contain the field "id" equal to "{{imported}}"
    And the workspace file "beta/collections/main/create-product.yaml" should contain the field "method" equal to "POST"

  @environment @file @clash
  Scenario: An environment file is imported with a new id and clashes by name
    Given I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "01JX0000000000000000000101",
        "format": "xenon_yaml", "shape": "file", "target_path": "{{temp_root}}/local.yaml"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/local.yaml"] }
      """
    And I save the response field "source_id" as "first"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{first}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.id" not equal to "01JX0000000000000000000101"
    And I save the response field "items.0.id" as "imported"
    And the workspace file "beta/environments/local.yaml" should contain the field "id" equal to "{{imported}}"
    And the workspace file "beta/environments/local.yaml" should contain the field "variables.1.secret" equal to "true"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Beta.local_id}}", "scope": "env:{{imported}}" }
      """
    Then the response should contain the field "rows.1.secret_state" equal to "missing"
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/local.yaml"] }
      """
    And I save the response field "source_id" as "second"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    Then the call should fail with code "IMP-04409"

  @workspace @zip
  Scenario: A workspace zip becomes a new workspace in an empty folder, ids kept
    Given I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "workspace",
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/acme.zip"
      }
      """
    And the zip file "acme.zip" is extracted to "unzipped"
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/acme.zip"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_folder": "{{temp_root}}/restored" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the response should contain the field "items.0.name" equal to "Acme"
    And the response should contain the field "workspaces" array length equal to 1
    And the response should contain the field "workspaces.0.created.collections" equal to "1"
    And the response should contain the field "workspaces.0.created.requests" equal to "9"
    And the response should contain the field "workspaces.0.created.environments" equal to "1"
    And I save the response field "workspaces.0.local_id" as "restored"
    And the folder "restored" should equal the folder "unzipped/acme"
    When I call "WorkspaceService.List"
    Then the response field "workspaces" should contain an item with "local_id" equal to "{{restored}}" and "name" equal to "Acme"
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{restored}}" }
      """
    Then the call should succeed
    And the response should contain the field "0.id" equal to "01JX0000000000000000000001"

  @workspace @folder
  Scenario: A workspace folder becomes a new workspace without its local-only files
    Given the workspace file "acme/.xenon/state.yaml" contains
      """
      local: only
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{Acme.path}}"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_folder": "{{temp_root}}/copy" } ] }
      """
    Then the call should succeed
    And the file "copy/workspace.yaml" should equal the file "workspaces/acme/workspace.yaml"
    And the folder "copy/collections" should equal the folder "workspaces/acme/collections"
    And the file "copy/.xenon/state.yaml" should not exist

  @workspace @errors
  Scenario Outline: A workspace import needs an empty or new folder (<case>)
    Given the file "taken/other.txt" exists with content
      """
      not empty
      """
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "workspace",
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/acme.zip"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/acme.zip"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true<target> } ] }
      """
    Then the call should fail with code "<code>"
    And the file "taken/workspace.yaml" should not exist

    Examples:
      | case               | target                                                      | code      |
      | not empty          | , "target_folder": "{{temp_root}}/taken"                    | WKS-01409 |
      | inside a workspace | , "target_folder": "{{Beta.path}}/nested"                   | WKS-02409 |
      | no folder          | , "target_local_id": "{{Beta.local_id}}"                    | IMP-43400 |

  @secrets @environment
  Scenario: Encrypted secrets go into the target vault with the right passphrase
    Given I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "01JX0000000000000000000101",
        "format": "xenon_yaml", "shape": "file", "include_secrets": true,
        "passphrase": "correct horse", "passphrase_confirmation": "correct horse",
        "target_path": "{{temp_root}}/local.yaml"
      }
      """
    And the log output is captured
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/local.yaml"] }
      """
    Then the call should succeed
    And the response should contain the field "encrypted_secrets" equal to "true"
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    Then the call should fail with code "IMP-50401"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "passphrase": "wrong horse", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    Then the call should fail with code "IMP-08401"
    And the workspace file "beta/environments/local.yaml" should not exist
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "passphrase": "correct horse", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    Then the call should succeed
    And I save the response field "items.0.id" as "imported"
    And the response should not contain "import-SECRET-key"
    And the workspace file "beta/environments/local.yaml" should not contain the field "secrets"
    And no file under the workspace folder should contain "import-SECRET-key"
    And the log output should not contain "import-SECRET-key"
    And the log output should not contain "correct horse"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Beta.local_id}}", "scope": "env:{{imported}}", "key": "api_key" }
      """
    Then the call should succeed
    And the response should contain the field "value" equal to "import-SECRET-key"

  @secrets @skip
  Scenario: Importing without secrets leaves them missing
    Given I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "01JX0000000000000000000101",
        "format": "xenon_yaml", "shape": "file", "include_secrets": true,
        "passphrase": "correct horse", "passphrase_confirmation": "correct horse",
        "target_path": "{{temp_root}}/local.yaml"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/local.yaml"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "skip_secrets": true, "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    Then the call should succeed
    And I save the response field "items.0.id" as "imported"
    When I call "VariableService.GetTable" with
      """
      { "local_id": "{{Beta.local_id}}", "scope": "env:{{imported}}" }
      """
    Then the response should contain the field "rows.1.key" equal to "api_key"
    And the response should contain the field "rows.1.secret_state" equal to "missing"

  @secrets @collection
  Scenario: A collection's encrypted secrets follow it to its new id
    Given I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "01JX0000000000000000000001",
        "format": "xenon_yaml", "shape": "zip", "include_secrets": true,
        "passphrase": "correct horse", "passphrase_confirmation": "correct horse",
        "target_path": "{{temp_root}}/shop-api.zip"
      }
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/shop-api.zip"] }
      """
    Then the response should contain the field "encrypted_secrets" equal to "true"
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "passphrase": "correct horse", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    Then the call should succeed
    And I save the response field "items.0.id" as "imported"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{Beta.local_id}}", "scope": "collection:{{imported}}", "key": "api_secret" }
      """
    Then the call should succeed
    And the response should contain the field "value" equal to "collection-SECRET-key"

  @secrets @workspace
  Scenario: A workspace zip's encrypted secrets go into the new workspace's vault
    Given I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "workspace",
        "format": "xenon_yaml", "shape": "zip", "include_secrets": true,
        "passphrase": "correct horse", "passphrase_confirmation": "correct horse",
        "target_path": "{{temp_root}}/acme.zip"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/acme.zip"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "passphrase": "wrong horse", "items": [ { "key": "0", "include": true, "target_folder": "{{temp_root}}/restored" } ] }
      """
    Then the call should fail with code "IMP-08401"
    And the file "restored/workspace.yaml" should not exist
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "passphrase": "correct horse", "items": [ { "key": "0", "include": true, "target_folder": "{{temp_root}}/restored" } ] }
      """
    Then the call should succeed
    And I save the response field "workspaces.0.local_id" as "restored"
    And the file "restored/xenon-secrets.age" should not exist
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{restored}}", "scope": "env:01JX0000000000000000000101", "key": "api_key" }
      """
    Then the call should succeed
    And the response should contain the field "value" equal to "import-SECRET-key"
    When I call "VaultService.Reveal" with
      """
      { "local_id": "{{restored}}", "scope": "collection:01JX0000000000000000000001", "key": "api_secret" }
      """
    Then the response should contain the field "value" equal to "collection-SECRET-key"

  @version
  Scenario: A request file from a newer format version is refused
    Given the file "request.yaml" exists with content
      """
      xenon: 2
      kind: request
      id: 01JX0000000000000000000900
      name: Next
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/request.yaml"] }
      """
    Then the call should fail with code "IMP-42422"
    And the error should contain the field "error.description" equal to "request.yaml was written by a newer Xenon (format 2); update Xenon to import it"

  @version
  Scenario: A bundle from a newer format version is refused
    Given the file "bundle.yaml" exists with content
      """
      xenon: 2
      kind: bundle
      root: collection
      item: {}
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/bundle.yaml"] }
      """
    Then the call should fail with code "IMP-42422"

  @version @zip
  Scenario: A zip whose manifest has a newer format version is refused
    Given the file "next/xenon-export.yaml" exists with content
      """
      xenon: 2
      kind: export
      root: collection
      """
    And the file "next/thing/collection.yaml" exists with content
      """
      xenon: 1
      kind: collection
      id: 01JX0000000000000000000900
      name: Thing
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/next"] }
      """
    Then the call should fail with code "IMP-42422"

  @roundtrip
  Scenario Outline: Importing an export and exporting it again gives the same bytes, ids aside (<item_kind> as <shape>)
    Given I call "TreeService.CreateCollection" with
      """
      { "local_id": "{{Beta.local_id}}", "name": "Holder" }
      """
    And I save the response field "id" as "beta_holder"
    And the workspace "Gamma" exists
    And I call "TreeService.CreateCollection" with
      """
      { "local_id": "{{Gamma.local_id}}", "name": "Holder" }
      """
    And I save the response field "id" as "gamma_holder"
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "<item_kind>", "item_id": "<item_id>",
        "format": "xenon_yaml", "shape": "<shape>", "target_path": "{{temp_root}}/first-<file>"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/first-<file>"] }
      """
    And I save the response field "source_id" as "first_source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{first_source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}"<beta_parent> } ] }
      """
    And I save the response field "items.0.id" as "beta_item"
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Beta.local_id}}", "item_kind": "<item_kind>", "item_id": "{{beta_item}}",
        "format": "xenon_yaml", "shape": "<shape>", "target_path": "{{temp_root}}/second-<file>"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/second-<file>"] }
      """
    And I save the response field "source_id" as "second_source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{second_source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Gamma.local_id}}"<gamma_parent> } ] }
      """
    And I save the response field "items.0.id" as "gamma_item"
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Gamma.local_id}}", "item_kind": "<item_kind>", "item_id": "{{gamma_item}}",
        "format": "xenon_yaml", "shape": "<shape>", "target_path": "{{temp_root}}/third-<file>"
      }
      """
    Then the call should succeed
    And the file "third-<file>" should equal the file "second-<file>" apart from ids

    Examples:
      | item_kind   | item_id                    | shape  | file                | beta_parent                             | gamma_parent                             |
      | collection  | 01JX0000000000000000000001 | bundle | shop-api.yaml       |                                         |                                          |
      | folder      | 01JX0000000000000000000003 | bundle | orders.yaml         | , "target_parent_id": "{{beta_holder}}" | , "target_parent_id": "{{gamma_holder}}" |
      | request     | 01JX0000000000000000000013 | file   | create-product.yaml | , "target_parent_id": "{{beta_holder}}" | , "target_parent_id": "{{gamma_holder}}" |
      | environment | 01JX0000000000000000000101 | file   | local.yaml          |                                         |                                          |

  @roundtrip @zip
  Scenario: A collection zip round trip gives the same files, ids aside
    Given the workspace "Gamma" exists
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "01JX0000000000000000000001",
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/first.zip"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/first.zip"] }
      """
    And I save the response field "source_id" as "first_source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{first_source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Beta.local_id}}" } ] }
      """
    And I save the response field "items.0.id" as "beta_item"
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Beta.local_id}}", "item_kind": "collection", "item_id": "{{beta_item}}",
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/second.zip"
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/second.zip"] }
      """
    And I save the response field "source_id" as "second_source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{second_source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Gamma.local_id}}" } ] }
      """
    And I save the response field "items.0.id" as "gamma_item"
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Gamma.local_id}}", "item_kind": "collection", "item_id": "{{gamma_item}}",
        "format": "xenon_yaml", "shape": "zip", "target_path": "{{temp_root}}/third.zip"
      }
      """
    Then the call should succeed
    When the zip file "second.zip" is extracted to "second"
    And the zip file "third.zip" is extracted to "third"
    Then the folder "third" should equal the folder "second" apart from ids
