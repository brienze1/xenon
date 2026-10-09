---
description: "Task list for XEN-002 — Environments, Variables and Secrets (Stage 2)"
---

# Tasks: Environments, Variables and Secrets (Stage 2)

**Input**: `specs/XEN-002-environments-and-secrets/` (plan, spec, research, data-model, contracts)
**Prerequisite**: XEN-001 implemented (package map, harness, send pipeline, watcher).

**Tests**: MANDATORY (constitution V). Feature files tagged `@all @xen_002 @<area>`, written
first and proven red. Frontend: Vitest + RTL with `platform/` mocked.

**Rules**: backend per `.claude/skills/architecture.md` + per-concern skills, in the XEN-001
package map; frontend per `frontend-architecture`. Codes from
[contracts/bindings.md](contracts/bindings.md); shapes from
[contracts/format-additions.md](contracts/format-additions.md).

## Format: `[ID] [P?] [Story] Description`
US1 switch & send · US2 manage tables · US3 secrets · US4 inspect & Variables panel ·
US5 colour/confirm/quick look · US6 dynamic variables

---

## Phase 1: Setup

- [ ] T001 Add Go deps to `go.mod`: `filippo.io/age v1.3.2`, `github.com/zalando/go-keyring v0.2.8`, `github.com/brianvoe/gofakeit/v7 v7.17.1`; add `@codemirror/autocomplete` (same major as other CM6 packages) to `frontend/package.json`
- [ ] T002 [P] Add properties `XENON_FORCE_PASSPHRASE` and `XENON_SCRYPT_LOG_N` (default 18) to `internal/infra/properties/properties.go`
- [ ] T003 [P] Vendor Postman's dynamic variable names (118, from `postmanlabs/postman-collection` `lib/superstring/dynamic-variables.js`, commit pinned in the file header) into `test/integration/resources/postman-dynamic-variables.txt`

---

## Phase 2: Foundational (blocks all stories)

### Harness additions
- [ ] T004 [P] Mocks in `test/integration/mock/`: `keychain.go` (modes available / unavailable / error; records entries; Reset), `dynamic_generator.go` (seeded generator wrapper, deterministic per scenario), `clipboard.go` (records last text) — injected once via setters in `step_definitions_test.go`
- [ ] T005 [P] Generic steps in `test/integration/steps/given_vault_steps_test.go` and `then_secret_steps_test.go`: `the keychain is available|unavailable`, `the vault passphrase is "<p>"`, `no file under the workspace folder should contain "<text>"`, `no file under the app data folder should contain "<text>"` (excluding `*.age`), `the clipboard should contain "<text>"`
- [ ] T006 [P] Generic step `the <N> "<METHOD>" request to "<path>" should have the header "<k>" matching "<regex>"` and `… body field "<path>" matching "<regex>"` in `then_api_steps_test.go`

### Domain + storage core
- [ ] T007 [P] Enums in `internal/domain/enums/`: `env_color`, `variable_scope`, `reference_status`, `vault_mode` (+ value packages)
- [ ] T008 [P] Entities in `internal/domain/entity/`: `environment.go` (prod-name defaults), `variable_row.go` (key validation, exact match), `scope_ref.go`, `variable_context.go`, `reference.go`, `resolved_reference.go`, `send_preflight.go`, `vault_status.go`
- [ ] T009 Models + persistence for environments `internal/integration/persistence/environment.go` + `model/environment.go` + `model/variable_row.go` (format additions §1–§2, secret rows without `value`, stray `value` on secret rows dropped on save with warning), use cases `find_environment`, `list_environments`, `save_environment`, `delete_environment`, adapter `environment_persistence.go`; extend `workspace.go`/`tree.go` persistence to read/write `variables` (globals, collection only)
- [ ] T010 Local variables store `internal/integration/persistence/local_variables.go` + model (data-model `variables.json`, owner-only perms) with use cases `find_local_variables`, `save_local_variables`, adapter `local_variables_persistence.go`
- [ ] T011 Pure use cases `tokenize_references.go` (Postman pattern `{{([^{}]*?)}}`), `build_variable_context.go` (precedence, disabled fall-through, current-over-initial), `resolve_text.go` (≤ 19 passes, dynamic last via injected func, cycles literal) with table tests covering every rule in data-model "Resolution rules"

**Checkpoint**: environments/variables can be stored and a text can be resolved in unit tests.

---

## Phase 3: User Story 1 — Switch environments and send (P1) 🎯 MVP

**Independent test**: two environments with different `base_url`; same request goes to two targets.

- [ ] T012 [P] [US1] BDD `test/integration/features/variables/resolution_precedence.feature` (@variables): env > collection > globals; disabled falls through; "No environment" uses collection/globals; unresolved sent literally
- [ ] T013 [P] [US1] BDD `test/integration/features/variables/resolution_fields.feature`: variables resolved in URL, query params (keys+values), path variables, headers (keys+values), raw body, form-data (key/value/text), urlencoded, GraphQL query+variables, binary file path, Bearer/Basic fields, inherited collection/folder auth and headers; History entry keeps unresolved text
- [ ] T014 [P] [US1] BDD `test/integration/features/variables/nesting_and_cycles.feature` and `exact_names.feature`: `{{a}}→{{b}}→v`; 19-pass limit; `a→b→a` literal; `{{ name }}` ≠ `{{name}}`; case-sensitive
- [ ] T015 [P] [US1] BDD `test/integration/features/environments/active_environment.feature`: `SetActive` per workspace persisted in `variables.json`, restored after restart, isolated between workspaces
- [ ] T016 [US1] Run T012–T015 → **red**
- [ ] T017 [US1] Use cases `resolve_request_variables.go` (walks every FR-012 field on a copy) and service changes in `internal/application/service/send_request.go`: resolve effective request → build context → resolve → auto headers → send → history (unresolved)
- [ ] T018 [US1] Environment service subset `manage_environment.go` (List, SetActive) + controllers `environment_list.go`, `environment_set_active.go`, DTOs, binding `internal/infra/server/binding/environment.go`; register; run US1 features → **green**
- [ ] T019 [P] [US1] Frontend `platform/services/environments.ts`; `features/environments/store.ts` (list, active per workspace, reset on workspace switch); `components/EnvironmentSwitcher.tsx` (top-right, "No environment", Cmd/Ctrl+Shift+E opens with focus — add to `app/hotkeys.ts`); RTL tests

---

## Phase 4: User Story 2 — Manage environments, collection variables, globals (P1)

**Independent test**: create env, add rows, set current values, restart, inspect files vs local.

- [ ] T020 [P] [US2] BDD `test/integration/features/environments/manage_environments.feature` (@environments): Create (file per format §2, unique name `ENV-02409`), Update/rename (file renamed), Duplicate (rows, initial, current, secrets copied as "<name> Copy"), Delete (file + current values + vault entries; active → none), malformed file listed with error
- [ ] T021 [P] [US2] BDD `test/integration/features/environments/initial_current_values.feature`: current overrides initial without touching the file; Reset / Persist / Reset all / Persist all; rename row moves current value; outside edit of initial keeps local override (watcher event `variables:changed`); active environment's file deleted outside Xenon → active becomes none and `variables:changed` carries the reason
- [ ] T022 [P] [US2] BDD `test/integration/features/environments/globals_and_collection.feature`: globals in `workspace.yaml`, collection variables in `collection.yaml`, folders reject variables (`folder.yaml` key ignored and preserved), invalid key `VAR-00400`, duplicate keys last-wins + flagged
- [ ] T023 [US2] Run T020–T022 → **red**
- [ ] T024 [US2] Services `manage_environment.go` (Create, Get, Update, Duplicate, Delete) and `variable_table.go` (GetTable, SaveTable with rollback ordering from plan decision 4, SetCurrent, Reset, Persist, AddVariable); controllers/DTOs `environment_*.go`, `variable_*.go`; binding `internal/infra/server/binding/variable.go`; event `variables:changed` emitted via publisher; watcher handler covers `environments/` and variable changes in YAML → **green**
- [ ] T025 [US2] Frontend `platform/services/variables.ts`; `features/environments/components/VariableTable.tsx` (initial | current | type | enabled | description, inherited-current styling, Reset/Persist per row and all, duplicate/space warnings, long values truncated with ellipsis and full value on hover/expand), `EnvironmentsManager.tsx` (create/rename/duplicate/delete), Globals entry in switcher menu, Variables tab in `features/sidebar/components/ItemSettingsDialog.tsx` for collections only; RTL tests

---

## Phase 5: User Story 3 — Secrets out of files (P1)

**Independent test**: secrets never found in the workspace folder; send works after restart.

- [ ] T026 [P] [US3] BDD `test/integration/features/secrets/secrets_never_in_files.feature` (@secrets): create secret rows in env/collection/globals; no workspace/app-data file contains the value; DTOs never contain it (inspect, table, list); logs don't contain it (SC-002)
- [ ] T027 [P] [US3] BDD `test/integration/features/secrets/keychain_mode.feature`: first secret probes keychain → keychain mode, identity stored under `dev.brienze.xenon.vault/<localId>`; restart → secret resolves with no input; Reveal returns value; CopySecret puts value on clipboard mock; workspaces have separate vaults
- [ ] T028 [P] [US3] BDD `test/integration/features/secrets/passphrase_mode.feature`: keychain unavailable → `Status.needs_passphrase_setup`; SetupPassphrase (mismatch `VLT-00400`, short `VLT-01400`); restart → locked; Unlock wrong `VLT-02401`, right → unlocked all vaults
- [ ] T029 [P] [US3] BDD `test/integration/features/secrets/missing_and_locked.feature` and `features/sending/preflight_missing_secrets.feature`: secret without value → `secret_state: missing`; Send → `SND-06409` listing names, nothing sent, no history; re-send with `acknowledge_missing_secrets` → literal placeholder sent; locked vault behaves the same
- [ ] T030 [P] [US3] BDD `test/integration/features/secrets/reset_vault.feature` and `plain_secret_switch.feature`: deleted/corrupt vault or lost keychain entry → broken status → Reset (confirm) → empty vault, secrets missing; plain→secret moves value and warns `git_history`; secret→plain needs `confirm_reveal` (`VAR-01409`); workspace removal deletes vault files and keychain entry; a row turned secret in the YAML (outside edit) moves its plain current value out of `variables.json`; orphan vault entries pruned; empty current value clears the override
- [ ] T031 [US3] Run T026–T030 → **red**
- [ ] T032 [US3] Keychain boundary: `internal/infra/adapter/keychain.go` (go-keyring, Get/Set), integration `webservice/keychain.go` + adapter `keychain_webservice.go`, use case `probe_keychain.go`
- [ ] T033 [US3] Vault persistence `internal/integration/persistence/vault.go`: age X25519 identity per workspace, binary vault (data-model JSON inside), atomic writes; identity storage via keychain or `vault-key.age` (scrypt, logN from properties); `app.json` vault mode + passphraseCheck; use cases `find/save_vault_secrets`, `find/save_vault_identity`, `delete_vault`
- [ ] T034 [US3] Service `vault.go` (state machine from data-model, session passphrase in memory, Status/Setup/Unlock/Reveal/CopySecret/Reset), clipboard boundary (`internal/infra/adapter/clipboard.go`, `webservice/clipboard.go`), `check_send_preflight.go` + Send changes (acknowledgements, `SND-06409`; secret values replaced by `•••••` in error messages/hints, History result summaries and logs); extend `variable_table.go` for secret rows and type switches; `remove_workspace.go` cleanup; controllers/DTOs `vault_*.go` (secret masking enforced in DTO constructors + `LogValue()`); binding `vault.go`; event `vault:status-changed` → **green**
- [ ] T035 [US3] Frontend `platform/services/vault.ts`; secret cells in `VariableTable` (dots, reveal via `Reveal`, copy via `CopySecret`, missing/locked badges); `PassphraseDialog.tsx` (setup ×2 / unlock), `MissingSecretsDialog.tsx` (Send anyway / Cancel → re-send with ack), `VaultBrokenNotice.tsx` (Reset vault), type-switch confirmations; RTL tests

---

## Phase 6: User Story 4 — Inspect, highlight, Variables panel (P2)

- [ ] T036 [P] [US4] BDD `test/integration/features/variables/inspect_and_suggest.feature`: `Inspect` returns references with field/start/end/status/scope for URL, tables and body, including inherited auth/headers; secrets masked; dynamic marked `dynamic`; `Suggest` prefix filtering across scopes + dynamic names, max 50; `AddVariable` to env/collection/globals, including `secret: true` (value goes to the vault — used by "Move to secret variable")
- [ ] T037 [US4] Run T036 → **red**; implement `inspect_request_variables.go` use case + `inspect_variables.go` service (no generation, masking), `Suggest`, controllers/DTOs, binding methods → **green**
- [ ] T038 [US4] `frontend/src/shared/ui/code-editor/`: `singleLine` variant plus generic `decorations` (`{from,to,tone,tooltip}`) and `completionSource` props; switch the URL bar and key/value table cells to it; RTL tests for decorations/completion (domain-free)
- [ ] T039 [US4] `features/request-builder/lib/variable-decorations.ts` (Inspect → decorations, debounced 120 ms; tones resolved/unresolved/secret/dynamic) + completion wiring via `Suggest`; RTL tests
- [ ] T040 [US4] `features/side-panel/components/VariablesPanel.tsx` (list with value/scope/status, inline edit → `SetCurrent`, secrets masked with reveal/copy, unresolved → "Add to environment / collection / globals"); "Move to secret variable" action for a literal value selected in a request field (asks name + scope, calls `AddVariable` with `secret: true`, replaces the selection with `{{name}}`); refresh on `variables:changed`; RTL tests

---

## Phase 7: User Story 5 — Colour, confirm-before-send, quick look (P2)

- [ ] T041 [P] [US5] BDD `test/integration/features/sending/confirm_before_send.feature`: env with flag → Send returns `SND-05409` with env details, nothing sent, no history; re-send with `confirmed_environment_id` → sent; mismatched id → still refused; `prod*` names default red + flag on, others none/off
- [ ] T042 [US5] Run T041 → **red**; implement confirm part of `check_send_preflight.go` and `default_environment_settings.go` → **green**
- [ ] T043 [US5] Frontend: switcher colour + glow from tokens (`--neon-*`, glow shadows), colour picker and flag in `EnvironmentsManager`, `ConfirmSendDialog.tsx` (names env, colour accent), `QuickLook.tsx` (eye button: active env + globals tables, editable, secrets masked); RTL tests

---

## Phase 8: User Story 6 — Dynamic variables (P3)

- [ ] T044 [P] [US6] BDD `test/integration/features/variables/dynamic_variables.feature`: every name in `resources/postman-dynamic-variables.txt` resolves and matches its format regex (Scenario Outline, 118 rows); two `{{$guid}}` differ; user variable `$guid` wins; `{{$nope}}` literal; `$timestamp`/`$isoTimestamp` follow the time mock; 1,000-iteration format check (SC-005)
- [ ] T045 [US6] Run T044 → **red**; implement `internal/integration/webservice/dynamic_values.go` (gofakeit v7 instance + `webservice/dynamicdata/` lists for the *custom* names in contracts/dynamic-variables.md), adapter `dynamic_value_generator.go`, use case `generate_dynamic_value.go`, injected into `resolve_text` → **green**; unit test asserts the Go name list equals the vendored list

---

## Phase 9: Polish

- [ ] T046 [P] Performance: BDD `features/variables/inspect_performance.feature` — 200 references Inspect < 100 ms, send-time resolution < 10 ms (SC-004)
- [ ] T047 [P] Security sweep test in `test/integration/features/secrets/secret_sweep.feature`: after a full scenario (create, send, history, duplicate, switch type, restart, a send that fails with a DNS error on `https://{{secret_host}}/` and a timeout with a secret in the query), grep workspace folder, app data (except `*.age`), log file, History result summaries and error messages for the secret value → 0 matches
- [ ] T048 Update `README.md` "Where data lives" with variables.json, vault files and keychain entry; `/refactor-diff` until clean; `make validate`; CI green on 3 OSes; walk [quickstart.md](quickstart.md)

---

## Dependencies & Execution Order
- Setup → Foundational → US1 → (US2, US3 can run in parallel after US1's T017) → US4 (needs US2/US3 for tables and masking) → US5 (needs US1 preflight hook and US2 manager) → US6 (needs resolver from Foundational; can run after Foundational in parallel with US1–US5) → Polish.
- Within each story: BDD red → domain/application → integration → binding → green → frontend.

## Parallel opportunities
- T002–T003, T004–T006, T007–T008 in parallel; all BDD files marked [P] within a story.
- US6 (T044–T045) is independent of UI work and can be built alongside US2–US5.

## Implementation Strategy
1. MVP = Foundational + US1 (variables resolve, environment switch).
2. + US2 + US3 = daily-usable with secrets (PRD stage 2 done in substance).
3. + US4, US5, US6, Polish.
