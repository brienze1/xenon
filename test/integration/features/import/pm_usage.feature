#language: en
#utf-8

@all @xen_003 @import @pm_usage
Feature: Report pm APIs that imported scripts use outside the MVP script subset
  Every import summary lists each pm.* API used by the imported scripts that is outside the MVP
  subset (variables, request url/headers/body, response code/status/headers/json/text/responseTime,
  pm.test, pm.expect, pm.sendRequest), grouped by API with the item, script and line that use it.
  Scripts using pm.require are also flagged, because packages only work inside Postman. Mentions
  inside comments and strings do not count.

  Background:
    Given the workspace "Acme" exists

  @collection
  Scenario: The preview and the summary group the APIs outside the subset with their locations
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/postman/full-v2.1.json"] }
      """
    Then the call should succeed
    And the response should contain the field "pm_usages" array length equal to 4
    And I save the response field "source_id" as "source"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the response should contain the field "pm_usages" array length equal to 4
    And the response should contain the field "pm_usages.0.api" equal to "pm.cookies.jar"
    And the response should contain the field "pm_usages.0.occurrences.0.path" equal to "Acme API/Bodies/Raw JSON"
    And the response should contain the field "pm_usages.0.occurrences.0.phase" equal to "preRequest"
    And the response should contain the field "pm_usages.0.occurrences.0.line" equal to "2"
    And the response should contain the field "pm_usages.1.api" equal to "pm.info.requestName"
    And the response should contain the field "pm_usages.1.occurrences.0.line" equal to "3"
    And the response should contain the field "pm_usages.2.api" equal to "pm.require"
    And the response should contain the field "pm_usages.2.occurrences.0.path" equal to "Acme API/Bodies/Form data"
    And the response should contain the field "pm_usages.2.occurrences.0.line" equal to "1"
    And the response should contain the field "pm_usages.3.api" equal to "pm.response.to"
    And the response should contain the field "pm_usages.3.occurrences.0.path" equal to "Acme API"
    And the response should contain the field "pm_usages.3.occurrences.0.phase" equal to "postResponse"
    And the response should not contain "pm.visualizer"
    And the response should not contain "pm.iterationData"
    And the response field "warnings" should contain an item with "code" equal to "pm_require" and "path" equal to "Acme API/Bodies/Form data"
    And the response field "warnings" should contain an item with "code" equal to "pm_api_outside_mvp" and "message" equal to "The script uses pm.cookies.jar, which is outside Xenon's script subset"

  @grouping
  Scenario: Occurrences of the same API in several scripts are grouped
    Given the file "usage.postman_collection.json" exists with content
      """
      {
        "info": { "name": "Usage", "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json" },
        "item": [
          { "name": "One", "event": [ { "listen": "prerequest", "script": { "exec": [
              "pm.execution.setNextRequest(null);",
              "/* pm.cookies.jar() in a block comment */",
              "const url = `${pm.request.url}`;",
              "pm.request.headers.add({ key: 'a', value: 'b' });",
              "pm.sendRequest('https://x.test', () => {});"
            ] } } ], "request": { "method": "GET", "url": "https://x.test/one" } },
          { "name": "Two", "event": [ { "listen": "test", "script": { "exec": [
              "pm.test('m', () => pm.expect(pm.request.method).to.eql('GET'));",
              "pm.execution.setNextRequest('One');"
            ] } } ], "request": { "method": "GET", "url": "https://x.test/two" } }
        ]
      }
      """
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{temp_root}}/usage.postman_collection.json"] }
      """
    Then the call should succeed
    And the response should contain the field "pm_usages" array length equal to 2
    And the response should contain the field "pm_usages.0.api" equal to "pm.execution.setNextRequest"
    And the response should contain the field "pm_usages.0.occurrences" array length equal to 2
    And the response should contain the field "pm_usages.0.occurrences.0.path" equal to "Usage/One"
    And the response should contain the field "pm_usages.0.occurrences.1.path" equal to "Usage/Two"
    And the response should contain the field "pm_usages.0.occurrences.1.line" equal to "2"
    And the response should contain the field "pm_usages.1.api" equal to "pm.request.method"
    And the response should contain the field "pm_usages.1.occurrences.0.phase" equal to "postResponse"
