# Research: Request Aliases (XEN-016)

Every decision below was settled against `origin/main` at `06d0d82` (2026-10-15). No unknowns remain.

## R1 — Where aliases live on disk

- **Decision**: a top-level `aliases:` list in the request's YAML file, after `examples:` and before
  `postman:`. Each item has `id`, `name` and the same sendable keys a request has (`method`, `url`,
  `params`, `pathVariables`, `headers`, `autoHeaders`, `auth`, `body`, `settings`, `scripts`,
  `inheritScripts`). Written only when the request has at least one alias.
- **Rationale**: Constitution I (one file per request, files are the data) and the issue ("stored
  in the request's YAML file"). Saved examples already set the precedent: `model.Request.Examples`
  is written as `yamlnode.Reserved("examples", …)` by `Fields()` (request save keeps whatever is on
  disk) and as a real field by `FieldsWithExamples()` (only `SaveRequestExamples` rewrites them),
  `internal/integration/persistence/model/request.go`, `model/example.go`.
- **Alternatives**: a sibling file per alias (rejected — the tree loader would have to skip it and
  git/sync/export would all see extra files); overrides-only storage (rejected by the user: full
  copy).

## R2 — Keeping the Default and aliases independent

- **Decision**: the Default save path is untouched: `saveRequest.Save` writes `model.Fields()`, in
  which `aliases` is a `Reserved` field, so a Default save never rewrites aliases. A new
  `SaveRequestAliases` use case (persistence `requestPersistence.SaveAliases`, writing
  `FieldsWithAliases()`) is the only writer of the `aliases` key, and it only replaces the alias
  being changed in the loaded list (`Request.WithAlias`), so every other alias and the Default
  round-trip unchanged.
- **Rationale**: `yamlnode.Patch` skips a `KeepExisting` field that exists and inserts it when the
  key is absent (`pkg/yamlnode/patch.go`), which is exactly the examples behaviour; reusing it gives
  independence with no new mechanism.
- **Alternatives**: sending the full alias list from the UI on every save (rejected — a stale tab
  would overwrite aliases changed by MCP or git).

## R3 — Stale-save protection for aliases

- **Decision**: alias create, save and delete check the file revision like a request save:
  `findRequest.FindRevision` + `RequestChange.BlocksSave(baseRevision, force)`, refusing with the
  existing `REQ-02409` conflict. Every alias call returns the updated request (with its new
  `Revision`) so the tab moves every version to the new base revision.
- **Rationale**: `manageExample.Save` does the same (`usecase.RejectStaleExampleSave`, `REQ-02409`);
  one file has one revision, shared by all versions.

## R4 — Sending an alias, and its scripts

- **Decision**: `dto.SendRequest` gains `alias_id`. The send service resolves the stored request
  once (it already does, for scripts: `send_request.go` loads `stored` and calls
  `request.WithScriptsOf(*stored)`) and uses `stored.Version(aliasId)` instead of `stored`, so an
  alias whose scripts were not edited in the tab runs **the alias's stored scripts**, not the
  Default's. An `alias_id` the stored request does not have is refused with `ALS-01404` before
  anything is sent. The history entry gets the alias id and name from the stored alias.
- **Rationale**: without this, `WithScriptsOf` would silently run the Default's scripts for an
  alias — a real bug the code reading surfaced. Taking the name from disk keeps History in Go
  (Constitution III).

## R5 — Tab state

- **Decision**: `entity.Tab` gains `AliasID string` (version on screen, empty = Default) and
  `VersionDrafts map[string]Request` (unsaved drafts of versions **not** on screen, keyed by alias
  id, `""` = Default). `Draft` keeps meaning "the draft of the version on screen", so tabs saved
  before XEN-016 load unchanged. `model`/`dto` for UI state mirror the two fields with
  `omitempty`.
- **Rationale**: drafts are restored from UI state today (`Tab.Draft *Request`,
  `features/tabs/store.ts` restore path); a map beside it keeps restore backward compatible.
- **Alternatives**: one draft list including the shown version (rejected — changes the meaning of
  `Draft` and every reader of it).

## R6 — Frontend shape

- **Decision**: the version switching lives in the request builder (`features/request-builder`):
  `TabDraft` gains `aliasId: string | null`, `aliases: AliasSummary[]` (id + name, from the loaded
  request) and `versions: Record<string, VersionState>` holding each non-shown version's draft,
  saved baseline and dirty flag. Switching swaps the shown `draft`/`savedName` with the target's
  version state. The Aliases dropdown is `features/request-builder/components/AliasMenu.tsx`,
  rendered on the right of a header row that holds the breadcrumb (`RequestBuilder.tsx`). Go calls
  go through a new `platform/services/aliases.ts`; mapping through `platform/mappers/request.ts`.
  The tab title (`features/tabs/hooks/use-tab-title.ts`) appends ` · <alias>`; the alias crumb is an
  extra item in `ItemBreadcrumb` (`features/sidebar/components/ItemBreadcrumb.tsx`) using the
  XEN-012 editable-name component.
- **Rationale**: the builder already owns drafts, dirty state, save and send for a request tab;
  a separate feature store would need to mirror all of it. `frontend-architecture` puts Go calls in
  `platform/services/*` and feature UI under `features/<feature>/components`.

## R7 — Alias name rules and error codes

- **Decision**: domain rules on `RequestAlias`/`Request`: trimmed, 1–200 runes (`isValidItemName`,
  `internal/domain/entity/item_name.go`), not `default` ignoring case, unique among the request's
  other aliases ignoring case and surrounding spaces. New codes (prefix `ALS`, like `EXM` for
  examples): `ALS-01404` alias not found, `ALS-03400` invalid name length, `ALS-04400` reserved name,
  `ALS-05409` name already used. Stale save reuses `REQ-02409`.

## R8 — Postman export

- **Decision**: a domain method `ExportNode.WithAliasesAsRequests()` (and the `Collection` it wraps)
  expands every request with aliases into the Default followed by one request per alias, in alias
  order, named `<request> (<alias>)`, id `<request id>-<alias id>`, no examples. `postmanExport.collection`
  (`internal/application/service/export_postman.go`) applies it to `node` as its first statement —
  before `usecase.EffectiveExportRoot`, `FindPostmanDrops(node)` and `FindLiteralCredentials(node)` —
  so the written collection, counts, drops and literal-credential warnings all see the expanded
  requests (literal credentials in an alias's auth are warned about like the Default's). Xenon YAML export copies the request files byte for byte
  (`internal/integration/persistence/xenon_export.go`, `fileEntries`/`exportedFiles`), so it needs
  no change; Xenon YAML import writes staged requests through `model.Request.Fields()`
  (`internal/integration/persistence/staging.go`), which carries `aliases` once the model maps them
  (locked by a scenario).

## R9 — Git sync, imports, duplicate, move, documentation

- **Git sync**: request files are not structurally merged (`PathPlan.StructuralKind` covers only
  workspace/collection/folder/environment files); a request file changed on both sides is a file
  conflict, as today. Aliases ride along; no sync code changes.
- **Imports**: the Postman/Bruno/cURL readers build `entity.Request` values with no aliases; Xenon
  YAML import reads files through `model.Request`, which now maps `aliases`. No reader changes
  except the model.
- **Duplicate / move**: duplicate copies the loaded entity (`Request.CopyAs`) and writes it through
  `Fields()`, where `Reserved` inserts `aliases` into the new file; move moves the file. No change.
- **Documentation**: `documentation_writer` and the OpenAPI generator read the Default fields of
  `entity.Request`; applying a document saves requests through `Fields()` (aliases reserved). No
  change; locked by a scenario.

## R10 — MCP surface

- **Decision**: follow the existing resource tools.
  - `tree` (read a request) — the request payload includes `aliases` (id, name, content).
  - `tree_edit` — new `kind: "alias"` with `request_id`: `create` (name), `update` (id + any of
    method/url/params/headers/body/auth/scripts), `rename` (id + name). Same rules and codes as the
    UI; recorded as AI activity with `aichangekind.Tree`.
  - `delete` — new action `alias` with `id` (alias id) and `owner_id` (request id).
  - `send` — `request_id` + optional `alias_id`; `alias_id` without `request_id` is refused as
    invalid arguments.
- **Rationale**: examples are already handled this way (`mcp_tree_edit_example.go`,
  `mcpDelete.deleteExample`, `MCPSendTarget`).

## R11 — Tab title and History label

- **Decision**: `<request name> · <alias name>` (U+00B7 with spaces) in the tab title and the
  History list row. Postman export uses `<request name> (<alias name>)` (user decision).

## R12 — OAuth 2.0 tokens and digest state per alias (user decision, 2026-10-15)

- **Fact**: tokens and digest state are keyed by `EffectiveRequest.AuthOwner`
  (`send_request.go:202` `s.tokens.current(ctx, localId, effective.AuthOwner, config)`), which is the
  id of the request or ancestor that owns the auth; the auth panel's token calls pass `owner_id`
  (`dto.OAuthTokenRequest` / `OAuthTokenReference` / `OAuthTokenUse`) and `oauthOwners.owner`
  resolves it through the tree.
- **Decision**: an alias with its own auth has the owner id `entity.AliasAuthOwner(requestId, aliasId)`
  = `<requestId>:<aliasId>` (ULIDs never contain `:`). `ResolveEffectiveRequest` is called with the
  alias version; when the send has an alias and the alias's auth is its own (`Auth.IsOwn()`), the
  send service replaces `AuthOwner` with the alias owner id. `oauthOwners.owner` splits an owner id
  containing `:` into request + alias and returns `request.Version(aliasId).Auth` with the request's
  ancestors. The frontend auth panel passes the alias owner id when an alias is shown. Deleting an
  alias calls `oauthTokens.delete` for its owner id (a missing token is not an error). Digest state
  follows the same `AuthOwner`, so it is per alias with no extra code.
- **Alternatives**: sharing the request's token (rejected by the user).
