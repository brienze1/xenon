---
description: "Task list for XEN-005 — Daily Comfort (Stage 5)"
---

# Tasks: Daily Comfort (Stage 5)

**Input**: `specs/XEN-005-daily-comfort/` · **Prerequisite**: XEN-001..004 implemented.

**Tests**: MANDATORY. Feature files tagged `@all @xen_005 @<area>`, written first and proven red.

**Rules**: backend per architecture skills in the established map; frontend per
`frontend-architecture`.

US1 auth · US2 cookies · US3 drag & drop · US4 tabs & hotkeys · US5 snippets

---

## Phase 1: Setup

- [ ] T001 Add `github.com/aws/aws-sdk-go-v2@v1.47.2`, `github.com/aws/smithy-go@v1.28.4`, `github.com/icholy/digest@v1.2.0`, `golang.org/x/net@v0.60.0` to `go.mod`
- [ ] T002 [P] Vendor test resources: AWS `aws-sig-v4-test-suite` vectors into `test/integration/resources/aws-sigv4-test-suite/`; snippet golden corpus `test/integration/resources/snippets/` (requests covering all body modes × auth types, expected output per language from postman-code-generators conventions)
- [ ] T003 [P] Mocks in `test/integration/mock/`: `oauth_provider.go` (authorize redirecting with code/state or error, token/refresh endpoints recording client auth, params and PKCE verifier check, configurable `expires_in`), `browser_opener.go` (follows the authorize URL to the provider and its redirect to the loopback listener; can simulate abandon), `digest_server.go` (RFC 7616 challenges for each algorithm/qop) — injected via setters

---

## Phase 2: Foundational

- [ ] T004 [P] Enums + entities from data-model.md (`oauth2_grant`, `client_authentication`, `token_placement`, `digest_algorithm`, `digest_qop`, `snippet_language`, `drop_position`; auth entities, `oauth2_token`, `cookie`, `digest_state`, `closed_tab`, `snippet_options`)
- [ ] T005 Persistence: auth sub-objects per format additions (round-trip with unknown-field preservation), `disableCookieJar` setting, vault `oauth2` section, `state.json` additions (closedTabs, cookieAllowlist, snippetOptions, digestState)
- [ ] T006 [P] Infra adapters with Get/Set: `internal/infra/adapter/browser.go` (Wails `app.Browser.OpenURL`), `loopback.go` (listener bound to the callback host/port)
- [ ] T007 Update XEN-003 Postman/Bruno importers: oauth2/apikey/awsv4/digest attributes map to these sub-objects; a literal `accessToken` moves to the vault (never written to files); BDD case added to `features/import/postman_collection.feature`

---

## Phase 3: User Story 1 — Auth types (P1) 🎯 MVP

- [ ] T008 [P] [US1] BDD `test/integration/features/auth/api_key.feature`: header/query, replacing same-named header (case-insensitive) or param, empty key+value sends nothing, inheritance
- [ ] T009 [P] [US1] BDD `test/integration/features/auth/oauth2_grants.feature`: each grant against the mock provider (PKCE S256/plain + generated verifier; client auth header/body; custom params by send_as; audience/resource; implicit token from fragment); token stored in vault only (grep workspace + app data except `*.age`); injection header/query with prefix and replacement rules; `mac` token not sent; inherited auth shares the token; Postman callback guidance flag in config DTO; non-https (non-loopback http) or non-http(s) authUrl refused before opening the browser
- [ ] T010 [P] [US1] BDD `test/integration/features/auth/oauth2_refresh.feature` and `oauth2_errors.feature`: auto-refresh before send when expired (time mock), refresh disabled/absent → `SND-07409` then send-anyway; two concurrent sends with one expired token cause exactly one refresh call at the mock provider; provider error `OAU-04401`, state mismatch `OAU-05400`, port busy `OAU-02409`, abandon → timeout `OAU-03408` (shortened via property in tests), cancel
- [ ] T011 [P] [US1] BDD `test/integration/features/auth/awsv4.feature` (every applicable test-suite vector; header and query modes; session token; stale headers removed; form-data signed empty) and `digest.feature` (each algorithm × qop; one retry; disable retry; state cached locally not in files; console shows both exchanges)
- [ ] T012 [US1] Run T008–T011 → **red**
- [ ] T013 [US1] Pure use cases `apply_api_key.go`, `apply_oauth_token.go`, `build_authorization_url.go`, `generate_pkce.go`, `compute_payload_hash.go`; webservices `oauth2.go`, `awsv4.go`, `digest.go`; service `oauth.go` (state machine from data-model) + controllers/bindings (bindings §1); `send_request.go` auth step + digest two-step + expired-token preflight → **green**
- [ ] T014 [US1] UI `features/request-builder/components/auth/`: `ApiKeyForm`, `OAuth2Form` (all fields, legacy badges, callback guidance), `TokenPanel` (masked token, expiry, scopes, Get new / Refresh / Use / Clear / reveal / copy, waiting-for-browser state with Cancel), `AwsV4Form`, `DigestForm` (advanced fields collapsible); RTL tests

---

## Phase 4: User Story 2 — Cookies (P1)

- [ ] T015 [P] [US2] BDD `test/integration/features/cookies/jar_send_store.feature`, `redirects.feature`, `disable_jar.feature`, `isolation.feature`: Set-Cookie stored, sent on matching requests and shown as auto-generated header, merged with manual Cookie header, redirect hops captured, disable jar per request, per-workspace isolation, `cookies_set` in SendResult
- [ ] T016 [P] [US2] BDD `test/integration/features/cookies/dialog_crud.feature`, `persistence_encrypted.feature`, `script_jar_allowlist.feature`: CRUD + domain ops + clear; restart keeps persistent cookies, drops session cookies; `cookies.age` not plaintext (grep value → 0 matches); locked vault keeps cookies in memory until unlock; `pm.cookies.jar()` methods allowed/denied by allowlist with Postman errors
- [ ] T017 [US2] Run → **red**; implement `webservice/cookiejar/` fork (+ unit tests ported from stdlib), `persistence/cookies.go`, service `cookies.go`, send integration, `pm.cookies.jar()` wiring in the XEN-004 sandbox, controllers/bindings (bindings §2), event `cookies:changed` → **green**
- [ ] T018 [US2] UI `features/cookies/`: `CookiesDialog` (domains list, cookie editor with raw `Set-Cookie`-style input, delete domain, clear all, allowlist tab), `ResponseCookiesTab`; RTL tests

---

## Phase 5: User Story 3 — Drag and drop (P2)

- [ ] T019 [P] [US3] BDD `test/integration/features/tree/drag_drop_moves.feature`: request across folders/collections (file moved, both orders rewritten, unique file name on clash), reorder within folder, folder reorder/nest/un-nest, collection reorder, into-own-descendant `TRE-02422`, watcher emits no external-change events for own moves, open tab ids unaffected
- [ ] T020 [US3] Run → **red**; implement `persistence/tree.go` Move + `move_tree_item` use case + `TreeService.Move` → **green**
- [ ] T021 [US3] UI sidebar dnd-kit tree (drop indicators before/inside/after, auto-expand 600 ms, edge auto-scroll, Esc cancel); RTL tests with dnd-kit test utilities

---

## Phase 6: User Story 4 — Tabs, hotkeys, quick search (P2)

- [ ] T022 [P] [US4] BDD `test/integration/features/tree/quick_search.feature`: ranking (name prefix > name contains > URL contains), workspace scope, limit
- [ ] T023 [US4] Run → **red**; implement `rank_quick_search.go` + `TreeService.QuickSearch` → **green**
- [ ] T024 [P] [US4] Vitest `frontend/src/features/tabs/store.test.ts` additions: every context-menu action (close/force variants, others, all, left, right, reveal, new request), duplicate tab = same request in two tabs with independent drafts, sidebar click focuses most recent, closed-tab stack (20, LIFO, drafts restored), Cmd+1..9
- [ ] T025 [P] [US4] Vitest `frontend/src/app/hotkeys.test.ts`: every FR-012 combination on mac/win/linux profiles; `event.code` matching; AltGraph events never trigger Ctrl+Alt shortcuts; Ctrl+Tab and Cmd+Shift+]/[ both cycle
- [ ] T026 [US4] Implement `features/tabs/components/TabContextMenu.tsx`, closed-tab stack in tabs store (persisted via UIState), duplicate-tab support, full `app/hotkeys.ts`, `features/search/QuickSearch.tsx`, `features/settings/ShortcutsList.tsx` → T024/T025 **green**; manual per-OS check of webview accelerators recorded in quickstart

---

## Phase 7: User Story 5 — Code snippets (P2)

- [ ] T027 [P] [US5] BDD `test/integration/features/snippets/generation_golden.feature` (every golden case × language × option set), `masking.feature` (secrets masked unless include_secrets; Copy matches shown text), `execution.feature` (CI runs generated curl/node/go/python snippets against ApiMock and compares received requests — SC-006)
- [ ] T028 [US5] Run → **red**; implement `materialise_auth.go` use case, `webservice/snippets/{curl,jsfetch,gonative,pythonrequests}.go`, `snippets.go` service, controllers/bindings (bindings §3) → **green**
- [ ] T029 [US5] UI `features/side-panel/components/CodePanel.tsx` (language picker, options per language remembered, include-secrets toggle, read-only CodeMirror, copy); RTL tests

---

## Phase 8: Polish

- [ ] T030 [P] Security sweep BDD `features/auth/secret_sweep.feature`: after OAuth, cookies, digest and snippet flows, no token/cookie/secret value in workspace files, unencrypted app data, logs or console (SC-004)
- [ ] T031 [P] Performance: move in a 1,000-request workspace < 200 ms; snippet generation < 50 ms (BDD timing assertions)
- [ ] T032 `/refactor-diff` until clean; `make validate`; CI on 3 OSes (incl. snippet execution); walk [quickstart.md](quickstart.md) on all three OSes including the AltGr check

---

## Dependencies & Execution Order
Setup → Foundational → US1 → US2 (needs vault from XEN-002 and send from US1's auth step) →
US3, US4, US5 (independent of each other) → Polish.

## Parallel opportunities
T002–T003, T006; all [P] BDD files per story; US3/US4/US5 in parallel.

## Implementation Strategy
MVP = US1 + US2 (every real API callable, sessions work) → US4 (muscle memory) → US3 → US5 →
Polish.
