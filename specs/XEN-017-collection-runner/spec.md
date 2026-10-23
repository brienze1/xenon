# Feature Specification: Collection Runner

**Feature Branch**: `xen-017`

**Created**: 2026-10-23

**Status**: Draft

**Input**: User description: "Collection runner inside the collection tab: run selected requests in order or as a load test, with Postman-style results. Clicking Run on a collection or folder opens its existing tab on a new Runner section (never a separate Runner tab), with request selection and order, run type (Functional or Performance), iterations and iteration data, Postman's run settings, live progress, results, past runs, export, and MCP access. Source: https://github.com/brienze1/xenon/issues/24"

## Context

Every request is sent by hand today, one tab at a time. A repeatable flow (log in, then create,
read, update and delete an object) cannot be run in one go with its test results in one place,
and an endpoint cannot be load-tested. Postman's **Collection Runner** covers both. Xenon adds the
same runner, with one deliberate difference: it lives **inside the collection's (or folder's)
own tab** as a **Runner** section, next to Overview, Authorization, Headers, Scripts, Variables
and Documentation. Xenon never opens a separate Runner tab.

The PRD lists "Collection runner and CLI" under **Non-goals for the MVP**. Constitution VI
("Postman parity, nothing more") puts anything outside the PRD out of scope, so the runner is a
**user decision**. The PRD records it (Decisions, Oct 23, 2026) and keeps "CLI" a non-goal.

### Decisions taken before this spec (user, 2026-10-23)

| Question (issue #24) | Decision |
| --- | --- |
| Scope of the first version | **Functional and Performance run types now.** Run method is **Local only**. Schedule and CLI/CI are out of scope. |
| Folders | **Every collection tab and every folder tab** gets a Runner section scoped to its contents. The sidebar context menu of a collection or folder gets **Run…**, which opens that item's tab on its Runner section. |
| Order | The selection and the custom order are **remembered per collection/folder in local UI state** on this machine and are never written to the git-tracked files. **Reset** restores the sidebar order and selects every request. |
| Aliases | Each row runs the request's **Default**. A per-row selector picks one of its aliases instead. |
| History | Runner sends are **not** recorded in History. Each finished run is stored locally in the app-data folder (never in the workspace or git). The **last 25 runs per collection/folder** are kept and listed so their results can be reopened. A stored run keeps the metadata, the tests, the console output and the sent request with secrets masked. Response bodies are kept **in memory only**, and only while "Persist responses for a session" is on. |
| Performance limits | Up to **100 virtual users** and up to **60 minutes**. Load profiles are **Fixed, Ramp up, Spike and Peak**. Pre-request and post-response **scripts run**. A stored performance run keeps the summary stats, the per-request stats and a per-second time series, and no individual responses. |
| Confirm before send | For an environment flagged "Confirm before send", the run is confirmed **once, when it starts**. |
| What a run sends; concurrency | A run sends every request **as saved on disk**, so unsaved tab drafts are ignored. The UI stays usable during a run (editing, sending from tabs). There is **one active run per collection/folder**. Runs on different collections or folders may overlap. |

### Settled from the issue, Postman parity and the code rather than asked

- **Keep variable values** follows Postman (Constitution VI). When it is on, values set by scripts
  during the run persist as current values, exactly as a normal send persists them. When it is
  off, changes are visible to later requests *within* the run and are discarded when the run
  ends, so the current values on disk are unchanged. Local-scope (`pm.variables`) values never
  outlive one request, as today.
- **Cookies** follow Postman.
  - With "Run collection without using stored cookies" off, the run starts from a copy of the
    workspace cookie jar. With it on, the run starts from an empty jar.
  - Either way, cookies set during the run are used by later requests in that run.
  - They are written back to the workspace jar only when "Save cookies after collection run" is
    on.
- **Run order** is the sidebar order of the collection or folder, flattened depth-first:
  - A folder's requests appear where the folder sits, in its own order.
  - Saved examples are not runnable rows.
  - Each request still inherits its ancestors' auth, headers and scripts, collection → folder →
    request, exactly as a single send does today (`ResolveEffectiveRequest`, `CollectScriptChain`).
- **`pm.execution.setNextRequest`** follows Postman:
  - It takes effect after the current request's post-response scripts finish, and the last call
    wins.
  - The target is matched among the run's rows by request id, or else by the first row in run
    order whose request name matches.
  - `null` ends the iteration after the current request. A target that is not in the run ends the
    iteration too, and a console warning names it.
  - Each iteration starts again at row 1.
  - Outside a run (a single send) it stays a no-op, as today.
- **`pm.execution.skipRequest()`** keeps today's rule: it is honoured in pre-request scripts only
  (`shims/pm.js`). A skipped row is reported as **skipped**.
- **Iteration data** follows Postman.
  - A CSV file must have a header row; each data row is one iteration. A JSON file must be an
    array of objects; each object is one iteration.
  - When a file is chosen, the iterations count defaults to the number of rows.
  - With more iterations than rows, the last row is reused.
  - Each row's fields are `pm.iterationData` and resolve as `{{field}}`. Iteration data sits
    between local variables and environment variables in precedence, as in Postman.
- **"Stop run if an error occurs"**: an error is a request that could not be sent (a network or
  timeout error) or a script error. A failed `pm.test` assertion is not an error, as in Postman.
  The run stops after that request and is marked **Stopped on error**.
- **Pause** waits for the request in flight to finish, then holds the run. **Stop** cancels the
  request in flight; that row is reported as cancelled and the run as **Stopped**. Every finished,
  stopped or failed run is stored.
- **History is untouched**: runner sends are never written to History (decision above), so the
  History caps and list stay exactly as they are.
- **Secrets in stored runs**: every vault secret value found in the stored sent request (URL,
  headers, body) and in the stored console output is replaced with `•••••`, the same masking the
  console already applies (`test/integration/features/console/console_masking.feature`).
- **Console**: unless "Turn off logs during run" is on, a run's script logs and requests appear in
  the app console as they do for a single send. They are also kept per row in the run results.
- **Missing secrets and expired tokens** follow the single-send contract (`SendAcknowledgement`,
  `internal/domain/entity/send_acknowledgement.go`). Missing secrets are acknowledged once for the
  whole run, like "Confirm before send". Expired tokens are always accepted inside a run, because
  Postman's runner never stops a run to ask for a token.
- **AI access (MCP)**: starting a run over MCP counts as sending for every request in it. The
  existing AI send policy (approval and "Confirm before send") applies once per run, as it does
  once per send today (`internal/integration/entrypoint/controller/mcp_send.go`).
- **Performance runs**:
  - Every virtual user loops over the selected rows in order for the duration.
  - Each user has its own variable and cookie state, starting from the run's environment and the
    jar as the cookie settings say.
  - Performance runs never persist variable values or cookies back.
  - An **error** is a request that could not be sent or got a status of 400 or higher.
  - `pm.execution.setNextRequest` is ignored in performance runs (one notice in the results),
    because each user loops over a fixed order. `skipRequest` is honoured.
  - Performance runs write nothing to the app console and do not report `pm.test` results, as in
    Postman's performance runs. Pause is not offered; Stop is.
  - Iteration data, when chosen, is consumed row by row by each user in turn and wraps around at
    the end.

## Clarifications

### Session 2026-10-23

- Decisions taken with the user before the spec are recorded in the table above.
- Q: How does the Runner section lay out setup, live progress/results and past runs? → A: Three inner sub-tabs — **Setup** (run order list on the left; run type, settings, environment and Start run on the right), **Results** (live progress, then the results of the current or opened run) and **Runs** (stored runs). Starting a run, Run again, or opening a stored run switches to Results.
- Q: Is the run configuration remembered per collection/folder? → A: Yes, all of it — run type, functional and performance settings, data file path and environment — locally per collection/folder like the order; a remembered data file that no longer exists is cleared with a notice.
- Q: Can stored runs be deleted by hand? → A: Yes — each run in the Runs sub-tab has Delete, and the sub-tab has **Clear runs** for that collection/folder behind a confirm dialog; the 25-run cap still applies.
- Q: Where is a run visible outside its Runner? → A: The collection/folder tab shows a running indicator while a run is active on it, and when a run ends a toast names the target with its status and passed/failed test totals and an **Open results** action that focuses (or opens) the tab on Runner › Results for that run.
- Settled from the code (not asked): the toast reuses the shared toast component (`frontend/src/shared/ui/toast/Toast.tsx`, already used for update notices).
- Settled (not asked): the Peak profile takes base users and a peak duration; the remaining time is split equally into a linear ramp up before the peak and a linear ramp down after it.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run a collection's requests in order (Priority: P1)

Luis opens his **Orders API** collection from the sidebar. The collection tab shows its sections
and a new one, **Runner**. He clicks Runner. Every request of the collection is listed in sidebar
order, numbered, each with its folder path, method and name, and every row is checked: "6 of 6
selected". He keeps Functional, leaves 1 iteration, and clicks **Start run**. While the run
progresses he sees the current request, a passed/failed counter and Stop/Pause. When it finishes,
the results list every request with its status, time, size and test results, and a summary shows
total requests, tests passed/failed/skipped, average response time and duration.

**Why this priority**: running a flow end to end with its tests in one place is the reason the
runner exists. Everything else builds on this slice.

**Independent Test**: give a collection three requests with `pm.test` assertions against a stub
server, start a functional run, and check that the three requests were sent in order with their
inherited auth and scripts and that the results hold the expected statuses and test outcomes.

**Acceptance Scenarios**:

1. **Given** a collection with requests in folders, **When** Luis opens its tab, **Then** the
   section strip includes **Runner** and the Runner lists every request in sidebar order, numbered,
   with folder path, method and name, all checked.
2. **Given** a selection, **When** he starts a functional run, **Then** the checked requests are
   sent in the shown order, each with its pre-request and post-response scripts and the inherited
   collection/folder auth, headers and scripts, and every `pm.test` result is collected.
3. **Given** a run in progress, **When** each request completes, **Then** the progress shows the
   current request and iteration and updated passed/failed counts.
4. **Given** a finished run, **When** results show, **Then** each row has method, name, resolved
   URL, status, time, size, its tests (name, pass/fail, error message), script errors and console
   output, and the summary shows totals, average response time and duration.
5. **Given** a request was edited in a tab but not saved, **When** the run sends it, **Then** the
   saved version on disk is sent, not the draft.
6. **Given** a run in progress, **When** Luis edits or sends other requests from tabs, **Then**
   both work and the run continues.

---

### User Story 2 - Choose and order what runs (Priority: P1)

Luis unchecks two requests ("4 of 6 selected"), drags **Delete order** to the end, and moves
**Login** to the top with the keyboard. He closes the tab and reopens it later: his selection and
order are still there. **Reset** puts every request back in sidebar order, all checked.

**Why this priority**: real flows run a subset in a specific order. Without it the runner only
replays the sidebar.

**Independent Test**: save a selection and a custom order for a collection, reload the UI state,
and check that the run uses exactly that selection and order, and that Reset restores the
sidebar order with every request checked.

**Acceptance Scenarios**:

1. **Given** the Runner list, **When** Luis unchecks rows, **Then** the counter reads "N of M
   selected" and only checked rows run.
2. **Given** the select-all checkbox, **When** he toggles it, **Then** every row is checked or
   unchecked, and it shows a mixed state when only some rows are checked.
3. **Given** the list, **When** he drags a row, or picks it up with Space and moves it with the
   arrow keys, **Then** the numbering follows the new order and the run uses it.
4. **Given** a custom selection and order, **When** the tab or the app is closed and reopened,
   **Then** both are restored, and nothing changed in the collection files on disk.
5. **Given** a request was added to or removed from the collection since the order was saved,
   **When** the Runner opens, **Then** removed requests are dropped, new requests appear checked
   at their sidebar position relative to their neighbours, and the rest keep the saved order.
6. **Given** **Reset**, **When** clicked, **Then** the sidebar order is restored and every row is
   checked.
7. **Given** a request with aliases, **When** its row shows, **Then** it runs the Default unless
   another alias is picked in that row's selector, and the choice is remembered with the order.

---

### User Story 3 - Run from a folder or the sidebar (Priority: P2)

Luis right-clicks the **Checkout** folder in the sidebar and picks **Run…**. The folder's tab opens
(or is focused) on its Runner section, listing only the requests inside that folder and its
subfolders.

**Why this priority**: folders group flows; running one flow without the rest of the collection
is common.

**Independent Test**: from the sidebar context menu of a folder, choose Run… and check that its
tab shows the Runner section listing only that folder's requests. Start a run and check that the
collection's scripts and auth still apply.

**Acceptance Scenarios**:

1. **Given** a collection or folder in the sidebar, **When** Luis opens its context menu,
   **Then** it contains **Run…**.
2. **Given** **Run…**, **When** chosen, **Then** that item's existing tab is focused (or opened)
   on its Runner section, and no other tab is created.
3. **Given** a folder's Runner, **When** a run starts, **Then** only that folder's requests (with
   its subfolders) are listed and sent, and each still inherits the collection's and the folder's
   auth, headers and scripts.

---

### User Story 4 - Iterations, data files and run settings (Priority: P2)

Luis picks `users.csv` (three rows, columns `email,password`) as iteration data. Iterations
becomes 3. He sets a 200 ms delay, turns on "Stop run if an error occurs" and turns off "Keep
variable values". The run executes the selection three times, each iteration using its row through
`{{email}}` and `pm.iterationData.get('email')`. When the run ends, his environment's current
values are as they were before.

**Why this priority**: data-driven flows and Postman's run settings are what make runs repeatable
and safe.

**Independent Test**: run a two-request collection with a 3-row CSV and check three iterations,
the row values in the sent requests, `pm.info.iteration`/`iterationCount` in scripts, the delay
between requests, and the variable and cookie settings' effect after the run.

**Acceptance Scenarios**:

1. **Given** iterations = N, **When** a run starts, **Then** the selection runs N times and
   `pm.info.iteration` (0-based) and `pm.info.iterationCount` (N) are set in every script.
2. **Given** a CSV or JSON data file, **When** chosen, **Then** iterations defaults to its row
   count. Each iteration exposes its row as `pm.iterationData` and as `{{field}}` variables, and
   rows past the end reuse the last row.
3. **Given** a malformed data file (CSV without a header row, JSON that is not an array of
   objects, unreadable file), **When** Luis starts the run, **Then** it does not start and an error
   names the problem.
4. **Given** a delay of D ms, **When** the run proceeds, **Then** at least D ms pass between the
   end of one request and the start of the next.
5. **Given** "Stop run if an error occurs" is on, **When** a request cannot be sent or a script
   throws, **Then** the run stops after that request with status **Stopped on error**. A failed
   assertion alone does not stop it.
6. **Given** "Keep variable values" is off, **When** a script sets an environment, collection or
   global variable, **Then** later requests in the run see the new value, and after the run the
   current values are unchanged. With it on, the values persist as after a normal send.
7. **Given** "Run collection without using stored cookies" is on, **When** the run starts,
   **Then** no stored cookie is sent, though cookies set during the run are sent by later requests
   in it.
8. **Given** "Save cookies after collection run" is on, **When** the run ends, **Then** cookies
   set during it are in the workspace jar. With it off, the workspace jar is unchanged.
9. **Given** an environment picked in the Runner that differs from the active one, **When** the
   run sends, **Then** it resolves variables against the picked environment, and the workspace's
   active environment is unchanged.
10. **Given** the run's environment is flagged "Confirm before send", **When** Luis clicks Start
    run, **Then** one confirmation dialog is shown. Confirming starts the whole run without
    further prompts. Cancelling starts nothing.
11. **Given** "Turn off logs during run" is on, **When** the run's scripts log, **Then** nothing
    from the run reaches the app console or the run's per-row console output.

---

### User Story 5 - Script control flow (Priority: P2)

Luis's Login request runs `pm.execution.setNextRequest('Get order')` to skip the sign-up step when
already registered. A teardown request's pre-request script calls `pm.execution.skipRequest()`
when there is nothing to delete.

**Why this priority**: existing Postman collections rely on these calls. Constitution VI requires
that they behave the same.

**Independent Test**: build collections whose scripts call `setNextRequest` with a name, an id,
`null` and an unknown name, and `skipRequest`. Check the sequence of sent requests and the
reported statuses.

**Acceptance Scenarios**:

1. **Given** a post-response script calls `pm.execution.setNextRequest('X')`, **When** the current
   request finishes, **Then** the run continues with row X (by id, or else the first row named X),
   even if that means going back.
2. **Given** several calls in one request, **When** it finishes, **Then** only the last call
   counts.
3. **Given** `setNextRequest(null)`, **When** the current request finishes, **Then** the iteration
   ends, and the next iteration (if any) starts at row 1.
4. **Given** `setNextRequest` names a request that is not in the run, **When** the current request
   finishes, **Then** the iteration ends and the console shows a warning naming the target.
5. **Given** a pre-request script calls `pm.execution.skipRequest()`, **When** the run reaches
   it, **Then** the request is not sent and its row is reported as **skipped**.
6. **Given** a single send outside a run, **When** a script calls `setNextRequest`, **Then**
   nothing changes (it stays a no-op).

---

### User Story 6 - Inspect, filter, re-run and export results (Priority: P2)

After a run Luis filters the results to **Failed**, opens one row and sees the request that was
sent and the response that came back (headers and body), as in the response viewer. He clicks
**Run again** to repeat it with the same settings, and exports the results as JUnit XML for a CI
report.

**Why this priority**: results are only useful if failures can be found, inspected and shared.

**Independent Test**: run a collection with one passing and one failing request. Check the
filters, the detail view, and that Run again reuses the settings. Export JSON and JUnit XML and
validate their contents.

**Acceptance Scenarios**:

1. **Given** results, **When** Luis filters by Passed, Failed or Skipped, **Then** only matching
   rows show. A row passes when all its tests pass and it had no error.
2. **Given** a row, **When** opened, **Then** the detail shows the sent request (method, resolved
   URL, headers and body, with secrets masked) and, while response bodies are persisted, the
   response status, headers and body with the response viewer's handling of large and binary
   bodies.
3. **Given** "Persist responses for a session" was off, **When** a row is opened, **Then** the
   response shows status, time and size, and the body area says responses were not persisted for
   this run.
4. **Given** **Run again**, **When** clicked, **Then** a new run starts with the same selection,
   order, aliases, run type, iterations, data file, settings and environment.
5. **Given** **Export**, **When** JSON is chosen, **Then** a file holds the run's settings,
   summary and every row with its tests, exactly as stored (secrets masked). **When** JUnit XML is
   chosen, **Then** a file holds one test suite per row and iteration, with one test case per
   `pm.test` (failures carry the error message).

---

### User Story 7 - Past runs (Priority: P3)

Luis reopens the Runner a day later. A **Runs** list shows the collection's recent runs (date,
run type, status, totals). He opens Tuesday's run and its results show as they did then.

**Why this priority**: comparing against a previous run matters, but each run is useful on its
own first.

**Independent Test**: finish more than 25 runs on one collection and check that only the 25 most
recent are listed and stored on disk, that each reopens with its results, and that nothing was
written to the workspace folder or to History.

**Acceptance Scenarios**:

1. **Given** finished runs, **When** the Runner opens, **Then** its Runs sub-tab shows them newest
   first, with start time, run type, status, environment name, total requests, tests
   passed/failed/skipped and duration.
2. **Given** a stored run, **When** opened, **Then** its results, filters, detail and export work
   as in User Story 6. Response bodies show only if they were persisted during this app session.
3. **Given** a 26th run finishes on the same collection or folder, **When** it is stored, **Then**
   the oldest run of that collection or folder is deleted.
4. **Given** any run, **When** it finishes, **Then** History has no new entries, and no file inside
   the workspace folder changed because of the run's storage.
5. **Given** two workspaces, **When** runs are listed, **Then** each workspace shows only its own
   runs.
6. **Given** a stored run, **When** Luis deletes it, **Then** it disappears from the list and from
   disk. **When** he chooses **Clear runs** and confirms, **Then** every stored run of that
   collection or folder is deleted. Cancelling the confirmation deletes nothing.
7. **Given** a run ends while Luis is on another tab, **When** it ends, **Then** a toast names the
   collection or folder with its status and test totals. **Open results** takes him to that run's
   results.

---

### User Story 8 - Performance run (Priority: P3)

Luis switches the run type to **Performance**, sets 20 virtual users, 2 minutes and a **Ramp up**
profile, and starts it. A live chart shows response time, throughput and error rate. When it ends,
a summary shows average, p90, p95, p99, min and max response time, requests per second and error
rate, overall and per request.

**Why this priority**: load testing is valuable but secondary to functional runs.

**Independent Test**: run a short performance test against a stub server with a known latency and
check that the request counts, percentiles, throughput and error rate fall within the expected
bounds, and that the profile shapes the number of active users over time.

**Acceptance Scenarios**:

1. **Given** Performance, **When** settings show, **Then** they include virtual users (1–100),
   duration (1 second–60 minutes), load profile (Fixed, Ramp up, Spike, Peak) and the profile's own
   parameters, and Start run rejects values out of range.
2. **Given** a running performance test, **When** it progresses, **Then** live metrics update at
   least once per second: active users, requests sent, average response time, throughput and error
   rate.
3. **Given** a finished performance run, **When** results show, **Then** overall and per-request
   stats are shown: total requests, average, p90, p95, p99, min and max response time, requests per
   second, error rate and the count per status code. A per-second time series is also shown.
4. **Given** a performance run, **When** it runs, **Then** pre-request and post-response scripts
   execute for every request, variables and cookies are never persisted afterwards, nothing is
   written to the app console, and `setNextRequest` is ignored with one notice in the results.
5. **Given** a running performance test, **When** Luis looks at the controls, **Then** only Stop
   is offered (no Pause).
6. **Given** the **Fixed** profile, **Then** all users are active for the whole duration. With
   **Ramp up**, users grow linearly from the initial count to the maximum over the ramp period, then
   hold. With **Spike**, a base count holds, jumps to the maximum for the spike period in the middle
   of the run, and returns to base. With **Peak**, a base count ramps to the maximum, holds the peak
   for the peak period, and ramps back down.
7. **Given** Stop, **When** clicked during a performance run, **Then** in-flight requests are
   cancelled, and the stats collected so far are stored as a **Stopped** run.

---

### User Story 9 - AI clients run collections (Priority: P3)

An AI client connected over MCP starts a run of the **Orders API** collection, polls its progress,
reads the results, and could stop it.

**Why this priority**: AI access mirrors every other Xenon feature, but the UI flow comes first.

**Independent Test**: through the MCP tools, start a functional run on a collection, read progress
until it finishes, read the results, and check that they match what the UI shows. Start another
run and stop it.

**Acceptance Scenarios**:

1. **Given** an MCP client with access, **When** it starts a run on a collection or folder (with an
   optional selection, order, aliases, run type, iterations, data, settings and environment),
   **Then** the run starts exactly as from the UI and the client receives the run id.
2. **Given** a run id, **When** the client reads it, **Then** it gets the status, the progress
   (current row and iteration, counts), and once finished the summary and the per-row results
   (tests and errors, with secrets masked).
3. **Given** a running run, **When** the client stops it, **Then** it stops as from the UI.
4. **Given** the run's environment is flagged "Confirm before send", **When** an AI client starts
   a run, **Then** the existing AI send approval flow asks once for the whole run.
5. **Given** an active run started by the AI, **When** the user opens the collection's Runner,
   **Then** the user sees that run's live progress.

### Edge Cases

- **Empty selection**: Start run is disabled while no row is checked, and the MCP start is rejected
  with an error.
- **Empty collection or folder**: the Runner shows an empty state and Start run is disabled.
- **A second run on the same collection or folder** while one is active is rejected: the UI shows
  the active run instead, and MCP gets an error naming the active run.
- **A request is deleted or renamed during a run**: the run sends what was saved when that row's
  turn came. A request missing from disk at its turn is reported as an error ("request not found").
  It counts toward "Stop run if an error occurs".
- **A request in the selection is moved out of the collection or folder** before the run: it is no
  longer listed, so it does not run.
- **Pre-request script fails**: the request is not sent, the row reports the script error, and it
  counts as an error.
- **Missing secrets** (a secret with no vault value): Start is refused with the same
  missing-secrets error a single send gets, naming the missing variables across every selected row.
  The UI shows the same dialog as for a send. **Run anyway** restarts the run with the missing
  secrets acknowledged, so their placeholders go out literally, exactly as "Send anyway" does today
  (`SND-06409`, `test/integration/features/sending/preflight_missing_secrets.feature`).
- **Expired OAuth 2.0 token** (own or inherited auth): rows never prompt, as in Postman's runner.
  The token is refreshed when it can be, as for a single send. Otherwise the stored token is sent
  as is, and the target's response shows the failure. A single send would instead refuse with
  `SND-07409`.
- **Iterations of 0, a negative delay, or a non-numeric field**: Start run rejects them with a
  field-level error. Iterations must be 1 or more and the delay 0 to 3,600,000 ms.
- **A data file with more columns in some rows than in the header**: extra cells are ignored, and
  missing cells are empty strings (CSV).
- **The workspace is switched or closed during a run**: the run is stopped and stored as Stopped.
- **The app quits during a run**: the run in progress is not stored.
- **A `setNextRequest` cycle** (A → B → A …) runs until stopped, as in Postman. Stop ends it.
- **A huge response in a functional run** with persisted responses: the body is held as the
  response viewer holds large bodies (spilling to a temporary file past the memory limit) and is
  shown with the same raw-first rule for very large bodies.
- **A performance run against an unreachable host**: every request counts as an error. The run
  continues for its duration unless stopped.

## Requirements *(mandatory)*

### Functional Requirements

**Placement and entry**

- **FR-001**: Every collection tab and every folder tab MUST include a **Runner** section next to
  its existing sections. Selecting it MUST NOT open a new tab. The Runner section MUST have three
  sub-tabs:
  - **Setup**: the run order list; run type, settings, environment and Start run.
  - **Results**: live progress, then the results of the current or opened run.
  - **Runs**: the stored runs.

  Starting a run, Run again, or opening a stored run MUST switch to Results.
- **FR-001a**: While a run is active on a collection or folder, that item's tab (when open) MUST
  show a running indicator.
- **FR-001b**: When a run ends (completed, stopped or stopped on error), a toast MUST name the
  target, the run status and the passed/failed test totals, with an **Open results** action. Open
  results focuses (or opens) the target's tab on Runner › Results for that run.
- **FR-002**: The sidebar context menu of every collection and folder MUST include **Run…**. It
  focuses (or opens) that item's tab and selects its Runner section on the Setup sub-tab (or on Results while a run is active there).

**Selection and order**

- **FR-003**: The Runner MUST list every request under the collection or folder (subfolders
  included, examples excluded) in sidebar order, flattened depth-first, numbered from 1. Each row
  MUST show the folder path relative to the run target, the method and the name.
- **FR-004**: Each row MUST have a checkbox. A select-all checkbox (checked, unchecked or mixed)
  MUST toggle every row. An "N of M selected" counter and a **Reset** action MUST be shown. Reset
  restores the sidebar order and checks every row.
- **FR-005**: Rows MUST be reorderable by pointer drag and drop and by keyboard (Space to pick up,
  arrow keys to move, Space to drop, Escape to cancel), with screen-reader announcements.
- **FR-006**: The selection, the order and the per-row alias choice MUST be remembered per
  collection/folder in the workspace's local UI state on this machine, and MUST NOT be written to
  any file inside the workspace folder.
- **FR-007**: When the remembered order is reconciled with the current tree, requests that no
  longer exist MUST be dropped. New requests MUST be inserted checked after the nearest preceding
  sidebar sibling that is in the list (or first). Remaining rows keep their remembered order.
- **FR-008**: A row whose request has aliases MUST offer an alias selector (Default first, then its
  aliases). The row runs the chosen version, and the Default when the remembered alias no longer
  exists.

**Run configuration**

- **FR-009**: The Runner MUST offer the run type **Functional** or **Performance**. The run method
  is Local, and no other run method is offered.
- **FR-010**: Functional settings MUST include:
  - iterations (an integer ≥ 1, default 1);
  - an optional iteration data file (CSV or JSON), picked with a file dialog;
  - a delay between requests (0–3,600,000 ms, default 0);
  - toggles, with Postman's defaults:
    - Persist responses for a session (off);
    - Turn off logs during run (off);
    - Stop run if an error occurs (off);
    - Keep variable values (on);
    - Run collection without using stored cookies (off);
    - Save cookies after collection run (off).
- **FR-011**: Performance settings MUST include:
  - virtual users (1–100, default 10);
  - duration (1 second to 60 minutes, default 10 minutes);
  - a load profile:
    - **Fixed**;
    - **Ramp up**: initial users (default 1) and ramp duration (default a quarter of the
      duration, at least 1 s);
    - **Spike**: base users (default a tenth of the virtual users, at least 1) and spike duration
      (default a tenth of the duration, at least 1 s);
    - **Peak**: base users (default a tenth of the virtual users, at least 1) and peak duration
      (default half the duration, at least 1 s). The ramps split the rest of the duration equally.
  - an optional iteration data file.

  Each profile parameter MUST fit within the duration and the user count. Defaults are recomputed
  when the duration or the user count changes, unless the user has edited that parameter.
- **FR-012**: The Runner MUST show the run's environment, defaulting to the workspace's active
  environment, and allow picking another environment or "No environment" for the run without
  changing the workspace's active environment.
- **FR-013**: The run configuration (run type, settings, data file path, environment) MUST be
  remembered per collection/folder in local UI state, like the order. A remembered data file that
  no longer exists MUST be cleared with a notice when the Runner opens.
- **FR-014**: Start run MUST validate the configuration before any request is sent, and reject:
  - an empty selection;
  - out-of-range values;
  - a malformed or unreadable data file;
  - missing secrets (unless the start acknowledges them, as "Send anyway" does);
  - an already active run on the same collection or folder.

  Each rejection MUST carry a specific error.
- **FR-015**: When the run's environment is flagged "Confirm before send", starting MUST require
  one confirmation for the whole run, and cancelling MUST start nothing.

**Functional execution**

- **FR-016**: A functional run MUST send the checked rows in the shown order, once per iteration.
  Each row sends its request (the chosen alias version) as saved on disk at the moment its turn
  comes, through the same pipeline as a single send: effective auth and headers inherited from
  ancestors, the pre-request chain collection → folder → request, the send, then the
  post-response chain. Every `pm.test` result MUST be collected.
- **FR-017**: Unsaved tab drafts MUST NOT affect a run. The user MUST be able to edit, save and
  send from tabs while a run is active.
- **FR-018**: `pm.info.iteration` (0-based) and `pm.info.iterationCount` MUST reflect the run in
  every script. `pm.iterationData` and `{{field}}` resolution MUST expose the current iteration's
  data row. Iteration data ranks between local and environment variables in precedence.
- **FR-019**: Iteration data parsing:
  - CSV MUST use the first row as the field names. Extra cells are ignored and missing cells are
    empty strings.
  - JSON MUST be an array of objects.
  - Files MUST be read as UTF-8, and a leading byte-order mark is ignored.
  - `pm.iterationData` keeps JSON value types. `{{field}}` uses the value's text form: strings as
    is, numbers and booleans as written, `null` as an empty string, and objects and arrays as
    compact JSON.
  - The iterations count MUST default to the row count when a file is chosen.
  - With fewer iterations than rows, only the first rows are used. Iterations beyond the last row
    reuse the last row.
- **FR-020**: `pm.execution.setNextRequest(target)` MUST, in a functional run:
  - count calls from both pre-request and post-response scripts of the row (collection, folder
    and request levels), including a row skipped with `skipRequest`, and apply the last one;
  - take effect when the row finishes (after its post-response scripts, or right after the skip);
  - resolve `target` by request id, or else by the first selected row in run order with that name;
  - end the iteration on `null` or on a target that is not a selected row of the run, the latter
    with a console warning naming the target.

  A `setNextRequest` cycle repeats until the user stops the run, as in Postman. There is no loop
  limit.

  Outside a run it MUST remain a no-op.
- **FR-021**: `pm.execution.skipRequest()` in a pre-request script MUST skip sending that row and
  report it as **skipped**.
- **FR-022**: The delay MUST elapse between the end of one row (its post-response scripts
  included) and the start of the next, across iteration boundaries too. No delay follows the last
  row of the run.
- **FR-023**: With "Stop run if an error occurs" on, the run MUST stop after the first row whose
  request could not be sent, whose request no longer exists, or whose scripts raised an error. Its
  status is **Stopped on error**. Failed assertions alone never stop a run.
- **FR-024**: With "Keep variable values" on, variable changes made by scripts (environment,
  collection, globals, secret variables and `pm.vault`) MUST persist exactly as after a single
  send. With it off, the same changes MUST be visible to later rows of the run and MUST be
  discarded at the end, leaving current values, the vault and the workspace files unchanged.
- **FR-025**: Cookies:
  - Without "Run collection without using stored cookies", a run MUST start from a copy of the
    workspace jar. With it, a run MUST start from an empty jar.
  - Cookies set during the run MUST be sent by later rows according to the usual cookie rules.
  - The workspace jar MUST change only when "Save cookies after collection run" is on, by merging
    the run's cookies when the run ends.
- **FR-026**: Runner sends MUST NOT be recorded in History.
- **FR-027**: With "Turn off logs during run" on, the run MUST NOT emit console entries to the app
  console or keep them in its results. Otherwise its entries MUST reach the app console as for a
  single send and be kept per row.

**Control and progress**

- **FR-028**: During a run the Runner MUST show live progress:
  - status (running, paused);
  - current row and iteration;
  - completed rows out of the planned total;
  - passed and failed test counts;
  - elapsed time.

  Progress MUST update after every row.
- **FR-029**: In a functional run:
  - **Pause** MUST hold the run once the row in flight completes. A pause during the delay holds
    before the next row starts.
  - **Resume** MUST continue with the next row, and paused time does not count as a delay.
  - **Stop** MUST cancel the request in flight, report that row as cancelled, and end the run with
    status **Stopped**.

  Performance runs offer **Stop** only, as Postman does, because pausing would distort the
  profile.
- **FR-030**: At most one run MUST be active per collection or folder. Runs on different
  collections or folders MAY run at the same time, alongside single sends. A collection and one
  of its folders are different targets.
- **FR-031**: Switching or closing the workspace MUST stop its active runs, storing them as
  **Stopped**.

**Results**

- **FR-032**: A functional run's results MUST list each executed row per iteration with:
  - iteration number, method, name, alias name (when not Default) and resolved URL;
  - status code, response time and response size;
  - outcome: passed, failed, skipped, error or cancelled;
  - its tests (name, passed/failed/skipped, error message), script errors and console entries.

  The results list MUST stay scrollable without visible lag for at least 10,000 row results.
  Only visible rows are rendered.
- **FR-033**: The summary MUST show:
  - run status (Completed, Stopped, Stopped on error);
  - start time and duration;
  - iterations and total requests;
  - tests passed, failed and skipped;
  - average response time.
- **FR-034**: Results MUST be filterable by **All**, **Passed**, **Failed** and **Skipped**:
  - Passed means no error and every test passed.
  - Failed means any failed test, an error, or a cancelled row.
  - Skipped means skipped by script.
- **FR-035**: Opening a row MUST show:
  - the sent request: method, resolved URL, headers and body, with secrets masked;
  - the response status, time and size;
  - when the run persisted responses and this app session still holds them, the response headers
    and body, rendered as the response viewer renders them (large and binary handling included);
  - otherwise, a note that responses were not persisted.
- **FR-036**: **Run again** MUST start a new run with the stored run's target, selection, order,
  aliases, run type, settings, data file and environment.
- **FR-037**: **Export** MUST save, through a save-file dialog:
  - **JSON**: the stored run as is (settings, summary, rows, tests, console, sent requests, with
    secrets masked);
  - **JUnit XML**: one `<testsuite>` per executed row and iteration, one `<testcase>` per test, a
    `<failure>` with the error message for failed tests, `<skipped/>` for skipped tests, and an
    `<error>` test case for a row that errored.

  Performance runs export JSON only (their summary, per-request stats and time series).

**Storage**

- **FR-038**: Every finished, stopped or failed run MUST be stored in the workspace's app-data
  folder (outside the workspace folder, never committed). The 25 most recent runs MUST be kept per
  collection or folder, and older ones deleted when a new run is stored. Removing a workspace from
  Xenon MUST delete its stored runs, as it deletes its UI state and local variables. Stored runs of
  a collection or folder that no longer exists MUST be deleted the next time that workspace's
  runner data is read. A stored run file that cannot be read MUST be left out of the list, with a
  logged warning, rather than failing the list.
- **FR-039**: A stored run MUST hold:
  - the run target;
  - the configuration (selection, order, aliases, settings, environment id and name, data file
    name);
  - the status and timing;
  - the summary;
  - per row: metadata, tests, script errors, console entries and the sent request.

  Every secret value MUST be replaced with `•••••` before the run is stored, exported or returned
  over MCP. This covers vault secrets known at the start and secret values set during the run. It
  applies wherever such a value appears:
  - the resolved URL;
  - the sent request's headers and body;
  - console entries;
  - test names and error messages;
  - script errors;
  - the row error.

  Response headers and bodies MUST NOT be stored. Iteration data values are not secrets: they
  appear as sent, as in Postman. Credentials belong in secret variables.
- **FR-040**: With "Persist responses for a session" on, a functional run's response headers and
  bodies MUST be kept in memory for the rest of the app session (bodies above the memory limit
  spill to temporary files, as the response viewer does) and dropped when the app quits or the run
  is deleted by the 25-run cap.
- **FR-041**: The Runner MUST list the stored runs of its collection or folder, newest first:
  - start time, run type and status;
  - environment name;
  - total requests;
  - tests passed, failed and skipped;
  - duration.

  Opening one MUST show its results. Runs MUST be scoped to their workspace.
- **FR-041a**: Each stored run MUST offer **Delete**. The Runs sub-tab MUST offer **Clear runs**,
  which deletes every stored run of that collection or folder after a confirm dialog. Deleting a
  run also drops any response bodies it holds in memory.

**Performance execution**

- **FR-042**: A performance run MUST start virtual users according to the profile. Each user loops
  over the checked rows in order until the duration ends or the run is stopped, running scripts as
  a functional run does. Each user MUST have its own variable and cookie state. Performance runs
  MUST NOT persist variables or cookies and MUST NOT write History.
- **FR-043**: Load profiles MUST shape the active user count over time:
  - **Fixed**: the maximum for the whole duration.
  - **Ramp up**: linear from the initial count to the maximum over the ramp duration, then the
    maximum.
  - **Spike**: the base count, the maximum for the spike duration centred in the run, then the
    base count.
  - **Peak**: the base count ramps linearly to the maximum over half of the time not taken by the
    peak duration, holds the maximum for the peak duration, then ramps linearly back to the base
    count over the other half.
- **FR-044**: Performance runs MUST:
  - ignore `setNextRequest`. The run's results show one notice, "pm.execution.setNextRequest is
    ignored in performance runs".
  - honour `skipRequest`;
  - consume iteration data rows in turn per user, wrapping around;
  - write nothing to the app console and keep no per-request console output, so 100 users cannot
    flood the console;
  - run `pm.test` assertions without reporting them, as Postman's performance runs do. Only the
    request-level metrics are reported.
- **FR-045**: Performance live metrics MUST update at least once per second: active users,
  requests sent, average response time, throughput (requests/s) and error rate. An error is a
  request that could not be sent or got a status ≥ 400.
- **FR-046**: A finished performance run MUST report, overall and per request:
  - total requests;
  - average, p90, p95, p99, min and max response time. Response time is the total request time
    the response viewer shows. Percentiles are accurate within 1% of the exact value.
  - throughput: completed requests divided by the elapsed run time, in requests per second;
  - error rate;
  - the count per status code.

  It MUST also report a per-second time series (active users, requests, average response time,
  errors). These are stored as in FR-038; no individual responses are kept.

**AI access**

- **FR-047**: The MCP server MUST let AI clients:
  - start a run on a collection or folder with the options the UI offers (all optional, defaulting
    to the remembered or default configuration);
  - read a run's status, progress and results (secrets masked);
  - list a target's stored runs;
  - stop a run.

  These operations MUST follow the existing AI access and send-approval policies, with approval
  asked once per run.
- **FR-048**: A run started over MCP MUST be visible live in the target's Runner section, and a
  run started in the UI MUST be readable over MCP.

**PRD**

- **FR-049**: The PRD MUST remove "Collection runner" from the MVP non-goals ("CLI" remains a
  non-goal), add a Collection runner section, and record the decisions in this spec as user
  decisions (Decisions, Oct 23, 2026). It MUST also change the "runner-only calls exist as no-ops"
  note so that it says they work inside runs.

### Key Entities

- **Run target**: the collection or folder a run belongs to (workspace, item id, kind).
- **Run plan**: the ordered, selected rows. Each row is a request id plus the chosen alias (Default
  when none).
- **Run configuration**: the run type, the functional settings (iterations, data file, delay,
  toggles), the performance settings (users, duration, profile and its parameters) and the
  environment.
- **Runner preferences**: the plan and configuration remembered per run target in local UI state.
- **Iteration data**: the ordered rows (field → value) parsed from a CSV or JSON file.
- **Run**: one execution. It has an id, target, configuration, status (Running, Paused, Completed,
  Stopped, Stopped on error), start time, duration, progress and results.
- **Row result**: one executed row in one iteration: the request identity, alias, sent request
  (masked), status, time, size, outcome, tests, script errors and console entries.
- **Performance stats**: overall and per-request aggregates, plus a per-second time series.
- **Stored run**: a finished Run as persisted in the app-data folder. It is capped at 25 per run
  target.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can start a functional run of a whole collection in 2 clicks from its tab
  (Runner, Start run) or 2 from the sidebar (Run…, Start run).
- **SC-002**: For a collection whose requests and scripts pass when sent one by one, a functional
  run reports the same statuses and test outcomes for 100% of the requests.
- **SC-003**: Live progress reflects each completed request within 1 second of its completion.
- **SC-004**: Exported JUnit XML is accepted by a standard JUnit report reader with one test case
  per `pm.test` and the same pass/fail counts as the run summary.
- **SC-005**: An existing Postman collection that uses `setNextRequest`, `skipRequest`,
  `pm.iterationData` and `pm.info.iteration` follows the same request sequence in Xenon as in
  Postman's runner.
- **SC-006**: A performance run of 100 virtual users for 60 minutes completes while the rest of
  the app stays responsive (tabs, sending, editing).
- **SC-007**: No secret value from the vault appears in any stored or exported run file (verified
  by scanning the files for every secret value).
- **SC-008**: After any run, History and every file inside the workspace folder are unchanged
  (except variable files when "Keep variable values" is on, exactly as a single send would change
  them).

## Assumptions

- The existing send pipeline (auth inheritance, script chains, OAuth tokens, cookies, response
  body spilling) is reused for each runner send. The runner adds sequencing, iteration context,
  variable and cookie isolation, and results, not a second HTTP engine.
- "Postman semantics" for settings, iteration data and `setNextRequest` refer to Postman's
  Collection Runner as of 2026, and constitution VI makes them binding where this spec does not
  say otherwise.
- The 25-run cap counts a collection and each of its folders separately.
- Performance runs execute on the user's machine. Reaching 100 virtual users depends on the
  machine and the target. Xenon does not throttle below the configured load.
- The CLI ("Postman CLI" and "Configure CI/CD" run methods) and scheduled runs stay out of scope.
  JUnit export covers the CI reporting need.
