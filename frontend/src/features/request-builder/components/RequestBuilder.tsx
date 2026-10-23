import { useCallback, useId, type ReactNode } from 'react'
import { isCurlRefusal } from '@/platform/services/requests'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { DecorationCardContext, type RenderDecorationCard } from '@/shared/ui/code-editor/decoration-card-context'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import type { RequestBuilderMode } from '@/features/request-builder/model'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { AliasMenu } from './AliasMenu'
import { ConfirmSendDialog } from './ConfirmSendDialog'
import { CurlPasteNotice } from './CurlPasteNotice'
import { ExpiredTokenDialog } from './ExpiredTokenDialog'
import { MissingSecretsDialog } from './MissingSecretsDialog'
import { RequestBar } from './RequestBar'
import { UnavailableAliasBanner } from './UnavailableAliasBanner'
import styles from './RequestBuilder.module.css'
import { RequestTabs } from './RequestTabs'
import { VariableCard } from './VariableCard'

const SAVE_NEXT_STEP = 'Try again, or check that the request file and its folder are writable.'

const CURL_NEXT_STEP = 'Check the cURL command and paste it again.'

const CONFLICT_MESSAGE = (
  <Banner.Message>This request changed on disk since you opened it. Keep your version or load the one from disk.</Banner.Message>
)

interface RequestBuilderProps {
  tabId: string
  /** The Scripts tab's editor, composed by the app. */
  scripts?: ReactNode
  /** The Authorization tab's editor, composed by the app. */
  auth?: ReactNode
  /** Where the request lives, shown above the URL bar; composed by the app. */
  breadcrumb?: ReactNode
  mode?: RequestBuilderMode
}

export function RequestBuilder({ tabId, scripts, auth, breadcrumb, mode = 'request' }: RequestBuilderProps) {
  const error = useRequestBuilderStore((state) => state.tabs[tabId]?.error ?? null)
  const conflict = useRequestBuilderStore((state) => state.tabs[tabId]?.conflict === true)
  const isSaved = useRequestBuilderStore((state) => (state.tabs[tabId]?.draft.id ?? '') !== '')
  const save = useRequestBuilderStore((state) => state.save)
  const openSaved = useRequestBuilderStore((state) => state.openSaved)
  const headingId = useId()
  const renderCard = useCallback<RenderDecorationCard>((name, close) => <VariableCard tabId={tabId} name={name} close={close} />, [tabId])

  function loadFromDisk() {
    const requestId = useRequestBuilderStore.getState().tabs[tabId]?.draft.id ?? ''
    void openSaved(tabId, requestId)
  }

  return (
    <DecorationCardContext value={renderCard}>
      <section aria-labelledby={headingId} className={styles.builder}>
        <h2 id={headingId} className="visually-hidden">
          {mode === 'example' ? 'Example Request' : 'Request'}
        </h2>
        <div className={styles.header}>
          {breadcrumb}
          {mode === 'request' && isSaved ? <AliasMenu tabId={tabId} /> : null}
        </div>
        <RequestBar tabId={tabId} mode={mode} />
        {mode === 'request' ? <UnavailableAliasBanner tabId={tabId} className={styles.banner} /> : null}
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
        <RequestTabs tabId={tabId} mode={mode} scripts={scripts} auth={auth} />
        <ConfirmSendDialog tabId={tabId} />
        <MissingSecretsDialog tabId={tabId} />
        <ExpiredTokenDialog tabId={tabId} />
      </section>
    </DecorationCardContext>
  )
}
