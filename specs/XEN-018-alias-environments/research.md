# Research: Alias Environments

All decisions below are grounded in `main` at `db23057` (2026-10-23). No open items remain.

## R1. Where the link is stored

- **Decision**: an `environments` string list inside each alias's mapping in the request YAML
  (`aliases[].environments`), written only when non-empty. Order is the user's pick order after
  normalization (R6).
- **Rationale**: aliases already live in the request file (`model.RequestAlias`,
  `internal/integration/persistence/model/request_alias.go`) and are written field by field through
  `Node()` / `pkg/yamlnode`. Omitting the key when empty keeps every existing file byte-identical
  (spec FR-002, SC-002). `yamlnode.StringList` already exists (used for `autoHeaders.disabled`).
- **Alternatives**: a separate index file mapping alias → environments (rejected: two files to keep
  in sync, and git conflicts split from the alias). Storing names (rejected: a rename breaks the
  link; ids are stable, `entity.Environment.ID`).

## R2. Where the guard runs

- **Decision**: in `sendRequest.Send` (`internal/application/service/send_request.go`), right
  after `s.alias(...)` resolves the alias and before `CheckSendPreflight`. If
  `!alias.IsAvailableIn(scopes.environment)`, return `usecase.RejectUnavailableAlias(...)`.
- **Rationale**: the app's `RequestService.Send` controller and the MCP `send` tool (`mcp_send.go`
  → the same `SendRequestService`) both go through this one function. `scopes.environment` is
  already resolved by `requestVariables.activeEnvironment`, which returns `nil` for no active
  environment *and* for a remembered active environment that was deleted. That is exactly the
  "no active environment" case of FR-003. Running before preflight means the guard runs before the
  confirm-before-send prompt (`AwaitsConfirmation`) and the MCP approval: `mcp_send` asks for
  approval only when `Send` returns the unconfirmed-environment error, which the guard
  pre-empts. It also runs before any script, HTTP call or `saveHistoryEntry`.
- **Alternatives**: a check in each controller (rejected: duplicated, and a future caller such as
  the XEN-017 runner could miss it).

## R3. Who computes availability for the UI

- **Decision**: Go returns each alias's `environment_ids` on `RequestService.Get` / alias writes.
  The frontend computes `isAliasAvailable(alias, activeId)` (empty → true; else
  `activeId !== null && ids.includes(activeId)`) in
  `features/request-builder/lib/alias-availability.ts`, only to grey out entries and disable Send.
- **Rationale**: the active environment changes through the environments store (`activeId`,
  `features/environments/store.ts`) without reloading requests. Asking Go on every switch would
  add a round trip for a one-line membership test. Constitution III is kept because Go is the only
  authority: a send the UI wrongly allowed is still refused with `ALS-09403`, and the UI shows that
  error like any other send error. The rule is mirrored, not moved, and Vitest cases mirror the BDD
  cases one to one.
- **Alternatives**: a Go `availableAliasIds` field recomputed on an environment-switch event
  (rejected: more wiring for the same result, and still stale between the event and the refetch).

## R4. A separate write for the link

- **Decision**: a new `ManageAliasService.SetEnvironments(ctx, localId, requestId, aliasId string,
  environmentIds []string, baseRevision string, force bool)`, exposed as the
  `AliasService.SetEnvironments` binding with DTO `SetAliasEnvironments`. `Create` gains an
  `environmentIds` argument. `Save` (content) keeps the stored set: `RequestAlias.Edited` replaces
  only `Version`, so the field survives unchanged.
- **Rationale**: the user decided the environments change saves immediately and never saves or
  discards unsaved content edits (spec Clarifications). Folding it into `Save` would force sending
  the draft content. The pattern matches `Delete`: `load` (stale check `REQ-02409`) →
  `ResolveAlias` (`ALS-01404`) → write with `SaveRequestAliases.SaveAliases`.
- **Alternatives**: extending `SaveAlias` with an optional field (rejected for the reason above).

## R5. Validating ids

- **Decision**: `SetEnvironments` and `Create` load `ListEnvironments.List(ctx, root)` once and
  refuse any normalized id that is not the `ID` of a listed environment with **`ALS-10400`**
  (BadRequest, "the environment does not exist in this workspace"). Malformed environment files
  (`NewMalformedEnvironment`, id prefixed with the error prefix) are not valid targets.
- **Rationale**: FR-011. Stored ids that *later* stop matching (deletion, git pull) are not
  re-validated on read or on content saves (FR-012). They are only dropped when the user saves a
  new choice that doesn't include them. The dialog submits only the boxes still ticked, and a
  leftover `Deleted environment` row can be unticked.
- **Note**: because validation happens on write only, a `SetEnvironments` call that still names a
  deleted id is refused. The Environments dialog therefore lists each deleted id as a flagged
  **Deleted environment** row with the text "Removed when you save", and never submits it. Saving
  the dialog drops deleted ids and keeps the ticked existing environments. If that leaves nothing
  ticked under "Only these environments", Save is disabled until the user ticks one or picks All
  environments. Keeping a dead id is never useful: it can only keep the alias unavailable.

## R6. Normalization

- **Decision**: `entity.NormalizedEnvironmentIDs(ids)` trims each id, drops blanks and drops
  duplicates, keeping the first occurrence's order. An empty result means All environments.
- **Rationale**: spec edge case "Duplicate ids… stores it once". Pure domain function, unit-tested.

## R7. Error codes

- **Decision**: `ALS-09403` unavailable alias (`errs.ForbiddenError`, precedent `CKE-09403`);
  `ALS-10400` unknown environment; `ALS-11400` invalid `SetAliasEnvironments` body (validator,
  precedent `ALS-06400`/`07400`/`08400`). The highest existing `ALS-` sequence is `08`.
- **Message for `ALS-09403`**: `the alias "<name>" is only for <label>, <label>`, with labels
  from R8. The UI builds its own banner text from its environment list and doesn't parse the
  message.

## R8. Environment labels (names, deleted ids)

- **Decision**: `entity.AliasEnvironmentLabels(ids, environments)` returns, per id, the current
  environment name, or `Deleted environment` when no environment has that id. Used by the
  `ALS-09403` message, the MCP alias read and the Postman export note. The UI does the same from
  its `EnvironmentSummary` list (name and colour) and shows a neutral dot for deleted ones.
- **Rationale**: FR-012 and the user's "Stay linked, flagged" decision.

## R9. History environment snapshot

- **Decision**: `entity.HistoryEnvironment{ID, Name, Color}`, set by
  `HistoryEntry.InEnvironment(scopes.environment)` in `Send` for every send entry. It is nil when no
  environment is active. It's persisted as `environment` (`omitempty`) in the history JSON model,
  exposed on the history DTO and in MCP history lines (name), and shown on the History row as a
  colour dot + name after the request/alias label.
- **Rationale**: user decision (record the active environment on every entry) plus Clarifications
  (dot + name on the row; AI reads include it). A snapshot, not a lookup, so a later rename or
  delete doesn't rewrite the past. `omitempty` keeps older entries and no-environment entries
  unchanged. Action entries (`NewActionHistoryEntry`) have no environment.

## R10. Postman export note

- **Decision**: `ExportNode.AliasRequests` takes the label lookup and, for a linked alias, appends
  a paragraph `Xenon alias for environments: HML, DEV` to the exported request's description
  (separated by a blank line when the Default has a description). Aliases set to All environments
  export unchanged. `exportPostman` loads `ListEnvironments` once.
- **Rationale**: user decision; Postman has no field for it.
- **Round trip**: importing that Postman file back gives plain requests (XEN-016: Postman imports
  create no aliases), and the note stays as description text. Nothing re-links the request, which is
  the expected outcome.

## R11. MCP shape

- **Decision**: `MCPTreeAlias.Environments []MCPAliasEnvironment{Id, Name}`, left out for an alias
  for all environments because MCP output drops every empty list (`mcp_compact_json.go`).
  `MCPTreeEdit.EnvironmentIds *[]string` (`environment_ids`), used by `create` and `update` with
  `kind:"alias"`: nil leaves the link unchanged, `[]` clears it. `create` passes the ids to `Create`.
  `update` calls `SetEnvironments` with `force: true` first, so an unknown id is refused before
  anything is written, and then saves content as today (unchanged fields keep their stored values).
  This is the same stale-write stance MCP alias edits already take (`SaveAlias(..., true)`). `send`
  is unchanged; it inherits the guard.
- **Rationale**: Clarifications (ids, as every MCP tool). The `tree` alias read needs names, so the
  tree controller resolves them through the existing environment listing.

## R12. UI placement

- **Decision**: `AliasMenu` rows show one dot + name per linked environment (`EnvironmentLook`
  colours from the environments feature). An unavailable row's `Dropdown.RadioItem` is `disabled`
  with `aria-disabled` and a greyed style; its ⋯ becomes a small submenu/list with
  **Environments…** and **Delete…**. `AliasEnvironmentsDialog` and `CreateAliasDialog` share an
  `EnvironmentPicker`: a radio for **All environments** / **Only these environments** plus a
  checkbox per environment, where at least one box is required for the second choice. The
  `UnavailableAliasBanner` sits under the request bar while the alias on screen is unavailable:
  "`mx-test-user` is only for HML." plus one **Switch to HML** button per linked environment that
  still exists (calls `useEnvironmentsStore.setActive(id)`).
- **Rationale**: spec FR-007–FR-010; `frontend-architecture` (feature components, shared dropdown
  primitives).
