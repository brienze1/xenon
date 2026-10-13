# Requirements Quality Checklist: Sandbox Parity, Isolation & Observability

**Purpose**: Unit-test the stage-4 requirements for Postman parity, sandbox isolation, limits and
console/secret handling before `/create-tests`.
**Created**: 2026-10-09
**Feature**: [spec.md](../spec.md) · [sandbox contract](../contracts/sandbox-api.md)
**Audience / depth**: implementer & reviewer · thorough

## Parity
- [x] CHK001 Is the execution order defined for both phases and all owner levels? [Spec §FR-002, US1 #1]
- [x] CHK002 Is the inherit toggle's effect defined for folders and requests, both phases? [Spec US1 #2, §FR-002]
- [x] CHK003 Is per-script scope isolation and the lifetime of the local scope specified? [Spec §FR-003]
- [x] CHK004 Is the exact pm surface enumerated (members, methods, no-ops, unsupported)? [Sandbox contract §1, §6; Spec §FR-008/012]
- [x] CHK005 Are assertion helper semantics and failure messages pinned to a reference? [Sandbox contract §2 — chai-postman 2.0.1]
- [x] CHK006 Are library versions pinned to a reference sandbox version? [Research R5 — postman-sandbox 6.7.4]
- [x] CHK007 Are async completion rules (timers, sendRequest, promises) specified? [Spec Edge Cases]
- [x] CHK008 Is test semantics for sync, `done`-async, `async ()` and legacy tests specified? [Spec US2, Edge Cases]
- [x] CHK009 Is parity measurable with an independent oracle? [Spec SC-001/005; research R7]
- [x] CHK010 Are deviations from Postman (vault mapping, Intl fallback, visualizer no-op) explicit? [Spec Edge Cases, Assumptions]

## Isolation & limits
- [x] CHK011 Is the sandbox's access boundary (FS, processes, network, app internals) specified? [Spec §FR-006]
- [x] CHK012 Is the exposed surface defined as an allowlist? [Gap→resolved, Spec Edge Cases]
- [x] CHK013 Are time, memory and stack limits defined with outcomes and defaults? [Spec §FR-006; research R3]
- [x] CHK014 Is the timeout configurable within a bounded range? [Spec §FR-006; bindings §4]
- [x] CHK015 Are unsupported runtime features defined to fail clearly? [Spec Edge Cases]

## Interaction with earlier stages
- [x] CHK016 Is it specified that scripts never change stored requests or History entries? [Spec §FR-004, Edge Cases]
- [x] CHK017 Are variable writes defined in terms of XEN-002 rules (current values, secrets, new rows)? [Spec §FR-010, Edge Cases]
- [x] CHK018 Are nested sends defined w.r.t. confirm-before-send and missing secrets? [Gap→resolved, Spec Edge Cases]
- [x] CHK019 Is History behaviour for skipped/stopped sends defined? [Conflict→resolved, data-model; tasks T014]

## Console & secrets
- [x] CHK020 Is every console entry type and its detail content defined? [Spec §FR-016; data-model ConsoleEntry]
- [x] CHK021 Is masking defined for all entry fields, including secrets set mid-run, and its limits stated? [Spec §FR-018, Edge Cases]
- [x] CHK022 Are console retention and size limits quantified? [Spec §FR-017]
- [x] CHK023 Is performance quantified for script overhead and stopping runaway scripts? [Spec SC-002/003]

## Notes
- All items pass after this run's fixes.
