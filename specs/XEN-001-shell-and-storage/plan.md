# Implementation Plan: Shell and Storage (Stage 1)

**Branch**: `xen-001` | **Date**: 2026-10-09 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/XEN-001-shell-and-storage/spec.md`

## Summary

Bootstrap Xenon as a Wails v3 desktop app and deliver stage 1 of the PRD: workspaces as folders
of versioned YAML (one file per request), sidebar tree with CRUD/search, History, Postman-style
tabs with drafts restored on restart, the full request builder (params/path vars, headers with
auto-headers, all body modes, inherit/none/Bearer/Basic auth, settings), sending with httptrace
timings and cancellation, and the response viewer — in the PRD's layout and visual identity, on
macOS, Windows and Linux. Go owns all logic behind thin Wails bindings; the backend follows the
Clean Architecture skills and the frontend follows `frontend-architecture`. The storage format
([contracts/workspace-format.md](contracts/workspace-format.md)) is designed now for every later
stage.

## Technical Context

**Language/Version**: Go 1.25+ (backend), TypeScript 5 + React 19 (frontend)

**Primary Dependencies**: Wails v3.0.0-beta.28 (`@wailsio/runtime` 3.0.0-beta.28),
`go.yaml.in/yaml/v3` v3.0.5, `fsnotify` v1.10.1, `wastebasket/v2` v2.0.3 (+ macOS cgo shim),
`oklog/ulid/v2`, `andybalholm/brotli`, `go-playground/validator/v10`; UI: Vite, Zustand 5,
CodeMirror 6, dnd-kit, TanStack Virtual, react-markdown (details: [research.md](research.md))

**Storage**: YAML files in the user's workspace folders; JSON in app data
(`os.UserConfigDir()/Xenon`); in-memory/temp files for responses

**Testing**: godog v0.16.0 BDD (`test/integration`), `go test`, gremlins mutation testing;
Vitest + React Testing Library for the UI

**Target Platform**: macOS (arm64/amd64), Windows 10/11 (WebView2), Linux (GTK4 + WebKitGTK 6.0,
Ubuntu 24.04+ class distros)

**Project Type**: desktop-app (Go backend + embedded web UI)

**Performance Goals**: SC-001…SC-003 — open a 1,000-request workspace < 2 s, search < 200 ms per
keystroke, 50 MB body shown in Raw < 2 s, UI never blocked by I/O

**Constraints**: offline-capable, no Xenon network services, atomic writes, no secrets in
workspace files/logs, identical behaviour on 3 OSes

**Scale/Scope**: ~4–10 workspaces, up to ~1,000 requests each, 1,000 history entries each, ~20
open tabs

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first** — workspace data is YAML in the chosen folder; local UI state/history in
  app data; no network service of our own; switching workspace swaps every store and the watcher
  (FR-004). ✅
- [x] **II. Secrets** — stage 1 has no secret type yet; history/UI state store requests as typed
  (no resolution), logger never logs request bodies/auth values (debug logs log ids only); the
  format reserves `secret: true` rows with no value (stage 2). ✅
- [x] **III. Go owns logic** — URL/params sync, bulk edit parsing, auto headers, effective
  auth/headers, search, persistence, sending, history are Go use cases; React renders and calls
  `platform/`. Large bodies come from the in-app `/xenon/response` route served by Go (local IPC,
  not network access by React — research R4). Wails imports only in `internal/infra/` and
  `cmd/`. ✅
- [x] **IV. Architecture skills** — package map below follows `architecture.md`; inapplicable
  skill parts (REST router, GORM, AWS, tracer) are not used and their roles are mapped
  explicitly (Bindings ↔ router, YAML persistence ↔ DB persistence). ✅
- [x] **V. BDD-first** — feature files listed per user story below; only outer boundaries faked
  (app-data root and workspace roots are temp dirs, request target is `httptest`, clock, trash,
  event emitter and dialog picker are setter-injected mocks). ✅
- [x] **VI. Postman parity** — tab preview/pin, dirty prompts, params↔URL sync, auto headers,
  body modes, defaults (timeout 0, SSL verify on) match Postman. History is beyond the PRD but was
  added by the user on 2026-10-09 (spec Clarifications; PRD updated). ✅
- [x] **VII. Cross-platform** — file-name rules (format §8), `filepath` everywhere, LF writes,
  per-OS trash, Cmd/Ctrl mapping, CI matrix on 3 OSes (research R13). ✅

**Post-design re-check (after Phase 1)**: all seven gates still pass; no Complexity Tracking
entries.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-001-shell-and-storage/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── workspace-format.md     ← the on-disk YAML contract (all stages)
│   └── bindings.md             ← Go ↔ UI services, DTOs, errors, events
├── checklists/
└── tasks.md                    ← /speckit-tasks
```

### Source Code (repository root)

Backend — names follow the skills' conventions (`{action}_{subject}.go`, adapters embed use
cases, `New{X}Service`, `New{X}Persistence`, `Get{X}`/`Set{X}` factories).

```text
cmd/
└── xenon/main.go                       ← application.Start() only

internal/
├── domain/
│   ├── entity/                         ← workspace.go, registered_workspace.go, collection.go
│   │                                     (Collection+Folder share it via Kind), request.go,
│   │                                     key_value.go, path_variable.go, auth.go, body.go,
│   │                                     file_ref.go, request_settings.go, effective_request.go,
│   │                                     response.go, timings.go, history_entry.go,
│   │                                     ui_state.go, tab.go, layout.go, tree_node.go
│   └── enums/                          ← item_kind, auth_type, body_mode, raw_language,
│                                         http_method, send_error_kind, sidebar_view,
│                                         split_orientation (+ {enum}/value.go each)
├── application/
│   ├── usecase/                        ← find_workspace_registry, save_workspace_registry,
│   │                                     load_workspace, save_workspace_file, load_tree,
│   │                                     save_tree_item, move_to_trash, find_request,
│   │                                     save_request, send_http_request, save_history_entry,
│   │                                     find_history, clear_history, find_ui_state,
│   │                                     save_ui_state, watch_workspace, publish_workspace_event,
│   │                                     and pure ones: derive_params_from_url,
│   │                                     build_url_from_params, parse_bulk_rows,
│   │                                     format_bulk_rows, resolve_effective_request,
│   │                                     compute_auto_headers, derive_file_name,
│   │                                     search_tree
│   ├── adapter/                        ← one interface per service below
│   └── service/                        ← create_workspace, open_workspace, rename_workspace,
│                                         remove_workspace, activate_workspace, locate_workspace,
│                                         list_workspaces, manage_tree_item (create/rename/
│                                         duplicate/delete), search_tree, item_settings,
│                                         edit_request (apply url/params/bulk/auto headers),
│                                         save_request, send_request, history, ui_state,
│                                         handle_file_change
├── integration/
│   ├── adapter/                        ← workspace_persistence, tree_persistence,
│   │                                     request_persistence, registry_persistence,
│   │                                     ui_state_persistence, history_persistence,
│   │                                     http_webservice, trash_webservice (OS boundary),
│   │                                     workspace_publisher, file_change_handler,
│   │                                     handler, error_handler, validator
│   ├── entrypoint/
│   │   ├── controller/                 ← one controller per binding method group:
│   │   │                                 workspace_*.go, tree_*.go, request_*.go,
│   │   │                                 history_*.go, ui_state_*.go, file_change.go,
│   │   │                                 error_handler.go
│   │   ├── dto/                        ← request/response DTOs from contracts/bindings.md
│   │   └── validator/custom.go
│   ├── persistence/                    ← YAML: workspace.go, tree.go, request.go
│   │   │                                 JSON app data: registry.go, ui_state.go, history.go
│   │   ├── model/                      ← yaml/json models with ToEntity()/New{X}() and
│   │   │                                 ordered yaml.Node builders (research R5)
│   │   └── yamlnode/                   ← node patch/merge helpers used only by persistence
│   ├── webservice/                     ← http_request.go (net/http + httptrace, R9),
│   │   ├── dto/                           response_body_store.go (memory/temp-file bodies)
│   │   └── enums/
│   └── publisher/                      ← workspace_event.go (+ dto/)
└── infra/
    ├── application.go                  ← Start(): builds app, registers services, Run();
    │                                     NewBindings() used by BDD without a window
    ├── properties/properties.go        ← XENON_DATA_DIR, LOG_LEVEL, XENON_ENV
    ├── adapter/                        ← event_emitter.go (Wails app.Event), trash.go +
    │                                     trash_darwin.go (cgo NSFileManager) +
    │                                     trash_other.go (wastebasket), dialog.go,
    │                                     file_watcher.go (fsnotify, recursive, debounced)
    ├── dependency/                     ← application/, integration/, controller/, external/
    │                                     (Get*/Set* factories; setters only for EventEmitter,
    │                                     Trash, DialogPicker, FileWatcher, HttpClient factory)
    └── server/
        ├── binding/                    ← workspace.go, tree.go, request.go, history.go,
        │                                 ui_state.go, dialog.go (thin, call controllers)
        ├── route/response_body.go      ← http.Handler for /xenon/response
        ├── event/events.go             ← application.RegisterEvent[...] constants
        └── marshal_error.go            ← BaseError → error DTO JSON

pkg/
├── errs/  logger/  timeutils/  utils/  collection/   ← per pkg-* skills
└── fileutils/                          ← WriteFileAtomic, SafeName helpers (no internal imports)

test/integration/
├── features/
│   ├── workspaces/                     ← create, open, rename, remove (keep/trash), switch
│   │                                     isolation, missing folder, overlap refusal
│   ├── tree/                           ← CRUD, duplicate ids, nesting, order persistence,
│   │                                     unknown-field preservation, malformed file, file names
│   ├── requests/                       ← params↔url sync, path vars, bulk edit, auto headers,
│   │                                     inherited auth/headers, save/save-as/revision conflict,
│   │                                     file refs (relative/absolute/missing)
│   ├── sending/                        ← methods incl. custom, body modes, auth, settings
│   │                                     (timeout, redirects, ssl), errors, cancel, large body
│   │                                     route, timings present
│   ├── history/                        ← recorded per send, per workspace, cap 1,000, clear
│   ├── ui_state/                       ← tabs/drafts/layout round trip per workspace
│   └── file_changes/                   ← external edit/delete → events + revision conflict
├── steps/                              ← split per bdd-step-definitions (given_workspace_…,
│                                         when_binding_…, then_response_…, then_file_…,
│                                         then_event_…, then_api_…)
├── mock/                               ← api.go (httptest target), time.go, trash.go,
│                                         event_emitter.go, dialog.go, file_watcher.go
└── resources/                          ← sample workspaces, files for form-data/binary

frontend/                               ← per .claude/skills/frontend-architecture
├── src/app/                            ← App.tsx, layout/AppShell + regions, hotkeys.ts,
│                                         bootstrap.ts (events → stores)
├── src/features/                       ← workspaces, sidebar (tree + history view + search),
│                                         tabs, request-builder, response-viewer, side-panel
│                                         (Info), console (empty drawer), settings (stub)
├── src/platform/                       ← services/{workspaces,tree,requests,history,
│                                         ui-state,dialogs}.ts, events.ts, errors.ts,
│                                         response-body.ts (fetch /xenon/response), mappers/
└── src/shared/                         ← ui/ (Button, IconButton, Tabs, Dropdown, Dialog,
                                          ContextMenu, SplitPane, KeyValueTable, CodeEditor,
                                          MethodBadge, EmptyState, Tooltip, Banner),
                                          hooks/, lib/, styles/tokens.css

build/                                  ← Wails scaffold; appicon.png from docs/brand/xenon-icon-1024.png
Makefile  Taskfile.yml  .github/workflows/ci.yml
```

**Structure Decision**: single repo, Go module `github.com/brienze1/xenon` at the root (Wails
convention), frontend in `frontend/`. Wails bindings sit in `internal/infra/server/binding/` —
the infra role the REST router has in the skills — and only adapt (ctx, DTO in → controller →
DTO out). Controllers keep the full controller-skill pattern. The YAML/JSON files are the
"database": persistence + model packages own them; the HTTP engine is a webservice; the file
watcher is an inbound entrypoint (controller via `file_change_handler`) and UI notifications go
out through a publisher. **CLAUDE.md is updated** to say bindings live in
`internal/infra/server/binding/` and call controllers.

## Key design decisions

1. **Ids, not paths**: every item has a ULID; the persistence layer keeps an in-memory index
   `id → path` per active workspace, rebuilt on load and patched on every write/watch event.
2. **Save conflicts**: `Request.revision` = SHA-256 of the file bytes at load; `Save` compares it
   with the current file and returns `REQ-02409` unless `force` (FR-017).
3. **Own-write suppression**: persistence registers (path, hash) for each write; the watcher
   handler ignores matching events for 2 s.
4. **Drafts**: tabs carry full request snapshots in `state.json`; the UI saves UI state debounced
   (300 ms) and on window close (Wails `OnShutdown`).
5. **Send pipeline** (service `send_request`): resolve effective request (ancestors) → validate
   file refs → compute auto headers → webservice send with trace → store body (memory ≤ 5 MB else
   temp file) → append history → return `SendResult`.
6. **Variables**: `{{…}}` sent literally in stage 1; resolution is a separate use case slot
   (`resolve_variables`) stage 2 plugs into `resolve_effective_request`.
7. **Logging**: JSON logger (pkg-logger) to `app-data/logs/xenon.log` (rotated at 10 MB × 3).
   The controller skill logs whole DTOs, so every DTO carrying a request/response implements
   `LogValue()` (ids, method, URL only) and the logger honours it — header values, auth and
   bodies never reach the log (constitution II).
8. **Responses survive switches**: the Go body store keeps each tab's last `SendResult` and body;
   the UI rehydrates via `RequestService.LastResult` after a workspace switch (FR-038).

## Makefile targets

Per the makefile skill, adapted to Wails (research R11):

| Target | Does |
|---|---|
| `make setup` | installs pinned tools: wails3, goimports, betteralign, deadcode, govulncheck, gremlins; `npm ci` in `frontend/` |
| `make dev` | `wails3 dev` |
| `make bindings` | `wails3 generate bindings -ts -clean` |
| `make build` | `wails3 build` (current OS) |
| `make package` | `wails3 package` |
| `make sort-imports` | `goimports -w` on Go dirs (excludes `frontend/`) |
| `make fix-struct-alignment` | betteralign loop (models use node builders, so order doesn't matter on disk) |
| `make vet` | `go vet ./...` |
| `make check-dead-code` | `deadcode -test ./...` |
| `make dependency-check` | `govulncheck ./...` |
| `make tests` | `go test ./... -coverpkg=./internal/... -coverprofile=coverage.out` |
| `make bdd [TAG=@x]` | godog suite in `test/integration/steps` with tag filter |
| `make validate-coverage` | ≥ 65 % |
| `make mutation-test` / `validate-mutation-coverage` / `validate-test-efficacy` | gremlins ≥ 72 %, efficacy 100 % |
| `make frontend-check` | `npm run lint && npm run typecheck && npm test` |
| `make validate` | all of the above in the skill's order + `frontend-check` |

## Complexity Tracking

No constitution violations to justify.
