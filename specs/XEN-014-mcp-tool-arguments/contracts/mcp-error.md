# Contract: errors (MCP tool results and Wails bindings)

## MCP tool error (text content of a result with `isError: true`)

```json
{
  "code": "WKS-00400",
  "message": "invalid request: path is required",
  "details": [ { "field": "path", "problem": "is required" } ]
}
```

- `details` is omitted when the failure has none, for example `MCP-08404`.
- Validation failures (`*-00400` and other validator codes): `message` is `invalid request: ` +
  `<field> <problem>` joined by `; `. `details` has one entry per failing field, in the
  validator's field order.
- `field` is always the **argument name the client sent**. When the action validated a request
  DTO that names the field differently, the attribute's first path segment is mapped back
  through the argument's `validated_as` tag (`local_id` → `id`, `apply_to_all` → `on_clash`,
  `text` → `paths`). The message uses the mapped names.
- Other failures keep their `message`. Their existing details map to
  `{field: attribute, problem: messages joined by ", "}`.
- The leak guard (`guardMCPText`) applies to the whole text, as today.

### Examples

| Call | Result |
|---|---|
| `workspace_edit {action: create, name: "Payments"}` | `WKS-00400`, `invalid request: path is required`, `[{path, is required}]` |
| `workspace_edit {action: create, name: "Payments", path: "rel/dir"}` | `WKS-00400`, `invalid request: path must be an absolute path`, `[{path, must be an absolute path}]` |
| `workspace_edit {action: create}` | `WKS-00400`, `invalid request: name is required; path is required`, two details |
| `workspace_edit {action: switch}` | `WKS-08400`, `invalid request: id is required`, `[{id, is required}]` (the validated field is `local_id`) |
| `workspace_edit {action: switch, id: "nope"}` | unchanged behaviour (not a validation failure) |

## Wails binding error (window) — shape unchanged

```json
{ "error": { "id": "<tx>", "code": "WKS-00400",
  "description": "invalid request: path must be an absolute path",
  "error_details": [ { "attribute": "path", "messages": ["must be an absolute path"] } ] } }
```

Only the wording changes. A nested field's `attribute` is its JSON path (`variables[2].name`).

## Problem wording (complete)

| Rule | Problem |
|---|---|
| `required` | `is required` |
| `required_if`, `required_unless` | `is required here` |
| `not_blank` | `must not be blank` |
| `oneof=a b c` | `must be one of a, b, c` |
| `max=N` (string / slice or map) | `must be at most N characters` / `must have at most N items` |
| `min=N` (string / slice or map) | `must be at least N characters` / `must have at least N items` |
| `gt=N` | `must be greater than N` |
| `gte=N` | `must be at least N` |
| `email` | `must be an email address` |
| `absolute_path` | `must be an absolute path` |
| `has_id` | `must have an id` |
| `variable_scope` | `must be a variable scope (globals, env:<id> or collection:<id>)` |
| `import_input` | `needs at least one path or some pasted text` |
| `safe_subfolder` | `must be a folder inside the workspace` |
| anything else | `is not valid` |
