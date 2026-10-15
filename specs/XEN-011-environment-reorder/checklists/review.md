# Requirements Review Checklist: Reorder Environments (XEN-011)

**Purpose**: Unit tests for the requirements — data integrity of `workspace.yaml` → `environments.order` and UX/accessibility of pointer and keyboard reordering. For the PR reviewer.
**Created**: 2026-10-15
**Feature**: [spec.md](../spec.md) · [plan.md](../plan.md) · [data-model.md](../data-model.md) · [contracts/](../contracts/)
**Depth**: standard

Every item was run against the artifacts. Items that failed on the first pass were fixed in the
artifacts named in the item, then ticked.

## Requirement Completeness — file format

- [x] CHK001 - Is the on-disk location, key path and value type of the order specified? [Completeness, Spec §FR-006] — `workspace.yaml` → `environments.order`, list of environment ids; shape example in research R2 and data-model.md.
- [x] CHK002 - Is the identity used in the order (id vs name vs file name) specified, with the reason? [Clarity, Spec §Assumptions] — ids, so renames never break the order (research R1).
- [x] CHK003 - Are requirements defined for every other writer of `workspace.yaml` preserving the order? [Completeness, Spec §FR-019] — failed first pass (only plan/tasks covered it); added FR-019 and extended T025 to collection reorder, globals edit and workspace rename.
- [x] CHK004 - Is preservation of unknown keys under `environments` required? [Completeness, Spec §FR-019, data-model.md] — FR-019 "any other keys under `environments`"; `yamlnode.Object` patches the existing mapping (research R4).
- [x] CHK005 - Is the "never written by listing" rule specified and measurable? [Measurability, Spec §FR-018, §SC-004] — byte-for-byte unchanged after open/list; locked in T010.
- [x] CHK006 - Is it specified exactly which operations write the order and which never do? [Completeness, Spec §FR-018, research R6] — reorder/create/duplicate/import write; delete only when listed; rename/update/list never.

## Requirement Clarity — arrangement and placement

- [x] CHK007 - Is the full arrangement rule specified, including the tie-breaker and where broken files go? [Clarity, Spec §FR-007] — stored order → readable alphabetical by name (then file name, research R5) → broken by file name.
- [x] CHK008 - Is "goes last" defined for a workspace that has no stored order yet? [Clarity, Spec §FR-009, Clarifications] — user decision: the first add materializes the arranged list + new id (US4 scenario 2).
- [x] CHK009 - Is Replace-import placement specified unambiguously? [Clarity, Spec §FR-010] — takes the replaced environment's position; order computed before the replaced file is deleted (plan note 4).
- [x] CHK010 - Are index semantics (0-based, position after the move, past-the-end, negative) specified for both entry points? [Clarity, contracts/environment-service.md, contracts/mcp-env-edit.md] — 0 = first; ≥ count → last; negative → `ENV-17400` / `ENV-18400`.
- [x] CHK011 - Is the batch-import order specified (several environments in one import)? [Clarity, Spec §US4 scenario 4] — last, in import order; plan note 4 re-lists per item because targets are resolved once per batch.

## Requirement Consistency

- [x] CHK012 - Do spec, research and plan agree on the delete rule? [Consistency, Spec §FR-012/§FR-018, research R6, plan note 3] — aligned in analyze pass 1 (I1): `IDs()` minus the id, written only when the stored order names it.
- [x] CHK013 - Do spec and UI contract agree on which rows are draggable while creating or renaming? [Consistency, Spec §Edge Cases, contracts/environments-view-ui.md] — failed first pass (spec said only the input row; contract disabled all rows while creating); spec edge case now says no row can be dragged while the input row shows.
- [x] CHK014 - Is the field name consistent across spec, data model, plan, contracts and tasks? [Consistency] — `environments.order` everywhere (user decision, Clarifications).
- [x] CHK015 - Is the switcher order consistent with the sidebar, including "No environment"? [Consistency, Spec §FR-008, §US1 scenario 3] — same order; "No environment" stays first.

## Scenario & Edge Case Coverage — data integrity

- [x] CHK016 - Are malformed values (non-mapping `environments`, non-list `order`, non-scalar items) addressed without breaking the workspace? [Edge Case, Spec §Edge Cases, research R3] — treated as no order; raw-node decoding avoids `WKS-16422`.
- [x] CHK017 - Are duplicate ids and stale ids addressed? [Edge Case, Spec §FR-007, §Edge Cases] — first occurrence wins; stale ignored and dropped on the next write.
- [x] CHK018 - Is an environment that is broken at the moment the order is written addressed? [Edge Case, Spec §Edge Cases] — failed first pass; added: it drops out of the order and returns in the alphabetical tail once fixed (its id cannot be read, so no other outcome exists).
- [x] CHK019 - Is the git merge behaviour specified, including concurrent reorders and add-vs-delete? [Coverage, Spec §FR-014, §SC-003, research R7] — same rule as `order` (`mergeOrder`), no conflict; T027 covers both cases.
- [x] CHK020 - Are concurrent writers within one machine addressed (UI and AI reorder at once)? [Coverage, Spec §Edge Cases] — failed first pass; added: each re-reads before writing, last write wins, window refreshes via `variables:changed`.
- [x] CHK021 - Is compatibility with older Xenon builds addressed? [Assumption, Spec §Assumptions] — failed first pass; added: older builds read and keep the key (verified `yamlnode.Patch` in `workspacePersistence.Save` and `treePersistence.writeFields`); they don't maintain it, and the next reorder tidies it.
- [x] CHK022 - Are export and whole-workspace import behaviours specified? [Coverage, Spec §Assumptions] — Xenon workspace export/import copy `workspace.yaml` (order kept); Postman exports unaffected.
- [x] CHK023 - Is atomicity of a failed write specified? [Exception Flow, Spec §FR-015] — "no partial order may be written"; writes use `fileutils.WriteFileAtomic`; read-only scenario in T010.
- [x] CHK024 - Are external on-disk changes required to refresh the UI? [Coverage, Spec §FR-013, research R8] — existing watcher → `variables:changed` → re-list; T026.

## UX & Accessibility Requirements

- [x] CHK025 - Are keyboard controls fully specified (pick up, move, drop, cancel, and what Enter/Space do outside a drag)? [Completeness, Spec §FR-002, contracts/environments-view-ui.md] — Space picks up, arrows move, Space/Enter drop, Escape cancels; Enter still opens, Space never opens.
- [x] CHK026 - Are screen-reader instructions and every announcement text specified, with positions? [Completeness, Spec §US3 scenario 3, contracts/environments-view-ui.md] — instructions and the four announcements with "{n} of {total}".
- [x] CHK027 - Is focus placement after a keyboard drop or cancel specified? [Completeness, contracts/environments-view-ui.md] — failed first pass; added: focus stays on the moved row.
- [x] CHK028 - Are all disabled states enumerated (Globals, broken rows, filtering, renaming, creating)? [Coverage, Spec §FR-003–FR-005, §Edge Cases] — listed in spec and contract.
- [x] CHK029 - Is the UI while filtering specified (hint or none)? [Clarity, Spec §Edge Cases] — failed first pass; added: no hint, no grab cursor, rows don't move; list and search box look as today.
- [x] CHK030 - Is drag feedback specified (how the landing position is shown, dragged-row style, cursor)? [Clarity, contracts/environments-view-ui.md] — failed first pass; added: rows slide to show the landing spot, `dragging` style, grab cursor on draggable rows only.
- [x] CHK031 - Is reduced motion addressed? [Non-Functional, contracts/environments-view-ui.md] — the global `prefers-reduced-motion` rule in `shared/styles/global.css` (transition-duration `!important`) covers the sortable transitions; contract references it.
- [x] CHK032 - Is a click still distinguishable from a drag? [Clarity, contracts/environments-view-ui.md] — pointer activation distance 5 px, as the tab strip and tree.
- [x] CHK033 - Are error-recovery requirements specified for a failed save (on-screen state and message)? [Recovery, Spec §FR-015, §US1 scenario 7] — list rolls back; banner with next step (`ACTION_NEXT_STEP`).
- [x] CHK034 - Is "no perceptible delay" made measurable? [Measurability, Spec §SC-002] — defined as "the list updates on drop, before the save completes" (optimistic update, contract store step 2).
- [x] CHK035 - Do reorders leave the active environment, open tabs and environment files alone, and is that testable? [Coverage, Spec §FR-016] — BDD (T010) for active id and files; Vitest `bootstrap.test.ts` for tabs (T011).

## Dependencies, Assumptions & Scope

- [x] CHK036 - Is the constitution's PRD-scope rule (Principle VI) addressed? [Assumption, Spec §Assumptions, research R13] — the PRD update is task T047.
- [x] CHK037 - Is AI-client parity specified (same validation, result, side effects)? [Consistency, Spec §FR-017, contracts/mcp-env-edit.md] — same service call; list result; `variables:changed`.
- [x] CHK038 - Are secrets excluded from everything this feature writes or announces? [Non-Functional, plan Constitution Check II] — ids only on disk; announcements use names already on screen.
- [x] CHK039 - Is out-of-scope work stated? [Scope, Spec §Assumptions] — moving between workspaces; Postman exports carry no order.
