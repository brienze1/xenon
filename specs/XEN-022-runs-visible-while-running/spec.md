# Feature Specification: Runs Are Visible While They Run

**Feature Branch**: `xen-022`

**Created**: 2026-10-27

**Status**: Draft

**Input**: User description: "Runs started from the AI assistant (MCP `run` tool, action start) and runs in progress are invisible in parts of Xenon. Today a collection/folder run started over MCP executes normally and `runs get` returns it, but `runs list` omits it (only finished, stored runs are listed), the target's Runs history in the UI omits it, and an open Runner tab for that collection/folder that already shows an earlier run keeps showing the earlier run instead of the new one — so the user cannot watch, review or stop the AI-started run. The same list gap affects UI-started runs while they are running. Users need: (1) a run started over MCP to appear in the target's Runner tab exactly like a UI-started one — live progress, rows, tests, console and stop/pause controls — the open tab follows the newly started run of its target and shows Results; (2) Runs history and `runs list` to include in-progress runs (running/paused) together with stored runs, newest first by start time; (3) `runs list` and `runs get` to agree — any run `get` can return is also in `list`, and a run never appears twice in the list (during the brief moment a finished run is both stored and still held as active, the stored copy wins). Only one run can be active per collection/folder at a time (a second start is already refused), so there is never more than one in-progress run per target. Out of scope: performance runs' throttled progress, the 60 s script timeout cap. Source: https://github.com/brienze1/xenon/issues/35"

## Context

XEN-017 added the collection/folder **Runner** (Setup, Results and Runs views in the target's
tab) and XEN-008 exposed runs to AI clients over MCP (`run` to start/pause/resume/stop, `runs`
to list/get). Reported on 2026-10-27 with Xenon 0.1.9 (issue #35), on a 50-request folder run
started by an AI client:

- The run executed and `runs get` returned it (`running`, 11/50 rows).
- `runs list` for the same folder did not include it — only the two finished runs.
- The folder's Runner tab, open on an earlier UI-started run, kept showing that earlier run. The
  new run appeared neither in **Results** nor in **Runs**.
- Earlier the same day, a UI-started run was also missing from `runs list` while it ran and
  appeared only once it finished.

### Settled from the code rather than asked

| Question | Answer and the fact that settles it |
| --- | --- |
| Why does `runs list` miss a running run while `runs get` finds it? | Listing reads only the stored runs, and a run is stored when it ends. Getting a run looks at the in-progress runs first, then at the stored ones. |
| Why does the open Runner tab keep the old run? | The tab follows only the run it already shows: progress of any other run of the same target is recorded for the running indicator but never shown in the tab. Runs started from the tab's own Start button switch the tab themselves; runs started anywhere else do not. |
| Can two runs of one target be in progress at once? | No. Starting a second run of a collection/folder that already has one in progress is refused. So "the newly started run of the target" is always unambiguous. |
| Is there a moment when a run is both stored and in progress? | Yes, briefly: a finished run is stored first and released from the in-progress runs right after. The list must show it once — the stored copy wins (FR-004). |
| Does an in-progress run have a summary? | Not today: the summary (requests, tests passed/failed/skipped, duration) is computed when the run ends, so an in-progress run would list as "0 requests · 0 passed". The list shows an in-progress run with the summary of what it has done so far (FR-003). |
| Can an in-progress run be deleted from Runs? | No — deleting an in-progress run is already refused. Runs therefore does not offer Delete on an in-progress run (FR-009). |
| What does Clear runs do while a run is in progress? | It deletes the stored runs only; the in-progress run is not stored, so it stays in the list (FR-010). |
| Switch the tab to the new run, or show a "new run started" notice? | Switch. The issue proposes either and recommends switching; it matches what Start in the tab already does (show Results of the new run), and a notice would leave the run unwatchable until clicked. |

## Clarifications

### Session 2026-10-27

No question needed the user; each was settled from the issue or the code (see the table above):

- Q: Follow the new run or raise a notice? → A: Follow it — the tab shows Results of the new run (FR-006).
- Q: Which summary does an in-progress run show in Runs and `runs list`? → A: The summary of the rows recorded so far (FR-003).
- Q: May an in-progress run be deleted from Runs? → A: No; Delete is not offered for it (FR-009), matching the existing refusal.
- Q: Does the tab keep following a run on every progress report? → A: No — once, when the run starts; an older run the user opens afterwards stays shown (FR-006). Following on every report would make Runs unusable during a run.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Runs history and `runs list` include in-progress runs (Priority: P1)

A user or an AI client starts a run of a collection or folder. While it runs (or is paused), the
target's **Runs** view and `runs list` for that target both show it at the top, with its status
(Running or Paused) and what it has done so far, above the finished runs. When it finishes, it
stays in the list exactly once, now with its final status and summary.

**Why this priority**: it is the root of the "invisible run" report for every client, AI and UI
alike, and it makes `list` and `get` agree. An AI client that cannot list a running run cannot
find, monitor or stop it.

**Independent Test**: start a run, call `runs list` (or open Runs) before it finishes, and see
the run listed first with status `running`; let it finish and list again — it is there once,
finished.

**Acceptance Scenarios**:

1. **Given** a folder with two stored runs and a run of it in progress, **When** an AI client
   lists the folder's runs, **Then** three runs are returned, the in-progress run first with
   status `running`, followed by the stored runs newest first.
2. **Given** a run in progress that is paused, **When** the runs of its target are listed,
   **Then** it is listed with status `paused`.
3. **Given** a run in progress that has completed some rows, **When** the runs of its target are
   listed, **Then** its summary counts the requests done and the tests passed, failed and skipped
   so far.
4. **Given** a run that just finished, **When** the runs of its target are listed, **Then** it is
   listed exactly once, with its final status and summary.
5. **Given** a run in progress on a collection, **When** the runs of a different collection or
   folder (or of another workspace) are listed, **Then** that run is not included.
6. **Given** any run that `runs get` returns for a target, **When** that target's runs are
   listed, **Then** the run is in the list.

---

### User Story 2 - The open Runner tab follows a run started elsewhere (Priority: P1)

A user has a folder's tab open with its Runner showing an earlier run (in Results or Runs). An AI
client starts a new run of that folder. The tab switches to **Results** of the new run and shows
it live — progress, rows, tests, console — with Pause/Resume and Stop, exactly as if the user had
pressed Start in the tab.

**Why this priority**: it is the other half of the report — the user must be able to watch and
stop what an AI client is sending on their behalf.

**Independent Test**: open a folder's Runner on an older run, start a run of that folder from an
AI client, and see the tab show the new run's Results, updating live, with Stop available.

**Acceptance Scenarios**:

1. **Given** a folder's tab showing an earlier run in Results, **When** a run of that folder
   started by an AI client reports progress, **Then** the tab shows Results of the new run and
   reads it as it progresses.
2. **Given** a folder's tab on Setup or Runs, **When** a run of that folder started elsewhere
   reports progress, **Then** the tab switches to Results of the new run.
3. **Given** the tab follows a run started elsewhere, **When** the user presses Stop (or Pause,
   then Resume), **Then** that run is stopped (paused, resumed).
4. **Given** a folder's tab and a run of a *different* folder starting, **When** it reports
   progress, **Then** the folder's tab is unchanged.
5. **Given** a run started elsewhere ends, **When** the tab is following it, **Then** the tab
   shows the finished run and the finished-run notice appears, as for a run started in the tab.
6. **Given** the tab follows a run in progress, **When** the user opens an older run from Runs
   and the in-progress run reports more progress, **Then** the tab keeps showing the older run.

---

### User Story 3 - The Runs view keeps up while a run is in progress (Priority: P2)

A user is looking at a target's **Runs** view when a run of that target starts (from the tab or
elsewhere). The new run appears at the top of the list without the user leaving and re-opening
the view, and an in-progress run cannot be deleted from the list.

**Why this priority**: polish on top of Story 1; the run is already reachable through Results
once Story 2 holds.

**Independent Test**: keep Runs open, start a run of the target, see the new run listed as
Running; its menu offers no Delete.

**Acceptance Scenarios**:

1. **Given** the target's Runs list has been read this session, **When** a run of the target not
   in that list reports progress, **Then** the Runs list is read again and shows it.
2. **Given** an in-progress run in the Runs list, **When** the user opens its actions, **Then**
   Delete is not offered.
3. **Given** a run in progress and stored runs, **When** the user clears the target's runs,
   **Then** the stored runs are deleted and the in-progress run stays listed.
4. **Given** the Runs list holds only an in-progress run, **When** the user looks at Clear runs,
   **Then** it is disabled.

### Edge Cases

- **Run finishes while being listed**: the finished run is briefly both stored and in progress;
  the list shows it once, as stored (FR-004).
- **Progress of the run the tab already shows**: unchanged behaviour — the tab keeps reading it.
- **Progress event arrives after the run's finished event** (events can cross): the tab does not
  switch back to a finished run, and the running indicator does not reappear (existing rule).
- **The target's tab was never opened this session**: nothing to switch; when the tab is opened
  while the run is in progress, it already shows the in-progress run (existing behaviour).
- **Another workspace's run**: ignored by the tab and not listed for this workspace.
- **Performance runs**: listed while in progress like functional runs; their throttled progress
  is out of scope and unchanged.
- **User opens an older run while a run is in progress**: the tab shows the older run and stays
  on it; the in-progress run keeps its running indicator and stays first in Runs (FR-006).
- **Clear runs while the tab follows the in-progress run**: the tab keeps showing it (the run
  shown is reset by Clear only when it is a stored run — existing rule).
- **User mid-edit on Setup** when a run starts elsewhere: the tab switches to Results; the
  Setup edits are kept and remembered as before (switching views never discards them).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Listing the runs of a collection or folder (Runner **Runs** view and the AI
  client's `runs list`) MUST include the target's in-progress run (status `running` or `paused`),
  if any, together with its stored runs.
- **FR-002**: The listed runs MUST be ordered by start time, newest first, so an in-progress run
  is first. Runs with the same start time keep the stored order (newest stored first).
- **FR-003**: An in-progress run in a list MUST carry the summary of what it has done so far:
  requests completed, tests passed, failed and skipped, and elapsed time.
- **FR-004**: A run MUST appear at most once in a list. When a run is both stored and still in
  progress (the moment it finishes), the stored copy MUST be listed.
- **FR-005**: Every run that getting a run by id returns for a target MUST also be returned by
  listing that target's runs (list and get agree).
- **FR-006**: When a run of a collection or folder whose tab Runner has been opened this session
  starts — its first progress report reaches the app and the tab does not already show it — the
  tab MUST switch to Results of that run and show it live (progress, rows, tests, console),
  whichever client started it. The tab follows a run once, when it starts: if the user then opens
  an older run from Runs, later progress of the in-progress run MUST NOT pull the tab back.
- **FR-007**: A tab following a run started elsewhere MUST offer the same Pause, Resume and Stop
  controls as for a run started in the tab, and they MUST control that run.
- **FR-008**: When the target's Runs list has already been read this session and a run of the
  target not in that list reports progress, the Runs list MUST be read again.
- **FR-009**: The Runs view MUST NOT offer Delete for an in-progress run.
- **FR-010**: Clearing a target's runs while one is in progress MUST leave the in-progress run
  in the list. Clear runs MUST be disabled when the list holds no stored run.
- **FR-011**: A progress report of a run that has already finished MUST NOT switch the tab or
  bring back the running indicator (existing rule, kept).
- **FR-012**: Runs of other targets and other workspaces MUST NOT be listed for a target and MUST
  NOT change a target's tab. A folder is its own target: a run of a folder is not listed for (and
  does not change the tab of) the collection or folder that contains it, and vice versa.
- **FR-013**: A listed in-progress run MUST expose no more than a listed stored run: id, target,
  type, status, environment name, start/finish time, progress counts with the current request's
  name, and summary counts — never request, response, header or variable values (secrets
  included).

### Key Entities

- **Run**: one execution of a collection or folder — id, target, type (functional/performance),
  status (running, paused, completed, stopped, stopped on error), start and finish time,
  environment, rows recorded, summary. *In progress* while running or paused; *stored* on this
  machine once it ends (25 most recent per target).
- **Run summary**: requests, tests passed/failed/skipped and duration of a run — final once it
  ends, so-far while it is in progress.
- **Runner tab state**: per collection/folder tab — the view shown (Setup, Results, Runs), the
  run shown in Results, and the Runs list read this session.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of runs that can be fetched by id are also returned when their target's runs
  are listed, at any moment of their life (in progress, finishing, finished).
- **SC-002**: A run started by an AI client appears in the open tab of its target within one
  progress report (the first completed row) — no reload, no restart, no click.
- **SC-003**: No run is ever listed twice for a target.
- **SC-004**: The user can stop an AI-started run from the tab in one click once it shows.

## Assumptions

- One in-progress run per target at most (enforced today; unchanged).
- Progress is reported after every completed row for functional runs and throttled for
  performance runs (unchanged).
- The 25-runs-kept limit applies to stored runs only; an in-progress run is listed in addition to
  them while it runs.
- Where the run was started (tab or AI client) is not displayed; the run looks the same either
  way.
- Out of scope: performance runs' throttled progress, the 60 s script timeout cap.
