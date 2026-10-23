#language: en
#utf-8

@newman @runner_newman
Feature: The runner follows the same request sequence as Postman's runner (SC-005)
  A collection steered by scripts (pm.execution.setNextRequest by name, pm.execution.skipRequest,
  pm.iterationData and pm.info.iteration) is run with a CSV data file by Xenon's runner, then
  exported as a Postman collection and run by Newman (Postman's own runner, pinned in tools/newman)
  with the same data file against the same mock server. Both runs must send the same requests in the
  same order with the same bodies. Run with make newman (CI job "newman"); it needs Node.js and is
  not part of make bdd.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "flow" variable is set to "01JR0000000000000000000070"
    And the "dev" variable is set to "01JR0000000000000000000071"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JR0000000000000000000000
      name: Acme
      order:
        - flow
      """
    And the workspace file "acme/collections/flow/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{flow}}
      name: Flow
      order:
        - a.yaml
        - b.yaml
        - c.yaml
        - d.yaml
      """
    And the workspace file "acme/collections/flow/a.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JR0000000000000000000072
      name: A
      method: POST
      url: "{{api}}/a"
      body:
        mode: raw
        raw:
          language: text
          content: "{{user}}"
      scripts:
        postResponse: |
          if (pm.iterationData.get("user") === "u2") {
            pm.execution.setNextRequest("C");
          }
      """
    And the workspace file "acme/collections/flow/b.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JR0000000000000000000073
      name: B
      method: POST
      url: "{{api}}/b"
      body:
        mode: raw
        raw:
          language: text
          content: "{{user}}-{{iter}}"
      scripts:
        preRequest: pm.variables.set("iter", String(pm.info.iteration))
      """
    And the workspace file "acme/collections/flow/c.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JR0000000000000000000074
      name: C
      method: POST
      url: "{{api}}/c"
      body:
        mode: raw
        raw:
          language: text
          content: "{{user}}"
      scripts:
        preRequest: |
          if (pm.iterationData.get("user") === "u3") {
            pm.execution.skipRequest();
          }
      """
    And the workspace file "acme/collections/flow/d.yaml" contains
      """
      xenon: 1
      kind: request
      id: 01JR0000000000000000000075
      name: D
      method: POST
      url: "{{api}}/d"
      body:
        mode: raw
        raw:
          language: text
          content: "{{user}}"
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: unused
          value: "1"
          enabled: true
          description: ""
      """
    And the file "data/users.csv" exists with content
      """
      user
      u1
      u2
      u3
      """

  Scenario: Xenon's runner and Newman send the same sequence for a script-steered, data-driven run
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{flow}}", "config": { "type": "functional", "rows": [{ "request_id": "01JR0000000000000000000072", "alias_id": "", "selected": true }, { "request_id": "01JR0000000000000000000073", "alias_id": "", "selected": true }, { "request_id": "01JR0000000000000000000074", "alias_id": "", "selected": true }, { "request_id": "01JR0000000000000000000075", "alias_id": "", "selected": true }], "functional": { "iterations": 3, "data_file": "{{temp_root}}/data/users.csv", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 10000 ms
    And the mock API should have received the requests "POST /a, POST /b, POST /c, POST /d, POST /a, POST /c, POST /d, POST /a, POST /b, POST /d" in order
    When I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "{{flow}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/flow.postman_collection.json"
      }
      """
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "{{dev}}",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/dev.postman_environment.json"
      }
      """
    And I run Newman on the file "flow.postman_collection.json" with the environment file "dev.postman_environment.json" and the data file "data/users.csv" in the workspace folder "acme"
    Then the mock API should have received the requests "POST /a, POST /b, POST /c, POST /d, POST /a, POST /c, POST /d, POST /a, POST /b, POST /d, POST /a, POST /b, POST /c, POST /d, POST /a, POST /c, POST /d, POST /a, POST /b, POST /d" in order
    And the 1 and 4 "POST" requests to "/a" should have the same query, body and headers "Content-Type"
    And the 3 and 6 "POST" requests to "/a" should have the same query, body and headers "Content-Type"
    And the 1 and 3 "POST" requests to "/b" should have the same query, body and headers "Content-Type"
    And the 2 and 4 "POST" requests to "/b" should have the same query, body and headers "Content-Type"
    And the 2 "POST" request to "/b" should have the body equal to
      """
      u3-2
      """
    And the 2 and 5 "POST" requests to "/d" should have the same query, body and headers "Content-Type"
