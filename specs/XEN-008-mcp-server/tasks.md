---
description: "Task list for XEN-008 — MCP server: full Xenon access for AI clients, secrets write-only"
---

# Tasks: MCP Server — Full Xenon Access for AI Clients, Secrets Write-Only

**Input**: `specs/XEN-008-mcp-server/` (spec, plan, research, data-model, contracts/mcp-tools.md,
contracts/bindings.md, contracts/stdio-bridge.md, quickstart)

**Prerequisites**: XEN-001 … XEN-007 merged on `main`.

"Design decision R#" refers to the numbered decisions in `research.md`.

**Tests**: MANDATORY (constitution V).
- Feature files are tagged `@all @xen_008 @<area>`. `/create-tests` writes them and proves them red; `/implement-task`
  turns them green.
- Frontend stories add Vitest tests.
- Only the outer boundaries are faked: the HTTP target (mock API), the keychain, the clock, the window focuser and the
  event emitter. A real MCP client talks to the real server.

**Stories**:
- US1 — connect, browse, send (P1)
- US2 — secrets write-only (P1)
- US3 — authoring (P2)
- US4 — environments and the rest (P2)
- US5 — approvals (P2)
- US6 — AI activity in History (P2)
- US7 — Settings as a tab (P3)

Every Go task follows its skill in `.claude/skills/` (entities, enums, usecases, services, adapters, controllers,
dtos, persistence, models, publishers, properties, dependency-injection, naming-conventions, comments). Every frontend
task follows `frontend-architecture`. Run `make sort-imports` after each batch of Go changes.

---

## Phase 1: Setup

- [x] T001 Add `github.com/modelcontextprotocol/go-sdk@v1.8.0` to `go.mod`/`go.sum` (`go get`, `go mod tidy`); importable only from `internal/infra/server/mcp`, `internal/infra/mcpbridge` and `test/integration` (design decision R1). — done by `/create-tests` (go.mod/go.sum; the dependency was already indirect at v1.6.1).
- [ ] T002 [P] Add `XENON_MCP_PORT` to `internal/infra/properties/properties.go` (`MCPPortOverride *int`): it replaces the default port 47821 while no port is saved in settings; a saved port always wins. Follow the properties skill, with a unit test in `properties_test.go`.
- [ ] T003 [P] Amend `docs/prd/xenon-mvp-prd.md` (FR-030):
  - Add a section "AI access (MCP)" covering: transport (HTTP + `xenon mcp`); loopback only, no token, on/off switch; secrets write-only, the leak guard, AI-written files written masked, and the two accepted residual risks (transformed values; sending to a destination the AI controls); approvals and "Dangerously skip approvals"; AI activity in History with per-source caps and filter; the 24-tool scheme.
  - Update the History line (sources, configurable caps) and the Settings description (it opens as a tab).
  - Add a decision entry dated 2026-10-13 to the decisions list.
- [x] T004 [P] Add a generic MCP test client in `test/integration/steps/when_mcp_steps_test.go`: — done by `/create-tests`. Deviations: steps are registered with `ctx.Step`, because godog binds Given/When/Then strictly. Clients are kept per name, so several sessions stay open at once. Also added: named background calls (`I start calling the MCP tool "<tool>" in the background as "<name>" with` / `the background call "<name>" should finish within N ms`), `the MCP endpoint should refuse connections`, `I post a body of N bytes to the MCP endpoint`, and in `then_mcp_steps_test.go` `the error should not contain`, `the response should have at most N characters`, `the response field "<f>" should have N characters`, `every item of the response field "<f>" should have the field "<k>" matching "<re>"`. Documenting these steps in `.claude/skills/bdd-step-definitions.md` is left for `/implement-task`.
  - It uses the SDK `StreamableClientTransport` against the address from `AIAccessService.Status`.
  - Steps: `I call the MCP tool "<tool>" with` (docstring), `I call the MCP tool "<tool>"`, `I call the MCP tool "<tool>" in the background with`, `the background MCP call finishes`, `I list the MCP tools`, `the MCP client is named "<name>"`, and `I post to the MCP endpoint with the headers` (table). The last step sets the route response, so the existing `the route response status should be N` applies.
  - Results become "the response": the parsed JSON text block, plus `is_error`. This way every existing `the response should …` step applies.
  - Make the existing step `the last call should have taken less than <n> ms` also measure the last MCP call.
  - Document the steps in `.claude/skills/bdd-step-definitions.md`.
- [x] T005 [P] Add generic steps in `test/integration/steps/when_program_steps_test.go` / `then_program_steps_test.go`: — done by `/create-tests`, in `when_mcp_steps_test.go`. The bridge step builds `./cmd/xenon` once per suite and runs `xenon mcp`. Until `internal/infra/mcpbridge` exists it fails at once instead of launching the GUI. Added `the bridge output should match "<re>"` and `the bridge error output should contain`.
  - `I run the stdio bridge with the input` (docstring): it runs `go run ./cmd/xenon mcp` (built once per suite) with the suite's `XENON_DATA_DIR` and `XENON_MCP_PORT`.
  - `the bridge output should contain "<text>"` and `the bridge exit code should be <n>`.
- [ ] T006 [P] (`/create-tests` added the mock `test/integration/mock/window_focuser.go`, `suite.window` and the step `the Xenon window should have been brought to the front N times`; left for `/implement-task`: `external.SetWindowFocuser` and injecting `suite.window`.) Add a window-focuser mock `test/integration/mock/window_focuser.go` (records focus calls), set through a new `external.SetWindowFocuser`. Add steps `the Xenon window should have been brought to the front` and `… should not have been brought to the front` in `then_app_steps_test.go`.
- [x] T007 [P] Add a generic echo capability to `test/integration/mock/api.go`: the step `the mock API echoes every request to "<path>"` makes the route return the received method, URL, headers (as a JSON object) and body as a JSON body, and also every header as a response header. — done by `/create-tests` (`SetEcho`, plus `the mock API echoes every request to "<path>" [with status N]`; the echo JSON is `{method,path,query,headers,body}` and every request header is reflected as `X-Echo-<Header>`). Also added `the "<var>" variable is set to the port of the mocked api url`.
- [ ] T008 [P] Test hooks in `test/integration/steps/hooks_test.go`:
  - Every scenario gets a free port in `XENON_MCP_PORT` (`{{mcp_port}}`, `{{mcp_url}}`); done by `/create-tests`.
  - `startApp()` starts the MCP server through its dependency getter and adds `binding.GetAIAccessService()` to the binding list.
  - `after()` stops the server.
  - Inject `suite.window` with `external.SetWindowFocuser`.
  - The last three are done by `/implement-task` once the production symbols exist.

---

## Phase 2: Foundational (blocks every story)

- [ ] T009 [P] Add the enums in `internal/domain/enums/` (+ value packages): `historysource` (user, ai), `historykind` (send, action), `historyfilter` (all, user, ai), `aiaccessstate` (off, running, error), `approvaloutcome` (pending, approved, denied, timed_out).
- [ ] T010 [P] Add the entities in `internal/domain/entity/` per data-model.md, each with table tests:
  - `ai_access_settings.go` (defaults, port validation → `SET-04400`) and `ai_access_status.go`.
  - `send_origin.go` and `ai_activity.go` (args capped at 2,048 bytes with `…`).
  - `ai_approval.go`: the transitions; a second answer → `MCP-05409`.
  - `known_secrets.go`: forms raw/QueryEscape/PathEscape/JSON-escaped/base64; longest-first `Mask`; whole-JSON-string rule under 4 characters; `LogValue` with counts only.
- [ ] T011 Extend `internal/domain/entity/app_settings.go`:
  - Add `HistoryUserCap`, `HistoryAICap` (default 1000, range 100–20000 → `SET-02400`/`SET-03400`) and `AIAccess`.
  - `Validate()` returns the first failing code in this order: script timeout, user cap, AI cap, port.
  - Add tests.
- [ ] T012 Extend the history entities:
  - `internal/domain/entity/history_entry.go` gains `Kind`, `Source`, `Client` and `Activity`. `Matches` also searches the activity tool, action and target name.
  - `history.go` replaces `HistoryCap` with `HistoryCaps` and `DefaultHistoryCaps`, plus `Compact(oldestFirst, caps)` per source and `WithoutSource`.
  - `history_query.go` gains a `Source` filter.
  - Update the existing tests in `history_test.go` and `storage_rules_test.go` to the per-source semantics.
- [ ] T013 Update the settings persistence:
  - `internal/integration/persistence/model/app_settings.go` adds `historyUserCap`, `historyAiCap` and `aiAccess{enabled,port,skipApprovals}`. Missing keys read as defaults.
  - Round-trip it in `persistence/registry.go`, and add a unit test with an old `app.json` that has no new keys.
- [ ] T014 Update the history persistence:
  - `internal/integration/persistence/model/history_entry.go` gains `kind`, `source`, `client` and `activity` (all omitempty). Old lines decode as user send.
  - `persistence/history.go` takes caps in `Save` and `Find` (compact on append; cap in memory on read).
  - `Clear(localId, source)`: `all` truncates; `user`/`ai` rewrite the file atomically, keeping the other source.
  - Update `history_test.go`.
- [ ] T015 Change the application adapters and services for settings and history:
  - `internal/application/adapter/manage_app_settings.go` and `service/manage_app_settings.go` handle the new fields.
  - `adapter/manage_history.go` and `service/manage_history.go`: `List(localId, query{source,…})`, `Clear(localId, source)`, caps read from settings. Publish `HistoryChanged` after every append and clear.
- [ ] T016 Add the publisher DTOs and events:
  - `internal/integration/publisher/dto/` gains `ai_access_status_changed.go` (`ai:access-status-changed`), `ai_approval_requested.go`, `ai_approval_resolved.go`, `history_changed.go` (`history:changed`) and `workspaces_changed.go` (`workspaces:changed`).
  - Add the publisher functions in `internal/integration/publisher/ai_events.go`.
  - Register them in `internal/infra/server/event/events.go`.
- [ ] T017 Add the ports in `internal/application/usecase/`:
  - `serve_ai_access.go`: `Apply(settings) AIAccessStatus`, `Status()`.
  - `focus_window.go`: `Focus()`.
  - `collect_known_secrets.go`: the active workspace's vault secrets and OAuth tokens, through the existing vault persistence `SecretsFor` and the OAuth token read.
  - `mask_known_secrets.go`: pure, with a table test.
  - Add the infra adapter `internal/infra/adapter/window_focuser.go` (Wails `Restore` + `Focus`, a no-op without an app), with `external` getter/setter in `internal/infra/dependency/external/window_focuser.go`.
- [ ] T018 Add the application services, adapter interfaces in `internal/application/adapter/` and implementations in `internal/application/service/`:
  - `guard_ai_output.go`: a session `KnownSecrets`; `Guard(localId, text)` unions the collected secrets, then masks; `Register(values...)`.
  - `manage_ai_access.go`: Status, and Apply on settings save.
  - `notify_ai_change.go`: maps the change kinds to the existing and new events.

  Each gets service tests with mocked ports.
- [ ] T019 Add the controller `internal/integration/entrypoint/controller/ai_access.go`, following the full controller pattern (validation, error handler, panic recovery):
  - `Status`, `Guard`, `Notify` now.
  - `PendingApprovals`, `AnswerApproval` and `RequestApproval` in US5.
  - `Record` in US6.

  Add DTOs in `entrypoint/dto/`: `ai_access_status.go`, `ai_guard.go`, `ai_change.go`.
- [ ] T020 Build the MCP server skeleton in `internal/infra/server/mcp/`:
  - `server.go`: an `http.Server` on `127.0.0.1:<port>` (the override from T002) serving `/mcp` through `mcp.NewStreamableHTTPHandler` (stateful). It implements the `serve_ai_access` port:
    - `Apply` stops, rebinds and starts within 2 s; a port in use gives status error `MCP-01409`.
    - `Status` returns the address and client count.
    - Status changes emit `ai:access-status-changed`.
  - `guard.go`: Host/Origin middleware returning 403 (design decision R3).
  - `call.go`: the per-call pipeline from contracts/mcp-tools.md → Conventions. It resolves `clientInfo.name` from the session and the active local id (`MCP-06404` when needed by the tool), calls the handler, compacts, serializes, calls `AIAccess.Guard`, and maps errors to tool errors carrying the Xenon code (FR-025).
  - `compact.go`: omit-empty, 100-row paging with an opaque cursor, the 16 KB body cap. Body text is guarded whole **before** any cap or slice (contract pipeline step 4).
  - `registry.go`: tool registration with annotations and the one-sentence descriptions from the contract.
- [ ] T021 Wire the lifecycle and DI:
  - Register a Wails service in `internal/infra/server/router/router.go` whose `ServiceStartup` calls `Apply(settings)` and whose `ServiceShutdown` stops the server (and, from US5, denies pending approvals).
  - Add the `binding/ai_access.go` binding (`Status` now).
  - Add the dependency getters in `internal/infra/dependency/{controller,binding,application}/ai_access.go` and `dependency/server/mcp.go`.

**Checkpoint**: the app starts an MCP server with zero tools; settings and history understand the new fields; the
existing suite is still green (`make bdd`).

---

## Phase 3: User Story 1 — Connect, browse, send (P1) 🎯 MVP

**Goal**: An AI client connects over HTTP or stdio, browses the tree level by level, and sends saved or ad-hoc
requests. **Independent test**: `connect.feature`, `stdio_bridge.feature` and `browse_and_send.feature` green.

- [x] T022 [P] [US1] Write `test/integration/features/mcp/connect.feature` (`@all @xen_008 @mcp`):
  - Connect and list tools: 24 tools; names; each description is one sentence; hints match the contract's hint summary (12 R, 2 W, 10 D; idempotent per Conventions) — SC-007.
  - The tools/list payload is 24,000 characters or fewer (SC-005).
  - Switching AI access off or on, and changing the port, through `SettingsService.Save` (US1 sc. 6).
  - Host or Origin of a non-loopback address → 403 (sc. 7), with a warning in the log naming the reason (`the log output should contain`).
  - A request body over 4 MiB → 413 (FR-002).
  - Port in use → status error `MCP-01409`, then picking another port → running.
  - Unknown action → `MCP-07400` as a tool error.
  - Connecting and listing tools take less than 1000 ms (`the last call should have taken less than 1000 ms`, SC-001).
  - Two clients (`the MCP client is named` A, then B) create items concurrently. Both succeed and both items exist (edge case "Two AI clients at once").
- [x] T023 [P] [US1] Write `test/integration/features/mcp/stdio_bridge.feature`: tools listed through the bridge; AI access off → exit 1 with the "off" message; server not running → exit 1 with the "not running" message; a non-Xenon HTTP server on the port (mock API route) → exit 1 with the "in use by another program" message, and nothing is forwarded (contracts/stdio-bridge.md).
- [x] T024 [P] [US1] Write `test/integration/features/mcp/browse_and_send.feature`:
  - `tree list` returns the top level only, with child counts; `parent_id` gives one level; `depth: "all"` gives the full tree.
  - Search, and get for each kind.
  - Send a saved request and an ad-hoc request against the mock API: status, headers, timing, tests; history recorded; nothing saved for the ad-hoc request.
  - A 40 KB text body → `truncated`, `size`, then `response body` with offset/length; a binary body → `binary: true`, then `export response_body` writes the file.
  - Paging with a cursor over 150 requests.
  - SC-002: browse to a saved request, send it and read its result in 4 calls (`tree list`, `tree list parent_id`, `send`, `response body`).
- [ ] T025 [US1] Run T022–T024 → **red** (`make bdd TAG=@mcp`). This is the red proof `/create-tests` commits; re-run it to confirm before implementing.
- [ ] T026 [P] [US1] Implement the `workspace` (read) tool in `internal/infra/server/mcp/tool_workspace.go` (contract §1).
- [ ] T027 [P] [US1] Implement the `tree` (read) tool in `internal/infra/server/mcp/tool_tree.go` (contract §2, actions list/get/settings/search/token), using `TreeSearch.GetTree` with depth trimming in `compact.go`.
- [ ] T028 [US1] Add the send origin:
  - `internal/integration/entrypoint/dto/send_request.go` gains `Origin` (`json:"-"`). The Wails binding sets `user`.
  - Pass the origin through `controller/request_send.go` → `application/adapter/send_request.go` → `service/send_request.go`.
  - The history entry carries `Source` and `Client`.
  - The service also calls `GuardAIOutput.Register` with the send's `VariableContext.SecretValues()` and the derived basic-auth value (design decision R6).
- [ ] T029 [US1] Implement the `send` tool in `internal/infra/server/mcp/tool_send.go` (contract §22, without the approval branch, which is US5): saved → `RequestGet` then `RequestSend`; ad hoc → `RequestSend`; tab `mcp:<session-id>`; `send_unresolved` and `send_expired_token` passed through (FR-029).
- [ ] T030 [US1] Implement the `response` (read) tool in `internal/infra/server/mcp/tool_response.go` (contract §6; mask the whole body, then slice). Implement the `export.response_body` action in `tool_export.go`: it writes the masked text, and a binary body containing a known secret form → `MCP-02403`. The rest of `export` is US4.
- [ ] T031 [US1] Implement the stdio bridge in `internal/infra/mcpbridge/bridge.go` and the `cmd/xenon/main.go` early branch for `mcp`, before the Wails app is built (contracts/stdio-bridge.md; design decision R4), including the `serverInfo.name == "xenon"` check. Add a unit test with a fake server for the pump.
- [ ] T032 [US1] Run T022–T024 → **green**; run the full `make bdd`.

---

## Phase 4: User Story 2 — Secrets write-only (P1)

**Goal**: The AI can write and use secrets but never read them. **Independent test**: `secrets_write_only.feature` and
`leak_guard.feature` green, with 0 plaintext hits.

- [x] T033 [P] [US2] Write `test/integration/features/mcp/secrets_write_only.feature`:
  - `env_edit add_var` with secret: the vault holds the value; the workspace file holds only `secret: true`; the response shows `•••••`.
  - Replace a value.
  - `env get` / `env inspect` → secret masked, plain values in full.
  - `snippet generate include_secrets:true` and `export run include_secrets:true` → `MCP-02403`.
  - `env_edit set_vars` turning a secret plain → `MCP-02403`, unchanged; plain → secret allowed (moves to the vault).
  - Locked vault (passphrase mode) send → `SND-06409` passed through.
  - The tool list has no reveal, copy or vault tool (US2 sc. 1–2, 4–6, 8–9).
- [x] T034 [P] [US2] Write `test/integration/features/mcp/leak_guard.feature` against `the mock API echoes every request to "/echo"`:
  - A Scenario Outline over secret value × form (raw, URL-encoded, JSON-escaped, base64; basic auth `user:secret`) × tool (`send`, `response last`, `response body`, `history list/get`, `console list/get`, `snippet generate`, `cookies list`, `env get/inspect`, `tree get/settings/search`, `docs get`, `workspace active`, `git status/changes`, `export options`, `import detect`). Each asserts `the response should not contain "<form>"` and contains `•••••`.
  - The same matrix for an OAuth 2.0 access token stored in the vault (FR-019).
  - Error messages: a failing call whose error text would contain the secret (an echo target returning 500 with the token in its body) returns it masked.
  - Slicing: a secret placed across the 16,384-byte boundary of an echoed body, and across a `response body` offset/length boundary, never appears in either part.
  - AI-written files: `export response_body` and `tree_edit create example from_last_response` on an echoed secret write `•••••` (`the file … should not contain`); the same body containing the secret as binary → `MCP-02403`.
  - Scripts: one copies the secret into a plain variable (`env get`), one into a test name (`send` tests), one into `console.log`.
  - A replaced secret: the old value echoed in a cached response is still masked.
  - A 3-character secret is masked as a whole value, but not inside words.
  - `the log output should not contain` the secret. The leak guard never alters the bytes the mock target received (FR-021).
  - `export run` (Xenon and Postman) to a path: the exported file does not contain any form of the secret (SC-004 "exports").
  - Redaction on disk: `env_edit add_var` with `secret:true` and a value, and `env_edit set_current` on a secret row, both while the vault is locked (`VLT-04409`); `tree_edit save_settings` with a secret variable value; and `import run` with a passphrase. The workspace's `history.jsonl` does not contain any of the values or the passphrase (constitution II; FR-031). Written as `@args_on_disk` in `history/ai_activity.feature`, because it needs a passphrase-mode vault while leak_guard's Background uses the keychain.
  - The short-value rule (masked only as a whole field value under 4 characters) is locked by the `known_secrets.go` table tests in T010, not by BDD: the MCP text nests escaped JSON, where a whole-value assertion cannot be phrased reliably.
  - Tools that only show a secret when one is planted literally (`workspace`, `git`, `import detect`, `export options`) are covered by the registry walk in T036, which checks that the guard wraps every tool.
- [ ] T035 [US2] Run T033–T034 → **red**.
- [ ] T036 [US2] Wire the leak guard into every result:
  - `internal/infra/server/mcp/call.go` calls `AIAccess.Guard` on the serialized result and on error messages.
  - `collect_known_secrets` includes OAuth tokens.
  - Unit test: the guard runs for every registered tool (registry walk).
- [ ] T037 [US2] Add the refusals in `internal/infra/server/mcp/refusals.go` (design decision R7), used by `tool_snippet.go`, `tool_export.go`, `tool_env_edit.go` and `tool_tree_edit.go` (`save_settings` variables). The secret→plain check compares against `VariableTable.GetTable`. `confirm_reveal` is never set.
- [ ] T038 [US2] Implement the `env` (read) tool in `tool_env.go` (contract §3) and `env_edit` in `tool_env_edit.go` (contract §15; `secret_value` handling), plus `snippet` in `tool_snippet.go` (contract §7, `IncludeSecrets=false`).
- [ ] T039 [US2] Run T033–T034 → **green**; run the full `make bdd`.

---

## Phase 5: User Story 3 — Authoring (P2)

**Goal**: The AI builds and edits collections, and the UI refreshes. **Independent test**: `authoring.feature` green.

- [x] T040 [P] [US3] Write `test/integration/features/mcp/authoring.feature`:
  - Create a collection, then a folder, then a request. Files match a binding-created item (`the file … should equal` with a resource fixture), and a `workspace:tree-changed` event is emitted within 1000 ms (SC-003).
  - Update fields; a request update emits `request:changed-on-disk` with the id.
  - Rename, move, duplicate; `delete target:item`.
  - Collection and folder settings save.
  - Examples: from the last response, rename, delete.
  - Docs: `docs get` YAML; `docs plan`; `docs_edit apply` without `confirm_deletes` leaves deletes unapplied, with `confirm_deletes:true` applies them.
  - Hints: `delete` and `docs_edit` are destructive; `tree` is read-only.
- [ ] T041 [US3] Run T040 → **red**.
- [ ] T042 [P] [US3] Implement `tree_edit` in `tool_tree_edit.go` (contract §14, including the oauth actions with metadata-only output; `create example from_last_response` writes the masked response).
- [ ] T043 [P] [US3] Implement `docs` and `docs_edit` in `tool_docs.go` and `tool_docs_edit.go` (contract §5, §16).
- [ ] T044 [P] [US3] Implement `delete` in `tool_delete.go` (contract §23, every `target`).
- [ ] T045 [US3] Notify after each successful write in `call.go`, mapping tool+action → change kinds (design decision R9) → `AIAccess.Notify`. Frontend: no code change is needed. `app/bootstrap.ts` already reloads the sidebar on `workspace:tree-changed` (when its `localId` is active) and routes `request:changed-on-disk` to `useTabsStore.handleChangedOnDisk`, which covers dirty tabs. Add a regression Vitest case in `app/bootstrap.test.ts` asserting both subscriptions.
- [ ] T046 [US3] Run T040 → **green**; run the full `make bdd`.

---

## Phase 6: User Story 4 — Environments, workspaces and the rest (P2)

**Goal**: Full parity for environments, workspaces, history reads, import/export, snippets, cookies, console, scripts,
settings and git. **Independent test**: `environments.feature`, `workspaces.feature` and `other_tools.feature` green.

- [x] T047 [P] [US4] Write `test/integration/features/mcp/environments.feature`: list; `set_active` emits `variables:changed` and the UI's active environment changes; create, update, duplicate; `delete target:environment`; `set_current` keeps the file unchanged; `set_vars` changes the file's initial value; reset and persist (US4 sc. 1–2).
- [x] T048 [P] [US4] Write `test/integration/features/mcp/workspaces.feature`: list marks the active workspace; create, open, rename; `switch` emits `workspaces:changed` with the new active id; `delete target:workspace` moves it to the trash (trash mock); with no active workspace, `tree list` → `MCP-06404` while `workspace list` and `workspace_edit create` still work; `tree get` with an id that belongs to another (non-active) workspace returns not found (US4 sc. 7, FR-008, workspace-boundary edge case).
- [x] T049 [P] [US4] Write `test/integration/features/mcp/other_tools.feature`, with git in its own `test/integration/features/mcp/git.feature`, using the existing git, cookie and import fixtures:
  - Import a Postman collection by path, including clashes, matching a binding import.
  - Export Xenon and Postman to a path, matching the binding export with secrets off.
  - Snippet languages and generate.
  - Cookies: list, upsert, allowlist, `delete target:cookie/cookie_domain/cookies`.
  - Console: list and get; `delete target:console`.
  - Scripts: check, reference, snippets.
  - Settings: get (no `ai_access`); `settings_edit` caps; `settings_edit` with `ai_access` → `MCP-02403`.
  - Git: status, changes, commit, pull with a conflict session, then resolve and finish, `git_push`, `delete target:git_link`.
  - OAuth `tree token` metadata only.
  - FR-016 events: `cookies_edit upsert` emits `cookies:changed`; `git_edit commit` and `git_push` emit `git:status-changed`; `import run` emits `workspace:tree-changed`.
- [ ] T050 [US4] Run T047–T049 → **red**.
- [ ] T051 [P] [US4] Implement `workspace_edit` in `tool_workspace_edit.go` (contract §13); `WorkspacesChanged` via Notify.
- [ ] T052 [P] [US4] Implement `history` (read) in `tool_history.go` (contract §4; source filter passed through).
- [ ] T053 [P] [US4] Implement `import` and `export` in `tool_import.go` and `tool_export.go` (contract §20, §21).
- [ ] T054 [P] [US4] Implement `cookies` and `cookies_edit` in `tool_cookies.go` and `tool_cookies_edit.go` (contract §8, §17).
- [ ] T055 [P] [US4] Implement `console`, `scripts`, `settings` and `settings_edit` in `tool_console.go`, `tool_scripts.go` and `tool_settings.go` (contract §9, §11, §12, §19).
- [ ] T056 [P] [US4] Implement `git`, `git_edit` and `git_push` in `tool_git.go`, `tool_git_edit.go` and `tool_git_push.go` (contract §10, §18, §24).
- [ ] T057 [US4] Frontend: subscribe to `workspaces:changed` in `frontend/src/app/bootstrap.ts` (reload the list, activate `active_local_id`) via `platform/events`; Vitest case in `bootstrap.test.ts`.
- [ ] T058 [US4] Run T047–T049 → **green**; run the full `make bdd`.

---

## Phase 7: User Story 5 — Approvals (P2)

**Goal**: AI sends to confirm-before-send environments wait for the human, unless "Dangerously skip approvals" is on.
**Independent test**: `approvals.feature` green, plus `ApprovalPrompt.test.tsx`.

- [x] T059 [P] [US5] Write `test/integration/features/mcp/approvals.feature`:
  - Bypass off: the send runs in the background → an `ai:approval-requested` event with client, method, raw URL and environment name and colour → the window is brought to the front. Then:
    - `AIAccessService.AnswerApproval approved:true` → the mock API received 1 request; the result is returned.
    - Deny → `MCP-03403`; the mock API received 0.
    - `120 seconds pass` → `MCP-04408`; 0 requests.
  - A second answer → `MCP-05409`.
  - Two pending approvals resolve in FIFO order.
  - Bypass on (`SettingsService.Save` `skip_approvals:true`) → sent with no event and no focus, and secrets are still masked.
  - `send_unresolved:true` passes the missing-secret acknowledgement; `send_expired_token:true` sends with an expired OAuth token, and without it the existing expired-token error is returned (FR-029).
  - App shutdown with a pending approval → denied, nothing sent.
  - The requesting MCP client disconnects while its approval is pending → `ai:approval-resolved` with outcome denied; nothing sent (FR-027).
- [ ] T060 [P] [US5] Write `frontend/src/features/ai-access/components/ApprovalPrompt.test.tsx`: renders the queue head (client, method, URL, environment colour chip); Approve and Deny call `platform/services/ai-access`; it disappears on `ai:approval-resolved`; it shows a "+N more" count.
- [ ] T061 [US5] Run T059–T060 → **red**.
- [ ] T062 [US5] Add the approval broker:
  - `internal/application/adapter/approve_ai_send.go` + `service/approve_ai_send.go` (design decision R8): an in-memory FIFO map; `Request` focuses the window, publishes, and waits with a `timeutils` `AfterFunc(120s)`; `Answer`; `DenyAll` on shutdown; `DenyForSession(sessionID)` when an MCP session closes (the server's session-close hook in `server.go` calls it).
  - Add `RequestApproval`, `AnswerApproval` and `PendingApprovals` to `controller/ai_access.go`, with DTOs `ai_approval.go` and `ai_approval_answer.go`.
- [ ] T063 [US5] Add the approval branch to `tool_send.go`: on `SND-05409` → if `skipApprovals`, re-send with `confirmed_environment_id`; otherwise `RequestApproval` → approved: re-send; denied: `MCP-03403`; timed out: `MCP-04408`. Extend `binding/ai_access.go` with `PendingApprovals` and `AnswerApproval`, and have `ServiceShutdown` call `DenyAll`.
- [ ] T064 [US5] Add the frontend approval prompt:
  - `frontend/src/features/ai-access/{store.ts, components/ApprovalPrompt.tsx, components/ApprovalPrompt.module.css}`.
  - `platform/services/ai-access.ts` and `platform/mappers/ai-access.ts`.
  - Subscribe to `ai:approval-requested`/`-resolved` in `app/bootstrap.ts` (plus `PendingApprovals` on start).
  - Mount it in `app/App.tsx`.
  - Shared `Dialog`, PRD palette tokens.
- [ ] T065 [US5] Run T059–T060 → **green**; run the full `make bdd` and `npm test`.

---

## Phase 8: User Story 6 — AI activity in History (P2)

**Goal**: Every AI call is visible in History, filterable by source, with per-source caps. **Independent test**:
`ai_activity.feature` and `app_settings.feature` green, plus `HistoryView.test.tsx`.

- [x] T066 [P] [US6] Write `test/integration/features/history/ai_activity.feature` (`@all @xen_008 @ai_activity`):
  - With `the MCP client is named "Claude Code"`: a read call, a write call, a failing call (error code) and a denied send each produce one entry with kind, source, client, tool, action, target, outcome and duration.
  - Sends are not duplicated as action rows. A denied send (approval), a timed-out send, and a send with an invalid argument (`MCP-07400`) each produce one action row (`tool: send`, outcome denied / timed_out / error), and no send row.
  - Action `args` are masked and capped at 2 KB.
  - `HistoryService.List source:user|ai|all`; `Clear source:ai` keeps user sends.
  - Caps: user 10000 / AI 100 → 101 AI calls keep 100 AI and every user entry; lowering a cap applies on the next List.
  - No active workspace → no entry, and the log contains the call.
  - `workspace_edit switch` is recorded in the new workspace.
  - `history:changed` is emitted.
  - Old history lines with no source read as user.
  - Through MCP: `history list source:ai` returns compact rows without `args`; `history get` on an action row returns its masked `args` (FR-035, US6 sc. 7).
- [x] T067 [P] [US6] Write `test/integration/features/settings/app_settings.feature` (`@all @xen_008 @app_settings`): `SettingsService.Get` defaults (1000, 1000, enabled, 47821, skip false); Save validates `SET-02400`, `SET-03400`, `SET-04400` and `SCR-01400`; an old `app.json` without the keys gets the defaults; the AI access fields are applied immediately (status event).
- [ ] T068 [P] [US6] Update `frontend/src/features/sidebar/components/HistoryView.test.tsx`: a filter (All/User/AI) next to Clear History, remembered via UI state; AI rows tagged "AI · <client>"; action rows show tool·action, target, outcome and duration, and expand to masked args; Clear History clears the filtered source; a reload on `history:changed`.
- [ ] T069 [US6] Run T066–T068 → **red**.
- [ ] T070 [US6] Add the recorder:
  - `internal/application/adapter/record_ai_activity.go` + `service/record_ai_activity.go`: append the action entry; errors logged only (FR-034).
  - Redact secret-carrying argument fields by tool and field name **before** the guard (design decision R10): `secret_value` anywhere; `value` in `env_edit` `add_var`/`set_current`; `value` beside `secret:true`; `passphrase` anywhere. The redaction is a pure use case `internal/application/usecase/redact_secret_args.go` with a table test.
  - `call.go` records an action row for any `send` that did not reach the send service (refused, denied, timed out, invalid).
  - `Record` in `controller/ai_access.go`.
  - `call.go` records every call except a `send` that reached the send service, with args redacted then guarded, and logs every call at Info (design decision R12).
- [ ] T071 [US6] Extend the controller and DTOs:
  - `controller/history.go` + `dto/list_history.go`, `dto/clear_history.go`, `dto/history_entry.go` and `dto/history_entry_detail.go` gain the source filter and the kind/source/client/activity fields.
  - `controller/settings.go` + `dto/app_settings.go` gain the new fields and their validation codes.
  - `dto/ui_state.go` + `entity/ui_state.go` + `persistence/ui_state.go` gain `history_source`.
- [ ] T072 [US6] Update the History frontend:
  - `frontend/src/features/sidebar/components/HistoryView.tsx` and a new `HistorySourceFilter.tsx`.
  - `ClearHistoryDialog.tsx` names the filtered source.
  - `platform/services/history.ts` and `platform/mappers/history.ts`.
  - Subscribe to `history:changed` in `app/bootstrap.ts`.
- [ ] T073 [US6] Run T066–T068 → **green**; run the full `make bdd` and `npm test`.

---

## Phase 9: User Story 7 — Settings as a tab (P3)

**Goal**: Settings opens as a main-area tab with every section. **Independent test**: `SettingsTab.test.tsx`,
`AiAccessSection.test.tsx` and `settings-tab.test.ts` green.

- [ ] T074 [P] [US7] Write `frontend/src/features/tabs/settings-tab.test.ts`: `openSettingsTab` creates one pinned `settings` tab and focuses it on a second call; it is persisted in UI state and restored without its draft; the close prompt appears when dirty.
- [ ] T075 [P] [US7] Write `frontend/src/features/settings/components/SettingsTab.test.tsx`:
  - Sections: Scripts, History, AI access (MCP), Git accounts, Shortcuts.
  - One Save validates and saves every field (inline errors for timeout, caps, port; nothing saved on error).
  - The dirty dot.
  - Git account actions act immediately.
  - The gear button and Cmd/Ctrl+, open the tab.
  - With no workspace, a full-page fallback with a close button.
- [ ] T076 [P] [US7] Write `frontend/src/features/settings/components/AiAccessSection.test.tsx`: status line (running address / off / error reason) updated by `ai:access-status-changed`; switch; port; the "Dangerously skip approvals" checkbox with its warning text; the stdio command; the HTTP and stdio configs with copy buttons (clipboard via `platform/`); the residual-risk note.
- [ ] T077 [US7] Run T074–T076 → **red**.
- [ ] T078 [US7] Add the tab kind `settings` in `frontend/src/features/tabs/{model.ts, store.ts}` (a singleton, pinned, reusing the environment-tab dirty and close-prompt logic). Render it in the main-area tab content switch.
- [ ] T079 [US7] Build the Settings tab:
  - `frontend/src/features/settings/components/SettingsTab.tsx` (+ `.module.css`) with the sections `ScriptsSection.tsx`, `HistorySection.tsx` and `AiAccessSection.tsx`, plus the existing `GitAccounts.tsx` and `ShortcutsList.tsx`.
  - `store.ts` holds one draft and dirty flag, and `open()` → `openSettingsTab`.
  - Update `model.ts`, `platform/services/settings.ts` and `platform/mappers/settings.ts` (new fields and codes). Create `platform/services/ai-access.ts` with `Status` if US5 has not created it yet; otherwise add `Status` to it. Either order works.
- [ ] T080 [US7] Replace the dialog:
  - Delete `SettingsDialog.tsx`, `SettingsDialog.module.css` and `SettingsDialog.test.tsx`.
  - `app/App.tsx` drops the dialog mount; the hotkeys and `SettingsButton` call `openSettingsTab`; with no workspace, render `SettingsTab` full page.
  - Move any styles still needed into `SettingsTab.module.css`.
- [ ] T081 [US7] Run T074–T076 → **green**; run `npm run lint && npm run typecheck && npm test`.

---

## Phase 10: Polish & Cross-Cutting

- [ ] T082 Performance check in `test/integration/features/mcp/browse_and_send.feature`: a scenario with a 1,000-request generated workspace asserts `tree list` top level is 8,000 characters or fewer (SC-005), that non-send calls answer within 200 ms using the existing step `the last call should have taken less than 200 ms` (extended in T004 to time MCP calls too), and that a `send` to an instant mock API route completes in less than 100 ms (SC-006).
- [ ] T083 [P] Run `make sort-imports` and `make validate`; fix lint findings in the changed files only.
- [ ] T084 [P] Run `/refactor-diff` until clean (the architecture skills; no SDK imports outside `infra`/`cmd`/tests, checked with `grep -rn "modelcontextprotocol" internal --include=*.go | grep -v internal/infra`, which must return nothing).
- [ ] T085 Walk through every step of `quickstart.md` manually with Claude Code (HTTP and stdio) on macOS, and record the results in the PR description draft.
- [ ] T086 Re-run the full suite (`make bdd`, `npm test`, `make validate`) → green; push the `xen-008` branch (no PR).

---

## Dependencies & Execution Order

- Setup (T001–T008), then Foundational (T009–T021), then the stories.
- US1 (T022–T032) is the MVP and comes before US2, because US2's leak tests need `send`, `response` and `tree`.
- US3, US4, US5 and US6 depend on US1 and US2, and are otherwise independent of each other. Exceptions: US4's
  `other_tools.feature` uses `delete` from US3 (T044); US6's send-row assertions use US5's denied outcome (T063).
- US7 (frontend only) depends only on Foundational (T011, T013 for the settings fields; T020 for status).
- Polish comes last.

## Parallel opportunities

- Setup: T002–T008 run together.
- Foundational: T009 and T010 run together.
- Feature files within a story run together: T022–T024, T033–T034, T047–T049, T059–T060, T066–T068, T074–T076.
- Tool files within a story run together: T026–T027, T042–T044, T051–T056.
- US7 can run alongside US3–US6.

## Implementation Strategy

1. **MVP** = Setup + Foundational + US1 + US2. An AI can connect, browse and send, and secrets are already
   write-only. US2 ships with US1, never later, because an MVP without the secret guard would violate constitution II.
2. Then US3 and US4 (parity), then US5 (approvals), then US6 (History activity).
3. US7 lands in parallel or last. Settings stays usable as a dialog until T080 replaces it.
