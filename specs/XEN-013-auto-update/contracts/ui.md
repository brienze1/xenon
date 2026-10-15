# Contract: UI (Settings section, toast, footer notice)

Texts are exact; `<v>` = a version, `<folder>` = a path. Every **What's new** / release-page
link calls `UpdateService.OpenReleasePage` (React never opens URLs itself).

## Settings → About & Updates (`features/settings/components/AboutUpdatesSection.tsx`)

It is the **first** section of the Settings tab, so opening Settings shows it without
scrolling. Always shows `Version <current>`. Below it, by state:

| State (reason) | Status line | Actions |
|---|---|---|
| `idle` | — | **Check for updates** |
| `checking` | `Checking…` | Check for updates (disabled) |
| `up_to_date` | `Xenon is up to date` · `Latest version <v>` | Check for updates |
| `available` | `Update available` · `Latest version <v>` · link **What's new** | **Update**, Check for updates |
| `no_package` | `No update package for this system` · `Latest version <v>` · What's new | Check for updates |
| `check_failed` | `Couldn't check for updates` | Check for updates |
| `downloading` | `Downloading… <n>%` (progress bar) · Latest version · What's new | — |
| `ready` | `Update ready` · Latest version · What's new | **Restart now** |
| `installing` | `Installing…` | — |
| `failed` (`download_failed`) | `The update could not be downloaded` | Update, Check for updates |
| `failed` (`unverified`) | `The update could not be verified — nothing was changed` | Update, Check for updates |
| `failed` (`not_packaged`) | `Updates install only in a packaged Xenon` | Check for updates |
| `failed` (`not_writable`) | `Xenon can't replace itself at <folder>. Move Xenon to a folder you own, such as ~/Applications, or install it from the release page.` + link to the release page | Check for updates |
| `failed` (`install_failed`) | `The update could not be installed — Xenon was not changed` | Update, Check for updates |

The **Update automatically** switch (`role="switch"`) sits in this section and belongs to the
settings form (draft + Save), default on.

## Toast (`shared/ui/toast/Toast.tsx`, used by `features/updates/components/UpdateReadyToast.tsx`)

- Fixed bottom-left, above the footer, `role="status"`, `aria-live="polite"`.
- Shown when an `update:status` event arrives with `state: ready` and `automatic: true`. It
  never auto-dismisses; it stays until Restart now, Later or ×. A failed automatic download
  shows no toast (the section shows the failure).
- Text: `Xenon <v> is ready to install.` Buttons: **Restart now** (calls `Restart`), **Later**
  (closes it). A close (×) button behaves as Later. Shown once per app run.

## Footer notice (`features/updates/components/FooterUpdateNotice.tsx`)

Rendered by `app/App.tsx` in the footer status slot, with or without an open workspace.

| State | Notice | Click |
|---|---|---|
| `ready` | `Update ready — Restart` | calls `Restart` |
| `available` | `Update available` | opens the Settings tab (`useTabsStore().openSettingsTab`, which also works with no workspace open) |
| anything else | nothing | — |
