# Feature Specification: Reorder Environments With Drag and Drop

**Feature Branch**: `xen-011`

**Created**: 2026-10-15

**Status**: Draft

**Input**: User description: "XEN-011 — Reorder environments with drag and drop (GitHub issue #8). Today the sidebar Environments list is always sorted alphabetically (case-insensitive), so users cannot keep the environments they use most at the top or group related ones together. Users need to drag and drop environments to reorder them in the sidebar Environments list, as they already reorder collections in the tree. Globals stays pinned at the top. The order persists in the workspace files so it syncs through git. Keyboard users can reorder too. The switcher lists environments in the same order. New, imported or duplicated environments go at the end; unlisted ones go after the ordered ones alphabetically; stale ids are ignored. Search still filters. AI clients can reorder through the MCP env_edit tool."

## Context

The sidebar's Environments view (PRD decisions log: "Where are environments managed?") lists the
globals and every environment of the workspace. Today that list, and the top-right environment
switcher, are always alphabetical (case-insensitive, by name). A user with `DEV`, `Dev (imported)`,
`HML`, `PROD` and a handful of personal environments cannot put the two or three they switch
between all day at the top, or keep an imported copy next to the one it came from.

Collections, folders and requests can already be reordered by drag and drop, and that order lives
in the workspace files so it survives a restart and travels through git. This feature gives the
environment list the same treatment.

## Clarifications

### Session 2026-10-15

- Q: While the search box filters the list, what does reordering do? → A: Disabled — rows are not
  draggable and keyboard pick-up does nothing until the search is cleared.
- Q: In a workspace never reordered, where does a created, duplicated or imported environment go?
  → A: At the end, as collections do: that first add saves the current arranged list with the new
  id appended, so new environments always go last from then on.
- Q: What is the field called in `workspace.yaml`? → A: A nested `environments:` mapping holding
  `order:` (the list of environment ids), i.e. `environments.order`, next to the collections'
  top-level `order`.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Drag an environment to a new place (Priority: P1)

The user opens the Environments view, drags `PROD` from the bottom of the list to just under
`Globals`, and drops it. `PROD` now shows first among the environments, in the sidebar and in the
environment switcher. After quitting and reopening Xenon the order is the same.

**Why this priority**: This is the feature. Everything else refines it.

**Independent Test**: In a workspace with three environments, drag the last one to the top, restart
the app, and check the sidebar and the switcher both show the new order.

**Acceptance Scenarios**:

1. **Given** environments listed `Alpha`, `Beta`, `Gamma`, **When** the user drags `Gamma` above
   `Alpha` and drops it, **Then** the list reads `Gamma`, `Alpha`, `Beta` immediately.
2. **Given** the user reordered the environments, **When** Xenon restarts and the workspace opens,
   **Then** the environments list in the saved order.
3. **Given** the user reordered the environments, **When** the user opens the environment switcher,
   **Then** it lists the environments in the same order as the sidebar ("No environment" stays
   first in the switcher, as today).
4. **Given** the Environments view, **Then** `Globals` stays the first row, cannot be dragged, and no
   environment can be dropped above it.
5. **Given** an environment file that cannot be read (shown as broken), **Then** it is listed after
   every readable environment, cannot be dragged, and is not a drop target.
6. **Given** the user drags an environment and drops it back where it started (or cancels with
   Escape), **Then** nothing is saved and the list is unchanged.
7. **Given** saving the new order fails (for example the workspace folder is read-only), **Then**
   the list returns to the previous order and the view shows an error banner with a next step.

---

### User Story 2 - The order travels with the workspace (Priority: P1)

The environment order is part of the workspace files, so a teammate who pulls the workspace from
git sees the same order, and two people who reorder on different machines get a sensible merged
order instead of a conflict.

**Why this priority**: Workspaces are shared through git (PRD: Git sync). An order that lived only
on one machine would silently disagree between teammates.

**Independent Test**: Reorder on one clone, commit and push, pull on another clone, and check the
order matches; reorder differently on both clones and pull, and check the result merges without a
conflict prompt.

**Acceptance Scenarios**:

1. **Given** a reordered workspace, **Then** the order is stored in the workspace's settings file
   (`workspace.yaml`) under `environments.order` as the list of environment ids, next to the
   collections' `order`, and is included in a commit like any other workspace change.
2. **Given** the order changes on disk outside Xenon (a git pull or a hand edit), **When** Xenon
   notices the change, **Then** the sidebar and switcher show the new order without a restart.
3. **Given** both sides of a pull changed the environment order, **When** Xenon merges the workspace
   file, **Then** the two orders merge the same way the collection order merges (no conflict is
   raised for the order).
4. **Given** the stored order names an environment id with no matching file (deleted elsewhere),
   **Then** that id is ignored and the rest of the list shows normally.
5. **Given** environment files on disk that the stored order does not name (added by hand or by a
   pull from an older Xenon), **Then** they list after the ordered environments, alphabetically
   (case-insensitive) among themselves.
6. **Given** a workspace that was never reordered and no environment has been added since this
   feature shipped, **Then** the environments list alphabetically, exactly as before, and the
   workspace file is not rewritten just by opening it or listing environments.

---

### User Story 3 - Reorder with the keyboard (Priority: P2)

A keyboard user focuses an environment row, presses Space to pick it up, moves it with the arrow
keys, and presses Space to drop it (Escape cancels). A screen reader announces the pick-up, each
new position and the drop, the same way reordering tabs is announced today.

**Why this priority**: Accessibility parity with the tab strip; it also gives a precise way to move
one step at a time.

**Independent Test**: With only the keyboard, move the second environment to the top and confirm
the order is saved and the announcements are made.

**Acceptance Scenarios**:

1. **Given** focus on an environment row, **When** the user presses Space, ArrowUp, Space, **Then**
   the environment moves one place up and the new order is saved.
2. **Given** an environment is picked up with the keyboard, **When** the user presses Escape,
   **Then** it returns to its original place and nothing is saved.
3. **Given** a keyboard drag, **Then** live announcements state which environment was picked up,
   its position out of the total as it moves, and where it was dropped.
4. **Given** focus on the `Globals` row or a broken environment row, **When** the user presses
   Space, **Then** nothing is picked up.

---

### User Story 4 - New environments go last; existing ones keep their place (Priority: P2)

Creating, duplicating or importing an environment puts it at the end of the list instead of
somewhere in the middle (also in a workspace that was never reordered: that first add saves the
current list with the new environment last); renaming an environment keeps it where it is;
deleting one removes it from the order.

**Why this priority**: Without this, the user's arrangement would be disturbed every time an
environment is added or renamed.

**Independent Test**: Reorder, then create, duplicate, import, rename and delete environments,
checking the position of each after every step.

**Acceptance Scenarios**:

1. **Given** environments in a saved order, **When** the user creates a new environment, **Then**
   it is listed last.
2. **Given** a never-reordered workspace with `Beta` and `Gamma`, **When** the user creates
   `Alpha`, **Then** the list reads `Beta`, `Gamma`, `Alpha` and `environments.order` now holds
   those three ids in that order.
3. **Given** environments in a saved order, **When** the user duplicates one, **Then** the copy is
   listed last.
4. **Given** environments in a saved order, **When** the user imports an environment whose name
   does not clash (or chooses "Keep both"), **Then** it is listed last; several imported at once
   are listed last in the order they were imported.
5. **Given** environments in a saved order, **When** an import replaces an existing environment
   ("Replace"), **Then** the replacing environment takes the replaced one's place.
6. **Given** environments in a saved order, **When** the user renames an environment, **Then** it
   keeps its place.
7. **Given** environments in a saved order, **When** the user deletes an environment, **Then** it is
   removed from the stored order and the others keep their relative order.

---

### User Story 5 - Reorder from an AI client (Priority: P3)

An AI client connected through Xenon's MCP server can set the environment order, so "put PROD
last" works from the assistant the same way it works by hand.

**Why this priority**: The PRD gives AI clients full Xenon access (XEN-008); reordering is a small
addition to the existing environment-editing tool.

**Independent Test**: Through the MCP `env_edit` tool, move an environment to a given position and
read the list back.

**Acceptance Scenarios**:

1. **Given** AI access is on, **When** an AI client asks `env_edit` to move an environment to a
   position, **Then** the order is saved exactly as a drag would save it and Xenon's sidebar
   refreshes.
2. **Given** the AI client names an environment id that does not exist, **Then** the tool returns
   a not-found error and the order is unchanged.

### Edge Cases

- **Filter active**: while the search box filters the list, reordering is disabled (rows are not
  draggable and keyboard pick-up does nothing); clearing the search re-enables it.
- **Renaming or creating in progress**: the row being renamed cannot be dragged; while the "new
  environment" input row is showing, no row can be dragged.
- **An environment file that is broken when the order is next written** (for example mid-merge):
  its id cannot be read, so it drops out of `environments.order`; once the file is fixed the
  environment lists in the alphabetical tail and can be dragged back into place.
- **Two reorders at nearly the same time** (the user and an AI client): each one re-reads the
  current order before writing, the last write wins, and the window refreshes to the saved order.
- **No hint while filtering**: rows simply show no grab cursor and do not move; the search box and
  list look as they do today.
- **Only one environment**: nothing to reorder; dragging it is harmless and saves nothing.
- **Duplicate ids in the stored order** (hand edit or bad merge): the first occurrence wins, later
  ones are ignored.
- **Malformed order value** (not a list of text values): treated as no order (alphabetical), and
  the next reorder writes a valid list.
- **A drop targets an index past the end**: the environment goes last.
- **Active environment**: reordering never changes which environment is active.
- **Open environment tabs**: reordering does not close, reorder or dirty any open environment tab.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Users MUST be able to move an environment to any position in the sidebar
  Environments list by dragging it with the pointer and dropping it.
- **FR-002**: Users MUST be able to move an environment with the keyboard (Space to pick up, arrow
  keys to move, Space to drop, Escape to cancel), with screen-reader announcements for pick-up,
  move and drop.
- **FR-003**: `Globals` MUST stay the first row of the Environments view; it MUST NOT be draggable
  and nothing may be dropped above it.
- **FR-004**: Broken (unreadable) environment rows MUST NOT be draggable and MUST NOT be drop
  targets.
- **FR-005**: While the search filter is non-empty, reordering MUST be disabled.
- **FR-006**: The system MUST save the new order to the workspace's settings file
  (`workspace.yaml`) under `environments.order` as an ordered list of environment ids, separate
  from the collections' top-level `order`, so it persists across restarts and
  is committed and synced like other workspace files.
- **FR-007**: The environment list returned to the sidebar and the switcher MUST be arranged as:
  environments named in the stored order, in that order (first occurrence wins; ids with no
  matching readable environment ignored); then every other readable environment, sorted
  alphabetically case-insensitively by name, as today; then the broken (unreadable) environment
  files, alphabetically by file name.
- **FR-008**: The environment switcher MUST list environments in the same order as the sidebar.
- **FR-009**: When an environment is created or duplicated, the system MUST place it last: the
  stored order becomes the current arranged list of readable environment ids (FR-007) with the new
  id appended — also when no order was stored before.
- **FR-010**: When an environment is imported without replacing another, the system MUST place it
  last in the stored order; when an import replaces an existing environment, the replacing
  environment MUST take the replaced one's position.
- **FR-011**: Renaming an environment MUST NOT change its position.
- **FR-012**: Deleting an environment MUST remove its id from the stored order.
- **FR-013**: When the order changes on disk outside Xenon, the sidebar and switcher MUST reflect
  it without a restart.
- **FR-014**: When a git pull merges two different changes to the environment order, the system
  MUST merge them with the same rule as the collection order, without raising a conflict.
- **FR-015**: A failed save of the order MUST restore the previous order on screen and show an
  error with a next step; no partial order may be written.
- **FR-016**: Reordering MUST NOT change the active environment, open tabs, or any environment
  file.
- **FR-017**: The AI `env_edit` tool MUST offer a reorder action that moves one environment to a
  given position, with the same validation and result as the in-app reorder.
- **FR-018**: A workspace with no stored order MUST list alphabetically as today and MUST NOT have
  its workspace file rewritten by opening it or listing environments; only a reorder, create,
  duplicate or import writes `environments.order`, and a delete writes it only when the stored
  order names the deleted id.
- **FR-019**: Every other change that rewrites `workspace.yaml` (collection reorder, globals
  edits, workspace rename, sync settings, imports that touch the workspace file) MUST keep
  `environments.order` and any other keys under `environments` exactly as they were.

### Key Entities

- **Environment order**: the ordered list of environment ids stored in `workspace.yaml` under
  `environments.order`, next to (and independent of) the collections' top-level `order`. May be
  absent (alphabetical order), may contain stale ids (ignored) and may omit environments (they
  follow, alphabetically).
- **Environment list**: what the sidebar and switcher show — every environment of the workspace,
  arranged by the environment order, plus the active environment id.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can move any environment to any position in a single drag, or in at most N
  arrow-key presses where N is the distance between the two positions.
- **SC-002**: After a reorder, the new order is visible in both the sidebar and the switcher with
  no perceptible delay (the list updates on drop, before the save completes), and 100% of reorders
  survive an app restart.
- **SC-003**: Two clones that reorder environments differently and then pull from each other end
  with the same order and zero conflict prompts about the order.
- **SC-004**: In a workspace never reordered, opening the workspace and listing environments
  leaves the displayed order alphabetical and the workspace file byte-for-byte unchanged.
- **SC-005**: Every acceptance scenario above is covered by an automated test (backend scenarios
  for persistence, merging, placement and the AI tool; UI tests for pointer and keyboard
  reordering).

## Assumptions

- Postman does not let users reorder environments; this is a Xenon addition on top of the
  Environments view, which is itself a recorded Xenon decision. The PRD's "Drag and drop" section
  and decisions log are updated to list it, so it is in scope under the constitution.
- Ids, not names or file names, identify environments in the stored order, so renaming (which can
  rename the file) never breaks the order.
- Older Xenon builds that do not know `environments` read `workspace.yaml` without error and,
  because they patch only the keys they own, keep `environments.order` when they save the file
  (verified: `yamlnode.Patch` in every `workspace.yaml` writer). They list environments
  alphabetically and do not update the order on create or delete; the next reorder in a current
  build tidies it (stale ids ignored, unlisted environments in the alphabetical tail).
- Moving environments between workspaces stays out of scope (export and import, as today).
- Exporting a whole workspace as Xenon YAML copies `workspace.yaml`, so the order travels with it;
  Postman exports have no environment order and are unaffected.
- Importing a whole Xenon workspace (a copy of its folder) keeps the source's environment order,
  because `workspace.yaml` is copied as-is. Environments imported into an existing workspace
  (Postman, Bruno, Xenon environment files, Postman dumps and workspace folders) follow FR-010.
