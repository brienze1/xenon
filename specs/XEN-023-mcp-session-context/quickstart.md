# Quickstart: validating XEN-023

## Prerequisites (fresh worktree)

```bash
cd frontend && npm ci && cd .. && make bindings   # Vitest needs generated bindings
```

## Automated

```bash
make bdd TAG=@xen_023          # all new/changed scenarios
make bdd TAG=@mcp               # every MCP feature (workspaces.feature changed)
make bdd TAG=@environments      # environment features (revision added)
cd frontend && npx vitest run src/app/bootstrap.test.ts src/features/workspaces src/features/environments src/features/sidebar
make validate                   # final gate
```

Expected: all green. The `@xen_023` scenarios are listed in `plan.md` → Tests.

## By hand (desktop app + two MCP clients)

1. Open Xenon on workspace **Acme**, environment **Dev**. Connect two MCP clients (e.g. two Claude Code sessions with
   the Xenon MCP).
2. Session 1: `workspace active` → Acme, Dev active. Session 2: `workspace_edit switch` to **Globex**, then
   `env_edit set_active` **Prod**.
   - Window: still Acme / Dev; sidebar, tabs and environment switcher unchanged; a running collection run in Acme keeps
     running.
   - Session 1: `workspace active` → still Acme / Dev; `send` resolves `{{base_url}}` with Dev.
3. Session 1: `send` with `"environment": "Prod"` → resolved with Prod; next plain `send` → Dev again.
4. Session 2: `workspace_edit create` **Payments** → appears in the window's workspace list; window stays on Acme.
5. In the window, open Dev's variables, type a change (don't save). Session 1: `env_edit add_var` on Dev. Save in the
   window → "This environment changed on disk since it was opened" notice with **Load from disk** / **Keep mine**.
6. History (AI filter): rows show `claude-code · <6-char tag>` with different tags per session; typing a tag in the
   search shows only that session's rows.
