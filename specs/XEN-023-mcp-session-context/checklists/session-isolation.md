# Session Isolation, Concurrency & Secrets Checklist: Per-Session MCP Context

**Purpose**: Unit tests for the requirements — concurrency and data integrity, MCP session isolation, window decoupling, and secret safety
**Created**: 2026-11-03
**Feature**: [spec.md](../spec.md) · [plan.md](../plan.md) · [research.md](../research.md)
**Depth / audience**: standard, PR reviewer

## Requirement Completeness — concurrency & data integrity

- [x] CHK001 - Is the set of operations that must be serialised per workspace enumerated? [Completeness, Spec §FR-016] — FR-016 lists the categories; research R7 and T041 list every service plus `vaultWriter.update`.
- [x] CHK002 - Are requirements defined for writes made outside Xenon, which cannot be serialised? [Gap → fixed, Spec §Edge Cases] — added: revision checks still catch external edits to requests/environments, and other files stay last-write-wins.
- [x] CHK003 - Is the conflict outcome fully specified: error code, message, and that nothing is written? [Completeness, Spec §FR-018] — `ENV-19409`, "the environment changed on disk since it was opened", nothing written.
- [x] CHK004 - Are the meaning and scope of `force` defined, including who uses it? [Clarity, Spec §FR-018, §FR-019, contracts/wails-bindings-and-events.md] — only the window's "Keep mine"; MCP patches the latest file and never needs it (FR-020).
- [x] CHK005 - Is the lost-update guarantee for files without revisions stated explicitly? [Completeness, Spec §Assumptions] — folder/collection/workspace settings and non-environment tables are serialised but last-write-wins (user decision).
- [x] CHK006 - Is the window's state after each conflict choice defined? [Completeness, Spec §FR-019] — Load from disk discards the draft and shows the latest; Keep mine saves forced and returns the new revision (contracts).
- [x] CHK007 - Is the request revision race fix stated in a testable way? [Measurability, Spec §FR-017, US3 scenario 4] — exactly one of two simultaneous saves wins; the other gets `REQ-02409`.
- [x] CHK008 - Is it stated that the lock never covers network I/O, so sends and runs are not blocked? [Gap → fixed, Spec §FR-016] — added to FR-016.
- [x] CHK009 - Is deadlock avoidance for nested operations addressed? [Completeness, research R7] — one reentrant lock per workspace, carried in the context.
- [x] CHK010 - Is the derivation of the environment revision defined so that every reader computes the same value? [Clarity, data-model.md §Environment] — SHA-256 of the stored file bytes, the same scheme as requests.
- [x] CHK011 - Are the conflict rules consistent between environment settings saves and variable-table saves? [Consistency, Spec §FR-018] — both use the same revision and code.
- [x] CHK012 - Is it specified which window variable operations are revision-checked and which are not? [Clarity, contracts/wails-bindings-and-events.md] — table save and settings update are checked; single-key set_current/reset/persist/add are patches under the lock.

## Requirement Completeness — session isolation

- [x] CHK013 - Is the exact moment a session's context is seeded defined? [Clarity, Spec §FR-002, research R2] — on the MCP `initialized` handshake, with a fallback for calls that arrive before it.
- [x] CHK014 - Is cleanup defined for both normal and abnormal session end? [Completeness, Spec §FR-009] — disconnect or session delete; the listener's close watcher covers dropped connections.
- [x] CHK015 - Is behaviour defined for calls without a session id? [Edge Case, Spec §Edge Cases] — treated as one anonymous session per call and never stored.
- [x] CHK016 - Is cross-session leakage of stored responses addressed? [Gap → fixed, Spec §Context] — added: the `response` tool reads only the calling session's key.
- [x] CHK017 - Are pending approvals scoped per session? [Consistency, Spec §Context] — added: approvals carry the session id, and closing a session denies only its own.
- [x] CHK018 - Is behaviour defined when a session's workspace is removed or its environment deleted? [Edge Case, Spec §FR-010] — `MCP-06404`, or resolve with no environment.
- [x] CHK019 - Is it defined whether a session follows later changes in the window? [Clarity, Spec §Edge Cases] — no; the session keeps its snapshot.
- [x] CHK020 - Is the lifetime of stored session context bounded? [Completeness, Spec §Key Entities] — in memory, one entry per live session, deleted on close.
- [x] CHK021 - Is re-entering a workspace within a session specified? [Clarity, Spec §FR-004, US2 scenario 7] — the earlier choice is restored.
- [x] CHK022 - Are runs started by a session defined across switch and close? [Gap → fixed, Spec §Edge Cases, §Context] — added: runs belong to their workspace and keep running; stop/list from that workspace.
- [x] CHK023 - Do all workspace-scoped tools follow the session (history, cookies, console, git, docs, scripts, delete)? [Coverage, Spec §FR-003] — FR-003 covers every workspace-scoped tool; `mcp_call.go` is the single resolution point (plan).

## Requirement Completeness — window decoupling

- [x] CHK024 - Does the requirement enumerate every piece of window state that MCP must not change? [Completeness, Spec §FR-011, §FR-013] — workspace, stored active environment, tabs, selection, runs, file watching and per-workspace UI state.
- [x] CHK025 - Is the one allowed exception, the window's workspace being removed, defined precisely? [Clarity, Spec §FR-014] — the same behaviour as a removal by the user.
- [x] CHK026 - Is AI create/open defined for when the window has no workspace open? [Edge Case, Spec §Edge Cases] — the workspace is registered and the window stays on none.
- [x] CHK027 - Is the window's own environment behaviour stated as unchanged? [Consistency, Spec §Assumptions]
- [x] CHK028 - Is it stated when the window learns about a conflicting AI change to an environment it is editing? [Gap → fixed, Spec §Edge Cases] — added: at save time; dirty drafts are kept until then.
- [x] CHK029 - Are live data refreshes required for every kind of AI change? [Coverage, Spec §FR-015] — tree, requests, variables, history, cookies and the workspace list.

## Non-Functional — security & secrets

- [x] CHK030 - Is it stated that the environment revision cannot reveal secret values? [Gap → fixed, Spec §Key Entities] — the file never holds secrets, so neither does the hash.
- [x] CHK031 - Is the session tag's sensitivity addressed? [Gap → fixed, Spec §Key Entities] — it is a display label only and the endpoint is loopback-only.
- [x] CHK032 - Is the session context restricted to non-secret data? [Completeness, data-model.md §AISessionContext] — ids only.
- [x] CHK033 - Are approval prompts for protected environments defined for the session environment and for per-call overrides? [Coverage, Spec §Edge Cases] — the prompt uses the environment the call actually resolves with.
- [x] CHK034 - Is it stated that new log and DTO fields carry no secret values? [Completeness, plan §Constitution Check II] — environment ids, revisions and session tags only.

## Acceptance Criteria & Dependencies

- [x] CHK035 - Are the success criteria measurable without knowing the implementation? [Measurability, Spec §SC-001–SC-005] — 100% / 0-lost thresholds.
- [x] CHK036 - Does every functional requirement trace to a named BDD or Vitest test? [Traceability, plan §Tests] — FR-001–FR-023 each map to a test row.
- [x] CHK037 - Is the `tools/list` size budget dependency documented with a measured baseline? [Dependency, Spec §FR-023] — 38,849 measured, budget 39,500.
- [x] CHK038 - Is backward compatibility of stored data addressed? [Assumption, data-model.md] — history `session` is `omitempty` so old entries read fine; the revision is never written to YAML, so files are unchanged.
- [x] CHK039 - Are the user's three decisions (never drive the window, lock everything plus environment revisions, session tag) carried consistently into spec, plan and tasks? [Consistency, Spec §Clarifications] — FR-011, FR-016–FR-020, FR-021–FR-022, and Phases 3, 5 and 6.
