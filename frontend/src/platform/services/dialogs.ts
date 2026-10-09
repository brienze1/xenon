import * as DialogService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/dialogservice'
import { toAppError } from '../errors'

export async function pickFolder(title: string): Promise<string | null> {
  try {
    const picked = await DialogService.PickFolder({ title })
    return picked?.path ? picked.path : null
  } catch (error) {
    throw toAppError(error)
  }
}
