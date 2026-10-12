import { useState } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { environmentScope, GLOBALS_SCOPE } from '@/platform/mappers/variable'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { useDiscardGuard } from '@/features/environments/hooks/use-discard-guard'
import { useEnvironmentsStore } from '@/features/environments/store'
import { environmentTabScope, useTabsStore } from '@/features/tabs/store'
import { DiscardChangesConfirm } from './DiscardChangesConfirm'
import styles from './QuickLook.module.css'
import { VariableTable, type DraftLifetime } from './VariableTable'

const EYE_GLYPH = (
  <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.icon}>
    <path d="M2.06 12.35a1 1 0 0 1 0-.7 10.75 10.75 0 0 1 19.88 0 1 1 0 0 1 0 .7 10.75 10.75 0 0 1-19.88 0" />
    <circle cx="12" cy="12" r="3" />
  </svg>
)

const NO_ACTIVE_ENVIRONMENT = <p className={styles.note}>No environment is active. Pick one in the switcher to see its variables here.</p>

/** The eye next to the switcher: the active environment's variables and the globals, editable in place. */
export function QuickLook() {
  const [open, setOpen] = useState(false)

  return (
    <>
      <IconButton label="Quick look at variables" onClick={() => setOpen(true)}>
        {EYE_GLYPH}
      </IconButton>
      {open ? <QuickLookContent onClose={() => setOpen(false)} /> : null}
    </>
  )
}

interface QuickLookContentProps {
  onClose: () => void
}

function QuickLookContent({ onClose }: QuickLookContentProps) {
  const active = useEnvironmentsStore((state) => state.environments.find((environment) => environment.id === state.activeId) ?? null)
  const tabScopes = useTabsStore(useShallow((state) => state.tabs.flatMap((tab) => environmentTabScope(tab) ?? [])))
  const scopes = active ? [environmentScope(active.id), GLOBALS_SCOPE] : [GLOBALS_SCOPE]
  // A table also open in a tab keeps its edits for that tab, so closing the quick look does not drop them.
  const { confirming, guard, discard, keepEditing } = useDiscardGuard(scopes.filter((scope) => !tabScopes.includes(scope)))
  const lifetime = (scope: string): DraftLifetime => (tabScopes.includes(scope) ? 'workspace' : 'view')
  const close = () => guard(onClose)

  return (
    <Dialog.Root open onOpenChange={(next) => (next ? undefined : close())} className={styles.dialog}>
      <Dialog.Title>Quick Look</Dialog.Title>
      <Dialog.Body>
        <section className={styles.section} aria-label="Active environment variables">
          <h3 className={styles.heading}>
            {active ? (
              <>
                <span className={styles.swatch} data-color={active.color} aria-hidden="true" />
                <span className={styles.headingName} translate="no" title={active.name}>
                  {active.name}
                </span>
              </>
            ) : (
              'Environment'
            )}
          </h3>
          {active ? <VariableTable scope={environmentScope(active.id)} label={`${active.name} variables`} draftLifetime={lifetime(environmentScope(active.id))} /> : NO_ACTIVE_ENVIRONMENT}
        </section>
        <section className={styles.section} aria-label="Global variables">
          <h3 className={styles.heading}>Globals</h3>
          <VariableTable scope={GLOBALS_SCOPE} label="Global variables" draftLifetime={lifetime(GLOBALS_SCOPE)} />
        </section>
      </Dialog.Body>
      <Dialog.Actions>
        {confirming ? (
          <DiscardChangesConfirm onKeepEditing={keepEditing} onDiscard={discard} />
        ) : (
          <Button variant="secondary" onClick={close}>
            Close
          </Button>
        )}
      </Dialog.Actions>
    </Dialog.Root>
  )
}
