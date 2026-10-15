# Research: Auto-Update Xenon From GitHub Releases

All facts below were verified against the code at `25acd4d` (origin/main, 2026-10-15). Each entry:
Decision · Rationale · Alternatives considered.

## R1 — Version: one source of truth, embedded at build time

- **Decision**: A new `VERSION` file at the repo root holds `major.minor` (`0.1`). The release
  workflow computes the full version (`R2`) and passes it to every build as
  `XENON_VERSION=<x.y.z>`; the three OS Taskfiles add
  `-X github.com/brienze1/xenon/internal/infra/properties.version={{.XENON_VERSION}}` to their
  production `-ldflags`, next to the existing `gitHubClientID` flag (`build/darwin/Taskfile.yml:47`,
  `build/windows/Taskfile.yml:60`, `build/linux/Taskfile.yml:62`). On macOS the workflow also
  stamps `CFBundleShortVersionString` and `CFBundleVersion` in the packaged bundle's `Info.plist`
  with `plutil -replace` before signing. `properties.Application.Version` becomes: env
  `XENON_VERSION` → ldflag `version` → `defaultVersion` (`0.1.0`, unchanged so existing export
  and MCP scenarios keep their expectations). A new `properties.Application.Packaged` is true
  only when the ldflag `version` is non-empty.
- **Rationale**: Mirrors the existing `gitHubClientID` ldflag pattern. The tag is the release
  version, the binary reports the same string, and nothing has to be committed back to `main` per
  release (no bot commits, no loop with the `push` trigger).
- **Alternatives**: committing the bumped version into `build/config.yml` on every merge (bot
  commit per merge, re-triggers CI); reading `VERSION` at runtime (the app has no repo at
  runtime). `build/config.yml` `info.version` stays as the Wails asset generator's input only;
  the plan does not use it for releases (documented in quickstart).

## R2 — Computing the next version

- **Decision**: `tools/release/next-version.mjs` (Node, like the other `tools/` scripts) reads
  `VERSION` (`major.minor`) and the list of `v*` tags, and prints `<major>.<minor>.<patch>` where
  `patch` = highest existing patch for that `major.minor` + 1, or `0` if none. A `VERSION` that
  is not `N.N` fails the job. Unit-tested with `node --test` (`tools/release/next-version.test.mjs`),
  run by a new `make release-check` target that `make validate` calls.
- **Rationale**: Testable, portable on all CI runners, no new Go entrypoint for a CI concern.
- **Alternatives**: a shell one-liner in the workflow (untested); a Go command in `cmd/`
  (`cmd/` is for app entrypoints).

## R3 — Release workflow

- **Decision**: `.github/workflows/release.yml`, `on: push: branches: [main]`,
  `permissions: contents: write`, `concurrency: { group: release, cancel-in-progress: false }`
  (merges release in order). Jobs:
  1. `version` (ubuntu-24.04): `fetch-depth: 0` + tags, run `next-version.mjs`, output `version`.
  2. `build` matrix (`macos-latest`, `windows-latest`, `ubuntu-24.04`), `needs: version`,
     `fail-fast: true`: `make setup`, `make bindings`, then
     - macOS: `wails3 task darwin:package:universal XENON_VERSION=…` → `bin/xenon.app`, stamp
       `Info.plist`, re-run the existing ad-hoc `codesign`, archive as
       `xenon-<v>-darwin-universal.tar.gz` (top entry `Xenon.app`).
     - Windows: `wails3 task windows:build XENON_VERSION=…` → `bin/xenon.exe`, archive as
       `xenon-<v>-windows-amd64.zip` (single entry `xenon.exe`).
     - Linux: `wails3 task linux:build XENON_VERSION=…` → `bin/xenon`, archive as
       `xenon-<v>-linux-amd64.tar.gz` (single entry `xenon`, mode 0755).
     Each uploads its archive as a workflow artifact.
  3. `publish` (ubuntu-24.04, `needs: [version, build]`): download artifacts, write
     `checksums.txt` with `sha256sum` (`<hex>  <file>` per line), then
     `gh release create v<v> --target $GITHUB_SHA --title "Xenon <v>" --notes-file notes.md`
     with all archives and `checksums.txt`. `notes.md` = GitHub-generated notes
     (`--generate-notes` content appended) plus the fixed "First install on macOS" paragraph
     (right-click → Open, or `xattr -dr com.apple.quarantine /path/to/Xenon.app`).
  A failed build stops `publish` (FR-004). No release is a draft or prerelease.
- **Rationale**: `needs` + `fail-fast` give all-or-nothing publishing; `gh` is preinstalled on
  runners; universal macOS covers Apple silicon and Intel in one package (FR-003).
- **Alternatives**: GoReleaser (new tool, does not build Wails bundles); NSIS installer / AppImage
  / deb as update payloads (need elevation or package managers; the user chose replacing the
  executable).

## R4 — Asking GitHub for the latest release

- **Decision**: `GET {UpdateApiURL}/releases/latest` with `Accept: application/vnd.github+json`,
  no auth. `UpdateApiURL` = env `XENON_UPDATE_API_URL` or
  `https://api.github.com/repos/brienze1/xenon` (new `properties.Update` block, read on every
  call like `Git.GitHubApiURL`, so BDD points it at the mock API). The endpoint already excludes
  drafts and prereleases. The answer's `tag_name` (leading `v` stripped), `html_url` and
  `assets[] {name, browser_download_url, size}` are mapped into a `Release`. 403/429 (rate limit),
  other non-200, network errors, unparsable JSON or an invalid version → `check_failed`.
  Timeout 15 s.
- **Rationale**: Public endpoint, 60 req/h unauthenticated — ample (spec Assumptions).
- **Alternatives**: listing `/releases` and filtering (more data, same result); Atom feed (no
  asset list).

## R5 — Version comparison

- **Decision**: A domain value `entity.Version{Major, Minor, Patch}` parsed from `N.N.N`
  (optional leading `v`; any pre-release/build suffix makes it invalid, since releases never
  carry one). `NewerThan` compares numerically. No dependency added.
- **Rationale**: Releases are always plain `x.y.z` (R2); a 20-line value object is simpler than a
  semver library (constitution: new deps need justification).
- **Alternatives**: `golang.org/x/mod/semver` (requires the `v` prefix, string API; not needed).

## R6 — Choosing the package for this machine

- **Decision**: `properties.Update.Platform` = env `XENON_UPDATE_PLATFORM` or
  `darwin-universal` on macOS, `<GOOS>-<GOARCH>` elsewhere. The wanted asset is
  `xenon-<version>-<platform>.<ext>` where `ext` is `zip` for `windows-*` and `tar.gz` otherwise.
  No matching asset → `no_package`. The platform env override exists so BDD runs the same
  scenarios identically on all three CI runners.
- **Rationale**: Deterministic names from R3; Windows/Linux arm64 simply find no asset.

## R7 — Download and verification

- **Decision**: Download only the `browser_download_url` the release listed (FR-022) — and the
  `checksums.txt` asset of the same release — with redirects followed (GitHub serves assets via a
  redirect to its object storage), 10-minute timeout, streamed to
  `<DataDir>/updates/<asset>.part` while hashing with SHA-256; progress (`received`, `total` from
  `size`/Content-Length) is published at most every 250 ms and at 100%. When the hash equals the
  `checksums.txt` line for that file name, the file is renamed to `<DataDir>/updates/<asset>` and
  the state becomes `ready`. Missing checksums file, no line for the asset, or a mismatch →
  delete the file, state `failed` with reason `unverified`.
- **Rationale**: Streaming hash avoids reading the file twice; `.part` never looks complete.
- **Alternatives**: signature verification (minisign/cosign) — rejected for this version: the
  user accepted the same-release trust model (spec Clarifications).

## R8 — Installing (replacing the app)

- **Decision**: The install target is `properties.Update.InstallPath` = env
  `XENON_INSTALL_PATH`, else derived from `Application.Executable`: on macOS the enclosing
  `*.app` directory when the executable is `<X>.app/Contents/MacOS/<bin>`; on Windows/Linux the
  executable itself. Installing requires `properties.Update.Installable` = `Application.Packaged` or an explicit
  `XENON_INSTALL_PATH` (BDD); otherwise state `failed` reason `not_packaged` (FR-016).
  Steps, all inside the target's parent folder so every rename is on one volume:
  1. Writability probe (always the first step, before the target is touched): create and
     delete `.<name>.xenon-update-probe` in the parent folder; any failure → `failed` reason
     `not_writable` with the folder (FR-015). No elevation.
  2. Extract the archive into `<parent>/.<name>.xenon-update/` (tar.gz keeps modes; zip for
     Windows). Reject unsafe entries (absolute paths, `..`, links pointing outside) and require
     exactly one top-level entry: a `*.app` directory for bundle platforms, or a regular file
     `xenon` / `xenon.exe` otherwise; anything else → `failed` reason `install_failed`. The
     entry is installed under the target's own name (an install named `Xenon.app` stays
     `Xenon.app`).
  3. Rename `<target>` → `<parent>/.<name>.xenon-old`, then the extracted entry → `<target>`.
     If the second rename fails, rename the old one back (FR-013); report `failed` reason
     `install_failed`.
  4. Remove `.<name>.xenon-old` (on Windows a running `.exe` cannot be deleted — it is left and
     removed at the next start, see step 5) and the extraction folder.
  5. At every start (`CheckOnStartup`), remove leftover `.<name>.xenon-old`,
     `.<name>.xenon-update/` and everything in `<DataDir>/updates/` (FR-021). Removal is best
     effort: a file still locked (Windows, an old `xenon mcp` bridge) is logged at Warn and
     retried at the next start; it never fails the check.
  On macOS, the ad-hoc signature travels inside the bundle; Go's HTTP client does not set
  `com.apple.quarantine`, so the new bundle opens without a Gatekeeper prompt.
- **Rationale**: Renaming a running bundle/binary is allowed on macOS and Linux and for `.exe`
  files on Windows (rename, not overwrite), giving an atomic swap with rollback.
- **Alternatives**: overwrite in place (not atomic, fails on Windows); helper updater process
  (extra binary to ship).

## R9 — Restarting

- **Decision**: After a successful install, an infra adapter starts the new executable
  (`<bundle>/Contents/MacOS/<bin>` on macOS, the target on Windows/Linux) detached, with env
  `XENON_RESTART_AFTER_PID=<current pid>`, then quits the app through Wails
  (`application.Get().Quit()`). In `internal/infra/run.go`, before `Start()`, a process started
  with that variable waits (polling every 100 ms, at most 15 s) for that pid to exit, so the
  single-instance lock (`UniqueID: dev.brienze.xenon`) is free. Quitting runs the normal shutdown,
  so UI state (open tabs and drafts) is already flushed.
- **Install on quit** (Later, FR-020): the update binding's `ServiceShutdown` installs a `ready`
  update without restarting.
- **Rationale**: No helper binary; works the same on all three OSes. BDD fakes the adapter
  (`mock/app_process.go`) and asserts the spawn path and the quit.
- **Alternatives**: `open -n` on macOS only (not portable); `syscall.Exec` (no Windows support,
  and replaces the process before shutdown runs).

## R10 — Startup check trigger

- **Decision**: The frontend's `app/bootstrap.ts` calls `UpdateService.CheckOnStartup()` once,
  after the shell has rendered (window usable — FR-018). Go does all the work in that call:
  cleanup (R8 step 5), check, then — with Update automatically on and an update offered —
  download and verify in a background goroutine; the call returns as soon as the check ends. The
  UI is never blocked because Wails calls are asynchronous.
- **Rationale**: Guarantees "after the window is usable" without timing guesses, and BDD calls the
  same binding. A Wails `ServiceStartup` hook runs before the window shows and cannot be driven by
  BDD.

## R11 — Update state and events

- **Decision**: One in-memory update state (single instance, guarded by a mutex) in
  `internal/integration/persistence/update_state.go`; every change is published as the
  `update:status` event carrying the full status DTO (`contracts/events.md`). Concurrency guard:
  a check/download/install request while the state is `checking`, `downloading` or `installing`
  returns the current status unchanged (FR-010).
  The store keys its state by `properties.Application.DataDir`: production has one data dir per
  process, while the BDD harness gives every scenario (and every `the app is restarted`) a new
  data dir, so each one starts from a fresh `idle` state and a fresh once-per-run startup guard
  even though DI singletons live for the whole test process.
- **Rationale**: Matches the in-memory session stores already in persistence (e.g.
  `sync_sessions.go`); one event with the full snapshot keeps the React store trivial.

## R12 — Auto-update setting storage

- **Decision**: `entity.AppSettings.AutoUpdate bool` (default `true` in `DefaultAppSettings`);
  model `AppSettings.AutoUpdate *bool` json `autoUpdate,omitempty` — `nil` reads as `true`, so
  machines that stored settings before this feature get "on" (spec Clarifications). Patch and DTO
  gain `AutoUpdate *bool` / `auto_update`.
- **Rationale**: Same pattern as the other settings in `app.json`; pointer keeps "never stored"
  distinguishable.

## R13 — Frontend placement

- **Decision**: A new `features/updates/` (store, model, `UpdateReadyToast`, `FooterUpdateNotice`)
  and a new domain-free `shared/ui/toast/Toast.tsx` (bottom-left, `role="status"`). The About &
  Updates section lives in `features/settings/components/AboutUpdatesSection.tsx` (a feature may
  read another feature's `store.ts` only), reading `features/updates/store.ts`. `app/App.tsx`
  mounts the toast and puts `FooterUpdateNotice` in the footer status slot **whether or not a
  workspace is open**. The auto-update switch is part of the existing settings form (draft +
  Save), like the AI access switch.
- **Rationale**: Follows the frontend-architecture boundary table.

## R14 — Logging

- **Decision**: Each service logs with `pkg/logger` (Started/Finished, plus Info lines for
  "update available <v>", "downloaded <file> <bytes>", "verified", "installed <v> at <path>",
  and Warn with the reason on failures). Only versions, file names, sizes, paths and statuses are
  logged (FR-023). BDD asserts with the existing log steps.
