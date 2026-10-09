# Requirements Quality Checklist: Secrets, Resolution & Integrity

**Purpose**: Unit-test the stage-2 requirements for secret safety, Postman-parity resolution and
multi-store data integrity before `/create-tests`.
**Created**: 2026-10-09
**Feature**: [spec.md](../spec.md) · [plan.md](../plan.md) · [data-model](../data-model.md)
**Audience / depth**: implementer & reviewer · thorough

## Secrets — completeness

- [x] CHK001 Is it specified exactly what a secret row stores in workspace files (and what it must never store)? [Spec §FR-017; format additions §1]
- [x] CHK002 Are all places where a secret value could be persisted enumerated (files, logs, History, UI state, temp files, local current values)? [Spec §FR-023; data-model]
- [x] CHK003 Are requirements defined for text Xenon generates from resolved requests (errors, hints, summaries, logs)? [Gap→resolved, Spec §FR-023]
- [x] CHK004 Is the one channel that may return a secret value to the UI specified? [Contracts bindings — `VaultService.Reveal`; `CopySecret` stays in Go]
- [x] CHK005 Are keychain-unavailable, wrong-passphrase, lost-key and corrupted-vault paths specified with outcomes? [Spec Edge Cases, §FR-019, §FR-023; data-model state machine]
- [x] CHK006 Is removal of vault, key and keychain entry specified on workspace removal? [Spec §FR-024]
- [x] CHK007 Is the plain↔secret switch specified in both directions, including the git-history warning? [Spec §FR-021]
- [x] CHK008 Are secrets introduced through git (row turned secret by a teammate) prevented from leaving plain copies locally? [Gap→resolved, Spec Edge Cases; data-model]

## Secrets — clarity & consistency

- [x] CHK009 Is "masked" defined concretely and consistently across editor, quick look, panel and hovers? [Spec §FR-020 — `•••••`, per-row reveal, copy]
- [x] CHK010 Is the keychain storage granularity (one entry per workspace) specified and consistent with the PRD rationale? [Spec §FR-018; research R4]
- [x] CHK011 Is the passphrase lifecycle (setup ×2, once per start, never stored, min length) unambiguous? [Spec §FR-019; bindings `VLT-*`]
- [x] CHK012 Is the scope of the keychain threat model stated (same-user processes on macOS)? [Assumption, research R4]
- [x] CHK013 Is passphrase change explicitly in or out of scope? [Gap→resolved: out, Spec Edge Cases]
- [x] CHK014 Are response bodies echoing secrets explicitly addressed? [Spec §FR-023 — server content, memory/owner-only temp only]

## Resolution — Postman parity

- [x] CHK015 Is precedence specified for every scope, including the stage-4 local scope? [Spec §FR-011]
- [x] CHK016 Is disabled-row behaviour (fall-through) specified? [Spec §FR-011, US2 #5]
- [x] CHK017 Are name-matching rules (exact, case-sensitive, spaces significant) specified? [Spec Edge Cases; data-model tokenizer]
- [x] CHK018 Is nested resolution bounded with the same limit as Postman, and cycle behaviour defined? [Spec §FR-013 — 19 passes]
- [x] CHK019 Is unresolved-variable behaviour consistent between sending (literal) and display (red)? [Spec §FR-014, §FR-025]
- [x] CHK020 Is every request field that resolves variables enumerated? [Spec §FR-012]
- [x] CHK021 Are initial vs current semantics (default, Reset, Persist, empty value) fully specified? [Spec §FR-008/009; Edge Cases "Empty current value"]
- [x] CHK022 Is dynamic-variable coverage defined by an authoritative, checkable list? [Spec §FR-016; contracts/dynamic-variables.md (118)]
- [x] CHK023 Is per-occurrence generation and user-variable precedence over dynamic names specified? [Spec US6 #2–#3, §FR-016]
- [x] CHK024 Is it specified that History and drafts keep unresolved text? [Spec §FR-012]
- [x] CHK025 Are the Xenon-only additions (confirm before send, passphrase fallback, folder variables excluded) explicitly marked as deliberate deviations from Postman? [Spec Clarifications; research R1]

## Data integrity

- [x] CHK026 Is the write ordering and rollback for the three stores (YAML, local values, vault) specified? [Plan decision 4]
- [x] CHK027 Is reconciliation after a crash between store writes defined? [Gap→resolved, data-model vault consistency]
- [x] CHK028 Are renames of variable keys specified to carry current values and vault entries? [Data-model LocalVariables]
- [x] CHK029 Are outside (git) edits to environment files and active-environment deletion specified? [Spec §FR-029, Edge Cases]
- [x] CHK030 Is malformed environment file handling consistent with XEN-001 FR-016? [Spec Edge Cases]

## Acceptance criteria & coverage

- [x] CHK031 Is "no secret leakage" objectively measurable? [Spec SC-002; tasks T026, T047]
- [x] CHK032 Is confirm-before-send enforcement measurable and independent of UI behaviour? [Spec SC-007; plan decision 3 — backend preflight]
- [x] CHK033 Are performance targets for highlighting/resolution quantified? [Spec SC-004; T046]
- [x] CHK034 Are cross-platform keychain behaviours covered by requirements and tests on all three OSes? [Spec SC-008; research R4; mock modes]

## Notes

- All items pass after this run's fixes (marked "Gap→resolved").
