# Quickstart: validating stage 2

Prerequisites as in [XEN-001 quickstart](../XEN-001-shell-and-storage/quickstart.md).

## Automated
```bash
make bdd TAG=@xen_002
make validate
```

## Manual smoke (≈10 min, `make dev`)
1. In "Acme", create environments `dev` and `prod`. `prod` turns red with "Confirm before send" on;
   `environments/prod.yaml` exists ([format](contracts/format-additions.md)).
2. Add `base_url` to both with different initial values. Send `GET {{base_url}}/get` (httpbin) with
   `dev` active, switch (Cmd/Ctrl+Shift+E) to `prod` → confirm prompt → target changes.
3. Set a current value for `base_url` in `dev` → file unchanged, request uses current; Reset →
   initial again; Persist → file updated.
4. Add secret `api_key = s3cr3t`; header `X-Key: {{api_key}}` → httpbin echoes it.
   `grep -r s3cr3t <workspace>` and `grep -r s3cr3t <app-data>` → no match (vault is binary
   ciphertext). Quit, relaunch, send again → works without typing anything.
5. Delete the app-data `vault.age` → secret shows "missing"; Send warns; Send anyway sends the
   literal `{{api_key}}`.
6. Hover `{{base_url}}` (green, value + scope), type `{{ba` → suggestions; open Variables panel →
   edit value in place; add an unresolved one to globals.
7. Body `{"id":"{{$guid}}","a":"{{$guid}}","t":{{$timestamp}}}` → two different UUIDs, Unix time.
8. Linux without Secret Service (or `XENON_FORCE_PASSPHRASE=1` dev property): passphrase setup
   on first secret, prompt once per start, wrong passphrase rejected, "Reset vault" path.
