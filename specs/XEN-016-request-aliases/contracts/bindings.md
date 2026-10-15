# Contract: Go ↔ React bindings for aliases

All calls go through `frontend/src/platform/services/*` (never the generated bindings directly).
Errors use the existing `errs` shape (`code`, `description`).

## New binding `AliasService` — `internal/infra/server/binding/alias.go`

Thin adapter over `adapter.AliasManageController`
(`internal/integration/entrypoint/controller/alias_manage.go`, full controller pattern: validate DTO →
`ToEntity` → `appAdapter.ManageAliasService` → error handler → panic recovery). Registered beside
`ExampleService` in the Wails service list (`internal/infra`).

| Method | Input DTO | Output | Behaviour |
| --- | --- | --- | --- |
| `Create(ctx, dto.CreateAlias)` | `{local_id*, request_id*, base_revision, name*}` | `dto.Request` (whole request incl. `aliases`, new `revision`) | Copies the **saved Default** into a new alias appended last. Name rules → `ALS-03400` / `ALS-04400` / `ALS-05409`. Stale `base_revision` → `REQ-02409`. |
| `Save(ctx, dto.SaveAlias)` | `{local_id*, request_id*, base_revision, force, alias: dto.RequestAlias*}` | `dto.Request` | Replaces the alias with `alias.id` (name + sendable content). Unknown id → `ALS-01404`. Name rules as above (a clash ignores the alias itself). Stale → `REQ-02409` unless `force`. Never touches the Default or other aliases. |
| `Delete(ctx, dto.AliasReference)` | `{local_id*, request_id*, alias_id*, base_revision}` | `dto.Request` | Removes the alias. Unknown id → `ALS-01404`. Stale → `REQ-02409`. When the last alias goes, the `aliases` key is removed from the file. |

`*` = `validate:"required,not_blank"`; `name` also `max=200`.

### `dto.RequestAlias`

```json
{ "id": "01J…", "name": "mx-test-user", "request": { /* dto.Request sendable fields */ } }
```

`request` reuses `dto.Request` (method, url, params, path_variables, headers, auto_headers, body, auth,
scripts, inherit_scripts, scripts_edited, settings); its `id`, `name`, `description` are ignored.

## Changed: `RequestService`

| Method | Change |
| --- | --- |
| `Get(dto.OpenRequest)` → `dto.Request` | `dto.Request` gains `aliases: dto.RequestAlias[]` (empty array when none). |
| `Save(dto.SaveRequest)` | Unchanged contract. Saves the Default only; `request.aliases` in the input is ignored; aliases on disk are kept (`Reserved`). |
| `SaveAs(dto.SaveRequestAs)` | Unchanged contract. The UI passes the version on screen; the new request has no aliases. |
| `Send(dto.SendRequest)` | New optional `alias_id`. When set, `request.id` must be a saved request having that alias, else `ALS-01404` (nothing sent). Scripts not edited in the tab come from the stored alias. History records `alias_id` + `alias_name`. |

## Changed: OAuth 2.0 token calls

`owner_id` in `OAuthTokenRequest` / `OAuthTokenReference` / `OAuthTokenUse` may be an alias owner id
`<requestId>:<aliasId>`; the frontend passes it when the shown alias has its own OAuth 2.0 auth. Go
resolves it to that alias's auth; unknown alias → `ALS-01404`. `AliasService.Delete` also deletes the
alias's token.

## Changed: History

`dto.HistoryEntry` gains `alias_id`, `alias_name` (omitted when empty). The History list label is
`<request.name> · <alias_name>` when `alias_name` is set.

## Changed: UI state

`dto.Tab` gains `alias_id` (string, omitted when empty) and `version_drafts`
(`{ [aliasId: string]: dto.Request }`, `""` = Default, omitted when empty).

## Error codes (new)

| Code | Kind | Message |
| --- | --- | --- |
| `ALS-01404` | not found | the alias does not exist |
| `ALS-03400` | bad request | the alias name must have 1 to 200 characters |
| `ALS-04400` | bad request | "Default" is reserved; choose another alias name |
| `ALS-05409` | conflict | this request already has an alias with that name |
