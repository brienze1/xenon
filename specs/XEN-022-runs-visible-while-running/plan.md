# Implementation Plan: Runs Are Visible While They Run

**Branch**: `xen-022` | **Date**: 2026-10-27 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/XEN-022-runs-visible-while-running/spec.md`

## Summary

Listing a target's runs reads only stored runs, and the Runner tab ignores runs it did not start.
Two small, independent fixes:

1. **Backend (Go)** — `FindRunsService.List` and `ListForAI` merge the target's in-progress run
   (from `FindActiveRuns.FindByWorkspace`) with its stored runs. A new domain function
   `entity.ListRuns(targetId, active, stored)` filters the active runs to the target, drops any
   active run whose id is already stored (the stored copy wins), gives each in-progress run its
   so-far summary, and sorts by `StartedAt` newest first. Both the Wails `RunnerService.List`
   binding and the MCP `runs list` tool go through these two methods, so both are fixed.
2. **Frontend (React/Zustand)** — `applyProgress` in `features/runner/store.ts` makes an open tab
   follow a newly started run of its target (switch `runId`, `view: 'results'`, read the run), and
   reads the Runs list again when it lacks the run. `clearRuns` keeps the in-progress run;
   `RunsList` offers no Delete on an in-progress run and disables Clear runs when nothing is
   stored.

## Technical Context

**Language/Version**: Go (version per `go.mod`), TypeScript 5 / React 19

**Primary Dependencies**: Wails v3 (bindings/events), Zustand — no new dependency

**Storage**: unchanged — stored runs in app-data `workspaces/<localId>/runs/<itemId>/<runId>.json`; in-progress runs in memory (`activeRunsPersistence`)

**Testing**: Godog BDD (`make bdd TAG=…`), Go unit tests (`internal/domain/entity/*_test.go`), Vitest (`frontend/src/features/runner/*.test.ts(x)`)

**Target Platform**: macOS, Windows, Linux desktop

**Project Type**: desktop app (Go backend + React UI in Wails)

**Performance Goals**: listing adds one in-memory scan of the workspace's active runs (at most one per target) — negligible

**Constraints**: one in-progress run per target (enforced by `activeRunsPersistence.Claim`, `RUN-02409`)

**Scale/Scope**: ≤ 25 stored runs + ≤ 1 in-progress run per target

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first** — no new state; runs stay in app-data / memory; workspace isolation kept: `FindByWorkspace(localId)` + target filter (FR-012). An in-progress run of another workspace cannot coexist anyway — activating a workspace stops the previous one's runs (`run_lifecycle.feature` "Activating another workspace stops the active run"); the tab side is locked by the existing Vitest "ignores progress of another workspace".
- [x] **II. Secrets** — a list returns only run summary views (`dto.NewRunSummaryViews`: id, target, type, status, environment name, times, progress (current request name and counts — already published by every `runner:progress` event and by `RunnerService.Active`), summary counts) — never rows, requests, responses or variable values. In-progress rows are masked at finish (`run_functional.go`), but they never reach a list, so listing an in-progress run exposes nothing new. `get` of an in-progress run is unchanged by this feature.
- [x] **III. Go owns logic** — merge/dedupe/sort/summary in Go domain; React only reacts to events and calls existing bindings.
- [x] **IV. Architecture skills** — domain function + method in `internal/domain/entity/run.go` (`entities.md`), service change in `internal/application/service/find_runs.go` (`services.md`); frontend changes stay inside `features/runner/` (`frontend-architecture`).
- [x] **V. BDD-first** — feature files named below; only HTTP targets / clock faked as today.
- [x] **VI. Postman parity** — Postman's runner shows the run in progress in its run view; nothing beyond the MVP.
- [x] **VII. Cross-platform** — no OS-specific code.

## Design

### Backend

`internal/domain/entity/run.go`

- Extract the summary computation from `Finish` into a private `(r *Run) summarize()` (rows,
  iterations, `Progress.ElapsedMs`, performance overall when present). `Finish` calls it — behaviour
  unchanged.
- `func (r Run) Listed() Run` — value receiver, returns a copy; for a run that `IsFinished()` it
  returns it unchanged, otherwise the copy with `summarize()` applied (so-far summary, FR-003).
- `func ListRuns(targetId string, active, stored []Run) []Run` — package-level, next to
  `RunsBeyondStoredLimit`:
  1. start with `stored` (already newest first);
  2. append each `active` run whose `Target.ItemID == targetId` and whose `ID` is not in `stored`,
     as `run.Listed()`;
  3. `slices.SortStableFunc` by `StartedAt` descending (stable keeps the stored order for equal
     times);
  4. never returns nil (`[]Run{}` when empty — DTO maps to `[]`).

`internal/application/service/find_runs.go`

- `List`: after `FindRuns`, `runs = entity.ListRuns(itemId, f.findActiveRuns.FindByWorkspace(ctx, localId), runs)`.
- `ListForAI`: same with `targetId`.
- No new dependency (`findActiveRuns` is already injected); DI unchanged.

Unchanged and confirmed: `Get` (active first, then stored — any run it returns is in-progress or
stored, both now listed: FR-005), `mcp_runs.go` controller, `RunnerService.List` binding, DTOs.

### Frontend

`frontend/src/features/runner/lib/results.ts`: `isRunInProgress(run: { status: RunStatus }): boolean` — `running` | `paused` (pure helper, unit-tested in `lib/lib.test.ts`).

`frontend/src/features/runner/store.ts`

- `applyProgress`, after the existing `finishedRuns`/workspace guard and `setActive`:
  - `starting = known?.id !== event.runId` (computed before `setActive`) — first report of a run
    this app does not yet hold as the target's active run.
  - target idle/unknown → return (unchanged).
  - if `target.runs !== null && !target.runs.some((r) => r.id === event.runId)` → `void get().loadRuns(itemId)` (FR-008).
  - if `target.runId !== event.runId`: when `target.runId === null || starting` →
    `patch(itemId, { runId: event.runId, run: null, view: 'results' })` and
    `void refreshRun(itemId, event.runId)`, then return (FR-006). Otherwise (user opened an older
    run during the run) → return without switching.
  - the existing same-run branch (patch progress, refresh functional) stays.
- `clearRuns` success: `runs: target.runs?.filter(isRunInProgress) ?? []` (FR-010); the shown-run
  reset rule is unchanged.
- `start()` already sets `active` before progress can be "starting" for the tab's own run; if a
  progress event wins the race the follow sets the same `runId` — idempotent.

`frontend/src/features/runner/components/RunsList.tsx`

- `RunItem`: render the actions menu (whose only item is Delete) only when
  `!isRunInProgress(run)` (FR-009).
- Clear runs `disabled={!runs.some((run) => !isRunInProgress(run))}` (FR-010).
- Empty-state and doc comment wording: "Runs of this collection or folder show up here."

## Tests (BDD-first)

| Requirement | Test |
| --- | --- |
| FR-001, FR-002, FR-005 | `test/integration/features/runner/run_storage.feature` — new `@in_progress_listed`: a stored run, then a slow run in progress → `RunnerService.List` length 2, `0.id` = in-progress, `0.status` = `running`, `1.id` = stored; `RunnerService.Get` of it agrees |
| FR-001 (paused) | same file, `@paused_listed`: pause the run → listed with `paused` |
| FR-003 | same scenario: rows `/login` (fast) then `/orders` (slow); after 300 ms the in-progress run lists `0.summary.total_requests` = 1 while still `running` |
| FR-004 | same file, `@listed_once`: after `runner:finished`, list has the run exactly once (length = stored count) with final status |
| FR-012 | same file, `@in_progress_other_target`: run on Shop in progress → List of the folder / another collection does not include it |
| FR-001 over MCP | `test/integration/features/mcp/runner.feature` — new `@list_in_progress`: MCP `run start` on a slow target, then `runs list` → `0.id` = run, `0.status` = `running`; `runs get` agrees |
| FR-006, FR-007, FR-011 | Vitest `frontend/src/features/runner/store.test.ts`: a tab showing `R1` follows progress of new `R2` (runId, view, getRun called); after `openRun(R1)` further `R2` progress keeps `R1`; progress after finish does nothing; pause/stop act on the followed run |
| FR-008 | Vitest `store.test.ts`: loaded Runs list without `R2` is read again on `R2` progress, not again for a run already listed |
| FR-010 | Vitest `store.test.ts`: clearRuns keeps the running entry; `RunsList.test.tsx`: Clear disabled when only an in-progress run is listed |
| FR-009 | Vitest `RunsList.test.tsx`: no "More actions" button on a Running/Paused row; present on a Completed row |
| FR-013 | No new test: the list response type (`dto.RunSummaryView`) is unchanged and carries no row/request/response/variable fields (Constitution Check II) |
| FR-003 (unit) | Go `internal/domain/entity/run_list_test.go`: `ListRuns` dedupe (stored wins), sort, target filter, so-far summary; `Listed` on a finished run is identity |

Every new scenario is tagged `@xen_022` plus its own tag (repo convention, e.g. `@xen_020 @reconcile`).

Only the HTTP target (slow responses via existing "respond after Nms" step) is faked. All steps
already exist in `test/integration/steps/` (`I call "RunnerService.*"`, `I call the MCP tool`,
`respond after`, `a "runner:…" event … within`, `response should contain the field … array length`).

## Project Structure

### Documentation (this feature)

```text
specs/XEN-022-runs-visible-while-running/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── runs-list.md
│   └── runner-store.md
├── checklists/
└── tasks.md
```

### Source Code (repository root)

```text
internal/domain/entity/run.go                    # summarize(), Listed(), ListRuns()
internal/domain/entity/run_list_test.go          # new unit tests
internal/application/service/find_runs.go       # List / ListForAI merge
test/integration/features/runner/run_storage.feature   # new scenarios
test/integration/features/mcp/runner.feature            # new scenario
frontend/src/features/runner/lib/results.ts      # isRunInProgress
frontend/src/features/runner/lib/lib.test.ts
frontend/src/features/runner/store.ts            # applyProgress follow, clearRuns keep
frontend/src/features/runner/store.test.ts
frontend/src/features/runner/components/RunsList.tsx
frontend/src/features/runner/components/RunsList.test.tsx
```

**Structure Decision**: existing layout; no new packages or folders.

## Complexity Tracking

None — no constitution violations.
