import { useEffect, useId, useState } from 'react'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { Tabs } from '@/shared/ui/tabs/Tabs'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { ResponseBody } from './ResponseBody'
import { ResponseError } from './ResponseError'
import { ResponseHeaders } from './ResponseHeaders'
import { StatusLine } from './StatusLine'
import styles from './ResponseViewer.module.css'

const SEND_NEXT_STEP = 'Check the URL and your connection, then send the request again.'

const EMPTY_RESPONSE = (
  <EmptyState.Root label="No response yet">
    <EmptyState.Art variant="response" />
    <EmptyState.Description>Enter the URL and click Send to get a response.</EmptyState.Description>
  </EmptyState.Root>
)

interface ResponseViewerProps {
  tabId: string
}

export function ResponseViewer({ tabId }: ResponseViewerProps) {
  const entry = useResponseViewerStore((state) => state.entries[tabId])
  const rehydrate = useResponseViewerStore((state) => state.rehydrate)
  const [panel, setPanel] = useState('body')
  const headingId = useId()

  useEffect(() => {
    void rehydrate(tabId)
  }, [tabId, rehydrate])

  const result = entry?.result ?? null

  return (
    <section aria-labelledby={headingId} className={styles.viewer}>
      <h2 id={headingId} className="visually-hidden">
        Response
      </h2>
      {entry?.failure ? (
        <Banner.Root tone="error">
          <Banner.Message>{withNextStep(entry.failure.description || entry.failure.code, SEND_NEXT_STEP)}</Banner.Message>
        </Banner.Root>
      ) : null}
      {entry?.sending ? (
        <p role="status" className={styles.sending}>
          Sending request…
        </p>
      ) : result === null ? (
        EMPTY_RESPONSE
      ) : result.error ? (
        <ResponseError failure={result.error} />
      ) : (
        <>
          <StatusLine result={result} />
          <Tabs.Root value={panel} onValueChange={setPanel} className={styles.tabs}>
            <Tabs.List label="Response">
              <Tabs.Tab value="body">Body</Tabs.Tab>
              <Tabs.Tab value="headers">Headers ({result.headers.length})</Tabs.Tab>
            </Tabs.List>
            <Tabs.Panel value="body" className={styles.panel}>
              <ResponseBody key={entry?.version} tabId={tabId} result={result} />
            </Tabs.Panel>
            <Tabs.Panel value="headers" className={styles.panel}>
              <ResponseHeaders headers={result.headers} />
            </Tabs.Panel>
          </Tabs.Root>
        </>
      )}
    </section>
  )
}
