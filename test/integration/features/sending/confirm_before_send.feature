#language: en
#utf-8

@all @xen_002 @sending @confirm_before_send
Feature: Environments that require a confirmation before every send
  An environment can turn on "Confirm before send" (on by default, with the neon red colour, for
  any name containing "prod"). While it is active, Go refuses every Send with SND-05409 and the
  environment's id, name and colour in the error details: nothing goes out and nothing is
  recorded in History. The UI asks, then sends again with confirmed_environment_id set to that
  environment's id. Any other id is refused the same way, so no request reaches a confirmed
  environment without an explicit confirmation, whatever the UI does.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "prod" variable is set to "01JC0000000000000000000010"
    And the "dev" variable is set to "01JC0000000000000000000011"
    And the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{prod}}
      name: prod
      color: red
      confirmBeforeSend: true
      variables:
        - key: api_key
          secret: true
          enabled: true
          description: ""
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      color: green
      confirmBeforeSend: false
      """

  @refused
  Scenario: A send to a confirmed environment is refused with its details and nothing is sent
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/orders" } }
      """
    Then the call should fail with code "SND-05409"
    And the error should contain the field "error.error_details.0.attribute" equal to "environment_id"
    And the error should contain the field "error.error_details.0.messages.0" equal to "{{prod}}"
    And the error should contain the field "error.error_details.1.attribute" equal to "name"
    And the error should contain the field "error.error_details.1.messages.0" equal to "prod"
    And the error should contain the field "error.error_details.2.attribute" equal to "color"
    And the error should contain the field "error.error_details.2.messages.0" equal to "red"
    And the "GET" request to "/orders" should have been received 0 times
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0

  @confirmed
  Scenario: Sending again with the environment's id sends the request
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "confirmed_environment_id": "{{prod}}", "request": { "method": "GET", "url": "{{api}}/orders" } }
      """
    Then the call should succeed
    And the "GET" request to "/orders" should have been received 1 times
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 1

  @refused
  Scenario Outline: A confirmation for another environment is still refused (<case>)
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "confirmed_environment_id": "<confirmed>", "request": { "method": "GET", "url": "{{api}}/orders" } }
      """
    Then the call should fail with code "SND-05409"
    And the "GET" request to "/orders" should have been received 0 times

    Examples:
      | case             | confirmed                  |
      | other env        | {{dev}}                    |
      | unknown id       | 01JC0000000000000000000099 |
      | different casing | 01jc0000000000000000000010 |

  @not_required
  Scenario Outline: No confirmation is needed with <active>
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "<id>" }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/orders" } }
      """
    Then the call should succeed
    And the "GET" request to "/orders" should have been received 1 times

    Examples:
      | active                       | id      |
      | an environment without it    | {{dev}} |
      | no environment               |         |

  @flag_off
  Scenario: Turning the flag off lets requests go without a confirmation
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    And I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}", "name": "prod", "description": "", "color": "red", "confirm_before_send": false }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/orders" } }
      """
    Then the call should succeed
    And the "GET" request to "/orders" should have been received 1 times

  @secrets
  Scenario: The confirmation comes before the missing secrets check, and both are needed
    Given I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/orders", "headers": [ { "key": "X-Key", "value": "{{api_key}}", "enabled": true } ] } }
      """
    Then the call should fail with code "SND-05409"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "confirmed_environment_id": "{{prod}}", "request": { "method": "GET", "url": "{{api}}/orders", "headers": [ { "key": "X-Key", "value": "{{api_key}}", "enabled": true } ] } }
      """
    Then the call should fail with code "SND-06409"
    And the "GET" request to "/orders" should have been received 0 times
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "confirmed_environment_id": "{{prod}}", "acknowledge_missing_secrets": true, "request": { "method": "GET", "url": "{{api}}/orders", "headers": [ { "key": "X-Key", "value": "{{api_key}}", "enabled": true } ] } }
      """
    Then the call should succeed
    And the 1 "GET" request to "/orders" should have the header "X-Key" equal to "{{api_key}}"

  @defaults
  Scenario Outline: A new environment named "<name>" defaults to colour <color> and confirmation <confirm>
    When I call "EnvironmentService.Create" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "<name>" }
      """
    Then the call should succeed
    And the response should contain the field "color" equal to "<color>"
    And the response should contain the field "confirm_before_send" equal to "<confirm>"
    And I save the response field "id" as "created"
    When I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{created}}" }
      """
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "GET", "url": "{{api}}/orders" } }
      """
    Then the "GET" request to "/orders" should have been received <sent> times

    Examples:
      | name       | color | confirm | sent |
      | Prod EU    | red   | true    | 0    |
      | PRODUCTION | red   | true    | 0    |
      | staging    | none  | false   | 1    |
      | pro        | none  | false   | 1    |

  @colours
  Scenario Outline: The colour can be set to <color>
    When I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}", "name": "dev", "description": "", "color": "<color>", "confirm_before_send": false }
      """
    Then the call should succeed
    And the response should contain the field "color" equal to "<color>"
    When I call "EnvironmentService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response field "environments" should contain an item with "name" equal to "dev" and "color" equal to "<color>"

    Examples:
      | color   |
      | none    |
      | green   |
      | cyan    |
      | magenta |
      | yellow  |
      | red     |
