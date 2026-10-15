# Data Model: Rename by Clicking the Name

Frontend state only. No Go entity, DTO or file format changes.

## Name draft (per tab type)

| Tab type | Where the draft lives | Initial value | Lifetime |
| --- | --- | --- | --- |
| Request (saved or never saved) | `request-builder` `tabs[tabId].draft.name` (existing field) | loaded request name | the builder tab: dropped by `dropTab` (close, discard); persisted in UI state when `dirty` **or name pending** (`keepDraft`) |
| Example | `request-builder` `tabs[tabId].draft.name` of the example tab (existing) | `example.name` | the example tab (`dropTab` / `examples.drop`) |
| Collection / folder | `item-settings` `items[itemId].draft.name` (existing field, now editable through `editName`) | loaded container name | the item entry: `items().drop(itemId)` on close / discard / preview reuse |
| Environment | `environments` `nameDrafts[environmentId]` (**new** `Record<string, string>`) | absent = no draft | until `dropName(id)` (close, discard, successful rename), environment deleted, or `reset()` on workspace switch |

**Saved name**: the name the sidebar shows.

- Request: `selectLocation(sidebar.tree, requestId).at(-1)`, or `null` when the request has no
  tree node (never saved).
- Example: the name of `exampleId` under `findExampleRequest(tree, exampleId)`.
- Collection / folder: the tree node's name, falling back to `items[itemId].base.name` while the
  tree has not loaded.
- Environment: `environments.find(id).name`.

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
