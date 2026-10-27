# Quickstart: Verify XEN-021

## Automated

```bash
cd frontend
npm ci                      # fresh worktree only
cd .. && make bindings      # fresh worktree only (Vitest imports the generated bindings)
cd frontend
npx vitest run src/features/runner/components/RunResults.test.tsx
npm run lint && npm run typecheck && npm test
cd .. && make validate
```

Expected: the new cases in `RunResults.test.tsx` (DOM contract, CSS contract, keyboard) are green
along with every existing case; `make validate` exits 0.

## By eye (macOS, `wails3 dev`)

Prerequisite: a folder with ≥ 10 requests whose post-response scripts each run 3+ `pm.test(…)`
calls and log 7+ console lines (`console.log` plus the requests' own `GET …`/`POST …` lines), and
one request that logs 40+ lines (a loop of `console.log`). Window at 1280 × 800 unless stated.

1. **Reported case** — run the folder; when finished, expand the first row (3 tests, 7 console
   lines). The tests chips, the 3 test lines and the 7 console lines read top to bottom; nothing
   overlaps; the next row starts right below the panel (SC-001).
2. **Several expanded** — expand rows 1, 3 and 5; each next row/heading starts where the previous
   one ends; collapse row 3 and the rows below move up with no gap; scroll rows 1 and 5 out of view and back —
   they keep their height and position (SC-002).
3. **Cap** — expand the 40-line row; the panel stops at 320 px and scrolls to its last line;
   the row below starts right after the panel (SC-003).
4. **Live run** — start the run again; expand row 1 while later rows keep arriving; the arriving
   rows stack below, nothing overlaps; the expanded row's content stays legible.
5. **Filters** — switch the results filter between All / Passed / Failed and back with rows
   expanded; switch an expanded row's tests filter; heights follow, no gaps.
6. **Past run and side panel** — open the finished run from **Runs**; expand a row; then click a
   row's name to open the detail side panel; the expanded rows re-wrap to the narrower width with
   no overlap.
7. **Collapsed look** — with everything collapsed, compare to `origin/main`: identical row and
   heading heights (SC-004).
8. **Keyboard** — Tab from an expanded row's name button: the details panel gets the focus ring;
   ↓ / Page Down / End scroll it to its last line (SC-005).
9. Capture before (`origin/main`) and after screenshots of steps 1 and 3 for the pull request.

Windows and Linux are covered by CI's quality gate (Clarifications).
