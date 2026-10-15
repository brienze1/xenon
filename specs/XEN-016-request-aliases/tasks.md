---

description: "Task list for XEN-016 Request Aliases"
---

# Tasks: Request Aliases

**Input**: Design documents from `specs/XEN-016-request-aliases/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/ (bindings.md, mcp.md, ui.md), quickstart.md

**Tests**: MANDATORY (constitution V, BDD-first). Every user story starts with Godog feature files in
`test/integration/features/` (tag `@aliases`) that must be proven red before implementation; frontend
behaviour gets Vitest tests. Write features against the existing generic step catalog in
`test/integration/steps/`; add a step only when no existing step can express the assertion, and keep
it generic.

**Organization**: grouped by user story (spec.md). Layers inside each story go innermost outward:
domain → application → integration → infra → frontend.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel (different files, no dependency on an unfinished task)
- **[Story]**: US1…US7 from spec.md

---

## Phase 1: Setup

- [ ] T001 Branch work on `xen-016` (fresh off `origin/main`); in a fresh worktree run `cd frontend && npm ci` and `make bindings` before any Vitest run (see `fresh-worktree-setup` note); confirm `make bdd` is green before changes

---

## Phase 2: Foundational (blocking all stories)

**Purpose**: the alias model, its file format and the read path every story needs.

- [ ] T002 Write red Go unit tests for the alias domain in `internal/domain/entity/request_alias_test.go`: `NewRequestAlias` copies only sendable fields; `Request.Version("")` is the request, `Version(id)` swaps sendable content and keeps id/name/description/examples/revision; `HasAliasNamed` ignores case and surrounding spaces and honours `exceptId`; `HasReservedName` for `default`/` DEFAULT `; `WithAliasAppended` / `WithAlias` / `WithoutAlias` return copies; `CopyAs` keeps `Aliases`
- [ ] T003 Implement `RequestAlias` (`ID`, `Name`, `Version Request`; `NewRequestAlias`, `HasValidName`, `HasReservedName`, `Renamed`, `Edited`) in `internal/domain/entity/request_alias.go`
- [ ] T004 Add `Aliases []RequestAlias` and `Alias`, `Version`, `HasAliasNamed`, `WithAliasAppended`, `WithAlias`, `WithoutAlias`, `SendableOf` to `internal/domain/entity/request.go` (data-model.md) — T002 green
- [ ] T005 [P] Implement the YAML model `RequestAlias` (`NewRequestAlias`, `ToEntity`, `fields()` reusing the request node helpers; scripts/inheritScripts written under the same rule as the request) in `internal/integration/persistence/model/request_alias.go`
- [ ] T006 Add `Aliases []RequestAlias \`yaml:"aliases" json:"-"\`` to `internal/integration/persistence/model/request.go`; map it in `NewRequest`/`ToEntity`; `Fields()` appends `optionalAliases` (Reserved, nothing when empty) after examples; add `FieldsWithAliases()` writing `aliasesField` (`Absent` when empty); key order `examples`, `aliases`, `postman`, `bruno`
- [ ] T007 Extend `internal/integration/persistence/model/model_test.go` with a request-with-aliases contract YAML and assert the existing `contractRequestYaml` (no aliases) is still written byte-for-byte
- [ ] T008 [P] Add `SaveRequestAliases` interface in `internal/application/usecase/save_request_aliases.go` and `SaveAliases` (writes `FieldsWithAliases()`) in `internal/integration/persistence/request.go`; wire the persistence as that use case in the infra DI
- [ ] T009 [P] Add `ResolveAlias` (`ALS-01404`) in `internal/application/usecase/resolve_alias.go`; `RejectInvalidAliasName` (`ALS-03400`), `RejectReservedAliasName` (`ALS-04400`), `RejectTakenAliasName` (`ALS-05409`) in `internal/application/usecase/reject_invalid_alias_name.go`; `CheckAliasName(request, alias)` in `internal/application/usecase/check_alias_name.go`
- [ ] T010 [P] Add `dto.RequestAlias` (+ constructor) in `internal/integration/entrypoint/dto/request_alias.go` and `Aliases []RequestAlias \`json:"aliases"\`` on `dto.Request` (`NewRequest` fills it, always a non-nil slice; `ToEntity` ignores it) in `internal/integration/entrypoint/dto/request.go`
- [ ] T011 [P] Add `RequestAlias` and `aliases` to `frontend/src/platform/mappers/request.ts` (+ cases in `request.test.ts`); regenerate bindings with `make bindings`

**Checkpoint**: a request file with `aliases:` loads, saves its Default without touching them, and `RequestService.Get` returns them.

---

## Phase 3: User Story 1 — Create an alias and send it (P1) 🎯 MVP

**Goal**: create an alias from the saved Default, switch to it in the tab, send it.

**Independent Test**: create an alias, edit its body, save, send; the send carries the alias body; one sidebar row.

### Tests (red first)

- [ ] T012 [P] [US1] Write `test/integration/features/aliases/alias_storage.feature` (`@aliases`): create copies the **saved** Default (an unsaved edit of the Default is not copied); file shape; each name rule with its code (`ALS-03400` empty and 201 chars, `ALS-04400` `default`/` DEFAULT `, `ALS-05409` existing name in other case); unknown alias → `ALS-01404`; stale `base_revision` → `REQ-02409`; request without aliases is byte-identical after save; deleting the last alias removes the `aliases` key; prove red with `make bdd TAG=@aliases`
- [ ] T013 [P] [US1] Write `test/integration/features/aliases/alias_send.feature` (`@aliases`): sending with `alias_id` sends the alias's method/URL/headers/body; an alias with its own stored pre-request script runs that script (not the Default's) when scripts were not edited in the tab; unknown `alias_id` → `ALS-01404` and the HTTP target receives nothing; the Default's post-response script sets environment variable `token` and the alias's sets `mxToken` — each send sets only its own (FR-015b); an alias with its own OAuth 2.0 (client credentials against the mock token endpoint) gets a token stored under `<requestId>:<aliasId>` without replacing the Default's token, and an alias inheriting auth uses the collection's token (FR-015a); prove red
- [ ] T014 [P] [US1] Write Vitest `frontend/src/features/request-builder/components/AliasMenu.test.tsx`: menu lists Default, aliases in order, separator, "Create new alias…"; selected item checked; create dialog shows Go errors inline and stays open; success switches the version; no menu for unsaved requests or example tabs

### Implementation

- [ ] T015 [US1] Add `ManageAliasService` (`Create(ctx, localId, requestId, name, baseRevision string) (*entity.Request, error)`, `Save(ctx, localId, requestId string, alias entity.RequestAlias, baseRevision string, force bool) (*entity.Request, error)`, `Delete(ctx, localId, requestId, aliasId, baseRevision string) (*entity.Request, error)`) in `internal/application/adapter/manage_alias.go`
- [ ] T016 [US1] Implement `Create` in `internal/application/service/manage_alias.go` following `manage_example.go`: workspace → `FindRevision` + `BlocksSave` (`REQ-02409`) → `findRequest.Find` → `entity.NewRequestAlias(idutils.NewUlid(), name, stored)` → `CheckAliasName` → `WithAliasAppended` → `saveRequestAliases.SaveAliases`; logs Started/Finished
- [ ] T017 [US1] Add `CreateAlias` DTO (+ constructor, `LogValue`, `ToEntity`) in `internal/integration/entrypoint/dto/create_alias.go`, validator in `internal/integration/entrypoint/controller/validate_alias.go`, controller `AliasManageController.Create` in `internal/integration/entrypoint/controller/alias_manage.go` (+ interface in `internal/integration/adapter/alias_manage_controller.go`)
- [ ] T018 [US1] Add binding `AliasService` with `Create` in `internal/infra/server/binding/alias.go`; wire service, controller and binding in the infra DI and Wails service registration
- [ ] T019 [US1] Send: add `AliasId string \`json:"alias_id"\`` to `internal/integration/entrypoint/dto/send_request.go` (constructor + `LogValue`); pass it through `internal/integration/entrypoint/controller/request_send.go`; in `internal/application/service/send_request.go` resolve the stored alias (`usecase.ResolveAlias`, refuse before sending) and use `stored.Version(aliasId)` for `WithScriptsOf` — T013 green
- [ ] T020 [US1] Per-alias OAuth owner (decision R12): add `AliasAuthOwner`/`SplitAuthOwner` in `internal/domain/entity/request_alias.go` (unit tests in `request_alias_test.go`); in `internal/application/service/send_request.go` set `effective.AuthOwner` to the alias owner id when the send has an alias whose auth `IsOwn()`; make `oauthOwners.owner` in `internal/application/service/oauth_owners.go` resolve an owner id with `:` to `request.Version(aliasId).Auth` with the request's ancestors (unknown alias → `ALS-01404`); in the frontend auth panel's token actions pass `<requestId>:<aliasId>` as `owner_id` when an alias with its own auth is shown (`frontend/src/features/auth/components/TokenPanel.tsx`, Vitest in `TokenPanel.test.tsx`) — T013 OAuth scenarios green
- [ ] T021 [P] [US1] Add `createAlias` in `frontend/src/platform/services/aliases.ts` (+ `aliases.test.ts`, mocking the generated binding the way `examples.test.ts` does)
- [ ] T022 [US1] In `frontend/src/features/request-builder/model.ts` add `aliasId`, `aliases: AliasSummary[]`, `versions: Record<string, VersionState>` to `TabDraft`; in `store.ts` load aliases in `openSaved`, add `selectVersion` (swap drafts, no prompt, pin a preview tab through the tabs store) and `createAlias` (pins; updates `baseRevision`); `send` passes `alias_id` of the shown version; every sub-tab (including Scripts and "inherit parent scripts") edits the shown version's draft; tests in `store.test.ts`
- [ ] T023 [US1] Build `AliasMenu.tsx` (+ `.module.css`) and `CreateAliasDialog.tsx` in `frontend/src/features/request-builder/components/` per contracts/ui.md, and render a header row (breadcrumb left, `AliasMenu` right) in `RequestBuilder.tsx` for saved requests only — T014 green
- [ ] T024 [US1] Tab title `<request> · <alias>` in `frontend/src/features/tabs/hooks/use-tab-title.ts` with a Vitest case; the snippet panel already generates from the shown draft — add a Vitest case in `frontend/src/features/side-panel/components/CodePanel.test.tsx` asserting it uses the alias draft when an alias is shown (FR-016)

**Checkpoint**: US1 green — `make bdd TAG=@aliases` (storage create cases + send) and Vitest.

---

## Phase 4: User Story 2 — Edit versions independently (P1)

**Goal**: save an alias or the Default without touching the others; per-version drafts.

**Independent Test**: save different content in the Default and an alias; reopen; each holds only its own.

### Tests (red first)

- [ ] T025 [P] [US2] Write `test/integration/features/aliases/alias_independence.feature` (`@aliases`): saving an alias changing each sendable field (method, URL path version, params, path variables, headers, disabled auto-headers, body, auth, scripts, inheritScripts, settings) leaves the Default and a second alias unchanged; saving the Default leaves every alias unchanged; saving an alias without edited scripts keeps its stored scripts; prove red
- [ ] T026 [P] [US2] Vitest in `frontend/src/features/request-builder/store.test.ts` and `frontend/src/features/tabs/store.test.ts`: switching keeps both drafts; the tab is dirty while any version is dirty; Cmd/Ctrl+S saves only the shown version; close prompt Save calls `saveAllVersions` (Default first, then aliases in order) and Don't Save discards all; "changed on disk" banner on an alias save conflict, where Keep Your Version force-saves the shown version and Load from Disk replaces only the shown version while the others keep their drafts and every version adopts the new revision (FR-009); a failing save in the close prompt stops, keeps earlier saves and leaves the tab open with the error

### Implementation

- [ ] T027 [US2] Implement `Save` in `internal/application/service/manage_alias.go` (revision check unless `force`, `ResolveAlias`, `stored.Edited(edit)` (keeps the stored alias scripts unless `edit.ScriptsEdited`; unit case in `request_alias_test.go`) + `Renamed`, `CheckAliasName` excluding itself, `WithAlias`, `SaveAliases`)
- [ ] T028 [US2] Add `SaveAlias` DTO in `internal/integration/entrypoint/dto/save_alias.go`, validator, controller `Save`, binding `AliasService.Save` — T025 green
- [ ] T029 [US2] Add `saveAlias` to `frontend/src/platform/services/aliases.ts`; in `frontend/src/features/request-builder/store.ts` route `save` to `saveAlias` when an alias is shown, add `saveAllVersions` (stops at the first error), extend `isDirty`, make the conflict banner's Keep Your Version / Load from Disk act on the shown version only (FR-009); in `frontend/src/features/tabs/store.ts` make the close prompt's Save call `saveAllVersions` — T026 green

---

## Phase 5: User Story 3 — Rename and delete aliases (P2)

**Goal**: rename through the alias crumb (XEN-012), delete from the ⋯ menu.

**Independent Test**: rename via crumb + Cmd/Ctrl+S, delete via ⋯; file holds the new name and no deleted alias.

### Tests (red first)

- [ ] T030 [P] [US3] Add to `test/integration/features/aliases/alias_storage.feature`: rename through `Save` with a new name (trimmed), rename clash/reserved/length errors write nothing, delete removes one alias and keeps the others and the Default, delete removes the alias's OAuth token, delete with a stale `base_revision` → `REQ-02409`; prove red
- [ ] T031 [P] [US3] Vitest: `frontend/src/features/sidebar/components/ItemBreadcrumb.test.tsx` (alias crumb is last, request name is plain text while an alias is shown); `AliasMenu.test.tsx` (⋯ on alias rows only, Delete… confirm/cancel, deleting the shown alias selects Default); request-builder store (alias name pending → dirty; Escape cancels; save sends the name)

### Implementation

- [ ] T032 [US3] Implement `Delete` in `internal/application/service/manage_alias.go` (also `oauthTokens.delete` for the alias owner id; a missing token is not an error; BDD case added to `alias_storage.feature` in T030); `AliasReference` DTO in `internal/integration/entrypoint/dto/alias_reference.go`; controller `Delete`; binding `AliasService.Delete` — T030 green
- [ ] T033 [US3] Add `deleteAlias` to `frontend/src/platform/services/aliases.ts`; `deleteAlias` store action (drops the version, selects Default if shown, pins a preview tab); `DeleteAliasDialog.tsx` and the ⋯ submenu in `AliasMenu.tsx`
- [ ] T034 [US3] Add the `aliasEditor` crumb to `frontend/src/features/sidebar/components/ItemBreadcrumb.tsx` and bind it in `RequestBuilder.tsx` to the XEN-012 editable name over the alias name draft (`editName` / cancel / inline errors from `ALS-03400/04400/05409`); while an alias is shown, `frontend/src/features/side-panel/components/InfoPanel.tsx` shows the request name as plain text (Vitest case in `InfoPanel.test.tsx`) — T031 green

---

## Phase 6: User Story 4 — The tab remembers its alias (P2)

**Goal**: restore the shown alias and per-version drafts after restart; sidebar open focuses as-is; new tab starts on Default.

**Independent Test**: select an alias, restart, the restored tab shows it; close and reopen → Default.

### Tests (red first)

- [ ] T035 [P] [US4] Write `test/integration/features/aliases/alias_tab_state.feature` (`@aliases`): UI state saved with a tab `alias_id` and `version_drafts` loads back identical; UI state written without them loads with empty values; prove red
- [ ] T036 [P] [US4] Vitest in `frontend/src/features/tabs/store.test.ts`: restore applies `aliasId` + `versionDrafts`; a restored alias missing from the request falls back to Default and drops its draft; `openPreview`/`openPinned` on an open request keep its alias; a new tab starts on Default; a file reload where the shown alias is gone selects Default; selecting, creating or deleting an alias in a preview tab pins it (FR-014a); `duplicateTab` on a tab showing an alias copies `aliasId`, `aliases` and every version draft (FR-012)

### Implementation

- [ ] T037 [US4] Add `AliasID`, `VersionDrafts` to `internal/domain/entity/tab.go` (`LogValue` includes `aliasId`), `alias_id`/`version_drafts` (`omitempty`) to `internal/integration/persistence/model/tab.go` and `internal/integration/entrypoint/dto/tab.go` with mapping both ways — T035 green
- [ ] T038 [US4] Map `aliasId`/`versionDrafts` in `frontend/src/platform/mappers/ui-state.ts` (+ `ui-state.test.ts`); persist and restore them in `frontend/src/features/tabs/store.ts` and the request-builder store; handle the missing-alias fallback on restore and on file reload; make `duplicateTab` copy `aliasId`, `aliases` and `versions` — T036 green

---

## Phase 7: User Story 5 — History shows which alias was sent (P3)

**Goal**: record alias id/name; History label `Request · alias`.

**Independent Test**: send the Default and an alias; History lists both, the alias one with its name.

- [ ] T039 [P] [US5] Add to `test/integration/features/aliases/alias_send.feature`: the History entry of an alias send has `alias_id` and `alias_name`; a Default send and an unsaved-tab send have none; prove red
- [ ] T040 [P] [US5] Vitest in `frontend/src/features/sidebar/components/HistoryView.test.tsx`: label `Login · mx-test-user` for an alias entry, `Login` otherwise; reopening an alias entry opens a draft as today
- [ ] T041 [US5] Add `AliasID`, `AliasName`, `ForAlias` in `internal/domain/entity/history_entry.go`; `alias_id`/`alias_name` in `internal/integration/persistence/model/history_entry.go` and `internal/integration/entrypoint/dto/history_entry.go`; `send_request.go` records `entry.ForAlias(alias)` — T039 green
- [ ] T042 [US5] Map `aliasId`/`aliasName` in `frontend/src/platform/mappers/history.ts` and render the label in `frontend/src/features/sidebar/components/HistoryView.tsx` — T040 green

---

## Phase 8: User Story 6 — Aliases travel with export, import and git (P3)

**Goal**: Postman export expansion; Xenon round-trip; imports create none; git, duplicate, move, Save As, documentation behave as specified.

**Independent Test**: export to Postman and Xenon YAML, import, duplicate, move, commit/pull, apply a documentation tab.

- [ ] T043 [P] [US6] Write `test/integration/features/aliases/alias_boundaries.feature` (`@aliases`): Postman export has `Login` then `Login (test-user)`, `Login (mx-test-user)` with alias content, examples only on `Login`, no secret value; Xenon YAML export → import keeps aliases; Postman, Bruno and cURL imports produce requests whose files have no `aliases` key; duplicate and move carry aliases; Save As with alias content creates a request without aliases and leaves the original unchanged; commit + push + pull of an alias change reaches the other clone; the collection documentation is generated from Defaults and applying it leaves aliases unchanged; prove red (scenarios that pass without code — Xenon export, imports, duplicate, move, git, documentation, Save As — are lock-in scenarios: note in the feature that they guard existing behaviour and confirm they pass once T006 lands)
- [ ] T044 [US6] Implement `ExportNode.WithAliasesAsRequests()` in `internal/domain/entity/export_node.go` (unit test in `internal/domain/entity/export_test.go`) and apply it to `node` as the first statement of `postmanExport.collection` in `internal/application/service/export_postman.go` (before `EffectiveExportRoot`, `FindPostmanDrops` and `FindLiteralCredentials`, so warnings cover alias auth); add a literal-credential-in-alias-auth case to T043 — T043 Postman scenarios green
- [ ] T045 [US6] Save As from an alias in `frontend/src/features/request-builder/store.ts`: `saveAs` sends the shown draft (already the case) — add a Vitest case asserting the shown alias draft is what `SaveAs` receives and the new tab starts on Default

---

## Phase 9: User Story 7 — AI clients work with aliases (P3)

**Goal**: MCP reads, `tree_edit kind:"alias"`, `delete` target `alias`, `send alias_id` (contracts/mcp.md).

**Independent Test**: list, create, update, rename, delete and send an alias through MCP; check the file and History.

- [ ] T046 [P] [US7] Write `test/integration/features/mcp/aliases.feature` (`@aliases`): `tree` request read shows `aliases` (secrets masked like the Default); `tree_edit` `kind: alias` `create`/`update`/`rename` with each error (`ALS-01404`, `ALS-03400`, `ALS-04400`, `ALS-05409`, unsupported action → invalid arguments); `delete` target `alias`; `send` with `request_id` + `alias_id` (History source AI + alias), `alias_id` alone → invalid arguments, unknown alias → `ALS-01404` and nothing sent; each change recorded as AI activity; prove red
- [ ] T047 [US7] Include aliases in the MCP request read in `internal/integration/entrypoint/controller/mcp_tree.go` / its DTO, through the existing leak guard
- [ ] T048 [US7] Add `internal/integration/entrypoint/controller/mcp_tree_edit_alias.go` (create / update / rename via `ManageAliasService`) and route `kind: "alias"` in `mcp_tree_edit.go`; reject other actions for that kind
- [ ] T049 [US7] Add the `alias` target to `internal/integration/entrypoint/controller/mcp_delete.go` (`id` + `owner_id`)
- [ ] T050 [US7] Add `AliasId` to `internal/integration/entrypoint/dto/mcp_send_target.go` and pass it from `resolveMCPSendTarget` (`mcp_send_target.go`) into the send; refuse `alias_id` without `request_id`; update the tool descriptions/schemas for `tree_edit`, `delete` and `send` — T046 green

---

## Phase 10: Polish & cross-cutting

- [ ] T051 Record the decision in `docs/prd/xenon-mvp-prd.md` → "Decisions (Oct 15, 2026)": request aliases (XEN-016, issue #19) — Postman has no equivalent; user decision; full-copy model, dropdown, crumb rename, Delete-only menu, one tab, history, Postman export as `Request (alias)`, MCP parity, documentation Defaults only, per-alias scripts and per-alias OAuth 2.0 tokens (FR-026)
- [ ] T052 Run `make sort-imports`, `make bdd` (full suite), `make frontend-check`, and `make validate`; all green
- [ ] T053 Walk through `specs/XEN-016-request-aliases/quickstart.md` manual steps in `wails3 dev`

---

## Requirement coverage

| Requirement | Tests | Implementation |
| --- | --- | --- |
| FR-001, FR-002, FR-004 | T002, T007, T012 | T003–T006, T008, T010, T011 |
| FR-003 | T025 | T027, T028 |
| FR-005 | T012, T030 | T009, T016, T027 |
| FR-006 | T012, T014 | T015–T018, T021–T023 |
| FR-007 | T030, T031 | T027, T034 |
| FR-008 | T030, T031 | T032, T033 |
| FR-009 | T012, T026 | T016, T027, T029 |
| FR-010 | T014 | T023 |
| FR-011 | T024, T031 | T024, T034 |
| FR-012 | T036 | T022, T038 |
| FR-013 | T026 | T029 |
| FR-014 | T035, T036 | T037, T038 |
| FR-014a | T036 | T022, T033 |
| FR-014b | T043, T045 | T045 |
| FR-015, FR-015b | T013 | T019, T022 |
| FR-015a | T013, T030 | T020, T032 |
| FR-016 | T024 | T024 |
| FR-017 | T039, T040 | T041, T042 |
| FR-018 | T043 | T044 |
| FR-019–FR-022a | T043 | T006 (model), no other code (lock-in scenarios) |
| FR-023–FR-025 | T046 | T047–T050 |
| FR-026 | — | T051 |
| SC-001–SC-006 | covered by the scenarios above | T052, T053 |

## Dependencies & Execution Order

- Phase 1 → Phase 2 (blocks everything).
- US1 (Phase 3) needs Phase 2. US2 needs US1's service/binding/store (T015–T023). US3 needs US2's `Save` (rename) and US1's menu. US4 needs US1's store fields. US5 needs US1's send path (T019). US6 needs Phase 2 (T006) only, except T045 (needs T022). US7 needs US1–US3 services (Create/Save/Delete) and T019.
- Polish after all stories.

### Parallel opportunities

- Phase 2: T005, T008, T009, T010, T011 in parallel after T004.
- Each story's red tests ([P]) run in parallel with each other.
- After US1: US4, US5 and US6 can proceed in parallel; US7 after US3.

## Implementation Strategy

1. MVP = Phases 1–3 (create + send an alias). 2. Add US2 (independent saves) — the contract users rely on.
3. US3, US4. 4. US5, US6, US7. 5. Polish and `make validate`.
