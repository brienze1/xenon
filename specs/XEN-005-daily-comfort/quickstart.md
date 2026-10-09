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
