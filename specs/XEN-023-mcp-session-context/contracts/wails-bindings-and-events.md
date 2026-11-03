# Contract: Wails bindings and events (XEN-023)

## `EnvironmentService` / `VariableService` (window)

| Binding | Change |
| --- | --- |
| `EnvironmentService.Get` → `Environment` | adds `revision: string` |
| `EnvironmentService.Update(UpdateEnvironment)` | request adds `base_revision: string`, `force: bool`; stale → error `ENV-19409` (409) "the environment changed on disk since it was opened"; response carries the new `revision` |
| `VariableService.GetTable` → `VariableTable` | adds `revision: string` (environment scopes; `""` otherwise) |
| `VariableService.SaveTable(SaveVariableTable)` | request adds `base_revision`, `force`; environment scope + stale + not forced → `ENV-19409`; response carries the new `revision` |
| `VariableService.SetCurrent` / `Reset` / `Persist` / `AddVariable` | unchanged contract (single-key patches applied to the latest file under the workspace lock); responses that return a table carry `revision` |
| `VariableService.Inspect` | unchanged contract (window resolves with the stored active environment) |

`frontend/src/platform/services/environments.ts` / `variables.ts` and their mappers carry `revision`,
`baseRevision`, `force`.

## `HistoryService` (window)

- `HistoryEntry` DTO adds `session?: string` (AI entries only). Search (`query`) matches the tag.

## Events

| Event | Payload | Window behaviour (new) |
| --- | --- | --- |
| `workspaces:changed` | `{active_local_id}` (registry's window workspace, unchanged) | `refreshList(activeLocalId)`: re-read list; switch only when the window's workspace is no longer listed (as `remove` does) (FR-014) |
| `variables:changed` | unchanged | unchanged (env list + clean drafts refresh; dirty drafts kept) |
| `request:changed-on-disk` | unchanged | unchanged (dirty tab → "changed" notice) |

No new events.
