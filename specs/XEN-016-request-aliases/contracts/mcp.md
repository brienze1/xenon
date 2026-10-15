# Contract: MCP tools and aliases

Same tools, same guard, same AI-activity recording as XEN-008. Aliases are not tree entries, so they
are addressed by their request.

## `tree` — read a request

The request payload (`dto.MCPTree…` request view) gains:

```json
"aliases": [ { "id": "01J…", "name": "mx-test-user", "method": "POST", "url": "…", "params": […],
               "headers": […], "body": {…}, "auth": {…}, "scripts": {…} } ]
```

Secrets follow the leak guard exactly as for the Default (values masked; secret variables by name).
Requests without aliases return `"aliases": []`.

## `tree_edit` — `kind: "alias"`

| Action | Arguments | Result |
| --- | --- | --- |
| `create` | `request_id`, `name` | Creates the alias from the saved Default; returns the alias (id, name, content). |
| `update` | `request_id`, `id`, any of `method`, `url`, `params`, `headers`, `body`, `auth`, `scripts` | Patches only the given fields of that alias; returns it. Default and other aliases untouched. |
| `rename` | `request_id`, `id`, `name` | Renames; returns it. |

Errors: unknown `request_id` → existing missing-item error (`MCP-08404`); unknown `id` → `ALS-01404`; name rules →
`ALS-03400` / `ALS-04400` / `ALS-05409`; other actions (`move`, `duplicate`, `save_settings`, token
actions) with `kind: "alias"` → invalid arguments (`MCP-07400`). Each successful call records AI activity with the
request as target and change kind `tree`.

## `delete` — target `alias`

Arguments: `"target": "alias"`, `id` (alias id), `owner_id` (request id). Removes the alias; result is the usual deleted
result with the request as target and change kind `tree`. Unknown → `ALS-01404`.

## `send`

`MCPSendTarget` gains `alias_id`. With `request_id` + `alias_id`, the alias's stored version is sent
(History: AI source, `alias_id`, `alias_name`; "Confirm before send" approvals as for any AI send).
`alias_id` with an ad-hoc `request` or without `request_id` → invalid arguments
(`"alias_id needs request_id"`). Unknown alias → `ALS-01404`, nothing sent.
