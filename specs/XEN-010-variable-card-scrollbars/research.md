# Research: Variable Card Value Field Without Scrollbars

All facts below were checked against the code at `93f5381` (`origin/main`, 2026-10-15) and, for
jsdom, by running jsdom 30.1.2 (the version pinned in `frontend/package.json`).

## R1 — Root cause

- **Decision**: The defect is the combination of an estimated height and browser-default overflow.
- **Facts**:
  - `VariableCard.tsx` sizes both value boxes with `rows={rowsFor(text)}`, where `rowsFor` counts
    36 characters a line (`CHARS_PER_ROW`), capped at `MAX_ROWS = 6`.
  - `VariableCard.module.css` `.value` sets padding `var(--space-2) var(--space-3)` (4 px / 8 px),
    `line-height: 1.5`, a 1 px border and `resize: none`, and nothing on `overflow` — so the UA
    default for a textarea, `overflow: auto`, applies.
  - The card is 320 px wide with 8 px padding and a 1 px border (`DecorationCards.module.css`), so
    the box is 302 px wide and has 284 px for text; 12 px monospace text (`--font-size-sm`,
    JetBrains Mono → ui-monospace/SF Mono/Menlo fallbacks, all 0.6 em) is ~7.2 px a character,
    so ~39 characters fit a line — not 36. Height and wrap therefore disagree for many values.
  - The app zooms with CSS `zoom` at 0.67–2 (`frontend/src/app/zoom.ts`); `1.5 × 12 px × zoom`
    is fractional at most levels, so even a one-line box can overflow its own content box by a
    sub-pixel.
  - With permanent scrollbars (macOS "Always", Windows, Linux), any overflow under `overflow:
    auto` draws a ~15 px vertical scrollbar; it narrows the text area so text that just fitted
    now overflows sideways and the horizontal bar is drawn too. In a 28 px tall box both bars
    cover the content — matching the report.
- **Rationale**: A fix that only tweaks the estimate (e.g. 39 characters) would still break at
  other zoom levels, fonts or window widths; the fix must make height and wrap agree by
  construction and stop the box from ever drawing bars it doesn't need.
- **Alternatives considered**: reproducing in WKWebView before deciding — not possible in this
  environment (no WebKit build available); the chosen fix does not depend on which of the above
  triggers fired, because it removes all of them (measured height, explicit overflow).

## R2 — Measure, don't estimate

- **Decision**: Size the box from `scrollHeight` after setting `height: auto`, in a layout effect,
  and refit on width changes with a `ResizeObserver` — the technique `WrapTextCell.tsx` already
  uses in the variable table.
- **Rationale**: `scrollHeight` is the content's laid-out height at the current width, font and
  zoom, so it always agrees with the wrapping. A layout effect runs before paint, so the card
  never shows a wrong height (FR-004). `WrapTextCell` has shipped this since the variable-table
  polish and works under zoom.
- **Alternatives considered**:
  - CSS `field-sizing: content` — not supported by every webview Xenon targets (WebKitGTK on
    Linux, older WKWebView), and it cannot express "cap at 6 lines then scroll" together with
    the existing `rows` minimum reliably; rejected.
  - A hidden mirror element to measure text — more DOM and duplicated styles for no gain over
    `scrollHeight`; rejected.

## R3 — Overflow ownership and the cap

- **Decision**: The hook sets `overflowX: hidden` always and `overflowY` to `hidden` until the
  content exceeds `maxLines × lineHeight + paddingY` by more than 1 px, then `auto` with the
  height pinned at the cap. The stylesheet also sets `overflow: hidden` as the default before the
  first measurement.
- **Rationale**: Hiding overflow while the box fits its content means a sub-pixel rounding error
  can never draw a scrollbar. The 1 px tolerance covers `scrollHeight` being rounded to an
  integer: at exactly 6 lines under a fractional zoom it may read one pixel over the cap, which
  must not turn scrolling on. `overflow-wrap: anywhere` (kept) guarantees there is never
  horizontal content to scroll to.
- **Alternatives considered**: a CSS `max-height` with `overflow-y: auto` — still draws a bar
  whenever rounding overflows below the cap, which is the bug; rejected.

## R4 — Testability in jsdom

- **Decision**: Tests stub `HTMLTextAreaElement.prototype.scrollHeight` (a getter derived from
  the element's value) and inject the box's metrics as a `<style>` rule; they assert the inline
  `height`, `overflowX` and `overflowY` the hook sets.
- **Facts** (jsdom 30.1.2, run on 2026-10-15): `scrollHeight`, `offsetHeight` and `clientHeight`
  are always 0; `getComputedStyle` applies `<style>` rules for `line-height` (`"18px"`),
  `padding-top/bottom` (`"4px"`) and `border-top-width` (`"1px"`), but does not expand the
  `overflow` shorthand (`overflowY` reads `"visible"`) and returns a unitless `line-height` raw
  (`"1.5"`). Vitest here does not inject CSS-module styles (`css.modules.classNameStrategy:
  'non-scoped'` only fixes the class names, so `.value` is the class name).
- **Consequences for the design**: the hook writes overflow as inline longhands (testable), the
  stylesheet uses `line-height: 18px` (a pixel value that the hook — and jsdom — read exactly),
  borders come from computed `border*Width` (not `offsetHeight - clientHeight`, which is 0 in
  jsdom), and when `lineHeight` is not a number the hook fits without a cap rather than producing
  `NaN`.
- **Alternatives considered**: enabling CSS processing for modules in `vitest.config.ts` — a
  global test-config change for one assertion; rejected.

## R5 — Where the hook lives

- **Decision**: `frontend/src/shared/hooks/use-fit-height.ts`, used by
  `features/request-builder/components/VariableCard.tsx` and
  `features/environments/components/WrapTextCell.tsx` (user decision, spec Clarifications).
- **Rationale**: frontend-architecture rule 2 forbids a feature importing another feature's
  hooks; a domain-free piece used by two features belongs in `shared/`. Naming follows
  `use-hotkey.ts` / `use-window-height.ts`.

## R6 — Packaged-app check on macOS

- **Decision**: `make package` (Wails `package` task) builds `bin/xenon.app`; set scrollbars to
  "Always" with System Settings → Appearance → Show scroll bars → Always (or
  `defaults write -g AppleShowScrollBars Always`, then relaunch the app); restore the user's
  previous setting afterwards (`defaults read -g AppleShowScrollBars` first; `defaults delete -g
  AppleShowScrollBars` restores the default "Automatically based on mouse or trackpad").
- **Rationale**: matches the issue's acceptance ("checked by eye in the packaged macOS app with
  scrollbars set to Always").
