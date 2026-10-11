import { useState } from 'react'
import type { ConsoleEntry, ConsoleHeader } from '@/platform/mappers/console'
import { fetchResponseText } from '@/platform/services/response-body'
import { useConsoleStore } from '@/features/console/store'
import { Button } from '@/shared/ui/button/Button'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { Tabs } from '@/shared/ui/tabs/Tabs'
import { ObjectTree } from './ObjectTree'
import styles from './ConsoleEntryDetail.module.css'

function Headers({ label, headers }: { label: string; headers: ConsoleHeader[] | null }) {
  if (!headers || headers.length === 0) {
    return <p className={styles.muted}>No headers.</p>
  }
  return (
    <dl aria-label={label} className={styles.headers}>
      {headers.map((header, index) => (
        <div key={index} className={styles.header}>
          <dt>{header.key}</dt>
          <dd>{header.value}</dd>
        </div>
      ))}
    </dl>
  )
}

function Body({ label, body }: { label: string; body: string | null }) {
  return body ? (
    <pre aria-label={label} className={styles.body}>
      {body}
    </pre>
  ) : (
    <p className={styles.muted}>No body.</p>
  )
}

function RoutedBody({ url }: { url: string }) {
  const [state, setState] = useState<{ text: string | null; failure: string | null; loading: boolean }>({ text: null, failure: null, loading: false })
  async function load() {
    setState({ text: null, failure: null, loading: true })
    try {
      setState({ text: await fetchResponseText(url), failure: null, loading: false })
    } catch (error) {
      setState({ text: null, failure: error instanceof Error ? error.message : String(error), loading: false })
    }
  }
  if (state.text !== null) {
    return <Body label="Response body" body={state.text} />
  }
  return (
    <div className={styles.routed}>
      <p className={styles.muted}>The body is over 1 MB or binary, so it is not loaded until you ask.</p>
      <Button size="sm" disabled={state.loading} onClick={() => void load()}>
        {state.loading ? 'Loading…' : 'Load Body'}
      </Button>
      {state.failure ? <p role="alert">{state.failure}</p> : null}
    </div>
  )
}

function RequestDetail({ entry }: { entry: ConsoleEntry }) {
  const [panel, setPanel] = useState('request')
  const response = entry.response
  return (
    <Tabs.Root value={panel} onValueChange={setPanel} className={styles.tabs}>
      <Tabs.List label="Request detail">
        <Tabs.Tab value="request">Request</Tabs.Tab>
        <Tabs.Tab value="response">Response</Tabs.Tab>
      </Tabs.List>
      <Tabs.Panel value="request" className={styles.panel}>
        <p className={styles.line} translate="no">
          {entry.request?.method} {entry.request?.url}
        </p>
        <Headers label="Request headers" headers={entry.request?.headers ?? null} />
        <Body label="Request body" body={entry.request?.body ?? null} />
      </Tabs.Panel>
      <Tabs.Panel value="response" className={styles.panel}>
        {response?.error ? (
          <p className={styles.error}>{response.error}</p>
        ) : (
          <>
            <p className={styles.line} translate="no">
              {response?.statusCode} {response?.statusText}
            </p>
            <Headers label="Response headers" headers={response?.headers ?? null} />
            {response?.bodyUrl ? <RoutedBody url={response.bodyUrl} /> : <Body label="Response body" body={response?.body ?? null} />}
          </>
        )}
      </Tabs.Panel>
    </Tabs.Root>
  )
}

function LogDetail({ entry }: { entry: ConsoleEntry }) {
  const values = entry.values ?? entry.arguments
  return (
    <div className={styles.panel}>
      {entry.line !== null ? (
        <p className={styles.muted}>
          Line {entry.line}, column {entry.column ?? 0}
        </p>
      ) : null}
      <div aria-label="Logged values" role="group" className={styles.values}>
        {values.map((value, index) => (
          <ObjectTree key={index} label={String(index)} value={value} />
        ))}
      </div>
    </div>
  )
}

export function ConsoleEntryDetail() {
  const selectedId = useConsoleStore((state) => state.selectedId)
  const detail = useConsoleStore((state) => state.detail)
  const failure = useConsoleStore((state) => state.failure)
  const select = useConsoleStore((state) => state.select)
  if (selectedId === null) {
    return null
  }
  return (
    <section aria-label="Console entry" className={styles.detail}>
      <div className={styles.heading}>
        <h3 className={styles.title}>{detail ? (detail.scriptLabel ?? 'Request') : 'Loading…'}</h3>
        {detail?.masked ? <span className={styles.masked}>Secrets masked</span> : null}
        <IconButton label="Close entry" onClick={() => void select(null)}>
          ×
        </IconButton>
      </div>
      {failure ? (
        <p role="alert" className={styles.error}>
          {failure.message}
        </p>
      ) : detail === null ? null : detail.kind === 'request' ? (
        <RequestDetail entry={detail} />
      ) : (
        <LogDetail entry={detail} />
      )}
    </section>
  )
}
