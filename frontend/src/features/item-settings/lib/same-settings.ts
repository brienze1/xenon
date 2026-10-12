import { toSaveItemSettingsDto, type ItemSettings } from '@/platform/mappers/tree'

/**
 * Whether two versions of a collection's or folder's settings say the same thing. The name is
 * left out (a rename never conflicts with edits) and so are the ids the table gives its rows.
 */
export function sameSettings(a: ItemSettings, b: ItemSettings): boolean {
  return JSON.stringify(toSaveItemSettingsDto('', a)) === JSON.stringify(toSaveItemSettingsDto('', b))
}
