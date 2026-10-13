import { CancelError } from '@wailsio/runtime'

export const CANCELLED_CODE = 'cancelled'
export const UNKNOWN_CODE = 'UNK-00500'
export const SAVE_CONFLICT_CODE = 'REQ-02409'

export interface AppErrorDetail {
  attribute: string
  messages: string[]
}

interface ErrorPayload {
  error?: {
    code?: unknown
    description?: unknown
    error_details?: unknown
  }
}

export class AppError extends Error {
  readonly code: string
  readonly description: string
  readonly details: AppErrorDetail[]

  constructor(code: string, description: string, details: AppErrorDetail[] = []) {
    super(description)
    this.name = 'AppError'
    this.code = code
    this.description = description
    this.details = details
  }

  get cancelled(): boolean {
    return this.code === CANCELLED_CODE
  }
}

export function toAppError(error: unknown): AppError {
  if (error instanceof AppError) {
    return error
  }
  if (error instanceof CancelError) {
    return new AppError(CANCELLED_CODE, 'You cancelled the request')
  }
  const payload = readPayload(error)
  if (payload?.error && typeof payload.error.code === 'string') {
    return new AppError(
      payload.error.code,
      typeof payload.error.description === 'string' ? payload.error.description : '',
      readDetails(payload.error.error_details),
    )
  }
  return new AppError(UNKNOWN_CODE, error instanceof Error ? error.message : String(error))
}

function readPayload(error: unknown): ErrorPayload | null {
  if (typeof error !== 'object' || error === null || !('cause' in error)) {
    return null
  }
  const cause = (error as { cause: unknown }).cause
  if (typeof cause === 'string') {
    try {
      return JSON.parse(cause) as ErrorPayload
    } catch {
      return null
    }
  }
  return typeof cause === 'object' && cause !== null ? (cause as ErrorPayload) : null
}

function readDetails(details: unknown): AppErrorDetail[] {
  if (!Array.isArray(details)) {
    return []
  }
  return details.map((detail: { attribute?: unknown; messages?: unknown }) => ({
    attribute: typeof detail.attribute === 'string' ? detail.attribute : '',
    messages: Array.isArray(detail.messages) ? detail.messages.map(String) : [],
  }))
}
