# State Consistency & Concurrency Checklist: Runner Follows the Current Tree

**Purpose**: Unit tests for the requirements — refresh triggers, coalescing and races, preservation of user edits, failure/empty states, workspace isolation, run correctness, and adequacy of the regression/guard tests
**Created**: 2026-10-26
**Feature**: [spec.md](../spec.md) · [plan.md](../plan.md) · [tasks.md](../tasks.md)
**Audience / depth**: PR reviewer, standard

## Requirement Completeness

- [x] CHK001 Are all sources of a tree change enumerated, including Xenon's own sidebar edits that emit no backend tree-changed event? [Completeness, Spec §Assumptions] — Assumptions list sidebar edits, AI edits, git pulls, imports, external edits; research R2 verifies own writes emit no event (`handle_file_change.go` `IsOwnWrite`), which is why the trigger is the sidebar tree.
- [x] CHK002 Is the set of Runner targets that refresh defined (read this session vs never shown)? [Completeness, Spec §FR-001, §Clarifications] — FR-001 + clarification 3; contract: `status !== 'idle'`.
- [x] CHK003 Are the state fields a refresh replaces, and those it must never touch, specified? [Completeness, Spec §FR-005, data-model.md] — data-model lists both sets.
- [x] CHK004 Are requirements defined for a Runner edit made while a re-read is in flight, not only before it starts? [Completeness, Spec §FR-003] — fixed: FR-003 now covers in-flight edits (the re-read is discarded and repeated after the edit is remembered).
- [x] CHK005 Is the behaviour defined when a refresh's re-read fails, including a deleted target? [Exception Flow, Spec §FR-010, §Edge Cases]
- [x] CHK006 Is the behaviour defined for a run active on the target during a refresh? [Completeness, Spec §FR-005, §Edge Cases]
- [x] CHK007 Is the behaviour defined when the start confirmation ("Confirm before send") is open while the tree changes? [Gap → resolved, Spec §Edge Cases] — fixed: added edge case; confirming runs the confirmed configuration, Setup refreshes for the next run (data-model keeps `pendingConfig` untouched).
- [x] CHK008 Is the empty-target state (every request removed) defined? [Gap → resolved, Spec §Edge Cases] — fixed: empty Setup list; start refused with the existing `RUN-03400` message (`RunConfig.Validate`, `EMPTY_SELECTION` in `platform/services/runner.ts`).

## Requirement Clarity

- [x] CHK009 Is "coalesce" quantified? [Clarity, Spec §FR-004] — one re-read in flight per target plus one trailing.
- [x] CHK010 Is the 1-second target measured from a defined reference point? [Measurability, Spec §SC-002] — "within 1 second of the sidebar showing them".
- [x] CHK011 Is the insertion position of a new request unambiguous for requests inside a brand-new subfolder? [Clarity, Spec §FR-002] — fixed: "right after the nearest request preceding it in sidebar order (depth first, subfolders included)", matching `ReconcileRunRows`'s `insertionPoint(rows, sidebar[:index])`; wording aligned across spec, plan, research, tasks, quickstart.
- [x] CHK012 Are "remembered rows" and "Runner rows (shown)" defined distinctly? [Clarity, Spec §Key Entities]

## Requirement Consistency

- [x] CHK013 Do FR-006 (run sends rows shown) and the edge case "an active run keeps the rows it started with" agree? [Consistency, Spec §FR-006, §Edge Cases] — both scope to the rows shown when Run is pressed.
- [x] CHK014 Is the spec's account of headless behaviour consistent with the code findings? [Consistency, Spec §Context, research R1] — aligned during analyze (I1).
- [x] CHK015 Are failure states consistent between a first read and a refresh? [Consistency, Spec §FR-010] — "exactly as a failed first read does".
- [x] CHK016 Is the deleted-request outcome stated consistently in spec and research? [Consistency, research R1, Spec §Edge Cases] — fixed: research R1 corrected from "skipped silently" to the real behaviour (failed "missing request" row, `run_row.go`), and the spec's race edge case states the same.

## Scenario Coverage (concurrency, isolation, run correctness)

- [x] CHK017 Is the burst scenario (an AI client creating many requests one by one) covered by requirements and acceptance? [Coverage, Spec §FR-004, §Edge Cases]
- [x] CHK018 Are workspace-isolation requirements defined for a workspace switch while a re-read is in flight? [Coverage, Spec §FR-008] — fixed: FR-008 now requires discarding a re-read that completes after the switch; research R3 shows the store reset precedes the new tree.
- [x] CHK019 Is "rows sent match rows shown" stated as a requirement? [Coverage, Spec §FR-006, Story 2]
- [x] CHK020 Is the race "Run pressed between a tree change and the refresh landing" addressed? [Coverage, Spec §Edge Cases] — fixed: added; the run sends the rows shown, a deleted request appears as a failed "missing request" row (existing behaviour), the next run is current.
- [x] CHK021 Are out-of-order responses (a slow first read finishing after a refresh) addressed? [Coverage, data-model.md, tasks T002/T004(i)] — the read-sequence rule applies only the latest read.

## Regression / Guard Test Adequacy

- [x] CHK022 Does every FR name the test that locks it? [Traceability, plan.md §Requirement → layer → test map] — FR-009 via the existing `runner_plan.feature:203` scenario (tasks Notes).
- [x] CHK023 Is it documented which tests must be red first and which are green guards, with justification? [Traceability, plan.md §Complexity Tracking, research R7]
- [x] CHK024 Do the guard scenarios specify concrete assertions (positions, `selected`, `planned`, request order)? [Measurability, tasks T006, T012]
- [x] CHK025 Is the UI-originated trigger (no backend event) covered by a test requirement? [Coverage, tasks T005]
- [x] CHK026 Is the plan for a guard that unexpectedly goes red defined? [Exception Flow, tasks §Tests] — fix the backend layer the failing step points to before continuing.

## Non-Functional, Dependencies & Assumptions

- [x] CHK027 Is the cost of re-reading every known target bounded and justified? [Non-Functional, Spec §Assumptions, plan §Technical Context] — targets are bounded by tabs opened in a session; ~100-row folders.
- [x] CHK028 Does the feature avoid exposing secret values? [Security, plan §Constitution Check II] — rows carry ids, names, methods, alias ids only.
- [x] CHK029 Is the dependency on backend reconciliation documented and verified against the code? [Dependency, research R1] — `run_planner.go` `plan()` → `treePersistence.Load` (index `Refresh`, full rescan) → `ReconcileRunRows`; `StartForAI` uses the same `plan()`.
- [x] CHK030 Is the assumption "the store resets before the new workspace's tree arrives" validated? [Assumption, research R3] — `useWorkspacesStore.activate` sets `activeLocalId` before `setTree`; sidebar `reset()` leaves `status: 'idle'`.
