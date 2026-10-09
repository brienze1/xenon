# Contract: Go ↔ UI bindings — stage 5 additions

New prefixes: `OAU` OAuth · `CKE` cookies · `SNP` snippets · `TRE` (existing) for moves.

## 1. `OAuthService`
| Method | Input | Output | Notes |
|---|---|---|---|
| `GetNewToken(req)` | `{ local_id, owner_id }` | `TokenInfo { token_type, scope, expires_at, has_refresh_token, has_id_token, use_token }` | uses the owner's resolved oauth2 config; cancellable (ctx) while waiting for the browser; `OAU-01400` invalid config, `OAU-02409` callback port busy, `OAU-03408` timed out, `OAU-04401` provider error (details: error, error_description), `OAU-05400` state mismatch |
| `Refresh(req)` | `{ local_id, owner_id }` | `TokenInfo` | `OAU-06404` no refresh token |
| `Get(req)` | `{ local_id, owner_id }` | `TokenInfo \| null` | never the token value |
| `Reveal(req)` | `{ local_id, owner_id, which: access\|refresh\|id }` | `{ value }` | only call returning a token |
| `Copy(req)` | same | — | clipboard from Go |
| `Use(req)` | `{ local_id, owner_id, use_token }` | `TokenInfo` | |
| `Clear(req)` | `{ local_id, owner_id }` | — | |
`RequestService.Send` preflight adds `SND-07409` token expired (no refresh possible) with
acknowledgement `send_with_expired_token: true`.

## 2. `CookieService`
| Method | Input | Output |
|---|---|---|
| `List(req)` | `{ local_id }` | `[{ domain, cookies: Cookie[] }]` (cookie values are shown in clear, as in Postman's Cookies dialog) |
| `Upsert(req)` | `{ local_id, cookie }` | `Cookie` (`CKE-00400` invalid) |
| `Delete(req)` | `{ local_id, domain, path, name }` | — |
| `DeleteDomain(req)` | `{ local_id, domain }` | — |
| `Clear(req)` | `{ local_id }` | — |
| `GetAllowlist(req)` / `SaveAllowlist(req)` | `{ local_id[, domains] }` | `{ domains }` |
`SendResult` gains `cookies_set: Cookie[]` (response Cookies tab). Event `cookies:changed { local_id }`.

## 3. `SnippetService`
| Method | Input | Output |
|---|---|---|
| `Languages()` | — | `[{ id: curl\|js_fetch\|go_native\|python_requests, options: [{ id, type, default, values? }] }]` |
| `Generate(req)` | `{ local_id, request (draft), language, options, include_secrets }` | `{ text, notices[] }` |
| `Copy(req)` | same as Generate | — (Go writes clipboard; avoids passing secrets to UI when masked view is shown) |

## 4. `TreeService` additions
| Method | Input | Output | Notes |
|---|---|---|---|
| `Move(req)` | `{ local_id, id, new_parent_id, index }` | `TreeNode[]` (updated tree) | `TRE-02422` into own descendant; collections: `new_parent_id = ""` reorders; cross-workspace impossible by contract |
| `QuickSearch(req)` | `{ local_id, query, limit }` | `[{ id, name, method, url, path }]` | Cmd/Ctrl+K, ranked: name prefix > name contains > URL contains |

## 5. `UIStateService` additions
`ClosedTabs` stack inside UIState (reopen handled client-side through `UIState.Save`).
