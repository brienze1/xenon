#language: en
#utf-8

@all @xen_005 @tree @quick_search
Feature: Quick search over the workspace's requests
  Cmd/Ctrl+K searches the current workspace's requests by name, URL and path (the collection and
  folders above them), ignoring case. Results are ranked: names starting with the text first,
  then names containing it, then URLs containing it, then paths containing it; within a rank the
  sidebar order is kept. At most "limit" results come back (50 when no limit is given), and a
  blank query returns nothing.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JQ0000000000000000000000
      name: Acme
      order:
        - billing
        - users
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JQ0000000000000000000001
      name: Billing
      order:
        - invoices
        - ping.yaml
        - order.yaml
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: 01JQ0000000000000000000002
      name: Invoices
      order:
        - create.yaml
        - pay.yaml
      """
    And the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JQ0000000000000000000003
      name: Create invoice
      method: POST
      url: https://api.example.test/v1/invoices
      """
    And the workspace file "acme/collections/billing/invoices/pay.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JQ0000000000000000000004
      name: Pay
      method: POST
      url: https://api.example.test/v1/payments
      """
    And the workspace file "acme/collections/billing/ping.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JQ0000000000000000000005
      name: Ping
      method: GET
      url: https://api.example.test/ping
      """
    And the workspace file "acme/collections/billing/order.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JQ0000000000000000000006
      name: Invoice totals
      method: GET
      url: https://api.example.test/v1/totals
      """
    And the workspace file "acme/collections/users/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JQ0000000000000000000007
      name: Users
      order:
        - list.yaml
      """
    And the workspace file "acme/collections/users/list.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JQ0000000000000000000008
      name: List users
      method: GET
      url: https://api.example.test/v1/users?with=invoices
      """
    And the workspace file "globex/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JQ0000000000000000000010
      name: Globex
      order:
        - billing
      """
    And the workspace file "globex/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JQ0000000000000000000011
      name: Billing
      """
    And the workspace file "globex/collections/billing/invoice.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JQ0000000000000000000012
      name: Invoice
      method: GET
      url: https://globex.example.test/invoice
      """

  @ranking
  Scenario: Name prefix comes before name contains, then URL, then path
    When I call "TreeService.QuickSearch" with
      """
      { "local_id": "{{Acme.local_id}}", "query": "INVOICE", "limit": 10 }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 4
    And the response should contain the field "0.id" equal to "01JQ0000000000000000000006"
    And the response should contain the field "1.id" equal to "01JQ0000000000000000000003"
    And the response should contain the field "2.id" equal to "01JQ0000000000000000000008"
    And the response should contain the field "3.id" equal to "01JQ0000000000000000000004"

  @fields
  Scenario: Each result carries the request's name, method, URL and path
    When I call "TreeService.QuickSearch" with
      """
      { "local_id": "{{Acme.local_id}}", "query": "create", "limit": 10 }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 1
    And the response should contain the field "0.id" equal to "01JQ0000000000000000000003"
    And the response should contain the field "0.name" equal to "Create invoice"
    And the response should contain the field "0.method" equal to "POST"
    And the response should contain the field "0.url" equal to "https://api.example.test/v1/invoices"
    And the response should contain the field "0.path" equal to "Billing / Invoices"

  @ranking @order
  Scenario: Within a rank the sidebar order is kept
    When I call "TreeService.QuickSearch" with
      """
      { "local_id": "{{Acme.local_id}}", "query": "/v1/", "limit": 10 }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 4
    And the response should contain the field "0.id" equal to "01JQ0000000000000000000003"
    And the response should contain the field "1.id" equal to "01JQ0000000000000000000004"
    And the response should contain the field "2.id" equal to "01JQ0000000000000000000006"
    And the response should contain the field "3.id" equal to "01JQ0000000000000000000008"

  @path
  Scenario: A request matches by the names of the collection and folders above it
    When I call "TreeService.QuickSearch" with
      """
      { "local_id": "{{Acme.local_id}}", "query": "billing / invoices", "limit": 10 }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 2
    And the response should contain the field "0.id" equal to "01JQ0000000000000000000003"
    And the response should contain the field "1.id" equal to "01JQ0000000000000000000004"
    And the response should contain the field "1.path" equal to "Billing / Invoices"

  @limit
  Scenario Outline: The limit caps the results - limit <limit>
    When I call "TreeService.QuickSearch" with
      """
      { "local_id": "{{Acme.local_id}}", "query": "e", "limit": <limit> }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to <count>

    Examples:
      | limit | count |
      | 1     | 1     |
      | 3     | 3     |
      | 0     | 5     |

  @isolation
  Scenario: Only the current workspace is searched
    When I call "TreeService.QuickSearch" with
      """
      { "local_id": "{{Globex.local_id}}", "query": "invoice", "limit": 10 }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 1
    And the response should contain the field "0.id" equal to "01JQ0000000000000000000012"
    And the response should contain the field "0.path" equal to "Billing"

  @empty
  Scenario Outline: A blank query or one that matches nothing returns no results - "<query>"
    When I call "TreeService.QuickSearch" with
      """
      { "local_id": "{{Acme.local_id}}", "query": "<query>", "limit": 10 }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 0

    Examples:
      | query              |
      |                    |
      | nothing like this  |

  @validation
  Scenario Outline: An invalid quick search is refused - <case>
    When I call "TreeService.QuickSearch" with
      """
      <payload>
      """
    Then the call should fail with code "<code>"

    Examples:
      | case               | payload                                                      | code      |
      | missing local id   | { "query": "ping", "limit": 10 }                             | TRE-40400 |
      | negative limit     | { "local_id": "{{Acme.local_id}}", "query": "ping", "limit": -1 } | TRE-40400 |
      | unknown workspace  | { "local_id": "01JQ0000000000000000000098", "query": "ping", "limit": 10 } | WKS-07404 |
