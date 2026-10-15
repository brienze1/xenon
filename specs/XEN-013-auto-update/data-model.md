# Data Model: Auto-Update Xenon From GitHub Releases

## Domain (`internal/domain/`)

### `entity.Version` (value object) — `entity/version.go`

| Field | Type | Rule |
|---|---|---|
| Major, Minor, Patch | int | ≥ 0 |

- `ParseVersion(text) (Version, error)` — accepts `N.N.N` with an optional leading `v`; anything
  else (pre-release suffix, missing part, non-digits) is invalid (`UPD-01422`).
- `NewerThan(other) bool` — numeric major → minor → patch.
- `String()` — `N.N.N`.

### `entity.ReleaseAsset` — `entity/release_asset.go`

| Field | Type | Notes |
|---|---|---|
| Name | string | e.g. `xenon-0.1.7-darwin-universal.tar.gz` |
| URL | string | the release's `browser_download_url` (only URL ever downloaded) |
| Size | int64 | bytes, used for progress when the server sends no length |

### `entity.Release` — `entity/release.go`

| Field | Type | Notes |
|---|---|---|
| Version | Version | from `tag_name` |
| PageURL | string | release page (`What's new`) |
| Assets | []ReleaseAsset | |

- `PackageFor(platform UpdatePlatform) (ReleaseAsset, bool)` — the asset named
  `xenon-<version>-<platform>.<ext>`.
- `Checksums() (ReleaseAsset, bool)` — the asset named `checksums.txt`.

### `entity.UpdatePlatform` — `entity/update_platform.go`

String value `darwin-universal | windows-amd64 | linux-amd64 | <goos>-<goarch>`;
`ArchiveExtension()` → `zip` when it starts with `windows-`, else `tar.gz`;
`IsBundle()` → true for `darwin-*` (install target is a `.app` directory).

### `entity.UpdateTarget` — `entity/update_platform.go`

| Field | Type | Notes |
|---|---|---|
| Path | string | install target (`.app` dir or executable), from `properties.Update.InstallPath` |
| Platform | UpdatePlatform | from `properties.Update.Platform` |
| Installable | bool | `properties.Update.Installable` (packaged build, or `XENON_INSTALL_PATH` set) |

Built by the update controller and passed to the services, so the application layer never reads
`properties`.

### `entity.UpdateStatus` — `entity/update_status.go`

| Field | Type | Notes |
|---|---|---|
| State | `enums.UpdateState` | see transitions |
| Reason | `enums.UpdateFailure` | set only when State is `check_failed` or `failed` |
| CurrentVersion | string | running version |
| LatestVersion | string | empty until a check succeeds |
| ReleaseURL | string | latest release page |
| Received, Total | int64 | download progress (bytes) |
| Folder | string | the unwritable folder, only with reason `not_writable` |
| Automatic | bool | true when the download was started by the startup check (drives the toast) |

### `enums.UpdateState` — `enums/update_state.go` + `enums/updatestate/value.go`

`idle · checking · up_to_date · available · no_package · check_failed · downloading · ready ·
installing · failed`

### `enums.UpdateFailure` — `enums/update_failure.go` + `enums/updatefailure/value.go`

`unreachable` (network/HTTP/rate limit/bad answer — the "Couldn't check for updates" case) ·
`download_failed` · `unverified` · `not_packaged` · `not_writable` · `install_failed`

### State transitions

```
idle ──check──► checking ──► up_to_date | available | no_package | check_failed
available ──update (manual)──► downloading ──► ready ──(UI flushes tabs, calls Restart)──► installing ──► (restart; process ends)
available ──startup, auto on──► downloading ──► ready
ready ──restart now / footer restart──► installing ──► (restart)
ready ──app quits normally──► installing ──► (quit)
downloading ──error──► failed(download_failed | unverified)
installing ──error──► failed(not_packaged | not_writable | install_failed)
failed | check_failed | up_to_date | no_package ──check──► checking
```

A request arriving while the state is `checking`, `downloading` or `installing` returns the
current status unchanged.

### `entity.AppSettings` (existing) — gains

| Field | Type | Default |
|---|---|---|
| AutoUpdate | bool | `true` |

`AppSettingsPatch.AutoUpdate *bool`; `ApplyTo` copies it when non-nil; `LogValue` includes
`auto_update`.

## Persistence (`internal/integration/persistence/`)

- **`model.AppSettings`** gains `AutoUpdate *bool \`json:"autoUpdate,omitempty"\``; `ToEntity`
  reads `nil` as `true`. Stored in `<DataDir>/app.json` under `settings`.
- **In-memory update state** (`update_state.go`): holds the one `UpdateStatus`, the found
  `Release`, the staged archive path, and the "install on quit" flag. Not written to disk.
- **Staging folder**: `<DataDir>/updates/` — `<asset>.part` while downloading, `<asset>` once
  verified. Emptied at every start.
- **Install side files** (in the install target's parent folder, removed after install or at the
  next start): `.<name>.xenon-update/` (extraction), `.<name>.xenon-old` (previous app),
  `.<name>.xenon-update-probe` (writability probe).

## Release files (published by CI)

| File | Content |
|---|---|
| `xenon-<v>-darwin-universal.tar.gz` | one entry: `Xenon.app/` (ad-hoc signed, Info.plist stamped with `<v>`) |
| `xenon-<v>-windows-amd64.zip` | one entry: `xenon.exe` |
| `xenon-<v>-linux-amd64.tar.gz` | one entry: `xenon` (mode 0755) |
| `checksums.txt` | `<sha256 hex>  <file name>` per archive, `sha256sum` format |
