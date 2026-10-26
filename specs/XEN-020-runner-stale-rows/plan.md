# Implementation Plan: Runner Follows the Current Tree

**Branch**: `xen-020` | **Date**: 2026-10-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/XEN-020-runner-stale-rows/spec.md`

## Summary

The Runner's Setup rows freeze at the first read of a session because the frontend Runner store
(`frontend/src/features/runner/store.ts`) loads each target once (`ensure`) and nothing re-reads
it; runs started from the Runner send those frozen rows. The backend already reconciles the
remembered rows against a fresh disk scan on every `RunnerService.Plan` and every MCP start
(research R1). The fix is frontend-only: a new `refreshAll()` action on the Runner store,
called from the existing sidebar-tree subscription in `frontend/src/app/bootstrap.ts` next to
`useItemSettingsStore.refreshAll()` and `useDocumentationStore.refreshAll()` (R2). A refresh
first remembers any pending Runner edit, then re-reads the plan; reads are coalesced and
sequence-guarded so the last tree change always wins and no user edit is lost (R4, R5). A failed
refresh shows the existing load-error banner (R6). Two Godog scenarios lock the backend
behaviour the fix relies on (R7).

## Technical Context

**Language/Version**: TypeScript 5.9 (React 19.3 + Vite) for the change; Go 1.26 backend unchanged

**Primary Dependencies**: Zustand 5 (Runner and sidebar stores), Vitest + Testing Library; Godog for the guard scenarios

**Storage**: N/A — remembered rows stay in the workspace's local UI state (`state.json` in app data) via the existing `RunnerService.SavePreferences`; nothing new is stored

**Testing**: Vitest (`frontend/src/features/runner/store.test.ts`, `frontend/src/app/bootstrap.test.ts`) — red first; Godog (`make bdd TAG=@xen_020`) — regression guards, green before and after

**Target Platform**: Xenon desktop (Wails v3) on macOS, Windows, Linux

**Project Type**: desktop-app (Go core + React webview)

**Performance Goals**: SC-002 — an open Runner shows the new contents within 1 s of the sidebar showing them (one `planRun` per known target per coalesced change; ~100-row folder reads ~100 request files, well under 1 s)

**Constraints**: at most one re-read in flight per target plus one trailing; no client-side row merging (Go owns reconciliation); no change to view, shown run or active run on refresh

**Scale/Scope**: targets already read in one session (bounded by collection/folder tabs opened); folders of ~100 requests (the report: 93)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first** — no new state; remembered rows stay in local UI state; workspace isolation kept (refresh only touches the active workspace's Runner store, reset on switch — R3).
- [x] **II. Secrets** — no values are read, logged or stored; rows carry request ids, names, methods and alias ids only.
- [x] **III. Go owns logic** — reconciliation stays in `entity.ReconcileRunRows`; the UI only asks `planRun` again through `platform/services/runner`.
- [x] **IV. Architecture skills** — store action in `features/runner/store.ts`; cross-store wiring in `app/bootstrap.ts` (the composition point, per `frontend-architecture` "subscriptions are opened once in `app/bootstrap.ts`"); no Go file changes.
- [x] **V. BDD-first** — Vitest specs written red first for every FR the frontend implements; Godog guards for FR-002/FR-007 (see BDD map below; they are green on `main` by design, R7).
- [x] **VI. Postman parity** — Postman's Runner always reflects the collection's current contents; this restores that; nothing beyond the PRD.
- [x] **VII. Cross-platform** — pure TypeScript store logic; no paths, shortcuts or OS APIs.

Post-design re-check (after Phase 1): all seven still ✅ — the design added no files outside the two frontend modules and their tests plus two feature-file scenarios.

## Requirement → layer → test map

| FR | Where it is implemented | Locked by |
| --- | --- | --- |
| FR-001 refresh known targets on tree change | `app/bootstrap.ts` sidebar subscription → `useRunnerStore.refreshAll()` | `bootstrap.test.ts` |
| FR-002 reconciliation rules | Backend, unchanged: `entity.ReconcileRunRows` via `run_planner.go` `plan()` | `runner_plan.feature` `@xen_020` scenario + existing `@reconcile` |
| FR-003 pending edit remembered first | `runner/store.ts` `refresh()` flushes `saveTimers` → `save()`; edit sequence discards stale responses | `store.test.ts` |
| FR-004 coalescing | `runner/store.ts` per-target in-flight + trailing flag, read sequence | `store.test.ts` |
| FR-005 view/run untouched | `refresh()` patches only `status: 'ready'`, `target`, `rows`, `config`, `environments`, `activeEnvironmentId`, `dataFileMissing`, `loadError`; `setActive` from the plan | `store.test.ts` |
| FR-006 Runner run sends refreshed rows | `start()` unchanged — reads `target.rows`, now current | `store.test.ts` |
| FR-007 headless start uses reconciled rows | Backend, unchanged: `start_run.go` `StartForAI` → `planner.plan` | `mcp/runner.feature` `@xen_020` scenario |
| FR-008 other workspaces ignored | Sidebar tree only holds the active workspace; `inWorkspace` drops cross-switch responses (R3) | `store.test.ts` (response after switch ignored) |
| FR-009 nothing in git-tracked files | Unchanged `SavePreferences` path (local UI state) | existing `runner_plan.feature` "never in the workspace files" scenarios |
| FR-010 failed refresh shows load error | `refresh()` failure clears `config`/`rows`/`target`, sets `loadError`; `RunnerSection` shows banner + Try Again | `store.test.ts` |

## Design

See [data-model.md](./data-model.md) for the store state and [contracts/runner-store.md](./contracts/runner-store.md)
for the action contract. In short:

1. `features/runner/store.ts`
   - Module-level maps beside `saveTimers` / `refreshing`: `reads: Map<string, number>` (latest
     read sequence per target), `edits: Map<string, number>` (edit sequence per target),
     `rereading: Map<string, boolean>` (in-flight refresh; `true` = one more requested).
     `clearTimers()` (called by `reset()`) clears them too.
   - `scheduleSave(itemId)` bumps `edits` for the target.
   - `load(itemId)` takes a read sequence; it applies its response only if it is still the
     latest read for the target (existing behaviour otherwise unchanged).
   - New internal `flushSave(itemId)`: if a save timer is pending, clear it and `await save(itemId)`.
   - New internal `refresh(itemId)`: if a refresh is in flight, mark trailing and return; else
     `await flushSave`, note `edits` and take a read sequence, call `planRun`; on response drop it
     if the workspace changed (`inWorkspace` → `null`), the target was reset, a newer read
     exists, or `edits` moved (then mark trailing); on success patch the plan fields (FR-005
     list) and `setActive(itemId, plan.activeRun)`; on failure patch
     `{ config: null, rows: [], target: null, loadError }`; finally run the trailing refresh if
     marked.
   - New action `refreshAll()`: `refresh` every target whose `status !== 'idle'`, in parallel.
2. `app/bootstrap.ts`: inside the existing `useSidebarStore.subscribe` branch
   `state.tree !== previous.tree && state.status === 'ready'`, add
   `void useRunnerStore.getState().refreshAll()`.
3. No change to `RunnerSection.tsx`: its banner already renders when `config == null && loadError`.
4. No Go change.

## BDD feature files (written by `/create-tests`)

- `test/integration/features/runner/runner_plan.feature` — add `@xen_020` scenario: *"A
  subfolder and requests added after the rows were remembered join selected, in sidebar order"*
  (remember rows via `RunnerService.SavePreferences`, write a new subfolder with two requests
  and one request into an existing folder, `RunnerService.Plan` → all rows, new ones selected,
  placed after the nearest preceding request in sidebar order, an unchecked row stays unchecked).
- `test/integration/features/mcp/runner.feature` — add `@xen_020` scenario: *"A run started
  without rows includes requests added after the rows were remembered"* (remember rows with one
  row unchecked, add a request to the target folder, MCP `run` `start` without `rows`, read the
  run: the new request ran, the unchecked one did not).
- Frontend (Vitest, red first): `frontend/src/features/runner/store.test.ts` — new
  `describe('refreshAll')`; `frontend/src/app/bootstrap.test.ts` — the sidebar tree change calls
  `useRunnerStore.refreshAll`.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-020-runner-stale-rows/
├── spec.md
├── plan.md              # this file
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── runner-store.md
├── checklists/
│   └── requirements.md
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
frontend/src/
├── app/
│   ├── bootstrap.ts                 # + runner refreshAll on sidebar tree change
│   └── bootstrap.test.ts            # + test
└── features/runner/
    ├── store.ts                     # + refreshAll, refresh, flushSave, read/edit sequences
    └── store.test.ts                # + describe('refreshAll')

test/integration/features/
├── runner/runner_plan.feature       # + @xen_020 guard scenario
└── mcp/runner.feature               # + @xen_020 guard scenario
```

**Structure Decision**: Existing single repo; frontend feature `runner` plus the `app/`
composition layer. Backend layers are untouched because the backend already satisfies FR-002
and FR-007 (research R1).

## Complexity Tracking

| Deviation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| The two `@xen_020` Godog scenarios are green before the implementation (Principle V asks for red first) | The defect is in the frontend store; the backend behaviour they lock (FR-002, FR-007) already holds (research R1). The red-first proof for the new behaviour is the Vitest suite (T004, T005, T010), which Principle V counts in the full suite | Dropping the Godog guards would leave FR-002 (new subfolders) and FR-007 (headless start) without an executable spec; forcing them red would mean breaking correct backend code |
