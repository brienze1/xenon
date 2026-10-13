import * as SyncService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/syncservice'
import { AppError, toAppError } from '../errors'
import {
  toPullResult,
  toSyncChange,
  toSyncConflict,
  toSyncSession,
  toSyncStatus,
  type ChosenResolution,
  type CloneInput,
  type LinkInput,
  type PullResult,
  type SyncChange,
  type SyncConflict,
  type SyncSession,
  type SyncStatus,
} from '../mappers/sync'
import { toWorkspaceSummary, type WorkspaceSummary } from '../mappers/workspace'

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

export function cloneAsWorkspace(input: CloneInput): Promise<WorkspaceSummary> {
  return call(async () =>
    toWorkspaceSummary(
      required(
        await SyncService.CloneAsWorkspace({
          repository: input.repository,
          branch: input.branch,
          subfolder: input.subfolder,
          account_id: input.accountId,
          clone_path: input.clonePath,
          name: input.name,
        }),
      ),
    ),
  )
}

export function unlinkRepository(localId: string, targetFolder: string): Promise<WorkspaceSummary> {
  return call(async () => toWorkspaceSummary(required(await SyncService.Unlink({ local_id: localId, target_folder: targetFolder }))))
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

/** The pull waiting for conflict resolution, or null when there is none. */
export function getSyncSession(localId: string): Promise<SyncSession | null> {
  return call(async () => {
    const session = await SyncService.Session({ local_id: localId })
    return session ? toSyncSession(session) : null
  })
}

export function resolveConflict(sessionId: string, itemId: string, resolution: ChosenResolution, fieldPath: string): Promise<SyncConflict> {
  return call(async () => toSyncConflict(required(await SyncService.Resolve({ session_id: sessionId, item_id: itemId, resolution, field_path: fieldPath }))))
}

export function finishPull(sessionId: string): Promise<SyncStatus> {
  return call(async () => toSyncStatus(required(await SyncService.FinishPull({ session_id: sessionId }))))
}

export function cancelPull(sessionId: string): Promise<SyncStatus> {
  return call(async () => toSyncStatus(required(await SyncService.CancelPull({ session_id: sessionId }))))
}
