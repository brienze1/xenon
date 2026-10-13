import { lazy, Suspense, useEffect, useMemo, useState } from 'react'
import type { ExampleResponse } from '@/platform/mappers/example'
import { initialBodyView, isHtml, isImage, languageFor, prettyBody, type BodyView } from '@/shared/lib/body-format'
import { formatBytes } from '@/shared/lib/format-bytes'
import { useExamplesStore } from '@/features/examples/store'
import styles from './ExampleTab.module.css'

const CodeEditor = lazy(() => import('@/shared/ui/code-editor/CodeEditor').then((module) => ({ default: module.CodeEditor })))

const VIEWS: Array<{ view: BodyView; label: string }> = [
  { view: 'pretty', label: 'Pretty' },
  { view: 'raw', label: 'Raw' },
  { view: 'preview', label: 'Preview' },
]

const BODY_VIEW_LEGEND = <legend className="visually-hidden">Body view</legend>

const LOADING_BODY = (
  <p role="status" className={styles.note}>
    Loading body…
  </p>
)

const LOADING_EDITOR = (
  <p role="status" className={styles.note}>
    Loading editor…
  </p>
)

interface ExampleBodyProps {
  tabId: string
  response: ExampleResponse
}

interface PreviewProps {
  response: ExampleResponse
  body: string | null
}

function Preview({ response, body }: PreviewProps) {
  if (isImage(response.contentType) && response.bodyUrl) {
    return <img className={styles.image} src={response.bodyUrl} alt="Example response preview" />
  }
  if (isHtml(response.contentType) && body !== null) {
    return <iframe className={styles.frame} title="Example response preview" sandbox="" srcDoc={body} />
  }
  return (
    <p className={styles.note}>
      No preview available for {response.contentType ? <span translate="no">{response.contentType}</span> : 'this response'}.
    </p>
  )
}

export function ExampleBody({ tabId, response }: ExampleBodyProps) {
  const [view, setView] = useState<BodyView>(() => initialBodyView(response.sizeBytes))
  const routedBody = useExamplesStore((state) => state.tabs[tabId]?.routedBody ?? null)
  const bodyStatus = useExamplesStore((state) => state.tabs[tabId]?.bodyStatus ?? 'idle')
  const setBody = useExamplesStore((state) => state.setBody)
  const loadRoutedBody = useExamplesStore((state) => state.loadRoutedBody)
  const editable = response.body !== null
  const body = response.body ?? routedBody
  const language = languageFor(response.contentType)

  useEffect(() => {
    if (!editable && !response.isBinary && bodyStatus === 'idle') {
      void loadRoutedBody(tabId)
    }
  }, [tabId, editable, response.isBinary, bodyStatus, loadRoutedBody])

  const pretty = useMemo(() => (body !== null && view === 'pretty' ? prettyBody(body, language) : null), [body, view, language])

  return (
    <div className={styles.body}>
      <fieldset className={styles.views}>
        {BODY_VIEW_LEGEND}
        {VIEWS.map((option) => (
          <label key={option.view} className={styles.view}>
            <input type="radio" name={`${tabId}-example-body-view`} className={styles.viewInput} checked={view === option.view} onChange={() => setView(option.view)} />
            {option.label}
          </label>
        ))}
      </fieldset>
      {!editable && !response.isBinary ? (
        <p className={styles.note}>This body is {formatBytes(response.sizeBytes)}; it is shown read-only and kept as it is when you save.</p>
      ) : null}
      {view === 'preview' ? (
        <Preview response={response} body={body} />
      ) : response.isBinary ? (
        <p className={styles.note}>Binary body ({formatBytes(response.sizeBytes)}). Use Preview to see images.</p>
      ) : !editable && bodyStatus !== 'ready' ? (
        LOADING_BODY
      ) : (
        <Suspense fallback={LOADING_EDITOR}>
          {view === 'pretty' ? (
            <CodeEditor key="pretty" label="Pretty example body" className={styles.editor} language={language} value={pretty ?? ''} />
          ) : (
            <CodeEditor
              key={editable ? 'raw-edit' : 'raw'}
              label="Raw example body"
              className={styles.editor}
              language="text"
              value={body ?? ''}
              onChange={editable ? (value) => setBody(tabId, value) : undefined}
            />
          )}
        </Suspense>
      )}
    </div>
  )
}
