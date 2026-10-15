# API Contract & Data-Safety Checklist: MCP Tools Declare Their Arguments; Clearer Workspace and Import Errors

**Purpose**: "Unit tests" for the written requirements on the MCP tool schema and error contract,
validation messaging, untrusted import input and secrets safety. For the implementer and the PR
reviewer.
**Created**: 2026-10-15
**Feature**: [spec.md](../spec.md) · [plan.md](../plan.md) · [contracts/](../contracts/)

Each item was run against the artifacts. Where an item initially failed, the artifact was fixed
and the fix is named.

## Requirement Completeness — tool schema contract

- [x] CHK001 Are requirements defined for every tool, not only the two named in the issue? [Completeness, Spec §FR-001, SC-001] All 24 tools are in scope; data-model.md Appendix A lists every argument of every tool.
- [x] CHK002 Is the type mapping specified for every Go field shape the argument DTOs use (pointer, slice, nested struct, map, `json.RawMessage`)? [Completeness, research.md R4] Yes, R4 has the full table plus the `schema` override and the `schema:"-"` exclusion.
- [x] CHK003 Is it specified how tools with no actions (`send`, `git_push`) and single-action tools are declared? [Completeness, Spec §Edge Cases, contracts/mcp-tool-schema.md] No selector for no-action tools; an optional selector and no "Used by" for single-action tools.
- [x] CHK004 Is it specified how an argument shared by several actions with different requirements is declared? [Completeness, data-model.md Appendix A notes 1–2] Declared once, with unioned "Used by"/"Required for"; `used_by` narrows it.
- [x] CHK005 Is the case of one DTO shared by two different tools (`MCPDocsInput` for `docs` and `docs_edit`) covered? [Completeness, data-model.md Appendix A note 1] Yes, by the `schema_for` tag and per-tool filtering of `used_by`/`required_for`.
- [x] CHK006 Are closed value sets required to be declared as enums, with the values listed? [Completeness, Spec §FR-002] Yes; Appendix A lists every enum value set (on_clash, color, format, language, kind, use, source, choice, console type).

## Requirement Clarity — tool schema contract

- [x] CHK007 Is "a short description" pinned to a testable form? [Clarity, Spec §FR-003, research.md R5] Yes: one sentence, plus generated "Used by …"/"Required for …" with defined list wording (`a`, `a and b`, `a, b and c`) in contracts/mcp-tool-schema.md.
- [x] CHK008 Is "which actions require it" defined from a verifiable source rather than an author's judgement? [Clarity, data-model.md "Argument DTO tagging"] Yes; required_for values come from the validator or controller rejections, verified per field in Appendix A (code at `06d0d82`).
- [x] CHK009 Is the exact published schema given for at least the acceptance tools? [Clarity, contracts/mcp-tool-schema.md] Yes; exact JSON is given for `workspace_edit` and `import`, plus a `tree.depth` excerpt.
- [x] CHK010 Is "derived from the same definitions so it cannot drift" made objectively checkable? [Measurability, Spec §FR-005, SC-001, research.md R6] Yes, by the AST drift test and the completeness test (T006).

## Requirement Consistency — tool schema contract

- [x] CHK011 Are the schema constraints consistent with what MCP clients accept? [Consistency, Spec §FR-006, research.md R3] Yes; there are no top-level combinators, and per-action requirements live in descriptions.
- [x] CHK012 Do the `import` argument requirements in the contract agree with the handler's real behaviour? [Consistency, contracts/mcp-tool-schema.md vs data-model.md Appendix A] Yes, after the sweep: `source_id` is not required (an empty discard is a no-op), and `on_clash`/`passphrase` are used only by clashes and run. Earlier drafts that said otherwise were corrected.
- [x] CHK013 Is the existing action-selector contract preserved, so existing clients and scenarios stay valid? [Consistency, Spec §FR-004] Yes; `environment_reorder.feature` @listing still holds (T012).

## Error Contract & Validation Messaging

- [x] CHK014 Is the MCP error payload shape fully specified, including when `details` is absent? [Completeness, contracts/mcp-error.md] `{code, message, details?}`, with `details` omitted when empty.
- [x] CHK015 Is plain-English wording specified for every validation rule in use, with a fallback? [Completeness, Spec §FR-009, contracts/mcp-error.md] There are 15 rules with wording, and unknown rules fall back to `is not valid`. The rule inventory was taken from every `validate:` tag in `internal/`.
- [x] CHK016 Is the message format for several failing fields defined, including order and separator? [Clarity, Spec §FR-010] `invalid request: ` + `<field> <problem>` joined by `; `, in field order.
- [x] CHK017 Are field names in errors required to be the names the client sends, not internal request-DTO names? [Consistency, Spec §FR-008, research.md R7b] Yes. **Fixed during plan**: about 20 arguments are validated under other names, now mapped back via `validated_as`.
- [x] CHK018 Is it defined what happens to validation attributes that have no client-facing argument (internal `local_id`, `tab_id`)? [Edge Case, data-model.md Appendix A note 3] They are kept as-is, because the client cannot send them.
- [x] CHK019 Is the impact on the window's error contract specified, and is it backward compatible? [Consistency, Spec §FR-011, research.md R7] The shape is unchanged. Only the wording changes, and the attribute becomes a path for nested fields only. No window code reads validation attributes (verified).
- [x] CHK020 Is the treatment of non-validation errors that already carry details specified? [Coverage, Spec §FR-012] Messages are kept, and details pass through as `{field, problem}` with messages joined by `, `.
- [x] CHK021 Are the existing tests whose expected text changes enumerated, so "every existing scenario stays green" is unambiguous? [Measurability, Spec §SC-006, research.md R11] Yes: `validator/custom_test.go`, `import_detection_test.go`, and `postman_v3_workspace.feature` @nothing. **Fixed in analyze pass 1** (I2).

## Import of Untrusted Folder Contents

- [x] CHK022 Is the depth of loose-folder expansion bounded? [Clarity, Spec §Assumptions, FR-013] One level only; subfolders are skipped and reported.
- [x] CHK023 Are hidden files, symbolic links and other non-regular entries addressed? [Edge Case, Spec §Edge Cases, FR-013/FR-014] **Fixed during checklist**: hidden entries, symlinks and non-regular entries are ignored silently and never followed, matching the existing folder walk (`import_source.go` `readImportFolder` skips non-regular entries).
- [x] CHK024 Are the existing per-file size, encoding and archive limits stated to apply, and is their interaction with loose folders defined? [Completeness, Spec §FR-013, Edge Cases] **Fixed during checklist**: the initial folder walk enforces the 200 MB/UTF-8 limits on every document before the layout is known, so a violation fails the detect for any folder. Zip extraction refusals inside a loose folder become skipped entries.
- [x] CHK025 Is the folder that gets expanded unambiguous when Xenon re-roots into `postman/`? [Ambiguity, Spec §Edge Cases, data-model.md ImportInput] **Fixed during checklist**: the new `ImportInput.Path` holds the path as passed, and expansion lists it.
- [x] CHK026 Is the reporting of skipped entries specified exactly (code, severity, message format, reason text)? [Clarity, Spec §FR-014, contracts/import-detect.md] Yes: `skipped_part`, `warning`, `Skipped in <folder>: <name> (<reason>), …`.
- [x] CHK027 Is the failure for a folder with nothing importable specified with an exact code and message listing the accepted layouts? [Clarity, Spec §FR-015, contracts/import-detect.md] Yes, `IMP-00415` with the exact text.
- [x] CHK028 Is a limit on the number of files in a loose folder addressed? [Gap → resolved, Spec §Assumptions] **Fixed during checklist**: there is deliberately no count cap beyond the per-file and archive limits, matching the existing folder import, which has none.
- [x] CHK029 Is the cleanup of temp dirs created while expanding a loose folder (zips) specified on success and on failure? [Recovery, data-model.md, tasks T032] They join the source's `TempDirs` (removed on discard/expiry), and they are discarded on detect failure.
- [x] CHK030 Is it guaranteed that recognised layouts are not affected? [Consistency, Spec §FR-016] Yes; only `detectFolder`'s default branch changes, and the existing import suite is the guard (T034/T035).
- [x] CHK031 Is the top-level kind rule complete for one input, identical inputs, the same kind with different versions, and different kinds? [Completeness, Spec §FR-017, contracts/import-detect.md] Yes, and the unit cases are in T025.

## Secrets Safety

- [x] CHK032 Is it stated that schemas expose names, types and descriptions only, never values? [Completeness, Spec §Edge Cases, plan Constitution II] Yes.
- [x] CHK033 Is it stated that validation details never echo a submitted value (e.g. a passphrase)? [Completeness, Spec §FR-012] **Fixed during checklist**: details carry field names and rule parameters only.
- [x] CHK034 Has every existing source of error details been reviewed for secret values before they newly reach AI clients? [Coverage, Spec §FR-012] **Fixed during checklist**: all `errs.ErrorDetails{…}` constructions in `internal/` were reviewed. They carry ids, names, colours, fingerprints, line/column, provider error codes, next steps, setting ranges and secret *names*. An approval id (MCP-05409) gives a client nothing, because only the window resolves approvals. No value is exposed.
- [x] CHK035 Does the error text stay behind the existing AI-output leak guard? [Consistency, Spec §FR-012, contracts/mcp-error.md] Yes, `guardMCPText` runs over the whole encoded error.
- [x] CHK036 Are human-only arguments (`include_secrets`, `ai_access`) treated consistently in the schema? [Consistency, data-model.md Appendix A] `include_secrets` is declared with "refused over MCP: humans only". `ai_access` is not declared (research.md R4), because any presence of it is refused.

## Non-Functional, Dependencies & Assumptions

- [x] CHK037 Is the cost of building schemas addressed? [Non-Functional, plan Performance Goals] They are built once at tool registration, and `tools/list` costs nothing extra per call.
- [x] CHK038 Is the growth of `tools/list` from fully declared nested objects acknowledged as a decision? [Assumption, Spec §Assumptions] **Fixed during checklist**: recorded as a deliberate trade-off. **Revised during implementation**: the size collided with XEN-008's 24,000-character budget, and the user raised it to 36,000 (spec Clarifications).
- [x] CHK039 Are new dependencies avoided or justified? [Dependency, plan Technical Context, research.md R1] There is no new dependency; a stdlib reflector lives in `pkg/jsonschema`.
- [x] CHK040 Is the assumption that clients send declared types without coercion stated? [Assumption, Spec §Assumptions, contracts/mcp-tool-schema.md] Yes.
- [x] CHK041 Is stdio-bridge parity established from the code rather than assumed? [Dependency, Spec §Clarifications] Yes; the bridge only pumps JSON-RPC (`stdio_bridge.feature`).

## Notes

- All 41 items pass. Seven (CHK023–025, CHK028, CHK033–034, CHK038) failed on the first run and
  passed after the spec, data-model, contract and tasks fixes named in each item.
- No item depends on a decision outside this repo or outside the user's answers in
  spec.md Clarifications.
