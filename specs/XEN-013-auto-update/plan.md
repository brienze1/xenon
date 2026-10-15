# Implementation Plan: Auto-Update Xenon From GitHub Releases

**Branch**: `xen-013` | **Date**: 2026-10-15 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/XEN-013-auto-update/spec.md`

## Summary

Every merge to `main` runs a new release workflow. It computes the next version from the `VERSION`
file (`major.minor`) and the existing `v*` tags (next patch), builds macOS universal / Windows x64 /
Linux x64 with that version embedded by ldflag, and publishes `v<x.y.z>` with three archives and
`checksums.txt`.

The app gets an updater, all in Go:
- **Check.** Ask GitHub `releases/latest`, compare against the running version and pick the asset
  for this platform.
- **Download and verify.** Stream the asset into `<DataDir>/updates/` while hashing it with
  SHA-256, then compare against `checksums.txt`.
- **Install.** Swap the `.app` / `.exe` / binary with two renames in its own folder, with
  rollback if the second rename fails.
- **Restart.** Spawn the new executable, which waits for the old pid, then quit through Wails.

A new `UpdateService` binding exposes `GetStatus`, `Check`, `CheckOnStartup`, `Update`
(download + verify → `ready`), `Restart` (install + restart) and `OpenReleasePage`. Before every
`Restart` the UI flushes the tabs' UI state, so the restart loses no draft edit. It emits `update:status` with the full status on every change.

The setting `auto_update` (default on) joins the existing app settings in `app.json`.

React adds:
- the About & Updates section, first in Settings;
- a bottom-left toast (new `shared/ui/toast`);
- a footer notice;
- the one-time `CheckOnStartup` call from bootstrap.

Decisions and alternatives: [research.md](research.md).

## Technical Context

**Language/Version**: Go 1.26 (toolchain 1.27.2), TypeScript 5 + React 19, Node LTS for
`tools/release`.

**Primary Dependencies**: none new. Uses the Go stdlib (`archive/tar`, `archive/zip`,
`compress/gzip`, `crypto/sha256`, `os/exec`), the existing `pkg/http` client factory, Wails v3
(quit, only in `internal/infra/`), and the `gh` CLI on GitHub runners.

**Storage**:
- `<DataDir>/app.json` → `settings.autoUpdate` (absent = true).
- `<DataDir>/updates/` holds the staging downloads, emptied at every start.
- Install side files sit next to the installed app and are removed after the install or at the
  next start.
- The update state is in memory only.

**Testing**:
- godog BDD (`test/integration/features/updates/`).
- Faked boundaries:
  - the GitHub API and the downloads, via the generic mock API with archive fixtures;
  - the install location, a temp folder set with `XENON_INSTALL_PATH`;
  - the platform, set with `XENON_UPDATE_PLATFORM`;
  - the process restart/quit, through a new `mock/app_process.go`;
  - the browser opener (existing mock);
  - the event emitter.
- Vitest + RTL for the section, toast, footer notice, store, service and mapper.
- `node --test` for `tools/release/next-version.mjs`.
- Go unit tests for `entity.Version` and the installer's archive safety checks.

**Target Platform**: macOS 12+ (universal), Windows 10/11 x64, Linux x64 (webkitgtk-6 / gtk4
installed, as for the existing build).

**Project Type**: desktop app (Wails v3) + CI workflow.

**Performance Goals**:
- SC-003: the startup check never delays the window. It is triggered after render, and the
  call is async.
- Progress events are capped at 4/s.

**Constraints**:
- No elevation prompt.
- No new dependency.
- Only URLs listed by the release are downloaded.
- Nothing is installed without a matching SHA-256.
- Same behaviour on all three OSes.

**Scale/Scope**: one release per merge; archives of about 10–30 MB; one check per start, plus
manual checks.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first** — Xenon still needs no account and no Xenon service. The only network
  call is to GitHub's public releases API and the release downloads. It works fully offline:
  a failed check is a quiet status (FR-009). Workspaces are untouched.
- [x] **II. Secrets** — No secrets are involved: no token, unauthenticated API. Logs carry only
  versions, file names, sizes, paths and statuses (FR-023).
- [x] **III. Go owns logic** — Go does the check, download, hashing, extraction, file swap,
  restart and opening the release page. React only calls `platform/services/updates.ts` and
  renders the status. Wails is imported only in `internal/infra/` (binding, process adapter,
  router).
- [x] **IV. Architecture skills** — Every file below has its layer and folder per
  `architecture.md`. The frontend respects the boundary table: the Settings section reads
  `features/updates/store.ts`, and the toast lives in `shared/ui/`.
- [x] **V. BDD-first** — The feature files are listed per requirement below, written red by
  `/create-tests` before code. Only outer boundaries are faked (HTTP targets, the install-folder
  root, the process quit/spawn, the browser). Persistence, services,
  use cases and controllers stay real.
- [x] **VI. Postman parity / PRD scope** — The feature changes no Postman-parity behaviour.
  The PRD does not yet cover distribution, so the product owner's decisions (issue #14 and this
  spec's Clarifications) are added to the PRD decisions list in the same PR (task T051). That
  satisfies "out of scope until the PRD changes".
- [x] **VII. Cross-platform** — All three OSes self-install (rename-based swap, valid on each).
  Paths use `filepath`. Windows' locked `.exe` is handled by rename plus cleanup on the next
  start. CI builds and releases all three.

Post-design re-check: unchanged, all ✅. No Complexity Tracking entries.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-013-auto-update/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── bindings.md
│   ├── events.md
│   ├── release-assets.md
│   └── ui.md
├── checklists/
└── tasks.md
```

### Source Code (repository root)

```text
VERSION                                                  ← new: "0.1"
tools/release/next-version.mjs (+ next-version.test.mjs) ← new (R2)
.github/workflows/release.yml                            ← new (R3)
Makefile                                                 ← `release-check` target, called by `validate`
build/{darwin,windows,linux}/Taskfile.yml                ← `-X …properties.version={{.XENON_VERSION}}`

internal/domain/
├── entity/version.go, release.go, release_asset.go, update_platform.go, update_status.go   ← new
├── entity/app_settings.go, app_settings_patch.go       ← AutoUpdate
└── enums/update_state.go + updatestate/value.go, update_failure.go + updatefailure/value.go ← new

internal/application/
├── usecase/find_latest_release.go          (port → webservice)
├── usecase/download_release_asset.go       (port → webservice; streams + hashes, progress callback)
├── usecase/find_release_checksums.go       (port → webservice)
├── usecase/install_update.go               (port → persistence: probe, extract, swap, rollback)
├── usecase/clean_update_leftovers.go       (port → persistence)
├── usecase/restart_app.go                  (port → webservice over the process adapter)
├── usecase/open_release_page.go            (port → existing browser webservice pattern)
├── usecase/find_update_state.go, save_update_state.go (port → in-memory persistence)
├── usecase/publish_update_status.go        (port → publisher)
├── adapter/check_for_updates.go, download_update.go, apply_update.go, startup_update.go   (service interfaces)
└── service/check_for_updates.go, download_update.go, apply_update.go, startup_update.go   ← new

internal/integration/
├── adapter/update_controller.go, update_webservice.go, update_persistence.go,
│   update_status_publisher.go, app_process.go                              ← new interfaces
├── entrypoint/controller/update.go                                         ← new controller
├── entrypoint/dto/update_status.go                                         ← new; app_settings.go gains auto_update
├── webservice/update.go (+ dto/github_release.go)                          ← GitHub releases + downloads
├── webservice/app_restart.go                                               ← restart via adapter.AppProcess
├── persistence/update_install.go                                           ← extract/swap/rollback/cleanup
├── persistence/update_state.go                                             ← in-memory state
├── persistence/model/app_settings.go                                       ← autoUpdate *bool
└── publisher/update_status.go (+ dto/update_status.go: "update:status")    ← new

internal/infra/
├── properties/properties.go         ← `version` ldflag var, Application.Packaged, Update{ApiURL, Platform, InstallPath}
├── run.go                           ← XENON_RESTART_AFTER_PID wait before Start()
├── adapter/app_process.go           ← Spawn(path, env) detached + Quit() via Wails
├── server/binding/update.go         ← UpdateService (+ ServiceShutdown → install on quit)
├── server/router/update_routes.go, router.go
└── dependency/{binding,controller,application,integration,external,router}/…  ← factories

test/integration/
├── features/updates/check_for_updates.feature
├── features/updates/install_update.feature
├── features/updates/auto_update.feature
├── mock/app_process.go                              ← records spawn + quit
├── steps/…                                          ← register UpdateService in hooks_test.go startApp(); inject mock process; generic steps only if missing (see below)
└── resources/updates/                               ← archive fixtures per platform + checksums

frontend/src/
├── platform/services/updates.ts (+test), platform/mappers/update-status.ts (+test), platform/events.ts (onUpdateStatus)
├── platform/mappers/settings.ts                     ← autoUpdate
├── shared/ui/toast/Toast.tsx, Toast.module.css (+test)
├── features/updates/store.ts (+test), model.ts
├── features/updates/components/UpdateReadyToast.tsx, FooterUpdateNotice.tsx (+tests)
├── features/settings/components/AboutUpdatesSection.tsx (+test), SettingsTab.tsx (first section), model.ts/lib/settings-form.ts (autoUpdate)
├── app/App.tsx                                      ← mount toast; footer notice with or without workspace
└── app/bootstrap.ts                                 ← CheckOnStartup once after render
```

**Structure Decision**: The existing single-repo layout. The backend follows the 4-layer skills,
modelled on the existing GitHub webservice, window-focus webservice/adapter pair, export-progress
publisher and settings controller. The frontend follows the `frontend-architecture` boundary
table.

## Requirement → BDD lock

`/create-tests` writes these, tagged `@all @xen_013 @updates`:

| Feature file | Locks |
|---|---|
| `test/integration/features/updates/check_for_updates.feature` | FR-006, FR-007, FR-008, FR-009, FR-010, FR-024 and these edge cases: newer local build (no downgrade), invalid tag, no package for the platform, draft/prerelease answer, 403/429/500, closed port, `OpenReleasePage`. |
| `test/integration/features/updates/install_update.feature` | FR-011, FR-012, FR-013, FR-014, FR-015, FR-016, FR-021, FR-022, FR-023. Covers progress events, a bad checksum, a missing checksums line, a dropped download (closed port / 500), an unwritable folder (made portable by pointing `XENON_INSTALL_PATH` at `<test root>/blocker.txt/Xenon.app`, where `blocker.txt` is a regular file, so the writability probe fails the same way on all three OSes; chmod-based read-only folders are not used because Windows ignores directory mode bits), not packaged, the bundle swap (darwin-universal) and the single-file swap (linux-amd64, windows-amd64), rollback, the restart spawn path plus quit, and leftovers removed at the next start. |
| `test/integration/features/updates/auto_update.feature` | FR-017, FR-018, FR-018a, FR-018b, FR-019 (the `ready` + `automatic` event), FR-020 (`Restart`, and install on quit through the binding shutdown). Also covers the `auto_update` default on, persistence across a restart, and a quiet startup failure. |

Generic steps:
- Most needs reuse the existing catalog:
  - mock API stubs, including the resource-file body;
  - env vars;
  - binding calls and response fields;
  - events;
  - app-data files;
  - log assertions;
  - app restart.
- Add a new step only where none exists, each kept generic:
  - `Given the folder "<path>" under the test root contains the resource "<resource>"` and `Given the file "<path>" under the test root contains "<text>"`;
  - `Then the file "<path>" under the test root should (not) exist` / `…should contain "<text>"`;
  - `Then the app should have spawned "<path>" and quit`;
  - `When the app shuts down` (calls `ServiceShutdown()` on every registered binding that has one;
    the catalog's `the app is restarted` only swaps the data dir).
- `/create-tests` checks the catalog first and adds only what is missing.

Vitest locks the UI contract (`contracts/ui.md`):
- every state row of the section;
- toast show/Later/Restart now;
- the footer notice per state, with and without a workspace;
- the settings switch default and save;
- `bootstrap` calling `CheckOnStartup` once.

`node --test` locks R2:
- no tags → `x.y.0`;
- the next patch;
- a `VERSION` change resets to `.0`;
- other lines' tags are ignored;
- an invalid `VERSION` fails.

## Complexity Tracking

None. The Constitution Check has no violations.
