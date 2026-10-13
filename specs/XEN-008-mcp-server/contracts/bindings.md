# Contract: Wails Bindings and Events (UI side)

## New binding — `AIAccessService` (`internal/infra/server/binding/ai_access.go`)

| Method | Input | Output | Errors |
| --- | --- | --- | --- |
| `Status` | — | `{state: "off"\|"running"\|"error", address?, reason?, clients, stdio_command, http_config, stdio_config}` | — |
| `PendingApprovals` | — | `[{id, client, method, url, environment:{id,name,color}, created_at}]` (FIFO) | — |
| `AnswerApproval` | `{id, approved: bool}` | — | `MCP-05409` already resolved, `MCP-07400` unknown id |

`http_config` / `stdio_config` are the ready-to-paste JSON strings (research R4); `stdio_command` is
`"<executable> mcp"`.

## Changed bindings

| Binding.Method | Change |
| --- | --- |
| `SettingsService.Get` / `Save` | DTO gains `history_user_cap`, `history_ai_cap`, `ai_access:{enabled, port, skip_approvals}`. Save leaves **omitted fields unchanged** (so `{ "script_timeout_ms": 2500 }` keeps the caps and AI access), then validates all (`SCR-01400`, `SET-02400`, `SET-03400`, `SET-04400`) and applies AI access immediately (FR-004) |
| `HistoryService.List` | input gains `source` (`all`\|`user`\|`ai`, default `all`); rows gain `kind`, `source`, `client`, `activity{tool,action,target,outcome,code,duration_ms}` |
| `HistoryService.Get` | action entries return `activity` incl. masked `args` |
| `HistoryService.Clear` | input gains `source` (default `all`) — clears only that source |
| `UIStateService.Get` / `Save` | state gains `history_source` |
| `RequestService.Send` | unchanged for the UI; the binding sets origin `user` internally (no new JSON field) |

## New events (`internal/infra/server/event/events.go`)

| Event name | Payload | Emitted when |
| --- | --- | --- |
| `ai:access-status-changed` | `AIAccessStatus` | server starts, stops, fails, client count changes |
| `ai:approval-requested` | approval (as `PendingApprovals` row) | an AI send needs approval |
| `ai:approval-resolved` | `{id, outcome}` | approved, denied, timed out, app quitting |
| `history:changed` | `{local_id}` | any history append or clear (user or AI) |
| `workspaces:changed` | `{active_local_id}` | MCP created/opened/renamed/removed/switched a workspace |

Existing events reused by MCP writes: `workspace:tree-changed`, `request:changed-on-disk`, `variables:changed`,
`cookies:changed`, `git:status-changed`.

## Frontend subscriptions (`app/bootstrap.ts`)

`history:changed` → reload History view if `local_id` is active · `workspaces:changed` → reload workspace list and
activate `active_local_id` · `ai:approval-requested` / `-resolved` → approval store · `ai:access-status-changed` →
Settings tab status line.
