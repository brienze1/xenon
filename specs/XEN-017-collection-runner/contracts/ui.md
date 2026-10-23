# Contract: Runner UI

## Collection/folder tab

- `ItemSection` gains `'runner'`. The section strip order is: Overview, Authorization, Headers,
  Scripts, Variables (collections only), Documentation, **Runner**.
- The tab label shows a small pulsing dot (accent token) while `useRunnerStore.active[itemId]` is
  set (FR-001a). Its `aria-label` gets the suffix ", run in progress".

## Runner section — sub-tabs (Radix-style `Tabs`, the existing `shared/ui/tabs`)

### Setup

- **Left: Run order**
  - The header has the select-all checkbox (`aria-checked="mixed"` when partial), "N of M selected"
    and **Reset**.
  - Each row has a drag handle, a checkbox, its number, the folder path (muted), a method badge
    (`shared/ui/method-badge`), the name, and an alias select when the request has aliases
    (Default first).
  - The list is a dnd-kit `SortableContext` with `PointerSensor` and `KeyboardSensor`
    (`sortableKeyboardCoordinates`). Space picks a row up or drops it, the arrow keys move it, and
    Escape cancels. Screen-reader announcements follow `use-environment-reorder.ts`.
  - An empty target shows `shared/ui/empty-state` ("No requests to run") and Start run is disabled.
- **Right: configuration**
  - Run type: a segmented control, Functional or Performance.
  - Functional: Iterations (number), Iteration data (Select file… / file name / ×), Delay (ms), and
    six checkboxes in Postman's order with Postman's labels.
  - Performance: Virtual users, Duration (minutes and seconds), Load profile (select) and the
    profile's fields, Iteration data, and a small preview line of the user curve (inline SVG).
  - Environment: a select listing "No environment" and every environment, defaulting to the active
    one. A confirm-before-send environment shows its colour.
  - **Start run** is the primary button, disabled while the selection is empty or a run is active
    on this target. Field errors from `RUN-04400` show inline under the field.
  - A `data_file_missing` notice shows as a `shared/ui/banner`.
- **Saving preferences**: every change saves through `RunnerService.SavePreferences`, debounced by
  300 ms.

### Results

- **During a run**: a progress header shows the status, current request and iteration, a
  completed/planned bar, passed and failed counts, and elapsed time, with **Pause/Resume** and
  **Stop**. Performance runs show **Stop** only.
- **Functional results**:
  - The summary strip shows the status, duration, iterations, total requests, tests
    passed/failed/skipped and average response time.
  - Filter chips: All, Passed, Failed, Skipped.
  - The rows are grouped by iteration and virtualized, so only visible rows render (FR-032,
    10,000 rows). Each shows an outcome icon, method, name (and `· alias`),
    URL, status, time and size. A row expands to its tests (`TestResultsTab`), script errors and
    console entries.
  - **Opening a row** shows a side panel with:
    - Request: method, URL, headers and body;
    - Response: `StatusLine`, `ResponseHeaders`, and `ResponseBody` keyed by
      `run:<runId>:<index>`; or the note "Responses were not persisted for this run."
- **Performance results**: the summary stats table (overall plus a row per request) and an
  inline-SVG time-series chart (response time, throughput, errors, and active users as a stepped
  line).
- **Actions**: **Run again** and **Export** (JSON / JUnit XML). JUnit is hidden for performance
  runs.

### Runs

- Stored runs, newest first: start time, type, a status badge, environment, totals and duration.
- Click a run to open it in Results. Each row has a ⋯ menu with Delete. **Clear runs** opens a
  confirm dialog (`shared/ui/dialog`).

## Sidebar

The context menu (and the ⋯ menu) of collections and folders gets **Run…** after Open. It opens the
tab pinned, on the Runner section, on the Setup sub-tab (or Results while a run is active there).

## Toast

On `runner:finished`, `shared/ui/toast/Toast` shows "<target> run <completed|stopped|stopped on
error> · N passed · M failed" with **Open results**. Open results focuses or opens the tab on
Runner › Results for that run. The toast auto-dismisses after 8 s.

## Store — `features/runner/store.ts`

- **State**: `targets: Record<itemId, { plan, config, view: 'setup'|'results'|'runs', runId?,
  run?, runs?, failure? }>` and `active: Record<itemId, RunSummaryView>`.
- **Actions**: `load`, `updateConfig` (debounced save), `reset`, `start` (handles `SND-05409` by
  setting `confirmation`, and `SND-06409` with the missing-secrets dialog whose **Run anyway** restarts
  with `acknowledge_missing_secrets`, as `request-builder` does for sends), `pause`, `resume`, `stop`, `open(runId)`,
  `deleteRun`, `clearRuns`, `export`, `applyProgress(event)`, `applyFinished(event)` and
  `reset()` on workspace switch.
