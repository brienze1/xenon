import { useEffect, useId, useMemo, useState, type ReactNode } from 'react'
import type { SendResult } from '@/platform/mappers/request'
import { formatBytes } from '@/shared/lib/format-bytes'
import { Button } from '@/shared/ui/button/Button'
import { CodeEditor } from '@/shared/ui/code-editor/CodeEditor'
import { initialBodyView, isHtml, isImage, languageFor, prettyBody, type BodyView } from '@/features/response-viewer/lib/body-view'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import styles from './ResponseBody.module.css'

interface ResponseBodyProps {
  tabId: string
  result: SendResult
}

const VIEWS: Array<{ view: BodyView; label: string }> = [
  { view: 'pretty', label: 'Pretty' },
  { view: 'raw', label: 'Raw' },
  { view: 'preview', label: 'Preview' },
]

type CopyState = 'idle' | 'copied' | 'failed'

const COPY_LABELS: Record<CopyState, string> = { idle: 'Copy Body', copied: 'Copied', failed: 'Copy Failed' }

const COPY_ANNOUNCEMENTS: Record<CopyState, ReactNode> = {
  idle: null,
  copied: (
    <>
      <span translate="no">Xenon</span> copied the response body to the clipboard.
    </>
  ),
  failed: (
    <>
      <span translate="no">Xenon</span> couldn’t copy to the clipboard. Select the body text and copy it manually.
    </>
  ),
}

const LOADING_BODY = <p className={styles.note}>Loading body…</p>

export function ResponseBody({ tabId, result }: ResponseBodyProps) {
  const [view, setView] = useState<BodyView>(() => initialBodyView(result.sizeBytes))
  const [copy, setCopy] = useState<CopyState>('idle')
  const body = useResponseViewerStore((state) => state.entries[tabId]?.body ?? null)
  const bodyStatus = useResponseViewerStore((state) => state.entries[tabId]?.bodyStatus ?? 'idle')
  const loadBody = useResponseViewerStore((state) => state.loadBody)
  const saveBody = useResponseViewerStore((state) => state.saveBody)
  const language = languageFor(result.contentType)
  const isBinary = result.isBinary
  const viewName = useId()

  useEffect(() => {
    if (!isBinary && bodyStatus === 'idle') {
      void loadBody(tabId)
    }
  }, [tabId, isBinary, bodyStatus, loadBody])

  const shown = useMemo(() => (body !== null && view === 'pretty' ? prettyBody(body, language) : (body ?? '')), [body, view, language])

  async function copyBody() {
    if (body === null) {
      return
    }
    try {
      await navigator.clipboard.writeText(body)
      setCopy('copied')
    } catch {
      setCopy('failed')
    }
  }

  return (
    <div className={styles.body}>
      <div className={styles.toolbar}>
        <fieldset className={styles.views}>
          <legend className="visually-hidden">Body view</legend>
          {VIEWS.map((option) => (
            <label key={option.view} className={styles.view}>
              <input
                type="radio"
                name={viewName}
                className={styles.viewInput}
                checked={view === option.view}
                onChange={() => setView(option.view)}
              />
              {option.label}
            </label>
          ))}
        </fieldset>
        <div className={styles.actions}>
          <Button size="sm" variant="ghost" disabled={body === null} onClick={() => void copyBody()}>
            {COPY_LABELS[copy]}
          </Button>
          <span role="status" aria-live="polite" className="visually-hidden">
            {COPY_ANNOUNCEMENTS[copy]}
          </span>
          <Button size="sm" variant="ghost" onClick={() => void saveBody(tabId)}>
            Save to File
          </Button>
        </div>
      </div>
      {view === 'preview' ? (
        <BodyPreview result={result} body={body} />
      ) : result.isBinary ? (
        <p className={styles.note}>Binary body ({formatBytes(result.sizeBytes)}). Use Preview or Save to File.</p>
      ) : bodyStatus === 'loading' ? (
        LOADING_BODY
      ) : (
        <CodeEditor
          key={view}
          label={view === 'pretty' ? 'Pretty response body' : 'Raw response body'}
          className={styles.editor}
          language={view === 'pretty' ? language : 'text'}
          value={shown}
        />
      )}
    </div>
  )
}

interface BodyPreviewProps {
  result: SendResult
  body: string | null
}

function BodyPreview({ result, body }: BodyPreviewProps) {
  if (isImage(result.contentType)) {
    return <img className={styles.image} src={result.bodyUrl} alt="Response preview" />
  }
  if (isHtml(result.contentType) && body !== null) {
    return <iframe className={styles.frame} title="Response preview" sandbox="" srcDoc={body} />
  }
  return (
    <p className={styles.note}>
      No preview available for {result.contentType ? <span translate="no">{result.contentType}</span> : 'this response'}.
    </p>
  )
}
