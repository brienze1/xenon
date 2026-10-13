#language: en
#utf-8

@all @xen_002 @variables @exact_names
Feature: Variable names match exactly, as in Postman
  A reference is "{{" + name + "}}" where the name is any text without "{" or "}" (Postman's
  pattern {{([^{}]*?)}}). The name is used exactly as written: it is case-sensitive and spaces
  are part of it, so "{{ name }}" looks up " name ". A user variable whose name starts with "$"
  is an ordinary variable.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "dev" variable is set to "01JE0000000000000000000010"
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: name
          value: plain
          enabled: true
          description: ""
        - key: " spaced "
          value: with-spaces
          enabled: true
          description: ""
        - key: Token
          value: upper
          enabled: true
          description: ""
        - key: my-var.v2:x
          value: punctuated
          enabled: true
          description: ""
        - key: $custom
          value: user-defined
          enabled: true
          description: ""
        - key: dup
          value: first
          enabled: true
          description: ""
        - key: dup
          value: last
          enabled: true
          description: ""
        - key: dup_disabled
          value: kept
          enabled: true
          description: ""
        - key: dup_disabled
          value: ignored
          enabled: false
          description: ""
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """

  @exact
  Scenario Outline: "<reference>" is sent as "<sent>"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/names", "headers": [ { "key": "X-Value", "value": "<reference>" } ] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/names" should have the header "X-Value" equal to "<sent>"

    Examples:
      | reference          | sent               |
      | {{name}}           | plain              |
      | {{ name }}         | {{ name }}         |
      | {{ spaced }}       | with-spaces        |
      | {{spaced}}         | {{spaced}}         |
      | {{Token}}          | upper              |
      | {{token}}          | {{token}}          |
      | {{TOKEN}}          | {{TOKEN}}          |
      | {{my-var.v2:x}}    | punctuated         |
      | {{$custom}}        | user-defined       |
      | {{}}               | {{}}               |
      | {{{name}}}         | {plain}            |
      | {{name}}{{Token}}  | plainupper         |
      | {name}             | {name}             |
      | {{dup}}            | last               |
      | {{dup_disabled}}   | kept               |
