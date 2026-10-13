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

export async function pickFile(title: string): Promise<string | null> {
  try {
    return pickedPath(await DialogService.PickFile({ title }))
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
