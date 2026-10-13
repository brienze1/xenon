#language: en
#utf-8

@all @xen_008 @mcp @mcp_approvals
Feature: AI sends to "Confirm before send" environments wait for the human
  When an AI send would use an environment marked "Confirm before send" and "Dangerously skip approvals" is off, Xenon
  brings its window to the front and asks the human to Approve or Deny. The prompt shows the client, the method, the
  URL as written (never resolved) and the environment. The request is sent once on Approve; on Deny, after 120 seconds
  with no answer, when the requesting client disconnects or when Xenon quits, nothing is sent and the AI is told why.
  Approvals queue in arrival order. With "Dangerously skip approvals" on, such sends go out without a prompt; secrets
  stay write-only either way.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the "api" variable is set to the mocked api url
    And the "prod" variable is set to "01JU0000000000000000000011"
    And the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{prod}}
      name: Prod
      color: red
      confirmBeforeSend: true
      variables:
        - key: base_url
          value: "{{api}}"
          enabled: true
          description: ""
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}" }
      """
    And the 1 "DELETE" request to "/users/7" returns status 204 with the following response
      """
      """
    And the MCP client is named "Claude Code"

  @approve
  Scenario: Approved, the request is sent once and the AI gets the response
    When I start calling the MCP tool "send" in the background with
      """
      { "request": { "method": "DELETE", "url": "{{base_url}}/users/7" } }
      """
    Then an "ai:approval-requested" event with the field "client" equal to "Claude Code" should have been emitted within 2000 ms
    And the last "ai:approval-requested" event should contain the field "method" equal to "DELETE"
    And the last "ai:approval-requested" event should contain the field "url" equal to "{{base_url}}/users/7"
    And the last "ai:approval-requested" event should contain the field "environment.name" equal to "Prod"
    And the last "ai:approval-requested" event should contain the field "environment.color" equal to "red"
    And the Xenon window should have been brought to the front 1 time
    And the "DELETE" request to "/users/7" should have been received 0 times
    When I call "AIAccessService.PendingApprovals"
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.url" equal to "{{base_url}}/users/7"
    And I save the response field "0.id" as "approval"
    When I call "AIAccessService.AnswerApproval" with
      """
      { "id": "{{approval}}", "approved": true }
      """
    Then the call should succeed
    And the background call should finish within 5000 ms
    And the call should succeed
    And the response should contain the field "status" equal to "204"
    And the "DELETE" request to "/users/7" should have been received 1 times
    And the last "ai:approval-resolved" event should contain the field "outcome" equal to "approved"

  @deny
  Scenario: Denied, nothing is sent and the AI is told so
    When I start calling the MCP tool "send" in the background with
      """
      { "request": { "method": "DELETE", "url": "{{base_url}}/users/7" } }
      """
    Then an "ai:approval-requested" event with the field "client" equal to "Claude Code" should have been emitted within 2000 ms
    When I call "AIAccessService.PendingApprovals"
    And I save the response field "0.id" as "approval"
    And I call "AIAccessService.AnswerApproval" with
      """
      { "id": "{{approval}}", "approved": false }
      """
    Then the background call should finish within 5000 ms
    And the call should fail with code "MCP-03403"
    And the error should contain the field "error.message" containing "denied by the user in Xenon"
    And the "DELETE" request to "/users/7" should have been received 0 times
    And the last "ai:approval-resolved" event should contain the field "outcome" equal to "denied"

  @timeout
  Scenario: With no answer for 120 seconds the send is not made
    When I start calling the MCP tool "send" in the background with
      """
      { "request": { "method": "DELETE", "url": "{{base_url}}/users/7" } }
      """
    Then an "ai:approval-requested" event with the field "client" equal to "Claude Code" should have been emitted within 2000 ms
    When 119 seconds pass
    And I call "AIAccessService.PendingApprovals"
    Then the response should contain the field "" array length equal to 1
    When 1 second passes
    Then the background call should finish within 5000 ms
    And the call should fail with code "MCP-04408"
    And the error should contain the field "error.message" containing "approval timed out"
    And the "DELETE" request to "/users/7" should have been received 0 times
    And the last "ai:approval-resolved" event should contain the field "outcome" equal to "timed_out"
    When I call "AIAccessService.PendingApprovals"
    Then the response should contain the field "" array length equal to 0

  @answered_twice
  Scenario: An approval can be answered only once
    When I start calling the MCP tool "send" in the background with
      """
      { "request": { "method": "DELETE", "url": "{{base_url}}/users/7" } }
      """
    Then an "ai:approval-requested" event with the field "client" equal to "Claude Code" should have been emitted within 2000 ms
    When I call "AIAccessService.PendingApprovals"
    And I save the response field "0.id" as "approval"
    And I call "AIAccessService.AnswerApproval" with
      """
      { "id": "{{approval}}", "approved": false }
      """
    And I call "AIAccessService.AnswerApproval" with
      """
      { "id": "{{approval}}", "approved": true }
      """
    Then the call should fail with code "MCP-05409"
    And the background call should finish within 5000 ms
    And the "DELETE" request to "/users/7" should have been received 0 times

  @fifo
  Scenario: Approvals from two clients queue in arrival order and are answered one by one
    Given the 1 "GET" request to "/orders" returns status 200 with the following response
      """
      []
      """
    When I start calling the MCP tool "send" in the background as "first" with
      """
      { "request": { "method": "DELETE", "url": "{{base_url}}/users/7" } }
      """
    Then an "ai:approval-requested" event with the field "client" equal to "Claude Code" should have been emitted within 2000 ms
    Given the MCP client is named "Cursor"
    When I start calling the MCP tool "send" in the background as "second" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/orders" } }
      """
    Then an "ai:approval-requested" event with the field "client" equal to "Cursor" should have been emitted within 2000 ms
    When I call "AIAccessService.PendingApprovals"
    Then the response should contain the field "" array length equal to 2
    And the response should contain the field "0.client" equal to "Claude Code"
    And the response should contain the field "1.client" equal to "Cursor"
    And I save the response field "0.id" as "first_id"
    And I save the response field "1.id" as "second_id"
    When I call "AIAccessService.AnswerApproval" with
      """
      { "id": "{{second_id}}", "approved": true }
      """
    Then the background call "second" should finish within 5000 ms
    And the call should succeed
    And the "GET" request to "/orders" should have been received 1 times
    When I call "AIAccessService.AnswerApproval" with
      """
      { "id": "{{first_id}}", "approved": false }
      """
    Then the background call "first" should finish within 5000 ms
    And the call should fail with code "MCP-03403"

  @disconnect
  Scenario: When the requesting client disconnects, its pending approval is withdrawn
    When I start calling the MCP tool "send" in the background with
      """
      { "request": { "method": "DELETE", "url": "{{base_url}}/users/7" } }
      """
    Then an "ai:approval-requested" event with the field "client" equal to "Claude Code" should have been emitted within 2000 ms
    When the MCP session is closed
    Then an "ai:approval-resolved" event with the field "outcome" equal to "denied" should have been emitted within 2000 ms
    When I call "AIAccessService.PendingApprovals"
    Then the response should contain the field "" array length equal to 0
    And the "DELETE" request to "/users/7" should have been received 0 times

  @quit
  Scenario: Quitting Xenon denies a pending approval and sends nothing
    When I start calling the MCP tool "send" in the background with
      """
      { "request": { "method": "DELETE", "url": "{{base_url}}/users/7" } }
      """
    Then an "ai:approval-requested" event with the field "client" equal to "Claude Code" should have been emitted within 2000 ms
    When the app is restarted
    Then an "ai:approval-resolved" event with the field "outcome" equal to "denied" should have been emitted within 2000 ms
    And the "DELETE" request to "/users/7" should have been received 0 times

  @skip_approvals
  Scenario: With "Dangerously skip approvals" on, the send goes out with no prompt and secrets stay masked
    Given I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "env:{{prod}}", "key": "admin_token", "secret": true, "value": "prod-admin-token-xyz" }
      """
    And I call "SettingsService.Save" with
      """
      { "ai_access": { "enabled": true, "port": {{mcp_port}}, "skip_approvals": true } }
      """
    And the mock API echoes every request to "/echo"
    And the emitted events are forgotten
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/echo", "headers": [ { "key": "Authorization", "value": "Bearer {{admin_token}}" } ] } }
      """
    Then the call should succeed
    And the response should contain the field "status" equal to "200"
    And the event "ai:approval-requested" should have been emitted 0 times
    And the Xenon window should have been brought to the front 0 times
    And the 1 "GET" request to "/echo" should have the header "Authorization" equal to "Bearer prod-admin-token-xyz"
    And the response should not contain "prod-admin-token-xyz"

  @not_protected
  Scenario: An environment without "Confirm before send" needs no approval
    Given I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}", "name": "Prod", "color": "red", "confirm_before_send": false }
      """
    And the emitted events are forgotten
    When I call the MCP tool "send" with
      """
      { "request": { "method": "DELETE", "url": "{{base_url}}/users/7" } }
      """
    Then the call should succeed
    And the event "ai:approval-requested" should have been emitted 0 times
    And the "DELETE" request to "/users/7" should have been received 1 times

  @acknowledgements
  Scenario: Missing secrets can be acknowledged by the AI, which sends the reference unresolved
    Given I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}", "name": "Prod", "color": "red", "confirm_before_send": false }
      """
    And the workspace file "acme/environments/prod.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{prod}}
      name: Prod
      color: red
      confirmBeforeSend: false
      variables:
        - key: base_url
          value: "{{api}}"
          enabled: true
          description: ""
        - key: missing_token
          secret: true
          enabled: true
          description: ""
      """
    And the 1 "GET" request to "/me" returns status 200 with the following response
      """
      {}
      """
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/me", "headers": [ { "key": "X-Token", "value": "{{missing_token}}" } ] } }
      """
    Then the call should fail with code "SND-06409"
    When I call the MCP tool "send" with
      """
      { "request": { "method": "GET", "url": "{{base_url}}/me", "headers": [ { "key": "X-Token", "value": "{{missing_token}}" } ] }, "send_unresolved": true }
      """
    Then the call should succeed
    And the 1 "GET" request to "/me" should have the header "X-Token" equal to "{{missing_token}}"

  @acknowledgements
  Scenario: An expired OAuth token is only sent when the AI says so
    Given I call "EnvironmentService.Update" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{prod}}", "name": "Prod", "color": "red", "confirm_before_send": false }
      """
    And the "oauth" variable is set to the mocked oauth provider url
    And the oauth provider is configured with
      """
      { "expires_in": 120 }
      """
    And the current time is "2026-10-13T10:00:00Z"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JU0000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: 01JU0000000000000000000001
      name: Billing
      auth:
        type: oauth2
        oauth2:
          grant_type: client_credentials
          accessTokenUrl: "{{oauth}}/token"
          clientId: xenon
          clientSecret: s3cret
          autoRefresh: false
      order:
        - me.yaml
      """
    And the workspace file "acme/collections/billing/me.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JU0000000000000000000002
      name: Me
      method: GET
      url: "{{api}}/me"
      auth:
        type: inherit
      """
    And I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "01JU0000000000000000000001" }
      """
    And the current time is "2026-10-13T10:05:00Z"
    When I call the MCP tool "send" with
      """
      { "request_id": "01JU0000000000000000000002" }
      """
    Then the call should fail with code "SND-07409"
    And the "GET" request to "/me" should have been received 0 times
    When I call the MCP tool "send" with
      """
      { "request_id": "01JU0000000000000000000002", "send_expired_token": true }
      """
    Then the call should succeed
    And the 1 "GET" request to "/me" should have the header "Authorization" equal to "Bearer access-1"
    And the response should not contain "access-1"
