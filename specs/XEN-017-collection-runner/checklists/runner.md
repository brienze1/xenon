# Runner Requirements Quality Checklist: Collection Runner

**Purpose**: Unit tests for the requirements of XEN-017. They check whether the run semantics,
data and secret handling, lifecycle, performance metrics and MCP access are complete, clear,
consistent and measurable before implementation.
**Created**: 2026-10-23
**Feature**: [spec.md](../spec.md)
**Depth / audience**: standard; the PR reviewer and the implementer of `/create-tests` and
`/implement-task`.

Every item was evaluated against the artifacts:

- An item that passed as written is ticked with the section that makes it pass.
- An item that failed was fixed in spec.md, research.md, data-model.md, contracts/ or tasks.md in
  this run, then ticked. Its note says "Fixed:".

## Execution semantics — Postman parity

- [x] CHK001 Are Postman's default values specified for every functional setting? [Completeness, Spec §FR-010] — each toggle, iterations and delay has its default in FR-010.
- [x] CHK002 Is "error" for "Stop run if an error occurs" defined so that it is distinct from a failed assertion? [Clarity, Spec §FR-023] — FR-023 covers a send failure, a missing request and a script error; failed assertions never stop a run.
- [x] CHK003 Is the resolution order of a `setNextRequest` target specified, including duplicate names? [Clarity, Spec §FR-020] — by id, else the first selected row in run order with that name.
- [x] CHK004 Is it specified which scripts' `setNextRequest` calls count (pre-request vs post-response, levels) and when they take effect? [Completeness, Spec §FR-020] — Fixed: both phases and all levels count, the last call wins, and it takes effect when the row finishes.
- [x] CHK005 Is the behaviour defined when `setNextRequest` targets a request that exists but is unchecked? [Edge Case, Spec §FR-020] — Fixed: "a target that is not a selected row" ends the iteration with a warning.
- [x] CHK006 Is the interplay of `skipRequest` and `setNextRequest` in the same row defined? [Consistency, Spec §FR-020, §FR-021] — Fixed: calls in a skipped row count, and they apply right after the skip.
- [x] CHK007 Are requirements defined for `setNextRequest` cycles (infinite loops)? [Edge Case, Gap] — Fixed: an edge case and FR-020 say the cycle runs until stopped, with no loop limit, as in Postman.
- [x] CHK008 Is the precedence of iteration data relative to the other variable scopes specified? [Clarity, Spec §FR-018] — between local and environment, as in Postman.
- [x] CHK009 Is the behaviour specified when iterations are fewer or more than the data rows? [Completeness, Spec §FR-019] — Fixed: fewer iterations use the first rows; more reuse the last row.
- [x] CHK010 Is the text form of non-string JSON values in `{{field}}` resolution specified? [Clarity, Spec §FR-019] — Fixed: strings as is, numbers and booleans as written, null as empty, objects and arrays as compact JSON; `pm.iterationData` keeps the types.
- [x] CHK011 Are data file encoding and malformed-file cases specified? [Edge Case, Spec §FR-019, Edge Cases] — Fixed: UTF-8 with a BOM ignored; a missing header, a non-array of objects, or an unreadable file gives `RUN-05400`; ragged CSV rows are defined.
- [x] CHK012 Is the delay's placement defined (between rows, across iterations, after the last row)? [Clarity, Spec §FR-022] — Fixed: between rows, across iteration boundaries, and none after the last row.
- [x] CHK013 Does "Keep variable values" define which scopes it covers, including the vault and secret variables? [Completeness, Spec §FR-024] — Fixed: environment, collection, globals, secret variables and `pm.vault`; when off, the vault is also unchanged.
- [x] CHK014 Are all four combinations of the two cookie settings unambiguous? [Clarity, Spec §FR-025] — FR-025 covers the start jar and the end merge independently, plus the in-run cookie use.
- [x] CHK015 Is it specified which version and which inherited configuration an alias row sends? [Clarity, Spec §FR-016, §FR-008] — the chosen alias version through the single-send pipeline; a missing alias falls back to Default.
- [x] CHK016 Is it specified whether a run sends the saved request or an unsaved draft, and when the saved request is read? [Clarity, Spec §FR-016, §FR-017] — the saved request on disk, read at the moment the row's turn comes.
- [x] CHK017 Are the expired-token and missing-secret behaviours in a run defined consistently with single sends? [Consistency, Spec Edge Cases, §FR-014] — missing secrets use `SND-06409` with Run anyway; expired tokens are accepted (Postman never prompts in runs), unlike `SND-07409` for single sends.

## Data handling and secrets

- [x] CHK018 Is every place a secret value may appear in a stored, exported or MCP-returned run enumerated for masking? [Completeness, Spec §FR-039] — Fixed: URL, headers, body, console, test names and messages, script errors and the row error.
- [x] CHK019 Do the masking requirements cover secrets that come into existence during the run? [Coverage, Spec §FR-039] — Fixed: secret values set during the run are masked as well.
- [x] CHK020 Is the treatment of iteration data values (secret or not) stated explicitly? [Assumption, Spec §FR-039] — Fixed: they are not secrets and appear as sent, as in Postman; credentials belong in secret variables (Constitution II defines secrets as vault values).
- [x] CHK021 Are the storage location and the exclusion from the workspace and git specified? [Completeness, Spec §FR-038, SC-008] — the app-data folder, never the workspace, never committed.
- [x] CHK022 Is the retention rule quantified and its unit (per collection or folder) defined? [Clarity, Spec §FR-038, Assumptions] — 25 per target, with a collection and its folders counted separately.
- [x] CHK023 Are deletion paths defined for stored runs: manual, cap, workspace removal, and deleted targets? [Coverage, Spec §FR-038, §FR-041a] — Fixed: workspace removal and orphan deletion added alongside manual delete, Clear and the cap.
- [x] CHK024 Is the lifecycle of in-memory persisted responses defined (start, limit, drop)? [Completeness, Spec §FR-040] — kept for the session only; spill past the memory limit; dropped on quit, delete or cap.
- [x] CHK025 Is it stated what a stored run must never contain? [Clarity, Spec §FR-039] — response headers and bodies are never stored; secrets are masked.
- [x] CHK026 Is the handling of corrupt or unreadable stored-run files defined? [Edge Case, Gap] — Fixed: the file is left out of the list with a logged warning.
- [x] CHK027 Is the scope of the data file path defined (remembered locally; only the base name stored on the run)? [Clarity, Spec §FR-013, §FR-039] — the path lives in local UI state, and the run stores only the data file name.
- [x] CHK028 Is the workspace boundary for runs and preferences stated? [Consistency, Spec §FR-041, US7-5, Constitution I] — runs are scoped to their workspace.

## Concurrency and lifecycle

- [x] CHK029 Is the concurrency rule for runs explicit, including a collection and its own folder? [Clarity, Spec §FR-030] — Fixed: one run per target; different targets (a collection and one of its folders included) may overlap, alongside single sends.
- [x] CHK030 Are pause semantics defined for a pause during the delay and for resumed timing? [Completeness, Spec §FR-029] — Fixed: the run holds before the next row, and paused time does not count as delay.
- [x] CHK031 Is it specified whether performance runs can be paused? [Gap, Spec §FR-029] — Fixed: Stop only, as in Postman, because pausing would distort the profile.
- [x] CHK032 Are stop semantics specified for the in-flight request and the resulting statuses? [Clarity, Spec §FR-029, §FR-033] — the in-flight row is cancelled and the run status is Stopped.
- [x] CHK033 Are workspace switch, workspace removal and app quit during a run defined? [Coverage, Spec §FR-031, Edge Cases] — a switch or close stops and stores the run; a quit does not store it.
- [x] CHK034 Are tree changes during a run defined (request deleted, renamed, moved)? [Edge Case, Spec Edge Cases] — the saved state at the row's turn is used; a missing request is a row error.
- [x] CHK035 Are the run status values and their transitions consistent across spec and data model? [Consistency, Spec §FR-033, data-model §Run] — running, paused, completed, stopped and stopped_on_error, with terminal states stored.
- [x] CHK036 Is the rejection of a second run on the same target specified for both UI and MCP? [Consistency, Spec Edge Cases, §FR-047] — `RUN-02409`; the UI shows the active run, and MCP gets an error naming it.
- [x] CHK037 Is live-progress latency quantified? [Measurability, Spec §FR-028, SC-003] — within 1 second of each completed request.

## Performance mode metrics

- [x] CHK038 Are the virtual-user and duration limits quantified? [Clarity, Spec §FR-011] — 1–100 users; 1 s to 60 min.
- [x] CHK039 Does every load-profile parameter have a defined default? [Completeness, Spec §FR-011] — Fixed: defaults for every Ramp up, Spike and Peak parameter, recomputed until edited.
- [x] CHK040 Is each load profile's user curve defined precisely enough to compute the expected active users at any second? [Measurability, Spec §FR-043] — Fixed and settled (Peak ramps split the remaining time equally; Spike is centred).
- [x] CHK041 Are "response time", "throughput" and "error" defined? [Clarity, Spec §FR-045, §FR-046] — Fixed: total request time as the response viewer shows it; throughput is completed requests over elapsed seconds; an error is not sent or status ≥ 400.
- [x] CHK042 Is percentile accuracy specified so it can be verified objectively? [Measurability, Spec §FR-046] — Fixed: within 1% of the exact value.
- [x] CHK043 Are console output and `pm.test` reporting in performance runs specified? [Gap, Spec §FR-044] — Fixed: no app console output, and `pm.test` runs but is not reported, as in Postman.
- [x] CHK044 Is the per-user isolation of variables, cookies and data rows defined? [Completeness, Spec §FR-042, §FR-044] — own state per user, never persisted; data rows wrap per user.
- [x] CHK045 Is app responsiveness under maximum load quantified? [Non-Functional, SC-006] — 100 users × 60 minutes with the UI staying usable (tabs, sending, editing).
- [x] CHK046 Is the content of a stored performance run defined (and what is not stored)? [Completeness, Spec §FR-046, Decisions table] — summary, per-request stats and the per-second series; no individual responses.

## MCP access

- [x] CHK047 Are the MCP operations enumerated with their defaults? [Completeness, Spec §FR-047, contracts/mcp.md] — start (defaults from the remembered or default configuration), get, list and stop.
- [x] CHK048 Is the confirm-before-send approval for AI-started runs specified as once per run? [Clarity, Spec §FR-047, US9-4] — the existing AI approval runs once per run.
- [x] CHK049 Is visibility between UI-started and AI-started runs defined in both directions? [Consistency, Spec §FR-048] — each is visible from the other side.
- [x] CHK050 Are secret masking and the leak guard required for MCP run output? [Coverage, Spec §FR-039, §FR-047] — masked as stored, plus the leak guard.
- [x] CHK051 Is it specified that runs never record History, including MCP-started runs? [Consistency, Spec §FR-026, contracts/mcp.md] — FR-026 applies to all runs, and the MCP contract records no `history` change.

## UI requirements relevant to the above

- [x] CHK052 Are keyboard reordering and its accessibility (announcements) specified? [Coverage, Spec §FR-005] — Space, arrows and Escape, with screen-reader announcements.
- [x] CHK053 Are requirements defined for very large result sets? [Non-Functional, Gap, Spec §FR-032] — Fixed: at least 10,000 rows stay scrollable without lag, rendering only the visible rows.
- [x] CHK054 Are empty and invalid states defined for the Runner (empty target, empty selection, field errors, missing data file)? [Coverage, Spec Edge Cases, §FR-013, §FR-014] — an empty state, a disabled Start, inline field errors, and the missing-file notice.
- [x] CHK055 Is the Runner layout (sub-tabs, where results and past runs appear) specified unambiguously? [Clarity, Spec §FR-001, Clarifications] — the Setup, Results and Runs sub-tabs and when each is selected.

## Dependencies and assumptions

- [x] CHK056 Is the dependency on the existing single-send pipeline stated, with the guarantee that single sends do not change? [Assumption, Spec Assumptions, SC-008] — reuse is stated, and the plan's constraint plus tasks T008 and T070 keep the existing suites green.
- [x] CHK057 Is the PRD change (removing the non-goal, recording decisions) a stated requirement? [Traceability, Spec §FR-049, Constitution VI] — FR-049.
- [x] CHK058 Are out-of-scope items explicit (Schedule, CLI, CI/CD run methods)? [Scope, Spec Decisions, Assumptions] — Local only; the CLI remains a PRD non-goal.
