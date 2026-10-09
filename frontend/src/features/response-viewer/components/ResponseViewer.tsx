import { useEffect, useState } from 'react'
import { Banner } from '@/shared/ui/banner/Banner'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { Tabs } from '@/shared/ui/tabs/Tabs'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { ResponseBody } from './ResponseBody'
import { ResponseError } from './ResponseError'
import { ResponseHeaders } from './ResponseHeaders'
import { StatusLine } from './StatusLine'
import styles from './ResponseViewer.module.css'

interface ResponseViewerProps {
  tabId: string
}

export function ResponseViewer({ tabId }: ResponseViewerProps) {
  const entry = useResponseViewerStore((state) => state.entries[tabId])
  const rehydrate = useResponseViewerStore((state) => state.rehydrate)
  const [panel, setPanel] = useState('body')

  useEffect(() => {
    void rehydrate(tabId)
  }, [tabId, rehydrate])

  const result = entry?.result ?? null

  return (
    <section aria-label="Response" className={styles.viewer}>
      {entry?.failure ? (
        <Banner.Root tone="error">
          <Banner.Message>{entry.failure.description || entry.failure.code}</Banner.Message>
        </Banner.Root>
      ) : null}
      {entry?.sending ? (
        <p role="status" className={styles.sending}>
          Sending request…
        </p>
      ) : result === null ? (
        <EmptyState.Root label="No response yet">
          <EmptyState.Art variant="response" />
          <EmptyState.Description>Enter the URL and click Send to get a response.</EmptyState.Description>
        </EmptyState.Root>
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
