# Research: Import Postman v3 (XEN-009)

All facts below were checked on 2026-10-14. Code was read on `origin/main` `6f61014`. The v3 shapes are
real Postman CLI 1.73.0 output:

- 12 collections fetched with `postman collection get --json`; only key names and types were inspected, values were
  redacted.
- v2.1 fixtures converted with `postman collection migrate`.
- Environments and globals written with `postman environment new/var set` and `postman globals get`.

No `NEEDS CLARIFICATION` remains.

## R1. The v3 shapes Xenon must read

**Decision**: Support exactly these shapes. Every key named here is mapped; any other key is kept for export
(passthrough, R6).

| Shape | Where it comes from | Structure (verified) |
|---|---|---|
| Collection JSON | `postman collection get <id> --json` | `{id, name, $kind:"collection", description?, variables?, auth?, scripts, items[]}`. `items[]` mixes folders `{$kind:"collection", id, name, description?, auth?, scripts, variables?, items[]}` and requests `{$kind:"http-request", id, name, url, method, headers[], queryParams[], pathVariables[], body?, auth?, scripts[], settings{}, description?, examples[]}`. Examples are inline: `{$kind:"http-example", id, name, request{url, method, headers[], queryParams[], body?}, response{statusCode, statusText, headers[], body?}}`. JSON has no `order`; array order is the order. `queryParams[].value` may be `null`. |
| Collection folder | `postman workspace pull`, `postman collection migrate` | `<collection>/.resources/definition.yaml` (`$kind: collection`, `id?`, `name?`, `description?`, `variables?`, `auth?`, `scripts?`, `order?`). Each subfolder may have `.resources/definition.yaml`. Requests are `*.request.yaml` (`$kind: http-request`, `name?`, `order`, the same request keys, `examples: ./.resources/<request>.resources/examples`). Examples are `.resources/<request>.resources/examples/*.example.yaml` (`$kind: http-example`, `name?`, `request`, `response`, `order`). |
| Environment | `postman/environments/<name>.environment.yaml`; `postman environment get <id> --json` | `{name, values:[{key, value, enabled, type?, secret?, source?}]}`. Cloud JSON adds `id, owner, color, createdAt, updatedAt, lastRevision, lastUpdatedBy, team`. |
| Globals | `postman/globals/workspace.globals.yaml`; `postman globals get [-w id] --json` | The same `{name, values}`. Cloud JSON adds `workspace, id, createdAt, …`. Local JSON adds `id: "postman/globals/workspace.globals.yaml", path`. |
| Workspace folder | `postman workspace pull` | `postman/collections/<c>/…`, `postman/environments/*.environment.yaml`, `postman/globals/workspace.globals.yaml`, and other parts (`specs/`, `flows/`, `mocks/`, …). `.postman/resources.yaml` sits next to `postman/` at the repo root. |

**Shape variants seen in real output** (each must be read):

- `headers`, `queryParams`, `pathVariables`, `auth.credentials` and example `response.headers` appear **both** as a list of `{key, value, …}` and
  as a YAML map `{Content-Type: application/json}`. `collection migrate` writes maps; `collection get --json` writes
  lists.
- Collection and folder `auth` is a **list** `[{id, name, type, credentials}]` (multi-auth shape). The CLI writes
  every collection or folder auth this way. Request auth is a single object `{type, credentials}`. `noauth` has no
  credentials.
- Body `content` is a string for raw types; a list for `formdata` (`{key, type: text|file, value | src, disabled?,
  contentType?, description?}`) and `urlencoded`; `{query, variables}` for `graphql`; `{src}` for `file`.
- Script `type` is `beforeRequest`/`afterResponse` on requests and `http:beforeRequest`/`http:afterResponse` on
  collections and folders.
- Variables: `{key, value, disabled?, secret?, description?}`.

**Alternatives considered**: reading only the documented schema in the `collection-schema-v3` skill. Rejected,
because real output differs from it: map headers, list auth, inline JSON examples.

## R2. Translate v3 into the existing v2.1 DTOs instead of a second mapper

**Decision**: The new v3 DTOs in `internal/integration/webservice/postmanv3/dto/` do two things:

1. Decode v3 (JSON or YAML).
2. **Build the equivalent Postman v2.1 document**: `info{_postman_id, name, description, schema v2.1}`, `item`,
   `auth`, `event`, `variable`, `request{url{raw, query, variable}, header, body, auth, description}`,
   `protocolProfileBehavior`, `response[]`.

The v3 webservice method marshals that document and runs it through the **existing** v2.1 path
(`dto.Collection.ToEntity(dto.NewConversion(...))`). Environments and globals become v2.1 environment JSON
(`{name, values, _postman_variable_scope}`) and go through the existing `dto.Environment.ToEntity`.

**Rationale**:

- Fidelity equal to v2.1 by construction. Acceptance says "same tree, requests, scripts, auth and variables as the
  v2.1 export of the same collection".
- The same warnings, missing-file checks, auth type mapping, settings mapping and export passthrough apply with no
  duplicated logic.
- When v2.1 mapping improves later, v3 gets the improvement for free.

**Alternatives considered**: a second v3 → entity mapper. Rejected: it doubles every mapping rule and drifts from
v2.1.

## R3. Kinds and versions

**Decision**:

| Source | kind | version |
|---|---|---|
| v3 collection JSON or folder | `postman_collection` (existing) | `3` |
| single `*.request.yaml` | `postman_request` (**new**) | `3` |
| v3 environment (YAML/JSON) | `postman_environment` (existing) | `3` |
| v3 globals (YAML/JSON) | `postman_globals` (existing) | `3` |
| workspace folder | `postman_workspace` (**new**) | `3` |

`ImportInput.IsPostmanV3()` (version `3`) routes these sources to the v3 reader in `service/import_readers.go`. The
UI label stays `sourceLabel(kind, version)` (for example, "Postman collection v3").

**Rationale**:

- The kind tells the UI and MCP clients *what* the source is; the version tells *which* format it is in. v2.0 and
  v2.1 already share `postman_collection` this way.
- Only the two genuinely new shapes get new kinds.

**Alternatives considered**: `postman_v3_*` kinds for everything. Rejected, because it duplicates labels and splits
one concept in two.

## R4. Detection rules (domain `ImportProbe`, use case `DetectImportFormat`)

**Decision**: These rules are evaluated after the existing Xenon, dump and Bruno checks, so FR-007 holds.

1. **Folder → `postman_workspace` v3** when any of these is true:
   - the top level has `postman/collections/`, `postman/environments/` or `postman/globals/` entries;
   - the top level has `collections/<c>/` holding v3 entries, or `environments/*.environment.yaml`, or
     `globals/*.globals.yaml`;
   - the folder itself is named `collections` and its subfolders hold v3 entries. This covers picking
     `postman/collections`.
2. **Folder → `postman_collection` v3** when it holds any `*.request.yaml` or `.resources/definition.yaml`. This
   replaces the `HasPostmanV3Entry` → IMP-00415 branch.
3. **File `*.request.yaml` → `postman_request` v3**. This replaces the `IsPostmanV3File` → IMP-00415 branch.
4. **JSON (file or pasted)** with no v2 `info.schema`:
   - `$kind == "collection"` with `items` → `postman_collection` v3;
   - top-level `name` + `values[]` and no `_postman_variable_scope` → v3 environment or globals (R5).
5. **YAML file** (`.yaml`/`.yml`, not a Xenon file): the probe now also parses YAML for `$kind`, `name`, `values`
   and `workspace`.
   - `$kind: collection` → `postman_collection` v3. This is a lone `definition.yaml`, which is imported as an empty
     collection.
   - `$kind: http-request` → `postman_request` v3.
   - `name` + `values[]` → environment or globals (R5).
   - Anything else keeps today's "not a Xenon, Postman, Bruno or cURL export" error. `misc/notes.yaml` is unchanged.
   - A YAML file that fails to parse during probing also keeps that error, unless its name ends with
     `.request.yaml`. Then rule 3 detects it by name, and the v3 reader raises `IMP-11400` for the parse error.

`postmanV3Msg` and its two IMP-00415 call sites are deleted. `IsMalformed` handling of JSON is unchanged. A YAML
file that fails to parse during probing falls through to the existing unsupported error. The v3 reader reports
malformed YAML inside a v3 folder (R9).

**Workspace folder re-rooting**: when the picked folder (or unpacked zip) has `postman/` with collections,
environments or globals under it, the v3 reader treats `postman/` as the workspace root. Only `postman/`'s other
top-level parts are reported as skipped. The rest of the repository is neither imported nor reported. To avoid
reading a whole repository's JSON and YAML files (and tripping the per-file size limit on unrelated files),
`readImportFolder` itself re-roots: if `<root>/postman/collections`, `/environments` or `/globals` exists, it walks
only `<root>/postman`. The probe name stays the picked folder's name.

## R5. Environment vs globals for v3

**Decision**: A v3 variables document is **globals** if any of these is true:

- its file name ends with `.globals.yaml`, `.globals.yml` or `.globals.json`;
- it sits in a `globals/` folder;
- it has a top-level `workspace` field (cloud `globals get --json`);
- its `id` ends with `.globals.yaml` (local `globals get --json`).

Otherwise it is an **environment**.

**Rationale**: These are the only signals the real outputs carry. Environments never have `workspace`, and their
local files are `*.environment.yaml`.

## R6. Passthrough for export

**Decision**: v3 keys with no v2.1 equivalent go into the generated v2.1 document at the matching level under their
v3 key names. The existing `Raw` passthrough then keeps them. Examples:

- request `settings.maxRedirects`, `protocolVersion` and `disabledSystemHeaders` go to `protocolProfileBehavior`,
  which is the same key in v2.1;
- variable `source` stays `source`;
- multi-auth `rules` and extra entries are kept under a `postmanV3` key on the collection or folder.

Ids map to `_postman_id` (collection) and `id` (folder, request, example).

**Rationale**: FR-018 and constitution VI (round trip). Export stays v2.1 (out of scope: v3 export).

## R7. Secrets

**Decision**:

- **Environment and globals**:
  - `secret: true` or `type: secret` → v2.1 `type: "secret"`, then the existing `EnvironmentValue.ToEntity` →
    `WithoutSecretValue` → vault through `import_variables_writer`.
  - The CLI mask `********` on a secret → value `""` (secret with no value → `secret_state: missing`).
  - A vault link (`source`, no value) → value `""`, with `source` kept in passthrough and one `linked_secret` warning.
- **Collection variables**: today the v2.1 `KeyValue.ToVariableRow` ignores `type: secret`, so a secret collection
  variable's value would reach the collection file. That violates constitution II, and v3 `secret: true` needs it
  fixed. Decision:
  - `dto.Collection.ToEntity` marks `type: secret` rows `Secret: true`, blanks their `InitialValue`, and collects
    `key → value` for non-empty values.
  - `ReadPostmanCollection` returns them (`entity.ImportCollection.Secrets`).
  - `importRead.addTree` sets `ImportItem.CollectionSecrets`.
  - The existing `importWriter.keepCollectionSecrets` writes them to the vault (scope = the new collection id, or the
    target collection when imported into a parent).
  - This also fixes v2.1 collections with `type: secret` variables. No existing fixture or scenario has one (checked
    with grep), so no existing scenario changes.
- The v3 translator never logs values. Its debug logs carry counts only, as `postman.go` does today.

## R8. Warnings for v3-only cases

**Decision**: Reuse existing codes where they fit, and add two new codes.

| Case | Code | Message (exact) |
|---|---|---|
| Non-HTTP request (`grpc-request`, `websocket-request`, `socket.io-request`, `mqtt-request`, `mcp-request`, `llm-request`) | `unsupported_request_type` (existing) | `<name> is a <kind> request, which Xenon does not support; it was skipped` |
| Extra multi-auth entries | `unsupported_auth` (existing) | `Postman multi-auth on <name> lists <n> auths; Xenon imported the first (<first type>) and kept <other types> for export only` |
| Collection or folder scripts of types `graphql:*` or `grpc:*` | `protocol_script` (**new**) | `The <type> script on <name> runs only for <GraphQL|gRPC> requests in Postman; it is kept for export and Xenon does not run it` |
| Vault-linked secret | `linked_secret` (**new**) | `<key> is linked to a Postman vault secret; enter its value in Xenon` |
| Workspace part not imported | `skipped_part` (existing) | `The workspace's <part> part (<n> file|files) is not a collection, an environment or globals and was skipped` |

Warnings raised by the v3 translator carry the import path (`entity.JoinImportPath`) and an empty item id, as dump
skipped parts do. Warnings raised by the reused v2.1 conversion keep their item ids.

## R9. Errors

**Decision**:

- A v3 document that fails to decode → `IMP-11400` `"<path> is not a valid Postman v3 file: <yaml/json error>"`
  (BadRequest), raised at its origin in the v3 webservice.
- An item with an unknown `$kind` → the same error with `unknown $kind "<k>"`, unless the kind ends in `-request`.
  Those are skipped with the R8 `unsupported_request_type` warning, so protocols Postman adds later do not break
  imports.
- **The reason never quotes content.** It is exactly one of:
  - `line <n>: a value has the wrong type` for a yaml.v3 `TypeError` (its first error's line);
  - `line <n>: <message>` for a yaml.v3 syntax error, with the `yaml: line <n>: ` prefix stripped (syntax messages
    never quote values);
  - `byte <offset>: invalid JSON` for a `json.SyntaxError`;
  - `a value has the wrong type` for a JSON document whose value has the wrong type (JSON is re-encoded as YAML before decoding, so its line numbers would be meaningless);
  - `unknown $kind "<k>"`.

  This keeps secret values out of error text, logs and the MCP output (constitution II).
- Code range `IMP-11xxx` is unused today (checked with grep).
- `IMP-00415` remains only for the generic "not a … export" and mixed-Bruno cases.

## R10. Ordering

**Decision**:

- **JSON**: array order.
- **YAML folders**: children (folders by their definition `order`, requests and examples by their `order`) sort
  ascending by `order`. Those without `order` come after. Ties and the unordered group sort by file or folder name
  with `strings.ToLower` and then byte order, so the result is stable on every OS (constitution VII).

**Alternatives considered**: directory listing order. Rejected, because it differs between OSes.

## R11. Names

**Decision**:

| Item | Name rule |
|---|---|
| Collection | definition `name`, else folder name |
| Folder | definition `name`, else folder name |
| Request | `name`, else file stem without `.request.yaml` |
| Example | `name`, else stem without `.example.yaml` |
| Environment | `name`, else stem without `.environment.yaml` |

## R12. Reading `.resources/`

**Decision**:

- `persistence/import_source.go` `isImportDocument` skips dot-segments **except** `.resources`. This lets v3
  definitions and examples into `Documents`.
- No other format uses `.resources`. Bruno and Xenon folders were checked: their fixtures and writers never create it.
- A single `*.request.yaml` file also reads its sibling examples
  `<dir>/.resources/<stem>.resources/examples/*.example.yaml` as extra documents.

## R13. UI and MCP

**Decision**:

- **Frontend**:
  - `postman_workspace` joins `postman_dump` as a multi-target kind in `ImportPreviewStep`. `DUMP_KIND` becomes a
    set `ASSIGNMENT_KINDS`, so the existing per-item workspace assignment (`DumpAssignment`) is reused unchanged.
  - Labels: `postman_workspace: 'Postman workspace'` and `postman_request: 'Postman request'`.
  - `itemKindLabel` already covers `globals`.
- **MCP**: the `import` tool already calls `ManageImportService`. Only its description text changes, to mention
  "Postman v2.1/v3". No code path changes.

## R14. Fixtures

**Decision**: Fixtures are generated, never hand-copied from a real workspace. All values are fake
(`s3cr3t-fixture` style).

- **`postman/v3/rich-v2.1.json`** is the ground truth: a v2.1 collection covering every FR-008–FR-015 field.
  - It is written in the **canonical v2.1 form the translator emits** (contracts/v3-to-v21-mapping.md):
    - every folder, request and example has an `id`;
    - `request.url` is `{raw, query, variable}`;
    - auth rows carry `type: "string"`;
    - examples carry `_postman_previewlanguage`;
    - events have no ids.
  - That form lets the equality check below hold byte for byte.
  - It holds **no v3-only case**, because those cannot exist in v2.1.
- **`postman/v3/rich/`** is the output of `postman collection migrate rich-v2.1.json`, unchanged. Verified on
  2026-10-14: `migrate` keeps item and example ids.
- **`postman/v3/rich.json`** is the same collection in `collection get --json` shape. It is built from `rich/` by a
  small documented script in the fixtures README: nested `items`, inline `examples`, list-form headers and
  credentials, and no `order`.
- **SC-001** is asserted with the existing step `the folder "<a>" should equal the folder "<b>" apart from ids` across
  the three imports of `rich`. That step compares bytes after replacing ULIDs.
- **`postman/v3/extras/`** is hand-written and holds every v3-only case:
  - map vs list headers and credentials;
  - a multi-auth list with 2 entries;
  - a `graphql:beforeQuery` collection script;
  - a `graphql-request`, a `grpc-request` and a `websocket-request`;
  - a missing and duplicate `order`;
  - a name that differs from the file name;
  - folder variables;
  - `settings.maxRedirects`;
  - a `secret: true` collection variable.

  These are asserted field by field.
- **`postman/v3/workspace/`** is a repo-like folder: `postman/collections/{Billing,Users}`,
  `postman/environments/{staging,prod}.environment.yaml`, `postman/globals/workspace.globals.yaml`,
  `postman/specs/…`, `postman/flows/…`, `.postman/resources.yaml`, plus `tools/latin1.json`, a non-UTF-8 file at the repo root. Any file the walk reads is UTF-8 checked, so reading it would fail with IMP-01400; that proves only `postman/` is read.
- **`postman/v3/env/`** holds environment and globals variants:
  - YAML with `type: secret`, `secret: true` and a vault `source`;
  - CLI JSON with `********` and cloud fields;
  - cloud globals JSON with `workspace`;
  - local globals JSON with `id: postman/globals/workspace.globals.yaml`.
- **`postman/v3/large-1000-v3.zip`** is `postman collection migrate large-1000.json`, zipped (1,011 files, about
  4 MB unpacked).
- **`postman/v3/malformed/`** is a collection folder with one broken `*.request.yaml`.
- The existing `Get users.request.yaml` and `.resources/definition.yaml` move to `postman/v3/native-git/`, because
  `postman/v3/` now holds every v3 fixture. The format-detection rows use that path.
