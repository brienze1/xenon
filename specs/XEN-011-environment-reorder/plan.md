# Implementation Plan: Reorder Environments With Drag and Drop

**Branch**: `xen-011` | **Date**: 2026-10-15 | **Spec**: [spec.md](spec.md) | **Issue**: [#8](https://github.com/brienze1/xenon/issues/8)

**Input**: Feature specification from `specs/XEN-011-environment-reorder/spec.md`

## Summary

Store an id-keyed environment order in `workspace.yaml` under `environments.order`, arrange the
environment list by it (stored order → other readable environments alphabetically → broken files),
keep it up to date on create / duplicate / import / delete, merge it in git like the collection
`order`, and expose a `Reorder` action on the `EnvironmentService` binding and the MCP `env_edit`
tool. In the UI, the Environments view becomes a `@dnd-kit/sortable` list (pointer + keyboard, the
tab strip's pattern) with Globals pinned and reordering disabled while the search filters; the
switcher follows the store order with no change. Every decision is recorded in
[research.md](research.md).

## Technical Context

**Language/Version**: Go (version per `go.mod`); TypeScript + React (per `frontend/package.json`)

**Primary Dependencies**: Wails v3 bindings, `go.yaml.in/yaml/v3` via `pkg/yamlnode`,
`@dnd-kit/core` 6.3.1 + `@dnd-kit/sortable` 10.0.0 (already installed), Zustand

**Storage**: workspace folder files — `workspace.yaml` gains `environments.order`; environment
files unchanged

**Testing**: Godog BDD (`test/integration/`, `make bdd`), Go unit tests next to entities/models,
Vitest + Testing Library for the frontend; `make validate` as the gate

**Target Platform**: macOS, Windows, Linux desktop (Wails webview)

**Project Type**: desktop app (Go backend + React frontend, single repo)

**Performance Goals**: list re-arrangement is O(n·m) over a handful of environments (n, m < 100);
the UI reorders optimistically before the save returns (SC-002)

**Constraints**: a never-reordered workspace's `workspace.yaml` is byte-for-byte unchanged by
opening/listing (FR-018); a malformed `environments` value must never fail loading the workspace

**Scale/Scope**: workspaces hold tens of environments at most

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first** — the order is a plain YAML list in `workspace.yaml`, committed with the
  workspace; no app-data or network state involved; workspace isolation untouched.
- [x] **II. Secrets** — the stored list holds environment ids only; no names, values or secrets
  are written, logged or announced (announcements use names already visible on screen).
- [x] **III. Go owns logic** — arranging, placement and persistence live in Go (entity + service +
  persistence); React only calls `platform/services/environments.ts` → the Wails binding and
  shows the returned list.
- [x] **IV. Architecture skills** — every file below has its layer per `architecture.md`
  (entity methods in domain, orchestration in the existing `manageEnvironment` service and
  `importVariablesWriter`, merge rule in the existing use-case function, model/DTO/controller in
  integration, binding + DI in infra); frontend files stay inside `features/sidebar`,
  `features/environments` and `platform/` per `frontend-architecture`.
- [x] **V. BDD-first** — the feature files listed under "BDD feature files" lock every FR; only
  the file system (temp workspace roots, read-only paths), the file watcher trigger and git
  remotes are faked, as today.
- [x] **VI. Postman parity** — Postman has no environment reordering; the Environments view is
  already a recorded Xenon decision, and this change adds the PRD bullet + decisions-log entry
  (research R13), so the feature is inside the PRD. Nothing Postman-visible changes (Postman
  exports carry no environment order).
- [x] **VII. Cross-platform** — no paths or shortcuts beyond Space/arrows/Escape (identical on all
  OSes); file writes reuse `fileutils.WriteFileAtomic`.

**Post-design re-check (after Phase 1)**: all seven gates still hold; no complexity exceptions.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-011-environment-reorder/
├── spec.md
├── plan.md              # this file
├── research.md          # R1–R13 decisions
├── data-model.md        # workspace.yaml field, entity/model/DTO changes
├── quickstart.md        # validation guide
├── contracts/
│   ├── environment-service.md   # Wails binding Reorder + changed behaviours
│   ├── mcp-env-edit.md          # env_edit "reorder" action
│   └── environments-view-ui.md  # store action, sortable view, announcements
├── checklists/
└── tasks.md             # /speckit-tasks
```

### Source code changes, by layer

```text
internal/domain/entity/
├── workspace.go                 # + EnvironmentOrder []string (nil = none), WithEnvironmentOrder, LogValue
├── environment_list.go          # NewEnvironmentList(envs, order, activeId) arranges per FR-007; +
│                                #   IDs, OrderMoving, OrderAppending, OrderReplacing, OrderRemoving
└── environment_list_test.go     # NEW unit tests for arranging + order operations

internal/application/
├── adapter/manage_environment.go         # + Reorder(ctx, localId, id string, index int) (*entity.EnvironmentList, error)
├── service/manage_environment.go         # + loadWorkspace / saveWorkspaceFile deps; list() reads the order;
│                                         #   Reorder; Create/Duplicate append; Delete removes when listed
├── service/import_variables_writer.go    # environment(): append id, or swap replaced id in place
├── service/manage_import.go              # importVariablesWriter{…} gets listEnvironments (already a constructor param)
├── usecase/plan_environment_import.go    # NewEnvironmentList(existing, nil, "") — signature follow-up
└── usecase/merge_structural_file.go      # "environments" known field; mergeOrder for "environments.order"

internal/integration/
├── persistence/model/workspace.go        # + Environments yaml.Node (raw) + EnvironmentOrder (yaml:"-");
│                                         #   ToEntity lenient parse; Fields emits environments.order when non-nil
├── persistence/model/workspace_test.go   # NEW unit tests: absent / malformed / round-trip / preserved siblings
├── adapter/environment_manage_controller.go  # + Reorder(ctx, dto.ReorderEnvironment) (*dto.EnvironmentList, error)
├── entrypoint/dto/reorder_environment.go # NEW ReorderEnvironment {local_id, id, index min=0}
├── entrypoint/dto/mcp_env_edit.go        # + Index *int, ReorderEnvironment(localId)
├── entrypoint/controller/environment_manage.go  # + Reorder (ENV-17400), triggerAutoSync Saved
├── entrypoint/controller/validate_environment.go # + validateReorderEnvironment (ENV-18400)
└── entrypoint/controller/mcp_env_edit.go # + "reorder" action → service.Reorder → NewMCPEnvList result

internal/infra/
├── server/binding/environment.go         # + Reorder(ctx, dto.ReorderEnvironment) (*dto.EnvironmentList, error)
└── dependency/application/manage_environment.go  # pass LoadWorkspace + SaveWorkspaceFile (workspace persistence)

frontend/src/
├── platform/services/environments.ts     # + reorderEnvironment(localId, id, index)
├── features/environments/store.ts        # + reorderEnvironment(id, index) optimistic with rollback
├── features/environments/store.test.ts   # + optimistic / rollback / no-op tests
├── features/environments/components/EnvironmentSwitcher.test.tsx  # + order follows store
├── app/bootstrap.test.ts                  # + reorder keeps environment tabs open
├── features/sidebar/components/EnvironmentsView.tsx         # DndContext + SortableContext, sortable rows,
│                                                             #   Globals pinned, disabled while filtering/renaming
├── features/sidebar/components/EnvironmentsView.module.css  # dragging / drop-indicator styles (tokens only)
└── features/sidebar/components/EnvironmentsView.test.tsx    # + pointer drag, keyboard drag, announcements,
                                                              #   filter-disabled, Globals not draggable, error banner

docs/prd/xenon-mvp-prd.md                 # Drag and drop bullet + decisions-log entry (research R13)
```

`frontend/bindings/**` is regenerated with `make bindings` (never hand-edited).

**Structure Decision**: single repo, existing layout. No new packages: every change extends an
existing file in its layer, plus one DTO file and two unit-test files.

### Backend design notes (what each change must do)

1. **Entity** (`environment_list.go`): arrangement and order operations exactly per
   data-model.md / research R5–R6. `OrderMoving` clamps `index` to `len(IDs())-1` after removal
   (index ≥ count → last).
2. **Model** (`model/workspace.go`): `ToEntity` sets `EnvironmentOrder` from the raw node only when
   it is a mapping whose `order` is a sequence of scalar nodes; otherwise `nil`. `NewWorkspace`
   copies the entity order into the `yaml:"-"` field; `Fields()` emits
   `yamlnode.Object("environments", yamlnode.StringList("order", order))` right after the `order`
   field (so `Patch` inserts it next to `order` when absent) only when that field is non-nil. Because tree `writeOrder`, staging and globals save all go through
   `model.NewWorkspace(stored.ToEntity()…)`, they preserve the order with no further change — a
   BDD scenario locks that a collection reorder and a globals edit keep `environments.order`.
3. **Service** (`manage_environment.go`):
   - `list(ctx, root, activeId)` loads the workspace (`loadWorkspace.Load`) and calls
     `entity.NewEnvironmentList(environments, workspace.EnvironmentOrder, activeId)`.
   - `writeOrder(ctx, root, order)`: load → `WithEnvironmentOrder(order)` →
     `saveWorkspaceFile.Save`.
   - `Reorder(ctx, localId, id, index)`: resolve workspace; `list`; `!list.Contains(id)` →
     `errs.NotFoundError("the environment does not exist", "ENV-01404")`; `order :=
     list.OrderMoving(id, index)`; write only when `order` differs from `list.IDs()` (a no-op
     move never writes and never materializes an order); publish `VariablesChanged` with the
     environment's scope; return the re-listed `EnvironmentList` with the active id.
   - `Create` / `Duplicate`: before `saveNamed`, list (`m.list`) and compute
     `order := list.OrderAppending(newId)` from that pre-save list (the new id is not in it, so it
     is appended exactly once), then after the save succeeds `writeOrder(order)`.
   - `Delete`: before deleting the file load the workspace and the list; after the file is
     deleted, if the workspace's stored `EnvironmentOrder` contains the id,
     `writeOrder(list.OrderRemoving(id))` (displayed positions of the others are unchanged);
     otherwise no write.
   - Error from an order write after a successful create/duplicate is returned (the environment
     file exists; listing shows it in the alphabetical tail — a valid state, retried by the next
     add or reorder).
4. **Import** (`import_variables_writer.go` → `environment`): import targets are resolved once
   for the whole batch (`importTargets.resolve`), so `target.environments` is stale after the
   first item. At the start of `environment(…)` — **before** the replaced file is deleted and the
   new one saved — the writer re-lists with a new `listEnvironments usecase.ListEnvironments`
   dependency (already a parameter of `NewManageImportService`; add it to the
   `importVariablesWriter{…}` literal in `internal/application/service/manage_import.go`), loads
   the workspace, builds `entity.NewEnvironmentList(fresh, workspace.EnvironmentOrder, "")` and
   computes `OrderReplacing(write.Replaced.ID, write.Environment.ID)` when replacing, else
   `OrderAppending(write.Environment.ID)`. After `saveEnvironment.Save` succeeds it writes that
   order. A failed order write fails the item with that error (the environment file stays; it
   lists in the alphabetical tail).
   Whole-workspace Xenon imports (`importWorkspaceWriter`) copy the folder, `workspace.yaml`
   included, so the source's `environments.order` is kept as-is — no change there.
5. **Merge** (`merge_structural_file.go`): add `"environments": true` to `knownStructuralFields`
   and `const environmentOrderField = "environments.order"`; call `mergeOrder` when
   `path == orderField || path == environmentOrderField`.
6. **Controller / binding / MCP**: per contracts. `Reorder` in `environment_manage.go` follows the
   controller skill template (logger Started/Finished, panic recovery, error handler, validator
   with `ENV-17400`, `triggerAutoSync(ctx, e.autoSyncService.Saved, request.LocalId)`).

### BDD feature files (written red by `/create-tests`)

All tagged `@all @xen_011`, reusing only existing steps (`I call "EnvironmentService.Reorder"
with`, `the workspace file … should contain the field "environments.order" with the items …`,
`… should be unchanged`, `… should not contain the field "environments"`, `the response should
contain the field "environments.0.name" equal to …`, `the workspace path … is read-only`,
`the file watcher reports a change to the workspace path …`, `I call the MCP tool "env_edit"
with`, git clone/commit/pull steps from `features/git/`). No new step definitions are needed.

| Feature file | Locks |
| --- | --- |
| `test/integration/features/environments/reorder_environments.feature` | FR-006, FR-007 (stored order first, stale ids ignored, duplicates first-wins, unlisted alphabetical, broken last, malformed value → alphabetical and workspace still loads), FR-011, FR-016 (active id unchanged, environment files unchanged), FR-018 (list never writes), reorder no-op writes nothing, index past the end → last, negative index → `ENV-17400`, unknown id → `ENV-01404`, FR-015 read-only workspace folder → error and order unchanged |
| `test/integration/features/environments/environment_order_placement.feature` | FR-009 (create/duplicate last, incl. never-reordered materialization), FR-012 (delete removes listed id; delete in a never-ordered workspace leaves `workspace.yaml` unchanged) |
| `test/integration/features/environments/environment_order_preserved.feature` | FR-019: collection reorder (`TreeService.Move` at workspace level), a globals edit and a workspace rename keep `environments.order` and its unknown sibling keys |
| `test/integration/features/import/import_environment_order.feature` | FR-010 (import no-clash and Keep both appended in import order — two environments in one import both land last, in order; Replace takes the replaced slot); whole-workspace Xenon import keeps the source's `environments.order` |
| `test/integration/features/file_changes/environment_order_changes.feature` | FR-013 (external edit of `workspace.yaml` → `variables:changed` emitted and `List` returns the new order) |
| `test/integration/features/git/conflicts_environment_order.feature` | FR-014 / SC-003 (two clones reorder differently; pull merges `environments.order` with no conflict; an addition on one side and a deletion on the other combine) |
| `test/integration/features/mcp/environment_reorder.feature` | FR-017 (`env_edit` reorder moves and returns the list, emits `variables:changed`; unknown id `ENV-01404`; missing/negative index `ENV-18400`; `reorder` listed in the tool's actions) |

### Frontend tests (Vitest)

| Test file | Locks |
| --- | --- |
| `features/sidebar/components/EnvironmentsView.test.tsx` | FR-001 pointer drag calls `reorderEnvironment` with the target index; FR-002 keyboard Space/ArrowUp/Space and Escape cancel; announcements text with positions; FR-003 Globals not draggable / first; FR-004 broken rows not draggable; FR-005 filter disables drag and Space pick-up; renaming row not draggable; FR-015 failed reorder restores the list (store) and shows the error banner |
| `features/environments/store.test.ts` | optimistic update, rollback + error on failure, no call for a no-op move |
| `features/environments/components/EnvironmentSwitcher.test.tsx` | FR-008 menu order equals store order (non-alphabetical) |
| `app/bootstrap.test.ts` | FR-016 a reorder keeps open environment tabs open and clean |

## Complexity Tracking

No constitution violations; nothing to justify.
