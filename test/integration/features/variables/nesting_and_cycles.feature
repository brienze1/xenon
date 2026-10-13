#language: en
#utf-8

@all @xen_002 @variables @nesting_and_cycles
Feature: Nested references resolve up to Postman's limit and cycles stay literal
  A value may itself contain "{{other}}". Go replaces references in passes over the text while
  replacements keep happening, at most 19 passes (Postman's VARS_SUBREPLACE_LIMIT). Whatever
  "{{...}}" text remains after that - a cycle, a chain deeper than the limit, or a name nobody
  defines - is sent literally.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "dev" variable is set to "01JN0000000000000000000010"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JN0000000000000000000000
      name: Acme
      variables:
        - key: last
          value: Brienze
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
        - key: greeting
          value: "hello {{name}}"
          enabled: true
          description: ""
        - key: name
          value: "{{first}} {{last}}"
          enabled: true
          description: ""
        - key: first
          value: Luis
          enabled: true
          description: ""
        - key: a
          value: "{{b}}"
          enabled: true
          description: ""
        - key: b
          value: "{{a}}"
          enabled: true
          description: ""
        - key: self
          value: "<{{self}}>"
          enabled: true
          description: ""
        - key: dangling
          value: "x-{{nobody}}"
          enabled: true
          description: ""
        - key: v1
          value: "{{v2}}"
          enabled: true
          description: ""
        - key: v2
          value: "{{v3}}"
          enabled: true
          description: ""
        - key: v3
          value: "{{v4}}"
          enabled: true
          description: ""
        - key: v4
          value: "{{v5}}"
          enabled: true
          description: ""
        - key: v5
          value: "{{v6}}"
          enabled: true
          description: ""
        - key: v6
          value: "{{v7}}"
          enabled: true
          description: ""
        - key: v7
          value: "{{v8}}"
          enabled: true
          description: ""
        - key: v8
          value: "{{v9}}"
          enabled: true
          description: ""
        - key: v9
          value: "{{v10}}"
          enabled: true
          description: ""
        - key: v10
          value: "{{v11}}"
          enabled: true
          description: ""
        - key: v11
          value: "{{v12}}"
          enabled: true
          description: ""
        - key: v12
          value: "{{v13}}"
          enabled: true
          description: ""
        - key: v13
          value: "{{v14}}"
          enabled: true
          description: ""
        - key: v14
          value: "{{v15}}"
          enabled: true
          description: ""
        - key: v15
          value: "{{v16}}"
          enabled: true
          description: ""
        - key: v16
          value: "{{v17}}"
          enabled: true
          description: ""
        - key: v17
          value: "{{v18}}"
          enabled: true
          description: ""
        - key: v18
          value: "{{v19}}"
          enabled: true
          description: ""
        - key: v19
          value: "{{v20}}"
          enabled: true
          description: ""
        - key: v20
          value: end
          enabled: true
          description: ""
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """

  @nested
  Scenario: A value referencing other variables, across scopes, resolves recursively
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/nested", "headers": [ { "key": "X-Greeting", "value": "{{greeting}}!" } ] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/nested" should have the header "X-Greeting" equal to "hello Luis Brienze!"

  @limit
  Scenario: A chain resolving within 19 passes is complete, one level deeper stays literal
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/deep",
          "headers": [ { "key": "X-Nineteen", "value": "{{v2}}" }, { "key": "X-Twenty", "value": "{{v1}}" } ]
        }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/deep" should have the header "X-Nineteen" equal to "end"
    And the 1 "GET" request to "/deep" should have the header "X-Twenty" equal to "{{v20}}"

  @cycles
  Scenario: A cycle stops at the pass limit and the remaining reference is sent literally
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/cycle",
          "headers": [
            { "key": "X-Cycle", "value": "{{a}}" },
            { "key": "X-Self", "value": "{{self}}" },
            { "key": "X-Dangling", "value": "{{dangling}}" }
          ]
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "200"
    And the 1 "GET" request to "/cycle" should have the header "X-Cycle" equal to "{{b}}"
    And the 1 "GET" request to "/cycle" should have the header "X-Self" equal to "<<<<<<<<<<<<<<<<<<<{{self}}>>>>>>>>>>>>>>>>>>>"
    And the 1 "GET" request to "/cycle" should have the header "X-Dangling" equal to "x-{{nobody}}"
