#language: en
#utf-8

@all @xen_003 @import @curl
Feature: Import a cURL command
  A cURL command copied from a terminal, browser DevTools, docs, Windows cmd or PowerShell becomes a
  request. Line continuations (bash \, cmd ^, PowerShell `) and each shell's quoting are
  normalised first, so the same command in the three syntaxes gives the same request (SC-005).
  Supported flags map as in FR-024, no-effect flags are ignored silently and unknown flags are
  listed in the notices. Pasted into the URL bar, RequestService.ParseCurl returns the request
  for the tab; in the import dialog the command becomes a new saved request, named after its
  method and path, in the chosen collection or folder. Text that is not a cURL command is refused
  (IMP-06400), as are a command without a URL (IMP-40400), an unreadable command (IMP-41400)
  and a URL that is not HTTP(S) (IMP-07400).

  Background:
    Given the workspace "Acme" exists

  @corpus
  Scenario Outline: The corpus command "<command>" in <dialect> syntax yields the expected request
    Given the "command" variable is set to the field "commands.<index>.<dialect>" of the resource file "import/curl/corpus.yaml"
    When I call "RequestService.ParseCurl" with
      """
      { "text": "{{command}}" }
      """
    Then the call should succeed
    And the response should match the field "commands.<index>.expected" of the resource file "import/curl/corpus.yaml"

    Examples:
      | command                    | dialect    | index |
      | simple get                 | bash       | 0     |
      | simple get                 | cmd        | 0     |
      | simple get                 | powershell | 0     |
      | explicit method            | bash       | 1     |
      | explicit method            | cmd        | 1     |
      | explicit method            | powershell | 1     |
      | long request flag          | bash       | 2     |
      | long request flag          | cmd        | 2     |
      | long request flag          | powershell | 2     |
      | headers                    | bash       | 3     |
      | headers                    | cmd        | 3     |
      | headers                    | powershell | 3     |
      | json data                  | bash       | 4     |
      | json data                  | cmd        | 4     |
      | json data                  | powershell | 4     |
      | data implies post          | bash       | 5     |
      | data implies post          | cmd        | 5     |
      | data implies post          | powershell | 5     |
      | data raw                   | bash       | 6     |
      | data raw                   | cmd        | 6     |
      | data raw                   | powershell | 6     |
      | data ascii                 | bash       | 7     |
      | data ascii                 | cmd        | 7     |
      | data ascii                 | powershell | 7     |
      | data binary file           | bash       | 8     |
      | data binary file           | cmd        | 8     |
      | data binary file           | powershell | 8     |
      | data urlencode             | bash       | 9     |
      | data urlencode             | cmd        | 9     |
      | data urlencode             | powershell | 9     |
      | form fields                | bash       | 10    |
      | form fields                | cmd        | 10    |
      | form fields                | powershell | 10    |
      | form long flag             | bash       | 11    |
      | form long flag             | cmd        | 11    |
      | form long flag             | powershell | 11    |
      | get with data              | bash       | 12    |
      | get with data              | cmd        | 12    |
      | get with data              | powershell | 12    |
      | head                       | bash       | 13    |
      | head                       | cmd        | 13    |
      | head                       | powershell | 13    |
      | upload file                | bash       | 14    |
      | upload file                | cmd        | 14    |
      | upload file                | powershell | 14    |
      | url flag                   | bash       | 15    |
      | url flag                   | cmd        | 15    |
      | url flag                   | powershell | 15    |
      | basic user                 | bash       | 16    |
      | basic user                 | cmd        | 16    |
      | basic user                 | powershell | 16    |
      | basic flag                 | bash       | 17    |
      | basic flag                 | cmd        | 17    |
      | basic flag                 | powershell | 17    |
      | digest                     | bash       | 18    |
      | digest                     | cmd        | 18    |
      | digest                     | powershell | 18    |
      | cookie                     | bash       | 19    |
      | cookie                     | cmd        | 19    |
      | cookie                     | powershell | 19    |
      | user agent                 | bash       | 20    |
      | user agent                 | cmd        | 20    |
      | user agent                 | powershell | 20    |
      | referer                    | bash       | 21    |
      | referer                    | cmd        | 21    |
      | referer                    | powershell | 21    |
      | json flag                  | bash       | 22    |
      | json flag                  | cmd        | 22    |
      | json flag                  | powershell | 22    |
      | insecure                   | bash       | 23    |
      | insecure                   | cmd        | 23    |
      | insecure                   | powershell | 23    |
      | location                   | bash       | 24    |
      | location                   | cmd        | 24    |
      | location                   | powershell | 24    |
      | compressed ignored         | bash       | 25    |
      | compressed ignored         | cmd        | 25    |
      | compressed ignored         | powershell | 25    |
      | silent and verbose ignored | bash       | 26    |
      | silent and verbose ignored | cmd        | 26    |
      | silent and verbose ignored | powershell | 26    |
      | unknown flag listed        | bash       | 27    |
      | unknown flag listed        | cmd        | 27    |
      | unknown flag listed        | powershell | 27    |
      | quotes inside              | bash       | 28    |
      | quotes inside              | cmd        | 28    |
      | quotes inside              | powershell | 28    |
      | query string kept          | bash       | 29    |
      | query string kept          | cmd        | 29    |
      | query string kept          | powershell | 29    |
      | devtools style             | bash       | 30    |
      | devtools style             | cmd        | 30    |
      | devtools style             | powershell | 30    |
      | variables kept             | bash       | 31    |
      | variables kept             | cmd        | 31    |
      | variables kept             | powershell | 31    |

  @flags
  Scenario Outline: Long and short forms of the supported flags map the same way - <flags>
    When I call "RequestService.ParseCurl" with
      """
      { "text": "curl <flags> https://api.test/x" }
      """
    Then the call should succeed
    And the response should contain the field "request.<field>" equal to "<value>"
    And the response should contain the field "notices" array length equal to 0

    Examples:
      | flags                               | field                     | value                   |
      | --header 'X-A: 1'                   | headers.0.key             | X-A                     |
      | -H'X-A: 1'                          | headers.0.value           | 1                       |
      | --cookie 'a=1'                      | headers.0.key             | Cookie                  |
      | --user-agent 'Agent/2'              | headers.0.value           | Agent/2                 |
      | --referer 'https://r.test'          | headers.0.key             | Referer                 |
      | --insecure                          | settings.verify_ssl       | false                   |
      | --location                          | settings.follow_redirects | true                    |
      | --get --data 'q=1'                  | url                       | https://api.test/x?q=1  |
      | --head                              | method                    | HEAD                    |
      | --upload-file '/tmp/f.bin'          | method                    | PUT                     |
      | -XPOST                              | method                    | POST                    |
      | -sSL                                | settings.follow_redirects | true                    |
      | --data-binary 'raw text'            | body.raw.content          | raw text                |
      | -d @/tmp/body.json                  | body.binary.src           | /tmp/body.json          |
      | -F 'doc=@/tmp/d.pdf;type=application/pdf' | body.formdata.0.src | /tmp/d.pdf             |
      | --user 'ana'                        | auth.basic.username       | ana                     |
      | --max-time 5 --retry 2              | method                    | GET                     |

  @quoting
  Scenario: ANSI-C quoting is decoded
    When I call "RequestService.ParseCurl" with
      """
      { "text": "curl $'https://api.test/x' -H $'X-Q: it\\'s' --data-raw $'{\\\"a\\\":\\n1}'" }
      """
    Then the call should succeed
    And the response should contain the field "request.url" equal to "https://api.test/x"
    And the response should contain the field "request.headers.0.value" equal to "it's"
    And the response should contain the field "request.body.raw.content" equal to
      """
      {"a":
      1}
      """

  @notices
  Scenario: Several URLs use the first HTTP one and list the others
    When I call "RequestService.ParseCurl" with
      """
      { "text": "curl https://api.test/one https://api.test/two --frobnicate" }
      """
    Then the call should succeed
    And the response should contain the field "request.url" equal to "https://api.test/one"
    And the response should contain the field "notices" with the items "Ignored unknown option --frobnicate, Ignored the extra URL https://api.test/two"

  @refused
  Scenario Outline: Text that is not a usable cURL command is refused - <case>
    When I call "RequestService.ParseCurl" with
      """
      { "text": "<text>" }
      """
    Then the call should fail with code "<code>"
    And the error should contain the field "error.description" equal to "<message>"

    Examples:
      | case             | text                           | code      | message                                         |
      | plain URL        | https://api.test/users         | IMP-06400 | The text is not a cURL command                  |
      | other command    | wget https://api.test/users    | IMP-06400 | The text is not a cURL command                  |
      | no URL           | curl -X POST                   | IMP-40400 | The cURL command has no URL                     |
      | broken quoting   | curl 'https://api.test/x       | IMP-41400 | The cURL command could not be read              |
      | ftp URL          | curl ftp://files.test/a.txt    | IMP-07400 | Only HTTP and HTTPS URLs can be imported, not ftp://files.test/a.txt |
      | file URL         | curl file:///etc/passwd        | IMP-07400 | Only HTTP and HTTPS URLs can be imported, not file:///etc/passwd |

  @validation
  Scenario: A parse without text is rejected
    When I call "RequestService.ParseCurl" with
      """
      { "text": "  " }
      """
    Then the call should fail with code "REQ-18400"

  @dialog
  Scenario: The import dialog creates a named request in the chosen folder
    Given I call "TreeService.CreateCollection" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "Main" }
      """
    And I save the response field "id" as "main"
    And I call "TreeService.CreateFolder" with
      """
      { "local_id": "{{Acme.local_id}}", "parent_id": "{{main}}", "name": "Users" }
      """
    And I save the response field "id" as "users"
    When I call "ImportService.Detect" with
      """
      { "text": "curl -X POST 'https://api.test/users?invite=1' -H 'Content-Type: application/json' --digest -u 'ana:pw' -d '{\"name\":\"Ana\"}'" }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "curl"
    And the response should contain the field "items" array length equal to 1
    And the response should contain the field "items.0.kind" equal to "request"
    And the response should contain the field "items.0.name" equal to "POST /users"
    And the response should contain the field "items.0.counts.requests" equal to "1"
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{users}}", "name": "Invite user" } ] }
      """
    Then the call should succeed
    And the response should contain the field "items.0.status" equal to "imported"
    And the response should contain the field "items.0.name" equal to "Invite user"
    And the response should contain the field "workspaces.0.created.requests" equal to "1"
    And the response should contain the field "warnings" array length equal to 0
    And the workspace file "acme/collections/main/users/folder.yaml" should contain the field "order" with the items "invite-user.yaml"
    And the workspace file "acme/collections/main/users/invite-user.yaml" should contain the field "name" equal to "Invite user"
    And the workspace file "acme/collections/main/users/invite-user.yaml" should contain the field "method" equal to "POST"
    And the workspace file "acme/collections/main/users/invite-user.yaml" should contain the field "url" equal to "https://api.test/users?invite=1"
    And the workspace file "acme/collections/main/users/invite-user.yaml" should contain the field "params.0.key" equal to "invite"
    And the workspace file "acme/collections/main/users/invite-user.yaml" should contain the field "headers.0.key" equal to "Content-Type"
    And the workspace file "acme/collections/main/users/invite-user.yaml" should contain the field "body.mode" equal to "raw"
    And the workspace file "acme/collections/main/users/invite-user.yaml" should contain the field "body.raw.language" equal to "json"
    And the workspace file "acme/collections/main/users/invite-user.yaml" should contain the field "body.raw.content" equal to
      """
      {"name":"Ana"}
      """
    And the workspace file "acme/collections/main/users/invite-user.yaml" should contain the field "auth.type" equal to "digest"
    And the workspace file "acme/collections/main/users/invite-user.yaml" should contain the field "auth.digest.username" equal to "ana"
    And the workspace file "acme/collections/main/users/invite-user.yaml" should contain the field "auth.digest.password" equal to "pw"

  @dialog @default_name
  Scenario: Without a name the request is named after its method and path
    Given I call "TreeService.CreateCollection" with
      """
      { "local_id": "{{Acme.local_id}}", "name": "Main" }
      """
    And I save the response field "id" as "main"
    When I call "ImportService.Detect" with
      """
      { "text": "curl https://api.test" }
      """
    Then the response should contain the field "items.0.name" equal to "GET api.test"
    When I call "ImportService.Detect" with
      """
      { "text": "curl.exe -X DELETE '{{base_url}}/users/{{id}}?hard=true'" }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "curl"
    And the response should contain the field "items.0.name" equal to "DELETE /users/{{id}}"
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}", "target_parent_id": "{{main}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/main/delete-users-{{id}}.yaml" should contain the field "name" equal to "DELETE /users/{{id}}"

  @dialog @refused
  Scenario: A cURL request needs a collection or folder to go into
    When I call "ImportService.Detect" with
      """
      { "text": "curl https://api.test/users" }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should fail with code "IMP-35400"
    And the error should contain the field "error.description" equal to "A request must be imported into a collection or folder"

  @dialog @refused
  Scenario: Detecting text that is not a usable cURL command is refused
    When I call "ImportService.Detect" with
      """
      { "text": "curl ftp://files.test/a.txt" }
      """
    Then the call should fail with code "IMP-07400"
