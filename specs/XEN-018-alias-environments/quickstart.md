# Quickstart: Alias Environments

## Prerequisites

Fresh worktree: `cd frontend && npm ci && cd .. && make bindings` (Vitest needs the generated
bindings).

## Automated validation

```bash
make bdd TAG=@alias_environments       # the four feature files in plan.md
make bdd                               # whole suite stays green (XEN-016 alias features included)
cd frontend && npx vitest run src/features/request-builder src/features/sidebar src/platform/services/aliases.test.ts
make validate                          # final gate
```

## Manual check (desktop app, `make dev`)

1. Create environments **HML** (green) and **Prod** (red), and a request **Login** with aliases
   `mx-test-user` and `v2`.
2. Aliases ⋯ on `mx-test-user` → **Environments…** → Only these environments → HML → Save. The
   request file now has `environments: [<HML id>]` under that alias only; `v2` has no
   `environments` key.
3. Activate HML: `mx-test-user` shows the green HML dot and sends.
4. Activate Prod with `mx-test-user` on screen: the tab keeps it, Send is disabled, and the banner
   reads "mx-test-user is only for HML. [Switch to HML]". In the dropdown `mx-test-user` is greyed
   out; Default and `v2` are selectable.
5. Press **Switch to HML**: HML becomes active and Send is enabled.
6. Activate no environment: `mx-test-user` is greyed out.
7. From an AI client (MCP) with Prod active: `send` with `alias_id` = `mx-test-user` fails with
   `ALS-09403`, with no approval prompt.
8. History rows show the environment dot + name of each send.
9. Delete HML: the alias shows "Deleted environment" and is unavailable everywhere. The request
   file is unchanged.
10. Export the collection to Postman: the `Login (mx-test-user)` request's description ends with
    "Xenon alias for environments: …".
