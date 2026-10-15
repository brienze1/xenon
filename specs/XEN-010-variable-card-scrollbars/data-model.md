# Data Model: Variable Card Value Field Without Scrollbars

No persisted data, DTOs or Go entities change. The only "model" is the value box's sizing state,
derived on every fit (see [contracts/use-fit-height.md](contracts/use-fit-height.md)).

## Value box sizing state

| Field | Source | Notes |
|---|---|---|
| `content` | `node.scrollHeight` after `height: auto` | laid-out content + vertical padding, whole px |
| `lineHeight` | computed `line-height` | 18 px for both consumers (`.value`, `.cell`) |
| `paddingY` | computed `padding-top + padding-bottom` | card: 4 + 4 = 8 px; table cell: 6 + 6 = 12 px |
| `borderY` | computed `border-top-width + border-bottom-width` | card: 1 + 1 = 2 px; table cell: 0 |
| `maxLines` | caller | card: 6; table cell: none |
| `cap` | `maxLines × lineHeight + paddingY` | card: 6 × 18 + 8 = 116 px (118 px with border) |
| `capped` | `cap` defined and `content − cap > 1` | 1 px tolerance for `scrollHeight` rounding |
| `height` (inline) | `(capped ? cap : content) + borderY` px | `box-sizing: border-box` is global |
| `overflowX` (inline) | always `hidden` | FR-001 |
| `overflowY` (inline) | `capped ? 'auto' : 'hidden'` | FR-002, FR-006 |

## Transitions

```text
fit (layout effect on value/maxLines change, or ResizeObserver on width change)
  content ≤ cap + 1 (or no cap) → height = content + borderY, overflowY hidden   [grows/shrinks]
  content > cap + 1             → height = cap + borderY,     overflowY auto     [pinned at 6 lines]
```

Card examples (12 px text, 18 px lines, 8 px padding, 2 px border):

| Content | `scrollHeight` | Height | overflowY |
|---|---|---|---|
| empty / one line | 26 | 28 px | hidden |
| 3 lines | 62 | 64 px | hidden |
| 6 lines | 116 (or 117 rounded) | 118 / 119 px | hidden |
| 9 lines | 170 | 118 px | auto |
