import type { SendResult } from '@/platform/mappers/request'
import { ResponseBody } from '@/features/response-viewer/components/ResponseBody'
import { ResponseHeaders } from '@/features/response-viewer/components/ResponseHeaders'
import { StatusLine } from '@/features/response-viewer/components/StatusLine'
import styles from './RunRowResponse.module.css'

interface RunRowResponseProps {
  /** `run:<runId>:<index>`: the response viewer entry holding this persisted response. */
  responseKey: string
  result: SendResult
}

/** A persisted run response, rendered as the response viewer renders a send (large and binary handling included). */
export function RunRowResponse({ responseKey, result }: RunRowResponseProps) {
  return (
    <div className={styles.response}>
      <StatusLine result={result} />
      <ResponseHeaders headers={result.headers} />
      <div className={styles.body}>
        <ResponseBody tabId={responseKey} result={result} />
      </div>
    </div>
  )
}
