# Contract: `useFitHeight`

**File**: `frontend/src/shared/hooks/use-fit-height.ts` (domain-free; imports only `react`)

```ts
export function useFitHeight(ref: RefObject<HTMLTextAreaElement | null>, value: string, maxLines?: number): void
```

## Inputs

| Param | Meaning |
|---|---|
| `ref` | the textarea to size; a `null` `current` (not rendered, e.g. the locked-secret card) is a no-op |
| `value` | the textarea's current text; a change triggers a refit |
| `maxLines` | optional cap in lines; omitted means grow without limit |

## Behaviour

1. Fits in a `useLayoutEffect` (before paint) whenever `value` or `maxLines` changes, and again
   whenever the node's size changes (`ResizeObserver`; not created when `ResizeObserver` is
   undefined). The observer is disconnected on cleanup.
2. A fit: `style.height = 'auto'`; `content = scrollHeight`; read computed `lineHeight`,
   `paddingTop`, `paddingBottom`, `borderTopWidth`, `borderBottomWidth` as px numbers.
3. `cap = maxLines * lineHeight + paddingTop + paddingBottom` when `maxLines` is given and
   `lineHeight` parses to a finite number; otherwise there is no cap.
4. `capped = cap !== undefined && content - cap > 1`.
5. Sets `style.height = \`${(capped ? cap : content) + borderTop + borderBottom}px\``.
6. Sets `style.overflowX = 'hidden'` and `style.overflowY = capped ? 'auto' : 'hidden'`.

## Guarantees

- Never a horizontal scrollbar; a vertical one only when the content is more than 1 px taller
  than `maxLines` lines.
- With no `maxLines` and no border, the height equals `scrollHeight` — identical to
  `WrapTextCell`'s previous effect.
- Writes only the textarea's inline `height`, `overflowX`, `overflowY`; reads nothing outside the
  node.

## Consumers

| Consumer | Call | Expected |
|---|---|---|
| `VariableCard` → `NewVariable` | `useFitHeight(ref, value, MAX_LINES)` (`MAX_LINES = 6`) | 1–6 lines, then scrolls vertically |
| `VariableCard` → `DefinedVariable` | `useFitHeight(ref, draft, MAX_LINES)` | same; no-op while the vault is locked (no textarea) |
| `WrapTextCell` | `useFitHeight(ref, value)` | grows without limit, never scrolls (unchanged) |

## Test obligations (`use-fit-height.test.tsx`)

With `scrollHeight` stubbed and metrics as inline styles (`lineHeight: '18px'`, padding 4 px,
1 px border):

1. No cap: height = `scrollHeight + 2` px, `overflowY` and `overflowX` `hidden`.
2. Cap 6: content 116 → 118 px, hidden; content 117 → 119 px, hidden (tolerance); content 170 →
   118 px, `overflowY` `auto`, `overflowX` `hidden`.
3. Refits when `value` changes (grow then shrink).
4. Unparseable `lineHeight` with a cap → fits uncapped (no `NaN` height).
5. Null ref → no throw.
