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

- [ ] T001 [P] Vendor the Postman collection v2.1.0 JSON schema into `test/integration/resources/postman-collection-v2.1.0.schema.json`; add `tools/newman/package.json` (pinned `newman`) and Makefile target `newman` (export corpus → run with Newman against ApiMock started by the BDD harness in a helper mode)
- [ ] T002 [P] Test helper `test/integration/steps/normalise_postman_test.go` implementing export-formats §4 plus generic steps `the exported Postman file should equal "<fixture>" after normalisation`, `the exported file should validate against the Postman v2.1 schema`, `the export file "<path>" should not contain "<text>"`

---

## Phase 2: Foundational

- [ ] T003 [P] Enums + entities `export_request.go`, `export_summary.go`, enums `export_format`, `export_shape`, `export_item_kind`
- [ ] T004 `internal/integration/persistence/export_file.go` (temp file beside target, zip streaming with UTF-8/forward-slash names, atomic rename, cleanup on cancel/failure); service skeleton `internal/application/service/export.go` + controllers/bindings (`Options`, `PickTarget`, `Export`, `ShowInFolder`) + event `export:progress`
- [ ] T005 [P] Pure use cases `effective_export_root.go`, `apply_current_values_as_initial.go`, `find_literal_credentials.go` with table tests

---

## Phase 3: User Story 1 — Postman export (P1) 🎯 MVP

- [ ] T006 [P] [US1] BDD `test/integration/features/export/roundtrip_corpus.feature`: every stage-3 Postman fixture: import → export → equal after normalisation (SC-001); exports validate against the v2.1 schema
- [ ] T007 [P] [US1] BDD `test/integration/features/export/postman_collection.feature` and `postman_folder_request.feature`: Xenon-created collection exports (all request features) valid; folder/request exports carry effective auth, scripts (respecting inherit toggles) and collection variables at the root
- [ ] T008 [P] [US1] BDD `test/integration/features/export/postman_environment_globals.feature` and `postman_workspace_zip.feature`: secrets `type: secret` + empty value (0 secret bytes in output — SC-004); current-values option; globals JSON; zip layout
- [ ] T009 [P] [US1] BDD `test/integration/features/export/postman_dropped_fields.feature`: every FR-005 item dropped and listed in the summary; unsupported auth restored (postman) / omitted with note (bruno); disabled scripts → `event.disabled`; disabled body restored; binary example body → empty + note; literal credential warning
- [ ] T010 [US1] Run T006–T009 → **red**
- [ ] T011 [US1] `webservice/postman/writer.go` + dto `New{X}` constructors (inverse mapping + passthrough merge per research R1), environment/globals writers, workspace zip; wire into `export.go` → **green**
- [ ] T012 [US1] CI `make newman`: every Xenon-created corpus collection exported and run with Newman against ApiMock; requests received equal Xenon's sends (SC-002)

---

## Phase 4: User Story 2 — Xenon YAML export (P1)

- [ ] T013 [P] [US2] BDD `test/integration/features/export/xenon_yaml_levels.feature`: workspace zip (excludes `.xenon/`, staging, temp files; no current values/tokens/cookies), collection/folder zip with manifest, bundle format (children order, nested folders), request/environment files
- [ ] T014 [P] [US2] BDD `test/integration/features/export/xenon_secrets.feature` and `current_values.feature`: include secrets → armored age block (single files) or `xenon-secrets.age` (zips), passphrase rules `EXP-01400`, without option → no secret bytes; current values written as initial only in the export, workspace unchanged
- [ ] T015 [US2] Run → **red**; implement `webservice/xenonyaml/writer.go`, `secrets.go` (age scrypt, logN from properties) → **green**

---

## Phase 5: User Story 3 — Xenon YAML import (P1)

- [ ] T016 [P] [US3] BDD `test/integration/features/import/xenon_yaml_import.feature`: each kind detected; collection/folder/request/environment imported with new ids and clash handling; workspace zip/folder → new workspace in an empty target folder (ids kept); encrypted secrets with right passphrase → vault, wrong → `IMP-08401`, skip → missing; newer format version refused; Xenon YAML → import → export byte-identical (FR-015, SC-003)
- [ ] T017 [US3] Run → **red**; implement `webservice/xenonyaml/reader.go` + secrets decryption + id remapping, ImportService passphrase/target-folder handling → **green**

---

## Phase 6: User Story 4 — Dialog and summary (P2)

- [ ] T018 [P] [US4] BDD `test/integration/features/export/atomic_cancel.feature`: cancel mid-export leaves no file; write failure leaves no file and reports `EXP-02500`; progress events for > 200 items
- [ ] T019 [US4] Run → **red** then **green** with T004 refinements
- [ ] T020 [US4] UI `features/export/`: `ExportDialog` (format, shape per item kind from `Options`, include secrets hidden for Postman with explanation, passphrase ×2, current-values checkbox, save location), `ExportSummary` (counts, dropped fields, warnings, Show in folder); "Export…" in sidebar and environment context menus; File → Export; Import dialog passphrase prompt; RTL tests

---

## Phase 7: Polish

- [ ] T021 [P] Performance BDD: 1,000-request workspace export < 5 s (SC-005)
- [ ] T022 `/refactor-diff` until clean; `make validate`; `make newman`; CI on 3 OSes; walk [quickstart.md](quickstart.md)

---

## Dependencies & Execution Order
Setup → Foundational → US1 → US2 → US3 (needs US2 writer for round-trips) → US4 → Polish.

## Parallel opportunities
T001–T002; T003, T005; [P] BDD files per story; US4 UI alongside US2/US3 backend.

## Implementation Strategy
MVP = US1 (free Postman export with fidelity) → US2 + US3 (Xenon YAML out and in, secrets) →
US4 → Polish.
