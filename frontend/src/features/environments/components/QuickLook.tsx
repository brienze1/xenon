import { useState } from 'react'
import { environmentScope, GLOBALS_SCOPE } from '@/platform/mappers/variable'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { useEnvironmentsStore } from '@/features/environments/store'
import styles from './QuickLook.module.css'
import { VariableTable } from './VariableTable'

const EYE_GLYPH = <span aria-hidden="true">◉</span>

const NO_ACTIVE_ENVIRONMENT = <p className={styles.note}>No environment is active. Pick one in the switcher to see its variables here.</p>

/** The eye next to the switcher: the active environment's variables and the globals, editable in place. */
export function QuickLook() {
  const [open, setOpen] = useState(false)
  const active = useEnvironmentsStore((state) => state.environments.find((environment) => environment.id === state.activeId) ?? null)

  return (
    <>
      <IconButton label="Quick look at variables" onClick={() => setOpen(true)}>
        {EYE_GLYPH}
      </IconButton>
      <Dialog.Root open={open} onOpenChange={setOpen} className={styles.dialog}>
        {open ? (
          <>
            <Dialog.Title>Quick Look</Dialog.Title>
            <Dialog.Body>
              <section className={styles.section} aria-label="Active environment variables">
                <h3 className={styles.heading}>
                  {active ? (
                    <>
                      <span className={styles.swatch} data-color={active.color} aria-hidden="true" />
                      <span translate="no">{active.name}</span>
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
              <Button variant="secondary" onClick={() => setOpen(false)}>
                Close
              </Button>
            </Dialog.Actions>
          </>
        ) : null}
      </Dialog.Root>
    </>
  )
}
