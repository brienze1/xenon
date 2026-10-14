# Implementation Plan: Import Postman v3 Collections, Environments and Git-Native Folders

**Branch**: `xen-009` | **Date**: 2026-10-14 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/XEN-009-postman-v3-import/spec.md`

## Summary

Xenon stops refusing Postman v3 (IMP-00415) and imports these sources:

- v3 collection JSON (`postman collection get --json`);
- git-native collection folders and single `*.request.yaml` files;
- v3 environments and globals (YAML and the CLI's JSON);
- whole `postman workspace pull` folders.

A new **v3 DTO package translates every v3 document into the equivalent Postman v2.1 document**, which then runs
through the **existing v2.1 mapping**. Fidelity, warnings, the missing-file checks and export passthrough are
therefore identical to v2.1 by construction (research R2). The contract is in
[contracts/v3-to-v21-mapping.md](contracts/v3-to-v21-mapping.md).

Detection gains v3 rules after the existing Xenon, dump and Bruno checks (R4), so detection of every existing
format is unchanged. Two kinds are new: `postman_request` and `postman_workspace`. v3 collections, environments and
globals reuse their kinds with version `3` (R3).

Secret values go to the vault:

- environment and globals secrets as in v2.1;
- collection variables with `secret: true` or v2.1 `type: secret` (R7);
- CLI-masked and vault-linked secrets as secrets with no value (R7, R8).

A workspace folder imports as several items with per-item workspace assignment, reusing the data-dump UI (R13).
Parts of the workspace that are not imported are reported. The MCP `import` tool needs only a description change,
because it already shares `ManageImportService`.

## Technical Context

**Language/Version**: Go 1.26 (toolchain 1.27.2); TypeScript 5 + React 19.

**Primary Dependencies**: No new dependencies. YAML uses the existing `go.yaml.in/yaml/v3` v3.0.5; JSON uses
`encoding/json`.

**Storage**: No change to workspace files. The output equals a v2.1 import. Secrets go to the existing per-workspace
age vault.

**Testing**:

- Godog BDD (`test/integration/`), driving `ImportService.*` bindings and the MCP `import` tool against real files.
- Fixtures are generated with `postman collection migrate` (R14).
- Only the outer boundaries are faked: file-system roots, keychain, clock.
- Go unit tests for the pure domain predicates (`import_probe_test.go`) and the detection table
  (`import_detection_test.go`).
- Vitest for the two frontend touches.

**Target Platform / Project Type**: desktop app (Wails v3) on macOS, Windows and Linux.

**Performance Goals**: A v3 collection of 1,000 requests (`large-1000-v3.zip`, generated with `collection migrate`)
meets the existing large-import bound: Detect < 2 s, Import < 8 s, progress at least every 250 ms (SC-006).

**Constraints**:

- No secret value in workspace files, logs, the preview or warnings (FR-019).
- No behaviour change for existing formats (FR-007, SC-005), except the deliberate v2.1 fix that routes secret
  collection variables to the vault (R7). No existing fixture has such a variable.
- Fixed orders independent of the OS (R10).

**Scale/Scope**: The same limits as other imports: a 200 MB file limit and a 1 GB archive limit. Only `postman/` is
read in a repository folder (R4).

## Constitution Check

- [x] **I. Local-first** — Files only. There is no Postman cloud access; the user or an AI client provides the
  files. Each item lands in exactly the workspace chosen for it. ✅
- [x] **II. Secrets** — Secret values go to the vault (R7). The CLI mask and vault links never become stored values.
  The v3 webservice logs counts only. Fixtures use fake values. BDD asserts that no workspace file or log line
  contains a fixture secret (SC-003). The R7 fix also closes an existing v2.1 path where a secret collection
  variable's value would have reached the collection file. ✅
- [x] **III. Go owns logic** — All parsing, detection and mapping is in Go. The frontend only adds two labels and
  treats `postman_workspace` like `postman_dump`. ✅
- [x] **IV. Architecture skills** — Every change has a layer (see Project Structure):
  - probe predicates in domain;
  - detection in a use case;
  - the new reader port in application;
  - the v3 contract DTOs and webservice in `integration/webservice` (`postmanv3/dto`, the same pattern as
    `postman/dto` and `bruno/dto`);
  - source reading in `integration/persistence`;
  - wiring in `infra/dependency`.

  Errors carry codes at their origin (`IMP-11400`). DTOs are created through constructors. ✅
- [x] **V. BDD-first** — Feature files are listed per requirement below. `/create-tests` writes them red first.
  Only the outer boundaries are faked. ✅
- [x] **VI. Postman parity, nothing more** — v3 is Postman's own current format, and the change preserves Postman's
  semantics. Auth inheritance, script levels and variable scopes go through the v2.1 path. The PRD changes in this
  task (FR-026) to list v3 import, so the feature is in scope before code lands. The product owner requested it in
  issue #4. Multi-auth is out of Xenon (Clarifications). ✅
- [x] **VII. Cross-platform** —
  - Paths are walked with `filepath` and turned into `ToSlash` entries.
  - Sorting is case-folded and does not depend on OS listing order.
  - CRLF in YAML is handled by yaml.v3.
  - Fixtures avoid characters that are illegal on Windows. Request names that need them use `name:` inside the file,
    as Postman does.

  ✅

**Post-design re-check (after Phase 1)**: all gates still pass. No new persistence, no Wails imports below `infra`, and
no secret-bearing field in any DTO returned to the UI or MCP. `ImportCollection.Secrets` stays inside the
application layer, as `ImportItem.CollectionSecrets` does today.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-009-postman-v3-import/
├── spec.md · plan.md · research.md · data-model.md · quickstart.md
├── contracts/
│   ├── v3-to-v21-mapping.md   # the whole v3 → v2.1 translation
│   └── import-service.md      # Detect/Clashes/Import/MCP results, errors, labels for v3
├── checklists/
└── tasks.md
```

### Source code (additions and changes)

```text
internal/domain/
├── enums/importkind/value.go            CHANGE: + PostmanRequest, PostmanWorkspace
├── enums/warningcode/value.go           CHANGE: + ProtocolScript, LinkedSecret
├── entity/import_probe.go               CHANGE: + PostmanKind, HasValues, HasWorkspaceField, DocumentID; v3 predicates
│                                                (R4/R5); replaces IsPostmanV3File/HasPostmanV3Entry
├── entity/import_input.go               CHANGE: + IsPostmanV3()
├── entity/import_collection.go          NEW: Root, Warnings, Secrets
└── entity/postman_v3_source.go          NEW: Collections, Request, Environments, Globals, Warnings, Skipped + SkippedWarnings()

internal/application/
├── usecase/detect_import_format.go      CHANGE: R4 rules; postmanV3Msg removed
├── usecase/read_postman_collection.go   CHANGE: returns *entity.ImportCollection
├── usecase/read_postman_v3.go           NEW port: ReadPostmanV3(ctx, ImportInput) (*PostmanV3Source, error)
├── service/import_readers.go            CHANGE: + readPostmanV3; IsPostmanV3() routing; postmanV3(); collection secrets
├── service/import_read.go               CHANGE: addCollection(kind, ImportCollection) sets CollectionSecrets
└── service/manage_import.go             CHANGE: constructor takes ReadPostmanV3

internal/integration/
├── adapter/postman_webservice.go        CHANGE: embeds usecase.ReadPostmanV3
├── webservice/postman.go                CHANGE: ReadCollection returns ImportCollection (secrets from dto)
├── webservice/postman_v3.go             NEW: ReadPostmanV3, dispatch by input.Kind, v2.1 reuse, IMP-11400
├── webservice/postman/dto/collection.go CHANGE: secret collection variables → Secret rows + secret map (R7)
├── webservice/postman/dto/key_value.go  CHANGE: IsSecret(); ToVariableRow marks secret, blanks value
├── webservice/postmanv3/dto/            NEW: collection.go, item.go, request.go, example.go, auth.go, credentials.go,
│                                             key_values.go, body.go, script.go, variable.go, environment.go,
│                                             definition.go, tree.go (folder → nested collection, R10/R11),
│                                             workspace.go (split collections/environments/globals/skipped),
│                                             conversion.go (R8 warnings), postman_v21.go (v2.1 builders)
├── persistence/import_source.go         CHANGE: .resources allowed; postman/ re-root; request-file sibling examples;
│                                                 YAML probing
├── persistence/model/import_probe.go    CHANGE: + Kind($kind), Values, Workspace, ID (json+yaml tags)
└── entrypoint/controller/mcp_import.go  CHANGE: tool description text

internal/infra/dependency/application/manage_import.go   CHANGE: pass GetPostmanWebservice() as ReadPostmanV3

frontend/src/features/import-export/
├── lib/labels.ts (+ labels.test.ts)                     CHANGE: postman_request, postman_workspace labels
└── components/ImportPreviewStep.tsx (+ test)            CHANGE: ASSIGNMENT_KINDS includes postman_workspace

docs/prd/xenon-mvp-prd.md                                CHANGE: "What can be imported" + scope line (~:301-307, :495)

test/integration/
├── features/import/format_detection.feature            CHANGE: v3 rows detected; new v3 detect rows
├── features/import/postman_v3_collection.feature        NEW
├── features/import/postman_v3_environment_globals.feature NEW
├── features/import/postman_v3_workspace.feature         NEW
├── features/import/postman_collection.feature           CHANGE: + v2.1 secret collection variable → vault scenario
├── features/import/large_collection.feature             CHANGE: + v3 1,000-request scenario
├── features/mcp/other_tools.feature                     CHANGE: + import tool on a v3 workspace folder
└── resources/import/postman/v3/                          NEW fixtures (R14) + README section
```

**Structure Decision**: This is the single-repo Wails layout. The backend follows the 4-layer skills and the frontend
follows `frontend-architecture`. The v3 contract package sits beside `postman/dto` and `bruno/dto`, the existing
precedent for per-format DTO packages inside `integration/webservice`.

## BDD Coverage (what `/create-tests` writes)

All scenarios are tagged `@all @xen_009 @import` plus the feature tag.

| Feature file | Requirements | Scenarios (outline) |
|---|---|---|
| `import/format_detection.feature` (CHANGE) | FR-001–FR-007 | Move the `postman v3 folder` and `postman v3 file` rows to the `@detect` outline (`postman_collection 3`, `postman_request 3`). Add rows: v3 JSON, env YAML, env JSON, globals YAML, globals JSON (cloud `workspace` field), workspace folder, `postman/` folder, `postman/collections` folder, workspace zip. Add a pasted v3 JSON row. Every pre-existing row is unchanged (FR-007). |
| `import/postman_v3_collection.feature` (NEW) | FR-008–FR-015, FR-017, FR-018, FR-023–FR-025, SC-001, SC-004 | (1) v3 JSON, v3 folder and v2.1 of `rich` import into three workspaces whose collection folders are equal apart from ids (research R14). The v3-only cases use `extras/` and are asserted field by field. (2) Bodies per type, with disabled rows and the missing file → repick warning. (3) Auth at three levels: inherit, noauth, map/list credentials. (4) Scripts at three levels and graphql-request scripts. (5) Examples inline (JSON) and as files (YAML), with order. (6) Order rules: `order`, missing `order`, ties by name. (7) Names from `name` versus file and folder names. (8) Multi-auth list with 2 entries → first imported + `unsupported_auth` warning. (9) `graphql:` collection script → `protocol_script` warning, kept for export. (10) grpc/websocket requests → `unsupported_request_type` warnings. (11) Folder variables → `folder_variables` warning. (12) `settings.maxRedirects` kept (export round-trip shows `protocolProfileBehavior.maxRedirects`). (13) A single `*.request.yaml` imported into a folder with its example; top-level is refused `IMP-35400`. (14) Malformed YAML → `IMP-11400` naming the file. (15) Clash keep_both / replace / skip on a v3 collection. (16) Secret collection variable (`secret: true`) → vault, not in the collection file. |
| `import/postman_v3_environment_globals.feature` (NEW) | FR-016, FR-019, FR-020, FR-023, SC-003 | (1) Env YAML: plain, disabled and secret (`type: secret` and `secret: true`) rows; vault reveal; no value in files or logs. (2) Env JSON from the CLI: `********` mask → secret missing; cloud fields kept in passthrough. (3) Vault-linked row → missing + `linked_secret` warning; `source` kept. (4) Globals YAML and JSON merge into workspace globals with clash rules. (5) Name falls back to the file stem. |
| `import/postman_v3_workspace.feature` (NEW) | FR-005, FR-021, FR-022, SC-002 | (1) Detect lists 2 collections, 2 environments and globals, in order. (2) Per-item targets across two workspaces. (3) Skipped `specs`/`flows` parts reported. (4) Repo-root pick reads only `postman/` (a non-UTF-8 `tools/latin1.json` outside `postman/` does not trip IMP-01400). (5) `postman/collections` pick → collections only. (6) Zip of the workspace. |
| `import/postman_collection.feature` (CHANGE) | FR-019 (R7) | A v2.1 collection variable `type: secret` goes to the vault, with `secret: true` in the collection file. |
| `import/large_collection.feature` (CHANGE) | SC-006, FR-024 | `postman/v3/large-1000-v3.zip`: Detect < 2 s, Import < 8 s, progress at most 250 ms apart. |
| `mcp/other_tools.feature` (CHANGE) | FR-025 | The `import` tool `detect` then `run` on the v3 workspace fixture returns `postman_workspace`/`3`, imports every item, and gives the same warnings. |

The existing step catalog covers every assertion above:

- `I call "ImportService.…" with`;
- response field checks;
- workspace file fields;
- `the folder … should equal the folder … apart from ids`;
- `no file under the workspace folder should contain …`;
- `the log output should not contain …`;
- `VaultService.Reveal`;
- `the file … exists with content`;
- `the last call should have taken less than …`.

No new step is needed.

## Complexity Tracking

| Item | Why needed | Simpler alternative rejected because |
|---|---|---|
| v2.1 behaviour change: `type: secret` collection variables go to the vault (R7) | The v3 path reuses the v2.1 mapping, and FR-019 and constitution II forbid the secret value in the collection file. | A v3-only special case would leave v2.1 leaking the same value and fork the mapping that R2 deliberately shares. |
| `readImportFolder` re-roots to `postman/` | Without it, picking a repository walks every JSON and YAML file in it and can fail on an unrelated large or non-UTF-8 file. | Requiring users to pick `postman/` contradicts US4's "or the repository folder that contains it". |
