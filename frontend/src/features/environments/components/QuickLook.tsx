import { useState } from 'react'
import { environmentScope, GLOBALS_SCOPE } from '@/platform/mappers/variable'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { useDiscardGuard } from '@/features/environments/hooks/use-discard-guard'
import { useEnvironmentsStore } from '@/features/environments/store'
import { DiscardChangesConfirm } from './DiscardChangesConfirm'
import styles from './QuickLook.module.css'
import { VariableTable } from './VariableTable'

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
  const scopes = active ? [environmentScope(active.id), GLOBALS_SCOPE] : [GLOBALS_SCOPE]
  const { confirming, guard, discard, keepEditing } = useDiscardGuard(scopes)
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
          {active ? <VariableTable scope={environmentScope(active.id)} label={`${active.name} variables`} /> : NO_ACTIVE_ENVIRONMENT}
        </section>
        <section className={styles.section} aria-label="Global variables">
          <h3 className={styles.heading}>Globals</h3>
          <VariableTable scope={GLOBALS_SCOPE} label="Global variables" />
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
