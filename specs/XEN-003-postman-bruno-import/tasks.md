---
description: "Task list for XEN-003 — Import (Postman, Bruno, cURL) and Saved Examples (Stage 3)"
---

# Tasks: Import and Saved Examples (Stage 3)

**Input**: `specs/XEN-003-postman-bruno-import/` · **Prerequisite**: XEN-001 and XEN-002 implemented.

**Tests**: MANDATORY (constitution V). Feature files tagged `@all @xen_003 @<area>`, written first
and proven red. Mapping assertions follow [contracts/postman-mapping.md](contracts/postman-mapping.md)
and [contracts/bruno-mapping.md](contracts/bruno-mapping.md) row by row.

**Rules**: backend per architecture skills in the XEN-001 map (readers in
`internal/integration/webservice/{postman,bruno,curl}/` per the plan's Structure Decision);
frontend per `frontend-architecture`.

US1 Postman collection · US2 Postman environments/globals · US3 data dump · US4 saved examples ·
US5 Bruno · US6 cURL

---

## Phase 1: Setup

- [x] T001 Add `mvdan.cc/sh/v3` to `go.mod`
- [x] T002 [P] Build the fixture corpus in `test/integration/resources/import/`: `postman/full-v2.1.json` (every mapped field: nested folders, all body modes incl. disabled body, all auth types incl. hawk/ntlm/jwt/asap, scripts at all levels incl. disabled and `pm.require`, collection/folder/request variables, examples with text bodies, protocolProfileBehavior, raw-string header, string `request`, object descriptions), `postman/full-v2.0.json` (same content, v2.0 auth maps), `postman/dev.postman_environment.json` (plain, disabled, secret with/without value), `postman/globals.json`, `postman/dump.zip` + unzipped `postman/dump/` (archive.json, 3 collections, 2 environments, one unknown part), `postman/v1.json`, `postman/v3/` (request.yaml layout); `bruno/bru-collection/` and `bruno/yaml-collection/` exported from Bruno 3.1 (folders, seq, all body modes, auth incl. ntlm, vars, assert, scripts, tests, docs, settings, environments with secrets); `curl/corpus.yaml` (≥ 30 commands, each in bash, cmd and PowerShell form with the expected request)
- [x] T003 [P] Add `.xenon-import-*/` to the `.gitignore` block writer and to the tree scanner/watcher ignore list (XEN-001 persistence)

---

## Phase 2: Foundational

- [x] T004 [P] Enums in `internal/domain/enums/`: `import_kind`, `import_item_kind`, `warning_code`, `clash_resolution`, `body_encoding`; `auth_type` gains `unsupported` (+ value pkgs)
- [x] T005 [P] Entities in `internal/domain/entity/`: `import_source.go`, `import_item.go`, `import_warning.go`, `import_plan.go`, `clash.go`, `import_result.go`, `pm_usage.go`, `saved_example.go`, `example_request.go`, `example_response.go`; `auth.go` gains `Unsupported`; `request.go` gains `Examples`
- [x] T006 Persistence: `scripts.disabled` and `srcState` (format additions §4–§5; send pipeline treats `repick` as missing), request model/files gain `examples:` (format additions §3, base64 for binary bodies), `postman:`/`bruno:` passthrough nodes on every kind (preserved via the XEN-001 node patcher), `unsupported` auth (§1) — unit tests round-trip each addition byte-stable
- [x] T007 Import pipeline core: `internal/infra/adapter/zip.go` (1 GB total uncompressed cap enforced while extracting, absolute/`..`/symlink entries refused → `IMP-03400`; single files > 200 MB → `IMP-02413`; non-UTF-8 → `IMP-01400`), use cases `read_import_source.go`, `detect_import_format.go` (research R9 order), staging persistence `internal/integration/persistence/staging.go` + use cases `stage_import_item.go`, `commit_import_item.go`, `rollback_import_item.go`, `find_clashes.go`, `resolve_clash_name.go`; service `internal/application/service/import.go` (Detect → cache 30 min → Clashes → Import with progress publisher `import_progress.go`, ctx cancel before writing); controllers/DTOs `import_*.go`; binding `internal/infra/server/binding/import.go` with `PickSources`; event `import:progress`
- [x] T008 [P] Pure use case `scan_pm_usage.go` with the MVP pm subset as a shared constant `internal/domain/enums/pmapi/value.go` (also used by stage 4), unit tests
- [x] T009 BDD `test/integration/features/import/format_detection.feature` (@import): each fixture detected as its kind; v1 → `IMP-01400`; v3 YAML → `IMP-00415` naming "Postman v3 YAML"; random JSON → `IMP-00415`; zip-slip and over-size zips refused; cURL text detected — red, then green with T007
- [x] T010 [P] Frontend `platform/services/import.ts`, `features/import/store.ts`, `components/ImportDialog.tsx` (Cmd/Ctrl+O in `app/hotkeys.ts`, File → Import… menu item and sidebar toolbar Import button, drop zone over the window, picker, paste box), `ImportProgress.tsx`, `ClashPrompt.tsx`, `ImportSummary.tsx` (counts, warnings with "open item", pm usage grouped); RTL tests

**Checkpoint**: sources detect, plans/clashes compute, writing is atomic — no format mapping yet.

> Implementation notes (T001–T016): the import service is `ManageImportService` (+ `import_targets.go`,
> `import_writer.go` helpers); readers follow the `{Subject}{Type}` adapter naming
> (`PostmanWebservice` in `webservice/postman/`, `ImportSourcePersistence` reads/caches sources,
> `ImportStagingPersistence` stages/commits/rolls back); the UI lives in `features/import-export/`
> (the existing feature folder from `frontend-architecture`) with `platform/services/import.ts`.
> Archive and file limits are properties (`XENON_IMPORT_ARCHIVE_LIMIT`, `XENON_IMPORT_FILE_LIMIT`).
> Existing workspaces get the `.xenon-import-*/` gitignore line only when their block is first
> written (XEN-001 never rewrites an existing block); staging folders are hidden and ignored by the
> watcher either way. The native File → Import… menu is macOS-only (global menu bar); Windows/Linux
> use Cmd/Ctrl+O and the sidebar button.

---

## Phase 3: User Story 1 — Postman collection (P1) 🎯 MVP

- [x] T011 [P] [US1] BDD `test/integration/features/import/postman_collection.feature`: import `full-v2.1.json` as new collection; assert tree order, every request row of postman-mapping.md (method incl. custom, url raw, query disabled/description, path variables, headers incl. raw string, each body mode, disabled body → none + passthrough, form-data missing files flagged), auth mapping incl. inherit/noauth/unsupported, scripts per level incl. disabled (`scripts.disabled`), file references stored with `srcState: repick` and blocked from sending until re-picked, collection variables, folder/request variables → passthrough + warnings, protocolProfileBehavior → settings + passthrough, examples count; every imported item has a new ULID and its Postman id in passthrough (FR-020)
- [x] T012 [P] [US1] BDD `test/integration/features/import/postman_v20.feature` (v2.0 gives the same Xenon result as v2.1) and `postman_into_folder.feature` (into existing folder: re-parenting, wrapper folder when target has its own auth/headers/scripts, collection variables merged, summary note)
- [x] T013 [P] [US1] BDD `test/integration/features/import/pm_usage.feature` and `atomicity.feature`: summary lists `pm.require` and pm APIs outside the subset with locations; a write failure mid-collection (read-only dir mock) leaves no trace of that collection while other items succeed; cancel before writing writes nothing; every summary warning's `item_key`/path resolves to an existing item (SC-007)
- [x] T014 [US1] Run T011–T013 → **red**
- [x] T015 [US1] `internal/integration/webservice/postman/dto/` (collection v2.0/v2.1 structs incl. flexible string/object fields) with `ToEntity` + passthrough extraction per postman-mapping.md; `collection.go` reader + adapter `postman_reader.go`; v2.0 auth normalisation → **green**
- [x] T016 [US1] UI: sidebar shows unsupported auth read-only in `AuthPanel.tsx` ("unsupported auth type (kept)"); summary links open items; RTL tests

---

## Phase 4: User Story 2 — Postman environments and globals (P1)

- [x] T017 [P] [US2] BDD `test/integration/features/import/postman_environment_globals.feature`: env rows (plain/disabled), secrets with value → vault (no workspace file contains the value; summary/warnings/logs don't either), secret without value → missing; globals merge; clash on env name (replace / keep both " (imported)" / skip) and on variable keys (keep existing, imported value recorded in summary — masked for secrets)
- [x] T018 [US2] Run → **red**; implement `webservice/postman/environment.go` + dto, environment passthrough (excluding secret values), commit through XEN-002 multi-store save → **green**

---

> Implementation notes (T017–T025): environments and globals are read by `PostmanWebservice`
> (`ReadEnvironment`, `ReadDump`) into an in-memory `ImportVariables` (rows without secret values +
> the secret values, never logged or in passthrough); `ManageImportService` plans them with the pure
> `PlanEnvironmentImport` / `PlanGlobalsImport` and writes them through `importVariablesWriter`
> (vault first, then the file; replace = delete the old file, save the new one, restore both on
> failure, then drop the old environment's local and secret values). Dump parts other than
> `collection/` and `environment/` become source-level `skipped_part` warnings. Examples are served
> by `ManageExampleService` (binding `ExampleService`), persisted through the request persistence's
> `SaveExamples`; the send pipeline keeps the unresolved sent request on the stored response for
> "Save response as example". UI-state tabs carry `kind` + `example_id`. The UI keeps the existing
> `features/import-export/` folder (`DumpAssignment`, `WorkspacePicker`, `NewWorkspaceForm`) and adds
> `features/examples/`.

## Phase 5: User Story 3 — Data dump (P2)

- [x] T019 [P] [US3] BDD `test/integration/features/import/postman_dump.feature`: zip and folder forms list all items; plan assigns items to three workspaces (one created inline via `WorkspaceService.Create` within the flow); each workspace gets exactly its items; unknown part skipped + listed; combined summary per workspace
- [x] T020 [US3] Run → **red**; implement `webservice/postman/dump.go` (archive.json + subfolders) and multi-workspace planning in `service/import.go` → **green**
- [x] T021 [US3] UI `features/import/components/DumpAssignment.tsx` (table: include, name, kind, counts, target workspace picker with "New workspace…" inline create via the workspaces store, bulk "assign selected to…"); RTL tests

---

## Phase 6: User Story 4 — Saved examples (P2)

- [x] T022 [P] [US4] BDD `test/integration/features/examples/examples_crud.feature` (@examples): Create/Get/Save/Rename/Duplicate/Delete; stored in the request file `examples:`; tree exposes example nodes with code; revision conflict `REQ-02409`
- [x] T023 [P] [US4] BDD `test/integration/features/examples/save_response_as_example.feature` and `examples_large_body.feature`: from a sent tab (unresolved request snapshot, received response incl. headers); 8 MB body stored and served via `body_url`; binary body base64 round trip; `EXM-02404` without a response
- [x] T024 [US4] Run → **red**; implement `service/example.go`, persistence via request persistence, controllers/DTOs `example_*.go`, binding `example.go`, response route `example:<id>` → **green**
- [x] T025 [US4] UI: sidebar example nodes + context menu (rename, duplicate, delete); `features/examples/components/ExampleTab.tsx` (request editors in example mode, response editor: status code/text, headers table, body CodeMirror with Pretty/Raw/Preview, Raw-first > 5 MB); tabs store supports `kind: example`; "Save response as example" in response viewer + `SaveAsExampleDialog.tsx`; RTL tests

---

## Phase 7: User Story 5 — Bruno (P2)

- [ ] T026 [P] [US5] BDD `test/integration/features/import/bruno_bru.feature` and `bruno_yaml.feature`: tree and seq order, every row of bruno-mapping.md §1–§2 and §5, secrets missing, gRPC/WS requests skipped + listed, mixed formats refused
- [ ] T027 [P] [US5] BDD `test/integration/features/import/bruno_scripts.feature`: each §3 translation row, each §4 assertion operator, vars → `pm.variables.set`, untranslated lines prefixed and listed with file:line; ≥ 90% translation rate on the corpus (SC-006)
- [ ] T028 [US5] Run → **red**
- [ ] T029 [US5] `internal/integration/webservice/bruno/bru_parser.go` (line grammar: blocks, dictionaries, `~`, `'''` multiline, annotations) with unit tests per block type; `yaml_reader.go`; `dto/` with ToEntity; adapter `bruno_reader.go`
- [ ] T030 [US5] Pure use cases `translate_bruno_script.go`, `convert_bruno_assertions.go`, `convert_bruno_vars.go` (tables from bruno-mapping.md §3–§4) with table tests → run US5 features **green**

---

## Phase 8: User Story 6 — cURL (P2)

- [ ] T031 [P] [US6] BDD `test/integration/features/import/curl.feature`: every corpus entry in bash/cmd/PowerShell yields the expected request (SC-005); each FR-024 flag; unknown flags in notices; non-HTTP refused `IMP-07400`; not-curl text `IMP-06400`; dialog import creates a named request in the chosen folder
- [ ] T032 [US6] Run → **red**; implement `webservice/curl/normalize.go`, `tokenizer.go` (mvdan.cc/sh), `flags.go`, dto + `derive_curl_request_name.go`; `RequestService.ParseCurl` controller/binding → **green**
- [ ] T033 [US6] UI: URL-bar paste detection (`features/request-builder`), apply as one undoable draft change, notices toast; RTL test

---

## Phase 9: Polish

- [ ] T034 [P] Performance BDD `features/import/large_collection.feature`: generated 1,000-request collection < 10 s, progress events ≤ 250 ms apart (SC-002)
- [ ] T035 [P] Clash BDD `features/import/clashes.feature`: every clash kind × resolution, apply-to-all, `IMP-04409` when unresolved
- [ ] T036 Run the importer on Luis's real Postman exports (SC-001 acceptance) and record the pm-usage report in `docs/prd/pm-usage-report.md` (settles the PRD item on pm usage beyond the MVP subset); `/refactor-diff` until clean; `make validate`; CI on 3 OSes; walk [quickstart.md](quickstart.md)

---

## Dependencies & Execution Order
Setup → Foundational → US1 → US2 → US3 (needs US1+US2 readers) ; US4 after Foundational (independent of readers, but US1 fills examples) ; US5 and US6 after Foundational (independent) → Polish.

## Parallel opportunities
Fixture corpus (T002) with everything; US4, US5, US6 can proceed in parallel after Phase 2; all
[P] BDD files within a story.

## Implementation Strategy
MVP = Foundational + US1 + US2 (Postman collections and environments in). Then US4 (examples),
US3, US5, US6, Polish.
