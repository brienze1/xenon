import { useState, useTransition, type FormEvent } from 'react'
import { useShallow } from 'zustand/react/shallow'
import type { AppError } from '@/platform/errors'
import type { SavedAlias } from '@/platform/mappers/request'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useEnvironmentsStore } from '@/features/environments/store'
import { choiceOf, chosenIds, isChoiceComplete } from '@/features/request-builder/lib/environment-choice'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { EnvironmentPicker } from './EnvironmentPicker'
import styles from './AliasDialog.module.css'

interface AliasEnvironmentsDialogProps {
  tabId: string
  alias: SavedAlias | null
  onClose: () => void
}

/** Limits an alias to some environments, or allows it in all; saved at once, leaving the alias's unsaved edits as they are. */
export function AliasEnvironmentsDialog({ tabId, alias, onClose }: AliasEnvironmentsDialogProps) {
  return (
    <Dialog.Root open={alias !== null} onOpenChange={(open) => (open ? undefined : onClose())}>
      {alias ? <AliasEnvironmentsForm key={alias.id} tabId={tabId} alias={alias} onClose={onClose} /> : null}
    </Dialog.Root>
  )
}

interface AliasEnvironmentsFormProps {
  tabId: string
  alias: SavedAlias
  onClose: () => void
}

function AliasEnvironmentsForm({ tabId, alias, onClose }: AliasEnvironmentsFormProps) {
  const setAliasEnvironments = useRequestBuilderStore((state) => state.setAliasEnvironments)
  const environments = useEnvironmentsStore(useShallow((state) => state.environments))
  const [choice, setChoice] = useState(() => choiceOf(alias.environmentIds))
  const [error, setError] = useState<AppError | null>(null)
  const [saving, startSave] = useTransition()
  const complete = isChoiceComplete(choice, environments)

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!complete) {
      return
    }
    startSave(async () => {
      const failure = await setAliasEnvironments(tabId, alias.id, chosenIds(choice, environments))
      if (failure === null) {
        onClose()
        return
      }
      setError(failure)
    })
  }

  return (
    <form aria-label={`Environments for ${alias.name}`} className={styles.form} noValidate onSubmit={onSubmit}>
      <Dialog.Title>
        Environments for <span translate="no">{alias.name}</span>
      </Dialog.Title>
      <Dialog.Body>
        <EnvironmentPicker choice={choice} storedIds={alias.environmentIds} complete={complete} onChange={setChoice} />
        <p className={styles.hint}>An alias limited to environments can only be sent while one of them is active.</p>
        {error ? (
          <Banner.Root tone="error" className={styles.saveError}>
            <Banner.Message>
              <NoTranslateText text={error.description || error.code} terms={[]} />
            </Banner.Message>
          </Banner.Root>
        ) : null}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={onClose}>
          Cancel
        </Button>
        <Button type="submit" variant="primary" disabled={saving || !complete}>
          {saving ? 'Saving…' : 'Save'}
        </Button>
      </Dialog.Actions>
    </form>
  )
}
