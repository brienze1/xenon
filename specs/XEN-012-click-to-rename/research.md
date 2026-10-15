# Research: Rename by Clicking the Name

All findings below were verified against `origin/main` at `93f5381` on 2026-10-15.

## R1. What a request save does with the name

- **Decision**: a request content save sends the **saved** name (`draft` with `name` replaced by
  the tree name). The draft name reaches disk only through `TreeService.Rename`.
- **Rationale**: `saveRequest.Save` (`internal/application/service/save_request.go`) writes the
  whole `entity.Request`, `Name` included, and never renames the file. Only `manageTreeItem.Rename`
  derives a new file name (`usecase.DeriveFileName`) and updates the parent order. If the draft
  name went out with the content save, the YAML name would no longer match the file name.
- **Alternatives considered**: renaming first, then saving the content. Rejected because a
  content-save failure (conflict `REQ-02409`, validation) would leave the name written but the
  rest unsaved, which breaks FR-011.

## R2. Collection/folder settings save and the name

- **Decision**: `items().save` stays as it is. The name goes through `sidebar.rename` afterwards.
- **Rationale**: `Collection.WithSettings` (`internal/domain/entity/collection.go`) copies only
  the description, auth, headers and scripts. The settings save ignores the name.

## R3. Example save and the name

- **Decision**: `examples().save` stays as it is; it already sends `tab.example` (the saved name).
  The pending example name lives in the example tab's builder `draft.name` (loaded as
  `{ ...example.request, name: example.name }`). It is written by `sidebar.renameExample` after
  the example save.
- **Rationale**: `SavedExample.Edited` takes `edit.Name`, which is `tab.example.name`, the saved
  name, so the content save never renames.

## R4. Validation that exists today

- Tree items: `dto.RenameItem.Name` is `required,not_blank,max=200` (`TRE-09400`). There are no
  character rules: file names are derived safely and de-duplicated, so names never clash.
- Examples: `SavedExample.HasValidName` checks 1–200 runes after trimming
  (`usecase.RejectInvalidExampleName`).
- Environments: `dto.UpdateEnvironment.Name` is `required,not_blank`. `Environment.HasValidName`
  checks 1–200 runes. A clash returns `ENV-02409` (`DUPLICATE_ENVIRONMENT_NAME`) when another
  environment `HasName` it, compared case-insensitively after trimming.
- **Decision**:
  - `maxLength={200}` on the input. The browser enforces it for typing and pasting (FR-012).
  - Empty-name and environment-clash checks run in the frontend before any write (FR-010).
    The environment clash check uses the store's `environments` list with the same rule as
    `HasName`, excluding the environment itself.
  - Any other rename error the backend returns is shown inline with today's text
    (`error.description` plus a next-step hint).
- **Messages**:
  - Environment, empty: "Type a name for the environment." (today's `NAME_REQUIRED`).
  - Environment, clash: "An environment named “<name>” already exists. Choose another name."
    (today's `errorText`).
  - Request / example / collection / folder, empty: "Type a name for the <request|example|collection|folder>."
    Today these tabs silently ignore an empty name, so there is no existing message to keep. This
    one follows the environment wording.
  - Backend errors: `withNextStep(description || code, 'Choose another name, then save.')`. The
    old hint ("…then press Enter.") would be wrong now that Enter no longer saves.

## R5. Where the saved name comes from

- **Decision**: each tab's saved name comes from the store that owns its draft and loaded it:
  - request-builder `savedName`, a new field on the builder tab;
  - the examples store's `example.name`;
  - the item-settings `base.name`;
  - `environments[].name`.

  The sidebar tree is the fallback when that store doesn't know it, e.g. a draft restored from
  the UI state. A request with no `requestId` (never saved) has no saved name.
- **Rationale**: the draft and its saved name are loaded together, so they cannot disagree.
  - The tree first plan (tree name as the saved name) was rejected while implementing it. A
    git-sync rename can land in the tree before the request's changed-on-disk reload, which would
    make a clean tab look renamed. That marks it dirty, and the reload is then skipped.
  - For collections and folders, the item settings' base and the tree can also disagree
    briefly, e.g. test fixtures, or a load that lags the tree.
  - Renames made elsewhere still reach every store:
    - `adoptRename` (requests, examples) updates `savedName`, or the example;
    - a tree change reloads every open item (`refreshAll` in `app/bootstrap.ts`);
    - the environment list reloads after a rename.
- **Change**: `useTabTitle` returns the tree name for a saved request, falling back to the
  builder's `savedName`, and the tree example name for an example tab. It returns the draft name
  only for a request the tree doesn't have. Today it returns `draft.name`, which would show
  unsaved names in the tab label.

## R6. Dirty flags

- **Decision**:
  - `request-builder` gains `editName`, which patches `draft.name` without setting `dirty`; it is used for saved requests and example tabs. `setName` stays the content edit that sets `dirty`, used only by never-saved requests: they have no saved name to compare against, and the name is part of what the Save dialog writes. An example tab's draft also has `id === ''`, so the draft can't tell the cases apart; the tabs store, which knows the tab type, picks the action. A pending
    name is derived: `isNamePending(draftName, savedName)` is
    `savedName !== null && draftName.trim() !== savedName`. That way Escape, or typing back to
    the saved name, clears the dot without a flag to reset (FR-006, edge case "unchanged after
    trimming").
  - The same derived rule applies to `item-settings` `draft.name` and `environments`
    `nameDrafts[id]`.
  - `tabs/store.ts` `isDirty(tabId)` and `TabItem`'s `dirty` OR in the pending name.
  - `keepDraft` in the UI-state snapshot becomes `requestId === null || dirty || namePending`,
    so a pending request name survives a restart like the request's other edits.
- **Rationale**: this keeps one rule ("derive, don't duplicate", frontend-architecture State
  section) and matches how the stores already expose `isItemDirty` and `isScopeDirty`.
- Never-saved requests: the draft name is the only name, so it is never "pending". The tab is
  dirty through its existing rule, and the Save dialog proposes the draft name as today.

## R7. Renames made elsewhere (FR-014)

- `request-builder.adoptRename(requestId)` and `examples.adoptRename(exampleId)` currently
  overwrite `draft.name` in every tab showing the item. `item-settings.load` overwrites
  `draft.name` with the loaded name.
- **Decision**:
  - `adoptRename` (both stores) gains a `previousName` argument: the saved name before the
    rename. A tab whose trimmed draft name equals `previousName` takes the new name. A tab with a
    pending draft keeps it and only gets the new `baseRevision` (request) or the example's new
    saved name and revision (example).
  - `app/tree-actions.ts` `adoptTreeRename(id, previousName)` passes it through.
    `TreeItem.onRenamed(id, previousName)` sends the node's name from before the F2 or
    context-menu rename, and `SidebarPanel` and `App.tsx` forward it. Only the signature changes;
    the sidebar rename behaviour is the same (FR-013).
  - `item-settings.load` keeps `draft.name` when it was pending against the previous
    `base.name`. Otherwise it takes the loaded name, as today.
  - `item-settings.adoptName` is deleted. Its only caller was the Overview Name field, which
    this feature removes, and the collection/folder saved name comes from the tree (R5).
  - Environments need nothing: `nameDrafts` is separate from `environments[]`.

## R8. One save routine

- **Decision**: add `saveTabChanges(tab): Promise<Result<boolean, AppError>>` in
  `tabs/store.ts`. It is the order from the plan Summary, per tab type. `save()` and
  `resolveClose('save')` both call it, replacing their per-type branches. Its result keeps
  today's meaning: `ok:false` shows `closeError` in the prompt; `value:false` means waiting for a
  confirmation (documentation diff, secret-type switch), which keeps the tab open and focused.
- The pre-check failure and the rename failure are stored in a new `nameErrors: Record<tabId,
  AppError | NameProblem>` slice of the tabs store, which the name editor renders inline. It is
  cleared by editing the name, Escape, discard and closing the tab.
- **Order for a saved request**:
  1. pre-check;
  2. if `dirty`: `builder.save(tabId, { name: savedName })`, then `syncSiblings`;
  3. if the name is pending: `sidebar.rename(requestId, trimmed)`;
  4. `builder.adoptRename(requestId, previousSavedName)`, which refreshes `baseRevision` for
     every tab on the request.
- **Rationale**: one routine means the hotkey and the prompt can't drift apart.

## R9. Breadcrumb composition

- **Decision**:
  - `ItemBreadcrumb` (sidebar feature) stops calling `rename` and `renameExample` and drops
    `RenameInput`. It takes a `nameEditor: ReactNode` prop rendered as the current crumb.
  - `app/layout/TabBreadcrumb.tsx` builds that editor from `request-builder` (draft name,
    `editTabName`/`cancelTabName`) and `tabs` (`nameErrors`), and passes `EditableName`.
  - The `onRenamed` prop goes away.
- **Rationale**: the frontend-architecture rule "only `app/` composes several features". The
  sidebar feature must not depend on the request-builder draft for its own breadcrumb.

## R10. Hotkeys and keyboard

- `Mod+S` is a global `useHotkeys` binding (`app/use-app-hotkeys.ts` → `App.tsx`), and it
  already fires from inside inputs (the URL bar relies on it). The name input lets it bubble.
- Inside the input, Escape and Enter call `stopPropagation`, so tab and sidebar handlers don't
  see them (the same pattern as `RenameInput`).
- The display state is a `<button type="button">`. That makes it focusable, and Enter or Space
  activate it natively (FR-002).
