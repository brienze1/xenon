# Feature Specification: Request Pane Content Never Overlaps

**Feature Branch**: `xen-019`

**Created**: 2026-10-23

**Status**: Draft

**Input**: User description: "XEN-019 — request tab content must never overlap when the request/response divider is dragged up. Today, when the user drags the divider between the request pane and the response pane upward (or the window is small), the request pane's content draws on top of itself — most visibly on the Scripts sub-tab, where the hint paragraph ("Type pm. for the API…") is drawn over the Pre-request/Post-response phase tabs and the "Inherit parent scripts" checkbox. Users need every request sub-tab (Params, Authorization, Headers, Body, Scripts, Settings), example tabs, collection/folder tabs (Overview, Authorization, Headers, Scripts, Variables, Documentation), and environment and globals tabs to stay legible at any pane size: no text or control overlaps another; fixed rows keep their natural height; editors and tables shrink first down to a sensible minimum, after which the content area scrolls vertically. Decisions already made by the product owner: (1) the breadcrumb/URL bar and the request sub-tab row stay pinned and only the panel under the sub-tab row scrolls; (2) the request pane gets a minimum height of about 200px in stacked mode in addition to the existing 20%–80% ratio limit; if extra banners push the content past that, the request builder as a whole scrolls as a fallback rather than clipping; (3) the Scripts hint keeps its full text and wraps inside its own row below the editor; (4) in side-by-side mode no new minimum width is added. Taller panes must look exactly as they do today. UI-only. Source: GitHub issue brienze1/xenon#26."

## Context

A layout bug reported in GitHub issue [brienze1/xenon#26](https://github.com/brienze1/xenon/issues/26).
A request tab shows the request pane (breadcrumb, URL bar, sub-tab row, active sub-tab) above the
response pane, separated by a draggable divider. The divider can be dragged so the request pane
takes as little as 20% of the area. At Xenon's minimum window size (960 × 600) that is roughly
75–100 px — less than the breadcrumb, URL bar and sub-tab row need on their own (about 110 px).

What goes wrong today (verified against `main` at `db23057`): the Scripts sub-tab is laid out so
that every part of it is allowed to shrink to nothing. When the pane is shorter than its content,
the phase-tab area collapses, its contents spill out of it, and the hint paragraph — placed right
after the collapsed area — is drawn on top of the Pre-request / Post-response tabs and the
"Inherit parent scripts" checkbox. The panel under the sub-tab row already knows how to scroll,
but the Scripts content never grows past the panel's height, so it never scrolls. The same
"shrink to nothing" pattern exists in the Scripts section of collection/folder tabs. Separately,
when the request pane is shorter than the URL bar plus the sub-tab row, those rows are cut off
at the pane's edge with no way to reach them.

## Clarifications

### Session 2026-10-23 (product owner, before specification)

- Q: When the request pane is short, what scrolls? → A: Only the panel under the sub-tab row;
  the breadcrumb, URL bar and sub-tab row stay pinned.
- Q: Should the request pane get a pixel minimum height in addition to the 20% ratio? → A: Yes,
  about 200 px in stacked mode, so the URL bar, sub-tab row and a couple of content rows always
  fit. If banners push the content past that, the whole request builder scrolls as a fallback.
- Q: What happens to the Scripts hint? → A: Keep the full text; it wraps inside its own row below
  the editor and never shrinks.
- Q: In side-by-side mode, should the request pane get a minimum width? → A: No. The overlap fix
  applies in both modes; the new minimum is height-only, in stacked mode.

### Session 2026-10-23 (speckit-clarify)

- Q: The issue asked for a by-eye check on macOS, Windows and Linux; what must the PR's manual
  verification cover? → A: macOS by eye (before/after screenshots at the highest divider
  position) plus CI's quality gate for Windows and Linux; no manual Windows/Linux step blocks
  the merge.
- Q: When exactly does the whole request builder start scrolling (FR-006)? → A (settled from
  the layout, no user input needed): when the pinned rows plus banners leave the panel under
  the sub-tab row less than 80 px — the room it has at the 200 px floor with no banner. Above
  that, only the panel scrolls.
- Q: Does the pinned URL bar/sub-tab rule apply to example tabs? → A (settled from the code:
  example tabs have no divider; their request part is already one scroll box above the response
  part): no. The example request part keeps scrolling as a whole, as today; FR-004 is for
  request tabs.
- Q: Is the 200 px floor in screen pixels or UI pixels, given View → Zoom (0.67–2) scales the
  whole UI? → A (settled from `app/zoom.ts`: the body is scaled with CSS `zoom`, so every size in
  the UI scales together): UI pixels — the floor scales with the zoom like every other size.
- Q: What wins when the 200 px floor would exceed the 80 % maximum (a small split area, e.g. the
  minimum window at a high zoom)? → A (settled for consistency with the existing 20 %–80 % limits,
  which guarantee the response pane at least 20 %): the floor never takes more than 80 % of the
  split area; the response pane keeps at least 20 %.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Scripts tab stays legible in a short request pane (Priority: P1)

A user editing a request's scripts drags the divider up to give the response more room. The
Pre-request / Post-response tabs, the "Inherit parent scripts" checkbox, the Snippets toolbar,
the editor and the hint all stay readable and clickable, each in its own space. When there is
not enough room for all of them, the area under the request's sub-tab row scrolls so the user
can reach every part.

**Why this priority**: This is the reported bug; today the Scripts tab is unusable with a short
request pane.

**Independent Test**: Open a request, select Scripts, drag the divider to its highest point and
check that no text or control is drawn over another and that scrolling the panel reaches the hint.

**Acceptance Scenarios**:

1. **Given** a request tab on Scripts with the divider at its highest point, **When** the user
   looks at the request pane, **Then** the hint is not drawn over the phase tabs or the
   "Inherit parent scripts" checkbox, and no other text or control overlaps.
2. **Given** the same short pane, **When** the user scrolls the area under the sub-tab row,
   **Then** the phase tabs, toolbar, editor (at its minimum height) and the full hint text are
   reachable in that order.
3. **Given** a narrow request pane, **When** the hint text is longer than one line, **Then** it
   wraps onto more lines inside its own row and pushes nothing on top of anything else.
4. **Given** a tall request pane (the default 1280 × 800 window with the divider at its lowest
   point, 80%), **When** the user opens Scripts, **Then** it looks exactly as it does today: the editor fills the free
   height and no scrollbar appears.

---

### User Story 2 - URL bar and sub-tab row always reachable (Priority: P1)

A user who drags the divider all the way up still sees the breadcrumb, the URL bar with its Send
button and the sub-tab row, plus at least a couple of rows of the active sub-tab, at any window
size Xenon allows.

**Why this priority**: Without a floor the URL bar and sub-tabs are cut off at small window
sizes, which is the same bug in another form.

**Independent Test**: Shrink the window to 960 × 600, drag the divider to its highest point and
check that the request pane is at least about 200 px tall and shows the URL bar and sub-tab row
in full.

**Acceptance Scenarios**:

1. **Given** the stacked layout at the minimum window size and 100% zoom, **When** the user drags
   the divider to its highest point, **Then** the request pane stays at least 200 px tall and the
   breadcrumb, URL bar and sub-tab row are fully visible.
2. **Given** a tall window where 20% of the area is more than 200 px, **When** the user drags the
   divider up, **Then** it stops at 20%, as today.
3. **Given** a request pane at its minimum height that also shows a banner (an error, an
   on-disk conflict or the cURL paste notice), **When** the content no longer fits, **Then** the
   request pane scrolls as a whole instead of clipping or overlapping anything.
4. **Given** the side-by-side layout, **When** the user drags the divider, **Then** the existing
   20%–80% limits apply and no new minimum width is added.

---

### User Story 3 - Every other stacked tab behaves the same (Priority: P2)

The same rule — fixed rows keep their height, editors and tables shrink to a minimum, then the
content scrolls — holds on every request sub-tab (Params, Authorization, Headers, Body, Scripts,
Settings), on example tabs (request part and response part), on collection and folder tabs
(Overview, Authorization, Headers, Scripts, Variables, Documentation), and on environment and
globals tabs.

**Why this priority**: These share the stacked layout; the bug must not survive on a tab the
report did not mention.

**Independent Test**: At the minimum window size (and, for request tabs, the divider at its
highest point), visit each listed tab and check that nothing overlaps and the content scrolls.

**Acceptance Scenarios**:

1. **Given** each request sub-tab with the divider at its highest point, **When** the user views
   it, **Then** nothing overlaps and the panel scrolls to reach the rest of the content.
2. **Given** a collection or folder tab on Scripts at the minimum window size, **When** the user
   views it, **Then** the hint is not drawn over the phase tabs or the inherit checkbox (a
   folder shows the checkbox; a collection does not).
3. **Given** an example, collection/folder, environment or globals tab at the minimum window
   size, **When** the user views any of its sections, **Then** nothing overlaps and the tab's
   content scrolls.

### Edge Cases

- Window at its minimum size with the console drawer open: the request pane still keeps its
  200 px floor; the response pane gets the rest.
- A small split area (the minimum window at a high View → Zoom, where the UI is scaled up and
  the area holds fewer UI pixels): the floor never takes more than 80% of the area, so the
  response pane keeps at least 20%; the request pane's content then scrolls (FR-006).
- A long script or a long body: the editor keeps its own internal scrolling; the panel only
  scrolls when the editor is already at its minimum height.
- A disabled script ("This script is disabled and does not run" with Enable) adds to the toolbar
  row; that row wraps or grows and is never overlapped.
- Keyboard resizing of the divider (arrow keys, Home, End) respects the same 200 px floor as
  dragging.
- The remembered divider position is not rewritten by the floor: a saved 20% ratio stays saved
  and is shown at the floor height whenever 20% is less than 200 px.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: At any request pane height, no text or control in the request pane MUST be drawn
  over another text or control, on any request sub-tab (Params, Authorization, Headers, Body,
  Scripts, Settings).
- **FR-002**: Fixed rows MUST keep their natural height — at least the height of their content,
  including wrapped lines — and never shrink: the request sub-tab
  row; the Scripts phase-tab row with the "Inherit parent scripts" checkbox; the Scripts toolbar
  (Snippets, disabled-script notice); the Scripts hint; every other toolbar, heading or note row
  in the affected tabs.
- **FR-003**: Editors MUST shrink first, down to their existing minimum heights (Scripts editor
  160 px, body editors 160 px, example response editor 120 px, documentation editor 360 px,
  collection/folder Scripts section 320 px; the example response panel gets a new 160 px minimum,
  room for its view switch and 120 px editor); below that the content MUST scroll vertically
  instead of shrinking further. Tables (params, headers, form fields, variables) keep their
  natural row heights and never shrink; the content around them scrolls.
- **FR-004**: In a request tab, the breadcrumb, URL bar and sub-tab row MUST stay pinned at the
  top of the request pane; only the panel under the sub-tab row scrolls (subject to FR-006). An
  example tab's request part keeps scrolling as one box, as today.
- **FR-005**: In the stacked layout the request pane MUST be at least 200 px tall (UI pixels,
  scaling with View → Zoom), in addition to the existing 20%–80% ratio limits; the larger of
  200 px and the stored position wins, but the floor never makes the request pane taller than 80%
  of the split area, so the response pane keeps at least 20%. The floor MUST apply to dragging
  and to keyboard resizing alike.
- **FR-006**: The panel under the sub-tab row MUST keep at least 80 px of height; when the
  pinned rows plus any banners would leave it less (for example a banner shown while the pane is
  at its floor), the request pane MUST scroll as a whole rather than clip or overlap its content.
- **FR-007**: The Scripts hint MUST keep its current full text and MUST wrap onto as many lines
  as it needs inside its own row below the phase editor.
- **FR-008**: In the side-by-side layout no new minimum width MUST be added; the existing ratio
  limits stay, and FR-001–FR-003 and FR-007 still apply.
- **FR-009**: The same rules (FR-001–FR-003, FR-007) MUST hold on example tabs (request and
  response parts), collection and folder tabs (Overview, Authorization, Headers, Scripts,
  Variables, Documentation), and environment and globals tabs. What scrolls: an example tab's
  request part scrolls as one box (as today) and its response part scrolls as one box under the
  "Example Response" heading (new — today it is cut off); collection, folder, environment and
  globals tabs scroll as a whole tab, as today.
- **FR-010**: When the request pane is tall enough for all its content, every affected tab MUST
  look as it does today: editors fill the free height and no scrollbar appears.
- **FR-011**: The remembered divider position MUST NOT be changed by the 200 px floor; the floor
  only affects how the saved position is displayed.
- **FR-012**: Scrolling MUST NOT hurt keyboard use: a control that receives focus inside a scroll
  area is scrolled into view (the platform's default behaviour, which no rule may disable), focus
  rings at the edges of a scroll area are not cut off, and no new horizontal scrollbar appears in
  the request pane or the example response part.

### Key Entities

- **Request/response divider position**: the remembered share of the area the request pane
  takes (20%–80%), already stored with the workspace's layout. Unchanged by this feature; the
  200 px floor is applied when displaying it.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: With the divider at its highest point at the 960 × 600 minimum window size, 0
  overlapping texts or controls on all 6 request sub-tabs, both example parts, all 6
  collection/folder sections, and the environment and globals tabs.
- **SC-002**: At the minimum window size the request pane is never shorter than 200 px in the
  stacked layout, and the URL bar and sub-tab row are fully visible 100% of the time.
- **SC-003**: Every part of the Scripts sub-tab (phase tabs, inherit checkbox, toolbar, editor,
  full hint) can be reached by scrolling at the shortest request pane.
- **SC-004**: At the default 1280 × 800 window with the divider at its lowest point (80%, where
  every tab's content fits), before/after screenshots of each affected tab are visually
  identical.
- **SC-005**: The full quality gate (lint, typecheck, unit tests, backend suite) passes on CI
  (macOS, Windows, Linux), and the pull request includes macOS before/after screenshots at the
  highest divider position.

## Assumptions

- UI-only: no stored data, file format or backend behaviour changes; the divider ratio stays the
  remembered value it is today.
- The existing editor minimum heights are the "sensible minimum" the issue asks for; they are not
  changed.
- "About 200 px" is fixed at exactly 200 px.
- The response pane gets no new minimum of its own; the existing 80% ratio limit stays its floor.
- Horizontal overflow at very narrow side-by-side widths is out of scope (decision 4); only
  vertical overlap is fixed.
- The by-eye check is done in the running app on macOS only (Clarifications); Windows and Linux
  are covered by the CI quality gate, and no manual check on them blocks the merge.
