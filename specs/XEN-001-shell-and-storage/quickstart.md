# Quickstart: validating stage 1

## Prerequisites
- Go 1.25+, Node LTS, `make setup` (installs pinned wails3 v3.0.0-beta.28 and Go tools).
- macOS: Xcode CLT · Windows: WebView2 (preinstalled) · Linux: `libgtk-4-dev libwebkitgtk-6.0-dev`.

## Automated
```bash
make validate          # Go gates + BDD + coverage/mutation + frontend lint/typecheck/tests
make bdd TAG=@xen_001  # only this stage's scenarios
```
Expected: all green on macOS, Windows and Linux (CI matrix).

## Manual smoke (≈10 min) — run `make dev`
1. Welcome screen shows the marker wordmark → **Create workspace** "Acme" in an empty temp folder.
   Check the folder: `workspace.yaml`, `collections/`, `environments/`, `.gitignore`
   ([format](contracts/workspace-format.md)).
2. **New tab** (Cmd/Ctrl+T) → GET `https://httpbin.org/get?a=1&b=2` → Params shows `a`, `b`;
   untick `b` → URL updates. Send → status green 200, hover time shows DNS/connect/TLS/TTFB/
   download; Body Pretty folds; search, copy, save to file work.
3. Send to `https://httpbin.org/delay/10`, press **Cancel** → "Cancelled". Set Settings timeout
   2000 → "Timed out after 2000 ms".
4. Save the tab into new collection "Demo" (Cmd/Ctrl+S) → `collections/demo/collection.yaml` and
   `collections/demo/<name>.yaml` exist; edit + save → `git diff` touches one file.
5. Create folder "Nested" → request inside; set collection auth Bearer `abc`, request auth
   Inherit → Send to `https://httpbin.org/headers` → `Authorization: Bearer abc` echoed.
6. Single-click two requests → preview tab replaced; double-click → pinned. Edit → dot; close →
   Save/Don't save/Cancel prompt.
7. Edit a request file in a text editor while open (clean tab → reloads; dirty tab → banner).
8. Create workspace "Globex", switch back and forth → no cross-over; History shows only own sends.
9. Hide sidebar/console, resize right panel, switch split to side by side → quit → relaunch:
   everything restored, including unsaved tabs.
10. Remove "Globex" with "Move folder to trash" → folder is in the OS trash.
