# Contract: Wails bindings (React ↔ Go)

All shapes are the JSON of the DTOs in `internal/integration/entrypoint/dto/`. Errors use the
standard `{ code, description }` error DTO.

## `RequestService.Get` / every alias write's response — `Request.aliases[]` (changed)

```json
{ "id": "01J…", "name": "mx-test-user", "environment_ids": ["01HML…"], "request": { … } }
```

`environment_ids` is always present: `[]` means All environments. Ids are returned as stored,
including ids of deleted environments.

## `AliasService.Create(CreateAlias)` (changed)

```json
{ "local_id": "…", "request_id": "…", "name": "canary", "base_revision": "…", "environment_ids": ["01HML…"] }
```

`environment_ids` is optional (missing or `[]` means All environments). Errors: the existing
`ALS-03400`/`04400`/`05409`/`06400`, `REQ-02409`, and **`ALS-10400`** for an unknown id.

## `AliasService.SetEnvironments(SetAliasEnvironments)` → `Request` (new)

```json
{ "local_id": "…", "request_id": "…", "alias_id": "01J…", "environment_ids": ["01HML…", "01DEV…"], "base_revision": "…" }
```

- Replaces the alias's set with the normalized ids (`[]` means All environments). The alias's
  content and the Default are untouched. The response is the saved request, with a new `revision`.
- Errors: `ALS-11400` (missing/blank `local_id`/`request_id`/`alias_id`), `ALS-01404` (unknown
  alias), `REQ-02409` (stale `base_revision`), `ALS-10400` (an id that is not an environment of
  the workspace).

## `AliasService.Save(SaveAlias)` (unchanged contract)

`alias.environment_ids` in the body is **ignored**. The stored set is kept. Only `SetEnvironments`
and `Create` write the link.

## `RequestService.Send` (unchanged request shape)

New refusal: **`ALS-09403`**, description `the alias "mx-test-user" is only for HML`. It's returned
when `alias_id` names a linked alias and the active environment is not in its set, including when no
environment is active. Nothing is sent and no history entry is written.

## History DTOs (changed)

`HistoryEntry` (list and detail) gains:

```json
"environment": { "id": "01HML…", "name": "HML", "color": "green" }
```

The field is omitted when no environment was active, and on older entries.
