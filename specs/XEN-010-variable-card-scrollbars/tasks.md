---

description: "Task list for XEN-010 — variable card value field without scrollbars"
---

# Tasks: Variable Card Value Field Without Scrollbars

**Input**: Design documents from `specs/XEN-010-variable-card-scrollbars/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/use-fit-height.md](contracts/use-fit-height.md),
[quickstart.md](quickstart.md)

**Tests**: MANDATORY and red-first (constitution V). This feature changes no Go behaviour, so there
is no Godog `.feature` (plan → Complexity Tracking); the red-first tests are Vitest cases. Each
test task ends by running the tests and confirming the new cases **fail** before the matching
implementation task starts.

**Organization**: grouped by user story (spec.md: US1 P1 undefined card, US2 P2 defined card,
US3 P3 zoom/scrollbar setting).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel (different files, no dependency on an unfinished task)
- **[Story]**: US1 / US2 / US3

## Shared test fixtures (used by T002, T004, T007, T009)

- **Metrics**: the card box's metrics, injected because Vitest does not load CSS-module styles
  (research R4): append `<style>.value { line-height: 18px; padding: 4px 8px; border: 1px solid; }</style>`
  to `document.head` in `beforeEach`, remove it in `afterEach`. (`classNameStrategy: 'non-scoped'`
  keeps the class name `value`.)
- **Layout stub**: in `beforeEach`, `Object.defineProperty(HTMLTextAreaElement.prototype,
  'scrollHeight', { configurable: true, get() { return 8 + 18 * (this.value === '' ? 1 :
  this.value.split('\n').length) } })`; in `afterEach`, `delete` it so jsdom's own (0) returns.
  Line breaks stand in for wrapped lines.
- **Expected card heights** (data-model.md): 1 line → `28px`; 3 lines → `64px`; 6 lines → `118px`;
  9 lines → `118px` with `overflowY: 'auto'`.

---

## Phase 1: Setup

**Purpose**: a known-green baseline

- [ ] T001 In `frontend/`, run `npm ci`, then `npx vitest run src/features/request-builder/components/VariableCard.test.tsx src/features/environments/components/VariableTable.test.tsx` and confirm both files pass on the untouched code (baseline for SC-005/SC-006).

---

## Phase 2: Foundational — the shared hook (blocks all stories)

**Purpose**: `useFitHeight` per [contracts/use-fit-height.md](contracts/use-fit-height.md)

- [ ] T002 Write `frontend/src/shared/hooks/use-fit-height.test.tsx`: a tiny test component renders `<textarea ref={ref} rows={1} value={value} onChange={…} style={{ lineHeight: '18px', padding: '4px 8px', border: '1px solid' }} />` and calls `useFitHeight(ref, value, maxLines)`; use the layout stub above. Cases (contract "Test obligations" 1–5): (1) no cap, one line → `style.height === '28px'`, `overflowY === 'hidden'`, `overflowX === 'hidden'`; (2) cap 6 with 6 lines → `118px` hidden; with a stub returning 117 → `119px` hidden (1 px tolerance); 9 lines → `118px`, `overflowY === 'auto'`, `overflowX === 'hidden'`; (3) rerender 1 → 3 → 1 lines → `28px` → `64px` → `28px`; (4) `lineHeight: 'normal'` with cap 6 and 9 lines → `170 + 2 = 172px`, hidden (uncapped, never `NaN`); (5) a component whose ref is never attached calls the hook without throwing. Run `npx vitest run src/shared/hooks/use-fit-height.test.tsx` and confirm it fails (module not found).
- [ ] T003 Create `frontend/src/shared/hooks/use-fit-height.ts` exporting `useFitHeight(ref: RefObject<HTMLTextAreaElement | null>, value: string, maxLines?: number): void` exactly as plan.md → Design (layout effect on `[ref, value, maxLines]`; `fit()` = height auto → `scrollHeight`, computed `lineHeight`/`padding*`/`border*Width` via `parseFloat`, cap only when `maxLines !== undefined && Number.isFinite(lineHeight)`, `capped = content - cap > 1`, inline `height`/`overflowX: 'hidden'`/`overflowY`; `ResizeObserver` on the node when defined, disconnected on cleanup; early return when `ref.current` is null). One short JSDoc line on the export ("Fits a textarea's height to its text, up to `maxLines` lines, and only then lets it scroll — vertically."). Import only from `react`. Run T002's file → green.

**Checkpoint**: the hook is green on its own; no consumer uses it yet.

---

## Phase 3: User Story 1 — Type a value for an undefined variable (P1) 🎯 MVP

**Goal**: the "is not defined" card's box fits its text with no scrollbars up to 6 lines.

**Independent test**: T004's cases pass; quickstart §3 steps 1–4.

- [ ] T004 [US1] In `frontend/src/features/request-builder/components/VariableCard.test.tsx`, add the shared fixtures (metrics + layout stub) to the existing `beforeEach`/`afterEach`, and a nested `describe('value box sizing')` with undefined-card cases (render `inspect({ environment: 'env:e1' })` + `<VariableCard tabId={TAB} name="authPrvateUrl" close={vi.fn()} />`, box = `getByRole('textbox', { name: 'Value of authPrvateUrl' })`): (a) empty → `style.height === '28px'`, `overflowY === 'hidden'`, `overflowX === 'hidden'`, `rows === 1`; (b) `userEvent.type` `https://auth.dev` → still `28px`, hidden; (c) type `a{Shift>}{Enter}{/Shift}b{Shift>}{Enter}{/Shift}c` (3 lines) → `64px`, hidden; (d) 9 lines → `118px`, `overflowY === 'auto'`, `overflowX === 'hidden'`; (e) from 3 lines, `userEvent.clear` → `28px`. Run the file and confirm (a)–(e) fail while the three existing cases still pass.
- [ ] T005 [US1] In `frontend/src/features/request-builder/components/VariableCard.module.css` `.value`: add `display: block;` and `overflow: hidden;`, change `line-height: 1.5;` to `line-height: 18px;`; keep `resize: none`, `overflow-wrap: anywhere`, padding, border, background, colour and font lines as they are.
- [ ] T006 [US1] In `frontend/src/features/request-builder/components/VariableCard.tsx` `NewVariable`: add `const valueRef = useRef<HTMLTextAreaElement>(null)`, call `useFitHeight(valueRef, value, MAX_LINES)` (import from `@/shared/hooks/use-fit-height`), and on its textarea replace `rows={rowsFor(value)}` with `ref={valueRef}` and `rows={1}`. Add `const MAX_LINES = 6` (keep `MAX_ROWS`/`rowsFor` for now — `DefinedVariable` still uses them until T008). Run the card tests → T004 (a)–(e) and the existing cases green.

**Checkpoint**: US1 delivered — the reported defect is fixed for the undefined card.

---

## Phase 4: User Story 2 — Edit a defined variable's value in place (P2)

**Goal**: the defined card's box behaves exactly like US1's.

**Independent test**: T007's cases pass; quickstart §3 step 5.

- [ ] T007 [US2] In `VariableCard.test.tsx` `describe('value box sizing')`, add defined-card cases using the existing `usage()` helper and `inspect({ environment: 'env:e1' }, [usage(…)])`, box = `getByRole('textbox', { name: 'Value of creatorAdsUrl' })`: (a) `displayValue: 'https://ads.dev'` → `28px`, hidden, `overflowX` hidden; (b) `displayValue: 'a\nb\nc'` → opens at `64px`, hidden; (c) `secret: true, displayValue: ''` (status `resolved`) → `28px`, hidden, placeholder `Secret — type to replace…`; (d) `displayValue` of 9 lines → `118px`, `overflowY === 'auto'`; (e) `status: 'locked_secret', secret: true` → `queryByRole('textbox')` is null and the "Unlock the vault to use this secret" button renders (hook no-op). Run and confirm (a)–(d) fail; (e) passes already and must keep passing.
- [ ] T008 [US2] In `VariableCard.tsx` `DefinedVariable`: add `const valueRef = useRef<HTMLTextAreaElement>(null)` and `useFitHeight(valueRef, draft, MAX_LINES)` (unconditional, before the return; the ref stays null on the locked branch); on its textarea replace `rows={rowsFor(draft)}` with `ref={valueRef}` and `rows={1}`. Then delete `CHARS_PER_ROW`, `MAX_ROWS` and `rowsFor`, moving `rowsFor`'s doc comment ("Grows the value box with its text, like Postman's, up to a few lines.") onto `MAX_LINES`. Run the card tests → all green.

**Checkpoint**: both card forms fixed; `rowsFor` is gone.

---

## Phase 5: User Story 3 — The same at any zoom level and scrollbar setting (P3)

**Goal**: no scrollbar at exactly 6 lines when rounding reads a pixel over (fractional zoom), and
the box looks the same with either scrollbar setting.

**Independent test**: T009 passes; quickstart §3 step 6.

- [ ] T009 [US3] In `VariableCard.test.tsx` `describe('value box sizing')`, add: with the layout stub overridden to return `117` for a 6-line value (the 1 px rounding at a fractional zoom), the undefined card's box is `119px` with `overflowY === 'hidden'`. Run it: it passes with T003's tolerance (the hook's own tolerance case in T002 was the red-first proof); if it fails, fix `use-fit-height.ts` — not the test.

**Checkpoint**: all three stories covered by tests.

---

## Phase 6: Polish & Cross-Cutting

- [ ] T010 [P] In `frontend/src/features/environments/components/WrapTextCell.tsx`, replace the `useLayoutEffect` block (and its comment) with `useFitHeight(ref, value)` from `@/shared/hooks/use-fit-height`; drop the now-unused `useLayoutEffect` import; leave `WrapTextCell.module.css` and the JSX unchanged. Run `npx vitest run src/features/environments/components/VariableTable.test.tsx` → passes unchanged (FR-012, SC-006).
- [ ] T011 Run `make validate` from the repo root → green (eslint, typecheck, all Vitest, Go gates untouched).
- [ ] T012 Packaged macOS by-eye check per [quickstart.md](quickstart.md) §3: note the current `AppleShowScrollBars`, set it to `Always`, `make package`, open `bin/xenon.app`, walk steps 1–7 (both card forms, secret placeholder, zoom 67%/100%/200%, variable table incl. window resize), take screenshots of steps 1 and 5 for the implementation report, then restore the previous setting.

---

## Dependencies & Execution Order

- T001 → T002 → T003 (hook) → every story.
- US1: T004 → T005, T006 (T005 and T006 touch different files but T006's run needs T005; do T005 first).
- US2: T007 → T008; needs T006 done (`MAX_LINES` exists, same file).
- US3: T009 after T003 (and T006 for the card under test).
- Polish: T010 after T003 (independent of US1–US3, different feature); T011 after T010 and all
  stories; T012 after T011.

```text
T001 → T002 → T003 ─┬→ T004 → T005 → T006 → T007 → T008 → T009 ─┬→ T011 → T012
                    └→ T010 ─────────────────────────────────────┘
```

## Parallel Opportunities

- T010 (environments feature) can run alongside T004–T009 (request-builder feature) once T003 is
  green.
- Within a story, tasks share `VariableCard.tsx`/`VariableCard.test.tsx`, so they are sequential.

## Implementation Strategy

- **MVP**: Phases 1–3 (T001–T006) fix the reported undefined-variable card.
- Then US2 (T007–T008) and US3 (T009), then T010–T012. Push only after T011 is green and T012
  has been checked (no PR — `/implement-task` pushes the branch).

## Requirement Coverage

| Requirement | Tasks |
|---|---|
| FR-001 | T002, T003, T004(d), T005 |
| FR-002 | T002, T004(a–c), T007(a–c), T009 |
| FR-003, FR-005 | T003, T006, T008 |
| FR-004 | T004(b,c,e), T007(b) |
| FR-006 | T002, T004(d), T007(d) |
| FR-007 | T007(c), T012 |
| FR-008 | T006, T007, T008 |
| FR-009 | T004/T007 keep the existing cases green, T011 |
| FR-010 | T002, T004, T007, T009 |
| FR-011 | T011, T012 |
| FR-012 | T003, T010 |
| SC-001–SC-004 | T012 (by eye), T004/T007/T009 (rules) |
| SC-005 | T001, T004, T008 |
| SC-006 | T001, T010 |
