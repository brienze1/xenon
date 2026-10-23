# Quickstart: validating the Collection Runner (XEN-017)

## Prerequisites

- A fresh worktree on `xen-017`: `cd frontend && npm ci`, then `make bindings` (the memory note:
  this is needed before Vitest).
- The Go toolchain and Wails v3 CLI, as in the repo README.

## Automated

```bash
make bdd TAG=@runner            # every runner feature (test/integration/features/runner/)
make bdd TAG=@mcp_runner        # MCP run / runs tools
make bdd TAG=@pick_file         # data-file filter on the file dialog
cd frontend && npm test -- runner item-settings sidebar tabs   # Vitest
make validate                   # full gate: Go unit + BDD + lint + typecheck + Vitest
```

Expected: everything is green, and the pre-existing suites are unchanged. In particular
`@history`, `@cookies`, `@variables`, `@sending` and `@scripts` prove that a single send behaves
exactly as before.

## Manual (dev app: `wails3 dev`)

1. **Functional run**
   1. Create a collection **Demo** with three requests to `https://httpbin.org`: `POST /post`
      with a test, `GET /get` and `GET /status/500`.
   2. Open Demo's tab and choose Runner. All three are listed and checked: "3 of 3 selected".
   3. Click Start run. Progress updates per request.
   4. In Results, row 3 is Failed when its test expects 200. Filter Failed shows only row 3.
2. **Order**
   1. Drag row 3 to the top. Use the keyboard to move row 2 down.
   2. Close and reopen the tab: the order is kept. `git status` in the workspace shows no change.
   3. Reset restores the sidebar order.
3. **Data**
   1. Pick a 3-row `users.csv`. Iterations becomes 3.
   2. Run it. Results are grouped into iterations 1–3, and `{{email}}` resolves in each URL.
4. **Control flow**: add `pm.execution.setNextRequest(null)` to request 1's tests and run. Only
   request 1 runs per iteration.
5. **Keep values off**: a test sets `pm.environment.set('x', Date.now())`. After the run, the
   environment's current value of `x` is unchanged.
6. **Performance**
   1. Switch to Performance with 5 users, 20 s and Ramp up (initial 1, ramp 10 s).
   2. The live chart shows the user count climbing.
   3. The summary shows p90/p95/p99 and throughput.
7. **Past runs**: the Runs sub-tab lists the runs. Open one, then delete one, then Clear runs.
8. **Sidebar**: right-click a folder and choose Run…. The folder's tab opens on Runner.
9. **Toast**: start a run, switch to another tab, and wait. A toast appears, and Open results goes
   back to the run.
10. **MCP**: from an MCP client, call `run {action:"start", target:"Demo"}`, then `runs
    {action:"get", run_id}` until the status is `completed`.
