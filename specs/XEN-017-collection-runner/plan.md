# Implementation Plan: Collection Runner

**Branch**: `xen-017` | **Date**: 2026-10-23 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/XEN-017-collection-runner/spec.md`

## Summary

Collections and folders get a **Runner** section in their existing tab, with three sub-tabs: Setup,
Results and Runs. A new Go service, `ManageRunnerService`, plans, validates and executes runs in
the background.

- **Sending**: every row goes through the **same send core** as a single send (auth inheritance,
  script chains, OAuth, digest, AWS and masking). That core is refactored so that a domain
  `RunContext` can change exactly the runner-specific behaviour:
  - the environment override;
  - the iteration info and data row;
  - the variable overlay (when values are not kept);
  - a run-scoped cookie jar;
  - console muting and a per-row collector;
  - no History;
  - a run-keyed response cache.
- **Script control flow**: `pm.execution.setNextRequest` becomes real inside runs, and
  `pm.info.iteration`, `iterationCount` and `pm.iterationData` are fed.
- **Performance runs**: virtual users follow a load profile and record into a bounded-memory
  latency histogram, with per-second series.
- **Events and storage**: progress and finish events drive the UI. Finished runs are stored as JSON
  in the app-data folder (25 per target) and are exportable as JSON or JUnit.
- **Preferences**: rows, order, aliases and settings are remembered in UI state.
- **MCP**: two tools, `run` (start/stop) and `runs` (get/list).
- **PRD**: removes the runner from the non-goals.

See [research.md](./research.md) for every decision.

## Technical Context

**Language/Version**: Go 1.26 (backend), TypeScript 5 + React 19 (frontend)

**Primary Dependencies**: Wails v3, goja, `encoding/csv`, `encoding/xml`, `encoding/json`, Zustand
and dnd-kit (already present). No new dependency.

**Storage**:

- `<DataDir>/workspaces/<localId>/runs/<targetId>/<runId>.json` (stored runs, capped at 25 per
  target);
- `state.json` `runners` (preferences);
- the in-memory active-run registry and response cache.

**Testing**: Godog BDD (`test/integration/`, `make bdd TAG=@runner`), Go unit tests for the pure
domain functions (load profile curve, histogram percentiles, reconcile, iteration data parsing,
JUnit builder), and Vitest for the UI.

**Target Platform**: desktop — macOS, Windows, Linux

**Project Type**: desktop app (Go core + React webview)

**Performance Goals**:

- Progress reaches the UI within 1 s of each row (SC-003).
- A 100-user, 60-minute performance run keeps the UI responsive (SC-006): at most 1 aggregated
  event per second, with histograms of bounded memory.

**Constraints**:

- A single send behaves byte-for-byte as today (History, variables, cookies, console). This is
  proven by the unchanged suites.
- No secret value in a stored or exported run (SC-007).
- No workspace file is written by run storage or preferences (SC-008).

**Scale/Scope**:

- Functional runs: tens to hundreds of rows × iterations.
- Performance runs: ≤ 100 users × ≤ 3600 s.
- Stored runs: ≤ 25 per target.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first**
  - Runs execute locally.
  - Stored runs and preferences live in the app-data folder, keyed by workspace `localId`.
  - Nothing is networked except the user's own requests, and the workspace boundary holds (US7-5).
- [x] **II. Secrets**
  - Stored runs and exports mask every vault secret value (FR-039), using the console's
    `MaskSecrets`.
  - Response bodies are never stored on disk; they are held in memory only.
  - The MCP output passes the leak guard.
  - BDD scans the stored files for the secret values (SC-007).
- [x] **III. Go owns logic**
  - Planning, reconciling, validation, data-file reading and parsing, execution, statistics,
    storage and export are all Go.
  - React renders and edits preferences through `platform/services/runner.ts`.
  - No Wails import outside `internal/infra`.
- [x] **IV. Architecture skills**
  - Every new file is placed per layer below.
  - In-memory registries live in `integration/persistence`, following `response_body.go`.
  - Frontend files follow `frontend-architecture`: a new `features/runner`; `app/` composes it
    into the item tab and the sidebar; cross-feature access is to `store.ts` only.
- [x] **V. BDD-first**
  - The feature files are listed per requirement below.
  - Only the HTTP target (mock API), the file system root, the dialog, the keychain and the clock
    are faked.
- [x] **VI. Postman parity**
  - The runner was outside the PRD, so it is recorded as a **user decision** (FR-049, Decisions
    Oct 23, 2026).
  - Settings, iteration data, `setNextRequest`, `skipRequest` and `pm.info` follow Postman.
  - Xenon's deliberate differences (in-tab section, no History, local runs list) are user
    decisions recorded in the PRD.
- [x] **VII. Cross-platform**
  - Paths come from `filepath` and app-data helpers.
  - The file-dialog filters work on all three OSes.
  - No new shortcuts.

**Post-design re-check**: all seven still pass, and there are no Complexity Tracking entries.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-017-collection-runner/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── bindings.md
│   ├── events.md
│   ├── mcp.md
│   └── ui.md
├── checklists/
└── tasks.md
```

### Source Code (repository root)

```text
internal/domain/
├── enums/run_type.go, run_status.go, run_row_outcome.go, load_profile_kind.go,
│   run_result_filter.go, run_export_format.go (+ value packages)                 NEW
└── entity/
    ├── run_target.go, run_row.go, run_plan_row.go (FlattenRunRows, ReconcileRunRows)     NEW
    ├── functional_settings.go, performance_settings.go, load_profile.go (ActiveUsersAt)  NEW
    ├── run_config.go (Validate), runner_preferences.go, iteration_data.go                NEW
    ├── run_context.go (NoRunContext, RunCookieJarID, RunResponseKey), run_variables.go   NEW
    ├── next_request.go, run_row_result.go (NewRunRowOutcome), run_summary.go,
    │   run_progress.go, run.go                                                            NEW
    ├── latency_histogram.go, request_stats.go, performance_second.go,
    │   performance_stats.go                                                               NEW
    ├── script_info.go          CHG  Iteration field
    ├── script_context.go       CHG  IterationData
    ├── script_run.go           CHG  NextRequest
    ├── send_execution.go       CHG  NextRequest()
    ├── variable_context.go     CHG  WithIterationData layer (after local, before environment)
    ├── cookie.go               CHG  MergeCookies
    ├── ui_state.go             CHG  Runners map[string]RunnerPreferences
    └── ai_approval.go          CHG  optional Summary on the approval request
internal/application/
├── usecase/
│   ├── find_runs.go, save_run.go, delete_runs.go (interfaces)                       NEW
│   ├── manage_active_runs.go (Claim/Release/Find/ListActive)                        NEW
│   ├── find_iteration_data.go (read file bytes)                                     NEW
│   ├── parse_iteration_data.go (pure, RUN-05400)                                    NEW
│   ├── build_run_junit.go (pure)                                                    NEW
│   ├── publish_run_event.go                                                         NEW
│   ├── reject_invalid_run.go (RUN-01404 … RUN-07404 constructors)                   NEW
│   └── manage_cookie_jar.go    CHG  DropJar
├── adapter/manage_runner.go    NEW  ManageRunnerService
└── service/
    ├── send_core.go            NEW  shared prepare/execute/settle extracted from send_request.go, taking RunContext
    ├── send_request.go         CHG  Send = core with NoRunContext (behaviour unchanged)
    ├── send_run.go             CHG  RunContext: overlay vs persist, jar id, console mute/collect, response key,
    │                                iteration info + data, setNextRequest warning in performance
    ├── request_variables.go    CHG  environment override; overlay applied
    ├── script_variables.go     CHG  per-workspace lock around persist
    ├── cookie_jar.go           CHG  jar id parameter; seed / merge / drop
    ├── manage_runner.go        NEW  Plan, SavePreferences, Start, Pause, Resume, Stop, Get, List, Active,
    │                                Delete, Clear, Export, Response, StopWorkspace
    ├── run_functional.go       NEW  row loop, iterations, delay, setNextRequest, stop on error, gate
    ├── run_performance.go      NEW  profile controller, virtual users, stats, per-second ticks
    ├── run_row.go              NEW  one row → core with RunContext → RunRowResult (masking)
    ├── activate_workspace.go   CHG  StopWorkspace for the previous workspace
    └── remove_workspace.go     CHG  StopWorkspace
internal/integration/
├── persistence/
│   ├── run.go                  NEW  stored runs + 25 cap + mutex per workspace
│   ├── model/run.go            NEW
│   ├── active_runs.go          NEW  in-memory registry
│   ├── iteration_data.go       NEW
│   ├── local_variables.go      CHG  mutex
│   └── model/ui_state.go       CHG  runners
├── webservice/
│   ├── cookie_jar.go           CHG  DropJar
│   ├── sandbox/pm_execution.go CHG  setNextRequest host function
│   ├── sandbox/shims/pm.js     CHG  setNextRequest → host; iterationData from context
│   ├── sandbox/shims/legacy.js CHG  postman.setNextRequest → host
│   └── dto/script_sandbox_info.go, script_sandbox_context.go  CHG  iteration, iterationData
├── publisher/run_event.go, dto/run_progress_changed.go, dto/run_finished.go          NEW
├── adapter/runner_controller.go                                                     NEW
└── entrypoint/
    ├── dto/run_config.go, run_plan.go, run.go, run_summary_view.go, start_run.go,
    │   run_reference.go, run_row_reference.go, run_target_reference.go,
    │   save_runner_preferences.go, export_run.go, mcp_run.go, mcp_runs.go             NEW
    ├── dto/pick_file.go, ui_state.go, ai_approval_request.go                         CHG
    └── controller/runner.go, validate_runner.go, mcp_run.go, mcp_runs.go             NEW
internal/infra/
├── server/binding/runner.go    NEW; binding/dialog.go CHG (filters)
├── server/event/events.go      CHG  RunProgressChanged, RunFinished
├── server/mcp/tool_run.go, tool_runs.go NEW; controllers.go, routes.go CHG
└── dependency/** (service, controller, persistence, publisher, mcp_routes) + Wails service list  CHG
frontend/src/
├── platform/services/runner.ts (+ test), platform/mappers/runner.ts (+ test)        NEW
├── platform/services/dialogs.ts CHG  filters; platform/events.ts CHG  onRunProgress, onRunFinished
├── features/runner/model.ts, store.ts (+ test)                                      NEW
├── features/runner/components/RunnerSection, RunOrderList, RunSettings, PerformanceSettings,
│   EnvironmentSelect, RunProgress, RunResults, RunRowDetail, PerformanceResults, RunsList,
│   ClearRunsDialog, RunFinishedToast (+ .module.css, + .test.tsx)                    NEW
├── features/runner/hooks/use-run-order-reorder.ts (+ test)                          NEW
├── features/item-settings/model.ts, components/ItemTab.tsx (+ tests)  CHG  'runner' section, runner render prop
├── features/tabs/components/TabItem.tsx (+ test)              CHG  running indicator
├── features/sidebar/components/TreeItem.tsx (+ test)          CHG  Run… item, onRunItem prop
├── app/layout/ItemWorkspace.tsx, app/App.tsx, app/bootstrap.ts CHG  compose runner, toast, event wiring, Run… action
docs/prd/xenon-mvp-prd.md       CHG  non-goal removed; Collection runner section; L507 note; Decisions (Oct 23, 2026)
test/integration/features/runner/*.feature, test/integration/features/mcp/runner.feature  NEW
test/integration/features/dialogs/pick_file.feature  CHG
```

**Structure Decision**: this is the existing Xenon layout (a Go 4-layer backend, the `frontend/`
Wails SPA and the Godog suite). The runner is a new vertical slice that reuses the send core
through a domain `RunContext` rather than calling the public `Send`.

## BDD feature files (written by `/create-tests`, all tagged `@runner` unless noted)

Runs are asynchronous. The scenarios start a run with `RunnerService.Start`, then wait for the
`runner:finished` event using the existing event steps (`an? "<event>" event with the field …
should have been emitted within N ms`). They read the results with `RunnerService.Get`.

| File | Locks |
| --- | --- |
| `runner/runner_plan.feature` | FR-003 (flattened order, subfolders, no examples, folder path), FR-006 (preferences in `state.json`, no workspace file changed), FR-007 (reconcile: removed / added / kept), FR-008 (alias choice, missing alias → Default), FR-013 (config remembered; missing data file cleared → `data_file_missing`), `RUN-07404` |
| `runner/run_functional.feature` | FR-016 (order, inherited auth/headers/scripts, collection → folder → request chain, tests collected), FR-017 (saved, not draft: a send with a different draft does not affect the run), FR-021 (skipRequest → skipped), FR-026 (History unchanged), FR-032/FR-033 (row fields, summary), folder target (US3-3), alias row sends the alias, `RUN-06404` for a request deleted before its turn |
| `runner/run_validation.feature` | FR-014 (each rejection: empty selection `RUN-03400`, ranges `RUN-04400`, data file `RUN-05400`, missing secrets, active run `RUN-02409`), FR-015 (`SND-05409` → confirmed start runs every row with no further rejection) |
| `runner/run_iterations.feature` | FR-018 (`pm.info.iteration`/`iterationCount`, `pm.iterationData`, `{{field}}` precedence over environment), FR-019 (CSV header, extra/missing cells, JSON array, last-row reuse, default count), FR-022 (delay: request receive times at the mock API are ≥ D ms apart) |
| `runner/run_control_flow.feature` | FR-020 (by name, by id, backwards, last call wins, `null`, unknown target + console warning, next iteration starts at row 1), single send stays a no-op, legacy `postman.setNextRequest` |
| `runner/run_settings.feature` | FR-012 (environment override, active environment unchanged; "none"), FR-023 (stop on error: network error, script error; failed assertion continues), FR-024 (keep values on/off, `variables.json` unchanged when off), FR-025 (stored cookies used / skipped; save-cookies merges / does not), FR-027 (logs off → no `console:entries`, empty row console) |
| `runner/run_lifecycle.feature` | FR-028 (`runner:progress` per row), FR-029 (pause holds, resume continues, stop → cancelled row + `stopped`), FR-030 (second run on the same target refused; different targets overlap; tab send during a run works), FR-031 (workspace switch stops and stores), `runner:finished` payload |
| `runner/run_newman_parity.feature` (`@newman`, CI `newman` job) | SC-005: the runner and Newman (Postman's runner) send the identical request sequence for a collection using `setNextRequest`, `skipRequest`, `pm.iterationData` and `pm.info.iteration` with a CSV data file |
| `runner/run_storage.feature` | FR-038 (location, 25 cap per target, folder and collection counted apart), FR-039 (masked secrets in URL, headers, body, console; no response stored; SC-007 scan), FR-041 (List newest first, fields), FR-041a (Delete, Clear), workspace isolation |
| `runner/run_responses.feature` | FR-040 (`RunnerService.Response` with persisted responses; `RUN-01404` without them or after the run is deleted) |
| `runner/run_export.feature` | FR-037 (JSON = stored run; JUnit suites/cases/failure/skipped/error; performance → JSON only, JUnit refused) |
| `runner/run_performance.feature` | FR-011 (ranges and profile validation), FR-042 (scripts run, per-user variables and cookies, nothing persisted, no History), FR-043 (active users over time per profile, from `runner:progress` metrics with short durations), FR-044 (setNextRequest ignored + warning, data wraps), FR-045/FR-046 (counts, percentiles within bounds against a mock with fixed latency, error rate with 500s, status counts, series length = duration) |
| `mcp/runner.feature` (`@mcp_runner`) | FR-047 (start with defaults / explicit rows and settings, get, list, stop, approval once for confirm environments, errors), FR-048 (UI-started run readable over MCP; MCP-started run on `Active`) |
| `dialogs/pick_file.feature` (CHG, `@pick_file`) | `filters` reach the native dialog; omitted filters behave as before |

The pure domain functions get Go unit tests next to them:

| Function | File |
| --- | --- |
| `LoadProfile.ActiveUsersAt` | `load_profile_test.go` |
| `LatencyHistogram` | `latency_histogram_test.go` |
| `ReconcileRunRows` | `run_plan_row_test.go` |
| `ParseIterationData` | `parse_iteration_data_test.go` |
| `BuildRunJUnit` | `build_run_junit_test.go` |
| `NewRunRowOutcome` | `run_row_result_test.go` |

UI-only requirements are locked by Vitest:

| Requirements | Test file |
| --- | --- |
| FR-001 (section and sub-tabs) | `ItemTab.test.tsx`, `RunnerSection.test.tsx` |
| FR-001a | `TabItem.test.tsx` |
| FR-001b | `RunFinishedToast.test.tsx`, `bootstrap.test.ts` |
| FR-002 | `TreeItem.test.tsx` |
| FR-004, FR-005 | `RunOrderList.test.tsx`, `use-run-order-reorder.test.ts` |
| FR-009–FR-011 (form) | `RunSettings.test.tsx`, `PerformanceSettings.test.tsx` |
| FR-028/FR-029 controls | `RunProgress.test.tsx` |
| FR-034 filters | `RunResults.test.tsx` |
| FR-035 detail | `RunRowDetail.test.tsx` |
| FR-036 Run again | `store.test.ts` |
| FR-041/FR-041a UI | `RunsList.test.tsx`, `ClearRunsDialog.test.tsx` |
| FR-046 chart | `PerformanceResults.test.tsx` |
| Mapping | `platform/services/runner.test.ts`, `platform/mappers/runner.test.ts` |

FR-049 (PRD) is verified by review in `/implement-task`. It is a documentation change with no
runtime behaviour.

## Complexity Tracking

No constitution violations.
