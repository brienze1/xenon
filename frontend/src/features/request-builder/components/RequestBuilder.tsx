import { useId, type ReactNode } from 'react'
import { isCurlRefusal } from '@/platform/services/requests'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import type { RequestBuilderMode } from '@/features/request-builder/model'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { ConfirmSendDialog } from './ConfirmSendDialog'
import { CurlPasteNotice } from './CurlPasteNotice'
import { MissingSecretsDialog } from './MissingSecretsDialog'
import { RequestBar } from './RequestBar'
import styles from './RequestBuilder.module.css'
import { RequestTabs } from './RequestTabs'

const SAVE_NEXT_STEP = 'Try again, or check that the request file and its folder are writable.'

const CURL_NEXT_STEP = 'Check the cURL command and paste it again.'

const CONFLICT_MESSAGE = (
  <Banner.Message>This request changed on disk since you opened it. Keep your version or load the one from disk.</Banner.Message>
)

interface RequestBuilderProps {
  tabId: string
  onSave: () => void
  /** The Scripts tab's editor, composed by the app. */
  scripts?: ReactNode
  mode?: RequestBuilderMode
}

export function RequestBuilder({ tabId, onSave, scripts, mode = 'request' }: RequestBuilderProps) {
  const error = useRequestBuilderStore((state) => state.tabs[tabId]?.error ?? null)
  const conflict = useRequestBuilderStore((state) => state.tabs[tabId]?.conflict === true)
  const save = useRequestBuilderStore((state) => state.save)
  const openSaved = useRequestBuilderStore((state) => state.openSaved)
  const headingId = useId()

  function loadFromDisk() {
    const requestId = useRequestBuilderStore.getState().tabs[tabId]?.draft.id ?? ''
    void openSaved(tabId, requestId)
  }

  return (
    <section aria-labelledby={headingId} className={styles.builder}>
      <h2 id={headingId} className="visually-hidden">
        {mode === 'example' ? 'Example Request' : 'Request'}
      </h2>
      <RequestBar tabId={tabId} onSave={onSave} mode={mode} />
      {mode === 'request' ? <CurlPasteNotice tabId={tabId} /> : null}
      {mode === 'request' && conflict ? (
        <Banner.Root tone="warning" className={styles.banner}>
          {CONFLICT_MESSAGE}
          <Banner.Actions>
            <Button variant="secondary" size="sm" onClick={() => void save(tabId, true)}>
              Keep Your Version
            </Button>
            <Button variant="ghost" size="sm" onClick={loadFromDisk}>
              Load from Disk
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      {mode === 'request' && error ? (
        <Banner.Root tone="error" className={styles.banner}>
          <Banner.Message>
            <NoTranslateText text={withNextStep(error.description || error.code, isCurlRefusal(error) ? CURL_NEXT_STEP : SAVE_NEXT_STEP)} terms={[error.code]} />
          </Banner.Message>
        </Banner.Root>
      ) : null}
      <RequestTabs tabId={tabId} mode={mode} scripts={scripts} />
      <ConfirmSendDialog tabId={tabId} />
      <MissingSecretsDialog tabId={tabId} />
    </section>
  )
}
