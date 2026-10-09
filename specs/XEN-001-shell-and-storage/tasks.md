---
description: "Task list for XEN-001 — Shell and Storage (Stage 1)"
---

# Tasks: Shell and Storage (Stage 1)

**Input**: Design documents from `specs/XEN-001-shell-and-storage/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: MANDATORY (constitution V, BDD-first). Every story starts with Godog feature files
in `test/integration/features/` that are run and **proven red** before implementation; frontend
work gets Vitest + React Testing Library tests with `platform/` mocked. Every feature file
carries the tags `@all @xen_001` plus its area tag.

**Rules for every task**: backend files follow `.claude/skills/architecture.md` and the
per-concern skill named in the task; frontend files follow
`.claude/skills/frontend-architecture/SKILL.md`. Error codes come from
[contracts/bindings.md](contracts/bindings.md); file shapes from
[contracts/workspace-format.md](contracts/workspace-format.md).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel (different files, no dependency on an unfinished task)
- **[Story]**: US1 send · US2 organise · US3 tabs · US4 workspaces · US5 history · US6 window

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Wails project, tooling, CI — nothing feature-specific.

- [ ] T001 Scaffold the app at repo root with `wails3 init -n xenon -t react -mod github.com/brienze1/xenon -skipgomodtidy` (wails3 v3.0.0-beta.28), keep `build/`, `frontend/`, `Taskfile.yml`, `build/config.yml`; delete the template's `greetservice.go`; move `main.go` to `cmd/xenon/main.go` and point `build/Taskfile.yml`/`build/config.yml` at `./cmd/xenon`
- [ ] T002 Pin Go deps in `go.mod` (go ≥ 1.25): `github.com/wailsapp/wails/v3 v3.0.0-beta.28`, `go.yaml.in/yaml/v3 v3.0.5`, `github.com/fsnotify/fsnotify v1.10.1`, `github.com/Bios-Marcel/wastebasket/v2 v2.0.3`, `github.com/oklog/ulid/v2`, `github.com/andybalholm/brotli`, `github.com/go-playground/validator/v10`, `github.com/cucumber/godog v0.16.0`
- [ ] T003 [P] Frontend deps in `frontend/package.json` with exact versions: React 19 + react-dom 19, `@wailsio/runtime 3.0.0-beta.28`, zustand 5, CodeMirror 6 packages (`@codemirror/view`, `state`, `commands`, `language`, `search`, `lang-json`, `lang-xml`, `lang-html`, `lang-javascript`), `@dnd-kit/core`, `@dnd-kit/sortable`, `@tanstack/react-virtual`, `react-markdown`; dev: vitest, `@testing-library/react`, `@testing-library/user-event`, jsdom, eslint 9, `typescript-eslint`, `eslint-plugin-boundaries`, `eslint-plugin-react-hooks`; scripts `lint`, `typecheck`, `test`, `check`
- [ ] T004 [P] Configure `frontend/tsconfig.json` (strict, `@/*` → `src/*`) and `frontend/vite.config.ts` (alias `@`, Vitest jsdom env, setup file `src/test/setup.ts`)
- [ ] T005 [P] Write `frontend/eslint.config.js` enforcing the frontend-architecture import table with `eslint-plugin-boundaries` (elements: app, features, platform, shared, bindings; only platform may import `bindings/**` and `@wailsio/runtime`; feature→other-feature only `store.ts`; no barrel `index.ts`)
- [ ] T006 [P] Create the `frontend/src/` skeleton exactly as in the frontend-architecture skill (`app/`, `features/*`, `platform/`, `shared/`), removing the template's `App.tsx` contents
- [ ] T007 [P] Self-host fonts in `frontend/public/fonts/` (Inter, JetBrains Mono, Permanent Marker — with their license files) and declare them in `frontend/src/shared/styles/fonts.css`
- [ ] T008 [P] Write `frontend/src/shared/styles/tokens.css` with every PRD palette token, method colours (`--method-get` … `--method-options`), glow shadows per neon colour, spacing/radius scales, font stacks; and `global.css` (reset, dark background, base typography)
- [ ] T009 [P] Generate platform icons: copy `docs/brand/xenon-icon-1024.png` to `build/appicon.png` and run `wails3 task common:generate:icons` (or the template's icon task) so `build/darwin/icons.icns` and `build/windows/icon.ico` come from it
- [ ] T010 [P] Write the `Makefile` with every target in plan.md "Makefile targets" (setup, dev, bindings, build, package, sort-imports, fix-struct-alignment, vet, check-dead-code, dependency-check via govulncheck, tests, bdd with `TAG`, validate-coverage ≥ 65, mutation-test, validate-mutation-coverage ≥ 72, validate-test-efficacy 100, frontend-check, validate) per `.claude/skills/makefile.md`, excluding `frontend/` and `build/` from Go tools
- [ ] T011 [P] Extend `.gitignore` with `frontend/bindings/`, `build/bin/`, `coverage.out`, `mutation_results.log`
- [ ] T012 [P] Write `.github/workflows/ci.yml`: matrix `macos-latest`, `windows-latest`, `ubuntu-24.04`; setup-go (go.mod version), setup-node (LTS), `make setup`, Linux apt `build-essential pkg-config libgtk-4-dev libwebkitgtk-6.0-dev`, `make validate`, `make build`; on every OS run `make bindings` before `npm run typecheck` so the UI is type-checked against freshly generated bindings (bindings are git-ignored)
- [ ] T013 [P] Add pkg packages per pkg skills: `pkg/errs/` (BaseError + factories from pkg-errs.md, plus `FileError(err, code)` mapping `fs.ErrNotExist`→404, `fs.ErrPermission`→403, other→500), `pkg/logger/` (JSON logger, file sink path set by infra), `pkg/timeutils/` (Now + SetCurrentTimeProvider), `pkg/utils/` (Ptr), `pkg/collection/`
- [ ] T014 [P] Add `pkg/fileutils/write_atomic.go` (`WriteFileAtomic`: temp `.<name>.xenon-tmp` in same dir, write, Sync, Close, Rename, Windows retry 5×20 ms, cleanup) with unit tests in `pkg/fileutils/write_atomic_test.go`

**Checkpoint**: `make dev` opens an empty window; `make validate` runs (with no features yet).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: harness, cross-cutting backend plumbing, minimal workspace + app shell every story
needs. ⚠️ No user-story work starts before this phase is done.

### BDD harness

- [ ] T015 Create `test/integration/steps/step_definitions_test.go` per bdd-testing/testing-setup skills: `TestFeatures` (godog, `Paths: ../features`, `-scenarios` tag flag, Strict), `testUtils`, `InitializeScenario` creating singleton mocks and injecting them once via setters, `before()` that resets mocks and creates a fresh temp `XENON_DATA_DIR` + temp workspace root per scenario, `startApp()` calling `application.NewBindings()` once (no window)
- [ ] T016 [P] Create mocks in `test/integration/mock/`: `api.go` (httptest target per mock-api.md, plus `SetDelay(path, d)` for cancel/timeout and `SetRedirect`), `time.go` (mock-time.md), `trash.go` (records trashed paths), `event_emitter.go` (records emitted events), `dialog.go` (queued picker answers), `file_watcher.go` (lets steps push synthetic file events) — all implementing the real interfaces, with `Reset()` and RWMutex
- [ ] T017 [P] Create generic steps in `test/integration/steps/given_workspace_steps_test.go`: `the workspace "<name>" exists`, `the workspace file "<path>" contains` (docstring), `the file "<path>" exists with content` (docstring), `the current time is "<t>"`, `the "<VAR>" variable is set to the mocked api url`
- [ ] T018 [P] Create generic steps in `test/integration/steps/when_binding_steps_test.go`: `I call "<Service>.<Method>" with` (JSON docstring, `{{var}}` substitution from scenario variables), `I call "<Service>.<Method>" and cancel it after <N>ms`, `I save the response field "<path>" as "<var>"`
- [ ] T019 [P] Create generic steps in `test/integration/steps/then_response_steps_test.go`: `the call should succeed`, `the call should fail with code "<code>"`, `the response should contain the field "<path>" equal to "<value>"`, `… array length equal to <N>`, `… field "<path>" should exist`
- [ ] T020 [P] Create generic steps in `test/integration/steps/then_file_steps_test.go`: `the workspace file "<path>" should exist`/`should not exist`, `the workspace file "<path>" should contain the field "<yamlpath>" equal to "<v>"`, `the workspace file "<path>" should equal` (docstring), `the app data file "<path>" should contain the field …`, `the trash should contain <N> items`
- [ ] T021 [P] Create generic steps in `test/integration/steps/then_event_steps_test.go` and `then_api_steps_test.go`: `the event "<name>" should have been emitted <N> times`, `… with the field "<p>" equal to "<v>"`; `the <N> "<METHOD>" request to "<path>" should have the header "<k>" equal to "<v>"`, `… body equal to` (docstring), `… query "<k>" equal to "<v>"`

### Backend plumbing

- [ ] T022 Write `internal/infra/properties/properties.go` per properties.md: `XENON_DATA_DIR` (default `os.UserConfigDir()/Xenon`, Linux `xenon`), `LOG_LEVEL`, `XENON_ENV`, `AppVersion`
- [ ] T023 [P] Write the error handler controller `internal/integration/entrypoint/controller/error_handler.go` and `internal/integration/adapter/error_handler.go` per error-handler.md (fallback `UNK-00500`)
- [ ] T024 [P] Write `internal/integration/entrypoint/validator/custom.go` per controllers.md (`CustomValidator().Struct(code, &dto)`)
- [ ] T025 [P] Write `internal/infra/server/marshal_error.go` (`Options.MarshalError`: `errs.BaseError` → `ToDto(txId)` JSON; other errors → `UNK-00500` DTO) and `internal/infra/server/event/events.go` (`application.RegisterEvent` for the three events in contracts/bindings.md §8)
- [ ] T026 [P] Write external-boundary adapters with Get/Set factories in `internal/infra/adapter/` + `internal/infra/dependency/external/`: `event_emitter.go` (Wails `app.Event.Emit`), `dialog.go` (Wails `app.Dialog`), `file_watcher.go` (interface only here; implementation in T094)
- [ ] T027 Write `internal/infra/application.go`: `Start()` (builds Wails app with SingleInstance `dev.brienze.xenon`, MarshalError, services from `NewBindings()`, window options from research R10, logger file sink `app-data/logs/xenon.log`) and `NewBindings()` (returns binding structs only, used by BDD); `cmd/xenon/main.go` calls `application.Start()`

### Shared domain + storage core

- [ ] T028 [P] Create domain enums with value packages in `internal/domain/enums/`: `item_kind`, `auth_type`, `body_mode`, `raw_language`, `http_method`, `send_error_kind`, `sidebar_view`, `split_orientation` (enums.md)
- [ ] T029 [P] Create value-object entities in `internal/domain/entity/`: `key_value.go`, `path_variable.go`, `auth.go`, `body.go`, `file_ref.go`, `request_settings.go` with behaviour from data-model.md (defaults, `FileRef.Resolve(workspaceRoot)`, method token validation)
- [ ] T030 [P] Create `internal/domain/entity/request.go`, `collection.go` (Kind collection|folder), `workspace.go`, `registered_workspace.go` (entities.md; ULID ids; name rules)
- [ ] T031 Write pure use case `internal/application/usecase/derive_file_name.go` (format §8 rules incl. Windows reserved names, 64-char cut, case-insensitive uniqueness, reserved `collection`/`folder` names) with table tests
- [ ] T032 Write `internal/integration/persistence/yamlnode/` helpers: load file → `yaml.Node`, patch owned keys in contract order preserving unknown keys, encode indent 2 + LF; plus unit tests for unknown-field and comment preservation
- [ ] T033 Write models in `internal/integration/persistence/model/`: `workspace.go`, `collection.go`, `request.go`, `key_value.go`, `auth.go`, `body.go`, `settings.go` with yaml tags, `ToEntity()`, `New{X}(entity)` and ordered node builders (models.md + research R5)
- [ ] T034 Write `internal/integration/persistence/registry.go` + model `registry.go` (app.json read/write via `WriteFileAtomic`) and use cases `find_workspace_registry.go` / `save_workspace_registry.go` + adapter `registry_persistence.go`

### Minimal workspace + shell (needed by every story)

- [ ] T035 Write BDD `test/integration/features/workspaces/create_workspace.feature` (@workspaces): create in empty folder writes `workspace.yaml` (fields per format §3), `collections/`, `environments/`, `.gitignore` block; registry gains the entry and it becomes active; invalid name → `WKS-00400`; non-empty folder → `WKS-01409`. Run → **red**
- [ ] T036 Implement create + activate workspace: use cases `load_workspace.go`, `save_workspace_file.go`; persistence `workspace.go`; services `create_workspace.go`, `activate_workspace.go`, `list_workspaces.go` (+ adapters); controllers `workspace_create.go`, `workspace_activate.go`, `workspace_list.go` with DTOs; binding `internal/infra/server/binding/workspace.go` (`List`, `Create`, `Activate`); factories in `internal/infra/dependency/`. Run T035 → **green**
- [ ] T037 [P] Write `frontend/src/platform/errors.ts` (parse `RuntimeError.cause` into `AppError {code, description, details}`, `CancelError` → `cancelled`), `frontend/src/platform/events.ts` (typed `on*` wrappers), `frontend/src/platform/services/workspaces.ts` (List/Create/Activate) + `mappers/workspace.ts`, with Vitest tests
- [ ] T038 [P] Build shared UI primitives in `frontend/src/shared/ui/` (one folder each, CSS Modules, compound APIs per vercel-composition-patterns): `button`, `icon-button`, `tabs`, `dropdown`, `dialog`, `context-menu`, `split-pane` (drag + orientation), `key-value-table` (rows, enable, description, bulk-edit toggle via callbacks), `code-editor` (CodeMirror 6 wrapper, lazy-loaded languages), `method-badge`, `empty-state` (marker font + noise texture), `tooltip`, `banner`; with RTL tests for key-value-table, split-pane, tabs
- [ ] T039 Build the window frame `frontend/src/app/layout/AppShell.tsx` + `regions/` (TopBar with workspace switcher slot, Sidebar, MainArea, RightPanel, ConsoleDrawer, Footer) and `frontend/src/app/bootstrap.ts` (load workspace list, activate last, subscribe events); welcome screen in `frontend/src/features/workspaces/components/WelcomeScreen.tsx` (create workspace via folder picker)
- [ ] T040 Write `frontend/src/shared/hooks/use-hotkey.ts` (Cmd↔Ctrl normalisation, scopes) and `frontend/src/app/hotkeys.ts` table (stage 1 shortcuts, FR-042) with tests

**Checkpoint**: an empty workspace can be created and opened; the shell renders; harness green.

---

## Phase 3: User Story 1 — Send a request and read the response (P1) 🎯 MVP

**Goal**: build any stage-1 request in a new tab and inspect the full response.
**Independent test**: new tab → send to the mock API → verify status, timings, body views,
headers, errors, cancel — no saving involved.

### Tests (write first, prove red)

- [ ] T041 [P] [US1] BDD `test/integration/features/requests/params_url_sync.feature` (@requests): `ApplyUrl`/`ApplyParams` round trips, disabled rows kept, path variables derived from `:id` segments in order, encoding rules
- [ ] T042 [P] [US1] BDD `test/integration/features/requests/bulk_edit.feature`: `ParseBulk`/`FormatBulk` (`key:value`, `//` disabled, values with `:`, blank lines)
- [ ] T043 [P] [US1] BDD `test/integration/features/requests/auto_headers.feature`: Content-Type per body mode/raw language, Host, `User-Agent: Xenon/<version>`, Accept, Accept-Encoding, Connection, Authorization from Bearer/Basic; disabled auto header omitted; user header with same name wins
- [ ] T044 [P] [US1] BDD `test/integration/features/sending/send_basics.feature` (@sending): GET with query, custom method `purge` sent as `PURGE`, POST raw JSON/XML/text/HTML/JS, form-data text + file, urlencoded, binary file, GraphQL as JSON POST, Bearer and Basic headers reach the mock; result has status/text/headers (repeated kept)/size/timings fields present
- [ ] T045 [P] [US1] BDD `test/integration/features/sending/send_settings.feature`: timeout → `error.kind timeout` with hint; follow redirects on/off (3xx shown when off); SSL verify on against httptest TLS self-signed → `tls` with hint, off → 200; encode URL on/off
- [ ] T046 [P] [US1] BDD `test/integration/features/sending/send_errors.feature`: unresolvable host → `dns`; closed port → `connection_refused`; invalid URL → `invalid_url`; missing body file → `missing_file`; cancel after 100 ms on a delayed endpoint → `cancelled`
- [ ] T047 [P] [US1] BDD `test/integration/features/sending/response_body.feature`: body ≤ 1 MB text → `inline_body`; 8 MB body → no inline, `body_url` served by the route with exact bytes and content type; binary → `is_binary`; `SaveResponseBody` writes exact bytes; `ReleaseResponse` → route 404
- [ ] T048 [US1] Run T041–T047 with `make bdd TAG=@xen_001` and paste the red summary into the commit message of the test commit

### Backend implementation

- [ ] T049 [P] [US1] Pure use cases in `internal/application/usecase/`: `derive_params_from_url.go`, `build_url_from_params.go`, `parse_bulk_rows.go`, `format_bulk_rows.go`, `compute_auto_headers.go` (usecases.md pure functions)
- [ ] T050 [P] [US1] Entities `internal/domain/entity/response.go`, `timings.go`, `effective_request.go` (+ `SendError` with kind/hint)
- [ ] T051 [US1] Use case `resolve_effective_request.go` (ancestor auth/headers walk per data-model; ancestors passed in; `{{…}}` left literal) in `internal/application/usecase/`
- [ ] T052 [US1] Webservice `internal/integration/webservice/http_request.go` (+ adapter `http_webservice.go`, use case `send_http_request.go`): fresh Transport per send, httptrace timings, redirect/TLS/timeout settings, decompression (gzip/deflate/br), error-kind mapping from research R9
- [ ] T053 [US1] Body store `internal/integration/webservice/response_body_store.go`: per-tab memory ≤ 5 MB else temp file under `os.TempDir()/xenon-<pid>/`, release, cleanup on shutdown
- [ ] T054 [US1] Services `internal/application/service/edit_request.go` (ApplyUrl/ApplyParams/bulk/auto headers) and `send_request.go` (pipeline from plan "Key design decisions" 5, without history — US5 adds it) + adapters
- [ ] T055 [US1] Controllers + DTOs: `request_edit.go`, `request_send.go`, `request_response.go` in `internal/integration/entrypoint/controller/`; DTOs `request.go`, `send_result.go`, `bulk.go` in `internal/integration/entrypoint/dto/` (contracts/bindings.md §3)
- [ ] T056 [US1] Binding `internal/infra/server/binding/request.go` (`ApplyUrl`, `ApplyParams`, `ParseBulk`, `FormatBulk`, `AutoHeaders`, `Send`, `ReleaseResponse`, `SaveResponseBody`) and route `internal/infra/server/route/response_body.go` mounted at `/xenon/response`; register both in `application.go`
- [ ] T057 [US1] Run US1 features → **green**; run full suite

### Frontend implementation

- [ ] T058 [P] [US1] `frontend/src/platform/services/requests.ts` + `platform/response-body.ts` (fetch `/xenon/response/{tab}` as text/blob) + `mappers/request.ts`, with tests
- [ ] T059 [US1] `frontend/src/features/request-builder/store.ts` (draft per tab, edit actions calling ApplyUrl/ApplyParams debounced 150 ms, send/cancel keeping the cancellable promise) with tests
- [ ] T060 [P] [US1] Request bar `features/request-builder/components/RequestBar.tsx` (method dropdown + custom entry with PRD method colours, URL input in JetBrains Mono, Send↔Cancel with glow, Save)
- [ ] T061 [P] [US1] Sub-tabs in Postman order `RequestTabs.tsx` + panels: `ParamsPanel.tsx` (query + path variables tables), `AuthPanel.tsx` (inherit/none/bearer/basic; reserved types listed as "later stage"), `HeadersPanel.tsx` (user rows + greyed auto headers with toggles), `BodyPanel.tsx` (all modes, raw language picker, CodeMirror, file pickers, missing-file marker), `ScriptsPanel.tsx` (stage-4 placeholder), `SettingsPanel.tsx`
- [ ] T062 [US1] `features/response-viewer/store.ts` + components `StatusLine.tsx` (class colours, size, time with timing tooltip), `ResponseBody.tsx` (Pretty/Raw/Preview, folding, search, copy, save; Raw-first above 5 MB; binary handling), `ResponseHeaders.tsx`, `ResponseError.tsx` (kind message + hint), empty "no response yet" state
- [ ] T063 [US1] RTL tests for request-builder and response-viewer components (params↔URL interaction via mocked platform, Send→Cancel, Raw-first rule, error rendering) in their feature folders

**Checkpoint**: US1 independently usable — an unsaved tab sends and shows responses.

---

## Phase 4: User Story 2 — Organise requests as files (P1)

**Goal**: collections/folders/requests on disk, sidebar CRUD, search, settings inheritance.
**Independent test**: build a tree, inspect files, restart, compare.

### Tests (write first, prove red)

- [ ] T064 [P] [US2] BDD `test/integration/features/tree/tree_crud.feature` (@tree): create collection/folder (nested)/request → files per format §1/§4/§5 with explicit defaults; rename renames file + parent order; duplicate → "<name> Copy", new ids for all descendants, placed after original; delete recursive returns `deleted_request_ids`
- [ ] T065 [P] [US2] BDD `test/integration/features/tree/tree_order_and_names.feature`: order persisted/restored; unlisted files appended by name; listed-missing skipped; same-name siblings get `-2`; names with `/:?*` and Windows reserved names; case-insensitive uniqueness; request named "collection"
- [ ] T066 [P] [US2] BDD `test/integration/features/tree/tree_robustness.feature`: malformed YAML → error node, rest loads, file untouched; unknown fields and comments preserved after save; duplicate ids → second re-identified with warning; directories without marker ignored; `xenon: 2` file → `WKS-04422`
- [ ] T067 [P] [US2] BDD `test/integration/features/tree/settings_inheritance.feature`: collection Bearer + request inherit → header sent; folder overrides collection; collection inherit = no auth; header override by name across collection → folder → request (case-insensitive)
- [ ] T068 [P] [US2] BDD `test/integration/features/tree/search.feature`: name/URL substring case-insensitive, requests only, ancestors returned, scoped to the workspace
- [ ] T069 [P] [US2] BDD `test/integration/features/requests/save_request.feature`: Get/Save/SaveAs round trip; `revision` mismatch → `REQ-02409`; `force` overwrites; file refs stored relative inside workspace (`/` separators) and absolute outside
- [ ] T070 [US2] Run T064–T069 → **red**

### Backend implementation

- [ ] T071 [P] [US2] Entities `internal/domain/entity/tree_node.go` + pure use case `search_tree.go`
- [ ] T072 [US2] Persistence `internal/integration/persistence/tree.go` (scan collections dir, build tree with id→path index, order rules, error nodes, duplicate-id handling, own-write registry for the watcher) + use cases `load_tree.go`, `save_tree_item.go` + adapter `tree_persistence.go`
- [ ] T073 [US2] Persistence `internal/integration/persistence/request.go` (node round trip, revision hash, relative/absolute `src` conversion) + use cases `find_request.go`, `save_request.go` + adapter `request_persistence.go`
- [ ] T074 [US2] Services `manage_tree_item.go`, `search_tree.go`, `item_settings.go`, `save_request.go`; extend `send_request.go` to load ancestors via tree persistence for saved requests
- [ ] T075 [US2] Controllers/DTOs `tree_*.go`, `request_get.go`, `request_save.go` and binding `internal/infra/server/binding/tree.go` + `request.go` additions (`Get`, `Save`, `SaveAs`); register
- [ ] T076 [US2] Run US2 features → **green**; full suite green

### Frontend implementation

- [ ] T077 [P] [US2] `platform/services/tree.ts` + `mappers/tree.ts` with tests
- [ ] T078 [US2] `features/sidebar/store.ts` (tree, expanded, selection, search query → `Search` binding debounced 100 ms) with tests
- [ ] T079 [US2] Sidebar components: `CollectionTree.tsx` (virtualized with TanStack Virtual, method badges, error nodes), `SidebarToolbar.tsx` (new collection, search box), item `ContextMenu` (new folder/request, rename inline, duplicate, delete with confirm), `ItemSettingsDialog.tsx` (auth + headers for collection/folder), empty states
- [ ] T080 [US2] `features/side-panel/components/InfoPanel.tsx` (name, markdown description editable, location breadcrumb) and RTL tests for sidebar and info panel

**Checkpoint**: US1 + US2 — requests can be saved, organised and found.

---

## Phase 5: User Story 3 — Tabs like Postman (P2)

**Goal**: preview/pin, dirty state, save prompts, save-into-collection, restore, reorder.
**Independent test**: exercise tab behaviours on a saved tree and restart.

### Tests (write first, prove red)

- [ ] T081 [P] [US3] BDD `test/integration/features/ui_state/ui_state_roundtrip.feature` (@ui_state): `UIStateService.Save/Get` round trip of tabs (order, active, preview id, pinned, drafts incl. unsaved new tab, base revision), defaults when absent, written under `workspaces/<localId>/state.json`, never inside the workspace folder
- [ ] T082 [P] [US3] Vitest `frontend/src/features/tabs/store.test.ts`: preview replace on single click, pin on edit/double-click, focus existing tab, dirty dot, close prompt outcomes (save/don't save/cancel), save-into-collection flow, delete-closes-tab with "keep as new tab" offer, reorder
- [ ] T083 [US3] Run T081–T082 → **red**

### Implementation

- [ ] T084 [P] [US3] Backend: entities `ui_state.go`, `tab.go`, `layout.go`; persistence `ui_state.go` + model; use cases `find_ui_state.go`, `save_ui_state.go`; service `ui_state.go`; controllers/DTOs `ui_state_*.go`; binding `internal/infra/server/binding/ui_state.go`; run T081 → **green**
- [ ] T085 [US3] `platform/services/ui-state.ts` + `features/tabs/store.ts` (tab model, drafts, persisted via UIState debounced 300 ms and on `beforeunload`/shutdown event) — make T082 green
- [ ] T086 [US3] Tab strip `features/tabs/components/TabStrip.tsx` (dnd-kit sortable, italic preview titles, dirty dot, method badge, + button, active underline glow) and `SaveRequestDialog.tsx` (name + collection/folder picker) and `CloseTabPrompt.tsx`
- [ ] T087 [US3] Wire hotkeys Cmd/Ctrl+S, Cmd/Ctrl+Shift+S, Cmd/Ctrl+T, Cmd/Ctrl+W, Cmd/Ctrl+Enter in `app/hotkeys.ts`; RTL test for tab strip interactions

**Checkpoint**: tabs behave like Postman and survive restarts.

---

## Phase 6: User Story 4 — Keep companies apart (P2)

**Goal**: full workspace management and hard isolation.
**Independent test**: two workspaces, switch 100×, restart.

### Tests (write first, prove red)

- [ ] T088 [P] [US4] BDD `test/integration/features/workspaces/manage_workspaces.feature`: Open existing folder; overlap refusal `WKS-02409` (same, parent, child); missing `workspace.yaml` `WKS-03404`; Rename (workspace.yaml only, folder not moved); Remove keep files; Remove + trash (trash mock receives the folder); active falls back to next / none; Locate a moved folder
- [ ] T089 [P] [US4] BDD `test/integration/features/workspaces/isolation.feature`: tree, search, ui state, history (after US5) and file events are scoped per `local_id`; Activate on a deleted folder → `WKS-06404` and `workspace:folder-missing` event
- [ ] T090 [US4] Run T088–T089 → **red**

### Implementation

- [ ] T091 [US4] Trash boundary: `internal/infra/adapter/trash.go` (interface + Get/Set), `trash_darwin.go` (cgo `NSFileManager trashItemAtURL`), `trash_other.go` (wastebasket v2); integration adapter `trash_webservice.go` + use case `move_to_trash.go`
- [ ] T092 [US4] Services `open_workspace.go`, `rename_workspace.go`, `remove_workspace.go`, `locate_workspace.go` (+ overlap validation on registry), controllers/DTOs and binding methods `Open`, `Rename`, `Remove`, `Locate`; run T088–T089 → **green**
- [ ] T093 [US4] Frontend `features/workspaces/` — `store.ts` (switch = reset every workspace-scoped store then load snapshot), `WorkspaceSwitcher.tsx` (top-left, create/open/rename/remove dialogs incl. keep/trash choice), `MissingFolderNotice.tsx` (Locate/Remove); RTL tests including "switch resets stores"

### File changes from outside (FR-017, spans US2–US4)

- [ ] T094 [US4] BDD `test/integration/features/file_changes/external_changes.feature` (@file_changes): synthetic watcher events → `workspace:tree-changed`; request file modified → `request:changed-on-disk` with new revision; deleted → `deleted: true`; own writes suppressed; then **green** after: `internal/infra/adapter/file_watcher.go` (fsnotify recursive, debounce 150 ms, ignore Chmod/`*.xenon-tmp`, overflow → full reload), handler controller `file_change.go`, service `handle_file_change.go`, publisher `internal/integration/publisher/workspace_event.go` + use case `publish_workspace_event.go`; activation starts/stops the watcher
- [ ] T095 [US4] Frontend: `bootstrap.ts` subscriptions → sidebar refetch; tabs store handles `request:changed-on-disk` (clean → reload; dirty → `Banner` with Keep mine (Save with `force`) / Load from disk; deleted → Keep as unsaved tab / Close); RTL test

**Checkpoint**: multiple companies fully separated; outside edits handled.

---

## Phase 7: User Story 5 — History (P3)

**Goal**: every send recorded per workspace and re-openable.
**Independent test**: sends in two workspaces, restart, re-open.

- [ ] T096 [P] [US5] BDD `test/integration/features/history/history.feature` (@history): every Send (ok and error, saved and unsaved) appends an entry with request as written (no resolution) and result summary; List newest-first with query filter and paging; Get returns the exact request; Clear empties; cap at 1,000 with compaction beyond 1,200 lines (use the time mock); stored at `workspaces/<localId>/history.jsonl`. Run → **red**
- [ ] T097 [US5] Backend: entity `history_entry.go`; persistence `history.go` (+ model, JSONL append, compaction via `WriteFileAtomic`); use cases `save_history_entry.go`, `find_history.go`, `clear_history.go`; service `history.go`; extend `send_request.go` to append history; controllers/DTOs `history_*.go`; binding `history.go`. Run T096 → **green**
- [ ] T098 [US5] Frontend: `platform/services/history.ts`; sidebar `HistoryView.tsx` (grouped by day, virtualized, method/URL/status/time, search, Clear with confirm) and Collections/History switch in `features/sidebar/`; opening an entry creates an unsaved tab via tabs store; RTL tests

---

## Phase 8: User Story 6 — Familiar, comfortable window (P3)

**Goal**: Postman placement, hideable/resizable regions remembered per workspace, identity.
**Independent test**: resize/hide/toggle everything, restart, compare.

- [ ] T099 [P] [US6] Vitest `frontend/src/app/layout/AppShell.test.tsx` (regions placed; console hidden by default; hide/show/resize update layout state; side-by-side switch) — **red** first
- [ ] T100 [US6] Implement layout persistence: `features/workspaces/store.ts` exposes layout from UIState; `AppShell` + `SplitPane` read/write it (sizes clamped: sidebar 200–600 px, right panel 240–600 px, console 120 px–60 %, split ratio 0.2–0.8); window bounds persistence in Go (`internal/infra/application.go` window events → app.json, clamp to screen)
- [ ] T101 [P] [US6] Footer `app/layout/regions/Footer.tsx` (console toggle button, layout toggles) and `features/console/components/ConsoleDrawer.tsx` empty state ("Console arrives with scripts")
- [ ] T102 [US6] Visual identity pass: method colours everywhere, glow only on Send/active tab/focused input/running request, empty-state illustrations (`frontend/src/shared/ui/empty-state/illustrations/` SVGs in neon palette), marker font only on wordmark/welcome/empty states; run `vercel-web-design-guidelines` review on `frontend/src/**` and fix findings

---

## Phase 9: Polish & Cross-Cutting Concerns

- [ ] T103 [P] Performance check against SC-002/SC-003: BDD `test/integration/features/tree/large_workspace.feature` (generate 1,000 requests; `GetTree` < 2 s; `Search` < 200 ms) and `sending/large_body.feature` (50 MB via route < 2 s)
- [ ] T104 [P] Crash-safety test `pkg/fileutils/write_atomic_kill_test.go` (subprocess killed mid-write 100×; target is always old or new content — SC-007)
- [ ] T105 [P] Update `README.md` (what Xenon is, `make setup`, `make dev`, platform prerequisites, where data lives)
- [ ] T106 Run `/refactor-diff` until clean, then `make validate` and `make frontend-check` on all three OSes via CI
- [ ] T107 Walk through [quickstart.md](quickstart.md) manually on macOS, Windows and Linux; tick SC-001…SC-008a

---

## Dependencies & Execution Order

### Phase dependencies
- Setup (1) → Foundational (2) → user stories. Polish (9) after the stories it measures.
- US1 (3) and US2 (4) are both P1: US1 first (MVP), US2 next; US2's T074 extends US1's send
  service with ancestor loading.
- US3 (5) needs US2 (saved requests to open in tabs).
- US4 (6) needs Foundational; its file-change part (T094–T095) needs US2 and US3.
- US5 (7) needs US1 (send pipeline); T089's history isolation assertion activates once US5 lands.
- US6 (8) needs US3's UIState persistence.

### Within each story
BDD features first and **red** → domain → application → integration → infra binding → **green**
→ frontend platform → store → components → Vitest/RTL.

### Parallel opportunities
- Setup: T003–T014 in parallel after T001–T002.
- Foundational: T016–T021 (harness pieces), T023–T026, T028–T030, T037–T038 in parallel.
- Each story's BDD feature files are mutually parallel ([P]).
- Frontend tasks of a story can run alongside its backend once the binding contract (DTOs) is
  fixed — contracts/bindings.md defines them up front.

## Parallel Example: User Story 1

```text
Together: T041, T042, T043, T044, T045, T046, T047   (feature files)
Then together: T049, T050                             (pure use cases, entities)
Backend chain: T051 → T052 → T053 → T054 → T055 → T056 → T057
Alongside backend: T058, T060, T061                   (platform + components on the contract)
```

## Implementation Strategy

1. **MVP**: Phases 1–3 (US1) — Xenon sends requests from unsaved tabs. Demo it.
2. **Daily-usable**: + US2 + US3 — saved, organised, tabbed; Luis can start using it for simple
   calls (PRD: real work moves in at stage 3).
3. **Multi-company**: + US4.
4. **Comfort**: + US5, US6, Polish.
Each increment ends with `/refactor-diff` clean and the full suite green.

## Notes
- `[P]` = different files, no unfinished dependencies.
- Prove red before green; quote the red summary in the test commit message.
- Commit per task or logical group on branch `xen-001`; never commit `REFACTORING_REPORT.md`.
