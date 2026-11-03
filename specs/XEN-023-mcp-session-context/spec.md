# Feature Specification: Per-Session MCP Context

**Feature Branch**: `xen-023`

**Created**: 2026-11-03

**Status**: Draft

**Input**: User description: "MCP sessions get their own workspace/environment context, so AI actions stop driving the Xenon UI and parallel AI sessions stop clashing. Today the MCP server and the Xenon window share one 'where am I' state: the active workspace and the active environment. When an AI client switches/opens/creates a workspace or sets the active environment through the Xenon MCP, the user's window follows (sidebar, tabs, environment switcher change under them, and switching even stops the user's running collection runs); and several parallel AI sessions overwrite each other's choice, so session B's send/run/tree quietly runs against the workspace or environment session A picked. Users need: (1) per-session context, seeded from the window at connect; (2) UI decoupled — no MCP action ever changes what the user is looking at; data changes still refresh live and never overwrite unsaved edits; (3) safe concurrent writes — per-workspace serialisation, atomic revision checks, environments gain a revision check; (4) visibility — a short session tag on AI activity and AI-sent history; (5) tool descriptions say the context is per-session. Source: https://github.com/brienze1/xenon/issues/39"

## Context

XEN-008 gave AI clients an MCP server inside Xenon. Every tool call resolves "the current workspace" from the one
active workspace the window also uses, and "the current environment" from that workspace's one active environment,
which is also what the window's environment switcher shows. So the AI and the user share a single cursor, and every
AI session shares it with every other AI session.

### Settled from the code rather than asked

| Question | Answer and the fact that settles it |
| --- | --- |
| Where does an MCP call get its workspace today? | From the app's registry of workspaces: the one marked active, the same one the window shows. |
| What does an MCP `workspace_edit switch` do to the window today? | It activates the workspace for the whole app, publishes "workspaces changed" with the new active id, and the window follows it (resets its stores, loads the new tree, tabs and environments). It also stops any collection run in progress in the previously active workspace. |
| Do MCP `create`/`open` activate the workspace? | Yes: registering a workspace in the registry also marks it active, and the window follows. |
| Where is the active environment kept? | Per workspace, in Xenon's local (non-committed) variable state, next to the current values. The window's environment switcher reads it, and `env_edit set_active` writes it; the window reloads it on any variables change. |
| Which MCP tools resolve variables with "the active environment"? | `send`, `run` (when no `environment` is given), `snippet`, and `env inspect`. Run already accepts an environment override (`environment`: name, id or `none`); send does not. |
| Does an AI send touch a UI tab? | No. Its response is stored under a per-session key that no window tab uses; only history and the cookie jar refresh. |
| Can one session read another session's last response? | No. The `response` tool reads the per-session key of the calling session only, so stored responses are already isolated per session. |
| Are pending approvals per session? | Yes. Approvals carry the session id, and closing a session denies only that session's pending approvals. |
| Are runs tied to a session? | No. A run belongs to its workspace and target; `run stop` and `runs` act on the calling session's current workspace. |
| Can an AI edit overwrite unsaved edits in an open request tab? | No. A request changed on disk while its tab has unsaved edits gets a "changed" notice (keep as new, keep mine, load from disk); a clean tab silently reloads. This stays. |
| Is there a "changed since you read it" check today? | For requests, aliases and examples only: a revision (hash of the file) is compared before saving and a stale save is refused with `REQ-02409`. The comparison and the write are not done as one step, so a write can still slip in between. Environments, folder/collection settings, workspace settings and variable tables have no check. |
| Are writes serialised? | Partly. Tree and request writes share one lock; environment saves, workspace settings saves, local variables (active environment, current values) and vault saves are not serialised against each other. Files are written atomically (temp file + rename), so a file is never half-written — the risk is lost updates, not torn files. |
| Does the window's environment editor send the whole table on save? | Yes: saving a variable table sends every row, so rows added by someone else since the editor loaded are dropped. This is why environments need the revision check. |
| How are MCP sessions identified? | Each client connection has a session id from the MCP transport; a session ends when the client disconnects or sends a session delete. The stdio bridge holds one session for its lifetime. |
| What does the activity record today? | The client name (e.g. `claude-code`), tool, action, target, outcome, code, arguments and duration — not the session. Two sessions of the same client look identical. |

## Clarifications

### Session 2026-11-03 (decided by the user before specification)

- Q: Should an AI still be able to change the workspace or environment the user is looking at? → A: **Never.** No MCP action changes the window's workspace or environment; there is no AI action to drive the UI.
- Q: How far should concurrent-write safety go? → A: **Serialise every read-modify-write per workspace, make the request/alias/example revision check atomic, and add the same revision check to environments** (window editor and MCP), refusing stale saves with a clear conflict error. Folder/collection settings, workspace settings and non-environment variable tables are serialised but keep last-write-wins.
- Q: Should activity identify the session? → A: **Yes.** Record a short session tag with each AI activity row and AI-sent history entry, show it next to the client name, and make it searchable in the history filter.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - AI work never moves my window (Priority: P1)

I'm working in workspace "Acme" with environment "Dev" selected. A Claude session using the Xenon MCP switches to
"Globex", sets environment "Prod", opens or creates another workspace, and edits requests. My window stays on
"Acme"/"Dev" with my tabs and selection untouched; I only see data refresh (the workspace list gains the new
workspace, edited requests reload if I have them open and clean).

**Why this priority**: This is the reported pain — the user's own work is hijacked by AI activity, including stopped runs.

**Independent Test**: With the window on one workspace/environment, call `workspace_edit switch`, `create`, `open` and `env_edit set_active` over MCP; verify the window's active workspace and environment are unchanged and no follow happens.

**Acceptance Scenarios**:

1. **Given** the window is on workspace A, **When** an MCP session switches to workspace B, **Then** the window's active workspace is still A, no "workspaces changed" event moves the window, and A's running collection run (if any) keeps running.
2. **Given** the window is on workspace A, **When** an MCP session creates or opens workspace C, **Then** C appears in the window's workspace list but the window stays on A.
3. **Given** the window is on A with environment Dev active, **When** an MCP session (in A) sets environment Prod active, **Then** the window's environment switcher still shows Dev and the stored active environment of A is still Dev.
4. **Given** a request is open in a window tab with unsaved edits, **When** an MCP session edits the same request, **Then** the tab keeps the user's edits and shows the existing "changed on disk" notice; a clean tab reloads silently.
5. **Given** the window is on workspace A, **When** an MCP session removes workspace A, **Then** the window leaves A exactly as if the user had removed it (moves to the workspace the registry picks, or to no workspace).

---

### User Story 2 - Each AI session has its own workspace and environment (Priority: P1)

Two Claude sessions work in parallel. Session 1 works in "Acme" with "Dev"; session 2 switches to "Globex" and
"Prod". Each session's `tree`, `send`, `run`, `env`, `snippet` and every other workspace-scoped tool uses only that
session's own workspace and environment.

**Why this priority**: Without isolation, parallel sessions silently send to the wrong workspace or environment (possibly production).

**Independent Test**: Open two MCP sessions, switch them to different workspaces/environments, and verify each one's `workspace active`, `env list`, `send` and `run` reflect only its own choice.

**Acceptance Scenarios**:

1. **Given** the window is on A with environment Dev, **When** a new MCP session connects and asks `workspace active`, **Then** it gets A with Dev marked active.
2. **Given** sessions S1 and S2 both start in A, **When** S2 switches to B, **Then** S1's `workspace active` still answers A and S1's `tree` lists A's items.
3. **Given** sessions S1 and S2 both in A, **When** S2 sets environment Prod active, **Then** S1's `send` still resolves variables with Dev and S1's `env list` still marks Dev active.
4. **Given** a session in A with Dev, **When** it calls `send` with `environment: "Prod"` (or `"none"`), **Then** that one send resolves with Prod (or with no environment) and the session's environment stays Dev.
5. **Given** a session in A whose environment is Prod (the window's is Dev) and a folder whose run configuration uses the active environment, **When** the session starts a `run` without `environment`, **Then** the run uses Prod.
6. **Given** the window has no workspace open, **When** a session connects, **Then** its workspace-scoped tools answer the existing "no active workspace" error (`MCP-06404`) until it switches to or opens/creates a workspace.
7. **Given** a session switched to B, then back to A, **When** it sends, **Then** it uses the environment it chose for A earlier in the session (or A's stored active environment if it never chose one).

---

### User Story 3 - Concurrent edits never lose a write silently (Priority: P2)

Two sessions, or a session and the user, edit the same workspace at the same time. Files are never corrupted, and
a write is never silently lost: patch-style AI edits apply on top of the latest file, and a save based on a stale
copy (a window editor holding an old revision) is refused with a clear "changed since you opened it" conflict.

**Why this priority**: Parallel sessions make concurrent writes routine; lost writes are invisible data loss.

**Independent Test**: Fire concurrent MCP edits to the same request/environment from two sessions; verify both edits land. Save an environment from the window with a stale revision; verify a conflict error.

**Acceptance Scenarios**:

1. **Given** two sessions editing different fields of the same request at the same time, **When** both calls finish, **Then** the request file contains both edits.
2. **Given** two sessions adding different variables to the same environment at the same time, **When** both calls finish, **Then** the environment contains both variables.
3. **Given** the window loaded environment Dev at revision R1 and an MCP session then added a variable (revision R2), **When** the window saves Dev's variable table based on R1, **Then** the save is refused with conflict `ENV-19409` "the environment changed on disk since it was opened", nothing is written, and the editor offers to load the latest version or keep its own (a forced save).
4. **Given** the window saves a request based on revision R1 at the same moment another writer saves it, **When** both run, **Then** exactly one save wins and the other gets `REQ-02409`; no write is lost without an error.

---

### User Story 4 - I can tell which session did what (Priority: P3)

The History "AI" activity and AI-sent history rows show the client name plus a short session tag (e.g.
`claude-code · 3f9a1c`), and typing the tag in the history search finds that session's rows.

**Why this priority**: Visibility makes the isolation trustworthy, but the feature works without it.

**Independent Test**: Make calls from two sessions of the same client; verify history entries carry different session tags and the search filters by tag.

**Acceptance Scenarios**:

1. **Given** two sessions of client `claude-code`, **When** each performs an action, **Then** their history rows carry different session tags.
2. **Given** history with rows from two sessions, **When** the user searches for one session's tag, **Then** only that session's AI rows match.

---

### Edge Cases

- A session's current workspace is removed or unlinked (by the user or another session): its workspace-scoped calls answer `MCP-06404` until it switches/opens/creates; non-workspace-scoped tools keep working.
- A session's chosen environment is deleted: the session resolves with no environment (as the window does today for a deleted active environment) and `env list` marks none active.
- A session ends (disconnect or session delete): its context is discarded; a reconnect is a new session seeded from the window again.
- The user switches the window's workspace or environment after a session connected: the session is unaffected (it keeps its own snapshot).
- `env_edit set_active` with an empty/omitted id clears the session's environment for its current workspace (resolve with none); it never touches the stored active environment.
- `workspace_edit rename` renames a workspace for everyone (it is data), but changes no one's current workspace.
- An AI `create`/`open` while the window has no workspace open: the workspace is registered but the window stays on "no workspace".
- Approval prompts for protected environments ("confirm before send") still trigger on the environment the call actually resolves with (session's, or the per-call override).
- A run started by a session keeps running when that session switches workspace or ends; `run stop`/`runs` reach it from any session whose current workspace is the run's workspace (and the window shows it as today).
- Files edited outside Xenon (an editor, the git CLI) cannot be serialised; the revision checks still catch them for requests and environments (the revision is computed from the file bytes), and other files stay last-write-wins.
- The window's environment editor learns about a conflict when it saves (`ENV-19409`), not earlier: a dirty variable draft is kept as is while AI changes arrive (clean drafts refresh, as today).
- Calls from a transport without a session id (none exist today; the stdio bridge has one session) behave as one anonymous session per call: seeded from the window, discarded after the call.

## Requirements *(mandatory)*

### Functional Requirements

**Session context**

- **FR-001**: Each MCP session MUST have its own context: a current workspace (or none) and, per workspace it has used, a chosen environment (or none).
- **FR-002**: A new session's context MUST be seeded when the session connects from the window's active workspace and that workspace's stored active environment at that moment (none if the window has no workspace).
- **FR-003**: Every workspace-scoped MCP tool call MUST resolve its workspace from the calling session's context, never from the window's active workspace.
- **FR-004**: `workspace_edit switch`, `open` and `create` MUST set only the calling session's current workspace. On first entering a workspace in a session, its environment MUST be seeded from that workspace's stored active environment at that moment; re-entering MUST restore the session's earlier choice for that workspace.
- **FR-005**: `env_edit set_active` MUST set only the calling session's environment for its current workspace (empty/omitted id = none) and MUST NOT change the workspace's stored active environment.
- **FR-006**: `workspace list` MUST mark the session's current workspace as active; `workspace active` MUST return the session's workspace with the session's environment marked active; `env list` MUST mark the session's environment active.
- **FR-007**: `send`, `snippet` and `env inspect` MUST resolve variables with the session's environment; `run` without `environment` MUST use the session's environment wherever its run configuration says "the active environment" (a configuration pinned to a specific environment keeps it).
- **FR-008**: `send` MUST accept an optional `environment` argument (environment name or id, or `none`) that applies to that call only, with the same meaning and the same "unknown environment" rejection as `run`'s.
- **FR-009**: When a session ends, its context MUST be discarded.
- **FR-010**: If a session's current workspace is no longer registered, workspace-scoped calls MUST answer `MCP-06404`; if its chosen environment no longer exists, calls MUST resolve with no environment.

**Window decoupling**

- **FR-011**: MCP actions MUST NOT change the window's active workspace, the workspace's stored active environment, the window's tabs or the window's selection; no MCP action exists to change them.
- **FR-012**: MCP `create`/`open` MUST register the workspace in the workspace list without making it the window's active workspace.
- **FR-013**: MCP `switch` MUST NOT stop collection runs, start/stop file watching for the window, or alter the window's per-workspace UI state.
- **FR-014**: When AI changes the workspace list, the window MUST refresh the list without switching; the only exception is when the window's own workspace was removed or unlinked, in which case it MUST behave exactly as when the user removes it (move to the workspace the registry picks, or to none).
- **FR-015**: AI data changes (tree, requests, variables, history, cookies, workspace list) MUST keep refreshing the window live; a request tab with unsaved edits MUST keep its edits and show the existing "changed on disk" notice.

**Concurrent writes**

- **FR-016**: All read-modify-write operations on one workspace's files and local state (requests, aliases, examples, folders/collections and their settings, ordering, environments, workspace settings, active environment and current values, vault) MUST be serialised per workspace, whether they come from the window or from MCP. The serialised step covers only reading and writing local files and state, never network I/O, so a send or run waits on another writer only for that brief step.
- **FR-017**: The request/alias/example revision check MUST be performed in the same serialised step as the write it guards.
- **FR-018**: Environments MUST carry a revision (derived from the stored environment file) that the window receives when it loads an environment and sends back when saving the environment's settings or variable table; a save whose revision differs from the stored one MUST be refused with `ENV-19409` "the environment changed on disk since it was opened" and write nothing, unless the save is explicitly forced.
- **FR-019**: When the window's environment save is refused with `ENV-19409`, the environment editor MUST keep the user's draft and offer "Load from disk" (discard the draft, show the latest) and "Keep mine" (save again, forced).
- **FR-020**: MCP edits MUST apply on top of the latest stored version, so concurrent MCP edits never drop each other's changes and never fail with a conflict: single-step edits (`env_edit` add_var, set_vars, set_current, reset, persist, reorder, duplicate) run inside the serialised step; edits that read and then save (`tree_edit` request create-with-fields/update, alias and example edits, `env_edit` update) save with the revision they read and, when another write landed in between, read again and re-apply their change (up to 50 attempts; each conflict means another write succeeded).

**Visibility and documentation**

- **FR-021**: Each AI activity record and each AI-sent history entry MUST store a session tag: the first 6 characters of the MCP session id.
- **FR-022**: The window's history MUST show the session tag next to the client name on AI rows, and the history search MUST match the tag.
- **FR-023**: The MCP tool descriptions of `workspace`, `workspace_edit`, `env_edit` and `send` MUST say that the workspace/environment they show or change is this session's; each stays one sentence (the existing tool-list rule `^[A-Z][^.]*\.$`), and the whole `tools/list` response MUST stay within a budget raised from 39,000 to 39,500 characters (it is 38,849 today; the new `send` argument and wording need about 170).

### Key Entities

- **MCP Session Context**: one per live MCP session — session id, current workspace (local id or none), and a map of workspace → chosen environment (id or none). In memory only; discarded when the session ends.
- **Environment Revision**: a fingerprint of an environment's stored file; returned with an environment, sent back on window saves, compared atomically before writing. The file never holds secret values (they live in the vault), so the fingerprint reveals none.
- **Session Tag**: the first 6 characters of a session id, stored with AI activity and AI-sent history entries. It is a display label only, never used to address a session; the MCP endpoint is loopback-only, so the partial id grants nothing.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In 100% of tested MCP actions (switch, open, create, set_active, edits, sends, runs) the window's active workspace and environment are unchanged afterwards.
- **SC-002**: With two parallel sessions in different workspaces/environments, 100% of each session's calls act on its own workspace and environment.
- **SC-003**: Under concurrent edits from two sessions to the same request or environment, 0 edits are lost: every edit is either present in the final file or its caller got a conflict error.
- **SC-004**: A window environment save based on a stale copy is refused 100% of the time, with nothing written.
- **SC-005**: Every AI activity and AI-sent history row shows which session produced it, and a search by tag returns only that session's rows.

## Assumptions

- Session context lives in memory only; it does not survive a Xenon restart (sessions do not either).
- Current values (`set_current`, scripts setting variables) remain shared workspace-local data, like any other data change; only the choice of *which* environment is active is per-session.
- Folder/collection settings, workspace settings and collection/workspace/global variable tables are serialised (no corruption, MCP patch edits merge) but keep last-write-wins between the window's full saves and other writers (user decision); requests, aliases, examples and environments are revision-checked.
- The window's own behaviour (its sends, runs, environment switcher) is unchanged: it keeps using the stored active workspace and environment.
- Out of scope: an MCP action to drive the window (decided: never); per-session current values; revision checks for folder/collection/workspace settings.
