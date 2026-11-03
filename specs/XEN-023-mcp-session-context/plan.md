# Implementation Plan: Per-Session MCP Context

**Branch**: `xen-023` | **Date**: 2026-11-03 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/XEN-023-mcp-session-context/spec.md`

## Summary

MCP calls stop reading "where am I" from the window's registry/active environment and read it from a per-session
context instead; MCP workspace/environment selection writes only that context. The window's `workspaces:changed`
handler refreshes instead of following. Writes to a workspace are serialised by one reentrant per-workspace lock that
also makes the existing request/alias/example revision check atomic; environments gain the same revision check for
window saves (`ENV-19409`) with a Load-from-disk / Keep-mine notice. History rows for AI carry a 6-character session
tag. Design decisions and the facts behind them: [research.md](./research.md).

## Technical Context

**Language/Version**: Go (per `go.mod`), TypeScript 5 / React 19

**Primary Dependencies**: Wails v3, `github.com/modelcontextprotocol/go-sdk/mcp` (session ids, `InitializedHandler`), Zustand — no new dependency

**Storage**: session contexts in memory only; environment revision derived from file bytes (not stored); history entries gain `session` (`omitempty`, backward compatible)

**Testing**: Godog BDD (`make bdd TAG=@xen_023`), Go unit tests for new domain methods, Vitest

**Target Platform**: macOS, Windows, Linux desktop

**Project Type**: desktop app (Go backend + React UI in Wails)

**Performance Goals**: no measurable change — the lock covers only local read-modify-write steps (milliseconds), never network I/O; session lookup is a map read per call

**Constraints**: lock must be reentrant per context (services call services); one lock per workspace, so no cross-workspace ordering deadlock

**Scale/Scope**: a handful of concurrent MCP sessions; one context entry each

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first** — no new network service; session context is process memory; workspace isolation strengthened (a session's calls are pinned to its own workspace).
- [x] **II. Secrets** — no secret value is added to any payload: the session context holds ids only; the revision is a hash of the environment file, which never contains secret values (secrets live in the vault); the session tag is a prefix of a random session id.
- [x] **III. Go owns logic** — context, locking, revision checks and environment resolution are all Go; React only renders the notice/tag and calls bindings.
- [x] **IV. Architecture skills** — every new file is placed per `architecture.md` below (entity, usecase port, persistence adapter, service, controller, DTO, infra wiring); frontend changes stay in `features/workspaces`, `features/environments`, `features/sidebar`, `app/bootstrap.ts`, `platform/`.
- [x] **V. BDD-first** — feature files listed per requirement below; only outer boundaries (HTTP targets, file roots) are faked.
- [x] **VI. Postman parity** — Postman has no MCP-session concept; this changes Xenon-only AI behaviour inside the PRD's AI-access feature; nothing beyond it.
- [x] **VII. Cross-platform** — no OS-specific code (mutexes and hashing are portable).

Post-design re-check: unchanged, all ✅.

## Design

### Domain (`internal/domain/entity/`)

- `ai_session_context.go` — `AISessionContext` (see [data-model.md](./data-model.md)) + unit test.
- `ai_session_tag.go` — `AISessionTag(sessionId)`.
- `app_registry.go` — `Add(workspace)`.
- `environment.go` — `Revision` field.
- `run_context.go` — `InEnvironment(id)`; `run_config.go` — `InActiveEnvironment(id)`; `run_plan.go` — `EnvironmentID` delegates to new `ResolveEnvironmentRef` (`environment_ref.go`).
- `snippet_input.go` — `EnvironmentID`.
- `send_origin.go` — `Session`; `history_entry.go` — `Session`, set from origin / activity; `Matches` includes the tag for AI entries.

### Application

Ports (`internal/application/usecase/`):
- `find_ai_session_context.go`, `save_ai_session_context.go`, `delete_ai_session_context.go`.
- `lock_workspace.go` — `LockWorkspace(ctx, localId) (context.Context, func())`. `lock_local_variables.go` deleted.
- `reject_stale_environment_write.go` — `ENV-19409` conflict.

Services (`internal/application/service/`, adapters in `internal/application/adapter/`):
- **New** `manage_ai_session.go` — `ManageAISessionService`: `Open(ctx, sessionId)`, `Current(ctx, sessionId) (AISessionContext, error)` (validates the workspace is still registered → `Leave()` otherwise; seeds when missing; does not store when `sessionId == ""`), `Enter(ctx, sessionId, localId)`, `ChooseEnvironment(ctx, sessionId, environmentId)` (checks it exists via `FindEnvironment`, unknown → existing not-found), `ResolveEnvironment(ctx, localId, ref) (string, bool, error)`, `Close(ctx, sessionId)`. Dependencies: `FindWorkspaceRegistry`, `LoadWorkspace`, `FindLocalVariables`, `FindEnvironment`, `ListEnvironments`, and the three session ports.
- `create_workspace.go` / `open_workspace.go` — `CreateInBackground` / `OpenInBackground` (shared validation; `registry.Add`).
- `send_request.go` — `Send(..., environmentId, ...)` → `NoRunContext().InEnvironment(environmentId)`.
- `generate_snippet.go`, `inspect_variables.go` — `loadIn` with the given environment.
- `start_run.go` — `StartForAI(..., activeEnvironmentId, ...)` → `config.InActiveEnvironment(...)` after `Configured`.
- `record_ai_activity.go` — `Record(ctx, localId, client, session, activity)`.
- `manage_environment.go` (`Update(ctx, localId, environment, baseRevision, force)`), `manage_variable_table.go` (`SaveTable(..., baseRevision, force)`) — revision check under the lock.
- Lock taken at the start of every mutating method listed in research R7 (`save_request.go`, `manage_tree_item.go`, `manage_alias.go`, `manage_example.go`, `manage_item_settings.go`, `manage_documentation.go`, `manage_environment.go`, `reorder_environment.go`, `set_active_environment.go`, `manage_variable_table.go`, `script_variables.go`, `manage_vault.go`, `manage_oauth.go`, `manage_cookies.go`, `rename_workspace.go`, `manage_import.go`, `handle_file_change.go`, `pull_workspace.go`, `finish_pull.go`, `sync_pull.go`), plus inside `vault_writer.go` `update` (covers vault writes during sends).

### Integration

- `persistence/ai_session_context.go` — in-memory store.
- `persistence/workspace_lock.go` — per-workspace reentrant lock; `local_variables.go` loses `LockVariables`/`workspaceLocks`.
- `persistence/environment.go` + `model/environment.go` — compute `Revision` on read (hash of bytes, not serialised).
- `persistence/model/history_entry.go` — `session,omitempty`.
- Controllers (`entrypoint/controller/`):
  - `mcp_call.go` — inject `ManageAISessionService` (replaces `ListWorkspacesService`); `run` reads `Current(call.Session)` → `NewMCPToolCall(call, tool, action, ctx.LocalID, ctx.EnvironmentID())`; `recordedWorkspace` uses the session context after the call; new `OpenSession`; `CloseSession` also calls `Close`; activity record carries the session tag.
  - `mcp_workspace_edit.go` — switch/create/open via session service + background create/open; drop `ActivateWorkspaceService`; switch returns no AI changes.
  - `mcp_workspace.go` — list/active from the session (`call.LocalId`, `call.EnvironmentId`).
  - `mcp_env.go` — list marks `call.EnvironmentId`; inspect passes it.
  - `mcp_env_edit.go` — `set_active` → `ChooseEnvironment`; no `Variables` change for it; drop `SetActiveEnvironmentService`.
  - `mcp_send.go` — resolve `arguments.Environment` (or `call.EnvironmentId`) → pass to `Send` (both the first and the confirmed attempt); origin carries the tag.
  - `mcp_run.go` — pass `call.EnvironmentId` as `activeEnvironmentId`.
  - `mcp_snippet.go` — pass `call.EnvironmentId`.
  - `environment_manage.go`, `variable_table.go` — pass `base_revision`/`force`.
  - Tool descriptions per [contracts/mcp-tools.md](./contracts/mcp-tools.md).
- DTOs (`entrypoint/dto/`): `mcp_tool_call.go` (`EnvironmentId`, `SessionTag`), `mcp_send.go` (`Environment`), `send_request.go` (`EnvironmentId`), `environment.go`/`variable_table.go` (`revision`), `update_environment.go`/`save_variable_table.go` (`base_revision`, `force`), `history_entry.go` + `mcp_history_line.go` (`session`), `ai_activity_record.go` (session).
- `adapter/mcp_call_controller.go` — `OpenSession(ctx, dto.MCPSession) error`.

### Infrastructure

- `server/mcp/listener.go` — `track` calls `l.routes.OpenSession(ctx, session.ID())` after recording the session.
- `server/adapter/mcp_routes.go` + `server/mcp/routes.go` — `OpenSession`.
- `dependency/` — wire `AISessionContextPersistence`, `WorkspaceLock`, `ManageAISessionService`; update the constructors whose signatures change (mcp call/workspace edit/env edit/send/run/snippet controllers, send/run/snippet/inspect/record services, every service that gains `LockWorkspace`).

### Frontend

- `app/bootstrap.ts` — `onWorkspacesChanged(id => refreshList(id))`.
- `features/workspaces/store.ts` — `refreshList` replaces `follow` (R6).
- `features/environments/store.ts` — revisions on tables/environments, `baseRevision` on drafts, `conflicts`, `loadFromDisk(scope)`, `keepMine(scope)`; `EnvironmentTab.tsx` — notice.
- `platform/services/environments.ts`, `variables.ts`, mappers — `revision`, `baseRevision`, `force`.
- `platform/mappers/history.ts` — `session`; `features/sidebar/components/HistoryView.tsx`, `HistoryActionRow.tsx` — `aiTag(client, session)`.

## Tests (BDD-first)

All new scenarios tagged `@xen_023` plus their own tag.

| Requirement | Test |
| --- | --- |
| FR-001, FR-002, FR-006 | `test/integration/features/mcp/session_context.feature` `@seeded`: window on Acme + Dev → a new session's `workspace active` = Acme with Dev active; `workspace list` marks Acme |
| FR-003, FR-004 | same file `@isolated_workspaces`: clients `one`/`two`; `two` switches to Globex → `one`'s `workspace active` = Acme, `one`'s `tree` lists Acme items |
| FR-005, FR-007 | same file `@isolated_environments`: `two` sets Prod → `one`'s `send` resolves `{{base_url}}` with Dev (HTTP mock sees Dev host); `two`'s `send` uses Prod; `variables.json` still `activeEnvironmentId` = Dev |
| FR-008 | same file `@send_environment_override`: `send` with `environment: "Prod"` and `"none"`; next `send` back on session env; unknown → invalid-arguments error |
| FR-007 (run) | same file `@run_session_environment`: `run start` without environment on a target using the active environment → run's environment = session env |
| FR-004 (re-entry) | same file `@reenter_keeps_choice` |
| FR-009 | same file `@closed_session`: close session, reconnect after the window changed → seeded from the window again |
| FR-010 | same file `@workspace_removed` (MCP-06404) and `@environment_deleted` (resolves with none) |
| FR-011, FR-012, FR-013 | `test/integration/features/mcp/workspaces.feature` — rewrite `@switch` (no `workspaces:changed` with the new id; `app.json` `active_local_id` unchanged; a running run keeps running) and `@create_open_rename` (create/open listed, `app.json` active unchanged); update the Feature description |
| FR-005 (window untouched) | `test/integration/features/mcp/environments.feature` — set_active scenario asserts `variables.json` unchanged and no `variables:changed` |
| FR-016, FR-017, FR-020 | `test/integration/features/mcp/concurrent_writes.feature` `@parallel_request_edits` / `@parallel_env_edits`: two sessions issue N background edits to different fields/variables of the same request/environment → all present |
| FR-018 | `test/integration/features/environments/environment_revision.feature` `@stale_table_save`, `@stale_settings_save`, `@forced_save`, `@revision_returned` via `VariableService.SaveTable` / `EnvironmentService.Update` |
| FR-021, FR-022 | `test/integration/features/history/ai_activity.feature` `@session_tag`: two clients' entries carry different 6-char `session`; `HistoryService.List` with the tag as query returns only that session's entries |
| FR-023 | `test/integration/features/mcp/connect.feature`: `@tool_list` budget raised to 39500 characters; new `@per_session_descriptions`: listed `workspace`, `workspace_edit`, `env_edit`, `send` descriptions contain "session" |
| FR-014, FR-015 | Vitest `frontend/src/app/bootstrap.test.ts` (workspaces:changed → `refreshList`, no `activate`); `features/workspaces/store.test.ts` (`refreshList` keeps the window, switches only when its workspace was removed) |
| FR-019 | Vitest `features/environments/store.test.ts` (ENV-19409 keeps draft, sets conflict; keepMine forces; loadFromDisk discards) and `components/EnvironmentTab.test.tsx` (notice + buttons) |
| FR-022 (UI) | Vitest `features/sidebar/components/HistoryView.test.tsx` (tag rendered next to client) |
| Domain units | Go `ai_session_context_test.go`, `environment_ref_test.go`, `history_entry_test.go` (tag match) |

Only HTTP targets and file roots are faked; MCP clients are real SDK sessions (`the MCP client is named …`).

## Project Structure

### Documentation (this feature)

```text
specs/XEN-023-mcp-session-context/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── mcp-tools.md
│   └── wails-bindings-and-events.md
├── checklists/
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
internal/domain/entity/          ai_session_context.go, ai_session_tag.go, environment_ref.go, + edits
internal/application/usecase/    find/save/delete_ai_session_context.go, lock_workspace.go, reject_stale_environment_write.go
internal/application/adapter/    manage_ai_session.go, + signature edits
internal/application/service/    manage_ai_session.go, + edits listed above
internal/integration/persistence/ ai_session_context.go, workspace_lock.go, environment.go, model/*
internal/integration/entrypoint/ controller/mcp_*.go, environment_manage.go, variable_table.go, dto/*
internal/infra/server/mcp/       listener.go, routes.go ; internal/infra/server/adapter/mcp_routes.go
internal/infra/dependency/       wiring
frontend/src/                    app/bootstrap.ts, features/{workspaces,environments,sidebar}, platform/{services,mappers}
test/integration/features/       mcp/{session_context,concurrent_writes}.feature, environments/environment_revision.feature, edits
```

**Structure Decision**: single repo, existing layout; no new top-level directories.

## Complexity Tracking

No constitution violations.
