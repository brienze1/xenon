# Contract: MCP tools for the runner

The guard, workspace scoping and AI-activity recording are the same as for XEN-008. The tool
argument schemas are described as in XEN-014, with each property's `description` filled.

## `run` — destructive, open-world, workspace-scoped

| Argument | Type | Notes |
| --- | --- | --- |
| `action` | `"start" \| "stop"` | required |
| `target` | string | `start`: collection or folder path (`Orders API/Checkout`), resolved as by `tree` |
| `item_id` | string | `start`: alternative to `target` |
| `rows` | `[{request_id, alias_id?}]` | `start`: optional. Runs these rows in this order. Omitted: the remembered rows, else every request in sidebar order |
| `type` | `"functional" \| "performance"` | `start`: optional, default the remembered value, else `functional` |
| `functional` | object | `start`: optional, the `dto.RunConfig.functional` shape. Missing fields take the remembered values, else the defaults |
| `performance` | object | `start`: optional, the `dto.RunConfig.performance` shape |
| `environment` | string | `start`: optional, environment name or id, or `"none"` |
| `run_id` | string | `stop`: required |

- **`start` result**: `{ run_id, status: "running", planned }`. The run continues after the tool
  returns.
- **Approval**: when the run's environment asks for a confirmation, the AI approval flow
  (`approveAISendService.Request`) runs once with summary `Run <target> · <N> requests` before the
  start. A denial or a timeout gives the existing AI-send rejections.
- **Errors**: the same codes as `RunnerService.Start`, plus `MCP-07400` for invalid argument
  combinations.
- **Changes recorded**: `cookies` when `save_cookies` is set, and `variables` when
  `keep_variable_values` is on. `history` is never recorded.
- **`stop` result**: the `dto.RunSummaryView` of the run.

## `runs` — read-only, workspace-scoped

| Argument | Type | Notes |
| --- | --- | --- |
| `action` | `"get" \| "list"` | required |
| `run_id` | string | `get` |
| `target` / `item_id` | string | `list` |

`get` returns `dto.Run`, with secrets masked as stored. Console entries and sent requests are
included, and response bodies never are. `list` returns `dto.RunSummaryView[]`, newest first.

## Leak guard

Run payloads pass through `guardAIOutputService` like every other read tool. Stored runs are
already masked (FR-039), so the guard is defence in depth.
