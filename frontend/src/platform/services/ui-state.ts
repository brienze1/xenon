import * as UIStateService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/uistateservice'
import { toAppError } from '../errors'
import { toUIState, toUIStateDto, type UIState } from '../mappers/ui-state'

export async function getUIState(localId: string): Promise<UIState> {
  try {
    return toUIState(await UIStateService.Get({ local_id: localId }))
  } catch (error) {
    throw toAppError(error)
  }
}

export async function saveUIState(localId: string, state: UIState): Promise<void> {
  try {
    await UIStateService.Save(toUIStateDto(localId, state))
  } catch (error) {
    throw toAppError(error)
  }
}
