#language: en
#utf-8

@all @xen_016 @aliases @alias_send
Feature: Sending an alias sends that version, runs its own scripts and uses its own OAuth token
  The tab sends the version on screen with the alias it shows. Go checks the alias exists on the
  saved request (nothing is sent otherwise), runs the alias's stored scripts when the tab did not
  edit them - so each version's scripts can set different variables in the usual scopes - and gives
  an alias with its own OAuth 2.0 auth its own token slot, "<request id>:<alias id>". History records
  the alias id and name of every alias send.

  Background:
    Given the workspace "Acme" exists
    And the keychain is available
    And the "api" variable is set to the mocked api url
    And the "oauth" variable is set to the mocked oauth provider url
    And the "billing" variable is set to "01JS0000000000000000000001"
    And the "login" variable is set to "01JS0000000000000000000002"
    And the "dev" variable is set to "01JS0000000000000000000009"
    And the "mx" variable is set to "01JS0000000000000000000010"
    And the "unknown" variable is set to "01JS0000000000000000000099"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JS0000000000000000000000
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
      order:
        - login.yaml
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: token
          value: ""
          enabled: true
          description: ""
        - key: mxToken
          value: ""
          enabled: true
          description: ""
      """
    And the workspace file "acme/collections/billing/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: "{{api}}/login"
      body:
        mode: raw
        raw:
          language: json
          content: '{"username":"luis"}'
      scripts:
        postResponse: pm.environment.set("token", pm.response.json().access_token)
      aliases:
        - id: {{mx}}
          name: mx-test-user
          method: POST
          url: "{{api}}/login"
          body:
            mode: raw
            raw:
              language: json
              content: '{"username":"mx-test"}'
          scripts:
            preRequest: ""
            postResponse: pm.environment.set("mxToken", pm.response.json().access_token)
          inheritScripts: true
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{dev}}" }
      """

  @content @history
  Scenario: Sending an alias sends its content and History records the alias
    Given the 1 "POST" request to "/login" returns status 200 with the following response
      """
      {"access_token":"tok-mx"}
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{mx}}",
        "request": {
          "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login",
          "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"username\":\"mx-test\"}" } }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "status_code" equal to "200"
    And the 1 "POST" request to "/login" should have the body field "username" equal to "mx-test"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the call should succeed
    And the response should contain the field "" array length equal to 1
    And the response should contain the field "0.source_request_id" equal to "{{login}}"
    And the response should contain the field "0.alias_id" equal to "{{mx}}"
    And the response should contain the field "0.alias_name" equal to "mx-test-user"

  @history
  Scenario: A Default send and an unsaved-tab send record no alias
    Given the 1 "POST" request to "/login" returns status 200 with the following response
      """
      {"access_token":"tok-default"}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-2", "request": { "method": "GET", "url": "{{api}}/ping" } }
      """
    Then the call should succeed
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 2
    And the response field "0.alias_id" should not exist
    And the response field "0.alias_name" should not exist
    And the response field "1.alias_id" should not exist
    And the response field "1.alias_name" should not exist
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{mx}}", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 3
    And the response should contain the field "0.alias_name" equal to "mx-test-user"

  @scripts @variables
  Scenario: Each version's stored scripts set their own variables
    Given the 1 "POST" request to "/login" returns status 200 with the following response
      """
      {"access_token":"tok-default"}
      """
    And the 2 "POST" request to "/login" returns status 200 with the following response
      """
      {"access_token":"tok-mx"}
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.token" equal to "tok-default"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "mxToken"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{mx}}", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.mxToken" equal to "tok-mx"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.token" equal to "tok-default"

  @scripts @edited
  Scenario: Scripts edited in the tab win over the alias's stored scripts
    Given the 1 "POST" request to "/login" returns status 200 with the following response
      """
      {"access_token":"tok-edited"}
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{mx}}",
        "request": {
          "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login",
          "scripts": { "pre_request": "", "post_response": "pm.environment.set(\"token\", \"from-editor\")" }
        }
      }
      """
    Then the call should succeed
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should contain the field "current.env:{{dev}}.token" equal to "from-editor"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "tok-edited"
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "mxToken"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.alias_name" equal to "mx-test-user"

  @errors
  Scenario: Sending an alias the request does not have fails and sends nothing
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{unknown}}", "request": { "id": "{{login}}", "name": "Login", "method": "POST", "url": "{{api}}/login" } }
      """
    Then the call should fail with code "ALS-01404"
    And the "POST" request to "/login" should have been received 0 times
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0

  @oauth2
  Scenario: An alias with its own OAuth 2.0 auth has its own token, separate from the Default's
    Given the workspace file "acme/collections/billing/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: GET
      url: "{{api}}/me"
      auth:
        type: oauth2
        oauth2:
          grant_type: client_credentials
          accessTokenUrl: "{{oauth}}/token"
          clientId: xenon
          clientSecret: s3cret
          scope: api
      aliases:
        - id: {{mx}}
          name: mx-test-user
          method: GET
          url: "{{api}}/me"
          auth:
            type: oauth2
            oauth2:
              grant_type: client_credentials
              accessTokenUrl: "{{oauth}}/token"
              clientId: mx
              clientSecret: mx-s3cret
              scope: api
      """
    When I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{login}}" }
      """
    Then the call should succeed
    When I call "OAuthService.GetNewToken" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{login}}:{{mx}}" }
      """
    Then the call should succeed
    And the response should contain the field "owner_id" equal to "{{login}}:{{mx}}"
    And the oauth provider should have received 2 "client_credentials" token requests
    And the 1 "client_credentials" token request should have the header "Authorization" equal to "Basic eGVub246czNjcmV0"
    And the 2 "client_credentials" token request should have the header "Authorization" equal to "Basic bXg6bXgtczNjcmV0"
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}", "tab_id": "tab-1",
        "request": {
          "id": "{{login}}", "name": "Login", "method": "GET", "url": "{{api}}/me",
          "auth": { "type": "oauth2", "oauth2": { "grant_type": "client_credentials", "access_token_url": "{{oauth}}/token", "client_id": "xenon", "client_secret": "s3cret", "scope": "api" } }
        }
      }
      """
    Then the call should succeed
    And the 1 "GET" request to "/me" should have the header "Authorization" equal to "Bearer access-1"
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "alias_id": "{{mx}}",
        "request": {
          "id": "{{login}}", "name": "Login", "method": "GET", "url": "{{api}}/me",
          "auth": { "type": "oauth2", "oauth2": { "grant_type": "client_credentials", "access_token_url": "{{oauth}}/token", "client_id": "mx", "client_secret": "mx-s3cret", "scope": "api" } }
        }
      }
      """
    Then the call should succeed
    And the 2 "GET" request to "/me" should have the header "Authorization" equal to "Bearer access-2"
    When I call "OAuthService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{login}}" }
      """
    Then the response should contain the field "token_type" equal to "Bearer"
    When I call "RequestService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "{{login}}" }
      """
    And I save the response field "revision" as "revision"
    And I call "AliasService.Delete" with
      """
      { "local_id": "{{Acme.local_id}}", "request_id": "{{login}}", "alias_id": "{{mx}}", "base_revision": "{{revision}}" }
      """
    Then the call should succeed
    When I call "OAuthService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{login}}:{{mx}}" }
      """
    Then the response should contain the field "" equal to "null"
    When I call "OAuthService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "owner_id": "{{login}}" }
      """
    Then the response should contain the field "token_type" equal to "Bearer"
