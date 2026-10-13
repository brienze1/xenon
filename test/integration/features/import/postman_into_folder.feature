#language: en
#utf-8

@all @xen_003 @import @postman_into_folder
Feature: Import a Postman collection into an existing collection or folder
  Importing into an existing folder makes the collection's top-level folders and requests children
  of that folder. When the target folder has no auth, headers or scripts of its own, the imported
  collection's auth and scripts are applied to it; otherwise the imported items are placed in a new
  sub-folder named after the collection, which carries them. Collection variables are added to the
  containing collection, keeping existing values on a "Keep both" clash. The summary says which
  option was applied.

  Background:
    Given the workspace "Acme" exists
    And I call "TreeService.CreateCollection" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "Main" }
      """
    And I save the response field "id" as "main"
    And I call "TreeService.CreateFolder" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{main}}", "name": "Target" }
      """
    And I save the response field "id" as "target"
    And I call "TreeService.CreateFolder" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{main}}", "name": "Owned" }
      """
    And I save the response field "id" as "owned"
    And I call "TreeService.SaveSettings" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{owned}}", "description": "", "auth": { "type": "basic", "basic": { "username": "own", "password": "pw" } }, "headers": [] }
      """
    And the workspace file "acme/collections/main/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{main}}
      name: Main
      description: ""
      order: [target, owned]
      auth:
        type: inherit
      headers: []
      variables:
        - key: base_url
          value: https://main.acme.test
          enabled: true
          description: ""
      """
    And the file "small.postman_collection.json" exists with content
      """
      {
        "info": { "_postman_id": "small-1", "name": "Small API", "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json" },
        "auth": { "type": "bearer", "bearer": [ { "key": "token", "value": "{{small_token}}", "type": "string" } ] },
        "event": [ { "listen": "prerequest", "script": { "type": "text/javascript", "exec": [ "pm.variables.set(\"small\", 1);" ] } } ],
        "variable": [ { "key": "base_url", "value": "https://small.acme.test" }, { "key": "page", "value": "2" } ],
        "item": [
          { "name": "Things", "item": [ { "name": "List things", "request": { "method": "GET", "url": "{{base_url}}/things" } } ] },
          { "name": "Health", "request": { "method": "GET", "url": "{{base_url}}/health" } }
        ]
      }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/small.postman_collection.json"] }
      """
    And I save the response field "source_id" as "source"

  @merge
  Scenario: Importing into a folder without settings of its own applies the collection's settings to it
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{target}}", "resolution": "keep_both" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the response should contain the field "items.0.id" equal to "{{target}}"
    And the response field "warnings" should contain an item with "code" equal to "merged_into_folder" and "path" equal to "Main/Target"
    And the response field "warnings" should contain an item with "code" equal to "variable_kept" and "message" equal to "base_url keeps its value; the imported value was https://small.acme.test"
    And the workspace file "acme/collections/main/target/folder.yaml" should contain the field "order" with the items "things, health.yaml"
    And the workspace file "acme/collections/main/target/folder.yaml" should contain the field "auth.type" equal to "bearer"
    And the workspace file "acme/collections/main/target/folder.yaml" should contain the field "auth.bearer.token" equal to "{{small_token}}"
    And the workspace file "acme/collections/main/target/folder.yaml" should contain the field "scripts.preRequest" equal to
      """
      pm.variables.set("small", 1);
      """
    And the workspace file "acme/collections/main/target/things/list-things.yaml" should contain the field "url" equal to "{{base_url}}/things"
    And the workspace file "acme/collections/main/target/health.yaml" should contain the field "method" equal to "GET"
    And the workspace file "acme/collections/main/collection.yaml" should contain the field "variables" array length equal to 2
    And the workspace file "acme/collections/main/collection.yaml" should contain the field "variables.0.value" equal to "https://main.acme.test"
    And the workspace file "acme/collections/main/collection.yaml" should contain the field "variables.1.key" equal to "page"
    And the workspace file "acme/collections/main/collection.yaml" should contain the field "variables.1.value" equal to "2"
    And the workspace file "acme/collections/main/collection.yaml" should contain the field "order" with the items "target, owned"
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.children.0.children.0.name" equal to "Things"
    And the response should contain the field "0.children.0.children.1.name" equal to "Health"

  @merge @replace
  Scenario: Replacing on a variable clash overwrites the existing value
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{target}}", "resolution": "replace" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/main/collection.yaml" should contain the field "variables.0.value" equal to "https://small.acme.test"

  @wrap
  Scenario: Importing into a folder with its own auth places the items in a sub-folder named after the collection
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{owned}}", "resolution": "keep_both" } ] }
      """
    Then the call should succeed
    And the response field "warnings" should contain an item with "code" equal to "wrapped_in_folder" and "path" equal to "Main/Owned/Small API"
    And the workspace file "acme/collections/main/owned/folder.yaml" should contain the field "order" with the items "small-api"
    And the workspace file "acme/collections/main/owned/folder.yaml" should contain the field "auth.type" equal to "basic"
    And the workspace file "acme/collections/main/owned/small-api/folder.yaml" should contain the field "name" equal to "Small API"
    And the workspace file "acme/collections/main/owned/small-api/folder.yaml" should contain the field "kind" equal to "folder"
    And the workspace file "acme/collections/main/owned/small-api/folder.yaml" should contain the field "auth.type" equal to "bearer"
    And the workspace file "acme/collections/main/owned/small-api/folder.yaml" should contain the field "order" with the items "things, health.yaml"
    And the workspace file "acme/collections/main/owned/small-api/folder.yaml" should contain the field "postman.info._postman_id" equal to "small-1"
    And the workspace file "acme/collections/main/owned/small-api/things/list-things.yaml" should contain the field "method" equal to "GET"

  @collection_target
  Scenario: Importing into a collection without settings of its own merges the items into it
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{main}}", "resolution": "keep_both" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/main/collection.yaml" should contain the field "order" with the items "target, owned, things, health.yaml"
    And the workspace file "acme/collections/main/collection.yaml" should contain the field "auth.type" equal to "bearer"
    And the workspace file "acme/collections/main/things/list-things.yaml" should contain the field "url" equal to "{{base_url}}/things"

  @clash
  Scenario: Requests that clash with children of the target folder are resolved
    Given the workspace file "acme/collections/main/target/health.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB0000000000000000000099
      name: Health
      method: HEAD
      url: https://main.acme.test/health
      """
    And the workspace file "acme/collections/main/target/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{target}}
      name: Target
      description: ""
      order: [health.yaml]
      auth:
        type: inherit
      headers: []
      """
    When I call "ImportService.Clashes" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{target}}" } ] }
      """
    Then the call should succeed
    And the response field "clashes" should contain an item with "kind" equal to "request" and "existing_name" equal to "Health"
    And the response field "clashes" should contain an item with "kind" equal to "variable" and "existing_name" equal to "base_url"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{target}}" } ] }
      """
    Then the call should fail with code "IMP-04409"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{target}}" } ], "apply_to_all": "keep_both" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/main/target/folder.yaml" should contain the field "order" with the items "health.yaml, things, health-(imported).yaml"
    And the workspace file "acme/collections/main/target/health-(imported).yaml" should contain the field "name" equal to "Health (imported)"
    And the workspace file "acme/collections/main/target/health.yaml" should contain the field "method" equal to "HEAD"
