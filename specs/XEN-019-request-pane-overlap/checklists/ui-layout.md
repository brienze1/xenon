# UI Layout Requirements Quality Checklist: Request Pane Content Never Overlaps

**Purpose**: Unit tests for the requirements of XEN-019 — overlap/scroll behaviour at minimum sizes, divider limits, visual regression, accessibility of scrolled and pinned content, cross-platform webview, and the verification strategy
**Created**: 2026-10-23
**Feature**: [spec.md](../spec.md) · audience: PR reviewer · depth: standard

Each item was run against the artifacts; failures were fixed in the artifacts before ticking
(noted as "fixed").

## Requirement Completeness

- [x] CHK001 Are overlap requirements defined for every tab the issue names (6 request sub-tabs, both example parts, 6 collection/folder sections, environment, globals)? [Completeness, Spec §FR-001, §FR-009] — all listed by name
- [x] CHK002 Is every fixed row that must not shrink enumerated? [Completeness, Spec §FR-002] — sub-tab row, phase-tab row with inherit checkbox, Scripts toolbar, hint, plus "every other toolbar, heading or note row"
- [x] CHK003 Are shrink-first minimums quantified for every editor, and is table behaviour stated? [Completeness, Spec §FR-003] — fixed: FR-003 now states tables keep natural row heights and adds the new 160 px example response panel minimum
- [x] CHK004 Is the scrolling region defined for each tab type? [Completeness, Spec §FR-004, §FR-006, §FR-009] — fixed: FR-009 now says what scrolls on example, collection/folder, environment and globals tabs
- [x] CHK005 Are divider-floor requirements defined for both pointer dragging and keyboard resizing? [Completeness, Spec §FR-005, Edge Cases]

## Requirement Clarity

- [x] CHK006 Is "about 200 px" fixed to an exact value? [Clarity, Spec §Assumptions] — exactly 200 px
- [x] CHK007 Is the trigger for scrolling the whole request builder quantified? [Clarity, Spec §FR-006] — panel under the sub-tab row below 64 px
- [x] CHK008 Is "natural height" defined so that wrapped text counts? [Clarity, Spec §FR-002] — fixed: "at least the height of their content, including wrapped lines"
- [x] CHK009 Is the unit of the 200 px floor defined relative to View → Zoom? [Clarity, Gap] — fixed: UI pixels that scale with the zoom (FR-005, Clarifications; `app/zoom.ts` scales the body with CSS `zoom`)

## Requirement Consistency

- [x] CHK010 Is "header stays pinned" (FR-004) consistent with "the builder scrolls as a whole" (FR-006)? [Consistency, Spec §FR-004, §FR-006] — FR-004 is explicitly "subject to FR-006"
- [x] CHK011 Is the 200 px floor consistent with the 80 % maximum when the split area is small? [Conflict, Spec §FR-005, Edge Cases] — fixed: the floor never exceeds 80 % of the area, so the response keeps ≥ 20 % (FR-005, edge case, data-model display rule, contract `--split-max`)
- [x] CHK012 Are side-by-side requirements consistent across spec, plan and data model? [Consistency, Spec §FR-008] — no new minimum width anywhere; `RequestWorkspace` passes no `minSize` side-by-side (contract)
- [x] CHK013 Are "request pane", "request builder" and "panel under the sub-tab row" used consistently? [Consistency] — the pane is the split area slot, the builder its content, the panel the scroll box below the sub-tab row; FR-006 and the Clarifications use them the same way

## Acceptance Criteria Quality

- [x] CHK014 Can "0 overlapping texts or controls" be objectively verified? [Measurability, Spec §SC-001] — an enumerated list of tabs at a fixed window size and divider position, checked by eye (quickstart steps 1–6)
- [x] CHK015 Does "visually identical" have a defined baseline and conditions? [Measurability, Spec §SC-004] — fixed: 1280 × 800 default window, divider at 80 % (where every tab's content fits), before = `origin/main`
- [x] CHK016 Are the acceptance scenarios tied to measurable conditions (window size, divider position, zoom)? [Acceptance Criteria, Spec §User Stories] — 960 × 600, highest divider point, 100 % zoom where it matters

## Scenario Coverage

- [x] CHK017 Are requirements defined for banner states (error, on-disk conflict, cURL notice) at the floor? [Coverage, Spec §US2 AS3, §FR-006]
- [x] CHK018 Are requirements defined for the disabled-script notice in the toolbar? [Coverage, Spec §Edge Cases]
- [x] CHK019 Are requirements defined for the console drawer being open? [Coverage, Spec §Edge Cases]
- [x] CHK020 Is behaviour defined when the window shrinks after the divider was positioned? [Coverage, Spec §FR-011] — the stored ratio is kept and shown at the floor

## Edge Case Coverage

- [x] CHK021 Is wrapping of the long hint at narrow widths specified? [Edge Case, Spec §FR-007, §US1 AS3]
- [x] CHK022 Is the split between editor-internal scrolling and panel scrolling defined for long scripts/bodies? [Edge Case, Spec §Edge Cases]
- [x] CHK023 Is the extreme small-area case (minimum window at high zoom) defined? [Edge Case, Gap] — fixed: new edge case, response keeps ≥ 20 %, request content scrolls

## Non-Functional Requirements

- [x] CHK024 Are keyboard-focus requirements defined for controls inside the new scroll areas (scrolled into view, focus rings not cut off)? [Accessibility, Gap] — fixed: FR-012; quickstart step 9
- [x] CHK025 Is horizontal-scroll behaviour of the new scroll boxes specified? [Coverage, Gap] — fixed: FR-012 "no new horizontal scrollbar"; contract `overflow-x: hidden` on the builder and the example response part, with wide tables still scrolling inside their own panel (research R4)
- [x] CHK026 Are cross-platform webview expectations and their verification stated? [Non-Functional, Spec §SC-005, Clarifications] — WebKit, WebView2 and WebKitGTK via CI; a manual check on macOS only, decided by the product owner
- [x] CHK027 Is resize performance addressed? [Non-Functional, Plan §Technical Context] — a pure-CSS floor and fixes add no JavaScript to resizing; no requirement is needed in the spec

## Dependencies & Assumptions

- [x] CHK028 Is the 960 × 600 window minimum documented as a verified dependency? [Dependency, Spec §Context] — from `internal/domain/entity/window_state.go`
- [x] CHK029 Is the assumption that the existing editor minimums are "sensible" recorded? [Assumption, Spec §Assumptions]
- [x] CHK030 Is the UI-only scope (no stored data, format or backend change) stated? [Assumption, Spec §Assumptions, data-model.md]

## Verification Strategy

- [x] CHK031 Is the absence of a Godog feature justified, with alternative red-first tests named? [Traceability, Plan §Complexity Tracking, tasks T003–T011]
- [x] CHK032 Are the limits of jsdom (no layout) and the by-eye complement documented? [Traceability, research R5, quickstart.md]
- [x] CHK033 Is the screenshot evidence for the PR defined (which steps and which baseline)? [Measurability, quickstart.md, task T014] — steps 1 and 7, before on `origin/main`
