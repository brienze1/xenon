import type { AppError } from '@/platform/errors'
import type { EnvironmentColor } from '@/platform/mappers/environment'
import type { VariableRow } from '@/platform/mappers/variable'

export type EnvironmentsStatus = 'idle' | 'loading' | 'ready'

/** The environment editors reachable from the switcher menu. */
export type EnvironmentEditor = 'manager' | 'globals'

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
