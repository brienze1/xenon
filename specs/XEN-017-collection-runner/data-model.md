# Data Model: Collection Runner (XEN-017)

Domain entities live in `internal/domain/entity/` and enums in `internal/domain/enums/`. Every type
has a single file named after it (`naming-conventions.md`).

## Enums

| Enum (file) | Values |
| --- | --- |
| `RunType` (`run_type.go`, `runtype/`) | `functional`, `performance` |
| `RunStatus` (`run_status.go`, `runstatus/`) | `running`, `paused`, `completed`, `stopped`, `stopped_on_error` |
| `RunRowOutcome` (`run_row_outcome.go`, `runrowoutcome/`) | `passed`, `failed`, `skipped`, `error`, `cancelled` |
| `LoadProfileKind` (`load_profile_kind.go`, `loadprofilekind/`) | `fixed`, `ramp_up`, `spike`, `peak` |
| `RunResultFilter` (`run_result_filter.go`, `runresultfilter/`) | `all`, `passed`, `failed`, `skipped` |
| `RunExportFormat` (`run_export_format.go`, `runexportformat/`) | `json`, `junit` |

## Entities

### RunTarget (`run_target.go`)

| Field | Type | Rule |
| --- | --- | --- |
| `LocalID` | string | workspace |
| `ItemID` | string | collection or folder id (ULID) |
| `Kind` | `ContainerKind` | collection or folder |
| `Name` | string | the display name when the run started |

### RunRow (`run_row.go`)

| Field | Type | Rule |
| --- | --- | --- |
| `RequestID` | string | saved request id |
| `AliasID` | string | `""` = Default (`IsDefaultAliasID`) |
| `Selected` | bool | |

### RunPlanRow (`run_plan_row.go`) — what the Runner lists

The fields are those of `RunRow`, plus `Number` (1-based position), `Name`, `Method`, `FolderPath
[]string` (relative to the target) and `Aliases []RequestAlias` (id and name only).

Functions:

- `ReconcileRunRows(remembered []RunRow, sidebar []RunPlanRow) []RunPlanRow` (FR-007, FR-008):
  - drops rows for missing requests;
  - inserts new requests checked, after the nearest preceding sidebar sibling that is present, or
    first;
  - resets aliases that no longer exist to the Default;
  - numbers the result.
- `FlattenRunRows(target Collection, tree)` walks the sidebar order depth-first and skips examples
  (FR-003).

### FunctionalSettings (`functional_settings.go`)

| Field | Type | Default | Rule |
| --- | --- | --- | --- |
| `Iterations` | int | 1 | ≥ 1 (`RUN-04400`) |
| `DataFile` | string | `""` | absolute path, or empty |
| `DelayMs` | int | 0 | 0–3,600,000 (`RUN-04400`) |
| `PersistResponses` | bool | false | |
| `DisableLogs` | bool | false | |
| `StopOnError` | bool | false | |
| `KeepVariableValues` | bool | true | |
| `SkipStoredCookies` | bool | false | |
| `SaveCookies` | bool | false | |

### LoadProfile (`load_profile.go`)

| Field | Type | Rule |
| --- | --- | --- |
| `Kind` | `LoadProfileKind` | |
| `InitialUsers` | int | Ramp up only; 1 ≤ x ≤ VirtualUsers |
| `RampSeconds` | int | Ramp up only; 1 ≤ x ≤ duration |
| `BaseUsers` | int | Spike and Peak only; 1 ≤ x ≤ VirtualUsers |
| `SpikeSeconds` | int | Spike only; 1 ≤ x ≤ duration |
| `PeakSeconds` | int | Peak only; 1 ≤ x < duration |

`ActiveUsersAt(elapsed, duration, max) int` (FR-043; research R13).

### PerformanceSettings (`performance_settings.go`)

| Field | Type | Default | Rule |
| --- | --- | --- | --- |
| `VirtualUsers` | int | 10 | 1–100 (`RUN-04400`) |
| `DurationSeconds` | int | 600 | 1–3600 (`RUN-04400`) |
| `Profile` | `LoadProfile` | Fixed | parameters within the duration and the user count |
| `DataFile` | string | `""` | |

### RunConfig (`run_config.go`)

| Field | Type |
| --- | --- |
| `Type` | `RunType` |
| `Rows` | `[]RunRow` (the order is the run order; only `Selected` rows execute) |
| `Functional` | `FunctionalSettings` |
| `Performance` | `PerformanceSettings` |
| `EnvironmentID` | string (`""` = active at plan time, resolved at start; `NoEnvironmentID` = none) |

`Validate() error` checks every rule above, plus a non-empty selection (`RUN-03400`).
`SelectedRows()`.

### RunnerPreferences (`runner_preferences.go`)

This is `RunConfig` minus nothing. It is stored per target in `UIState.Runners map[string]
RunnerPreferences` (FR-006, FR-013).

### IterationData (`iteration_data.go`)

| Field | Type |
| --- | --- |
| `FileName` | string (base name only, stored on the run) |
| `Rows` | `[]map[string]any` |

`Row(iteration) map[string]any` returns the last row past the end (FR-019).
`WrappedRow(n)` wraps around (FR-044). `Count()`.

### RunContext (`run_context.go`) — passed into the shared send core (research R1)

| Field | Type | Meaning |
| --- | --- | --- |
| `RunID` | string | `""` for a single send (`NoRunContext()`) |
| `EnvironmentID` | string | the override (R3) |
| `Iteration`, `IterationCount` | int | `pm.info` (R6) |
| `IterationData` | `map[string]any` | the current row |
| `Variables` | `*RunVariables` | nil means persist as today |
| `CookieJarID` | string | `localId` or a run jar id (R5) |
| `MuteConsole` | bool | FR-027 |
| `Collect` | `func(ConsoleEntry)` | the row collector |
| `ResponseKey` | string | `""` = keep no response (R11) |
| `Acknowledgement` | `SendAcknowledgement` | confirmed environment, given once per run |
| `Performance` | bool | ignores `setNextRequest` with a warning |

### RunVariables (`run_variables.go`)

This holds the overlay `ScriptScopes` (environment, collection, globals) and the vault changes.
`Apply(before, after ScriptScopes, vaultBefore, vaultAfter)` and `Over(VariableContext,
ScriptScopes)` (R4).

### NextRequest (`next_request.go`)

`{ Target string; Stop bool }` comes from `ScriptRun.NextRequest`. `SendExecution.NextRequest()`
returns the last value set.

### RunRowResult (`run_row_result.go`)

| Field | Type |
| --- | --- |
| `Index` | int (execution order, 0-based, unique within the run) |
| `Iteration` | int |
| `RequestID`, `AliasID`, `AliasName`, `Name`, `Method`, `FolderPath` | identity |
| `URL` | resolved URL, masked |
| `Status` | int (0 = not sent) |
| `DurationMs` | int64 |
| `SizeBytes` | int64 |
| `Outcome` | `RunRowOutcome` |
| `Tests` | `[]TestResult` |
| `ScriptErrors` | `[]ScriptError` |
| `Console` | `[]ConsoleEntry` (masked; empty when logs are off) |
| `SentRequest` | `SentRequest` (method, URL, headers, body text; masked) |
| `Error` | string (send or lookup error, masked) |
| `ResponseKey` | string (`""` unless responses were persisted) |

The outcome is derived (FR-034) by `NewRunRowOutcome(execution)`:

| Situation | Outcome |
| --- | --- |
| Skipped by a script | skipped |
| Cancelled | cancelled |
| Not sent, a script error, or the request not found | error |
| Any failed test | failed |
| Otherwise | passed |

### RunSummary (`run_summary.go`)

`Iterations`, `TotalRequests`, `TestsPassed`, `TestsFailed`, `TestsSkipped`, `AvgResponseMs`,
`DurationMs`, plus counts per outcome. It is computed from the rows.

### RunProgress (`run_progress.go`)

`CurrentRow` (number), `CurrentName`, `Iteration`, `Completed`, `Planned` (rows × iterations; for
performance runs, the requests so far), `TestsPassed`, `TestsFailed` and `ElapsedMs`. For
performance runs it also holds `Metrics PerformanceSecond`.

### Performance statistics

- **`LatencyHistogram`** (`latency_histogram.go`): `Record(d)`, `Percentile(p)`, `Min`, `Max`,
  `Mean`, `Count`. Log-linear buckets with ≤1% relative error (R13).
- **`RequestStats`** (`request_stats.go`): `RequestID`, `Name`, `Histogram`, `Errors`,
  `StatusCounts map[int]int` and `Throughput(duration)`.
- **`PerformanceSecond`** (`performance_second.go`): `Second`, `ActiveUsers`, `Requests`,
  `AvgResponseMs`, `Errors`.
- **`PerformanceStats`** (`performance_stats.go`): `Overall RequestStats`, `PerRequest
  []RequestStats` and `Series []PerformanceSecond`. It is safe for concurrent `Record` calls
  (it holds its own mutex, a deliberate exception kept inside the aggregate).

### Run (`run.go`)

| Field | Type |
| --- | --- |
| `ID` | ULID |
| `Target` | `RunTarget` |
| `Config` | `RunConfig` |
| `EnvironmentName` | string |
| `DataFileName` | string |
| `Status` | `RunStatus` |
| `StartedAt`, `FinishedAt` | time |
| `Progress` | `RunProgress` |
| `Rows` | `[]RunRowResult` (functional) |
| `Summary` | `RunSummary` |
| `Performance` | `*PerformanceStats` (performance runs only) |
| `Notices` | `[]string` (run-level notices, e.g. the performance `setNextRequest` warning) |

State transitions:

```
running ──pause──► paused ──resume──► running
running|paused ──last row / duration ends──► completed
(performance runs never enter paused)
running|paused ──stop / workspace closed──► stopped
running ──error with StopOnError──► stopped_on_error
```

These end states are terminal, and a run is stored only in a terminal state.

## Persistence models (`internal/integration/persistence/model/`)

- **`run.go`**: the JSON model of `Run` (snake_case), with `ToEntity` and `NewRun`. Stored at
  `<DataDir>/workspaces/<localId>/runs/<targetId>/<runId>.json`. The cap of 25 per target is
  applied by the persistence after each save (R12).
- **`ui_state.go`**: gains `runners map[string]runnerPreferences` (`omitempty`). Older
  `state.json` files without it load as empty.

## Validation error codes (new; `RUN-` family)

| Code | When |
| --- | --- |
| `RUN-01404` | run id not found (neither active nor stored) |
| `RUN-02409` | a run is already active on this target (the details hold the active run id) |
| `RUN-03400` | empty selection, or a target with no requests |
| `RUN-04400` | a setting out of range (the details name the field) |
| `RUN-05400` | an iteration data file that is unreadable or malformed (the details give the reason) |
| `RUN-06404` | a row's request no longer exists (a row error, not a start error) |
| `RUN-07404` | target collection or folder not found |

Existing errors are reused unchanged: `SND-05409` (confirm before send) and the missing-secrets
rejection.
