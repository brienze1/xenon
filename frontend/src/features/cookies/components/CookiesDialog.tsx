import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { Tabs } from '@/shared/ui/tabs/Tabs'
import { useCookiesStore, type CookiesView } from '@/features/cookies/store'
import { AllowlistPanel } from './AllowlistPanel'
import { CookieErrorText } from './CookieErrorText'
import { CookieJarPanel } from './CookieJarPanel'
import styles from './CookiesDialog.module.css'

const COOKIES_TITLE = <Dialog.Title>Cookies</Dialog.Title>

const LOADING = (
  <p role="status" className={styles.empty}>
    Loading cookies…
  </p>
)

const JAR_NEXT_STEP = 'Close the dialog and open it again; if it keeps failing, check the app data folder.'

export function CookiesDialog() {
  const open = useCookiesStore((state) => state.open)
  const close = useCookiesStore((state) => state.close)
  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : close())} className={styles.dialog}>
      {open ? <CookiesContent onClose={close} /> : null}
    </Dialog.Root>
  )
}

interface CookiesContentProps {
  onClose: () => void
}

function CookiesContent({ onClose }: CookiesContentProps) {
  const view = useCookiesStore((state) => state.view)
  const loading = useCookiesStore((state) => state.loading)
  const error = useCookiesStore((state) => state.error)
  const hasCookies = useCookiesStore((state) => state.domains.length > 0)
  const setView = useCookiesStore((state) => state.setView)
  const clearAll = useCookiesStore((state) => state.clearAll)
  return (
    <>
      {COOKIES_TITLE}
      <Dialog.Body>
        {error ? (
          <Banner.Root tone="error">
            <Banner.Message>
              <CookieErrorText error={error} nextStep={JAR_NEXT_STEP} />
            </Banner.Message>
          </Banner.Root>
        ) : null}
        <Tabs.Root value={view} onValueChange={(next) => setView(next as CookiesView)}>
          <Tabs.List label="Cookies">
            <Tabs.Tab value="cookies">Cookies</Tabs.Tab>
            <Tabs.Tab value="allowlist">Allowlist</Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="cookies" className={styles.panel}>
            {loading ? LOADING : <CookieJarPanel />}
          </Tabs.Panel>
          <Tabs.Panel value="allowlist" className={styles.panel}>
            {loading ? LOADING : <AllowlistPanel />}
          </Tabs.Panel>
        </Tabs.Root>
      </Dialog.Body>
      <Dialog.Actions>
        {view === 'cookies' ? (
          <Button variant="danger" disabled={!hasCookies} onClick={() => void clearAll()}>
            Clear All
          </Button>
        ) : null}
        <Button variant="secondary" onClick={onClose}>
          Close
        </Button>
      </Dialog.Actions>
    </>
  )
}
