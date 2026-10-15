# Implementation Plan: Request Aliases

**Branch**: `xen-016` | **Date**: 2026-10-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/XEN-016-request-aliases/spec.md`

## Summary

A saved request gains an ordered list of **aliases**, each a full, independent copy of the request's
sendable content, stored in the request's own YAML file under `aliases:` (written only when there is
at least one). The Default keeps today's save path untouched — `aliases` is a `Reserved` field in
`model.Request.Fields()`, exactly like `examples` — and a new alias service (mirroring
`manageExample`) is the only writer of the list. React gets an `AliasService` binding (Create, Save,
Delete), `RequestService.Get` returns aliases, `Send` takes an `alias_id` (scripts come from the stored
alias; History records the alias). The request tab gets an Aliases dropdown (top-right), per-version
drafts in the request-builder store, an extra alias crumb (XEN-012 rename) and a `Request · alias`
title. Postman export expands aliases into sibling requests `Request (alias)`; MCP gets alias reads,
`tree_edit kind:"alias"`, `delete` action `alias` and `send alias_id`. See [research.md](./research.md).

## Technical Context

**Language/Version**: Go 1.26 (backend), TypeScript 5 + React 19 (frontend)

**Primary Dependencies**: Wails v3, `gopkg.in/yaml.v3` via `pkg/yamlnode`, Zustand; no new dependency

**Storage**: the request's YAML file (`aliases:` key); per-workspace UI state (tab `alias_id`,
`version_drafts`); history store (`alias_id`, `alias_name`)

**Testing**: Godog BDD (`test/integration/`, `make bdd`), Go unit tests for domain methods, Vitest for UI

**Target Platform**: desktop — macOS, Windows, Linux

**Project Type**: desktop app (Go core + React webview)

**Performance Goals**: switching versions is in-memory (no disk read); alias create/save/delete is one
atomic file write, same cost as an example save

**Constraints**: a request without aliases is written byte-for-byte as today; secrets rules unchanged

**Scale/Scope**: a handful of aliases per request; no limit enforced

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first** — aliases live in the request's YAML file; tab selection and drafts in local
  UI state; nothing networked; workspace boundary unchanged.
- [x] **II. Secrets** — aliases hold the same fields as the Default and go through the same model,
  masking, MCP leak guard and Postman-export secret rules; no new place a value can land.
- [x] **III. Go owns logic** — name rules, copying the Default, independence, script resolution,
  history labels and Postman expansion are Go; React only switches drafts and calls
  `platform/services/aliases.ts`. No Wails import outside `internal/infra`.
- [x] **IV. Architecture skills** — every new file is placed per layer below (entity → usecase/adapter/
  service → persistence/model/dto/controller → infra binding/DI); frontend files per
  `frontend-architecture` (`platform/services`, `platform/mappers`, `features/request-builder`).
- [x] **V. BDD-first** — feature files listed per requirement below; only the filesystem root, clock and
  HTTP target are faked.
- [x] **VI. Postman parity** — Postman has no aliases: recorded as a **user decision** in the PRD
  (FR-026); Postman round-trip of everything else is unchanged and aliases export as plain requests.
- [x] **VII. Cross-platform** — no paths or shortcuts beyond the existing Cmd/Ctrl+S; YAML only.

Post-design re-check: all seven still pass; no Complexity Tracking entries.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-016-request-aliases/
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
├── request_alias.go            NEW  RequestAlias, NewRequestAlias, name checks, Renamed, Edited, AliasAuthOwner, SplitAuthOwner
├── request.go                  CHG  Aliases field; Alias, Version, HasAliasNamed, WithAlias*, WithoutAlias, SendableOf
├── history_entry.go            CHG  AliasID, AliasName, ForAlias
├── tab.go                      CHG  AliasID, VersionDrafts
└── export_node.go              CHG  WithAliasesAsRequests
internal/application/
├── usecase/save_request_aliases.go      NEW  SaveRequestAliases interface
├── usecase/resolve_alias.go             NEW  ALS-01404
├── usecase/reject_invalid_alias_name.go NEW  ALS-03400 / ALS-04400 / ALS-05409
├── usecase/check_alias_name.go          NEW  CheckAliasName
├── adapter/manage_alias.go              NEW  ManageAliasService
├── service/manage_alias.go              NEW  Create / Save / Delete (pattern of manage_example.go)
├── service/send_request.go              CHG  aliasId → stored.Version(aliasId) for scripts; alias AuthOwner; entry.ForAlias
├── service/oauth_owners.go              CHG  resolve "<requestId>:<aliasId>" owner ids (research R12)
└── service/export_postman.go            CHG  node.WithAliasesAsRequests() before WriteCollection
internal/integration/
├── persistence/model/request_alias.go   NEW  YAML model + fields()
├── persistence/model/request.go         CHG  Aliases; optionalAliases in Fields(); FieldsWithAliases()
├── persistence/model/history_entry.go   CHG  alias_id, alias_name
├── persistence/model/tab.go            CHG  alias_id, version_drafts
├── persistence/request.go               CHG  SaveAliases
├── entrypoint/dto/request_alias.go, create_alias.go, save_alias.go, alias_reference.go (+ constructors)  NEW
├── entrypoint/dto/request.go, send_request.go, history_entry.go, tab.go, mcp_send_target.go  CHG
├── entrypoint/controller/alias_manage.go            NEW
├── entrypoint/controller/validate_alias.go          NEW
├── entrypoint/controller/request_send.go            CHG  alias_id passthrough
├── entrypoint/controller/mcp_tree_edit_alias.go     NEW  create / update / rename
├── entrypoint/controller/mcp_tree_edit.go, mcp_delete.go, mcp_send_target.go, mcp_tree*.go  CHG
└── adapter/alias_manage_controller.go               NEW
internal/infra/
├── server/binding/alias.go               NEW  AliasService
└── (DI wiring + Wails service registration) CHG
frontend/src/
├── platform/services/aliases.ts (+ .test.ts)     NEW
├── platform/mappers/request.ts, history.ts, ui-state.ts  CHG
├── features/request-builder/model.ts, store.ts   CHG  aliasId, aliases, versions; selectVersion, createAlias, deleteAlias, saveAllVersions
├── features/request-builder/components/AliasMenu.tsx (+ .module.css, .test.tsx)  NEW
├── features/request-builder/components/CreateAliasDialog.tsx, DeleteAliasDialog.tsx  NEW
├── features/request-builder/components/RequestBuilder.tsx  CHG  header row with AliasMenu
├── features/sidebar/components/ItemBreadcrumb.tsx          CHG  aliasEditor crumb
├── features/tabs/hooks/use-tab-title.ts, store.ts          CHG  title; persist/restore aliasId + versionDrafts; close prompt saves all versions
└── features/sidebar/components/HistoryView.tsx             CHG  "Request · alias" label
test/integration/features/aliases/        NEW  (see below)
test/integration/features/mcp/aliases.feature  NEW
docs/prd/xenon-mvp-prd.md                 CHG  Decisions (Oct 15, 2026): request aliases
```

**Structure Decision**: the existing Xenon layout (Go 4-layer backend + `frontend/` Wails SPA + Godog
suite). Aliases follow the saved-examples precedent at every layer.

## BDD feature files (written by `/create-tests`, all tagged `@aliases`)

| File | Locks |
| --- | --- |
| `test/integration/features/aliases/alias_storage.feature` | FR-001, FR-002 (file shape; request without aliases byte-identical; last alias deleted removes the key), FR-004, FR-005 (each name rule + code), FR-006 (copy of the **saved** Default), FR-008, FR-009 (`REQ-02409`), `ALS-01404` |
| `test/integration/features/aliases/alias_independence.feature` | FR-003 both directions, every sendable field (method, URL, params, path variables, headers, auto-headers, body, auth, scripts, inherit-scripts, settings), Save of the Default after an alias was created keeps it |
| `test/integration/features/aliases/alias_send.feature` | FR-015 (alias content sent), FR-015a (per-alias OAuth token), FR-015b (alias scripts set their own variables; stored alias scripts run, R4), unknown `alias_id` refused and nothing sent, FR-017 history `alias_id`/`alias_name`, Default send has none |
| `test/integration/features/aliases/alias_boundaries.feature` | FR-014b (Save As from alias content → no aliases), FR-018 (Postman export expansion, names, no examples, secrets never), FR-019 (Xenon export → import), FR-020 (Postman/Bruno/cURL imports have none), FR-021 (commit + pull carries aliases), FR-022 (duplicate, move), FR-022a (documentation Defaults only; apply keeps aliases) |
| `test/integration/features/aliases/alias_tab_state.feature` | FR-014 persistence half: UI state round-trips `alias_id` and `version_drafts`; older UI state without them loads |
| `test/integration/features/mcp/aliases.feature` | FR-023, FR-024, FR-025 incl. errors and AI activity |

UI-only requirements (FR-010–FR-014a, the UI half of FR-015a, FR-016, FR-017 label) are locked by Vitest:
`AliasMenu.test.tsx`, `features/request-builder/store.test.ts`, `features/tabs/store.test.ts`,
`ItemBreadcrumb.test.tsx`, `HistoryView.test.tsx`, a `use-tab-title` test, and
`platform/services/aliases.test.ts`.

## Complexity Tracking

No constitution violations.
