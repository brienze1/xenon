# Contract: UI

## Aliases dropdown (`AliasMenu`)

- Order unchanged: Default, aliases, separator, Create new alias….
- Each alias row shows only the alias name; its environments are not listed in the dropdown (they
  are shown in the Environments… dialog and the Send-blocked banner).
- A row is **unavailable** when `isAliasAvailable(alias, activeId)` is false. It is then rendered
  greyed out, its radio item is `disabled` (not selectable by mouse or keyboard), and its accessible
  name ends with ", not available in <active env name | no environment>".
- Every alias row's ⋯ opens a small menu: **Environments…** (opens `AliasEnvironmentsDialog`) and
  **Delete…** (the existing `DeleteAliasDialog`). This works for available and unavailable rows.

## Environment picker (`EnvironmentPicker`, used by both dialogs)

- Radio group: **All environments** (default) | **Only these environments**.
- With "Only these environments": one checkbox per environment of the workspace (colour dot +
  name, in the switcher's order; malformed environment files excluded). At least one is required,
  otherwise the confirm button is disabled with the hint "Pick at least one environment."
- Deleted ids of the alias are listed as flagged **Deleted environment** rows reading "Removed when
  you save". They are never submitted.

## `AliasEnvironmentsDialog`

Title "Environments for <alias>". **Save** calls `setAliasEnvironments`, which saves immediately
and leaves the alias's unsaved content draft as it is. Errors are shown inline (code description).
**Cancel** closes without writing.

## `CreateAliasDialog`

Name field as today, plus the picker preset to All environments. **Create** passes the chosen ids.

## Unavailable alias on screen

When the tab shows an alias that is unavailable:

- the Send button in `RequestBar` is disabled (`aria-disabled`, tooltip = banner text);
- the store's `send(tabId)` returns without calling Go (covers Mod+Enter and the send dialogs);
- `UnavailableAliasBanner` appears under the request bar: "`<alias>` is only for <env>, <env>."
  plus one **Switch to <env>** button per linked environment that still exists. Pressing one calls
  `useEnvironmentsStore.setActive(id)`. The banner disappears and Send is enabled as soon as the
  alias is available.
- When none of the alias's environments exists any more, the banner reads "`<alias>` is only for
  environments that were deleted. Change its environments from the Aliases menu." and shows no
  Switch button.
- The alias's unsaved edits are untouched throughout.

If a send still reaches Go and fails with `ALS-09403`, the response pane shows the error like any
other send error.

## History row (`HistoryView`)

After the request/alias label (or at the start of the line when there is no alias label), a colour
dot + the environment name recorded on the entry. Nothing is shown when the entry has no
environment. The row's accessible label includes ", in <env>".
