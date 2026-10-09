# Quickstart: validating stage 4

## Automated
```bash
make bdd TAG=@xen_004
make parity          # runs tools/parity oracle (Node, postman-sandbox 6.7.4) + Go parity test
make validate
```

## Manual (≈15 min, `make dev`)
1. Collection script `console.log("C pre")`, folder `console.log("F pre")`, request
   `console.log("R pre")` (same for post). Send → console shows C, F, R pre → request → C, F, R
   post. Turn off "Inherit parent scripts" on the request → only R runs.
2. Pre-request: `pm.request.headers.upsert({key:"X-Sig", value: CryptoJS.HmacSHA256("a","k").toString()})`
   → httpbin `/headers` echoes it; saved request unchanged.
3. Post-response: `pm.environment.set("token", pm.response.json().headers.Host)` → next request
   with `{{token}}` sends it; Variables panel shows the current value.
4. Tests: `pm.test("ok", () => pm.response.to.have.status(200))`, a failing test,
   `pm.test.skip("later")`, legacy `tests["legacy"] = responseCode.code === 200` → Test Results
   (3/4 passed + 1 skipped) with filters.
5. `const r = await pm.sendRequest("https://httpbin.org/uuid"); console.log(r.json())` → console
   shows the nested request and the logged object.
6. `while(true){}` → stopped after 10 s, request not sent, error names the script; app stays
   responsive. Change the timeout in Settings to 2 s → stopped after 2 s.
7. `console.log(pm.environment.get("api_key"))` with a secret → console shows `•••••`.
8. Editor: `pm.response.to.have.` → completions; snippets menu inserts "Status code is 200";
   a syntax error is underlined.
