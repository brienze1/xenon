# Data Model: Rename by Clicking the Name

Frontend state only. No Go entity, DTO or file format changes.

## Name draft (per tab type)

| Tab type | Where the draft lives | Initial value | Lifetime |
| --- | --- | --- | --- |
| Request (saved or never saved) | `request-builder` `tabs[tabId].draft.name` (existing field) | loaded request name | the builder tab: dropped by `dropTab` (close, discard); persisted in UI state when `dirty` **or name pending** (`keepDraft`) |
| Example | `request-builder` `tabs[tabId].draft.name` of the example tab (existing) | `example.name` | the example tab (`dropTab` / `examples.drop`) |
| Collection / folder | `item-settings` `items[itemId].draft.name` (existing field, now editable through `editName`) | loaded container name | the item entry: `items().drop(itemId)` on close / discard / preview reuse |
| Environment | `environments` `nameDrafts[environmentId]` (**new** `Record<string, string>`) | absent = no draft | until `dropName(id)` (close, discard, successful rename), environment deleted, or `reset()` on workspace switch |

**Saved name**: the name the store that owns the draft loaded or last saved, so the draft and
the saved name always move together. The sidebar tree is the fallback only when that store doesn't
know it.

- Request: `request-builder` `tabs[tabId].savedName` (**new** field, set by `openSaved`, `save`,
  `saveAs`, `adoptRename` and `duplicateTab`; null for a never-saved request or a draft restored
  from the UI state), falling back to the tree node's name. `null` for a request never saved.
- Example: `examples` `tabs[tabId].example.name`, falling back to the tree example's name.
- Collection / folder: `item-settings` `items[itemId].base.name`, falling back to the tree node's
  name. A tree change reloads every open item (`refreshAll`), so the base follows renames made
  elsewhere.
- Environment: `environments.find(id).name`.

The tree alone would race: a git-sync rename can change the tree before the request's
changed-on-disk reload, which would make a clean tab look renamed for a moment, and dirty enough
to skip that reload. The logic lives once, in `tabs/lib/tab-name.ts` (`tabNameOf`). The tabs store
reads it through `getState()`; `use-name-pending.ts` reads it through narrow selectors.

**Derived: name pending** (`tabs/lib/name-draft.ts`):

```
isNamePending(draft, saved) = saved !== null && draft !== undefined && draft.trim() !== saved
```

- A pending name makes the tab dirty. It is ORed into `tabs/store.ts` `isDirty` and into
  `TabItem`'s `dirty` (through `use-name-pending.ts`).
- When the trimmed draft equals the saved name, nothing is pending. That covers Escape, typing
  back to the saved name, and a rename elsewhere landing on the draft's value (FR-014).

## Name error (tabs store)

`nameErrors: Record<tabId, NameError>` where

```
NameError = { kind: 'required'; subject: 'request' | 'example' | 'collection' | 'folder' | 'environment' }
          | { kind: 'clash'; name: string }            // environment only
          | { kind: 'failed'; error: AppError }        // what the rename operation returned
```

- **Set** by `saveTabChanges`: on a failed pre-check (`required`, `clash`) or a failed rename
  (`failed`).
- **Cleared** by `editName` on that tab (any keystroke), Escape (`dropName`), discard, `closeTab`,
  and a successful save.
- **Rendered** by the tab's `EditableName` `error` slot. Message texts are in research.md R4.

## State transitions of a tab's name

```
            click / Enter on text
  Display ───────────────────────► Editing (input, text selected)
     ▲                                │  typing → editName(draft) [pending iff trimmed ≠ saved]
     │  Escape → dropName (draft := saved, error cleared)
     ├────────────────────────────────┤
     │  Enter / blur / Tab → Display showing the draft (still pending)
     │
  Cmd/Ctrl+S or prompt Save:
     pre-check fails ─────────► error shown, nothing written, still pending
     other part fails / waits ─► nothing renamed, still pending
     rename fails ────────────► error shown, other parts already saved, still pending (the content is no longer dirty)
     rename ok ───────────────► tree / env list reload → saved = draft → not pending → dot gone (if nothing else dirty)
  Close prompt Discard ───────► draft dropped with the tab
```

There is only one write the user can't undo by discarding. If the rename call itself fails after
the content save succeeded, the content stays saved and the name stays pending with its error.
The pre-checks cover every validation the backend applies to names (empty, length, environment
clash), so that case is limited to I/O failures (disk, permissions). FR-011 is about the name not
being written when anything else fails, which this order guarantees.
