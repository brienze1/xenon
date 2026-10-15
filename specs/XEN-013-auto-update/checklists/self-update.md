# Self-Update Safety & Reliability Checklist: Auto-Update Xenon From GitHub Releases

**Purpose**: Release-gate check of the *requirements* for integrity, install atomicity, failure handling, cross-platform behaviour, release versioning, update UX and data preservation.
**Created**: 2026-10-15
**Feature**: [spec.md](../spec.md)
**Audience**: the implementer and the PR reviewer

**Note**: The `/speckit-checklist` command generated these items. Each one was then graded against
the artifacts. Where an item failed, the artifacts were fixed in this run, and the item notes
which change resolved it.

## Download Integrity & Trust

- [x] CHK001 - Is the integrity mechanism specified, both the algorithm and where the expected digest comes from? [Completeness, Spec §FR-011, §FR-012] — Yes: SHA-256 against `checksums.txt` from the same release (research R7, contracts/release-assets.md).
- [x] CHK002 - Is the trust boundary of checksum verification stated, including what it does not protect against? [Clarity, Spec §Clarifications] — Yes. It guards against corrupt or truncated downloads, not a compromised GitHub account; the user accepted this. Signature verification is the rejected alternative (research R7).
- [x] CHK003 - Is the outcome defined when the checksums file is missing or has no line for the package? [Edge Case, Spec §Edge Cases, §FR-012] — Yes: unverifiable, nothing installed (`unverified`).
- [x] CHK004 - Are download sources restricted to locations the release itself lists? [Completeness, Spec §FR-022]
- [x] CHK005 - Are archive-content safety rules specified (path traversal, links, the expected single top-level entry)? [Gap → resolved, research R8 step 2] — Yes: absolute paths, `..` and outside-pointing links are rejected, and exactly one `*.app` or `xenon`/`xenon.exe` entry is required.
- [x] CHK006 - Is it specified that a partial download can never be mistaken for a complete one? [Clarity, research R7] — Yes: `.part` file, renamed only after verification.

## Install Atomicity & Rollback

- [x] CHK007 - Is "replace the installed app as a whole" defined per OS (bundle, `.exe`, executable)? [Clarity, Spec §FR-013]
- [x] CHK008 - Is rollback defined for a failure at every install step, including the swap itself? [Coverage, Spec §Edge Cases, research R8 step 3] — Yes: the previous app is renamed back, and the outcome is `install_failed`.
- [x] CHK009 - Is the order of checks specified so the target is never touched before writability is known? [Clarity, research R8 step 1] — Yes: the writability probe always runs first.
- [x] CHK010 - Are leftover files from interrupted installs or downloads named, with a cleanup rule? [Completeness, Spec §FR-021, data-model §Persistence] — Yes. Cleanup is best effort and a locked file is retried at the next start (added in this run).
- [x] CHK011 - Is the Windows constraint (a running `.exe` cannot be overwritten) addressed explicitly? [Coverage, Spec §Edge Cases, research R8]

## Failure Handling

- [x] CHK012 - Is one user-visible outcome defined for every check failure class (offline, unreachable, error status, rate limit, unreadable answer, invalid version, draft/prerelease)? [Completeness, Spec §FR-009, contracts/release-assets.md] — Yes: all map to "Couldn't check for updates".
- [x] CHK013 - Is it stated that failures never block startup and never raise dialogs? [Clarity, Spec §FR-009, contracts/bindings.md] — Yes: failures are statuses, not errors.
- [x] CHK014 - Is behaviour defined when the install location is not writable, including that no elevation is attempted? [Completeness, Spec §FR-015]
- [x] CHK015 - Is a failed *automatic* download's UX specified (no toast, where it is shown, when it retries)? [Gap → resolved, Spec §FR-018b] — Added FR-018b.
- [x] CHK016 - Is a full disk during download covered? [Edge Case → resolved, Spec §Edge Cases] — Added: "could not be downloaded" and the partial file is removed.
- [x] CHK017 - Are timeouts for the API call and the download quantified? [Clarity, research R4 (15 s), R7 (10 min)]
- [x] CHK018 - Is the concurrency rule (one operation at a time) defined with the response to a second request? [Consistency, Spec §FR-010, contracts/bindings.md] — Yes: the current status is returned unchanged.

## Cross-Platform

- [x] CHK019 - Are the supported update targets per OS and processor explicit, with the outcome for unsupported ones? [Completeness, Spec §FR-003, §Assumptions, §US2-3] — Yes: macOS universal, Windows x64 and Linux x64; anything else gets "No update package for this system".
- [x] CHK020 - Are runtime prerequisites of the portable packages documented (Linux GTK 4 and WebKitGTK 6, Windows WebView2)? [Assumption → resolved, Spec §Assumptions] — Added in this run.
- [x] CHK021 - Is macOS Gatekeeper behaviour specified for the first install and for self-updates? [Coverage, Spec §FR-005, research R8] — Yes: release-notes guidance for the first install, and Go downloads carry no quarantine attribute.
- [x] CHK022 - Is App Translocation (running a quarantined app from its download folder) covered? [Edge Case → resolved, Spec §Edge Cases] — Added: it maps to the "can't replace itself" case.
- [x] CHK023 - Is the restart mechanism defined portably, including the single-instance lock? [Completeness, research R9] — Yes: spawn with `XENON_RESTART_AFTER_PID`, wait at most 15 s, then quit.
- [x] CHK024 - Are other running processes of the same executable (the `xenon mcp` bridge) addressed? [Gap → resolved, Spec §Edge Cases] — Added: the bridge keeps the old version until its client restarts it, and Windows lock cleanup is retried.

## Release Pipeline & Versioning

- [x] CHK025 - Is the version formula unambiguous, including the first release and a `VERSION` change? [Clarity, Spec §FR-001, research R2]
- [x] CHK026 - Is the single source of truth named, and is every place the version must appear listed (binary, Info.plist, tag)? [Completeness, Spec §FR-002, research R1]
- [x] CHK027 - Is all-or-nothing publishing defined? [Completeness, Spec §FR-004, research R3]
- [x] CHK028 - Are asset names and archive layouts specified exactly, so the updater and CI agree? [Consistency, contracts/release-assets.md, data-model §Release files]
- [x] CHK029 - Is ordering defined for back-to-back merges (no duplicate or skipped versions)? [Edge Case, research R3] — Yes: the `release` concurrency group without cancel.
- [x] CHK030 - Is the non-downgrade rule explicit? [Clarity, Spec §FR-008, §Edge Cases] — Yes: strictly newer only.
- [x] CHK031 - Is the role of `build/config.yml`'s version clarified, so it does not compete as a second source? [Consistency, research R1, quickstart §2]

## Update UX (Settings, Toast, Footer)

- [x] CHK032 - Is there exact copy for every state and reason in the section? [Completeness, contracts/ui.md]
- [x] CHK033 - Are the toast's position, trigger, actions, dismissal and persistence specified? [Clarity, Spec §FR-019, contracts/ui.md] — Yes. "Never auto-dismisses" was added in this run.
- [x] CHK034 - Are the footer notice rules consistent between the spec and the UI contract? [Consistency → resolved, Spec §FR-018a, contracts/ui.md] — FR-018a was aligned: "Update available" shows whenever the state is available.
- [x] CHK035 - Is the footer notice required with no workspace open? [Coverage, contracts/ui.md, tasks T046]
- [x] CHK036 - Are the Update automatically default and the reading of absent stored values specified? [Clarity, Spec §FR-017, research R12]
- [x] CHK037 - Are accessibility attributes specified for the toast and the switch? [Coverage, contracts/ui.md] — Yes: `role="status"`, `aria-live="polite"`, `role="switch"`.
- [x] CHK038 - Is opening external release pages kept in Go, consistent with constitution III? [Consistency, contracts/bindings.md `OpenReleasePage`]

## Data Preservation Across Restart

- [x] CHK039 - Is it specified that tabs, drafts and layout survive an update restart, with a measurable criterion? [Completeness, Spec §FR-014, §SC-005]
- [x] CHK040 - Is the race between the debounced UI-state save and the restart addressed? [Gap → resolved during analyze, contracts/bindings.md, tasks T032] — Yes: the UI flushes tab state before `Restart`.
- [x] CHK041 - Is "install on quit" (Later) defined so a normal quit loses nothing? [Clarity, Spec §FR-020, research R9] — Yes: it runs in the binding's `ServiceShutdown` after the normal shutdown path.

## Observability & Privacy

- [x] CHK042 - Is the logging requirement scoped, including what may never be logged? [Clarity, Spec §FR-023, research R14]
- [x] CHK043 - Is the absence of secrets and credentials in this feature stated? [Coverage, plan §Constitution Check II] — Yes: an unauthenticated API, no token.

## Acceptance & Testability

- [x] CHK044 - Does every FR map to a named BDD feature file or Vitest/node test? [Traceability, plan §Requirement → BDD lock]
- [x] CHK045 - Can the not-writable case be tested the same way on all three CI OSes? [Measurability → resolved, plan §Requirement → BDD lock] — Yes: a file-as-parent trick instead of chmod.
- [x] CHK046 - Are the parts that cannot be verified in CI (real publish, real macOS swap) given an explicit manual procedure? [Coverage, quickstart §2–§3, tasks T052]
