# Quickstart: verifying XEN-019

## Automated

```bash
cd frontend
npm ci                                   # fresh worktree
npx vitest run src/shared/ui/split-pane src/app/layout/RequestWorkspace.test.tsx \
  src/features/scripts src/features/request-builder src/features/examples
npm run lint && npm run typecheck && npm test
cd .. && make validate                   # final gate (Go + BDD + frontend)
```

Expected: the new cases (see [tasks.md](tasks.md)) are red before the CSS/prop changes and
green after; every pre-existing case stays green.

## By eye (macOS, running app)

Run the app (`wails3 dev` from the repo root, or the `run` skill), set the window to its minimum
size (960 × 600), stacked layout.

1. Open a request → **Scripts**. Drag the divider to its highest point.
   - The request pane is about 200 px tall; breadcrumb, URL bar and sub-tab row are fully visible.
   - Nothing overlaps; scrolling the area under the sub-tab row reaches the phase tabs, the
     "Inherit parent scripts" checkbox, Snippets, the 160 px editor and the full hint, in order.
2. Repeat for **Params, Authorization, Headers, Body, Settings** — nothing overlaps, the panel scrolls.
3. Trigger a banner while at the floor (e.g. paste an invalid cURL command into the URL bar) —
   the whole request pane scrolls; nothing is clipped or overlapped.
4. Open the console drawer — the request pane keeps its ~200 px floor.
5. Switch to side-by-side — the overlap fix holds; the divider limits are 20 %–80 %, no minimum width.
6. Open an **example** tab, a **folder** tab (Scripts shows the inherit checkbox), a
   **collection** tab (Overview, Authorization, Headers, Scripts, Variables, Documentation), an
   **environment** tab and **Globals** — nothing overlaps; each scrolls.
7. At the default 1280 × 800 window, drag the divider to its lowest point (80 %) — every tab above
   looks as it does on `main`.
8. At the minimum window, set View → Zoom to the maximum — the response pane keeps at least 20 %
   of the area; the request content scrolls; nothing overlaps.
9. Tab through the Scripts controls in a short pane — each focused control scrolls into view and
   its focus ring is not cut off; no horizontal scrollbar appears in the request pane.

Take before (`main`) / after screenshots of steps 1 and 7 for the pull request. Windows and
Linux are covered by CI's `make validate`; no manual check there (spec Clarifications).
