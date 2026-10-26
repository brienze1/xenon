# Quickstart: validating XEN-020

## Automated

```bash
# Frontend (fresh worktree: npm ci + make bindings first)
cd frontend && npx vitest run src/features/runner/store.test.ts src/app/bootstrap.test.ts

# Backend guard scenarios (green before and after the fix)
make bdd TAG=@xen_020

# Full gate
make validate
```

Expected: the `refreshAll` tests in `store.test.ts` and the Runner case in `bootstrap.test.ts`
fail on `main` and pass after the change; `@xen_020` scenarios pass on both; nothing else turns red.

## Manual (the issue's repro)

1. Open a workspace with a folder of several requests; open the folder's tab → **Runner**.
   Note the row count; uncheck one row, drag another to the top.
2. Leave the Runner open. Add requests to an existing subfolder and add a new subfolder with
   requests — once from the sidebar (New Request / New Folder) and once from an AI client
   (`tree_edit` create).
3. Within a second of the sidebar showing them, the Runner shows them: checked, each after its
   nearest preceding request in sidebar order. The unchecked row is still unchecked; the dragged row is still first.
4. Delete one request from the sidebar → its row disappears.
5. Click **Run** → the run's row count equals the checked rows shown.
6. Delete the folder itself while its Runner is shown → the Runner shows the load-error banner
   with **Try Again**; no crash.
7. From an AI client, `run` `start` on the folder without `rows` → the run includes the new requests.
