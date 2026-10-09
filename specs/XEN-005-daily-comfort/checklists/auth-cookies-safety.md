# Requirements Quality Checklist: Auth, Cookies & Interaction Safety

**Purpose**: Unit-test the stage-5 requirements for token/cookie safety, protocol correctness and
keyboard/drag interactions before `/create-tests`.
**Created**: 2026-10-09
**Feature**: [spec.md](../spec.md) · [contracts](../contracts/)
**Audience / depth**: implementer & reviewer · thorough

## Auth correctness
- [x] CHK001 Are all OAuth 2.0 grants and configuration fields specified with Postman's names? [Spec §FR-002; format additions]
- [x] CHK002 Is the authorization-code flow fully specified (browser, loopback, state, PKCE, timeout, cancel)? [Spec §FR-003; data-model state machine]
- [x] CHK003 Are token injection rules (placement, prefix, replacement, unsupported types) unambiguous? [Spec §FR-006, Edge Cases]
- [x] CHK004 Is refresh behaviour defined, including concurrency and failure paths? [Spec §FR-005 — single-flight added]
- [x] CHK005 Are AWS SigV4 inputs, modes, payload hashing and stale-header removal defined? [Spec §FR-007; research R2]
- [x] CHK006 Is the Digest exchange (algorithms, qop, retry limit, state storage) defined? [Spec §FR-008, Edge Cases]
- [x] CHK007 Are deviations from Postman (loopback callback, RFC nc) explicit? [Spec Assumptions]
- [x] CHK008 Is protocol correctness measurable against independent vectors/servers? [Spec SC-001–SC-003]

## Secrets & safety
- [x] CHK009 Is token storage restricted to the vault with masked display and explicit reveal/copy? [Spec §FR-004; bindings §1]
- [x] CHK010 Are imported literal tokens kept out of files? [Format additions; tasks T007]
- [x] CHK011 Is the browser-opening surface restricted to safe schemes? [Gap→resolved, Spec §FR-003]
- [x] CHK012 Is the loopback listener's exposure bounded (host, single request, lifetime)? [Spec §FR-003; research R1]
- [x] CHK013 Are cookies encrypted at rest and isolated per workspace, with locked-vault behaviour defined? [Spec §FR-009; research R4]
- [x] CHK014 Is the no-leak requirement measurable across files, local data, logs and console? [Spec SC-004; tasks T030]
- [x] CHK015 Are snippet secret-masking rules and the copy behaviour unambiguous? [Spec §FR-019]

## Interaction
- [x] CHK016 Is every shortcut listed with platform mapping and a conflict rule for AltGr layouts? [Spec §FR-012, SC-005]
- [x] CHK017 Are tab-menu actions and their unsaved-change handling defined (close vs force)? [Spec US4 #1]
- [x] CHK018 Is duplicate-tab behaviour consistent with XEN-001's one-tab-per-request rule and conflict banner? [Spec Clarifications; data-model Tabs]
- [x] CHK019 Are drag-and-drop valid/invalid drops, persistence and file moves defined? [Spec §FR-017, US3]
- [x] CHK020 Is the closed-tab stack bounded and draft-preserving? [Spec §FR-013]

## Notes
- All items pass after this run's fixes.
