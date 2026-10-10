import * as EnvironmentService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/environmentservice'
import { toAppError } from '../errors'
import { toEnvironmentList, type EnvironmentList } from '../mappers/environment'

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

export function listEnvironments(localId: string): Promise<EnvironmentList> {
  return call(async () => toEnvironmentList(await EnvironmentService.List({ local_id: localId })))
}

export function setActiveEnvironment(localId: string, id: string | null): Promise<string | null> {
  return call(async () => {
    const active = await EnvironmentService.SetActive({ local_id: localId, id: id ?? '' })
    return active?.active_id ? active.active_id : null
  })
}
