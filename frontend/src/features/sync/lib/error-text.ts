import type { AppError } from '@/platform/errors'
import type { SyncError } from '@/platform/mappers/sync'
import { withNextStep } from '@/shared/lib/next-step'

const NEXT_STEP = 'next_step'
const FALLBACK_STEP = 'Try again.'

/** A sync failure as one sentence plus the specific next step Go attached (spec FR-019). */
export function errorText(error: AppError): string {
  const nextStep = error.details.find((detail) => detail.attribute === NEXT_STEP)?.messages[0]
  return withNextStep(error.description || error.code, nextStep ?? FALLBACK_STEP)
}

export function lastErrorText(error: SyncError): string {
  return withNextStep(error.message || error.code, error.nextStep || FALLBACK_STEP)
}
