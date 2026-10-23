# Contract: Stacked-layout CSS rules and the SplitPane floor (XEN-019)

The Vitest cases lock these rules (computed styles with the real CSS injected, see
[research R5](../research.md)). Anything not listed keeps its current value.

## `SplitPane.Root` — new optional prop

```ts
interface RootProps {
  // …existing: direction, ratio, onRatioChange, min, max, className, children
  /** Smallest size of the first pane in px, applied in the column direction; the ratio still applies above it. */
  minSize?: number
}
```

- When `minSize` is given, the root's inline style carries `--split-min-size: <minSize>px` and
  `--split-max: <max>` next to `--split-ratio`; when it is not, both properties are absent.
- `SplitPane.module.css`: `.column > .pane:first-child { min-height: min(var(--split-min-size, 0px), calc(var(--split-max, 1) * 100%)); }`.
- `ratio`, `onRatioChange`, `min`, `max`, keyboard steps and `aria-value*` are unchanged; the
  floor never changes the ratio passed to `onRatioChange`.

## `LAYOUT_LIMITS` (`features/workspaces/lib/layout.ts`)

```ts
split: { min: 0.2, max: 0.8, minSize: 200 }
```

`clampLayout` is unchanged (it clamps the ratio only).

## `RequestWorkspace`

Passes `minSize={LAYOUT_LIMITS.split.minSize}` when `layout.split === 'stacked'`, and no
`minSize` when `'side-by-side'`.

## CSS rules (computed values the tests assert)

| File · selector | Rule after the change | Asserted as |
|---|---|---|
| `ScriptsEditor.module.css` `.header` | `flex: none` | `flexShrink` `0` |
| `ScriptsEditor.module.css` `.tabs` | `min-height: auto` (replaces `0`) | `minHeight` `auto` |
| `ScriptsEditor.module.css` `.panel` | `min-height: auto` (replaces `0`); `overflow: visible` | `minHeight` `auto`; `overflow` `visible` |
| `ScriptsEditor.module.css` `.phase` | no `min-height` (delete the `0`; not a `Tabs` part) | `minHeight` `auto` |
| `ScriptsEditor.module.css` `.toolbar` | `flex: none; flex-wrap: wrap` | `flexShrink` `0`; `flexWrap` `wrap` |
| `ScriptsEditor.module.css` `.editor` | unchanged `flex: 1; min-height: 160px` | `minHeight` `160px` |
| `ScriptsEditor.module.css` `.hint` | `flex: none` | `flexShrink` `0` |
| `RequestBuilder.module.css` `.builder` | `overflow-x: hidden; overflow-y: auto` | `overflowX` `hidden`; `overflowY` `auto` |
| `RequestTabs.module.css` `.tabs` | `min-height: auto` (replaces `0`) | `minHeight` `auto` |
| `RequestTabs.module.css` `.panel` | `min-height: 80px` (keeps `overflow: auto`) | `minHeight` `80px` |
| `ExampleTab.module.css` `.response` | `overflow-x: hidden; overflow-y: auto` (drops nothing else) | `overflowX` `hidden`; `overflowY` `auto` |
| `ExampleTab.module.css` `.tabs` | `min-height: auto` (replaces `0`) | `minHeight` `auto` |
| `ExampleTab.module.css` `.panel` | `min-height: 160px` (keeps `overflow: auto`) | `minHeight` `160px` |

Why `auto` is written out: every `.tabs` above is also a `Tabs.Root` (`Tabs.module.css .root { min-height: 0 }`) and the Scripts `.panel` is also a `Tabs.Panel` (`.panel { min-height: 0 }`), so only an explicit `min-height: auto` in the feature stylesheet overrides them.

Cascade order: `ScriptsEditor.tsx`, `RequestTabs.tsx` and `ExampleTab.tsx` import `Tabs` before
their own stylesheet, so their single-class rules come after `Tabs.module.css` and win at equal
specificity (the existing `padding-top` overrides already rely on this). Tests inject
`Tabs.module.css` first, then the component's stylesheet, to reproduce that order.

## Verified unchanged (no rule edits)

`ItemTab.module.css`, `DocumentationPanel.module.css`, `EnvironmentTab.module.css`,
`VariableTable.module.css`, `BodyPanel.module.css`, `ParamsPanel.module.css`,
`HeadersPanel.module.css`, `SettingsPanel.module.css`, `AuthForm.module.css`,
`Tabs.module.css` — see research R2.
