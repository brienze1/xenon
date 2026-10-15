#language: en
#utf-8

@all @xen_003 @import @format_detection
Feature: Detect the format of an import source
  The Import dialog hands Go the picked or dropped paths, or pasted text. Go works out what it is
  (Postman collection v2.0/v2.1, environment, globals or data dump as a zip or a folder; Postman v3
  JSON, git-native folders, request files, environments, globals and `postman workspace pull`
  folders; Bruno .bru or YAML collection; cURL) and refuses anything else before anything is
  written, naming the problem. Archives and files are untrusted: zip entries with absolute paths, ".." segments or
  symbolic links are refused, archives may not unpack to more than the archive limit (1 GB) and
  single files may not exceed the file limit (200 MB); input must be UTF-8.

  Background:
    Given the workspace "Acme" exists

  @detect
  Scenario Outline: Each supported source is detected as its kind - <source>
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/<source>"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "<kind>"
    And the response should contain the field "version" equal to "<version>"
    And the response should contain the field "source_id" matching "^[0-9A-HJKMNP-TV-Z]{26}$"

    Examples:
      | source                                | kind                | version |
      | postman/full-v2.1.json                | postman_collection  | 2.1     |
      | postman/full-v2.0.json                | postman_collection  | 2.0     |
      | postman/dev.postman_environment.json  | postman_environment |         |
      | postman/globals.json                  | postman_globals     |         |
      | postman/dump                          | postman_dump        |         |
      | postman/dump.zip                      | postman_dump        |         |
      | misc/zip-nested.zip                   | postman_dump        |         |
      | bruno/bru-collection                  | bruno_bru           |         |
      | bruno/yaml-collection                 | bruno_yaml          |         |

  @detect @xen_009 @postman_v3_detect
  Scenario Outline: Postman v3 sources are detected with version 3 - <source>
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/<source>"] }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "<kind>"
    And the response should contain the field "version" equal to "3"
    And the response should contain the field "source_id" matching "^[0-9A-HJKMNP-TV-Z]{26}$"

    Examples:
      | source                                         | kind                |
      | postman/v3/native-git                          | postman_collection  |
      | postman/v3/native-git/Get users.request.yaml   | postman_request     |
      | postman/v3/rich.json                           | postman_collection  |
      | postman/v3/rich                                | postman_collection  |
      | postman/v3/extras                              | postman_collection  |
      | postman/v3/large-1000-v3.zip                   | postman_collection  |
      | postman/v3/env/staging.environment.yaml        | postman_environment |
      | postman/v3/env/cli-environment.json            | postman_environment |
      | postman/v3/env/workspace.globals.yaml          | postman_globals     |
      | postman/v3/env/cloud-globals.json              | postman_globals     |
      | postman/v3/env/local-globals.json              | postman_globals     |
      | postman/v3/workspace                           | postman_workspace   |
      | postman/v3/workspace/postman                   | postman_workspace   |
      | postman/v3/workspace/postman/collections       | postman_workspace   |
      | postman/v3/workspace.zip                       | postman_workspace   |

  @detect @collection
  Scenario: A Postman collection lists its single item with counts
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/full-v2.1.json"] }
      """
    Then the call should succeed
    And the response should contain the field "items" array length equal to 1
    And the response should contain the field "items.0.kind" equal to "collection"
    And the response should contain the field "items.0.name" equal to "Acme API"
    And the response should contain the field "items.0.counts.folders" equal to "3"
    And the response should contain the field "items.0.counts.requests" equal to "14"
    And the response should contain the field "items.0.counts.examples" equal to "3"
    And the response should contain the field "items.0.counts.variables" equal to "3"
    And the response should contain the field "items.0.counts.scripts" equal to "7"
    And the response should contain the field "items.0.warnings_count" not equal to "0"

  @detect @paste
  Scenario: Pasted Postman JSON is detected like a file
    When I call "ImportService.Detect" with
      """
      { "text": "{\"info\":{\"name\":\"Pasted\",\"schema\":\"https://schema.getpostman.com/json/collection/v2.1.0/collection.json\"},\"item\":[]}" }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "postman_collection"
    And the response should contain the field "items.0.name" equal to "Pasted"

  @detect @curl
  Scenario Outline: Pasted cURL commands are detected - <case>
    When I call "ImportService.Detect" with
      """
      { "text": "<text>" }
      """
    Then the call should succeed
    And the response should contain the field "kind" equal to "curl"

    Examples:
      | case            | text                                    |
      | plain           | curl https://api.test/users             |
      | with flags      | curl -X POST https://api.test/users     |
      | upper case      | CURL https://api.test/users             |

  @refused
  Scenario Outline: Unsupported or broken sources are refused before anything is written - <case>
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/<source>"] }
      """
    Then the call should fail with code "<code>"
    And the error should contain the field "error.description" equal to "<message>"
    And the workspace folder "acme" should contain the entries ".gitignore, collections, environments, workspace.yaml"

    Examples:
      | case                | source                 | code       | message                                                                       |
      | postman v1          | postman/v1.json        | IMP-01400  | Postman v1 collections are not supported; export the collection as v2.1        |
      | random JSON         | misc/random.json       | IMP-00415  | random.json is not a Xenon, Postman, Bruno or cURL export                           |
      | other YAML          | misc/notes.yaml        | IMP-00415  | notes.yaml is not a Xenon, Postman, Bruno or cURL export                            |
      | malformed JSON      | misc/malformed.json    | IMP-01400  | malformed.json is not valid JSON                                              |
      | not UTF-8           | misc/latin1.json       | IMP-01400  | latin1.json is not UTF-8 text                                                 |
      | mixed bruno formats | bruno/mixed-collection | IMP-00415  | mixed-collection mixes Bruno .bru and YAML files                              |

  @refused @archive
  Scenario Outline: Unsafe archive entries are refused - <case>
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/misc/<archive>"] }
      """
    Then the call should fail with code "IMP-03400"
    And the error should contain the field "error.description" equal to "<archive> contains the unsafe entry <entry>"
    And no file under the workspace folder should contain "evil"

    Examples:
      | case          | archive          | entry          |
      | parent folder | zip-slip.zip     | ../evil.json   |
      | absolute path | zip-absolute.zip | /tmp/evil.json |
      | symbolic link | zip-symlink.zip  | link.json      |

  @refused @limits
  Scenario: An archive that unpacks to more than the archive limit is refused
    Given the "XENON_IMPORT_ARCHIVE_LIMIT" env var is set to "1000"
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dump.zip"] }
      """
    Then the call should fail with code "IMP-02413"
    And the error should contain the field "error.description" equal to "dump.zip unpacks to more than 1000 bytes"

  @refused @limits
  Scenario: A file larger than the file limit is refused
    Given the "XENON_IMPORT_FILE_LIMIT" env var is set to "100"
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/globals.json"] }
      """
    Then the call should fail with code "IMP-39413"
    And the error should contain the field "error.description" equal to "globals.json is larger than 100 bytes"

  @refused
  Scenario Outline: Text that is neither JSON nor cURL is refused - <case>
    When I call "ImportService.Detect" with
      """
      { "text": "<text>" }
      """
    Then the call should fail with code "<code>"

    Examples:
      | case        | text                     | code      |
      | plain words | hello there              | IMP-00415 |
      | broken JSON | { broken                 | IMP-01400 |
      | url only    | https://api.test/users   | IMP-00415 |

  @refused @not_found
  Scenario: A path that does not exist is refused
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/missing.json"] }
      """
    Then the call should fail with code "IMP-08404"

  @validation
  Scenario: A detect without paths or text is rejected
    When I call "ImportService.Detect" with
      """
      { "paths": [], "text": "  " }
      """
    Then the call should fail with code "IMP-09400"

  @discard
  Scenario: A discarded source can no longer be imported
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/dump.zip"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Discard" with
      """
      { "source_id": "{{source}}" }
      """
    Then the call should succeed
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [] }
      """
    Then the call should fail with code "IMP-05404"
