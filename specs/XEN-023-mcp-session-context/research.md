# Research: Per-Session MCP Context (XEN-023)

All findings verified on `main` @ `c87aed8` (2026-11-03). No open unknowns remain.

## R1 — Where the session context lives

- **Decision**: An in-memory store keyed by MCP session id, implemented in the integration persistence layer
  (`internal/integration/persistence/ai_session_context.go`) behind usecase interfaces
  (`FindAISessionContext`, `SaveAISessionContext`, `DeleteAISessionContext`), and owned by a new application service
  `ManageAISessionService` (`internal/application/service/manage_ai_session.go`).
- **Rationale**: This is the pattern already used for other process-lifetime state reached through ports: pending
  AI approvals (`persistence/ai_approval.go` + `usecase/queue_ai_approval.go`, `find_ai_approvals.go`) and active runs
  (`usecase/claim_active_run.go`, `find_active_runs.go`). Sessions die with the process, so nothing is written to disk.
- **Alternatives considered**: a field on the `mcpCall` controller (rejected: business state in a controller violates
  `controllers.md`); persisting in app-data (rejected: sessions do not survive a restart, so persisted context would
  only go stale).

## R2 — When the context is seeded

- **Decision**: On the MCP `initialized` handshake. `listener.track` (`internal/infra/server/mcp/listener.go:110`)
  already sees every new session; it calls a new `routes.OpenSession(ctx, sessionId)` →
  `MCPCallController.OpenSession(ctx, dto.MCPSession)` → `ManageAISessionService.Open`. As a safety net,
  `ManageAISessionService.Current` seeds a missing context the same way (a call that reaches Xenon before
  `initialized`, or a call without a session id, which is seeded per call and not stored).
- **Rationale**: FR-002 says "at the moment the session connects". `untrackWhenClosed` already calls
  `routes.CloseSession`, so open and close are symmetric in the same file.
- **Alternatives considered**: seeding lazily on the first tool call only (rejected: a user who switches the window
  between connect and first call would see the session start somewhere else than when it connected).

## R3 — How a session's environment is expressed to the existing variable resolution

- **Decision**: `dto.MCPToolCall` gains `EnvironmentId`, always explicit: an environment id, or `entity.NoEnvironmentID`
  (`"none"`, `internal/domain/entity/run_config.go:10`). It is never `""`.
- **Rationale**: `requestVariables.loadIn` (`internal/application/service/request_variables.go:32`) already
  distinguishes `""` = the workspace's stored active environment (window behaviour), `"none"` = no environment, and an
  id = that environment (a deleted id resolves with none — `runEnvironment`). Passing an explicit id/`none` from MCP
  reuses this unchanged and makes FR-010's "deleted environment resolves with none" automatic.
- **Plumbing**:
  - Send: `send_request.go` builds `entity.NoRunContext()`; it becomes
    `entity.NoRunContext().InEnvironment(environmentId)` (new value method, only sets `EnvironmentID`), and
    `SendRequestService.Send` gains an `environmentId` parameter. `send_core.go:39` already passes
    `runContext.EnvironmentID` to `loadIn`. The window's `RequestService.Send` passes `""` (unchanged behaviour).
  - Snippet: `entity.SnippetInput` gains `EnvironmentID`; `generate_snippet.go:83` switches `load` → `loadIn`.
  - Inspect: `InspectVariablesService.Inspect` gains `environmentId`; `inspect_variables.go:47` switches to `loadIn`.
    The window's `VariableService.Inspect` passes `""`.
  - Run: `StartForAI` gains `activeEnvironmentId`; after `plan.Configured(overrides)`, a run whose configured
    environment is `""` (= "the active one") gets the session environment via new
    `RunConfig.InActiveEnvironment(id)`. A run config that pins an environment, or an explicit `environment`
    argument, is unchanged.

## R4 — Resolving `send`'s `environment` argument

- **Decision**: Extract `RunPlan.EnvironmentID(ref)` (`internal/domain/entity/run_plan.go:41`) into a package
  function `entity.ResolveEnvironmentRef(environments []Environment, ref) (id string, known bool)` used by both
  `RunPlan.EnvironmentID` and `ManageAISessionService.ResolveEnvironment(ctx, localId, ref)`. Unknown → the controller
  rejects with `RejectInvalidArguments(tool, "environment must be an environment name or id, or none")` — the same
  message and code `run` uses (`mcp_run.go:22,102`).
- **Rationale**: identical semantics for `run` and `send` (FR-008) from one function.

## R5 — Registering a workspace without activating it

- **Decision**: New domain method `AppRegistry.Add(workspace)` (appends, leaves `ActiveLocalID`), and new service
  methods `CreateWorkspaceService.CreateInBackground(ctx, name, path)` and
  `OpenWorkspaceService.OpenInBackground(ctx, path)` that share the existing validation and call `Add` instead of
  `Register` (`app_registry.go:17` sets `ActiveLocalID`).
- **Rationale**: FR-012. The window's own create/open keep activating.
- **Switch**: `mcp_workspace_edit.go switchTo` stops calling `ActivateWorkspaceService.Activate` (which stops runs —
  `activate_workspace.go:66` — re-points the file watcher and loads UI state). It calls
  `ManageAISessionService.Enter(ctx, sessionId, localId)`, which checks the workspace is registered
  (`FindWorkspaceRegistry.FindWorkspace`, not found → existing workspace-not-found error) and that its folder loads
  (`LoadWorkspace.Load`, so a missing folder fails as it does today). The auto-sync `Opened` schedule stays (it is data
  sync, not window state).

## R6 — Window refresh instead of follow

- **Decision**: Keep the `workspaces:changed` event and payload (`{active_local_id}` = registry's window workspace).
  In `frontend/src/app/bootstrap.ts:95` replace `follow(activeLocalId)` with a new store action
  `refreshList(activeLocalId)` in `features/workspaces/store.ts`: re-read the list; if the window's
  `activeLocalId` is still in it, only update `workspaces`; if it is gone (removed/unlinked by AI), do exactly what
  `remove` does when the removed workspace was active (`store.ts:203-210`): `activate(list.activeLocalId)` or
  `leaveWorkspace(null)`. `follow` is deleted (no other caller — only bootstrap uses it).
- **Rationale**: FR-014. Only AI publishes `workspaces:changed` (`notify_ai_change.go:70`), so the window's own
  actions are unaffected.

## R7 — Write serialisation

- **Facts**: files are written atomically (`pkg/fileutils/write_atomic.go`), so torn files are not the risk; lost
  updates are. Locks today: one tree/request index lock (`workspace_index_cache.go:20`), separate per-workspace
  local-variables lock used only by `script_variables.go:26`; environment, workspace.yaml, vault and active-environment
  saves are unlocked. The request revision check (`save_request.go:55-67`) reads the revision and writes under
  separately-taken locks (`request.go:44,101`).
- **Decision**: One per-workspace, context-reentrant write lock: usecase
  `LockWorkspace.LockWorkspace(ctx, localId) (context.Context, func())` (`internal/application/usecase/lock_workspace.go`),
  implemented in `internal/integration/persistence/workspace_lock.go` (a `sync.Map` of `*sync.Mutex` per local id,
  as `local_variables.go:77` does; the returned context carries a marker so a nested call for the same workspace is a
  no-op instead of a deadlock). It replaces `LockLocalVariables` (deleted; `script_variables.go` moves to it).
  Every application service method that reads then writes workspace files or workspace-local state takes it first and
  releases it on return:
  `save_request.go` (Save), `manage_tree_item.go` (all mutating methods), `manage_alias.go`, `manage_example.go`,
  `manage_item_settings.go`, `manage_documentation.go` (save), `manage_environment.go` (Create/Update/Duplicate/Delete),
  `reorder_environment.go`, `set_active_environment.go`, `manage_variable_table.go` (all mutating methods),
  `script_variables.go`, `manage_vault.go` (mutating), `manage_oauth.go` (mutating), `manage_cookies.go` (mutating), `rename_workspace.go`, `manage_import.go` (import into a
  workspace), `handle_file_change.go` (variable fix-up), `pull_workspace.go`/`finish_pull.go`/`sync_pull.go` (file
  writes).
  In addition, `vaultWriter.update` (`internal/application/service/vault_writer.go:26`), the one read-modify-write of
  a workspace vault, takes the lock itself, so vault writes made during a send (OAuth token refresh in
  `send_pipeline.go`) or by any other helper are serialised too; reentrancy makes this a no-op under an outer lock.
- **Rationale**: one lock order (a single lock per workspace) cannot deadlock across workspaces; reentrancy makes
  service-calls-service safe. Sends and runs never hold it while waiting on the network: only the script-variable
  write-back at the end of a send takes it (as `LockVariables` does today).
- **Alternatives considered**: per-file locks (rejected: request + workspace.yaml ordering writes span files; many
  locks invite ordering deadlocks); a global lock (rejected: needlessly serialises unrelated workspaces).

## R8 — Environment revision

- **Decision**: `entity.Environment.Revision` = SHA-256 (hex) of the environment file bytes, computed by
  `environmentPersistence.Find`/`List` exactly like requests (`workspace_index_cache.go:164`); not stored in the file.
  Returned in `dto.Environment.revision` and `dto.VariableTable.revision` (environment scopes only; empty for other
  scopes). Window saves send `base_revision` and `force` on `dto.UpdateEnvironment` and `dto.SaveVariableTable`.
  `ManageEnvironmentService.Update` and `ManageVariableTableService.SaveTable` compare, under the workspace lock,
  the stored revision with `base_revision` when the scope is an environment and `base_revision` is non-empty and
  `force` is false; mismatch → `errs.ConflictError("the environment changed on disk since it was opened",
  "ENV-19409")` via a new `usecase.RejectStaleEnvironmentWrite()` (mirrors `reject_stale_alias_write.go`).
- **Empty `base_revision`** means "patch the latest": used only by MCP paths, which read the stored environment and
  apply their change inside the same locked step (FR-020). The window always sends its revision.
- **Next free code**: `ENV-19409` (existing ENV codes end at `ENV-18400`).

## R9 — Window conflict UX for environments

- **Decision**: In `features/environments/store.ts`, tables keep `revision` from the loaded `VariableTable`; drafts
  remember `baseRevision`. `saveDraft`/`saveTable` and `updateEnvironment` send it. On `ENV-19409` the draft is kept
  and `conflicts[scope] = true`; `EnvironmentTab.tsx` renders a notice (reusing the request tab's `TabNotice`
  wording pattern) with **Load from disk** (`discardDraft(scope)` + reload) and **Keep mine** (save again with
  `force: true`).
- **Rationale**: matches the request tab's existing "changed" notice (`features/tabs/store.ts:1166-1217`) so users see
  one pattern.

## R10 — Session tag

- **Decision**: `entity.AISessionTag(sessionId) string` = first 6 characters (whole id when shorter; `""` for no
  session). Stored as `session` on `HistoryEntry` (model + DTO; `omitempty`), set for AI actions
  (`RecordAIActivityService.Record` gains a session parameter) and AI sends (`entity.SendOrigin.Session`).
  `HistoryEntry.Matches` (`entity/history_entry.go:115`) also matches the tag for AI entries. Frontend: the history
  mapper reads `session`; `aiTag(client, session)` renders `client · tag`; the MCP `history` tool lines include it.
- **Rationale**: FR-021/022. Six characters of a random session id are enough to tell concurrent sessions apart.

## R11 — Test harness support

- Separate MCP sessions already exist: `the MCP client is named "<name>"` keeps one session per name
  (`test/integration/steps/mcp_helpers_test.go:22`); background calls (`I start calling the MCP tool … in the
  background as "<name>" with`) give concurrency. Window state is observable through the app-data files
  (`app.json` `active_local_id`; `workspaces/<localId>/variables.json` `activeEnvironmentId`) with the existing
  "the app data file … should contain the field …" step, and window actions through `I call "<Service>.<Method>" with`.
