# Feature Specification: The Eye Opens the Environment Tab

**Feature Branch**: `xen-015`

**Created**: 2026-10-15

**Status**: Draft

**Input**: User description: "The eye button next to the environment switcher (top right of the app) opens the active environment's main-area tab instead of the Quick Look modal. Today the eye opens a Quick Look dialog with two editable variable tables (active environment + globals); with a real environment (~30 variables, several long URLs) the table overflows the dialog, long values wrap, the dialog grows past the window with its own scrollbar, and the full environment tab already does the same job better. Clicking the eye opens the active environment's tab (pinned), or focuses it if already open; with no active environment it opens the Globals tab; the label becomes 'Open environment' / 'Open globals'; the Quick Look modal and code only it used are removed; the PRD and XEN-002 spec stop describing a quick look. Frontend-only; Vitest covers the eye; make validate stays green. Source: https://github.com/brienze1/xenon/issues/18"

## Context

The top bar's environment switcher has an eye button beside it. Today it opens **Quick Look**, a
dialog with two editable variable tables: the active environment's and the globals'. With a real
environment (the HML one: about 30 variables, several long URLs) long values wrap onto several
lines, the dialog grows past the window and gets its own scrollbar, and it is hard to read and
edit.

Since the Quick Look was built, every environment and the globals got a full **main-area tab**
(opened from the sidebar's Environments view or the switcher's "Globals" item) with the name,
colour, "Confirm before send" and the same variables table, laid out for the whole main area and
saved with the tab's normal draft / Cmd+S / close-prompt cycle. The dialog duplicates that tab,
badly. This feature points the eye at the tab and removes the dialog.

## Clarifications

### Session 2026-10-15

Decided by the user before the spec was written (the issue's open questions):

- Q: What does the eye do when no environment is active? → A: It opens the **Globals** tab.
- Q: With an active environment, should the eye also open Globals? → A: No — only the
  environment's tab. Globals stay one click away through the switcher's "Globals" item and the
  sidebar's Environments view.
- Q: Preview or pinned tab? → A: **Pinned**, like the switcher's "Globals" item.
- Q: Label / tooltip? → A: "Open environment" when an environment is active, "Open globals" when
  none is.

Settled from the code rather than asked:

- The switcher (and so the eye) is only shown while a workspace is open, so the eye never has to
  handle "no workspace" (`TopBar end={hasActive ? … : SETTINGS_BUTTON}` in `App.tsx`).
- Opening a target whose tab already exists focuses that tab and, when asked for pinned, pins it;
  it never opens a second tab for the same environment or for the globals (tab store
  `openEnvironmentTab`).
- An environment whose file is broken cannot be made active from the switcher (its menu entry is
  disabled), and if the active one breaks later its tab already shows the error with a next step,
  so the eye needs no special case for it.
- The discard-changes guard and its "Discard unsaved variable changes?" confirmation, and the
  variable table's "drop edits when it closes" draft lifetime, are used only by the Quick Look.
  The environment, globals and collection-variables tabs all keep their edits until saved or
  discarded. They go with the dialog.

Decided by the user during planning:

- Q: The variables table also has its own "Discard Changes / Save Variables" footer that only
  the Quick Look shows (every tab hides it and saves with Cmd+S and the close prompt). Remove it
  too? → A: Yes — remove both the "drop edits on close" lifetime and the footer; every variables
  table is saved and discarded by its tab.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The eye opens the active environment's tab (Priority: P1)

Luis has the HML environment active. He clicks the eye next to the switcher. No dialog appears:
the HML environment's tab opens in the main area as a pinned tab and is focused, showing its name,
colour, "Confirm before send" and all of its variables in the full-width table. If the HML tab was
already open (as a preview or pinned), the eye focuses that tab instead of opening a second one,
and a preview tab becomes pinned.

**Why this priority**: It is the whole point of the issue — the eye must lead to a readable,
editable view of the active environment's variables.

**Independent Test**: With an environment active, click the eye; check that exactly one tab for
that environment exists, it is pinned and focused, and no dialog is open. Repeat with the tab
already open as a preview and as a pinned tab.

**Acceptance Scenarios**:

1. **Given** environment "HML" is active and no tab shows it, **When** Luis clicks the eye,
   **Then** a pinned "HML" environment tab opens and is focused, and no dialog appears.
2. **Given** "HML" is active and its tab is already open but another tab is focused, **When** he
   clicks the eye, **Then** the existing "HML" tab is focused and no new tab is added.
3. **Given** "HML" is active and its tab is open as the preview tab, **When** he clicks the eye,
   **Then** that tab becomes pinned and focused, and no new tab is added.
4. **Given** "HML" is active and its tab has unsaved variable edits, **When** he clicks the eye,
   **Then** the tab is focused with the edits and the unsaved dot still there.
5. **Given** "HML" is active, **When** he hovers or focuses the eye, **Then** its label reads
   "Open environment".

---

### User Story 2 - With no environment active, the eye opens the globals (Priority: P2)

Luis has "No environment" selected. He clicks the eye and the Globals tab opens pinned and focused
(or the existing Globals tab is focused and pinned). The eye's label reads "Open globals", so he
knows what it will open before clicking.

**Why this priority**: Without it the eye would be dead or confusing whenever no environment is
active; the globals are the only variables that still apply then.

**Independent Test**: Select "No environment", click the eye, and check that exactly one Globals
tab exists, pinned and focused, with no dialog.

**Acceptance Scenarios**:

1. **Given** no environment is active and no Globals tab is open, **When** Luis clicks the eye,
   **Then** a pinned Globals tab opens and is focused, and no dialog appears.
2. **Given** no environment is active and the Globals tab is already open, **When** he clicks the
   eye, **Then** that tab is focused (pinned if it was a preview) and no new tab is added.
3. **Given** no environment is active, **Then** the eye's label reads "Open globals"; **When** he
   then picks "HML" in the switcher, **Then** the label reads "Open environment".

---

### User Story 3 - The Quick Look is gone everywhere (Priority: P3)

Nothing in the app opens the Quick Look dialog any more, and the product documents no longer
promise one: the PRD and the environments spec describe the eye as opening the environment's
tab.

**Why this priority**: Leaving the dialog's code or the documents behind would keep a second,
divergent way to edit variables and mislead the next feature that reads the PRD.

**Independent Test**: Search the frontend and the product documents for the Quick Look; the only
mentions left are the history in this spec and in the issue.

**Acceptance Scenarios**:

1. **Given** any environment state, **When** Luis clicks the eye, **Then** no dialog titled
   "Quick Look" appears.
2. **Given** the PRD and the XEN-002 environments spec, **When** they are read, **Then** the eye is
   described as opening the active environment's tab (or the Globals tab with no active
   environment), not as a quick look.

### Edge Cases

- **Active environment changes**: the eye always follows the environment active at the moment of
  the click; switching environments does not open, close or retarget any tab by itself.
- **Active environment deleted**: deleting the active environment makes "No environment" active,
  so the eye then opens the Globals tab (and the deleted environment's tab is closed by the
  existing tab clean-up).
- **Active environment's file is broken**: the eye opens its tab, which shows the existing error
  banner with its next step; nothing new is needed.
- **Unsaved edits in an open environment or Globals tab**: the eye only focuses (and may pin) the
  tab; it never discards, saves or prompts.
- **Many tabs open / tab strip scrolled**: focusing an existing tab uses the same focus behaviour
  as every other way of opening a tab.
- **Keyboard users**: the eye is reachable by Tab and activated by Enter or Space like any icon
  button, and its accessible name is the label above.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: With an environment active, activating the eye MUST open that environment's tab in
  the main area as a pinned tab and focus it — the same tab the sidebar's Environments view opens.
- **FR-002**: If the active environment's tab is already open, activating the eye MUST focus it
  and pin it if it was a preview, and MUST NOT open a second tab for it.
- **FR-003**: With no environment active, activating the eye MUST open the Globals tab pinned and
  focused, or focus (and pin) the existing Globals tab, without opening a second one.
- **FR-004**: With an environment active, activating the eye MUST NOT open or focus the Globals
  tab.
- **FR-005**: Activating the eye MUST NOT open any dialog.
- **FR-006**: The eye's accessible name and tooltip MUST be "Open environment" while an
  environment is active and "Open globals" while none is, updating as soon as the active
  environment changes.
- **FR-007**: Activating the eye MUST NOT change, save or discard any variable edits; unsaved edits
  in an already-open tab stay with that tab.
- **FR-008**: The Quick Look dialog MUST be removed, together with everything only it used: its
  visuals, its tests, the "discard unsaved variable changes?" guard and confirmation, the
  variable table's mode that drops edits when the table closes, and the variable table's own
  "Discard Changes / Save Variables" footer. Every remaining variables table keeps its edits
  until its tab saves or discards them, exactly as today.
- **FR-009**: The eye MUST keep its current look and position next to the switcher.
- **FR-010**: The PRD (environment switcher section and the environments decision) and the XEN-002
  environments spec MUST describe the eye as opening the active environment's tab, or the Globals
  tab when no environment is active, and MUST NOT describe a quick look (the PRD may note that
  XEN-015 replaced it). Only XEN-002's `spec.md` is edited: its `plan.md`, `tasks.md` and
  `checklists/` record how XEN-002 was built and stay as they are.
- **FR-011**: Automated frontend tests MUST cover FR-001 to FR-007, and the project's full
  validation MUST pass.

### Key Entities

- **Active environment**: the environment chosen in the switcher for the open workspace, or none
  ("No environment").
- **Environment tab target**: what a variables tab shows — one environment, or the globals. At
  most one tab exists per target.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: One click on the eye shows all of the active environment's variables in the main
  area — including an environment of about 30 variables with long URLs — with no dialog and no
  window-height overflow of its own.
- **SC-002**: Clicking the eye any number of times in a row never produces more than one tab for
  the same environment or for the globals.
- **SC-003**: Zero code paths open the Quick Look dialog; searching the app's source for it finds
  nothing.
- **SC-004**: 100% of the eye's behaviours in FR-001 to FR-007 are covered by automated tests, and
  the full project validation passes.

## Assumptions

- The environment and Globals tabs, their layout, saving and unsaved-changes handling stay exactly
  as they are; this feature only changes what the eye opens.
- The eye's glyph is reused unchanged; only its label and action change.
- The labels follow the existing sentence-case style of icon-button labels ("Quick look at
  variables" today).
- The backend, the stored files and the BDD suite are not touched; nothing on disk changes.
- Other ways of opening these tabs (sidebar single/double click, switcher "Globals", "Manage
  Environments…", reopen closed tab) keep their behaviour.
