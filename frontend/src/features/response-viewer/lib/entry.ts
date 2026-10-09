import type { ResponseEntry } from '@/features/response-viewer/model'
import type { SendResult } from '@/platform/mappers/request'

export const EMPTY_ENTRY: ResponseEntry = { version: 0, sending: false, result: null, failure: null, body: null, bodyStatus: 'idle' }

export function withResult(previous: ResponseEntry | undefined, result: SendResult): ResponseEntry {
  const inline = result.inlineBody
  return {
    version: (previous?.version ?? 0) + 1,
    sending: false,
    result,
    failure: null,
    body: inline,
    bodyStatus: inline === null ? 'idle' : 'ready',
  }
}
