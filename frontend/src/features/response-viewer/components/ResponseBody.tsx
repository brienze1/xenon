import { useEffect, useMemo, useState, type KeyboardEvent } from 'react'
import type { SendResult } from '@/platform/mappers/request'
import { formatBytes } from '@/shared/lib/format-bytes'
import { rovingIndex } from '@/shared/lib/roving-focus'
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

const RADIO_KEYS: Record<string, string> = { ArrowDown: 'ArrowRight', ArrowUp: 'ArrowLeft' }

export function ResponseBody({ tabId, result }: ResponseBodyProps) {
  const [view, setView] = useState<BodyView>(() => initialBodyView(result.sizeBytes))
  const [copied, setCopied] = useState(false)
  const body = useResponseViewerStore((state) => state.entries[tabId]?.body ?? null)
  const bodyStatus = useResponseViewerStore((state) => state.entries[tabId]?.bodyStatus ?? 'idle')
  const loadBody = useResponseViewerStore((state) => state.loadBody)
  const saveBody = useResponseViewerStore((state) => state.saveBody)
  const language = languageFor(result.contentType)
  const isBinary = result.isBinary

  useEffect(() => {
    if (!isBinary && bodyStatus === 'idle') {
      void loadBody(tabId)
    }
  }, [tabId, isBinary, bodyStatus, loadBody])

  const shown = useMemo(() => (body !== null && view === 'pretty' ? prettyBody(body, language) : (body ?? '')), [body, view, language])

  function onViewKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const key = RADIO_KEYS[event.key] ?? event.key
    const current = VIEWS.findIndex((option) => option.view === view)
    const index = rovingIndex(VIEWS.length, current, key, { orientation: 'horizontal' })
    const next = index === null ? undefined : VIEWS[index]
    if (index === null || !next) {
      return
    }
    event.preventDefault()
    setView(next.view)
    event.currentTarget.querySelectorAll<HTMLElement>('[role="radio"]')[index]?.focus()
  }

  async function copy() {
    if (body !== null) {
      await navigator.clipboard.writeText(body)
      setCopied(true)
    }
  }

  return (
    <div className={styles.body}>
      <div className={styles.toolbar}>
        <div role="radiogroup" aria-label="Body view" className={styles.views} onKeyDown={onViewKeyDown}>
          {VIEWS.map((option) => (
            <button
              key={option.view}
              type="button"
              role="radio"
              aria-checked={view === option.view}
              tabIndex={view === option.view ? 0 : -1}
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
