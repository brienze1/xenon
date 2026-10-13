# Implementation Plan: Export (Stage 6)

**Branch**: `xen-006` (stacked on `xen-005`) | **Date**: 2026-10-09 | **Spec**: [spec.md](spec.md)

## Summary

Postman export as the exact inverse of the stage-3 mapping plus passthrough merge; Xenon YAML
export (workspace zip, collection/folder zip or bundle, request/environment files) with optional
passphrase-encrypted secrets and the per-export current-values option; Xenon YAML import through
the stage-3 pipeline; fidelity proven by a normalised round-trip of the fixture corpus and Newman
runs of exported collections.

## Technical Context

**Language/Version**: Go 1.25+, TypeScript 5 + React 19

**Primary Dependencies**: existing (`filippo.io/age`, `archive/zip`, `go.yaml.in/yaml/v3`);
CI-only: `newman` (pinned) and the Postman collection v2.1.0 JSON schema

**Storage**: export files written by the user; no new persistent state

**Testing**: godog BDD; round-trip corpus with normaliser; JSON-schema validation; Newman in CI;
Vitest + RTL

**Target Platform / Project Type**: desktop-app, macOS/Windows/Linux

**Performance Goals**: SC-005 (1,000-request workspace export < 5 s)

**Constraints**: no secrets in Postman exports; Xenon exports include secrets only encrypted and
opted-in; atomic writes

**Scale/Scope**: workspaces up to thousands of requests

## Constitution Check

- [x] **I. Local-first** — free export/import at every level, plain files. ✅
- [x] **II. Secrets** — Postman: names only; Xenon: opt-in, passphrase-encrypted, self-contained;
  tokens/cookies/current values never exported (current values only when asked, plain rows
  only). ✅
- [x] **III. Go owns logic** — writers, readers, encryption in Go; UI is the dialog. ✅
- [x] **IV. Architecture skills** — writers next to the stage-3 readers in
  `webservice/{postman,xenonyaml}` (external formats), service in application. ✅
- [x] **V. BDD-first** — features per story; round-trip corpus; Newman. ✅
- [x] **VI. Postman parity** — inverse mapping + passthrough; normalised equality (FR-014). ✅
- [x] **VII. Cross-platform** — portable zip names; CI on 3 OSes. ✅

**Post-design re-check**: all gates pass.

## Project Structure

### Documentation
```text
specs/XEN-006-export/
├── plan.md · research.md · data-model.md · quickstart.md
├── contracts/ export-formats.md · bindings.md
├── checklists/
└── tasks.md
```

### Source code (additions)
```text
internal/domain/entity/  export_request.go, export_summary.go
internal/domain/enums/   export_format, export_shape, export_item_kind (+ value pkgs)
internal/application/
├── usecase/  pure: effective_export_root (inherited auth/scripts/vars for folder/request
│             exports), apply_current_values_as_initial, find_literal_credentials; ports:
│             write_postman_export, write_xenon_export, encrypt_export_secrets
├── adapter/  export_service
└── service/  export.go
internal/integration/
├── webservice/postman/   writer.go + dto `New{X}` constructors (inverse mapping, passthrough merge)
├── webservice/xenonyaml/ writer.go (zip/bundle/file, manifest), reader.go (import kinds),
│                         secrets.go (age scrypt armored)
├── persistence/          export_file.go (temp + atomic rename, zip streaming)
└── entrypoint/controller/ export_*.go; import controllers accept passphrase
internal/infra/server/binding/ export.go (+ PickTarget, ShowInFolder via Wails/OS adapter)
test/integration/
├── features/export/  postman_collection, postman_folder_request, postman_environment_globals,
│                     postman_workspace_zip, postman_dropped_fields, roundtrip_corpus,
│                     xenon_yaml_levels, xenon_secrets, current_values, atomic_cancel
├── features/import/  xenon_yaml_import (all kinds, passphrase right/wrong/skip)
└── resources/        postman-collection-v2.1.0.schema.json; tools/newman/ (package.json)
frontend/src/features/export/  ExportDialog, ExportSummary; context-menu entries in sidebar and
                               environments; File → Export
```

**Structure Decision**: unchanged layering; writers live beside the corresponding readers so a
format's mapping is in one package.

## Key design decisions
1. Folder/request Postman exports compute the **effective root** (auth, scripts respecting
   inherit toggles, collection variables) so behaviour in Postman matches Xenon.
2. Passthrough merge never touches Xenon-owned arrays (R1).
3. Secrets block is keyed by exported ids; on import remapped to new ids.
4. Workspace zip import = unpack + register (no id regeneration).

## Complexity Tracking

No constitution violations to justify.
