# Requirements Quality Checklist: Import Fidelity, Untrusted Input & Atomicity

**Purpose**: Unit-test the stage-3 requirements for lossless import, safe handling of untrusted
input and failure behaviour before `/create-tests`.
**Created**: 2026-10-09
**Feature**: [spec.md](../spec.md) · [plan.md](../plan.md) · [mappings](../contracts/)
**Audience / depth**: implementer & reviewer · thorough

## Fidelity & round-trip

- [x] CHK001 Is every Postman v2.1 field either mapped or explicitly routed to passthrough? [Contracts postman-mapping "every other source key goes to passthrough"; Spec §FR-007]
- [x] CHK002 Is the passthrough rule defined so that edits/deletions in Xenon win on export? [Research R2 — owned arrays regenerated]
- [x] CHK003 Are auth inheritance and explicit "no auth" distinguished? [Spec §FR-005]
- [x] CHK004 Is the treatment of auth types Xenon cannot send defined (stored, shown, export)? [Format additions §1; Spec §FR-005]
- [x] CHK005 Are disabled scripts represented as a first-class field rather than only passthrough? [Gap→resolved, Spec §FR-020b; format additions §4]
- [x] CHK006 Are Postman folder/request-level variables (no Xenon place) handled without loss and surfaced? [Spec §FR-007, US1 #5]
- [x] CHK007 Are examples mapped completely, including passthrough fields (cookies, timings, preview language)? [Postman-mapping "Example"]
- [x] CHK008 Is v2.0 → v2.1 normalisation specified? [Spec §FR-008; mapping Auth]
- [x] CHK009 Are Bruno ordering, settings and unsupported request types specified? [Bruno-mapping §1–§2]
- [x] CHK010 Is the script translation coverage defined by a published table with explicit non-translatable calls? [Bruno-mapping §3–§4]
- [x] CHK011 Is the "import into folder" behaviour unambiguous for collection-level auth/headers/scripts/variables? [Spec Edge Cases; plan decision 4]
- [x] CHK012 Is the fidelity bar measurable? [Spec SC-001 — fixture corpus field-by-field + real exports]

## Untrusted input & secrets

- [x] CHK013 Are zip limits defined for size, path traversal and symlinks? [Gap→resolved, Spec Edge Cases "Untrusted archives"; research R9]
- [x] CHK014 Are input size and encoding limits specified for single files? [Gap→resolved, Spec Edge Cases]
- [x] CHK015 Is the risk of imported file references uploading local files addressed? [Gap→resolved, Spec §FR-020a; format additions §5]
- [x] CHK016 Are imported secret values kept out of workspace files, passthrough, summaries and logs? [Spec §FR-009, SC-004; plan gate II]
- [x] CHK017 Is it stated that imported scripts are untrusted code that only ever runs in the stage-4 sandbox? [Assumption; constitution III]
- [x] CHK018 Are unsupported/unknown formats rejected before anything is written, with a reason? [Spec §FR-002, Edge Cases]

## Failure & atomicity

- [x] CHK019 Is the unit of atomicity defined (per collection/environment) and what remains on failure? [Spec §FR-017; research R10]
- [x] CHK020 Is cancellation defined before and during writing? [Spec §FR-019; data-model state machine]
- [x] CHK021 Are multi-workspace (dump) partial failures reported per item and workspace? [Data-model ImportResult; US3 #3]
- [x] CHK022 Are clash resolutions defined for every kind (collection, environment, request, variable) including "apply to all"? [Spec §FR-018]
- [x] CHK023 Is the behaviour of in-progress staging data on crash defined (ignored by tree/watcher, git-ignored, cleaned)? [Research R10; format additions §6]

## Acceptance criteria & UX

- [x] CHK024 Are warning categories enumerated and each linked to an item? [Data-model ImportWarning codes; SC-007]
- [x] CHK025 Is the pm-usage report defined with grouping and locations? [Spec §FR-016; data-model PmUsage]
- [x] CHK026 Are performance and progress requirements quantified? [Spec SC-002, §FR-019]
- [x] CHK027 Is cURL equivalence across shells measurable? [Spec SC-005]

## Notes

- All items pass after this run's fixes (marked "Gap→resolved").
