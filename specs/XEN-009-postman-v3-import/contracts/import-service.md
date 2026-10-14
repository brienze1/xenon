# Contract: Import service, bindings and MCP tool for v3 sources

There are no new bindings or tools. `ImportService.Detect / Clashes / Import / Discard` (Wails) and the MCP `import`
tool (`detect` / `clashes` / `run` / `discard`) keep their request and response DTOs. This contract fixes what they
return for v3 sources.

## Detect

| Source | `kind` | `version` | `items[]` |
|---|---|---|---|
| v3 collection JSON (file or pasted) | `postman_collection` | `3` | 1 × `{kind: "collection", name, counts}` |
| v3 collection folder or zip (incl. `postman/v3` fixture) | `postman_collection` | `3` | 1 × collection |
| `*.request.yaml` | `postman_request` | `3` | 1 × `{kind: "request"}` |
| `*.environment.yaml`, environment JSON | `postman_environment` | `3` | 1 × `{kind: "environment"}` |
| `*.globals.yaml`, globals JSON | `postman_globals` | `3` | 1 × `{kind: "globals"}` |
| workspace folder or zip (`postman/`, its parent, or `postman/collections`) | `postman_workspace` | `3` | collections (sorted by folder name), then environments (sorted by file name), then globals |

`warnings[]` contains the R8 warnings with `path` set, plus the skipped parts of a workspace (`code: "skipped_part"`,
`path: <part name>`).

No secret value appears in any Detect field. Variable counts include secrets. `encrypted_secrets` is `false`.

## Clashes and Import

These are unchanged from other formats:

- `clash: keep_both | replace | skip` per item;
- `IMP-04409` when a clash is unresolved;
- a `postman_request` item with no parent placement fails with `IMP-35400` "A request must be imported into a
  collection or folder";
- workspace items each take their own `target_local_id` (the dump assignment).

## Errors

| Code | When | `error.description` |
|---|---|---|
| `IMP-11400` | a v3 document in the source fails to decode, or has an unknown `$kind` that does not end in `-request` | `<relative path> is not a valid Postman v3 file: <reason>`. `<reason>` is `line <n>: <error class>` or `unknown $kind "<k>"`, and never quotes file content (research R9). |
| `IMP-00415` | unchanged generic case | `<name> is not a Xenon, Postman, Bruno or cURL export` |
| (removed) | — | `Postman v3 YAML is not supported yet; export the collection as v2.1 JSON` no longer exists |

## MCP `import` tool description (text change only)

`Import Postman (v2.0/v2.1 JSON, v3 JSON or git-native folders, environments, globals, data dumps, workspace
folders), Bruno or Xenon files from paths, or a pasted cURL command.`

## UI

- Source label: `sourceLabel(kind, version)`, for example "Postman collection v3", "Postman workspace v3" or
  "Postman request v3".
- A `postman_workspace` preview shows the per-item workspace assignment table and the "Not Imported" list of skipped
  parts, exactly like `postman_dump`.
