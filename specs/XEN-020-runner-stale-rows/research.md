# Research: Runner Follows the Current Tree (XEN-020)

All findings are from `origin/main` at `2a46f14` (2026-10-26).

## R1 — Where the stale rows come from

- **Decision**: The defect is in the frontend Runner store only.
- **Evidence**:
  - `frontend/src/features/runner/store.ts` `ensure(itemId)` returns early once a target's
    `status !== 'idle'`; `RunnerSection.tsx` calls `ensure` on mount only. Nothing else ever
    calls `load` again for a target, so the rows from the first `planRun` stay for the session.
  - `start()` builds the run config from those rows (`configWithRows(target.config, target.rows)`)
    and the backend runs exactly them: `internal/application/service/run_starter.go`
    `target()` → `start.Config.PlannedRows(...)` keeps only the config's selected rows. New
    requests are therefore skipped; a deleted one is skipped by `preflight` and becomes a failed
    "missing request" row in the run (`run_row.go` `runRow` → `row.MissingRequest()`).
  - The debounced `save()` writes the frozen rows back as the remembered rows.
- **Backend is correct**: `run_planner.go` `plan()` loads the tree with `loadTree.Load`, which
  is `treePersistence.Load` → `workspaceIndexCache.Refresh` (a full disk rescan on every call),
  then `entity.ReconcileRunRows(remembered, sidebar)` (new requests join selected after their
  nearest preceding request in sidebar order, removed ones drop, order/selection/alias kept, unknown alias →
  Default via `withKnownAlias`). `RunnerService.Plan` and the MCP `StartForAI` both go through
  `planner.plan`, so a headless start without rows always sees the current tree. Locked today by
  `runner_plan.feature` `@reconcile` scenarios (requests in existing folders only).
- **Alternatives considered**: reconciling inside `runStarter.target()` (adding unknown sidebar
  requests to explicit rows) — rejected: explicit rows (UI or AI) are an explicit selection, and
  the UI shows rows that must match what runs; fixing the UI's freshness fixes both.

## R2 — Which signal means "the tree changed"

- **Decision**: React to the sidebar tree being replaced (`useSidebarStore` `tree` reference
  changes while `status === 'ready'`), in the existing subscription in `frontend/src/app/bootstrap.ts`.
- **Rationale**: The backend `workspace:tree-changed` event is **not** emitted for Xenon's own
  writes (`handle_file_change.go`: `change.IsOwnWrite()` returns early), so sidebar edits made in
  the UI never produce it; the sidebar refreshes itself through `useSidebarStore` (`refresh()`
  after create/rename/move/delete). Every tree change — sidebar edits, AI tree edits
  (`notify_ai_change.go`), git sync (`sync_worktree_writer.go`), imports (`import_writer.go`),
  external edits (`handle_file_change.go`) — ends in a new `useSidebarStore.tree`. The same
  subscription already refreshes item settings (`useItemSettingsStore.refreshAll`) and
  documentation (`useDocumentationStore.refreshAll`) on exactly this condition; the Runner was
  simply missing from it.
- **Alternatives considered**: subscribing to `onWorkspaceTreeChanged` — misses own writes;
  re-reading on every Runner mount (`ensure` → `load` always) — misses an already-open Runner and
  still lets an open Runner start a stale run.

## R3 — Workspace switch safety

- **Decision**: No extra guard needed.
- **Evidence**: `useWorkspacesStore.activate` sets `activeLocalId` (the bootstrap subscription
  then calls `useRunnerStore.reset()` synchronously) **before** `useSidebarStore.setTree(snapshot.tree)`,
  so the refresh triggered by the new tree sees no targets. `useSidebarStore.reset()` leaves
  `status` not `'ready'`, so it triggers nothing. Responses arriving after a switch are already
  dropped by the store's `inWorkspace` helper (it returns `null` when `activeLocalId` changed).

## R4 — Not losing the user's Runner edits

- **Decision**: Before re-reading a target, flush its pending debounced save (cancel the timer
  and `await save(itemId)`). While a re-read is in flight, any new edit (`setRows`,
  `updateConfig`, data-file changes — all go through `scheduleSave`) bumps a per-target edit
  sequence; a re-read whose start sequence no longer matches is discarded and one trailing
  re-read is queued (it flushes the newer edit first).
- **Rationale**: the backend reconciles the remembered rows, so remembering first and then
  reading gives "edits kept + tree applied" without merging rows on the client (Constitution III:
  Go owns the reconciliation logic).

## R5 — Bursts of tree changes

- **Decision**: One re-read in flight per target plus at most one trailing re-read (same shape as
  the store's existing `refreshRun` coalescing). A per-target read sequence also makes the
  first `load` and a refresh safe against out-of-order responses: only the latest read applies.

## R6 — Refresh failure

- **Decision**: A failed re-read clears the target's `config`/`rows`/`target` and sets `loadError`,
  so `RunnerSection` shows its existing error banner with **Try Again** (it only shows it when
  `config == null`). Today a failed `load` keeps the old config, which would hide the error and
  keep showing rows that no longer exist (deleted target).

## R7 — BDD coverage of an already-correct backend

- **Decision**: Add two Godog scenarios that lock FR-002 (a whole new subfolder joins, in sidebar
  order) and FR-007 (MCP `run start` without rows after requests were added). They are expected
  to be **green before implementation** — they guard backend behaviour that already holds (R1).
  The red-first proof for this fix is the Vitest suite (`store.test.ts`, `bootstrap.test.ts`),
  because the defect lives in the frontend store (Constitution V: the full suite includes
  frontend Vitest).
