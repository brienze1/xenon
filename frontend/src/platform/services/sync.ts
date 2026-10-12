import * as SyncService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/syncservice'
import { AppError, toAppError } from '../errors'
import { toPullResult, toSyncChange, toSyncStatus, type LinkInput, type PullResult, type SyncChange, type SyncStatus } from '../mappers/sync'

export const SUBFOLDER_HAS_WORKSPACE = 'GIT-01409'
export const SYNC_BUSY = 'GIT-05409'

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

function required<T>(value: T | null): T {
  if (value === null) {
    throw new AppError('UNK-00500', 'Xenon returned no answer')
  }
  return value
}

export function getSyncStatus(localId: string): Promise<SyncStatus> {
  return call(async () => toSyncStatus(required(await SyncService.Status({ local_id: localId }))))
}

export function linkRepository(input: LinkInput): Promise<SyncStatus> {
  return call(async () =>
    toSyncStatus(
      required(
        await SyncService.Link({
          local_id: input.localId,
          repository: input.repository,
          branch: input.branch,
          subfolder: input.subfolder,
          account_id: input.accountId,
          clone_path: input.clonePath,
          mode: input.mode ?? '',
        }),
      ),
    ),
  )
}

export function listChanges(localId: string): Promise<SyncChange[]> {
  return call(async () => ((await SyncService.Changes({ local_id: localId })) ?? []).map(toSyncChange))
}

export function commitChanges(localId: string, message: string, itemIds: string[]): Promise<SyncStatus> {
  return call(async () => toSyncStatus(required(await SyncService.Commit({ local_id: localId, message, item_ids: itemIds }))))
}

export function pushWorkspace(localId: string): Promise<SyncStatus> {
  return call(async () => toSyncStatus(required(await SyncService.Push({ local_id: localId }))))
}

export function pullWorkspace(localId: string): Promise<PullResult> {
  return call(async () => toPullResult(required(await SyncService.Pull({ local_id: localId }))))
}

export function saveSyncSettings(localId: string, accountId: string, autoPull: boolean, autoPush: boolean): Promise<void> {
  return call(() => SyncService.SaveSettings({ local_id: localId, account_id: accountId, auto_pull: autoPull, auto_push: autoPush }))
}
