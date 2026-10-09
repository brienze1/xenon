# Research: Daily Comfort (Stage 5)

Verified 2026-10-09 against postman-runtime authorizers (`oauth2.js`, `aws4.js`, `digest.js`,
`apikey.js`), postman-code-generators (curl, js-fetch, golang, python-requests), Postman docs v12
(OAuth 2.0, cookies, navigation, shortcuts), RFC 8252/9700/7616/6265, and module proxy versions.

## R1. OAuth 2.0
- **Decision**: own Go implementation (no SDK needed): authorization URL building, loopback
  listener (`net.Listen` on the callback host:port, path match, one request, HTML "you can close
  this window" page), `state` (random 128-bit) check, PKCE (S256 default, verifier 64 random
  chars when empty), token exchange / refresh with client authentication as Basic header or body,
  custom parameters with `send_as` body|url|header, audience/resource, `id_token` capture, expiry
  from `expires_in`. Browser opened via Wails `app.Browser.OpenURL` (infra adapter).
- **Field names** = Postman's exported keys (format additions §2) for lossless import/export.
- **Callback**: loopback only (RFC 8252 §7.3); `localhost` accepted but `127.0.0.1` default;
  default port 53682 (configurable). Postman's `oauth.pstmn.io` callback → guidance (spec US1 #9).
- **Grants**: implicit/password supported, labelled legacy (RFC 9700).

## R2. AWS Signature v4
- **Decision**: `github.com/aws/aws-sdk-go-v2/aws/signer/v4` (module `aws-sdk-go-v2` v1.47.2 +
  `smithy-go` v1.28.4; no service clients). Xenon computes the payload SHA-256 and sets
  `X-Amz-Content-Sha256` before `SignHTTP`; query mode via `PresignHTTP` with `X-Amz-Expires`
  (default 86400 like aws4). `DisableURIPathEscaping` for service `s3`. Form-data bodies signed as
  empty (Postman parity). Stale `Authorization`/`X-Amz-*` headers and query parameters removed
  first.
- **Test**: AWS SigV4 test suite vectors (`aws-sig-v4-test-suite`) vendored as BDD resources.

## R3. Digest
- **Decision**: `github.com/icholy/digest` v1.2.0 (MIT) low-level `ParseChallenge` + `Digest`;
  `-sess` variants via precomputed `Options.A1`. The send service performs the two-step exchange
  itself so the console shows both exchanges; RFC-correct hexadecimal `nc`; challenge state
  (realm, nonce, opaque, nc) cached locally per auth owner (not in files). One retry per send.

## R4. Cookie jar
- **Decision**: fork `net/http/cookiejar` (BSD) into `internal/integration/webservice/cookiejar/`
  adding `All()`, `Set/Delete(domain, path, name)`, `DeleteDomain`, `Clear`, `Export/Import`
  (JSON entries); keep `golang.org/x/net/publicsuffix`. One jar per workspace in memory; persisted
  to `workspaces/<localId>/cookies.age` encrypted to the workspace vault identity (XEN-002),
  written debounced (1 s) after changes and on shutdown; session cookies (no expiry) are not
  persisted. If the vault is locked, cookies stay in memory and are persisted after unlock.
- `http.Client.Jar` captures redirect hops; jar cookies are merged with a manual Cookie header
  (Go's `AddCookie` behaviour = Postman's merge). Per-request "Disable cookie jar" → no Jar on the
  client for that send.
- Script allowlist: per workspace in local data; `pm.cookies.jar()` checks the URL host against it.
- **Alternatives**: `juju/persistent-cookiejar` (unmaintained since 2017/2022).

## R5. Code snippets
- **Decision**: Go generators ported from postman-code-generators conventions (option names and
  defaults, flag order, body rendering per mode, `--globoff`, `--data-raw` when `@` present, Go
  generator gofmt'd output — fixing upstream's non-compiling fractional timeout — Python JSON
  via `json.dumps`, fetch with `Headers`/`FormData`/`URLSearchParams`). Input = the effective
  request after variable resolution and auth materialisation (OAuth/API key/Basic/Bearer/AWS/
  Digest headers computed as for a send, Digest using cached challenge when present), with
  secrets masked unless "Include secrets".
- Snippets are generated in Go (`SnippetService.Generate`), so the UI never needs secret values
  unless the user opted in.

## R6. Hotkeys
- **Decision**: a single keydown handler in `app/hotkeys.ts` matching on `event.code` plus
  modifiers; when `event.getModifierState("AltGraph")` is true the event is ignored for Ctrl+Alt
  shortcuts; Cmd on macOS / Ctrl elsewhere. Webview accelerators: WebView2 accelerator keys left
  enabled and the handler calls `preventDefault` for owned shortcuts; verified per OS in the
  quickstart. Ctrl+Tab and Cmd+Shift+] / [ both map to next/previous tab.

## R7. Drag and drop
- **Decision**: `@dnd-kit` sortable tree (already in the stack) in the UI; Go `TreeService.Move`
  performs the move: file/directory rename into the new parent (unique file name via XEN-001
  naming), both parents' `order` rewritten, own-write suppression for the watcher, id → path index
  updated, open tabs unaffected (they reference ids).
