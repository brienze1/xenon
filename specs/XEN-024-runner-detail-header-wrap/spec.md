# Feature Specification: Runner Row Detail Keeps Header Names Readable

**Feature Branch**: `xen-024`

**Created**: 2026-11-17

**Status**: Draft

**Input**: User description: "XEN-024 — In a collection or folder tab's Runner → Results view, opening a result row shows a details side panel whose Request section lists the request headers that were sent as a name/value table. Today each header name (Content-Type, User-Agent, Accept, Accept-Encoding, Connection) is broken one character per line in a column one letter wide, so the panel becomes a very tall column of single letters and users must scroll a long way to reach the body and the response. Users need each header name to stay on one line (or wrap only at a sensible width, never one character per line) in a left column wide enough for typical names such as Accept-Encoding and Content-Type, with values filling the remaining width; very long names or values must wrap inside their own cell without making the panel scroll sideways. The result must be the same in the macOS WebKit webview and on Windows/Linux. The same defect pattern (a name column allowed to break anywhere in an automatically sized table) also exists in the API documentation view's Parameters table (Name column) and must be fixed there too; other key/value tables are not affected. UI-only defect: the sent request and response data are correct. Source: GitHub issue brienze1/xenon#41."

## Context

A layout bug reported in GitHub issue [brienze1/xenon#41](https://github.com/brienze1/xenon/issues/41).
A collection or folder tab has a Runner section whose Results view lists one row per request
sent. Opening a row shows a details side panel; its **Request** section shows the method and URL,
then a two-column list of the headers that were sent (name on the left, value on the right),
then the body, and below that the **Response** section.

What goes wrong today (verified against `main` at `cc82759`): every header name in that table is
broken after each character, so a name like `Content-Type` is drawn as a column of single letters
next to its value. The reported request (`POST …/campaigns/510876/briefings/7/guidelines`, with the
default `Content-Type`, `User-Agent`, `Accept`, `Accept-Encoding` and `Connection` headers) turns
the panel into a tall strip of letters and pushes the body and response far down. The sent request
and the response data are correct; only the layout is wrong.

The same layout pattern — a name column allowed to break between any two characters, inside a
table whose column widths are decided by its content — also exists in the **Parameters** table of
the API documentation view (the **Name** column), where a long example value can squeeze the names
the same way. The other key/value tables in the app (response headers, request-builder headers,
params and form body, sync conflict rows, console entry headers) give their columns fixed shares
or a minimum width and are not affected; they are out of scope.

## Clarifications

### Session 2026-11-17 (speckit-clarify)

- Q: A header name longer than the name column's cap — wrap or ellipsize? → A: Wrap within its
  own cell (at hyphens/spaces, mid-word only when one segment exceeds the cap); never cut off, so
  every character stays visible and selectable (product owner).
- Q: How is the visual result verified, given jsdom has no layout? → A (settled by the XEN-019 /
  XEN-021 precedent for UI-only layout fixes): Vitest locks the DOM and style contract with the
  real stylesheets injected; the visual result is checked by eye on macOS with before/after
  screenshots in the pull request; CI's quality gate runs on macOS, Windows and Linux.
- Q: How are the header pairs exposed to assistive technology once the layout is fixed? → A
  (settled from the code: the console entry detail already shows request headers as a labelled
  name/value list, and a table restyled as a grid loses its table semantics in WebKit): as a
  name/value list labelled "Request headers", each name the term for the value after it; no new
  focus stops are added.
- Q: Are the documentation Parameters table's other columns (In, Description) affected? → A
  (settled from the code: they hold short words and prose without break-anywhere wrapping): no;
  only the Name column gets the one-line/cap rule, Example keeps wrapping within its cell.
- Q: What is the Name cap in the documentation Parameters table, which is a four-column table
  sized by its content? → A (settled in planning, research R3): 32 characters of the name's own
  font — about 40% of the table at the 960 px minimum window width — because a percentage cap
  cannot be expressed reliably inside an automatically sized table on the app's oldest supported
  webview (macOS 12).

- Q: 40% of the default details panel is narrower than `Accept-Encoding` — which wins? → A
  (settled from the code: `NewLayout` opens the 280 px sidebar and the 320 px right panel, so in
  the 1280 px default window the details panel sits at its 280 px minimum and 40% is ≈ 14
  characters; the issue requires the column to fit `Accept-Encoding`): the cap is the larger of
  40% and 16 characters.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Read the sent headers of a runner row at a glance (Priority: P1)

A user runs a folder, opens a result row, and reads the request headers that were sent. Each
header name sits on one line in a left column wide enough for it; the value fills the rest of the
row. The headers take one short line each, so the body and the response are visible without a
long scroll.

**Why this priority**: This is the reported defect. Today the panel is effectively unreadable for
any request that sends headers, which is every request (the app adds default headers).

**Independent Test**: Open the details panel of a runner row whose request sent the default
headers (`Content-Type`, `User-Agent`, `Accept`, `Accept-Encoding`, `Connection`) and confirm that
every name is on a single line and each header occupies one line.

**Acceptance Scenarios**:

1. **Given** a run whose row sent `Content-Type: application/json`, `User-Agent: Xenon/0.1.14`,
   `Accept: */*`, `Accept-Encoding: gzip, deflate, br` and `Connection: keep-alive`, **When** the
   user opens that row's details panel in the default 1280×800 window, **Then** each header name
   is drawn on one line, and each header (name and value) occupies exactly one line.
2. **Given** the same row, **When** the details panel is shown, **Then** all names share one left
   column whose width fits the longest name (`Accept-Encoding`), and every value starts at the same
   left edge to the right of that column.
3. **Given** the same row, **When** the details panel is shown, **Then** the panel has no
   horizontal scrollbar.

---

### User Story 2 - Long names and values stay inside the panel (Priority: P2)

A user opens a row whose request carried an unusually long header (for example a custom
`X-Very-Long-Custom-Correlation-Header-Name-For-Tracing-Abcdef` name, or a multi-hundred-character
`Authorization`/cookie value). The long content wraps inside its own cell; the panel never
scrolls sideways and the other headers keep their readable layout.

**Why this priority**: Real requests carry long tokens and custom headers; the fix must not trade
one broken layout for another.

**Independent Test**: Open the details panel of a row that sent the default headers plus one
header with a 61-character name and a 400-character value without spaces; confirm no horizontal
scrolling and that the short names still sit on one line.

**Acceptance Scenarios**:

1. **Given** a row with a header value of 400 characters and no spaces, **When** the details panel
   is shown, **Then** the value wraps within its cell and the panel does not scroll sideways.
2. **Given** a row with a header name longer than the name column's cap, **When** the details
   panel is shown, **Then** that name wraps within the name column (at its hyphens where possible),
   the panel does not scroll sideways, and the other names stay on one line.
3. **Given** a row with a header name longer than the cap, **When** the details panel is shown,
   **Then** the name column is no wider than its cap (the larger of 40% of the header list's width
   and 16 characters) and the values keep the remaining width.

---

### User Story 3 - API documentation parameter names stay readable (Priority: P3)

A user views the generated API documentation for a request whose parameters include a long
example value (for example a long token). Each parameter name stays readable on one line in the
**Name** column instead of being squeezed to one character per line.

**Why this priority**: Same defect pattern in a less-used view; fixing it keeps the rule
consistent across the app's name/value tables.

**Independent Test**: Open the documentation view for a request with a parameter named
`Accept-Encoding` (or `page_size`) and a 400-character example value; confirm the name is on one
line and the documentation view does not scroll sideways.

**Acceptance Scenarios**:

1. **Given** a documented request with a parameter whose example value is 400 characters without
   spaces, **When** the Parameters table is shown, **Then** the parameter name (and its `required`
   marker) is drawn on one line and the example value wraps within its own cell.
2. **Given** a parameter name longer than 32 characters, **When** the Parameters table is shown,
   **Then** that name wraps within a block at most 32 characters wide and the view does not
   scroll sideways.

---

### Edge Cases

- **No headers sent**: the header list is not shown (unchanged behaviour).
- **Header name longer than the cap with no hyphen or space** (e.g. 80 letters): it breaks inside
  its own cell only once it reaches the cap, never at one character per line, and never pushes the
  panel wider.
- **Many headers (e.g. 30)**: each takes one line when names and values fit; the panel scrolls
  vertically as it does today.
- **Empty header value**: the row still shows its name on one line; the value cell is empty.
- **Panel at its minimum width (280 px — which is what the default layout gives: 1280 px window
  with the 280 px sidebar and 320 px right panel open)**: 40% of the list is only about 14
  characters, so the 16-character floor applies; all five default names (longest:
  `Accept-Encoding`, 15) stay on one line and nothing scrolls sideways.
- **Secret values (masked)**: masked values are shown exactly as today; only the layout changes.
- **Header value with spaces** (e.g. `gzip, deflate, br`): wraps at spaces when it does not fit,
  before breaking inside words.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: In the runner row details panel, each sent header name MUST be drawn on a single
  line whenever it fits within the name column's cap; it MUST never be broken after every
  character.
- **FR-002**: The header list MUST span the full width of the Request section; names sit in a
  left column sized to the longest name, values take the remaining width, and all values start at
  the same left edge.
- **FR-003**: The name column's cap MUST be the larger of 40% of the header list's width and 16
  characters of the list's own monospaced text (so the floor scales with the text size), so names of up to 16 characters (`Accept-Encoding`, `Accept-Language`,
  `X-Correlation-Id`) always fit on one line, even at the panel's 280 px minimum. A name longer
  than the cap wraps within its own cell, preferring hyphens and spaces, and breaks inside a word
  only when a single segment is still longer than the cap.
- **FR-004**: Header values MUST wrap within their own cell — at spaces where possible, and inside
  a word when a single word is longer than the cell — so no value widens the panel.
- **FR-005**: The runner row details panel MUST NOT scroll horizontally because of header names or
  values, whatever their length.
- **FR-006**: In the API documentation view's Parameters table, each parameter name MUST be drawn
  on one line (with its `required` marker after it on the same line) whenever the name is at most
  32 characters long; a longer name wraps within a 32-character-wide block (hyphens and spaces
  first, mid-word only when one segment is longer), never one character per line. Example values
  keep wrapping within their own cell (FR-004), and the view never scrolls sideways because of a
  name or example. The request **Body** fields table in the same view is drawn by the same
  Parameters table and follows the same rule.
- **FR-007**: The layout rules above MUST hold in every webview the app ships on — WebKit (macOS,
  Linux) and WebView2 (Windows).
- **FR-008**: What the panels show MUST NOT change: same headers, in the same order, same masking
  of secrets, same names and values; same parameters, columns and markers in the documentation
  view. Only their layout changes.
- **FR-009**: The other key/value tables (response headers, request-builder headers/params/body,
  sync conflict rows, console entry headers) MUST NOT change.
- **FR-010**: The runner header list MUST stay accessible as name/value pairs: it carries the
  accessible name "Request headers", and each name is announced as the term whose value follows it
  — the same structure the console entry detail already uses for request headers.

### Key Entities

- **Sent header**: a name/value pair recorded with a runner row as it was sent (secrets masked).
  Unchanged by this feature; only how it is laid out changes.
- **Documented parameter**: a name, location, description, example and required flag shown in the
  API documentation view. Unchanged; only the Name column's layout changes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For the reported request (5 default headers), the header list in the details panel
  is 5 lines tall — one line per header — instead of one line per character of the names (today
  roughly 60 lines).
- **SC-002**: In the default 1280×800 window, a user who opens the details panel of a request that
  sent the 5 default headers and a URL of up to 120 characters sees the start of the request body
  without scrolling the panel.
- **SC-003**: For a row with a 400-character header value and a 61-character header name, the
  details panel shows 0 px of horizontal scroll.
- **SC-004**: In the documentation view, a parameter named `Accept-Encoding` (marked required) next
  to a 400-character example is drawn on 1 line, and a 51-character name wraps to 2 lines at most
  32 characters wide.
- **SC-005**: The by-eye check on macOS (before/after screenshots of the reported row) shows names
  on one line; the automated quality gate passes on macOS, Windows and Linux.

## Assumptions

- The 40% cap on the runner name column comes from the issue's proposed scope ("such as ~40%");
  the 16-character floor comes from the issue's other expectation, that the column is wide enough
  for typical names such as `Accept-Encoding`, which 40% alone is not at the panel's default
  280 px width (see Clarifications).
- Long names past the cap wrap rather than being cut off with an ellipsis (product owner's
  decision, see Clarifications), so every character stays visible and selectable.
- The details panel's own size, position and minimum width (280 px) are unchanged.
- The response viewer inside the details panel (its Headers tab) is the shared response viewer,
  which already uses fixed equal columns and is not affected.
- This is a UI-only defect: no stored data, run format or backend behaviour changes.
- Selecting and copying header lines from the list gives each name and value on their own lines
  (a name/value list, not table cells); no copy action exists for this list today and none is
  added.
- The panel's colours, fonts, theme handling (light/dark) and vertical scrolling are unchanged;
  header names and values stay excluded from page translation.
