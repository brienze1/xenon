# Implementation Plan: Import and Saved Examples (Stage 3)

**Branch**: `xen-003` (stacked on `xen-002`) | **Date**: 2026-10-09 | **Spec**: [spec.md](spec.md)

## Summary

An import pipeline (detect → parse into Xenon entities + passthrough + warnings → plan with
targets and clash resolutions → atomic per-item write) for Postman collections v2.0/v2.1,
environments, globals and data dumps, Bruno `.bru` and OpenCollection YAML (with bru→pm script
translation), and cURL; a pm-usage scanner in every summary; and saved examples stored in
request files with sidebar, tabs and "Save response as example". Lossless passthrough blocks keep
everything stage 6 needs for Postman round-trip.

## Technical Context

**Language/Version**: Go 1.25+, TypeScript 5 + React 19

**Primary Dependencies** (new): `mvdan.cc/sh/v3` (shell tokenizing for cURL); stdlib
`archive/zip`, `encoding/json`; existing `go.yaml.in/yaml/v3`

**Storage**: XEN-001/002 formats + additions ([contracts/format-additions.md](contracts/format-additions.md));
temp dirs for unpacked zips; staging dirs `.xenon-import-*/` inside the target workspace

**Testing**: godog BDD over a fixture corpus in `test/integration/resources/import/` (Postman
v2.0/v2.1 collections covering every mapped field, environments, globals, a dump; Bruno `.bru` and
YAML collections generated with Bruno 3.1; ≥ 30 cURL commands in bash/cmd/PowerShell); Go unit
tests for parsers, translator and tokenizer; Vitest + RTL

**Target Platform / Project Type**: desktop-app on macOS, Windows, Linux

**Performance Goals**: SC-002 — 1,000-request collection imported < 10 s; progress events
≤ 250 ms apart; UI responsive

**Constraints**: no secret values in workspace files (Postman secrets → vault); atomic per item;
zip safety (size cap 1 GB, zip-slip refused); nothing written before the user confirms the plan

**Scale/Scope**: dumps of ~50 collections / ~20 environments; collections up to a few thousand
requests

## Constitution Check

- [x] **I. Local-first** — imports read local files/text; results are plain YAML in the chosen
  workspace; staging inside the workspace folder, git-ignored. ✅
- [x] **II. Secrets** — Postman `secret` values go to the vault (XEN-002 multi-store save);
  import summaries, warnings, logs and passthrough never contain secret values (passthrough for
  environment rows excludes `value` of secret rows). ✅
- [x] **III. Go owns logic** — detection, parsing, translation, mapping, clash detection,
  writing and pm scanning are Go; UI shows previews/plans/summaries. ✅
- [x] **IV. Architecture skills** — see Structure Decision (parsers as integration webservices of
  external formats; mapping in DTO `ToEntity`; pure use cases for translation/scanning). ✅
- [x] **V. BDD-first** — fixture-driven features per story; only the file system roots, picker,
  clock and emitter are faked. ✅
- [x] **VI. Postman parity** — lossless passthrough (R2), mapping contract, examples feature
  (user decision 2026-10-09). ✅
- [x] **VII. Cross-platform** — zip paths normalised, Windows reserved names via XEN-001 naming,
  cmd/PowerShell cURL syntax supported. ✅

**Post-design re-check**: all gates pass.

## Project Structure

### Documentation
```text
specs/XEN-003-postman-bruno-import/
├── plan.md · research.md · data-model.md · quickstart.md
├── contracts/
│   ├── format-additions.md   ← unsupported auth, postman:/bruno: passthrough, examples
│   ├── postman-mapping.md    ← field-by-field Postman → Xenon
│   ├── bruno-mapping.md      ← Bruno → Xenon, bru→pm table, assertion table
│   └── bindings.md           ← ImportService, ExampleService
├── checklists/
└── tasks.md
```

### Source code (additions to the XEN-001/002 map)
```text
internal/domain/
├── entity/   saved_example.go, example_request.go, example_response.go, import_source.go,
│             import_item.go, import_warning.go, import_plan.go, clash.go, import_result.go,
│             pm_usage.go; auth.go gains Unsupported
└── enums/    import_kind, import_item_kind, warning_code, clash_resolution, body_encoding
              (+ value pkgs); auth_type gains `unsupported`

internal/application/
├── usecase/  read_import_source, detect_import_format, stage_import_item, commit_import_item,
│             rollback_import_item, find_clashes; pure: scan_pm_usage, translate_bruno_script,
│             convert_bruno_assertions, convert_bruno_vars, derive_curl_request_name,
│             resolve_clash_name
├── adapter/  import_service, example_service
└── service/  import.go (detect/plan/import orchestration, progress via publisher),
              example.go

internal/integration/
├── adapter/  postman_reader, bruno_reader, curl_reader, import_staging_persistence,
│             example_persistence (via request persistence), import_progress_publisher
├── webservice/                      ← external-format contracts (see Structure Decision)
│   ├── postman/ collection.go, environment.go, dump.go, dto/ (v2.0+v2.1 structs with ToEntity
│   │            and passthrough extraction per postman-mapping.md)
│   ├── bruno/   bru_parser.go (line grammar), yaml_reader.go, dto/ (ToEntity per bruno-mapping.md)
│   └── curl/    normalize.go (bash/cmd/PowerShell), tokenizer.go (mvdan.cc/sh), flags.go (FR-024
│                table), dto/
├── persistence/ request.go/model extended for examples + passthrough + unsupported auth;
│                staging.go (staging dir create/move/remove)
├── entrypoint/controller/ import_*.go, example_*.go, request_parse_curl.go
├── entrypoint/dto/        import_preview.go, import_plan.go, import_result.go, example.go
└── publisher/             import_progress.go

internal/infra/
├── adapter/ zip.go (archive/zip with size cap + zip-slip guard), dialog.go gains multi-pick
└── server/binding/ import.go, example.go; event/events.go adds import:progress;
                    route/response_body.go serves `example:<id>` bodies

test/integration/
├── features/import/   postman_collection, postman_v20, postman_environment_globals,
│                      postman_dump, bruno_bru, bruno_yaml, bruno_scripts, curl, clashes,
│                      atomicity, pm_usage, format_detection
├── features/examples/ examples_crud, save_response_as_example, examples_large_body
└── resources/import/  fixture corpus (see Technical Context)

frontend/src/
├── features/import/     store.ts, components/ ImportDialog (drop zone, picker, paste),
│                        DumpAssignment (per-item workspace picker + inline new workspace),
│                        ClashPrompt (replace/keep both/skip + apply to all), ImportProgress,
│                        ImportSummary (counts, warnings with links, pm usage)
├── features/examples/   store.ts, components/ ExampleTab (request + response editors),
│                        SaveAsExampleDialog
├── features/sidebar/    example nodes under requests (+ context menu)
├── features/request-builder/ URL bar cURL paste → ParseCurl (undoable)
├── features/response-viewer/ "Save response as example"
└── platform/services/   import.ts, examples.ts
```

**Structure Decision**: the skills have no "importer" category. Parsing an external system's
data contract is what the webservice skill describes ("external API clients — contract logic
only, DTOs with ToEntity"), so the Postman/Bruno/cURL readers live in
`internal/integration/webservice/{postman,bruno,curl}/` with their own `dto/` packages; they read
files/text the application hands them (no HTTP). Script translation and pm scanning are
business rules over script text → pure application use cases. No new top-level folders.

## Key design decisions

1. **Common output**: readers return `[]ImportItem` holding ready Xenon entities (with new ULIDs,
   passthrough nodes and warnings) — the rest of the pipeline is source-agnostic.
2. **Two-phase write**: stage every included item into `.xenon-import-<source>/<item>/` using the
   normal persistence code, then commit by renaming into `collections/` (or atomic environment
   multi-store save); rollback deletes staging. Watcher ignores staging dirs.
3. **Clash detection before writing** (`Clashes`), so the UI asks once with "apply to all";
   `Import` refuses unresolved clashes (`IMP-04409`).
4. **Import into an existing folder**: top-level children re-parented; collection-level auth,
   headers, scripts applied to the target folder only when the folder has none of its own
   (otherwise kept on a new wrapper folder named after the collection) — recorded in the summary.
5. **Examples** live in `Request.Examples`; their tabs reuse request-builder and response-viewer
   components in an "example" mode (no Send).
6. **cURL in the URL bar**: the UI detects a leading `curl ` on paste, calls `ParseCurl`, applies
   the result to the tab draft as one undoable edit.
7. **pm usage**: `scan_pm_usage` runs on every imported script; the MVP subset list is a constant
   shared with stage 4's sandbox (one source of truth).

## Complexity Tracking

No constitution violations to justify.
