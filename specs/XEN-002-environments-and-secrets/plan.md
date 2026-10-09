# Implementation Plan: Environments, Variables and Secrets (Stage 2)

**Branch**: `xen-002` (stacked on `xen-001`) | **Date**: 2026-10-09 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/XEN-002-environments-and-secrets/spec.md`

## Summary

Add per-workspace environments, collection variables and globals with Postman's initial/current
model, a backend resolver (precedence, 19-pass nesting, 118 dynamic variables) wired into the
XEN-001 send pipeline with a preflight for confirm-before-send and missing secrets, and a
per-workspace age-encrypted vault whose identity lives in the OS keychain (passphrase fallback).
UI: environment switcher + quick look, environment/globals/collection-variable editors, `{{…}}`
highlighting and autocomplete via CodeMirror, and the Variables side panel. Builds strictly on
XEN-001's package map, format and bindings conventions.

## Technical Context

**Language/Version**: Go 1.25+, TypeScript 5 + React 19 (unchanged)

**Primary Dependencies** (new): `filippo.io/age` v1.3.2, `github.com/zalando/go-keyring` v0.2.8,
`github.com/brianvoe/gofakeit/v7` v7.17.1; UI: `@codemirror/autocomplete` (same CM6 line as
XEN-001)

**Storage**: environment YAML in `environments/`; globals/collection variables in existing YAML;
local `variables.json`, `vault.age`, `vault-key.age` per workspace in app data; OS keychain

**Testing**: godog BDD with keychain mock (available / unavailable), seeded dynamic generator,
frozen clock; Vitest + RTL

**Target Platform**: macOS, Windows, Linux (keychain backends per OS, passphrase fallback)

**Project Type**: desktop-app

**Performance Goals**: SC-004 — Inspect/highlight ≤ 100 ms per keystroke for 200 references;
resolution < 10 ms per send; passphrase unlock ≈ 1 s (scrypt logN 18)

**Constraints**: no secret value in workspace files, logs, History, UI state, temp files, or
bridge responses (except Reveal); Postman parity of resolution

**Scale/Scope**: ~10 environments × ~100 variables per workspace

## Constitution Check

- [x] **I. Local-first** — environments are files in the workspace; current values, active env
  and vault are local app data; nothing networked. ✅
- [x] **II. Secrets** — secret rows have no value in files (format additions §1); vault encrypted
  with age, identity in keychain or passphrase-wrapped; DTOs mask secrets (bindings rule);
  `CopySecret` never crosses the bridge; logger redaction from XEN-001 extended to variable DTOs;
  History stores unresolved requests; BDD asserts no secret bytes in workspace/app-data files
  other than the encrypted vault (SC-002). ✅
- [x] **III. Go owns logic** — resolution, precedence, dynamic generation, masking, vault and
  preflight all in Go; UI renders Inspect results. ✅
- [x] **IV. Architecture skills** — package map below follows the skills and XEN-001's mapping. ✅
- [x] **V. BDD-first** — features listed per story; boundaries faked: keychain, dynamic generator
  seed, clock, emitter, clipboard. ✅
- [x] **VI. Postman parity** — R1/R2; confirm-before-send and the passphrase fallback are PRD /
  user decisions (spec Clarifications). ✅
- [x] **VII. Cross-platform** — keychain per OS + fallback tested via mock; file permissions as
  XEN-001. ✅

**Post-design re-check**: all gates pass; no Complexity Tracking entries.

## Project Structure

### Documentation
```text
specs/XEN-002-environments-and-secrets/
├── plan.md · research.md · data-model.md · quickstart.md
├── contracts/
│   ├── format-additions.md     ← environment files, variable rows (format v1 slots)
│   ├── dynamic-variables.md    ← the 118 names and formats
│   └── bindings.md             ← EnvironmentService, VariableService, VaultService, Send changes
├── checklists/
└── tasks.md
```

### Source code (additions to XEN-001's map)
```text
internal/domain/
├── entity/       environment.go, variable_row.go, scope_ref.go, variable_context.go,
│                 reference.go, resolved_reference.go, send_preflight.go, vault_status.go
└── enums/        env_color, variable_scope, reference_status, vault_mode (+ value pkgs)

internal/application/
├── usecase/      find_environment, save_environment, delete_environment, list_environments,
│                 find_local_variables, save_local_variables, find_vault_secrets,
│                 save_vault_secrets, delete_vault, find_vault_identity, save_vault_identity,
│                 probe_keychain, generate_dynamic_value, copy_to_clipboard,
│                 pure: tokenize_references, resolve_text (passes ≤ 19, precedence, dynamic
│                 last), build_variable_context, resolve_request_variables (all FR-012 fields),
│                 inspect_request_variables, check_send_preflight, default_environment_settings
├── adapter/      environment_service, variable_service, vault_service, (send_request extended)
└── service/      manage_environment.go, variable_table.go, inspect_variables.go, vault.go,
                  send_request.go (extended: preflight + resolution)

internal/integration/
├── adapter/      environment_persistence, local_variables_persistence, vault_persistence,
│                 keychain_webservice (OS boundary), dynamic_value_generator, clipboard_webservice
├── entrypoint/controller/   environment_*.go, variable_*.go, vault_*.go; request_send.go extended
├── entrypoint/dto/          environment.go, variable_row.go, inspect.go, vault_status.go, ...
├── persistence/ environment.go (+ model), local_variables.go (+ model), vault.go (age encrypt/
│                decrypt; identity wrap/unwrap), workspace.go/tree.go extended for variables
└── webservice/  keychain.go (go-keyring), dynamic_values.go (gofakeit + lists in
                 webservice/dynamicdata/*.go), clipboard.go (Wails clipboard via infra adapter)

internal/infra/
├── adapter/      keychain.go (go-keyring wrapper, Get/Set for mock), clipboard.go
├── dependency/   factories for the above
└── server/binding/  environment.go, variable.go, vault.go; event/events.go adds
                     variables:changed, vault:status-changed

test/integration/
├── features/environments/   manage_environments, initial_current_values, globals_and_collection
├── features/variables/      resolution_precedence, resolution_fields, nesting_and_cycles,
│                            exact_names, dynamic_variables, inspect_and_suggest
├── features/secrets/        secrets_never_in_files, keychain_mode, passphrase_mode,
│                            missing_and_locked, reset_vault, plain_secret_switch
├── features/sending/        confirm_before_send, preflight_missing_secrets
├── mock/                    keychain.go (available/unavailable/error), dynamic_generator seed,
│                            clipboard.go
└── resources/postman-dynamic-variables.txt

frontend/src/
├── features/environments/   store.ts, components/ EnvironmentSwitcher, QuickLook,
│                            EnvironmentsManager, VariableTable (initial/current/type/enabled/
│                            description, Reset/Persist, secret reveal/copy), ConfirmSendDialog,
│                            MissingSecretsDialog, PassphraseDialog, VaultBrokenNotice
├── features/side-panel/     components/VariablesPanel.tsx (+ Add to scope)
├── features/request-builder/ variable decorations + completion feeding (lib/variable-decorations.ts)
├── platform/services/       environments.ts, variables.ts, vault.ts
└── shared/ui/code-editor/   singleLine variant, generic decorations + completionSource props
```

**Structure Decision**: unchanged from XEN-001; new domain lives in the same layers. The
keychain, dynamic value generator and clipboard are outer boundaries (integration webservice +
infra adapter with setter), so BDD controls them.

## Key design decisions

1. **One resolver, two modes**: `resolve_text` used by Send (generates dynamic values) and
   Inspect (marks them `dynamic`, no generation, masks secrets). Pure function over a
   `VariableContext`; dynamic generation passed in as a function → deterministic tests.
2. **Field coverage** (FR-012): resolver walks a request copy: URL, params (key/value), path
   variable values, headers (key/value), auth sub-fields of the active type, body by mode
   (raw content, form-data key/value/text, urlencoded key/value, GraphQL query + variables,
   binary `src`), plus inherited auth/headers from ancestors. File `src` is resolved before the
   XEN-001 missing-file check.
3. **Preflight in the backend**: `Send` returns `SND-05409` (confirm) / `SND-06409` (missing or
   locked secrets) before any network I/O or History write; UI re-sends with acknowledgements.
   SC-007 holds even if the UI misbehaves.
4. **Atomic multi-store save**: `SaveTable` writes YAML (initial values), `variables.json`
   (current), and `vault.age` (secrets). Order: vault → local → YAML; on a later failure the
   earlier writes are rolled back from in-memory copies, so a failed save leaves all three as
   before.
5. **Vault identity storage** abstracted by `find/save_vault_identity`: keychain mode reads the
   keychain; passphrase mode unwraps `vault-key.age` with the session passphrase held in memory.
   Mode probe once per app (stored in `app.json`). Properties: `XENON_FORCE_PASSPHRASE` (dev/test: skip the
   probe and use passphrase mode) and `XENON_SCRYPT_LOG_N` (BDD only: lower scrypt cost).
6. **Outside edits**: XEN-001's watcher also covers `environments/`; changes emit
   `variables:changed`; local overrides are keyed by scope+key so they survive.
7. **Removal**: `WorkspaceService.Remove` deletes `variables.json`, vault files and the keychain
   entry (errors logged, removal continues).

## Complexity Tracking

No constitution violations to justify.
