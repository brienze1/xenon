---

description: "Task list for XEN-021 — runner result rows never overlap when expanded"
---

# Tasks: Runner Result Rows Never Overlap When Expanded

**Input**: Design documents from `specs/XEN-021-runner-row-overlap/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/layout-contract.md](contracts/layout-contract.md),
[quickstart.md](quickstart.md)

**Tests**: MANDATORY and red-first (constitution V). This task is frontend-only, so there is **no
Godog `.feature`** (plan Complexity Tracking); the red-first tests are the Vitest cases T002, T004,
T006 and T008 in `frontend/src/features/runner/components/RunResults.test.tsx`, which
`/create-tests` writes and proves red. They assert the DOM contract and, with the real stylesheet
injected, the computed-style contract (research R6).

**Organization**: Tasks are grouped by user story (spec.md) so each story can be implemented and
verified on its own.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: US1 expanded row legible · US2 rows below make room · US3 live and past runs
- All paths are repository-relative; frontend commands run in `frontend/`
- Every test and implementation task edits one of the same three files
  (`RunResults.test.tsx`, `RunResults.tsx`, `RunResults.module.css`), so no task is `[P]`

---

## Phase 1: Setup

**Purpose**: A fresh worktree that can run Vitest.

- [x] T001 Install frontend dependencies and generate the Wails bindings: `cd frontend && npm ci`, then `make bindings` from the repo root; confirm `npx vitest run src/features/runner/components/RunResults.test.tsx` is green on the untouched code

---

## Phase 2: Foundational (Blocking Prerequisites)

None. The test helper `injectStylesheets` already exists in `frontend/src/test/stylesheets.ts`
(XEN-019) and `*.test.tsx` files may already import it (`frontend/eslint.config.js` boundaries);
no shared code is added.

---

## Phase 3: User Story 1 — Expanded row shows every part legibly (Priority: P1) 🎯 MVP

**Goal**: The details panel's parts never shrink, the panel grows to its content up to 320 px and
then scrolls inside itself, and it is keyboard-focusable (FR-001, FR-002, FR-003, FR-009).

**Independent Test**: T002 and T004 green; quickstart by-eye steps 1, 3 and 8.

### Tests for User Story 1 (red first)

- [x] T002 [US1] Add a case "keeps an expanded row's parts from shrinking and caps the panel at 320 px" to `frontend/src/features/runner/components/RunResults.test.tsx`: import `injectStylesheets` from `@/test/stylesheets` and call `injectStylesheets('src/features/runner/components/RunResults.module.css')`; seed a run with one row (`runRow(0, { name: 'Login', error: 'timeout', tests: [one passed test], scriptErrors: [one], console: [one entry] })` — build the console entry and script error like the existing expand case); render, click `Details of Login`; take the panel as `screen.getByRole('button', { name: 'Details of Login' }).closest('[role="listitem"]')!.querySelector('.expansion')` (Vitest keeps CSS-module class names unscoped — `frontend/vite.config.ts` `classNameStrategy: 'non-scoped'` — and the class keeps this case independent of T005); assert `getComputedStyle(panel)`: `maxHeight` `'320px'`, `overflow` `'auto'` (jsdom does not expand the `overflow` shorthand into `overflowY`), `flexGrow` not `'1'`, `minHeight` not `'0px'`; and for every `panel.children` element `getComputedStyle(child).flexShrink` is `'0'`. Run it and confirm it is red (today `maxHeight` is not `'320px'`, `flexGrow` is `'1'`, `minHeight` is `'0px'` and no child declares `flexShrink` `'0'`)
- [x] T004 [US1] Add a case "lets the keyboard reach and scroll an expanded row's details" to `frontend/src/features/runner/components/RunResults.test.tsx`: seed one row `runRow(0, { name: 'Login', aliasName: 'mx-test-user' })`, render, click the chevron `Details of Login · mx-test-user`; assert `screen.getByRole('group', { name: 'Details of Login · mx-test-user' })` exists, has attribute `tabindex` `'0'`, and contains the tests slot text (`Tests:`); then focus the row's name button (`within(row).getByRole('button', { name: 'Login · mx-test-user' })`), `await userEvent.tab()`, and assert the group has focus. Run and confirm red (no such group today)

### Implementation for User Story 1

- [x] T003 [US1] In `frontend/src/features/runner/components/RunResults.module.css`: in `.expansion` remove `flex: 1` and `min-height: 0` and add `max-height: 320px` (keep `display: flex; flex-direction: column; gap; margin; padding; overflow: auto; border-radius; background`); add the rule `.expansion > * { flex: none; }` right after `.expansion`. Do not touch `frontend/src/features/response-viewer/components/TestResultsTab.module.css` (Phase 0 finding R4). T002 and every existing case green (depends on T002)
- [x] T005 [US1] In `frontend/src/features/runner/components/RunResults.tsx` `ResultRow`, give the expansion `<div className={styles.expansion}>` the props `role="group"`, `aria-label={`Details of ${name}`}` and `tabIndex={0}` (children and their order unchanged); in `RunResults.module.css` add `.expansion:focus-visible` to the existing `.chip:focus-visible, .chevron:focus-visible, .name:focus-visible` selector group. T004 and every existing case green (depends on T003, T004)

**Checkpoint**: US1 complete — an expanded row's parts no longer paint over each other; long rows scroll in their panel.

---

## Phase 4: User Story 2 — Rows below make room (Priority: P1)

**Goal**: Every slot is measured by the virtualizer and sized by its own content, so each item
starts where the previous one ends, also with several rows expanded and across filter changes
(FR-004, FR-005, FR-007).

**Independent Test**: T006 green; quickstart by-eye steps 2, 5, 6 and 7.

### Tests for User Story 2 (red first)

- [x] T006 [US2] Add a case "measures every row instead of giving it a fixed height" to `frontend/src/features/runner/components/RunResults.test.tsx`: `injectStylesheets('src/features/runner/components/RunResults.module.css')`; seed a run with rows `runRow(0, { name: 'Login' })`, `runRow(1, { name: 'Get order' })`, `runRow(2, { name: 'Login', iteration: 1 })`; render; expand `Details of Login` (first) and `Details of Get order`; then for every `within(results()).getAllByRole('listitem')` assert `item.dataset.index` equals its position as a string (`'0'`, `'1'`, …), `item.style.height` is `''` and `item.style.transform` starts with `'translateY('`; assert the first item's `p` (the iteration heading) has computed `height` `'28px'`; assert each row element (`item.querySelector('.row')` for the row items) has computed `height` not `'100%'`. Run and confirm red (today every slot has an inline `height` and no `data-index`, `.iteration` is `'100%'`, `.row` is `'100%'`)

### Implementation for User Story 2

- [x] T007 [US2] In `frontend/src/features/runner/components/RunResults.tsx` `RowResults`: delete the `useEffect(() => { virtualizer.measure() }, [expanded, virtualizer])` block and remove `useEffect` from the `react` import (its only use); on each slot `<div role="listitem" className={styles.slot}>` add `data-index={virtual.index}` and `ref={virtualizer.measureElement}` and change its `style` to `{ transform: `translateY(${virtual.start}px)` }`; change the `EXPANDED_ROW_HEIGHT` doc comment to `/** First estimate of an expanded row before it is measured; the rendered row sets its real height. */` (value stays 240; `estimateSize`, `itemHeight`, `getItemKey`, `overscan` unchanged). In `RunResults.module.css`: remove `height: 100%` from `.row`; change `.iteration`'s `height: 100%` to `height: 28px`. T006 and every existing case green, including "renders only the visible rows of 10,000 results" (depends on T005, T006)

**Checkpoint**: US2 complete — the list places items by their real heights.

---

## Phase 5: User Story 3 — Same behaviour on live and past runs (Priority: P2)

**Goal**: The measured layout holds when the run is refreshed while in progress and when a past
run is shown (FR-006). The mechanism is US2's; this phase locks it for a changing run.

**Independent Test**: T008 green; quickstart by-eye steps 4 and 6.

### Tests for User Story 3 (red first)

- [x] T008 [US3] Add a case "keeps an expanded row expanded and measured while a live run refreshes" to `frontend/src/features/runner/components/RunResults.test.tsx`: seed `run({ status: 'running', rows: [runRow(0, { name: 'Login', console: [entry 'one'] })] })`, render with `const { rerender } = render(...)`, expand `Details of Login`; then `rerender` with a new run object (same id) whose `rows` are `[runRow(0, { name: 'Login', console: [entry 'one', entry 'two'] }), runRow(1, { name: 'Get order' })]`; assert the `Details of Login` group still exists and its "Console" list has 2 items, and every direct child of the "Run results" list (`Array.from(results().children)` — `getAllByRole('listitem')` would also match the console's own `<li>` items) has a `data-index` and `style.height` `''`. Run and confirm red (no `data-index` / inline height today, and no group before T005)

### Implementation for User Story 3

- [x] T009 [US3] No code change: T008 is satisfied by T005 and T007 (the `expanded` set is keyed by row index and survives the refresh today; measuring follows the slot's size). Run `npx vitest run src/features/runner/components/RunResults.test.tsx` and confirm T008 and every case green (depends on T007, T008)

**Checkpoint**: all stories complete.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T010 Verify FR-008: `git diff --stat origin/main -- frontend/src/features/response-viewer frontend/src/app/layout/runner-slots.tsx frontend/src/features/runner/components/RunRowDetail.tsx frontend/src/features/runner/components/RunRowDetail.module.css frontend/src/features/runner/components/PerformanceResults.tsx` prints nothing, and `npx vitest run src/features/response-viewer` is green (expected outcome decided: no edits there)
- [x] T011 By-eye check on macOS in `wails3 dev` per [quickstart.md](quickstart.md) steps 1–8; capture before (`origin/main`) and after screenshots of steps 1 and 3 for the pull request description (step 9) *Done 2026-10-27 in Chromium (Playwright) against a throwaway Vite harness that renders the real `RunResults` (and, for the before shots, the `origin/main` copy) with the app's global styles, `RUNNER_SLOTS` and a 50-row fixture run — a real run needs the Go backend, which a browser does not have. Measured: before, the 3-test/7-line row is a fixed 240 px with the test list (y 262–348) drawn over the console (y 280–418); after, 0 overlaps inside every details panel, headings 28 px, collapsed rows 32 px, every item starts within 0.5 px of the previous one with rows 1, 3 and 5 expanded and at 960 × 600 (steps 1, 2, 7); the 40-line panel is 320 px tall with 940 px of content and scrolls to line 40 (step 3); Tab from the name button focuses the panel with the focus ring and End scrolls it to its last line without moving the list (step 8). Screenshots in `screenshots/`. Steps 4 (live run) and 6 (opened from Runs, side panel) need the backend and were not checked by eye — covered by T008 (refresh keeps the row expanded and measured) and by the narrower-width re-measure above.*
- [x] T012 Run `cd frontend && npm run lint && npm run typecheck && npm test`, then `make validate` from the repo root; all green *(done: 1665/1665 Vitest, lint and typecheck clean, `make validate` exit 0)*
- [x] T013 Run `/refactor-diff` until it reports clean, then re-run T012 *(clean on the first pass: the diff has no `.go` source files in its scope)*

---

## Dependencies & Execution Order

- **T001** → everything.
- **US1**: T002 → T003; T004 → T005 (T005 after T003, same CSS file).
- **US2**: T006 → T007 (T007 after T005, same TSX file).
- **US3**: T008 → T009 (after T007).
- **Polish** (T010–T013) after all stories.

All test tasks edit `RunResults.test.tsx` and all implementation tasks edit `RunResults.tsx` /
`RunResults.module.css`, so they run in ID order. `/create-tests` writes T002, T004, T006 and T008
together and proves each red.

## Parallel Example

```text
# No [P] tasks: one test file and one component. The only independent work is verification:
T010 (git diff + response-viewer tests)   T011 (by-eye check)   — after T009
```

## Implementation Strategy

- **MVP**: Phase 1 + US1 (T001–T005) removes the reported overlap inside an expanded row.
- **Then** US2 (T006–T007) removes the overlap/gaps between rows; US3 (T008–T009) locks live runs;
  finish with T010–T013.
