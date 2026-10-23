# Alias Environments Requirements Quality Checklist: Alias Environments

**Purpose**: Unit tests for the requirements of XEN-018 (security of the send guard, storage
compatibility, UX states, MCP contract, Postman export), run before implementation
**Created**: 2026-10-23
**Feature**: [spec.md](../spec.md)
**Depth / audience**: standard; PR reviewer and implementer

## Requirement Completeness

- [x] CHK001 Are availability rules defined for every combination of linked/unlinked alias × active/none/deleted active environment? [Completeness, Spec §FR-003] — FR-003 plus Edge Cases ("Active environment that no longer exists" counts as none); research R2
- [x] CHK002 Is the error behaviour of a refused send fully specified (code, message content, nothing sent, no history, no scripts, no confirmation)? [Completeness, Spec §FR-005] — FR-005; contracts/bindings.md `ALS-09403`; research R7
- [x] CHK003 Are requirements defined for every place a send of a saved alias can originate (app button, hotkey, send dialogs, MCP, future runner)? [Coverage, Spec §FR-006] — FR-006; Context "Where the guard lives"; contracts/ui.md (store `send` covers Mod+Enter and dialogs); runner inherits via the one backend send (Assumptions)
- [x] CHK004 Are validation rules for writing a link (unknown ids, duplicates, malformed environment files, missing fields) specified with error codes? [Completeness, Spec §FR-011] — FR-011; data-model.md validation table (`ALS-10400`, `ALS-11400`, `ALS-01404`, `REQ-02409`); research R5
- [x] CHK005 Is the on-disk format of the link, including its absence, specified precisely enough to guarantee byte-identical legacy files? [Completeness, Spec §FR-002] — data-model.md YAML section (key after `name`, emitted only when non-empty); research R1
- [x] CHK006 Is the history record format (fields, omission rule, backwards compatibility) specified? [Completeness, Spec §FR-016] — data-model.md History JSON (`environment` omitempty, `historyVersion` stays 1); research R9
- [x] CHK007 Are all UI surfaces that show or change the link enumerated (dropdown row, ⋯ menu, Environments dialog, Create dialog, banner, Send button, History row)? [Completeness, Spec §FR-007–FR-010, FR-016] — contracts/ui.md has a section for each
- [x] CHK008 Are the MCP read and write shapes, including how to leave the link unchanged vs clear it, specified? [Completeness, Spec §FR-013] — contracts/mcp.md (missing = unchanged, `[]` = all; `environments` always present)

## Requirement Clarity

- [x] CHK009 Is "unavailable" defined once and used with the same meaning in spec, plan, contracts and tasks? [Clarity, Spec §FR-003] — FR-003 is the single definition; data-model `IsAvailableIn`; research R3 mirrors it in the UI helper with identical cases (T002 ↔ T012)
- [x] CHK010 Is "greyed out and not selectable" specified in observable terms (disabled state, keyboard, accessible name)? [Clarity, Spec §FR-009] — contracts/ui.md: `disabled`, not selectable by mouse or keyboard, accessible name suffix
- [x] CHK011 Is the blocked-Send message's wording and its actions specified, including the all-deleted case? [Clarity, Spec §FR-010] — contracts/ui.md banner text, one Switch button per existing environment, all-deleted wording with no button
- [x] CHK012 Is the content of the Postman export note specified exactly (text, separator, placement)? [Clarity, Spec §FR-015] — research R10 (`Xenon alias for environments: …`, blank line before when a description exists)
- [x] CHK013 Is the "Deleted environment" label and its behaviour in the dialog unambiguous (shown, not submitted, removed on save)? [Clarity, Spec §FR-012] — contracts/ui.md picker section; research R5 note

## Requirement Consistency

- [x] CHK014 Do the decisions table, FRs, Edge Cases and contracts agree on deleted environments (link kept, never widened, flagged)? [Consistency, Spec §FR-012] — all four say the id stays and the alias is never widened; checked after the analyze fixes
- [x] CHK015 Are the Save vs SetEnvironments responsibilities consistent (content save never changes the link; link write never changes content)? [Consistency, Spec Clarifications] — contracts/bindings.md (`Save` ignores `environment_ids`); research R4; T007, T009, T015
- [x] CHK016 Is the Default's treatment consistent everywhere (never linkable, always available, sends unaffected)? [Consistency, Spec §FR-004] — FR-004, US1 scenario 6, plan guard note, T010, T020
- [x] CHK017 Are error codes consistent across spec references, data-model, contracts and tasks? [Consistency] — `ALS-09403`, `ALS-10400`, `ALS-11400`, `REQ-02409`, `ALS-01404` match in data-model.md, contracts/, research R7 and tasks T009/T010/T017/T018/T022/T038

## Acceptance Criteria Quality

- [x] CHK018 Can every success criterion be objectively measured? [Measurability, Spec §SC-001–SC-005] — SC-001/002 are 100% rates checked by BDD; SC-003 is a 15-second timed task; SC-004 is checked against the dropdown contract; SC-005 has a BDD rename scenario (T009)
- [x] CHK019 Does each user story's acceptance scenarios map to at least one BDD or Vitest task? [Traceability] — US1 → T009/T010/T013–T015; US2 → T031; US3 → T033/T034; US4 → T038; history/export → T043/T044

## Scenario Coverage

- [x] CHK020 Are alternate flows covered (multi-environment set, All environments alias, Default)? [Coverage, Spec §US1] — US1 scenarios 4–6
- [x] CHK021 Are exception flows covered (unknown id, stale revision, unknown alias, concurrent environment deletion while the dialog is open)? [Coverage, Exception Flow] — FR-011, Edge Cases; a deletion between opening the dialog and saving yields `ALS-10400`, shown inline (contracts/ui.md dialog errors)
- [x] CHK022 Are recovery flows defined (relinking an alias whose environments were deleted; returning to an available state)? [Coverage, Recovery] — FR-012 (save a new choice), FR-010 (Send re-enabled; Switch to buttons)
- [x] CHK023 Is the environment switch while an alias with unsaved edits is on screen covered? [Coverage, Spec §US1 scenario 7] — keep alias and drafts, block Send (FR-010, T015)

## Edge Case Coverage

- [x] CHK024 Are rename, recolour, duplicate (request and environment), Save As, and malformed active environment cases addressed? [Edge Case] — Edge Cases list rename/duplicate/Save As. A malformed active environment can't match a link: its id carries the error prefix and is never accepted as a link target (research R5), so linked aliases are unavailable, as with no environment
- [x] CHK025 Is the "Confirm before send" interaction ordering specified? [Edge Case, Spec §Edge Cases] — the guard runs before confirmation (Edge Cases, research R2)
- [x] CHK026 Is behaviour for an ad-hoc (unsaved) request or an MCP send naming an alias without a saved request defined? [Edge Case] — unaffected (Assumptions); MCP already rejects a stray `alias_id` (`MCPSendTarget.HasStrayAlias`), unchanged

## Non-Functional Requirements

- [x] CHK027 Is the security objective (no test credential reaches another environment) stated and traced to an enforcement point that the UI can't bypass? [Security, Spec §Context, §FR-005/006] — Context; constitution check II/III; research R2/R3
- [x] CHK028 Is it specified that the refusal message and logs carry no secret values? [Security, Gap] — research R7: the message holds only the alias and environment names; constitution check II in plan.md
- [x] CHK029 Are accessibility requirements specified for the new interactive UI (disabled rows, banner buttons, picker, history label)? [Accessibility] — contracts/ui.md: accessible names, `aria-disabled`, labelled checkboxes and radios, history accessible label
- [x] CHK030 Is the performance impact on send bounded? [Performance] — plan Technical Context: one in-memory check, and an environment listing only on the refusal path

## Dependencies & Assumptions

- [x] CHK031 Is the git/team sharing assumption grounded (same ids across teammates) and its limit stated? [Assumption, Spec §Context] — "Git and teams" bullet: shared workspace files share ids; separately created environments are different
- [x] CHK032 Is the dependency on the unmerged runner (XEN-017) resolved without leaving work for it? [Dependency] — Assumptions: it inherits the guard through the same backend send; no change required here
- [x] CHK033 Is the Postman parity principle (constitution VI) addressed with a recorded decision? [Dependency, Constitution VI] — plan Constitution Check VI; T052 records the PRD decision

## Ambiguities & Conflicts

- [x] CHK034 Is the Postman round trip of the export note defined so nobody expects re-linking on import? [Ambiguity] — research R10 "Round trip": imports create plain requests and the note stays as description text
- [x] CHK035 Is any requirement left with an unresolved marker or a decision deferred to the implementer? [Ambiguity] — none; every question was answered from code or by the user (spec Decisions + Clarifications)
