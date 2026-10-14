# Data Model: Import Postman v3 (XEN-009)

No workspace file format changes. Imported v3 items are written exactly as imported v2.1 items are: the same
collection, folder, request, example and environment YAML, and the same `postman:` passthrough blocks. The changes
are in the import pipeline's in-memory model.

## Domain (`internal/domain/`)

### `enums/importkind` (CHANGE)

| Value | New? | Meaning |
|---|---|---|
| `postman_request` | new | A single Postman v3 `*.request.yaml` (version `3`) |
| `postman_workspace` | new | A `postman workspace pull` folder (version `3`) |
| `postman_collection`, `postman_environment`, `postman_globals` | existing | Also carry version `3` for v3 sources |

### `enums/warningcode` (CHANGE)

| Value | New? | Severity | Used for |
|---|---|---|---|
| `protocol_script` | new | warning | Collection or folder `graphql:*` / `grpc:*` scripts (research R8) |
| `linked_secret` | new | warning | A variable linked to a Postman vault secret (R8) |

The `unsupported_request_type`, `unsupported_auth` and `skipped_part` codes are reused with the R8 messages.

### `entity.ImportProbe` (CHANGE)

New fields, filled by the persistence probe:

| Field | Type | Source |
|---|---|---|
| `PostmanKind` | string | top-level `$kind` of a JSON or YAML document (`collection`, `http-request`, …) |
| `HasValues` | bool | top-level `values` array present |
| `HasWorkspaceField` | bool | top-level `workspace` field (cloud globals JSON) |
| `DocumentID` | string | top-level `id` (local `globals get --json` gives the globals file path) |

New predicates. Each is pure, with no I/O:

| Predicate | True when |
|---|---|
| `IsPostmanV3Collection()` | (`IsJson` or YAML) and no v2 schema and `PostmanKind == "collection"` |
| `IsPostmanV3Request()` | file name ends with `.request.yaml` or `PostmanKind == "http-request"` |
| `IsPostmanV3Variables()` | `Name`-bearing document with `HasValues` and empty `Scope` |
| `IsPostmanV3Globals()` | R5 rule: file name suffix `.globals.(yaml|yml|json)`, `HasWorkspaceField`, or `DocumentID` ending with `.globals.yaml` |
| `IsPostmanV3Workspace()` | folder rules R4.1 over `Entries` and `Name` |
| `IsPostmanV3CollectionFolder()` | folder with a `*.request.yaml` or `.resources/definition.yaml` entry (replaces `HasPostmanV3Entry`) |

`IsPostmanV3File` and `HasPostmanV3Entry` are renamed or replaced by these predicates.

### `entity.ImportInput` (CHANGE)

- `IsPostmanV3() bool` returns `Version == "3"`.

### `entity.ImportCollection` (NEW)

The result of reading one Postman collection, whether v2.x or v3:

| Field | Type | Notes |
|---|---|---|
| `Root` | `ImportNode` | collection tree |
| `Warnings` | `[]ImportWarning` | conversion warnings |
| `Secrets` | `map[string]string` | secret collection-variable values by key (R7); never logged or serialised to the UI |

### `entity.PostmanV3Source` (NEW)

What the v3 reader returns for any v3 input:

| Field | Type | Notes |
|---|---|---|
| `Collections` | `[]ImportCollection` | one per collection (JSON, folder, or each folder under a workspace's `collections/`) |
| `Request` | `*ImportCollection` | set only for `postman_request`: `Root` is a single request node |
| `Environments` | `[]ImportVariables` | |
| `Globals` | `*ImportVariables` | |
| `Warnings` | `[]ImportWarning` | source-level warnings (skipped parts) |
| `Skipped` | `[]PostmanDumpPart` | workspace parts not imported (R8 message via `SkippedWarnings`) |

`SkippedWarnings()` follows the `PostmanDump.SkippedWarnings` pattern but uses the workspace wording from R8.

### `entity.VariableRow` (no change)

`Secret` already exists. Rows from v3 `secret: true` / `type: secret` arrive with `Secret: true` and an empty
`InitialValue`.

## Application (`internal/application/`)

| Element | Change |
|---|---|
| `usecase.ReadPostmanCollection` | `ReadCollection(ctx, ImportDocument) (*entity.ImportCollection, error)`. The return type changes to carry `Secrets`. |
| `usecase.ReadPostmanV3` (NEW port) | `ReadPostmanV3(ctx, entity.ImportInput) (*entity.PostmanV3Source, error)` |
| `usecase.DetectImportFormat` | Implements R4. Removes `postmanV3Msg`. |
| `service.importReaders` | New field `readPostmanV3`. `read()` routes `input.IsPostmanV3()` → `postmanV3()`. That adds each collection as a tree with secrets, the request as an `importitemkind.Request` tree, environments as `Environment` items, globals as a `Globals` item, and the skipped-part warnings. The `collection()` helper sets `CollectionSecrets` from `ImportCollection.Secrets`. |
| `service.importRead` | `addTree` gains a secrets-aware variant (`addCollection(kind, ImportCollection)`) that sets `ImportItem.CollectionSecrets`. |

Clash handling, targets, `RejectTopLevelRequest`, planning and writing are unchanged and format-agnostic.

## Integration (`internal/integration/`)

| Element | Change |
|---|---|
| `adapter.PostmanWebservice` | Embeds `usecase.ReadPostmanV3`. |
| `webservice/postman_v3.go` (NEW) | `(*postmanWebservice) ReadPostmanV3`. Logs debug Started/Finished with counts only. Dispatches on `input.Kind`. Builds v2.1 documents through the v3 DTOs, then reuses `p.ReadCollection` / `p.ReadEnvironment` logic. |
| `webservice/postmanv3/dto/` (NEW) | v3 contract DTOs: `Collection`, `Item` (folder or request by `$kind`), `Request`, `Example`, `Auth` (single or list), `Credentials` (list or map), `KeyValues` (list or map), `Body`, `Script`, `Variable`, `Environment`, `Definition`, `Tree` (assembles a folder's documents into a nested `Collection`, applying R10 order and R11 names), `Conversion` (v3-only warnings, R8), and `ToPostmanV21()` builders that return v2.1 documents (`map[string]any`, marshalled once). |
| `webservice/postman/dto/collection.go`, `key_value.go` | `type: secret` collection variables become secret rows with values collected (R7). |
| `persistence/import_source.go` | `.resources` is allowed in `isImportDocument`. Re-roots to `postman/` (R4). A single `*.request.yaml` also reads its sibling examples (R12). The YAML probe fills the new probe fields. |
| `persistence/model/import_probe.go` | Adds `Kind` (`$kind`), `Values`, `Workspace`, `ID` with both `json` and `yaml` tags. `ToEntity` maps them. |
| `entrypoint/controller/mcp_import.go` | Description text only. |

## Infrastructure (`internal/infra/`)

| Element | Change |
|---|---|
| `dependency/application/manage_import.go` | Passes `GetPostmanWebservice()` as the `ReadPostmanV3`. |

## Frontend (`frontend/src/features/import-export/`)

| Element | Change |
|---|---|
| `lib/labels.ts` | Adds `postman_request: 'Postman request'` and `postman_workspace: 'Postman workspace'`. |
| `components/ImportPreviewStep.tsx` | `DUMP_KIND` becomes `ASSIGNMENT_KINDS = new Set(['postman_dump', 'postman_workspace'])`. |

## State and lifecycle

These are unchanged: detect → preview (`ImportSource` held in memory with expiry) → clashes → import (atomic per
item) → summary.
