# Implementation Plan: The Eye Opens the Environment Tab

**Branch**: `xen-015` | **Date**: 2026-10-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/XEN-015-eye-opens-environment-tab/spec.md`

## Summary

The eye next to the environment switcher stops opening the Quick Look dialog. It calls a new
`onOpenEnvironment(environmentId)` callback when an environment is active and the existing
`onOpenGlobals()` callback when none is; `app/App.tsx` wires both to the tab store's
`openEnvironmentTab(target, 'pinned')`, which already opens-or-focuses (and pins) the one tab per
target. Its label is "Open environment" / "Open globals". The Quick Look component and everything
only it used are deleted: its CSS and tests, `useDiscardGuard`, `DiscardChangesConfirm`, and the
variable table's `draftLifetime` and `savedByTab` props (the "drop edits on close" lifetime and the
in-table Save/Discard footer). The PRD and the XEN-002 spec are reworded. Frontend only; Vitest
is the red-first gate (see research R4).

## Technical Context

**Language/Version**: TypeScript 5 (strict), React 19

**Primary Dependencies**: Vite, Zustand, Vitest + Testing Library + user-event (all already in
`frontend/package.json`; no new dependency)

**Storage**: N/A — nothing on disk changes; tab layout is already persisted by the existing UI
state flow

**Testing**: Vitest (`npm test` in `frontend/`), plus `make validate` (Go + BDD + frontend lint,
typecheck, Vitest)

**Target Platform**: Wails v3 desktop webview on macOS, Windows, Linux

**Project Type**: desktop-app (Go backend + React UI); this feature touches only the UI

**Performance Goals**: N/A — one click calls an existing synchronous store action

**Constraints**: frontend import boundaries (`frontend-architecture` rule 2: a feature may not
import another feature's `model.ts`), PRD tokens for styling, no dead code left behind

**Scale/Scope**: 8 edited frontend source files, 6 deleted files, 3 edited test files and 2
docs

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Local-first** — ✅ no state added; no network or disk access.
- [x] **II. Secrets** — ✅ no secret handling changes; secrets keep showing masked in the tabs the
  eye now opens (unchanged `VariableTable` rows).
- [x] **III. Go owns logic** — ✅ React only switches which tab is shown; no Go call is added.
  No Wails import is added.
- [x] **IV. Architecture skills** — ✅ every touched file stays in its `frontend-architecture`
  place; the switcher acts on tabs through a callback composed in `app/` (it may not import
  `features/tabs/model.ts` for `EnvironmentTabTarget`), the same way `onOpenGlobals` and
  `onManage` already work. No file is added; only edits and deletions.
- [x] **V. BDD-first** — ✅ with justification (Complexity Tracking): no Go behaviour changes, so
  no Godog scenario can observe this. The red-first tests are Vitest cases (the constitution's
  suite includes "frontend lint + typecheck + Vitest"), written and proven red before the change.
  `/create-tests` writes **no `.feature` file** for this task; it writes the red Vitest cases
  T002, T003 and T006 (see tasks.md). Only `platform/` services are mocked, per the constitution.
- [x] **VI. Postman parity** — ✅ Postman's eye opens a quick-look popover; the user decided
  (issue #18) that Xenon's tab is the better view, and the PRD is updated in the same change so
  the PRD stays the source of truth (FR-010). Nothing beyond the MVP is added.
- [x] **VII. Cross-platform** — ✅ pure UI behaviour; no paths, keychain or shortcuts involved.

## Project Structure

### Documentation (this feature)

```text
specs/XEN-015-eye-opens-environment-tab/
├── plan.md              # This file
├── research.md          # Phase 0 decisions R1–R6
├── data-model.md        # UI state involved (no persisted data)
├── quickstart.md        # How to validate by hand and by test
├── contracts/
│   └── eye-button.md    # The switcher's eye: props, label, click contract
├── checklists/
│   ├── requirements.md
│   └── ux.md
└── tasks.md
```

### Source Code (repository root)

```text
frontend/src/
├── app/
│   ├── App.tsx                                  EDIT  wire onOpenEnvironment → openEnvironmentTab({kind:'environment'}, 'pinned')
│   └── App.test.tsx                             EDIT  eye opens / focuses the env tab and the Globals tab in the real app
└── features/environments/
    ├── components/
    │   ├── EnvironmentSwitcher.tsx              EDIT  eye IconButton inline (glyph moved from QuickLook), onOpenEnvironment prop, label
    │   ├── EnvironmentSwitcher.module.css       EDIT  + .icon (moved from QuickLook.module.css)
    │   ├── EnvironmentSwitcher.test.tsx         EDIT  replace the two quick-look tests with eye tests
    │   ├── QuickLook.tsx                        DELETE
    │   ├── QuickLook.module.css                 DELETE
    │   ├── QuickLook.test.tsx                   DELETE
    │   ├── DiscardChangesConfirm.tsx            DELETE (only QuickLook used it)
    │   ├── DiscardChangesConfirm.module.css     DELETE
    │   ├── VariableTable.tsx                    EDIT  drop DraftLifetime/draftLifetime and savedByTab, the unmount discard effect,
    │   │                                              the footer, requestSave and discard(); keep confirmSave/afterSave
    │   ├── VariableTable.module.css             EDIT  drop .footer rules
    │   ├── VariableTable.test.tsx               EDIT  save/discard through the store (as the tab does); drop the "clears on close"
    │   │                                              test; focus-on-invalid test goes through the confirm path
    │   ├── EnvironmentTab.tsx                   EDIT  drop draftLifetime/savedByTab props
    │   ├── EnvironmentTab.test.tsx              UNCHANGED  its "no Save/Discard buttons" assertions stay true (now unconditionally)
    │   └── GlobalsTab.tsx                       EDIT  drop draftLifetime/savedByTab props
    └── hooks/
        └── use-discard-guard.ts                 DELETE (only QuickLook used it; the hooks/ folder is then empty and goes too)
frontend/src/app/layout/ItemWorkspace.tsx        EDIT  drop draftLifetime/savedByTab props
docs/prd/xenon-mvp-prd.md                        EDIT  environment switcher bullet + environments decision
specs/XEN-002-environments-and-secrets/spec.md   EDIT  US secrets scenario 3, US environments intro + scenario 4, FR (secrets), FR-028
```

No file under `internal/`, `cmd/`, `pkg/` or `test/integration/` changes. **No BDD feature
file**: nothing in Go observes which tab the eye opens (Constitution V justification below).

**Structure Decision**: frontend-only change in the existing `frontend/src` layout; no file is
added, so no placement decision beyond moving the eye glyph and its `.icon` class into the
switcher that now owns the button.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| V. BDD-first: no Godog `.feature` for these requirements | The behaviour is which main-area tab the UI focuses; no Go code is involved, so no step in `test/integration/steps/` can observe it | A Godog scenario would assert nothing real; Vitest is the constitution's frontend gate and is written red-first here exactly like a feature file would be. The issue (brienze1/xenon#18) itself names Vitest as the coverage |
