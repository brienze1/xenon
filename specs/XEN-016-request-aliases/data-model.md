# Data Model: Request Aliases (XEN-016)

## Domain — `internal/domain/entity`

### `RequestAlias` (new, `request_alias.go`)

| Field | Type | Notes |
| --- | --- | --- |
| `ID` | `string` | ULID (`idutils.NewUlid()`), unique within the request, never shown |
| `Name` | `string` | trimmed; 1–200 runes; not `default` (any case); unique among the request's aliases ignoring case/spaces |
| `Version` | `Request` | the alias's sendable content: `Method`, `URL`, `Params`, `PathVariables`, `Headers`, `DisabledAutoHeaders`, `Body`, `Auth`, `Scripts`, `InheritScripts`, `Settings`. Its `ID`, `Name`, `Description`, `Examples`, `Aliases`, `Postman`, `Bruno` are always empty |

Functions / methods:

- `NewRequestAlias(id, name string, of Request) RequestAlias` — copies only the sendable content of `of`.
- `(a RequestAlias) HasValidName() bool` — `isValidItemName`.
- `(a RequestAlias) HasReservedName() bool` — `strings.EqualFold(strings.TrimSpace(a.Name), "default")`.
- `(a RequestAlias) Renamed(name string) RequestAlias`.
- `(a RequestAlias) Edited(edit Request) RequestAlias` — replaces `Version` with the sendable content of
  `edit.WithScriptsOf(a.Version)`: scripts and inherit-scripts are taken from the edit only when
  `edit.ScriptsEdited`, otherwise the stored alias scripts are kept (the same rule a request save applies
  through the `Reserved` scripts field, which a whole-list `aliases` write cannot use). The stored alias
  keeps `ScriptsEdited = true` in the model so its scripts are always written.

### `Request` (changed, `request.go`)

New field: `Aliases []RequestAlias` (file order = creation order).

New methods (all value receivers, returning copies, mirroring the examples methods):

- `Alias(id string) (RequestAlias, bool)`
- `Version(aliasId string) Request` — `""` → the request itself; otherwise the request with its sendable
  content replaced by the alias's (`ID`, `Name`, `Description`, `Examples`, `Revision` kept). Used by send
  (scripts) and MCP reads.
- `HasAliasNamed(name, exceptId string) bool` — case-insensitive, trimmed.
- `WithAliasAppended(alias RequestAlias) Request`
- `WithAlias(alias RequestAlias) Request` — replaces the alias with the same id
- `WithoutAlias(id string) Request`
- `SendableOf(source Request) Request` — copies the sendable fields of `source` onto `r` (used by
  `NewRequestAlias` and `Edited`).
- `CopyAs` keeps `Aliases` (duplicate carries them); `MapText`/`TextFields` stay Default-only (they
  serve variable inspection of the version being sent, which is already a `Version(...)`).

### `AliasAuthOwner` (new, in `request_alias.go`)

`AliasAuthOwner(requestId, aliasId string) string` → `requestId + ":" + aliasId`;
`SplitAuthOwner(owner string) (requestId, aliasId string)` → alias id empty when there is no `:`.

### `HistoryEntry` (changed, `history_entry.go`)

New fields `AliasID string`, `AliasName string` (empty for the Default / unsaved requests).
`NewHistoryEntry(id, sentAt, request, execution, origin)` is unchanged; a new
`(h HistoryEntry) ForAlias(alias RequestAlias) HistoryEntry` sets both fields.

### `Tab` (changed, `tab.go`)

New fields `AliasID string` (version on screen, `""` = Default) and `VersionDrafts map[string]Request`
(unsaved drafts of the versions not on screen; key = alias id, `""` = Default). `Draft` keeps its meaning
(the version on screen). `LogValue` adds `aliasId`.

### `ExportNode` (changed, `export_node.go`)

New `(n ExportNode) WithAliasesAsRequests() ExportNode`: walks the node; every request with aliases is
followed by one request per alias: `Version(alias.ID)` with `ID = <request id>-<alias id>`,
`Name = <request name> (<alias name>)`, no `Examples`, no `Aliases`. The Default itself is exported
without `Aliases`.

## Application — `internal/application`

- `usecase/save_request_aliases.go` — `SaveRequestAliases` interface:
  `SaveAliases(ctx, root string, request entity.Request) (*entity.Request, error)`.
- `usecase/resolve_alias.go` — `ResolveAlias(request, aliasId) (*entity.RequestAlias, error)` → `ALS-01404`.
- `usecase/reject_invalid_alias_name.go` — `RejectInvalidAliasName()` `ALS-03400`,
  `RejectReservedAliasName()` `ALS-04400`, `RejectTakenAliasName()` `ALS-05409`;
  `usecase/check_alias_name.go` — `CheckAliasName(request, alias) error` applying the three rules.
- `adapter/manage_alias.go` — `ManageAliasService` with `Create`, `Save`, `Delete` (signatures in
  `contracts/bindings.md`).
- `service/manage_alias.go` — implementation, mirroring `manageExample`: find workspace → revision check
  (`REQ-02409`) → `findRequest.Find` → change → `CheckAliasName` → `saveRequestAliases.SaveAliases`.
- `service/send_request.go` — takes `aliasId`; uses `stored.Version(aliasId)` for scripts; records
  `entry.ForAlias(alias)`.

## Integration — `internal/integration`

### Persistence model (`persistence/model`)

- `request_alias.go` — `RequestAlias` model:

  ```yaml
  aliases:
    - id: 01J…
      name: mx-test-user
      method: POST
      url: '{{base}}/login'
      params: []
      pathVariables: []
      headers: []
      autoHeaders: {disabled: []}
      auth: {type: inherit}
      body: {mode: raw, raw: '{"username":"mx"}', …}
      settings: {…}
      scripts: {…}        # only when edited, same rule as the request
      inheritScripts: true
  ```

  `NewRequestAlias(entity.RequestAlias)`, `ToEntity()`, `fields()` using the same node helpers as
  `Request.fields`.
- `request.go` — `Aliases []RequestAlias \`yaml:"aliases" json:"-"\``; `Fields()` appends
  `optionalAliases` (`Reserved`, nothing when empty); new `FieldsWithAliases()` appends `aliasesField`
  (`Absent` when empty); key order: … `examples`, `aliases`, `postman`, `bruno`.
- `history_entry.go` — `AliasId`, `AliasName` (`json:"alias_id,omitempty"`, `json:"alias_name,omitempty"`).
- `tab.go` (UI-state tab model) — `AliasId string \`json:"alias_id,omitempty"\``,
  `VersionDrafts map[string]Request \`json:"version_drafts,omitempty"\``.

### Persistence (`persistence/request.go`)

`SaveAliases(ctx, root, request)` — `r.write(ctx, root, request.ID, stored.FieldsWithAliases())`.

### DTOs (`entrypoint/dto`) — see `contracts/bindings.md`

`RequestAlias`, `CreateAlias`, `SaveAlias`, `AliasReference`; `Request.Aliases`; `SendRequest.AliasId`;
`HistoryEntry.AliasId/AliasName`; UI-state `Tab.AliasId/VersionDrafts`; MCP `MCPTreeEdit` gains nothing
new (uses `Kind: "alias"`, `RequestId`, `Id`, `Name`, …), `MCPDelete` uses `Id` + `OwnerId`,
`MCPSendTarget.AliasId`.

## Frontend — `frontend/src`

- `platform/mappers/request.ts` — `RequestAlias { id; name; version: RequestDraft }`,
  `SavedRequest.aliases`.
- `platform/services/aliases.ts` — `createAlias`, `saveAlias`, `deleteAlias`.
- `features/request-builder/model.ts` — `TabDraft.aliasId: string | null`,
  `TabDraft.aliases: AliasSummary[]`, `TabDraft.versions: Record<string, VersionState>`;
  `VersionState { draft; saved; savedName; dirty }`; `DEFAULT_VERSION = ''`.
- `platform/mappers/ui-state.ts` — `aliasId`, `versionDrafts`.
- `platform/mappers/history.ts` — `aliasId`, `aliasName`.

## State transitions (one request tab)

```
open (new tab) ─► Default
Default ──select alias A──► A            (no prompt; drafts kept per version; preview tab pins)
any ──Create new alias "N"──► N          (file written: copy of saved Default; tab pins)
A ──delete A (confirm)──► Default        (file written; A's draft dropped)
A ──A deleted elsewhere / missing on restore──► Default
Cmd+S ──► saves the version on screen (Default → Request.Save, alias → Alias.Save)
close prompt Save ──► saves every dirty version, Default first, then aliases in file order
```
