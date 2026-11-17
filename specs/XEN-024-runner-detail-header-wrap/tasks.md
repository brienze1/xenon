---

description: "Task list for XEN-024 — runner row detail keeps header names readable"
---

# Tasks: Runner Row Detail Keeps Header Names Readable

**Input**: Design documents from `/specs/XEN-024-runner-detail-header-wrap/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/layout-contract.md, quickstart.md

**Tests**: MANDATORY and red-first (constitution V). This task changes no Go code, so there is **no
Godog feature file** (plan.md → Complexity Tracking); the red-first tests are the Vitest cases
T002, T003, T009 and T011, each proven red before the code it locks is written.

**Organization**: Tasks are grouped by user story (US1 P1, US2 P2, US3 P3 from spec.md).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to

## Path Conventions

Frontend paths are relative to the repo root (`frontend/src/...`). Vitest commands run from
`frontend/`; paths passed to `injectStylesheets` are relative to `frontend/`.

---

## Phase 1: Setup

- [x] T001 Prepare the worktree for Vitest: run `npm ci` in `frontend/`, then `make bindings` from the repo root (fresh worktrees lack `frontend/node_modules` and the generated bindings)

---

## Phase 2: Foundational

No foundational work: no shared component, store, mapper or backend change is needed. User
stories start directly after Phase 1.

---

## Phase 3: User Story 1 — Read the sent headers of a runner row at a glance (Priority: P1) 🎯 MVP

**Goal**: The runner row details panel shows the sent headers as a labelled name/value list laid
out as a capped name column plus a value column (contract C1), so names no longer break one
character per line.

**Independent Test**: In `RunRowDetail.test.tsx`, a row that sent the 5 default headers renders a
`Request headers` list of terms/definitions with the C1 grid contract; by eye, the reported row
shows 5 header lines.

### Tests for User Story 1 (write first, prove red) ⚠️

- [x] T002 [US1] In `frontend/src/features/runner/components/RunRowDetail.test.tsx` add `it('shows the sent headers as a labelled list of names and values', …)`: render `RunRowDetail` (runId `R1`, `response` slot and `onClose` as in the existing cases) for `runRow(0, { sentRequest: { method: 'POST', url: 'https://api.test/campaigns/510876/briefings/7/guidelines', headers: [Content-Type: application/json, User-Agent: Xenon/0.1.14, Accept: */*, Accept-Encoding: gzip, deflate, br, Connection: keep-alive, Authorization: •••••], body: '{}' } })`; within the `Request` region assert: `getByLabelText('Request headers')` is a `DL` element; `getAllByRole('term').map(textContent)` equals the 6 names in sent order; each term's `nextElementSibling` is a `DD` whose text is the matching value (the `Authorization` definition is `•••••`); and `container.querySelector('table')` is `null`
- [x] T003 [US1] In the same file add `it('lays the headers out as a capped name column and a value column', …)`: call `injectStylesheets('src/features/runner/components/RunRowDetail.module.css')` (import from `@/test/stylesheets`), render the same row, and assert contract C1 from `specs/XEN-024-runner-detail-header-wrap/contracts/layout-contract.md` on the `Request headers` element and its first `dt`/`dd`: `display` is `grid`; `getPropertyValue('grid-template-columns').replace(/\s+/g, '')` is `fit-content(max(40%,16ch))minmax(0,1fr)`; `dt` and `dd` `overflow-wrap` are `anywhere`; `dt` `min-width` is `0` or `0px`; `dd` `margin-left` is `0` or `0px`
- [x] T004 [US1] From `frontend/` run `npx vitest run src/features/runner/components/RunRowDetail.test.tsx` and record that T002, T003 and T009 fail on the current `<table>` markup (red) while the two pre-existing cases pass

### Implementation for User Story 1

- [x] T005 [US1] In `frontend/src/features/runner/components/RunRowDetail.tsx` replace the `<table className={styles.headers}>…</table>` block with `<dl className={styles.headers} aria-label="Request headers">` containing, per header, a keyed `Fragment` (`key={`${header.key}-${index}`}`, import `Fragment` from `react`) of `<dt translate="no">{header.key}</dt><dd translate="no">{header.value}</dd>`; keep the `row.sentRequest.headers.length > 0` guard; remove the visually hidden caption (the `aria-label` replaces it)
- [x] T006 [US1] In `frontend/src/features/runner/components/RunRowDetail.module.css` replace the `.headers`, `.headers th, .headers td`, `.headers th` and `.headers td` rules with: `.headers { display: grid; grid-template-columns: fit-content(max(40%, 16ch)) minmax(0, 1fr); column-gap: var(--space-3); row-gap: var(--space-1); margin: 0; font-family: var(--font-mono); font-size: var(--font-size-sm); }`, `.headers dt { min-width: 0; color: var(--color-text-muted); overflow-wrap: anywhere; }`, `.headers dd { min-width: 0; margin: 0; color: var(--color-text); overflow-wrap: anywhere; }`; leave every other rule untouched
- [x] T007 [US1] In `frontend/src/features/runner/components/RunResults.test.tsx` (line 182) replace `within(detail).getByRole('rowheader', { name: 'Authorization' }).nextSibling` with `within(detail).getByText('Authorization', { selector: 'dt' }).nextElementSibling` (still `toHaveTextContent('•••••')`) — Testing Library computes no accessible name for the `term` role, so `getByRole('term', { name })` cannot be used (verified by probe)
- [x] T008 [US1] From `frontend/` run `npx vitest run src/features/runner/components/` and confirm T002, T003 and every pre-existing runner case are green

**Checkpoint**: US1 is shippable on its own — the reported defect is fixed.

---

## Phase 4: User Story 2 — Long names and values stay inside the panel (Priority: P2)

**Goal**: Long names wrap inside the capped column (never ellipsized, per the product owner) and
long values wrap inside their cell; the panel never scrolls sideways.

**Independent Test**: A row with a 61-character name and a 400-character unbroken value renders
both in full, with no truncation styling; by eye, no horizontal scroll.

### Tests for User Story 2 ⚠️

- [x] T009 [US2] In `frontend/src/features/runner/components/RunRowDetail.test.tsx` add `it('shows long header names and values in full, wrapping instead of truncating', …)`: with `RunRowDetail.module.css` injected, render a row whose headers are `X-Very-Long-Custom-Correlation-Header-Name-For-Tracing-Abcdef` (61 chars) → `'a'.repeat(400)`; assert the term's text is the full 61-character name and the definition's text is the full 400 characters; for both `dt` and `dd` assert computed `white-space` is not `nowrap`, `text-overflow` is not `ellipsis`, and `overflow-wrap` is `anywhere`; and neither element has a `title` attribute. Write it right after T003 and run it in T004 with the other red cases, before T005 (on the current markup there is no `dt`, so it fails)

### Implementation for User Story 2

- [x] T010 [US2] No extra production code: the `fit-content(max(40%, 16ch))` column and `overflow-wrap: anywhere` on `dt`/`dd` from T006 already satisfy FR-003–FR-005. From `frontend/` run `npx vitest run src/features/runner/components/RunRowDetail.test.tsx` and confirm T009 is green

**Checkpoint**: US1 + US2 hold together.

---

## Phase 5: User Story 3 — API documentation parameter names stay readable (Priority: P3)

**Goal**: The documentation Parameters (and Body fields) table keeps each name on one line up to 32
characters, wrapped inside a 32-character block past that (contract C2).

**Independent Test**: `ApiDocView.test.tsx` renders an opened operation whose Name cells satisfy C2;
by eye, `Accept-Encoding required` sits on one line next to a 400-character example.

### Tests for User Story 3 (write first, prove red) ⚠️

- [x] T011 [P] [US3] Create `frontend/src/features/documentation/components/ApiDocView.test.tsx`: import `generatedDocument` from `@/features/documentation/test-fixtures`, `injectStylesheets` from `@/test/stylesheets`; in `it('keeps parameter names on one line inside a block capped at 32 characters', …)` inject `src/features/documentation/components/ApiDocView.module.css`, render `<ApiDocView kind="collection" document={{ ...generatedDocument, operations: [{ ...generatedDocument.operations[0]!, parameters: [{ name: 'Accept-Encoding', in: 'header', description: 'Encodings', example: 'x'.repeat(400), required: true, variable: false }, { name: 'a-very-long-parameter-name-that-exceeds-the-cap-xyz', in: 'query', description: '', example: '1', required: false, variable: false }] }] }} />`, open the operation with `userEvent.click(screen.getByText('/invoices'))`, then within the `List invoices parameters` region assert for the first body row: the Name `td` contains a `span` whose text is `Accept-Encoding`, followed in the same `td` by the ` required` marker; the `td` computed `white-space` is `nowrap` and `overflow-wrap` is not `anywhere`; the span's computed `display` is `inline-block`, `width` is `max-content`, `max-width` is `256px` (jsdom's value for `32ch`), `white-space` is `normal`, `overflow-wrap` is `anywhere`; the Example `td` keeps `overflow-wrap: anywhere`; and the second row's span text is the full 51-character name
- [x] T012 [US3] From `frontend/` run `npx vitest run src/features/documentation/components/ApiDocView.test.tsx` and record that T011 fails on the current markup (red)

### Implementation for User Story 3

- [x] T013 [US3] In `frontend/src/features/documentation/components/ApiDocView.tsx` (`Parameters`) change the Name cell to `<td className={styles.nameCell} translate="no"><span className={styles.parameterName}>{parameter.name}</span>{parameter.required ? <span className={styles.required}> required</span> : null}</td>`; the Example cell keeps `className={styles.code}`
- [x] T014 [US3] In `frontend/src/features/documentation/components/ApiDocView.module.css` add, after `.code`: `.nameCell { font-family: var(--font-mono); white-space: nowrap; }` and `.parameterName { display: inline-block; width: max-content; max-width: 32ch; white-space: normal; overflow-wrap: anywhere; vertical-align: top; }`; leave `.code` and every other rule unchanged
- [x] T015 [US3] From `frontend/` run `npx vitest run src/features/documentation/` and confirm T011 and every pre-existing documentation case (including `DocumentationPanel.test.tsx`'s `statusqueryInvoice statusopen` text assertion) are green

**Checkpoint**: all three stories hold.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T016 By-eye check on macOS per `specs/XEN-024-runner-detail-header-wrap/quickstart.md` (scratchpad Vite harness outside the repo): scenarios 1–3 at 280 px (default layout) and 420 px (right panel hidden), before (`origin/main` copies) and after; confirm 5 header lines with every default name on one line at both widths, no single-letter lines, the start of the request body visible without scrolling in the 280 px × 800 px case (SC-002), `scrollWidth === clientWidth` on the panel and on the docs view; save the PNGs to `specs/XEN-024-runner-detail-header-wrap/screenshots/` (`before-long-header.png`, `after-long-header.png`, `after-reported-row.png`, `before-docs-parameters.png`, `after-docs-parameters.png` — the before shot of the reported row needs the long value to reproduce, research R7) and delete any `.playwright-mcp/` folder
- [ ] T017 Confirm FR-009 holds: `git diff --stat origin/main -- frontend/` lists only `RunRowDetail.tsx`, `RunRowDetail.module.css`, `RunRowDetail.test.tsx`, `RunResults.test.tsx`, `ApiDocView.tsx`, `ApiDocView.module.css`, `ApiDocView.test.tsx`
- [ ] T018 Run `make validate` from the repo root (lint, typecheck, Vitest, Go tests and BDD suite) and keep it green
- [ ] T019 Run `/refactor-diff` until it reports clean; never commit `REFACTORING_REPORT.md`

---

## Dependencies & Execution Order

- **Phase 1 (T001)** → everything.
- **US1 (T002–T008)**: T002 → T003 → (T009) → T004 (red) → T005 → T006 → T007 → T008. MVP.
- **US2 (T009–T010)**: T009 is written right after T003 and proven red in T004 (before T005);
  T010 runs after T006.
- **US3 (T011–T015)**: independent of US1/US2 (different feature folder); T011 → T012 (red) →
  T013 → T014 → T015.
- **Polish (T016–T019)**: after all stories.

## Parallel Opportunities

- T011 (US3 test, `documentation/`) can be written in parallel with T002/T003 (US1 tests,
  `runner/`) — different files.
- T013/T014 (US3) can run in parallel with T005–T007 (US1).

### Parallel example

```text
Developer A: T002 → T003 → T009 → T004 → T005 → T006 → T007 → T008 → T010
Developer B: T011 → T012 → T013 → T014 → T015
Then: T016 → T017 → T018 → T019
```

## Implementation Strategy

1. **MVP**: Phase 1 + US1 — fixes the reported panel.
2. **Increment**: US2 locks the no-truncation decision and the long-content behaviour.
3. **Increment**: US3 fixes the same pattern in the documentation view.
4. **Polish**: by-eye screenshots, scope check, `make validate`, `/refactor-diff`.
