import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useRequestBuilderStore } from '@/features/request-builder/store'

interface ExpiredTokenDialogProps {
  tabId: string
}

export function ExpiredTokenDialog({ tabId }: ExpiredTokenDialogProps) {
  const owner = useRequestBuilderStore((state) => state.tabs[tabId]?.expiredTokenOwner ?? null)
  const send = useRequestBuilderStore((state) => state.send)
  const dismiss = useRequestBuilderStore((state) => state.dismissExpiredToken)
  const setSubTab = useRequestBuilderStore((state) => state.setSubTab)
  const close = () => dismiss(tabId)

  function getNewToken() {
    dismiss(tabId)
    setSubTab(tabId, 'auth')
  }

  function sendAnyway() {
    const acknowledgement = useRequestBuilderStore.getState().tabs[tabId]?.acknowledgement
    dismiss(tabId)
    void send(tabId, { ...acknowledgement, expiredToken: true })
  }

  return (
    <Dialog.Root open={owner !== null} onOpenChange={(open) => (open ? undefined : close())}>
      {owner !== null ? (
        <>
          <Dialog.Title>Access Token Expired</Dialog.Title>
          <Dialog.Body>
            <p>The OAuth 2.0 access token expired and could not be refreshed. Get a new token, or send the request with the expired one.</p>
          </Dialog.Body>
          <Dialog.Actions>
            <Button variant="secondary" onClick={close}>
              Cancel
            </Button>
            <Button variant="secondary" onClick={sendAnyway}>
              Send Anyway
            </Button>
            <Button variant="primary" onClick={getNewToken}>
              Get New Access Token
            </Button>
          </Dialog.Actions>
        </>
      ) : null}
    </Dialog.Root>
  )
}
