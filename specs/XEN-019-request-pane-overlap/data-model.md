# Data Model: Request Pane Content Never Overlaps (XEN-019)

No stored data changes. The workspace layout state (`LayoutState` in
`frontend/src/platform/mappers/ui-state.ts`, persisted through the UI-state service as
`split` / `split_ratio`) keeps its shape and values.

## Layout limits (constants, `features/workspaces/lib/layout.ts`)

| Name | Value | Meaning |
|---|---|---|
| `LAYOUT_LIMITS.split.min` | `0.2` | Smallest stored ratio (unchanged) |
| `LAYOUT_LIMITS.split.max` | `0.8` | Largest stored ratio (unchanged) |
| `LAYOUT_LIMITS.split.minSize` | `200` (px) | **New.** Smallest displayed height of the request pane in the stacked layout |

## Display rule

Displayed request-pane height (stacked) = `max(ratio × H, min(200 px, 0.8 × H))`, where `H` is
the split area's height in UI pixels (scaled by View → Zoom) and `ratio` the stored value
(0.2–0.8), so the response pane always keeps at least 20 %. The stored ratio is never rewritten by the
floor (FR-011). In the side-by-side layout the displayed width is `ratio × W`, as today.

## Fixed sizes referenced by the requirements

| Item | Minimum |
|---|---|
| Request sub-tab panel (under the sub-tab row) | 80 px (new) |
| Scripts editor | 160 px (existing) |
| Body editors | 160 px (existing) |
| Example response panel | 160 px (new; view switch + 120 px editor) |
| Example response editor | 120 px (existing) |
| Collection/folder Scripts section | 320 px (existing) |
| Documentation editor | 360 px (existing) |
