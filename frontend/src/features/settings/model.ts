/** A shortcut as the Settings list shows it; the app passes its hotkey table. */
export interface ShortcutEntry {
  label: string
  combos: readonly string[]
}

/** The Settings tab's form, as typed: numbers stay text until Save validates them. */
export interface SettingsDraft {
  scriptTimeoutSeconds: string
  historyUserCap: string
  historyAiCap: string
  aiEnabled: boolean
  aiPort: string
  skipApprovals: boolean
  autoUpdate: boolean
}

/** The form fields Save validates, each with its own inline error. */
export type SettingsField = 'scriptTimeout' | 'historyUserCap' | 'historyAiCap' | 'aiPort'

export type SettingsFieldErrors = Partial<Record<SettingsField, string>>
