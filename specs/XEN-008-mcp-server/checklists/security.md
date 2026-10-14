# Security & Contract Requirements Checklist: MCP Server — Secrets Write-Only

**Purpose**: Unit tests for the requirements covering secret handling, the leak guard, local-only access control,
approvals, and the MCP tool / settings contract. This is a formal release gate for the PR reviewer.

**Created**: 2026-10-13

**Feature**: [spec.md](../spec.md) · [contracts/mcp-tools.md](../contracts/mcp-tools.md)

Every item was evaluated against the artifacts. Where an item first failed, the artifacts were fixed (noted as
"fixed:"), and then the item was ticked.

## Secret Handling — Completeness

- [x] CHK001 Are all paths through which a secret's value could reach the AI listed? [Completeness, Spec §FR-015, §FR-018, §FR-020] — FR-020 lists every surface: listings, inspection, resolved previews, history, console, snippets, test names and messages, script errors, response headers and bodies, and errors. FR-015 removes reveal/copy/vault. FR-018 refuses include-secrets and secret→plain.
- [x] CHK002 Are all write-only secret operations (create, replace, plain→secret, delete) specified? [Completeness, Spec §FR-017]
- [x] CHK003 Are the scopes that can hold secrets named for the write-only rules? [Completeness, Spec §FR-017] — fixed: FR-017 now names globals, environments and collection variables, and states that folders hold no variables (verified in code: `entity/collection.go` `OwnVariables` returns empty for folders).
- [x] CHK004 Is converting a secret to plain specified as refused, with its error? [Completeness, Spec §FR-018; research R7 `MCP-02403`]
- [x] CHK005 Are secrets that leave the AI response path (History arguments, app log) covered? [Completeness, Spec §FR-031] — fixed: FR-031 redacts secret-carrying arguments by name before the guard, and the app log records no arguments.
- [x] CHK006 Are OAuth tokens and git credentials covered by the write-only rule? [Coverage, Spec §FR-014, §FR-015, §FR-019] — OAuth gives metadata only and its tokens are in the guard's known set. Git account and credential management is not exposed.
- [x] CHK007 Are files written on the AI's behalf from response content addressed? [Completeness, Spec §FR-021; research R6] — they are written masked, and a binary body containing a secret is refused.

## Leak Guard — Clarity

- [x] CHK008 Is the mask token defined unambiguously? [Clarity, Spec §FR-019, Constitution II] — `•••••`.
- [x] CHK009 Are the encoded forms the guard recognises listed exhaustively? [Clarity, Spec §FR-019] — raw, URL-encoded, JSON-escaped, base64, and basic-auth `user:password`.
- [x] CHK010 Is the short-value rule quantified? [Clarity, Spec §FR-019] — values of 4 or more characters are masked inside text. Shorter values are masked only as a field's entire value.
- [x] CHK011 Is the guard's order relative to truncation and paging specified? [Clarity, Spec §FR-022; contracts pipeline step 4] — the body is masked whole, then sliced, and sizes refer to the masked text.
- [x] CHK012 Is the set of "known" secret values defined, including values replaced or deleted during the session? [Clarity, Spec §FR-019, Edge Cases] — the active workspace's secrets and OAuth tokens, plus every value seen this session.
- [x] CHK013 Is the guard's scope bounded so it never alters what is sent, the human's own files, or the UI? [Consistency, Spec §FR-021]

## Leak Guard — Residual Risk & Assumptions

- [x] CHK014 Are the accepted residual risks documented, with who accepted them and the mitigations? [Assumption, Spec Context] — transformed values, and sending to an AI-controlled destination. The product owner accepted them (decision 2026-10-13). Mitigations: "Confirm before send" and client approval of `send`. Shown in the Settings note (FR-003) and the PRD (T003).
- [x] CHK015 Is behaviour specified when the vault is locked? [Edge Case, Spec US2 sc. 8; research R6]

## Access Control

- [x] CHK016 Is the network exposure boundary defined (bind address, Host/Origin rules)? [Completeness, Spec §FR-002; research R3]
- [x] CHK017 Is the accepted exposure of having no token stated explicitly? [Assumption, Spec Context] — fixed: any process of the same OS user can connect. This is accepted because such a process can already read the vault key.
- [x] CHK018 Are logging requirements for rejected requests defined? [Gap, Spec §FR-002] — fixed: a warning with the reason, never the body.
- [x] CHK019 Is a request-size limit specified? [Gap, Spec §FR-002] — fixed: 4 MiB, then 413 (research R2; contract Limits).
- [x] CHK020 Is it specified that MCP cannot change its own access settings? [Completeness, Spec §FR-006; contract §19 `MCP-02403`]
- [x] CHK021 Is switching on/off and changing the port timed and specified? [Clarity, Spec §FR-004, US1 sc. 6 — within 2 s]
- [x] CHK022 Is the stdio bridge's protection against forwarding to a foreign server specified? [Edge Case, Spec Edge Cases "Port already in use"; contracts/stdio-bridge.md step 4]
- [x] CHK023 Is rate limiting deliberately in or out of scope? [Gap, Spec Assumptions] — fixed: intentionally excluded, with the reason (local and single-user; the AI history cap bounds storage).

## Approvals

- [x] CHK024 Is it explicit which AI actions require human approval, and which rely on the client's own approval? [Clarity, Spec Context, §FR-026, US3 sc. 3]
- [x] CHK025 Is the approval timeout quantified? [Clarity, Spec §FR-027 — 120 s]
- [x] CHK026 Is the approval prompt's content specified so that no secret is shown? [Completeness, Spec §FR-026 — the URL as written, unresolved]
- [x] CHK027 Are all ways an approval can end specified (approve, deny, timeout, quit, client disconnect)? [Coverage, Spec §FR-027] — fixed: client session end → denied and the prompt removed (T059, T062).
- [x] CHK028 Are multiple concurrent approvals specified? [Coverage, Spec US5 sc. 6 — FIFO]
- [x] CHK029 Is the scope of "Dangerously skip approvals" bounded, with its default and warning? [Clarity, Spec §FR-003, §FR-028 — default OFF; secrets unaffected]
- [x] CHK030 Are the missing-secret and expired-token acknowledgements distinguished from human approval? [Clarity, Spec §FR-029]

## MCP Tool Contract Quality

- [x] CHK031 Is the full tool catalog specified with names, actions, arguments, outputs and hints? [Completeness, contracts/mcp-tools.md — 24 tools]
- [x] CHK032 Is the destructive classification rule objective and consistent between spec and contract? [Consistency, Spec §FR-023; contract Hints] — removal-based; 12 R / 2 W / 10 D.
- [x] CHK033 Is the error format specified for every failure mode, with stable codes? [Completeness, Spec §FR-025; data-model Error codes; contract Conventions]
- [x] CHK034 Are the operations that are never exposed listed exhaustively? [Completeness, Spec §FR-015; contract "Never exposed"]
- [x] CHK035 Are pagination and compaction semantics defined (page size, cursor, omitted fields)? [Clarity, Spec §FR-022; research R5]
- [x] CHK036 Is behaviour for unknown actions and invalid arguments defined? [Edge Case, contract Conventions `MCP-07400`]
- [x] CHK037 Is client identity defined, including a missing client name? [Clarity, data-model SendOrigin — defaults to "AI client"]
- [x] CHK038 Is the workspace isolation boundary specified for MCP calls? [Coverage, Spec §FR-008, Edge Cases "Item id not found or from another workspace"]
- [x] CHK039 Are concurrent clients specified? [Coverage, Spec Edge Cases "Two AI clients at once"]
- [x] CHK040 Are protocol and version compatibility assumptions documented? [Assumption, Spec Assumptions; research R1 — automatic negotiation from 2024-11-05 to 2026-07-28]

## Settings Contract Quality

- [x] CHK041 Are validation ranges and error codes defined for every new settings field? [Completeness, data-model AppSettings / AIAccessSettings]
- [x] CHK042 Are defaults defined, including reading an older `app.json`? [Completeness, data-model — missing keys read as defaults]
- [x] CHK043 Are the semantics of raising and lowering history caps defined? [Clarity, Spec §FR-034, US6 sc. 8]

## Non-Functional & Measurability

- [x] CHK044 Are performance requirements quantified for connect, browse, list size and send? [Measurability, Spec SC-001, SC-005, SC-006]
- [x] CHK045 Is "secrets never leak" made objectively verifiable? [Measurability, Spec SC-004 — 0 hits across the tool × form × surface matrix (T034)]
- [x] CHK046 Is the hint accuracy measurable? [Measurability, Spec SC-007; T022]
