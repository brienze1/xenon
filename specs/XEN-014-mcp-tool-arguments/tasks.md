---
description: "Task list for XEN-014 — MCP tools declare their arguments; clearer workspace and import errors"
---

# Tasks: MCP Tools Declare Their Arguments; Clearer Workspace and Import Errors

**Input**: Design documents from `specs/XEN-014-mcp-tool-arguments/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md (Appendix A is the tag source of
truth), contracts/, quickstart.md

**Tests**: MANDATORY (constitution V, BDD-first). Each user story starts with Godog scenarios in
`test/integration/features/` that are proven red before implementation. Go unit tests and Vitest
cover the pure pieces.

**Organization**: Tasks are grouped by user story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: US1 (argument schema), US2 (readable errors), US3 (loose folders, mixed kind)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Fixtures shared by the BDD scenarios.

- [ ] T001 [P] Create the loose-folder fixtures:
  - `test/integration/resources/import/loose/mixed/collection.json`: a minimal Postman v2.1
    collection, `info.name` "Loose Mixed", one `GET https://api.test/ping` request, no auth,
    scripts or variables, so it detects with zero warnings;
  - `mixed/dev.postman_environment.json` (copy of `postman/dev.postman_environment.json`);
  - `mixed/globals.json` (copy of `postman/globals.json`);
  - `collections-only/a.json` and `collections-only/b.json` (copies of `mixed/collection.json`
    with `info.name` "Loose A" and "Loose B" and distinct `_postman_id`s);
  - `with-extras/` = the three `mixed/` files plus `notes.txt`, `.DS_Store` and `sub/readme.md`;
  - `nothing/notes.txt` and `nothing/sub/readme.md`.

  Add a line for `loose/` to `test/integration/resources/import/README.md`.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The standalone schema reflector. US1 builds on it; US2 reuses the argument metadata
it reads.

- [ ] T002 [P] Write the unit tests in `pkg/jsonschema/properties_test.go`. Cover:
  - string / bool / int / float / pointer / slice / slice-of-struct / nested struct / map;
  - `json:"-"`, `schema:"-"`, `schema:"integer|string"` on a `json.RawMessage`;
  - `json.RawMessage` without a `schema` tag → error;
  - the `enum`, `description`, `required_for`, `used_by`, `validated_as` and `schema_for` tags
    read into `Property`;
  - recursion guard on a self-referencing type;
  - `Property.Map()` rendering one type as a string and two as an array, and never rendering
    `RequiredFor`/`UsedBy`/`ValidatedAs`/`SchemaFor`.

  Prove the tests red.
- [ ] T003 Implement `pkg/jsonschema/property.go` (type `Property` per data-model.md, with
  `Map() map[string]any`) and `pkg/jsonschema/properties.go`
  (`Properties(t reflect.Type) (map[string]Property, error)`). Use only the stdlib, with no
  `internal/` or other `pkg/` import (pkg-generic). Make T002 green.

**Checkpoint**: `go test ./pkg/jsonschema/...` is green.

---

## Phase 3: User Story 1 — AI clients discover every tool argument and call `import` (P1) 🎯 MVP

**Goal**: `tools/list` declares every argument of all 24 tools, with type, description and
per-action requirement. `import` accepts an array `paths` from any client.

**Independent Test**: `make bdd TAG=@xen_014_schema`. `tools/list` shows the typed arguments of
`workspace_edit`, `import` and `tree`, and `import run` with an array `paths` imports the
collection.

### Tests for User Story 1 (write first, prove red)

- [ ] T004 [US1] Write `test/integration/features/mcp/tool_arguments.feature` with tags
  `@all @xen_014 @mcp @mcp_tool_arguments`, Background `Given the workspace "Acme" exists`, and
  the following `@xen_014_schema` scenarios:
  - **@schema workspace_edit.** `When I list the MCP tools`, then the existing step
    `the response field "" should contain an item with "name" equal to "workspace_edit" and "<path>" equal to "<value>"`
    holds for each row:
    - `inputSchema.properties.id.type` = `string`
    - `inputSchema.properties.name.type` = `string`
    - `inputSchema.properties.path.type` = `string`
    - `inputSchema.properties.path.description` = `Absolute path of the workspace folder. Used by create and open. Required for create and open.`
    - `inputSchema.required.0` = `action`
  - **@schema import.**
    - `inputSchema.properties.paths.type` = `array`
    - `inputSchema.properties.paths.items.type` = `string`
    - `inputSchema.properties.on_clash.enum.0` = `keep_both`
    - `inputSchema.properties.source_id.type` = `string`
    - `inputSchema.properties.passphrase.type` = `string`
  - **@schema tree depth.**
    - `inputSchema.properties.depth.type.0` = `integer`
    - `inputSchema.properties.depth.type.1` = `string`
  - **@schema object.** `every item of the response field "" should have the field "inputSchema.type" matching "^object$"`.
    The absence of `oneOf`/`anyOf`/`allOf` is locked by the T005 and T006 unit tests.
  - **@schema settings_edit.** The tool with `name` = `settings_edit` has
    `inputSchema.properties.script_timeout_ms.type` = `integer`. The absence of `ai_access` is
    locked by the T005 unit test, because the step catalog has no item-scoped "field should not
    exist" step and this does not justify a new one.
  - **@import_array.** `I call the MCP tool "import" with { "action": "run", "paths": ["{{resources}}/import/postman/full-v2.1.json"] }`
    → `the call should succeed`. Then `tree list` contains the imported collection's name, with
    the same assertions as `other_tools.feature` @import.
  - **@no_details.** `I call the MCP tool "tree" with { "action": "get", "id": "nope" }` →
    `the call should fail with code "MCP-08404"` and `the error should not contain "details"`.

  Run `make bdd TAG=@xen_014_schema` and record that the schema scenarios fail. `@import_array`
  already passes and is the FR-007 regression guard.
- [ ] T005 [P] [US1] Write `internal/integration/entrypoint/dto/mcp_tool_test.go` with the
  `InputSchema()` unit tests for:
  - union of two registrations sharing a field (type agrees; Used by and Required for are
    unions);
  - "Used by"/"Required for" wording (`a`, `a and b`, `a, b and c`, in the tool's action order);
  - `used_by` narrowing;
  - `required_for:"always"` → top-level `required`;
  - a single-action tool gets no "Used by";
  - a no-action tool (`send`) has no selector property;
  - `schema_for` filtering by tool name;
  - the real `settings_edit` tool (built from `dto.MCPSettingsSave`) declares no `ai_access`;
  - no `oneOf`/`anyOf`/`allOf` key;
  - `ArgumentOf` (exact match, longest prefix, `[]` index wildcard, `@action` qualifier, no match
    → unchanged).

  Prove the tests red.
- [ ] T006 [P] [US1] Write `internal/integration/entrypoint/controller/mcp_tool_arguments_test.go`
  (research.md R6):
  1. **AST drift test.** Parse `controller/mcp_*.go` (non-test), group methods by receiver type,
     and collect decoded DTO types (`var arguments dto.X` followed by `call.Decode(&arguments)`)
     and registered types (`dto.NewMCPToolArguments[dto.Y]` in `Tool()`). Fail when a decoded
     type is not registered.
  2. **Completeness test.** Construct all 24 MCP tool controllers with nil dependencies and call
     only `Tool()`. Fail when `InputSchema()` reports a type conflict, a top-level property lacks
     `type` or `description`, or a `used_by`/`required_for` name matches no action of the
     registering tools.

  Prove the tests red.

### Implementation for User Story 1

- [ ] T007 [US1] Create `internal/integration/entrypoint/dto/mcp_tool_arguments.go`:
  - `type MCPToolArguments struct { Actions []string; Shape reflect.Type }`;
  - `func NewMCPToolArguments[T any](actions ...string) MCPToolArguments`.
- [ ] T008 [US1] Change `internal/integration/entrypoint/dto/mcp_tool.go`:
  - add ``Arguments []MCPToolArguments `json:"-"` ``;
  - change the constructor to `NewMCPTool(tool entity.AITool, arguments ...MCPToolArguments)`;
  - rewrite `InputSchema()` to merge `jsonschema.Properties` of every registration per
    data-model.md "MCPTool" and Appendix A notes 1–3;
  - add `ArgumentOf(action, attribute string) string` (research.md R7b; the action argument
    applies `@action` qualifiers).

  Make T005 green.
- [ ] T009 [P] [US1] Tag the argument DTOs exactly as listed in data-model.md Appendix A (rows
  console … export) in `internal/integration/entrypoint/dto/`: `mcp_console_list.go`,
  `mcp_console_get.go`, `mcp_cookies_list.go`, `mcp_cookie_upsert.go`,
  `mcp_cookie_allowlist.go`, `mcp_delete.go`, `mcp_docs_input.go`, `mcp_env_scope.go`,
  `mcp_env_inspect.go`, `mcp_env_edit.go`, `mcp_env_activation.go`, `mcp_env_variable.go`,
  `mcp_env_variables.go`, `mcp_variable_row.go`, `mcp_export.go`, `mcp_export_target.go`,
  `mcp_save_response_body.go`.
- [ ] T010 [P] [US1] Tag the remaining argument DTOs exactly as listed in data-model.md
  Appendix A (rows git_edit … workspace_edit): `mcp_git_link.go`, `mcp_git_settings.go`,
  `mcp_git_commit.go`, `mcp_git_resolve.go`, `mcp_git_resolution.go`, `mcp_history_list.go`,
  `mcp_history_get.go`, `mcp_import.go`, `mcp_response_body.go`, `mcp_script_check.go`,
  `mcp_send.go`, `mcp_send_target.go`, `mcp_settings_save.go` (`ai_access` → `schema:"-"`),
  `mcp_snippet_input.go`, `mcp_tree_list.go` (`depth` → `schema:"integer|string"`),
  `mcp_tree_search.go`, `mcp_tree_token_owner.go`, `mcp_tree_item.go`, `mcp_tree_edit.go`,
  `mcp_workspace_edit.go`.

  Nested shared request DTOs (`Request`, `Body`, `Auth`, `Scripts`, `KeyValue`,
  `ExampleResponseEdit`) need no tags. Their fields are reflected by type, and a nested
  `description` is optional (research.md R5).
- [ ] T011 [US1] Register the decoded DTOs in each controller's `Tool()` in
  `internal/integration/entrypoint/controller/`. Pass
  `dto.NewMCPTool(<tool>AITool, dto.NewMCPToolArguments[dto.X](actions…)…)` with the
  registrations from data-model.md Appendix A, column "Registration", in:
  `mcp_console.go`, `mcp_cookies.go`, `mcp_cookies_edit.go`, `mcp_delete.go`, `mcp_docs.go`,
  `mcp_docs_edit.go`, `mcp_env.go`, `mcp_env_edit.go`, `mcp_export.go`, `mcp_git.go` (none),
  `mcp_git_edit.go`, `mcp_git_push.go` (none), `mcp_history.go`, `mcp_import.go`,
  `mcp_response.go`, `mcp_scripts.go`, `mcp_send.go`, `mcp_settings.go` (none),
  `mcp_settings_edit.go`, `mcp_snippet.go`, `mcp_tree.go`, `mcp_tree_edit.go`,
  `mcp_workspace.go` (none), `mcp_workspace_edit.go`.

  Make T006 green.
- [ ] T012 [US1] Run `make bdd TAG=@xen_014_schema` until green, then `make bdd TAG=@mcp`. The
  existing `environment_reorder.feature` @listing (`inputSchema.properties.action.enum.3`) and
  `stdio_bridge.feature` must stay green.

**Checkpoint**: US1 is independently shippable. Every tool declares its arguments, and `import` is
callable from Claude Code.

---

## Phase 4: User Story 2 — Validation errors say which argument is wrong and why (P2)

**Goal**: Validation errors name the field (as the client sent it) and the rule in plain English,
over MCP (`details`) and in the window (`error_details`, description).

**Independent Test**: `make bdd TAG=@xen_014_errors`. `workspace_edit create` without `path`, with
a relative path, with nothing, and `switch` without `id` all return readable field-level errors.
`WorkspaceService.Create` returns the same wording.

### Tests for User Story 2 (write first, prove red)

- [ ] T013 [US2] Add `@xen_014_errors @validation` scenarios to
  `test/integration/features/mcp/tool_arguments.feature`:
  - **missing path.** `I call the MCP tool "workspace_edit" with { "action": "create", "name": "Payments" }`:
    - `the call should fail with code "WKS-00400"`;
    - `the error should contain the field "error.message" equal to "invalid request: path is required"`;
    - `"error.details.0.field"` = `path`;
    - `"error.details.0.problem"` = `is required`.
  - **relative path.** `{ "action": "create", "name": "Payments", "path": "relative/dir" }` gives
    `WKS-00400`, `error.details.0.problem` = `must be an absolute path`.
  - **both missing.** `{ "action": "create" }` gives
    `error.message` = `invalid request: name is required; path is required`,
    `error.details.1.field` = `path`.
  - **id mapped back.** `{ "action": "switch" }` gives `WKS-08400`,
    `error.message` = `invalid request: id is required`, `error.details.0.field` = `id`.
  - **import on_clash mapped back.**
    `I call the MCP tool "import" with { "action": "run", "paths": ["{{resources}}/import/postman/full-v2.1.json"], "on_clash": "merge" }`
    gives `IMP-13400` and `error.details.0.field` = `on_clash`.

  Prove them red.
- [ ] T014 [P] [US2] Extend `test/integration/features/workspaces/create_workspace.feature` with
  `@xen_014 @xen_014_errors` steps on the existing @validation scenarios:
  - **relative folder path** adds:
    - `the error should contain the field "error.error_details.0.messages.0" equal to "must be an absolute path"`;
    - `the error should contain the field "error.description" equal to "invalid request: path must be an absolute path"`.
  - **blank name** adds `"error.error_details.0.messages.0"` = `must not be blank`.

  Prove them red.
- [ ] T015 [P] [US2] Rewrite `internal/integration/entrypoint/validator/custom_test.go`'s wording
  assertions to the new contract. Add one table case per rule in contracts/mcp-error.md
  "Problem wording":
  - `required`, `required_if`, `required_unless`, `not_blank`, `oneof`;
  - `max`/`min` on a string and on a slice;
  - `gt`, `gte`, `email`, `absolute_path`, `has_id`, `variable_scope`, `import_input`,
    `safe_subfolder`;
  - an unregistered tag → `is not valid`.

  Also add:
  - a nested-path case (`variables[1].name`);
  - the description join (`invalid request: name must not be blank; path must be an absolute path`).

  Prove the tests red.
- [ ] T016 [P] [US2] Write `internal/integration/entrypoint/dto/mcp_tool_error_test.go`:
  - a validation `BadRequestError` with details `local_id`/`[is required]` through a tool whose
    `id` has `validated_as:"local_id"` → `{field: id, problem: is required}` and message
    `invalid request: id is required`;
  - a `ConflictError` with `missing_secrets: [A, B]` → `{field: missing_secrets, problem: "A, B"}`
    with the message unchanged;
  - a `NotFoundError` with no details → JSON with no `details` key.

  Prove the tests red.

### Implementation for User Story 2

- [ ] T017 [US2] Create `internal/integration/entrypoint/validator/problem.go` with
  `problemOf(fieldError validator.FieldError) string`, implementing the wording table in
  contracts/mcp-error.md. Pick the string/slice variant by `fieldError.Kind()`, join the `oneof`
  params with `, `, and fall back to `is not valid`.
- [ ] T018 [US2] Change `internal/integration/entrypoint/validator/custom.go` `Struct`:
  - `Attribute` = `fieldError.Namespace()` without its first segment (JSON names are already
    registered through `jsonName`);
  - `Messages` = `[problemOf(fieldError)]`;
  - description = `invalid request: ` + `<attribute> <problem>` joined by `; `.

  Make T015 green.
- [ ] T019 [US2] Create `internal/integration/entrypoint/dto/mcp_tool_error_detail.go`
  (`MCPToolErrorDetail{Field string json:"field"; Problem string json:"problem"}`). Change
  `internal/integration/entrypoint/dto/mcp_tool_error.go`:
  - add ``Details []MCPToolErrorDetail `json:"details,omitempty"` ``;
  - `NewMCPToolError(failure errs.BaseError, tool MCPTool, action string)` maps each detail with
    `tool.ArgumentOf(action, attribute)` and joins messages with `, `;
  - when the description starts with `invalid request: `, rebuild the message from the mapped
    details.

  Make T016 green.
- [ ] T020 [US2] In `internal/integration/entrypoint/controller/mcp_call.go`, pass `definition`
  and `toolCall.Action` into `respond(...)` and on to `dto.NewMCPToolError`. In `panicked(...)`,
  pass `dto.MCPTool{}` and `""` (panics carry no details). Update
  `internal/integration/entrypoint/controller/mcp_call_test.go` call sites.
- [ ] T021 [US2] Run `make bdd TAG=@xen_014_errors`, then `make bdd TAG=@workspaces` and
  `make bdd TAG=@mcp`, until green.

**Checkpoint**: US2 is shippable on its own. Its schema-dependent part (`id` → `local_id`
mapping) needs T008/T009–T011 from US1.

---

## Phase 5: User Story 3 — A folder of loose export files imports; mixed batches say `mixed` (P3)

**Goal**: An unrecognised folder expands into its top-level files. Skipped entries are reported.
A folder with nothing importable fails with the layout message. A batch reports its shared kind
or `mixed`.

**Independent Test**: `make bdd TAG=@xen_014_loose` against the `loose/` fixtures.

### Tests for User Story 3 (write first, prove red)

- [ ] T022 [US3] Write `test/integration/features/import/loose_folder.feature` with tags
  `@all @xen_014 @import @xen_014_loose` and Background `Given the workspace "Acme" exists`:
  - **@expand.** `ImportService.Detect { "paths": ["{{resources}}/import/loose/mixed"] }`:
    - succeeds;
    - `items` array length 3;
    - `kind` = `mixed`, `version` = ``;
    - `items.0.kind` = `collection`, `items.1.kind` = `environment`, `items.2.kind` = `globals`
      (name order: `collection.json`, `dev.postman_environment.json`, `globals.json`);
    - `the response should contain the field "warnings" array length equal to 0` (the fixtures
      carry no item warnings, T001).

    Then `ImportService.Import` with the source and `apply_to_all: keep_both` adds the
    collection to the tree, using the existing import steps from `postman_collection.feature`.
  - **@skipped.** `with-extras` gives 3 items and one warning with:
    - `warnings.0.code` = `skipped_part`;
    - `warnings.0.message` containing `notes.txt (not a Xenon, Postman, Bruno or cURL export)`;
    - `warnings.0.message` containing `sub (folder)`;
    - the response not containing `.DS_Store`.
  - **@nothing.** `nothing` → `the call should fail with code "IMP-00415"` and
    `error.description` = `nothing has no file Xenon can import; pass a Postman v3 workspace or collection folder, a Postman data dump (archive.json), a Bruno collection (bruno.json or opencollection.yml), a Xenon workspace folder, or a folder of standalone export files`.
  - **@kind.** `collections-only` → `kind` = `postman_collection`, `version` = `2.1`.
  - **@kind @mixed_paths.** `{ "paths": ["{{resources}}/import/postman/globals.json", "{{resources}}/import/postman/full-v2.1.json"] }`
    → `kind` = `mixed`, `version` = ``.
  - **@mcp.** `I call the MCP tool "import" with { "action": "detect", "paths": ["{{resources}}/import/loose/mixed"] }`
    → succeeds, `kind` = `mixed`, `items` length 3.

  Prove them red.
- [ ] T023 [P] [US3] Update `test/integration/features/import/postman_v3_workspace.feature`
  @nothing: the expected `error.description` becomes `specs-only has no file Xenon can import; pass a Postman v3 workspace or collection folder, a Postman data dump (archive.json), a Bruno collection (bruno.json or opencollection.yml), a Xenon workspace folder, or a folder of standalone export files`,
  with code `IMP-00415` unchanged. Prove it red.
- [ ] T024 [P] [US3] Update `internal/application/usecase/import_detection_test.go`:
  - the folder cases `specs only` and `empty dir` expect `(importkind.Folder, "", nil)`;
  - add a case for `RejectUnimportableFolder("x")` (code `IMP-00415`, the layout message).

  Prove the tests red.
- [ ] T025 [P] [US3] Write `internal/domain/entity/import_source_test.go` cases for
  `NewImportSource`:
  - same kind and version → that kind;
  - same kind, different version → `mixed`/"";
  - different kinds → `mixed`/"";
  - one input → its kind;
  - none → empty.

  Write `NewSkippedEntriesWarning` message cases in
  `internal/domain/entity/import_warning_test.go`. Prove the tests red.
- [ ] T026 [P] [US3] Add `mixed → "Multiple sources"` and `sourceLabel('mixed', '')` cases to
  `frontend/src/features/import-export/lib/labels.test.ts`. Prove the test red.

### Implementation for User Story 3

- [ ] T027 [P] [US3] Add `Folder enums.ImportKind = "folder"` and
  `Mixed enums.ImportKind = "mixed"` to `internal/domain/enums/importkind/value.go`.
- [ ] T028 [US3] Domain changes:
  - `internal/domain/entity/import_input.go`: new field `Path` and `IsLooseFolder()`.
    `internal/integration/persistence/import_source.go` `readPath` sets `input.Path = path` on
    every returned input.
  - new `internal/domain/entity/import_skipped_entry.go`: `ImportSkippedEntry{Name, Reason}`.
  - `internal/domain/entity/import_warning.go`: `NewSkippedEntriesWarning(folder string, entries []ImportSkippedEntry)` →
    `warningcode.SkippedPart` with message `Skipped in <folder>: <name> (<reason>), …`; a reason
    that starts with `<name> ` drops that prefix.
  - `internal/domain/entity/import_source.go`: `NewImportSource` computes the shared kind and
    version, or `importkind.Mixed`/"".

  Make T025 green.
- [ ] T029 [US3] In `internal/application/usecase/detect_import_format.go`:
  - `detectFolder` default → `return importkind.Folder, "", nil`;
  - add `RejectUnimportableFolder(name string) error` (`errs.ClientError`, `IMP-00415`,
    `http.StatusUnsupportedMediaType`, the message from contracts/import-detect.md).

  Make T024 green.
- [ ] T030 [US3] Create the port `internal/application/usecase/read_import_folder_files.go`
  (`ReadImportFolderFiles` interface, per data-model.md). Embed it in the persistence adapter
  interface `internal/integration/adapter/import_source_persistence.go`.
- [ ] T031 [US3] Implement `ReadFolderFiles` in
  `internal/integration/persistence/import_source.go`:
  - `os.ReadDir(folder.Path)` (already name-sorted);
  - hidden entries (leading `.`) and non-regular, non-directory entries such as symbolic links
    are skipped silently and never followed;
  - directories → `ImportSkippedEntry{Name, "folder"}`;
  - regular files → `i.readPath(filepath.Join(root, name))`, collecting temp dirs;
  - a read error → `ImportSkippedEntry{Name, <BaseError description>}`;
  - a failure to list the folder → `errs.NotFoundError(name+" cannot be found", "IMP-08404")`.

  Log Started/Finished per persistence.md.
- [ ] T032 [US3] In `internal/application/service/import_readers.go`:
  - add `readImportFolderFiles usecase.ReadImportFolderFiles` to `importReaders`;
  - in `detected`, after `DetectImportFormat`, replace each `IsLooseFolder()` input in place with
    its expanded files, each detected with `DetectImportFormat` (failures become skipped entries
    with the error description);
  - return `usecase.RejectUnimportableFolder(probe.Name)` when none survive;
  - append the expansion's temp dirs to the temp dirs passed to `NewImportSource`;
  - add `NewSkippedEntriesWarning` to the source warnings when anything was skipped.

  Change `detected` to return the extra temp dirs so `manage_import.go` `Detect` can discard
  them on failure. Wire the port in `internal/application/service/manage_import.go`
  (`NewManageImportService` gets `readImportFolderFiles`) and in the DI container
  (`internal/infra/dependency/application/manage_import.go`: pass one more
  `integration.GetImportSourcePersistence()` argument, in the constructor's new parameter position).
- [ ] T033 [P] [US3] Add `mixed: 'Multiple sources'` to `KIND_LABELS` in
  `frontend/src/features/import-export/lib/labels.ts`. Make T026 green.
- [ ] T034 [US3] Run `make bdd TAG=@xen_014_loose`, `make bdd TAG=@import` and
  `go test ./internal/...` until green.

**Checkpoint**: All three stories work.

---

## Phase 6: Polish & Cross-Cutting

- [ ] T035 Run `make sort-imports`, then the full `make bdd` (every pre-existing scenario green),
  `cd frontend && npx vitest run`, and `make validate`. Fix anything they report.
- [ ] T036 [P] Add a short "MCP tool arguments" note to `docs/prd/xenon-mvp-prd.md` under the MCP
  section, recording:
  - tool schemas declare every argument;
  - validation errors name the argument and rule;
  - loose folders import their top-level files;
  - `mixed` source kind.

  These are the user decisions in spec.md Clarifications.
- [ ] T037 Run quickstart.md "Expected outcomes" 1–4 against the BDD run output, and confirm:
  - SC-001: T006 green;
  - SC-003: T013–T015 green;
  - SC-004: T022 @expand green, a folder giving the same 3 items and 0 warnings as its three
    files detected one by one;
  - SC-005: T022 @kind green.

---

## Dependencies & Execution Order

- **Setup (T001)** has no dependency.
- **Foundational (T002 → T003)** blocks US1.
- **US1 (T004–T012)**:
  - T007 → T008 → (T009, T010 in parallel) → T011 → T012;
  - T004/T005/T006 are written first and proven red.
- **US2 (T013–T021)**:
  - T017 → T018 can start after Foundational;
  - T019/T020 need T008 (`ArgumentOf`) and the US1 `validated_as` tags (T009/T010) for the
    mapped-name scenarios.
- **US3 (T022–T034)**:
  - independent of US1/US2 (needs only T001);
  - T027 → T028 → T029 → T030 → T031 → T032; T033 is in parallel.
- **Polish (T035–T037)** comes after all stories.

## Parallel Opportunities

- T001 and T002 can run together.
- US3 can run alongside US1 after T001: it touches different files (domain import,
  usecase/service import, persistence, frontend).
- Within US1: T005 ∥ T006, and T009 ∥ T010.
- Within US2: T014 ∥ T015 ∥ T016.
- Within US3: T023 ∥ T024 ∥ T025 ∥ T026, and T033 ∥ T028–T032.

## Implementation Strategy

1. **MVP = US1.** Declared arguments unblock `import` from Claude Code, which is the reported bug.
2. **US2** next. Readable errors depend on US1's tags for name mapping.
3. **US3** last, or in parallel with US1, because it is independent.
4. Each story ends with its BDD tag green. Polish runs the full suite and `make validate`.
