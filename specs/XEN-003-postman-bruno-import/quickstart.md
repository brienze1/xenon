# Quickstart: validating stage 3

## Automated
```bash
make bdd TAG=@xen_003
make validate
```

## Manual (≈15 min, `make dev`)
1. Cmd/Ctrl+O → drop `test/integration/resources/import/postman/full-v2.1.json` → preview shows
   counts → import into Acme as a new collection → summary lists unsupported auth, `pm.require`,
   missing files, folder variables and pm APIs outside the MVP subset; each warning opens its
   item. Compare a few requests against [postman-mapping.md](contracts/postman-mapping.md).
2. Import `postman/dev.postman_environment.json` (has secrets) → environment created; grep the
   workspace folder for the secret value → no match; Variables show it as set (vault).
3. Drop the dump zip `postman/dump.zip` → assignment screen → assign items to Acme, Globex and a
   new workspace created inline → one summary per workspace.
4. Import the same collection again → clash prompt → "Keep both" + apply to all → "(imported)"
   copies exist.
5. Drop `bruno/bru-collection/` and `bruno/yaml-collection/` → same tree and order as in Bruno;
   scripts show `pm.*`; untranslated lines carry `// XENON: untranslated` and are listed.
6. Paste `curl -u a:b -F file=@/tmp/x.txt -G -d q=1 https://httpbin.org/anything` into a tab's URL
   bar → request filled (undo restores); the same command in cmd (`^`) and PowerShell (`` ` ``)
   gives the same request.
7. Send a request → "Save response as example" → example appears under the request; open it,
   edit status and body, save, restart → persists in the request file `examples:`.
