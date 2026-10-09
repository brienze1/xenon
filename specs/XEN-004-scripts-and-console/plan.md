# Implementation Plan: Scripts, Tests and Console (Stage 4)

**Branch**: `xen-004` (stacked on `xen-003`) | **Date**: 2026-10-09 | **Spec**: [spec.md](spec.md)

## Summary

A Postman-compatible sandbox on goja: one fresh runtime per script, lazily-loaded embedded
libraries pinned to postman-sandbox 6.7.4, Go-implemented `pm` services (variables, sendRequest,
runRequest, vault, cookies, crypto) with thin JS shims for Postman's object model, chai-postman
assertions and the legacy API; interrupt-based limits (time, stack, memory); an execution
pipeline wrapped around XEN-001's send; tests collected into Test Results; a session console fed
by events with secret masking; editor support (reference, snippets, syntax check). A Node-based
parity oracle (postman-sandbox itself) proves equivalence in CI.

## Technical Context

**Language/Version**: Go 1.25+, TypeScript 5 + React 19; embedded JavaScript assets

**Primary Dependencies** (new): `github.com/dop251/goja` and `github.com/dop251/goja_nodejs`
(pinned pseudo-versions, research R1); asset toolchain: esbuild + the npm packages in research
R5/R6 (build-time only, in `tools/sandbox-assets/`); parity oracle: `postman-sandbox@6.7.4`
(dev/CI only, `tools/parity/`)

**Storage**: none new besides `app.json` `settings.scriptTimeoutMs`

**Testing**: godog BDD for pipeline behaviour; Go parity test over ≥ 200 corpus cases diffed
against the Node oracle; unit tests per shim; Vitest + RTL for console/Test Results/editor

**Target Platform / Project Type**: desktop-app, macOS/Windows/Linux

**Performance Goals**: SC-002 (< 100 ms overhead for 4 typical scripts), SC-003 (stop within
limit + 0.5 s)

**Constraints**: sandbox has no FS/process/network access except Go-mediated `sendRequest`/
`runRequest`; secrets masked in console; scripts never modify stored requests

**Scale/Scope**: scripts up to a few thousand lines; console 5,000 entries

## Constitution Check

- [x] **I. Local-first** — no network beyond what scripts explicitly send; libraries embedded,
  never fetched. ✅
- [x] **II. Secrets** — console masking at entry creation; History keeps unresolved requests;
  `pm.vault` uses the XEN-002 vault; logs never contain resolved headers/bodies. ✅
- [x] **III. Go owns logic** — scripts run in Go (goja), never in the webview; the UI only shows
  results, console and editor aids from Go. ✅
- [x] **IV. Architecture skills** — see Structure Decision. ✅
- [x] **V. BDD-first** — features per story + parity oracle; boundaries faked: HTTP target
  (ApiMock), clock, keychain, emitter. ✅
- [x] **VI. Postman parity** — oracle-backed (R7); scope decisions recorded in spec
  Clarifications. ✅
- [x] **VII. Cross-platform** — pure-Go engine; assets embedded; parity suite runs on 3 OSes. ✅

**Post-design re-check**: all gates pass.

## Project Structure

### Documentation
```text
specs/XEN-004-scripts-and-console/
├── plan.md · research.md · data-model.md · quickstart.md
├── contracts/
│   ├── sandbox-api.md     ← exact pm/globals/require/legacy surface + error kinds
│   └── bindings.md        ← Send result additions, ScriptService, ConsoleService, SettingsService
├── checklists/
└── tasks.md
```

### Source code (additions)
```text
internal/domain/
├── entity/  script_ref.go, send_execution.go, script_run.go, script_error.go, test_result.go,
│            console_entry.go, nested_request.go, app_settings.go
└── enums/   script_phase, script_owner_kind, script_error_kind, test_status,
             console_entry_kind, console_level, send_outcome (+ value pkgs)

internal/application/
├── usecase/  run_script (adapter to the sandbox), collect_script_chain (pure: order +
│             inherit + disabled), append_console_entry, find_console_entries, clear_console,
│             mask_secrets (pure, used by console + errors), find_app_settings, save_app_settings
├── adapter/  script_service, console_service, settings_service; send_request extended
└── service/  send_request.go (pipeline per data-model state machine), script_tools.go
              (Check/Reference/Snippets), console.go, settings.go

internal/integration/
├── webservice/sandbox/            ← script engine (see Structure Decision)
│   ├── runtime.go                 (fresh runtime, event loop, limits, wrapper, completion)
│   ├── limits.go                  (timeout, stack, memory guard)
│   ├── pm_variables.go, pm_send_request.go, pm_execution.go, pm_vault.go, pm_cookies.go,
│   │   pm_info.go, pm_test.go, console_sink.go, globals_crypto.go, globals_encoding.go,
│   │   globals_url.go, require_registry.go (embedded modules + lazy globals)
│   ├── shims/ (embedded JS) pm_request.js, pm_response.js, property_list.js, url.js,
│   │   chai_postman.js, legacy.js, intl_install.js
│   ├── assets/ (embedded, generated) vendored libraries + bundles + MANIFEST.json
│   └── dto/ script context/result structures
├── adapter/   script_runner.go (application port implemented by sandbox), console_store
├── persistence/ console ring buffer (memory) console.go; settings via registry model
├── publisher/ console_entries.go (batched events)
└── entrypoint/controller/ script_*.go, console_*.go, settings_*.go; request_send.go extended

internal/infra/server/binding/  script.go, console.go, settings.go; event/events.go adds console:entries

tools/
├── sandbox-assets/  package.json (exact versions), build.mjs (esbuild bundles + MANIFEST)
└── parity/          package.json (postman-sandbox 6.7.4), oracle.mjs (corpus case → outcome JSON)

test/integration/
├── features/scripts/   execution_order, inheritance, request_mutation, errors_and_limits,
│                       variables_from_scripts, send_request, run_request_skip, legacy_api,
│                       libraries, unsupported_apis, vault_and_cookies
├── features/tests/     test_results, async_tests
├── features/console/   console_entries, console_masking, console_filters
├── parity/             parity_test.go (runs corpus through Xenon; diffs with oracle output)
└── resources/scripts/corpus/*.json (+ oracle outputs regenerated in CI)

frontend/src/
├── features/console/        store.ts (event-fed, virtualized list), ConsoleDrawer, ConsoleEntryRow,
│                            ConsoleEntryDetail, ConsoleToolbar (filter/search/clear)
├── features/response-viewer/ TestResultsTab (count badge, filters), ScriptErrorPanel
├── features/request-builder/ ScriptsPanel (two editors, inherit toggle), script editor wiring
├── features/sidebar/         ItemSettingsDialog Scripts tab (+ inherit toggle on folders)
├── features/settings/        script timeout setting
├── shared/ui/code-editor/    JS mode + generic completion source + lint diagnostics props
└── platform/services/        scripts.ts, console.ts, settings.ts
```

**Structure Decision**: the sandbox is an outer boundary that runs third-party code against an
external contract (Postman's sandbox API) — the same role a webservice plays for an external
API — so it lives in `internal/integration/webservice/sandbox/` behind the application port
`ScriptRunner`. All Xenon behaviour reachable from scripts (variables, sending, vault, console,
tests) is called through application adapters passed into the run context, so business rules
stay in application/domain. JS shims and assets are data embedded in that package.

## Key design decisions

1. **One runtime per script**, compiled programs cached, libraries lazy (R2) → meets SC-002.
2. **Pipeline owns order** (`collect_script_chain`): owners list from the tree, cut at the
   nearest `inheritScripts: false`, disabled scripts dropped.
3. **Local scope** is a map on `SendExecution`, plugged into XEN-002's resolver as the
   strongest scope (the slot reserved there).
4. **Pre-request mutations** are applied to a request copy inside the execution; resolution of
   `{{…}}` happens after pre-request scripts (so scripts can set variables used by the request).
5. **Nested sends** (`sendRequest`, `runRequest`) reuse the send service with a flag that
   suppresses History and (for sendRequest) scripts; `runRequest` runs the target's scripts with a
   child `SendExecution` and merges variable changes back.
6. **Masking** is a pure function over text with the current secret set; applied to console
   entries, script error messages and logs.
7. **Parity first**: the oracle corpus is built before the shims and drives them (red → green).

## Complexity Tracking

No constitution violations to justify.
