import { useId, useState, type ReactNode } from 'react'
import type { KeyValueRow } from '@/platform/mappers/request'
import { withNextStep } from '@/shared/lib/next-step'
import { formatNumber } from '@/shared/lib/number-format'
import { addRow, patchRow, removeRow } from '@/shared/lib/rows'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { KeyValueTable, type KeyValueExamples } from '@/shared/ui/key-value-table/KeyValueTable'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { Tabs } from '@/shared/ui/tabs/Tabs'
import { useExamplesStore } from '@/features/examples/store'
import { ExampleBody } from './ExampleBody'
import styles from './ExampleTab.module.css'

const SAVE_NEXT_STEP = 'Try again, or check that the request file and its folder are writable.'
const LOAD_NEXT_STEP = 'Close this tab and open the example again from the sidebar.'

const CONFLICT_MESSAGE = (
  <Banner.Message>The request file changed on disk since you opened this example. Keep your version or load the one from disk.</Banner.Message>
)

const LOADING = (
  <p role="status" className={styles.note}>
    Loading example…
  </p>
)

const HEADER_EXAMPLES: KeyValueExamples = { key: 'Content-Type', value: 'application/json', description: '' }

const NO_HEADERS: KeyValueRow[] = []

const BODY_TAB = <Tabs.Tab value="body">Body</Tabs.Tab>

const STATUS_CODE_LABEL = <span className={styles.label}>Status code</span>

const STATUS_TEXT_LABEL = <span className={styles.label}>Status text</span>

interface ExampleTabProps {
  tabId: string
  /** The editors of the example's request snapshot. */
  requestEditor: ReactNode
}

export function ExampleTab({ tabId, requestEditor }: ExampleTabProps) {
  const status = useExamplesStore((state) => state.tabs[tabId]?.status ?? 'loading')
  const error = useExamplesStore((state) => state.tabs[tabId]?.error ?? null)
  const conflict = useExamplesStore((state) => state.tabs[tabId]?.conflict === true)
  const save = useExamplesStore((state) => state.save)
  const open = useExamplesStore((state) => state.open)
  const headingId = useId()

  function loadFromDisk() {
    const tab = useExamplesStore.getState().tabs[tabId]
    if (tab) {
      void open(tabId, tab.requestId, tab.exampleId)
    }
  }

  return (
    <div className={styles.tab}>
      {conflict ? (
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
      {error ? (
        <Banner.Root tone="error" className={styles.banner}>
          <Banner.Message>
            <NoTranslateText text={withNextStep(error.description || error.code, status === 'error' ? LOAD_NEXT_STEP : SAVE_NEXT_STEP)} terms={[error.code]} />
          </Banner.Message>
        </Banner.Root>
      ) : null}
      {status === 'loading' ? (
        LOADING
      ) : status === 'error' ? null : (
        <>
          <div className={styles.request}>{requestEditor}</div>
          <section aria-labelledby={headingId} className={styles.response}>
            <h2 id={headingId} className={styles.heading}>
              Example Response
            </h2>
            <ExampleResponseEditor tabId={tabId} />
          </section>
        </>
      )}
    </div>
  )
}

interface ExampleResponseEditorProps {
  tabId: string
}

function ExampleResponseEditor({ tabId }: ExampleResponseEditorProps) {
  const response = useExamplesStore((state) => state.tabs[tabId]?.response ?? null)
  const bulkHeaders = useExamplesStore((state) => state.tabs[tabId]?.bulkHeaders ?? null)
  const setCode = useExamplesStore((state) => state.setCode)
  const setStatusText = useExamplesStore((state) => state.setStatusText)
  const setHeaders = useExamplesStore((state) => state.setHeaders)
  const setBulkHeaders = useExamplesStore((state) => state.setBulkHeaders)
  const setBulkHeadersText = useExamplesStore((state) => state.setBulkHeadersText)
  const [panel, setPanel] = useState('body')
  if (!response) {
    return null
  }
  const headers = response.headers ?? NO_HEADERS

  return (
    <>
      <div className={styles.statusLine}>
        <label className={styles.field}>
          {STATUS_CODE_LABEL}
          <input
            type="number"
            name="example-status-code"
            inputMode="numeric"
            min={100}
            max={599}
            className={styles.codeInput}
            value={response.code}
            onChange={(event) => setCode(tabId, Number(event.target.value))}
          />
        </label>
        <label className={styles.field}>
          {STATUS_TEXT_LABEL}
          <input
            name="example-status-text"
            autoComplete="off"
            className={styles.textInput}
            value={response.status}
            onChange={(event) => setStatusText(tabId, event.target.value)}
          />
        </label>
        {response.contentType ? (
          <span className={styles.contentType} translate="no">
            {response.contentType}
          </span>
        ) : null}
      </div>
      <Tabs.Root value={panel} onValueChange={setPanel} className={styles.tabs}>
        <Tabs.List label="Example response">
          {BODY_TAB}
          <Tabs.Tab value="headers">Headers ({formatNumber(headers.length, 0)})</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="body" className={styles.panel}>
          <ExampleBody tabId={tabId} response={response} />
        </Tabs.Panel>
        <Tabs.Panel value="headers" className={styles.panel}>
          <KeyValueTable
            examples={HEADER_EXAMPLES}
            label="Response headers"
            rows={headers}
            mode={bulkHeaders === null ? 'table' : 'bulk'}
            bulkText={bulkHeaders ?? ''}
            onModeChange={(mode) => setBulkHeaders(tabId, mode === 'bulk')}
            onBulkTextChange={(text) => setBulkHeadersText(tabId, text)}
            onRowChange={(id, patch) => setHeaders(tabId, patchRow(headers, id, patch))}
            onRowAdd={(row) => setHeaders(tabId, addRow(headers, row))}
            onRowRemove={(id) => setHeaders(tabId, removeRow(headers, id))}
          />
        </Tabs.Panel>
      </Tabs.Root>
    </>
  )
}
