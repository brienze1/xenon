# Quickstart: validating stage 6

## Automated
```bash
make bdd TAG=@xen_006
make newman      # exports corpus collections and runs them with Newman against the mock server
make validate
```

## Manual (≈10 min, `make dev`)
1. Import `postman/full-v2.1.json` (stage 3), export it as Postman, open the export in Postman →
   identical tree, scripts, auth, examples; the export summary lists nothing dropped.
2. Export a folder and a single request as Postman → inherited auth/scripts present at the root.
3. Export an environment with secrets as Postman → secrets have type secret, empty values.
4. Export the workspace as Xenon YAML with "Include secrets" (passphrase) → import it as a new
   workspace on another machine/user account → secrets restored after entering the passphrase;
   wrong passphrase rejected; "import without secrets" marks them missing.
5. Export a collection as a bundled `.yaml` with "Export current values as initial values" → the
   bundle carries current values; the workspace files are unchanged.
6. Cancel a large export midway → no file left behind.
