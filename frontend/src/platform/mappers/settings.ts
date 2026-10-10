import type { AppSettings as AppSettingsDto } from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

export const DEFAULT_SCRIPT_TIMEOUT_MS = 10_000
export const MIN_SCRIPT_TIMEOUT_MS = 1_000
export const MAX_SCRIPT_TIMEOUT_MS = 60_000

/** App-wide settings, stored by Go in app.json. */
export interface AppSettings {
  scriptTimeoutMs: number
}

export function defaultAppSettings(): AppSettings {
  return { scriptTimeoutMs: DEFAULT_SCRIPT_TIMEOUT_MS }
}

export function toAppSettings(dto: AppSettingsDto | null): AppSettings {
  return dto ? { scriptTimeoutMs: dto.script_timeout_ms } : defaultAppSettings()
}

export function toAppSettingsDto(settings: AppSettings): AppSettingsDto {
  return { script_timeout_ms: settings.scriptTimeoutMs }
}
