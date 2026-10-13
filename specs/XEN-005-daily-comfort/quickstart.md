# Quickstart: validating stage 5

## Automated
```bash
make bdd TAG=@xen_005
make validate
```

## Manual (≈20 min, `make dev`)
1. OAuth 2.0 (code + PKCE) against a real IdP test tenant (e.g., a dev Keycloak/Auth0) with
   redirect `http://127.0.0.1:53682/callback` → browser login → token shown (masked) → send works;
   `grep` workspace folder for the token → no match. Shorten the IdP token lifetime → send after
   expiry refreshes silently.
2. Client credentials, API key (header and query), AWS SigV4 against an API Gateway test stage,
   Digest against `https://httpbin.org/digest-auth/auth/u/p/MD5` (and SHA-256).
3. `https://httpbin.org/cookies/set?a=1` → Cookies tab shows `a`; next `/cookies` sends it;
   Cookies dialog edit/delete; restart → still there; `cookies.age` is binary.
4. Drag a request to another folder, reorder folders and collections; restart; check files.
5. Tab menu: every item; Duplicate tab → two tabs on one request, save in one → banner in the
   other. Every shortcut on the current OS; on a German layout type `@` (AltGr+Q) and `{` in the URL
   bar → characters typed, no shortcut.
6. Code panel: four languages, include/mask secrets, copy; run the cURL snippet in a terminal →
   same response.

## Webview accelerators (T026, walked per OS at the stage end, T032)
`app/hotkeys.ts` calls `preventDefault` on every shortcut it owns, matched on `event.code` with
AltGr ignored. Check on each OS that the webview does not take the key first:

| Shortcut | macOS (WKWebView) | Windows (WebView2) | Linux (WebKitGTK) |
|---|---|---|---|
| Cmd/Ctrl+F (Find in response) | | | |
| Cmd/Ctrl+L (Focus URL) | | | |
| Cmd/Ctrl+K, Cmd/Ctrl+E, Cmd/Ctrl+D | | | |
| Cmd/Ctrl+1…9, Cmd/Ctrl+Shift+]/[ | | | |
| Ctrl+Tab / Ctrl+Shift+Tab | | | |
| Cmd/Ctrl+W, Cmd/Ctrl+Alt+W, Cmd/Ctrl+Shift+T | | | |
| Cmd/Ctrl+, and Cmd/Ctrl+\ | | | |
| AltGr+Q (`@`), AltGr+7 (`{`) on German; AltGr+C (`ć`) on Polish in the URL bar | | | |

Automated coverage: `frontend/src/app/hotkeys.test.ts` (every combo on macOS/Windows/Linux
profiles, `event.code` matching, AltGr typing never firing a shortcut).
