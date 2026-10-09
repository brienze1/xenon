import { useEffect, useMemo, useState } from 'react'
import type { SendResult } from '@/platform/mappers/request'
import { formatBytes } from '@/shared/lib/format-bytes'
import { Button } from '@/shared/ui/button/Button'
import { CodeEditor } from '@/shared/ui/code-editor/CodeEditor'
import { initialBodyView, isHtml, isImage, languageFor, prettyBody, type BodyView } from '../lib/body-view'
import { useResponseViewerStore } from '../store'
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

export function ResponseBody({ tabId, result }: ResponseBodyProps) {
  const [view, setView] = useState<BodyView>(() => initialBodyView(result.sizeBytes))
  const [copied, setCopied] = useState(false)
  const body = useResponseViewerStore((state) => state.entries[tabId]?.body ?? null)
  const bodyStatus = useResponseViewerStore((state) => state.entries[tabId]?.bodyStatus ?? 'idle')
  const loadBody = useResponseViewerStore((state) => state.loadBody)
  const saveBody = useResponseViewerStore((state) => state.saveBody)
  const language = languageFor(result.contentType)

  useEffect(() => {
    if (!result.isBinary && bodyStatus === 'idle') {
      void loadBody(tabId)
    }
  }, [tabId, result, bodyStatus, loadBody])

  const shown = useMemo(() => (body !== null && view === 'pretty' ? prettyBody(body, language) : (body ?? '')), [body, view, language])

  async function copy() {
    if (body !== null) {
      await navigator.clipboard.writeText(body)
      setCopied(true)
    }
  }

  return (
    <div className={styles.body}>
      <div className={styles.toolbar}>
        <div role="radiogroup" aria-label="Body view" className={styles.views}>
          {VIEWS.map((option) => (
            <button
              key={option.view}
              type="button"
              role="radio"
              aria-checked={view === option.view}
              className={styles.view}
              onClick={() => setView(option.view)}
            >
              {option.label}
            </button>
          ))}
        </div>
        <div className={styles.actions}>
          <Button size="sm" variant="ghost" disabled={body === null} onClick={() => void copy()}>
            {copied ? 'Copied' : 'Copy'}
          </Button>
          <Button size="sm" variant="ghost" onClick={() => void saveBody(tabId)}>
            Save to file
          </Button>
        </div>
      </div>
      {view === 'preview' ? (
        <BodyPreview result={result} body={body} />
      ) : result.isBinary ? (
        <p className={styles.note}>Binary body ({formatBytes(result.sizeBytes)}). Use Preview or Save to file.</p>
      ) : bodyStatus === 'loading' ? (
        <p className={styles.note}>Loading body…</p>
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
  return <p className={styles.note}>No preview available for {result.contentType || 'this response'}.</p>
}
