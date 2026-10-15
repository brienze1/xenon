import { Button } from '@/shared/ui/button/Button'
import { Toast } from '@/shared/ui/toast/Toast'
import { useUpdatesStore } from '@/features/updates/store'

/** "Xenon <v> is ready to install." after an automatic download is verified; once per run. */
export function UpdateReadyToast() {
  const shown = useUpdatesStore((state) => state.toast === 'open' && state.status.state === 'ready')
  const version = useUpdatesStore((state) => state.status.latestVersion)
  const restart = useUpdatesStore((state) => state.restart)
  const closeToast = useUpdatesStore((state) => state.closeToast)
  if (!shown) {
    return null
  }
  return (
    <Toast.Root closeLabel="Close" onClose={closeToast}>
      <Toast.Message>
        <span translate="no">Xenon {version}</span> is ready to install.
      </Toast.Message>
      <Toast.Actions>
        <Button variant="primary" size="sm" onClick={() => void restart()}>
          Restart now
        </Button>
        <Button variant="ghost" size="sm" onClick={closeToast}>
          Later
        </Button>
      </Toast.Actions>
    </Toast.Root>
  )
}
