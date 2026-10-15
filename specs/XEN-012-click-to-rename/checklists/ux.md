# UX & State Requirements Quality Checklist: Rename by Clicking the Name

**Purpose**: Unit tests for the requirements of in-place renaming. Covers the editing UX,
draft/dirty/save consistency across the five tab types, save order and failure handling,
validation messages, accessibility, and scope boundaries.
**Created**: 2026-10-15
**Feature**: [spec.md](../spec.md)
**Audience / timing**: implementer, before `/create-tests`. Depth: standard.

Every item below was run against spec.md, plan.md, research.md, data-model.md, contracts/ and
tasks.md. Items that failed on the first pass were fixed in the artifacts and are marked
"(fixed)".

## Requirement Completeness

- [x] CHK001 Are the tab types that get an editable name, and the ones that don't, listed explicitly? [Completeness, Spec §Context table, §FR-001, Edge Cases "Globals and Settings"]
- [x] CHK002 Is the location of the name editor specified for every tab type? [Completeness, Spec §FR-003, Plan "Where each tab's name editor sits"]
- [x] CHK003 Are the visual affordances that mark the name as editable (cursor, hover, tooltip, focus ring) specified? [Completeness, Spec §FR-002, contracts/editable-name.md "Not editing"] (fixed: added to FR-002 and the contract)
- [x] CHK004 Is the display of a name too long for its space specified? [Completeness, Spec §FR-002] (fixed: one-line ellipsis, full name in the tooltip)
- [x] CHK005 Is it specified what removes an inline name error? [Completeness, Spec §FR-010, data-model.md "Name error"] (fixed: FR-010 now lists edit, Escape, discard, save, close)
- [x] CHK006 Are requirements defined for the never-saved request, which has no saved name? [Completeness, Spec Edge Cases "Never-saved request", contracts/tab-save.md `setName`] (fixed during analyze: `setName` still marks a never-saved request dirty)
- [x] CHK007 Is the saved-name source defined per tab type, so "shows the new name only after saving" is unambiguous? [Completeness, Spec §FR-009, data-model.md "Saved name", research.md R5]

## Requirement Clarity

- [x] CHK008 Is "pending name" defined precisely, including trimming? [Clarity, Spec §FR-004, data-model.md `isNamePending`]
- [x] CHK009 Is the behaviour of each key in the input (Enter, Escape, Tab, Mod+S) defined unambiguously? [Clarity, Spec §FR-006, §FR-007, Clarifications Q "Enter", contracts/editable-name.md]
- [x] CHK010 Is where keyboard focus goes after the input closes specified? [Clarity, Spec §FR-007, contracts/editable-name.md] (fixed: back to the name text after Enter/Escape; left alone after blur)
- [x] CHK011 Are the validation messages given as exact texts? [Clarity, research.md R4, Spec Assumptions]
- [x] CHK012 Is "too long" given as a number, and is how it is enforced stated? [Clarity, Spec §FR-012: 200 characters, input refuses more]

## Requirement Consistency

- [x] CHK013 Is the dirty/close/save behaviour of a pending name the same as for every other draft in all five tab types? [Consistency, Spec §FR-004, §FR-005, §FR-008, contracts/tab-save.md `isDirty`]
- [x] CHK014 Do Cmd/Ctrl+S and the close prompt's Save follow one save order? [Consistency, research.md R8: both call `saveTabChanges`]
- [x] CHK015 Is FR-003 ("one name draft") consistent with the request tab showing the name in two places? [Consistency, Spec §FR-003] (fixed during analyze: reworded)
- [x] CHK016 Is the "same request in two tabs" edge case consistent with the rename-elsewhere rule? [Consistency, Spec Edge Cases, §FR-014, contracts/tab-save.md `adoptRename`] (fixed during analyze)
- [x] CHK017 Are the tab label rules consistent for requests **and examples** (saved name only)? [Consistency, Spec §FR-009, research.md R5, tasks T019] (fixed: example tab titles also come from the tree; T019 and T010 extended)

## Acceptance Criteria Quality

- [x] CHK018 Can each user story's acceptance scenarios be checked objectively (DOM, store or call order)? [Measurability, Spec User Stories 1–5]
- [x] CHK019 Is SC-004's per-tab-type coverage traceable to concrete test tasks? [Measurability, Plan "BDD / test files…" table, tasks T002–T053]
- [x] CHK020 Is the all-or-nothing rule written so that a test can tell when it is violated? [Measurability, Spec §FR-011, contracts/tab-save.md steps 2–4]

## Scenario Coverage

- [x] CHK021 Are primary flows defined for every tab type (rename + other edit + save)? [Coverage, Spec US1–US4]
- [x] CHK022 Are exception flows defined for an invalid name, a failure in another part of the save, and a rename failure? [Coverage, Spec Edge Cases, §FR-010, §FR-011, data-model.md transitions]
- [x] CHK023 Are save flows that wait for a confirmation (documentation diff, secret-type switch) covered? [Coverage, Spec Edge Cases "Collection save that waits…", tasks T026, T039]
- [x] CHK024 Are recovery requirements defined for a rename that fails after the other parts were written? [Recovery, data-model.md last paragraph: the content stays saved, the name stays pending with its error, and the pre-checks limit this to I/O failures]
- [x] CHK025 Is the close flow (Save / Discard / Cancel, bulk close) covered for a name-only change? [Coverage, Spec US5, §FR-004, tasks T034]

## Edge Case Coverage

- [x] CHK026 Are an empty name, a whitespace-only name, an unchanged-after-trim name, and pasting over the limit covered? [Edge Case, Spec Edge Cases, §FR-012]
- [x] CHK027 Is a rename from outside the tab (sidebar F2, git sync, another tab) covered while a draft is pending? [Edge Case, Spec §FR-014, research.md R5/R7, tasks T009]
- [x] CHK028 Is a disk change event arriving for a request whose only change is a pending name covered? [Edge Case, contracts/tab-save.md `isDirty`, tasks T009/T017] (fixed during analyze)
- [x] CHK029 Is editing the name, or pressing save again, while a save is running covered? [Edge Case, Spec Edge Cases "Editing while a save is running", contracts/tab-save.md in-flight set] (fixed)
- [x] CHK030 Are Duplicate Tab, Save As and app restart with a pending name covered? [Edge Case, Spec Edge Cases "Name draft lifetime", "Duplicate Tab", "Save As"; research.md R6 `keepDraft`]
- [x] CHK031 Is an environment-name clash defined with the same rule the backend uses (case and spaces ignored, self excluded)? [Edge Case, research.md R4, contracts/tab-save.md `findNameClash`]

## Non-Functional Requirements

- [x] CHK032 Are accessibility requirements specified: accessible names, error announcement, `aria-invalid`/`aria-describedby`, keyboard-only use? [Accessibility, Spec §FR-017, contracts/editable-name.md]
- [x] CHK033 Is the rendering-cost expectation stated (typing re-renders only what selects the draft)? [Performance, Plan Technical Context "Performance Goals"]
- [x] CHK034 Is cross-platform hotkey behaviour (Cmd on macOS, Ctrl elsewhere) stated? [Constitution VII, Plan Constitution Check VII, tasks T014]

## Dependencies & Assumptions

- [x] CHK035 Is the assumption that no backend change is needed backed by code facts? [Assumption, research.md R1–R4: the request save writes the name without a file rename; the settings save ignores the name; the example save sends the saved name]
- [x] CHK036 Is the assumption that Mod+S fires from inside inputs backed by evidence? [Assumption, research.md R10: global `useHotkeys` binding already used from the URL bar; tasks T014 proves it end to end]
- [x] CHK037 Is the deviation from Constitution V (no new Godog features) recorded with a justification and a replacement red-first test plan? [Dependency, Plan Complexity Tracking, tasks "Tests" header]

## Scope Boundaries

- [x] CHK038 Is it explicit that the sidebar tree's F2/context-menu rename and the workspace rename dialog are unchanged? [Scope, Spec §FR-013, Clarifications Q1, Plan "Unchanged on purpose"]
- [x] CHK039 Is the one change in the sidebar rename path (`onRenamed` gains the previous name) bounded so it can't change sidebar behaviour? [Scope, research.md R7, tasks T023, which re-runs the sidebar F2 specs]
- [x] CHK040 Is "Globals where a name applies" from the issue resolved rather than left open? [Scope, Spec Assumptions: the Globals tab has no name, so nothing changes]

## Notes

- 40/40 items pass. Ten needed artifact fixes. Six were fixed in this pass (CHK003, CHK004,
  CHK005, CHK010, CHK017, CHK029; touching spec.md, contracts/, research.md and tasks.md). Four
  were fixed earlier, during analyze (CHK006, CHK015, CHK016, CHK028).
