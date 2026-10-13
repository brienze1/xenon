# Data Model: XEN-008 MCP Server

All new state is either app-level settings (in `app.json`), per-workspace history (in app-data), per-workspace UI
state, or in-memory runtime state. Nothing is written to workspace folders.

## AppSettings (changed) — `internal/domain/entity/app_settings.go`

| Field | Type | Default | Validation | Error code |
| --- | --- | --- | --- | --- |
| ScriptTimeoutMs | int | 10000 | 1000–60000 (unchanged) | existing |
| HistoryUserCap | int | 1000 | 100–20000 | `SET-02400` |
| HistoryAICap | int | 1000 | 100–20000 | `SET-03400` |
| AIAccess | AIAccessSettings | see below | see below | — |

Stored in `<DataDir>/app.json` under `settings` as `scriptTimeoutMs`, `historyUserCap`, `historyAiCap`,
`aiAccess{enabled,port,skipApprovals}`. Missing keys read as their defaults, so existing installs need no migration.

## AIAccessSettings (new) — `entity/ai_access_settings.go`

| Field | Type | Default | Validation |
| --- | --- | --- | --- |
| Enabled | bool | true | — |
| Port | int | 47821 (`DefaultAIPort`) | 1024–65535 → `SET-04400` |
| SkipApprovals | bool | false | — |

- These fields cannot be changed through MCP. The `settings_edit` tool rejects any `ai_access` key with `MCP-02403`
  (FR-006).
- `properties.MCPPortOverride` (env `XENON_MCP_PORT`), when set, replaces the **default** port (47821) while no port
  has been saved in settings. A saved port always wins. It is never persisted. Tests use it to give each scenario a free
  port.

## AIAccessStatus (new, runtime) — `entity/ai_access_status.go`

`State aiaccessstate (off | running | error)`, `Address string` (`http://127.0.0.1:<port>/mcp` when running), `Reason
string` (when error, for example `port 47821 is in use`), `Clients int` (open sessions). It is emitted as
`ai:access-status-changed` on every transition.

## SendOrigin (new) — `entity/send_origin.go`

`Source historysource (user | ai)`, `Client string` (only when ai). The Wails binding builds `user`. The MCP entrypoint
builds `ai` from the session's `clientInfo.name`, defaulting to `"AI client"` when it is empty. It is never accepted
from tool arguments.

## HistoryEntry (changed) — `entity/history_entry.go`

| Field | Type | Notes |
| --- | --- | --- |
| ID, SentAt | existing | `SentAt` is the record time for actions too |
| Kind | historykind (send \| action) | missing → send |
| Source | historysource (user \| ai) | missing → user |
| Client | string | ai only |
| SourceRequestID, Outcome, Result, Request | existing | send only |
| Activity | *AIActivity | action only |

`Matches(query)` also matches `Activity.Tool`, `Activity.Action` and `Activity.Target.Name` for action rows.

## AIActivity (new) — `entity/ai_activity.go`

| Field | Type | Notes |
| --- | --- | --- |
| Tool | string | for example `tree_edit` |
| Action | string | for example `create` |
| Target | {ID, Name, Path string} | empty when the call names no item |
| Outcome | string | `ok` \| `error` \| `denied` \| `timed_out` |
| Code | string | Xenon error code when Outcome = error |
| DurationMs | int64 | |
| Args | string | JSON, cut to 2,048 bytes with `…` appended when longer, with secret-carrying fields redacted by tool and field name (`secret_value` anywhere; `value` in `env_edit` `add_var`/`set_current`; `value` beside `secret:true`; `passphrase` anywhere — research R10), then the leak guard |

## History (changed) — `entity/history.go`

- `HistoryCap` is replaced by `HistoryCaps{User, AI int}`. `DefaultHistoryCaps = {1000, 1000}`.
- `Compact(oldestFirst, caps) (kept, changed)` keeps the newest `caps.User` entries with source=user and the newest
  `caps.AI` with source=ai, preserving order.
- `Page(query)` takes `HistoryQuery{Source historyfilter (all|user|ai), Text, Limit, Cursor}`.
- `WithoutSource(source)` returns the entries to keep on a filtered Clear.

## AIApproval (new, in memory) — `entity/ai_approval.go`

| Field | Type |
| --- | --- |
| ID | ULID (via `idutils`) |
| Client | string |
| Method | string |
| RawURL | string (as written; never resolved) |
| Environment | {ID, Name, Color} — from the `SND-05409` error details |
| CreatedAt | time |
| Outcome | approvaloutcome (pending → approved \| denied \| timed_out) |

Transitions: `pending → approved` (UI approve), `pending → denied` (UI deny, or app quit), and `pending → timed_out`
(120 s through the clock). Every terminal transition is final. A second answer for the same id returns `MCP-05409`
(already resolved).

## KnownSecrets (new, in memory) — `entity/known_secrets.go`

- `values set[string]` is the session-seen set. It only grows.
- `Forms(value)` returns raw, QueryEscape, PathEscape, JSON-escaped and base64 forms, deduplicated.
- `Add(values...)`.
- `Mask(text) string` replaces longest-first. A form of 4 or more characters is replaced anywhere. Shorter forms are
  replaced only as a whole JSON string value (`"<v>"` → `"•••••"`).
- `LogValue()` returns `{count}` only.

## UIState (changed) — `entity/ui_state.go`

Adds `HistorySource historyfilter` (`all` default), persisted in `<DataDir>/workspaces/<localId>/state.json`.

## Enums (new)

`historysource`: user, ai · `historykind`: send, action · `historyfilter`: all, user, ai · `aiaccessstate`: off,
running, error · `approvaloutcome`: pending, approved, denied, timed_out.

## Error codes (new)

| Code | Meaning |
| --- | --- |
| `MCP-01409` | Port in use (status reason) |
| `MCP-02403` | Human-only action refused (include secrets, secret→plain, AI access settings) |
| `MCP-03403` | Send denied by the user in Xenon |
| `MCP-04408` | Approval timed out |
| `MCP-05409` | Approval already resolved |
| `MCP-06404` | No active workspace |
| `MCP-07400` | Unknown action for this tool / invalid arguments |
| `MCP-08404` | Item id not found in the active workspace (unknown id, or an id from another workspace) |
| `SET-02400` / `SET-03400` / `SET-04400` | History user cap / AI cap / port out of range |
