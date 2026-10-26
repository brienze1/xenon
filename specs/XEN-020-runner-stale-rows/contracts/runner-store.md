# Contract: Runner store refresh (XEN-020)

The UI-facing contract of `useRunnerStore` (`frontend/src/features/runner/store.ts`) gains one
action. No Wails binding, Go service, MCP tool or event changes; `planRun` and
`saveRunnerPreferences` from `platform/services/runner` are reused as they are.

## `refreshAll(): Promise<void>`

Doc comment: `/** Reads every Runner target shown this session again (after the workspace tree changed). */`

- Targets: every `targets[itemId]` whose `status !== 'idle'`. Targets never read this session are
  not touched. Resolves when the reads it started (including their trailing reads) settled; a call
  that finds a read already in flight only requests that read's trailing read.
- Per target, in order:
  1. If a refresh for the target is already in flight, request one trailing refresh and return.
  2. If a remembered-rows save is pending (debounce timer), cancel the timer and await the save.
  3. Read `planRun(activeLocalId, itemId)`.
  4. Drop the response when: the active workspace changed, the store was reset, a newer read of
     the target started, or the target was edited since step 2 (then request a trailing refresh).
  5. Success: set `status: 'ready'` and replace `target`, `rows`, `config`, `environments`, `activeEnvironmentId`,
     `dataFileMissing`; set `loadError: null`; set `active[itemId]` from `plan.activeRun`.
     Nothing else changes (view, shown run, failure, confirmation, starting…).
  6. Failure: set `status: 'ready'`, `config: null`, `rows: []`, `target: null`, `loadError: <AppError>`.
  7. If a trailing refresh was requested, run it (from step 2).
- Never throws; errors land in `loadError` (6) or, for the flushed save, in `failure` (existing
  `save()` behaviour).

## Caller

`frontend/src/app/bootstrap.ts`, existing subscription:

```ts
useSidebarStore.subscribe((state, previous) => {
  if (state.tree !== previous.tree && state.status === 'ready') {
    useTabsStore.getState().handleTreeChanged(state.tree)
    void useItemSettingsStore.getState().refreshAll()
    void useDocumentationStore.getState().refreshAll()
    void useRunnerStore.getState().refreshAll()   // XEN-020
  }
  …
})
```

## Unchanged, relied upon

- `RunnerService.Plan` (`planRun`) returns rows reconciled against a fresh disk scan
  (`run_planner.go` `plan()` → `entity.ReconcileRunRows`).
- MCP `run` `start` without `rows` uses the same `plan()` (`start_run.go` `StartForAI`).
- `start()` sends `configWithRows(target.config, target.rows)` — current after a refresh.
