# Implementation Plan: Daily Comfort (Stage 5)

**Branch**: `xen-005` (stacked on `xen-004`) | **Date**: 2026-10-09 | **Spec**: [spec.md](spec.md)

## Summary

Remaining auth types (API key; OAuth 2.0 with system-browser + loopback + PKCE and vault-held,
auto-refreshed tokens; AWS SigV4 via the AWS SDK signer; Digest two-step via icholy/digest), a
per-workspace persistent encrypted cookie jar with dialog, response tab and script access, sidebar
drag and drop with file moves, Postman's tab menu with same-request duplicate tabs, a closed-tab
stack, the complete hotkey set with AltGr safety, quick search, and Go-generated code snippets
with secret masking.

## Technical Context

**Language/Version**: Go 1.25+, TypeScript 5 + React 19

**Primary Dependencies** (new): `github.com/aws/aws-sdk-go-v2` v1.47.2 (signer only) +
`github.com/aws/smithy-go` v1.28.4, `github.com/icholy/digest` v1.2.0,
`golang.org/x/net/publicsuffix` (x/net v0.60.0); forked stdlib cookiejar (BSD)

**Storage**: auth configs in workspace YAML; tokens in the vault; `cookies.age`; `state.json`
additions

**Testing**: godog BDD with: a mock OAuth provider (httptest: authorize endpoint redirecting to
the loopback callback, token/refresh endpoints recording client auth and params), a "browser
opener" mock that follows the authorize URL, AWS SigV4 test-suite vectors, a Digest test server
(RFC 7616 examples), cookie-setting ApiMock routes; snippet corpus executed against ApiMock in
CI (curl, node, go, python available on runners); Vitest + RTL

**Target Platform / Project Type**: desktop-app, macOS/Windows/Linux

**Performance Goals**: drag-and-drop move < 200 ms for a 1,000-request workspace; snippet
generation < 50 ms

**Constraints**: tokens/cookies never in workspace files or unencrypted on disk; loopback-only
OAuth redirect; hotkeys AltGr-safe

**Scale/Scope**: tens of OAuth configs, thousands of cookies per workspace

## Constitution Check

- [x] **I. Local-first** — no Xenon service; OAuth talks only to the user's provider; loopback
  listener bound to 127.0.0.1 only. ✅
- [x] **II. Secrets** — tokens in vault (masked DTOs; Reveal/Copy only); cookies encrypted with
  the vault identity; snippets masked by default; Postman-imported literal tokens moved to the
  vault. ✅
- [x] **III. Go owns logic** — OAuth, signing, digest, jar, moves, snippet generation in Go; UI
  only renders. ✅
- [x] **IV. Architecture skills** — see structure; signing/jar/OAuth HTTP calls are integration
  webservices; browser opener and loopback listener are infra adapters with setters. ✅
- [x] **V. BDD-first** — features per story with outer boundaries mocked (provider, browser
  opener, AWS vectors, digest server, clock). ✅
- [x] **VI. Postman parity** — field names, injection rules, snippet conventions, tab menu and
  shortcuts per research; deviations recorded in spec Assumptions. ✅
- [x] **VII. Cross-platform** — loopback/browser per OS via Wails; AltGr handling; CI matrix runs
  snippet execution and hotkey tests on 3 OSes. ✅

**Post-design re-check**: all gates pass.

## Project Structure

### Documentation
```text
specs/XEN-005-daily-comfort/
├── plan.md · research.md · data-model.md · quickstart.md
├── contracts/ format-additions.md · bindings.md
├── checklists/
└── tasks.md
```

### Source code (additions)
```text
internal/domain/entity/   api_key_auth.go, oauth2_config.go, oauth2_token.go, awsv4_auth.go,
                          digest_auth.go, digest_state.go, cookie.go, closed_tab.go,
                          snippet_options.go
internal/domain/enums/    oauth2_grant, client_authentication, token_placement, digest_algorithm,
                          digest_qop, snippet_language, drop_position (+ value pkgs)

internal/application/
├── usecase/  pure: build_authorization_url, generate_pkce, apply_api_key, apply_oauth_token,
│             compute_payload_hash, build_digest_header_inputs, rank_quick_search,
│             materialise_auth (for snippets); ports: open_browser, await_oauth_callback,
│             exchange_oauth_token, refresh_oauth_token, sign_aws_request, digest_response,
│             find/save_cookies, find/save_oauth_token, move_tree_item, generate_snippet
├── adapter/  oauth_service, cookie_service, snippet_service; tree/send services extended
└── service/  oauth.go, cookies.go, snippets.go, send_request.go (auth step per type, digest
              two-step, cookie jar per workspace, expired-token preflight), tree.go (Move,
              QuickSearch)

internal/integration/
├── webservice/ oauth2.go (token/refresh HTTP), awsv4.go (aws-sdk-go-v2 signer), digest.go
│               (icholy/digest low-level), cookiejar/ (forked jar + export/import),
│               snippets/ curl.go, jsfetch.go, gonative.go, pythonrequests.go (+ golden tests)
├── persistence/ cookies.go (encrypted with vault identity), tree.go Move, ui_state additions,
│                vault.go oauth2 section
└── entrypoint/controller/ oauth_*.go, cookie_*.go, snippet_*.go, tree_move.go, tree_quick_search.go

internal/infra/
├── adapter/  browser.go (Wails app.Browser.OpenURL; Get/Set), loopback.go (127.0.0.1 listener;
│             Get/Set)
└── server/binding/ oauth.go, cookie.go, snippet.go; event/events.go adds cookies:changed

test/integration/
├── features/auth/      api_key, oauth2_grants, oauth2_refresh, oauth2_errors, awsv4, digest
├── features/cookies/   jar_send_store, redirects, dialog_crud, persistence_encrypted,
│                       disable_jar, script_jar_allowlist, isolation
├── features/tree/      drag_drop_moves, quick_search
├── features/snippets/  generation_golden, masking, execution (runs generated snippets)
├── mock/               oauth_provider.go, browser_opener.go, digest_server.go
└── resources/          aws-sigv4-test-suite/, snippets/golden/

frontend/src/
├── features/request-builder/ AuthPanel: ApiKeyForm, OAuth2Form (+ TokenPanel, legacy badges,
│                             callback guidance), AwsV4Form, DigestForm
├── features/cookies/         store.ts, CookiesDialog (domains, CRUD, allowlist), ResponseCookiesTab
├── features/sidebar/         dnd-kit tree with drop indicators, auto-expand/scroll
├── features/tabs/            TabContextMenu, closed-tab stack, duplicate-tab handling
├── features/side-panel/      CodePanel (language, options, include secrets, copy)
├── features/search/          QuickSearch (Cmd/Ctrl+K)
├── features/settings/        ShortcutsList
└── app/hotkeys.ts            full table (FR-012), event.code matching, AltGraph guard
```

**Structure Decision**: unchanged layering. OAuth token endpoints, AWS signing, digest and the
cookie jar talk HTTP or implement HTTP-client concerns → integration webservices. The browser
opener and loopback listener touch the OS/Wails → infra adapters behind setters (mocked in BDD).
Snippet generators are output formats of an external contract (Postman's code generators) →
integration webservice package, fed by an application use case that materialises the request.

## Key design decisions

1. **Auth pipeline step** in `send_request.go` after variable resolution: apply auth by type
   (API key / OAuth token / Bearer / Basic / AWS sign last, after all headers final / Digest
   two-step).
2. **Token ownership** = the auth owner whose config is effective (inheritance walk), so inherited
   auth shares one token.
3. **Cookie jar per workspace** loaded on activation (decrypt) and kept in memory; send uses it
   unless `disableCookieJar`.
4. **Duplicate tabs** reference the same RequestID; XEN-001 revision conflict handling already
   covers concurrent edits; sidebar focus picks most recent.
5. **Snippets in Go** from the same materialised request the send would use (minus Digest
   challenge if none cached → notice).

## Complexity Tracking

No constitution violations to justify.
