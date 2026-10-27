# Feature Specification: Runner Result Rows Never Overlap When Expanded

**Feature Branch**: `xen-021`

**Created**: 2026-10-27

**Status**: Draft

**Input**: User description: "XEN-021 — In a collection or folder tab's Runner → Results list, expanding a request row must show its details legibly. Today an expanded row is always drawn at one fixed height, so when a row has several tests and several console lines the tests panel and the console list are squeezed into the same band and drawn on top of each other, and anything that doesn't fit is crushed rather than revealed. Users need an expanded row to show, top to bottom and without overlap, the row's error (if any), its tests panel (filter chips + test lines), its script errors, and its console lines. The expanded row must grow to fit its content, and past a maximum height it scrolls inside its own panel; the rows below move down to make room with no gaps or overlap — also while a run is still in progress and adding rows or updating the expanded row's content, when several rows are expanded at once, after the run finishes, and when a past run is opened from Runs. This is a UI-only defect: the run data, tests and console output are correct. Out of scope: the row detail side panel (RunRowDetail), and changing what the expansion shows. Source: GitHub issue brienze1/xenon#34."

## Context

A layout bug reported in GitHub issue [brienze1/xenon#34](https://github.com/brienze1/xenon/issues/34).
A collection or folder tab has a Runner section whose Results view lists one row per request
sent, grouped by iteration. Each row has a chevron that expands it in place to show the row's
error (if any), its tests (filter chips plus one line per test), its script errors and its
console lines.

What goes wrong today (verified against `main` at `4e25bc5`): every expanded row is given the same
fixed height regardless of what it holds, and the parts inside the expansion are allowed to shrink
below their natural height. When a row has more content than fits — the reported case is 3 tests
and 7 console lines on a 50-request folder run — the tests panel and the console list are squeezed
into the same band and drawn over each other, and the next row starts right after the fixed height.
Run data, tests and console output are correct; only the layout is wrong.

## Clarifications

### Session 2026-10-27 (product owner, during specification)

- Q: How tall may an expanded row's details panel get before it scrolls inside itself? → A: A
  fixed 320 px. Up to that height the panel grows to fit its content; past it, the panel scrolls
  inside itself and the row stays at the cap.

### Session 2026-10-27 (speckit-clarify)

- Q: Can keyboard users scroll a details panel that has reached the cap? → A: Yes. The details
  panel of an expanded row is keyboard-focusable (one Tab stop per expanded row, after the row's
  own controls), has the accessible name "Details of <row name>", shows the app's focus ring,
  and scrolls with the arrow, Page Up/Down, Home and End keys while focused.
- Q: Is every expanded panel focusable, or only one that actually overflows? → A (settled from
  the code: whether a panel overflows is only known after layout, and a stop that appears and
  disappears as content arrives during a live run is worse than a stable one): every expanded
  details panel is focusable.
- Q: Does the 320 px cap include the panel's padding? → A (settled: it is the visible panel
  box): yes — the panel's outer box, padding included, is at most 320 px tall; the row's margin
  below the panel is outside the cap.
- Q: When a row above the visible part of the list changes height, does the view jump? → A
  (settled from the list's existing scrolling behaviour, no new behaviour added): the list keeps
  the items the user is looking at in place; no auto-scroll is added.
- Q: How is the visual result verified across platforms? → A (settled by the XEN-019 precedent
  for UI-only layout fixes): by eye on macOS with before/after screenshots in the pull request,
  plus CI's quality gate on Windows and Linux.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Expanded row shows every part legibly (Priority: P1)

A user runs a folder and expands a row that has several tests and several console lines. They see
the row's error (if any), the tests panel (filter chips and test lines), the script errors and the
console lines stacked top to bottom, each fully readable, nothing drawn on top of anything else.
If the content is taller than the maximum, the details panel scrolls inside itself so every line
can still be reached.

**Why this priority**: This is the reported defect; without it the expansion is unreadable.

**Independent Test**: Expand a row with 3 tests and 7 console lines; every test line and every
console line is readable and none shares vertical space with another. Expand a row with 40 console
lines; the details panel stops at the maximum height and scrolls to the last line.

**Acceptance Scenarios**:

1. **Given** a finished run whose row has 3 passing tests and 7 console lines, **When** the user
   expands the row, **Then** the filter chips, the 3 test lines and the 7 console lines appear in
   that order, top to bottom, with no line overlapping another.
2. **Given** a row with an error, failed tests, script errors and console lines, **When** the user
   expands it, **Then** the parts appear in the order error → tests → script errors → console,
   without overlap.
3. **Given** a row whose expansion content is taller than 320 px, **When** the user expands it,
   **Then** the details panel is 320 px tall and scrolls inside itself; scrolling it reaches the
   last console line.
4. **Given** a row whose expansion content is shorter than 320 px (e.g. "No tests" and one console
   line), **When** the user expands it, **Then** the row is only as tall as its content, with no
   empty band below it.

---

### User Story 2 - Rows below make room (Priority: P1)

When a row is expanded, collapsed, or its content changes, the rows below it move to sit directly
after it — no gap, no overlap — including when several rows are expanded at once.

**Why this priority**: Overlap with the next row is the second half of the reported defect.

**Independent Test**: Expand rows 1 and 3 of a 10-row run; row 2 starts right below row 1's
expansion and row 4 right below row 3's; collapse row 1 and everything below moves up with no gap.

**Acceptance Scenarios**:

1. **Given** a run with several rows, **When** the user expands one row, **Then** the next row (or
   iteration heading) starts immediately below the expanded row's details, with no gap and no
   overlap.
2. **Given** two or more expanded rows, **When** the user collapses one, **Then** the rows below it
   move up to close the space, and the other expanded rows keep their full height.
3. **Given** an expanded row, **When** the user changes the results filter (All / Passed / Failed /
   Skipped) so the list changes, **Then** every visible row and heading is placed directly after the
   previous one with no gap or overlap.

---

### User Story 3 - Same behaviour on live and past runs (Priority: P2)

The layout holds while a run is still in progress — new rows arriving, an expanded row's content
being filled in — after the run finishes, and when a past run is opened from Runs.

**Why this priority**: The defect was observed during a live run; the same list shows past runs.

**Independent Test**: Start a 50-request folder run, expand the first row while later rows keep
arriving; the expansion stays legible and the arriving rows stack below it. Open the same run from
Runs afterwards and expand the same row; it looks the same.

**Acceptance Scenarios**:

1. **Given** a run in progress with an expanded row, **When** new rows arrive, **Then** they appear
   below without overlapping the expanded row.
2. **Given** a run in progress, **When** an already expanded row's content changes (more tests or
   console lines shown after a refresh of the results), **Then** the row's height follows its new
   content (up to the cap) and the rows below move accordingly.
3. **Given** a finished run opened from Runs, **When** the user expands a row, **Then** it behaves
   exactly as in User Story 1 and 2.

---

### Edge Cases

- A row with no tests shows the compact "No tests" empty state; it must not stretch to fill the
  maximum height nor be squeezed.
- A test with a long multi-line error message wraps inside the tests panel; the panel grows (up to
  the cap, then scrolls) instead of drawing over the console.
- Long console summaries wrap within the panel's width as today; a single unbreakable token wider
  than the panel scrolls sideways inside the panel; nothing spills into the next row.
- The tests panel's own filter (All / Passed / Failed / Skipped) changes the number of test lines;
  the row's height follows the change.
- The side detail panel (opened by clicking a row's name) narrows the list; expanded rows re-wrap to
  the narrower width and the list re-flows without overlap.
- Expanded rows scrolled out of view and back keep their correct height and position.
- An iteration heading and a collapsed row keep their current heights (28 px and 32 px).
- A performance run has no row list; nothing changes there.
- Rows the user expanded stay expanded when the run's results are refreshed during a live run and
  when the results filter hides and shows them again (as today); a refresh never collapses them.
- When the mouse wheel or trackpad scrolls a capped details panel to its top or bottom, further
  scrolling moves the Results list itself (the platform's normal hand-off between nested scroll
  areas); the panel never traps the wheel.
- A focused details panel that the user scrolls far out of the rendered part of the list loses
  focus the same way the row's chevron and name buttons do today; no focus restoration is added.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: An expanded row MUST show, in this order and without any part drawn over another: the
  row's error (if any), the tests panel, the script errors (if any), and the console lines (if any).
- **FR-002**: The parts of an expanded row's details MUST NOT shrink below their natural height;
  when the details are taller than the available height, the details panel scrolls instead.
- **FR-003**: An expanded row's details panel MUST grow to fit its content up to a maximum of
  320 px, and past that MUST scroll vertically inside itself so all of its content stays reachable.
- **FR-004**: The list MUST place every row and iteration heading directly after the previous one,
  using each expanded row's actual rendered height, so there is no gap and no overlap between items.
- **FR-005**: When a row is expanded or collapsed, when its content changes, or when the list's
  width changes, the list MUST re-place the items below it accordingly.
- **FR-006**: FR-001–FR-005 MUST hold while a run is in progress (rows arriving, results being
  refreshed), after it finishes, when several rows are expanded at once, under every results
  filter, and when a past run is opened from Runs.
- **FR-007**: Collapsed rows (32 px) and iteration headings (28 px) MUST look as they do today; what
  the expansion shows, the row detail side panel, and the performance results MUST NOT change.
- **FR-008**: The tests panel embedded in a runner row MUST keep its behaviour in the request's
  response viewer (Tests tab) unchanged.
- **FR-009**: An expanded row's details panel MUST be keyboard-focusable, MUST carry the
  accessible name "Details of <row name>" (the same name the row's chevron uses), MUST show the
  app's focus ring when focused, and MUST scroll with the arrow, Page Up/Down, Home and End keys
  while focused.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For the reported case (3 tests, 7 console lines), 0 lines of the expansion overlap
  another line, and 0 lines overlap the next row.
- **SC-002**: With 3 rows expanded at once in a 50-row run, every row and heading starts within 1 px
  of where the previous one ends (no gap, no overlap).
- **SC-003**: A row with 40 console lines shows a details panel no taller than 320 px, and the last
  console line can be reached by scrolling that panel.
- **SC-004**: Collapsed lists look identical before and after the change (same row and heading
  heights).
- **SC-005**: With only the keyboard, a user can move focus to a capped details panel and scroll
  it to its last line.
- **SC-006**: The project's full quality gate (lint, typecheck, unit tests, backend suites) stays
  green.

## Assumptions

- The fix is frontend-only; no backend, run data, or stored-format change is involved.
- The 320 px cap applies to the details panel's outer box, padding included (the row's own 32 px
  line sits above it and the panel's bottom margin sits below it); it is a fixed value, not
  scaled to the window, per the product owner's choice. Like every size in the UI it is in UI
  pixels, so View → Zoom scales it together with the rest of the interface.
- Verification of the visual result is by eye in the running app on macOS, with screenshots in the
  pull request, plus CI's quality gate for Windows and Linux (same policy as XEN-019).
- When rows above the visible area change height, the list keeps what the user is looking at in
  place; no new auto-scroll behaviour is added.
