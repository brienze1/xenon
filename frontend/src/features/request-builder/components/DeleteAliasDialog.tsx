import { useState, useTransition } from 'react'
import type { AppError } from '@/platform/errors'
import type { SavedAlias } from '@/platform/mappers/request'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useTabsStore } from '@/features/tabs/store'
import styles from './AliasDialog.module.css'

interface DeleteAliasDialogProps {
  tabId: string
  alias: SavedAlias | null
  requestName: string
  onClose: () => void
}

/** Confirms deleting an alias; its unsaved edits go with it and a tab showing it goes back to the Default. */
export function DeleteAliasDialog({ tabId, alias, requestName, onClose }: DeleteAliasDialogProps) {
  return (
    <Dialog.Root open={alias !== null} onOpenChange={(open) => (open ? undefined : onClose())}>
      {alias ? <DeleteAliasConfirm key={alias.id} tabId={tabId} alias={alias} requestName={requestName} onClose={onClose} /> : null}
    </Dialog.Root>
  )
}

interface DeleteAliasConfirmProps {
  tabId: string
  alias: SavedAlias
  requestName: string
  onClose: () => void
}

function DeleteAliasConfirm({ tabId, alias, requestName, onClose }: DeleteAliasConfirmProps) {
  const deleteAlias = useRequestBuilderStore((state) => state.deleteAlias)
  const pin = useTabsStore((state) => state.pin)
  const [deleting, startDelete] = useTransition()
  const [error, setError] = useState<AppError | null>(null)

  function confirm() {
    startDelete(async () => {
      const failure = await deleteAlias(tabId, alias.id)
      if (failure === null) {
        pin(tabId)
        onClose()
        return
      }
      setError(failure)
    })
  }

  return (
    <>
      <Dialog.Title>Delete alias?</Dialog.Title>
      <Dialog.Body>
        “{alias.name}” will be removed from {requestName || 'the request'}. This can’t be undone.
        {error ? (
          <Banner.Root tone="error" className={styles.saveError}>
            <Banner.Message>
              <NoTranslateText text={error.description || error.code} terms={[error.code]} />
            </Banner.Message>
          </Banner.Root>
        ) : null}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={onClose}>
          Cancel
        </Button>
        <Button variant="danger" disabled={deleting} onClick={confirm}>
          {deleting ? 'Deleting…' : 'Delete'}
        </Button>
      </Dialog.Actions>
    </>
  )
}
