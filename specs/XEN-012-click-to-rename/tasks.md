---

description: "Task list for XEN-012 click-to-rename"
---

# Tasks: Rename by Clicking the Name, Saved With the Save Hotkey

**Input**: Design documents from `specs/XEN-012-click-to-rename/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/editable-name.md, contracts/tab-save.md, quickstart.md

**Tests**: MANDATORY and first. This feature changes no Go behaviour, so its red-first tests are
Vitest + React Testing Library specs (plan.md → Complexity Tracking). Every test task comes
before the code it covers, and must be run and seen failing for the expected reason before that
code is written. The existing Godog rename scenarios are the regression lock for FR-015.

**Organization**: grouped by user story (spec.md). All paths are relative to the repository root.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel (different files, no dependency on an unfinished task)
- **[Story]**: the user story the task serves (US1–US5)

---

## Phase 1: Setup

**Purpose**: a green baseline to compare against.

- [ ] T001 Run `make frontend-check` and `make bdd TAG=@rename` on the task branch before any change, and record both as green in the PR description draft (baseline for FR-015 and FR-013)

---

## Phase 2: Foundational (blocking prerequisites)

**Purpose**: the shared component, the pending-name rule, and the tabs-store plumbing that every
story uses.

### Tests first (red)

- [ ] T002 [P] Write `frontend/src/shared/ui/editable-name/EditableName.test.tsx` covering every row of contracts/editable-name.md: button display with `aria-label`; click → focused input with all text selected; focus + Enter → input; Space → input; typing calls `onChange` per keystroke; Enter closes without `onChange`/`onCancel` and focus returns to the button; Escape calls `onCancel`, closes, and focus returns to the button; blur closes and keeps `value`; `maxLength` defaults to 200 and a 300-character paste keeps 200; `error` renders `role="alert"` with `aria-invalid` and `aria-describedby`; Enter/Escape keydown do not reach a parent `onKeyDown`; Mod+S keydown does reach it
- [ ] T003 [P] Write `frontend/src/features/tabs/lib/name-draft.test.ts` for `isNamePending(draft, saved)` (null saved → false; trimmed equal → false; trailing spaces only → false; different → true; undefined draft → false) and for `nameErrorMessage(error)` returning the exact texts in research.md R4 for `required` (request, example, collection, folder, environment), `clash` and `failed` (`withNextStep(description || code, 'Choose another name, then save.')`)
- [ ] T004 Run `cd frontend && npx vitest run editable-name name-draft` and confirm both files fail because the modules do not exist

### Implementation

- [ ] T005 [P] Create `frontend/src/shared/ui/editable-name/EditableName.tsx` and `EditableName.module.css` per contracts/editable-name.md: local `editing` state only, a `<button type="button">` display (`translate="no"`, `title="Click to rename"`, min-width and focus ring from the design tokens), a controlled `<input>` with `maxLength` (default 200), `spellCheck={false}` and `onFocus` select-all, Enter/Escape `stopPropagation`, blur closes, error `<p role="alert">` linked with `useId`
- [ ] T006 [P] Create `frontend/src/features/tabs/lib/name-draft.ts` exporting `isNamePending`, the `NameError` union from data-model.md, and `nameErrorMessage` with the messages from research.md R4 (reusing `withNextStep` from `@/shared/lib/next-step`)
- [ ] T007 Run `cd frontend && npx vitest run editable-name name-draft` and confirm green

**Checkpoint**: the shared component and the pending rule exist; no tab uses them yet.

---

## Phase 3: User Story 1 — Rename a request in its tab and save with the hotkey (P1) 🎯 MVP

**Goal**: a saved or never-saved request is renamed from the breadcrumb or the Info panel as a
draft, shows the dot, and is saved by Cmd/Ctrl+S in the order from contracts/tab-save.md.

**Independent test**: quickstart.md manual steps 1–8 and 10 for a saved request and a
never-saved request, plus the Vitest files below.

### Tests first (red)

- [ ] T008 [P] [US1] Extend `frontend/src/features/request-builder/store.test.ts`: `editName` changes `draft.name` without setting `dirty`, and `setName` still sets `dirty` (never-saved requests); `dropName(tabId, saved)` restores the name without touching `dirty`; `save(tabId, false, { name: 'Saved' })` sends `name: 'Saved'` to the mocked `saveRequest` while the store keeps the draft name; `adoptRename(requestId, previousName)` updates `baseRevision` on every tab of the request but replaces `draft.name` only on tabs whose trimmed draft equals `previousName`
- [ ] T009 [P] [US1] Write `frontend/src/features/tabs/name-drafts.test.ts` (request section; `platform/` mocked as in `store.test.ts`): editing the name of a saved request makes `isDirty` true and Escape (`cancelTabName`) makes it false again; save with only a pending name calls `renameItem` and not `saveRequest`; save with a pending name and a dirty URL calls `saveRequest` with the saved name **before** `renameItem` with the trimmed draft; an empty name sets `nameErrors[tab]` to `required` and calls neither; a `saveRequest` conflict (`REQ-02409`) stops the save without calling `renameItem` and leaves the name pending; a `renameItem` failure sets `nameErrors[tab]` to `failed` and the name stays pending; a second `save()` while the first is running does not rename twice; a name edited during a running save stays pending after it; a sidebar rename (`adoptTreeRename(id, previous)`) during a pending draft keeps the draft and the tab stays dirty; a draft equal to the new saved name stops being pending; the UI-state snapshot keeps the draft of a tab whose only change is a pending name; `handleChangedOnDisk` for that request does not reload (and so does not wipe) a tab whose only change is a pending name; a never-saved request whose only change is its name opens the close prompt
- [ ] T010 [P] [US1] Extend `frontend/src/app/layout/TabBreadcrumb.test.tsx`: the saved request's breadcrumb shows its name as a button labelled "Rename request"; clicking turns it into an input with the text selected; focusing it and pressing Enter does the same; typing does not change the breadcrumb ancestors or the tab label (`useTabTitle` still returns the tree name); a never-saved request's name is also click-to-edit and its tab label follows the draft; an example tab's label (`useTabTitle`) is the tree example name even while its draft name differs
- [ ] T011 [P] [US1] Extend `frontend/src/features/sidebar/components/ItemBreadcrumb.test.tsx`: the current crumb renders the `nameEditor` node it is given; `rename`/`renameExample` from the sidebar store are never called by the breadcrumb; ancestors still open their tabs
- [ ] T012 [P] [US1] Extend `frontend/src/features/side-panel/components/InfoPanel.test.tsx`: there is no always-visible Name input; the name is click-to-edit text bound to the same draft as the breadcrumb (editing in one shows in the other); blur does not call `renameItem`; a `nameErrors` entry for the tab renders inline
- [ ] T013 [P] [US1] Extend `frontend/src/features/tabs/components/TabStrip.test.tsx`: a request tab whose only change is a pending name shows the "Unsaved changes" dot, and the dot goes away after Escape
- [ ] T014 [US1] Extend `frontend/src/app/App.test.tsx` (next to "opens, saves and closes tabs with the keyboard"): open a saved request, click the breadcrumb name, type a new name, press Mod+S (Meta on mac, Ctrl elsewhere, as the existing test does): `renameItem` is called with the trimmed name, and the dot is gone after the tree reloads.
- [ ] T015 [US1] Run `cd frontend && npx vitest run request-builder/store name-drafts TabBreadcrumb ItemBreadcrumb InfoPanel TabStrip App.test` and confirm the new cases fail for the expected reasons (missing actions or props, old behaviour)

### Implementation

- [ ] T016 [US1] In `frontend/src/features/request-builder/store.ts`: add `editName` (patches `draft.name`, no `dirty`, no `curlPaste`) next to the unchanged `setName`; add `dropName(tabId, savedName)`; add the `options?: { name?: string }` parameter to `save` and send `{ ...draft, name: options.name }` when given; change `adoptRename(requestId, previousName)` per contracts/tab-save.md
- [ ] T017 [US1] In `frontend/src/features/tabs/store.ts`: add `nameErrors`, `editTabName`, `cancelTabName` and the private `saveTabChanges(tab)` (guarded by a per-tab in-flight set so a second save during a running one returns without writing) implementing the order in contracts/tab-save.md for request tabs (saved name from `useSidebarStore` tree via `selectLocation`); OR the pending name into `isDirty`; make `save()` and `resolveClose('save')` call `saveTabChanges` (unsaved requests still open the Save dialog); include the pending name in `keepDraft` and in the dirty check `handleChangedOnDisk` uses before reloading a tab; clear `nameErrors[tabId]` in `closeTab`
- [ ] T018 [P] [US1] Create `frontend/src/features/tabs/hooks/use-name-pending.ts` (reactive: tab kind → owning store draft name vs saved name through `isNamePending`) and use it in `frontend/src/features/tabs/components/TabItem.tsx` so `dirty` includes it
- [ ] T019 [P] [US1] In `frontend/src/features/tabs/hooks/use-tab-title.ts` return the tree name for a request that has a tree node, the tree example name for an example tab (today it returns the builder draft name, which would show a pending example name), and the draft name (or "Untitled Request") only for a request with no tree node
- [ ] T020 [US1] In `frontend/src/features/sidebar/components/ItemBreadcrumb.tsx` replace the `editing` state, `RenameInput` and `commitRename` with a `nameEditor: ReactNode` prop rendered in the current crumb; remove the `onRenamed` prop
- [ ] T021 [US1] In `frontend/src/app/layout/TabBreadcrumb.tsx` compose `<EditableName>` (value = the builder draft name, `label` "Rename request" on request tabs and "Rename example" on example tabs — example tabs only start editing for real once T051 routes `editTabName` to them, `onChange` → `editTabName`, `onCancel` → `cancelTabName`, `error` → `nameErrorMessage(nameErrors[tabId])`) and pass it as `nameEditor`; for a never-saved request use the same editor
- [ ] T022 [US1] In `frontend/src/features/side-panel/components/InfoPanel.tsx` replace the Name `<input>`, `commitName`, `FailedRename` and `RENAME_NEXT_STEP` with `<EditableName>` bound to the same draft, `editTabName`/`cancelTabName` and `nameErrors` from the tabs store
- [ ] T023 [US1] In `frontend/src/app/tree-actions.ts` change `adoptTreeRename(id, previousName)`; in `frontend/src/features/sidebar/components/TreeItem.tsx` call `onRenamed(node.id, node.name)` (name captured before the rename; the rename flow itself unchanged); update the prop type in `frontend/src/features/sidebar/components/SidebarPanel.tsx` and the call site in `frontend/src/app/App.tsx`
- [ ] T024 [US1] Run the T015 command and confirm green; run `cd frontend && npm test` to confirm no other spec broke (in particular `CollectionTree.test.tsx` and `SidebarPanel.test.tsx` F2 rename, FR-013)

**Checkpoint**: request renaming is complete and shippable on its own.

---

## Phase 4: User Story 2 — Rename a collection or folder in its tab (P1)

**Goal**: the heading is the only name editor; the Overview loses its Name field; the save
order runs settings → variables → documentation → rename.

**Independent test**: quickstart.md steps 1–6, 8 and 10 for a collection and a folder.

### Tests first (red)

- [ ] T025 [P] [US2] Extend `frontend/src/features/item-settings/store.test.ts`: `editName` sets `draft.name` without `dirty`; `dropName` restores it; `load` keeps a draft name that was pending against the previous `base.name` and takes the loaded name otherwise; `adoptName` no longer exists
- [ ] T026 [P] [US2] Extend `frontend/src/features/tabs/name-drafts.test.ts` (collection and folder section): a pending name makes `isDirty` true; save with a pending name and dirty auth calls `saveItemSettings` before `renameItem`; a pending documentation confirmation (`value:false`) stops before `renameItem`; a failing collection-variables save stops before `renameItem`; an empty name calls nothing and sets `required` (`collection` / `folder`)
- [ ] T027 [P] [US2] Extend `frontend/src/features/item-settings/components/ItemTab.test.tsx`: the heading is a button labelled "Rename collection" / "Rename folder"; for both a collection and a folder tab, clicking or focusing it and pressing Enter opens the input with the text selected; the Overview has no Name input; a `nameErrors` entry renders inline under the heading; and in `frontend/src/app/App.test.tsx` next to "opens a collection as a tab from the sidebar, saves it with Cmd/Ctrl+S…": rename the collection from its heading and press Mod+S, and `renameItem` is called after `saveItemSettings`
- [ ] T028 [US2] Run `cd frontend && npx vitest run item-settings name-drafts ItemTab App.test` and confirm the new cases fail

### Implementation

- [ ] T029 [US2] In `frontend/src/features/item-settings/store.ts` add `editName` and `dropName`, make `load` keep a pending draft name, delete `adoptName`
- [ ] T030 [US2] In `frontend/src/features/tabs/store.ts` extend `editTabName`, `cancelTabName`, `isDirty` and `saveTabChanges` to collection and folder tabs (saved name = tree node name, falling back to `base.name`); rename runs after today's `saveItemTab` succeeds with `value:true`
- [ ] T031 [P] [US2] In `frontend/src/features/tabs/hooks/use-name-pending.ts` add collection and folder tabs
- [ ] T032 [US2] In `frontend/src/features/item-settings/components/ItemTab.tsx` render `<EditableName>` inside the existing `<h2>` (value = `draft.name`, label "Rename collection"/"Rename folder", handlers and error from the tabs store); in `frontend/src/features/item-settings/components/ItemOverview.tsx` delete the Name field, `commitName`, `onNameKeyDown`, `nameError` and `RENAME_NEXT_STEP`
- [ ] T033 [US2] Run the T028 command and confirm green

**Checkpoint**: US1 and US2 work independently.

---

## Phase 5: User Story 5 — Close a tab with an unsaved name (P2)

**Goal**: the close prompt and bulk close treat a pending name as unsaved, and Save / Discard /
Cancel behave as for any edit.

**Independent test**: quickstart.md step 9 for every tab type implemented so far.

### Tests first (red)

- [ ] T034 [US5] Extend `frontend/src/features/tabs/name-drafts.test.ts` (close section, for request, collection and folder tabs): `requestClose` on a tab whose only change is its name opens the prompt; `resolveClose('discard')` closes it, calls no rename and drops the draft; `resolveClose('cancel')` keeps the draft; `resolveClose('save')` renames then closes; `resolveClose('save')` with an empty name keeps the tab open, sets `closeError` and `nameErrors`; `closeAll` queues a name-only dirty tab
- [ ] T035 [US5] Run `cd frontend && npx vitest run name-drafts` and confirm only the new close cases fail (the T017/T030 save routine already routes `resolveClose`; any case that passes already must be checked to assert the right thing, not deleted)

### Implementation

- [ ] T036 [US5] In `frontend/src/features/tabs/store.ts` make `resolveClose('save')` mirror `nameErrors[tab]` into `closeError` when `saveTabChanges` returns `ok:false` from a name check, and make `closeTab` drop every name draft and error for the closed tab
- [ ] T037 [US5] Run the T035 command and confirm green

---

## Phase 6: User Story 3 — Rename an environment in its tab (P2)

**Goal**: the heading is the only name editor; the Rename form is gone; the clash is caught
before anything is written.

**Independent test**: quickstart.md steps 1–9 for an environment.

### Tests first (red)

- [ ] T038 [P] [US3] Extend `frontend/src/features/environments/store.test.ts`: `editName`/`dropName` manage `nameDrafts`; `findNameClash` matches another environment ignoring case and surrounding spaces and never matches the environment itself; `deleteEnvironment` and `reset` clear the draft
- [ ] T039 [P] [US3] Extend `frontend/src/features/tabs/name-drafts.test.ts` (environment section): pending name → dirty; save with a pending name and dirty variables calls `saveVariableTable` before `updateEnvironment`; a clash (`"  DEV "` vs `"dev"`) sets `nameErrors` to `clash` and calls neither; an empty name sets `required` (`environment`); a variable save waiting for the secret-type confirmation (`value:false`) does not rename; success drops the draft; closing the tab drops the draft
- [ ] T040 [P] [US3] Extend `frontend/src/features/environments/components/EnvironmentTab.test.tsx`: the name appears once, as a button labelled "Rename environment" in the heading; clicking it, or focusing it and pressing Enter, opens the input with the text selected; there is no "Environment name" input and no Rename button; the clash message "An environment named “dev” already exists. Choose another name." and "Type a name for the environment." render inline; replace the existing "renames the environment and follows a rename made elsewhere" case with the draft-based equivalent (rename elsewhere keeps the pending draft); and in `frontend/src/app/App.test.tsx`: rename an environment from its tab heading and press Mod+S, and `updateEnvironment` is called with the new name
- [ ] T041 [US3] Run `cd frontend && npx vitest run environments name-drafts EnvironmentTab App.test` and confirm the new cases fail

### Implementation

- [ ] T042 [US3] In `frontend/src/features/environments/store.ts` add `nameDrafts`, `editName`, `dropName`, `findNameClash`; clear the draft in `deleteEnvironment` and `reset`
- [ ] T043 [US3] In `frontend/src/features/tabs/store.ts` extend `editTabName`, `cancelTabName`, `isDirty`, `saveTabChanges` (pre-check clash; `saveDraft` → `renameEnvironment` → `dropName`) and `closeTab` to environment tabs
- [ ] T044 [P] [US3] In `frontend/src/features/tabs/hooks/use-name-pending.ts` add environment tabs
- [ ] T045 [US3] In `frontend/src/features/environments/components/EnvironmentTab.tsx` render `<EditableName>` inside the `<h2>` (value = `nameDrafts[id] ?? environment.name`), delete the `rename` form, the name `useState`/`syncedName`, `NAME_REQUIRED`, `nameRequired` and the name part of `errorText` (moved to `nameErrorMessage`); keep the colour and Confirm-before-send controls and their error banner
- [ ] T046 [US3] Run the T041 command and confirm green

---

## Phase 7: User Story 4 — Rename a saved example in its tab (P3)

**Goal**: the example breadcrumb edits a draft; the example save runs before `renameExample`.

**Independent test**: quickstart.md steps 1–6 and 9 for an example.

### Tests first (red)

- [ ] T047 [P] [US4] Extend `frontend/src/features/examples/store.test.ts`: `adoptRename(exampleId, previousName)` updates `tab.example` and replaces the builder draft name only when it equals `previousName`
- [ ] T048 [P] [US4] Extend `frontend/src/features/tabs/name-drafts.test.ts` (example section): a pending example name → dirty; save with a pending name and an edited status code calls `saveExample` (with the saved name) before `renameExample`; an invalid code error from `saveExample` stops before `renameExample`; an empty name sets `required` (`example`); discard on close drops the draft; and extend `frontend/src/app/layout/TabBreadcrumb.test.tsx`: on an example tab the breadcrumb name is a button labelled "Rename example", and clicking it or focusing it and pressing Enter opens the input with the text selected
- [ ] T049 [US4] Run `cd frontend && npx vitest run examples/store name-drafts TabBreadcrumb` and confirm the new cases fail

### Implementation

- [ ] T050 [US4] In `frontend/src/features/examples/store.ts` change `adoptRename(exampleId, previousName)` per contracts/tab-save.md
- [ ] T051 [US4] In `frontend/src/features/tabs/store.ts` extend `editTabName`, `cancelTabName`, `isDirty` and `saveTabChanges` to example tabs (saved name = the tree example name; rename via `sidebar.renameExample` then `examples.adoptRename`)
- [ ] T052 [P] [US4] In `frontend/src/features/tabs/hooks/use-name-pending.ts` add example tabs
- [ ] T053 [US4] Run the T049 command and the T010/T011 breadcrumb specs and confirm green

---

## Phase 8: Polish & cross-cutting

- [ ] T054 Grep `frontend/src` for `RENAME_NEXT_STEP`, `commitName`, `adoptName`, `onRenamed(current` and `name="item-name"` outside `features/sidebar/components/RenameInput.tsx`, and remove every leftover reference (dead code after US1–US4)
- [ ] T055 Run `make frontend-check` (lint incl. import boundaries, typecheck, all Vitest) and fix until green
- [ ] T056 Run `make bdd` (full Godog suite, unchanged) and confirm green: FR-015
- [ ] T057 Run `make validate` and confirm green
- [ ] T058 Walk through quickstart.md manual steps 1–11 in `wails3 dev` for all six item kinds (saved request, never-saved request, example, collection, folder, environment)

---

## Dependencies & execution order

- Phase 1 → Phase 2 → US1 (Phase 3). US1 creates `saveTabChanges`, `editTabName`,
  `cancelTabName`, `nameErrors` and `use-name-pending.ts`, which US2–US5 extend.
- US2 (Phase 4) depends on US1. US5 (Phase 5) depends on US1 and US2 (it tests request and
  collection close). US3 (Phase 6) and US4 (Phase 7) depend on US1 only and can run in parallel
  with each other after US2 if they are staffed separately. They touch the same
  `tabs/store.ts`, so merge them one after the other.
- Phase 8 runs after every story.

## Parallel examples

- Phase 2: T002 ∥ T003, then T005 ∥ T006.
- US1 tests: T008 ∥ T009 ∥ T010 ∥ T011 ∥ T012 ∥ T013 ∥ T014 (different files); implementation T018 ∥ T019 after T016/T017.
- US3 tests: T038 ∥ T039 ∥ T040.

## Implementation strategy

MVP = Phases 1–3 (requests, the most renamed item). Then US2 (collections/folders) to finish the
P1 scope, US5 (close prompt), US3 (environments) and US4 (examples). Each phase ends at a green
checkpoint.
