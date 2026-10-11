#language: en
#utf-8

@all @xen_006 @export @xenon_export @xenon_secrets
Feature: Include secrets in a Xenon YAML export, encrypted with a passphrase
  "Include secrets" asks for a passphrase twice (at least 8 characters, both entries equal,
  EXP-01400 otherwise) and adds the exported items' secret values encrypted with it: age with a
  scrypt passphrase recipient, ASCII-armored. A single file carries the block under a top-level
  secrets: key; a zip carries it as xenon-secrets.age next to the manifest. The block is keyed by
  the exported items' ids and needs nothing but the passphrase. Without the option no secret value
  is in the export at all.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
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
        - key: unset_secret
          secret: true
      """
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JX0000000000000000000100
      name: Acme
      variables:
        - key: signing_key
          secret: true
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:01JX0000000000000000000101", "key": "api_key", "value": "env-SECRET-value" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "collection:01JX0000000000000000000001", "key": "api_secret", "value": "collection-SECRET-value" }
      """
    And I call "VariableService.SetCurrent" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "signing_key", "value": "globals-SECRET-value" }
      """

  @environment @file
  Scenario: An environment file carries its secrets in an armored age block
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "01JX0000000000000000000101",
        "format": "xenon_yaml", "shape": "file", "include_secrets": true,
        "passphrase": "correct horse", "passphrase_confirmation": "correct horse",
        "target_path": "{{temp_root}}/local.yaml"
      }
      """
    Then the call should succeed
    And the response should contain the field "secrets" equal to "included"
    And the response should not contain "env-SECRET-value"
    And the file "local.yaml" should contain the field "id" equal to "01JX0000000000000000000101"
    And the file "local.yaml" should contain the field "secrets" matching "^-----BEGIN AGE ENCRYPTED FILE-----\n"
    And the file "local.yaml" should not contain "env-SECRET-value"
    And the field "secrets" of the file "local.yaml" decrypted with the passphrase "correct horse" should contain the field "version" equal to "1"
    And the field "secrets" of the file "local.yaml" decrypted with the passphrase "correct horse" should contain the field "secrets.env:01JX0000000000000000000101.api_key" equal to "env-SECRET-value"
    And the field "secrets" of the file "local.yaml" decrypted with the passphrase "correct horse" should not contain the field "secrets.env:01JX0000000000000000000101.unset_secret"
    And the field "secrets" of the file "local.yaml" decrypted with the passphrase "correct horse" should not contain the field "secrets.globals"
    And the workspace file "acme/environments/local.yaml" should be unchanged

  @collection @bundle
  Scenario: A collection bundle carries the collection's secrets
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "01JX0000000000000000000001",
        "format": "xenon_yaml", "shape": "bundle", "include_secrets": true,
        "passphrase": "correct horse", "passphrase_confirmation": "correct horse",
        "target_path": "{{temp_root}}/shop-api.yaml"
      }
      """
    Then the call should succeed
    And the response should contain the field "secrets" equal to "included"
    And the file "shop-api.yaml" should contain the field "secrets" matching "^-----BEGIN AGE ENCRYPTED FILE-----\n"
    And the file "shop-api.yaml" should not contain "collection-SECRET-value"
    And the field "secrets" of the file "shop-api.yaml" decrypted with the passphrase "correct horse" should contain the field "secrets.collection:01JX0000000000000000000001.api_secret" equal to "collection-SECRET-value"

  @workspace @zip
  Scenario: A workspace zip carries every scope's secrets in xenon-secrets.age
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "workspace",
        "format": "xenon_yaml", "shape": "zip", "include_secrets": true,
        "passphrase": "correct horse", "passphrase_confirmation": "correct horse",
        "target_path": "{{temp_root}}/acme.zip"
      }
      """
    Then the call should succeed
    And the response should contain the field "secrets" equal to "included"
    And the file "acme.zip" should not contain "env-SECRET-value"
    And the file "acme.zip" should not contain "collection-SECRET-value"
    And the file "acme.zip" should not contain "globals-SECRET-value"
    When the zip file "acme.zip" is extracted to "unzipped"
    Then the file "unzipped/xenon-export.yaml" should contain the field "secrets" equal to "file"
    And the file "unzipped/xenon-secrets.age" decrypted with the passphrase "correct horse" should contain the field "secrets.env:01JX0000000000000000000101.api_key" equal to "env-SECRET-value"
    And the file "unzipped/xenon-secrets.age" decrypted with the passphrase "correct horse" should contain the field "secrets.collection:01JX0000000000000000000001.api_secret" equal to "collection-SECRET-value"
    And the file "unzipped/xenon-secrets.age" decrypted with the passphrase "correct horse" should contain the field "secrets.globals.signing_key" equal to "globals-SECRET-value"
    And the file "unzipped/acme/environments/local.yaml" should not contain the field "secrets"

  @request @file
  Scenario: A request has no secrets of its own, so its file carries none
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "request", "item_id": "01JX0000000000000000000031",
        "format": "xenon_yaml", "shape": "file", "include_secrets": true,
        "passphrase": "correct horse", "passphrase_confirmation": "correct horse",
        "target_path": "{{temp_root}}/ping.yaml"
      }
      """
    Then the call should succeed
    And the response should contain the field "secrets" equal to "omitted"
    And the file "ping.yaml" should equal the resource file "export/shop/collections/shop-api/ping.yaml"

  @without_option
  Scenario Outline: Without "Include secrets" no secret value is exported (<kind> as <shape>)
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "<kind>", "item_id": "<id>",
        "format": "xenon_yaml", "shape": "<shape>", "target_path": "{{temp_root}}/export.<extension>"
      }
      """
    Then the call should succeed
    And the response should contain the field "secrets" equal to "omitted"
    And the file "export.<extension>" should not contain "env-SECRET-value"
    And the file "export.<extension>" should not contain "collection-SECRET-value"
    And the file "export.<extension>" should not contain "globals-SECRET-value"
    And the file "export.<extension>" should not contain "BEGIN AGE ENCRYPTED FILE"

    Examples:
      | kind        | id                         | shape  | extension |
      | workspace   |                            | zip    | zip       |
      | collection  | 01JX0000000000000000000001 | zip    | zip       |
      | collection  | 01JX0000000000000000000001 | bundle | yaml      |
      | environment | 01JX0000000000000000000101 | file   | yaml      |

  @passphrase @errors
  Scenario Outline: The passphrase must have 8 characters and be entered twice the same (<case>)
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "01JX0000000000000000000101",
        "format": "xenon_yaml", "shape": "file", "include_secrets": true,
        "passphrase": "<passphrase>", "passphrase_confirmation": "<confirmation>",
        "target_path": "{{temp_root}}/local.yaml"
      }
      """
    Then the call should fail with code "EXP-01400"
    And the file "local.yaml" should not exist

    Examples:
      | case      | passphrase    | confirmation  |
      | too short | short12       | short12       |
      | mismatch  | correct horse | correct house |
      | empty     |               |               |

  @postman @errors
  Scenario: Secrets cannot be included in a Postman export
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "01JX0000000000000000000101",
        "format": "postman", "shape": "file", "include_secrets": true,
        "passphrase": "correct horse", "passphrase_confirmation": "correct horse",
        "target_path": "{{temp_root}}/local.postman_environment.json"
      }
      """
    Then the call should fail with code "EXP-00400"
    And the file "local.postman_environment.json" should not exist
