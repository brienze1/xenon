import type { AppError } from '@/platform/errors'
import type { Example, ExampleResponse } from '@/platform/mappers/example'

export type ExampleStatus = 'loading' | 'ready' | 'error'
export type BodyStatus = 'idle' | 'loading' | 'ready' | 'error'

/** One open example tab: the example as stored, the response being edited and save state. */
export interface ExampleTabState {
  requestId: string
  exampleId: string
  example: Example | null
  response: ExampleResponse | null
  /** Response headers as `Key: value` lines while bulk editing; `null` in table mode. */
  bulkHeaders: string | null
  /** The body served by the response route, shown read-only. */
  routedBody: string | null
  bodyStatus: BodyStatus
  status: ExampleStatus
  dirty: boolean
  saving: boolean
  conflict: boolean
  error: AppError | null
}

export interface SaveExampleDialogState {
  tabId: string
  requestId: string
}
