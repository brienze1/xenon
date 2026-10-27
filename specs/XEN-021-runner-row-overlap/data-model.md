# Data Model: Runner Result Rows Never Overlap When Expanded (XEN-021)

No stored data changes. Runs, run rows (`RunRowResult` in `frontend/src/platform/mappers/runner.ts`),
tests and console entries keep their shape and values; the Go side is untouched.

## List items (unchanged, `features/runner/lib/results.ts`)

| Item | Key | Content |
|---|---|---|
| Iteration heading | `iteration-<iteration>-<first row index>` | "Iteration N" + row count |
| Result row | `row-<row index>` | the row's line; when expanded, its details panel |

The key is what the virtualizer caches a measured size under (research R2).

## Sizes

| Item | Size | Source |
|---|---|---|
| Iteration heading | 28 px | `.iteration { height: 28px }`; estimate `ITERATION_HEADING_HEIGHT` |
| Collapsed row | 32 px | `.line { height: 32px }`; estimate `RESULT_ROW_HEIGHT` |
| Expanded row | 32 px + details panel + 4 px bottom margin (`--space-2`) | measured; first estimate `EXPANDED_ROW_HEIGHT` = 240 |
| Details panel | its content height, at most **320 px** (padding included), then scrolls | `.expansion { max-height: 320px; overflow: auto }` |

## Placement rule

`start(item i) = Σ measured size(item j) for j < i`, the measured size being the slot's border-box
height as rendered (estimate until first measured). Re-computed whenever a slot's size changes.
No gap and no overlap follow by construction (FR-004).

## State (unchanged)

- `expanded: ReadonlySet<number>` — row indexes the user expanded, local to `RowResults`; kept
  across live refreshes and filter changes as today.
- `filter`, `openIndex` — unchanged.
