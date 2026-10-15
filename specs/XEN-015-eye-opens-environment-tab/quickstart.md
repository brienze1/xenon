# Quickstart: Validate the Eye

## Prerequisites

Fresh worktree: `cd frontend && npm ci`, then `make bindings` from the repo root (Vitest imports
the generated bindings' types).

## Automated

```bash
cd frontend
npx vitest run src/features/environments/components/EnvironmentSwitcher.test.tsx \
               src/features/environments/components/VariableTable.test.tsx \
               src/features/environments/components/EnvironmentTab.test.tsx \
               src/app/App.test.tsx
cd .. && make validate
```

Expected: all green; no test references `QuickLook`, `draftLifetime` or `savedByTab`.

```bash
grep -rnE "QuickLook|useDiscardGuard|DiscardChangesConfirm|draftLifetime|savedByTab|DraftLifetime" frontend/src
```

Expected: no output.

## By hand (`make dev` or the packaged app)

1. Open a workspace with an environment of ~30 variables (e.g. HML) and make it active.
2. Click the eye next to the switcher → the HML tab opens pinned (no italic preview title) and
   focused; no dialog. Hovering the eye shows "Open environment".
3. Focus another tab, click the eye again → the same HML tab is focused; the tab count is
   unchanged.
4. Close the HML tab, single-click HML in the sidebar (preview), focus another tab, click the eye
   → the preview tab becomes pinned and focused.
5. Edit a variable in the HML tab (unsaved dot), focus another tab, click the eye → the HML tab
   is focused with the edit and dot intact.
6. Pick "No environment" → the eye's tooltip reads "Open globals"; click it → the Globals tab
   opens pinned and focused.
