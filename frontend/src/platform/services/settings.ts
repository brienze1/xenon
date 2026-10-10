import * as SettingsService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/settingsservice'
import { toAppError } from '../errors'
import { toAppSettings, toAppSettingsDto, type AppSettings } from '../mappers/settings'

export const SCRIPT_TIMEOUT_OUT_OF_RANGE = 'SCR-01400'

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

export function getSettings(): Promise<AppSettings> {
  return call(async () => toAppSettings(await SettingsService.Get()))
}

export function saveSettings(settings: AppSettings): Promise<AppSettings> {
  return call(async () => toAppSettings(await SettingsService.Save(toAppSettingsDto(settings))))
}
