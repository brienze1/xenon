# Implementation Plan: Variable Card Value Field Without Scrollbars

**Branch**: `xen-010` | **Date**: 2026-10-15 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/XEN-010-variable-card-scrollbars/spec.md`

## Summary

The variable card's value box guesses its height from a character count (`rowsFor`: 36
characters a line, at most 6) and inherits the browser's `overflow: auto`, so any mismatch
between the guess and the real layout draws scrollbars — and with permanent scrollbars a vertical
one forces a horizontal one. The fix measures instead of guessing: a new domain-free hook,
`useFitHeight`, sets the box's height from its laid-out content (`scrollHeight`), caps it at 6
lines, and owns the overflow — sideways never, vertically only past the cap. The hook replaces
the ad-hoc fitting in the variable table's `WrapTextCell` (same behaviour, no cap), as decided in
Clarifications. Frontend only; no Go changes.

## Technical Context

**Language/Version**: TypeScript 5 (strict), React 19, CSS Modules

**Primary Dependencies**: none new (React `useLayoutEffect`, the browser's `ResizeObserver`)

**Storage**: N/A (no data changes)

**Testing**: Vitest 5 + Testing Library in jsdom 30 (`frontend/vitest.config.ts`); `make validate`
runs `frontend-check` (eslint, `tsc --noEmit`, `vitest run`)

**Target Platform**: Wails v3 webview — WKWebView on macOS, WebView2 on Windows, WebKitGTK on Linux

**Project Type**: desktop app (Go backend + React UI)

**Performance Goals**: the box fits on the first paint (layout effect, before paint) and on every
keystroke with one reflow; no visible jump when the card opens

**Constraints**: works under CSS `zoom` (the app's View → Zoom, 0.67–2) and with permanent
scrollbars; jsdom has no layout, so tests stub `scrollHeight` and inject the box's metrics as a
stylesheet (research R4)

**Scale/Scope**: 1 new hook + test, 2 components changed (+ 1 stylesheet), 1 test file extended

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first** — ✅ no state, storage or network involved.
- [x] **II. Secrets** — ✅ the box never shows a secret's value (placeholder only), unchanged; no
  value reaches logs or fixtures (tests use fake URLs).
- [x] **III. Go owns logic** — ✅ pure presentation; no disk/network access, no bindings touched.
- [x] **IV. Architecture skills** — ✅ `useFitHeight` is domain-free, used by two features
  (`request-builder`, `environments`), so it goes in `shared/hooks/` (frontend-architecture:
  "a generic, domain-free" piece shared by features moves to `shared/`; rule 2 forbids one feature
  importing another's hooks). Consumers stay in their features.
- [x] **V. BDD-first** — ✅ with justification (Complexity Tracking): no Go behaviour changes, so
  no Godog scenario can observe this; the red-first tests are Vitest cases (the constitution's
  suite includes "frontend lint + typecheck + Vitest"), written and proven red before the fix.
  `/create-tests` writes no `.feature` for this task; it writes the red Vitest cases T002, T004, T007 and T009.
  Only the browser layout (`scrollHeight`) is faked.
- [x] **VI. Postman parity** — ✅ Postman's card value box grows with its text and shows no
  scrollbars on a short value; this restores that. Nothing beyond the PRD.
- [x] **VII. Cross-platform** — ✅ the defect is worst where scrollbars are permanent (Windows,
  Linux by default; macOS "Always"); measuring + `overflow: hidden` is engine-independent
  (WebKit, WebView2, WebKitGTK).

**Post-design re-check (after Phase 1)**: all seven still ✅ — the design adds one shared hook and
no dependency, keeps the V justification below.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-010-variable-card-scrollbars/
├── plan.md              # This file
├── research.md          # Phase 0: root cause, measuring, zoom, jsdom testability
├── data-model.md        # Phase 1: the value box's sizing state
├── quickstart.md        # Phase 1: how to verify (tests + packaged macOS app)
├── contracts/
│   └── use-fit-height.md  # Phase 1: the shared hook's contract
├── checklists/
│   ├── requirements.md
│   └── ui.md
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
frontend/src/
├── shared/hooks/
│   ├── use-fit-height.ts          # NEW — useFitHeight(ref, value, maxLines?)
│   └── use-fit-height.test.tsx    # NEW — the hook's own rules (fit, cap, tolerance, overflow, refit)
├── features/request-builder/components/
│   ├── VariableCard.tsx           # rowsFor/CHARS_PER_ROW removed; both boxes rows={1} + useFitHeight(…, 6)
│   ├── VariableCard.module.css    # .value: display block, line-height 18px, overflow hidden
│   └── VariableCard.test.tsx      # + sizing/overflow cases for both card forms (FR-010)
└── features/environments/components/
    └── WrapTextCell.tsx           # its layout effect replaced by useFitHeight(ref, value)
```

No file under `internal/`, `cmd/`, `pkg/` or `test/integration/` changes. No BDD feature file:
nothing in Go observes textarea sizing (Constitution V justification).

**Structure Decision**: frontend-only change in the existing `frontend/src` layout. The shared
hook sits next to `use-hotkey.ts` and `use-window-height.ts` in `shared/hooks/` (kebab-case file,
`use` prefix, colocated test — frontend-architecture naming). `shared/` imports nothing from
`features/`, so the hook takes a ref and plain numbers.

### Design

**`useFitHeight(ref, value, maxLines?)`** (full contract: [contracts/use-fit-height.md](contracts/use-fit-height.md))

In a `useLayoutEffect` keyed on `value` and `maxLines`, and again from a `ResizeObserver` on the
node (width changes rewrap the text; skipped when `ResizeObserver` is undefined, as in jsdom):

1. `style.height = 'auto'`, then read `content = node.scrollHeight` (content + vertical padding).
2. Read the computed `lineHeight`, `paddingTop/Bottom`, `borderTop/BottomWidth` (px).
3. `cap = maxLines * lineHeight + paddingY` when `maxLines` is given and `lineHeight` is a
   number; otherwise no cap.
4. `capped = cap !== undefined && content - cap > 1` (1 px tolerance: `scrollHeight` is rounded
   to whole pixels, so at exactly `maxLines` lines under a fractional zoom it can read 1 px over).
5. `style.height = (capped ? cap : content) + borderY + 'px'` — `box-sizing: border-box` is
   global (`shared/styles/global.css`), so the border is added back.
6. `style.overflowX = 'hidden'` always; `style.overflowY = capped ? 'auto' : 'hidden'`.

**`VariableCard`**: delete `CHARS_PER_ROW` and `rowsFor`; keep `MAX_ROWS = 6`, renamed
`MAX_LINES` with its comment ("Grows the value box with its text, like Postman's, up to a few
lines"). Both `DefinedVariable` and `NewVariable` get a `useRef<HTMLTextAreaElement>(null)`,
`rows={1}`, `ref`, and `useFitHeight(ref, draft|value, MAX_LINES)`. The locked-secret branch
renders no textarea; the defined box's ref is attached only when the textarea renders (the hook
returns early on a null ref). Nothing else in the card changes (FR-009).

**`VariableCard.module.css` `.value`**: add `display: block`, `overflow: hidden` (the stylesheet
default before the hook's first run, and the rule that the hook's inline styles refine), and
change `line-height: 1.5` to `line-height: 18px` (the same 18 px at 12 px text, now a pixel value
like `WrapTextCell.module.css`, so the line height the hook reads is exact). Keep `resize: none`,
`overflow-wrap: anywhere` (breaks an unbroken token anywhere — edge case "one unbroken long
token"), padding, border and font.

**`WrapTextCell`**: replace its `useLayoutEffect` (height auto → `scrollHeight`, `ResizeObserver`
refit) with `useFitHeight(ref, value)`. Its stylesheet already has `overflow: hidden`, `border: 0`
and `line-height: 18px`; with no cap and no border the hook computes the same height it does
today, and the inline `overflowX/Y: hidden` match the stylesheet — no visible change (FR-012).

### Root cause → requirement map

| Requirement | Covered by |
|---|---|
| FR-001 no sideways scrollbar, wrap | `overflowX: hidden` (hook) + `overflow-wrap: anywhere` (CSS) |
| FR-002 no vertical scrollbar ≤ 6 lines | `overflowY: hidden` until `content - cap > 1` |
| FR-003 / FR-005 height = measured content | hook steps 1–5; `rowsFor` deleted |
| FR-004 follows edits, right on first paint | layout effect on `value`; `ResizeObserver` on width |
| FR-006 stop at 6 lines, scroll vertically | `cap`, `overflowY: auto` when capped |
| FR-007 placeholders visible | longest is 37 chars, 39 fit (spec Clarifications) at every zoom |
| FR-008 both card forms | both textareas use the hook |
| FR-009 nothing else changes | only sizing props/CSS touched; existing card tests unchanged |
| FR-010 unit test | T004, T007, T009 (card), T002 (hook) |
| FR-011 by-eye check + validate | T011, T012 |
| FR-012 shared, table unchanged | T010; existing `VariableTable.test.tsx` unchanged |

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| V. BDD-first: no Godog `.feature` for these requirements | The behaviour is textarea sizing inside the webview; no Go code is involved, so no step in `test/integration/steps/` can observe it | A Godog scenario would have to assert nothing real (or a fake), which is worse than none; Vitest is the constitution's frontend gate and is written red-first here exactly like a feature file would be. The issue's own acceptance (brienze1/xenon#7, written by the product owner) names a Vitest for `VariableCard` as the test |
