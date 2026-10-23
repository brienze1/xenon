---

description: "Task list for XEN-017 Collection Runner"
---

# Tasks: Collection Runner

**Input**: Design documents from `specs/XEN-017-collection-runner/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/ (bindings.md,
events.md, mcp.md, ui.md), quickstart.md

**Tests**: MANDATORY (constitution V, BDD-first).

- Every user story starts with Godog feature files in `test/integration/features/` (tag `@runner`,
  or `@mcp_runner` for MCP) that must be proven red before implementation.
- Pure domain functions get Go unit tests, and frontend behaviour gets Vitest tests.
- Write features against the existing generic step catalog in `test/integration/steps/`, e.g. `I
  call "<Service.Method>" with`, the mock API request steps, the event steps, and the app-data file
  steps. Add a step only when no existing step can express the assertion, and keep it generic.

**Organization**: tasks are grouped by user story (spec.md). Inside each story, layers go
innermost outward: domain → application → integration → infra → frontend.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel (different files, no dependency on an unfinished task)
- **[Story]**: US1…US9 from spec.md

---

## Phase 1: Setup

- [ ] T001 Set up the work: on `xen-017` (rebased on a freshly fetched `origin/main`), run `cd frontend && npm ci` and `make bindings` before any Vitest run (fresh-worktree note). Then confirm `make bdd` and `cd frontend && npm test` are green before any change.

---

## Phase 2: Foundational (blocking all stories)

**Purpose**: the run model, the shared send core with `RunContext`, the active-run registry, the
event publisher and the `RunnerService` skeleton every story needs.

- [x] T002 [P] Add the enums `RunType`, `RunStatus`, `RunRowOutcome`, `LoadProfileKind`, `RunResultFilter` and `RunExportFormat`, each with its value package, in `internal/domain/enums/` (`run_type.go`, `runtype/value.go`, …), following `.claude/skills/enums.md`.
- [x] T003 [P] Write red Go unit tests in `internal/domain/entity/run_config_test.go` for `RunConfig.Validate`:
  - an empty selection gives `RUN-03400`;
  - iterations < 1, a delay outside 0–3,600,000, users outside 1–100 and a duration outside 1–3600 each give `RUN-04400`, naming the field;
  - profile parameters outside the duration or user count give `RUN-04400`;
  - Postman's defaults come from `NewRunConfig()`.
- [x] T004 Add the following entities to `internal/domain/entity/` per data-model.md: `RunTarget` (`run_target.go`), `RunRow` (`run_row.go`), `FunctionalSettings` (`functional_settings.go`), `LoadProfile` (fields only, `load_profile.go`), `PerformanceSettings` (`performance_settings.go`), `RunConfig` (`NewRunConfig`, `Validate`, `SelectedRows`, `run_config.go`) and `RunnerPreferences` (`runner_preferences.go`). T003 must go green.
- [ ] T005 [P] Add the `RUN-` error constructors in `internal/application/usecase/reject_invalid_run.go`:
  - `RejectUnknownRun` (`RUN-01404`);
  - `RejectActiveRun` (`RUN-02409`, details `run_id`);
  - `RejectEmptyRun` (`RUN-03400`);
  - `RejectInvalidRunSetting` (`RUN-04400`, details `field`);
  - `RejectInvalidIterationData` (`RUN-05400`, details `reason`);
  - `NewMissingRunRequest` (`RUN-06404`);
  - `RejectUnknownRunTarget` (`RUN-07404`).

  Follow the `errs` helpers used by `reject_unconfirmed_send.go`.
- [x] T006 Add `RunContext` (`NoRunContext()`, `RunCookieJarID(localId, runId)`, `RunResponseKey(runId, index)`) in `internal/domain/entity/run_context.go`. Add `RunVariables` (`Apply`, `Over`) in `internal/domain/entity/run_variables.go`, with red-then-green unit tests in `run_variables_test.go`:
  - an overlay change is visible through `Over`;
  - an unset removes the value;
  - vault changes are tracked.
- [x] T007 Add `Iteration int` to `ScriptInfo` (`internal/domain/entity/script_info.go`) and `IterationData map[string]any` to `ScriptContext` (`internal/domain/entity/script_context.go`). Add `VariableContext.WithIterationData(row)`, a layer after local and before environment, in `internal/domain/entity/variable_context.go`, with unit cases in `variable_context_test.go` showing that iteration data wins over environment and loses to local.
- [x] T008 Extract the shared send core from `internal/application/service/send_request.go` into `internal/application/service/send_core.go`. The core takes an `entity.RunContext`, and `Send` calls it with `NoRunContext()`.
  - `send_run.go` reads its jar id, console mute and collector, response key, overlay and iteration info from the context.
  - `request_variables.go` `loadChain` takes the environment override (`""` = active, `entity.NoEnvironmentID` = none) and applies an overlay when one is given.
  - `cookie_jar.go` takes a jar id.

  No behaviour changes for a single send: `make bdd TAG=@sending`, `@history`, `@cookies`, `@variables`, `@scripts`, `@console` and `@aliases` must stay green.
- [x] T009 Add `DropJar(ctx, jarId)` to `usecase.ManageCookieJar` (`internal/application/usecase/manage_cookie_jar.go`) and implement it in `internal/integration/webservice/cookie_jar.go`. Add `entity.MergeCookies(workspace, run)` (same name, domain and path replace; others add) in `internal/domain/entity/cookie.go`, with unit cases in `cookie_test.go`.
- [x] T010 [P] Add a `sync.Mutex` to `localVariablesPersistence` in `internal/integration/persistence/local_variables.go`. Add a per-workspace lock around the find → apply → save sequence of `scriptVariables.persist` in `internal/application/service/script_variables.go` (research R4).
- [x] T011 [P] Add `usecase.ManageActiveRuns` (`Claim`, `Release`, `Find`, `ListActive`, `ListWorkspace`) in `internal/application/usecase/manage_active_runs.go`, and its in-memory, mutex-guarded implementation in `internal/integration/persistence/active_runs.go`. `Claim` returns `RUN-02409` for an occupied target.
- [x] T012 [P] Add the entities `RunRowResult` (`NewRunRowOutcome`), `RunSummary` (computed from rows), `RunProgress` and `Run` (state transitions per data-model.md) in `internal/domain/entity/`. Add unit tests for `NewRunRowOutcome` in `run_row_result_test.go` and for the transitions in `run_test.go`.
- [x] T013 [P] Add `usecase.PublishRunEvent` (`Progress`, `Finished`) in `internal/application/usecase/publish_run_event.go`. Add the publisher in `internal/integration/publisher/run_event.go`, throttled to 10/s per run with the last state always emitted, and the DTOs `RunProgressChanged` (`runner:progress`) and `RunFinished` (`runner:finished`) in `internal/integration/publisher/dto/`. Register both in `internal/infra/server/event/events.go`.
- [ ] T014 Add `adapter.ManageRunnerService` in `internal/application/adapter/manage_runner.go` with the full method set from contracts/bindings.md. Add `service.manageRunner` in `internal/application/service/manage_runner.go`; it builds its own send core with the same dependencies as `NewSendRequestService`. Add `adapter.RunnerController` in `internal/integration/adapter/runner_controller.go`.
- [ ] T015 Add the DTOs per contracts/bindings.md in `internal/integration/entrypoint/dto/`:
  - `run_config.go`, `run_plan.go`, `run.go`, `run_summary_view.go`;
  - `start_run.go`, `run_reference.go`, `run_row_reference.go`, `run_target_reference.go`;
  - `save_runner_preferences.go`, `export_run.go`.

  Each comes with `ToEntity` and constructors, and `validate` tags on the required fields.
- [ ] T016 Add `controller.runner` in `internal/integration/entrypoint/controller/runner.go`, following the full controller skill pattern (validate, `ToEntity`, service, error handler, panic recovery), with the DTO validation codes `RUN-08400`…`RUN-12400` in `validate_runner.go`. Add the binding `RunnerService` in `internal/infra/server/binding/runner.go`, wire DI under `internal/infra/dependency/`, and register the service in the Wails service list.
- [ ] T017 [P] Add `frontend/src/platform/services/runner.ts` and `frontend/src/platform/mappers/runner.ts`, with tests `runner.test.ts` for each. Add `onRunProgress` and `onRunFinished` to `frontend/src/platform/events.ts`. Run `make bindings`.
- [ ] T018 [P] Add the empty feature shell `frontend/src/features/runner/model.ts` and `store.ts` (state shape per contracts/ui.md, `reset()`), and register its reset on workspace switch in `frontend/src/app/bootstrap.ts`.

**Checkpoint**: the single-send suites are green on the shared core, and `RunnerService` is
callable with no behaviour behind it yet.

---

## Phase 3: User Story 1 — Run a collection's requests in order (P1) 🎯 MVP

**Goal**: start a functional run on a collection and see live progress and the results.

**Independent Test**: start a functional run on a 3-request collection against the mock API. The
requests are sent in order with inherited auth and scripts, the results hold their statuses and
tests, and History is unchanged.

- [x] T019 [P] [US1] Write the red `test/integration/features/runner/run_functional.feature` (`@runner @run_functional`), covering:
  - FR-016: order; inherited auth and headers; the collection → folder → request script chain; tests collected.
  - FR-017: a draft differing from disk does not affect the run.
  - FR-021: `skipRequest` gives `skipped`.
  - FR-026: no History entry, and no `history:changed`.
  - FR-032 and FR-033: row fields and summary.
  - An alias row sends the alias.
  - `RUN-06404` for a request deleted before its turn.
  - SC-002: each row's status and test outcomes equal those of the same request sent alone.
  - SC-008: no workspace file changes when keep values is off.
- [x] T020 [P] [US1] Write the red `test/integration/features/runner/run_lifecycle.feature` (`@runner @run_lifecycle`), covering:
  - FR-028: `runner:progress` after every row.
  - FR-029: pause holds, resume continues, stop gives a `cancelled` row and status `stopped`.
  - FR-030: a second start on the same target gives `RUN-02409`, while a run on another collection overlaps and a tab send during a run works.
  - FR-031: activating another workspace stops the run and stores it as `stopped`.
  - The `runner:finished` payload.
  - SC-003: each `runner:progress` event is emitted within 1000 ms of its row's response.
- [x] T021 [P] [US1] Write the red `test/integration/features/runner/run_validation.feature` (`@runner @run_validation`), covering:
  - FR-014: an empty selection gives `RUN-03400`; each out-of-range field gives `RUN-04400`; missing secrets give `SND-06409`, and a restart with `acknowledge_missing_secrets` runs; an active run gives `RUN-02409`.
  - An expired, non-refreshable OAuth 2.0 token is sent without `SND-07409`.
  - FR-015: `SND-05409` on start; a restart with `confirmed_environment_id` sends every row with no further rejection.

  Run `make bdd TAG=@runner` and record that T019–T021 are red.
- [x] T022 [US1] Implement `FlattenRunRows` (depth-first sidebar order, examples skipped, folder path relative to the target) in `internal/domain/entity/run_plan_row.go`, with unit tests in `run_plan_row_test.go`.
- [x] T023 [US1] Implement `manageRunner.Start`, `Get` and `Active` in `internal/application/service/manage_runner.go`:
  1. Resolve the target (`RUN-07404`) and validate the config.
  2. Resolve the environment and run the preflight once: confirm (`SND-05409`) and missing secrets over every selected row.
  3. `Claim` the target.
  4. Start the run goroutine on a context detached from the call, then return the summary view.
- [x] T024 [US1] Implement the row loop in `internal/application/service/run_functional.go` and the row step in `internal/application/service/run_row.go`:
  - The row step loads the saved request at its turn (`FindRequest` and `Version(aliasId)`; a missing request gives a `RUN-06404` row error).
  - It calls the core with a `RunContext`: no History, run jar seeded from the workspace jar, the confirmed acknowledgement.
  - It builds the `RunRowResult` with the sent request and console masked through `usecase.MaskSecrets`, then publishes progress.
  - At the end it computes the summary, saves the run, `Release`s the target, drops the run jar and publishes `Finished`.
- [x] T025 [US1] Implement `Pause`, `Resume` and `Stop` (a gate checked before each row; context cancel gives a `Cancelled` outcome mapped to `cancelled`) in `internal/application/service/manage_runner.go` and `run_functional.go`. Add `StopWorkspace(localId)` and call it from `internal/application/service/activate_workspace.go` and `remove_workspace.go`.
- [x] T026 [US1] Add the stored-run persistence needed to finish a run: `usecase.SaveRun` and `usecase.FindRuns` (`internal/application/usecase/save_run.go`, `find_runs.go`), with the model in `internal/integration/persistence/model/run.go` and the implementation in `internal/integration/persistence/run.go` (atomic write to `runs/<targetId>/<runId>.json`, per-workspace mutex). T019–T021 must go green.
- [ ] T027 [P] [US1] FR-001: add `'runner'` to `ItemSection` in `frontend/src/features/item-settings/model.ts`. Add a `runner` render prop and the Runner tab trigger last in `frontend/src/features/item-settings/components/ItemTab.tsx`, with cases in `ItemTab.test.tsx`. Compose `RunnerSection` in `frontend/src/app/layout/ItemWorkspace.tsx`.
- [ ] T028 [US1] FR-001, FR-009, SC-001 (Runner, then Start run): build `RunnerSection` (sub-tabs Setup / Results / Runs; Start and Run again switch to Results) in `frontend/src/features/runner/components/RunnerSection.tsx`, a read-only `RunOrderList` (numbers, folder path, method badge, name, all checked) and Start run, with `RunnerSection.test.tsx`.
- [ ] T029 [US1] Build `RunProgress` (status, current row and iteration, bar, counts, elapsed, Pause/Resume/Stop) and `RunResults` (summary strip and rows grouped by iteration, each expanding to `TestResultsTab`, script errors and console) in `frontend/src/features/runner/components/`, with tests. In `frontend/src/features/runner/store.ts`:
  - implement `start` (handling `SND-05409` with a confirm dialog that reuses the request-builder `ConfirmSendDialog` pattern, and `SND-06409` with the missing-secrets dialog whose **Run anyway** restarts with `acknowledge_missing_secrets`), `pause`, `resume`, `stop`, `applyProgress` and `applyFinished`;
  - wire the events in `frontend/src/app/bootstrap.ts`;
  - add `store.test.ts`.

**Checkpoint**: a collection runs end to end from its tab with live progress and results (MVP).

---

## Phase 4: User Story 2 — Choose and order what runs (P1)

**Goal**: selection, the select-all checkbox, the counter, Reset, reordering (pointer and
keyboard), alias choice per row, all remembered locally.

**Independent Test**: save preferences, reload the plan, check that the run uses that selection
and order, and that Reset restores the sidebar order.

- [ ] T030 [P] [US2] Write the red `test/integration/features/runner/runner_plan.feature` (`@runner @runner_plan`), covering:
  - FR-003: flattened order with subfolders, examples excluded.
  - FR-006: `SavePreferences` writes `runners` in `state.json`, and no workspace file changes.
  - FR-007: reconcile drops removed requests, inserts added ones checked after the preceding sibling, and keeps the rest.
  - FR-008: the alias choice is kept; a removed alias gives `""`.
  - FR-013: the config is remembered; a missing data file is cleared and `data_file_missing` is true.
  - `RUN-07404`.

  Prove it red.
- [ ] T031 [US2] Implement `ReconcileRunRows` in `internal/domain/entity/run_plan_row.go`, with unit tests covering removed, added-at-sibling, added-first and stale-alias cases.
- [ ] T032 [US2] Add `Runners map[string]RunnerPreferences` to `internal/domain/entity/ui_state.go`, the model field `runners` (`omitempty`, older files load) to `internal/integration/persistence/model/ui_state.go`, and `runners` to `internal/integration/entrypoint/dto/ui_state.go`. `UIStateService.Save` must preserve `runners`.
- [ ] T033 [US2] Implement `manageRunner.Plan` and `SavePreferences` in `internal/application/service/manage_runner.go`:
  - reconcile the rows;
  - include the environments and the active run;
  - clear a missing data file (`data_file_missing`);
  - delete preferences of targets that no longer exist in the tree when the plan is read.

  T030 must go green.
- [ ] T034 [US2] FR-004, FR-005, FR-008: make `RunOrderList` interactive in `frontend/src/features/runner/components/RunOrderList.tsx`:
  - row checkboxes, a select-all checkbox with `aria-checked="mixed"`, "N of M selected" and Reset;
  - a per-row alias select (Default first);
  - dnd-kit `SortableContext` with `PointerSensor` and `KeyboardSensor`, using `frontend/src/features/runner/hooks/use-run-order-reorder.ts` (modelled on `features/sidebar/hooks/use-environment-reorder.ts`, with announcements).

  Add `RunOrderList.test.tsx` and `use-run-order-reorder.test.ts`.
- [ ] T035 [US2] Implement `load`, `updateConfig` (`SavePreferences` debounced by 300 ms) and `reset` in `frontend/src/features/runner/store.ts`, with tests.

---

## Phase 5: User Story 3 — Run from a folder or the sidebar (P2)

**Goal**: folder tabs have the Runner too, and **Run…** in the sidebar opens it.

**Independent Test**: from a folder's context menu choose Run…. Its tab shows Runner with only
that folder's requests, and a run still applies the collection's scripts and auth.

- [ ] T036 [P] [US3] Add folder-target scenarios to `test/integration/features/runner/run_functional.feature` (only the folder's requests, including subfolders; the collection and folder auth and scripts still apply) and prove them red.
- [x] T037 [US3] Make target resolution in `manage_runner.go` accept folder ids, and make `FlattenRunRows` start at the folder. Each row's ancestors still come from `FindAncestors`. T036 must go green.
- [ ] T038 [P] [US3] FR-002, SC-001 (Run…, then Start run): add **Run…** after Open for containers, in both the context menu and the ⋯ menu, with a new `onRunItem(id, kind)` prop, in `frontend/src/features/sidebar/components/TreeItem.tsx`, with `TreeItem.test.tsx` cases. In `app/` (where `onOpenItem` is wired), implement it: `useTabsStore.openItemTab(target, 'pinned')`, then set the item section to `runner` through `useItemSettingsStore.setSection`, and select the sub-tab (Results while a run is active there, else Setup) in `useRunnerStore`.

---

## Phase 6: User Story 4 — Iterations, data files and run settings (P2)

**Goal**: iterations, CSV/JSON data, delay, stop on error, keep values, cookie settings, logs off,
and the environment override.

**Independent Test**: run with a 3-row CSV and check the 3 iterations, the row values, `pm.info`,
the delay, and the effects of the variable and cookie settings.

- [ ] T039 [P] [US4] Write the red `test/integration/features/runner/run_iterations.feature` (`@runner @run_iterations`), covering:
  - FR-018: `pm.info.iteration` and `iterationCount`, `pm.iterationData.get`, and `{{field}}` overriding an environment variable of the same name.
  - FR-019: the CSV header, extra and missing cells, a UTF-8 BOM, a JSON array with number, boolean, null and object values (typed in `pm.iterationData`, text form in `{{}}`), fewer iterations than rows, last-row reuse, the default count, and the `RUN-05400` reasons.
  - FR-022: mock API receive times are ≥ the delay apart, across iterations too, and the run finishes without a trailing delay.
- [ ] T040 [P] [US4] Write the red `test/integration/features/runner/run_settings.feature` (`@runner @run_settings`), covering:
  - FR-012: the environment override, the active environment unchanged, and `"none"`.
  - FR-023: stop on a network error and on a script error; a failed assertion continues.
  - FR-024: keep values on persists; off leaves `variables.json`, the vault and the workspace files unchanged while later rows see the value (including a `pm.vault.set` and a secret variable).
  - FR-025: stored cookies are sent, or skipped; save-cookies merges into the jar, or does not.
  - FR-027: logs off gives no `console:entries` and an empty row console.

  Prove T039 and T040 red.
- [ ] T041 [P] [US4] Add the `filters` field to `internal/integration/entrypoint/dto/pick_file.go` and pass it in `internal/infra/server/binding/dialog.go`. Extend `test/integration/features/dialogs/pick_file.feature` (red, then green) and `frontend/src/platform/services/dialogs.ts` (+ test).
- [ ] T042 [US4] Add `IterationData` (`Row`, `WrappedRow`, `Count`) in `internal/domain/entity/iteration_data.go` and `usecase.ParseIterationData` in `internal/application/usecase/parse_iteration_data.go`, with unit tests in `parse_iteration_data_test.go`. Add `usecase.FindIterationData` in `internal/application/usecase/find_iteration_data.go`, implemented in `internal/integration/persistence/iteration_data.go`.
- [ ] T043 [US4] Feed the iteration to the sandbox:
  - `internal/integration/webservice/dto/script_sandbox_info.go` takes the iteration from `ScriptInfo`;
  - `script_sandbox_context.go` gains `iterationData`;
  - `internal/integration/webservice/sandbox/shims/pm.js` fills its `iterationData` scope from `context.iterationData`;
  - `send_run.go` builds `ScriptInfo` and `ScriptContext` from the `RunContext`, and resolves request variables with `WithIterationData`.
- [ ] T044 [US4] Implement the iterations, the default count from the data rows, the delay (timer between the end of one row and the start of the next, honouring stop and pause), stop on error (`stopped_on_error`), and the keep-values overlay (`RunVariables` in `RunContext` when off) in `internal/application/service/run_functional.go`.
- [ ] T045 [US4] Implement the cookie settings in `internal/application/service/run_functional.go`:
  - seed the run jar, or leave it empty for skip-stored;
  - at the end, merge with `MergeCookies` into the workspace jar and `settle` when save-cookies is on;
  - always `DropJar`.

  Also implement logs-off (`MuteConsole`, no row collection). T039 and T040 must go green.
- [ ] T046 [US4] FR-009, FR-010, FR-012, FR-013: build `RunSettings` (Postman labels and defaults, inline `RUN-04400` field errors, data file pick with the CSV/JSON filter, a clear button, the `data_file_missing` banner) and `EnvironmentSelect` in `frontend/src/features/runner/components/`, with tests.

---

## Phase 7: User Story 5 — Script control flow (P2)

**Goal**: `pm.execution.setNextRequest` (and the legacy `postman.setNextRequest`) changes the run
flow, while a single send stays a no-op.

**Independent Test**: scripts calling setNextRequest by name, by id, with `null` and with an
unknown name give the expected request sequences.

- [ ] T047 [P] [US5] Write the red `test/integration/features/runner/run_control_flow.feature` (`@runner @run_control_flow`), covering:
  - FR-020: by name, by id, backwards, last call wins (including a pre-request call and a call in a skipped row), `null`, an unchecked row treated as not in the run, and an unknown target with its console warning text;
  - the next iteration restarts at row 1;
  - a single send stays a no-op;
  - legacy `postman.setNextRequest`.
- [ ] T048 [US5] Add `NextRequest` (`internal/domain/entity/next_request.go`), `ScriptRun.NextRequest` (`script_run.go`) and `SendExecution.NextRequest()` (`send_execution.go`, last set wins across the chains), with unit tests.
- [ ] T049 [US5] In `internal/integration/webservice/sandbox/pm_execution.go`, add the `setNextRequest` host function, which records `{Target, Stop}` on the execution and returns it in the sandbox result. Make `shims/pm.js` `setNextRequest(next)` call the host (a missing argument or `null` means stop), and make `shims/legacy.js` `postman.setNextRequest` call the same host.
- [ ] T050 [US5] Resolve the next row in `internal/application/service/run_functional.go`: by id, else by the first row with that name, else end the iteration with the console warning `Couldn't find request "<target>" in this run; the iteration ends.` A stop ends the iteration. T047 must go green.


- [ ] T050a [US5] SC-005: write `test/integration/features/runner/run_newman_parity.feature` (`@newman @runner_newman`, run by `make newman` in the CI `newman` job, not by `make bdd`). It runs a collection that uses `setNextRequest` (by name and `null`), `skipRequest`, `pm.iterationData` and `pm.info.iteration`, with a 3-row CSV, through `RunnerService.Start`. Then it exports the collection as Postman and runs Newman with the same data file against the same mock. The method, path, query and body sequence received by the mock must be identical. Extend the generic Newman step in `test/integration/steps/when_newman_steps_test.go` with an optional `with the data file "<path>"` variant that passes `-d` (and `tools/newman/run.mjs` accordingly). Prove it red before T049 and T050, then green.
---

## Phase 8: User Story 6 — Inspect, filter, re-run and export results (P2)

**Goal**: filters, the row detail with the sent request and the response, Run again, and JSON and
JUnit export.

**Independent Test**: with a passing and a failing row, check the filters and the detail, that Run
again reuses the settings, and the export file contents.

- [ ] T051 [P] [US6] Write the red `test/integration/features/runner/run_export.feature` (`@runner @run_export`), covering:
  - FR-037: JSON equals the stored run, masked.
  - JUnit suites per row and iteration, `<failure message>`, `<skipped/>`, and `<error>` for errored or cancelled rows.
  - SC-004: a `<testsuites>` root, and `tests`, `failures`, `errors`, `skipped` and `time` attributes on every suite, whose totals equal the run summary.
  - A performance run with `junit` gives `RUN-04400`.

  Also write the red `test/integration/features/runner/run_responses.feature` (`@runner @run_responses`) for FR-040: `RunnerService.Response` returns the response when persisted, and `RUN-01404` when not or after the run is deleted.
- [ ] T052 [US6] Add `usecase.BuildRunJUnit` in `internal/application/usecase/build_run_junit.go`, with unit tests. Implement `manageRunner.Export`, which writes through the existing export file writer (`internal/integration/persistence/export_file.go`).
- [ ] T053 [US6] Implement persisted responses: when `PersistResponses` is on, `RunContext.ResponseKey = RunResponseKey(runId, index)` and `sendRun.send` saves under that key. Implement `manageRunner.Response`, which finds the response by key or returns `RUN-01404`. T051 must go green.
- [ ] T054 [US6] FR-032 (virtualized list, with a test rendering 10,000 rows and only the visible ones mounted), FR-034, FR-035: build the filter chips (All / Passed / Failed / Skipped, per FR-034) in `RunResults.tsx`, and `RunRowDetail.tsx` (Request: method, URL, headers, body; Response: `StatusLine`, `ResponseHeaders`, and `ResponseBody` with `tabId` set to the run response key, or the "not persisted" note), with tests.
- [ ] T055 [US6] FR-036, FR-037: implement **Run again** (`start` with the stored run's target and config) and **Export** (`PickSaveFile`, then `RunnerService.Export`; JUnit hidden for performance runs) in `frontend/src/features/runner/store.ts` and the Results header, with tests.

---

## Phase 9: User Story 7 — Past runs (P3)

**Goal**: the Runs sub-tab lists, opens, deletes and clears stored runs (25 per target), plus the
finish toast and the tab indicator.

**Independent Test**: after 26 runs only 25 are listed and on disk. Each opens; Delete and Clear
work; nothing is in History or the workspace; workspaces stay isolated.

- [ ] T056 [P] [US7] Write the red `test/integration/features/runner/run_storage.feature` (`@runner @run_storage`), covering:
  - FR-038: the path, the 25 cap per target, and a collection and its folder counted apart.
  - FR-039: secrets masked in the URL, headers, body and console; no response in the file; a scan of the files for every secret value (SC-007).
  - FR-041: List newest first, with its fields.
  - FR-041a: Delete and Clear; Delete of an active run gives `RUN-02409`.
  - Workspace isolation.
  - Removing a workspace deletes its `runs/` folder.
  - Runs of a deleted folder are gone after the next `Plan`.
  - An unreadable run file is left out of `List`.
  - Secrets set by a script during the run, and secrets in test error messages, are masked.
- [ ] T057 [US7] Add the 25-per-target cap after each save in `internal/integration/persistence/run.go`, plus `DeleteRun`, `ClearRuns` and `DeleteWorkspaceRuns` (`internal/application/usecase/delete_runs.go`). Make `FindRuns` skip unreadable files with a logged warning. Call `DeleteWorkspaceRuns` from `internal/application/service/remove_workspace.go`. Make `Plan` delete the runs folders of targets that no longer exist. Implement `manageRunner.List`, `Delete` and `Clear`; each releases the persisted response keys of the deleted runs through `ReleaseResponse`. T056 must go green.
- [ ] T058 [US7] Build `RunsList.tsx` (newest first; the FR-041 fields; open switches to Results; a ⋯ menu with Delete) and `ClearRunsDialog.tsx` (`shared/ui/dialog`) in `frontend/src/features/runner/components/`, with tests.
- [ ] T059 [US7] FR-001a, FR-001b: add the running indicator to `frontend/src/features/tabs/components/TabItem.tsx`, reading `useRunnerStore.active` (pulsing accent dot, `aria-label` suffix), with `TabItem.test.tsx`. Add `RunFinishedToast.tsx`, rendered from `frontend/src/app/App.tsx` on `runner:finished` with **Open results** (focus or open the tab, set section `runner`, Results, that run; auto-dismiss after 8 s). Add tests in `RunFinishedToast.test.tsx` and `bootstrap.test.ts`. Restore the indicators after a reload with `RunnerService.Active`.

---

## Phase 10: User Story 8 — Performance run (P3)

**Goal**: virtual users following a load profile, with live metrics and final statistics.

**Independent Test**: a short run against a mock API with a fixed latency gives request counts,
percentiles, throughput and error rate within bounds, and the active-user curve follows the
profile.

- [ ] T060 [P] [US8] Write the red `test/integration/features/runner/run_performance.feature` (`@runner @run_performance`), with durations of a few seconds, covering:
  - FR-011: range and profile validation.
  - FR-042: scripts run; per-user variables and cookies; nothing persisted; no History.
  - FR-043: the active users per second follow each profile, read from the `runner:progress` metrics.
  - FR-044: `setNextRequest` is ignored with one run notice, data wraps around, and no `console:entries` are emitted.
  - FR-029: Pause on a performance run gives `RUN-04400`.
  - FR-045 and FR-046: counts, p90/p95/p99 within ±10% of the mock latency, error rate with 500s, status counts, and series length equal to the duration.
- [ ] T061 [P] [US8] Implement `LoadProfile.ActiveUsersAt` in `internal/domain/entity/load_profile.go` (Fixed, Ramp up, Spike centred, and Peak with equal ramps around the peak, per FR-043), with table-driven unit tests in `load_profile_test.go`.
- [ ] T062 [P] [US8] Implement `LatencyHistogram` (log-linear, ≤1% relative error, exact min, max, sum and count), `RequestStats`, `PerformanceSecond` and `PerformanceStats` (internal mutex) in `internal/domain/entity/`, with unit tests checking the percentiles against sorted samples.
- [ ] T063 [US8] Implement `internal/application/service/run_performance.go` (no Pause, so `Pause` on a performance run gives `RUN-04400`; no app console and no per-request console; `pm.test` not reported; the `setNextRequest` notice added once to `Run.Notices`; throughput = completed requests / elapsed seconds):
  - a 100 ms controller tick that starts or retires virtual users to `ActiveUsersAt`;
  - each user loops over the selected rows through `run_row.go` with its own `RunVariables` and jar `RunCookieJarID(localId, runId)+":"+vu`, never merged back;
  - `RunContext.Performance` set, so `setNextRequest` is ignored with a warning;
  - data rows wrap;
  - errors are not-sent or status ≥ 400;
  - one `runner:progress` per second with metrics;
  - at the end, store the stats and series and drop every user jar.

  T060 must go green.
- [ ] T064 [US8] Build `PerformanceSettings.tsx` (users, duration, profile and its fields with the FR-011 defaults recomputed until edited, the inline-SVG curve preview) and `PerformanceResults.tsx` (stats table and inline-SVG time series) in `frontend/src/features/runner/components/`, with tests.

---

## Phase 11: User Story 9 — AI clients run collections (P3)

**Goal**: the MCP `run` (start/stop) and `runs` (get/list) tools.

**Independent Test**: start a run over MCP, read it until it finishes, check the results match
`RunnerService.Get`, then stop another run.

- [ ] T065 [P] [US9] Write the red `test/integration/features/mcp/runner.feature` (`@mcp_runner`), covering:
  - FR-047: start with defaults, and with explicit rows, settings and environment by name; `get` until completed; `list`; `stop`.
  - The approval asked once for a confirm environment, with denial and timeout.
  - The errors `RUN-02409`, `RUN-03400` and `MCP-07400`.
  - Recorded changes: no `history`.
  - FR-048: a UI-started run is readable over MCP, and an MCP-started run appears in `RunnerService.Active`.
- [ ] T066 [US9] Add the optional `Summary` to `entity.AIApproval` (`internal/domain/entity/ai_approval.go`) and to `dto.AIApprovalRequest` (`internal/integration/entrypoint/dto/ai_approval_request.go`). Show it in place of the method and URL in the approval dialog (`frontend/src/features/ai-access/`), with a test.
- [ ] T067 [US9] Add `controller.mcpRun` (`internal/integration/entrypoint/controller/mcp_run.go`; destructive and open-world; resolves the target like `tree`; approval on `SND-05409` as in `mcp_send.go`; records changes) and `controller.mcpRuns` (`mcp_runs.go`; read-only; output through `guardAIOutputService`), with DTOs `dto/mcp_run.go` and `dto/mcp_runs.go` whose argument descriptions are filled per XEN-014.
- [ ] T068 [US9] Register the tools in `internal/infra/server/mcp/tool_run.go`, `tool_runs.go`, `controllers.go` and `routes.go`, and wire them in `internal/infra/dependency/server/mcp_routes.go` and the controller DI. T065 must go green.

---

## Phase 12: Polish & cross-cutting

- [ ] T069 Update `docs/prd/xenon-mvp-prd.md` (FR-049):
  - the non-goal line becomes "CLI (useful, but not needed to replace daily use)";
  - add a "Collection runner" section that summarises the spec;
  - change the L507 runner-only note to "work inside collection runs; outside a run they are no-ops";
  - add **Decisions (Oct 23, 2026)** listing every decision from the spec's decision tables and clarifications.
- [ ] T070 Run `make sort-imports`, then the full `make bdd` (every pre-existing scenario included), `cd frontend && npm run lint && npm run typecheck && npm test`, and `make validate`. All must be green.
- [ ] T071 Run `/refactor-diff` until clean, then re-run `make validate`.
- [ ] T072 Walk through quickstart.md's manual checks in `wails3 dev` (functional, order, data, control flow, keep values, performance, past runs, sidebar Run…, toast, MCP). Include SC-006: a 100-user, 60-minute performance run against a local mock, during which tabs open, edit and send without visible lag.

---

## Dependencies & execution order

- **Phase 1 → Phase 2** blocks everything. T008 (the shared core) is the critical path, and the
  single-send suites must stay green after it.
- **US1 (Phase 3)** depends on Phase 2 only. It is the MVP.
- **US2** depends on US1 (it extends `RunOrderList` and `manage_runner.go`).
- **US3, US4, US5, US6 and US7** each depend on US1 and can proceed in parallel after it:
  - US4 touches `run_functional.go` iterations and settings.
  - US5 touches control flow in the same file, so sequence US4 → US5 when one person does both.
  - US6 owns `run_export.feature` and `run_responses.feature`, and US7 owns `run_storage.feature`. Their feature files do not overlap.
- **US8** depends on US1 and US4 (overlay, jar seeding, data rows).
- **US9** depends on US1. Its approval part reuses the US1 preflight.
- **Polish** comes last.

## Parallel examples

- **Phase 2**: T002, T003, T005, T010, T011, T013, T017 and T018 can run together. After T004:
  T006, T007 and T012.
- **US1**: T019, T020 and T021 (feature files) together; then T027 (frontend) alongside T022–T026
  (backend).
- **US4**: T039, T040 and T041 together; T042 alongside T043.
- **US8**: T060, T061 and T062 together.

## Implementation strategy

1. MVP = Phase 1 + Phase 2 + US1. Demo a functional run from the collection tab.
2. Add US2 (selection and order), then US3 to US7 in priority order, keeping the full suite green
   after each story.
3. Add US8 (performance) and US9 (MCP).
4. Polish: PRD, validate, refactor-diff, then the manual quickstart.
