# Implementation Plan: MCP Server — Full Xenon Access for AI Clients, Secrets Write-Only

**Branch**: `xen-008` | **Date**: 2026-10-13 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/XEN-008-mcp-server/spec.md`

## Summary

Xenon runs an MCP server (official Go SDK `github.com/modelcontextprotocol/go-sdk` v1.8.0) on
`http://127.0.0.1:<port>/mcp` (Streamable HTTP; default port **47821**). A `xenon mcp` subcommand bridges stdio to it.
The server is a second entrypoint next to the Wails bindings. It exposes **24 resource tools with an `action`
argument** (12 read-only, 10 that change data, of which 8 are hinted destructive, and 2 destructive-only; see [contracts/mcp-tools.md](contracts/mcp-tools.md)). Each
tool adapts its arguments to the existing controller DTOs and calls the **same controllers** the bindings call, then
runs the result through a compactor (ids and names, a 16 KB body cap) and a final **leak guard**.

The guard masks every known secret value (vault secrets, OAuth tokens, values seen this session, in raw, URL-escaped,
JSON-escaped and base64 forms) with `•••••`. Anything that would reveal a secret is refused at the tool layer:
reveal/copy, include-secrets snippets and exports, secret→plain, and vault setup/unlock/reset.

A send blocked by an environment's existing "Confirm before send" check (`SND-05409`) goes to an **approval broker**.
The broker brings the window to the front, emits `ai:approval-requested`, and waits up to 120 s for the UI's
Approve/Deny before re-sending with the existing acknowledgement. The "Dangerously skip approvals" setting bypasses
the broker.

Every MCP call is recorded in the active workspace's **History**:
- History entries gain `source` (user/ai), `client` and `kind` (send/action).
- Caps are per source and configurable (Settings → History, default 1,000 each, range 100–20,000).
- The History view gets an All / User / AI filter, and Clear History clears what the filter shows.

MCP writes publish the same UI events as file-watcher changes, so the open window refreshes. **Settings moves from a
modal dialog to a main-area tab** with sections for Scripts, History, AI access (MCP), Git accounts and Shortcuts.

## Technical Context

**Language/Version**: Go 1.26 (toolchain 1.27.2), TypeScript 5 + React 19

**Primary Dependencies** (new): `github.com/modelcontextprotocol/go-sdk` **v1.8.0**. It provides the MCP server, the
Streamable HTTP handler with built-in localhost/DNS-rebinding protection, `StdioTransport` and
`StreamableClientTransport` for the bridge, and tool annotations. It is imported **only** in `internal/infra/` and
`cmd/`, the same rule that applies to Wails.

**Storage**: unchanged locations.
- `<DataDir>/app.json` → `settings` gains `historyUserCap`, `historyAiCap` and `aiAccess {enabled, port, skipApprovals}`.
- `<DataDir>/workspaces/<localId>/history.jsonl` lines gain `kind`, `source`, `client` and `activity{…}`. Old lines read
  as `kind=send, source=user`.

**Testing**:
- godog BDD in `test/integration/` drives a real MCP client (SDK `StreamableClientTransport`) against the real server
  started by `startApp()` on an ephemeral port (`XENON_MCP_PORT=0`).
- Only the outer boundaries are faked: the HTTP target (mock API, including an echo route), the keychain, the clock
  (approval timeout), the window focuser and the event emitter.
- Vitest + RTL cover the frontend (Settings tab, approval prompt, History filter).

**Target Platform / Project Type**: desktop app (Wails v3) on macOS, Windows and Linux.

**Performance Goals**: SC-001 (connect and list tools in under 1 s); SC-006 (non-send calls under 200 ms on 1,000
requests; a send to an instant local target in under 100 ms); SC-005 (tool list of 6,000 tokens or fewer; top-level tree of 2,000 tokens or
fewer).

**Constraints**:
- Loopback only; Host/Origin checks.
- No new business logic in the MCP layer (FR-007).
- The leak guard never touches what is sent, stored or shown to the human (FR-021).
- History writes never fail a tool call (FR-034).

**Scale/Scope**: workspaces up to 1,000+ requests; up to a few concurrent AI clients; history caps up to 20,000 per
source.

## Constitution Check

- [x] **I. Local-first** — The server is local to the user's own app and bound to loopback. It is not a Xenon cloud
  service. Settings and history stay in app-data. Tools act only on the active workspace, so the workspace boundary
  holds (FR-008, edge case "item from another workspace"). ✅
- [x] **II. Secrets** — Secret values never leave the vault toward the AI:
  - Reveal/copy/vault tools are absent (FR-015).
  - Include-secrets and secret→plain are refused (FR-018).
  - Every output passes the leak guard (FR-019/020).
  - History action rows store arguments only after secret-carrying fields are redacted by name and the guard has
    run (research R10). The guard never writes its known-secret set anywhere (data-model).
  - The leak BDD suite covers every data-returning tool × 4 encodings (SC-004).

  The two accepted residual risks are the product owner's documented decisions (spec Context; research R6): a value a
  script transforms, and an AI sending `{{secret}}` to a destination it controls. ✅
- [x] **III. Go owns logic** — All MCP, masking, approval and history logic is in Go. React only renders the Settings
  tab, the approval prompt and the History filter, through `platform/`. The SDK is imported only in `infra/` and
  `cmd/`. ✅
- [x] **IV. Architecture skills** — Every new file has a layer (Project Structure below). MCP tool handlers are thin
  adapters in `infra/server/mcp/` (the router role, like `binding/`) that call controllers. MCP-specific behaviour is
  in controllers, services and use cases. ✅
- [x] **V. BDD-first** — Feature files are listed per requirement (BDD Coverage below). Only outer boundaries are faked,
  and a real MCP client talks to the real server. ✅
- [x] **VI. Postman parity, nothing more** — MCP is beyond the current PRD. **Resolved by FR-030**: the PRD gains an
  "AI access (MCP)" section in this task (task T003), so the feature is in scope before any code lands. The product
  owner requested it (GitHub issue #3). Recorded in Complexity Tracking. ✅
- [x] **VII. Cross-platform** —
  - TCP loopback works the same on all three OSes.
  - The `xenon mcp` bridge is the same binary; the ready-to-paste config uses the absolute executable path from
    `os.Executable()` with OS-native separators.
  - Cmd+, maps to Ctrl+, on Windows and Linux.

  ✅

**Post-design re-check (after Phase 1)**: all gates still pass. The design added no persistence outside app-data, no
Wails/SDK imports below `infra`, and no secret-bearing field in any DTO returned to MCP.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-008-mcp-server/
├── spec.md · plan.md · research.md · data-model.md · quickstart.md
├── contracts/
│   ├── mcp-tools.md        # the 24 tools: names, hints, actions, args, compact outputs, refusals
│   ├── bindings.md         # new/changed Wails bindings + events for the UI
│   └── stdio-bridge.md     # `xenon mcp` behaviour and client configs
├── checklists/
└── tasks.md
```

### Source code (additions and changes)

```text
cmd/xenon/main.go                         CHANGE: `xenon mcp` → infra/mcpbridge.Run() before any Wails app is built
                                           (so SingleInstance never fires for the bridge)

internal/domain/
├── entity/
│   ├── app_settings.go                   CHANGE: + HistoryUserCap, HistoryAICap, AIAccess (validation + defaults)
│   ├── ai_access_settings.go             NEW: Enabled, Port, SkipApprovals; DefaultAIPort = 47821; port range
│   ├── ai_access_status.go               NEW: state (off|running|error), address, reason
│   ├── ai_approval.go                    NEW: approval request (id, client, method, raw URL, env id/name/colour,
│   │                                           created-at, outcome) + transitions
│   ├── ai_activity.go                    NEW: action entry payload (tool, action, target, outcome, duration, args)
│   ├── history.go / history_entry.go     CHANGE: Kind, Source, Client, Activity; per-source caps
│   │   history_query.go                         (Compact(caps), Filter(source)); HistoryCap const → defaults
│   ├── known_secrets.go                  NEW: secret value set + encoded forms (raw/url/json/base64/basic),
│   │                                           longest-first Mask(text) with the 4-char rule
│   └── send_origin.go                    NEW: who sent (user | ai + client) carried by the send command
└── enums/  historysource (user|ai), historykind (send|action), aiaccessstate (off|running|error),
            approvaloutcome (pending|approved|denied|timed_out)  (+ value pkgs)

internal/application/
├── usecase/
│   ├── mask_known_secrets.go             NEW (pure): apply KnownSecrets to text
│   ├── collect_known_secrets.go          NEW (port-using): vault secrets + OAuth tokens of the active workspace
│   ├── serve_ai_access.go                NEW (port): Apply(settings) / Status() — implemented in infra/server/mcp
│   ├── focus_window.go                   NEW (port): bring the main window to front — infra adapter
│   └── (reuse) reject_unconfirmed_send.go, mask_secrets.go, snippet secret forms
├── adapter/
│   ├── guard_ai_output.go                NEW interface
│   ├── approve_ai_send.go                NEW interface
│   ├── record_ai_activity.go             NEW interface
│   ├── manage_ai_access.go               NEW interface (status, apply on settings save)
│   ├── notify_ai_change.go               NEW interface (publish UI refresh events after MCP writes)
│   └── manage_history.go / manage_app_settings.go / send_request.go   CHANGE signatures (source filter, caps, origin)
└── service/
    ├── guard_ai_output.go                NEW: session-scoped KnownSecrets (seen-set grows, never shrinks) + mask
    ├── approve_ai_send.go                NEW: in-memory broker, FIFO, 120 s timeout via timeutils clock, quit → deny
    ├── record_ai_activity.go             NEW: append action entry (masked args ≤ 2 KB), errors logged only
    ├── manage_ai_access.go               NEW: status; applies settings to the server port
    ├── notify_ai_change.go               NEW: TreeChanged / RequestChangedOnDisk / VariablesChanged / HistoryChanged
    │                                          / CookiesChanged / GitStatusChanged / WorkspacesChanged
    ├── manage_app_settings.go            CHANGE: new fields; after save → manage_ai_access.Apply; trim caps lazily
    ├── manage_history.go                 CHANGE: source filter on list and clear; caps from settings
    └── send_request.go                   CHANGE: history entry carries SendOrigin; guard registers the send's
                                                 secret forms (incl. derived basic-auth header) for this session

internal/integration/
├── entrypoint/
│   ├── controller/
│   │   ├── ai_access.go                  NEW: Status, AnswerApproval, PendingApprovals (UI); Guard, Record,
│   │   │                                        Notify, RequestApproval (MCP router) — full controller pattern
│   │   ├── history.go                    CHANGE: source filter on List/Clear
│   │   ├── settings.go                   CHANGE: new fields
│   │   └── request_send.go               CHANGE: accepts origin (set by the entrypoint, never by tool args)
│   └── dto/  ai_access_status.go, ai_approval.go, ai_approval_answer.go, ai_activity.go, ai_guard.go,
│             history_query.go (+source), clear_history.go (+source), app_settings.go (+fields),
│             send_request.go (+origin, json:"-" for bindings), history_entry.go (+kind/source/client/action)
├── persistence/
│   ├── history.go                        CHANGE: per-source compaction, clear-by-source rewrite
│   ├── model/history_entry.go            CHANGE: kind/source/client/action fields (omitempty; old lines → user send)
│   └── model/app_settings.go             CHANGE: new fields with defaults on read
└── publisher/
    ├── ai_events.go                      NEW: AIApprovalRequested, AIApprovalResolved, AIAccessStatusChanged,
    │                                          HistoryChanged, WorkspacesChanged
    └── dto/ (matching event DTOs + event names)

internal/infra/
├── server/
│   ├── mcp/                              NEW (router role; only package importing the SDK besides mcpbridge)
│   │   ├── server.go                     http.Server on 127.0.0.1:<port>, StreamableHTTPHandler, Apply/Status,
│   │   │                                 Host/Origin guard middleware, shutdown on app quit (ServiceShutdown)
│   │   ├── registry.go                   registers the 24 tools with annotations + one-sentence descriptions
│   │   ├── call.go                       per-call pipeline: client name → controller → compact → guard → record →
│   │   │                                 notify; errors → tool errors with Xenon code (FR-025)
│   │   ├── compact.go                    list paging (100 + cursor), omit-empty, body cap 16 KB / 64 KB reads
│   │   └── tool_<resource>.go            one file per tool: args struct → controller DTO(s) → controller call
│   ├── binding/ai_access.go              NEW Wails binding: Status, AnswerApproval, PendingApprovals
│   ├── binding/history.go, settings.go   CHANGE: new fields pass-through
│   ├── event/events.go                   CHANGE: register the new events
│   └── router/router.go                  CHANGE: register AIAccess binding + the MCP server service
├── adapter/window_focuser.go             NEW: wails window Restore()+Focus() (no-op without app)
├── mcpbridge/bridge.go                   NEW: read port from app.json; StdioTransport ⇄ StreamableClientTransport pump;
│                                           exit 1 with one line when unreachable/disabled
├── properties/properties.go              CHANGE: XENON_MCP_PORT override (tests: 0 → ephemeral)
└── dependency/…                          NEW getters for the services/controllers/binding/server; Set* for focuser

frontend/src/
├── features/settings/
│   ├── components/SettingsTab.tsx        NEW (replaces SettingsDialog.tsx, which is deleted) + sections:
│   │   ScriptsSection, HistorySection, AiAccessSection (status, switch, port, skip approvals + warning,
│   │   stdio command, HTTP/stdio configs with copy, residual-risk note), GitAccounts (moved), ShortcutsList
│   ├── store.ts                          CHANGE: draft/dirty/save for all form fields; open() → tabs.openSettingsTab
│   └── model.ts                          CHANGE: field types
├── features/tabs/                        CHANGE: tab kind 'settings' (singleton, pinned, dirty, close prompt,
│                                          restored without edits); no-workspace full-page fallback
├── features/ai-access/                   NEW: ApprovalPrompt.tsx (queue, Approve/Deny, client/method/URL/env),
│                                          store.ts (pending approvals from events + PendingApprovals on start)
├── features/sidebar/components/HistoryView.tsx   CHANGE: source filter (All/User/AI) next to Clear History,
│   HistorySourceFilter.tsx (NEW), ClearHistoryDialog.tsx (clears the filtered source), action rows + AI tag
├── platform/services/ ai-access.ts (NEW), history.ts, settings.ts (CHANGE); platform/mappers likewise
└── app/  App.tsx (SettingsDialog → tab; mount ApprovalPrompt), bootstrap.ts (subscribe to new events)

docs/prd/xenon-mvp-prd.md                 CHANGE: "AI access (MCP)" section + History/Settings changes (FR-030)
```

**Structure Decision**: This is the existing single-repo layout. The MCP server takes the router role in
`internal/infra/server/mcp/`, next to `binding/`. Tool handlers only adapt, and everything with behaviour goes through
controllers, then services, then use cases.

### BDD coverage (feature files `/create-tests` writes)

| Feature file | Locks |
| --- | --- |
| `test/integration/features/mcp/connect.feature` | US1 sc. 1, 3, 6, 7; FR-001–004, FR-022–025, SC-001, SC-005, SC-007 (tool list: 24 tools, hints, one-sentence descriptions) |
| `test/integration/features/mcp/stdio_bridge.feature` | US1 sc. 2; FR-005 (running, disabled, not running) |
| `test/integration/features/mcp/browse_and_send.feature` | US1 sc. 3–5; FR-010, FR-022 (truncation, binary, paging), FR-024 |
| `test/integration/features/mcp/secrets_write_only.feature` | US2 sc. 1–2, 4–6, 8–9; FR-015, FR-017, FR-018 |
| `test/integration/features/mcp/leak_guard.feature` | US2 sc. 3, 7; FR-019–021, SC-004. Scenario Outline over every data-returning tool × raw/url/json/base64 forms, using an echo route; session-seen values after a replace; short-value rule |
| `test/integration/features/mcp/authoring.feature` | US3 sc. 1–5; FR-009, FR-012, FR-013 (docs), FR-016 (events) |
| `test/integration/features/mcp/environments.feature` | US4 sc. 1–2; FR-011 |
| `test/integration/features/mcp/workspaces.feature` | US4 sc. 7; FR-008, "no active workspace" edge case |
| `test/integration/features/mcp/other_tools.feature` | US4 sc. 4–6; FR-013, FR-014 (import, export, snippets, cookies, console, scripts, settings, oauth metadata, git) |
| `test/integration/features/mcp/approvals.feature` | US5 sc. 1–6; FR-026–029 |
| `test/integration/features/history/ai_activity.feature` | US6 sc. 1–8; FR-031–035 |
| `test/integration/features/settings/app_settings.feature` | FR-003 fields, FR-004 apply-now, FR-006 (not changeable via MCP), FR-034 caps validation |

Every XEN-008 feature file is tagged `@all @xen_008 @<area>` (`@mcp`, `@ai_activity`, `@app_settings`), following the
existing `@all @xen_00N @<area>` convention.

Frontend (Vitest): `SettingsTab.test.tsx`, `settings-tab.test.ts` (tabs store), `ApprovalPrompt.test.tsx`,
`HistoryView.test.tsx` (filter and clear), `AiAccessSection.test.tsx` (US7 sc. 1–6; FR-036–038).

New generic steps, added to the shared catalog (`bdd-step-definitions.md`):
- `I call the MCP tool "<tool>" with` (docstring) and `I call the MCP tool "<tool>"`. The result becomes "the
  response", so the existing `the response should …` steps apply.
- `I call the MCP tool "<tool>" in the background with` and `the background MCP call finishes`, for approvals.
- `I list the MCP tools`.
- `the MCP client is named "<name>"`.
- `I post to the MCP endpoint with the headers` (table). The status is asserted with the existing
  `the route response status should be N` step.
- `I run the stdio bridge with the input` (docstring), plus `the bridge output should contain` and
  `the bridge exit code should be N`.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
| --- | --- | --- |
| Principle VI: feature not in the MVP PRD | The product owner requested it in GitHub issue #3 / XEN-008 | Waiting for a separate PRD change only adds ceremony; this task amends the PRD itself (FR-030, task T003) before code |
| New dependency `modelcontextprotocol/go-sdk` | MCP protocol, Streamable HTTP, sessions, stdio transport and tool annotations | Hand-rolling JSON-RPC and Streamable HTTP plus protocol-version negotiation is large and easy to get wrong; the official SDK is maintained by the protocol authors |
