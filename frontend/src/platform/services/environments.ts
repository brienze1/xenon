import * as EnvironmentService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/environmentservice'
import { AppError, toAppError } from '../errors'
import {
  toEnvironment,
  toEnvironmentList,
  type Environment,
  type EnvironmentList,
  type EnvironmentUpdate,
} from '../mappers/environment'

const EMPTY_RESPONSE = 'UNK-00204'
export const DUPLICATE_ENVIRONMENT_NAME = 'ENV-02409'
export const ENVIRONMENT_NOT_FOUND = 'ENV-01404'

function required<T>(value: T | null): T {
  if (value === null) {
    throw new AppError(EMPTY_RESPONSE, 'The backend returned no data')
  }
  return value
}

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

export function getEnvironment(localId: string, id: string): Promise<Environment> {
  return call(async () => toEnvironment(required(await EnvironmentService.Get({ local_id: localId, id }))))
}

export function createEnvironment(localId: string, name: string): Promise<Environment> {
  return call(async () => toEnvironment(required(await EnvironmentService.Create({ local_id: localId, name }))))
}

export function updateEnvironment(localId: string, update: EnvironmentUpdate): Promise<Environment> {
  return call(async () =>
    toEnvironment(
      required(
        await EnvironmentService.Update({
          local_id: localId,
          id: update.id,
          name: update.name,
          description: update.description,
          color: update.color,
          confirm_before_send: update.confirmBeforeSend,
        }),
      ),
    ),
  )
}

export function duplicateEnvironment(localId: string, id: string): Promise<Environment> {
  return call(async () => toEnvironment(required(await EnvironmentService.Duplicate({ local_id: localId, id }))))
}

/** Deletes the environment and returns the workspace's active environment afterwards. */
export function deleteEnvironment(localId: string, id: string): Promise<string | null> {
  return call(async () => {
    const active = await EnvironmentService.Delete({ local_id: localId, id })
    return active?.active_id ? active.active_id : null
  })
}
