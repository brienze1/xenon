# Contract: Go ↔ React bindings for the runner

All calls go through `frontend/src/platform/services/runner.ts`, never through the generated
bindings directly. Errors use the existing `errs` shape (`code`, `description`, `details`).

## New binding `RunnerService` — `internal/infra/server/binding/runner.go`

`RunnerService` is a thin adapter over `adapter.RunnerController`
(`internal/integration/entrypoint/controller/runner.go`). That controller follows the full
controller pattern:

1. validate the DTO;
2. `ToEntity`;
3. call `appAdapter.ManageRunnerService`;
4. handle errors;
5. recover from panics.

`RunnerService` is registered next to `RequestService` in the Wails service list.

| Method | Input DTO | Output | Behaviour |
| --- | --- | --- | --- |
| `Plan(ctx, dto.RunTargetReference)` | `{local_id*, item_id*}` | `dto.RunPlan` | Returns the target, its reconciled rows, the remembered configuration (or the defaults), the environments, an active-run summary when one exists, and `data_file_missing` when a remembered file vanished (that file is cleared). Unknown target → `RUN-07404`. |
| `SavePreferences(ctx, dto.SaveRunnerPreferences)` | `{local_id*, item_id*, config: dto.RunConfig*}` | — | Remembers the rows, type, settings, data file and environment in UI state (FR-006, FR-013). It writes no workspace file. |
| `Start(ctx, dto.StartRun)` | `{local_id*, item_id*, config: dto.RunConfig*, confirmed_environment_id, acknowledge_missing_secrets}` | `dto.RunSummaryView` (id, status `running`) | Validates and returns at once. The run continues in the background, and events report its progress. Rejections: `RUN-03400`, `RUN-04400`, `RUN-05400`, `RUN-02409`, `RUN-07404`, `SND-05409` (confirm), and `SND-06409` missing secrets (unless `acknowledge_missing_secrets`). Every row runs with `ExpiredToken` acknowledged. |
| `Pause(ctx, dto.RunReference)` | `{local_id*, run_id*}` | `dto.RunSummaryView` | Only from `running`, and only for functional runs (performance → `RUN-04400`). Unknown run → `RUN-01404`. |
| `Resume(ctx, dto.RunReference)` | same | `dto.RunSummaryView` | Only from `paused`. |
| `Stop(ctx, dto.RunReference)` | same | `dto.RunSummaryView` | It cancels the request in flight. The run ends as `stopped` and is stored. |
| `Get(ctx, dto.RunReference)` | same | `dto.Run` | Returns the active run, or else the stored run. |
| `List(ctx, dto.RunTargetReference)` | `{local_id*, item_id*}` | `dto.RunSummaryView[]` | The stored runs of the target, newest first, at most 25. |
| `Active(ctx, dto.WorkspaceReference)` | `{local_id*}` | `dto.RunSummaryView[]` | The active runs of the workspace. The UI uses it to restore the indicators after a reload. |
| `Response(ctx, dto.RunRowReference)` | `{local_id*, run_id*, index*}` | `dto.SendResult` | Returns the persisted response of that row (status, headers, timings, `body_url`/`inline_body`) when this session holds it. Otherwise → `RUN-01404`. |
| `Delete(ctx, dto.RunReference)` | same | — | Deletes the stored run and releases its persisted responses. An active run → `RUN-02409`. |
| `Clear(ctx, dto.RunTargetReference)` | same | — | Deletes every stored run of the target (FR-041a). |
| `Export(ctx, dto.ExportRun)` | `{local_id*, run_id*, format* ("json"\|"junit"), path*}` | — | Writes the file. `junit` for a performance run → `RUN-04400`. |

`*` means `validate:"required,not_blank"`. DTO validation failures use these codes:

| Code | Calls |
| --- | --- |
| `RUN-08400` | Plan, List, Clear |
| `RUN-09400` | SavePreferences, Start |
| `RUN-10400` | Pause, Resume, Stop, Get, Delete, Response |
| `RUN-11400` | Export |
| `RUN-12400` | Active |

### `dto.RunConfig`

```json
{
  "type": "functional",
  "rows": [{ "request_id": "01J…", "alias_id": "", "selected": true }],
  "functional": {
    "iterations": 1, "data_file": "", "delay_ms": 0,
    "persist_responses": false, "disable_logs": false, "stop_on_error": false,
    "keep_variable_values": true, "skip_stored_cookies": false, "save_cookies": false
  },
  "performance": {
    "virtual_users": 10, "duration_seconds": 600, "data_file": "",
    "profile": { "kind": "fixed", "initial_users": 0, "ramp_seconds": 0,
                 "base_users": 0, "spike_seconds": 0, "peak_seconds": 0 }
  },
  "environment_id": ""
}
```

In `environment_id`, `""` means the active environment and `"none"` means no environment.

### `dto.RunPlan`

```json
{
  "target": { "item_id": "…", "kind": "collection", "name": "Orders API" },
  "rows": [{ "number": 1, "request_id": "…", "alias_id": "", "selected": true,
             "name": "Login", "method": "POST", "folder_path": ["Auth"],
             "aliases": [{ "id": "…", "name": "mx-test-user" }] }],
  "config": { /* dto.RunConfig */ },
  "active_environment_id": "…",
  "environments": [{ "id": "…", "name": "dev", "color": "green", "confirm_before_send": false }],
  "active_run": null,
  "data_file_missing": false
}
```

### `dto.Run`

```json
{
  "id": "01J…", "target": {…}, "config": {…}, "environment_name": "dev", "data_file_name": "users.csv",
  "status": "completed", "started_at": "…", "finished_at": "…",
  "progress": { "current_row": 3, "current_name": "Get order", "iteration": 0, "completed": 6, "planned": 6,
                "tests_passed": 10, "tests_failed": 1, "elapsed_ms": 1840, "metrics": null },
  "summary": { "iterations": 1, "total_requests": 6, "tests_passed": 10, "tests_failed": 1, "tests_skipped": 0,
               "avg_response_ms": 120, "duration_ms": 1840,
               "passed": 5, "failed": 1, "skipped": 0, "errors": 0, "cancelled": 0 },
  "rows": [{
    "index": 0, "iteration": 0, "request_id": "…", "alias_id": "", "alias_name": "", "name": "Login",
    "method": "POST", "folder_path": ["Auth"], "url": "https://api…/login", "status": 200,
    "duration_ms": 98, "size_bytes": 512, "outcome": "passed",
    "tests": [{ "name": "status is 200", "status": "passed", "error": "" }],
    "script_errors": [], "console": [ /* dto.ConsoleEntry */ ],
    "sent_request": { "method": "POST", "url": "…", "headers": [{ "key": "Authorization", "value": "•••••" }], "body": "…" },
    "error": "", "response_body_url": ""
  }],
  "performance": null,
  "notices": []
}
```

- When the run persisted responses and this session still holds them, `response_body_url` is the
  in-app body route URL for `run:<runId>:<index>`, and the row detail loads the response with
  `RunnerService.Response`.
- `performance` holds:
  - `overall` and `per_request`: `{request_id, name, count, avg_ms, p90_ms, p95_ms, p99_ms,
    min_ms, max_ms, throughput, error_rate, status_counts}`;
  - `series`: `[{second, active_users, requests, avg_ms, errors}]`.

## Changed: `DialogService.PickFile`

`dto.PickFile` gains an optional `filters: [{ display_name, pattern }]`, for example `{
"display_name": "Data files", "pattern": "*.csv;*.json" }`, which is passed to the native dialog.
Calls without `filters` behave as today.

## Changed: `RequestService` / send pipeline

The public contract is unchanged. Internally, `Send` runs through the shared core with
`NoRunContext()` (research R1). `pm.info.iteration` becomes `0` and `pm.info.iterationCount`
becomes `1` for a single send, as today.

## Changed: UI state

`dto.UIState` gains `runners: { [itemId]: dto.RunConfig }` (omitted when empty). It is written
only through `RunnerService.SavePreferences`, and `UIStateService.Save` keeps it as is.
