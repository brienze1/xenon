#language: en
#utf-8

@all @xen_004 @scripts @script_tools
Feature: Script editor tools
  The script editor asks Go for three things. Check compiles a script without running it (top-level
  await allowed, as when it runs) and lists its syntax errors with V8's wording, line and column.
  Reference lists the sandbox surface for autocomplete: every pm member, the response assertion
  helpers and the globals, each with its path, signature, a short description and its kind
  (object, function or property); it is generated at build time from the shims. Snippets lists
  Postman's standard snippets with the exact text Postman inserts and the phase they belong to.

  @check_valid
  Scenario Outline: Check finds no errors in a valid script - <case>
    When I call "ScriptService.Check" with
      """
      { "text": "<text>" }
      """
    Then the call should succeed
    And the response should contain the field "errors" array length equal to 0

    Examples:
      | case            | text                                                                 |
      | empty           |                                                                      |
      | statements      | pm.test('ok', () => pm.response.to.have.status(200))                 |
      | top-level await | const res = await pm.sendRequest('https://x.io');\\nconsole.log(res) |
      | return          | if (!pm.response) { return }                                         |

  @check_errors
  Scenario Outline: Check reports a syntax error with its location - <case>
    When I call "ScriptService.Check" with
      """
      { "text": "<text>" }
      """
    Then the call should succeed
    And the response should contain the field "errors" array length equal to 1
    And the response should contain the field "errors.0.message" equal to "<message>"
    And the response should contain the field "errors.0.line" equal to "<line>"
    And the response should contain the field "errors.0.column" equal to "<column>"

    Examples:
      | case             | text                                     | message                              | line | column |
      | unexpected token | const a = ;                              | SyntaxError: Unexpected token ';'    | 1    | 11     |
      | third line       | pm.test('a', () => {})\\n\\nlet = = 2    | SyntaxError: Unexpected token '='    | 3    | 7      |
      | unclosed         | pm.test('a', () => {                     | SyntaxError: Unexpected end of input | 1    | 21     |

  @reference
  Scenario Outline: The reference describes <path>
    When I call "ScriptService.Reference"
    Then the call should succeed
    And the response field "members" should contain an item with "path" equal to "<path>" and "kind" equal to "<kind>"

    Examples:
      | path                                 | kind     |
      | pm                                   | object   |
      | pm.variables                         | object   |
      | pm.variables.get                     | function |
      | pm.variables.replaceIn               | function |
      | pm.environment.name                  | property |
      | pm.environment.set                   | function |
      | pm.environment.unset                 | function |
      | pm.collectionVariables.toObject      | function |
      | pm.globals.clear                     | function |
      | pm.iterationData.toJSON              | function |
      | pm.request.url                       | object   |
      | pm.request.url.getQueryString        | function |
      | pm.request.method                    | property |
      | pm.request.headers.upsert            | function |
      | pm.request.headers.toObject          | function |
      | pm.request.body.update               | function |
      | pm.request.addQueryParams            | function |
      | pm.request.auth                      | object   |
      | pm.response.code                     | property |
      | pm.response.responseTime             | property |
      | pm.response.json                     | function |
      | pm.response.text                     | function |
      | pm.response.headers.get              | function |
      | pm.response.to.have.status           | function |
      | pm.response.to.have.header           | function |
      | pm.response.to.have.jsonBody         | function |
      | pm.response.to.have.jsonSchema       | function |
      | pm.response.to.be.ok                 | property |
      | pm.response.to.be.notFound           | property |
      | pm.response.to.be.clientError        | property |
      | pm.info.requestName                  | property |
      | pm.info.eventName                    | property |
      | pm.cookies.get                       | function |
      | pm.cookies.jar                       | function |
      | pm.sendRequest                       | function |
      | pm.execution.skipRequest             | function |
      | pm.execution.runRequest              | function |
      | pm.execution.location                | property |
      | pm.execution.setNextRequest          | function |
      | pm.test                              | function |
      | pm.test.skip                         | function |
      | pm.expect                            | function |
      | pm.vault.get                         | function |
      | pm.visualizer.set                    | function |
      | console.log                          | function |
      | console.clear                        | function |
      | setTimeout                           | function |
      | atob                                 | function |
      | Buffer                               | function |
      | URLSearchParams                      | function |
      | crypto.randomUUID                    | function |
      | crypto.subtle.digest                 | function |
      | structuredClone                      | function |
      | require                              | function |
      | CryptoJS                             | object   |
      | _                                    | function |

  @reference_details
  Scenario: Reference entries carry a signature and a description
    When I call "ScriptService.Reference"
    Then the call should succeed
    And the response field "members" should contain an item with "path" equal to "pm.environment.set" and "signature" equal to "set(key: string, value: any): void"
    And the response field "members" should contain an item with "path" equal to "pm.environment.set" and "description" equal to "Sets the current value of a variable in the active environment."
    And the response field "members" should contain an item with "path" equal to "pm.response.to.have.status" and "signature" equal to "status(code: number | string)"
    And the response field "members" should contain 0 items with "description" equal to ""
    And the response field "members" should contain 0 items with "signature" equal to ""

  @snippets
  Scenario: Snippets are Postman's standard snippets
    When I call "ScriptService.Snippets"
    Then the call should succeed
    And the response should contain the field "" array length equal to 21
    And the response field "" should contain an item with "id" equal to "status-code-200" and "title" equal to "Status code: Code is 200"
    And the response field "" should contain an item with "id" equal to "status-code-200" and "phase" equal to "postResponse"
    And the response field "" should contain an item with "id" equal to "get-environment-variable" and "phase" equal to "any"
    And the response should contain the field "0.id" equal to "get-environment-variable"
    And the response should contain the field "0.text" equal to
      """
      pm.environment.get("variable_key");
      """
    And the response should contain the field "5.id" equal to "set-global-variable"
    And the response should contain the field "5.text" equal to
      """
      pm.globals.set("variable_key", "variable_value");
      """
    And the response should contain the field "11.id" equal to "status-code-200"
    And the response should contain the field "11.text" equal to
      """
      pm.test("Status code is 200", function () {
          pm.response.to.have.status(200);
      });
      """
