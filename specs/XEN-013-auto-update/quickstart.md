# Quickstart: validating auto-update

## 1. Automated suites (every branch)

```bash
make bindings
make bdd TAG=@xen_013          # check / install / auto-update scenarios (GitHub + downloads faked)
cd frontend && npm test -- updates settings toast   # Vitest: section, toast, footer, store, mappers
node --test tools/release/next-version.test.mjs  # next-version rules
make validate                  # full gate (includes release-check)
```

Expected: everything green. BDD drives the real controllers, services and persistence. Only these
are faked: the mock API (`XENON_UPDATE_API_URL`), the install folder (`XENON_INSTALL_PATH`), the
platform (`XENON_UPDATE_PLATFORM`) and the process spawn/quit. See [plan.md](plan.md) →
Requirement → BDD lock.

## 2. Release pipeline (after the PR merges)

1. Merge to `main`. The **Release** workflow runs (`.github/workflows/release.yml`).
2. Expected:
   - Release `v0.1.0` (the first one, since `VERSION` = `0.1` and there are no tags yet).
   - Assets: `xenon-0.1.0-darwin-universal.tar.gz`, `xenon-0.1.0-windows-amd64.zip`,
     `xenon-0.1.0-linux-amd64.tar.gz` and `checksums.txt`.
   - Release notes ending with the "First install on macOS" paragraph.
   - `gh release view v0.1.0 --repo brienze1/xenon` lists the assets.
   - `sha256sum -c checksums.txt` passes on the downloaded files.
3. The next merge publishes `v0.1.1`. A PR that changes `VERSION` to `0.2` publishes `v0.2.0`.

`build/config.yml`'s `info.version` only feeds the Wails asset generator. Release versions come
from `VERSION` plus the tags, and are injected at build time.

## 3. Manual check on macOS (the acceptance run)

1. Download `xenon-<older>-darwin-universal.tar.gz` from an earlier release. Extract it to
   `~/Applications/Xenon.app`, then run `xattr -dr com.apple.quarantine ~/Applications/Xenon.app`
   (first install only).
2. Open Xenon, then Settings. The first section shows `Version <older>`.
3. Click **Check for updates**. Expected: `Update available · Latest version <newest>`, and an
   **Update** button.
4. Click **Update**. Expected: the download percentage, then `Installing…`. Xenon quits and
   reopens on `<newest>`, with tabs and drafts restored and no Gatekeeper prompt.
5. Auto mode: install `<older>` again with **Update automatically** on (the default), then open
   it. Expected:
   - the window is usable immediately;
   - a bottom-left toast `Xenon <newest> is ready to install.` appears with Restart now and Later.
   Then:
   - **Later**: the footer shows `Update ready — Restart`. Quit with Cmd+Q and reopen; Xenon is
     now on `<newest>`.
6. Offline: turn off Wi-Fi and click Check for updates. Expected: `Couldn't check for updates`
   and nothing else changes.
7. Unwritable: copy the old build to a root-owned folder (e.g. `/Applications`, as a standard
   user without write access) and click Update. Expected: `Xenon can't replace itself at
   /Applications…`, with no admin prompt.

On Windows and Linux, steps 2–6 are the same with the portable `xenon.exe` / `xenon` in a
user-owned folder.
