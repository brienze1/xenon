#language: en
#utf-8

@all @xen_017 @runner @run_performance
Feature: Performance runs load an endpoint with virtual users and report statistics
  A performance run starts virtual users (1–100) for a duration (1 s – 60 min) following a load
  profile (Fixed, Ramp up, Spike, Peak). Each user loops over the selected rows with its own
  variables and cookies, running scripts, but nothing is persisted, logged to the console or
  recorded in History; setNextRequest is ignored (one notice) and pm.test results are not reported.
  Live metrics arrive at least once per second. The result holds count, average, p90/p95/p99, min,
  max, throughput, error rate (not sent or status >= 400) and status counts, overall and per request,
  plus a per-second series. Durations here are a few real seconds against a mock with 50 ms latency.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "shop" variable is set to "01JR0000000000000000000001"
    And the "orders" variable is set to "01JR0000000000000000000002"
    And the "login" variable is set to "01JR0000000000000000000003"
    And the "create" variable is set to "01JR0000000000000000000004"
    And the "get" variable is set to "01JR0000000000000000000005"
    And the "logout" variable is set to "01JR0000000000000000000006"
    And the "dev" variable is set to "01JR0000000000000000000010"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JR0000000000000000000000
      name: Acme
      order:
        - shop
      """
    And the workspace file "acme/collections/shop/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{shop}}
      name: Shop
      order:
        - login.yaml
        - orders
        - logout.yaml
      auth:
        type: bearer
        bearer:
          token: "{{token}}"
      headers:
        - key: X-Suite
          value: runner
          enabled: true
          description: ""
      scripts:
        preRequest: console.log("shop pre")
      """
    And the workspace file "acme/collections/shop/orders/folder.yaml" contains
      """
      xenon: 1
      kind: folder
      id: {{orders}}
      name: Orders
      order:
        - create.yaml
        - get.yaml
      """
    And the workspace file "acme/collections/shop/login.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{login}}
      name: Login
      method: POST
      url: "{{api}}/login"
      auth:
        type: inherit
      scripts:
        postResponse: pm.test("login ok", function () { pm.response.to.have.status(200); });
      """
    And the workspace file "acme/collections/shop/orders/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: POST
      url: "{{api}}/orders"
      auth:
        type: inherit
      scripts:
        postResponse: pm.test("created", function () { pm.response.to.have.status(201); });
      """
    And the workspace file "acme/collections/shop/orders/get.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{get}}
      name: Get
      method: GET
      url: "{{api}}/orders/1"
      auth:
        type: inherit
      """
    And the workspace file "acme/collections/shop/logout.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{logout}}
      name: Logout
      method: POST
      url: "{{api}}/logout"
      auth:
        type: inherit
      """
    And the workspace file "acme/environments/dev.yaml" contains
      """
      xenon: 1
      kind: environment
      id: {{dev}}
      name: dev
      variables:
        - key: token
          value: t-123
          enabled: true
          description: ""
      """
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" contains
      """
      { "version": 1, "activeEnvironmentId": "{{dev}}", "current": {} }
      """
    And the 1 "POST" request to "/orders" returns status 201 with the following response
      """
      {"id":"o-1"}
      """
    And the "load" variable is set to "01JR0000000000000000000060"
    And the "p" variable is set to "01JR0000000000000000000061"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JR0000000000000000000000
      name: Acme
      order:
        - shop
        - load
      """
    And the workspace file "acme/collections/load/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{load}}
      name: Load
      order:
        - p.yaml
      """
    And the workspace file "acme/collections/load/p.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{p}}
      name: P
      method: GET
      url: "{{api}}/p"
      """
    And the mock API echoes every request to "/p"
    And the "GET" requests to "/p" respond after 50ms

  @validation
  Scenario Outline: Performance settings out of range are refused - <case>
    When I call "RunnerService.Start" with
      """
      <payload>
      """
    Then the call should fail with code "RUN-04400"
    And the "GET" request to "/p" should have been received 0 times

    Examples:
      | case                     | payload |
      | no virtual users         | { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 0, "duration_seconds": 3, "data_file": "", "profile": { "kind": "fixed", "initial_users": 0, "ramp_seconds": 0, "base_users": 0, "spike_seconds": 0, "peak_seconds": 0 } }, "environment_id": "" } } |
      | 101 virtual users        | { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 101, "duration_seconds": 3, "data_file": "", "profile": { "kind": "fixed", "initial_users": 0, "ramp_seconds": 0, "base_users": 0, "spike_seconds": 0, "peak_seconds": 0 } }, "environment_id": "" } } |
      | zero duration            | { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 2, "duration_seconds": 0, "data_file": "", "profile": { "kind": "fixed", "initial_users": 0, "ramp_seconds": 0, "base_users": 0, "spike_seconds": 0, "peak_seconds": 0 } }, "environment_id": "" } } |
      | over 60 minutes          | { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 2, "duration_seconds": 3601, "data_file": "", "profile": { "kind": "fixed", "initial_users": 0, "ramp_seconds": 0, "base_users": 0, "spike_seconds": 0, "peak_seconds": 0 } }, "environment_id": "" } } |
      | ramp starts above max    | { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 4, "duration_seconds": 3, "data_file": "", "profile": { "kind": "ramp_up", "initial_users": 5, "ramp_seconds": 1, "base_users": 0, "spike_seconds": 0, "peak_seconds": 0 } }, "environment_id": "" } } |
      | ramp longer than run     | { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 4, "duration_seconds": 3, "data_file": "", "profile": { "kind": "ramp_up", "initial_users": 1, "ramp_seconds": 4, "base_users": 0, "spike_seconds": 0, "peak_seconds": 0 } }, "environment_id": "" } } |
      | spike longer than run    | { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 4, "duration_seconds": 3, "data_file": "", "profile": { "kind": "spike", "initial_users": 0, "ramp_seconds": 0, "base_users": 1, "spike_seconds": 4, "peak_seconds": 0 } }, "environment_id": "" } } |
      | peak fills the whole run | { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 4, "duration_seconds": 3, "data_file": "", "profile": { "kind": "peak", "initial_users": 0, "ramp_seconds": 0, "base_users": 1, "spike_seconds": 0, "peak_seconds": 3 } }, "environment_id": "" } } |

  @fixed
  Scenario: A fixed load reports response-time statistics, throughput, errors and a per-second series
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 2, "duration_seconds": 3, "data_file": "", "profile": { "kind": "fixed", "initial_users": 0, "ramp_seconds": 0, "base_users": 0, "spike_seconds": 0, "peak_seconds": 0 } }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 10000 ms
    And the "runner:progress" events should have been emitted at most 1500 ms apart
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "config.type" equal to "performance"
    And the response should contain the field "status" equal to "completed"
    And the response should contain the field "performance.overall.name" equal to ""
    And the response field "performance.overall.count" should be between 20 and 130
    And the response field "performance.overall.avg_ms" should be between 45 and 200
    And the response field "performance.overall.min_ms" should be between 45 and 200
    And the response field "performance.overall.p90_ms" should be between 45 and 300
    And the response field "performance.overall.p95_ms" should be between 45 and 300
    And the response field "performance.overall.p99_ms" should be between 45 and 400
    And the response field "performance.overall.max_ms" should be between 45 and 1000
    And the response field "performance.overall.throughput" should be between 6 and 45
    And the response field "performance.overall.error_rate" should be between 0 and 0
    And the response field "performance.overall.status_counts.200" should be between 20 and 130
    And the response should contain the field "performance.per_request" array length equal to 1
    And the response should contain the field "performance.per_request.0.request_id" equal to "{{p}}"
    And the response should contain the field "performance.per_request.0.name" equal to "P"
    And the response should contain the field "performance.series" array length equal to 3
    And the response should contain the field "performance.series.0.second" equal to "0"
    And the response should contain the field "performance.series.2.active_users" equal to "2"
    And the response should contain the field "rows" array length equal to 0

  @errors
  Scenario: Requests that fail or get a status of 400 or more count as errors
    Given the mock API echoes every request to "/p" with status 500
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 1, "duration_seconds": 2, "data_file": "", "profile": { "kind": "fixed", "initial_users": 0, "ramp_seconds": 0, "base_users": 0, "spike_seconds": 0, "peak_seconds": 0 } }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 10000 ms
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response field "performance.overall.error_rate" should be between 1 and 1
    And the response field "performance.overall.status_counts.500" should be between 1 and 100
    And the response field "performance.series.0.errors" should be between 1 and 100

  @profiles
  Scenario: Ramp up grows the users linearly to the maximum, then holds
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 4, "duration_seconds": 4, "data_file": "", "profile": { "kind": "ramp_up", "initial_users": 1, "ramp_seconds": 2, "base_users": 0, "spike_seconds": 0, "peak_seconds": 0 } }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 10000 ms
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response field "performance.series.0.active_users" should be between 1 and 2
    And the response field "performance.series.1.active_users" should be between 2 and 4
    And the response should contain the field "performance.series.2.active_users" equal to "4"
    And the response should contain the field "performance.series.3.active_users" equal to "4"

  @profiles
  Scenario: A spike holds the base load, jumps to the maximum in the middle of the run, and returns
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 3, "duration_seconds": 3, "data_file": "", "profile": { "kind": "spike", "initial_users": 0, "ramp_seconds": 0, "base_users": 1, "spike_seconds": 1, "peak_seconds": 0 } }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 10000 ms
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "performance.series.0.active_users" equal to "1"
    And the response should contain the field "performance.series.1.active_users" equal to "3"
    And the response should contain the field "performance.series.2.active_users" equal to "1"

  @profiles
  Scenario: A peak ramps up, holds the maximum for the peak, and ramps back down
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 3, "duration_seconds": 3, "data_file": "", "profile": { "kind": "peak", "initial_users": 0, "ramp_seconds": 0, "base_users": 1, "spike_seconds": 0, "peak_seconds": 1 } }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 10000 ms
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response field "performance.series.0.active_users" should be between 1 and 3
    And the response should contain the field "performance.series.1.active_users" equal to "3"
    And the response field "performance.series.2.active_users" should be between 1 and 3

  @scripts_and_isolation
  Scenario: Scripts run for every request, but nothing is persisted, logged or recorded
    Given the workspace file "acme/collections/load/p.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{p}}
      name: P
      method: GET
      url: "{{api}}/p"
      scripts:
        preRequest: |
          pm.request.headers.add({ key: "X-Pre", value: "yes" })
          pm.environment.set("perf_marker", "perf-value-1")
          pm.execution.setNextRequest("P")
        postResponse: pm.test("echoed", function () { pm.response.to.have.status(200); });
      """
    And the requests to "/p" respond with the header "Set-Cookie" equal to "perf=1; Path=/"
    And I call "ConsoleService.Clear"
    And the emitted events are forgotten
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 1, "duration_seconds": 1, "data_file": "", "profile": { "kind": "fixed", "initial_users": 0, "ramp_seconds": 0, "base_users": 0, "spike_seconds": 0, "peak_seconds": 0 } }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 10000 ms
    And the 1 "GET" request to "/p" should have the header "X-Pre" equal to "yes"
    And the event "console:entries" should have been emitted 0 times
    And the event "history:changed" should have been emitted 0 times
    And the event "variables:changed" should have been emitted 0 times
    And the app data file "workspaces/{{Acme.local_id}}/variables.json" should not contain "perf-value-1"
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response should contain the field "notices" with the items "pm.execution.setNextRequest is ignored in performance runs"
    And the response should contain the field "rows" array length equal to 0
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0

  @data_wraps
  Scenario: Each virtual user takes data rows in turn and wraps around
    Given the workspace file "acme/collections/load/p.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{p}}
      name: P
      method: GET
      url: "{{api}}/p?q={{q}}"
      """
    And the file "data/q.csv" exists with content
      """
      q
      first
      second
      """
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 1, "duration_seconds": 1, "data_file": "{{temp_root}}/data/q.csv", "profile": { "kind": "fixed", "initial_users": 0, "ramp_seconds": 0, "base_users": 0, "spike_seconds": 0, "peak_seconds": 0 } }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 10000 ms
    And the 1 "GET" request to "/p" should have the query "q" equal to "first"
    And the 2 "GET" request to "/p" should have the query "q" equal to "second"
    And the 3 "GET" request to "/p" should have the query "q" equal to "first"

  @stop_only
  Scenario: A performance run cannot be paused, only stopped, and keeps the stats so far
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 2, "duration_seconds": 30, "data_file": "", "profile": { "kind": "fixed", "initial_users": 0, "ramp_seconds": 0, "base_users": 0, "spike_seconds": 0, "peak_seconds": 0 } }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    When I wait 1200 ms
    And I call "RunnerService.Pause" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should fail with code "RUN-04400"
    When I call "RunnerService.Stop" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 2000 ms
    And the last "runner:finished" event should contain the field "status" equal to "stopped"
    When I call "RunnerService.Get" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}" }
      """
    Then the call should succeed
    And the response field "performance.overall.count" should be between 1 and 1000

  @junit_refused
  Scenario: A performance run exports as JSON only
    When I call "RunnerService.Start" with
      """
      { "local_id": "{{Acme.local_id}}", "item_id": "{{load}}", "config": { "type": "performance", "rows": [{ "request_id": "{{p}}", "alias_id": "", "selected": true }], "functional": { "iterations": 1, "data_file": "", "delay_ms": 0, "persist_responses": false, "disable_logs": false, "stop_on_error": false, "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false }, "performance": { "virtual_users": 1, "duration_seconds": 1, "data_file": "", "profile": { "kind": "fixed", "initial_users": 0, "ramp_seconds": 0, "base_users": 0, "spike_seconds": 0, "peak_seconds": 0 } }, "environment_id": "" } }
      """
    Then the call should succeed
    And I save the response field "id" as "run"
    And a "runner:finished" event with the field "run_id" equal to "{{run}}" should have been emitted within 10000 ms
    When I call "RunnerService.Export" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}", "format": "junit", "path": "{{temp_root}}/out/perf.xml" }
      """
    Then the call should fail with code "RUN-04400"
    When I call "RunnerService.Export" with
      """
      { "local_id": "{{Acme.local_id}}", "run_id": "{{run}}", "format": "json", "path": "{{temp_root}}/out/perf.json" }
      """
    Then the call should succeed
    And the file "out/perf.json" should contain the field "performance.series" array length equal to 1
