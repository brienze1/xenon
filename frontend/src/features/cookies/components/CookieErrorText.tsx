import type { AppError } from '@/platform/errors'
import { withNextStep } from '@/shared/lib/next-step'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'

interface CookieErrorTextProps {
  error: AppError
  nextStep: string
}

export function CookieErrorText({ error, nextStep }: CookieErrorTextProps) {
  return <NoTranslateText text={withNextStep(error.description || error.code, nextStep)} terms={[error.code]} />
}
