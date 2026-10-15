# Contract: Wails bindings

All calls are async from React (via `frontend/src/platform/services/updates.ts`); errors follow the
existing `errorpayload` shape. Status DTO = `contracts/events.md` → `UpdateStatus`.

## `UpdateService` (new) — `internal/infra/server/binding/update.go`

Controller: `internal/integration/entrypoint/controller/update.go` (`adapter.UpdateController`).

| Method | Input | Output | Behaviour |
|---|---|---|---|
| `GetStatus(ctx)` | — | `UpdateStatus` | Current in-memory status; `idle` with `current_version` before any check. |
| `Check(ctx)` | — | `UpdateStatus` | Runs a check (FR-007/008/009) and returns the resulting status. Busy → current status unchanged. |
| `CheckOnStartup(ctx)` | — | `UpdateStatus` | Cleanup of leftovers (FR-021), then a check. When the result is `available` and Update automatically is on, starts the background download (`automatic: true`) and returns immediately with `downloading`. With the setting off, returns `available` and downloads nothing (FR-018a). Called once per app run; a second call returns the current status. |
| `Update(ctx)` | — | `UpdateStatus` | From `available` (or `failed` with `download_failed`/`unverified`/`install_failed`): download + verify (progress events) and return `ready` with `automatic: false`, or the failure status. It does **not** restart: the UI first flushes the open tabs' UI state (`useTabsStore().flushPersist()`), then calls `Restart`, so no draft edit is lost to the debounced save (FR-014, SC-005). From `ready`: returns `ready`. Other states → current status unchanged. |
| `OpenReleasePage(ctx)` | — | `UpdateStatus` | Opens the latest release page (`release_url`) in the system browser through the existing `adapter.Browser` (the UI's **What's new** and release-page links). No release known → current status unchanged, nothing opened. |
| `Restart(ctx)` | — | `UpdateStatus` | From `ready`: install and restart (FR-020). On success the process quits (the returned status is `installing`); on failure returns `failed` with its reason. Other states → current status unchanged. The UI always flushes the tabs' UI state before calling it. |

`ServiceShutdown()` on the binding: if the state is `ready`, install without restarting
(Later → install on quit).

Errors: none of these methods returns an error for update failures — failures are statuses
(`check_failed`/`failed` + `reason`) so the UI never shows an error dialog (FR-009). Only a panic
or an internal bug returns the standard error payload.

## `SettingsService` (existing) — `Get` / `Save`

`dto.AppSettings` gains:

| JSON field | Type | Notes |
|---|---|---|
| `auto_update` | `boolean` (optional on Save) | Default `true`; omitted on Save = unchanged. |

Persisted as `settings.autoUpdate` in `<DataDir>/app.json`; absent → `true`.
