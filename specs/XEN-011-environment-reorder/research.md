# Research: Reorder Environments With Drag and Drop (XEN-011)

Every decision below is settled from the code on `origin/main` (93f5381) or by the user in
`spec.md` → Clarifications. Nothing is left for the implementer to decide.

## R1 — What identifies an environment in the stored order

- **Decision**: environment **ids** (ULIDs from each environment file's `id:` field).
- **Rationale**: renaming an environment renames its file (`manageEnvironment.saveNamed` →
  `DeriveFileName`, then `environmentPersistence.Save` removes the old file when
  `Environment.IsMovedFrom`). An id-keyed order survives renames with no rewrite; issue #8 itself
  speaks of "ids in the order". The collection `order` is keyed by file/folder names and has to
  be rewritten on rename (`TreeNode.OrderRenaming`); ids avoid that class of bug.
- **Alternatives**: file names (rejected: churn on every rename); names (rejected: case-insensitive
  uniqueness rules make them a poor key).

## R2 — Where and under which key the order is stored

- **Decision**: `workspace.yaml`, nested mapping `environments:` with key `order:` — a list of ids
  (user decision, Clarifications). The collections' top-level `order:` is untouched.
- **Rationale**: the workspace file already holds workspace-level arrangement (`order`) and is
  committed and synced; the nested mapping leaves room for future environment-level workspace
  settings.
- **Shape on disk**:

  ```yaml
  order:
    - billing-api
  environments:
    order:
      - 01JA00000000000000000000P1
      - 01JA00000000000000000000D1
  variables: []
  ```

## R3 — Reading the field leniently

- **Finding**: `readWorkspace` (`internal/integration/persistence/workspace.go`) decodes the whole
  document into `model.Workspace` and fails the workspace with `WKS-16422` ("invalid fields") when
  a typed field does not fit. A hand-edited `environments: oops` would therefore break the whole
  workspace if the field were typed `struct{ Order []string }`.
- **Decision**: `model.Workspace.Environments` is a `yaml.Node` (raw); `ToEntity` extracts
  `environments.order` only when `environments` is a mapping whose `order` is a sequence of
  scalars, and otherwise yields **nil** (no order → alphabetical). This satisfies the spec's
  "malformed order value" edge case without breaking the workspace.
- **Older Xenon builds**: `Decode` is non-strict and `yamlnode.Patch` only touches the keys it is
  given, so an older build reading or saving this `workspace.yaml` ignores and preserves
  `environments:`.

## R4 — "Absent" vs "empty" order, and never writing on list

- **Decision**: `entity.Workspace.EnvironmentOrder []string` is `nil` when the file has no usable
  `environments.order`, and a (possibly empty) slice when it has one. `model.Workspace.Fields()`
  emits `yamlnode.Object("environments", yamlnode.StringList("order", ids))` right after the
  `order` field **only when `EnvironmentOrder != nil`**; when nil, no `environments` field is emitted, so `Patch` leaves the
  key (or its absence, or a malformed value) exactly as it is on disk.
- **Rationale**: every writer of `workspace.yaml` round-trips through
  `model.NewWorkspace(stored.ToEntity()…)` — globals save (`workspacePersistence.Save`), the tree's
  workspace-level `writeOrder` (`internal/integration/persistence/tree.go`), staging
  (`staging.go`) and rename. With the field on the entity and emitted only when non-nil, none of
  them can drop or invent the order, and a workspace that was never reordered is byte-for-byte
  unchanged (FR-018, SC-004). Listing never writes.
- Because `Object` patches an existing mapping in place (`yamlnode.Patch` recurses into
  `Fields` when the current node is a mapping), unknown sibling keys under `environments:` are
  preserved; a non-mapping malformed value is replaced on the next write.

## R5 — How the arranged list is computed

- **Decision**: `entity.NewEnvironmentList(environments, order, activeId)` arranges:
  1. readable environments whose id appears in `order`, in `order`'s sequence (first occurrence
     wins, ids with no readable environment skipped);
  2. remaining readable environments by today's `sortKey()` (lower-cased name, then file name);
  3. malformed environments, by file name.
  `activeId` handling is unchanged.
- **Rationale**: matches FR-007 and the spec's broken-rows-last rule; mirrors the existing
  `arrangedNames` helper (`internal/domain/entity/item_order.go`) but keyed by id.
- The only other caller, `withEnvironmentFileName` in
  `internal/application/usecase/plan_environment_import.go`, only needs sibling file names and
  passes `nil` order.

## R6 — Order-changing operations (all in `entity.EnvironmentList`)

Mirroring the tree's `TreeNode.OrderAppending / OrderRemoving / OrderMoving / OrderImporting`:

| Operation | New stored order |
| --- | --- |
| Reorder (move `id` to `index`) | arranged readable ids with `id` removed, then inserted at `index` (clamped to the end) |
| Create / duplicate / import (no clash or Keep both) | arranged readable ids + new id appended |
| Import "Replace" | arranged readable ids with the replaced id swapped for the new id in place |
| Delete | arranged readable ids minus the id (`IDs()` without it; displayed positions of the others unchanged) — written **only** when the stored order names the id |
| Rename / update / variable edits | unchanged (no write) |

- **Decision on "never reordered" workspaces**: create, duplicate and import materialize the
  order (arranged list + new id) — user decision (Clarifications), same as collections, whose
  `OrderAppending` appends to the full `ChildOrder()`.
- **Index semantics**: `index` is the 0-based position among readable environments **after** the
  move (the position the moved environment ends at), identical to `TreeService.Move`'s `index`.
  An index ≥ the number of readable environments places it last. A negative index is rejected by
  validation.
- A reorder that produces the same order as the current arranged list is still accepted and
  writes nothing (no-op save avoided by comparing before writing).

## R7 — Git merge of the order

- **Finding**: `MergeStructuralFile` (`internal/application/usecase/merge_structural_file.go`)
  merges `workspace.yaml` field by field; scalar lists at path `order` go through `mergeOrder`
  (theirs' sequence, then mine's additions; deletions on either side honoured). Any other
  conflicting path in a non-environment file whose top field is not in `knownStructuralFields` is
  silently resolved as "kept theirs"; a known one becomes a field conflict.
- **Decision**: add `"environments": true` to `knownStructuralFields` and apply `mergeOrder` when
  the path is `order` **or** `environments.order` (constant `environmentOrderField =
  "environments.order"`). Result: two clones that reorder differently merge without a conflict
  (FR-014, SC-003); additions on one side and deletions on the other combine as for collections.

## R8 — Refreshing the UI on external changes

- **Finding**: `variableChangesAt` (`internal/integration/persistence/tree_change.go`) maps any
  change to `workspace.yaml` to `VariableScopes: [globals]`; `handleFileChange` then publishes
  `variables:changed`; the frontend's `handleVariablesChanged`
  (`frontend/src/features/environments/store.ts`) calls `load()`, which re-lists environments.
- **Decision**: no new event or watcher wiring. FR-013 is already met once the list honours the
  stored order; a BDD scenario locks it.

## R9 — Events after an in-app or AI reorder

- **Decision**: `manageEnvironment.Reorder` publishes `variables:changed` with the moved
  environment's scope, as `Create`/`Duplicate` do via `announced`. This refreshes the switcher in
  the window when an AI client reorders (US5) and is harmless for an in-app reorder (the store
  re-lists the same order). The binding controller triggers auto-sync like `Update`
  (`triggerAutoSync(ctx, autoSyncService.Saved, localId)`).

## R10 — Keyboard reordering pattern

- **Finding**: the collection tree (`CollectionTree.tsx`) has **no** keyboard reordering — only a
  `PointerSensor`. The tab strip (`frontend/src/features/tabs/components/TabStrip.tsx`) has it:
  `@dnd-kit/sortable` + `KeyboardSensor` with `sortableKeyboardCoordinates`, `KEYBOARD_CODES`
  (Space picks up; Space/Enter drops; Escape cancels), screen-reader instructions and
  announcements.
- **Decision**: the Environments view follows the **tab strip** pattern with
  `verticalListSortingStrategy` (issue #8's "as the collection tree allows" is corrected by this
  finding). Announcements name the environment and its position: `Picked up {name}.`,
  `{name} moved to position {n} of {total}.`, `{name} dropped at position {n} of {total}.`,
  `Reordering cancelled; {name} is back at position {n} of {total}.`

## R11 — Optimistic UI and failure

- **Decision**: the environments store's new `reorderEnvironment(id, index)` sets the moved list
  immediately (SC-002), calls `reorderEnvironment` in `platform/services/environments.ts`, and on
  success replaces the list with the returned one; on failure restores the previous list and
  returns the `AppError`, which `EnvironmentsView` shows in its existing error banner with the
  existing `ACTION_NEXT_STEP` (FR-015).

## R12 — Error codes

Free `ENV-` sequence numbers (used today: 01–16): `ENV-17400` — reorder request validation in the
binding controller; `ENV-18400` — reorder validation in the MCP controller. Not found reuses
`ENV-01404` (from `FindEnvironment`). A failed `workspace.yaml` write surfaces the existing
`WKS-25500`/`WKS-26500` codes from `workspacePersistence.Save`.

## R13 — Constitution VI (PRD scope)

- **Finding**: Postman has no environment reordering; the Environments view is already a recorded
  Xenon decision (PRD decisions log, "Where are environments managed?").
- **Decision**: the user opened issue #8, which is the product decision. The PRD is updated in the
  same change: one bullet in "Collections and folders → Drag and drop" ("Environments can be
  reordered in the Environments view; Globals stays first.") and one decisions-log entry
  recording XEN-011's choices (ids in `environments.order`, new environments last, disabled while
  filtering). With that, the feature is inside the PRD and Principle VI holds.
