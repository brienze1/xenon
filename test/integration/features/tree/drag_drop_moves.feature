#language: en
#utf-8

@all @xen_005 @tree @drag_drop_moves
Feature: Move sidebar items by drag and drop
  Dropping an item in the sidebar moves it: a request or folder dropped on another parent moves
  its file or directory there (with a unique file name when a sibling already uses its name, the
  display name unchanged), and the order of both parents is rewritten; dropping it between its
  siblings only reorders. Collections only reorder at the top level. A folder cannot be dropped
  into itself or its own descendants. Items keep their ids, so open tabs keep pointing at them,
  and Xenon's own moves are never reported as outside changes.

  Background:
    Given the workspace "Globex" exists
    And the "billing" variable is set to "01JM0000000000000000000001"
    And the "invoices" variable is set to "01JM0000000000000000000002"
    And the "drafts" variable is set to "01JM0000000000000000000003"
    And the "create" variable is set to "01JM0000000000000000000004"
    And the "ping" variable is set to "01JM0000000000000000000005"
    And the "status" variable is set to "01JM0000000000000000000006"
    And the "users" variable is set to "01JM0000000000000000000007"
    And the "list" variable is set to "01JM0000000000000000000008"
    And the "archive" variable is set to "01JM0000000000000000000009"
    And the workspace file "globex/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JM0000000000000000000000
      name: Globex
      order:
        - billing
        - users
        - archive
      """
    And the workspace file "globex/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      order:
        - invoices
        - ping.yaml
        - status.yaml
      """
    And the workspace file "globex/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{invoices}}
      name: Invoices
      order:
        - drafts
        - create.yaml
      """
    And the workspace file "globex/collections/billing/invoices/drafts/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{drafts}}
      name: Drafts
      order: []
      """
    And the workspace file "globex/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: https://api.example.test/invoices
      """
    And the workspace file "globex/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{ping}}
      name: Ping
      method: GET
      url: https://api.example.test/ping
      """
    And the workspace file "globex/collections/billing/status.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{status}}
      name: Status
      method: GET
      url: https://api.example.test/status
      """
    And the workspace file "globex/collections/users/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{users}}
      name: Users
      order:
        - ping.yaml
        - list.yaml
      """
    And the workspace file "globex/collections/users/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JM000000000000000000000A
      name: Ping
      method: GET
      url: https://api.example.test/users/ping
      """
    And the workspace file "globex/collections/users/list.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{list}}
      name: List
      method: GET
      url: https://api.example.test/users
      """
    And the workspace file "globex/collections/archive/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{archive}}
      name: Archive
      order: []
      """

  @request @across
  Scenario: A request dropped on a folder or on another collection moves its file there
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{status}}", "new_parent_id": "{{invoices}}", "index": 1 }
      """
    Then the call should succeed
    And the response should contain the field "0.children.1.name" equal to "Ping"
    And the response should contain the field "0.children.0.children.1.id" equal to "{{status}}"
    And the workspace file "globex/collections/billing/status.yaml" should not exist
    And the workspace file "globex/collections/billing/invoices/status.yaml" should contain the field "id" equal to "{{status}}"
    And the workspace file "globex/collections/billing/invoices/status.yaml" should contain the field "url" equal to "https://api.example.test/status"
    And the workspace file "globex/collections/billing/collection.yaml" should contain the field "order" with the items "invoices, ping.yaml"
    And the workspace file "globex/collections/billing/invoices/folder.yaml" should contain the field "order" with the items "drafts, status.yaml, create.yaml"
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{status}}", "new_parent_id": "{{archive}}", "index": 0 }
      """
    Then the call should succeed
    And the response should contain the field "2.children.0.id" equal to "{{status}}"
    And the workspace file "globex/collections/archive/status.yaml" should contain the field "name" equal to "Status"
    And the workspace file "globex/collections/billing/invoices/folder.yaml" should contain the field "order" with the items "drafts, create.yaml"
    And the workspace file "globex/collections/archive/collection.yaml" should contain the field "order" with the items "status.yaml"

  @request @tabs
  Scenario: A moved request keeps its id, so its open tab still loads it
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{create}}", "new_parent_id": "{{users}}", "index": 5 }
      """
    Then the call should succeed
    And the workspace file "globex/collections/users/collection.yaml" should contain the field "order" with the items "ping.yaml, list.yaml, create.yaml"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{create}}" }
      """
    Then the call should succeed
    And the response should contain the field "id" equal to "{{create}}"
    And the response should contain the field "name" equal to "Create"
    And the response should contain the field "method" equal to "POST"

  @request @names
  Scenario: A request whose file name is taken in the new parent gets a unique file name and keeps its name
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{ping}}", "new_parent_id": "{{users}}", "index": 0 }
      """
    Then the call should succeed
    And the response should contain the field "1.children.0.id" equal to "{{ping}}"
    And the response should contain the field "1.children.0.name" equal to "Ping"
    And the workspace file "globex/collections/users/ping-2.yaml" should contain the field "id" equal to "{{ping}}"
    And the workspace file "globex/collections/users/ping-2.yaml" should contain the field "name" equal to "Ping"
    And the workspace file "globex/collections/users/ping.yaml" should contain the field "id" equal to "01JM000000000000000000000A"
    And the workspace file "globex/collections/users/collection.yaml" should contain the field "order" with the items "ping-2.yaml, ping.yaml, list.yaml"
    And the workspace file "globex/collections/billing/collection.yaml" should contain the field "order" with the items "invoices, status.yaml"

  @request @reorder
  Scenario Outline: A request dropped between its siblings only reorders - index <index>
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{status}}", "new_parent_id": "{{billing}}", "index": <index> }
      """
    Then the call should succeed
    And the workspace file "globex/collections/billing/status.yaml" should contain the field "id" equal to "{{status}}"
    And the workspace file "globex/collections/billing/collection.yaml" should contain the field "order" with the items "<order>"

    Examples:
      | index | order                          |
      | 0     | status.yaml, invoices, ping.yaml |
      | 1     | invoices, status.yaml, ping.yaml |
      | 2     | invoices, ping.yaml, status.yaml |
      | 9     | invoices, ping.yaml, status.yaml |

  @folder @reorder
  Scenario: A folder dropped between its siblings reorders
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{invoices}}", "new_parent_id": "{{billing}}", "index": 2 }
      """
    Then the call should succeed
    And the workspace file "globex/collections/billing/collection.yaml" should contain the field "order" with the items "ping.yaml, status.yaml, invoices"
    And the response should contain the field "0.children.2.id" equal to "{{invoices}}"
    And the response should contain the field "0.children.2.children.1.id" equal to "{{create}}"

  @folder @nest
  Scenario: A folder dropped on another collection moves its directory with everything below it
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{invoices}}", "new_parent_id": "{{users}}", "index": 1 }
      """
    Then the call should succeed
    And the workspace folder "globex/collections/users/invoices" should exist
    And the workspace file "globex/collections/billing/invoices/folder.yaml" should not exist
    And the workspace file "globex/collections/users/invoices/folder.yaml" should contain the field "order" with the items "drafts, create.yaml"
    And the workspace file "globex/collections/users/invoices/create.yaml" should contain the field "id" equal to "{{create}}"
    And the workspace file "globex/collections/users/invoices/drafts/folder.yaml" should contain the field "id" equal to "{{drafts}}"
    And the workspace file "globex/collections/users/collection.yaml" should contain the field "order" with the items "ping.yaml, invoices, list.yaml"
    And the workspace file "globex/collections/billing/collection.yaml" should contain the field "order" with the items "ping.yaml, status.yaml"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{create}}" }
      """
    Then the response should contain the field "url" equal to "https://api.example.test/invoices"

  @folder @unnest
  Scenario: A nested folder dropped on its collection is un-nested
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{drafts}}", "new_parent_id": "{{billing}}", "index": 3 }
      """
    Then the call should succeed
    And the workspace file "globex/collections/billing/drafts/folder.yaml" should contain the field "name" equal to "Drafts"
    And the workspace file "globex/collections/billing/invoices/folder.yaml" should contain the field "order" with the items "create.yaml"
    And the workspace file "globex/collections/billing/collection.yaml" should contain the field "order" with the items "invoices, ping.yaml, status.yaml, drafts"

  @collection @reorder
  Scenario Outline: Collections reorder at the top level - index <index>
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{archive}}", "new_parent_id": "", "index": <index> }
      """
    Then the call should succeed
    And the workspace file "globex/workspace.yaml" should contain the field "order" with the items "<order>"
    And the response should contain the field "<position>.id" equal to "{{archive}}"
    And the workspace file "globex/collections/archive/collection.yaml" should contain the field "id" equal to "{{archive}}"

    Examples:
      | index | order                   | position |
      | 0     | archive, billing, users | 0        |
      | 1     | billing, archive, users | 1        |

  @refused
  Scenario Outline: A folder cannot be dropped into itself or its descendants - <case>
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{invoices}}", "new_parent_id": "<parent>", "index": 0 }
      """
    Then the call should fail with code "TRE-02422"
    And the workspace file "globex/collections/billing/invoices/folder.yaml" should contain the field "order" with the items "drafts, create.yaml"
    And the workspace file "globex/collections/billing/collection.yaml" should contain the field "order" with the items "invoices, ping.yaml, status.yaml"

    Examples:
      | case       | parent       |
      | itself     | {{invoices}} |
      | descendant | {{drafts}}   |

  @refused
  Scenario Outline: Items only move to a place that can hold them - <case>
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "<id>", "new_parent_id": "<parent>", "index": 0 }
      """
    Then the call should fail with code "<code>"

    Examples:
      | case                         | id                         | parent                     | code      |
      | request at the top level     | {{ping}}                   |                            | TRE-38422 |
      | folder at the top level      | {{invoices}}               |                            | TRE-38422 |
      | collection inside a folder   | {{archive}}                | {{invoices}}               | TRE-39422 |
      | parent that is a request     | {{status}}                 | {{ping}}                   | TRE-01404 |
      | parent that does not exist   | {{status}}                 | 01JM0000000000000000000099 | TRE-01404 |
      | item that does not exist     | 01JM0000000000000000000099 | {{billing}}                | TRE-02404 |

  @own_writes
  Scenario: Xenon's own moves are not reported as outside changes
    Given I call "WorkspaceService.Activate" with
      """
      { "local_id": "{{Globex.local_id}}" }
      """
    When I call "TreeService.Move" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{invoices}}", "new_parent_id": "{{users}}", "index": 0 }
      """
    Then the call should succeed
    When the file watcher reports the removal of the workspace path "globex/collections/billing/invoices"
    And the file watcher reports a change to the workspace path "globex/collections/users/invoices"
    And the file watcher reports a change to the workspace path "globex/collections/users/invoices/create.yaml"
    And the file watcher reports a change to the workspace path "globex/collections/billing/collection.yaml"
    And the file watcher reports a change to the workspace path "globex/collections/users/collection.yaml"
    And I call "TreeService.Move" with
      """
      { "local_id": "{{Globex.local_id}}", "id": "{{ping}}", "new_parent_id": "{{users}}", "index": 0 }
      """
    And the file watcher reports the removal of the workspace path "globex/collections/billing/ping.yaml"
    And the file watcher reports a change to the workspace path "globex/collections/users/ping-2.yaml"
    Then the event "workspace:tree-changed" should have been emitted 0 times
    And the event "request:changed-on-disk" should have been emitted 0 times

  @not_found
  Scenario: Moves on an unknown workspace fail
    When I call "TreeService.Move" with
      """
      { "local_id": "01JM0000000000000000000098", "id": "{{ping}}", "new_parent_id": "{{users}}", "index": 0 }
      """
    Then the call should fail with code "WKS-07404"

  @validation
  Scenario Outline: Invalid input is rejected - <case>
    When I call "TreeService.Move" with
      """
      <payload>
      """
    Then the call should fail with code "TRE-14400"

    Examples:
      | case             | payload                                                                                     |
      | no workspace     | { "local_id": "", "id": "{{ping}}", "new_parent_id": "{{users}}", "index": 0 }              |
      | no item          | { "local_id": "{{Globex.local_id}}", "id": " ", "new_parent_id": "{{users}}", "index": 0 }   |
      | negative index   | { "local_id": "{{Globex.local_id}}", "id": "{{ping}}", "new_parent_id": "{{users}}", "index": -1 } |
