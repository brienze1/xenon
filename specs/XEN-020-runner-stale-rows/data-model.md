# Data Model: Runner Follows the Current Tree (XEN-020)

No persisted data changes. The remembered rows (`RunnerPreferences` per target in the
workspace's local UI state) and the `RunPlan` returned by `RunnerService.Plan` keep their XEN-017
shape. What changes is client-side state in `frontend/src/features/runner/store.ts`.

## RunnerTarget (existing, per collection/folder id)

Fields a refresh **replaces** from the new plan:

| Field | Source |
| --- | --- |
| `target` | `plan.target` |
| `rows` | `plan.rows` (reconciled by the backend) |
| `config` | `plan.config` (remembered settings with the reconciled rows) |
| `environments` | `plan.environments` |
| `activeEnvironmentId` | `plan.activeEnvironmentId` |
| `dataFileMissing` | `plan.dataFileMissing` |
| `loadError` | `null` on success |

Fields a refresh **never touches** (FR-005): `status`, `view`, `runId`, `run`, `runs`, `failure`,
`fieldErrors`, `confirmation`, `missingSecrets`, `acknowledgement`, `pendingConfig`, `starting`.
The store's `active[itemId]` entry is set from `plan.activeRun` (same as `load`).

On a failed refresh (FR-010): `config = null`, `rows = []`, `target = null`, `loadError = error`;
the other fields stay.

## Module-level bookkeeping (new, beside `saveTimers` / `refreshing` / `finishedRuns`)

| Name | Type | Meaning | Cleared by |
| --- | --- | --- | --- |
| `reads` | `Map<string, number>` | Sequence of the latest plan read started for a target (`load` or refresh). A response applies only when its sequence is still the latest. | `clearTimers()` (store `reset`) |
| `edits` | `Map<string, number>` | Sequence bumped by every `scheduleSave` (rows, settings, data file). A refresh response is dropped when it moved during the read. | `clearTimers()` |
| `rereading` | `Map<string, boolean>` | Target has a refresh in flight; `true` = one trailing refresh requested. | `clearTimers()`, and per target when its refresh chain ends |

## State transitions of one target during a refresh

```text
ready ──tree change──▶ (pending save? flush it) ──▶ read in flight
read in flight ──another tree change or an edit──▶ trailing requested
read in flight ──response, still latest, no edit──▶ plan fields replaced ──▶ trailing? ──▶ read again
read in flight ──response, edit happened──▶ dropped ──▶ trailing read (flushes the edit first)
read in flight ──failure──▶ config=null + loadError (banner, Try Again)
read in flight ──workspace switched / store reset──▶ dropped, nothing patched
```
