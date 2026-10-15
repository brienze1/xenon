import type { TabDraft, VersionState } from '@/features/request-builder/model'

/** The key of the Default among a tab's versions. */
export const DEFAULT_VERSION = ''

/** The key of a version among a tab's versions. */
export function versionKey(aliasId: string | null): string {
  return aliasId ?? DEFAULT_VERSION
}

/** Whether a version kept aside has unsaved edits, its name included. */
export function isVersionDirty(version: VersionState): boolean {
  return version.dirty || (version.savedName !== null && version.draft.name.trim() !== version.savedName)
}

/** Whether a version of the tab other than the one on screen has unsaved edits. */
export function hasDirtyHiddenVersion(tab: Pick<TabDraft, 'versions'>): boolean {
  return Object.values(tab.versions).some(isVersionDirty)
}
