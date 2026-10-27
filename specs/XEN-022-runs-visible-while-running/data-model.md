# Data Model: Runs Are Visible While They Run

No stored format changes. Entities touched (`internal/domain/entity/run.go`):

## Run (existing)

| Field | Use here |
| --- | --- |
| `ID` | identity for de-duplication (stored copy wins) |
| `Target.ItemID`, `Target.LocalID` | filter in-progress runs to the listed target / workspace |
| `Status` | `running`/`paused` = in progress; `completed`/`stopped`/`stopped_on_error` = finished (`IsFinished`) |
| `StartedAt` | sort key, newest first |
| `Rows`, `Progress.ElapsedMs`, `Performance` | inputs to the so-far summary |
| `Summary` | final once finished; filled by `Listed()` for an in-progress run in a list |

### New behaviour

- `(r *Run) summarize()` — private; sets `Summary` from `Iterations()`, `Rows`,
  `Progress.ElapsedMs`, plus `WithPerformance(Performance.Overall)` when `Performance != nil`.
  `Finish` uses it (unchanged result).
- `(r Run) Listed() Run` — copy; finished → as is; in progress → `summarize()` applied.
- `ListRuns(targetId string, active, stored []Run) []Run` — stored ∪ (active with
  `Target.ItemID == targetId` and `ID ∉ stored`, each `Listed()`), stable-sorted by `StartedAt`
  desc; never nil.

### State transitions (unchanged)

`running ⇄ paused → completed | stopped | stopped_on_error`. Listed as in progress while
`running`/`paused` (from memory); listed as stored once saved; never both.

## Runner tab state (frontend, `features/runner/model.ts` `RunnerTarget`, unchanged shape)

| Field | Change |
| --- | --- |
| `runId` | set to a newly started run of the target on its first progress report |
| `view` | `'results'` when following a newly started run |
| `run` | `null` until the followed run is read |
| `runs` | read again when a progress report names a run it lacks; Clear keeps in-progress entries |
