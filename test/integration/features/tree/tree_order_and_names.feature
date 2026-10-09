#language: en
#utf-8

@all @xen_001 @tree @tree_order_and_names
Feature: Sidebar order and file names
  The order of items is stored in the parent's file and restored exactly; hand-added files show
  up instead of vanishing. File names are derived from item names so they are valid on macOS,
  Windows and Linux and unique among siblings even on case-insensitive file systems, while the
  display name stays in the file.

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
      order: []
      """

  @order
  Scenario: Items keep the order they were created in, not name order
    When I call "TreeService.CreateRequest" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{billing}}", "name": "Charlie" }
      """
    And I call "TreeService.CreateRequest" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{billing}}", "name": "Alpha" }
      """
    And I call "TreeService.CreateFolder" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{billing}}", "name": "Bravo" }
      """
    Then the workspace file "acme/collections/billing/collection.yaml" should contain the field "order" with the items "charlie.yaml, alpha.yaml, bravo"
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "0.children.0.name" equal to "Charlie"
    And the response should contain the field "0.children.1.name" equal to "Alpha"
    And the response should contain the field "0.children.2.name" equal to "Bravo"

  @order
  Scenario: Unlisted files are appended by name and listed files that are gone are skipped
    Given the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      order:
        - b.yaml
        - gone.yaml
        - a.yaml
      """
    And the workspace file "acme/collections/billing/a.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB000000000000000000000A
      name: A
      """
    And the workspace file "acme/collections/billing/b.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB000000000000000000000B
      name: B
      """
    And the workspace file "acme/collections/billing/d.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB000000000000000000000D
      name: D
      """
    And the workspace file "acme/collections/billing/c.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB000000000000000000000C
      name: C
      """
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "0.children" array length equal to 4
    And the response should contain the field "0.children.0.name" equal to "B"
    And the response should contain the field "0.children.1.name" equal to "A"
    And the response should contain the field "0.children.2.name" equal to "C"
    And the response should contain the field "0.children.3.name" equal to "D"
    When I call "TreeService.CreateRequest" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{billing}}", "name": "E" }
      """
    Then the call should succeed
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "order" with the items "b.yaml, a.yaml, c.yaml, d.yaml, e.yaml"

  @order
  Scenario: Unlisted collections are appended to the workspace by name
    Given the workspace file "acme/collections/archive/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JB000000000000000000000E
      name: Archive
      """
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.name" equal to "Billing"
    And the response should contain the field "1.name" equal to "Archive"

  @names
  Scenario: Siblings with the same name each get their own file
    When I call "TreeService.CreateRequest" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{billing}}", "name": "Get user" }
      """
    And I call "TreeService.CreateRequest" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{billing}}", "name": "Get user" }
      """
    And I call "TreeService.CreateRequest" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{billing}}", "name": "Get user" }
      """
    Then the workspace file "acme/collections/billing/get-user.yaml" should contain the field "name" equal to "Get user"
    And the workspace file "acme/collections/billing/get-user-2.yaml" should contain the field "name" equal to "Get user"
    And the workspace file "acme/collections/billing/get-user-3.yaml" should contain the field "name" equal to "Get user"

  @names
  Scenario Outline: Names that are not valid file names get a safe file name - <name>
    When I call "TreeService.<method>" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{billing}}", "name": "<name>" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "<name>"
    And the workspace folder "acme/collections/billing" should contain the entries "collection.yaml, <entry>"

    Examples:
      | method        | name              | entry               |
      | CreateRequest | GET /users/:id?   | get-users-id.yaml   |
      | CreateRequest | a*b<c>d\|e        | a-b-c-d-e.yaml      |
      | CreateRequest | CON               | con_.yaml           |
      | CreateRequest | lpt1.backup       | lpt1_.backup.yaml   |
      | CreateFolder  | aux               | aux_                |
      | CreateRequest | collection        | collection-2.yaml   |
      | CreateRequest | ...               | untitled.yaml       |
      | CreateFolder  | Ünïcödé Földer    | ünïcödé-földer      |

  @names
  Scenario: A name that only differs by case from an existing file gets its own file
    Given the workspace file "acme/collections/billing/Users.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB000000000000000000000U
      name: Users (hand made)
      """
    When I call "TreeService.CreateRequest" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{billing}}", "name": "users" }
      """
    Then the call should succeed
    And the workspace folder "acme/collections/billing" should contain the entries "Users.yaml, collection.yaml, users-2.yaml"

  @names
  Scenario: A rename that only changes letter case renames the file through a temporary name
    Given the workspace file "acme/collections/billing/Users.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB000000000000000000000U
      name: Users
      """
    When I call "TreeService.Rename" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JB000000000000000000000U", "name": "USERS" }
      """
    Then the call should succeed
    And the response should contain the field "name" equal to "USERS"
    And the workspace folder "acme/collections/billing" should contain the entries "collection.yaml, users.yaml"
    And the workspace file "acme/collections/billing/users.yaml" should contain the field "name" equal to "USERS"
    And the workspace file "acme/collections/billing/collection.yaml" should contain the field "order" with the items "users.yaml"

  @names
  Scenario: A rename to a name that derives the same file name keeps the file
    When I call "TreeService.CreateRequest" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{billing}}", "name": "Get user" }
      """
    And I save the response field "id" as "user"
    And I call "TreeService.Rename" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{user}}", "name": "GET  user" }
      """
    Then the call should succeed
    And the workspace folder "acme/collections/billing" should contain the entries "collection.yaml, get-user.yaml"
    And the workspace file "acme/collections/billing/get-user.yaml" should contain the field "name" equal to "GET  user"
