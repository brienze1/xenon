# Feature Specification: Scripts, Tests and Console (Stage 4)

**Feature Branch**: `xen-004`

**Created**: 2026-10-09

**Status**: Draft

**Input**: User description: "Xenon Stage 4 — Pre-request and post-response scripts with Postman's pm API and execution order, tests and Test Results, the console. Full pm API, every library Postman's sandbox ships, legacy API shim. Source: docs/prd/xenon-mvp-prd.md 'Scripts', 'Console'."

## Context

Stage 4 of 7, the one the PRD says decides whether Postman can be dropped. Scripts were stored
since stage 1 (and imported in stage 3); now they run. Existing Postman scripts must run
**unchanged**: same `pm` API, same order, same libraries, including old collections written with
the legacy API. Scripts run in a sandbox with no file system or shell access. The console shows
every request (including those sent from scripts) and every script log and error.

## Clarifications

### Session 2026-10-09

- Q: How much of the `pm` API? → A: The full Postman `pm` API; runner-only calls (e.g.
  `setNextRequest`, `iterationData`) exist but do nothing, since the collection runner is a PRD
  non-goal.
- Q: Sandbox libraries? → A: Everything Postman's sandbox ships (`require()`-able libraries,
  Node-style modules and globals).
- Q: Legacy API (`postman.*`, `tests[...]`, `responseBody`, …)? → A: Supported through a shim.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Scripts run around every send, in Postman's order (Priority: P1)

Luis's collections have pre-request scripts that compute signatures and set tokens, and
post-response scripts that store values from responses. When he sends a request, the scripts of
its collection, folders and the request itself run in Postman's order, and their effects apply.

**Why this priority**: Without scripts, most of Luis's real requests cannot be sent (signing,
token chaining).

**Independent Test**: A collection with scripts at collection, two folder levels and request,
each logging its name and setting a variable; send and verify order, effects and the inherit
toggle.

**Acceptance Scenarios**:

1. **Given** scripts at collection, folder A, nested folder B and the request, **When** Luis
   sends, **Then** they run in this order: collection pre-request → A pre-request → B
   pre-request → request pre-request → request is sent → collection post-response → A
   post-response → B post-response → request post-response.
2. **Given** "Inherit parent scripts" is turned off on folder B, **When** a request in B is sent,
   **Then** the collection and A scripts are skipped (both phases) and B's and the request's run.
   **Given** it is off on the request, **Then** only the request's own scripts run.
3. **Given** a pre-request script changes the request (`pm.request.headers.upsert`,
   `pm.request.url.query.add`, `pm.request.body.update`, `pm.request.method`), **When** sent,
   **Then** the request goes out with those changes, and the saved request is not changed.
4. **Given** a top-level `const x` in the collection script, **Then** it is not visible in the
   request script (each script has its own scope); values pass through `pm.variables` and the
   other variable scopes.
5. **Given** a pre-request script throws or has a syntax error, **When** Luis sends, **Then** the
   request is not sent, and the error, the script that failed (collection/folder/request,
   pre-request) and the line and column are shown in the response area and the console.
6. **Given** a post-response script throws, **Then** the response is still shown, the error is
   reported with its location, and the remaining post-response scripts still run.
7. **Given** a script that runs longer than the time limit (default 10 s, including time spent
   waiting for its timers and requests), **Then** it is stopped and reported as timed out; a
   pre-request timeout stops the send.
8. **Given** a disabled script (imported with `disabled`), **Then** it is not run.

---

### User Story 2 - Tests and Test Results (Priority: P1)

Post-response scripts contain `pm.test(...)` with `pm.expect` and `pm.response.to.*` assertions.
After each send, the response's Test Results tab lists every test with pass/fail, and the tab
shows the count.

**Why this priority**: Tests are how Luis checks responses daily; the PRD lists them in the MVP.

**Independent Test**: A request with passing, failing, skipped, async and legacy tests; check the
Test Results tab and count.

**Acceptance Scenarios**:

1. **Given** `pm.test("status 200", () => pm.response.to.have.status(200))` and a 200 response,
   **Then** Test Results shows it as passed; with a 404 it fails with the message
   `expected response to have status code 200 but got 404`.
2. **Given** tests in collection, folder and request scripts, **Then** all are listed in run
   order, each labelled with the script it came from; the tab title shows e.g. "Test Results
   (3/4)".
3. **Given** `pm.test.skip("later")`, **Then** it is listed as skipped.
4. **Given** an async test `pm.test("x", (done) => { setTimeout(() => { pm.expect(1).to.eql(1);
   done(); }, 10) })`, **Then** it is recorded when `done` is called; `done(err)` fails it.
5. **Given** legacy `tests["has id"] = responseBody.includes("id")`, **Then** it appears in Test
   Results like any other test.
6. **Given** Luis filters Test Results by All / Passed / Failed / Skipped, **Then** only those
   tests are listed.
7. **Given** a pre-request script contains tests, **Then** they also appear in Test Results.

---

### User Story 3 - Variables, chained requests and the full pm API (Priority: P1)

Scripts read and write variables in every scope, send extra requests with `pm.sendRequest` (callback
or `await`), run other saved requests, read cookies and request/response details, and use the
Postman libraries (`CryptoJS`, `_`, `moment`, `uuid`, `require('ajv')`, …) — all exactly as in
Postman.

**Why this priority**: Token flows and signing depend on these; "unchanged scripts" is the PRD's
bar.

**Independent Test**: A script corpus taken from real-world Postman usage (signing with
CryptoJS, OAuth token fetch with `pm.sendRequest`, JSON-schema validation with ajv, date maths with
moment, lodash transforms) runs with the same outputs as in Postman.

**Acceptance Scenarios**:

1. **Given** `pm.environment.set("token", res.json().access_token)` in a post-response script,
   **Then** the active environment's current value is updated (XEN-002 rules: local, not written
   to the workspace file; a secret row stays secret and goes to the vault), and the next request
   using `{{token}}` sends it.
2. **Given** `pm.variables.set("x", 1)` in a pre-request script, **Then** `{{x}}` resolves for that
   send (local scope wins over all others) and `pm.variables.get("x")` returns it in the
   post-response scripts of the same send; it is gone for the next send.
3. **Given** `pm.collectionVariables.set`, `pm.globals.set`, `unset`, `clear`, `has`, `toObject`,
   `replaceIn`, **Then** each behaves as in Postman on the request's collection and the
   workspace globals.
4. **Given** `const res = await pm.sendRequest({ url: "{{base_url}}/token", method: "POST",
   … })` at top level (or the callback form), **Then** the request is sent, appears in the
   console, and its response is available to the script; Xenon waits for it before continuing.
5. **Given** `pm.execution.skipRequest()` in a pre-request script, **Then** the request is not
   sent and the response area says it was skipped by a script.
6. **Given** `pm.execution.runRequest("<request id>")`, **Then** that saved request runs with its
   own scripts, its variable changes merge back, and it appears in the console.
7. **Given** `pm.info.requestName`, `pm.info.eventName`, `pm.cookies.get("session")`,
   `pm.response.headers.get(...)`, `pm.response.responseTime`, **Then** they return Postman's
   values; runner-only APIs (`pm.execution.setNextRequest`, `pm.iterationData`,
   `postman.setNextRequest`) do nothing and return Postman's defaults (iteration 0 of 1).
8. **Given** scripts using `CryptoJS.HmacSHA256`, `_`, `moment()`, `require("uuid")`,
   `require("ajv")`, `require("xml2js")`, `require("cheerio")`, `require("csv-parse/lib/sync")`,
   `tv4`, `atob`/`btoa`, `Buffer`, `URL`, `TextEncoder`, `crypto.randomUUID()`, **Then** they
   produce the same results as in Postman.
9. **Given** a script calls an API Xenon cannot support (`pm.require`, `pm.datasets`,
   `pm.mock`, `pm.state`, `pm.message`) or a sandbox global it lacks, **Then** it fails with an
   error naming the API and saying it is not supported in Xenon.

---

### User Story 4 - The console (Priority: P1)

A bottom drawer (hidden by default, toggled from the footer and Cmd/Ctrl+Alt+C) logs every
request sent — including those from `pm.sendRequest` and `runRequest` — and every script log and
error, in order. Each request entry expands to show exactly what was sent and received after
variables were replaced. Secrets are masked.

**Why this priority**: Debugging scripts and chained requests is impossible without it; it is in
the PRD's MVP.

**Independent Test**: Send a request whose scripts log at each level and call `pm.sendRequest`;
inspect, filter, search and clear the console; check secret masking.

**Acceptance Scenarios**:

1. **Given** Luis sends a request, **Then** the console gets an entry with method, URL, status,
   time and size; expanding it shows the resolved request (URL, headers, body) and the response
   (headers, body).
2. **Given** scripts call `console.log/info/warn/error/debug` with strings and objects, **Then**
   entries appear in order between the request entries, with level, the originating script, and
   objects shown expandable.
3. **Given** a script error or test failure, **Then** an error entry appears with message and
   location.
4. **Given** a secret value appears anywhere in a console entry (URL, header, body, log text),
   **Then** it is shown as `•••••`.
5. **Given** the filter (All / Requests / Logs / Errors), search box and Clear button, **Then**
   they filter, search and clear entries; `console.clear()` in a script clears too.
6. **Given** the app restarts, **Then** the console starts empty (it is a session log).

---

### User Story 5 - Writing scripts comfortably (Priority: P2)

Scripts are edited in a code editor with JavaScript highlighting, `pm.*` autocomplete (with the
exact Postman API surface), variable-name completion inside `pm.*.get/set("…")`, inline syntax
errors, and a snippets menu with Postman's common snippets. Scripts tabs exist on requests,
folders and collections, with the "Inherit parent scripts" toggle on folders and requests.

**Why this priority**: Comfort for writing new scripts; imported ones already run without it.

**Independent Test**: Open the Scripts tab, type `pm.re`, accept completions, insert a snippet,
introduce a syntax error.

**Acceptance Scenarios**:

1. **Given** Luis types `pm.` in a script, **Then** completions list the pm API members with
   short descriptions; `pm.response.to.have.` lists the assertion helpers.
2. **Given** he types `pm.environment.get("` , **Then** names of the active environment's
   variables are suggested.
3. **Given** a syntax error, **Then** it is underlined with the message before sending.
4. **Given** the snippets menu, **Then** inserting e.g. "Status code is 200" adds Postman's
   snippet text at the cursor.
5. **Given** collection and folder settings, **Then** they have Pre-request and Post-response
   editors; folders and requests show the "Inherit parent scripts" toggle (default on).

---

### Edge Cases

- **Async work at script end**: the script finishes when its top-level code, pending timers,
  pending `pm.sendRequest` calls and cookie-jar calls have completed (Postman behaviour);
  promises not tied to these are not awaited.
- **Unhandled promise rejection** inside a script is reported as a script error.
- **Async test with `async () => {}`** (no `done`): a failure after an `await` is reported as a
  script error rather than a test failure — Postman behaviour, documented in the editor hint.
- **Runaway memory** (e.g., an endless array): the script is stopped with an error before the app
  becomes unresponsive.
- **Deep recursion**: stopped with a stack-overflow error naming the script.
- **Non-string variable values** set from scripts: kept as-is within the same send; stored as
  strings (objects JSON-encoded) when written to environment/collection/globals.
- **`pm.environment.set` with no active environment**: does nothing and logs a console warning.
- **Setting a variable that does not exist** in the target scope: creates the row (empty initial
  value in the file, value as local current value), as Postman does.
- **Scripts on examples**: examples (stage 3) have no scripts.
- **Localisation APIs** (`Intl`, `toLocaleString`): number and date formatting for the `en-US`
  locale behave as in Postman; other locales fall back to `en-US` with a console warning.
- **Unsupported runtime features** (streams, compression streams, `SharedArrayBuffer`,
  `Atomics`, `WeakRef`, async generators): fail with a clear error naming the feature.
- **`pm.cookies.jar()`** uses the workspace cookie jar delivered in stage 5; until then it fails
  with "cookie jar not available yet" (both stages ship together in the MVP).
- **`{{vault:name}}` and `pm.vault`**: Postman's vault maps to Xenon's workspace-level secrets
  (globals secret rows): `pm.vault.get/set/unset` read and write them; `{{vault:name}}` resolves
  to the same value.
- **History** keeps the request as written (unchanged by pre-request scripts); the console shows
  what was actually sent.

## Requirements *(mandatory)*

### Functional Requirements

**Execution**

- **FR-001**: Collections, folders and requests MUST each have a pre-request and a post-response
  script; folders and requests MUST have "Inherit parent scripts" (default on).
- **FR-002**: A send MUST run scripts in Postman's order (collection → folders outer to inner →
  request, for each phase), skipping everything above a level whose inherit toggle is off, and
  skipping disabled scripts.
- **FR-003**: Each script MUST run in its own scope; data passes only through `pm` scopes. The
  local scope (`pm.variables`) lives for one send (pre + post scripts of that send).
- **FR-004**: Changes to `pm.request` in pre-request scripts MUST apply to the sent request only,
  never to the stored request or History entry.
- **FR-005**: A pre-request script error, timeout or `skipRequest` MUST stop the send (no
  network call); a post-response script error MUST NOT hide the response and MUST NOT stop the
  remaining post-response scripts.
- **FR-006**: Scripts MUST run in a sandbox with no access to the file system, processes, the
  network (except `pm.sendRequest`/`runRequest`) or the app's internals; a script MUST be stopped
  after its time limit (default 10 s, configurable 1–60 s in Settings), on excessive memory use,
  and on stack overflow, with an error naming the script.
- **FR-007**: Errors MUST report the script (scope + phase), message and line/column.

**pm API**

- **FR-008**: The full Postman `pm` API MUST be provided with Postman's behaviour: `pm.variables`,
  `pm.environment` (incl. `name`), `pm.collectionVariables`, `pm.globals`, `pm.iterationData`
  (empty), `pm.request` (Request/Url/PropertyList/RequestBody API), `pm.response` (code, status,
  headers, responseTime, responseSize, cookies, json, jsonp, text, reason, size, mime,
  contentInfo, dataURI, encoding, `to.*`), `pm.info`, `pm.cookies` (+ `jar()` per stage 5),
  `pm.sendRequest` (callback and promise), `pm.execution` (`skipRequest`, `runRequest`,
  `location`, `setNextRequest` no-op), `pm.test` (sync, `done` async, `skip`, `index`),
  `pm.expect`, `pm.vault` (mapped per Edge Cases), `pm.visualizer` (no-op).
- **FR-009**: The Postman response assertion helpers (`pm.response.to.be.ok/success/error/…`,
  `to.have.status/header/body/jsonBody/jsonSchema/responseTime/responseSize`, status-name
  properties) MUST be provided with Postman's semantics and messages.
- **FR-010**: Variable writes from scripts MUST follow XEN-002: current values (local), secret
  rows stay secret, new rows created per Edge Cases, and `variables:changed` emitted.
- **FR-011**: `pm.sendRequest` and `runRequest` MUST resolve `{{variables}}`, honour the
  workspace's settings and appear in the console; they MUST NOT be added to History.
- **FR-012**: Unsupported APIs (`pm.require`, `pm.datasets`, `pm.mock`, `pm.state`, `pm.message`)
  MUST throw an error naming the API.

**Libraries and legacy**

- **FR-013**: The sandbox MUST provide every library Postman's sandbox ships, at Postman's versions,
  via `require()` and the same globals: ajv, atob, btoa, chai, cheerio, crypto-js, csv-parse
  (sync), lodash (and legacy lodash 3 as the legacy `_` global), moment, postman-collection
  (its commonly used classes), tv4, uuid, xml2js, backbone, and Node-style modules path, assert,
  buffer, util, url, punycode, querystring, string_decoder, stream, timers, events; plus globals
  `atob`, `btoa`, `Buffer`, `URL`, `URLSearchParams`, `TextEncoder`/`TextDecoder`,
  `structuredClone`, `queueMicrotask`, `setTimeout`/`setInterval`/`setImmediate` (+ clear),
  `crypto` (`getRandomValues`, `randomUUID`, `subtle.digest`/`sign`/`verify` for HMAC/SHA),
  `AbortController`, `Event`/`EventTarget`, `Blob`, `console`.
- **FR-014**: The legacy API MUST be supported: globals `tests`, `globals`, `environment`, `data`,
  `request`, `responseCookies`, `responseHeaders`, `responseTime`, `responseCode`,
  `responseBody`, `iteration`, `postman` (`set/get/clear EnvironmentVariable(s)`,
  `set/get/clear GlobalVariable(s)`, `setNextRequest` no-op, `getResponseCookie`,
  `getResponseHeader`), `CryptoJS`, `tv4`, `xml2Json`, `cheerio`, `_` (lodash 3); `tests[...]`
  results are collected at script end.

**Tests and console**

- **FR-015**: Test results (name, passed/failed/skipped, error message, originating script) MUST
  be shown in a Test Results tab of the response with a count, filterable by status, for every
  send, including tests from pre-request scripts and legacy `tests[...]`.
- **FR-016**: A console drawer (hidden by default; footer button and Cmd/Ctrl+Alt+C) MUST log, in
  order: every request sent (main, `pm.sendRequest`, `runRequest`) with method, URL, status,
  time, size and expandable resolved request/response; every `console.*` call (levels log,
  info, warn, error, debug; objects expandable; `console.clear`); script errors and test
  failures — each with its originating script.
- **FR-017**: The console MUST support filtering (All, Requests, Logs, Errors), search and Clear;
  it is per app session, capped at the most recent 5,000 entries, with bodies over 1 MB opened on
  demand.
- **FR-018**: Every occurrence of a secret value in console entries MUST be masked `•••••`
  (XEN-002 FR-020/023).

**Editing**

- **FR-019**: Script editors MUST provide JavaScript highlighting, inline syntax errors, `pm.*`
  autocomplete matching FR-008/FR-009 (with descriptions), variable-name completion inside
  `pm.<scope>.get/set/unset/has("…")`, and a snippets menu with Postman's standard snippets.
- **FR-020**: Collection and folder settings MUST have Pre-request and Post-response editors;
  folders and requests MUST show the inherit toggle.

### Key Entities

- **Script**: text, phase (pre-request | post-response), owner (collection | folder | request),
  disabled flag; plus owner's inherit-parent-scripts flag.
- **Script run**: one script's execution: owner, phase, logs, test results, error, duration.
- **Send execution**: the chain for one send: local variables, pre runs, sent request (as
  modified), response, post runs, nested requests (sendRequest/runRequest).
- **Test result**: name, status (passed/failed/skipped), error message, script, index.
- **Console entry**: kind (request | log | error | test-failure), level, timestamp, originating
  script, payload (request/response details or log arguments), masked.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A corpus of at least 200 real-world Postman scripts (signing, token chaining, schema
  validation, legacy-API scripts) produces identical variable changes, sent requests and test
  results in Xenon and Postman for ≥ 98% of scripts; every difference is listed with its cause.
- **SC-002**: Script overhead for a send with four typical scripts (one per level) is under
  100 ms in total.
- **SC-003**: A script stuck in an endless loop is stopped within its time limit + 0.5 s, and the
  app stays responsive throughout.
- **SC-004**: 0 secret values appear in console entries for a scenario that logs, sends and
  echoes secrets.
- **SC-005**: Every API listed in FR-008, FR-009, FR-013 and FR-014 is covered by an automated
  parity check.
- **SC-006**: Luis's own collections (imported in stage 3) run with their scripts unchanged, with
  any failure traced to an item in the stage 3 pm-usage report.

## Assumptions

- Postman's sandbox (postman-sandbox 6.x and its pinned library versions) is the parity
  reference.
- The collection runner is a PRD non-goal; runner-only APIs are no-ops.
- `pm.visualizer` is not rendered in the MVP (no Visualize tab).
- `pm.require` / Package Library depend on Postman's cloud and cannot be supported.
- Scripts are the user's own (or imported from collections the user chose to import); the
  sandbox still treats them as untrusted code.
