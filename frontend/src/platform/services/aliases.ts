import * as AliasService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/aliasservice'
import { AppError, toAppError } from '../errors'
import { toRequestAliasDto, toSavedRequest, type RequestDraft, type SavedRequest } from '../mappers/request'

const EMPTY_RESPONSE = 'UNK-00204'

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

/** Creates an alias of the saved request as a copy of its saved Default, limited to `environmentIds` (none: all environments). */
export function createAlias(localId: string, requestId: string, name: string, environmentIds: string[], baseRevision: string): Promise<SavedRequest> {
  return call(async () =>
    toSavedRequest(
      required(await AliasService.Create({ local_id: localId, request_id: requestId, base_revision: baseRevision, name, environment_ids: environmentIds })),
    ),
  )
}

/** Limits an alias to `environmentIds` (none: all environments) without touching its content. */
export function setAliasEnvironments(localId: string, requestId: string, aliasId: string, environmentIds: string[], baseRevision: string): Promise<SavedRequest> {
  return call(async () =>
    toSavedRequest(
      required(
        await AliasService.SetEnvironments({
          local_id: localId,
          request_id: requestId,
          alias_id: aliasId,
          environment_ids: environmentIds,
          base_revision: baseRevision,
        }),
      ),
    ),
  )
}

/** Saves one alias (its name is the draft's name); the Default and the other aliases stay as they are on disk. */
export function saveAlias(localId: string, aliasId: string, draft: RequestDraft, baseRevision: string, force: boolean): Promise<SavedRequest> {
  return call(async () =>
    toSavedRequest(
      required(
        await AliasService.Save({ local_id: localId, request_id: draft.id, base_revision: baseRevision, force, alias: toRequestAliasDto(aliasId, draft) }),
      ),
    ),
  )
}

export function deleteAlias(localId: string, requestId: string, aliasId: string, baseRevision: string): Promise<SavedRequest> {
  return call(async () =>
    toSavedRequest(required(await AliasService.Delete({ local_id: localId, request_id: requestId, alias_id: aliasId, base_revision: baseRevision }))),
  )
}
