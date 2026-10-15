# Implementation Plan: Rename by Clicking the Name, Saved With the Save Hotkey

**Branch**: `xen-012` | **Date**: 2026-10-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/XEN-012-click-to-rename/spec.md`

## Summary

Every name shown inside a main-area tab (request, example, collection, folder, environment)
becomes one shared click-to-edit component, `EditableName` (`frontend/src/shared/ui/editable-name/`).
Typing into it edits a **name draft** owned by the store that already owns that tab's other
drafts. The tabs store treats a pending name like any other unsaved edit: the dirty dot, the
close prompt, bulk close. One save routine in the tabs store, `saveTabChanges`, serves both
Cmd/Ctrl+S and the close prompt's Save. It runs in this order:

1. **Pre-check the name** (FR-010, FR-011). An empty name, or an environment name another
   environment already has, stops the save before anything is written. The error shows inline
   under the name.
2. **Save the tab's other changes** exactly as today. A request save sends the **saved** name, so
   the YAML `name` never changes without the matching file rename. If this step fails or waits
   for a confirmation, the save stops and the name stays pending.
3. **Rename** through the existing operation (`TreeService.Rename`, `ExampleService.Rename`,
   `EnvironmentService.Update`). If it fails, its error shows inline and the name stays pending.

The **saved name** is always what the sidebar tree (requests, examples, collections, folders) or
the environment list (environments) shows. So the tab label, the breadcrumb ancestors and the
sidebar change only after a successful save (FR-009). A rename made elsewhere moves the saved
name under a pending draft without touching the draft (FR-014).

There is no backend change. The Go rename, save and validation code is untouched (FR-015).

## Technical Context

**Language/Version**: TypeScript 5 (React 19, Vite); Go 1.x is untouched

**Primary Dependencies**: React, Zustand, CSS Modules + design tokens. No new dependency.

**Storage**: N/A. Names are written by the existing Go rename operations. Request drafts are
persisted with the UI state as they are today.

**Testing**: Vitest + React Testing Library (`npm test` in `frontend/`, run by
`make frontend-check` inside `make validate`). The existing Godog rename scenarios run unchanged
as a regression lock.

**Target Platform**: Wails v3 desktop webview on macOS, Windows and Linux

**Project Type**: desktop-app (Go backend + React UI)

**Performance Goals**: typing in the name input re-renders only the name editor, the tab item and
the components that select that draft (narrow Zustand selectors)

**Constraints**: the frontend-architecture import boundaries (shared ← features ← app); no
`platform/` call from components; names are at most 200 characters (`maxItemNameLength`)

**Scale/Scope**: about 22 frontend source files changed or added (see the tree below), plus their tests; 0 Go files changed

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first**: no new state outside the existing stores and UI state; workspace
  `reset()` clears the new name drafts along with the other drafts.
- [x] **II. Secrets**: names are not secrets; no variable value is touched.
- [x] **III. Go owns logic**: renames still go through `platform/services/*` (`renameItem`,
  `renameExample`, `updateEnvironment`) via the stores; React never touches the disk.
- [x] **IV. Architecture skills**: `EditableName` is domain-free and goes in `shared/ui/`. Name
  drafts live in each feature's own `store.ts`. Cross-feature reads go only through `store.ts`.
  The breadcrumb's name editor is composed in `app/layout/TabBreadcrumb.tsx` and passed into
  the sidebar's `ItemBreadcrumb` as props.
- [x] **V. BDD-first**: there is no Go behaviour to specify, so there are no new `.feature`
  files (see Complexity Tracking). The red-first tests are Vitest specs, written and proven red
  before the code (the tests-first tasks that open every phase of tasks.md). The existing rename features (`tree/tree_crud.feature`,
  `tree/tree_order_and_names.feature`, `environments/manage_environments.feature`,
  `examples/examples_crud.feature`) must stay green.
- [x] **VI. Postman parity**: the inline name in the tab header follows Postman's click-to-edit
  tab title. Saving with the tab, rather than right away, is the user's explicit choice
  (issue #9).
- [x] **VII. Cross-platform**: Cmd/Ctrl+S comes from the existing `Mod+S` binding; no paths
  involved.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-012-click-to-rename/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── editable-name.md     # the shared component's props and behaviour
│   └── tab-save.md          # name-draft store actions and the tab save order
├── checklists/
└── tasks.md
```

### Source Code (repository root)

```text
frontend/src/
├── shared/ui/editable-name/
│   ├── EditableName.tsx              # NEW: text ⇄ input, Enter/Escape/blur, maxLength 200, inline error
│   ├── EditableName.module.css       # NEW
│   └── EditableName.test.tsx         # NEW
├── features/request-builder/
│   ├── store.ts                      # setName no longer marks content dirty; save(tabId, { name }) override;
│   │                                 #   adoptRename(requestId, previousName) keeps pending drafts
│   └── store.test.ts                 # extended
├── features/examples/store.ts        # adoptRename keeps a pending example name draft
├── features/item-settings/
│   ├── store.ts                      # editName / dropName; load keeps a pending draft name; adoptName deleted
│   ├── components/ItemTab.tsx        # heading → EditableName
│   ├── components/ItemOverview.tsx   # Name field removed
│   └── components/ItemTab.test.tsx   # extended
├── features/environments/
│   ├── store.ts                      # nameDrafts, editName / dropName, findNameClash
│   ├── components/EnvironmentTab.tsx # heading → EditableName; rename form removed
│   └── components/EnvironmentTab.test.tsx
├── features/side-panel/components/
│   ├── InfoPanel.tsx                 # Name input → EditableName on the same draft
│   └── InfoPanel.test.tsx
├── features/sidebar/components/
│   ├── ItemBreadcrumb.tsx            # current crumb rendered from a `nameEditor` slot; no rename call
│   ├── ItemBreadcrumb.test.tsx
│   ├── TreeItem.tsx                  # onRenamed(id, previousName): signature only, rename behaviour unchanged
│   └── SidebarPanel.tsx              # forwards onRenamed(id, previousName)
├── features/tabs/
│   ├── store.ts                      # nameErrors; isDirty includes the pending name; saveTabChanges used by
│   │                                 #   save() and resolveClose(); closeTab/dropItem drop name drafts; keepDraft
│   ├── lib/name-draft.ts             # NEW: pure helpers (isNamePending, trimmed compare, empty-name messages)
│   ├── lib/name-draft.test.ts        # NEW
│   ├── hooks/use-tab-title.ts        # saved request: tree name (draft name only for never-saved requests)
│   ├── hooks/use-name-pending.ts     # NEW: reactive "is the tab's name pending" for TabItem
│   ├── components/TabItem.tsx        # dirty includes the pending name
│   └── name-drafts.test.ts           # NEW: store-level save / discard / close / clash scenarios
└── app/
    ├── tree-actions.ts               # adoptTreeRename(id, previousName)
    ├── App.tsx                       # passes adoptTreeRename to SidebarPanel (new signature)
    └── layout/
        ├── TabBreadcrumb.tsx         # composes EditableName on the builder draft into ItemBreadcrumb
        └── TabBreadcrumb.test.tsx    # extended
```

**Structure Decision**: this is frontend-only work inside the existing `frontend/src` layout.
Unchanged on purpose: `features/sidebar/components/RenameInput.tsx` and the rename flow in
`TreeItem.tsx` (sidebar F2 / context-menu rename, FR-013; only its `onRenamed` callback gains the
previous name), `features/workspaces/components/RenameWorkspaceDialog.tsx`
(FR-013), and every Go file.

### Where each tab's name editor sits (FR-003)

| Tab | Editor | Draft owner | Saved name source | Rename call |
| --- | --- | --- | --- | --- |
| Request (saved) | Breadcrumb last crumb + Info panel | `request-builder` `draft.name` | sidebar tree node | `sidebar.rename` |
| Request (never saved) | Breadcrumb last crumb + Info panel | `request-builder` `draft.name` | none (draft is the name) | none; Save dialog proposes it |
| Example | Breadcrumb last crumb | `request-builder` `draft.name` of the example tab | tree example under its request | `sidebar.renameExample` |
| Collection / folder | Heading (`ItemTab`) | `item-settings` `draft.name` | sidebar tree node (fallback `base.name`) | `sidebar.rename` |
| Environment | Heading (`EnvironmentTab`) | `environments` `nameDrafts[id]` | `environments[].name` | `environments.renameEnvironment` |

### BDD / test files that lock each requirement

No new `.feature` file; see Complexity Tracking. Coverage by Vitest:

| Requirement | Test file |
| --- | --- |
| FR-001, FR-002, FR-006, FR-007, FR-012, FR-016, FR-017 | `shared/ui/editable-name/EditableName.test.tsx` (FR-016: every tab test below finds the name through the same "Rename <kind>" button that only `EditableName` renders) |
| FR-001, FR-003 (request, never-saved request), FR-009 | `app/layout/TabBreadcrumb.test.tsx`, `features/sidebar/components/ItemBreadcrumb.test.tsx`, `features/side-panel/components/InfoPanel.test.tsx` |
| FR-001, FR-003 (collection/folder) | `features/item-settings/components/ItemTab.test.tsx` |
| FR-001, FR-003, FR-010 (environment) | `features/environments/components/EnvironmentTab.test.tsx` |
| FR-004, FR-005, FR-008, FR-009, FR-010, FR-011, FR-014 for all five tab types | `features/tabs/name-drafts.test.ts` |
| FR-004 (dot) | `features/tabs/components/TabStrip.test.tsx` |
| FR-005 via the real Mod+S binding | `app/App.test.tsx` (request, collection, environment). The `Mod+S` → `useTabsStore.save()` binding doesn't depend on the tab type, so folder and example "save with the hotkey" (SC-004) are covered by `name-drafts.test.ts` calling `save()`, the action the binding runs. |
| Pending-name helpers | `features/tabs/lib/name-draft.test.ts` |
| Store details (setName vs dirty, save name override, adoptRename) | `features/request-builder/store.test.ts`, `features/item-settings/store.test.ts`, `features/environments/store.test.ts`, `features/examples/store.test.ts` |
| FR-015 regression | existing `tree/tree_crud.feature`, `tree/tree_order_and_names.feature`, `environments/manage_environments.feature`, `examples/examples_crud.feature` (unchanged) |
| FR-013 regression | existing `features/sidebar/components/CollectionTree.test.tsx`, `SidebarPanel.test.tsx` (F2 rename) and the workspace switcher tests (unchanged, must stay green) |

`/create-tests` for this task writes the red Vitest specs above (the tests-first tasks that open every phase of tasks.md). The Godog
harness drives Go bindings, not React, and no binding changes.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
| --- | --- | --- |
| No new Godog `.feature` files (Principle V) | The feature changes only frontend behaviour; every Go binding it calls is unchanged and already covered by the rename scenarios listed above | A Godog scenario can't drive React; a feature that re-asserts the unchanged `TreeService.Rename` contract would test nothing new. Red-first still holds: the Vitest specs are written and proven red before the code. |
