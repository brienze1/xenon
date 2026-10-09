# Requirements Quality Checklist: Sync Safety & Correctness

**Purpose**: Unit-test the stage-7 requirements for data safety, credential handling and merge
correctness before `/create-tests`.
**Created**: 2026-10-09
**Feature**: [spec.md](../spec.md) · [plan](../plan.md) · [bindings](../contracts/bindings.md)
**Audience / depth**: implementer & reviewer · thorough

## Data safety
- [x] CHK001 Is it specified that Xenon never touches checkouts other than its own clones? [Spec §FR-002, SC-004]
- [x] CHK002 Is preservation of uncommitted local work defined for fast-forward and replay pulls? [Spec §FR-009, Edge Cases]
- [x] CHK003 Is Cancel defined to restore the exact pre-pull state? [Spec §FR-012]
- [x] CHK004 Is "no change lost" measurable under concurrent edits? [Spec SC-002; tasks T006]
- [x] CHK005 Is concurrency between sync operations bounded? [Spec §FR-020; bindings `GIT-05409`]
- [x] CHK006 Are unlink and re-link paths defined without data loss? [Spec §FR-003, §FR-021]
- [x] CHK007 Is the interplay with the file watcher during pulls defined? [Gap→resolved, tasks T019]

## Merge semantics
- [x] CHK008 Are automatic merge rules defined per file kind (request, environment, structural)? [Spec §FR-011; research R2/R3]
- [x] CHK009 Is the history shape defined (linear, one combined commit, original messages kept)? [Spec §FR-009 — aligned with plan]
- [x] CHK010 Are delete-vs-modify and field-level conflicts defined with their choices? [Spec §FR-011]
- [x] CHK011 Is "Keep both" naming and identity (new id) defined? [Spec §FR-011]

## Credentials & trust
- [x] CHK012 Are credential storage locations restricted (keychain/fallback; never config/URLs/files)? [Spec §FR-014]
- [x] CHK013 Is SSH host-key trust defined (pre-trusted, TOFU, mismatch)? [Spec §FR-015]
- [x] CHK014 Are host-specific token conventions and their verification defined? [Spec §FR-013; research R4]
- [x] CHK015 Is the external dependency (GitHub OAuth App client id) owned and its blast radius stated? [Spec Assumptions]
- [x] CHK016 Are org-level access restrictions surfaced with next steps? [Spec Edge Cases, §FR-019]
- [x] CHK017 Is untrusted repository content addressed (symlinks, scripts)? [Gap→resolved, Spec Edge Cases]

## Scope & errors
- [x] CHK018 Is what is (not) synced enumerated? [Spec §FR-017, SC-003]
- [x] CHK019 Is every error class mapped to a user-facing next step? [Spec §FR-019; bindings error codes]
- [x] CHK020 Are unsupported Git features (hooks, signing, LFS, branch switching) explicitly scoped out with behaviour? [Spec Assumptions, §FR-005, Edge Cases]
- [x] CHK021 Are remote changes to the shared sync settings defined? [Gap→resolved, Spec Edge Cases]

## Notes
- All items pass after this run's fixes.
