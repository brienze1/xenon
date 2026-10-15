---

description: "Task list for XEN-011 — Reorder environments with drag and drop"
---

# Tasks: Reorder Environments With Drag and Drop

**Input**: Design documents from `specs/XEN-011-environment-reorder/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: MANDATORY (constitution V, BDD-first). Every user story starts with Godog feature files in `test/integration/features/` that are proven red before implementation (`/create-tests` writes all of them up front, red against `origin/main`); frontend stories also get Vitest tests. `/create-tests XEN-011` writes the `.feature` files listed below; `/implement-task XEN-011` does the rest.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested on its own.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story the task belongs to (US1–US5 from spec.md)
- Every backend file follows its `.claude/skills/*.md` skill; every frontend file follows `.claude/skills/frontend-architecture/SKILL.md`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Branch and baseline

- [ ] T001 Work on branch `xen-011` (fresh from `origin/main`), run `make bdd` and `make validate` once to record a green baseline before any change

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The stored order, its lenient parsing and the arranged list — every story reads them.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T002 Add `EnvironmentOrder []string` (nil = no stored order), `WithEnvironmentOrder(order []string) Workspace` (non-nil copy) and `"environmentOrder": len(...)` in `LogValue` to `internal/domain/entity/workspace.go`
- [ ] T003 Change `NewEnvironmentList(environments []Environment, order []string, activeId string)` in `internal/domain/entity/environment_list.go` to arrange per decision R5 (stored order first-occurrence-wins skipping unknown/malformed ids → other readable environments by `sortKey()` → malformed environments by file name) and add `IDs()`, `OrderMoving(id, index)` (index ≥ count → last), `OrderAppending(id)`, `OrderReplacing(previous, id)`, `OrderRemoving(id)` per data-model.md
- [ ] T004 [P] Add unit tests for T003 in `internal/domain/entity/environment_list_test.go` (stale ids, duplicate ids, unlisted alphabetical tail, malformed last, each order operation including index past the end and replacing an absent id)
- [ ] T005 Update the one other caller to `entity.NewEnvironmentList(existing, nil, "")` in `internal/application/usecase/plan_environment_import.go`
- [ ] T006 Add `Environments yaml.Node \`yaml:"environments"\`` and `EnvironmentOrder []string \`yaml:"-"\`` to `internal/integration/persistence/model/workspace.go`; `NewWorkspace` copies the entity order into `EnvironmentOrder`; `ToEntity` sets the entity order from `environments.order` only when `environments` is a mapping whose `order` is a sequence of scalars (else nil); `Fields()` emits `yamlnode.Object("environments", yamlnode.StringList("order", order))` right after the `order` field, only when `EnvironmentOrder != nil`
- [ ] T007 [P] Add unit tests for T006 in `internal/integration/persistence/model/workspace_test.go` (absent → nil and no `environments` field emitted; `environments: oops` and `environments: {order: oops}` → nil without a decode error; round-trip of a list; empty list emitted as `order: []`)
- [ ] T008 Load the workspace in `manageEnvironment.list` and pass `workspace.EnvironmentOrder` to `entity.NewEnvironmentList`; add `loadWorkspace usecase.LoadWorkspace` and `saveWorkspaceFile usecase.SaveWorkspaceFile` to the struct and constructor in `internal/application/service/manage_environment.go`, plus a private `writeOrder(ctx, root, order)` (load → `WithEnvironmentOrder` → `saveWorkspaceFile.Save`)
- [ ] T009 Wire the workspace persistence (already implementing `LoadWorkspace` and `SaveWorkspaceFile`) into `NewManageEnvironmentService` in `internal/infra/dependency/application/manage_environment.go`; run `make sort-imports`

**Checkpoint**: `List` honours a hand-written `environments.order`; nothing writes it yet.

---

## Phase 3: User Story 1 - Drag an environment to a new place (Priority: P1) 🎯 MVP

**Goal**: Pointer drag-and-drop reorders the sidebar list; the order is saved in `workspace.yaml` and shown by the switcher.

**Independent Test**: In a workspace with three environments, drag the last one to the top, restart, and see the same order in the sidebar and the switcher.

### Tests for User Story 1 (write first, prove red)

- [ ] T010 [P] [US1] Write `test/integration/features/environments/reorder_environments.feature` (`@all @xen_011 @environments @reorder_environments`) from plan.md's BDD table: `EnvironmentService.Reorder` moves and returns the arranged list; `environments.order` written with the ids; `List` after a hand-written order (stale id ignored, duplicate id first-wins, unlisted alphabetical tail, broken file last); malformed `environments` value → alphabetical and the workspace still loads, and the next reorder replaces it with a valid `environments.order` list; never-reordered workspace: `List` leaves `acme/workspace.yaml` unchanged; same-position reorder leaves it unchanged; index past the end → last; negative index → `ENV-17400`; unknown id → `ENV-01404`; `the workspace path "acme" is read-only` → error and order unchanged; `active_id` and environment files unchanged after a reorder
- [ ] T011 [P] [US1] Add Vitest cases to `frontend/src/features/environments/store.test.ts`: `reorderEnvironment` applies the moved list at once, replaces it with the returned list, restores the previous list and returns the error on failure, and makes no call for a same-position move; and to `frontend/src/app/bootstrap.test.ts`: a reorder of the environments store leaves open environment tabs open and clean (FR-016)
- [ ] T012 [P] [US1] Add Vitest cases to `frontend/src/features/sidebar/components/EnvironmentsView.test.tsx`: a pointer drag of the last row onto the first calls the store with index 0; Globals stays first and is not draggable; broken rows are not draggable; rows are not draggable while the search box has text or a row is being renamed; a failed reorder shows the error banner with the next step
- [ ] T013 [P] [US1] Add a Vitest case to `frontend/src/features/environments/components/EnvironmentSwitcher.test.tsx`: the menu lists environments in store order when that order is not alphabetical
- [ ] T014 [US1] Run `make bdd TAG=@reorder_environments` and the three Vitest files; record that the new cases fail for the missing behaviour (red)

### Implementation for User Story 1

- [ ] T015 [US1] Add `Reorder(ctx context.Context, localId, id string, index int) (*entity.EnvironmentList, error)` to `ManageEnvironmentService` in `internal/application/adapter/manage_environment.go`
- [ ] T016 [US1] Implement `manageEnvironment.Reorder` in `internal/application/service/manage_environment.go` per plan.md note 3 (not found → `ENV-01404`; write only when `OrderMoving` differs from `IDs()`; publish `VariablesChanged` with the environment's scope; return the re-listed list with the active id)
- [ ] T017 [P] [US1] Create `internal/integration/entrypoint/dto/reorder_environment.go` with `ReorderEnvironment{LocalId, Id, Index}` (`local_id`/`id` `required,not_blank`; `index` `min=0`)
- [ ] T018 [US1] Add `Reorder(ctx, dto.ReorderEnvironment) (*dto.EnvironmentList, error)` to `EnvironmentManageController` in `internal/integration/adapter/environment_manage_controller.go` and implement it in `internal/integration/entrypoint/controller/environment_manage.go` with the controller skill template, validator code `ENV-17400`, and `triggerAutoSync(ctx, e.autoSyncService.Saved, request.LocalId)`
- [ ] T019 [US1] Add `Reorder(ctx, dto.ReorderEnvironment) (*dto.EnvironmentList, error)` to the binding in `internal/infra/server/binding/environment.go`; run `make bindings` and `make sort-imports`
- [ ] T020 [US1] Add `reorderEnvironment(localId, id, index)` to `frontend/src/platform/services/environments.ts` per contracts/environment-service.md
- [ ] T021 [US1] Add `reorderEnvironment(id, index)` to the store in `frontend/src/features/environments/store.ts` per contracts/environments-view-ui.md (optimistic, rollback on failure)
- [ ] T022 [US1] Make the rows sortable in `frontend/src/features/sidebar/components/EnvironmentsView.tsx`: `DndContext` (`PointerSensor` distance 5 px, `closestCenter`) + `SortableContext` (`verticalListSortingStrategy`) over readable environment ids; `useSortable({ id, disabled })` per readable row with `disabled` while searching, renaming or creating; Globals and broken rows outside the sortable items; text selection suspended during a drag; drop → target index among readable environments → `reorderEnvironment`; error → existing banner
- [ ] T023 [P] [US1] Add dragging and drop-position styles with design tokens only in `frontend/src/features/sidebar/components/EnvironmentsView.module.css`
- [ ] T024 [US1] Run `make bdd TAG=@reorder_environments` and the US1 Vitest files until green

**Checkpoint**: MVP — pointer reordering works end to end and persists.

---

## Phase 4: User Story 2 - The order travels with the workspace (Priority: P1)

**Goal**: The order survives other writers of `workspace.yaml`, refreshes on external changes and merges in git without conflicts.

**Independent Test**: Reorder on one clone, push, pull on another; reorder differently on both and pull — same order, no conflict prompt.

### Tests for User Story 2 (write first, prove red)

- [ ] T025 [P] [US2] Write `test/integration/features/environments/environment_order_preserved.feature` (`@all @xen_011 @environments @environment_order_preserved`): a workspace-level `TreeService.Move` of a collection, a globals variable edit and a `WorkspaceService.Rename` each keep `environments.order` and an unknown sibling key under `environments` (FR-019)
- [ ] T026 [P] [US2] Write `test/integration/features/file_changes/environment_order_changes.feature` (`@all @xen_011 @file_changes @environment_order_changes`): after `acme/workspace.yaml` is rewritten with another `environments.order` and `the file watcher reports a change to the workspace path "acme/workspace.yaml"`, a `variables:changed` event is emitted and `EnvironmentService.List` returns the new order
- [ ] T027 [P] [US2] Write `test/integration/features/git/conflicts_environment_order.feature` (`@all @xen_011 @git @conflicts_environment_order`) with the clone/commit/push/pull steps of `features/git/conflicts_structural.feature`: two clones reorder differently → pull succeeds with no conflict and `environments.order` merged by the `order` rule; one clone adds an id while the other removes another → both changes kept
- [ ] T028 [US2] Run `make bdd TAG=@xen_011` for these three files; record the results. `/create-tests` writes every XEN-011 feature file before any implementation, so all three were proven red against `origin/main` there; at this point T027 is still red (merge rule missing) while T025 and T026 already pass through Phase 2 (every writer round-trips the entity; the watcher already re-lists) and stay as regression locks

### Implementation for User Story 2

- [ ] T029 [US2] In `internal/application/usecase/merge_structural_file.go` add `"environments": true` to `knownStructuralFields`, add `const environmentOrderField = "environments.order"`, and call `mergeOrder` when `path == orderField || path == environmentOrderField`
- [ ] T030 [US2] Run `make bdd TAG=@xen_011` and `make bdd TAG=@git` until green

**Checkpoint**: The order is safe under every other writer and in git.

---

## Phase 5: User Story 3 - Reorder with the keyboard (Priority: P2)

**Goal**: Space / arrows / Space (Escape cancels) reorders, with position announcements.

**Independent Test**: Keyboard only — move the second environment to the top; the order is saved and announced.

### Tests for User Story 3 (write first, prove red)

- [ ] T031 [US3] Add Vitest cases to `frontend/src/features/sidebar/components/EnvironmentsView.test.tsx`: Space, ArrowUp, Space on the second row calls the store with index 0; Escape after a move makes no call; the live region reads `Picked up {name}.`, `{name} moved to position 1 of 3.`, `{name} dropped at position 1 of 3.` and the cancel text; Space on Globals, on a broken row and while the search box has text picks nothing up; Space never opens the environment; Enter still opens it
- [ ] T032 [US3] Run that Vitest file; record the new cases red

### Implementation for User Story 3

- [ ] T033 [US3] In `frontend/src/features/sidebar/components/EnvironmentsView.tsx` add `KeyboardSensor` (`sortableKeyboardCoordinates`, key codes `start: Space`, `end: Space|Enter`, `cancel: Escape`) and the `accessibility` prop with the screen-reader instructions and announcements from contracts/environments-view-ui.md; suppress the row button's Space click while a keyboard drag starts or runs
- [ ] T034 [US3] Run the Vitest file until green

**Checkpoint**: Keyboard parity with the tab strip.

---

## Phase 6: User Story 4 - New environments go last; existing ones keep their place (Priority: P2)

**Goal**: Create, duplicate and import append; Replace keeps the slot; rename keeps the place; delete removes the id.

**Independent Test**: Reorder, then create, duplicate, import, rename and delete, checking positions after each step.

### Tests for User Story 4 (write first, prove red)

- [ ] T035 [P] [US4] Write `test/integration/features/environments/environment_order_placement.feature` (`@all @xen_011 @environments @environment_order_placement`): create and duplicate are listed last and appended to `environments.order`; in a never-reordered workspace with `Beta` and `Gamma`, creating `Alpha` writes `environments.order` as Beta, Gamma, Alpha; rename keeps the position and leaves `environments.order` unchanged; delete of a listed id removes it; delete in a never-ordered workspace leaves `acme/workspace.yaml` unchanged
- [ ] T036 [P] [US4] Write `test/integration/features/import/import_environment_order.feature` (`@all @xen_011 @import @import_environment_order`) with the existing import steps: a no-clash environment import and a "Keep both" import are appended; two environments in one import land last in import order; a "Replace" import takes the replaced environment's position; a whole-workspace Xenon import keeps the source's `environments.order`
- [ ] T037 [US4] Run `make bdd TAG=@environment_order_placement` and `make bdd TAG=@import_environment_order`; record the red results (rename and whole-workspace import scenarios are expected to pass already and act as regression locks)

### Implementation for User Story 4

- [ ] T038 [US4] In `internal/application/service/manage_environment.go`: `Create` and `Duplicate` list before `saveNamed`, compute `OrderAppending(newId)` and call `writeOrder` after the save succeeds; `Delete` loads the workspace and list before deleting and, after the file is gone, calls `writeOrder(list.OrderRemoving(id))` only when the stored `EnvironmentOrder` contains the id
- [ ] T039 [US4] Add `listEnvironments usecase.ListEnvironments` to `importVariablesWriter` in `internal/application/service/import_variables_writer.go` and to its literal in `internal/application/service/manage_import.go`; in `environment(…)` re-list and load the workspace before any delete or save, compute `OrderReplacing(write.Replaced.ID, write.Environment.ID)` for Replace or `OrderAppending(write.Environment.ID)` otherwise, and write it with `saveWorkspaceFile.Save` after `saveEnvironment.Save` succeeds; a failed write fails the item
- [ ] T040 [US4] Run `make sort-imports`, then `make bdd TAG=@environment_order_placement`, `make bdd TAG=@import_environment_order` and `make bdd TAG=@environments` until green

**Checkpoint**: The user's arrangement is never disturbed by adding, renaming or deleting.

---

## Phase 7: User Story 5 - Reorder from an AI client (Priority: P3)

**Goal**: `env_edit` gains a `reorder` action.

**Independent Test**: Through MCP, move an environment to a position and read the list back.

### Tests for User Story 5 (write first, prove red)

- [ ] T041 [P] [US5] Write `test/integration/features/mcp/environment_reorder.feature` (`@all @xen_011 @mcp @environment_reorder`) with the existing MCP steps: `env_edit` `reorder` moves an environment and returns `items` in the new order; a `variables:changed` event is emitted; `EnvironmentService.List` shows the same order; unknown id → `ENV-01404`; missing index and negative index → `ENV-18400`; `I list the MCP tools` shows `reorder` among `env_edit`'s actions
- [ ] T042 [US5] Run `make bdd TAG=@environment_reorder`; record red

### Implementation for User Story 5

- [ ] T043 [P] [US5] Add `Index *int \`json:"index"\`` and `ReorderEnvironment(localId string) ReorderEnvironment` (nil index → -1) to `internal/integration/entrypoint/dto/mcp_env_edit.go`
- [ ] T044 [P] [US5] Add `validateReorderEnvironment` with code `ENV-18400` to `internal/integration/entrypoint/controller/validate_environment.go`
- [ ] T045 [US5] In `internal/integration/entrypoint/controller/mcp_env_edit.go` add `mcpEnvEditReorderAction = "reorder"` after `duplicate` in `envEditAITool.Actions`, update the tool description per contracts/mcp-env-edit.md, and add the `reorder` case calling `manageEnvironmentService.Reorder` and returning `dto.NewMCPToolResult(dto.NewMCPEnvList(*list))`
- [ ] T046 [US5] Run `make bdd TAG=@environment_reorder` and `make bdd TAG=@mcp` until green (no existing feature or Go test lists `env_edit`'s actions verbatim — verified on `origin/main` — so no existing expectation changes)

**Checkpoint**: AI clients reorder exactly like the UI.

---

## Phase 8: Polish & Cross-Cutting Concerns

- [ ] T047 [P] Add the environment reorder to `docs/prd/xenon-mvp-prd.md`: a "Drag and drop" bullet ("Environments can be reordered in the Environments view, by mouse or keyboard; Globals stays first; disabled while searching.") and a decisions-log entry for XEN-011 (ids in `workspace.yaml` → `environments.order`, new environments last, Replace keeps the slot, git merge like `order`)
- [ ] T048 Run `make bdd` (full suite) and `cd frontend && npx vitest run` until green
- [ ] T049 Run `/refactor-diff` until clean, then `make validate`
- [ ] T050 Walk through `specs/XEN-011-environment-reorder/quickstart.md` steps 1–9 in `make dev`

---

## Dependencies & Execution Order

### Phase dependencies

- **Setup (Phase 1)** → **Foundational (Phase 2)** → user stories.
- **US1 (Phase 3)** depends on Phase 2. It is the MVP.
- **US2 (Phase 4)** depends on Phase 2 (and on US1's `Reorder` for the git scenarios that reorder through the binding).
- **US3 (Phase 5)** depends on US1's sortable view (T022).
- **US4 (Phase 6)** depends on Phase 2 and US1's service changes (T016 shares the file).
- **US5 (Phase 7)** depends on US1's `ManageEnvironmentService.Reorder` (T015–T016).
- **Polish (Phase 8)** after all stories.

### Within each story

Feature files / Vitest cases first and proven red → domain → application → integration → infra →
frontend → green.

### Parallel opportunities

- T004 and T007 alongside T005/T008.
- In US1: T010–T013 together; T017 and T023 alongside the service work.
- US2's three feature files (T025–T027) together; US2 can proceed in parallel with US3 once US1 is done.
- US4's two feature files (T035, T036) together.
- US5's T043 and T044 together.
- T047 any time.

## Parallel Example: User Story 1

```text
T010 reorder_environments.feature   | T011 store.test.ts   | T012 EnvironmentsView.test.tsx   | T013 EnvironmentSwitcher.test.tsx
then: T015 → T016 → (T017 ∥ T023) → T018 → T019 → T020 → T021 → T022 → T024
```

## Implementation Strategy

### MVP first

1. Phases 1–2, then Phase 3 (US1) → pointer reorder persisted and shown in the switcher.
2. Stop and validate with quickstart steps 1–3.

### Incremental delivery

US2 (git-safe) → US3 (keyboard) → US4 (placement) → US5 (AI) → Polish, each ending on a green
`make bdd TAG=@xen_011`.
