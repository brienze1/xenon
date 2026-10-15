# Research: XEN-014

Every decision below was settled from the code at `06d0d82` or by the user (see spec
Clarifications). Nothing is open.

## R1 — Where the argument contract comes from

- **Decision**: The argument contract comes from the existing MCP argument DTOs
  (`internal/integration/entrypoint/dto/mcp_*.go`, the types each controller passes to
  `call.Decode`).
  - Struct tags on those DTOs carry the description, the actions that require the field, closed
    enums and type overrides.
  - A reflector builds the schema from those tags.
  - Each controller's `Tool()` registers the DTO type(s) its actions decode.
- **Rationale**: The DTO is what the handler decodes, so a renamed or retyped field changes the
  schema with it (FR-005).
- **Alternatives considered**:
  - A hand-written schema per tool: drifts, which is the bug.
  - `github.com/google/jsonschema-go`'s `For[T]`: an indirect dependency today. It uses its own
    `jsonschema:"…"` tag, emits `["null", T]` unions for pointers, and offers no hook for
    per-action requirements. A 150-line stdlib reflector in `pkg/jsonschema` is simpler and
    exactly fits the contract.

## R2 — One DTO per action, several per tool

- **Fact**: 8 tools decode more than one DTO type (`console`, `cookies_edit`, `env`, `env_edit`, `export`, `git_edit`, `history`, `tree`). For example:
  - `env_edit` decodes `MCPEnvEdit`, `MCPEnvActivation`, `MCPEnvVariable` and `MCPEnvVariables`;
  - `tree` decodes `MCPTreeList`, `MCPTreeSearch`, `MCPTreeTokenOwner` and `MCPTreeItem`;
  - `export` decodes `MCPExport` and `MCPSaveResponseBody`, the latter in
    `mcp_export_response_body.go`.

  Several actions decode nothing (`workspace list`, `git status`, …).
- **Decision**: `dto.NewMCPToolArguments[T](actions ...string)` registers the actions that
  decode `T`, and `Tool()` passes one registration per DTO type.
  - The tool's `properties` are the union of all registered DTOs' fields.
  - A field present in several registrations is declared once, and its "Used by" list is the
    union of their actions.
  - The registrations must agree on the field's type; a disagreement fails the completeness
    unit test (R6).
- **Rationale**: This mirrors the code exactly. One field can serve several actions (`id`) with
  different requirements.

## R3 — Per-action requirements cannot be conditional schema

- **Fact**: Claude's tool-use API rejects `oneOf`, `anyOf` and `allOf` at the top level of
  `input_schema`, and MCP clients pass the schema through to the model API.
- **Decision**:
  - Top-level `required` holds the action selector (when the tool has more than one action) plus
    any field tagged `required_for:"always"`.
  - Per-action requirements are appended to the field's description: `Required for create and
    open.` (FR-003, FR-006).
- **Alternatives considered**: `if/then` and `oneOf` keyed on `action` were rejected because
  clients refuse them.

## R4 — Types the reflector must handle

- **Fact**: The argument DTO fields are:
  - `string`, `bool`, `int`, `*int`, `[]string`;
  - nested structs and pointers to structs (`*Body`, `*Auth`, `*Scripts`,
    `*ExampleResponseEdit` in `MCPTreeEdit`);
  - slices of structs (`MCPEnvVariables` rows);
  - `map[string]string`;
  - two `json.RawMessage` fields: `MCPTreeList.depth`, an integer ≥ 1 or `"all"`, and
    `MCPSettingsSave.ai_access`.
- **Decision**:

  | Go field type | Schema |
  |---|---|
  | `string`, `bool` | `string`, `boolean` |
  | Any int kind | `integer` |
  | Any float kind | `number` |
  | Pointer | The pointee's schema |
  | Slice | `array` with `items` |
  | Struct | `object` with nested `properties` |
  | Map | `object` with `additionalProperties` |

  More rules:
  - `json:"-"` fields are skipped.
  - `json.RawMessage` requires a `schema:"…"` tag naming its JSON types (`depth` →
    `schema:"integer|string"`, which emits `"type": ["integer","string"]`). Without the tag the
    completeness test fails.
  - `ai_access` is tagged `schema:"-"` and not declared. It exists only so `settings_edit` can
    refuse it (`RejectHumanOnlyAIAction`), and declaring it would invite clients to send it.
    This is the only field excluded from every tool. (`schema_for` only scopes a field of a DTO
    shared by two tools to the tool that reads it; see data-model.md Appendix A, note 1.) FR-001
    covers arguments an action *uses*; a field that is read only to refuse it is not one.
  - Recursion is guarded by type: a type already on the current path becomes a plain `object`.

## R5 — Description wording

- **Decision**:
  - Each top-level field's `description:"…"` tag is one sentence.
  - For tools with more than one action, `InputSchema()` appends ` Used by <a>, <b> and <c>.`
    and, when tagged, ` Required for <x> and <y>.`
  - Nested fields keep their own `description` tag when present; it is optional for them.
- **Example**: `workspace_edit.path` → `Absolute path of the workspace folder. Used by create
  and open. Required for create and open.`

## R6 — Keeping the registration honest (SC-001)

- **Decision**: `controller/mcp_tool_arguments_test.go` does two things.
  1. **AST drift check.** It parses every `controller/mcp_*.go`, groups methods by receiver type,
     collects each `var arguments dto.X` / `dto.X{}` decoded in that receiver's methods and every
     `dto.NewMCPToolArguments[dto.Y]` in its `Tool()`, and fails if any decoded `X` is not a
     registered `Y`.
  2. **Schema completeness check.** It builds every tool controller (nil services are fine,
     because only `Tool()` is called). For each tool, `InputSchema()` must:
     - build with no type conflict;
     - give every top-level property a type and a non-empty description;
     - declare no `oneOf`/`anyOf`/`allOf`;
     - leave no `json.RawMessage` field without a `schema` tag.
- **Rationale**: The two checks together make "a decoded argument missing from the schema" fail
  CI.

## R7 — Validation wording and field paths

- **Fact**:
  - `validator.CustomValidator().Struct` already builds one `errs.ErrorDetails` per failing field.
  - `Attribute` is `fieldError.Field()`, the leaf JSON name.
  - Messages are `failed on the '<tag>' rule`.
  - The description is the fixed `invalid request`.
  - Tags in use: `required`, `required_if`, `required_unless`, `not_blank`, `oneof`, `max`,
    `min`, `gt`, `gte`, `email`, `absolute_path`, `has_id`, `variable_scope`, `import_input`,
    `safe_subfolder`; `dive` and `omitempty` are modifiers.
- **Decision**:
  - `Attribute` becomes the JSON path from the root: `fieldError.Namespace()` without its first
    segment, the struct name, e.g. `variables[2].name`. Top-level fields stay as they are
    (`path`), so every existing `error_details.0.attribute` assertion still holds.
  - `Messages[0]` becomes the problem. The wording table lives in `validator/problem.go`:

    | Rule | Problem |
    |---|---|
    | `required` | `is required` |
    | `required_if`, `required_unless` | `is required here` |
    | `not_blank` | `must not be blank` |
    | `oneof` | `must be one of <param, comma-separated>` |
    | `max` (string) / `max` (slice) | `must be at most <n> characters` / `must have at most <n> items` |
    | `min` (string) / `min` (slice) | `must be at least <n> characters` / `must have at least <n> items` |
    | `gt` | `must be greater than <n>` |
    | `gte` | `must be at least <n>` |
    | `email` | `must be an email address` |
    | `absolute_path` | `must be an absolute path` |
    | `has_id` | `must have an id` |
    | `variable_scope` | `must be a variable scope (globals, env:<id> or collection:<id>)` |
    | `import_input` | `needs at least one path or some pasted text` |
    | `safe_subfolder` | `must be a folder inside the workspace` |
    | any other rule | `is not valid` |

  - The description becomes `invalid request: ` + `<attribute> <problem>` joined by `; `, in
    validator order (struct field order).
- **Impact on existing tests**:
  - `validator/custom_test.go` asserts the `'<rule>'` wording and is updated.
  - Frontend tests use hand-written fixtures and are unaffected.
  - No BDD scenario asserts `invalid request` or the old wording; `grep` finds none.
  - The window reads details only for `SND-05409`, `SND-06409` and `SND-07409`, `DOC-*` and
    sync errors, none of them validation errors.

## R7b — Request-DTO names vs MCP argument names

- **Fact** (a sweep of all 33 argument DTOs, recorded in data-model.md Appendix A):
  - Validation runs on the request DTOs the arguments are converted into, so about 20 MCP
    arguments are validated under another name. Examples: `workspace_edit.id` → `local_id`,
    `import.on_clash` → `apply_to_all` and `items[].resolution`, `export.target.kind` →
    `item_kind`, `git_edit.url` → `repository`.
  - Nested request fields are validated through prefixes such as `request.` and
    `example.request.` for `tree_edit`.
  - Today's errors use the leaf name only (`fieldError.Field()`), so `auth.type` is reported as
    `type`.
- **Decision**:
  - Attributes become full JSON paths (R7).
  - Each MCP argument field lists, in `validated_as`, the request paths it feeds (optionally
    limited to an action with `@action`).
  - `MCPTool.ArgumentOf(action, attribute)` maps the longest matching prefix back to the argument's own
    path, and `NewMCPToolError(failure, tool, action)` applies it. The validation message is rebuilt from
    the mapped details.
  - Internal attributes (`local_id` from the active workspace, `tab_id`, server-built ids) match
    nothing and stay as they are, because the client cannot send them.
- **Alternatives considered**:
  - Validating the MCP DTOs directly was rejected: one DTO serves several actions with
    different rules, and the request DTOs are shared with the window.
  - Renaming the request-DTO JSON fields was rejected because it would break the window
    bindings.

## R8 — MCP error details

- **Fact**: `dto.NewMCPToolError(failure)` keeps only `Code()` and `Description()`.
  `failure.ToDto("").Error.ErrorDetails` already exposes the details, and
  `unconfirmed_send_environment.go` reads them that way today.
- **Decision**:
  - `MCPToolError` gains ``Details []MCPToolErrorDetail `json:"details,omitempty"` ``.
  - `MCPToolErrorDetail` is `{field, problem}`: `field` is the attribute and `problem` is the
    messages joined by `, `.
  - The leak guard still runs over the whole encoded text.

## R9 — Loose folder expansion

- **Facts**:
  - `importSourcePersistence.readPath` turns a folder into one `ImportInput` with
    `Probe.Entries` (the recursive file listing) and `Documents`.
  - `DetectImportFormat` → `detectFolder` returns `IMP-00415 <name> is not a Xenon, Postman,
    Bruno or cURL export` when no layout matches.
  - Detection runs in `importReaders.detected` (service) before any reader.
- **Decision**:
  1. `detectFolder`'s default branch returns `(importkind.Folder, "", nil)`.
     `entity.ImportInput.IsLooseFolder()` is `Kind == importkind.Folder`.
  2. A new port `usecase.ReadImportFolderFiles.ReadFolderFiles(ctx, input)` is implemented by
     `importSourcePersistence`. It lists `input.Root` with `os.ReadDir`, sorted by name:
     - hidden entries are ignored silently;
     - subfolders become skipped entries with reason `folder`;
     - regular files go through the existing `readPath`, so zips are extracted and the size
       and UTF-8 limits apply;
     - a file that fails to read is skipped with its error description as the reason;
     - the result is the file inputs, their temp dirs and the skipped entries.
  3. `importReaders.detected` replaces each loose-folder input, in place, with the inputs it
     expands to, after detecting each of them with `DetectImportFormat`:
     - a file that fails detection is added to the skipped entries with its error description;
     - a file that is itself `Folder` cannot occur, because only regular files are read.

     If no input survives, it returns `usecase.RejectUnimportableFolder(name)`: `ClientError`
     `IMP-00415`, HTTP 415, message
     `<name> has no file Xenon can import; pass a Postman v3 workspace or collection folder, a Postman data dump (archive.json), a Bruno collection (bruno.json or opencollection.yml), a Xenon workspace folder, or a folder of standalone export files`.
     Temp dirs from the expansion join the source's temp dirs, so `Discard` removes them.
  4. When there are skipped entries, one source warning is added with code `skipped_part`
     (existing code) and message `Skipped in <folder>: notes.txt (not a Xenon, Postman, Bruno or
     cURL export), sub (folder)`.
- **Rationale**:
  - Reading stays in persistence and detection stays in the use-case function.
  - The service only orchestrates, so the skills' layering holds.
  - Reusing `readPath` keeps every existing safety limit.
- **Alternatives considered**: Probing the files inside `readImportFolder` was rejected because
  it would duplicate detection inside persistence.

## R10 — Top-level kind of a batch

- **Fact**: `entity.NewImportSource` copies `inputs[0].Kind` and `inputs[0].Version`.
- **Decision**:
  - When every input has the same kind and the same version, report that kind and version.
  - Otherwise report `importkind.Mixed` (`"mixed"`) with version `""`.
  - Expanded loose-folder inputs count individually, so a folder of v2.1 collections reports
    `postman_collection` / `2.1`.
  - The frontend `ASSIGNMENT_KINDS` (`postman_dump`, `postman_workspace`) does not include
    `mixed`, so the regular item table shows. `KIND_LABELS.mixed = 'Multiple sources'`.

## R11 — Existing scenarios whose expectation changes

- `import/postman_v3_workspace.feature` @nothing: the `specs-only` folder has only a subfolder,
  so its description becomes the R9 layout message. The code stays `IMP-00415`.
- `application/usecase/import_detection_test.go`: the folder cases `specs only` and `empty dir`
  expect `(importkind.Folder, nil)`. `req folder` and `env folder` are Xenon probes and are
  unchanged.
- `validator/custom_test.go`: new wording.
- No other BDD scenario asserts these texts (grep for `is not a Xenon`, `IMP-00415`,
  `invalid request` and `failed on the` under `test/integration/features`).
