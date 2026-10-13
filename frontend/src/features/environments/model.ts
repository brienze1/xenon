import type { AppError } from '@/platform/errors'
import type { EnvironmentColor } from '@/platform/mappers/environment'
import type { VariableRow } from '@/platform/mappers/variable'
import type { DraftRow, TypeSwitches } from './lib/variable-draft'

export type EnvironmentsStatus = 'idle' | 'loading' | 'ready'

export type PassphrasePrompt = 'setup' | 'unlock'

export interface TableState {
  rows: VariableRow[]
  warnings: string[]
  status: EnvironmentsStatus
  error: AppError | null
}

/** How an environment warns before sending: its neon colour and "Confirm before send". */
export interface EnvironmentLook {
  color: EnvironmentColor
  confirmBeforeSend: boolean
}

/** The unsaved edits of one variable table. `base` is the stored rows the edits started from. */
export interface VariableDraft {
  base: VariableRow[]
  rows: DraftRow[]
  dirty: boolean
  /** The type switches waiting for confirmation before a save. */
  confirming: TypeSwitches | null
  failure: AppError | null
  saving: boolean
}
