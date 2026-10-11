# Feature Specification: Daily Comfort (Stage 5)

**Feature Branch**: `xen-005`

**Created**: 2026-10-09

**Status**: Draft

**Input**: User description: "Xenon Stage 5 — Daily comfort: sidebar drag and drop, tab right-click menu, full hotkeys and shortcuts list, per-workspace cookie jar, code snippets, and the remaining auth types (API key, OAuth 2.0, AWS Signature, Digest). Source: docs/prd/xenon-mvp-prd.md; builds on XEN-001..004."

## Context

Stage 5 of 7: everything that makes a full working day comfortable once the core works. Most of
it is UI behaviour Postman users rely on without thinking (drag and drop, tab menu, hotkeys),
plus three things real APIs need: a cookie jar like a browser, OAuth 2.0 token flows, and AWS /
Digest signing. Code snippets let Luis hand a request to a colleague or a script.

## Clarifications

### Session 2026-10-09

- Q: Where does OAuth 2.0 authorization-code login happen? → A: In the system browser, with
  Xenon catching the redirect on a loopback address (`http://127.0.0.1:<port>/…`), with PKCE.
- Q: Expired OAuth 2.0 tokens? → A: Tokens are secrets kept in the workspace vault and
  auto-refreshed before a send when a refresh token exists; otherwise Luis is asked to get a new
  one.
- Q: Cookies across restarts? → A: Persisted per workspace, encrypted with the workspace's vault
  key; clearable any time.
- Q: Duplicate tab — copy or same request? → A: Same request in a second tab, as in Postman; saving
  one tab makes the other show XEN-001's "changed on disk" banner (keep mine / load from disk).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - OAuth 2.0 and the remaining auth types (Priority: P1)

Luis's company APIs use OAuth 2.0 (authorization code with PKCE via the company SSO, client
credentials for service-to-service), some use API keys, one uses AWS Signature (API Gateway), and
an old one uses Digest. He configures each in the Authorization tab (or on a collection/folder for
inheritance) and requests just work, including after tokens expire.

**Why this priority**: Without these, a share of Luis's real APIs cannot be called — the PRD calls
OAuth 2.0 one of the riskiest parts.

**Independent Test**: Against test providers, obtain tokens with each grant type, send requests,
let a token expire and send again; sign an AWS request and a Digest request and have the servers
accept them.

**Acceptance Scenarios**:

1. **Given** OAuth 2.0 with grant "Authorization code (with PKCE)", auth URL, token URL, client id,
   scope and callback URL `http://127.0.0.1:53682/callback`, **When** Luis clicks "Get new access
   token", **Then** his system browser opens the provider's login, and after he signs in Xenon
   receives the code on the loopback address, exchanges it (with the PKCE verifier) and shows the
   token (type, expiry, scopes; ID token if returned) with "Use token".
2. **Given** grant types client credentials, password, implicit and authorization code without
   PKCE, **Then** each can obtain a token with Postman's fields (client authentication as Basic
   header or in the body; custom auth/token/refresh request parameters sent in body, URL or
   header; audience/resource), implicit and password marked "legacy" in the UI.
3. **Given** a token, **Then** requests carry it as `Authorization: <prefix> <token>` (default
   prefix `Bearer`) or as `access_token` query parameter per "Add auth data to"; any existing
   `Authorization` header or `access_token` parameter is replaced.
4. **Given** a token stored for a request or an inherited collection/folder auth, **Then** the token
   value is in the workspace vault — never in the workspace files — and is masked wherever shown.
5. **Given** the token expired and a refresh token exists, **When** Luis sends, **Then** Xenon
   refreshes it first (using the refresh URL or the token URL and the refresh parameters) and
   sends with the new token; if refreshing fails or no refresh token exists, he is told the token
   expired and offered "Get new access token" (or "Send anyway").
6. **Given** API key auth (key, value, in header or query), **Then** the request carries it, replacing
   a same-named header (case-insensitive) or query parameter.
7. **Given** AWS Signature auth (access key, secret key — secrets allowed via `{{variables}}`,
   session token, region default `us-east-1`, service default `execute-api`, "add auth data to
   query"), **Then** the request is signed with AWS Signature Version 4 (headers or presigned
   query) including a payload hash, and AWS accepts it.
8. **Given** Digest auth (username, password; realm, nonce, algorithm, qop, nonce count, client
   nonce, opaque — advanced fields filled from the server's challenge), **When** Luis sends,
   **Then** Xenon sends, receives the 401 challenge, answers it, and shows the final response; the
   console shows both exchanges.
9. **Given** a configuration imported from Postman that used Postman's web callback URL
   (`https://oauth.pstmn.io/v1/callback`), **Then** the auth panel explains that the provider must
   allow Xenon's loopback callback and shows the exact URL to register.

---

### User Story 2 - Cookies like a browser (Priority: P1)

Cookies set by responses are stored per workspace and sent automatically on later requests to the
same domain, like a browser. Luis can see, edit and delete them by domain in a Cookies dialog, and
see the cookies of each response in a Cookies tab. They survive restarts.

**Why this priority**: Session-based APIs and login flows break without a jar; the response Cookies
tab is part of the PRD's response viewer.

**Independent Test**: Log in against a test server that sets a session cookie, call an endpoint that
requires it, restart, call again; edit and delete cookies in the dialog.

**Acceptance Scenarios**:

1. **Given** a response sets `Set-Cookie: session=abc; Path=/; HttpOnly`, **Then** the next request
   to that domain sends `Cookie: session=abc`, shown as an auto-generated header that can be
   switched off.
2. **Given** a redirect chain where intermediate responses set cookies, **Then** all of them are
   stored.
3. **Given** the response Cookies tab, **Then** it lists the cookies the response set: name, value,
   domain, path, expiry, HttpOnly, Secure.
4. **Given** the Cookies dialog, **Then** it lists cookies grouped by domain; Luis can add a domain,
   add, edit and delete cookies, delete a domain, and clear all.
5. **Given** a request with a manual `Cookie` header, **Then** jar cookies are merged with it (as in
   Postman); a request setting "Disable cookie jar" sends and stores no jar cookies.
6. **Given** two workspaces, **Then** each has its own jar; cookies never cross workspaces.
7. **Given** a restart, **Then** cookies (except session cookies without expiry, which are kept
   until the app quits) are still there; on disk they are encrypted.
8. **Given** a script calls `pm.cookies.jar().get(url, name, cb)` / `set` / `unset` / `clear` /
   `getAll`, **Then** it works for domains on the workspace's script allowlist (managed in the
   Cookies dialog), and fails with Postman's error for other domains.

---

### User Story 3 - Drag and drop in the sidebar (Priority: P2)

Luis reorganises collections by dragging: requests into other folders or collections, requests
within a folder to reorder, folders to reorder or nest, collections to reorder. The order is saved
to disk.

**Why this priority**: Daily convenience; menus can do the moves meanwhile.

**Independent Test**: Perform each kind of drag, restart, check order and files on disk.

**Acceptance Scenarios**:

1. **Given** a request in folder A, **When** Luis drags it onto folder B (or a collection),
   **Then** it moves there (its file moves with it) and its open tab stays open and points to the
   new location.
2. **Given** requests in a folder, **When** he drags one between two others, **Then** the order
   changes and is saved.
3. **Given** folders, **When** he drags a folder between siblings or onto another folder, **Then**
   it reorders or nests; dropping a folder into its own descendant is refused.
4. **Given** collections, **When** he drags one, **Then** collections reorder.
5. **Given** a drag in progress, **Then** a drop indicator shows "before / inside / after", the tree
   auto-expands a folder held over for 600 ms, and auto-scrolls near the edges; Esc cancels.
6. **Given** a move would create a name clash with a sibling's file, **Then** the moved item gets a
   unique file name (display names may stay equal).

---

### User Story 4 - Tab menu and every hotkey (Priority: P2)

Tabs have Postman's right-click menu, and every PRD shortcut works, Cmd on macOS and Ctrl on
Windows/Linux. Settings lists all shortcuts.

**Why this priority**: Muscle memory is a PRD goal; it is polish over working basics.

**Independent Test**: Exercise each menu item and each shortcut on all three OSes.

**Acceptance Scenarios**:

1. **Given** a tab's right-click menu, **Then** it offers: New request, Duplicate tab, Close tab,
   Force close tab, Close other tabs, Close all tabs, Close tabs to the left, Close tabs to the
   right, Force close all tabs, Reveal in sidebar — each doing what it says; "close" items ask
   about unsaved changes, "force" items discard them without asking.
2. **Given** Duplicate tab, **Then** a second tab opens on the same saved request (with a copy of
   the first tab's unsaved edits, if any); saving in one tab makes the other show the "changed on
   disk" banner. Clicking the request in the sidebar focuses the most recently active of its tabs.
3. **Given** the shortcuts in FR-012, **Then** each performs its action; Ctrl+Tab / Ctrl+Shift+Tab
   and Cmd/Ctrl+Shift+] / [ both cycle tabs.
4. **Given** a keyboard layout where AltGr produces characters (e.g., German, Polish), **Then**
   typing those characters in text fields never triggers Ctrl+Alt shortcuts.
5. **Given** Settings → Shortcuts, **Then** every shortcut is listed with its key combination for
   the current OS.

---

### User Story 5 - Code snippets (Priority: P2)

The right side panel has a Code view showing the current request as cURL, JavaScript fetch, Go
net/http or Python requests, ready to copy. Secrets are masked unless Luis chooses to include them.

**Why this priority**: Handy for sharing and scripting; not needed to send requests.

**Independent Test**: For requests covering every body mode and auth type, compare snippets to
expected outputs and run them against a test server.

**Acceptance Scenarios**:

1. **Given** the Code view, **Then** Luis picks a language and sees the snippet for the current
   request (including unsaved edits), with variables resolved, inherited auth and headers applied,
   and auth materialised into headers/query parameters.
2. **Given** the request uses secrets, **Then** the snippet shows `•••••` for them by default;
   "Include secrets" shows real values (and copying then puts real values on the clipboard).
3. **Given** snippet options (cURL: multi-line, long flags, line continuation `\` `^` or `` ` ``,
   quote type; all: follow redirects, timeout, trim body), **Then** the snippet follows them and
   they are remembered.
4. **Given** "Copy", **Then** the exact snippet text is on the clipboard.
5. **Given** a generated snippet is run against the same server, **Then** it sends an equivalent
   request (method, URL, headers, body).

---

### Edge Cases

- **OAuth callback port busy**: Xenon reports the port in use and lets Luis change the callback URL.
  A callback URL that is not a loopback address (not `127.0.0.1`, `[::1]` or `localhost` with
  `http`) is refused with an explanation.
- **OAuth login abandoned**: the wait for the browser redirect times out after 5 minutes (Cancel
  available immediately); no token is stored.
- **OAuth state mismatch / provider error**: shown with the provider's error and description; no
  token stored.
- **Tokens and git**: token values and refresh tokens are never written to workspace files; the
  auth configuration (URLs, client id, scopes) is. A client secret typed literally is stored as
  typed (as with other auth fields) — `{{secret_variable}}` is recommended in the field hint.
- **Inherited OAuth**: a token obtained on a collection's auth is shared by all requests inheriting
  it.
- **Token types**: a `mac` token type is not sent (as in Postman), with a notice.
- **AWS form-data bodies** are signed as empty payloads (Postman behaviour), noted in the panel.
- **Digest `nc`** follows RFC 7616 (hexadecimal), not Postman's decimal quirk; Digest retry happens
  once per send (no loops); "disable retry" sends without answering the challenge.
- **Cookie limits**: invalid cookies (bad domain, public-suffix domains) are ignored as a browser
  would; SameSite and `__Host-`/`__Secure-` prefixes are stored but not enforced (as in Postman).
- **Drag across workspaces** is not possible (PRD); moving a request whose file is open in another
  editor follows XEN-001 FR-017.
- **Shortcut conflicts with the OS** (e.g., Linux desktop environments taking Ctrl+Alt+arrows) are
  documented in Settings → Shortcuts.
- **Snippets for unsupported auth types** (stage 3 "unsupported") omit auth with a comment line.

## Requirements *(mandatory)*

### Functional Requirements

**Auth**

- **FR-001**: Authorization MUST support API key (key, value, in header|query), OAuth 2.0, AWS
  Signature and Digest, at request, folder and collection level with inheritance (XEN-001
  FR-014), in addition to stage-1 types.
- **FR-002**: OAuth 2.0 MUST support grants authorization code, authorization code with PKCE
  (S256 or plain; verifier generated when empty), implicit, password and client credentials, with
  Postman's configuration fields (token name, callback URL, auth URL, access token URL, client
  id/secret, username/password, scope, state, client authentication header|body, add token to
  header|query, header prefix, refresh token URL, custom auth/token/refresh request parameters
  with send-as body|URL|header, audience, resource, auto-refresh on/off default on).
- **FR-003**: Authorization-code grants MUST open the system browser and receive the redirect on a
  loopback listener at the configured callback URL (default `http://127.0.0.1:53682/callback`;
  must be loopback http), verify `state`, and stop listening after success, error, cancel or 5
  minutes. Xenon MUST only open `https:` authorization URLs (or `http:` on a loopback host) in
  the browser; any other scheme is refused with an explanation.
- **FR-004**: Tokens (access, refresh, ID) MUST be stored only in the workspace vault, keyed by the
  owning auth (collection/folder/request id), shown masked with reveal/copy, with expiry, scopes and
  token type; actions: Get new access token, Refresh, Use token (access or ID token), Clear.
- **FR-005**: Before a send whose effective auth is OAuth 2.0 with an expired token, Xenon MUST
  refresh it when a refresh token exists and auto-refresh is on; otherwise warn with "Get new
  access token" / "Send anyway". Concurrent sends sharing an expired token MUST trigger a single
  refresh (the others wait for its result), since providers may rotate refresh tokens.
- **FR-006**: Token injection MUST follow Postman: remove existing `Authorization` header /
  `access_token` query parameter, then add `Authorization: <prefix> <token>` or `?access_token=`.
- **FR-007**: AWS Signature MUST sign requests with Signature Version 4 (header mode with
  `X-Amz-Content-Sha256` payload hash, or query presign mode), removing stale `X-Amz-*` and
  `Authorization` values first, using resolved variables for keys.
- **FR-008**: Digest MUST perform the challenge–response exchange (401/403 with `WWW-Authenticate:
  Digest`), supporting algorithms MD5, MD5-sess, SHA-256, SHA-256-sess, SHA-512-256,
  SHA-512-256-sess and qop auth / auth-int, storing the server's realm/nonce/opaque for reuse
  (locally, not in workspace files), with "disable retry".

**Cookies**

- **FR-009**: Each workspace MUST have a cookie jar that stores cookies from every response
  (including redirect hops) and sends matching cookies with requests (merged with any manual
  `Cookie` header), with a per-request "Disable cookie jar" setting; cookies persist across
  restarts encrypted with the workspace vault key; session cookies last until the app quits.
- **FR-010**: A Cookies dialog MUST list cookies by domain with add/edit/delete per cookie, add and
  delete domain, clear all, and manage the script allowlist; the response MUST have a Cookies tab
  listing the cookies it set.
- **FR-011**: `pm.cookies.jar()` (XEN-004) MUST operate on the workspace jar for allowlisted
  domains with Postman's callback API and errors.

**Tabs, hotkeys, drag and drop**

- **FR-012**: Shortcuts (Cmd on macOS, Ctrl on Windows/Linux): Send Cmd+Enter; Save Cmd+S; Save as
  Cmd+Shift+S; New tab Cmd+T; Close tab Cmd+W; Force close tab Cmd+Alt+W; Reopen last closed tab
  Cmd+Shift+T; Next/previous tab Ctrl+Tab / Ctrl+Shift+Tab and Cmd+Shift+] / Cmd+Shift+[; Go to tab
  1–8 Cmd+1…8 (Cmd+9 last tab); Focus URL Cmd+L; Search requests Cmd+K; Find in response Cmd+F;
  Toggle sidebar Cmd+\; Toggle console Cmd+Alt+C; Rename selected item Cmd+E; Duplicate selected
  item Cmd+D; Import Cmd+O; Settings Cmd+,; Switch environment Cmd+Shift+E. Matching MUST use
  physical key codes and ignore AltGr combinations.
- **FR-013**: "Reopen last closed tab" MUST restore the last 20 closed tabs in reverse order,
  including their unsaved drafts.
- **FR-014**: Cmd/Ctrl+K MUST open a quick search over the current workspace's requests (name,
  URL, path) and open the chosen one in a tab.
- **FR-015**: The tab right-click menu MUST offer the items in User Story 4 #1 with the described
  behaviour.
- **FR-016**: Settings MUST list every shortcut with its combination for the current OS
  (remapping is not in the MVP).
- **FR-017**: The sidebar MUST support dragging requests between and within folders/collections,
  folders (reorder, nest, un-nest) and collections (reorder), persisting order and moving files;
  invalid drops (into own descendant, across workspaces) are refused.

**Code snippets**

- **FR-018**: The right side panel MUST have a Code view generating cURL, JavaScript fetch, Go
  net/http and Python requests for the current request (draft included) with resolved variables,
  inherited auth/headers and materialised auth, following Postman's code-generator conventions and
  options; options are remembered per language.
- **FR-019**: Secrets MUST be masked in snippets unless "Include secrets" is on; copy copies exactly
  what is shown.

### Key Entities

- **OAuth2Config** (in files): grant type, URLs, client id/secret (as typed or `{{var}}`), scope,
  state, PKCE method, client authentication, token placement, header prefix, custom parameters,
  audience, resource, auto-refresh, callback URL.
- **OAuth2Token** (vault): access token, refresh token, ID token, token type, expiry, scopes,
  obtained-at, owner auth id, which token to use.
- **Cookie** (encrypted local store per workspace): name, value, domain, path, expires, HttpOnly,
  Secure, SameSite (stored), host-only, created/last-access.
- **Cookie allowlist** (per workspace, local): domains scripts may access.
- **DigestState** (local): realm, nonce, opaque, nonce count per auth owner.
- **Closed-tab stack** (local, per workspace): last 20 closed tabs with drafts.
- **Snippet options** (local): per language.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Each OAuth 2.0 grant type obtains a token from a standards-compliant test provider
  in 100% of runs; an expired token is refreshed automatically before the send in 100% of runs
  where a refresh token exists.
- **SC-002**: AWS Signature v4 requests are accepted by AWS's reference test vectors (100% of the
  published SigV4 test suite cases that apply to HTTP requests).
- **SC-003**: Digest auth succeeds against RFC 7616 test servers for every supported algorithm and
  qop.
- **SC-004**: 0 OAuth tokens or cookie values appear in any workspace file or unencrypted local
  file.
- **SC-005**: Every PRD shortcut works on macOS, Windows and Linux, and no shortcut fires while
  typing AltGr characters on German and Polish layouts.
- **SC-006**: Generated snippets for a corpus covering all body modes and auth types reproduce the
  original request (method, URL, headers, body) when executed, for all four languages.
- **SC-007**: Reordering by drag and drop takes one gesture and the result survives a restart in
  100% of cases.

## Assumptions

- Postman (postman-runtime authorizers, postman-code-generators, cookie behaviour) is the parity
  reference; documented deviations: loopback callback instead of Postman's web relay, RFC-correct
  Digest `nc`, "Close tabs to the left/right" added (PRD).
- Implicit and password grants are supported for compatibility but labelled legacy (RFC 9700).
- Shortcut remapping is a post-MVP nice-to-have (PRD).
- SC-002 "applicable" SigV4 vectors are the 20 the AWS SDK signer reproduces for an HTTP client:
  path normalisation, single encoding of raw UTF-8/space paths, header line folding, repeated
  header names (Xenon sends one header per name), encoded-key query ordering and the
  token-added-after-signing case are out (tasks.md Phase 3 notes). As in Postman's aws4,
  `X-Amz-Content-Sha256` is sent for S3 only; the payload hash is always signed.
