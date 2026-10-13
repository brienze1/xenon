import type { AppError } from '@/platform/errors'
import type { SendResult } from '@/platform/mappers/request'

export type BodyStatus = 'idle' | 'loading' | 'ready' | 'error'

export interface ResponseEntry {
  version: number
  sending: boolean
  result: SendResult | null
  failure: AppError | null
  body: string | null
  bodyStatus: BodyStatus
}
