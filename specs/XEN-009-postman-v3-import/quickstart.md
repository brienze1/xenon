# Quickstart: validate Postman v3 import (XEN-009)

## Prerequisites

- The repo builds (`make build`). The BDD harness runs (`make bdd`).
- The fixtures under `test/integration/resources/import/postman/v3/` are present (task list, Phase 2).
- Optional, for regenerating fixtures: Postman CLI ≥ 1.73.0 (`postman --version`). No login is needed for
  `collection migrate`.

## Automated validation

```bash
make bdd TAG=@xen_009            # every v3 scenario
make bdd TAG=@import             # all import scenarios, old and new, must be green (SC-005)
make bdd TAG=@format_detection   # v3 rows now detected; other rows unchanged
make validate                    # final gate
```

Expected:

- every `@xen_009` scenario passes;
- no pre-existing `@import` scenario changes outcome;
- the v3 1,000-request zip imports within the large-collection bound (detect < 2 s, import < 8 s).

## Manual walk-through (desktop app)

1. **v3 JSON**: `postman collection get <id> --json > /tmp/c.json` (or use
   `test/integration/resources/import/postman/v3/rich.json`). Open Import (Cmd/Ctrl+O) and pick the file.
   - The preview says "Postman collection v3", with counts equal to the v2.1 fixture's.
   - Import it as a new collection. The tree, bodies, auth (Inherit / No auth / Bearer), scripts and examples match
     `rich-v2.1.json` imported side by side.
2. **Git-native folder**: pick `…/postman/v3/rich/`. The result is the same as step 1.
3. **Single request**: pick `…/postman/v3/rich/Users/Get user.request.yaml`.
   - Choosing "new collection" is refused with "A request must be imported into a collection or folder".
   - Importing into a folder adds the request with its example.
4. **Environment with secrets**: pick `…/postman/v3/workspace/postman/environments/staging.environment.yaml`.
   - The environment file shows `secret: true` rows and no values.
   - Revealing the row in the environment editor shows the secret value.
   - The vault-linked row is "missing", and the summary has a `linked_secret` warning.
5. **Workspace folder**: pick `…/postman/v3/workspace/` (the repo-like parent of `postman/`).
   - The preview says "Postman workspace v3" and lists 2 collections, 2 environments and Globals.
   - "Not Imported" lists `flows` and `specs`.
   - Assign the items to two workspaces and import. Everything lands where it was assigned.
6. **MCP**: with AI access on, call the `import` tool (`detect`, then `run`) on the same paths. Kinds, items and
   warnings match the dialog.

## Regenerating fixtures

```bash
postman collection migrate test/integration/resources/import/postman/v3/rich-v2.1.json -o /tmp/rich
# then re-apply the hand-added cases listed in test/integration/resources/import/README.md (v3 section)
```
