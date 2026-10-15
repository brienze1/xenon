# Implementation Plan: MCP Tools Declare Their Arguments; Clearer Workspace and Import Errors

**Branch**: `xen-014` | **Date**: 2026-10-15 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/XEN-014-mcp-tool-arguments/spec.md`

## Summary

Three fixes to the MCP surface and to import detection.

1. **Tool argument schema (US1, FR-001–007).** The MCP argument DTOs get two tags on top of `json`:
   - `description:"…"` holds one sentence about the argument.
   - `required_for:"a,b"` lists the actions that require it.

   Optional tags:
   - `enum:"…"` declares a closed value set;
   - `schema:"…"` overrides the type for `json.RawMessage` fields;
   - `used_by:"…"` narrows the actions that read a field of a shared DTO;
   - `validated_as:"…"` names the field in the request DTO the action validates, when that name
     differs.

   A new standalone package, `pkg/jsonschema`, reflects a struct type into JSON-Schema properties.
   Each controller's `Tool()` registers the DTO types its actions decode, using
   `dto.NewMCPToolArguments[T](actions…)`. `dto.MCPTool.InputSchema()` merges those types
   into one `properties` map:
   - it keeps the action selector as today;
   - each top-level description gets "Used by …" and "Required for …" from the registration;
   - it never uses top-level `oneOf`/`anyOf`/`allOf`.

   Unit tests keep the schema honest:
   - an AST test fails if any controller decodes a DTO it did not register;
   - a schema test fails if any top-level argument lacks a description or two registered
     DTOs disagree on an argument's type.
2. **Readable validation errors (US2, FR-008–012).**
   - `validator.CustomValidator().Struct` writes each detail as `{attribute: <json path>,
     messages: [<plain-English problem>]}` and the description as `invalid request: <field>
     <problem>; …`.
   - `dto.MCPToolError` gains `details: [{field, problem}]`, taken from the failure's
     `ErrorDetails`. Each attribute is mapped back to the argument the client sent through
     `validated_as` (`local_id` → `id`), and the validation message is rebuilt from those
     mapped names.
   - The window's error shape is unchanged.
3. **Loose folders and mixed batches (US3, FR-013–019).**
   - `DetectImportFormat` returns the new internal kind `importkind.Folder` for a folder that
     matches no known layout, instead of failing.
   - `importReaders.detected` expands each such input. A new port, `usecase.ReadImportFolderFiles`
     (implemented by `importSourcePersistence`), reads the folder's top-level files one by one.
     Each file is detected as if passed alone. Failures and subfolders become one `skipped_part`
     source warning.
   - A folder with no importable file fails with a new `IMP-00415` layout message.
   - `entity.NewImportSource` reports the shared kind and version, or the new `importkind.Mixed`
     (`mixed`) with an empty version.
   - The window labels `mixed` "Multiple sources".

Decisions and alternatives: [research.md](research.md).

## Technical Context

**Language/Version**: Go 1.26 (toolchain 1.27.2); TypeScript 5 + React 19 (one label).

**Primary Dependencies**: no new dependency. The schema reflector uses only the stdlib
(`reflect`, `strings`) in `pkg/jsonschema`. The MCP SDK (`modelcontextprotocol/go-sdk` v1.8.0)
already accepts a `map[string]any` input schema. `go-playground/validator/v10` is already used.

**Storage**: none. Import sources stay in memory (`importSourcePersistence.sources`), and the
temp dirs of zip files inside a loose folder join the source's `TempDirs`.

**Testing**:
- godog BDD in `test/integration/features/mcp/tool_arguments.feature`,
  `test/integration/features/import/loose_folder.feature` and
  `test/integration/features/workspaces/create_workspace.feature` (new scenarios). Only the file
  system root is faked (`{{temp_root}}`, `{{resources}}`), as today.
- Go unit tests:
  - `pkg/jsonschema`;
  - `dto.MCPTool.InputSchema`;
  - the controller argument-drift AST test;
  - `validator` wording;
  - `DetectImportFormat`;
  - `entity.NewImportSource`.
- Vitest for `labels.ts`.

**Target Platform**: macOS, Windows, Linux (desktop, Wails v3). The MCP clients are Claude Code,
Claude Desktop and any MCP client, over HTTP or the `xenon mcp` stdio bridge.

**Project Type**: desktop app (Go backend + React UI) with an embedded MCP server.

**Performance Goals**: schemas are built once per tool, at `addTool` registration (startup).
`tools/list` costs nothing extra per call.

**Constraints**:
- The input schema must be accepted by Claude's tool-use API, so no top-level
  `oneOf`/`anyOf`/`allOf`.
- Error codes are unchanged.
- No secret value appears in a schema or an error detail.
- Recognised folder layouts behave exactly as before.

**Scale/Scope**:
- 24 tools and 33 argument DTO types.
- About 15 validation rules get wording.
- 2 new import kinds, of which only `mixed` is visible.
- One frontend label.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first**: Nothing new is stored. Loose-folder import reads only paths the user
  or client passed, and the files land in the active workspace exactly as single-file imports do.
  Workspace isolation is unchanged: `import` stays workspace-scoped.
- [x] **II. Secrets**: Schemas hold argument names, types and descriptions only. Validation
  details name fields and rules, never values; `passphrase` failures say `is required`, never
  the value. The MCP error text still goes through `guardMCPText` (leak guard). `MCPImport`
  still logs `passphrase_length`, not the passphrase.
- [x] **III. Go owns logic**:
  - Schema building, validation wording, folder expansion and kind resolution are all in Go.
  - React only maps `mixed` to a label.
  - No Wails import is added outside `internal/infra/`; `tool_definition.go` keeps publishing
    `InputSchema()`.
- [x] **IV. Architecture skills**: Every file below has its layer:
  - `pkg/jsonschema` follows pkg-generic (no internal import, no other pkg);
  - argument DTOs and the tool DTO follow dtos.md;
  - the drift test and registration follow controllers.md;
  - the new port follows usecases.md and its implementation follows persistence.md;
  - kinds follow enums.md;
  - the skipped entry follows entities.md;
  - the frontend label follows frontend-architecture.
- [x] **V. BDD-first**: The feature files are listed per requirement below and written red by
  `/create-tests`. Only the file-system root is faked.
- [x] **VI. Postman parity / PRD scope**: There is no Postman behaviour change. Importing
  Postman-exported files is in the PRD's import scope; a folder of them is the same files.
- [x] **VII. Cross-platform**: Folder listing uses `os.ReadDir` + `filepath`. Hidden-entry
  detection uses the leading `.` on every OS (as `isHiddenEntry` already does). Field paths use
  JSON names, so they are OS-agnostic.

Post-design re-check: unchanged, all ✅. No Complexity Tracking entries.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-014-mcp-tool-arguments/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── mcp-tool-schema.md
│   ├── mcp-error.md
│   └── import-detect.md
├── checklists/
└── tasks.md
```

### Source Code (repository root)

```text
pkg/jsonschema/                                   ← new standalone package (pkg-generic)
├── properties.go          Properties(reflect.Type) → map[string]Property; Property.Map()
├── property.go            type Property {Type, Items, Properties, Enum, Description, RequiredFor}
└── properties_test.go

internal/domain/
├── enums/importkind/value.go                     ← + Folder ("folder", internal), Mixed ("mixed")
├── entity/import_source.go                       ← NewImportSource: shared kind/version or Mixed
├── entity/import_input.go                        ← IsLooseFolder()
├── entity/import_skipped_entry.go                ← new: {Name, Reason}
└── entity/import_warning.go                      ← NewSkippedEntriesWarning(folder, []ImportSkippedEntry)

internal/application/
├── usecase/detect_import_format.go               ← detectFolder default → importkind.Folder, nil;
│                                                   RejectUnimportableFolder(name) (IMP-00415)
├── usecase/read_import_folder_files.go           ← new port: ReadFolderFiles(ctx, input) ([]ImportInput, []string, []ImportSkippedEntry, error)
└── service/import_readers.go                     ← detected(): expand Folder inputs in place
    service/manage_import.go                      ← wire readImportFolderFiles into importReaders

internal/integration/
├── adapter/import_source_persistence.go          ← embed usecase.ReadImportFolderFiles
├── persistence/import_source.go                  ← ReadFolderFiles (top-level files via readPath)
├── entrypoint/validator/custom.go                ← field path + plain-English problems + message
├── entrypoint/validator/problem.go               ← new: rule → problem wording
├── entrypoint/dto/mcp_tool.go                    ← Arguments; InputSchema() merges properties
├── entrypoint/dto/mcp_tool_arguments.go          ← new: NewMCPToolArguments[T](actions…)
├── entrypoint/dto/mcp_tool_error.go              ← + Details; NewMCPToolError(failure, tool, action) maps fields via tool.ArgumentOf(action, attribute)
├── entrypoint/controller/mcp_call.go             ← respond()/panicked() pass the tool definition and action to NewMCPToolError
├── entrypoint/dto/mcp_tool_error_detail.go       ← new: {field, problem}
├── entrypoint/dto/mcp_*.go (33 argument DTOs + MCPExportTarget, MCPGitResolution, MCPVariableRow)
│                                                 ← description / required_for / used_by / enum / schema / validated_as / schema_for
│                                                   tags, exactly as listed in data-model.md Appendix A
└── entrypoint/controller/mcp_*.go (24 tools)     ← Tool() registers its argument DTOs
    entrypoint/controller/mcp_tool_arguments_test.go ← new: AST drift test + schema completeness test

internal/infra/server/mcp/tool_definition.go      ← unchanged (still publishes InputSchema())

frontend/src/features/import-export/lib/labels.ts (+ labels.test.ts) ← mixed: 'Multiple sources'

test/integration/
├── features/mcp/tool_arguments.feature           ← new (US1 + US2 over MCP)
├── features/import/loose_folder.feature          ← new (US3)
├── features/workspaces/create_workspace.feature  ← + readable details/description scenarios (FR-011)
├── features/import/postman_v3_workspace.feature  ← @nothing: expected description → layout message
└── resources/import/loose/                       ← new fixtures (see quickstart)
```

**Structure Decision**: Single Go module with the existing 4 layers plus `frontend/`. Every
change above sits in the layer that already owns that concern. The only new folder is
`pkg/jsonschema`.

## Requirement → layer → BDD map

| Requirement | Layer(s) | Locked by |
|---|---|---|
| FR-001–005 typed, described arguments from the decoded DTOs | pkg/jsonschema, entrypoint/dto, entrypoint/controller | `mcp/tool_arguments.feature` @schema (workspace_edit, import, tree depth, delete); Go drift + completeness tests (SC-001) |
| FR-006 no top-level combinators | entrypoint/dto | `mcp/tool_arguments.feature` @schema (`inputSchema.oneOf` absent); dto unit test |
| FR-007 `import run` with array `paths` (HTTP + stdio) | (already works server-side) | `mcp/tool_arguments.feature` @import_array; `mcp/stdio_bridge.feature` existing listing |
| FR-008, FR-010 MCP `details` + readable message | entrypoint/dto, validator | `mcp/tool_arguments.feature` @validation (missing path, relative path, both missing) |
| FR-009 wording of every rule | validator | `validator/custom_test.go` table over every rule in use |
| FR-011 window keeps shape, readable wording | validator | `workspaces/create_workspace.feature` @validation (description + `error_details.0.messages.0`) |
| FR-012 non-validation details pass through | entrypoint/dto | `dto/mcp_tool_error_test.go` (a `ConflictError` with `missing_secrets: [A, B]` → `{field: missing_secrets, problem: "A, B"}`; an error without details → no `details` key); `mcp/tool_arguments.feature` @no_details (`MCP-08404` for an unknown id carries no `details`) |
| FR-013–015 loose folder expansion, skipped warning, layout error | domain, usecase, service, persistence | `import/loose_folder.feature` @expand, @skipped, @nothing; `import/postman_v3_workspace.feature` @nothing |
| FR-016 known layouts unchanged | usecase | every existing `import/*.feature` scenario |
| FR-017 shared kind or `mixed` | domain | `import/loose_folder.feature` @kind (same-kind folder, mixed paths) |
| FR-018 window label | frontend | `labels.test.ts` |
| FR-019 window and MCP both | service | `loose_folder.feature` runs through `ImportService.Detect`, plus an MCP `import detect` scenario |
| FR-020 gates | all | `make bdd`, `make validate` |

## Complexity Tracking

No constitution violations.
