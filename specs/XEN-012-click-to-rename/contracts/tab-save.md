# Contract: Name drafts in the stores and the tab save order

## Store actions (new or changed)

### `features/request-builder/store.ts`

| Action | Contract |
| --- | --- |
| `setName(tabId, name)` | Unchanged: a content edit that sets `dirty`. Used only for a never-saved request, whose name is part of the content the Save dialog writes. |
| `editName(tabId, name)` | **New.** Patches `draft.name` without touching `dirty` or `curlPaste`. Used for saved requests and example tabs, whose pending name is derived against the tree (an example tab's draft also has `id === ''`, so the draft can't tell the two cases apart itself). |
| `TabDraft.savedName` | **New** field: the request's name on disk, set by `openSaved`, `save`, `saveAs`, `adoptRename`, and `loadDraft`'s optional fifth argument (`duplicateTab` passes it on); null for a never-saved request or a draft restored from the UI state. |
| `save(tabId, force?, options?: { name?: string })` | Same as today, but when `options.name` is given the request is sent with `draft.name` replaced by it. The tab's `draft.name` in the store is not changed by the save. |
| `adoptRename(requestId, previousName)` | Re-reads the request. For each tab on it: always takes `baseRevision`; replaces `draft.name` with the new saved name only when `draft.name.trim() === previousName`. |

### `features/examples/store.ts`

| Action | Contract |
| --- | --- |
| `adoptRename(exampleId, previousName)` | As today (re-reads the example, updates `tab.example`). It calls `builder().editName(tabId, renamed.name)` (not `setName`, so the tab doesn't turn dirty) only when the example tab's draft name trimmed equals `previousName`. |

### `features/item-settings/store.ts`

| Action | Contract |
| --- | --- |
| `editName(itemId, name)` | **New.** Sets `draft.name`. Does not set `dirty` (pending is derived). |
| `load(itemId)` | Same as today, except it keeps `draft.name` when it was pending against the previous `base.name`. |
| `save(itemId)` | Same as today, but writes `base.name` (the saved name) and keeps the draft's name afterwards. |
| `isItemNamePending(state, itemId)` | **New** selector: `draft.name.trim() !== base.name`. The tab's Save button uses it. |
| `adoptName` | **Deleted** (its only caller, the Overview Name field, is removed). |

### `features/environments/store.ts`

| Action / selector | Contract |
| --- | --- |
| `nameDrafts: Record<string, string>` | **New** state; `reset()` clears it. |
| `editName(environmentId, name)` | **New.** Sets `nameDrafts[id] = name`. |
| `dropName(environmentId)` | **New.** Deletes `nameDrafts[id]`. |
| `findNameClash(state, environmentId, name)` | **New** selector. Returns the other environment whose name equals `name` ignoring case and surrounding spaces (the same rule as Go's `Environment.HasName`), or null. |
| `deleteEnvironment(id)` | Also deletes `nameDrafts[id]`. |

### `features/tabs/store.ts`

| Member | Contract |
| --- | --- |
| `nameErrors: Record<string, NameError>` | **New** (see data-model.md). |
| `selectNameErrorText(state, tabId)`, `selectItemTabId(state, itemId)`, `selectEnvironmentTabId(state, environmentId)`, `isTabDirty(tabId)` | **New** exported selectors. Other features may use a store, never its `lib/` or `hooks/`, so the item, environment and side-panel editors read the error text and their tab id through these. |
| `editTabName(tabId, name)` | **New.** Routes by tab type: a never-saved request → `builder.setName`; a saved request or an example → `builder.editName`; a collection/folder → `items.editName`; an environment → `environments.editName`. Clears `nameErrors[tabId]`. |
| `cancelTabName(tabId)` | **New.** Puts the saved name back: `editName(…, savedName)` on the request builder or the item settings, `dropName(environmentId)` on the environments store. Clears `nameErrors[tabId]`. |
| `isDirty(tabId)` (internal) | Today's rule OR the tab's name is pending. Every caller uses it: close prompt, bulk close, preview reuse, and `handleChangedOnDisk`'s decision whether to reload or to show the "changed on disk" notice. |
| `save(tabId?)` | Calls `saveTabChanges(tab)`. Unsaved requests still open the Save dialog. Settings and Globals tabs are unchanged (no name). |
| `resolveClose('save')` | Calls `saveTabChanges(tab)`. On `ok:false` it shows `closeError`; on `value:false` it keeps the tab open and focused (today's behaviour). On success it closes the tab and prompts for the next. |
| `closeTab(tabId)` | Same as today. It also drops the environment name draft (environment tabs) and clears `nameErrors[tabId]`. Request and item drafts already go with `dropTab` / `dropItem`. |

## `saveTabChanges(tab)`: the order (FR-005, FR-010, FR-011)

```
1. name = the tab's draft name; saved = the tab's saved name (data-model.md)
   pending = isNamePending(name, saved)
2. if pending:
     if name.trim() === ''                         → nameErrors[tab] = required; return ok:false (nothing written)
     if environment && findNameClash(...) !== null → nameErrors[tab] = clash;    return ok:false (nothing written)
3. save the other changes exactly as today:
     request  : if dirty → builder.save(tabId, false, { name: saved }); then syncSiblings
     example  : if examples.isDirty → examples.save(tabId)
     item     : today's saveItemTab(tab) (settings → collection variables → documentation)
     env      : if isScopeDirty → environments.saveDraft(scope)
   ok:false or value:false → return as is (the name is NOT written)
4. if pending:
     request    : sidebar.rename(requestId, name.trim())            then builder.adoptRename(requestId, saved)
     example    : sidebar.renameExample(exampleId, name.trim())     then examples.adoptRename(exampleId, saved)
     collection/folder : sidebar.rename(itemId, name.trim())
     env        : environments.renameEnvironment(id, name.trim())   then environments.dropName(id) on success
   failure → nameErrors[tab] = failed(error); return ok:false
5. clear nameErrors[tab]; return ok:true, value:true
```

`saveTabChanges` keeps a per-tab in-flight set. A call for a tab that is already saving returns `ok:true, value:false` without writing. Step 1 reads the draft once, so an edit typed during the save stays pending afterwards.

A pending name alone (no other dirty part) skips step 3. A tab with nothing pending and nothing
dirty does nothing, as today.
