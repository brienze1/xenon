---

description: "Task list for XEN-018 Alias Environments"
---

# Tasks: Alias Environments

**Input**: Design documents from `specs/XEN-018-alias-environments/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/ (bindings.md, mcp.md, ui.md), quickstart.md

**Tests**: MANDATORY (constitution V, BDD-first). Every story starts with Godog feature files in
`test/integration/features/` (tag `@alias_environments`) that must be proven red before
implementation; frontend behaviour gets Vitest tests. Write features against the existing generic
step catalog in `test/integration/steps/` (the XEN-016 alias features show the steps for request
files, binding calls, sends, history and MCP calls); add a step only when no existing step can
express the assertion, and keep it generic.

**Organization**: grouped by user story (spec.md). Layers inside each story go innermost outward:
domain → application → integration → infra → frontend.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel (different files, no dependency on an unfinished task)
- **[Story]**: US1…US4 from spec.md

---

## Phase 1: Setup

- [ ] T001 Work on branch `xen-018` (fresh off `origin/main`); in a fresh worktree run `cd frontend && npm ci` and `make bindings` before any Vitest run; confirm `make bdd` is green before changes

---

## Phase 2: Foundational (blocking all stories)

**Purpose**: the link on the alias, its file format, and the read path every story needs.

- [ ] T002 Write red Go unit tests in `internal/domain/entity/alias_environments_test.go` and `internal/domain/entity/request_alias_test.go`: `NormalizedEnvironmentIDs` (trim, drop blanks, dedupe keeping first order, nil when empty); `AliasEnvironmentLabels` (current name per id, `Deleted environment` for an unknown id); `IsAvailableIn` (unlinked → true for nil and any environment; linked → true only when the active id is in the set; nil active → false; set of two); `WithEnvironments` normalizes; `Edited` and `Renamed` keep `EnvironmentIDs`; `NewRequestAlias(id, name, of, ids)` stores normalized ids
- [ ] T003 Implement `NormalizedEnvironmentIDs` and `AliasEnvironmentLabels` (constant `deletedEnvironmentLabel = "Deleted environment"`) in `internal/domain/entity/alias_environments.go`
- [ ] T004 Add `EnvironmentIDs []string`, `IsLinked`, `IsAvailableIn(active *Environment)` and `WithEnvironments` to `internal/domain/entity/request_alias.go`, and change `NewRequestAlias` to take `environmentIds []string`. Update its callers (`internal/application/service/manage_alias.go` passes `nil` until T019, and any test helpers). T002 goes green
- [ ] T005 Add `Environments []string \`yaml:"environments"\`` to `internal/integration/persistence/model/request_alias.go`: `NewRequestAlias` copies `EnvironmentIDs`; `ToEntity` sets `EnvironmentIDs` from it; `Node()` emits `yamlnode.StringList("environments", …)` right after `name`, and only when non-empty
- [ ] T006 Extend `internal/integration/persistence/model/model_test.go` with an alias carrying `environments` (round-trip, key order). Assert the existing request-with-aliases contract YAML (no `environments`) is still written byte-for-byte
- [ ] T007 [P] Add `EnvironmentIds []string \`json:"environment_ids"\`` to `dto.RequestAlias` in `internal/integration/entrypoint/dto/request_alias.go`, filled by `NewRequestAlias` in `internal/integration/entrypoint/dto/request_alias_constructor.go` (always a non-nil slice). `RequestAlias.ToEntity` leaves `EnvironmentIDs` empty: `Save` ignores it (contracts/bindings.md) and keeps the stored set through `Edited`
- [ ] T008 [P] Add `environmentIds: string[]` to `SavedAlias` in `frontend/src/platform/mappers/request.ts` (+ cases in `frontend/src/platform/mappers/request.test.ts`) after `make bindings`

**Checkpoint**: a request file with `environments` under an alias loads, `RequestService.Get` returns `environment_ids`, and files without it are unchanged.

---

## Phase 3: User Story 1 — Limit an alias to environments and be stopped from sending it elsewhere (P1) 🎯 MVP

**Goal**: set an alias's environments from its ⋯ menu; Go refuses to send it in any other environment; the tab blocks Send with a Switch-to banner.

**Independent Test**: link an alias to HML, then send with HML active (sent), Prod active (refused, nothing sent), no environment (refused).

### Tests (red first)

- [ ] T009 [P] [US1] Write `test/integration/features/aliases/alias_environment_storage.feature` (`@alias_environments`), covering the plan.md row: `AliasService.SetEnvironments` writes `environments` under that alias only; `[]` removes the key; duplicates stored once; unknown id → `ALS-10400` and the file unchanged; malformed environment file id → `ALS-10400`; unknown alias → `ALS-01404`; stale `base_revision` → `REQ-02409`; blank `alias_id` → `ALS-11400`; an alias content `Save` keeps the link; a request file without `environments` loads and saves byte-for-byte; deleting a linked environment leaves the request file byte-for-byte unchanged and `RequestService.Get` still returns its id; duplicating the request keeps links; duplicating the environment links nothing new; renaming or recolouring a linked environment leaves the request file byte-for-byte unchanged and the alias still sends with it active (SC-005); Save As from an alias creates a request with no aliases and no `environments` key (spec edge case)
- [ ] T010 [P] [US1] Write `test/integration/features/aliases/alias_environment_send.feature` (`@alias_environments`), covering the plan.md row: linked alias sent with its environment active (HTTP call made); refused with `ALS-09403` and the description naming the environments, with Prod active, with no environment active and with the remembered active environment deleted. Each refusal checks no HTTP call, no history entry and no pre-request script run. Also: a two-environment set (DEV sent, Prod refused); an All environments alias and the Default always sent; a "Confirm before send" active environment refuses the linked alias with `ALS-09403`, not the confirmation error; an alias whose only environment was deleted is refused with every environment
- [ ] T011 [US1] Run `make bdd TAG=@alias_environments` and record that T009/T010 fail for the expected reason (missing binding / send not refused). Commit nothing red to `main`
- [ ] T012 [P] [US1] Write Vitest `frontend/src/features/request-builder/lib/alias-availability.test.ts`, mirroring T002's availability cases and labels (deleted id → `Deleted environment`)
- [ ] T013 [P] [US1] Write Vitest `frontend/src/features/request-builder/components/AliasEnvironmentsDialog.test.tsx`: opens preset from the alias (All, or Only these with its boxes ticked); Save with Only these and no box is disabled with "Pick at least one environment."; a deleted id shows "Deleted environment · Removed when you save" and is not submitted; Save calls `setAliasEnvironments` with the ticked ids in switcher order; an error shows inline; Cancel writes nothing
- [ ] T014 [P] [US1] Write Vitest `frontend/src/features/request-builder/components/UnavailableAliasBanner.test.tsx` and `frontend/src/features/request-builder/components/RequestBar.test.tsx`: with an unavailable alias on screen, Send is disabled and the banner reads "`mx-test-user` is only for HML." with a **Switch to HML** button per existing linked environment (none for deleted ones). Pressing it calls `setActive(id)`. After the active environment becomes HML the banner disappears and Send is enabled. When every linked environment was deleted, the banner reads "… only for environments that were deleted. Change its environments from the Aliases menu." with no button. Available alias / Default: no banner
- [ ] T015 [P] [US1] Extend `frontend/src/features/request-builder/store.test.ts`: `send(tabId)` does not call the send service while the alias on screen is unavailable, and calls it once available; `setAliasEnvironments` calls the service with the tab's revision, replaces `aliases`/`revision` from the response and leaves the alias's draft content untouched; switching the active environment never changes `aliasId` or drafts; linking the alias on screen to another environment (via `setAliasEnvironments`) makes `send` a no-op right away (FR-010 "linked elsewhere")
- [ ] T016 [P] [US1] Extend `frontend/src/platform/services/aliases.test.ts` for `setAliasEnvironments` (DTO shape per contracts/bindings.md, error mapping)

### Implementation

- [ ] T017 [P] [US1] Add `RejectUnavailableAlias(aliasName string, labels []string)` returning `errs.ForbiddenError("the alias \"<name>\" is only for <labels joined by \", \">", "ALS-09403")` in `internal/application/usecase/reject_unavailable_alias.go`
- [ ] T018 [P] [US1] Add `RejectUnknownEnvironment()` returning `errs.BadRequestError("the environment does not exist in this workspace", "ALS-10400")` in `internal/application/usecase/reject_unknown_environment.go`
- [ ] T019 [US1] In `internal/application/service/manage_alias.go` add a `listEnvironments usecase.ListEnvironments` dependency and `SetEnvironments(ctx, localId, requestId, aliasId string, environmentIds []string, baseRevision string, force bool) (*entity.Request, error)`. It runs `load` → `ResolveAlias` → `checkEnvironments(ctx, root, ids)` (normalize; each id must equal the `ID` of a listed environment with an empty `Error`, else `RejectUnknownEnvironment`) → `saveRequestAliases.SaveAliases(root, request.WithAlias(stored.WithEnvironments(ids)))`, logged Started/Finished like `Delete`. Add the method to `internal/application/adapter/manage_alias.go`, and pass `ListEnvironments` in `internal/infra/dependency/application/manage_alias.go`
- [ ] T020 [US1] In `internal/application/service/send_request.go`, right after `s.alias(...)` in `Send`, add: `if !alias.IsAvailableIn(scopes.environment) { return nil, s.unavailableAlias(ctx, workspace.Path, alias) }`. The helper lists the environments once (inject `ListEnvironments` through `NewSendRequestService` and its DI in `internal/infra/dependency/application/`) and returns `usecase.RejectUnavailableAlias(alias.Name, entity.AliasEnvironmentLabels(alias.EnvironmentIDs, environments))`. The Default (`IsDefaultAliasID`) resolves to an unlinked zero alias and is never refused
- [ ] T021 [P] [US1] Add `dto.SetAliasEnvironments` (`local_id`, `request_id`, `alias_id` all `validate:"required,not_blank"`; `environment_ids`; `base_revision`; `LogValue`) in `internal/integration/entrypoint/dto/set_alias_environments.go`
- [ ] T022 [US1] Add `validateSetAliasEnvironments` (`ALS-11400`) to `internal/integration/entrypoint/controller/validate_alias.go`, and `SetEnvironments` to the controller in `internal/integration/entrypoint/controller/alias_manage.go` (same defer/recover/error-handler shape as `Delete`; `force` false) and to the `AliasManageController` interface in `internal/integration/adapter/alias_manage_controller.go`
- [ ] T023 [US1] Add `SetEnvironments(ctx, dto.SetAliasEnvironments) (*dto.Request, error)` to `internal/infra/server/binding/alias.go`; run `make bindings`. T009/T010 go green
- [ ] T024 [P] [US1] Add `setAliasEnvironments` to `frontend/src/platform/services/aliases.ts` (T016 green)
- [ ] T025 [P] [US1] Implement `isAliasAvailable(alias, activeId)` and `aliasEnvironmentLabels(alias, environments)` (returns `{ id, name, color, deleted }[]`) in `frontend/src/features/request-builder/lib/alias-availability.ts` (T012 green)
- [ ] T026 [US1] In `frontend/src/features/request-builder/store.ts` add `setAliasEnvironments(tabId, aliasId, environmentIds)` (returns `AppError | null`). Make `send(tabId)` return early when the tab's alias is unavailable for `useEnvironmentsStore.getState().activeId`. T015 goes green
- [ ] T027 [P] [US1] Implement `frontend/src/features/request-builder/components/EnvironmentPicker.tsx` (+ `EnvironmentPicker.module.css`): All environments / Only these radio; checkboxes per non-malformed environment in switcher order with colour dots; flagged deleted rows; exposes `value: string[]` and `valid` (contracts/ui.md)
- [ ] T028 [US1] Implement `frontend/src/features/request-builder/components/AliasEnvironmentsDialog.tsx` on `EnvironmentPicker` (title "Environments for <alias>", Save → `setAliasEnvironments`, inline error). T013 goes green
- [ ] T029 [US1] In `frontend/src/features/request-builder/components/AliasMenu.tsx`, turn each alias row's ⋯ into a menu with **Environments…** (opens `AliasEnvironmentsDialog`) and **Delete…** (existing `DeleteAliasDialog`)
- [ ] T030 [US1] Implement `frontend/src/features/request-builder/components/UnavailableAliasBanner.tsx` and render it under the request bar from `frontend/src/features/request-builder/components/RequestBar.tsx`, which also disables the Send button (`aria-disabled`, tooltip = banner text) while the alias on screen is unavailable. T014 goes green

**Checkpoint**: US1 independently works: link from ⋯, guard in Go, blocked Send with Switch-to.

---

## Phase 4: User Story 2 — See which environment each alias belongs to (P1)

**Goal**: dots and names in the dropdown; unavailable aliases greyed out and not selectable.

**Independent Test**: open the dropdown under HML, Prod and no environment; check selectability and labels.

- [ ] T031 [P] [US2] Extend `frontend/src/features/request-builder/components/AliasMenu.test.tsx`: a linked row shows one dot + current name per environment (a renamed environment shows its new name; a deleted id shows "Deleted environment"); with HML active the Prod-linked row is disabled (not selectable by click or keyboard) and its accessible name ends ", not available in HML"; with no environment every linked row is disabled; Default and All environments rows stay selectable; the ⋯ menu of a disabled row still opens Environments… and Delete…
- [ ] T032 [US2] Render the environment labels (`aliasEnvironmentLabels`) and the disabled/greyed state (`isAliasAvailable` with `useEnvironmentsStore` `activeId` and `environments`) in `frontend/src/features/request-builder/components/AliasMenu.tsx` and `frontend/src/features/request-builder/components/AliasMenu.module.css`. T031 goes green

**Checkpoint**: US1 + US2 together are the full P1 experience.

---

## Phase 5: User Story 3 — Choose environments when creating an alias (P2)

**Goal**: the Create New Alias dialog sets the link from the start.

**Independent Test**: create an alias with HML picked and one with the preset; read both back.

- [ ] T033 [P] [US3] Add scenarios to `test/integration/features/aliases/alias_environment_storage.feature`: `AliasService.Create` with `environment_ids` writes them; without (or `[]`) writes no `environments` key; an unknown id → `ALS-10400` and no alias created. Prove red
- [ ] T034 [P] [US3] Write Vitest `frontend/src/features/request-builder/components/CreateAliasDialog.test.tsx`: the picker opens on All environments; creating with HML picked calls `createAlias(tabId, name, [hmlId])`; with the preset, `createAlias(tabId, name, [])`; Only these with nothing ticked disables Create; after creating an alias linked to HML while Prod is active, the tab shows it with Send disabled and the banner (US3 scenario 4)
- [ ] T035 [US3] Change `Create` in `internal/application/service/manage_alias.go` and `internal/application/adapter/manage_alias.go` to take `environmentIds []string` (after `name`): `checkEnvironments`, then `entity.NewRequestAlias(id, name, request, ids)`. Update callers in `internal/integration/entrypoint/controller/alias_manage.go` and `internal/integration/entrypoint/controller/mcp_tree_edit_alias.go` (MCP passes `nil` until T040)
- [ ] T036 [US3] Add `EnvironmentIds []string \`json:"environment_ids"\`` to `dto.CreateAlias` in `internal/integration/entrypoint/dto/create_alias.go` (constructor + `LogValue`); the controller passes it. Run `make bindings`. T033 goes green
- [ ] T037 [US3] Add `environmentIds` to `createAlias` in `frontend/src/platform/services/aliases.ts` and in `frontend/src/features/request-builder/store.ts`. Add `EnvironmentPicker` under the name field of `frontend/src/features/request-builder/components/CreateAliasDialog.tsx`. T034 goes green

---

## Phase 6: User Story 4 — AI clients read, set and respect the link (P2)

**Goal**: MCP reads and writes the link; MCP sends are refused like app sends.

**Independent Test**: through MCP, read linked/unlinked aliases, set and clear a link, send a linked alias with another environment active.

- [ ] T038 [P] [US4] Write `test/integration/features/mcp/alias_environments.feature` (`@alias_environments`), covering the plan.md row: `tree get` → `aliases.N.environments` with id + name, `[]` for all, `Deleted environment` for a deleted id; `tree_edit` create with `environment_ids`; update with `environment_ids` sets the link, `[]` clears it, and omitting it keeps it; unknown id → `ALS-10400`; `send` with `alias_id` and Prod active → `ALS-09403`, no approval requested, no HTTP call; AI activity recorded as for other failed sends. Prove red
- [ ] T039 [P] [US4] Add `MCPAliasEnvironment{Id, Name}` and `Environments []MCPAliasEnvironment \`json:"environments"\`` to `internal/integration/entrypoint/dto/mcp_tree_alias.go`. `NewMCPTreeAlias(alias, labels)` and `NewMCPTreeAliases` take the labels; always a non-nil slice
- [ ] T040 [US4] Add `EnvironmentIds *[]string \`json:"environment_ids"\`` to `dto.MCPTreeEdit` in `internal/integration/entrypoint/dto/mcp_tree_edit.go`, with the description from contracts/mcp.md and `used_by:"create,update"`. `CreateAlias` in `internal/integration/entrypoint/dto/mcp_tree_edit_alias.go` carries the ids; add `SetAliasEnvironments(localId string, saved Request) SetAliasEnvironments`
- [ ] T041 [US4] In `internal/integration/entrypoint/controller/mcp_tree_edit_alias.go`, `createAlias` passes the ids. For the `update` action only (`rename` ignores `environment_ids`, matching `used_by:"create,update"`), `saveAlias` calls `manageAliasService.SetEnvironments(..., force true)` after the content save when `EnvironmentIds != nil`. The content save runs as today even when only `environment_ids` was given (unchanged fields keep their stored values through `NewEditedRequestAlias`, so it rewrites the same content). `aliasEdited` resolves labels for the result. Inject environment listing in `internal/infra/dependency/controller/mcp_tree_edit.go`
- [ ] T042 [US4] In `internal/integration/entrypoint/controller/mcp_tree.go`, fill alias environment labels on `get` of a request (list environments once per call; inject in `internal/infra/dependency/controller/mcp_tree.go`). T038 goes green (the send scenario is already green from T020; confirm)

---

## Phase 7: Cross-cutting — History environment and Postman export (FR-015, FR-016)

**Goal**: every send records the active environment; export notes the link.

**Independent Test**: send with and without an active environment and read history; export a request with a linked alias to Postman.

- [ ] T043 [P] Write `test/integration/features/aliases/alias_environment_boundaries.feature` (`@alias_environments`), covering the plan.md row: history entries of Default and alias sends carry `environment` id/name/colour; no `environment` without an active environment; renaming the environment afterwards leaves the entry as recorded; a pre-existing history file without `environment` loads; Postman export of a request with an HML-linked alias → the `Login (mx-test-user)` request description ends with `Xenon alias for environments: HML`, after a blank line when the Default has a description; an All environments alias exports unchanged; a deleted id exports as `Deleted environment`; a Xenon YAML export → import keeps `environments`. Add the MCP `history` environment assertion to `test/integration/features/mcp/alias_environments.feature`. Prove red
- [ ] T044 [P] Write Vitest cases in `frontend/src/features/sidebar/components/HistoryView.test.tsx`: an entry with an environment shows its dot + name after the alias label (or at the start without one), and its accessible label includes ", in HML"; no environment, nothing shown
- [ ] T045 Add `HistoryEnvironment{ID, Name, Color}` in `internal/domain/entity/history_environment.go`; add `Environment *HistoryEnvironment` and `InEnvironment(active *Environment)` to `internal/domain/entity/history_entry.go` (+ unit test in `internal/domain/entity/history_entry_test.go`)
- [ ] T046 Persist it: `internal/integration/persistence/model/history_environment.go` (new, `id`/`name`/`color`) and `Environment *HistoryEnvironment \`json:"environment,omitempty"\`` mapped both ways in `internal/integration/persistence/model/history_entry.go` (send entries only)
- [ ] T047 In `internal/application/service/send_request.go`, chain `.InEnvironment(scopes.environment)` onto the `NewHistoryEntry(...).ForAlias(alias)` call
- [ ] T048 [P] Expose it: `internal/integration/entrypoint/dto/history_environment.go` (new) and `Environment *HistoryEnvironment \`json:"environment,omitempty"\`` in `internal/integration/entrypoint/dto/history_entry.go`. Add the environment name to `internal/integration/entrypoint/dto/mcp_history_line.go`, and the object to `internal/integration/entrypoint/dto/mcp_history_send.go`. Run `make bindings`
- [ ] T049 [P] Map `environment` in `frontend/src/platform/mappers/history.ts`; render the dot + name in `frontend/src/features/sidebar/components/HistoryView.tsx` (+ `HistoryView.module.css`). T044 goes green
- [ ] T050 Change `ExportNode.WithAliasesAsRequests` / `WithAliasRequests` / `AliasRequests` in `internal/domain/entity/export_node.go` to take `labels func(ids []string) []string`. A linked alias appends `Xenon alias for environments: <labels joined by ", ">` to the exported request's `Description` (blank line first when it is non-empty). Unit test it in `internal/domain/entity/export_node_test.go`
- [ ] T051 In `internal/application/service/export_postman.go` `collection`, list environments once through the existing `p.items.listEnvironments` and pass `func(ids) { return entity.AliasEnvironmentLabels(ids, environments) }` to both `WithAliasesAsRequests` and `WithAliasRequests`. The only other caller is the unit test `internal/domain/entity/request_alias_test.go:118`; pass a labels func there too. T043 goes green

---

## Phase 8: Polish & cross-cutting

- [ ] T052 Add "Decisions (Oct 23, 2026): alias environments" to `docs/prd/xenon-mvp-prd.md`, recording the decisions table from spec.md (set of environments, greyed out, keep + block + Switch to, no environment = unavailable, deleted = stays linked and flagged, export note, history environment)
- [ ] T053 Run `make sort-imports`, the full `make bdd` (every XEN-016 alias feature included) and `cd frontend && npx vitest run`; all green
- [ ] T054 Run `/refactor-diff` until clean, then `make validate` (final gate)
- [ ] T055 Walk through `specs/XEN-018-alias-environments/quickstart.md` manual steps 1–10 in `make dev` and confirm each expected outcome

---

## Dependencies & Execution Order

- **Setup (T001)** → **Foundational (T002–T008)** → stories.
- **US1 (T009–T030)** first: it adds `SetEnvironments`, the send guard, the picker and the banner.
- **US2 (T031–T032)** needs T025 (availability helpers) and T029 (row ⋯ menu).
- **US3 (T033–T037)** needs T019 (`checkEnvironments`) and T027 (picker).
- **US4 (T038–T042)** needs T019 (`SetEnvironments`), T020 (guard) and T035 (`Create` signature).
- **Phase 7 (T043–T051)** needs only Foundational; T047 touches `send_request.go` after T020.
- **Polish (T052–T055)** last.

### Parallel opportunities

- T007 ∥ T008 after T005.
- US1 tests T009, T010, T012–T016 all in parallel; T017 ∥ T018 ∥ T021 ∥ T024 ∥ T025 ∥ T027.
- US4 T038 ∥ T039; Phase 7 T043 ∥ T044, T048 ∥ T049.

```text
# US1 tests together
T009 alias_environment_storage.feature   T010 alias_environment_send.feature
T012 alias-availability.test.ts          T013 AliasEnvironmentsDialog.test.tsx
T014 Banner/RequestBar tests             T015 store.test.ts   T016 aliases.test.ts
```

## Implementation Strategy

1. MVP = Phases 1–3 (US1): the guard alone delivers the safety promise (SC-001).
2. Add US2 for the complete P1 dropdown experience, then US3 and US4.
3. Phase 7 for history and export, then Polish and `make validate`.
