---
description: "Task list for XEN-006 — Export (Stage 6)"
---

# Tasks: Export (Stage 6)

**Input**: `specs/XEN-006-export/` · **Prerequisite**: XEN-001..005 implemented.

**Tests**: MANDATORY. Feature files tagged `@all @xen_006 @<area>`, written first and proven red.
Formats per [contracts/export-formats.md](contracts/export-formats.md).

US1 Postman export · US2 Xenon YAML export · US3 Xenon YAML import · US4 dialog & summary

---

## Phase 1: Setup

- [x] T001 [P] Vendor the Postman collection v2.1.0 JSON schema into `test/integration/resources/postman-collection-v2.1.0.schema.json`; add `tools/newman/package.json` (pinned `newman`) and Makefile target `newman` (export corpus → run with Newman against ApiMock started by the BDD harness in a helper mode)
- [x] T002 [P] Test helper `test/integration/steps/normalise_postman_test.go` implementing export-formats §4 plus generic steps `the exported Postman file should equal "<fixture>" after normalisation`, `the exported file should validate against the Postman v2.1 schema`, `the export file "<path>" should not contain "<text>"`

---

## Phase 2: Foundational

- [x] T003 [P] Enums + entities `export_request.go`, `export_summary.go`, enums `export_format`, `export_shape`, `export_item_kind`
- [x] T004 `internal/integration/persistence/export_file.go` (temp file beside target, zip streaming with UTF-8/forward-slash names, atomic rename, cleanup on cancel/failure); service skeleton `internal/application/service/export.go` + controllers/bindings (`Options`, `PickTarget`, `Export`, `ShowInFolder`) + event `export:progress`
- [x] T005 [P] Pure use cases `effective_export_root.go`, `apply_current_values_as_initial.go`, `find_literal_credentials.go` with table tests

---

## Phase 3: User Story 1 — Postman export (P1) 🎯 MVP

- [x] T006 [P] [US1] BDD `test/integration/features/export/roundtrip_corpus.feature`: every stage-3 Postman fixture: import → export → equal after normalisation (SC-001); exports validate against the v2.1 schema
- [x] T007 [P] [US1] BDD `test/integration/features/export/postman_collection.feature` and `postman_folder_request.feature`: Xenon-created collection exports (all request features) valid; folder/request exports carry effective auth, scripts (respecting inherit toggles) and collection variables at the root
- [x] T008 [P] [US1] BDD `test/integration/features/export/postman_environment_globals.feature` and `postman_workspace_zip.feature`: secrets `type: secret` + empty value (0 secret bytes in output — SC-004); current-values option; globals JSON; zip layout
- [x] T009 [P] [US1] BDD `test/integration/features/export/postman_dropped_fields.feature`: every FR-005 item dropped and listed in the summary; unsupported auth restored (postman) / omitted with note (bruno); disabled scripts → `event.disabled`; disabled body restored; binary example body → empty + note; literal credential warning
- [x] T010 [US1] Run T006–T009 → **red**
- [x] T011 [US1] `webservice/postman/writer.go` + dto `New{X}` constructors (inverse mapping + passthrough merge per research R1), environment/globals writers, workspace zip; wire into `export.go` → **green**
- [x] T012 [US1] CI `make newman`: every Xenon-created corpus collection exported and run with Newman against ApiMock; requests received equal Xenon's sends (SC-002)

---

## Phase 4: User Story 2 — Xenon YAML export (P1)

- [x] T013 [P] [US2] BDD `test/integration/features/export/xenon_yaml_levels.feature`: workspace zip (excludes `.xenon/`, staging, temp files; no current values/tokens/cookies), collection/folder zip with manifest, bundle format (children order, nested folders), request/environment files
- [x] T014 [P] [US2] BDD `test/integration/features/export/xenon_secrets.feature` and `current_values.feature`: include secrets → armored age block (single files) or `xenon-secrets.age` (zips), passphrase rules `EXP-01400`, without option → no secret bytes; current values written as initial only in the export, workspace unchanged
- [x] T015 [US2] Run → **red**; implement `webservice/xenonyaml/writer.go`, `secrets.go` (age scrypt, logN from properties) → **green**

---

## Phase 5: User Story 3 — Xenon YAML import (P1)

- [x] T016 [P] [US3] BDD `test/integration/features/import/xenon_yaml_import.feature`: each kind detected; collection/folder/request/environment imported with new ids and clash handling; workspace zip/folder → new workspace in an empty target folder (ids kept); encrypted secrets with right passphrase → vault, wrong → `IMP-08401`, skip → missing; newer format version refused; Xenon YAML → import → export byte-identical (FR-015, SC-003)
- [x] T017 [US3] Run → **red**; implement `webservice/xenonyaml/reader.go` + secrets decryption + id remapping, ImportService passphrase/target-folder handling → **green**

---

## Phase 6: User Story 4 — Dialog and summary (P2)

- [x] T018 [P] [US4] BDD `test/integration/features/export/atomic_cancel.feature`: cancel mid-export leaves no file; write failure leaves no file and reports `EXP-02500`; progress events for > 200 items
- [x] T019 [US4] Run → **red** then **green** with T004 refinements
- [x] T020 [US4] UI `features/export/`: `ExportDialog` (format, shape per item kind from `Options`, include secrets hidden for Postman with explanation, passphrase ×2, current-values checkbox, save location), `ExportSummary` (counts, dropped fields, warnings, Show in folder); "Export…" in sidebar and environment context menus; File → Export; Import dialog passphrase prompt; RTL tests

---

## Phase 7: Polish

- [x] T021 [P] Performance BDD: 1,000-request workspace export < 5 s (SC-005)
- [x] T022 `/refactor-diff` until clean; `make validate`; `make newman`; CI on 3 OSes; walk [quickstart.md](quickstart.md) — the manual quickstart walk is deferred to the user's manual test session

---

## Dependencies & Execution Order
Setup → Foundational → US1 → US2 → US3 (needs US2 writer for round-trips) → US4 → Polish.

## Parallel opportunities
T001–T002; T003, T005; [P] BDD files per story; US4 UI alongside US2/US3 backend.

## Implementation Strategy
MVP = US1 (free Postman export with fidelity) → US2 + US3 (Xenon YAML out and in, secrets) →
US4 → Polish.

---

## Implementation notes (T001–T012)

- T001: schema vendored as published (draft-04). `tools/newman` pins newman 6.2.2 (6.2.3 was under two
  weeks old). `make newman` = `npm ci` in tools/newman + the `@newman` BDD scenario with
  `XENON_NEWMAN=on`; the harness itself is the "helper mode" (no separate server process).
- T002: the normaliser lives in `steps/normalise_postman_test.go`; generic steps in
  `then_export_file_steps_test.go` are path-generic (`the file "<path>" …` under the temp root, zip
  aware) rather than "the exported Postman file". Schema validation uses
  `github.com/santhosh-tekuri/jsonschema/v6` v6.0.2 (test-only dependency). Normaliser rules beyond
  §4, all "defaults written by Postman vs absent": a string `request` is `{method: GET, url}`, methods
  compared upper-cased, `options.raw.language: text`, auth attributes compared as a key set with
  their `type` annotations and Postman's attribute defaults (apikey `in: header`, digest `MD5`, oauth2
  `addTokenTo`/`client_authentication: header`, …) ignored, and literal OAuth `accessToken`/`tokenType`
  (moved to the vault on import, never exported) left out.
- T003: also `export_secrets` (summary secrets), `ExportScope`, `ExportNode`, `ExportEntry`,
  `ExportProgress`, `ExportFormatOption`; `literal_credential` warning code.
- T004: the service is `service/export.go` + `export_items.go` (loading) + `export_postman.go`
  (documents); `Options` returns `suggested_file_name` (base) plus one per format; `PickTarget`
  opens the native save dialog with the name the UI passes. Show in folder goes through a Wails file
  manager adapter (`app.Env.OpenFileManager`) mocked in BDD. Streaming writes use the new
  `fileutils.WriteStreamAtomic`. Xenon YAML export answers `EXP-03400` until US2 (T015).
- T005: ancestor headers (Postman has no collection/folder headers) are copied onto every request
  of the export and listed as changed; the drop scan is `FindPostmanDrops`.
- T009/T011 deviations: Bruno-only auth is exported as `noauth` (spec edge case) rather than omitted
  (contract table), so the request still sends without auth as in Xenon. Collection secret variables
  are exported as empty `string` variables and listed in the summary, because the official v2.1.0
  schema has no `secret` collection variable type (environments/globals keep `type: secret`). The
  full stage-3 fixtures (`full-v2.1.json`, `full-v2.0.json`) use auth types newer than the v2.1.0
  schema (`jwt`, `asap`), so their round trip is checked for equality only; every other export is
  also schema-validated. Item ids: the passthrough id, else the Xenon id only for Xenon-created
  trees (otherwise a Postman item without an id would gain one).
- T011 fix outside the feature: leftover auth attributes in a typed sub-object (e.g. an OAuth 2.0
  `accessToken: {{token_var}}` reference) were written to the YAML but not loaded back; the model
  now loads them into `Auth.Attributes` so they survive into the export (SC-001).
- T012: CI job `newman` runs on macOS, Windows and Linux. Newman matches Xenon on method, path,
  query, body (form/JSON aware) and request/auth/script headers for the whole shop corpus; AWS v4
  is compared by shape. With the url host/path split disabled the scenario fails, so it is sensitive.
- Mutation: `gremlins --diff` in unit mode over the XEN-006 code (BDD-only service/infra/entrypoint/
  publisher files excluded) kills 201/201 covered mutants. Locally the integration mode killed every
  mutant within seconds, including a deliberately equivalent one, so it was not used as a signal.

## Implementation notes (T013–T017)

- Placement deviation (plan: `webservice/xenonyaml/{writer,reader,secrets}.go`): Xenon YAML is the
  workspace's own on-disk format, owned by `persistence/model`, so the writer, reader and secrets
  codec live in `persistence/xenon_export.go` + `persistence/xenon_import.go` (adapter
  `XenonExportPersistence`: `WriteXenonExport`, `EncryptExportSecrets`, `ReadXenonExport`,
  `DecryptExportSecrets`) instead of duplicating every model in a webservice DTO package.
  `pkg/agecrypt` gained `EncryptArmoredWithPassphrase` / `DecryptArmoredWithPassphrase`.
- T013/T015 export: files are exported byte for byte from disk; only variable files whose rows
  change with "current values as initial" are re-encoded (`variables:` patched, rest kept).
  Every zip (workspace too) has `xenon-export.yaml` at its root next to one folder named after the
  workspace/collection/folder directory; `xenon-secrets.age` sits beside the manifest. Left out of
  zips: `.xenon/`, `.xenon-import-*/`, `*.xenon-tmp` and also `.git/` (local repository metadata).
  A request export never carries collection secrets (its file has none); a folder export carries
  no collection scope either (collection.yaml is not in it). Summary `secrets`: `included` only
  when at least one value was encrypted, else `omitted`. The bundle is named after the item and
  always has a `secrets:` key (`null` without secrets).
- T016/T017 import: `Detect` reads a `xenon/kind/root` header from YAML files, from the manifest
  of a zip or from a top-level `workspace.yaml`; kinds `xenon_workspace_zip`,
  `xenon_workspace_folder`, `xenon_zip`, `xenon_bundle`, `xenon_request`, `xenon_environment`;
  preview `encrypted_secrets`. Newer format → `IMP-42422`. A missing passphrase is `IMP-50401`
  (distinct from the contract's wrong-passphrase `IMP-08401`) so the UI can tell "ask" from "retry";
  `skip_secrets: true` imports them as missing. Decryption happens before anything is written.
  New ids for every item, example and environment; the export id is kept on the item to find its
  secrets (environment → XEN-003 vault path; collection → its new id when imported at the top
  level; into a parent the values are not kept). A folder export imported into a parent stays a
  folder (no Postman-style merge/wrap); at the top level it becomes a collection.
  Workspace imports take `items[].target_folder` (then `target_local_id` is not required): the
  folder must be empty or new (`WKS-01409`, inside another workspace `WKS-02409`, none
  `IMP-43400`); files are copied into a staging sibling then renamed, registered with ids kept,
  and secrets written to the new workspace's vault. Byte identity (FR-015/SC-003) is asserted as
  export → import → export → import → export with the last two exports equal apart from ULIDs,
  since a hand-written first file need not be in Xenon's canonical encoding.
- The detection refusal now reads "is not a Xenon, Postman, Bruno or cURL export".
- Frontend: only the import mapper changed (new plan/preview fields from the regenerated
  bindings); the passphrase / target-folder UI is T020.
- Mutation: gremlins unit mode over the T013–T017 diff (BDD-only service/infra/entrypoint/
  publisher files excluded) kills 180/180 covered mutants.
- CI: the XEN-004 parity case `pm-test-async-done` reordered its async tests once on a loaded
  macOS runner (timers of 1/5/10 ms on separate Go timers); its delays are now 1/60/120 ms, with
  the oracle expectation unchanged.

## Implementation notes (T018–T021)

- T018/T019: `features/export/atomic_cancel.feature`. Red before the change: the progress and
  mid-export cancel scenarios (only a start and an end event existed, so a cancel on the 2nd/3rd
  event came after the file was complete). Already green before the change and kept as guards:
  the write-failure scenarios (`EXP-02500`, temp file removed; the target is an existing folder
  so the rename fails portably), the already-cancelled export and the small-export events.
  Progress rule (`ExportProgress.IsMilestone`): an archive of more than 200 files announces
  `export:progress` after every 50th file, plus the start (0) and the end (total); a cancel
  between files stops the zip, removes the temp file and keeps a file it would have replaced
  (`EXP-05499`). `SaveArchive` takes a `written(done)` callback. New generic step
  `I call "<Service>.<Method>" and cancel it on the <n> "<event>" event with` (the event mock
  runs a hook synchronously on the n-th emission, so the cancel is deterministic on every OS).
- T020 UI: the export dialog is its own feature folder `features/export/` (own store, as the
  plan says) next to `features/import-export/`. Entry points: Export… in the sidebar context
  menu (collections, folders, requests; not examples), the environment context menu and an
  Export… button in the environment editor, Export Workspace… in the workspace switcher (every
  OS) and the macOS File menu (`export:requested`; no accelerator, Cmd+Shift+E is the
  environment switcher). Save location: `PickTarget` now takes the shape and suggests the
  matching Xenon YAML extension (`.zip`/`.yaml`, BDD in `postman_collection.feature`).
  Progress bar only when the export has more than 200 files; Cancel returns to the options
  with "no file was written". The summary shows counts, the secrets line (when the item has
  secrets), dropped/changed fields, warnings and Show in Folder. Import dialog: `IMP-50401`
  asks for the passphrase, `IMP-08401` asks again (wrong), "Import Without Secrets" sends
  `skip_secrets`; a workspace export row offers Choose Folder… instead of a workspace and
  location, sends `target_folder`, and the workspace list is reloaded after the import. The
  passphrase lives in the store only while the dialog is open and is cleared after use.
  Globals have no Export… entry of their own (the workspace export carries them).
- T021: `features/export/large_workspace.feature`, 1,000 requests across 4 collections, Xenon
  YAML and Postman zips each under 5 s (about 0.3 s locally); green on the first run (the
  writer already met SC-005), so it is a guard.
- Chunk 2 leftovers (BDD first, both red before the fix): a collection export imported into
  another collection merges its variables into that collection, so its encrypted secret
  values now go to that collection's vault scope, following the clash resolution like the
  globals import (shared pure use case `MergeImportedSecrets`; keep_both/skip keep the
  existing value, replace overwrites, other secrets of the target untouched). A workspace
  import that fails after copying (load, vault or registry) removes the copied folder and the
  new vault, registers the workspace last, and the result no longer carries a workspace entry
  with an empty id (only `items[].status: failed` and `failed[]`).
- Mutation: gremlins unit mode over this chunk's diff kills 19/19 covered mutants.

## Implementation notes (T022)

- `/refactor-diff` over the XEN-006 diff (`xen-005...HEAD`), three passes, the fourth scan
  clean: export/import rules moved onto entities (`ExportPickerTitle`,
  `Environment.ExportCounts`, `NewVariablesExportCounts`, `WellFormedEnvironments`,
  `ExportableChildren`, `ExportPath`, `ExportNode.ID/Auth/HasContainerHeaders/ScopeRefWithin`,
  `XenonExport.IsWorkspace/IsEnvironment/ExportedSecrets`, `XenonImport.IsWorkspace/
  IsEnvironment`, `ImportPlan.HasPassphrase`, `ImportPlanItem.HasTargetFolder`,
  `ImportItem.HasCollectionSecrets`); `exportDocuments`, `ExportBundleChildren` and the import
  passphrase unlock in their own files; manifest, bundle and Postman environment/globals
  constructors format the export time themselves (formats unchanged); Xenon zip read failures
  get unique codes `IMP-51422` (collection/folder) and `IMP-52422` (request); archive entry
  times come from `timeutils`.
- Left as is: `IMP-01400` is shared by detection and the non-UTF-8 read because the XEN-003
  contract names it for both; `exportedAt` stays RFC 3339 without milliseconds because
  `contracts/export-formats.md` fixes it.
- `make validate` green locally with the mutation step left to CI (gremlins unit mode over the
  changed domain, model and Postman DTO packages: no surviving mutant in XEN-006 code);
  `make bdd`, `make parity`, `make newman`, `make frontend-check` and `make build` green.
- Manual quickstart walk: deferred to the user's manual test session.
