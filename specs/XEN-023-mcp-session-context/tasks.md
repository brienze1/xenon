---

description: "Task list for XEN-023 Per-Session MCP Context"
---

# Tasks: Per-Session MCP Context

**Input**: Design documents from `specs/XEN-023-mcp-session-context/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/mcp-tools.md, contracts/wails-bindings-and-events.md, quickstart.md

**Tests**: MANDATORY (constitution V, BDD-first). Every story starts with tests proven red before implementation; frontend stories also get Vitest tests.

**Organization**: grouped by user story. Phase 2 builds the session context every MCP story needs. US1 and US2 depend on it; US3 (concurrent writes) and US4 (session tag) are independent of each other and of US1/US2 except where noted.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1–US4)

---

## Phase 1: Setup

- [ ] T001 Fresh worktree on branch `xen-023` (from freshly fetched `origin/main`): run `npm ci` in `frontend/` and `make bindings` so Vitest and the BDD harness start (`.claude/skills/makefile.md`)

---

## Phase 2: Foundational — session context (blocks US1, US2)

**Purpose**: every MCP call resolves its workspace/environment from a per-session context (research R1–R3).

- [ ] T002 [P] Add `AISessionContext` (`NewAISessionContext`, `Enter`, `HasEntered`, `ChooseEnvironment`, `EnvironmentID`, `Leave`) per data-model.md in `internal/domain/entity/ai_session_context.go` with table tests in `internal/domain/entity/ai_session_context_test.go` (seed with/without workspace, empty stored env → `none`, re-entry keeps choice, `EnvironmentID` never `""`)
- [ ] T003 [P] Add ports `FindAISessionContext`, `SaveAISessionContext`, `DeleteAISessionContext` in `internal/application/usecase/find_ai_session_context.go`, `save_ai_session_context.go`, `delete_ai_session_context.go` (pattern: `usecase/find_ai_approvals.go`)
- [ ] T004 Implement the in-memory store (`sync.Mutex` + map; `Save` with empty session id is a no-op) in `internal/integration/persistence/ai_session_context.go` with its persistence adapter interface in `internal/integration/adapter/ai_session_context_persistence.go` (pattern: `persistence/ai_approval.go`) and DI in `internal/infra/dependency/integration/ai_session_context_persistence.go`
- [ ] T005 Add `ManageAISessionService` adapter in `internal/application/adapter/manage_ai_session.go` and service in `internal/application/service/manage_ai_session.go`: `Open` (seed from `FindWorkspaceRegistry.Find().ActiveLocalID` + `FindLocalVariables` active environment), `Current` (seed when missing; `Leave()` when the workspace is no longer registered; not stored for `""` session), `Enter` (registered + `LoadWorkspace.Load` check; seed environment from stored active only when `!HasEntered`), `ChooseEnvironment` (`FindEnvironment` must find it unless empty), `ResolveEnvironment(ctx, localId, ref)` (via `ListEnvironments` + `entity.ResolveEnvironmentRef`), `Close`; DI in `internal/infra/dependency/application/manage_ai_session.go`
- [ ] T006 [P] Add `ResolveEnvironmentRef(environments []Environment, ref string) (string, bool)` in `internal/domain/entity/environment_ref.go` (+ `environment_ref_test.go`) and make `RunPlan.EnvironmentID` in `internal/domain/entity/run_plan.go` delegate to it
- [ ] T007 Extend `dto.MCPToolCall` with `EnvironmentId` and `SessionTag()`; change `NewMCPToolCall(call, tool, action, localId, environmentId)` in `internal/integration/entrypoint/dto/mcp_tool_call.go` (update all callers)
- [ ] T008 In `internal/integration/entrypoint/controller/mcp_call.go`: replace `ListWorkspacesService` with `ManageAISessionService`; `run` reads `Current(ctx, call.Session)` for `LocalId`/`EnvironmentId`; `recordedWorkspace` reads the session context after the call; add `OpenSession(ctx, dto.MCPSession)` (validate → `Open`); `CloseSession` also calls `Close`; update `internal/integration/adapter/mcp_call_controller.go` and DI `internal/infra/dependency/controller/mcp_call.go`
- [ ] T009 Seed on connect: add `OpenSession(ctx, sessionId)` to `internal/infra/server/adapter/mcp_routes.go` and `internal/infra/server/mcp/routes.go`; call it from `listener.track` in `internal/infra/server/mcp/listener.go` after recording the session; adjust `internal/infra/server/mcp/server_test.go` fakes

**Checkpoint**: `make bdd TAG=@mcp` still green (sessions seeded from the window give the same results as before for a single session).

---

## Phase 3: User Story 1 — AI work never moves my window (P1) 🎯 MVP

**Goal**: no MCP action changes the window's workspace, stored active environment, tabs or selection (FR-011–FR-015).

**Independent Test**: MCP `switch`/`create`/`open`/`set_active` leave `app.json` `active_local_id` and `variables.json` `activeEnvironmentId` unchanged and emit no follow.

### Tests (write first, prove red)

- [ ] T010 [P] [US1] Rewrite `@switch` and `@create_open_rename` and the Feature description in `test/integration/features/mcp/workspaces.feature`: switch emits no `workspaces:changed` with the new id, `app.json` `active_local_id` unchanged, the session's `workspace active` = new workspace, a slow run in the window's workspace keeps running (`runner:` events / `RunnerService.Get` status `running`); create/open list the workspace while `app.json` keeps the window's workspace; tag `@xen_023`
- [ ] T011 [P] [US1] Add a set_active scenario to `test/integration/features/mcp/environments.feature`: `env_edit set_active` Prod → `workspaces/<localId>/variables.json` `activeEnvironmentId` still Dev, no `variables:changed` event, `EnvironmentService.List` `active_id` still Dev; tag `@xen_023 @set_active_session_only`
- [ ] T012 [P] [US1] Vitest in `frontend/src/app/bootstrap.test.ts`: replace "follows workspace changes made through MCP" with "refreshes the workspace list on MCP changes without switching" (spy `refreshList`)
- [ ] T013 [P] [US1] Vitest in `frontend/src/features/workspaces/store.test.ts`: replace the `follow` cases with `refreshList` cases — keeps the window's workspace when still listed; activates the registry's workspace when the window's was removed; leaves to none when the list is empty
- [ ] T014 [US1] Run T010–T013 and record that they fail (red)

### Implementation

- [ ] T015 [P] [US1] Add `AppRegistry.Add` in `internal/domain/entity/app_registry.go` (+ test in `app_registry_test.go`: active unchanged)
- [ ] T016 [US1] Add `CreateInBackground` / `OpenInBackground` to `internal/application/adapter/create_workspace.go`, `open_workspace.go` and `internal/application/service/create_workspace.go`, `open_workspace.go`, sharing the existing validation and calling `registry.Add`
- [ ] T017 [US1] Rework `internal/integration/entrypoint/controller/mcp_workspace_edit.go`: create/open → background variants then `ManageAISessionService.Enter(call.Session, id)`; switch → `Enter` + keep `scheduleAutoSync(...Opened...)`, result with no AI changes; remove `ActivateWorkspaceService`; `Handle` needs the `MCPToolCall` for the session; update DI `internal/infra/dependency/controller/mcp_workspace_edit.go`
- [ ] T018 [US1] In `internal/integration/entrypoint/controller/mcp_env_edit.go`: `set_active` → `ManageAISessionService.ChooseEnvironment(call.Session, id)`, result `NewMCPEnvActive(id or "")`, no `aichangekind.Variables` for this action; remove `SetActiveEnvironmentService`; update DI `internal/infra/dependency/controller/mcp_env_edit.go`
- [ ] T019 [US1] Replace `follow` with `refreshList(activeLocalId)` in `frontend/src/features/workspaces/store.ts` (research R6; reuse the `remove` branch for a vanished window workspace) and call it from `frontend/src/app/bootstrap.ts:95`
- [ ] T020 [US1] Run T010–T013 green, then `make bdd TAG=@mcp`, the workspaces/bootstrap Vitest suites and `frontend/src/features/tabs/store.test.ts` (its existing changed-on-disk cases lock FR-015: dirty tab keeps edits + notice, clean tab reloads)

**Checkpoint**: the window is never driven by MCP.

---

## Phase 4: User Story 2 — Each AI session has its own workspace and environment (P1)

**Goal**: per-session workspace/environment for every tool; `send` environment override (FR-001–FR-010, FR-023).

**Independent Test**: two named MCP clients in different workspaces/environments each see only their own context.

### Tests (write first, prove red)

- [ ] T021 [P] [US2] Create `test/integration/features/mcp/session_context.feature` (`@all @xen_023 @mcp @mcp_session_context`) with scenarios `@seeded`, `@isolated_workspaces`, `@isolated_environments` (HTTP mock asserts the Dev vs Prod host per client), `@send_environment_override` (Prod, `none`, unknown → invalid arguments, next send back on the session env), `@run_session_environment`, `@reenter_keeps_choice`, `@closed_session` (`the MCP session is closed`, window switched via `WorkspaceService.Activate`, reconnect seeded from the window), `@workspace_removed` (`MCP-06404`), `@environment_deleted` (resolves with none), `@snippet_and_inspect_use_session_env`; use only existing steps (`the MCP client is named`, `I call the MCP tool`, `I call "<Service>.<Method>" with`, app-data and HTTP mock steps)
- [ ] T022 [P] [US2] In `test/integration/features/mcp/connect.feature`: raise the `@tool_list` budget step to `at most 39500 characters`, and add `@xen_023 @per_session_descriptions`: listed tools `workspace`, `workspace_edit`, `env_edit`, `send` have descriptions containing "session"
- [ ] T023 [US2] Run T021–T022 and record that they fail (red)

### Implementation

- [ ] T024 [P] [US2] Add `RunContext.InEnvironment` in `internal/domain/entity/run_context.go` and `RunConfig.InActiveEnvironment` in `internal/domain/entity/run_config.go` (+ unit tests)
- [ ] T025 [P] [US2] Add `EnvironmentID` to `entity.SnippetInput` (`internal/domain/entity/snippet_input.go`); `internal/application/service/generate_snippet.go` uses `variables.loadIn(..., input.EnvironmentID)`
- [ ] T026 [P] [US2] Add `environmentId` to `InspectVariablesService.Inspect` (`internal/application/adapter/inspect_variables.go`, `internal/application/service/inspect_variables.go` → `loadIn`); the window's `internal/integration/entrypoint/controller/variable_inspect.go` passes `""`
- [ ] T027 [US2] Add `environmentId` to `SendRequestService.Send` (`internal/application/adapter/send_request.go`, `internal/application/service/send_request.go` → `entity.NoRunContext().InEnvironment(environmentId)`); add `EnvironmentId` to `dto.SendRequest` (`internal/integration/entrypoint/dto/send_request.go`); window controller `request_send.go` keeps `""`
- [ ] T028 [US2] Add `activeEnvironmentId` to `StartRunService.StartForAI` (`internal/application/adapter/start_run.go`, `internal/application/service/start_run.go`: `config = config.InActiveEnvironment(activeEnvironmentId)` after `Configured`)
- [ ] T029 [US2] `internal/integration/entrypoint/dto/mcp_send.go`: add `Environment` argument (description per contracts/mcp-tools.md) and pass the environment id into `SendRequest`; `internal/integration/entrypoint/controller/mcp_send.go`: resolve `arguments.Environment` via `ManageAISessionService.ResolveEnvironment` (unknown → `RejectInvalidArguments(call.Tool, "environment must be an environment name or id, or none")`), else `call.EnvironmentId`; pass it to both the first and the confirmed send; DI `internal/infra/dependency/controller/mcp_send.go`
- [ ] T030 [P] [US2] `internal/integration/entrypoint/controller/mcp_run.go`: pass `call.EnvironmentId` to `StartForAI`; `mcp_snippet.go`: set `EnvironmentID` on the snippet input from `call.EnvironmentId`; `mcp_env.go`: `inspect` passes `call.EnvironmentId`, `list` marks `call.EnvironmentId` active (`dto.NewMCPEnvList` takes the active id)
- [ ] T031 [P] [US2] `internal/integration/entrypoint/controller/mcp_workspace.go`: `list` marks `call.LocalId` active (`dto.NewMCPWorkspaceList(registry, activeLocalId)`), `active` loads the session's workspace from the registry and marks `call.EnvironmentId`; no workspace → `RejectInactiveAIWorkspace`
- [ ] T032 [US2] Set the four new tool descriptions exactly as in contracts/mcp-tools.md in `mcp_workspace.go`, `mcp_workspace_edit.go`, `mcp_env_edit.go`, `mcp_send.go` (`internal/integration/entrypoint/controller/`)
- [ ] T033 [US2] Run T021–T022 green, then `make bdd TAG=@mcp` and `make bdd TAG=@sending` (window sends unchanged)

**Checkpoint**: parallel sessions are isolated.

---

## Phase 5: User Story 3 — Concurrent edits never lose a write silently (P2)

**Goal**: per-workspace write serialisation, atomic revision checks, environment revisions with a window conflict notice (FR-016–FR-020).

**Independent Test**: parallel MCP edits all land; a stale window environment save gets `ENV-19409`.

### Tests (write first, prove red)

- [ ] T034 [P] [US3] Create `test/integration/features/mcp/concurrent_writes.feature` (`@all @xen_023 @mcp @mcp_concurrent_writes`): `@parallel_request_edits` — clients `one` and `two` each start background `tree_edit` calls (`one` sets header A, `two` sets query param B on the same request, 5 rounds each via distinct keys), wait for both, request file contains every edit; `@parallel_env_edits` — both clients `env_edit add_var` distinct keys on Dev in the background, environment file contains all keys
- [ ] T035 [P] [US3] Create `test/integration/features/environments/environment_revision.feature` (`@all @xen_023 @environments @environment_revision`): `@revision_returned` (`EnvironmentService.Get` and `VariableService.GetTable` return the same non-empty `revision`); `@stale_table_save` (save with an old `base_revision` after an MCP `add_var` → `ENV-19409`, file unchanged); `@stale_settings_save` (same for `EnvironmentService.Update`); `@forced_save` (`force: true` writes); `@fresh_save` (current revision writes and returns a new revision)
- [ ] T036 [P] [US3] Vitest in `frontend/src/features/environments/store.test.ts`: save sends `baseRevision`; `ENV-19409` keeps the draft and sets the scope's conflict; `keepMine` resaves with `force: true`; `loadFromDisk` discards the draft and reloads
- [ ] T037 [P] [US3] Vitest in `frontend/src/features/environments/components/EnvironmentTab.test.tsx`: conflict notice text "This environment changed on disk since it was opened" with **Load from disk** and **Keep mine** buttons calling the store
- [ ] T038 [US3] Run T034–T037 and record that they fail (red)

### Implementation

- [ ] T039 [US3] Add port `LockWorkspace(ctx, localId) (context.Context, func())` in `internal/application/usecase/lock_workspace.go`; implement the reentrant per-workspace lock in `internal/integration/persistence/workspace_lock.go` (+ `workspace_lock_test.go`: serialises same id, independent ids don't block, nested call with the returned context does not deadlock); DI in `internal/infra/dependency/integration/`
- [ ] T040 [US3] Delete `internal/application/usecase/lock_local_variables.go`, `LockVariables` from `internal/integration/adapter/local_variables_persistence.go` and `LockVariables`/`workspaceLocks` from `internal/integration/persistence/local_variables.go`; replace the `LockLocalVariables` dependency with `LockWorkspace` in `internal/application/service/script_variables.go`, `send_pipeline.go`, `send_request.go`, `start_run.go`, `run_starter.go` and their DI files in `internal/infra/dependency/application/` (`send_request.go`, `start_run.go`)
- [ ] T041 [US3] Take `LockWorkspace` at the start of every mutating method (use the returned ctx for the rest of the method) in `internal/application/service/`: `save_request.go`, `manage_tree_item.go`, `manage_alias.go`, `manage_example.go`, `manage_item_settings.go`, `manage_documentation.go`, `manage_environment.go`, `reorder_environment.go`, `set_active_environment.go`, `manage_variable_table.go`, `manage_vault.go`, `manage_oauth.go`, `manage_cookies.go`, `rename_workspace.go`, `manage_import.go`, `handle_file_change.go` (variable fix-up), `pull_workspace.go`, `finish_pull.go`, `sync_pull.go`; also inside `vaultWriter.update` in `vault_writer.go` (add a `lockWorkspace` field, set by every constructor that builds a `vaultWriter`); inject the port via each service's DI file in `internal/infra/dependency/application/`
- [ ] T042 [P] [US3] Add `Revision` to `entity.Environment` (`internal/domain/entity/environment.go`); compute SHA-256 of the file bytes in `internal/integration/persistence/environment.go` `Find`/`List` (not serialised; `model/environment.go` untouched on write)
- [ ] T043 [P] [US3] Add `usecase.RejectStaleEnvironmentWrite()` → `errs.ConflictError("the environment changed on disk since it was opened", "ENV-19409")` in `internal/application/usecase/reject_stale_environment_write.go`
- [ ] T044 [US3] `ManageEnvironmentService.Update(ctx, localId, environment, baseRevision, force)` and `ManageVariableTableService.SaveTable(ctx, localId, scope, edits, confirmReveal, baseRevision, force)`: inside the lock, when the scope is an environment, `baseRevision != ""` and `!force`, compare with the stored `Revision` → `RejectStaleEnvironmentWrite` (`internal/application/adapter/` + `internal/application/service/manage_environment.go`, `manage_variable_table.go`); MCP single-step callers (`mcp_variable_rows.go`, add_var/set_current/reset/persist paths) pass `""`, `false`; `mcp_env_edit.go` `update` passes the revision it read (T045)
- [ ] T045 [US3] Add `retryMCPEdit` in `internal/integration/entrypoint/controller/mcp_edit_retry.go` (re-run on `REQ-02409`/`ENV-19409`, max 50 attempts, last error returned) and wrap the read → merge → save sequences in `mcp_tree_edit_request.go` (`createRequest` with fields, `updateRequest`), `mcp_tree_edit_alias.go`, `mcp_tree_edit_example.go` and `mcp_env_edit.go` `update` (pass `stored.Revision` as base revision instead of `""`); add `mcp_edit_retry_test.go` (stops on success, retries only conflicts, cap)
- [ ] T046 [US3] DTOs: `revision` on `internal/integration/entrypoint/dto/environment.go` and `variable_table.go` (environment scopes); `base_revision`, `force` on `update_environment.go` and `save_variable_table.go`; window controllers `environment_manage.go`, `variable_table.go` pass them
- [ ] T047 [P] [US3] Frontend platform: `revision`/`baseRevision`/`force` in `frontend/src/platform/services/environments.ts`, `variables.ts` and their mappers (+ `environments.test.ts`, `variables.test.ts` cases)
- [ ] T048 [US3] `frontend/src/features/environments/store.ts`: keep `revision` per table/environment, `baseRevision` on drafts, `conflicts` state, `keepMine(scope)`, `loadFromDisk(scope)`; `frontend/src/features/environments/components/EnvironmentTab.tsx` (+ `.module.css`) renders the conflict notice
- [ ] T049 [US3] Run T034–T037 green, then `make bdd TAG=@environments`, `make bdd TAG=@variables`, `make bdd TAG=@requests`, `make bdd TAG=@aliases`, `make bdd TAG=@git` and the environments Vitest suites

**Checkpoint**: no silent lost writes.

---

## Phase 6: User Story 4 — I can tell which session did what (P3)

**Goal**: 6-character session tag on AI activity and AI-sent history, shown and searchable (FR-021, FR-022).

**Independent Test**: two clients' entries carry different tags; searching a tag returns only that session.

### Tests (write first, prove red)

- [ ] T050 [P] [US4] Add `@session_tag` to `test/integration/features/history/ai_activity.feature`: MCP clients named `one` and `two` (one session each) each call `tree`, then `HistoryService.List` with source `ai` shows two action entries whose `session` values are 6 characters long and differ, `HistoryService.List` with `query` = one entry's `session` returns only that session's AI entries; tag `@xen_023`
- [ ] T051 [P] [US4] Vitest in `frontend/src/features/sidebar/components/HistoryView.test.tsx`: an AI row renders `client · tag`; an action row (`HistoryActionRow`) too
- [ ] T052 [US4] Run T050–T051 and record that they fail (red)

### Implementation

- [ ] T053 [P] [US4] Add `AISessionTag` in `internal/domain/entity/ai_session_tag.go` (+ test); `Session` on `SendOrigin` (`send_origin.go`) and `HistoryEntry` (`history_entry.go`, set from origin; `Matches` also checks the tag for AI entries, + test)
- [ ] T054 [US4] `RecordAIActivityService.Record(ctx, localId, client, session, activity)` (`internal/application/adapter/record_ai_activity.go`, `internal/application/service/record_ai_activity.go`); `dto.AIActivityRecord` carries the tag (`internal/integration/entrypoint/dto/ai_activity_record.go`); `mcp_call.go` passes `toolCall.SessionTag()`; `dto/mcp_send.go` origin carries the tag
- [ ] T055 [US4] Persist and expose: `session,omitempty` in `internal/integration/persistence/model/history_entry.go`, `internal/integration/entrypoint/dto/history_entry.go`, `dto/mcp_history_line.go`
- [ ] T056 [US4] Frontend: `session` in `frontend/src/platform/mappers/history.ts`; `aiTag(client, session)` used by `frontend/src/features/sidebar/components/HistoryView.tsx` and `HistoryActionRow.tsx` (visible text and `aria-label`)
- [ ] T057 [US4] Run T050–T051 green, then `make bdd TAG=@history`

---

## Phase 7: Polish & cross-cutting

- [ ] T058 `make sort-imports`, then `make bdd` (full suite) and the full Vitest run; all green
- [ ] T059 Update `docs/prd/xenon-mvp-prd.md`: replace the AI-access bullet "Tools act on the single active workspace. Switching workspace through MCP switches the window too, …" (line 379) with: each MCP session has its own workspace and environment, seeded from the window when it connects; MCP never switches the window, and data changes still show up in it at once; and add an XEN-023 entry to the delivered list noting the `tools/list` budget raised from 39,000 to 39,500 characters
- [ ] T060 `/refactor-diff` until clean, then `make validate`

---

## Dependencies & Execution Order

- Phase 1 → Phase 2 → {US1, US2}. US2's T030/T031 build on T007/T008. US1 and US2 edit different controller files (US2 does not touch `mcp_env_edit.go` or `mcp_workspace_edit.go` except the description text in T032, which runs after T017/T018), so they can proceed in parallel.
- US3 depends only on Phase 1 (lock/revision are independent of sessions); its MCP concurrency scenarios need two sessions, which exist already.
- US4 depends on Phase 2 T007 (`SessionTag`) only.
- Polish after all stories.

## Parallel Examples

- Phase 2: T002, T003, T006 together.
- US1 tests: T010, T011, T012, T013 together.
- US2: T024, T025, T026 together; T030 and T031 together.
- US3: T034–T037 together; T042, T043, T047 together.
- US4: T050, T051 together.

## Implementation Strategy

1. MVP = Phase 2 + US1 + US2 (the reported problem: window hijacked, sessions clash).
2. Then US3 (data safety), then US4 (visibility).
3. Each story ends green on its own suites before the next starts; `make validate` at the end.
