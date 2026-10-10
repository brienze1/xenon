#language: en
#utf-8

@all @xen_003 @import @postman_v20
Feature: Import a Postman collection (v2.0)
  Postman v2.0 collections are accepted and normalised like v2.1: auth given as key/value maps
  becomes the same Xenon auth as the v2.1 attribute lists, so both exports of the same collection
  give the same Xenon collection.

  Background:
    Given the workspace "Acme" exists
    And the workspace "Globex" exists
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/full-v2.0.json"] }
      """
    And I save the response field "source_id" as "v20"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{v20}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/full-v2.1.json"] }
      """
    And I save the response field "source_id" as "v21"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{v21}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Globex.local_id}}" } ] }
      """

  @counts
  Scenario: Both versions import the same number of items and warnings
    Then the call should succeed
    And the response should contain the field "workspaces.0.created.requests" equal to "14"
    And the response should contain the field "warnings" array length equal to 17
    When I call "TreeService.GetTree" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.children.0.children.1.children.1.name" equal to "Delete user"
    And the response should contain the field "0.children.1.children" array length equal to 8

  @auth
  Scenario Outline: v2.0 auth maps give the same Xenon auth as v2.1 lists - <field> in <file>
    Then the workspace file "acme/collections/acme-api/<file>" should contain the field "<field>" equal to "<value>"
    And the workspace file "globex/collections/acme-api/<file>" should contain the field "<field>" equal to "<value>"

    Examples:
      | file                         | field                        | value              |
      | collection.yaml              | auth.type                    | bearer             |
      | collection.yaml              | auth.bearer.token            | {{token}}          |
      | users/folder.yaml            | auth.type                    | none               |
      | users/admin/folder.yaml      | auth.basic.username          | admin              |
      | users/admin/folder.yaml      | auth.basic.password          | {{admin_pw}}       |
      | users/admin/get-user.yaml    | auth.apikey.key              | X-Api-Key          |
      | users/admin/get-user.yaml    | auth.apikey.in               | header             |
      | users/admin/delete-user.yaml | auth.digest.algorithm        | MD5                |
      | bodies/form-data.yaml        | auth.oauth2.grant_type       | client_credentials |
      | bodies/form-data.yaml        | auth.oauth2.useBrowser       | false              |
      | bodies/binary.yaml           | auth.awsv4.service           | s3                 |
      | users/propfind.yaml          | auth.type                    | unsupported        |
      | users/propfind.yaml          | auth.unsupported.type        | hawk               |
      | users/propfind.yaml          | auth.unsupported.data.0.key  | authId             |
      | users/propfind.yaml          | auth.unsupported.data.0.value| hawk-id            |
      | users/list-users.yaml        | auth.type                    | inherit            |

  @fields
  Scenario Outline: v2.0 requests keep the same fields as v2.1 - <field> in <file>
    Then the workspace file "acme/collections/acme-api/<file>" should contain the field "<field>" equal to "<value>"
    And the workspace file "globex/collections/acme-api/<file>" should contain the field "<field>" equal to "<value>"

    Examples:
      | file                         | field                    | value                                        |
      | users/list-users.yaml        | params.2.enabled         | false                                        |
      | users/list-users.yaml        | examples.0.response.code | 200                                          |
      | users/folder.yaml            | scripts.disabled.0       | preRequest                                   |
      | bodies/form-data.yaml        | body.formdata.1.srcState | repick                                       |
      | rebuilt-url.yaml             | url                      | https://api.acme.test:8443/v1/rebuilt?q=1#frag |
      | users/admin/delete-user.yaml | settings.encodeUrl       | false                                        |

  @passthrough
  Scenario: The v2.0 schema is kept for export
    Then the workspace file "acme/collections/acme-api/collection.yaml" should contain the field "postman.info.schema" equal to "https://schema.getpostman.com/json/collection/v2.0.0/collection.json"
    And the workspace file "globex/collections/acme-api/collection.yaml" should contain the field "postman.info.schema" equal to "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
