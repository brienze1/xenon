# Research: Runner Row Detail Keeps Header Names Readable

All items below were settled from the code on `origin/main` @ `cc82759` (2026-11-17) and from a
jsdom probe run in this worktree. No `NEEDS CLARIFICATION` remained after this phase.

## R1 — Why names break one character per line

- **Decision**: The cause is `overflow-wrap: anywhere` on both `th` and `td` of an automatically
  sized table (`RunRowDetail.module.css` `.headers th, .headers td`), with no table width.
- **Rationale**: `overflow-wrap: anywhere` (unlike `break-word`) lowers a cell's min-content width
  to one character. The auto table layout algorithm only guarantees each column its min-content
  width and shares the rest by max-content, so inside the 280 px+ side panel (`RunResults.module.css`
  `.withDetail` → `minmax(280px, 2fr)`) the name column collapses to one letter while values take
  the remaining width — exactly the issue's screenshot.
- **Alternatives considered**: none; the mechanism is defined by CSS Text 3 and CSS 2.1 §17.5.2.2.

## R2 — Runner header layout: grid list vs. fixed table vs. restyled table

- **Decision**: Replace the `<table>` with `<dl aria-label="Request headers">` whose `dt`/`dd`
  pairs are direct children of one grid,
  `grid-template-columns: fit-content(max(40%, 16ch)) minmax(0, 1fr)`;
  `dt` and `dd` keep `overflow-wrap: anywhere` and `min-width: 0`.
- **Rationale**: `fit-content(L)` sizes the column to `min(max-content, max(min-content, L))`;
  with break-anywhere names the min-content is one character, so the column is
  `min(longest name, max(40%, 16ch))` — exactly FR-002/FR-003. The 16-character floor (R6) keeps
  `Accept-Encoding` on one line in the default layout. Grid track sizing is precisely specified and
  identical in WebKit, Blink and WebKitGTK (fit-content() since Safari 10.1), unlike auto table
  layout, which is engine-specific. The console entry detail already presents request headers as a
  labelled `<dl>` (`ConsoleEntryDetail.tsx` `Headers`), so the runner panel now matches the app's
  existing pattern. All dt/dd as direct grid items keep every value aligned on one left edge.
- **Alternatives considered**:
  - *`table-layout: fixed; width: 100%`* (what `ResponseHeaders` does) — no squeeze, but the name
    column becomes a fixed 50% share regardless of names: wastes half the 280 px panel on
    `Accept`. Rejected (FR-002 asks for a column sized to the names).
  - *Keep `<table>`, `th { white-space: nowrap; width: 1% }`* — sizes to the names but has no cap:
    one 80-character name widens the table past the panel (violates FR-003/FR-005); `max-width`
    on table cells is undefined in CSS 2.1 and ignored by engines.
  - *Keep `<table>` but `display: grid` on it with `display: contents` rows* — WebKit has dropped
    table semantics for re-displayed tables and `display: contents` rows; risky for FR-010.
  - *Container query units for a percentage cap inside the table* — needs Safari 16+; the app
    supports macOS 12 (`LSMinimumSystemVersion` 12.0.0), whose WebKit may be Safari 15.
- **Ripple**: `RunResults.test.tsx:182` queries `getByRole('rowheader', { name: 'Authorization' })`;
  it becomes a `term`/`definition` query. No other code queries the runner header table
  (`grep -rn "rowheader\|Request headers" frontend/src`).

## R3 — Documentation Parameters table: how to cap the Name column

- **Decision**: Keep the four-column `<table>` (it has column headers: Name/In/Description/Example).
  The Name cell drops break-anywhere and gets `white-space: nowrap`; the name itself is wrapped in
  `<span class="parameterName">` with `display: inline-block; width: max-content; max-width: 32ch;
  white-space: normal; overflow-wrap: anywhere; vertical-align: top`.
- **Rationale**: The span's used width is `min(max-content, 32ch)` and is definite, so it is also
  its min-content contribution — the auto table cannot squeeze the column below it, and a name
  longer than 32 characters wraps inside the span instead of widening the table. 32ch in the mono
  name font is ≈ 250 px, about 40% of the documentation table at the 960 px minimum window
  width (`internal/domain/entity/window_state.go` `MinWindowWidth = 960`). Every feature used
  (`inline-block`, `max-content`, `ch`) works on the oldest supported webview. The `required`
  marker stays after the span on the same line because the cell is `nowrap`.
- **Alternatives considered**: a percentage cap (not expressible for content inside an auto table
  without container query units — R2); converting to a grid (loses table semantics — R2);
  `overflow-wrap: break-word` only (an unbreakable long name would then widen the table past the
  view, FR-006).
- **Scope note**: The same `Parameters` component renders the request **Body** fields table
  (`ApiDocView.tsx` `OperationDetails`), so the fix applies to both tables automatically.

## R4 — What jsdom can verify

- **Decision**: Tests assert the DOM structure and the computed-style contract with the real CSS
  module injected via `injectStylesheets`.
- **Rationale**: A probe in this worktree (jsdom 30.1.2) returned, from an injected rule:
  `display: grid`, `grid-template-columns` as authored (e.g. `fit-content(max(40%, 16ch)) minmax(0, 1fr)` — author text
  kept, so tests compare with whitespace removed), `overflow-wrap: anywhere`, `white-space: nowrap`,
  `width: max-content`, and `max-width: 32ch` resolved to `256px` (jsdom assumes 8 px per `ch`;
  tests assert `256px`). jsdom performs no layout, so line counts and widths are verified by eye
  (R5).
- **Alternatives considered**: Playwright WebKit in CI — its download hung on this machine before
  (XEN-021) and the user did not want it chased; not adopted.

## R5 — Visual verification

- **Decision**: By-eye on macOS with the scratchpad Vite harness (render `RunRowDetail` and
  `ApiDocView` with fixture data — the reported 5 default headers, a 61-character name and a
  400-character unbroken value; a docs parameter `Accept-Encoding` required + a 400-character
  example, and a 51-character name), at panel widths 280 px (default layout) and 420 px (right
  panel hidden), before
  (`origin/main` copies) and after. Screenshots go to `screenshots/` and the PR.
- **Rationale**: XEN-019/XEN-021 precedent for UI-only layout fixes; the harness keeps lint and
  typecheck clean because it lives outside the repo.

## R6 — Default panel width and the 16-character floor

- **Decision**: The cap is `max(40%, 16ch)`, not 40% alone.
- **Rationale**: `internal/domain/entity/layout.go` `NewLayout` shows the 280 px sidebar and the
  320 px right panel; in the 1280 px default window the runner's `rows` grid
  (`minmax(0, 3fr) minmax(280px, 2fr)`) leaves the details panel at its 280 px minimum, 256 px of
  content after padding. 40% = 102 px ≈ 14 characters of the 12 px JetBrains Mono (0.6 em = 7.2 px
  per character), and `Accept-Encoding` is 15 (108 px) — it would wrap in the very layout the bug
  was reported in. 16ch = 115 px fits every common request header name (`Accept-Encoding`,
  `Accept-Language`, `Content-Length`, `X-Correlation-Id`) and still leaves 141 px (55%) for
  values. `max()` inside `fit-content()` works in every shipped webview (Safari 11.1+), and jsdom
  returns the authored value `fit-content(max(40%, 16ch)) minmax(0, 1fr)` (probe in this worktree).
- **Alternatives considered**: a flat 50% cap (wastes width on short names in wide panels and
  still a percentage of an unknown width); a fixed 16ch cap (wraps long names early in wide
  panels).
