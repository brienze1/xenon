# Implementation Plan: Request Pane Content Never Overlaps

**Branch**: `xen-019` | **Date**: 2026-10-23 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/XEN-019-request-pane-overlap/spec.md`

## Summary

The Scripts sub-tab lets every container in its flex chain shrink to nothing (`min-height: 0`)
and its inner phase panel is a scroll container, so in a short request pane the phase-tab area
collapses and the hint is drawn on top of it; the request sub-tab panel, already a scroll box,
never overflows and never scrolls. The fix is CSS-first: containers keep their content-based
minimum, fixed rows get `flex: none`, and the scroll box under the sub-tab row finally scrolls.
Around it: a 200 px floor on the request pane in the stacked layout (a new `minSize` on the
shared `SplitPane`), a fallback that scrolls the whole request builder when banners leave the
sub-tab panel under 80 px, and the same treatment for the example tab's response part, which is
clipped today. Collection/folder, environment and globals tabs are verified unaffected beyond
the shared `ScriptsEditor` fix. Frontend only; no Go changes.

## Technical Context

**Language/Version**: TypeScript 5 (strict), React 19, CSS Modules

**Primary Dependencies**: none new

**Storage**: N/A (layout state shape and values unchanged — [data-model.md](data-model.md))

**Testing**: Vitest 5.0.3 + Testing Library in jsdom 30.1.2 (`frontend/vite.config.ts`
`test` block); real CSS injected as `<style>` via a test helper (research R5); `make validate`

**Target Platform**: Wails v3 webview — WKWebView (macOS), WebView2 (Windows), WebKitGTK (Linux)

**Project Type**: desktop app (Go backend + React UI)

**Performance Goals**: no new JS on resize — the floor and all fixes are pure CSS; dragging the
divider stays one style update per pointer move, as today

**Constraints**: taller panes must look identical (FR-010); the stored divider ratio is never
rewritten (FR-011); jsdom has no layout, so tests lock the CSS contract and the visual result is
checked by eye on macOS

**Scale/Scope**: 8 source files changed (1 TSX prop + 1 TSX caller + 1 TS constant + 5 CSS
modules), 1 new test helper, 5 test files extended

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first** — ✅ no new state; the remembered divider ratio is unchanged.
- [x] **II. Secrets** — ✅ no values displayed, logged or stored differently.
- [x] **III. Go owns logic** — ✅ pure presentation; no bindings, disk or network.
- [x] **IV. Architecture skills** — ✅ `SplitPane` stays domain-free in `shared/ui/split-pane/`
  (gains a generic `minSize` in px); the 200 px value lives with the other layout limits in
  `features/workspaces/lib/layout.ts`; `app/layout/RequestWorkspace.tsx` composes them. CSS
  changes stay in each component's colocated module. The test helper sits in `frontend/src/test/`
  next to `setup.ts` (test support, not product code).
- [x] **V. BDD-first** — ✅ with justification (Complexity Tracking): no Go behaviour changes, so
  no Godog scenario can observe this. The red-first tests are Vitest cases (the constitution's
  suite includes "frontend lint + typecheck + Vitest"), written and proven red before the change.
  `/create-tests` writes **no `.feature` file** for this task; it writes the red Vitest cases
  T003, T005, T007, T009, T011 (tasks.md). Only the browser's missing layout is worked around
  (real stylesheets injected); no component, store or `platform/` module is mocked beyond what
  the existing tests already mock.
- [x] **VI. Postman parity** — ✅ Postman keeps the URL bar and request tabs visible and scrolls
  the tab content when the request pane is short, and limits how far the divider moves; this
  matches it. Nothing beyond the PRD.
- [x] **VII. Cross-platform** — ✅ only standard CSS (`min()`, custom properties, flexbox
  minimums) supported by WebKit, WebKitGTK and Chromium-based WebView2; CI runs the suite on all
  three; manual check on macOS per Clarifications.

**Post-design re-check (after Phase 1)**: all seven still ✅ — one generic prop, no dependency,
the V justification below stands.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-019-request-pane-overlap/
├── plan.md              # This file
├── research.md          # Phase 0: root cause, affected tabs, floor, fallback, jsdom testability
├── data-model.md        # Phase 1: layout limits and the display rule (no stored-data change)
├── quickstart.md        # Phase 1: automated + by-eye verification
├── contracts/
│   └── layout-contract.md  # Phase 1: SplitPane.minSize, LAYOUT_LIMITS, CSS rule contract
├── checklists/
│   ├── requirements.md
│   └── ui-layout.md
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
frontend/src/
├── test/
│   ├── setup.ts                             # afterEach also removes injected test stylesheets
│   └── stylesheets.ts                       # NEW — injectStylesheets(...paths) for computed-style tests
├── shared/ui/split-pane/
│   ├── SplitPane.tsx                        # Root: optional minSize → --split-min-size
│   ├── SplitPane.module.css                 # .column > .pane:first-child min-height
│   └── SplitPane.test.tsx                   # + minSize cases
├── features/workspaces/lib/
│   └── layout.ts                            # LAYOUT_LIMITS.split.minSize = 200
├── app/layout/
│   ├── RequestWorkspace.tsx                 # passes minSize when stacked
│   └── RequestWorkspace.test.tsx            # + floor stacked / none side-by-side
├── features/scripts/components/
│   ├── ScriptsEditor.module.css             # root-cause fix (R1)
│   └── ScriptsEditor.test.tsx               # + layout-contract case
├── features/request-builder/components/
│   ├── RequestBuilder.module.css            # .builder overflow-y: auto (fallback)
│   ├── RequestTabs.module.css               # .tabs auto min; .panel min-height 80px
│   └── RequestBuilder.test.tsx              # + layout-contract case
└── features/examples/components/
    ├── ExampleTab.module.css                # .response scrolls; .tabs auto min; .panel 160px
    └── ExampleTab.test.tsx                  # + layout-contract case
```

No file under `internal/`, `cmd/`, `pkg/` or `test/integration/` changes. **No BDD feature
file**: nothing in Go observes CSS layout (Constitution V justification).

**Structure Decision**: frontend-only change in the existing `frontend/src` layout; every edit
is in the file that already owns the rule. Layer placement per `frontend-architecture`:
`shared/ui` (generic split pane), `features/*` (feature stylesheets and limits), `app/layout`
(composition).

### Design

Full rule list: [contracts/layout-contract.md](contracts/layout-contract.md).

**`ScriptsEditor.module.css`** (R1): `.header`, `.toolbar`, `.hint` → `flex: none`;
`.toolbar` → `flex-wrap: wrap`; `.tabs`, `.panel`, `.phase` → remove `min-height: 0`;
`.panel` → `overflow: visible`. `.scripts` keeps `height: 100%` (definite height chain → the
editor fills tall panes, FR-010). `.editor` keeps `flex: 1; min-height: 160px`. No TSX change;
the hint text stays (decision 3).

**`SplitPane`** (R3): `RootProps.minSize?: number`; when set, the root's inline style is
`{ '--split-ratio': value, '--split-min-size': `${minSize}px`, '--split-max': max }`, otherwise
only `--split-ratio`. `SplitPane.module.css` adds
`.column > .pane:first-child { min-height: min(var(--split-min-size, 0px), calc(var(--split-max, 1) * 100%)); }`. Ratio
clamping, `onRatioChange`, keyboard and `aria-value*` untouched (FR-011).

**`layout.ts`**: `split: { min: 0.2, max: 0.8, minSize: 200 }`. **`RequestWorkspace.tsx`**:
`minSize={layout.split === 'side-by-side' ? undefined : LAYOUT_LIMITS.split.minSize}`.

**`RequestBuilder.module.css` / `RequestTabs.module.css`** (R4): `.builder` → `overflow-x:
hidden; overflow-y: auto`; RequestTabs `.tabs` → remove `min-height: 0`; RequestTabs `.panel` → `min-height: 80px`
(keeps `flex: 1; overflow: auto; padding-block`).

**`ExampleTab.module.css`** (R2): `.response` → `overflow-x: hidden; overflow-y: auto`; `.tabs` → remove
`min-height: 0`; `.panel` → `min-height: 160px`.

**Test helper `frontend/src/test/stylesheets.ts`**: `injectStylesheets(...paths: string[])`
reads each path (relative to `frontend/`) with `readFileSync(path, 'utf8')` and appends one
`<style data-test-stylesheet>` per file to `document.head` in argument order; the global
`frontend/src/test/setup.ts` `afterEach` removes them (hooks cannot be registered inside a test). Tests call it with `Tabs.module.css`
first, then the component's stylesheet(s) (contract: cascade order).

### Requirement → coverage map

| Requirement | Change | Test |
|---|---|---|
| FR-001, FR-002, FR-003, FR-007 (Scripts) | ScriptsEditor.module.css | T003 |
| FR-004, FR-006 | RequestBuilder/RequestTabs.module.css | T005 |
| FR-005, FR-011 | SplitPane minSize + layout.ts + RequestWorkspace | T007, T009 |
| FR-008 | RequestWorkspace passes no minSize side-by-side | T009 |
| FR-009 (example) | ExampleTab.module.css | T011 |
| FR-009 (collection/folder, environment, globals) | ScriptsEditor fix (shared); others verified unchanged (research R2) | T003 + quickstart step 6 |
| FR-010 | definite height chain kept; no rule changed for tall panes | existing tests unchanged + quickstart step 7 |
| FR-005 cap / zoom | `--split-max` cap in the min-height | T007 + quickstart step 8 |
| FR-012 | `overflow-x: hidden` on the new scroll boxes; focus gutters kept | T005, T011 + quickstart step 9 |
| SC-004, SC-005 | — | T014 (by eye, screenshots), T015 (`make validate`) |

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| V. BDD-first: no Godog `.feature` for these requirements | The behaviour is CSS layout inside the webview; no Go code is involved, so no step in `test/integration/steps/` can observe it | A Godog scenario would assert nothing real; Vitest is the constitution's frontend gate and is written red-first here exactly like a feature file would be. The issue (brienze1/xenon#26) itself names a Vitest computed-style/class-contract check as the test. Precedent: XEN-010, XEN-015 |
