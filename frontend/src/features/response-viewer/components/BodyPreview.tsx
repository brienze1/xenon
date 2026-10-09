import type { SendResult } from '@/platform/mappers/request'
import { isHtml, isImage } from '@/features/response-viewer/lib/body-view'
import styles from './BodyPreview.module.css'

interface BodyPreviewProps {
  result: SendResult
  body: string | null
}

export function BodyPreview({ result, body }: BodyPreviewProps) {
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
