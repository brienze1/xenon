# Research: Request Pane Content Never Overlaps (XEN-019)

All findings verified against `origin/main` at `db23057` on 2026-10-23.

## R1 — Root cause of the Scripts overlap

- **Decision**: The overlap is caused by the `min-height: 0` chain in
  `frontend/src/features/scripts/components/ScriptsEditor.module.css`, combined with the inner
  phase panel being a scroll container. Fix it by letting the flexible containers keep their
  automatic (content-based) minimum and by making the fixed rows non-shrinkable.
- **Facts**:
  - Structure: `.scripts` (`height: 100%`, flex column) → [`.tabs` (`flex: 1; min-height: 0`,
    a `Tabs.Root`) → [`.header` (phase `Tabs.List` + "Inherit parent scripts"), `.panel`
    (`flex: 1; min-height: 0`, a `Tabs.Panel`, which also gets `overflow: auto` from
    `shared/ui/tabs/Tabs.module.css`)] → `.phase` (`flex: 1; min-height: 0`) → [`.toolbar`,
    `.editor` (`flex: 1; min-height: 160px`)]], then `.hint` (`<p>`, no flex rule).
  - `.scripts` is exactly as tall as the request sub-tab panel. When that is shorter than the
    content, `.tabs` (min-height 0) shrinks to almost nothing; its header spills out of the
    collapsed box (overflow visible), and the hint — laid out right after `.tabs` — is drawn
    over the phase tabs and the inherit checkbox. That is precisely the screenshot in #26.
  - The request sub-tab panel (`RequestTabs.module.css .panel`: `flex: 1; min-height: 0;
    overflow: auto`) is already a scroll box, but since `.scripts` never grows past its height,
    nothing ever overflows it and it never scrolls.
  - The inner phase panel is a scroll container (`overflow: auto` from `Tabs.module.css`). A
    scroll container's automatic minimum size is 0, so removing `min-height: 0` alone is not
    enough: it also needs `overflow: visible`, so its minimum becomes its content (toolbar +
    160 px editor). The editor keeps its own internal scrolling (CodeMirror), so the phase panel
    never needs to scroll.
- **Consequences**: In `ScriptsEditor.module.css`: `.tabs`, `.panel` and `.phase` drop
  `min-height: 0`; `.panel` gets `overflow: visible`; `.header`, `.toolbar` and `.hint` get
  `flex: none`; `.toolbar` gets `flex-wrap: wrap` (the disabled-script notice). With
  `.tabs` refusing to shrink below its content, the hint is laid out after it, `.scripts`
  overflows, and the request sub-tab panel scrolls (FR-001–FR-003, FR-007). On a tall pane the
  heights are definite as before and the editor still fills the free height (FR-010).
- **Alternatives considered**: `.scripts { min-height: 100% }` instead of `height: 100%` — makes
  the editor's height indefinite, so CodeMirror's `height: 100%` stops filling the free height on
  tall panes (breaks FR-010); rejected. Moving the hint above the tabs — a visual change the
  product owner did not ask for (decision 3); rejected.

## R2 — Which other tabs are actually affected

- **Decision**: Change only what can overlap or clip; record the rest as verified unaffected.
- **Facts**:
  - **Request sub-tabs Params, Authorization, Headers, Body, Settings**: their content is
    normal block flow (tables, `PanelSection` with no height constraint) inside the request
    sub-tab scroll panel — they overflow and scroll today; no overlap. No change.
  - **Request builder chrome** (`RequestBuilder.module.css .builder`: `height: 100%`, flex
    column; `RequestTabs.module.css .tabs`: `flex: 1; min-height: 0`): when the pane is shorter
    than breadcrumb + URL bar + sub-tab row (~110 px; the 20 % floor is ~75–100 px at the
    960 × 600 minimum window), `.tabs` collapses and the sub-tab row is clipped by
    `SplitPane .pane { overflow: hidden }`. Fixed by R3 (200 px floor) and R4 (fallback scroll).
  - **Collection/folder tabs** (`ItemTab.module.css`): `.tab` is `overflow: auto`; the Scripts
    section panel has `min-height: 320px` and Documentation `min-height: 360px`. `.sections`
    (`min-height: 0`) may shrink, but nothing is laid out after it, so its overflowing panels
    only extend `.tab`'s scroll area — no overlap. Inside the Scripts section the
    `ScriptsEditor` overlap is fixed by R1 (same component). Folder shows the inherit checkbox,
    collection does not (`inheritScripts === null`). No change to `ItemTab.module.css` or
    `DocumentationPanel.module.css`.
  - **Environment and globals tabs** (`EnvironmentTab.module.css .tab`: flex column,
    `height: 100%`, `overflow: auto`): every child keeps its automatic minimum (no
    `min-height: 0`, none is a scroll container except `VariableTable .scroller`, which sits in
    an unconstrained column), so the tab scrolls and nothing overlaps. No change.
  - **Example tabs** (`ExampleTab.module.css`): the request part `.request` is a scroll box
    (`flex: 1; min-height: 0; overflow: auto`) — fine. The response part `.response`
    (`flex: 1; min-height: 0`, **not** a scroll box) → `.tabs` (`flex: 1; min-height: 0`) →
    `.panel` (`overflow: auto`) → `.body` (`height: 100%`) → views + editor (`min-height:
    120px`). At the minimum window size `.tabs` collapses and the response tab list and editor
    are clipped with no way to reach them. Fix: `.response` gets `overflow-x: hidden; overflow-y:
  auto` (same reasoning as R4; it has 16 px side padding); `.tabs`
    drops `min-height: 0`; `.panel` gets `min-height: 160px` (view switch 32 px + gap + 120 px
    editor — it is a scroll container, so it needs an explicit minimum).
- **Alternatives considered**: blanket-removing every `min-height: 0` in the frontend — many are
  load-bearing for scroll boxes (sidebar, console, response viewer); rejected.

## R3 — The 200 px request-pane floor

- **Decision**: `SplitPane.Root` gets an optional `minSize` prop (pixels, first pane). It is
  written as the CSS custom properties `--split-min-size` and `--split-max` on the root, and
  `SplitPane.module.css` applies `.column > .pane:first-child { min-height:
  min(var(--split-min-size, 0px), calc(var(--split-max, 1) * 100%)); }`. `LAYOUT_LIMITS.split` gains `minSize: 200`, and
  `RequestWorkspace` passes it only when `layout.split` is `'stacked'` (decision 4).
- **Facts**: `.pane:first-child` is `flex: 0 0 calc(var(--split-ratio) * 100%)`; the stored
  ratio is clamped to 0.2–0.8 by `clampLayout` and `SplitPane.Root`. A CSS minimum on the pane
  shows the stored ratio at no less than 200 px without rewriting it (FR-011), and applies to
  dragging and to keyboard resizing alike because both only change the ratio (FR-005).
  The cap `calc(var(--split-max) * 100%)` (the root also carries `--split-max`, the `max` ratio,
  when `minSize` is set) keeps the floor from taking more than 80 % of the area, so the response
  pane keeps its 20 % when the area is small — the minimum window at a high View → Zoom, since
  `app/zoom.ts` scales the body with CSS `zoom` and the floor is in UI pixels (spec
  Clarifications, edge case). The divider's `aria-valuenow` keeps reporting the stored ratio.
  Window minimum is 960 × 600 (`internal/domain/entity/window_state.go`); with the top bar
  (40 px), tab strip, footer (26 px) and the console drawer open (≥ 120 px), the split area is
  still ≈ 380 px, so the response pane keeps ≈ 180 px.
- **Alternatives considered**: clamping the ratio in JS from the measured root height — needs a
  `ResizeObserver` and rewrites the remembered ratio when the window shrinks (breaks FR-011);
  rejected. A `min-width` in side-by-side mode — declined by the product owner (decision 4).

## R4 — Fallback scrolling of the whole request builder (FR-006)

- **Decision**: `RequestBuilder.module.css .builder` gets `overflow-x: hidden; overflow-y: auto`
  (a non-visible `overflow-y` would turn `overflow-x` into `auto`; `hidden` guarantees no new
  horizontal scrollbar, FR-012 — nothing in the builder is wider than the pane: the URL field has
  `min-width: 0`, the breadcrumb `min-width: 0`, and the sub-tab panel is its own scroll box, so
  wide tables still scroll sideways inside it; focus rings stay inside the 16 px side padding of
  the URL bar and the sub-tab gutter of `Tabs.module.css`);
  `RequestTabs.module.css .tabs` drops `min-height: 0`; `RequestTabs.module.css .panel` gets
  `min-height: 80px`.
- **Facts**: `.tabs` is a column flex item holding the sub-tab row (automatic minimum = its
  height) and the panel (`flex: 1`, basis 0, scroll container → its minimum is the explicit
  80 px). With `min-height: auto`, `.tabs` never gets shorter than sub-tab row + 80 px; when the
  pinned rows plus banners need more than the pane has, the builder overflows and — now being a
  scroll box — scrolls as a whole (FR-006). Otherwise the panel alone scrolls (FR-004). At the
  200 px floor with no banner: breadcrumb (~20) + URL bar (56) + sub-tab row (~34) ≈ 110 px,
  leaving ≈ 90 px ≥ 80 px, so the builder does not scroll in the normal case. 80 px is the
  clarified value (spec Clarifications).
- **Alternatives considered**: always scrolling the whole builder — contradicts decision 1;
  rejected.

## R5 — Testing in jsdom

- **Decision**: Red-first Vitest cases that render the real components with their real CSS
  module source injected as a `<style>` element and assert computed longhands. A small
  test-support helper, `frontend/src/test/stylesheets.ts`, reads CSS files with
  `node:fs` `readFileSync` (paths relative to `frontend/`) and appends one
  `<style data-test-stylesheet>` per file; the global `src/test/setup.ts` `afterEach` removes them.
- **Facts** (jsdom 30.1.2, Vitest 5.0.3, probed on 2026-10-23):
  - `css.modules.classNameStrategy: 'non-scoped'` in `vite.config.ts` keeps class names
    unscoped (`.hint` stays `hint`), so the injected source matches the rendered classes.
  - Importing a stylesheet with `?raw` under Vitest yields a non-string (`TypeError: symbol is
    not a function` on concatenation); `readFileSync` works.
  - `getComputedStyle` expands `flex` (`flex: none` → `flexShrink "0"`, `flexGrow "0"`), reads
    `min-height` exactly (`"0px"`, `"160px"`, `"auto"` when unset, and raw `min(…)` /
    `var(…)` strings), and reads `overflow-y` when written as a longhand, but does **not**
    expand the `overflow` shorthand into `overflowY` (it reads `"visible"`; `overflow` reads
    the shorthand value).
- **Consequences**: new rules that tests check for vertical scrolling are written as the
  `overflow-y` longhand; the scripts phase panel's `overflow: visible` is asserted through
  `overflow`. jsdom has no layout, so the tests lock the CSS contract
  ([contracts/layout-contract.md](contracts/layout-contract.md)); the visual outcome is checked
  by eye in the running app ([quickstart.md](quickstart.md)).
- **Alternatives considered**: enabling CSS processing in Vitest — still no layout in jsdom,
  and a global test-config change; rejected. Playwright screenshots — no visual-regression
  harness exists in this repo; out of scope.

## R6 — No Godog feature

- **Decision**: No `.feature` file. Precedent: XEN-010 and XEN-015 (UI-only), justified in the
  plan's Complexity Tracking.
- **Facts**: No file under `internal/`, `cmd/`, `pkg/` or `test/integration/` changes; nothing
  in Go observes CSS layout.
