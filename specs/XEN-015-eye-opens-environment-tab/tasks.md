---

description: "Task list for XEN-015 — the eye next to the switcher opens the environment tab"
---

# Tasks: The Eye Opens the Environment Tab

**Input**: Design documents from `specs/XEN-015-eye-opens-environment-tab/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/eye-button.md](contracts/eye-button.md),
[quickstart.md](quickstart.md)

**Tests**: MANDATORY and red-first (constitution V). No Go behaviour changes, so there is no
Godog `.feature` (plan → Complexity Tracking); the red-first tests are Vitest cases T002, T003 and T006,
which `/create-tests` writes and proves red. Each test task ends by running its file and
confirming the new cases **fail** on the untouched code.

**Organization**: grouped by user story (spec.md: US1 P1 eye opens the active environment's
tab, US2 P2 eye opens Globals with no active environment, US3 P3 the Quick Look is gone).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel (different files, no dependency on an unfinished task)
- **[Story]**: US1 / US2 / US3

## Shared test facts

- `EnvironmentSwitcher.test.tsx` fixtures: `dev` = `e1`, `hml` = `e2`, `prod` = `e3`; module-level
  `onOpenGlobals` and `onManage` mocks. Add `const onOpenEnvironment = vi.fn()` next to them and
  pass `onOpenEnvironment={onOpenEnvironment}` in **every** `render(<EnvironmentSwitcher … />)`
  and `rerender(…)` call in the file.
- `App.test.tsx`: the mocked `listEnvironments` returns one environment `e1` "dev" with
  `activeId: null`; `renderWorkspace()` opens workspace `A`. The environment tab is the region
  named `"dev environment"`, the globals tab the region named `"Globals"`. Mock
  `environmentServices.getEnvironment` to resolve
  `{ id: 'e1', name: 'dev', description: '', color: 'cyan', confirmBeforeSend: false, variables: [] }`
  (as the existing rename test does) and make `e1` active with
  `act(() => useEnvironmentsStore.setState({ activeId: 'e1' }))` after `renderWorkspace()`.
- Tab assertions use `useTabsStore.getState()`: `tabs` (each `TabMeta` has `kind`,
  `environmentId`, `pinned`), `activeTabId`, `previewTabId`.

---

## Phase 1: Setup

**Purpose**: a known-green baseline

- [ ] T001 In `frontend/`, run `npm ci`, then from the repo root `make bindings`, then in `frontend/` `npx vitest run src/features/environments src/app/App.test.tsx` and confirm every file passes on the untouched code.

---

## Phase 2: Foundational

None — every change belongs to one story. (The eye's callback prop is introduced by US1 and
reused by US2.)

---

## Phase 3: User Story 1 — The eye opens the active environment's tab (P1) 🎯 MVP

**Goal**: with an environment active, the eye opens or focuses that environment's pinned tab;
no dialog.

**Independent Test**: make `e1` active, click "Open environment" → exactly one pinned `e1` tab,
focused; no dialog.

### Tests (red first)

- [ ] T002 [P] [US1] In `frontend/src/features/environments/components/EnvironmentSwitcher.test.tsx`: add `onOpenEnvironment` per "Shared test facts"; delete the two tests `'opens a quick look at the active environment and the globals'` and `'tells that no environment is active in the quick look'`; add `it('opens the active environment in a tab from the eye')`: set `activeId: 'e2'`, render, assert `screen.getByRole('button', { name: 'Open environment' })` has `title` `"Open environment"`, click it, expect `onOpenEnvironment` called once with `'e2'`, `onOpenGlobals` not called, and `screen.queryByRole('dialog')` is `null`. Run `npx vitest run src/features/environments/components/EnvironmentSwitcher.test.tsx` and confirm the new case fails (no button named "Open environment").
- [ ] T003 [P] [US1] In `frontend/src/app/App.test.tsx`, add a `describe('the eye next to the environment switcher')` block (inside the existing `describe('App')` so its `beforeEach`/`afterEach` apply) with: (a) `it('opens the active environment in a pinned, focused tab and never twice')` — setup per "Shared test facts", click `'Open environment'`, `await screen.findByRole('region', { name: 'dev environment' })`, assert exactly one tab with `kind === 'environment' && environmentId === 'e1'`, it is `pinned`, and `activeTabId` is its id; focus another tab with `act(() => useTabsStore.getState().openEnvironmentTab({ kind: 'globals' }, 'pinned'))`, click the eye again, assert the tab count is unchanged and `activeTabId` is the `e1` tab's id again; assert `screen.queryByRole('dialog')` is `null`; then (FR-007) type `x` into the `e1` tab's `'Add variable'` textbox (inside the `'dev environment'` region), focus the globals tab again, click the eye, and assert `useEnvironmentsStore.getState().drafts['env:e1']?.dirty` is still `true` and the region again shows the added row (`'Name of x'`); (b) `it('pins an environment tab that was open as a preview')` — `act(() => useTabsStore.getState().openEnvironmentTab({ kind: 'environment', environmentId: 'e1' }, 'preview'))`, then open globals pinned as above, click the eye, assert the single `e1` tab is now `pinned`, `previewTabId` is not its id, and it is the `activeTabId`. Run `npx vitest run src/app/App.test.tsx` and confirm both new cases fail.

### Implementation

- [ ] T004 [US1] In `frontend/src/features/environments/components/EnvironmentSwitcher.tsx`: add prop `onOpenEnvironment: (environmentId: string) => void` with JSDoc `/** Opens one environment in a main-area tab. */` (contract `contracts/eye-button.md`); move the `EYE_GLYPH` constant from `QuickLook.tsx` into this file using `styles.icon`; replace `<QuickLook />` with `<IconButton label="Open environment" onClick={() => onOpenEnvironment(active.id)}>{EYE_GLYPH}</IconButton>` rendered only for the active-environment case for now (US2 completes the no-environment branch in T007 — until then render it as `active ? … : null`); remove the `QuickLook` import; import `IconButton` from `@/shared/ui/icon-button/IconButton`. Add the `.icon` rule (copied verbatim from `QuickLook.module.css`) to `frontend/src/features/environments/components/EnvironmentSwitcher.module.css`. T002 → green.
- [ ] T005 [US1] In `frontend/src/app/App.tsx`: add `function openEnvironment(environmentId: string) { openEnvironmentTab({ kind: 'environment', environmentId }, 'pinned') }` next to `openGlobals`, and pass `onOpenEnvironment={openEnvironment}` in `ENVIRONMENT_SWITCHER`. T003 → green; `npx vitest run src/features/environments src/app/App.test.tsx` green.

**Checkpoint**: US1 works; the eye shows nothing when no environment is active (completed by US2).

---

## Phase 4: User Story 2 — With no environment active, the eye opens the globals (P2)

**Goal**: with "No environment", the eye reads "Open globals" and opens or focuses the pinned
Globals tab.

**Independent Test**: `activeId: null`, click "Open globals" → exactly one pinned Globals tab,
focused; no dialog.

### Tests (red first)

- [ ] T006 [P] [US2] In `frontend/src/features/environments/components/EnvironmentSwitcher.test.tsx` add `it('opens the globals from the eye when no environment is active')`: `activeId: null`, click `screen.getByRole('button', { name: 'Open globals' })`, expect `onOpenGlobals` called once, `onOpenEnvironment` not called, no dialog; and `it('names what the eye opens after the active environment')`: render with `activeId: null`, assert a button named `'Open globals'`, then `act(() => useEnvironmentsStore.setState({ activeId: 'e1' }))` and assert a button named `'Open environment'` and none named `'Open globals'`. In `frontend/src/app/App.test.tsx` (same describe block as T003) add `it('opens the globals in a pinned, focused tab when no environment is active')`: no `activeId` change, click `'Open globals'`, `await screen.findByRole('region', { name: 'Globals' })`, assert exactly one tab with `kind === 'globals'`, `pinned`, and it is the `activeTabId`; click again → tab count unchanged. Run both files and confirm the new cases fail.

### Implementation

- [ ] T007 [US2] In `frontend/src/features/environments/components/EnvironmentSwitcher.tsx` render the eye always: `<IconButton label={active ? 'Open environment' : 'Open globals'} onClick={() => (active ? onOpenEnvironment(active.id) : onOpenGlobals())}>{EYE_GLYPH}</IconButton>` (hoist the two label strings to module constants `OPEN_ENVIRONMENT_LABEL` / `OPEN_GLOBALS_LABEL`). No `App.tsx` change — `onOpenGlobals` is already wired to the pinned Globals tab. T006 → green.

**Checkpoint**: US1 and US2 both work; the Quick Look component is no longer rendered anywhere.

---

## Phase 5: User Story 3 — The Quick Look is gone everywhere (P3)

**Goal**: no Quick Look code, no code only it used, no document describing it (FR-008, FR-010).

**Independent Test**: the `grep` in quickstart.md finds nothing; the reworded documents describe
the eye as opening the tab.

- [ ] T008 [P] [US3] Delete `frontend/src/features/environments/components/QuickLook.tsx`, `QuickLook.module.css` and `QuickLook.test.tsx` (unused since T004).
- [ ] T009 [P] [US3] Delete `frontend/src/features/environments/hooks/use-discard-guard.ts` (the `hooks/` folder is then empty — remove it), `frontend/src/features/environments/components/DiscardChangesConfirm.tsx` and `DiscardChangesConfirm.module.css` (only `QuickLook.tsx` imported them — research R5).
- [ ] T010 [US3] Rewrite `frontend/src/features/environments/components/VariableTable.test.tsx` so no test uses the footer or the removed props (research R6). Add helpers `const saveFromTab = () => act(() => useEnvironmentsStore.getState().saveDraft(SCOPE))` and `const discardFromTab = () => act(() => useEnvironmentsStore.getState().discardDraft(SCOPE))` (import `act` from `@testing-library/react`). Replace every `user.click(screen.getByRole('button', { name: 'Save Variables' }))` with `await saveFromTab()` and every `'Discard Changes'` click with `discardFromTab()`. In `'adds, renames and removes rows, then saves the whole table'` replace the two lines asserting the "Save Variables" button is enabled and clicking it again with: `await screen.findByRole('textbox', { name: 'Name of base_url' })`, `await saveFromTab()`, `expect(variables.saveVariableTable).toHaveBeenCalledTimes(1)` (a clean draft is not saved again — `saveDraft` returns early when not dirty; keep `expect(screen.getAllByRole('row')).toHaveLength(3)`). In `'shows a save error inline and keeps the edits'` replace the final "Save Variables enabled" assertion with `expect(useEnvironmentsStore.getState().drafts[SCOPE]?.dirty).toBe(true)`. Delete `'clears its unsaved-edits flag when it closes'`. In `'keeps unsaved edits through an unmount…'` drop both `draftLifetime="workspace"` props, rename it `'keeps unsaved edits through an unmount'`, and replace the "Discard Changes enabled" assertion with `expect(dirtyScopes()).toEqual([SCOPE])`. Rewrite `'moves focus to the first invalid name after a failed save'` to go through the confirm path: rows `[plain({ key: 'ok' }), plain({ key: 'other', initialValue: 'x', currentValue: 'x' })]`, type `'{{'` into `'Name of other'` (user-event types a single `{`, so the row is now `other{`), click `'Secret other{'` (the secret toggle of the renamed row), `await saveFromTab()`, mock `saveVariableTable` to reject `new AppError('VAR-00400', 'the variable name is invalid')`, click the confirmation's `'Make Secret'`, then expect the alert text and focus on `'Name of other{'`. Every other test keeps its assertions. Run the file — all green against the **current** component (these are behaviour-preserving rewrites; the footer still exists until T011).
- [ ] T011 [US3] In `frontend/src/features/environments/components/VariableTable.tsx` remove: the `DraftLifetime` type and its JSDoc, the `draftLifetime` and `savedByTab` props (interface keeps `scope` and `label`), the outer component's `discardDraft` selector and the unmount `useEffect` that discarded `'view'` drafts, the `savedByTab` prop of `VariableTableEditor`, the editor's `discardDraft` selector, `discard()`, `requestSave()` and the whole `{savedByTab ? null : (<div className={styles.footer}>…</div>)}` block. Keep `afterSave`, `confirmSave`, `focusFirstInvalidName` and `TypeSwitchConfirm`. Drop now-unused imports (`useEffect` stays — `loadTable` still uses it). In `frontend/src/features/environments/components/VariableTable.module.css` delete both `.footer` rule blocks. Run `VariableTable.test.tsx` → green.
- [ ] T012 [US3] Drop `draftLifetime="workspace"` and `savedByTab` from the `VariableTable` usages in `frontend/src/features/environments/components/EnvironmentTab.tsx`, `frontend/src/features/environments/components/GlobalsTab.tsx` and `frontend/src/app/layout/ItemWorkspace.tsx`. `EnvironmentTab.test.tsx`'s "no Discard/Save buttons" assertions stay unchanged and pass.
- [ ] T013 [P] [US3] In `docs/prd/xenon-mvp-prd.md`: replace the switcher bullet `- An eye button next to it opens a quick look at the active environment's variables, editable in place, as in Postman.` with `- An eye button next to it opens the active environment's tab in the main area (pinned; focused if it is already open). With no active environment it opens the Globals tab.`; in the "Where are environments managed?" decision replace `the top-right switcher stays for picking the active environment, with its quick look.` with `the top-right switcher stays for picking the active environment, and its eye opens the active environment's tab (the Globals tab when none is active; XEN-015 replaced the quick look).`
- [ ] T014 [P] [US3] In `specs/XEN-002-environments-and-secrets/spec.md` (the live product spec; its `plan.md`, `tasks.md` and `checklists/` are the historical record of that build and stay as they are): in US secrets scenario 3 replace `(environment editor, quick look, Variables panel, URL hover)` with `(environment and Globals tabs, Variables panel, URL hover)`; in the User Story 5 intro replace `An eye button\nnext to the switcher opens a quick look at the active environment's variables, editable in place.` with `An eye button\nnext to the switcher opens the active environment's tab (the Globals tab when none is active).`; replace User Story 5 scenario 4 with `4. **Given** the eye button next to the switcher, **When** he clicks it, **Then** the active\n   environment's tab opens pinned in the main area (or is focused if already open); with no\n   active environment the Globals tab opens instead (XEN-015).`; in FR-020 replace `environment editor,\n  quick look, Variables panel, hovers` with `environment and Globals\n  tabs, Variables panel, hovers`; in FR-028 replace `An eye button next to it MUST open a quick look of the active environment and globals,\n  editable in place.` with `An eye button next to it MUST open the active environment's tab\n  (the Globals tab when none is active).`

**Checkpoint**: all three stories done.

---

## Phase 6: Polish & Cross-Cutting

- [ ] T015 Run `grep -rnE "QuickLook|useDiscardGuard|DiscardChangesConfirm|draftLifetime|savedByTab|DraftLifetime" frontend/src` → no output; `grep -rniE "quick ?look" docs/prd specs/XEN-002-environments-and-secrets/spec.md` → only the T013 "XEN-015 replaced the quick look" note.
- [ ] T016 In `frontend/`: `npm run lint`, `npm run typecheck`, `npm test` → all green (the typecheck proves no caller still passes the removed props or omits `onOpenEnvironment`).
- [ ] T017 From the repo root run `make validate` → green (Go and BDD suites unchanged).
- [ ] T018 Walk the by-hand steps 1–6 of `quickstart.md` in `make dev` and confirm each expected result.

---

## Dependencies & Execution Order

- T001 → everything.
- US1: T002, T003 (parallel, red) → T004 → T005.
- US2: T006 (red; may be written alongside T002/T003) → T007 (after T004, same file).
- US3: T008 after T004 (QuickLook no longer imported); T009 after T008; T010 → T011 → T012;
  T013 and T014 any time.
- Polish: T015–T018 after all stories.

Story order: US1 → US2 (shares `EnvironmentSwitcher.tsx`) → US3. T010–T014 do not depend on US2.

## Parallel Examples

- Red phase (`/create-tests`): T002, T003, T006 together (two files; T002 and T006 edit the same
  switcher test file, so one author writes both).
- US3: T008, T009, T013, T014 together; T010 → T011 → T012 in sequence.

## Implementation Strategy

1. MVP = US1 (T001–T005): the eye opens the active environment's tab.
2. US2 (T006–T007) completes the eye for "No environment".
3. US3 (T008–T014) removes the dead dialog code and fixes the documents.
4. Polish (T015–T018), then `/refactor-diff` until clean.
