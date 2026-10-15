# Contract: GitHub release (published by CI, consumed by the updater)

## Tag and version

- Tag `v<major>.<minor>.<patch>`; `major.minor` from `VERSION`; patch = highest released patch in
  that line + 1, or 0 (`tools/release/next-version.mjs`).
- Release title `Xenon <major>.<minor>.<patch>`; never draft, never prerelease; target = merged
  commit.

## Assets (exact names)

| Name | Archive | Single top-level entry |
|---|---|---|
| `xenon-<v>-darwin-universal.tar.gz` | gzip tar | `Xenon.app/` directory |
| `xenon-<v>-windows-amd64.zip` | zip | `xenon.exe` |
| `xenon-<v>-linux-amd64.tar.gz` | gzip tar | `xenon` (0755) |
| `checksums.txt` | text | lines `<64 lowercase hex>  <asset name>` |

## Release notes

GitHub-generated change list, followed by:

> **First install on macOS.** Xenon is not notarized. After downloading it in a browser, open it
> with right-click → Open the first time, or run
> `xattr -dr com.apple.quarantine /path/to/Xenon.app`. Updates installed by Xenon itself do not
> need this.

## What the updater reads (`GET {XENON_UPDATE_API_URL}/releases/latest`)

```json
{
  "tag_name": "v0.1.7",
  "html_url": "https://github.com/brienze1/xenon/releases/tag/v0.1.7",
  "draft": false,
  "prerelease": false,
  "assets": [
    { "name": "xenon-0.1.7-linux-amd64.tar.gz", "size": 15728640,
      "browser_download_url": "https://github.com/brienze1/xenon/releases/download/v0.1.7/xenon-0.1.7-linux-amd64.tar.gz" },
    { "name": "checksums.txt", "size": 300,
      "browser_download_url": "https://github.com/brienze1/xenon/releases/download/v0.1.7/checksums.txt" }
  ]
}
```

A `draft: true` or `prerelease: true` answer (only possible from a misbehaving server) is treated
as `check_failed`/`unreachable`.
