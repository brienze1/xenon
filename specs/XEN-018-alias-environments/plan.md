# Implementation Plan: Alias Environments

**Branch**: `xen-018` | **Date**: 2026-10-23 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/XEN-018-alias-environments/spec.md`

## Summary

A request alias (XEN-016) gains an optional, ordered set of **environment ids**. Empty means All
environments and is written to the request file as no key at all, so every existing file loads and
saves byte-for-byte as today; a non-empty set is written as `environments: [<id>, …]` inside the
alias's YAML mapping. One domain rule, `RequestAlias.IsAvailableIn(active *Environment)`, decides
availability. The backend send (`sendRequest.Send`) checks it right after resolving the alias and
before preflight, scripts or HTTP, and refuses with **`ALS-09403`**. Every app and MCP send goes
through that one path, so both inherit the guard.

The link is written by a new, dedicated alias write, `ManageAliasService.SetEnvironments` (binding
`AliasService.SetEnvironments`), which checks every id against the workspace's environments
(**`ALS-10400`**) and never touches the alias's content. `Create` takes an optional initial set. The
content save (`Save`) keeps the stored set. MCP reads `environments` (id + name, or
`Deleted environment`) on each alias and sets them with `tree_edit kind:"alias"` +
`environment_ids`.

Every send's history entry records the active environment's id, name and colour as they were at
send time (`environment` on the entry, omitted when none). History rows and MCP history reads show
it. Postman export adds `Xenon alias for environments: …` to an alias request's description.

In the UI, the Aliases dropdown shows environment dots and names, greys out unavailable aliases and
adds an **Environments…** item to each ⋯ menu (opening a new `AliasEnvironmentsDialog`). The Create
dialog gets the same picker. The request bar disables the Send button (and the store's `send` ignores Mod+Enter) and shows a
banner with **Switch to <env>** buttons while the alias on screen is unavailable. See
[research.md](./research.md).

## Technical Context

**Language/Version**: Go 1.26 (backend), TypeScript 5 + React 19 (frontend)

**Primary Dependencies**: Wails v3, `go.yaml.in/yaml/v3` via `pkg/yamlnode`, Zustand; no new dependency

**Storage**: the request's YAML file (`aliases[].environments`, absent when empty); history store
(`environment` object per send entry, omitted when none)

**Testing**: Godog BDD (`test/integration/`, `make bdd`), Go unit tests for the domain rule, Vitest for UI

**Target Platform**: desktop — macOS, Windows, Linux

**Project Type**: desktop app (Go core + React webview)

**Performance Goals**: the guard adds one in-memory membership check to a send. Its error message
lists environments only on the refusal path (one `ListEnvironments`). Setting environments is one
atomic file write, the same cost as an alias delete.

**Constraints**: an alias with no link is written exactly as today (no `environments` key); history
entries without an environment are written exactly as today; secrets rules unchanged.

**Scale/Scope**: a handful of aliases per request and of environments per workspace; no limit enforced.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first**: the link lives in the request's YAML file and syncs with git like every
  other workspace file; the history environment lives in the local history store.
- [x] **II. Secrets**: no new value is stored: environment ids, names and colours are not secrets.
  The guard *reduces* leakage (a test credential can't reach another environment). The `ALS-09403`
  message holds only the alias's and environments' names.
- [x] **III. Go owns logic**: Go owns availability and enforces it on every send (the UI can't
  bypass it). Go also owns id validation, storage, the history snapshot and the export note. React
  only *displays* availability for greying out and disabling Send, from the environment ids Go
  returns and the active environment it already holds (research R3). A wrong UI state can never send
  a refused alias.
- [x] **IV. Architecture skills**: every file is placed per layer below. Domain rule on the entity,
  rejections as usecases, a new service method plus adapter, DTOs and controller in integration,
  binding and DI in infra. Frontend files per `frontend-architecture`.
- [x] **V. BDD-first**: the feature files are listed per requirement below. Only the filesystem root,
  the clock and the HTTP target are faked.
- [x] **VI. Postman parity**: aliases are a recorded user decision (PRD, Oct 15, 2026). This feature
  extends it at the user's request (issue #25) and is recorded in the PRD Decisions (Oct 23, 2026).
  Postman export stays valid: the link becomes description text.
- [x] **VII. Cross-platform**: no paths or shortcuts beyond the existing Cmd/Ctrl+Enter; YAML only.

Post-design re-check: all seven still pass; no Complexity Tracking entries.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-018-alias-environments/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── bindings.md
│   ├── mcp.md
│   └── ui.md
├── checklists/
└── tasks.md
```

### Source Code (repository root)

```text
internal/domain/entity/
├── request_alias.go            CHG  EnvironmentIDs; NewRequestAlias(id, name, of, environmentIds); IsAvailableIn; WithEnvironments; IsLinked
├── alias_environments.go       NEW  NormalizedEnvironmentIDs (trim, drop blanks, dedupe keeping first order); AliasEnvironmentLabels (names, "Deleted environment")
├── history_entry.go            CHG  Environment *HistoryEnvironment; InEnvironment(*Environment)
├── history_environment.go      NEW  HistoryEnvironment{ID, Name, Color}
└── export_node.go              CHG  AliasRequests appends the environments note to the description (labels passed in)
internal/application/
├── usecase/reject_unavailable_alias.go   NEW  ALS-09403 (ForbiddenError), message names alias + environments
├── usecase/reject_unknown_environment.go NEW  ALS-10400 (BadRequestError)
├── adapter/manage_alias.go               CHG  Create(+environmentIds); SetEnvironments(...)
├── service/manage_alias.go               CHG  Create stores the normalized set; SetEnvironments; checkEnvironments via ListEnvironments
├── service/send_request.go               CHG  guard after s.alias(); entry.InEnvironment(scopes.environment)
└── service/export_postman.go             CHG  pass environment labels (existing p.items.listEnvironments; no DI change) to WithAliasesAsRequests / WithAliasRequests
internal/integration/
├── persistence/model/request_alias.go    CHG  Environments []string `yaml:"environments"`; Node() writes it only when non-empty (yamlnode.StringList)
├── persistence/model/history_entry.go    CHG  Environment *HistoryEnvironment `json:"environment,omitempty"`
├── persistence/model/history_environment.go NEW
├── entrypoint/dto/request_alias.go       CHG  EnvironmentIds []string `json:"environment_ids"`
├── entrypoint/dto/create_alias.go        CHG  EnvironmentIds []string `json:"environment_ids"`
├── entrypoint/dto/set_alias_environments.go NEW  local_id, request_id, alias_id, environment_ids, base_revision
├── entrypoint/dto/history_entry.go       CHG  Environment *HistoryEnvironment
├── entrypoint/dto/history_environment.go NEW  id, name, color
├── entrypoint/dto/mcp_tree_alias.go      CHG  Environments []MCPAliasEnvironment (id, name)
├── entrypoint/dto/mcp_tree_edit.go       CHG  EnvironmentIds *[]string `json:"environment_ids"` (used_by create,update; kind alias)
├── entrypoint/dto/mcp_tree_edit_alias.go CHG  CreateAlias carries ids; SetAliasEnvironments(...)
├── entrypoint/dto/mcp_history_line.go, mcp_history_send.go CHG  environment name
├── entrypoint/controller/alias_manage.go CHG  SetEnvironments
├── entrypoint/controller/validate_alias.go CHG validateSetAliasEnvironments → ALS-11400
├── entrypoint/controller/mcp_tree_edit_alias.go CHG  create with ids; update/rename then SetEnvironments when environment_ids given
├── entrypoint/controller/mcp_tree.go     CHG  alias environment names (ListEnvironments via the existing environment service)
└── adapter/alias_manage_controller.go    CHG  SetEnvironments
internal/infra/
├── server/binding/alias.go               CHG  SetEnvironments
└── dependency/... (manage alias + tree controller wiring) CHG  ListEnvironments / FindEnvironment injection
frontend/src/
├── platform/services/aliases.ts (+ .test.ts)   CHG  createAlias(environmentIds), setAliasEnvironments
├── platform/mappers/request.ts                 CHG  SavedAlias.environmentIds
├── platform/mappers/history.ts                 CHG  HistoryEntry.environment
├── features/request-builder/lib/alias-availability.ts (+ .test.ts) NEW  isAliasAvailable, aliasEnvironmentLabels
├── features/request-builder/store.ts (+ test)  CHG  createAlias(name, environmentIds); setAliasEnvironments; send() is a no-op while the alias on screen is unavailable (covers button, Mod+Enter and dialogs)
├── features/request-builder/components/AliasMenu.tsx (+ .test.tsx, .module.css) CHG  dots, greyed-out rows, Environments… item
├── features/request-builder/components/AliasEnvironmentsDialog.tsx (+ .test.tsx) NEW
├── features/request-builder/components/EnvironmentPicker.tsx (+ .module.css) NEW  shared by both dialogs
├── features/request-builder/components/CreateAliasDialog.tsx (+ NEW .test.tsx) CHG  picker preset All environments
├── features/request-builder/components/UnavailableAliasBanner.tsx (+ .test.tsx) NEW  message + Switch to <env>
├── features/request-builder/components/RequestBar.tsx (+ NEW .test.tsx)  CHG  Send button disabled and UnavailableAliasBanner shown when unavailable
└── features/sidebar/components/HistoryView.tsx (+ test) CHG  environment dot + name on the row
docs/prd/xenon-mvp-prd.md                       CHG  Decisions (Oct 23, 2026): alias environments
```

**Structure Decision**: the existing Xenon layout (Go 4-layer backend + `frontend/` Wails SPA + Godog
suite). Everything extends the XEN-016 alias files. The new write mirrors `Delete` (an alias write
with stale-revision protection that saves immediately).

## BDD feature files (written by `/create-tests`, all tagged `@alias_environments`)

| File | Locks |
| --- | --- |
| `test/integration/features/aliases/alias_environment_storage.feature` | FR-001, FR-002 (file shape with a link; no `environments` key without one; existing file unchanged on load and save; clearing removes the key), FR-008 Go half (Create with ids, Create without = All), FR-011 (`ALS-10400` unknown id; duplicates stored once; `REQ-02409` stale write; `ALS-01404` unknown alias; `ALS-11400` invalid body), FR-012 (deleting an environment leaves request files untouched; link read back with the leftover id), content Save keeps the link, duplicating a request keeps links, duplicating an environment links nothing |
| `test/integration/features/aliases/alias_environment_send.feature` | FR-003, FR-004, FR-005, FR-006 (app half): linked alias with its environment active is sent; with another environment, none, or a deleted active environment, refused with `ALS-09403`, no HTTP call, no history entry, no script run; multi-environment set; All environments alias and Default always sent; a "Confirm before send" environment refuses the alias before asking for confirmation; an alias whose environments are all deleted is refused everywhere |
| `test/integration/features/aliases/alias_environment_boundaries.feature` | FR-015 (Postman export description note with names; All environments alias exports unchanged; deleted environment labelled), FR-014 (Xenon export → import keeps links), FR-016 Go half (history `environment` id/name/colour snapshot for Default and alias sends; none when no environment; renaming the environment later doesn't change the entry; old entries without it load) |
| `test/integration/features/mcp/alias_environments.feature` | FR-013, FR-006 (AI half): `tree get` lists each alias's `environments` (id + name; empty for all; `Deleted environment`); `tree_edit kind:"alias"` create/update with `environment_ids` sets and `[]` clears; unknown id `ALS-10400`; `send alias_id` refused with `ALS-09403`, no approval prompt, no HTTP call; `history` read shows the environment |

UI-only requirements (FR-007, FR-008 UI half, FR-009, FR-010, the FR-016 row) are locked by
Vitest: `alias-availability.test.ts`, `AliasMenu.test.tsx`, `AliasEnvironmentsDialog.test.tsx`,
`CreateAliasDialog.test.tsx`, `UnavailableAliasBanner.test.tsx`, `RequestBar.test.tsx`,
`features/request-builder/store.test.ts`, `HistoryView.test.tsx` and
`platform/services/aliases.test.ts`.

## Complexity Tracking

No constitution violations; nothing to justify.
