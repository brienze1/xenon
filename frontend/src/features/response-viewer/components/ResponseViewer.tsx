import { useEffect, useId, useState } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { withNextStep } from '@/shared/lib/next-step'
import { formatNumber } from '@/shared/lib/number-format'
import { Banner } from '@/shared/ui/banner/Banner'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { BoltRingArt } from '@/shared/ui/empty-state/illustrations/BoltRingArt'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { Tabs } from '@/shared/ui/tabs/Tabs'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { ResponseBody } from './ResponseBody'
import { ResponseError } from './ResponseError'
import { ResponseHeaders } from './ResponseHeaders'
import { SaveAsExampleButton } from './SaveAsExampleButton'
import { ScriptErrorPanel } from './ScriptErrorPanel'
import { StatusLine } from './StatusLine'
import { TestResultsTab } from './TestResultsTab'
import { testResultsTitle } from '@/features/response-viewer/lib/tests'
import styles from './ResponseViewer.module.css'

const SEND_NEXT_STEP = 'Check the URL and your connection, then send the request again.'

const EMPTY_RESPONSE = (
  <EmptyState.Root label="No response yet">
    <EmptyState.Art>
      <BoltRingArt />
    </EmptyState.Art>
    <EmptyState.Description>Enter the URL and click Send to get a response.</EmptyState.Description>
  </EmptyState.Root>
)

const SENDING = (
  <p role="status" className={styles.sending}>
    Sending request…
  </p>
)

const BODY_TAB = <Tabs.Tab value="body">Body</Tabs.Tab>

interface ResponseViewerProps {
  tabId: string
}

export function ResponseViewer({ tabId }: ResponseViewerProps) {
  const { failure, sending, version, result } = useResponseViewerStore(
    useShallow((state) => {
      const entry = state.entries[tabId]
      return {
        failure: entry?.failure ?? null,
        sending: entry?.sending ?? false,
        version: entry?.version,
        result: entry?.result ?? null,
      }
    }),
  )
  const rehydrate = useResponseViewerStore((state) => state.rehydrate)
  const [panel, setPanel] = useState('body')
  const headingId = useId()

  useEffect(() => {
    void rehydrate(tabId)
  }, [tabId, rehydrate])

  return (
    <section aria-labelledby={headingId} className={styles.viewer}>
      <h2 id={headingId} className="visually-hidden">
        Response
      </h2>
      {failure ? (
        <Banner.Root tone="error">
          <Banner.Message>
            <NoTranslateText text={withNextStep(failure.description || failure.code, SEND_NEXT_STEP)} terms={[failure.code]} />
          </Banner.Message>
        </Banner.Root>
      ) : null}
      <p aria-live="polite" className="visually-hidden">
        {!sending && result !== null && !result.error ? `Response received: ${result.statusCode} ${result.statusText}` : ''}
      </p>
      {sending ? (
        SENDING
      ) : result === null ? (
        EMPTY_RESPONSE
      ) : result.outcome === 'stopped_by_script_error' || result.outcome === 'skipped_by_script' ? (
        <>
          <ScriptErrorPanel outcome={result.outcome} errors={result.scriptErrors} />
          {result.tests.length > 0 ? <TestResultsTab tests={result.tests} summary={result.testSummary} /> : null}
        </>
      ) : result.error ? (
        <ResponseError failure={result.error} />
      ) : (
        <>
          <div className={styles.statusBar}>
            <StatusLine result={result} />
            <SaveAsExampleButton tabId={tabId} />
          </div>
          {result.scriptErrors.length > 0 ? <ScriptErrorPanel outcome={result.outcome} errors={result.scriptErrors} /> : null}
          <Tabs.Root value={panel} onValueChange={setPanel} className={styles.tabs}>
            <Tabs.List label="Response">
              {BODY_TAB}
              <Tabs.Tab value="headers">Headers ({formatNumber(result.headers.length, 0)})</Tabs.Tab>
              <Tabs.Tab value="tests">{testResultsTitle(result.testSummary)}</Tabs.Tab>
            </Tabs.List>
            <Tabs.Panel value="body" className={styles.panel}>
              <ResponseBody key={version} tabId={tabId} result={result} />
            </Tabs.Panel>
            <Tabs.Panel value="headers" className={styles.panel}>
              <ResponseHeaders headers={result.headers} />
            </Tabs.Panel>
            <Tabs.Panel value="tests" className={styles.panel}>
              <TestResultsTab key={version} tests={result.tests} summary={result.testSummary} />
            </Tabs.Panel>
          </Tabs.Root>
        </>
      )}
    </section>
  )
}
