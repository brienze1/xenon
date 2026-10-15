# Contract: Request tab UI for aliases

## Layout

`RequestBuilder.tsx` (mode `request`, saved request only) renders a header row: the breadcrumb on the
left, `AliasMenu` on the right (top-right corner of the request tab). Unsaved requests (`requestId`
null) and example tabs render no `AliasMenu`.

## `AliasMenu` — `features/request-builder/components/AliasMenu.tsx`

- Trigger: a button labelled with the version on screen (`Default` or the alias name), `aria-haspopup="menu"`,
  accessible name `Aliases: <version>`.
- Menu items, in order: `Default`; every alias in file order; a separator; `Create new alias…`.
  The selected version has `aria-checked="true"` (menuitemradio).
- Each alias item has a trailing `⋯` button (`aria-label="Alias actions for <name>"`) opening a
  submenu with one item, `Delete…`. The Default has none.
- `Create new alias…` opens a name prompt dialog (input `maxLength=200`, `Create` / `Cancel`).
  Errors from Go (`ALS-03400/04400/05409`) show inline under the input; the dialog stays open.
  On success the tab switches to the new alias.
- `Delete…` opens a confirm dialog: title `Delete alias?`, body `"<name>" will be removed from
  <request>. This can't be undone.`, buttons `Delete` (danger) / `Cancel`.
- Keyboard: full menu keyboard support from the design system's Menu (arrows, Enter, Escape).

## Store — `features/request-builder/store.ts`

| Action | Contract |
| --- | --- |
| `openSaved(tabId, requestId)` | Also loads `aliases`; starts on Default unless the tab was restored with an `aliasId` that exists. |
| `selectVersion(tabId, aliasId \| null)` | Stores the shown version's draft/saved/dirty into `versions[current]`, loads the target's (from `versions` or from the loaded request). No prompt. Pins a preview tab. |
| `createAlias(tabId, name)` | `platform/services/aliases.createAlias` with `baseRevision`; on success updates `aliases`, `baseRevision`, selects the new alias; pins a preview tab. Returns the error to the dialog on failure. |
| `deleteAlias(tabId, aliasId)` | Calls `deleteAlias`; drops `versions[aliasId]`; if shown, selects Default; updates `baseRevision`; pins a preview tab. |
| `save(tabId, force?, options?)` | Default on screen → `RequestService.Save` (today). Alias on screen → `AliasService.Save` with the shown draft and the name draft (XEN-012 name pending rules apply to the alias name). |
| `saveAllVersions(tabId)` | Used by the close prompt's Save: saves every dirty version (Default first, then aliases in file order); stops at the first error and keeps the tab open with that error. |
| `isDirty(tabId)` | True when the shown version or any entry in `versions` is dirty (or its name is pending). |
| `send(tabId)` | Sends the shown draft with `alias_id` = shown alias (or none). |

`duplicateTab` (tabs store) copies `aliasId`, `aliases` and `versions` into the new tab.
The auth panel's OAuth 2.0 token actions use `<requestId>:<aliasId>` as `owner_id` when an alias with
its own auth is shown.

Conflict banner (`REQ-02409` on an alias or Default save): **Keep Your Version** re-saves the shown
version with `force`; **Load from Disk** reloads the request and replaces only the shown version's
draft; every version adopts the new `baseRevision` and the other versions keep their drafts.

On a file reload (changed on disk / alias deleted elsewhere): refresh `aliases`; a shown alias that is
gone → Default and its draft is dropped.

## Tab title — `features/tabs/hooks/use-tab-title.ts`

Request tab with a shown alias: `<request title> · <alias name>` (saved alias name; a pending rename
shows after saving, as XEN-012).

## Breadcrumb — `features/sidebar/components/ItemBreadcrumb.tsx`

New optional prop `aliasEditor: ReactNode`. When given, the request's own (saved) name is rendered as
plain, non-editable text before it, and the `aliasEditor` (the XEN-012 editable name bound to the alias
name draft) is the last crumb. Renaming the request itself is done with Default on screen (or from the
sidebar), so a tab has at most one pending name at a time. For the same reason the side panel's
Info panel (`features/side-panel/components/InfoPanel.tsx`) shows the request name as plain text while an
alias is shown.

## Persistence of tab state

`features/tabs/store.ts` writes `aliasId` and `versionDrafts` (drafts of dirty non-shown versions)
into UI state and restores them; a restored alias id missing from the loaded request falls back to
Default and drops that draft.

## History row

`features/sidebar/components/HistoryView.tsx`: the label is `<request name> · <alias name>` when the
entry has `aliasName`.
