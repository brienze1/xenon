# Data Model: Per-Session MCP Context (XEN-023)

## AISessionContext (domain entity, in memory only)

`internal/domain/entity/ai_session_context.go`

| Field | Type | Meaning |
| --- | --- | --- |
| `SessionID` | string | MCP session id (`""` for a call without a session — never stored) |
| `LocalID` | string | The session's current workspace local id; `""` = no workspace |
| `Environments` | `map[string]string` | Per workspace local id → chosen environment id, or `NoEnvironmentID` (`"none"`). A workspace absent from the map has not been entered yet in this session. |

Methods (value semantics, return updated copies):

- `NewAISessionContext(sessionId, localId, storedActiveEnvironmentId string) AISessionContext` — seed (FR-002). An empty stored id is recorded as `NoEnvironmentID`; an empty `localId` gives an empty map.
- `Enter(localId, storedActiveEnvironmentId string) AISessionContext` — sets `LocalID`; seeds `Environments[localId]` only when absent (FR-004 re-entry keeps the earlier choice).
- `HasEntered(localId string) bool` — whether the map has an entry (lets the service read the stored active environment only on first entry).
- `ChooseEnvironment(environmentId string) AISessionContext` — sets `Environments[LocalID]`; `""` → `NoEnvironmentID` (FR-005).
- `EnvironmentID() string` — choice for `LocalID`, `NoEnvironmentID` when none/no workspace. Never `""` (R3).
- `Leave() AISessionContext` — clears `LocalID` (used when the workspace is no longer registered, FR-010); keeps the map.

Lifecycle: created on `initialized` (`Open`) → updated by `Enter`/`ChooseEnvironment` → deleted on session close (`Close`, FR-009).

Ports (`internal/application/usecase/`): `FindAISessionContext.Find(ctx, sessionId) (AISessionContext, bool)`,
`SaveAISessionContext.Save(ctx, context)`, `DeleteAISessionContext.Delete(ctx, sessionId)`. Implementation:
`internal/integration/persistence/ai_session_context.go` — `sync.Mutex` + `map[string]entity.AISessionContext`; `Save`
with an empty session id is a no-op.

## AppRegistry (existing, extended)

`internal/domain/entity/app_registry.go`

- New `Add(workspace RegisteredWorkspace)` — appends without touching `ActiveLocalID` (FR-012). `Register` unchanged
  (window create/open still activate).

## Environment (existing, extended)

`internal/domain/entity/environment.go`

- New `Revision string` — SHA-256 hex of the stored file bytes, set by persistence on read; never serialised into the
  YAML (FR-018).
- `dto.Environment` gains `revision`; `dto.VariableTable` gains `revision` (environment scopes; `""` otherwise).
- `dto.UpdateEnvironment` and `dto.SaveVariableTable` gain `base_revision` (string) and `force` (bool).
- Rule: stored revision ≠ non-empty `base_revision` and not `force` → `ENV-19409` (409), nothing written.

## RunContext / RunConfig (existing, extended)

- `RunContext.InEnvironment(environmentId string) RunContext` — copy with `EnvironmentID` set (R3).
- `RunConfig.InActiveEnvironment(environmentId string) RunConfig` — when `EnvironmentID == ""` (use active), replace
  with `environmentId`; otherwise unchanged (R3, run).
- `ResolveEnvironmentRef(environments []Environment, ref string) (string, bool)` — package function; `""`/`"none"`
  pass through; otherwise match id or name. `RunPlan.EnvironmentID` delegates to it (R4).

## SnippetInput (existing, extended)

- New `EnvironmentID string` — `""` = stored active (window), id/`none` from MCP.

## SendOrigin / HistoryEntry / AIActivity record (existing, extended)

- `SendOrigin.Session string` — session tag for AI sends.
- `HistoryEntry.Session string` (model `session,omitempty`; DTO `session,omitempty`) — set from `SendOrigin.Session`
  for sends and from the activity record for actions (FR-021).
- `AISessionTag(sessionId string) string` — first 6 characters (whole id if shorter).
- `HistoryEntry.Matches` — for AI entries, also true when the tag contains the needle (FR-022).

## MCPToolCall (DTO, extended)

`internal/integration/entrypoint/dto/mcp_tool_call.go`

- `LocalId` — now the session's workspace (was the registry's active).
- New `EnvironmentId` — the session's environment for `LocalId` (id or `none`, never `""`).
- New `SessionTag()` — `entity.AISessionTag(Session)`.

## MCPSend (DTO, extended)

- New `Environment string` (`json:"environment" description:"Environment name or id, or none, for this send only."`).
