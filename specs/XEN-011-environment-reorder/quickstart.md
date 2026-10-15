# Quickstart: validating XEN-011

## Prerequisites

- Branch `xen-011`, Go toolchain and Node per `Makefile` (`make setup` once).
- Bindings regenerated after the backend change: `make bindings`.

## Automated checks

```bash
make bdd TAG=@xen_011          # every XEN-011 scenario (see plan.md → BDD feature files)
make bdd                        # full suite stays green (tree order, globals, import, git merge, MCP)
cd frontend && npx vitest run src/features/sidebar/components/EnvironmentsView.test.tsx \
  src/features/environments/store.test.ts src/features/environments/components/EnvironmentSwitcher.test.tsx
make validate                   # final gate
```

Expected: all green; the XEN-011 scenarios were red before implementation (`/create-tests`).

## Manual run (`make dev`)

1. Open a workspace with environments `Alpha`, `Beta`, `Gamma` (never reordered). Confirm the
   list is alphabetical and `git diff workspace.yaml` is empty after opening.
2. Drag `Gamma` above `Alpha`. The list reads Gamma, Alpha, Beta at once; the top-right switcher
   matches; `workspace.yaml` now has `environments: { order: [<Gamma id>, <Alpha id>, <Beta id>] }`.
3. Quit and relaunch: same order.
4. Focus `Beta`, press Space, ArrowUp, Space: Beta moves up one; a screen reader (VoiceOver /
   NVDA) announces pick-up, "Beta moved to position 2 of 3", and the drop.
5. Type `a` in the search box: rows cannot be dragged and Space does not pick up. Clear it: reorder
   works again.
6. Create `Delta`, duplicate `Alpha`: both appear last. Rename `Gamma`: it stays first. Delete
   `Beta`: its id leaves `environments.order`.
7. Hand-edit `workspace.yaml` to list ids in another order (or add a bogus id): the sidebar updates
   within the file-watcher delay, the bogus id is ignored.
8. Make the workspace folder itself read-only (so `workspace.yaml` cannot be replaced) and drag: the list snaps back and the error banner shows the
   next step.
9. With AI access on, call `env_edit` `{ "action": "reorder", "id": "<Delta id>", "index": 0 }`
   from an MCP client: Delta jumps to the top in the open window.

See `contracts/` for exact request/response shapes and `data-model.md` for the file format.
