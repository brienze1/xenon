---

description: "Task list for XEN-019 — request pane content never overlaps"
---

# Tasks: Request Pane Content Never Overlaps

**Input**: Design documents from `specs/XEN-019-request-pane-overlap/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/layout-contract.md](contracts/layout-contract.md),
[quickstart.md](quickstart.md)

**Tests**: MANDATORY and red-first (constitution V). This task is frontend-only, so there is **no
Godog `.feature`** (plan Complexity Tracking); the red-first tests are the Vitest cases T003,
T005, T007, T009 and T011, which `/create-tests` writes and proves red. Each asserts the
computed styles of the real components with their real stylesheets injected (research R5).

**Organization**: Tasks are grouped by user story (spec.md) so each story can be implemented
and verified on its own.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: US1 Scripts legible · US2 URL bar and sub-tab row reachable · US3 other stacked tabs
- All paths are repository-relative; frontend commands run in `frontend/`

---

## Phase 1: Setup

**Purpose**: A fresh worktree that can run Vitest.

- [x] T001 Install frontend dependencies and generate the Wails bindings: `cd frontend && npm ci`, then `make bindings` from the repo root; confirm `npx vitest run src/features/scripts` is green on the untouched code

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The test helper every red test uses.

- [x] T002 Create `frontend/src/test/stylesheets.ts` exporting `injectStylesheets(...paths: string[]): void` — for each path (relative to `frontend/`, e.g. `'src/shared/ui/tabs/Tabs.module.css'`) read it with `readFileSync(path, 'utf8')` from `node:fs` and append a `<style data-test-stylesheet>` element with that text to `document.head`, in argument order. In `frontend/src/test/setup.ts`, extend the existing `afterEach` to also remove every `style[data-test-stylesheet]` (Vitest hooks cannot be registered from inside a test body, so cleanup lives in the global setup). In `frontend/eslint.config.js`, add the boundaries file category `test` (`src/**/*.test.{ts,tsx}`) and a policy allowing it to import the `test-support` element — test files only; product code stays forbidden (done with the red tests). Do not import stylesheets with `?raw` (yields a non-string under Vitest — Phase 0 finding R5)

**Checkpoint**: helper ready — story phases can start.

---

## Phase 3: User Story 1 — Scripts tab stays legible in a short request pane (Priority: P1) 🎯 MVP

**Goal**: The Scripts editor's fixed rows never shrink, its flexible containers keep their
content minimum, and the content overflows into the request sub-tab panel's scroll box instead
of being drawn over itself (FR-001, FR-002, FR-003, FR-007, FR-010).

**Independent Test**: T003 green; quickstart by-eye step 1.

### Tests for User Story 1 (red first)

- [x] T003 [US1] Add a case "keeps its rows from overlapping in a short pane" to `frontend/src/features/scripts/components/ScriptsEditor.test.tsx`: `injectStylesheets('src/shared/ui/tabs/Tabs.module.css', 'src/features/scripts/components/ScriptsEditor.module.css')`, render the existing `Harness` with `inherit={true}`, then assert with `getComputedStyle`: the parent of the "Script phase" tablist (`.header`) has `flexShrink` `'0'`; the hint paragraph (`screen.getByText(/for the API/).closest('p')`) has `flexShrink` `'0'`; `screen.getByRole('button', { name: 'Snippets for the Pre-request script' }).closest('.toolbar')` has `flexShrink` `'0'` and `flexWrap` `'wrap'`; the `.tabs`, visible `.panel` and `.phase` elements (via `container.querySelector`) have `minHeight` `'auto'`; the visible `.panel` has `overflow` `'visible'`; `.editor` has `minHeight` `'160px'`. Run it and confirm it is red (today `.tabs`/`.panel`/`.phase` read `'0px'` and `.hint`/`.header` read `flexShrink` `'1'`)

### Implementation for User Story 1

- [x] T004 [US1] In `frontend/src/features/scripts/components/ScriptsEditor.module.css`: add `flex: none` to `.header`, `.toolbar` and `.hint`; add `flex-wrap: wrap` to `.toolbar`; change `min-height: 0` to `min-height: auto` on `.tabs` and `.panel` (explicit, because these elements are also `Tabs.Root` / `Tabs.Panel`, whose own `min-height: 0` would otherwise stay in force) and remove `min-height: 0` from `.phase`; add `overflow: visible` to `.panel`. Keep `.scripts { height: 100%; min-height: 0 }` and `.editor { flex: 1; min-height: 160px }` unchanged (contract table). No change to `ScriptsEditor.tsx`. T003 and every existing case in `ScriptsEditor.test.tsx` green

**Checkpoint**: US1 complete — the reported overlap is gone on request, folder and collection Scripts (same component).

---

## Phase 4: User Story 2 — URL bar and sub-tab row always reachable (Priority: P1)

**Goal**: In the stacked layout the request pane is never shorter than 200 px (capped at 80 % of the split area); the breadcrumb,
URL bar and sub-tab row stay pinned while the sub-tab panel (min 64 px) scrolls, and the whole
builder scrolls only when banners leave the panel less than that (FR-004, FR-005, FR-006,
FR-008, FR-011).

**Independent Test**: T005, T007, T009 green; quickstart by-eye steps 1, 3, 4, 5.

### Tests for User Story 2 (red first)

- [x] T005 [P] [US2] Add a case "pins the URL bar and sub-tabs and falls back to scrolling the builder" to `frontend/src/features/request-builder/components/RequestBuilder.test.tsx`: `injectStylesheets('src/shared/ui/tabs/Tabs.module.css', 'src/features/request-builder/components/RequestBuilder.module.css', 'src/features/request-builder/components/RequestTabs.module.css')`, render the builder the way the file's existing cases do, then assert: the `section` labelled "Request" has `overflowY` `'auto'` and `overflowX` `'hidden'`; the element with class `tabs` that contains the "Request" tablist has `minHeight` `'auto'`; the visible sub-tab panel (`role="tabpanel"`) has `minHeight` `'64px'` and `contain` `'size'`. Run it and confirm it is red (today `overflowY` `'visible'`, `.tabs` `'0px'`, panel `'0px'`)
- [x] T007 [P] [US2] Add cases to `frontend/src/shared/ui/split-pane/SplitPane.test.tsx`: (a) with `minSize={200}` the root (`screen.getByRole('separator').parentElement`) has inline style `--split-min-size` `'200px'` and `--split-max` `'0.8'` (`style.getPropertyValue(…)`), and without `minSize` both are `''`; (b) with `injectStylesheets('src/shared/ui/split-pane/SplitPane.module.css')`, `direction="column"` and `minSize={200}`, the first pane's computed `minHeight` contains `--split-min-size` (jsdom normalises `min()`, so do not match the whole string) while in `direction="row"` it stays `'0px'`; (c) dragging/keyboard with `minSize` set still reports the clamped ratio to `onRatioChange` exactly as without it (reuse the existing keyboard case's steps: Home → `0.2`). Extend the file's `Example` helper with an optional `minSize` prop passed to `SplitPane.Root`. Run and confirm red (`minSize` is not a prop yet → typecheck/assertion failure)
- [x] T009 [P] [US2] Add cases to `frontend/src/app/layout/RequestWorkspace.test.tsx`: in the stacked layout the split root carries `--split-min-size` `'200px'`; in the side-by-side layout the property is `''`. Use the file's existing setup for each layout and locate the root as `screen.getByRole('separator', { name: 'Resize request and response' }).parentElement`. Run and confirm red

### Implementation for User Story 2

- [x] T006 [US2] In `frontend/src/features/request-builder/components/RequestBuilder.module.css` add `overflow-x: hidden; overflow-y: auto` to `.builder`; in `frontend/src/features/request-builder/components/RequestTabs.module.css` change `.tabs`'s `min-height: 0` to `min-height: auto` (it is also a `Tabs.Root`) and change `.panel`'s `min-height: 0` to `min-height: 64px` and add `contain: size` (so the panel's content does not raise the tabs' minimum — research R4) (keep `flex: 1; overflow: auto; padding-block: var(--space-4)`). T005 and the existing `RequestBuilder.test.tsx` cases green
- [x] T008 [US2] In `frontend/src/shared/ui/split-pane/SplitPane.tsx` add `minSize?: number` to `RootProps` with the doc comment "Smallest size of the first pane in px, applied in the column direction; the ratio still applies above it."; destructure it in `Root` and set the root's style to `{ '--split-ratio': value, ...(minSize === undefined ? {} : { '--split-min-size': `${minSize}px`, '--split-max': max }) } as CSSProperties`; do not change `clampRatio`, `actions`, `meta`, the divider or `aria-value*`. In `frontend/src/shared/ui/split-pane/SplitPane.module.css` add, after `.pane:first-child`, the rule `.column > .pane:first-child { min-height: min(var(--split-min-size, 0px), calc(var(--split-max, 1) * 100%)); }`. T007 and the existing `SplitPane.test.tsx` cases green
- [x] T010 [US2] In `frontend/src/features/workspaces/lib/layout.ts` change `split` to `{ min: 0.2, max: 0.8, minSize: 200 }` (leave `clampLayout` unchanged); in `frontend/src/app/layout/RequestWorkspace.tsx` pass `minSize={layout.split === 'side-by-side' ? undefined : LAYOUT_LIMITS.split.minSize}` to `SplitPane.Root`. T009 and the existing `RequestWorkspace.test.tsx` cases green (depends on T008)

**Checkpoint**: US2 complete — the request pane keeps its floor and nothing is clipped.

---

## Phase 5: User Story 3 — Every other stacked tab behaves the same (Priority: P2)

**Goal**: The example tab's response part scrolls instead of clipping; collection/folder,
environment and globals tabs are confirmed unaffected beyond the shared Scripts fix (FR-009).

**Independent Test**: T011 green, T013 confirms; quickstart by-eye step 6.

### Tests for User Story 3 (red first)

- [x] T011 [P] [US3] Add a case "scrolls the example response instead of clipping it" to `frontend/src/features/examples/components/ExampleTab.test.tsx`: `injectStylesheets('src/shared/ui/tabs/Tabs.module.css', 'src/features/examples/components/ExampleTab.module.css')`, render an example tab the way the file's existing cases do, then assert: the section labelled "Example Response" has `overflowY` `'auto'` and `overflowX` `'hidden'`; the element with class `tabs` containing the "Example response" tablist has `minHeight` `'auto'`; the visible response tabpanel has `minHeight` `'160px'`. Run and confirm red

### Implementation for User Story 3

- [x] T012 [US3] In `frontend/src/features/examples/components/ExampleTab.module.css` add `overflow-x: hidden; overflow-y: auto` to `.response`; change `.tabs`'s `min-height: 0` to `min-height: auto` (it is also a `Tabs.Root`); change `.panel`'s `min-height: 0` to `min-height: 160px` (keep `flex: 1; padding-top; overflow: auto`). T011 and the existing `ExampleTab.test.tsx` cases green
- [x] T013 [US3] Verify the tabs that Phase 0 finding R2 lists as unaffected stay unchanged: `git diff --stat origin/main -- frontend/src/features/item-settings frontend/src/features/documentation frontend/src/features/environments frontend/src/features/auth frontend/src/shared/ui/tabs` prints nothing, and `npx vitest run src/features/item-settings src/features/documentation src/features/environments` is green (expected outcome decided: no edits there; the collection/folder Scripts section is fixed by T004)

**Checkpoint**: all stories complete.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T014 By-eye check on macOS per [quickstart.md](quickstart.md) steps 1–9 (running app; window sizes, zoom and divider positions as each step states); capture before (`origin/main`) and after screenshots of steps 1 and 7 for the pull request description. *Done 2026-10-23 in Chromium (Playwright) against the Vite dev server, at the viewport sizes listed, with the workspace and tab state set through the app's own stores (no Go backend in a browser), plus `origin/main` served the same way for the before shots; screenshots in `screenshots/`. Steps 1–5 and 7–9 checked on request tabs (0 overlaps measured on all 6 sub-tabs; tall-pane before/after byte-identical). Step 6 (example, collection/folder, environment, globals tabs) needs backend data and was not checked by eye — covered by T011/T013 and Phase 0 finding R2. Found and fixed here: the sub-tab panel needed `contain: size`, and the measured pinned rows (128.5 px) set the panel minimum to 64 px.*
- [ ] T015 Run `cd frontend && npm run lint && npm run typecheck && npm test`, then `make validate` from the repo root; all green
- [ ] T016 Run `/refactor-diff` until it reports clean, then re-run T015

---

## Dependencies & Execution Order

- **T001 → T002** → story phases.
- **US1** (T003 → T004), **US2** and **US3** are independent of each other (different files).
- Inside **US2**: T005 → T006; T007 → T008 → T010; T009 → T010.
- Inside **US3**: T011 → T012; T013 after T004 and T012.
- **Polish** (T014–T016) after all stories.

## Parallel Example

```text
# After T002, the red tests touch different files:
T003 ScriptsEditor.test.tsx   T005 RequestBuilder.test.tsx   T007 SplitPane.test.tsx
T009 RequestWorkspace.test.tsx   T011 ExampleTab.test.tsx
```

## Implementation Strategy

- **MVP**: Phase 1–3 (US1) removes the reported overlap on every Scripts editor.
- **Then** US2 (floor + pinned chrome + fallback) and US3 (example response), each verified by its
  own tests; finish with the by-eye check, `make validate` and `/refactor-diff`.
