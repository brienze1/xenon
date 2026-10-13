#language: en
#utf-8

@all @xen_002 @variables @performance @inspect_performance
Feature: Inspecting and resolving many variables stays fast
  A request with 200 variable occurrences (resolved, nested, secret, dynamic and unresolved) is
  inspected in under 100 ms per keystroke, so highlighting and the Variables panel keep up with
  typing, and sending it stays fast (SC-004; the resolver itself is held under 10 ms by a unit
  test of ResolveRequestVariables).

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the "api" variable is set to the mocked api url
    And the "dev" variable is set to "01JP0000000000000000000010"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JP0000000000000000000000
      name: Acme
      variables:
        - key: region
          value: eu-west-1
          enabled: true
          description: ""
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: base_url
          value: "{{api}}"
          enabled: true
          description: ""
        - key: nested
          value: "{{region}}-{{base_url}}"
          enabled: true
          description: ""
      """
    And I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{dev}}", "key": "token", "value": "perf-s3cr3t", "secret": true }
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """
    And the "payload" variable is set to "{{nested}}/{{token}}/{{$guid}}/{{missing}}/{{region}}" repeated 40 times separated by ","

  @inspect
  Scenario: Every keystroke of a request with 200 occurrences is inspected in under 100 ms
    Given I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "POST", "url": "{{base_url}}/x", "body": { "mode": "raw", "raw": { "language": "text", "content": "{{payload}}" } } } }
      """
    When I call "VariableService.Inspect" with
      """
      { "local_id": "{{Acme.local_id}}", "request": { "method": "POST", "url": "{{base_url}}/x", "body": { "mode": "raw", "raw": { "language": "text", "content": "{{payload}}" } } } }
      """
    Then the call should succeed
    And the last call should have taken less than 100 ms
    And the response should contain the field "references" array length equal to 201
    And the response should contain the field "variables" array length equal to 6
    And the response should not contain "perf-s3cr3t"

  @send
  Scenario: Sending a request with 200 occurrences stays fast
    Given I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{base_url}}/warm" } }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{base_url}}/x", "body": { "mode": "raw", "raw": { "language": "text", "content": "{{payload}}" } } } }
      """
    Then the call should succeed
    And the last call should have taken less than 100 ms
    And the "POST" request to "/x" should have been received 1 times
    And the 1 "POST" request to "/x" should have the body split by "," into 40 parts all matching "^eu-west-1-http://127\.0\.0\.1:\d+/perf-s3cr3t/[0-9a-f-]{36}/\{\{missing\}\}/eu-west-1$"
