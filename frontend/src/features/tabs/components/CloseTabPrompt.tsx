import { useRequestBuilderStore } from '@/features/request-builder/store'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useTabsStore } from '../store'

export function CloseTabPrompt() {
  const tabId = useTabsStore((state) => state.closePrompt)
  const resolveClose = useTabsStore((state) => state.resolveClose)
  const name = useRequestBuilderStore((state) => (tabId ? state.tabs[tabId]?.draft.name : null) || 'Untitled Request')

  return (
    <Dialog.Root open={tabId !== null} onOpenChange={(open) => (open ? undefined : void resolveClose('cancel'))}>
      <Dialog.Title>Save changes?</Dialog.Title>
      <Dialog.Body>“{name}” has unsaved changes. Do you want to save them before closing?</Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={() => void resolveClose('cancel')}>
          Cancel
        </Button>
        <Button variant="secondary" onClick={() => void resolveClose('discard')}>
          Don&apos;t save
        </Button>
        <Button variant="primary" onClick={() => void resolveClose('save')}>
          Save
        </Button>
      </Dialog.Actions>
    </Dialog.Root>
  )
}
