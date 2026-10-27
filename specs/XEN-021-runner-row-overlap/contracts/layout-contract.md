# Layout Contract: Runner Results List (XEN-021)

The DOM and CSS contract the red Vitest cases lock. Paths are under
`frontend/src/features/runner/components/`.

## DOM (`RunResults.tsx`)

| Element | Contract |
|---|---|
| Each item of the list labelled "Run results" (`role="listitem"`, class `slot`) | has `data-index` = its virtual index (`"0"`, `"1"`, …); inline `style.height` is `''`; inline `style.transform` is `translateY(<start>px)` |
| Expanded row's details panel (class `expansion`) | `role="group"`, accessible name `Details of <name>` where `<name>` is `row.name` or `row.name · row.aliasName`; `tabIndex` 0 |
| Details panel children, in order | row error (`role="alert"`, if any) → tests slot → "Script errors" list (if any) → "Console" list (if any) — unchanged |
| Tab order within an expanded row | chevron → name button → details panel → controls inside the panel (tests filter chips) |

`RowResults` has no effect that calls `virtualizer.measure()`.

## CSS (`RunResults.module.css`), computed with the stylesheet injected

| Selector / element | Property | Value |
|---|---|---|
| `.row` | `height` | not set (computed `''`/`auto`, never `100%`) |
| `.line` | `height`, `flex-shrink` | `32px`, `0` (unchanged) |
| `.iteration` | `height` | `28px` |
| `.expansion` | `max-height` | `320px` |
| `.expansion` | `overflow-y` | `auto` (from the unchanged `overflow: auto`) |
| `.expansion` | `flex-grow` / `min-height` | `0` / not `0px` (the `flex: 1; min-height: 0` pair is gone) |
| every direct child of `.expansion` | `flex-shrink` | `0` |
| `.expansion:focus-visible` | `outline` / `box-shadow` | `none` / `var(--focus-ring)` |

## Unchanged

- `frontend/src/features/response-viewer/components/TestResultsTab.tsx` and `.module.css`.
- `frontend/src/app/layout/runner-slots.tsx`.
- `RunRowDetail.tsx`, `PerformanceResults.tsx` and their styles.
- `RESULT_ROW_HEIGHT` (32), `ITERATION_HEADING_HEIGHT` (28), `EXPANDED_ROW_HEIGHT` (240, now the
  first estimate only), `overscan` (10), `getItemKey`.
