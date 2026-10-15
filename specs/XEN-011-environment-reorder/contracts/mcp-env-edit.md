# Contract: MCP `env_edit` — `reorder` action (XEN-011)

Controller: `internal/integration/entrypoint/controller/mcp_env_edit.go`. Tool metadata:
`envEditAITool` (`Effect: Destructive`, `WorkspaceScoped: true`, unchanged).

## Tool listing

`Actions` gains `"reorder"` (constant `mcpEnvEditReorderAction`), appended after `duplicate`:

```text
create, update, duplicate, reorder, set_active, add_var, set_vars, set_current, reset, persist
```

The tool description becomes: "Create, update or reorder environments, switch the active one,
and edit variables (secret values are write-only)."

## Call

```json
{ "action": "reorder", "id": "<environment id>", "index": 2 }
```

| Argument | Rule | Failure |
| --- | --- | --- |
| `id` | required, not blank | `ENV-18400` |
| `index` | required integer ≥ 0 (0 = first; ≥ count → last) | `ENV-18400` when missing or negative |

Unknown argument names are rejected by `MCPToolCall.Decode` as today.

## Result

The arranged list in the same compact shape as `env` → `list` (`dto.NewMCPEnvList`):

```json
{ "items": [ { "id": "…", "name": "PROD", "color": "red", "confirm_before_send": true, "active": false }, … ] }
```

Result carries `changes: ["variables"]` like every `env_edit` action (`WithChanges(aichangekind.Variables)`).

## Errors

| Case | Code |
| --- | --- |
| Environment id not found | `ENV-01404` |
| Bad arguments | `ENV-18400` |
| Workspace file not writable | `WKS-26500` |

## Side effects

Same service call as the binding (`ManageEnvironmentService.Reorder`): writes
`environments.order`, emits `variables:changed` (so the window's sidebar and switcher refresh),
and records AI activity like other `env_edit` actions. No approval prompt (reordering never
sends a request).
