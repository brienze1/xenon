# Feature Specification: Rename by Clicking the Name, Saved With the Save Hotkey

**Feature Branch**: `xen-012`

**Created**: 2026-10-15

**Status**: Draft

**Input**: User description: "One consistent way to rename whatever a main-area tab shows. Today the environment tab repeats its name in an input with a Rename button, the collection/folder Overview and the request Info panel have an always-visible name input that saves on Enter or blur, and the request/example breadcrumb lets you click the name but saves straight away. None of them use the save hotkey or the tab's unsaved-changes dot. Expected: the name is shown as text; clicking it (or focusing it and pressing Enter) turns it into an input in place with the name selected; editing marks the tab dirty; Cmd/Ctrl+S saves the name with the tab's other pending changes; Escape cancels; leaving the field keeps the draft and the dirty dot; the close prompt applies; validation errors appear inline when saving with today's messages; the tab label, sidebar tree and breadcrumb show the new name only after saving. Backend rename operations do not change. Source: https://github.com/brienze1/xenon/issues/9"

## Context

Every edit inside a main-area tab (a request's URL, a collection's auth, an environment's
variables, …) follows one rule: it is a draft, the tab shows the yellow unsaved dot, Cmd/Ctrl+S
saves it, and closing the tab asks "discard changes?". Names are the exception, and each tab
renames differently:

| Tab | Where the name is edited today | When it is written |
| --- | --- | --- |
| Environment | Heading shows the name; a separate input + **Rename** button below it | On the button |
| Collection / folder | Heading shows the name; Overview has an always-visible **Name** input | On Enter or blur |
| Request | Breadcrumb's last crumb is clickable; Info panel has an always-visible **Name** input | Breadcrumb: on Enter; Info: on Enter or blur |
| Example | Breadcrumb's last crumb is clickable | On Enter |
| Globals, Settings | No editable name | — |

This feature replaces all of them with one pattern and brings the name into the tab's normal
draft/save/discard cycle. The operations that write a name to disk (renaming a tree item, an
example or an environment) stay exactly as they are; only when and from where they are called
changes.

## Clarifications

### Session 2026-10-15

- Q: Do the sidebar tree's inline rename (context menu / F2) and the workspace rename dialog
  change too? → A: No. They stay as they are and still save straight away. Only names shown
  inside main-area tabs change.
- Q: If the name can't be saved but the tab's other edits are valid, what does Cmd/Ctrl+S do?
  → A: It stops the whole save. The user's rule of thumb: if any field is not allowed, nothing is
  saved. Also, the name input must not accept more characters than allowed, so "too long" can't
  be typed at all.
- Q: A tab has a pending name draft and the item is renamed elsewhere (sidebar, git sync,
  another tab). What happens? → A: The tab keeps its draft and dirty dot. Views of the saved name
  update. Saving applies the draft. If the draft now equals the new saved name, it is no longer
  pending.
- Q: What does Enter do while typing in the name input? → A: It closes the input and keeps the
  draft pending (dirty dot). It does not save; only Cmd/Ctrl+S or the close prompt's Save do.
- Settled from the code rather than asked: every name has the same limit, 200 characters
  (`maxItemNameLength` in `internal/domain/entity/item_name.go`, used for tree items, examples
  and environments). An environment name clashes when another environment has it ignoring case
  and surrounding spaces (`Environment.HasName`). Tree item and example names can't clash (file
  names are de-duplicated on disk) and have no character rules.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Rename a request in its tab and save with the hotkey (Priority: P1)

Luis has a saved request open. He clicks its name in the breadcrumb above the URL bar; the name
turns into an input with the whole name selected. He types a new name. The tab shows the yellow
unsaved dot, but the tab label and the sidebar still show the old name. He presses Cmd+S: the
name and any other edits he had made to the request are saved together, the dot disappears and
the tab label, sidebar tree and breadcrumb all show the new name.

**Why this priority**: Requests are renamed far more often than anything else, and the request
tab is where the inconsistency (instant save in one place, blur-save in another) bites hardest.

**Independent Test**: Open a saved request, click its breadcrumb name, type a new name, check the
dirty dot and the unchanged tab label/sidebar, press Cmd/Ctrl+S, and check the request was
renamed on disk and everywhere in the UI.

**Acceptance Scenarios**:

1. **Given** a saved request open in a tab, **When** Luis clicks the name in the breadcrumb,
   **Then** the name becomes an input in place with the full current name selected.
2. **Given** the name is being edited, **When** Luis changes it, **Then** the tab shows the
   unsaved dot and the tab label, sidebar tree and breadcrumb ancestors still show the saved name.
3. **Given** a changed name and a changed URL in the same tab, **When** Luis presses Cmd/Ctrl+S,
   **Then** both are saved, the dot disappears, and the tab label, sidebar tree and breadcrumb show
   the new name.
4. **Given** the name is being edited, **When** Luis presses Escape, **Then** the input closes,
   the saved name is shown again, and the tab is dirty only if other edits are pending.
5. **Given** the name is being edited, **When** Luis clicks elsewhere in the tab (the field loses
   focus), **Then** the input closes showing the draft name, and the draft and the unsaved dot
   remain.
6. **Given** the Info side panel is open for the same request, **When** the name is edited in
   either the breadcrumb or the Info panel, **Then** both show the same draft name.

---

### User Story 2 - Rename a collection or folder in its tab (Priority: P1)

Luis opens a collection tab. Its name is the heading at the top of the tab; the Overview no
longer has a separate Name input. He clicks the heading, types a new name, edits the
collection's auth too, and presses Cmd+S. Both are saved; the sidebar shows the new name.

**Why this priority**: Collections and folders are the second most renamed items, and their
Overview currently writes the name on blur, outside the tab's save.

**Independent Test**: Open a collection (or folder) tab, click the heading, rename, check the
dirty dot, save with Cmd/Ctrl+S, and check the sidebar and tab label.

**Acceptance Scenarios**:

1. **Given** a collection or folder tab, **When** Luis looks at it, **Then** the name appears only
   as the heading text — there is no Name input in the Overview and no Rename button.
2. **Given** the heading, **When** Luis focuses it with the keyboard and presses Enter, **Then**
   it becomes an input with the name selected.
3. **Given** a renamed but unsaved collection, **When** Luis presses Cmd/Ctrl+S, **Then** the name
   and every other pending change of the tab (settings, variables, documentation) are saved.

---

### User Story 3 - Rename an environment in its tab (Priority: P2)

Luis opens an environment tab. Its name is the heading next to the colour swatch; there is no
input or Rename button below it. He clicks the heading and renames the environment; the tab shows
the unsaved dot together with any unsaved variable edits; Cmd+S saves both. If another
environment already has that name, the save fails and the same "already exists" message as today
appears under the name.

**Why this priority**: Environments are renamed less often, but their tab is the most visibly
inconsistent one (a heading and an input showing the same name).

**Independent Test**: Open an environment tab, rename via the heading, save with Cmd/Ctrl+S, then
try a name that another environment already uses and check the inline error.

**Acceptance Scenarios**:

1. **Given** an environment tab, **When** Luis looks at it, **Then** the name is shown once, as
   the heading, with no separate input or Rename button.
2. **Given** a draft name equal to another environment's name, **When** Luis saves, **Then** the
   message "An environment named “<name>” already exists. Choose another name." appears inline
   under the name, the draft and the unsaved dot stay, and the environment keeps its saved name.

---

### User Story 4 - Rename a saved example in its tab (Priority: P3)

Luis opens a saved example. He clicks its name in the breadcrumb, renames it, and saves it with
Cmd+S together with any edits to its response.

**Why this priority**: Examples are renamed rarely, but they share the breadcrumb with requests,
so leaving them on the old behaviour would keep two rules in one component.

**Independent Test**: Open an example tab, rename via the breadcrumb, save with Cmd/Ctrl+S, check
the sidebar shows the new example name.

**Acceptance Scenarios**:

1. **Given** an example tab, **When** Luis renames it in the breadcrumb and presses Cmd/Ctrl+S,
   **Then** the example's new name and its other pending edits are saved, and the sidebar shows
   the new name.

---

### User Story 5 - Close a tab with an unsaved name (Priority: P2)

Luis renames a request but does not save. He closes the tab: the usual "discard changes?" prompt
appears. **Save** renames and saves the request, then closes the tab; **Discard** closes it and
the request keeps its old name; **Cancel** leaves the tab open with the draft name.

**Why this priority**: Without this, an unsaved name would be lost silently — the exact problem
the dirty dot exists to prevent.

**Independent Test**: For each tab type with a name, rename without saving, close the tab and
check each of the three choices.

**Acceptance Scenarios**:

1. **Given** a tab whose only pending change is its name, **When** Luis closes it, **Then** the
   close prompt appears.
2. **Given** the close prompt, **When** Luis chooses Discard, **Then** the tab closes and the item
   keeps its saved name on disk and in the sidebar.
3. **Given** the close prompt, **When** Luis chooses Save and the name is invalid, **Then** the
   tab stays open with the prompt's error, exactly as for any other failed save.

---

### Edge Cases

- **Empty or whitespace-only name**: saving shows the inline "empty name" message used today for
  that tab type; nothing is saved for the name.
- **Name longer than 200 characters**: cannot be typed or pasted — the input stops at 200
  characters (pasted text is cut at the limit).
- **Name unchanged after trimming** (e.g. only trailing spaces added, or typed back to the saved
  name): the name is not a pending change — no dirty dot from the name, and saving does not call
  rename.
- **Never-saved request**: there is no item on disk to rename. Its name is shown as text and is
  click-to-edit like any other; the draft name is what the Save dialog proposes when the request
  is first saved (as today). The tab label of a never-saved request shows its draft name, since
  it has no saved name.
- **Same request open in two tabs**: each tab has its own draft. Saving the name in one tab
  changes the saved name. The other tab takes the new name if it had no pending name of its own,
  and keeps its own pending name if it had one (FR-014).
- **Item renamed elsewhere while a name draft is pending** (sidebar F2, git sync, another tab):
  the tab keeps its draft and dirty dot; the tab label, sidebar and other views show the new saved
  name; saving applies the draft; a draft equal (after trimming) to the new saved name stops being
  pending.
- **The name is invalid but other edits are valid**: the whole save stops — nothing is written,
  the error shows inline under the name, and every pending change stays pending.
- **Another part of the save fails** (e.g. a variable or example field the backend rejects, or a
  conflict): the name is not written either; it stays pending.
- **Typing past the limit**: the name input does not accept more than 200 characters.
- **Name draft lifetime**: a name draft lives exactly as long as the tab's other drafts of the
  same tab type. It is kept across tab switches and app restarts whenever the tab's other edits
  are, and it is dropped whenever they are (discard, close).
- **Duplicate Tab** on a request with a pending name: the new tab gets a copy of the whole draft,
  name included, as Duplicate Tab does today for every other field.
- **Save As** (Cmd/Ctrl+Shift+S) on a request with a pending name: the Save As dialog proposes
  "<draft name> Copy" (today's rule applied to the draft name). The original request is not
  renamed, and its tab keeps its pending name draft.
- **Collection save that waits for documentation confirmation**: the name is written only after
  the documentation change is confirmed and saved. If the confirmation is cancelled, the name
  stays pending (FR-011).
- **Editing while a save is running**: the name stays editable. The running save writes the name
  it read when it started. Afterwards the tab compares whatever the draft holds then against the
  new saved name, so a later edit stays pending (dot shown). A second Cmd/Ctrl+S during a running
  save is ignored, as it is today for request and example saves (`saving` guard).
- **Globals and Settings tabs**: have no editable name; their titles stay fixed text.
- **Rename outside tabs**: the sidebar tree's inline rename (context menu / F2) and the workspace
  rename dialog are unchanged and keep saving immediately. A sidebar rename of an item whose tab
  has a pending name draft follows the "renamed elsewhere" rule above.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: In request, example, collection, folder and environment tabs, the item's name MUST
  be shown as text with no always-visible name input and no Rename button.
- **FR-002**: The name text MUST be focusable and MUST turn into an input in place, with the full
  current (draft) name selected, when clicked or when focused and Enter is pressed. It MUST look
  editable on hover and keyboard focus, the way the breadcrumb's click-to-rename crumb does today
  (pointer cursor, hover background, "Click to rename" tooltip, visible focus ring). A name too
  long for its space MUST be cut with an ellipsis on one line, with the full name in the tooltip.
  The input takes the same space and typography as the text it replaces.
- **FR-003**: Each of these tabs MUST have exactly one name draft. It is edited from the
  breadcrumb's last crumb for request and example tabs, and from the heading for collection,
  folder and environment tabs. Request tabs also show the same draft in the Info side panel, with
  the same click-to-edit behaviour. The collection/folder Overview MUST NOT have a Name field.
- **FR-004**: A name that differs from the saved name after trimming MUST be a pending change of
  the tab: it marks the tab dirty (the unsaved dot) and counts for the close prompt and for every
  bulk-close action, like any other unsaved edit.
- **FR-005**: Cmd/Ctrl+S (and the close prompt's Save) MUST save a pending name together with the
  tab's other pending changes, by calling the existing rename operation for that item type
  (tree item, example or environment).
- **FR-006**: Escape while editing MUST close the input, drop the name draft and show the saved
  name again.
- **FR-007**: Leaving the input any other way (blur, Tab, clicking elsewhere) MUST close it and
  keep the draft name displayed and pending. After Enter or Escape, keyboard focus MUST return to
  the name text. After a blur, focus stays wherever the user moved it. Enter while editing MUST close the input and keep the
  draft (it does not save).
- **FR-008**: Discarding a tab's changes (close prompt → Discard) MUST drop the name draft; the
  item keeps its saved name.
- **FR-009**: The tab label, the sidebar tree, the breadcrumb ancestors and any other view of the
  item outside its own name editor MUST show the saved name until the save succeeds.
- **FR-010**: Save-time errors for the name (empty, environment name already taken, or any error
  the rename operation returns) MUST be shown inline next to the name, with the same messages as
  today, and the draft MUST stay pending. The empty-name and the environment-name-taken checks MUST
  run before anything of the tab is written. The error MUST stay visible until the name is edited
  again, Escape drops the draft, the tab's changes are discarded or saved, or the tab is closed.
- **FR-011**: A save MUST be all-or-nothing with respect to the name: an invalid name MUST stop
  the whole save before any of the tab's changes are written, and the name MUST NOT be written
  unless every other part of the tab's save succeeded. Any failure leaves all of the tab's
  unwritten changes pending and the tab dirty.
- **FR-012**: The name input MUST NOT accept more than 200 characters (the limit every item type
  shares today).
- **FR-013**: The sidebar tree's inline rename (context menu / F2) and the workspace rename dialog
  MUST stay unchanged and keep saving immediately.
- **FR-014**: When the saved name of an item changes outside the tab (sidebar rename, git sync,
  another tab's save) while the tab has a pending name draft, the tab MUST keep the draft and its
  dirty dot, and MUST stop treating the draft as pending if it now equals the saved name after
  trimming.
- **FR-015**: The backend rename operations, their validation and their error codes MUST NOT
  change; existing BDD rename scenarios MUST keep passing.
- **FR-016**: The editable name MUST be one shared UI component used by every tab type above, so
  the behaviour cannot drift between tabs.
- **FR-017**: The name editor MUST be keyboard accessible: it has an accessible label naming what
  is renamed (e.g. "Rename request"), the input is labelled, and a save error is announced and
  linked to the input.

### Key Entities

- **Name draft**: the unsaved name of the item a tab shows. Belongs to the tab's existing draft
  state for that tab type (request draft, example draft, item-settings draft, environment draft).
  Pending when its trimmed value differs from the saved name.
- **Saved name**: the item's name on disk, as the sidebar tree, the environment list and the tab
  label show it.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All five tab types with a name (request, example, collection, folder, environment)
  rename through the same interaction; zero Rename buttons and zero always-visible name inputs
  remain in tabs.
- **SC-002**: In 100% of cases an unsaved name shows the unsaved dot, and closing that tab asks
  before discarding it.
- **SC-003**: A rename plus other edits in the same tab is saved with a single save keystroke.
- **SC-004**: Frontend tests cover, for each of the five tab types: click to edit, Enter to edit,
  dirty dot, save with the hotkey, Escape to cancel, validation errors and discard on close; the
  existing BDD suite stays green and `make validate` passes.

## Assumptions

- "Globals where a name applies" from the issue resolves to none: the Globals tab has no name of
  its own (its title is the fixed word "Globals"), so it is unchanged.
- The validation messages are the ones shown today where one exists: "Type a name for the
  environment." for an empty environment name, the "already exists" message for a duplicate
  environment name, and the backend's description for any tree item or example rename error. An
  empty request, example, collection or folder name is silently ignored today, so it has no
  message to keep. It gets "Type a name for the <request|example|collection|folder>.", following
  the environment wording. The next-step hint after a backend error reads "Choose another name,
  then save." instead of today's "…then press Enter.", since Enter no longer saves. Tree items and
  examples have no "name taken" rule (file names are de-duplicated on disk) and no
  invalid-character rule (names are free text; file names are derived safely), so the issue's
  "name taken" and "invalid characters" apply only where the backend enforces them today.
- No backend, file format or BDD step changes are needed; the work is in the frontend.
