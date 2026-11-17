# Implementation Plan: Runner Row Detail Keeps Header Names Readable

**Branch**: `xen-024` | **Date**: 2026-11-17 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/XEN-024-runner-detail-header-wrap/spec.md`

## Summary

The runner row details panel draws the sent headers as an automatically sized `<table>` whose
cells all carry `overflow-wrap: anywhere`, so every cell's min-content width is one character and
the table algorithm squeezes the name column to one letter. The fix replaces that table with a
labelled name/value list (`<dl aria-label="Request headers">`, `dt`/`dd` pairs as direct children)
laid out as one CSS grid with columns `fit-content(max(40%, 16ch)) minmax(0, 1fr)`: the name column
is as wide as the longest name up to the larger of 40% of the list and 16 characters, values take the rest, and both wrap only inside their
own cell. The API documentation view's Parameters table (four columns, must stay a table) gets the
same guarantee by wrapping each name in an inline-block capped at 32 characters
(`width: max-content; max-width: 32ch`) so the table can never squeeze it below its own width.
Frontend only; no Go, data or Wails binding change.

## Technical Context

**Language/Version**: TypeScript 5 (strict), React 19, CSS Modules

**Primary Dependencies**: none new (React, Testing Library, Vitest already present)

**Storage**: N/A — no stored data or run format change ([data-model.md](data-model.md))

**Testing**: Vitest + Testing Library in jsdom (`frontend/vite.config.ts` `test` block); real CSS
injected with `injectStylesheets` from `frontend/src/test/stylesheets.ts` so `getComputedStyle`
reads the layout contract (verified in research R4: jsdom 30.1.2 returns `display`,
`grid-template-columns`, `overflow-wrap`, `white-space`, `width` and `max-width` from injected
rules); by-eye check in a scratchpad Vite harness; `make validate`

**Target Platform**: Wails v3 webview — WKWebView (macOS 12+, `build/darwin/Info.plist`
`LSMinimumSystemVersion`), WebView2 (Windows), WebKitGTK (Linux)

**Project Type**: desktop app (Go backend + React UI)

**Performance Goals**: N/A — static layout; no extra renders, no measuring, no JS layout code

**Constraints**: CSS features must work on the oldest supported webview (Safari 15-era WebKit on
macOS 12): grid `fit-content()` (Safari 10.1+), CSS `max()` (Safari 11.1+), `max-content` keyword
(Safari 11+) and `ch` units are allowed; container query units (Safari 16+) are not used. jsdom has no layout, so tests lock
the DOM/CSS contract and the visual result is checked by eye on macOS.

**Scale/Scope**: 2 components (`RunRowDetail`, `ApiDocView`), their 2 CSS modules, 2 test files
(one new), 1 assertion update in `RunResults.test.tsx`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first** — ✅ no state, file or network change.
- [x] **II. Secrets** — ✅ masked values are rendered exactly as today (FR-008); test fixtures use
  dummy header values only (`Bearer •••••`, `Xenon/0.1.14`).
- [x] **III. Go owns logic** — ✅ no disk/network access added; pure presentation change.
- [x] **IV. Architecture skills** — ✅ changes stay inside `features/runner/components/` and
  `features/documentation/components/` (component + colocated CSS module + colocated test, per
  `frontend-architecture`); no cross-feature imports added; `shared/` untouched.
- [x] **V. BDD-first** — ✅ with justification (Complexity Tracking): no Go behaviour changes, so no
  Godog scenario can observe this. The red-first tests are Vitest cases (the constitution's suite
  includes "frontend lint + typecheck + Vitest"), written and proven red before the change.
  `/create-tests` writes **no `.feature` file** for this task; it writes the red Vitest cases
  T002, T003, T009 and T011 in tasks.md.
- [x] **VI. Postman parity** — ✅ bug fix of an existing view; nothing beyond the PRD.
- [x] **VII. Cross-platform** — ✅ only CSS features available in every shipped webview (research
  R2/R3); by-eye on macOS plus CI's quality gate on macOS, Windows and Linux.

**Post-design re-check (after Phase 1)**: all seven gates still ✅ — the design adds no state, no
backend code, no new dependency and no CSS feature newer than the oldest supported webview.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-024-runner-detail-header-wrap/
├── plan.md              # This file
├── research.md          # Phase 0 — cause, layout choices, jsdom capabilities
├── data-model.md        # Phase 1 — no data change; displayed shapes
├── quickstart.md        # Phase 1 — how to validate (Vitest + by-eye)
├── contracts/
│   └── layout-contract.md   # Phase 1 — DOM + CSS contract the tests lock
├── checklists/
│   ├── requirements.md
│   └── ui-layout.md     # speckit-checklist output
├── screenshots/         # by-eye before/after (added by /implement-task)
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
frontend/src/features/runner/components/
├── RunRowDetail.tsx          # <table> of headers → <dl aria-label="Request headers"> with dt/dd pairs
├── RunRowDetail.module.css   # .headers: grid fit-content(max(40%, 16ch)) minmax(0,1fr); dt/dd wrap in their cell
├── RunRowDetail.test.tsx     # + red cases: list semantics, grid contract, wrap contract
└── RunResults.test.tsx       # rowheader query → term/definition query (line 182)

frontend/src/features/documentation/components/
├── ApiDocView.tsx            # Name cell: name in <span className={styles.parameterName}>
├── ApiDocView.module.css     # .nameCell (no break-anywhere, nowrap) + .parameterName cap 32ch
└── ApiDocView.test.tsx       # new: red case for the name-cell contract
```

**Structure Decision**: frontend-only change inside two existing feature folders. No Go file, no
`platform/` wrapper, no `shared/` component, no BDD feature file. No change to
`test/integration/` (see Complexity Tracking).

## Design

### Runner header list (FR-001 – FR-005, FR-007, FR-008, FR-010)

- Markup: `<dl className={styles.headers} aria-label="Request headers">` holding, per header, a
  `<Fragment key=…>` of `<dt translate="no">{key}</dt><dd translate="no">{value}</dd>` — `dt`/`dd`
  are direct grid items, so every row shares the same two columns. Rendered only when
  `row.sentRequest.headers.length > 0` (unchanged).
- CSS `.headers`: `display: grid; grid-template-columns: fit-content(max(40%, 16ch)) minmax(0, 1fr);
  column-gap: var(--space-3); row-gap: var(--space-1); margin: 0; font-family: var(--font-mono);
  font-size: var(--font-size-sm)`.
- CSS `.headers dt`: `min-width: 0; color: var(--color-text-muted); overflow-wrap: anywhere`.
  With `fit-content(max(40%, 16ch))` the column is `min(longest name's max-content, max(40%, 16ch))`
  (`ch` resolves in the list's 12 px mono font), so a name breaks only when it is longer than the
  cap; `anywhere` then lets it break inside its cell (hyphens and
  spaces are still preferred, being ordinary break opportunities) instead of overflowing.
- CSS `.headers dd`: `min-width: 0; margin: 0; color: var(--color-text); overflow-wrap: anywhere`.
- The old `.headers th, .headers td` rules are removed; nothing else in the module changes.

### Documentation Parameters Name cell (FR-006, FR-007, FR-008)

- Markup: `<td className={styles.nameCell} translate="no"><span className={styles.parameterName}>
  {parameter.name}</span>{required marker unchanged}</td>`. The Example cell keeps `.code`.
- CSS `.nameCell`: `font-family: var(--font-mono); white-space: nowrap` — no `overflow-wrap:
  anywhere`, so the name and its `required` marker stay on one line.
- CSS `.parameterName`: `display: inline-block; width: max-content; max-width: 32ch;
  white-space: normal; overflow-wrap: anywhere; vertical-align: top`. Its used width is
  `min(max-content, 32ch)` and is definite, so the auto table cannot squeeze it; longer names wrap
  inside it.

### Out of scope (FR-009)

`ResponseHeaders`, `HeadersPanel`, `ParamsPanel`, `BodyPanel`, `VersionRows` use
`table-layout: fixed` with `width: 100%` (equal fixed shares — no squeeze), and
`ConsoleEntryDetail` uses a per-row grid with a `minmax(8rem, auto)` name column and no
break-anywhere on names; none is touched.

## Test plan

| Requirement | Test (file) | Kind |
|---|---|---|
| FR-002, FR-010 | `RunRowDetail.test.tsx` — list named "Request headers"; terms are the header names in order, each followed by its definition (value); no `table` rendered | Vitest DOM (red) |
| FR-001, FR-003, FR-004, FR-005 | `RunRowDetail.test.tsx` — with `RunRowDetail.module.css` injected: list `display: grid`, `grid-template-columns` = `fit-content(max(40%, 16ch)) minmax(0, 1fr)`; `dt` and `dd` `overflow-wrap: anywhere`, `min-width: 0` | Vitest style contract (red) |
| FR-008 | `RunRowDetail.test.tsx` — masked value `•••••` shown as the definition of `Authorization`; existing `RunResults.test.tsx:182` updated to the term/definition query and still green | Vitest DOM |
| FR-006 | `ApiDocView.test.tsx` (new) — with `ApiDocView.module.css` injected: name cell `white-space: nowrap` and not `overflow-wrap: anywhere`; name span `display: inline-block`, `width: max-content`, `max-width` set (jsdom resolves `32ch` to `256px`), `overflow-wrap: anywhere`; `required` marker in the same cell | Vitest style contract (red) |
| FR-007, SC-001–SC-005 | By-eye on macOS in the scratchpad Vite harness: reported row, long name/value row, docs table; before/after screenshots in `screenshots/` and the PR | Manual, recorded |
| FR-009 | No diff in the listed files (`git diff --stat origin/main`) | Review |

No Godog scenario: nothing in Go changes and no step in `test/integration/steps/` can observe CSS
layout (Complexity Tracking).

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| V. BDD-first: no Godog `.feature` for these requirements | The behaviour is CSS layout inside the webview; no Go code is involved, so no step in `test/integration/steps/` can observe it | A Godog scenario would assert nothing real; Vitest is the constitution's frontend gate and is written red-first here exactly like a feature file would be. Issue #41 itself names a Vitest class/style-contract check as the test. Precedent: XEN-010, XEN-015, XEN-019, XEN-021 |
