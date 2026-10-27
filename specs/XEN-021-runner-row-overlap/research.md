# Research: Runner Result Rows Never Overlap When Expanded (XEN-021)

All findings verified against `origin/main` at `4e25bc5` on 2026-10-27, with
`@tanstack/react-virtual` / `@tanstack/virtual-core` 3.14.14 installed by `npm ci`.

## R1 — Root cause

- **Decision**: Two independent defects in `frontend/src/features/runner/components/`, fixed
  together: (a) the list gives every expanded row a fixed 240 px slot; (b) the expansion's
  children may shrink below their content height.
- **Facts**:
  - `RowResults` builds the virtualizer with `estimateSize: (index) => itemHeight(items[index],
    expanded)`, which returns `EXPANDED_ROW_HEIGHT = 240` for any expanded row. The slots carry
    neither `data-index` nor `ref={virtualizer.measureElement}`, so the estimate is final.
  - Each slot is `position: absolute` with inline `height: virtual.size`; `.row` is a flex column
    with `height: 100%`; `.expansion` is `flex: 1; min-height: 0; overflow: auto`. The row is
    therefore always 240 px; the next slot starts at `start + 240`.
  - `.expansion`'s children (the embedded `TestResultsTab` `.results` — itself `min-height: 0` —
    the script-errors `<ul>` and the console `<ul>`) have the default `flex-shrink: 1`. Lists with
    `display: flex` have an automatic minimum of their content, but `TestResultsTab .results`
    sets `min-height: 0`, so when the content is taller than the expansion the tests panel
    shrinks to less than its content and its overflowing test lines are painted over the console
    list laid out right after it. That is the screenshot in #34.
- **Alternatives considered**: only raising `EXPANDED_ROW_HEIGHT` — still wrong for short and long
  rows (gaps or overlap); rejected. Computing the height from counts of tests/console lines — test
  error messages and console lines wrap, so a count cannot predict the height; rejected.

## R2 — How the virtualizer measures (3.14.14)

- **Decision**: Measure every slot (rows *and* iteration headings) with
  `ref={virtualizer.measureElement}` and `data-index={virtual.index}`, and drop the inline
  `height`. Keep `estimateSize` (and `EXPANDED_ROW_HEIGHT = 240`) as the initial estimate only.
  Remove the `useEffect(() => virtualizer.measure(), [expanded, virtualizer])`.
- **Facts** (`node_modules/@tanstack/virtual-core/dist/esm/index.js`):
  - `measureElement(node)` reads the index from `indexAttribute` (default `data-index`; a missing
    one logs a warning and is ignored), observes the node with a `ResizeObserver`, and records
    its size. The default `measureElement` option uses the RO entry's `borderBoxSize`; without an
    entry it returns the **cached** size if the item's key has one, else `offsetHeight`.
  - Every later size change of an observed slot (expand/collapse, content arriving during a live
    run, the tests filter, a narrower list when the side panel opens) fires the RO → `resizeItem`
    → items below are re-placed (FR-004, FR-005, FR-006).
  - `virtualizer.measure()` clears `itemSizeCache` and falls back to estimates. With measured
    slots this would reset every expanded row to 240 px, and a slot whose size did not change gets
    no new RO entry to correct it — so the existing effect must go.
  - Sizes are cached by `getItemKey` (`row-<index>` / `iteration-<iteration>-<index>` from
    `features/runner/lib/results.ts`), so a row hidden by the results filter and shown again starts
    from its last measured size; `observe()` delivers an initial RO entry, correcting it.
  - Scroll anchoring: on a first measurement, an item whose top is above the scroll offset shifts
    `scrollTop` by the delta; on a re-measurement only items entirely above the fold do. This is
    the default (`shouldAdjustScrollPositionOnItemSizeChange` unset) and keeps what the user is
    looking at in place (spec Assumptions). No option change.
  - React 19 calls the stable `virtualizer.measureElement` ref once on mount and with `null` on
    unmount (which prunes disconnected nodes); no re-render churn.
- **Alternatives considered**: measuring only expanded rows — collapsed rows and headings would
  keep fixed inline heights while expanded ones are measured; two code paths for no gain, since
  collapsed rows measure to exactly their 32 px line. Rejected. Calling `measure()` on toggle —
  breaks other expanded rows (above); rejected.

## R3 — Intrinsic heights once the slot is no longer sized

- **Decision**: Every slot's content sets its own height. `.row` drops `height: 100%`; `.line`
  is already `height: 32px; flex: none` (collapsed row = 32 px, FR-007). `.iteration` changes
  `height: 100%` to `height: 28px` (= `ITERATION_HEADING_HEIGHT`, FR-007).
- **Facts**: `.slot` is `position: absolute; width: 100%` with no height, so `height: 100%` on a
  child resolves to `auto` (percent of an indefinite height) — `.iteration` would collapse to its
  text's line height without an explicit 28 px.

## R4 — The expansion panel

- **Decision**: `.expansion` loses `flex: 1; min-height: 0` and gets `max-height: 320px`; keeps
  `overflow: auto`, margin, padding, radius, background. A new rule `.expansion > * { flex: none; }`
  makes the error, tests panel, script errors and console non-shrinkable (FR-001, FR-002, FR-003).
- **Facts**: `global.css` sets `box-sizing: border-box` on every element, so the 320 px includes
  the panel's padding and excludes its bottom margin — as clarified. `TestResultsTab`'s
  `.results { min-height: 0 }` is harmless once its flex item cannot shrink, so
  `TestResultsTab.module.css` is **not** changed (FR-008: the response viewer's Tests tab keeps its
  layout). The "No tests" `EmptyState` root has `height: 100%`; in a panel with only a
  `max-height` (no definite height) that resolves to `auto`, so it keeps its compact natural height
  (edge case). No runner-specific class on `TestResultsTab` is needed.
- **Alternatives considered**: `max-height` on the whole slot — would cap the row's 32 px line
  into the scroll area too; rejected. A cap relative to the list height — product owner chose a
  fixed 320 px.

## R5 — Keyboard access to the capped panel

- **Decision**: The expansion `<div>` gets `role="group"`, `aria-label={`Details of ${name}`}`
  (the same `name` the chevron uses) and `tabIndex={0}`; CSS adds
  `.expansion:focus-visible { outline: none; box-shadow: var(--focus-ring); }` (the runner
  module's existing focus pattern). Focused, a scroll container scrolls with arrows, Page Up/Down,
  Home and End natively in WebKit, WebKitGTK and WebView2 (FR-009).
- **Facts**: No `eslint-plugin-jsx-a11y` in `frontend/eslint.config.js`, so `tabIndex` on a group
  is not flagged. Tab order: chevron → name button → (expanded) the tests filter chips come
  *inside* the panel, so the panel itself is the stop before them (it is their ancestor) — order
  is chevron, name, panel, filter chips, then the next row. `role="group"` (not `region`) because a
  landmark per expanded row would flood landmark navigation.

## R6 — Testability in jsdom

- **Decision**: Red-first Vitest cases in `RunResults.test.tsx`, as in XEN-019: DOM contract
  (`data-index`, no inline `height`, `role`/`aria-label`/`tabIndex`) and computed-style contract
  with the real stylesheet injected via `injectStylesheets('src/features/runner/components/RunResults.module.css')`
  from `frontend/src/test/stylesheets.ts`. The visual result is checked by eye on macOS.
- **Facts**: jsdom has no layout and no `ResizeObserver` (virtual-core then skips observing; the
  ref measurement falls back to `offsetHeight`, which the test file stubs to 600 for every
  element). With 600 px per item and a 600 px scroll element plus `overscan: 10`, the first 11
  items render — enough for every existing case (≤ 6 items). jsdom's `getComputedStyle` applies
  injected rules and returns declared values (`'320px'`, `'auto'`, `'0'`), which is what the
  contract asserts.
- **No Godog scenario**: no Go code changes; nothing in `test/integration/steps/` can observe CSS
  layout (Constitution V justification in plan.md; precedent XEN-010, XEN-015, XEN-019).
