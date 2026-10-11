---
description: "Task list for XEN-004 — Scripts, Tests and Console (Stage 4)"
---

# Tasks: Scripts, Tests and Console (Stage 4)

**Input**: `specs/XEN-004-scripts-and-console/` · **Prerequisite**: XEN-001..003 implemented.

**Tests**: MANDATORY (constitution V). Feature files tagged `@all @xen_004 @<area>`; the parity
corpus is written before the shims and drives them. The sandbox surface is
[contracts/sandbox-api.md](contracts/sandbox-api.md).

**Rules**: backend per architecture skills (sandbox in `internal/integration/webservice/sandbox/`
per the plan's Structure Decision); frontend per `frontend-architecture`.

US1 execution · US2 tests · US3 pm API & libraries · US4 console · US5 editor

---

## Phase 1: Setup

- [x] T001 Add `github.com/dop251/goja@v0.0.0-20261009133625-a4bedf5f6000` and `github.com/dop251/goja_nodejs@v0.0.0-20260918173711-b481721df8a2` to `go.mod`
- [x] T002 [P] `tools/sandbox-assets/`: `package.json` with exact versions from research R5/R6 (chai 4.4.1, lodash 4.18.1, lodash 3.10.2, moment 2.30.1, crypto-js 3.3.0, ajv 6.12.5, tv4 1.3.0, xml2js 0.6.2, cheerio 0.22.0, csv-parse 1.2.4, backbone 1.6.0, assert 2.0.0, buffer 6.0.3, events, path-browserify, punycode, querystring-es3, string_decoder, stream-browserify, timers-browserify, url, util, FormatJS Intl packages + en-US data, esbuild); `build.mjs` producing vendored files and bundles into `internal/integration/webservice/sandbox/assets/` with `MANIFEST.json` (name, version, sha256)
- [x] T003 [P] `tools/parity/`: `package.json` (`postman-sandbox@6.7.4`), `oracle.mjs` that reads a corpus case (script, phase, request, response, variable scopes, cookies) and writes the outcome JSON (variable mutations per scope, sent request after mutations, tests, console calls, errors)
- [x] T004 [P] Makefile targets `sandbox-assets` (rebuild + drift check against MANIFEST) and `parity` (run oracle over the corpus then `go test ./test/integration/parity/...`); CI job runs both on all 3 OSes

---

## Phase 2: Foundational

- [x] T005 [P] Enums + entities from data-model.md in `internal/domain/` (script_phase, script_owner_kind, script_error_kind, test_status, console_entry_kind, console_level, send_outcome; script_ref, send_execution, script_run, script_error, test_result, console_entry, nested_request, app_settings)
- [x] T006 [P] Pure use cases `collect_script_chain.go` (order, inherit cut, disabled) and `mask_secrets.go` with table tests
- [x] T007 Parity corpus `test/integration/resources/scripts/corpus/`: ≥ 1 case per row of sandbox-api.md §1–§5 plus ≥ 80 real-world scripts from public Postman collections (licence noted per file) — signing, OAuth token chains, ajv/tv4 schema checks, moment date maths, lodash transforms, xml2js/cheerio parsing, legacy-API test scripts; run `make parity` oracle half to produce expected outputs
- [x] T008 Sandbox core `internal/integration/webservice/sandbox/runtime.go` + `limits.go`: fresh runtime per script, wrapper `;(async()=>{…})()`, goja_nodejs event loop with completion = wrapper settled + no pending timers/sendRequest, `SetMaxCallStackSize`, interrupt timeout, memory guard (research R3), unhandled-rejection tracker, error mapping to `ScriptError` with line/column; sandbox objects built explicitly (no reflection-exposed Go structs) and a unit test asserting the global object's keys equal the sandbox-api.md allowlist; `require_registry.go` with embedded assets and lazy globals; application port `run_script.go` + adapter `script_runner.go`; unit tests for limits
- [x] T009 Go parity test `test/integration/parity/parity_test.go`: runs each corpus case through the sandbox with fixture-backed pm services and diffs with the oracle output (red until shims exist)

**Checkpoint**: an empty sandbox runs plain JS with limits; parity test runs (red).

**Phase 1–2 notes (implemented)**:
- Parity is red by design until T022: `make parity` = `parity-oracle` (regenerates
  `corpus/expected/` with postman-sandbox 6.7.4 under `TZ=UTC` and fails if it drifts from the
  committed outcomes) + `parity-diff` (`go test -tags parity ./test/integration/parity/...`, 2/249
  cases match today). The `parity` build tag keeps the diff out of `go test ./...` and
  `make validate`; CI runs `sandbox-assets` and `parity-oracle` as blocking steps and `parity-diff`
  with `continue-on-error` — T022 removes that flag.
- Corpus: 249 cases — 136 hand-written (every row of sandbox-api.md §1–§6) and 113 real-world
  scripts from 77 permissively licensed public repositories (MIT/Apache-2.0/BSD; repository,
  path, blob URL and licence in each case's `source`). Long token-like literals, credentials,
  e-mails and secret-named variable values are redacted to `redacted-value`; scripts whose
  outcome depends on time or randomness were excluded (oracle run three times, outcomes stable).
  `pm.cookies.jar()` is not in the corpus (stage 5 contract).
- Sandbox assets: Node builtins follow browserify 16's ranges (events 2.1.0, path-browserify
  0.0.1, punycode 1.4.1, querystring-es3 0.2.1, stream-browserify 2.0.2, string_decoder 1.3.0,
  timers-browserify 1.4.2, url 0.11.4, util 0.10.4; `assert` bundles its own util 0.12.5); FormatJS
  intl-getcanonicallocales 3.2.12, intl-locale 5.3.12, intl-pluralrules 6.3.15, intl-numberformat
  9.4.3, intl-datetimeformat 7.8.0 (+ en data, all time zones); esbuild 0.28.2; `uuid` is
  postman-sandbox's vendored file. `atob`/`btoa`/`postman-collection` are Go/shim-backed (T020/T021).
- The application port is `usecase.RunScript`; the integration adapter follows the webservice
  naming rule (`adapter.ScriptRunnerWebservice` in `script_runner_webservice.go`). A minimal
  in-sandbox `console` records calls until T020 routes them to the console sink. The global
  allowlist test keeps a `pendingGlobals` list that T020/T021 must empty.

---

## Phase 3: User Story 1 — Execution pipeline (P1) 🎯 MVP

- [x] T010 [P] [US1] BDD `test/integration/features/scripts/execution_order.feature` and `inheritance.feature`: order across collection/folder/nested folder/request for both phases (via console log entries); inherit off on folder and on request; disabled scripts skipped; separate scopes (`const` not shared)
- [x] T011 [P] [US1] BDD `test/integration/features/scripts/request_mutation.feature`: headers/query/body/method changes reach the ApiMock; stored request and History unchanged; variables set in pre-request resolve in the same send
- [x] T012 [P] [US1] BDD `test/integration/features/scripts/errors_and_limits.feature`: pre-request throw/syntax error → `outcome: stopped_by_script_error`, no request reaches the mock, error has script + line/column; post-response throw → response shown, later post scripts run; timeout (setting 1000 ms) on `while(true){}` and on a never-resolving `setTimeout` chain; stack overflow; memory guard; unhandled rejection; `skipRequest`
- [x] T013 [US1] Run T010–T012 → **red**
- [x] T014 [US1] Extend `internal/application/service/send_request.go` with the pipeline (data-model state machine), local scope plugged into the XEN-002 resolver, nested send support flagging History suppression (nested sends inherit the main send's confirm-before-send acknowledgement; missing/locked secrets in nested sends reject with an error); History entries gain `outcome` (sent / skipped_by_script / stopped_by_script_error / cancelled / error); `SendResult` additions (bindings §1); controllers/DTOs updated → **green**
- [x] T015 [US1] Settings: `app.json` `settings.scriptTimeoutMs`, `SettingsService` (use cases, controller, binding), `features/settings` UI field; BDD `features/scripts/settings_timeout.feature`

**Phase 3 notes (implemented)**:
- The request's own scripts and `inheritScripts` come from its saved file (the request DTO does
  not carry scripts yet — T031 adds them with the Scripts panel); an unsaved request has none.
  `inheritScripts` is read from collection/folder/request files (default true).
- Minimal pm surface for US1 over JS shims in `sandbox/shims/` (`property_list.js`, `url.js`,
  `pm_request.js`, `pm_response.js`, `pm.js`) and a Go bridge: `pm.variables` (has/get/set/unset/
  clear/replaceIn/toObject), `pm.request` (Url, header PropertyList, RequestBody), 
  `pm.execution.skipRequest`, `pm.sendRequest` (callback + promise). T020/T021 extend these.
- Script requests: `usecase.SendScriptRequest` is implemented in the send service
  (`script_requests.go`); they resolve with the send's local scope at call time, inherit the main
  send's confirmation, reject missing/locked secrets with "The request uses secrets that have no
  value on this machine or are locked: <names>", and never reach History
  (`features/scripts/nested_send_preflight.feature`).
- Post-response scripts run only when the request went out and got a response (outcome `sent`).
- A minimal `ConsoleService` (`List`, `Clear`; in-memory 5,000-entry buffer; log and error
  entries masked at creation) exists so the US1 features can assert order; T027 adds request
  entries, `Get`, the `console:entries` publisher and `console.clear()` wiring.
- Settings: `SettingsService` (`Get`/`Save`, `SCR-01400`), `app.json` `settings.scriptTimeoutMs`
  (out-of-range file values fall back to 10000), Settings dialog opened from the top-bar gear.

---

## Phase 4: User Story 3 — pm API, libraries, legacy (P1)

- [x] T016 [P] [US3] BDD `test/integration/features/scripts/variables_from_scripts.feature`: every pm scope method; writes become current values; secret rows stay in the vault; new rows created with empty initial value; no active env → warning; non-string values stringified on persist; `variables:changed` emitted
- [x] T017 [P] [US3] BDD `test/integration/features/scripts/send_request.feature` and `run_request_skip.feature`: callback and `await` forms, `{{vars}}` resolved, nested entries in console, not in History; `runRequest` runs target scripts and merges variables; `skipRequest`
- [x] T018 [P] [US3] BDD `test/integration/features/scripts/libraries.feature`, `legacy_api.feature`, `unsupported_apis.feature`, `vault_and_cookies.feature`: each require-able module and global loads and gives the corpus result; legacy globals map to pm; unsupported APIs throw named errors; `pm.vault` + `{{vault:x}}` on globals secrets; `pm.cookies` from response; `pm.cookies.jar()` per stage 5 contract
- [x] T019 [US3] Run T016–T018 → **red**
- [x] T020 [US3] Go pm services in `sandbox/`: `pm_variables.go` (calling XEN-002 use cases), `pm_send_request.go`, `pm_execution.go`, `pm_vault.go` (+ resolver support for `{{vault:name}}` in `internal/application/usecase/resolve_text.go`), `pm_cookies.go`, `pm_info.go`, globals (`globals_crypto.go`, `globals_encoding.go`, `globals_url.go`), `console_sink.go`
- [x] T021 [US3] JS shims `sandbox/shims/`: `property_list.js`, `url.js`, `pm_request.js`, `pm_response.js`, `legacy.js`, `intl_install.js`; `postman-collection` module re-exporting shim classes
- [x] T022 [US3] Run `make sandbox-assets`; wire every module in sandbox-api.md §4 into `require_registry.go`; run US3 features and `make parity` → **green** (≥ 98% corpus match; remaining differences listed in `test/integration/parity/KNOWN_DIFFERENCES.md` with cause)

**Phase 4 notes (implemented)**:
- BDD: `variables_from_scripts`, `send_request`, `run_request_skip`, `libraries`, `legacy_api`,
  `unsupported_apis`, `vault_and_cookies` (74 scenarios; red run 30 passed / 44 failed, recorded in
  the test commit). Library/legacy expectations were taken from postman-sandbox 6.7.4 itself.
  One new generic step: app data file field equal to a doc string.
- Shims (`sandbox/shims/`) port postman-collection 5.3.1's object model (`sdk_util`,
  `property_list`, `variable_scope`, `url`, `pm_request`, `pm_response`), chai-postman 2.0.1
  (`chai_postman`), the pm object and test runner (`pm`), the legacy interface (`legacy`, incl.
  liquid-json `JSON` and deprecation warnings), console serialisation as postman-sandbox teleports it
  (`console`), WHATWG globals (`globals`), `intl_install` and `sugar_install` (Sugar 1.4 ships with
  postman-sandbox; loaded on first use). `require('postman-collection')` re-exports the shim classes.
- Go (`sandbox/`): `host.go` (the shims' Go bridge), `console_sink.go`, `globals_crypto.go`,
  `globals_encoding.go`, `globals_url.go` (IDNA via the punycode asset, local time zone),
  `pm.go` (context in, effects out), `pm_execution.go` (skipRequest, runRequest),
  `pm_send_request.go`, `pm_legacy.go`, `v8_messages.go`. Vault, cookies and info travel in the
  context DTO (`webservice/dto/script_sandbox.go`) instead of separate `pm_vault/pm_cookies/pm_info`
  files; `pm_test.go` is a Go test-file name, so the test runner lives in `shims/pm.js`.
- Deviation (architecture): variable scopes live in the shims during a run (typed, as Postman's
  VariableScope) and the run hands the final scopes back; the application layer
  (`service/script_variables.go`) applies XEN-002's rules after each script — current values,
  secret rows in the vault, new names as rows with an empty initial value, `unset`/`clear` drop
  current values (rows stay), `variables:changed` once per script that changed something. The
  send resolves `{{…}}` from the scopes as scripts left them (`VariableContext.WithScriptScopes`).
- Pulled forward from US2 because ~130 corpus cases need them: the `pm.test` runner (sync, `done`,
  skip, index, legacy `tests[]`) and chai-postman. T023/T025 (Test Results BDD + UI) remain.
- Postman behaviour preferred where the corpus asserts it (documented in sandbox-api.md):
  `pm.require/mock/state/message` are undefined (V8 TypeError), `pm.datasets()` throws
  "pm.datasets is not supported in Xenon"; deprecation notices are console `warn`; `skipRequest`
  ends the script on the spot and throws a TypeError in post-response scripts (the
  `errors_and_limits` skip scenarios were updated); `pm.vault.get` of a missing name gives `null`;
  syntax/TypeError messages use V8's wording; `fs`/`os` modules are empty objects.
- `{{vault:name}}` resolves in `entity.VariableContext.Lookup` (the globals' secret rows, or the
  send's vault as scripts left it), so every resolution path gets it, not just `resolve_text.go`.
- runRequest runs the saved request (its collection/folder/request scripts) on the caller's
  scopes and persists to the caller's scopes; options `variables` are local overrides for that
  run; the child's local scope is not merged back; depth limit 10.
- Console: logs stream to the console as they happen (`ScriptLogSink`), so runRequest output keeps
  its order; `console.clear()` is recorded as a `clear` log and skipped by the console for now (US4
  T027 wires the clearing). Xenon-only notices (no environment, visualizer, Intl locale fallback)
  are console entries but not part of the script's Postman-visible console.
- Assets added (`make sandbox-assets`): liquid-json, http-reasons, mime, mime-format, sugar and a
  generated sugar method list (internal, not require-able).
- Parity: 246/249 (98.8%); the 3 differences and their causes are in
  `test/integration/parity/KNOWN_DIFFERENCES.md`, which the test reads (an unlisted difference or a
  listed case that matches again fails it). `parity-diff` runs with `TZ=UTC` and is a blocking CI
  step; the global allowlist test's "not yet built" list is empty.

---

## Phase 5: User Story 2 — Tests and Test Results (P1)

- [x] T023 [P] [US2] BDD `test/integration/features/tests/test_results.feature` and `async_tests.feature`: pass/fail with chai-postman messages for every helper in sandbox-api.md §2; skip; `done` async (pass and `done(err)`); `async () =>` rejection reported as script error; legacy `tests[]` collected; pre-request tests included; per-script labels and order; count
- [x] T024 [US2] Run → **red**; implement `pm_test.go` (runner semantics), `shims/chai_postman.js` (port of chai-postman 2.0.1 over shim objects), legacy `tests` collection → **green**
- [x] T025 [US2] UI `features/response-viewer/components/TestResultsTab.tsx` (count badge "n/m", filters All/Passed/Failed/Skipped, script label, error message) and `ScriptErrorPanel.tsx` (outcome + error in response area); RTL tests

---

## Phase 6: User Story 4 — Console (P1)

- [x] T026 [P] [US4] BDD `test/integration/features/console/console_entries.feature`, `console_masking.feature`, `console_filters.feature`: request entries (main + nested) with resolved request/response detail; log levels with objects; errors and test failures; ordering; secret masking in URL/headers/body/log text (SC-004); `List` pagination; `Clear` and `console.clear()`; ring buffer cap 5,000; bodies > 1 MB via `body_url`; a secret value set by a script mid-run is masked in later entries
- [x] T027 [US4] Run → **red**; implement `console.go` service + ring buffer + `console_entries.go` publisher (batched ≤ 50 ms) + controllers/bindings (bindings §3) → **green**
- [x] T028 [US4] UI `features/console/`: event-fed store, virtualized `ConsoleDrawer` (footer toggle + Cmd/Ctrl+Alt+C in `app/hotkeys.ts`), `ConsoleEntryRow`, `ConsoleEntryDetail` (request/response tabs, object tree), `ConsoleToolbar` (filter, search, clear); RTL tests

---

## Phase 7: User Story 5 — Editor (P2)

- [x] T029 [P] [US5] BDD `test/integration/features/scripts/script_tools.feature`: `Check` reports syntax errors with line/column; `Reference` contains every member of sandbox-api.md §1–§3; `Snippets` returns Postman's standard set
- [x] T030 [US5] Run → **red**; implement `script_tools.go` (Reference generated at build time from shim/Go definitions into an embedded JSON) → **green**
- [x] T031 [US5] UI: `shared/ui/code-editor` JS mode with generic completion and lint-diagnostic props; `features/request-builder/components/ScriptsPanel.tsx` (pre/post editors, inherit toggle, snippets menu, variable-name completion in `pm.<scope>.get/set("…")` using XEN-002 `Suggest`); Scripts tab in `ItemSettingsDialog` for collections/folders; RTL tests

---

## Phase 8: Polish

- [x] T032 [P] Performance BDD `features/scripts/performance.feature`: four typical scripts add < 100 ms (SC-002); endless loop stopped within limit + 0.5 s while another binding call stays responsive (SC-003)
- [ ] T033 Run Luis's imported collections (from XEN-003 T036) end to end; record results against the stage-3 pm-usage report in `docs/prd/pm-usage-report.md` (SC-006)
- [ ] T034 `/refactor-diff` until clean; `make validate`, `make parity`, `make sandbox-assets` drift check; CI green on 3 OSes; walk [quickstart.md](quickstart.md)

**Phases 5–8 notes (implemented)**:
- US2: the runner and chai-postman landed with T022, so `test_results`/`async_tests` (42 scenarios,
  messages taken from postman-sandbox 6.7.4) were red only on what the UI needs: `SendResult`
  gains `test_summary {total, passed, failed, skipped}` and every `script` ref gains `label`
  ("Billing · Post-response"). The tab title counts passed/total like Postman.
- US4 console: request entries for the main send and for `pm.sendRequest`/`runRequest` (nested,
  under the calling script, grouped by the main send id); failed tests are `test_failure`
  entries; errors keep line/column. `List` adds optional `filter` (all/requests/logs/errors) and
  `search`; `Get` returns the detail (request as sent, response headers/body, logged values);
  bodies over 1 MB (or binary) go to the response store under `console:<id>` and are served by
  the existing `/xenon/response/` route, released when dropped from the 5,000-entry ring or
  cleared. Masking uses the send's secrets plus those scripts set during the run (masked from
  the moment they are set, in the same script's logs and nested requests too). The publisher emits
  `console:entries` at most every 50 ms or 100 entries; Clear and `console.clear()` emit
  `{cleared: true}` and drop pending entries. Existing script features now list the console
  with `filter: logs` because request entries sit between the logs.
- US5: `ScriptService` (`Check`/`Reference`/`Snippets`). The reference is
  `sandbox/reference/reference.json`, generated by `make script-reference` from
  `reference/members.tsv` (path, signature, description) plus the live sandbox, which gives each
  member's kind; a sandbox test fails on drift or on a listed member the shims lack. Syntax
  errors past the script's end (unclosed blocks) point at its last character.
- Requests carry `scripts` and `inherit_scripts`; a send runs the carried scripts (saved or not)
  and falls back to the saved file when absent; Save and collection/folder `SaveSettings` write
  carried scripts (files keep theirs when the payload has none) — `ScriptsEdited` on the entity.
- UI: the scripts editor lives in `features/scripts` and is composed by the app into the
  request's Scripts tab and a Scripts tab of collection/folder settings (render props, like
  collection variables). Cmd/Ctrl+Alt+C toggles the console; Alt combos match `event.code`
  (Option changes `event.key` on macOS) and AltGr typing never fires a shortcut.
- T032 passes on the existing implementation (~33 ms for four typical scripts locally); new
  generic steps start a binding call in the background and wait for it.
- Each console call carries the script's variable scopes at that moment, so a value the script
  just stored in a secret row is masked in its own following logs too.

---

## Dependencies & Execution Order
Setup → Foundational (T007 corpus and T008 core gate everything) → US1 → US3 → US2 → US4
(US4 backend can start after US1; UI after T027) → US5 → Polish.

## Parallel opportunities
T002–T004; T005–T006; within stories all [P] BDD files; US4 and US5 UI work alongside US3
backend once contracts are fixed.

## Implementation Strategy
MVP = Foundational + US1 + US3 (scripts run with the full API) → US2 (tests) → US4 (console) →
US5 (editor comfort) → Polish.
