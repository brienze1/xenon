import type { ExportFormat, ExportItemRef, ExportShape } from '@/platform/mappers/export'

export type ExportStep = 'closed' | 'loading' | 'configure' | 'exporting' | 'summary'

/** The item the dialog exports, with the name shown in its title. */
export interface ExportTarget extends ExportItemRef {
  name: string
}

/** What the user picked in the dialog; the passphrase only lives here while the dialog is open. */
export interface ExportChoices {
  format: ExportFormat
  shape: ExportShape
  includeSecrets: boolean
  passphrase: string
  passphraseConfirmation: string
  currentValuesAsInitial: boolean
}

export interface ExportProgressState {
  done: number
  total: number
}
