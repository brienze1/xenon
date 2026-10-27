# Research: Runs Are Visible While They Run

All unknowns were settled from the code on `origin/main` @ `4e25bc5` (2026-10-27).

## R1 — Where the list gap comes from

- **Decision**: merge in-progress runs into `FindRunsService.List` / `ListForAI`.
- **Rationale**: `internal/application/service/find_runs.go` `List`/`ListForAI` call only
  `findRuns.FindRuns` (stored runs, `internal/integration/persistence/run.go`, written by
  `runExecutor.store` when a run ends). `Get`/`Response` go through `runLookup.run`, which checks
  `findActiveRuns.FindById` first. Both Wails `RunnerService.List` and MCP `runs list`
  (`controller/mcp_runs.go` → `ListForAI`) use these service methods, so one fix covers both.
- **Alternatives**: persist runs at start (rejected: rewrites the storage lifecycle, the 25-run
  pruning, and adds disk writes per row); merge in the controllers (rejected: business rule in the
  integration layer, two copies).

## R2 — Duplicate during finish

- **Decision**: drop an active run whose id is already stored; the stored copy wins.
- **Rationale**: `run_functional.go` finishes in this order: `change(... Finish ...)` →
  `store` (SaveRun + prune) → `releaseActiveRun.Release` → `publishRunEvent.Finished`. Between
  `store` and `Release` the run is in both sources. The stored copy is final and masked.

## R3 — Summary of an in-progress run

- **Decision**: compute the so-far summary at listing time (`Run.Listed()`), sharing the
  computation with `Finish` through a private `summarize()`.
- **Rationale**: `Run.Summary` is only set in `Finish` (`entity/run.go`), so a listed in-progress
  run would show zeros. Rows, `Progress.ElapsedMs` and `Performance` are kept up to date during the
  run (`Record`, `Starting`, `Measure`), so the same formula gives the so-far values.
- **Alternatives**: update `Summary` on every `Record` (rejected: changes what an in-progress
  `get` returns and touches the hot path for no listing benefit).

## R4 — Ordering

- **Decision**: stable sort by `StartedAt` descending after merging.
- **Rationale**: stored runs come newest first (file names are time-ordered ids, read
  backwards). Only one run per target can be in progress (`activeRunsPersistence.Claim` →
  `RUN-02409`), and it started after every stored run of that target finished, so it lands first;
  the stable sort keeps the stored order for equal times.

## R5 — Why the Runner tab keeps the old run

- **Decision**: in `applyProgress`, follow a run once when it starts.
- **Rationale**: `frontend/src/features/runner/store.ts` `applyProgress` returns early on
  `(target.runId ?? event.runId) !== event.runId`; only `start()` (the tab's own Start) sets
  `runId` to the new run. "Starts" is detectable without new backend data: before `setActive`,
  `active[itemId]?.id !== event.runId` means the app did not yet hold this run as the target's
  active run (`restoreActive` fills `active` at startup/workspace switch, so restored runs are not
  "starting"). Following only then lets the user open an older run during the run without being
  pulled back on the next progress report.
- **Alternatives**: a "new run started — view" notice (rejected per issue's recommendation;
  leaves the run unwatchable until clicked); following on every report (rejected: Runs unusable
  during a run); a new `runner:started` event (rejected: the first progress event already carries
  everything, and no backend change is needed).

## R6 — Delete / Clear with an in-progress run

- **Decision**: no Delete action on in-progress rows; Clear keeps the in-progress row and is
  disabled when nothing is stored.
- **Rationale**: `delete_runs.go` `Delete` refuses an active run (`ActiveConflict`); `Clear`
  deletes the stored folder only (the in-progress run is not in it). The UI must not offer an
  action that is always refused, and must not hide a run that still exists.
