---

description: "Task list for XEN-022 Runs Are Visible While They Run"
---

# Tasks: Runs Are Visible While They Run

**Input**: Design documents from `specs/XEN-022-runs-visible-while-running/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/runs-list.md, contracts/runner-store.md

**Tests**: MANDATORY (constitution V, BDD-first). Every story starts with tests proven red before implementation.

**Organization**: grouped by user story; US1 is backend-only, US2 and US3 are frontend-only, so US2 and US3 can be built in parallel with US1.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)

---

## Phase 1: Setup

**Purpose**: a working tree that can run every suite.

- [ ] T001 Fresh worktree on branch `xen-022` (from freshly fetched `origin/main`): run `npm ci` in `frontend/` and `make bindings` so Vitest and the BDD harness start (see `.claude/skills/makefile.md`)

---

## Phase 2: Foundational

No shared prerequisite: US1 touches only Go (`internal/`, `test/integration/`), US2/US3 only `frontend/src/features/runner/`. Nothing blocks the stories.

---

## Phase 3: User Story 1 — Runs history and `runs list` include in-progress runs (Priority: P1) 🎯 MVP

**Goal**: `RunnerService.List` and MCP `runs list` return the target's in-progress run with its so-far summary, first, together with the stored runs, each run once (contract `contracts/runs-list.md`).

**Independent Test**: start a slow run, list the target's runs before it finishes → it is element 0 with status `running`; after `runner:finished` it is listed exactly once.

### Tests for User Story 1 (write first, prove red)

- [ ] T002 [P] [US1] Add scenario `@xen_022 @in_progress_listed` to `test/integration/features/runner/run_storage.feature`: one finished stored run of `{{shop}}` (saved as `older`); `Given the "POST" requests to "/orders" respond after 3000ms`; start a run with rows `{{login}}` then `{{create}}` (saved as `run`); `When I wait 300 ms`; `RunnerService.List` `{ local_id, item_id: {{shop}} }` → array length 2, `0.id` = `{{run}}`, `0.status` = `running`, `0.summary.total_requests` = `1`, `1.id` = `{{older}}`; `RunnerService.Get` of `{{run}}` → `status` = `running`; then `RunnerService.Stop` `{{run}}` and wait for `runner:finished`
- [ ] T003 [P] [US1] Add scenario `@xen_022 @paused_listed` to `test/integration/features/runner/run_storage.feature`: slow `/orders` (3000ms), start rows `{{login}}`, `{{create}}`, `{{get}}`; wait 300 ms; `RunnerService.Pause`; `Then a "runner:progress" event with the field "status" equal to "paused" should have been emitted within 4000 ms` (the pause lands once the 3000 ms `/orders` request in flight completes, as in `run_lifecycle.feature` `@pause_resume`); `RunnerService.List` → `0.id` = run, `0.status` = `paused`; then stop it and wait for `runner:finished`
- [ ] T004 [P] [US1] Add scenario `@xen_022 @listed_once` to `test/integration/features/runner/run_storage.feature`: start one run of `{{shop}}`, wait for `runner:finished`, `RunnerService.List` → array length 1, `0.id` = run, `0.status` = `completed`
- [ ] T005 [P] [US1] Add scenario `@xen_022 @in_progress_other_target` to `test/integration/features/runner/run_storage.feature`: slow `/orders`, run of `{{shop}}` in progress; `RunnerService.List` `{ local_id, item_id: {{orders}} }` (the background's `orders` folder inside Shop) → array length 0; stop the run and wait for `runner:finished`. (FR-012's other-workspace half needs no new scenario: `FindByWorkspace(localId)` scopes it and activating another workspace already stops runs — `run_lifecycle.feature`.)
- [ ] T006 [P] [US1] Add scenario `@xen_022 @list_in_progress` to `test/integration/features/mcp/runner.feature`: `Given the "POST" requests to "/orders" respond after 3000ms`; MCP `run` `{ "action": "start", "target": "Shop" }` (save `run_id` as `run`); wait 300 ms; MCP `runs` `{ "action": "list", "target": "Shop" }` → `0.id` = `{{run}}`, `0.status` = `running`; MCP `runs` `{ "action": "get", "run_id": "{{run}}" }` → `status` = `running`; MCP `run` stop and wait for `runner:finished`
- [ ] T007 [P] [US1] Create `internal/domain/entity/run_list_test.go` (table tests, `package entity` like the other `*_test.go` there): `ListRuns` returns stored + the target's active run sorted by `StartedAt` desc (equal `StartedAt` keeps the stored order); drops an active run whose id is stored (stored copy kept); ignores active runs of other item ids; returns `[]Run{}` (non-nil) for no runs; an in-progress listed run's `Summary.TotalRequests` equals its recorded rows and `DurationMs` its `Progress.ElapsedMs`; `Listed()` on a finished run returns it unchanged; an in-progress performance run with a `Performance` report lists with the summary's performance figures (`WithPerformance(Performance.Overall)`), one without a report lists without them
- [ ] T008 [US1] Run `make bdd TAG=@xen_022` and `go test ./internal/domain/entity/ -run 'ListRuns|Listed'`; confirm T002, T003 and T006 fail on the missing in-progress run and T007 fails to compile; T004 and T005 already pass (they guard regressions)

### Implementation for User Story 1

- [ ] T009 [US1] In `internal/domain/entity/run.go` extract the summary computation of `Finish` into private `func (r *Run) summarize()` (`NewRunSummary(r.Iterations(), r.Rows, r.Progress.ElapsedMs)` plus `WithPerformance(r.Performance.Overall)` when `r.Performance != nil`) and call it from `Finish`
- [ ] T010 [US1] In `internal/domain/entity/run.go` add `func (r Run) Listed() Run` (copy; `summarize()` applied when `!IsFinished()`) and package-level `func ListRuns(targetId string, active, stored []Run) []Run` next to `RunsBeyondStoredLimit` (stored ∪ active with `Target.ItemID == targetId` and id not stored, as `Listed()`, `slices.SortStableFunc` by `StartedAt` desc, never nil); follow `.claude/skills/entities.md` and `comments.md`
- [ ] T011 [US1] In `internal/application/service/find_runs.go` make `List` return `entity.ListRuns(itemId, f.findActiveRuns.FindByWorkspace(ctx, localId), runs)` and `ListForAI` return `entity.ListRuns(targetId, f.findActiveRuns.FindByWorkspace(ctx, localId), runs)`; keep the `Finished` log line counting the merged runs (`.claude/skills/services.md`)
- [ ] T012 [US1] Run `make sort-imports`, then `make bdd TAG=@xen_022` and the entity tests until green

**Checkpoint**: US1 complete — `runs list` and Runs history (on next read) include in-progress runs.

---

## Phase 4: User Story 2 — The open Runner tab follows a run started elsewhere (Priority: P1)

**Goal**: a tab whose Runner was read this session switches to Results of a newly started run of its target, once, and shows it live (contract `contracts/runner-store.md`).

**Independent Test**: Vitest — tab showing run `R1`; progress of new `R2` → `runId` `R2`, `view` `results`, `getRun('L1','R2')` called.

### Tests for User Story 2 (write first, prove red)

- [ ] T013 [P] [US2] In `frontend/src/features/runner/store.test.ts` add: "follows a run of the target started elsewhere" — loaded target, `openRun('c1','R1')` (stored), then `applyProgress` for `R2` of `c1` → `target().runId === 'R2'`, `target().view === 'results'`, `service.getRun` called with `R2`
- [ ] T014 [P] [US2] In `frontend/src/features/runner/store.test.ts` add: "follows a new run once; an older run opened during it stays shown" — after following `R2`, `openRun('c1','R1')`, another `R2` progress → `runId` stays `R1`
- [ ] T015 [P] [US2] In `frontend/src/features/runner/store.test.ts` add: "a run started elsewhere is followed from Setup and Runs" (view `setup` / `runs` → `results`), "progress of another target leaves the tab alone", and "progress after the run finished does not switch the tab" (`applyFinished` for `R2` then late `R2` progress on a tab showing `R1` → stays `R1`)
- [ ] T016 [P] [US2] In `frontend/src/features/runner/store.test.ts` add: "pause and stop act on the followed run" — after following `R2`, `pause('c1')` / `resume('c1')` / `stop('c1')` call `service.pauseRun` / `service.resumeRun` / `service.stopRun` with `R2`; and "a followed run's finish reads it again and shows the notice" — `applyFinished` for `R2` → `service.getRun` called with `R2` again and `notice.runId === 'R2'`
- [ ] T017 [US2] Run `cd frontend && npx vitest run src/features/runner/store.test.ts`; confirm T013–T015's switching cases fail (T015's "another target" and T016 may already pass — regression guards)

### Implementation for User Story 2

- [ ] T018 [US2] In `frontend/src/features/runner/store.ts` `applyProgress`: compute `const starting = known?.id !== event.runId` before `setActive`; after the idle guard, when `target.runId !== event.runId`, if `target.runId === null || starting` then `patch(event.itemId, { runId: event.runId, run: null, view: 'results' })` and `void refreshRun(event.itemId, event.runId)`, and return either way; drop the old `runId === null` patch and `(target.runId ?? event.runId)` guard it replaces; keep the same-run branch unchanged (`.claude/skills/frontend-architecture/SKILL.md`)
- [ ] T019 [US2] Run the runner Vitest suite until green, including the pre-existing "ignores progress of another workspace and adopts a run started elsewhere" and "does not bring back the indicator of a run that ended before Start answered" cases

**Checkpoint**: US2 complete — AI-started runs show live in their target's open tab.

---

## Phase 5: User Story 3 — The Runs view keeps up while a run is in progress (Priority: P2)

**Goal**: the Runs list is read again when a run it lacks reports progress; in-progress rows offer no Delete; Clear keeps the in-progress run and is disabled with nothing stored.

**Independent Test**: Vitest — loaded `runs` without `R2`, `R2` progress → `service.listRuns` called again; `RunsList` with a Running row renders no "More actions" button.

### Tests for User Story 3 (write first, prove red)

- [ ] T020 [P] [US3] In `frontend/src/features/runner/lib/lib.test.ts` add `isRunInProgress` cases: `running`, `paused` → true; `completed`, `stopped`, `stopped_on_error` → false
- [ ] T021 [P] [US3] In `frontend/src/features/runner/store.test.ts` add: "reads the Runs list again when a run it lacks reports progress" (after `loadRuns('c1')` returning `[R1]`, `R2` progress → `service.listRuns` called a second time; a further `R2` progress once `runs` holds `R2` → no extra call) and "Clear runs keeps the in-progress run" (`runs` `[R2 running, R1 completed]`, `clearRuns('c1')` → `runs` = `[R2]`)
- [ ] T022 [P] [US3] In `frontend/src/features/runner/components/RunsList.test.tsx` add: a Running and a Paused row render no "More actions" button, a Completed row does; with only a Running row, **Clear runs** is disabled; with a Running and a Completed row it is enabled
- [ ] T023 [US3] Run `cd frontend && npx vitest run src/features/runner`; confirm T020–T022 fail

### Implementation for User Story 3

- [ ] T024 [US3] Add `export function isRunInProgress(run: { status: RunStatus }): boolean` (`running` or `paused`) to `frontend/src/features/runner/lib/results.ts`
- [ ] T025 [US3] In `frontend/src/features/runner/store.ts` `applyProgress`, right after the idle guard and before the `target.runId !== event.runId` branch added in T018: `if (target.runs !== null && !target.runs.some((run) => run.id === event.runId)) void get().loadRuns(event.itemId)`; in `clearRuns` success set `runs: target.runs?.filter(isRunInProgress) ?? []` (shown-run reset rule unchanged)
- [ ] T026 [US3] In `frontend/src/features/runner/components/RunsList.tsx`: render `RunItem`'s `ContextMenu` only when `!isRunInProgress(run)`; set Clear runs `disabled={!runs.some((run) => !isRunInProgress(run))}`; update the component doc comment and `NO_RUNS` description to "Runs of this collection or folder show up here."
- [ ] T027 [US3] Run the runner Vitest suite until green

**Checkpoint**: all stories complete.

---

## Phase 6: Polish & Cross-Cutting

- [ ] T028 Run the full gates: `make validate` (Go unit + full BDD suite) and in `frontend/` `npm run lint`, `npm run typecheck`, `npm test`; every pre-existing scenario stays green (notably `run_storage.feature` `@newest_first`, `@cap`, `mcp/runner.feature` list/stop scenarios, `run_lifecycle.feature` `@active`)
- [ ] T029 Walk through `specs/XEN-022-runs-visible-while-running/quickstart.md` "By hand" steps 1–6 in the desktop app (`make dev`) and record the outcome in the PR description
- [ ] T030 `/refactor-diff` until clean; never commit `REFACTORING_REPORT.md`

---

## Dependencies & Execution Order

- **Setup (T001)** → everything.
- **US1 (T002–T012)**: tests T002–T007 in parallel → T008 red → T009 → T010 → T011 → T012.
- **US2 (T013–T019)**: independent of US1 (frontend only; Vitest mocks `platform/`). T013–T016 in parallel → T017 → T018 → T019.
- **US3 (T020–T027)**: after US2's T018 (both edit `applyProgress` in `store.ts`). T020–T022 in parallel → T023 → T024 → T025 → T026 → T027. Its effect in the app relies on US1 (the re-read list must contain the in-progress run).
- **Polish (T028–T030)** after all stories.

## Parallel Example

```text
# US1 tests together:
T002, T003, T004, T005 (run_storage.feature — one editor, sequential within the file)
T006 (mcp/runner.feature) ‖ T007 (run_list_test.go)
# US1 and US2 in parallel: Go work (T009–T011) ‖ store.ts follow (T018)
```

## Implementation Strategy

1. **MVP = US1**: fixes `runs list`/`get` agreement for AI clients and the Runs history; shippable alone.
2. **US2**: the open tab follows AI-started runs — completes the issue's report.
3. **US3**: Runs view keeps up live and stops offering refused actions.
4. One PR for the whole task (spec + tests + implementation), merged on green CI.
