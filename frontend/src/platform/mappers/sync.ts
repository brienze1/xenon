import type {
  SyncChangeItem as SyncChangeItemDto,
  SyncConflictItem as SyncConflictItemDto,
  SyncPullResult as SyncPullResultDto,
  SyncStatus as SyncStatusDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

export type SyncState = 'up_to_date' | 'local_changes' | 'ahead' | 'behind' | 'diverged' | 'conflict' | 'error' | 'syncing' | 'not_linked'
export type ChangeKind = 'added' | 'modified' | 'deleted' | 'renamed'
export type LinkMode = 'adopt_repo' | 'replace_with_mine'

export interface SyncError {
  code: string
  message: string
  nextStep: string
}

export interface SyncLinkSettings {
  repository: string
  branch: string
  subfolder: string
  accountId: string
  clonePath: string
  autoPull: boolean
  autoPush: boolean
}

export interface SyncStatus {
  state: SyncState
  localChanges: number
  ahead: number
  behind: number
  lastSyncAt: string | null
  lastError: SyncError | null
  notice: string
  settings: SyncLinkSettings
}

export interface SyncChange {
  itemId: string
  itemName: string
  itemKind: string
  change: ChangeKind
  paths: string[]
}

export interface SyncConflict {
  itemId: string
  itemName: string
  itemKind: string
  kind: string
  path: string
}

export interface PullResult {
  status: SyncStatus
  conflicts: SyncConflict[]
}

export interface LinkInput {
  localId: string
  repository: string
  branch: string
  subfolder: string
  accountId: string
  clonePath: string
  mode: LinkMode | null
}

/** Accepts the binding DTO and the git:status-changed payload, which share one shape. */
export function toSyncStatus(dto: SyncStatusDto): SyncStatus {
  return {
    state: dto.state as SyncState,
    localChanges: dto.local_changes,
    ahead: dto.ahead,
    behind: dto.behind,
    lastSyncAt: dto.last_sync_at ?? null,
    lastError: dto.last_error ? { code: dto.last_error.code, message: dto.last_error.message, nextStep: dto.last_error.next_step } : null,
    notice: dto.notice ?? '',
    settings: {
      repository: dto.settings.repository,
      branch: dto.settings.branch,
      subfolder: dto.settings.subfolder,
      accountId: dto.settings.account_id,
      clonePath: dto.settings.clone_path,
      autoPull: dto.settings.auto_pull,
      autoPush: dto.settings.auto_push,
    },
  }
}

export function toSyncChange(dto: SyncChangeItemDto): SyncChange {
  return { itemId: dto.item_id, itemName: dto.item_name, itemKind: dto.item_kind, change: dto.change as ChangeKind, paths: dto.paths ?? [] }
}

function toSyncConflict(dto: SyncConflictItemDto): SyncConflict {
  return { itemId: dto.item_id, itemName: dto.item_name, itemKind: dto.item_kind, kind: dto.kind, path: dto.path }
}

export function toPullResult(dto: SyncPullResultDto): PullResult {
  return { status: toSyncStatus(dto.status), conflicts: (dto.session?.conflicts ?? []).map(toSyncConflict) }
}
