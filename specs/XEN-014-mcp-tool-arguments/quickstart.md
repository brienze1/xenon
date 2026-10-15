# Quickstart: validating XEN-014

## Prerequisites

- A fresh worktree on `xen-014` with `npm ci` (in `frontend/`) and `make bindings` done.
- Go toolchain per `go.mod`.

## Fixtures the BDD suite adds

`test/integration/resources/import/loose/`:
- `mixed/`:
  - `collection.json` (copy of `postman/full-v2.1.json`);
  - `dev.postman_environment.json` (copy of `postman/dev.postman_environment.json`);
  - `globals.json` (copy of `postman/globals.json`).
- `collections-only/`:
  - `a.json` and `b.json`, two v2.1 collections with distinct names (copies of
    `postman/full-v2.1.json` with `info.name` changed).
- `with-extras/`: everything in `mixed/` plus `notes.txt`, a hidden `.hidden-copy.json` collection (`.DS_Store` is git-ignored) and `sub/readme.md`.
- `nothing/`: `notes.txt` and `sub/readme.md`.

## Run

```bash
make bdd TAG=@xen_014          # the new scenarios
make bdd                       # everything, must stay green
go test ./pkg/jsonschema/... ./internal/integration/entrypoint/... ./internal/application/usecase/... ./internal/domain/...
(cd frontend && npx vitest run src/features/import-export/lib/labels.test.ts)
make validate                  # final gate
```

## Expected outcomes

1. **Schema** (`I list the MCP tools`):
   - `workspace_edit` has `inputSchema.properties.path.type = string`, and its description
     contains `Required for create and open.`
   - `import` has `inputSchema.properties.paths.type = array` and
     `inputSchema.properties.paths.items.type = string`.
   - `tree` has `inputSchema.properties.depth.type = ["integer","string"]`.
   - No tool has `inputSchema.oneOf`. See [contracts/mcp-tool-schema.md](contracts/mcp-tool-schema.md).
2. **Array paths**: `import {action: run, paths: ["<resources>/import/postman/full-v2.1.json"]}`
   succeeds, and the collection appears in `tree list`.
3. **Errors**: `workspace_edit {action: create, name: "Payments"}` fails with `WKS-00400`, the
   message `invalid request: path is required`, and `details.0 = {field: path, problem: is required}`.
   `WorkspaceService.Create` with a relative path gives
   `error.error_details.0.messages.0 = must be an absolute path`. See
   [contracts/mcp-error.md](contracts/mcp-error.md).
4. **Loose folder**:
   - `ImportService.Detect {paths: [<loose>/mixed]}` returns 3 items, no warnings and kind `mixed`.
   - `with-extras` gives 3 items and one `skipped_part` warning naming `notes.txt` and `sub`
     (not the hidden `.hidden-copy.json`).
   - `nothing` fails with `IMP-55415` and the layout message.
   - `collections-only` reports kind `postman_collection` / `2.1`.

   See [contracts/import-detect.md](contracts/import-detect.md).
5. **Real client check (manual, optional)**: with Xenon running and AI access on, ask Claude
   Code to "import <folder of Postman exports> into the active workspace". It calls `import`
   with `paths` as an array and succeeds on the first try.
