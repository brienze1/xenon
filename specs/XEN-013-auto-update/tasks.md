---

description: "Task list for XEN-013 — Auto-update Xenon from GitHub releases"
---

# Tasks: Auto-Update Xenon From GitHub Releases

**Input**: Design documents from `specs/XEN-013-auto-update/`

**Prerequisites**: plan.md, spec.md, research.md (design decisions R1–R14), data-model.md, contracts/ (bindings, events, release-assets, ui), quickstart.md

**Tests**: MANDATORY (constitution V, BDD-first). Each story starts with its Godog feature file in
`test/integration/features/updates/`, proven red before implementation; frontend work gets Vitest
tests; the release version script gets `node --test`.

**Organization**: grouped by user story (spec.md): US1 check (P1), US2 update (P1), US4 release
pipeline (P1), US3 automatic updates (P2).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: parallelizable (different files, no dependency on an unfinished task)
- **[Story]**: US1–US4 from spec.md

---

## Phase 1: Setup (Shared Infrastructure)

- [ ] T001 Create `VERSION` at the repo root containing `0.1` (single line, trailing newline)
- [ ] T002 [P] Add the release fixtures folder `test/integration/resources/updates/` with: `xenon-0.1.7-darwin-universal.tar.gz` (top entry `Xenon.app/Contents/MacOS/xenon`, a text file reading `xenon 0.1.7`, mode 0755, plus `Xenon.app/Contents/Info.plist`), `xenon-0.1.7-linux-amd64.tar.gz` (single entry `xenon`, text `xenon 0.1.7`, mode 0755), `xenon-0.1.7-windows-amd64.zip` (single entry `xenon.exe`, text `xenon 0.1.7`), `xenon-0.1.7-evil-amd64.tar.gz` (entry `../escape`), and `checksums.txt` with the real `sha256sum` lines for the three valid archives (the evil one is listed too, so unsafe-entry rejection is what fails it)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: version plumbing, properties, domain types, state store, event, binding skeleton —
needed by every story.

- [ ] T003 Add `var version = ""` (ldflag target), `Application.Packaged` (true when `version != ""`), and the `Application.Version` order env `XENON_VERSION` → `version` → `defaultVersion`, plus a new `Update` block `{ApiURL (env XENON_UPDATE_API_URL, default https://api.github.com/repos/brienze1/xenon), Platform (env XENON_UPDATE_PLATFORM, default darwin-universal on macOS else <GOOS>-<GOARCH>), InstallPath (env XENON_INSTALL_PATH, else derived from Application.Executable per design decision R8), Installable (Application.Packaged or XENON_INSTALL_PATH set)}` in `internal/infra/properties/properties.go`; extend `internal/infra/properties/properties_test.go` with default/override cases for each
- [ ] T004 [P] Add `-X github.com/brienze1/xenon/internal/infra/properties.version={{.XENON_VERSION}}` (only when `XENON_VERSION` is set, same `{{if}}` pattern as `XENON_GITHUB_CLIENT_ID`) to the production `BUILD_FLAGS` in `build/darwin/Taskfile.yml`, `build/windows/Taskfile.yml` and `build/linux/Taskfile.yml`
- [ ] T005 [P] Create `entity.Version` (`ParseVersion`, `NewerThan`, `String`; error code `UPD-01422`) in `internal/domain/entity/version.go` with unit tests in `internal/domain/entity/version_test.go` (v-prefix, suffix rejected, numeric ordering 0.1.10 > 0.1.9)
- [ ] T006 [P] Create enums `UpdateState` (`internal/domain/enums/update_state.go`, `internal/domain/enums/updatestate/value.go`) and `UpdateFailure` (`internal/domain/enums/update_failure.go`, `internal/domain/enums/updatefailure/value.go`) with the values in data-model.md, following `.claude/skills/enums.md`
- [ ] T007 [P] Create `entity.ReleaseAsset`, `entity.Release` (`PackageFor`, `Checksums`), `entity.UpdatePlatform` (`ArchiveExtension`, `IsBundle`) and `entity.UpdateStatus` (+ `LogValue`) in `internal/domain/entity/release_asset.go`, `release.go`, `update_platform.go`, `update_status.go`
- [ ] T008 Create the in-memory update state store (status, release, staged archive path, busy guard, once-per-run startup flag), keyed by `properties.Application.DataDir` per design decision R11, in `internal/integration/persistence/update_state.go`, its port use cases `internal/application/usecase/find_update_state.go` and `save_update_state.go`, the adapter `internal/integration/adapter/update_persistence.go`, and its factory under `internal/infra/dependency/integration/`
- [ ] T009 [P] Create the `update:status` publisher: event DTO + name in `internal/integration/publisher/dto/update_status.go`, publisher `internal/integration/publisher/update_status.go`, port `internal/application/usecase/publish_update_status.go`, adapter `internal/integration/adapter/update_status_publisher.go`, factory under `internal/infra/dependency/integration/`
- [ ] T010 [P] Create the response DTO `internal/integration/entrypoint/dto/update_status.go` (snake_case fields per contracts/events.md, `NewUpdateStatus(entity)`)
- [ ] T011 Create the `UpdateService` binding skeleton in `internal/infra/server/binding/update.go`, the controller interface `internal/integration/adapter/update_controller.go`, the controller `internal/integration/entrypoint/controller/update.go` with `GetStatus` (controller skill pattern: logger, panic recovery, error handler), the route `internal/infra/server/router/update_routes.go`, the router field/param/registration in `internal/infra/server/router/router.go`, and factories in `internal/infra/dependency/{binding,controller,router}/`
- [ ] T012 Register `UpdateService` in `startApp()` in `test/integration/steps/hooks_test.go`, then run `make bindings` so `frontend/bindings/…/updateservice` is generated

**Checkpoint**: `GetStatus` answers `idle` with the running version.

---

## Phase 3: User Story 1 — See whether Xenon is up to date (P1) 🎯 MVP

**Goal**: Check for updates in Settings shows current/latest versions and the right status.

**Independent Test**: with the mock API publishing 0.1.7 and `XENON_VERSION=0.1.4`, `UpdateService.Check` returns `available` / `latest_version 0.1.7`; the Settings section renders it.

### Tests for User Story 1 (write first, prove red)

- [ ] T013 [US1] Write `test/integration/features/updates/check_for_updates.feature` (`@all @xen_013 @updates @check_for_updates`) covering: idle status with the running version; available (0.1.4 vs 0.1.7, platform `linux-amd64`); up to date (equal, and running newer — no downgrade); no package (platform `linux-arm64`); check_failed/unreachable for 403, 429, 500, invalid JSON, invalid tag `v0.1.7-beta`, `prerelease: true`, a closed-port URL; busy guard (`Check` while a delayed check runs returns `checking`, the mock receives 1 request); `update:status` events for `checking` then the result; `OpenReleasePage` opens `release_url` through the browser mock and does nothing before a check; log lines for the result. Reuse the catalog steps (mock API, env var, binding call, response field, event, log); run `make bdd TAG=@check_for_updates` and record it red
- [ ] T014 [P] [US1] Write Vitest `frontend/src/features/settings/components/AboutUpdatesSection.test.tsx` for every state row in contracts/ui.md (status text, Version line, buttons enabled/disabled, What's new calling the service), plus `frontend/src/platform/mappers/update-status.test.ts` and `frontend/src/platform/services/updates.test.ts`; run and record red

### Implementation for User Story 1

- [ ] T015 [P] [US1] Create the GitHub releases response DTO (`tag_name`, `html_url`, `draft`, `prerelease`, `assets[]`) with `ToEntity()` in `internal/integration/webservice/dto/github_release.go`
- [ ] T016 [US1] Implement `FindLatestRelease` in `internal/integration/webservice/update.go` (`GET {Update.ApiURL}/releases/latest`, `Accept: application/vnd.github+json`, 15 s timeout, client from `httpClient.ClientFactory`; any failure → `UPD-02503` service-unavailable error mapped to reason `unreachable`), with port `internal/application/usecase/find_latest_release.go`, adapter `internal/integration/adapter/update_webservice.go`, factory `internal/infra/dependency/integration/update_webservice.go`
- [ ] T017 [P] [US1] Implement `OpenReleasePage` port `internal/application/usecase/open_release_page.go` on the update webservice using the existing `adapter.Browser` (as `github.go` does)
- [ ] T018 [US1] Implement the `CheckForUpdates` service (busy guard → `checking` + publish → find release → compare `entity.Version` → `up_to_date | available | no_package`, or `check_failed/unreachable`; save + publish; logs) in `internal/application/service/check_for_updates.go` with interface `internal/application/adapter/check_for_updates.go` and factory `internal/infra/dependency/application/check_for_updates.go`
- [ ] T019 [US1] Add `Check` and `OpenReleasePage` to the controller `internal/integration/entrypoint/controller/update.go` and the binding `internal/infra/server/binding/update.go`; run `make bindings`
- [ ] T020 [P] [US1] Add `UPDATE_STATUS` + `onUpdateStatus` to `frontend/src/platform/events.ts`, the mapper `frontend/src/platform/mappers/update-status.ts`, and `frontend/src/platform/services/updates.ts` (`getUpdateStatus`, `checkForUpdates`, `checkOnStartup`, `startUpdate`, `restartToUpdate`, `openReleasePage`)
- [ ] T021 [US1] Create `frontend/src/features/updates/model.ts` and `frontend/src/features/updates/store.ts` (Zustand: status, `load`, `check`, `update`, `restart`, `openReleasePage`, event subscription) with `frontend/src/features/updates/store.test.ts`
- [ ] T022 [US1] Create `frontend/src/features/settings/components/AboutUpdatesSection.tsx` (reads `features/updates/store.ts`; texts exactly per contracts/ui.md; styles from `SettingsSection.module.css`) and render it as the first section in `frontend/src/features/settings/components/SettingsTab.tsx`
- [ ] T023 [US1] Run `make bdd TAG=@check_for_updates` and the Vitest files from T014 — all green

**Checkpoint**: US1 works alone (check + display).

---

## Phase 4: User Story 2 — Update with one click (P1)

**Goal**: Update downloads, verifies, swaps the app and restarts.

**Independent Test**: with `XENON_INSTALL_PATH` at a temp `Xenon.app` and the fixture release, `UpdateService.Update` replaces the bundle content with the fixture and the process mock records the spawn of `<install>/Contents/MacOS/xenon` and a quit.

### Tests for User Story 2 (write first, prove red)

- [ ] T024 [US2] Create the process mock `test/integration/mock/app_process.go` (records `Spawn(path, env)` and `Quit()`; no real process) and inject it through a new `external.SetAppProcess` in `test/integration/steps/step_definitions_test.go`
- [ ] T025 [US2] Write `test/integration/features/updates/install_update.feature` (`@all @xen_013 @updates @install_update`) covering: darwin-universal bundle swap via `Update` (returns `ready`) then `Restart` (content of `Contents/MacOS/xenon` becomes `xenon 0.1.7`, spawn path `<install>/Contents/MacOS/xenon` with `XENON_RESTART_AFTER_PID`, quit); linux-amd64 and windows-amd64 single-file swaps; `downloading` progress events ending at `received == total`; bad checksum → `failed/unverified`, install untouched, `<DataDir>/updates/` empty; missing checksums line → `unverified`; download 500 / closed port → `failed/download_failed`; unsafe archive entry (platform `evil-amd64`, fixture `xenon-0.1.7-evil-amd64.tar.gz`) → `failed/install_failed`, install untouched; unwritable (`XENON_INSTALL_PATH=<root>/blocker.txt/Xenon.app` with `blocker.txt` a file) → `failed/not_writable` with `folder`; not packaged (no `XENON_INSTALL_PATH`, no ldflag version) → `failed/not_packaged`; only the release-listed URLs requested (mock request counts); leftovers `.Xenon.app.xenon-old`, `.Xenon.app.xenon-update/` and `<DataDir>/updates/*` removed by `CheckOnStartup`; log lines. Add only the missing generic steps listed in plan.md (test-root files/folders, spawned-and-quit) to `test/integration/steps/`. Run `make bdd TAG=@install_update` and record it red

### Implementation for User Story 2

- [ ] T026 [P] [US2] Implement `DownloadReleaseAsset` (stream to `<DataDir>/updates/<asset>.part`, SHA-256 while writing, progress callback ≤ every 250 ms + final, redirects followed, 10 min timeout, rename on success, delete on failure) and `FindReleaseChecksums` (parse `sha256sum` lines) in `internal/integration/webservice/update.go`, ports `internal/application/usecase/download_release_asset.go` and `find_release_checksums.go`
- [ ] T027 [P] [US2] Implement the installer in `internal/integration/persistence/update_install.go`: writability probe first, safe extraction (tar.gz with modes / zip; reject absolute, `..`, outside-pointing links; exactly one top-level entry per design decision R8), two-rename swap with rollback, side-file removal, `CleanLeftovers`; ports `internal/application/usecase/install_update.go` and `clean_update_leftovers.go`; unit tests in `internal/integration/persistence/update_install_test.go` for the extraction safety rules and for rollback (the second rename fails through an injected rename function → the previous app is back at its path and the error is `install_failed`), since a failing rename cannot be provoked from BDD
- [ ] T028 [P] [US2] Implement the process adapter: interface `internal/integration/adapter/app_process.go`; infra implementation `internal/infra/adapter/app_process.go` (detached `exec.Command` start with env, `Setpgid` on Unix via build-tagged files, `application.Get().Quit()`); settable factory `internal/infra/dependency/external/app_process.go`; webservice `internal/integration/webservice/app_restart.go` with port `internal/application/usecase/restart_app.go`
- [ ] T029 [US2] Add the restart handshake to `internal/infra/run.go`: when `XENON_RESTART_AFTER_PID` is set, wait (poll every 100 ms, max 15 s) for that pid to exit before `Start()`; unit test the wait helper with a fake process checker
- [ ] T030 [US2] Implement the `DownloadUpdate` service (state `downloading`, progress publish, checksum compare, `ready` or `failed/download_failed|unverified`) in `internal/application/service/download_update.go` and the `ApplyUpdate` service (`not_packaged` guard: the controller reads `properties.Update.InstallPath`/`Installable` — as the settings controller already reads `properties` — and passes them to the service as an `entity.UpdateTarget{Path, Platform, Installable}` added to `internal/domain/entity/update_platform.go`; then `installing`, install, restart; failures → `failed/not_writable|install_failed`) in `internal/application/service/apply_update.go`, with interfaces in `internal/application/adapter/` and factories in `internal/infra/dependency/application/`
- [ ] T031 [US2] Add `Update` (download + verify → `ready`, `automatic: false`; no restart) and `Restart` (from `ready`: apply = install + restart) to the controller and binding (`internal/integration/entrypoint/controller/update.go`, `internal/infra/server/binding/update.go`) per contracts/bindings.md; run `make bindings`
- [ ] T032 [US2] In `frontend/src/features/updates/store.ts`, make `restart` await `useTabsStore.getState().flushPersist()` before `restartToUpdate()`, and make `update` call `restart` when `startUpdate()` resolves with `ready`; extend `frontend/src/features/updates/store.test.ts` (flush happens before restart; no restart on failure). Wire Update / Restart now / failure rows into `frontend/src/features/settings/components/AboutUpdatesSection.tsx` (progress bar from `received/total`) and extend its Vitest file
- [ ] T033 [US2] Run `make bdd TAG=@install_update` and the section's Vitest — all green

**Checkpoint**: US1 + US2 give a full manual update.

---

## Phase 5: User Story 4 — Every merge publishes a release (P1)

**Goal**: CI publishes a versioned release with packages and checksums on each merge to `main`.

**Independent Test**: `node --test tools/release/` passes; after merge, `gh release view v0.1.0` lists the four assets.

- [ ] T034 [P] [US4] Write `tools/release/next-version.test.mjs` (`node --test`): no tags → `0.1.0`; tags `v0.1.0,v0.1.1` → `0.1.2`; `VERSION` `0.2` with `v0.1.x` tags → `0.2.0`; tags of other lines ignored; non-`v` tags ignored; `VERSION` `abc` → exit code 1. Run and record red
- [ ] T035 [US4] Implement `tools/release/next-version.mjs` (reads `VERSION`, takes tags from stdin one per line, prints the version) with a `tools/release/package.json` (`"type": "module"`, no dependencies)
- [ ] T036 [US4] Add the `release-check` target (`node --test tools/release/`) to `Makefile`, include it in `.PHONY`, and call it from `validate`
- [ ] T037 [US4] Create `.github/workflows/release.yml` per design decision R3 and contracts/release-assets.md: jobs `version` → `build` (matrix, `fail-fast: true`, `XENON_VERSION` passed to the Taskfile builds; macOS universal package, `plutil -replace CFBundleShortVersionString/CFBundleVersion`, ad-hoc re-sign, `tar -czf` with top entry `Xenon.app`; Windows `xenon.exe` zipped with PowerShell `Compress-Archive` (built into the runner); Linux `xenon` tar.gz) → `publish` (`sha256sum` → `checksums.txt`, notes with the macOS first-install paragraph, `gh release create v<v> --target $GITHUB_SHA`), `permissions: contents: write`, `concurrency: release` without cancel
- [ ] T038 [US4] Run `make release-check` (green) and lint the workflow with `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.7 .github/workflows/release.yml` (zero findings); the real publish is verified after merge per quickstart.md §2

**Checkpoint**: releases flow from `main`.

---

## Phase 6: User Story 3 — Updates arrive by themselves (P2)

**Goal**: startup check; auto download + toast + footer; setting default on; install on quit.

**Independent Test**: `CheckOnStartup` with the setting on emits `ready` + `automatic: true`; off → `available`, nothing downloaded; binding shutdown with `ready` installs without spawning.

### Tests for User Story 3 (write first, prove red)

- [ ] T039 [US3] Write `test/integration/features/updates/auto_update.feature` (`@all @xen_013 @updates @auto_update`): `SettingsService.Get` returns `auto_update: true` with no stored value and with an `app.json` written before this feature; Save `auto_update: false` persists `settings.autoUpdate false` and survives `the app is restarted`; on + newer → `CheckOnStartup` returns `downloading`, then a `ready` event with `automatic` `true` within 5000 ms and the staged archive in `<DataDir>/updates/`; off + newer → `available`, no download request received; startup check failure → `check_failed`, no error; automatic download failure (bad checksum) → `failed/unverified` event with `automatic` `true` and no `ready` event (FR-018b); second `CheckOnStartup` returns current status without a new request; `Restart` from `ready` installs + spawns + quits; app shutdown with `ready` installs and does not spawn. The catalog has no shutdown step (`the app is restarted` only swaps the data dir), so add the generic `When the app shuts down` step in `test/integration/steps/given_app_steps_test.go`, calling `ServiceShutdown()` on every registered binding that has one. Run `make bdd TAG=@auto_update` and record it red
- [ ] T040 [P] [US3] Write Vitest: `frontend/src/shared/ui/toast/Toast.test.tsx`, `frontend/src/features/updates/components/UpdateReadyToast.test.tsx` (shows on `ready`+`automatic`, never on `failed`, no auto-dismiss, Restart now → `restartToUpdate`, Later/× hides, once per run), `frontend/src/features/updates/components/FooterUpdateNotice.test.tsx` (ready; available with the setting on and off; none), settings switch default + save in `frontend/src/features/settings/components/SettingsTab.test.tsx`, `frontend/src/app/bootstrap.test.ts` (`checkOnStartup` called once); record red

### Implementation for User Story 3

- [ ] T041 [P] [US3] Add `AutoUpdate` to `internal/domain/entity/app_settings.go` (default `true`, `LogValue`) and `app_settings_patch.go` (`ApplyTo`, `LogValue`); `AutoUpdate *bool` json `autoUpdate,omitempty` (nil → true) in `internal/integration/persistence/model/app_settings.go`; `auto_update` in `internal/integration/entrypoint/dto/app_settings.go` (`NewAppSettings`, `ToEntity`, `LogValue`); extend the entity/model unit tests
- [ ] T042 [US3] Implement the `StartupUpdate` service (once-per-run guard, best-effort `CleanLeftovers` — a locked file is logged at Warn and never fails the check, check via `CheckForUpdates`, read `AppSettings.AutoUpdate`, start `DownloadUpdate` in a goroutine with `Automatic=true` when on) in `internal/application/service/startup_update.go` + interface + factory
- [ ] T043 [US3] Add `CheckOnStartup` to the controller/binding and `ServiceShutdown()` on `internal/infra/server/binding/update.go` calling a controller `InstallOnQuit(ctx)` (installs a `ready` update without restart); run `make bindings`
- [ ] T044 [P] [US3] Create `frontend/src/shared/ui/toast/Toast.tsx` + `Toast.module.css` (bottom-left fixed, `role="status"`, `aria-live="polite"`, close button, PRD palette tokens only)
- [ ] T045 [US3] Create `frontend/src/features/updates/components/UpdateReadyToast.tsx` and `FooterUpdateNotice.tsx` (texts per contracts/ui.md; footer `available` opens `useTabsStore().openSettingsTab`)
- [ ] T046 [US3] Mount `UpdateReadyToast` in `frontend/src/app/App.tsx` and render `FooterUpdateNotice` in the footer status slot whether or not a workspace is open; call `checkOnStartup()` once after the shell renders from `frontend/src/app/bootstrap.ts`
- [ ] T047 [US3] Add `autoUpdate` to `frontend/src/platform/mappers/settings.ts`, `frontend/src/features/settings/model.ts`, `frontend/src/features/settings/lib/settings-form.ts`, and the **Update automatically** switch (`role="switch"`, form draft + Save) to `AboutUpdatesSection.tsx`
- [ ] T048 [US3] Run `make bdd TAG=@auto_update` and the Vitest files from T040 — all green

**Checkpoint**: all stories work.

---

## Phase 7: Polish & Cross-Cutting

- [ ] T049 Run `make sort-imports`, then the full `make bdd` (all tags) and `make validate`; fix until green with every pre-existing scenario still passing
- [ ] T050 [P] Run `/refactor-diff` until it reports clean
- [ ] T051 [P] Add the update feature to the `docs/prd/xenon-mvp-prd.md` decisions list (constitution VI: the PRD must cover it) as a built user decision (release per merge, VERSION file, ad-hoc signing, all-OS self-install, ask-before-restart, default on, check-and-notify when off)
- [ ] T052 Walk through quickstart.md §1 locally; §2–§3 after merge (release publish and the macOS manual run)

---

## Dependencies & Execution Order

- Setup (T001–T002) → Foundational (T003–T012) → stories.
- US1 (T013–T023) first: US2 and US3 build on `CheckForUpdates` and the section.
- US2 (T024–T033) after US1.
- US4 (T034–T038) depends only on Foundational T003–T004 and can run in parallel with US1/US2.
- US3 (T039–T048) after US1 and US2 (uses `DownloadUpdate`, `ApplyUpdate`).
- Polish last.

Within a story: feature file red → domain → use case ports → integration (webservice/persistence/publisher) → service → controller/binding → frontend → green.

## Parallel Opportunities

- Foundational: T004, T005, T006, T007, T009, T010 in parallel after T003.
- US1: T014 alongside T013; T015, T017, T020 in parallel.
- US2: T026, T027, T028 in parallel after T025.
- US4 entirely parallel to US1/US2 (different files: `tools/`, `.github/`, `Makefile`).
- US3: T040, T041, T044 in parallel after T039.

## Implementation Strategy

MVP = Setup + Foundational + US1 (version visible, check works). Then US2 (manual update), US4
(releases exist so the updater has something real to fetch), US3 (automation). Each checkpoint
is independently demoable; the whole set ships in one PR per the task flow.
