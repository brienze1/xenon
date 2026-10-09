# Requirements Quality Checklist: Export Fidelity & Secrets

**Purpose**: Unit-test the stage-6 requirements for lossless export, secret handling and failure
behaviour before `/create-tests`.
**Created**: 2026-10-09
**Feature**: [spec.md](../spec.md) · [formats](../contracts/export-formats.md)
**Audience / depth**: implementer & reviewer · thorough

## Fidelity
- [x] CHK001 Is the Postman export defined as the exact inverse of the import mapping, including passthrough merge precedence? [Spec §FR-002; formats §1; research R1]
- [x] CHK002 Is "lossless" defined by an explicit normalisation rule set? [Spec §FR-014; formats §4]
- [x] CHK003 Are folder/request exports defined so they behave the same in Postman (inherited auth/scripts/variables)? [Spec §FR-003; formats §1 "Folder export"]
- [x] CHK004 Is every Xenon-only field that cannot be expressed in Postman enumerated with its treatment? [Spec §FR-005; formats §1 table]
- [x] CHK005 Are Xenon YAML formats (zip, bundle, file, manifest) fully specified? [Formats §2]
- [x] CHK006 Is Xenon YAML round-trip defined as byte-identical? [Spec §FR-015, SC-003]
- [x] CHK007 Is acceptance by real Postman tooling measurable? [Spec SC-002 — Newman]

## Secrets
- [x] CHK008 Is it unambiguous that Postman exports never contain secret values? [Spec §FR-004, Clarifications]
- [x] CHK009 Are passphrase rules, encryption scope and self-containment defined for Xenon exports? [Spec §FR-009; formats §3; research R2]
- [x] CHK010 Is the import path for encrypted secrets defined for right/wrong/skip? [Spec §FR-012]
- [x] CHK011 Are tokens, cookies and current values excluded by default and the current-values option scoped to plain rows? [Spec §FR-008, FR-010]
- [x] CHK012 Are literal credentials in auth fields surfaced to the user? [Spec Edge Cases, §FR-017]

## Failure & safety
- [x] CHK013 Is export atomicity and cancellation defined with no partial files? [Spec §FR-018]
- [x] CHK014 Are untrusted-archive protections reused for Xenon YAML imports? [XEN-003 research R9 / T007 guard applies to all zip imports]
- [x] CHK015 Is importing a newer format version defined? [Spec US3 #4]
- [x] CHK016 Is performance quantified? [Spec SC-005]

## Notes
- All items pass.
