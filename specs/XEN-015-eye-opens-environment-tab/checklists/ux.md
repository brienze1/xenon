# UX Requirements Quality Checklist: The Eye Opens the Environment Tab

**Purpose**: Unit tests for the requirements of the eye's interaction (open-or-focus, labels and
accessibility, no-environment state, unsaved-edits safety) and for the scope of the dead-code
removal and document updates.
**Created**: 2026-10-15
**Feature**: [spec.md](../spec.md) · [plan.md](../plan.md) · [tasks.md](../tasks.md)
**Depth / audience**: standard; PR reviewer and implementer.

## Requirement Completeness

- [x] CHK001 Are open-versus-focus outcomes specified for every prior tab state (no tab, preview tab, pinned tab) for the environment target? [Completeness, Spec §FR-001, §FR-002, US1 scenarios 1–3]
- [x] CHK002 Are the same open-versus-focus outcomes specified for the Globals target? [Completeness, Spec §FR-003, US2 scenarios 1–2]
- [x] CHK003 Is the tab mode (pinned) stated for both targets, not just one? [Completeness, Spec §FR-001, §FR-003 — both say pinned]
- [x] CHK004 Is the no-environment state fully defined (label, target, outcome)? [Completeness, Spec §FR-003, §FR-006, US2]
- [x] CHK005 Is the scope of code removed enumerated, including what "only it used" means? [Completeness, Spec §FR-008; plan research R5 lists each item with its other users]
- [x] CHK006 Are the document sections that must change named, and is it bounded which documents are not edited? [Completeness, Spec §FR-010 — PRD switcher section + environments decision, XEN-002 `spec.md` only; XEN-002 plan/tasks/checklists explicitly stay (fixed during this run)]

## Requirement Clarity

- [x] CHK007 Are the label strings given verbatim for both states? [Clarity, Spec §FR-006 — "Open environment" / "Open globals"]
- [x] CHK008 Is it specified when the label updates relative to a change of active environment? [Clarity, Spec §FR-006 — "as soon as the active environment changes"; US2 scenario 3]
- [x] CHK009 Is "focus an existing tab" defined by reference to existing behaviour rather than left open? [Clarity, Spec Edge Cases — "same focus behaviour as every other way of opening a tab"]
- [x] CHK010 Is "keep its current look and position" concrete enough to verify? [Clarity, Spec §FR-009; plan R3 — same glyph and `.icon` CSS moved verbatim, rendered right after the switcher dropdown]

## Requirement Consistency

- [x] CHK011 Are FR-003 (no env → Globals) and FR-004 (env active → never Globals) mutually exclusive and jointly exhaustive over the active-environment state? [Consistency, Spec §FR-003, §FR-004 — keyed on active / none]
- [x] CHK012 Is the eye's Globals behaviour consistent with the switcher's existing "Globals" menu item (both pinned, same tab)? [Consistency, Spec Clarifications "Pinned, like the switcher's Globals item"; contract — both use `onOpenGlobals`]
- [x] CHK013 Do the user stories, FRs and success criteria use the same terms (eye, environment tab, Globals tab, pinned, focused)? [Consistency — no drift between spec, plan, contract and tasks]
- [x] CHK014 Is removing the variable table's own save footer consistent with every remaining place that shows a variables table having another way to save? [Consistency, Spec §FR-008 — environment, Globals and collection-variables tabs save via the tab (Cmd+S, close prompt); research R5 confirms all three already hide the footer]

## Acceptance Criteria Quality

- [x] CHK015 Is "no dialog" objectively verifiable? [Measurability, Spec §FR-005, SC-003 — no element with the dialog role; source search finds no Quick Look]
- [x] CHK016 Is "never a second tab" measurable? [Measurability, Spec SC-002 — at most one tab per target after any number of clicks]
- [x] CHK017 Is automated-test coverage stated for every behavioural FR of the eye, including edit safety? [Acceptance Criteria, Spec §FR-011, SC-004 — FR-001 to FR-007 (FR-007 added after analyze)]
- [x] CHK018 Is SC-001's "no window-height overflow of its own" verifiable? [Measurability, Spec SC-001 — the tab lives in the main area with no dialog; quickstart step 2 checks it with a ~30-variable environment]

## Scenario & Edge Case Coverage

- [x] CHK019 Are requirements defined for unsaved edits in an already-open tab? [Coverage, Spec §FR-007, US1 scenario 4]
- [x] CHK020 Is the case "active environment deleted" addressed? [Edge Case, Spec Edge Cases — becomes "No environment", eye opens Globals]
- [x] CHK021 Is the case "active environment's file is broken" addressed? [Edge Case, Spec Edge Cases + Clarifications — the tab shows the existing error banner]
- [x] CHK022 Is switching environments while a tab is open addressed (no auto-retarget)? [Edge Case, Spec Edge Cases — "Active environment changes"]
- [x] CHK023 Is the "no workspace open" state addressed? [Edge Case, Spec Clarifications — the switcher, and so the eye, is only shown with an open workspace]
- [x] CHK024 Are repeated rapid clicks addressed? [Edge Case, Spec SC-002]

## Non-Functional Requirements

- [x] CHK025 Are keyboard and screen-reader requirements specified for the eye? [Accessibility, Spec Edge Cases "Keyboard users" + §FR-006 accessible name = tooltip]
- [x] CHK026 Is it stated that secret masking in the opened tabs is unaffected? [Security, Spec Assumptions — tabs unchanged; plan Constitution II]
- [x] CHK027 Are performance requirements intentionally absent? [Non-Functional — not applicable: one click calls an existing synchronous store action (plan Technical Context "Performance Goals: N/A")]

## Dependencies & Assumptions

- [x] CHK028 Are the existing open paths that must keep working listed? [Assumption, Spec Assumptions — sidebar single/double click, switcher "Globals", "Manage Environments…", reopen closed tab]
- [x] CHK029 Is the deviation from Postman's quick-look popover recorded with its authority? [Dependency, Spec Context + Clarifications (user decision, issue #18); plan Constitution VI; PRD updated per §FR-010]
- [x] CHK030 Is the frontend-only / no-BDD scope justified rather than assumed? [Assumption, Spec Assumptions; plan Complexity Tracking for Constitution V]
