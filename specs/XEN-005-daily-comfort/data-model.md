# Data Model: Daily Comfort (Stage 5)

File shapes: [contracts/format-additions.md](contracts/format-additions.md).

## Auth entities (`internal/domain/entity/`)
- **ApiKeyAuth** `{ Key, Value, In: header|query }`.
- **OAuth2Config** — fields of format §oauth2; validation: authUrl required for code/implicit
  grants, accessTokenUrl for all but implicit; redirect_uri must be loopback http
  (`127.0.0.1`, `[::1]`, `localhost`) with a port.
- **AwsV4Auth** `{ AccessKey, SecretKey, SessionToken, Region (default us-east-1), Service (default
  execute-api), AddAuthDataToQuery }`.
- **DigestAuth** `{ Username, Password, Algorithm (enum), Qop (auth|auth-int|""), Realm, Nonce,
  Opaque, NonceCount, ClientNonce, DisableRetryRequest }`.

## Local / vault data (per workspace, app data)
```
workspaces/<localId>/
├── vault.age            ← gains oauth2 tokens (XEN-002 vault JSON: "oauth2": { "<ownerId>": OAuth2Token })
├── cookies.age          ← encrypted cookie jar (persistent cookies only)
└── state.json           ← gains closedTabs[], cookieAllowlist[], snippetOptions, digestState{}
```

### OAuth2Token (vault)
`{ AccessToken, RefreshToken, IDToken, TokenType, Scope, ExpiresAt?, ObtainedAt, Owner
(collection|folder|request id), UseToken (access_token|id_token) }`. Expired = `ExpiresAt` ≤ now
+ 30 s.

### Cookie
`{ Name, Value, Domain, HostOnly, Path, Expires?, Secure, HttpOnly, SameSite, Created,
LastAccess }` — RFC 6265 storage model.

### DigestState (state.json, per auth owner)
`{ Realm, Nonce, Opaque, Algorithm, Qop, NonceCount (int) }` — never contains credentials.

### ClosedTab (state.json)
Tab snapshot (XEN-001 Tab incl. draft) + closedAt; max 20, LIFO.

### Tabs (XEN-001 change)
Multiple tabs may reference the same RequestID (Duplicate tab). Sidebar click focuses the most
recently active of them. Each tab keeps its own draft and base revision (XEN-001 conflict rules
apply between them).

## OAuth state machine (per owner)
```
NoToken --Get new--> Authorizing (browser + loopback, ≤ 5 min) --code--> Exchanging --> HasToken
Authorizing --cancel / timeout / error / state mismatch--> NoToken (error shown)
HasToken --expired + refresh token + autoRefresh--> Refreshing --> HasToken | Expired
HasToken --Clear--> NoToken
Send with Expired → prompt (Get new access token / Send anyway)
```
