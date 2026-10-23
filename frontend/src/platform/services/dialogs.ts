import * as DialogService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/dialogservice'
import { toAppError } from '../errors'

function pickedPath(picked: { path: string } | null): string | null {
  return picked?.path ? picked.path : null
}

export async function pickFolder(title: string): Promise<string | null> {
  try {
    return pickedPath(await DialogService.PickFolder({ title }))
  } catch (error) {
    throw toAppError(error)
  }
}

/** A file type choice of the native open dialog, e.g. `{ displayName: 'Data files', pattern: '*.csv;*.json' }`. */
export interface FileFilter {
  displayName: string
  pattern: string
}

export async function pickFile(title: string, filters: FileFilter[] = []): Promise<string | null> {
  try {
    return pickedPath(await DialogService.PickFile({ title, filters: filters.map((filter) => ({ display_name: filter.displayName, pattern: filter.pattern })) }))
  } catch (error) {
    throw toAppError(error)
  }
}

export async function pickSaveFile(title: string, suggestedName: string): Promise<string | null> {
  try {
    return pickedPath(await DialogService.PickSaveFile({ title, suggested_name: suggestedName }))
  } catch (error) {
    throw toAppError(error)
  }
}
