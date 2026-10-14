import {
  MAX_AI_PORT,
  MAX_HISTORY_CAP,
  MAX_SCRIPT_TIMEOUT_MS,
  MIN_AI_PORT,
  MIN_HISTORY_CAP,
  MIN_SCRIPT_TIMEOUT_MS,
  type AppSettings,
} from '@/platform/mappers/settings'
import { formatNumber } from '@/shared/lib/number-format'
import type { SettingsDraft, SettingsField, SettingsFieldErrors } from '@/features/settings/model'

const MS_PER_SECOND = 1000
export const MIN_TIMEOUT_SECONDS = MIN_SCRIPT_TIMEOUT_MS / MS_PER_SECOND
export const MAX_TIMEOUT_SECONDS = MAX_SCRIPT_TIMEOUT_MS / MS_PER_SECOND

export const FIELD_MESSAGES: Record<SettingsField, string> = {
  scriptTimeout: `Enter a whole number of seconds from ${MIN_TIMEOUT_SECONDS} to ${MAX_TIMEOUT_SECONDS}.`,
  historyUserCap: `Enter a whole number from ${formatNumber(MIN_HISTORY_CAP, 0)} to ${formatNumber(MAX_HISTORY_CAP, 0)}.`,
  historyAiCap: `Enter a whole number from ${formatNumber(MIN_HISTORY_CAP, 0)} to ${formatNumber(MAX_HISTORY_CAP, 0)}.`,
  aiPort: `Enter a port from ${MIN_AI_PORT} to ${MAX_AI_PORT}.`,
}

/** The fields in the order they appear, so the first invalid one gets focus. */
export const FIELD_ORDER: readonly SettingsField[] = ['scriptTimeout', 'historyUserCap', 'historyAiCap', 'aiPort']

export function toDraft(settings: AppSettings): SettingsDraft {
  return {
    scriptTimeoutSeconds: String(settings.scriptTimeoutMs / MS_PER_SECOND),
    historyUserCap: String(settings.historyUserCap),
    historyAiCap: String(settings.historyAiCap),
    aiEnabled: settings.aiAccess.enabled,
    aiPort: String(settings.aiAccess.port),
    skipApprovals: settings.aiAccess.skipApprovals,
  }
}

export function sameDraft(a: SettingsDraft, b: SettingsDraft): boolean {
  return (
    a.scriptTimeoutSeconds.trim() === b.scriptTimeoutSeconds.trim() &&
    a.historyUserCap.trim() === b.historyUserCap.trim() &&
    a.historyAiCap.trim() === b.historyAiCap.trim() &&
    a.aiEnabled === b.aiEnabled &&
    a.aiPort.trim() === b.aiPort.trim() &&
    a.skipApprovals === b.skipApprovals
  )
}

function wholeNumberIn(text: string, min: number, max: number): number | null {
  const value = Number(text)
  return text.trim() !== '' && Number.isInteger(value) && value >= min && value <= max ? value : null
}

export type DraftCheck = { ok: true; settings: AppSettings } | { ok: false; errors: SettingsFieldErrors }

/** Validates every field at once: the settings to save, or an inline error per invalid field. */
export function checkDraft(draft: SettingsDraft): DraftCheck {
  const values: Record<SettingsField, number | null> = {
    scriptTimeout: wholeNumberIn(draft.scriptTimeoutSeconds, MIN_TIMEOUT_SECONDS, MAX_TIMEOUT_SECONDS),
    historyUserCap: wholeNumberIn(draft.historyUserCap, MIN_HISTORY_CAP, MAX_HISTORY_CAP),
    historyAiCap: wholeNumberIn(draft.historyAiCap, MIN_HISTORY_CAP, MAX_HISTORY_CAP),
    aiPort: wholeNumberIn(draft.aiPort, MIN_AI_PORT, MAX_AI_PORT),
  }
  const invalid = FIELD_ORDER.filter((field) => values[field] === null)
  if (invalid.length > 0) {
    return { ok: false, errors: Object.fromEntries(invalid.map((field) => [field, FIELD_MESSAGES[field]])) }
  }
  const { scriptTimeout, historyUserCap, historyAiCap, aiPort } = values as Record<SettingsField, number>
  return {
    ok: true,
    settings: {
      scriptTimeoutMs: scriptTimeout * MS_PER_SECOND,
      historyUserCap,
      historyAiCap,
      aiAccess: { enabled: draft.aiEnabled, port: aiPort, skipApprovals: draft.skipApprovals },
    },
  }
}
