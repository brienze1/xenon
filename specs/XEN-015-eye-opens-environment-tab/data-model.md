# Data Model: The Eye Opens the Environment Tab

No persisted data changes. The feature reads existing UI state only.

| State | Owner | Used for |
| --- | --- | --- |
| `activeId: string \| null` | `features/environments/store.ts` | Which target the eye opens and its label |
| `environments[]` (`id`, `name`, `error`) | `features/environments/store.ts` | Resolving the active environment |
| `tabs[]`, `previewTabId`, focused tab | `features/tabs/store.ts` | Open-or-focus of one tab per target (unchanged) |
| `EnvironmentTabTarget` = `{ kind: 'environment', environmentId } \| { kind: 'globals' }` | `features/tabs/model.ts` | Built in `app/App.tsx` only |

## Eye state

| `activeId` | Label | Click |
| --- | --- | --- |
| an environment id | "Open environment" | `onOpenEnvironment(activeId)` → `{ kind: 'environment', environmentId }` pinned |
| `null` | "Open globals" | `onOpenGlobals()` → `{ kind: 'globals' }` pinned |

## Removed

- `DraftLifetime` type and the `draftLifetime` / `savedByTab` props of `VariableTable`: every
  variable table now keeps its draft in the environments store until it is saved or discarded
  by its tab (the former `'workspace'` + `savedByTab` behaviour).
