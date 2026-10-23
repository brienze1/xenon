# Research: Collection Runner (XEN-017)

All findings were verified on `origin/main` @ `db23057` (2026-10-23). Every decision below is
final. None is left open.

## R1 — Reuse the single-send pipeline through a run seam, not a second engine

- **Decision**: A run row sends through the existing `sendRequest` core
  (`internal/application/service/send_request.go`, `send_run.go`). That core is refactored so that
  `Send` (the single-send path, unchanged in behaviour) and a new runner path share one prepare →
  execute → settle sequence. The behaviours that differ are carried by a new domain value,
  `entity.RunContext`, which is passed into the shared core. A single send passes
  `entity.NoRunContext()`. Through it the runner selects:
  - the environment override;
  - the iteration info and data row;
  - the variable overlay (or none, which persists as today);
  - the cookie jar id;
  - console muting and the row collector;
  - no History and no tab response;
  - the response-cache key when responses are persisted.
- **Rationale**: FR-016 needs the exact single-send semantics: inherited auth and headers
  (`ResolveEffectiveRequest`), the script chain (`CollectScriptChain`), OAuth, digest, AWS, the
  cookie rules and secret masking. Duplicating them would drift.
- **Alternatives rejected**:
  - Calling `adapter.SendRequestService.Send` per row: it always writes History
    (`send_request.go:179`), always persists script variables, and always uses the workspace jar
    and the active environment.
  - Adding flags to `Send`'s public signature: that leaks runner concerns into every caller
    (UI, MCP).

## R2 — Saved requests, read at their turn

- **Decision**: Each row loads its request with `usecase.FindRequest.Find(root, requestId)` and
  takes `Version(aliasId)` just before the row runs (FR-016, FR-017). A missing file becomes a
  row **error** with code `RUN-06404` ("request not found"). The UI draft is never involved.
- **Rationale**: The MCP `send` tool already sends saved requests the same way
  (`controller/mcp_send_target.go`). Reading at the row's turn makes the "renamed or deleted during
  a run" edge case well-defined.

## R3 — Environment override

- **Decision**: `requestVariables.loadChain` (`service/request_variables.go:40`) currently takes
  the active environment from `local.ActiveEnvironment`. It gains an environment-id parameter:
  - `""` keeps today's behaviour (the active environment);
  - `entity.NoEnvironmentID` means no environment;
  - any other id loads that environment.

  The run resolves the id once at start and stores it on `RunContext`. The workspace's active
  environment is never written (FR-012).
- **Confirm before send** (FR-015): at start, `usecase.CheckSendPreflight` runs against the run's
  environment once. An unconfirmed start returns the existing `SND-05409` error
  (`usecase/reject_unconfirmed_send.go`). The UI then restarts with `confirmed_environment_id`,
  the same contract as Send. Rows then run with an acknowledgement that carries that id, so no row
  asks again.
- **Missing secrets** (FR-014): at start the runner evaluates `AwaitsSecretsAcknowledgement` for
  every selected row's effective request. Any miss rejects the start with the existing
  missing-secrets error (`SND-06409`), listing the union of the missing variables. A restart with
  `acknowledge_missing_secrets` carries `MissingSecrets: true` in every row's acknowledgement,
  which is the same contract as "Send anyway".
- **Expired tokens**: every row's acknowledgement sets `ExpiredToken: true`
  (`SendAcknowledgement.RejectsExpiredToken`), so a run never stops to ask, matching Postman's
  runner. Refreshable tokens are still refreshed by `oauthTokens.current`.

## R4 — Variable overlay when values are not kept

- **Decision**: Add a new domain value, `entity.RunVariables`, which holds the environment,
  collection and globals `ScriptScopes`, the vault changes and the per-collection identity.
  - **Keep values on (functional)**: rows persist exactly as today through
    `scriptVariables.persist` (FR-024).
  - **Keep values off, and every performance virtual user**: `sendRun.runScripts` records each
    script's changes into the overlay instead of calling `persist`.
  - Each row's scopes are loaded from disk and then overlaid. The tools for this already exist:
    `VariableContext.WithScriptScopes` and `ScriptScopes` merging
    (`entity/variable_context.go:37`).
- **Rationale**: Later rows must see the changes, and the end of the run must leave
  `variables.json` and the workspace files untouched (FR-024, SC-008). An overlay discarded at the
  end is exactly that.
- **Concurrency note**: `persistence/local_variables.go` has no mutex. With keep values on, two
  overlapping runs (on different collections), or a run plus a tab send, could race on the
  read-modify-write of `variables.json`. Fix: add a `sync.Mutex` to `localVariablesPersistence`,
  plus a per-workspace lock in `scriptVariables.persist` around its find → apply → save sequence.
  Performance runs never persist, so they add no contention.

## R5 — A run-scoped cookie jar

- **Decision**: The jar manager is keyed by an id (`usecase.ManageCookieJar`). Today that id is
  always `localId`, and requests carry it via `EffectiveRequest.WithCookieJar(id)`. A run uses its
  own jar id, `entity.RunCookieJarID(localId, runId)` (`<localId>#run:<runId>`):
  - Without "Run collection without using stored cookies", it is seeded with
    `JarState(localId).Cookies` (loaded first). With it, the jar starts empty.
  - Cookies set during the run land in the run jar.
  - When the run ends with "Save cookies after collection run" on, the run jar's cookies are merged
    into the workspace jar with `entity.MergeCookies(workspace, run)`: a run cookie replaces the
    workspace cookie with the same name, domain and path, and adds otherwise. The result is loaded
    with `LoadCookies(localId, merged)`, then `cookies.settle` saves it and publishes
    `CookiesChanged`.
  - Either way the run jar is dropped. This needs a new method on the jar manager,
    `ManageCookieJar.DropJar(ctx, jarId)`, implemented in `webservice/cookie_jar.go`.
- **Performance**: each virtual user gets `<localId>#run:<runId>:<vu>`, seeded the same way. These
  jars are never merged back (FR-042).

## R6 — Iteration info and iteration data

- **Domain**:
  - `entity.ScriptInfo` gains `Iteration int`. `NewScriptInfo` keeps `IterationCount: 1`,
    `Iteration: 0` for single sends.
  - `entity.ScriptContext` gains `IterationData map[string]any`.
- **Sandbox**:
  - `dto.NewScriptSandboxInfo` (`webservice/dto/script_sandbox_info.go`) passes the real iteration
    instead of the hard-wired 0.
  - The sandbox context DTO gets an `iterationData` object.
  - `shims/pm.js:70` fills its `iterationData` `VariableScope` from `context.iterationData`. It is
    already layered between local and environment, which matches Postman precedence.
- **Go resolution of `{{field}}`**: add `VariableContext.WithIterationData(row)`, a layer inserted
  directly after local and before environment, so request URLs, headers and bodies resolve data
  columns (FR-018).
- **Parsing** (FR-019): pure domain functions in a new use case, `usecase.ParseIterationData(name,
  bytes)`:
  - CSV via `encoding/csv`, with `FieldsPerRecord = -1`. The header row gives the field names,
    extra cells are dropped and missing cells become `""`.
  - JSON must be an array of objects. Values keep their JSON types for `pm.iterationData`, and
    their string form for `{{}}` resolution.
  - Errors use `RUN-05400` with the reason (no header row, not an array of objects, empty, or
    unreadable).
- **Reading the file**: a new persistence adapter, `usecase.FindIterationData`
  (`persistence/iteration_data.go`). React never reads files (Constitution III).

## R7 — `pm.execution.setNextRequest`

- **Sandbox**:
  - `shims/pm.js:227` `setNextRequest(next)` calls `host.setNextRequest(next === null ? null :
    String(next))`. A no-argument call is treated as `null`, as in Postman.
  - `sandbox/pm_execution.go` records the last call on the execution.
  - `entity.ScriptRun` gains `NextRequest *entity.NextRequest`, which holds `{Target string; Stop
    bool}`.
  - `entity.SendExecution.NextRequest()` returns the last one set across the post-response chain
    (pre-request calls also count, as in Postman: the last call wins).
- **Runner** (FR-020):
  - It resolves the target against its rows: first by request id, then by the first row in run
    order whose request name equals the target.
  - With `Stop` set, or no match, the iteration ends. A missing match also adds a console warning
    entry, `Couldn't find request "<target>" in this run; the iteration ends.`
- **Single send**: the call is recorded and ignored, which keeps it a no-op.
- **Performance runs**: the runner ignores it and adds one run notice (`Run.Notices`),
  `pm.execution.setNextRequest is ignored in performance runs`. Performance runs never write to
  the app console (FR-044).

## R8 — `skipRequest`

- **Decision**: no sandbox change. `SkippedByScript` already exists
  (`entity/send_execution.go:69`). The runner maps that outcome to the row outcome `skipped`
  (FR-021).

## R9 — Run lifecycle, pause, stop, and one run per target

- **Active runs**: kept in an in-memory store behind a new use case, `usecase.ManageActiveRuns`,
  implemented in `internal/integration/persistence/active_runs.go`. This follows the in-memory
  precedent of `persistence/response_body.go`.
  - It is keyed by workspace and target id, and guarded by a mutex.
  - `Claim(target)` fails with `RUN-02409` when a run is already active on that target (FR-030).
  - The run's goroutine is owned by the runner service and started with a context detached from the
    binding call (`context.WithoutCancel` plus its own cancel), so `Start` returns the run id
    immediately.
- **Pause and resume**: a gate (channel plus state) checked before each row and during the delay
  wait (FR-029). Performance runs have no pause: Stop only, as in Postman.
- **Stop**: cancels the run context. The in-flight send ends with outcome `Cancelled`
  (`send_execution.go:89`), and that row's outcome is `cancelled`.
- **Workspace switch or close** (FR-031): the existing workspace close/activate services call
  `ManageRunnerService.StopWorkspace(localId)`, which stops every active run of that workspace and
  stores them as `Stopped`.
- **App quit**: active runs are abandoned and not stored (edge case).
- **Real time for waits**: the delay between rows, the performance duration, the controller tick
  and the once-per-second metrics wait on real time (`time.NewTimer` / `time.Ticker`, or `select`
  on `time.After` with the run context). They never use `timeutils.AfterFunc`: the BDD fake clock
  fires those timers only when a scenario advances time, so a run would never end. Timestamps
  (`StartedAt`, `FinishedAt`, durations) still come from `timeutils.Now()`, which the fake clock
  keeps moving with real time.

## R10 — Progress events

- **Decision**: two new Wails events, registered in `infra/server/event/events.go` with DTOs in
  `integration/publisher/dto`:
  - `runner:progress`: a run snapshot (id, target, status, current row and iteration, completed and
    planned counts, tests passed and failed, elapsed time, and, for performance runs, the
    last-second metrics). It is emitted after every row (FR-028) and, for performance runs, once
    per second (FR-045). It is throttled to at most 10 per second per run.
  - `runner:finished`: the final summary (status and totals). It drives the toast (FR-001b) and
    the Runs list refresh.

  The publisher follows `publisher/console_entries.go` (a usecase interface `PublishRunEvent`,
  then the publisher, then `RegisterEvent`).

## R11 — Results and response bodies

- **Row results** collect:
  - the `SendExecution` data: status, time, size, tests (`TestSummary`), script errors and outcome;
  - the sent request, masked with `usecase.MaskSecrets` over the row's secret values (the same
    masking the console uses);
  - the console entries, captured by giving `sendRun` a row collector. The collector tees every
    `console.append`. It skips the app console when logs are off (FR-027), and keeps nothing when
    logs are off.
- **Persisted responses** (FR-040): when on, `sendRun.send` calls the existing
  `saveResponse.Save(ctx, key, response)` with `key = entity.RunResponseKey(runId, rowIndex)`
  (`run:<runId>:<n>`) in place of a tab id. The existing memory limit and temp-file spilling apply.
  - The row detail fetches the body through the existing `ResponseBodyService` / in-app body route
    by that key. No new body route is needed.
  - Keys are released (`ReleaseResponse`) on run delete, clear, or cap pruning.
  - The store is in memory, so keys die with the app.

## R12 — Stored runs

- **Location**: `<DataDir>/workspaces/<localId>/runs/<targetId>/<runId>.json`, written
  atomically with `fileutils.WriteFileAtomic`. This is outside the workspace folder (FR-038).
- **Cap**: after each write, the files of that target are listed newest first (ULID run ids sort
  by time), and everything past 25 is deleted.
- **Persistence**: a new file, `persistence/run.go` (`usecase.FindRuns`, `SaveRun`,
  `DeleteRun`, `ClearRuns`), plus a model in `persistence/model/run.go`. It has a mutex per
  workspace, like `history.go`.
- **Workspace boundary**: the path is keyed by `localId` (US7-5).
- **Removal**: `removeWorkspace.Remove` also calls `usecase.DeleteWorkspaceRuns` (the whole
  `runs/` folder), next to its existing `DeleteUIState`, `DeleteVault` and `DeleteLocalVariables`
  calls.
- **Orphans**: when `Plan` reads the tree, it deletes the `runs/<targetId>/` folders and
  preferences of targets that no longer exist.
- **Unreadable files**: `FindRuns` skips a file it cannot read or decode, and logs a warning.
- **History**: runs never call `saveHistoryEntry` (FR-026).

## R13 — Performance engine and statistics

- **Scheduling** (FR-043): a domain function, `entity.LoadProfile.ActiveUsersAt(elapsed)`, returns
  the target user count at a time offset for:
  - Fixed;
  - Ramp up (initial → max over the ramp);
  - Spike (base, max during a centred window, base);
  - Peak (base → max over half the non-peak time, hold, back down).

  A controller tick of 100 ms starts or retires virtual users to match the target. A retired user
  finishes its current request first.
- **Statistics** (FR-045, FR-046): `entity.LatencyHistogram`, a log-linear bucketed histogram with
  ≤1% relative error from 1 µs to 10 minutes, gives the percentiles in bounded memory over 60
  minutes × 100 users. No new dependency (Constitution: new dependencies need justification).
  - It also tracks exact min, max, sum and count.
  - It is kept per request and overall, plus a per-second time-series bucket: active users,
    requests, latency sum and count, and errors.
  - Error = outcome not sent, or status ≥ 400.
- **Load**: 100 goroutines doing HTTP work is light for Go. The app stays responsive because the
  UI receives at most one aggregated event per second for a performance run (SC-006).

## R14 — Exports

- **JSON**: the stored run model, serialised as is (already masked).
- **JUnit XML**: built by a pure application function, `usecase.BuildRunJUnit(run)`, with
  `encoding/xml`.
  - Each row and iteration is a `<testsuite name="<folder path>/<request name> [iteration N]"
    tests= failures= errors= skipped= time=>`. N is 1-based. The folder path is relative to the
    run target and joined with `/`, and it is omitted (no leading slash) for requests at the target's
    root: `Login [iteration 1]`, `Orders/Create [iteration 1]`.
  - Each `pm.test` is a `<testcase name= classname="<request name>">`, with `<failure
    message=>` or `<skipped/>`.
  - A row that errored or was cancelled gets one `<testcase name="request"><error message=>`.
- **Writing**: the export is written through the existing export file writer
  (`persistence/export_file.go`) to a path chosen with `DialogService.PickSaveFile`.
- **Performance runs**: JSON only (FR-037).

## R15 — Runner preferences in UI state

- **Decision**: `entity.UIState` (persisted to `<DataDir>/workspaces/<localId>/state.json` by
  `persistence/ui_state.go`) gains `Runners map[string]RunnerPreferences`, keyed by target id.
  - **Contents**:
    - the order (request ids);
    - the unchecked set (a request id plus alias pairs are stored as `rows: [{request_id,
      alias_id, selected}]`);
    - the run type;
    - the functional and performance settings;
    - the data file path;
    - the environment id.
  - **Stale entries**: preferences whose target no longer exists are deleted by the runner service
    the next time that workspace's runner plans are read. `UIState.Normalized()` does not know the
    tree, so it does not do this.
- **Reconcile** (FR-007): a pure domain function, `entity.ReconcileRunRows(remembered, sidebar)`,
  returns the rows for the Runner:
  - removed requests are dropped;
  - new requests are inserted, checked, after the nearest preceding sidebar sibling that is
    present (or first);
  - aliases that no longer exist fall back to the Default (FR-008).
- **Data file** (FR-013): when the runner service builds the plan, a remembered data file path
  that no longer exists is cleared and reported (`DataFileMissing: true` on the plan DTO, which
  the UI shows as a notice).

## R16 — MCP surface

- **Decision**: two tools, following the read/edit split of `git` / `git_edit`:
  - **`run`** (`aitooleffect.Destructive`, `OpenWorld: true`, workspace scoped):
    - `action: "start"`: takes the target (path or id, resolved as the `tree` tool does) and
      optional rows, run type, settings and environment. It returns the run id.
    - `action: "stop"`: takes `run_id`.
  - **`runs`** (read-only):
    - `action: "get"`: `run_id` → status, progress, and the summary and rows once finished.
    - `action: "list"`: a target → stored runs.

  Wiring follows the MCP checklist (`controller/mcp_run.go`, `controller/mcp_runs.go`,
  `infra/server/mcp/tool_run.go`, `tool_runs.go`, `controllers.go`, `routes.go`,
  `infra/dependency/server/mcp_routes.go`).
- **Confirm before send**: same as `mcp_send.go:97-120`. An `SND-05409` from start goes through
  `approveAISendService.Request` once, with an approval request that describes the run
  ("Run <target> · N requests"). The approval DTO gains an optional `summary` field, shown by the
  approval dialog in place of the method and URL.
- **AI activity and changes**:
  - `run start` records `aichangekind.Cookies` when save-cookies is on, and `Variables` when
    keep-values is on. It never records `History`.
  - `run stop` and `runs` are not recorded as changes.

## R17 — Frontend placement

- **New feature folder**: `frontend/src/features/runner/`. It contains:
  - `model.ts` and `store.ts` (Zustand, keyed by target id, with an active-runs map);
  - components:
    - `RunnerSection`, with sub-tabs Setup, Results and Runs;
    - `RunOrderList`, a dnd-kit sortable with `KeyboardSensor` and `sortableKeyboardCoordinates`,
      modelled on `sidebar/hooks/use-environment-reorder.ts`;
    - `RunSettings`, `PerformanceSettings`, `RunProgress`, `RunResults`, `RunRowDetail`,
      `PerformanceResults` (inline-SVG time series, no chart dependency), `RunsList` and
      `RunFinishedToast`.
- **Platform**: `platform/services/runner.ts`, `platform/mappers/runner.ts`, and the
  `platform/events.ts` additions (`onRunProgress`, `onRunFinished`). The events are wired once in
  `app/bootstrap.ts`.
- **Item settings**:
  - `features/item-settings/model.ts` adds `'runner'` to `ItemSection`.
  - `ItemTab.tsx` gains a `runner` render prop and a running indicator prop.
  - `app/layout/ItemWorkspace.tsx` composes `RunnerSection` (app may import features).
- **Sidebar**: `features/sidebar/components/TreeItem.tsx` adds **Run…** for containers. It calls a
  new `onRunItem(id, kind)` prop. The app wiring (`app/`) opens the item tab pinned and sets the
  section to `runner` through `useItemSettingsStore` and `useTabsStore`.
- **Tabs**: the tab strip shows the running indicator from `useRunnerStore`. A feature may import
  another feature's `store.ts`.
- **Toast**: `shared/ui/toast/Toast.tsx`, rendered from `app/` on `runner:finished`.
- **Response detail**: `ResponseBody` is keyed by `tabId`. The runner's row detail passes the run
  response key as that id, so the response-viewer store's per-key entry works unchanged.
- **Data files**: a CSV/JSON filter is needed. `dto.PickFile` (`entrypoint/dto/pick_file.go`)
  gains `Filters []FileFilter{DisplayName, Pattern}`, and `binding/dialog.go` passes them to the
  Wails dialog.

## R18 — PRD

- **Decision**: `docs/prd/xenon-mvp-prd.md`:
  - L32 becomes "CLI (useful, but not needed to replace daily use)". The collection runner is
    removed from the non-goals.
  - A "Collection runner" section is added.
  - L507's "runner-only calls such as `setNextRequest` exist as no-ops" becomes "work inside
    collection runs; outside a run they are no-ops".
  - A **Decisions (Oct 23, 2026)** block records every user decision from the spec.
