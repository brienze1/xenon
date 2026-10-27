# Quickstart: validating "Runs Are Visible While They Run"

## Automated

```bash
make bdd TAG=@xen_022            # new BDD scenarios (runner/run_storage.feature, mcp/runner.feature)
go test ./internal/domain/entity/ -run 'ListRuns|Listed'
cd frontend && npm test -- src/features/runner   # store.test.ts, RunsList.test.tsx, lib.test.ts
make validate                   # full gate
```

Expected: all green; before the change, the new BDD scenarios and Vitest cases are red.

## By hand (desktop app)

1. Open a workspace with a folder of ≥ 10 requests against a slow-ish target. Run it once from
   its tab (Runner → Start) and let it finish; leave the tab on that run's Results.
2. From an AI client connected over MCP: `run` `{ "action": "start", "target": "<folder>" }`.
3. The folder's tab switches to Results of the new run and updates row by row; Pause/Stop work.
4. Open **Runs**: the new run is first, badge **Running**, with the requests done so far; its row
   has no actions menu. From the AI client, `runs list` for the folder returns it first with
   `status: running`.
5. Open the older run from Runs: the tab stays on it while the new run keeps going.
6. When the new run ends: finished-run notice; Runs lists it once with its final status.
