#language: en
#utf-8

@all @xen_001 @tree @search
Feature: Sidebar search
  Typing in the sidebar search keeps only the requests whose name or URL contains the text
  (ignoring case), together with the folders above them, within the current workspace only.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JB0000000000000000000000
      name: Acme
      order:
        - billing
        - users
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JB0000000000000000000001
      name: Billing
      """
    And the workspace file "acme/collections/billing/invoices/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: 01JB0000000000000000000002
      name: Invoices
      """
    And the workspace file "acme/collections/billing/invoices/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB0000000000000000000003
      name: Create invoice
      method: POST
      url: https://api.example.test/v1/INVOICES
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
    And the workspace file "acme/collections/users/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JB0000000000000000000005
      name: Users
      """
    And the workspace file "acme/collections/users/list.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB0000000000000000000006
      name: List Users
      method: GET
      url: https://api.example.test/people
      """
    And the workspace file "globex/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JB0000000000000000000010
      name: Globex
      order:
        - users
      """
    And the workspace file "globex/collections/users/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JB0000000000000000000011
      name: Users
      """
    And the workspace file "globex/collections/users/list.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JB0000000000000000000012
      name: List users
      method: GET
      url: https://api.example.test/users
      """

  @success
  Scenario: A name match returns the request and every folder above it
    When I call "TreeService.Search" with
      """
      { "local_id": "{{Acme.local_id}}", "query": "INVOICE" }
      """
    Then the call should succeed
    And the response should contain the field "match_ids" with the items "01JB0000000000000000000003"
    And the response should contain the field "ancestor_ids" with the items "01JB0000000000000000000001, 01JB0000000000000000000002"

  @success
  Scenario: A URL match counts and folders or collections never match by themselves
    When I call "TreeService.Search" with
      """
      { "local_id": "{{Acme.local_id}}", "query": "users" }
      """
    Then the call should succeed
    And the response should contain the field "match_ids" with the items "01JB0000000000000000000006"
    And the response should contain the field "ancestor_ids" with the items "01JB0000000000000000000005"
    When I call "TreeService.Search" with
      """
      { "local_id": "{{Acme.local_id}}", "query": "/ping" }
      """
    Then the response should contain the field "match_ids" with the items "01JB0000000000000000000004"
    And the response should contain the field "ancestor_ids" with the items "01JB0000000000000000000001"

  @success
  Scenario: Several matches across collections
    When I call "TreeService.Search" with
      """
      { "local_id": "{{Acme.local_id}}", "query": "api.example.test" }
      """
    Then the response should contain the field "match_ids" with the items "01JB0000000000000000000003, 01JB0000000000000000000004, 01JB0000000000000000000006"
    And the response should contain the field "ancestor_ids" with the items "01JB0000000000000000000001, 01JB0000000000000000000002, 01JB0000000000000000000005"

  @isolation
  Scenario: Search only looks at the current workspace
    When I call "TreeService.Search" with
      """
      { "local_id": "{{Globex.local_id}}", "query": "users" }
      """
    Then the response should contain the field "match_ids" with the items "01JB0000000000000000000012"
    And the response should contain the field "ancestor_ids" with the items "01JB0000000000000000000011"

  @empty
  Scenario: Nothing matches
    When I call "TreeService.Search" with
      """
      { "local_id": "{{Acme.local_id}}", "query": "nothing like this" }
      """
    Then the call should succeed
    And the response should contain the field "match_ids" with the items ""
    And the response should contain the field "ancestor_ids" with the items ""
