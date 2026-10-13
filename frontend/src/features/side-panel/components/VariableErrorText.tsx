import type { AppError } from '@/platform/errors'
import { DUPLICATE_VARIABLE, INVALID_NEW_VARIABLE_KEY } from '@/platform/services/variables'
import { PASSPHRASE_SETUP_REQUIRED, VAULT_LOCKED } from '@/platform/services/vault'
import { withNextStep } from '@/shared/lib/next-step'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'

interface VariableErrorTextProps {
  error: AppError
  nextStep: string
}

function nextStepFor(error: AppError, fallback: string): string {
  switch (error.code) {
    case VAULT_LOCKED:
      return 'Unlock the vault, then try again.'
    case PASSPHRASE_SETUP_REQUIRED:
      return 'Choose a vault passphrase, then try again.'
    case DUPLICATE_VARIABLE:
      return 'Pick another name, then try again.'
    case INVALID_NEW_VARIABLE_KEY:
      return 'Remove any {, } or line breaks from the name, then try again.'
    default:
      return fallback
  }
}

/** A variable action's failure with what to do next: the code-specific step, or the caller's fallback. */
export function VariableErrorText({ error, nextStep }: VariableErrorTextProps) {
  return <NoTranslateText text={withNextStep(error.description || error.code, nextStepFor(error, nextStep))} terms={[error.code]} />
}
