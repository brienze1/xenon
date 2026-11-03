# Contract: MCP tools (XEN-023)

Every call resolves its workspace and environment from the **calling session's context** (FR-003). Nothing below
changes the Xenon window.

## Tool descriptions (FR-023)

One sentence each (tool-list rule `^[A-Z][^.]*\.$`); `tools/list` budget raised from 39,000 to 39,500 characters
(measured 38,849 on `main` @ `c87aed8`).

| Tool | Old `Description` | New `Description` |
| --- | --- | --- |
| `workspace` | `List workspaces or show the active one.` | `List workspaces or show this session's workspace.` |
| `workspace_edit` | `Create, open, switch or rename a workspace.` | `Create, open, switch or rename a workspace; switching only affects this MCP session.` |
| `env_edit` | `Create, update or reorder environments, switch the active one, and edit variables (secret values are write-only).` | `Create, update or reorder environments, pick this session's active one, and edit variables (secret values are write-only).` |
| `send` | `Send a saved request (or one of its aliases, with alias_id) or an ad-hoc request with the active environment and return the response.` | `Send a saved request (or one of its aliases, with alias_id) or an ad-hoc request with the session's environment and return the response.` |

`env`, `run` and `snippet` descriptions are unchanged (their behaviour follows the session without new wording).

## `workspace`

- `list` → `items[].active` is true only for the session's workspace.
- `active` → the session's workspace and its environments, `active` = session's environment. No workspace →
  `MCP-06404` (unchanged code).

## `workspace_edit`

| Action | Effect | Result changes (AI change kinds) |
| --- | --- | --- |
| `switch {id}` | session workspace = id (FR-004). Unknown id → existing not-found error. No run stopped, no watcher/UI-state change. | none |
| `create {name, path}` | registers the workspace without activating it (FR-012); session workspace = new id | `workspaces` |
| `open {path}` | same as create for an existing folder | `workspaces` |
| `rename {id, name}` | unchanged (data) | `workspaces` |

## `env`

- `list` → `active` marks the session's environment.
- `inspect` → resolves with the session's environment.

## `env_edit`

- `set_active {id?}` → session-only (FR-005). `id` must be an environment of the session's workspace (unknown →
  existing `ENV-01404`); omitted/empty → none. Result `{active_id}` as today. Result changes: **none** (no file
  changed). The stored active environment (`variables.json` `activeEnvironmentId`) is never written.
- All other actions: unchanged contract; they apply to the latest stored environment inside the workspace lock
  (FR-020) and never return `ENV-19409`.

## `send`

New optional argument:

```json
{ "environment": "Prod" }   // description: "Environment name or id, or none, for this send only."
```

- Absent → session environment. Unknown → `RejectInvalidArguments` with reason
  `environment must be an environment name or id, or none` (same as `run`).
- The session's environment is unchanged by the override.

## `run`

- `start` without `environment`: when the target's run configuration uses "the active environment", the session's
  environment is used; a configuration pinned to an environment, or an explicit `environment`, wins (unchanged).

## `snippet`

- Resolves variables with the session's environment.

## `history`

- Lines for AI entries gain `session` (the 6-character tag) next to `client`.

## Session lifecycle

| Event | Effect |
| --- | --- |
| `notifications/initialized` | context seeded from the window's workspace + its stored active environment (FR-002) |
| Session closed (disconnect / HTTP DELETE) | context deleted (FR-009); pending approvals denied (unchanged) |
| Session's workspace unregistered | workspace-scoped calls → `MCP-06404` (FR-010) |
