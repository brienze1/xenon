#language: en
#utf-8

@all @xen_003 @import @clashes
Feature: Resolve name clashes when importing
  Every kind of clash (a collection, an environment, a request in the target folder, a collection
  variable) can be resolved with Replace, Keep both (the imported one gets " (imported)"; a variable
  keeps its existing value and the imported value is recorded in the summary) or Skip, per item or
  once for everything with "apply to all". Clashes are listed before importing, and an import with
  an unresolved clash is refused with IMP-04409 before anything is written.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And I call "TreeService.CreateCollection" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "Main" }
      """
    And I save the response field "id" as "main"
    And I call "TreeService.CreateFolder" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{main}}", "name": "Users" }
      """
    And I save the response field "id" as "users"

  @collection
  Scenario Outline: A collection clash is resolved with <resolution>
    Given I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "first"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{first}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    And I save the response field "items.0.id" as "original"
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "second"
    And I call "ImportService.Clashes" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "clashes" array length equal to 1
    And the response should contain the field "clashes.0.kind" equal to "collection"
    And the response should contain the field "clashes.0.existing_name" equal to "Bruno Shop"
    And the response should contain the field "clashes.0.existing_id" equal to "{{original}}"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should fail with code "IMP-04409"
    And the workspace file "acme/workspace.yaml" should contain the field "order" with the items "main, bruno-shop"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "resolution": "<resolution>" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "<status>"
    And the workspace file "acme/workspace.yaml" should contain the field "order" with the items "<order>"

    Examples:
      | resolution | status   | order                                      |
      | replace    | replaced | main, bruno-shop                           |
      | keep_both  | imported | main, bruno-shop, bruno-shop-(imported)    |
      | skip       | skipped  | main, bruno-shop                           |

  @collection @replace
  Scenario: Replacing a collection swaps it for the imported one
    Given I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "first"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{first}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    And I save the response field "items.0.id" as "original"
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "second"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "resolution": "replace" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/bruno-shop/collection.yaml" should contain the field "id" not equal to "{{original}}"
    And the workspace folder "acme/collections" should contain the entries "bruno-shop, main"

  @environment
  Scenario Outline: An environment clash is resolved with <resolution>
    Given I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "first"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{first}}", "items": [ { "key": "1", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "second"
    And I call "ImportService.Clashes" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "1", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the response should contain the field "clashes" array length equal to 1
    And the response should contain the field "clashes.0.kind" equal to "environment"
    And the response should contain the field "clashes.0.existing_name" equal to "local"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "1", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should fail with code "IMP-04409"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "1", "include": true, "target_local_id": "{{Acme.local_id}}", "resolution": "<resolution>" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "<status>"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments" array length equal to <count>
    And the response should contain the field "environments.<last>.name" equal to "<name>"

    Examples:
      | resolution | status   | count | last | name             |
      | replace    | replaced | 1     | 0    | local            |
      | keep_both  | imported | 2     | 1    | local (imported) |
      | skip       | skipped  | 1     | 0    | local            |

  @request
  Scenario Outline: A request clash in the target folder is resolved with <resolution>
    Given I call "ImportService.Detect" with
      """
      { "text": "curl https://api.test/users" }
      """
    And I save the response field "source_id" as "first"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{first}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{users}}", "name": "List users" } ] }
      """
    When I call "ImportService.Detect" with
      """
      { "text": "curl -X HEAD https://api.test/users" }
      """
    And I save the response field "source_id" as "second"
    And I call "ImportService.Clashes" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{users}}", "name": "List users" } ] }
      """
    Then the response should contain the field "clashes" array length equal to 1
    And the response should contain the field "clashes.0.kind" equal to "request"
    And the response should contain the field "clashes.0.existing_name" equal to "List users"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{users}}", "name": "List users" } ] }
      """
    Then the call should fail with code "IMP-04409"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{users}}", "name": "List users", "resolution": "<resolution>" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "<status>"
    And the workspace file "acme/collections/main/users/folder.yaml" should contain the field "order" with the items "<order>"
    And the workspace file "acme/collections/main/users/list-users.yaml" should contain the field "method" equal to "<method>"

    Examples:
      | resolution | status   | order                                        | method |
      | replace    | replaced | list-users.yaml                              | HEAD   |
      | keep_both  | imported | list-users.yaml, list-users-(imported).yaml  | GET    |
      | skip       | skipped  | list-users.yaml                              | GET    |

  @variable
  Scenario Outline: A collection variable clash is resolved with <resolution>
    Given the workspace file "acme/collections/main/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{main}}
      name: Main
      description: ""
      order: [users]
      auth:
        type: inherit
      headers: []
      variables:
        - key: api_version
          value: v1
          enabled: true
          description: ""
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "source"
    When I call "ImportService.Clashes" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{main}}" } ] }
      """
    Then the response should contain the field "clashes" array length equal to 1
    And the response should contain the field "clashes.0.kind" equal to "variable"
    And the response should contain the field "clashes.0.existing_name" equal to "api_version"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{main}}" } ] }
      """
    Then the call should fail with code "IMP-04409"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{main}}", "resolution": "<resolution>" } ] }
      """
    Then the call should succeed
    And the response field "warnings" should contain <kept> items with "code" equal to "variable_kept"
    And the workspace file "acme/collections/main/collection.yaml" should contain the field "variables.0.key" equal to "api_version"
    And the workspace file "acme/collections/main/collection.yaml" should contain the field "variables.0.value" equal to "<value>"
    And the workspace file "acme/collections/main/collection.yaml" should contain the field "variables.1.key" equal to "unused"

    Examples:
      | resolution | kept | value |
      | replace    | 0    | v2    |
      | keep_both  | 1    | v1    |
      | skip       | 0    | v1    |

  @apply_to_all
  Scenario: "Apply to all" resolves every clash of every item at once
    Given I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "first"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{first}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" }, { "key": "1", "include": true, "target_local_id": "{{Acme.local_id}}" }, { "key": "2", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/bru-collection"] }
      """
    And I save the response field "source_id" as "second"
    And I call "ImportService.Clashes" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" }, { "key": "1", "include": true, "target_local_id": "{{Acme.local_id}}" }, { "key": "2", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the response should contain the field "clashes" array length equal to 3
    When I call "ImportService.Import" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" }, { "key": "1", "include": true, "target_local_id": "{{Acme.local_id}}", "resolution": "skip" }, { "key": "2", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should fail with code "IMP-04409"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{second}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" }, { "key": "1", "include": true, "target_local_id": "{{Acme.local_id}}", "resolution": "skip" }, { "key": "2", "include": true, "target_local_id": "{{Acme.local_id}}" } ], "apply_to_all": "keep_both" }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the response should contain the field "items.0.name" equal to "Bruno Shop (imported)"
    And the response should contain the field "items.1.status" equal to "skipped"
    And the response should contain the field "items.2.status" equal to "imported"
    And the response should contain the field "items.2.name" equal to "prod (imported)"
    And the workspace file "acme/workspace.yaml" should contain the field "order" with the items "main, bruno-shop, bruno-shop-(imported)"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "environments" array length equal to 3
