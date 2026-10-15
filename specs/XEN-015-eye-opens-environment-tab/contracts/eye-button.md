# Contract: The Switcher's Eye

## `EnvironmentSwitcher` props

```ts
interface EnvironmentSwitcherProps {
  /** Opens the globals in a main-area tab. */
  onOpenGlobals: () => void
  /** Opens one environment in a main-area tab. */
  onOpenEnvironment: (environmentId: string) => void
  /** Shows the environments in the sidebar, where they are created, renamed and opened. */
  onManage: () => void
}
```

`onOpenGlobals` keeps serving the "Globals" menu item and now also the eye with no active
environment.

## The eye

- Rendered right after the switcher dropdown, as an `IconButton` with the unchanged eye glyph.
- Accessible name and `title`: `"Open environment"` when `activeId` names an environment,
  `"Open globals"` when `activeId` is `null`.
- Click / Enter / Space:
  - active environment → `onOpenEnvironment(activeId)` exactly once; `onOpenGlobals` not called.
  - no active environment → `onOpenGlobals()` exactly once; `onOpenEnvironment` not called.
- Never renders a dialog.

## App wiring (`app/App.tsx`)

| Callback | Store call |
| --- | --- |
| `onOpenEnvironment(id)` | `useTabsStore.getState().openEnvironmentTab({ kind: 'environment', environmentId: id }, 'pinned')` |
| `onOpenGlobals()` | `useTabsStore.getState().openEnvironmentTab({ kind: 'globals' }, 'pinned')` (unchanged) |

`openEnvironmentTab` guarantees: at most one tab per target; an existing tab is focused and, for
`'pinned'`, pinned; no draft is saved or discarded.

## `VariableTable` props (after removal)

```ts
interface VariableTableProps {
  scope: string
  label: string
}
```

Drafts always live in the environments store until saved or discarded by the owning tab; the
table renders no Save / Discard footer.
