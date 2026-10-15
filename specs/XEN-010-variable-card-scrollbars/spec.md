# Feature Specification: Variable Card Value Field Without Scrollbars

**Feature Branch**: `xen-010`

**Created**: 2026-10-15

**Status**: Draft

**Input**: User description: "XEN-010 — Variable card value field shows scrollbars over its text. Hovering a {{variable}} in the request URL bar opens a small hover card. For an undefined variable the card offers a one-line value box ("Enter value…") plus the scopes to add it to; for a defined variable it shows the current value to edit in place. In the installed macOS build (after XEN-009) the undefined-variable card's value box draws a horizontal scrollbar and a vertical scroll thumb inside a field one line tall, covering the placeholder and any typed text, so the field is barely usable — most visibly when macOS "Show scroll bars" is set to "Always". Users need the value box to read like a plain text field: no scrollbars when empty or holding a single line, the placeholder and typed text fully visible at the card's width, long values wrapping onto more lines with the box growing to fit up to a small maximum height (only beyond that maximum may it scroll, and only vertically — never sideways), with the same behaviour in the defined-variable card and regardless of the system scrollbar setting or the app's zoom level. Acceptance: a frontend unit test for the variable card covers the value field's sizing/overflow rules where testable in jsdom, the fix is checked by eye in the packaged macOS app with scrollbars set to "Always", and make validate stays green. Source: GitHub issue brienze1/xenon#7."

## Context

A bug fix in the variable hover card added by the UI polish merged with XEN-007 (`0e59de5`, the
Postman-style card that opens when the pointer rests on a `{{variable}}` in the URL bar or another
variable-aware editor).
Reported in GitHub issue [brienze1/xenon#7](https://github.com/brienze1/xenon/issues/7), seen in
the installed macOS build at `93f5381` (after the XEN-009 merge).

The card has two forms, and both have a value box:

- **Undefined variable** ("`{{authPrvateUrl}}` is not defined"): an empty value box with the
  placeholder "Enter value…", then "Add to" and the Environment / Collection / Globals choices.
- **Defined variable**: the variable's current value in an editable box (placeholder "Enter
  value…", or for a secret "Secret — type to replace…" / "No value on this machine — enter
  one…"), then the scope it comes from and "Variables in request →". A secret whose vault is
  locked shows an "Unlock the vault to use this secret" link instead of a box.

What goes wrong today (verified against the code at `93f5381`): the box's height is a guess
made by counting characters (36 per line, at most 6 lines), and the box is allowed to show
scrollbars whenever its text is even slightly taller or wider than that guess. When the system
draws permanent scrollbars (macOS "Show scroll bars: Always", and Windows/Linux by default),
any mismatch — a fractional line height at a non-100% app zoom, a font whose real width
differs from the guess, a placeholder that wraps — makes the box draw a vertical scrollbar.
That scrollbar takes width from the text, so the text overflows sideways and a horizontal
scrollbar is drawn too. In a box one line tall, the two scrollbars cover everything.

## Clarifications

### Session 2026-10-15

- Q: The variable table already fits its value cells to their content; how does the card's fix
  reuse that? → A: One shared, domain-free fit-to-content behaviour (with an optional line cap)
  used by both the card's value box and the variable table's value cells; the table's visible
  behaviour does not change.
- Settled from the issue and the code rather than asked: the maximum height stays 6 lines of
  text, the bound the card already uses; past it the box scrolls vertically (the issue only
  asks for no scrollbars "when empty or holding one line" and growth "as `rowsFor` intends").
- Settled from the code: the defined-variable card shares the box and the defect, so it is in
  scope (the issue lists it under Expected); the locked-secret link has no box and is
  unaffected.
- Settled from the code: every placeholder the box can show fits on one line at the card's
  width — the longest, "No value on this machine — enter one…", is 37 characters; the box has
  284 px for text at 100% zoom and a 12 px monospace character is about 7.2 px wide, so 39
  characters fit. App zoom scales the text and the card together, so the ratio holds at every
  zoom level.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Type a value for an undefined variable (Priority: P1)

Luis hovers `{{authPrvateUrl}}` in the URL bar of a `HEAD` request. The card says the variable
is not defined and offers a value box. He sees "Enter value…" clearly, clicks the box, types a
URL and adds it to the active environment. At no point do scrollbars appear over the box.

**Why this priority**: This is the reported defect. The box is the only way to define a
variable from the URL bar, and today it is barely usable.

**Independent Test**: Open the card for an undefined variable with the system set to always
show scrollbars; the empty box and a box holding a short value both show their whole text with
no scrollbars.

**Acceptance Scenarios**:

1. **Given** the system shows scrollbars always, **When** the card for an undefined variable
   opens, **Then** the value box is one line tall, shows "Enter value…" in full and draws no
   scrollbar.
2. **Given** the card is open, **When** Luis types a value that fits on one line, **Then** the
   box stays one line tall, shows the whole value and draws no scrollbar.
3. **Given** the card is open, **When** Luis types or pastes a value longer than one line (a
   long URL or token), **Then** the text wraps, the box grows one line at a time to show all of
   it, and no scrollbar is drawn while the value fits in 6 lines.
4. **Given** the value is longer than 6 lines, **When** it is shown, **Then** the box stops at 6
   lines and the rest is reachable by scrolling vertically; it never scrolls sideways.
5. **Given** the value grew to several lines, **When** Luis deletes text, **Then** the box
   shrinks back to fit, down to one line.

---

### User Story 2 - Edit a defined variable's value in place (Priority: P2)

Luis hovers a defined `{{creatorAdsUrl}}`. The card shows its current value in the same kind of
box; he edits it and presses Enter. The box behaves exactly like the one in User Story 1.

**Why this priority**: Same box, same defect, used more often than the undefined form once a
workspace is set up — but the issue was reported on the undefined form.

**Independent Test**: Open the card for a defined variable with a short value, a long value and
an empty secret, with scrollbars set to always; each box shows its whole text or placeholder and
draws no scrollbar until the value passes 6 lines.

**Acceptance Scenarios**:

1. **Given** a defined variable whose value fits on one line, **When** its card opens, **Then**
   the box is one line tall, shows the whole value and draws no scrollbar.
2. **Given** a defined variable whose value needs 3 lines at the card's width, **When** its card
   opens, **Then** the box opens 3 lines tall with no scrollbar.
3. **Given** a secret variable (value never shown), **When** its card opens, **Then** the box is
   one line tall and its placeholder ("Secret — type to replace…" or "No value on this machine —
   enter one…") is shown in full with no scrollbar.
4. **Given** the card's value box, **When** Luis edits the value, **Then** the box grows and
   shrinks with it exactly as in User Story 1, scenarios 3–5.

---

### User Story 3 - The same at any zoom level and scrollbar setting (Priority: P3)

Luis zooms the app in or out (View → Zoom), or a teammate on Windows or Linux (where scrollbars
are always drawn) uses the same card. The value box looks and behaves the same: no scrollbars
until the value passes 6 lines.

**Why this priority**: The defect depends on the scrollbar setting and on fractional sizes that
zoom produces; the fix is only real if it holds for all of them.

**Independent Test**: Repeat User Story 1 scenarios 1–3 at the smallest (67%), 100% and largest
(200%) zoom levels with scrollbars set to always.

**Acceptance Scenarios**:

1. **Given** any of the app's zoom levels and the system set to always show scrollbars, **When**
   the card opens with an empty or one-line value, **Then** no scrollbar is drawn and the whole
   placeholder or value is visible.
2. **Given** the system set to show scrollbars only while scrolling, **When** the card opens,
   **Then** the box looks the same as with scrollbars always shown.

---

### Edge Cases

- **Value with line breaks** (pasted multi-line text): each line counts, wrapped lines included;
  the box grows to fit up to 6 lines, then scrolls vertically.
- **One unbroken long token** (no spaces, slashes or dashes): it breaks anywhere to wrap rather
  than overflowing sideways.
- **Value that changes from outside while the card is open** (the card is re-keyed when the
  defined value or its scope changes): the new box fits its new value from the first paint.
- **Card opened near the window's right edge**: the card keeps its width (it moves left), so the
  box's width — and therefore the wrapping — is unchanged.
- **Narrow window**: cannot shrink the card — the window's minimum width is 960 px
  (`entity.MinWindowWidth`), i.e. 480 CSS px even at 200% zoom, more than the card's 320 px plus
  its 16 px margin; the box's width is therefore fixed and only zoom changes how its text lays out.
- **Saving in progress** (the defined box is disabled while a save runs): sizing is unchanged.
- **Locked secret**: no box is shown, so nothing changes.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The variable card's value box MUST NOT show a horizontal scrollbar under any
  value, placeholder, zoom level or system scrollbar setting; text that is too wide MUST wrap,
  breaking inside a word when nothing else fits.
- **FR-002**: The value box MUST NOT show a vertical scrollbar while its content (value, or
  placeholder when empty) fits in 6 lines.
- **FR-003**: The value box's height MUST fit its content: one line when empty or holding a
  one-line value, growing a line for each wrapped or broken line, up to 6 lines.
- **FR-004**: The height MUST follow the content as it changes — growing as the user types or
  pastes, shrinking as they delete — and MUST be right on the first paint of the card.
- **FR-005**: The height MUST be measured from the content as actually laid out at the box's
  current width, not estimated from a character count, so wrapping and height always agree.
- **FR-006**: Once the content exceeds 6 lines, the box MUST stay 6 lines tall and let the user
  scroll vertically to the rest — with the wheel or trackpad, and by moving the caret with the
  keyboard, which keeps the caret's line in view; it MUST NOT scroll sideways.
- **FR-007**: Every placeholder the box can show ("Enter value…", "Secret — type to replace…",
  "No value on this machine — enter one…") MUST be visible in full in an empty box.
- **FR-008**: FR-001 to FR-007 MUST hold for both card forms: the undefined-variable box and the
  defined-variable box (including secret variables). The locked-secret link is unchanged.
- **FR-009**: The fix MUST NOT change anything else about the card: its width, the box's
  accessible name ("Value of <name>"), its focus ring and placeholder texts, the value it
  saves or adds, Enter-to-save on the defined box (Shift+Enter keeps a line break), saving when
  focus leaves or the card closes, the scope choices, and the error messages.
- **FR-010**: A frontend unit test for the variable card MUST cover the box's sizing and
  overflow rules that can be checked without a real layout engine: the box fits its measured
  content height, stops at 6 lines, allows vertical scrolling only once past 6 lines, never
  allows sideways scrolling, and applies the same rules to both card forms.
- **FR-011**: The fix MUST be checked by eye in the packaged macOS app with "Show scroll bars"
  set to "Always", against User Story 1 scenarios 1–4 and User Story 2 scenario 3, and the full
  validation suite MUST stay green.
- **FR-012**: The fit-to-content sizing MUST be one shared behaviour used by both the variable
  card's value box (with the 6-line cap) and the variable table's value cells (no cap). The
  variable table's cells MUST look and behave exactly as before: they grow to show their whole
  value, never scroll, and refit when their width changes (window or panel resize).

### Key Entities

- **Value box**: the editable text area in the variable card. Attributes that matter here: its
  content (value, or placeholder when empty), its width (fixed by the card), its line height,
  and the 6-line maximum height.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: With scrollbars set to always, 0 scrollbars are drawn on the value box for an empty
  box and for any value up to 6 lines, in both card forms, at the smallest, default and largest
  app zoom levels (67%, 100%, 200%); the sizing is measured at whatever zoom applies, so the
  levels in between behave the same.
- **SC-002**: 100% of the placeholder text and of any value up to 6 lines is visible without
  scrolling.
- **SC-003**: The box's height always equals its content's height (rounded to whole lines) up to
  6 lines; it never shows empty extra lines and never hides a line of a value that fits.
- **SC-004**: No value, however long, makes the box scroll sideways.
- **SC-005**: Every existing variable-card behaviour (adding, editing, saving, closing) keeps
  working: the existing card tests pass unchanged.
- **SC-006**: The variable table's value cells look and behave as before: the existing variable
  table tests pass unchanged.

## Assumptions

- The fix is in the frontend only; nothing about how variables are stored, resolved or saved
  changes, and no Go code changes.
- The card's 320 px width, its padding and its monospace 12 px text stay as they are.
- Windows and Linux are not checked by eye in this task: their scrollbars are permanent by
  default, which is exactly the condition macOS "Show scroll bars: Always" reproduces, and the
  sizing rules (FR-001–FR-006) do not depend on the platform; the shared unit tests run on every
  platform's CI (`.github/workflows/ci.yml` runs macOS, Windows and Ubuntu).
- "Checked by eye" means the implementer builds and runs the packaged macOS app, sets System
  Settings → Appearance → Show scroll bars to "Always", and confirms the scenarios named in
  FR-011; a screenshot is attached to the implementation report.
- Out of scope: other text areas in the app (beyond the variable table's value cells moving onto
  the shared sizing behaviour with no visible change, FR-012), the URL bar itself, and the card's
  layout beyond the value box.
