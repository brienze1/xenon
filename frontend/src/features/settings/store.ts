import { create } from 'zustand'
import { toAppError, type AppError } from '@/platform/errors'
import { defaultAppSettings, type AppSettings } from '@/platform/mappers/settings'
import { getSettings, saveSettings } from '@/platform/services/settings'

interface SettingsState {
  open: boolean
  settings: AppSettings
  loading: boolean
  error: AppError | null
  openSettings: () => Promise<void>
  close: () => void
  save: (settings: AppSettings) => Promise<AppError | null>
}

export const useSettingsStore = create<SettingsState>((set) => ({
  open: false,
  settings: defaultAppSettings(),
  loading: false,
  error: null,
  openSettings: async () => {
    set({ open: true, loading: true, error: null })
    try {
      set({ settings: await getSettings(), loading: false })
    } catch (error) {
      set({ error: toAppError(error), loading: false })
    }
  },
  close: () => set({ open: false, error: null }),
  save: async (settings) => {
    try {
      set({ settings: await saveSettings(settings), open: false, error: null })
      return null
    } catch (error) {
      return toAppError(error)
    }
  },
}))
