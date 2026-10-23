# Contract: MCP tools

## `tree` — `get` of a request: `aliases[]` (changed)

```json
{ "id": "01J…", "name": "mx-test-user", "environments": [ { "id": "01HML…", "name": "HML" } ], "method": "POST", "url": "…", … }
```

`environments` is always present (`[]` means all environments). An id that matches no environment
reads `{ "id": "<id>", "name": "Deleted environment" }`.

## `tree_edit` with `kind: "alias"` (changed)

New argument `environment_ids` (array of environment ids; ids from the `env` tool), used by
`create` and `update`:

- missing: the link is unchanged (`update`), or All environments (`create`);
- `[]`: All environments;
- `["01HML…", …]`: only those environments.

```json
{ "action": "create", "kind": "alias", "request_id": "{{login}}", "name": "mx-test-user", "environment_ids": ["{{hml}}"] }
{ "action": "update", "kind": "alias", "request_id": "{{login}}", "id": "{{alias}}", "environment_ids": [] }
```

The result is the alias as read by `tree` (with `environments`). An unknown id fails with
`ALS-10400`. The argument description reads: "Environments the alias is limited to (ids from env);
[] allows all environments; omit to keep the current setting. Alias only."

## `send` with `alias_id` (behaviour change only)

A linked alias whose environments don't include the active one fails with **`ALS-09403`**. No
approval prompt is shown, nothing is sent and no history entry is written. The AI activity is
recorded as a failed send, as for any other send error.

## `history` reads (changed)

Each send line includes the active environment's name at send time (for example
`env: HML`), or nothing when none was active. The send detail includes
`environment: { id, name, color }`.
