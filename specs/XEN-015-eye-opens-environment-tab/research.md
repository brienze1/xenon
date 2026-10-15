# Research: The Eye Opens the Environment Tab

All decisions below were settled from the code on `origin/main` (06d0d82) or by the user on
2026-10-15. Nothing is left open.

## R1 — How the switcher opens a tab

- **Decision**: `EnvironmentSwitcher` gets a new prop `onOpenEnvironment: (environmentId: string) => void`.
  The eye calls `onOpenEnvironment(active.id)` when an environment is active, else the existing
  `onOpenGlobals()`. `app/App.tsx` passes
  `(environmentId) => openEnvironmentTab({ kind: 'environment', environmentId }, 'pinned')`
  (its module-level `openEnvironmentTab` helper already exists, and `openGlobals` already opens
  `{ kind: 'globals' }` pinned).
- **Rationale**: `frontend-architecture` rule 2 forbids `features/environments` from importing
  `features/tabs/model.ts` (where `EnvironmentTabTarget` lives); callbacks composed in `app/` is
  the switcher's established pattern (`onOpenGlobals`, `onManage`). It also keeps the switcher
  testable with `vi.fn()` like the existing "Globals" menu test.
- **Alternatives**: calling `useTabsStore.getState().openEnvironmentTab(...)` from the switcher
  (allowed via `store.ts`, and the Quick Look already imported `features/tabs/store`) — rejected
  because the switcher would then hard-code the tab mode and duplicate `App.tsx`'s globals
  wiring; one place (`App.tsx`) decides how environment targets open.

## R2 — Open-or-focus, pinning, no duplicates

- **Decision**: no tab-store change. `openEnvironmentTab(target, 'pinned')` already finds an
  existing tab for the target (`isTargetOf`), pins it if needed and focuses it; otherwise it adds
  a pinned tab. Unsaved edits are untouched (it never calls discard or save).
- **Rationale**: verified in `features/tabs/store.ts` (`openEnvironmentTab`) and covered by
  `features/tabs/environment-tabs.test.ts`.

## R3 — Label and glyph

- **Decision**: `label={active ? 'Open environment' : 'Open globals'}` on the existing
  `IconButton` (which sets both `aria-label` and `title`). The eye SVG and its `.icon` CSS move
  unchanged from `QuickLook.tsx` / `QuickLook.module.css` into `EnvironmentSwitcher.tsx` /
  `.module.css`. `active` is the switcher's existing `environments.find(... activeId)`.
- **Rationale**: user decision (labels); the glyph and position stay (FR-009). Sentence case
  matches the replaced label "Quick look at variables".

## R4 — Tests instead of BDD

- **Decision**: red-first Vitest cases, no `.feature` file.
  - `EnvironmentSwitcher.test.tsx`: eye with active env → `onOpenEnvironment('e1')` and not
    `onOpenGlobals`, no dialog; eye with no env → `onOpenGlobals` and not `onOpenEnvironment`;
    label follows the active environment.
  - `App.test.tsx`: in the real app, the eye opens a pinned, focused env tab; a second click adds
    no tab; an existing preview tab gets pinned; with no env the Globals tab opens.
- **Rationale**: Constitution V allows Vitest as the frontend gate; Go is untouched. Precedent:
  XEN-010.

## R5 — Code only the Quick Look used (removed)

Verified with `grep` over `frontend/src`:

| Code | Other users | Decision |
| --- | --- | --- |
| `QuickLook.tsx`, `.module.css`, `.test.tsx` | `EnvironmentSwitcher.tsx` only | delete |
| `hooks/use-discard-guard.ts` | `QuickLook.tsx` only | delete (folder becomes empty) |
| `DiscardChangesConfirm.tsx`, `.module.css` | `QuickLook.tsx` only | delete |
| `VariableTable` `draftLifetime` (`'view'` default) | `EnvironmentTab`, `GlobalsTab`, `ItemWorkspace` all pass `'workspace'` | remove prop + `DraftLifetime` type + unmount-discard effect; tables always keep drafts |
| `VariableTable` `savedByTab` (`false` default → Save/Discard footer) | same three all pass `savedByTab` | user decision: remove prop, footer, `requestSave`, `discard`, the editor's `discardDraft` selector and `.footer` CSS; table is always saved by its tab |

Kept: `confirmSave` / `afterSave` / `focusFirstInvalidName` and `TypeSwitchConfirm` — a tab's
Cmd+S goes through the store's `saveDraft`, which can set `draft.confirming`, and the table still
renders that confirmation.

## R6 — Test rewrites forced by R5

- `VariableTable.test.tsx`: tests that clicked "Save Variables" / "Discard Changes" save and
  discard through `useEnvironmentsStore.getState().saveDraft(SCOPE)` / `discardDraft(SCOPE)`
  inside `act`, exactly what `features/tabs/store.ts` does on Cmd+S / Don't Save. Assertions that
  the footer button is enabled again become assertions on the store/draft (e.g. a second save of
  an unchanged table does not call the service). The "clears its unsaved-edits flag when it
  closes" test is deleted (that behaviour is removed); the "keeps unsaved edits through an
  unmount" test drops its `draftLifetime` prop and the "Discard Changes" assertion. The
  "moves focus to the first invalid name after a failed save" test goes through the confirm path
  (save → type-switch confirmation → confirm → service rejects `VAR-00400` → focus), since that
  is the remaining caller of `focusFirstInvalidName`.
- `EnvironmentTab.test.tsx`: the "no Save/Discard buttons" assertions stay valid unchanged.
- `QuickLook.test.tsx` is deleted with the component.
