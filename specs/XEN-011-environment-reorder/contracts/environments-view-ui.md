# Contract: Environments view and switcher UI (XEN-011)

Files: `frontend/src/features/sidebar/components/EnvironmentsView.tsx` (+ `.module.css`,
`.test.tsx`), `frontend/src/features/environments/store.ts` (+ `store.test.ts`),
`frontend/src/features/environments/components/EnvironmentSwitcher.test.tsx`.

## Store — `useEnvironmentsStore`

```ts
/** Moves the environment to `index` (0 = first environment) at once, then saves; restores and returns the error on failure. */
reorderEnvironment: (id: string, index: number) => Promise<AppError | null>
```

1. If `id` is unknown or the move changes nothing → return `null`, no call.
2. `set({ environments: moved })` immediately (optimistic).
3. `inWorkspace((localId) => reorderEnvironment(localId, id, index))`.
4. Success → `set({ environments: list.environments, activeId: list.activeId })`, return `null`.
5. Failure → `set({ environments: previous })`, return the `AppError`.

`index` counts readable environments only (broken rows are always after them).

## Environments view

- `Globals` row: outside the sortable context; never draggable; not a drop target.
- Readable environment rows: one `@dnd-kit/sortable` item each (`useSortable({ id, disabled })`)
  inside `SortableContext` with `verticalListSortingStrategy`.
- `disabled` is true when: the search box is non-empty, the row is being renamed, or the "new
  environment" input row is showing.
- Broken rows and the "new environment" row are not sortable items.
- Sensors: `PointerSensor` (activation distance 5 px, as the tab strip and tree) and
  `KeyboardSensor` with `sortableKeyboardCoordinates` and the tab strip's key codes
  (`start: Space`, `end: Space|Enter`, `cancel: Escape`). Enter without a drag in progress keeps
  opening the environment (Cmd/Ctrl+Enter pins) as today. Space picks the row up and must not
  also open the environment (the row button's Space click is suppressed while a keyboard drag is
  active or starting).
- On drag end over a different row: index = the target row's index among readable environments;
  call `reorderEnvironment(activeId, index)`; on a returned error show it in the view's existing
  error banner (`SidebarErrorText` with `ACTION_NEXT_STEP`).
- Text selection is suspended during a drag (`suspendTextSelection`), as in the tab strip.
- Screen-reader instructions: "To reorder an environment, press Space to pick it up, use the
  arrow keys to move it, then press Space again to drop it or Escape to cancel."
- Announcements (positions are 1-based among readable environments):
  - start: `Picked up {name}.`
  - over: `{name} moved to position {n} of {total}.`
  - end: `{name} dropped at position {n} of {total}.`
  - cancel: `Reordering cancelled. {name} is back at position {n} of {total}.`
- While dragging, the other rows slide to show where the dragged row will land (sortable list,
  as the tab strip), the dragged row gets the `dragging` style, and the pointer shows a grab
  cursor on draggable rows only; motion follows the global `prefers-reduced-motion` rule in
  `shared/styles/global.css`.
- After a keyboard drop or cancel, focus stays on the moved row.
- The dragged row keeps its context menu and click behaviour when not dragging.

## Switcher

No code change: it renders `useEnvironmentsStore().environments` in store order. A test asserts
the menu order follows the store order (not alphabetical) after a reorder.
