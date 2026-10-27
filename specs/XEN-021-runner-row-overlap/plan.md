# Implementation Plan: Runner Result Rows Never Overlap When Expanded

**Branch**: `xen-021` | **Date**: 2026-10-27 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/XEN-021-runner-row-overlap/spec.md`

## Summary

The Runner's Results list gives every expanded row a fixed 240 px slot and lets the expansion's
parts shrink below their content, so the tests panel and the console are painted over each other
and the next row starts too early. The fix measures each rendered slot with the list
virtualizer's own measuring (`measureElement` + `data-index`, the 240 px kept only as the initial
estimate, the size-cache-clearing `measure()` effect removed), lets every slot's content set its
own height, makes the expansion's parts non-shrinkable, caps the details panel at 320 px with
internal scrolling, and makes that panel keyboard-focusable. Frontend only: one TSX file, one CSS
module, one test file. No Go changes.

## Technical Context

**Language/Version**: TypeScript 5 (strict), React 19, CSS Modules

**Primary Dependencies**: `@tanstack/react-virtual` 3.14.14 (already used; no new dependency)

**Storage**: N/A — no stored data or run format change ([data-model.md](data-model.md))

**Testing**: Vitest + Testing Library in jsdom (`frontend/vite.config.ts` `test` block); real CSS
injected with `injectStylesheets` from `frontend/src/test/stylesheets.ts` (research R6);
`make validate`

**Target Platform**: Wails v3 webview — WKWebView (macOS), WebView2 (Windows), WebKitGTK (Linux)

**Project Type**: desktop app (Go backend + React UI)

**Performance Goals**: no extra renders per frame — sizes come from one `ResizeObserver` the
virtualizer already owns; a 50-row live run re-places rows only when a slot's size changes

**Constraints**: collapsed rows (32 px) and iteration headings (28 px) look identical (FR-007);
`TestResultsTab` and the response viewer are untouched (FR-008); jsdom has no layout, so tests lock
the DOM/CSS contract and the visual result is checked by eye on macOS

**Scale/Scope**: 2 source files changed (`RunResults.tsx`, `RunResults.module.css`), 1 test file
extended (`RunResults.test.tsx`)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first** — ✅ no state, file or format change.
- [x] **II. Secrets** — ✅ console lines are shown exactly as today (already masked upstream); no
  new display, log or storage of values.
- [x] **III. Go owns logic** — ✅ pure presentation; no binding, disk or network access.
- [x] **IV. Architecture skills** — ✅ changes stay inside `features/runner/components/`, in the
  component and its colocated CSS module (`frontend-architecture`); the shared
  `features/response-viewer` `TestResultsTab` and `app/layout/runner-slots.tsx` are not touched;
  no new import crosses a boundary.
- [x] **V. BDD-first** — ✅ with justification (Complexity Tracking): no Go behaviour changes, so no
  Godog scenario can observe this. The red-first tests are Vitest cases (the constitution's suite
  includes "frontend lint + typecheck + Vitest"), written and proven red before the change.
  `/create-tests` writes **no `.feature` file** for this task; it writes the red Vitest cases T002,
  T004, T006 and T008 (tasks.md). Nothing beyond what `RunResults.test.tsx` already mocks is mocked.
- [x] **VI. Postman parity** — ✅ Postman's runner result rows expand to their content and scroll
  long content; nothing beyond that.
- [x] **VII. Cross-platform** — ✅ `ResizeObserver`, `max-height`, `flex: none`, `tabindex` and
  `:focus-visible` are supported by WebKit, WebKitGTK and WebView2; CI runs the suite on all three;
  by-eye check on macOS per Clarifications.

**Post-design re-check (after Phase 1)**: all seven still ✅ — no new dependency, no boundary
crossed, the V justification below stands.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-021-runner-row-overlap/
├── plan.md              # This file
├── research.md          # Phase 0: root cause, virtualizer measuring, intrinsic heights, panel, keyboard, jsdom
├── data-model.md        # Phase 1: sizes and the placement rule (no stored-data change)
├── quickstart.md        # Phase 1: automated + by-eye verification
├── contracts/
│   └── layout-contract.md  # Phase 1: DOM and CSS contract of the Results list
├── checklists/
│   ├── requirements.md
│   └── ui-layout.md
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
frontend/src/features/runner/components/
├── RunResults.tsx          # slots measured; measure() effect removed; expansion focusable group
├── RunResults.module.css   # .row/.iteration intrinsic heights; .expansion cap + non-shrinking children + focus ring
└── RunResults.test.tsx     # + DOM contract, CSS contract and keyboard cases
```

No file under `internal/`, `cmd/`, `pkg/`, `test/integration/`,
`frontend/src/features/response-viewer/` or `frontend/src/app/` changes. **No BDD feature file**:
nothing in Go observes CSS layout (Constitution V justification).

**Structure Decision**: frontend-only change in the component that owns the defect, per
`frontend-architecture` (feature component + colocated CSS module + colocated test).

### Design

Full rule list: [contracts/layout-contract.md](contracts/layout-contract.md).

**`RunResults.tsx`** (research R2, R5):
- In `RowResults`, delete the `useEffect(() => { virtualizer.measure() }, [expanded, virtualizer])`
  block and drop `useEffect` from the React import (that effect is its only use in the file).
- Each slot `<div role="listitem" className={styles.slot}>` gets `data-index={virtual.index}` and
  `ref={virtualizer.measureElement}`; its `style` becomes `{ transform: `translateY(${virtual.start}px)` }`
  (no `height`).
- `estimateSize`, `itemHeight`, `RESULT_ROW_HEIGHT`, `ITERATION_HEADING_HEIGHT` stay.
  `EXPANDED_ROW_HEIGHT` stays at 240; its doc comment becomes "First estimate of an expanded row
  before it is measured; the rendered row sets its real height."
- In `ResultRow`, the expansion `<div className={styles.expansion}>` gets `role="group"`,
  `aria-label={`Details of ${name}`}` and `tabIndex={0}`. Its children and their order are unchanged.

**`RunResults.module.css`** (research R3, R4, R5):
- `.row`: remove `height: 100%`.
- `.iteration`: `height: 100%` → `height: 28px`.
- `.expansion`: remove `flex: 1` and `min-height: 0`; add `max-height: 320px`; keep the rest.
- New `.expansion > * { flex: none; }`.
- Add `.expansion:focus-visible` to the module's existing focus selector group
  (`.chip:focus-visible, .chevron:focus-visible, .name:focus-visible { outline: none; box-shadow: var(--focus-ring); }`).

### Requirement → coverage map

| Requirement | Change | Test |
|---|---|---|
| FR-001, FR-002 | `.expansion > * { flex: none }`; order unchanged | T002 (+ existing expand case) |
| FR-003 | `.expansion { max-height: 320px; overflow: auto }` | T002 + quickstart step 3 |
| FR-004, FR-005 | measured slots, no inline height, `measure()` effect removed | T006 + quickstart steps 2, 5, 6 |
| FR-006 | same mechanism on refresh / past runs | T008 + quickstart steps 4, 6 |
| FR-007 | `.line` 32 px (unchanged), `.iteration` 28 px, estimates unchanged | T006 + quickstart step 7 |
| FR-008 | `TestResultsTab` untouched | T010 (`git diff` empty) + response-viewer tests |
| FR-009 | `role="group"`, `aria-label`, `tabIndex={0}`, focus ring | T004 + quickstart step 8 |
| SC-001–SC-005 | — | T011 (by eye, screenshots) |
| SC-006 | — | T012 (`make validate`) |

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| V. BDD-first: no Godog `.feature` for these requirements | The behaviour is CSS layout and list measuring inside the webview; no Go code is involved, so no step in `test/integration/steps/` can observe it | A Godog scenario would assert nothing real; Vitest is the constitution's frontend gate and is written red-first here exactly like a feature file would be. The issue (brienze1/xenon#34) itself names a Vitest class/style-contract check as the test. Precedent: XEN-010, XEN-015, XEN-019 |
