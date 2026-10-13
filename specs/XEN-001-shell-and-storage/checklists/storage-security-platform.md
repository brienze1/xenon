# Requirements Quality Checklist: Storage, Security & Cross-Platform

**Purpose**: Unit-test the stage-1 requirements for data integrity, secret/credential safety and
cross-platform behaviour before `/create-tests`.
**Created**: 2026-10-09
**Feature**: [spec.md](../spec.md) · [plan.md](../plan.md) · [format](../contracts/workspace-format.md)
**Audience / depth**: implementer & reviewer · thorough

## Requirement Completeness — storage

- [x] CHK001 Is the complete folder layout, including how collections/folders are recognised, specified? [Spec §FR-009] — format §1 (marker files).
- [x] CHK002 Is every stored field of workspace, collection, folder and request files defined with type and default? [Spec §FR-012] — format §3–§6, defaults written explicitly (§7).
- [x] CHK003 Are reserved slots for later stages (variables, scripts, inheritScripts, all auth types, postman passthrough, secret rows) defined so no migration is needed? [Spec §FR-012] — format §2–§6, §10.
- [x] CHK004 Are atomicity requirements defined for every write, including app-data files? [Spec §FR-013] — format §7, data-model "Local data".
- [x] CHK005 Is preservation of unknown fields (and their order) specified? [Spec §FR-013] — format §7.
- [x] CHK006 Is how order is stored, and how unlisted/missing entries are reconciled, specified? [Spec §FR-010] — format §7 "Ordering".
- [x] CHK007 Are requirements defined for where local-only state lives and that it never enters the workspace folder? [Spec §FR-005, §FR-038c] — data-model "Local data"; `.gitignore` block format §9.
- [x] CHK008 Is the stable identity of items (independent of file paths) specified for tabs/history/UI state? [Gap→resolved] — format §2 `id` (ULID), data-model Tab/HistoryEntry.

## Requirement Clarity — storage

- [x] CHK009 Are file-name derivation rules precise enough to produce one deterministic name per input? [Spec §FR-011] — format §8 steps 1–6.
- [x] CHK010 Is "safe on macOS, Windows and Linux" defined (forbidden chars, reserved names, length, case)? [Spec §FR-011] — format §8.
- [x] CHK011 Is the external-change conflict behaviour unambiguous for clean, dirty and deleted cases? [Spec §FR-017, Clarifications] — FR-017 + contracts/bindings `REQ-02409`/`force`.
- [x] CHK012 Is the body-file path storage rule (relative vs absolute, separator) unambiguous? [Spec §FR-028a] — format §6.3.
- [x] CHK013 Is "large response" quantified? [Spec §FR-035] — 5 MB; bridge limit 1 MB (research R4).

## Requirement Consistency

- [x] CHK014 Are the spec's "remembered per workspace" state items consistent with the WorkspaceUIState fields? [Spec §FR-005, §FR-040] — data-model WorkspaceUIState covers tabs, drafts, sidebar, view, layout.
- [x] CHK015 Is FR-038 (responses survive switches) consistent with the per-workspace store reset? [Conflict→resolved] — `LastResult` binding; plan decision 8; tasks T053/T093.
- [x] CHK016 Are header-inheritance rules consistent between spec, data model and tests? [Spec §FR-014] — data-model "Effective request"; T067.
- [x] CHK017 Is History scope consistent across spec (US5), constitution VI and PRD? [Consistency] — user decision recorded in Clarifications and PRD; plan gate VI.

## Scenario & Edge-Case Coverage — storage

- [x] CHK018 Are requirements defined for malformed files, unknown versions and duplicate ids? [Spec Edge Cases, §FR-016] — format §2, §10; T066.
- [x] CHK019 Are crash-during-save and unwritable-folder outcomes specified? [Spec Edge Cases, SC-007] — Edge Cases; T104.
- [x] CHK020 Are overlapping/nested workspace folders addressed? [Spec Edge Cases] — data-model AppRegistry validation; `WKS-02409`.
- [x] CHK021 Are missing/moved workspace folders addressed? [Spec US4 #5] — `WKS-06404`, Locate.
- [x] CHK022 Are symlinks inside a workspace addressed? [Gap→resolved] — format §1, Edge Cases.
- [x] CHK023 Are CRLF/BOM files (git autocrlf) addressed? [Gap→resolved] — format §7.
- [x] CHK024 Are case-only renames and Windows long paths addressed? [Gap→resolved] — format §8.
- [x] CHK025 Is trash failure during workspace removal specified? [Gap→resolved] — Edge Cases; `WKS-05500`.

## Workspace isolation

- [x] CHK026 Is "hard boundary" defined in terms of every state type (tree, tabs, search, history, layout, events, watcher)? [Spec §FR-004] — FR-004, FR-038c; T089 scopes.
- [x] CHK027 Is isolation measurable? [Spec SC-006] — 100 alternating switches, T089.
- [x] CHK028 Is the identity used for local state separated from the shareable workspace `id` (two clones of one repo)? [Gap→resolved] — data-model RegisteredWorkspace.LocalID.

## Security & privacy

- [x] CHK029 Are requirements explicit that history and drafts store values as typed and never resolved values? [Spec §FR-038a] — FR-038a, Assumptions.
- [x] CHK030 Is leakage of headers/auth/bodies into logs prevented by a stated requirement? [Constitution II] — plan decision 7; T055 redaction + test.
- [x] CHK031 Are permissions for local files holding typed tokens specified? [Gap→resolved] — data-model "Local data" (0700/0600).
- [x] CHK032 Is the handling of temp files containing response bodies specified, including after crashes? [Gap→resolved] — data-model Response (owner-only, cleanup at startup).
- [x] CHK033 Is the stage-1 storage of typed auth values in request files an explicit, documented decision? [Assumption] — Spec Assumptions (Postman-equivalent; stage 2 vault).
- [x] CHK034 Are TLS verification defaults and the way to bypass them specified? [Spec §FR-029, §FR-037] — default on, per-request toggle, hint in error.
- [x] CHK035 Is it specified that a workspace cannot read/write outside its folder? [Gap→resolved] — symlinks ignored; body files are read only at send time by explicit user choice.
- [x] CHK036 Is the large-body route's exposure scoped (same-origin, in-app only)? [Contracts §6] — bindings.md §6.

## Cross-platform

- [x] CHK037 Are shortcut mappings defined per OS? [Spec §FR-042] — Cmd on macOS, Ctrl elsewhere.
- [x] CHK038 Are app-data locations defined per OS? [Gap→resolved] — data-model "Local data" root.
- [x] CHK039 Are trash semantics defined per OS, including macOS permission concerns? [Spec §FR-003] — research R8.
- [x] CHK040 Are Linux runtime/build prerequisites (GTK4/WebKitGTK 6.0, supported distros) documented? [Dependency] — plan Target Platform, research R13, quickstart.
- [x] CHK041 Is "same behaviour on all OSes" measurable? [Spec SC-008] — CI matrix runs the full suite on 3 OSes (T012, T106).
- [x] CHK042 Are file-watching limitations per OS documented? [Assumption] — research R7 (kqueue fds), Spec Assumptions (network drives).

## Acceptance criteria quality

- [x] CHK043 Are performance targets quantified and tied to tasks? [Spec SC-002, SC-003] — T103.
- [x] CHK044 Is "zero data loss" measurable? [Spec SC-007] — 100 kill-during-write runs, T104.
- [x] CHK045 Is restore completeness measurable? [Spec SC-004] — 100 % of listed state, T081/T107.

## Notes

- All items pass after the artifact fixes made in this run (marked "Gap→resolved" / "Conflict→resolved").
