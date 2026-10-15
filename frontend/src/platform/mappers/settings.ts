import type { AppSettings as AppSettingsDto } from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

export const DEFAULT_SCRIPT_TIMEOUT_MS = 10_000
export const MIN_SCRIPT_TIMEOUT_MS = 1_000
export const MAX_SCRIPT_TIMEOUT_MS = 60_000

export const DEFAULT_HISTORY_CAP = 1_000
export const MIN_HISTORY_CAP = 100
export const MAX_HISTORY_CAP = 20_000

export const DEFAULT_AI_PORT = 47_821
export const MIN_AI_PORT = 1_024
export const MAX_AI_PORT = 65_535

/** AI access (MCP): the local server switch, its port and the approval bypass. */
export interface AIAccessSettings {
  enabled: boolean
  port: number
  skipApprovals: boolean
}

/** App-wide settings, stored by Go in app.json. */
export interface AppSettings {
  scriptTimeoutMs: number
  historyUserCap: number
  historyAiCap: number
  aiAccess: AIAccessSettings
  /** Download a newer release in the background after the startup check (on by default). */
  autoUpdate: boolean
}

export function defaultAppSettings(): AppSettings {
  return {
    scriptTimeoutMs: DEFAULT_SCRIPT_TIMEOUT_MS,
    historyUserCap: DEFAULT_HISTORY_CAP,
    historyAiCap: DEFAULT_HISTORY_CAP,
    aiAccess: { enabled: true, port: DEFAULT_AI_PORT, skipApprovals: false },
    autoUpdate: true,
  }
}

export function toAppSettings(dto: AppSettingsDto | null): AppSettings {
  const defaults = defaultAppSettings()
  if (!dto) {
    return defaults
  }
  return {
    scriptTimeoutMs: dto.script_timeout_ms ?? defaults.scriptTimeoutMs,
    historyUserCap: dto.history_user_cap ?? defaults.historyUserCap,
    historyAiCap: dto.history_ai_cap ?? defaults.historyAiCap,
    aiAccess: {
      enabled: dto.ai_access?.enabled ?? defaults.aiAccess.enabled,
      port: dto.ai_access?.port ?? defaults.aiAccess.port,
      skipApprovals: dto.ai_access?.skip_approvals ?? defaults.aiAccess.skipApprovals,
    },
    autoUpdate: dto.auto_update ?? defaults.autoUpdate,
  }
}

export function toAppSettingsDto(settings: AppSettings): AppSettingsDto {
  return {
    script_timeout_ms: settings.scriptTimeoutMs,
    history_user_cap: settings.historyUserCap,
    history_ai_cap: settings.historyAiCap,
    ai_access: {
      enabled: settings.aiAccess.enabled,
      port: settings.aiAccess.port,
      skip_approvals: settings.aiAccess.skipApprovals,
    },
    auto_update: settings.autoUpdate,
  }
}
