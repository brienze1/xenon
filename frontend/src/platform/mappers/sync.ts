import type {
  SyncChangeItem as SyncChangeItemDto,
  SyncConflictItem as SyncConflictItemDto,
  SyncEnvironmentVersion as SyncEnvironmentVersionDto,
  SyncFieldConflict as SyncFieldConflictDto,
  SyncItemVersion as SyncItemVersionDto,
  SyncPullResult as SyncPullResultDto,
  SyncSession as SyncSessionDto,
  SyncStatus as SyncStatusDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import { toRequestDraft, type RequestDraft } from './request'

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

export type ConflictKind = 'both_modified' | 'delete_modify' | 'field'
export type ConflictResolution = 'unresolved' | 'mine' | 'theirs' | 'both' | 'keep' | 'delete' | 'merged'
export type ChosenResolution = Exclude<ConflictResolution, 'unresolved' | 'merged'>

export interface SyncVariableVersion {
  key: string
  value: string
  description: string
  secret: boolean
  enabled: boolean
}

export interface SyncEnvironmentVersion {
  name: string
  color: string
  variables: SyncVariableVersion[]
}

/** One side of a conflict, rendered as the Xenon item when it parses, always with its file text. */
export interface SyncItemVersion {
  text: string
  request: RequestDraft | null
  environment: SyncEnvironmentVersion | null
}

/** A field changed differently on both sides; null values mean the side removed the field. */
export interface SyncFieldConflict {
  path: string
  mine: string | null
  theirs: string | null
  resolution: ConflictResolution
}

export interface SyncConflict {
  itemId: string
  itemName: string
  itemKind: string
  kind: ConflictKind
  resolution: ConflictResolution
  path: string
  uncommitted: boolean
  base: SyncItemVersion | null
  mine: SyncItemVersion | null
  theirs: SyncItemVersion | null
  fields: SyncFieldConflict[]
}

export interface SyncSession {
  id: string
  localId: string
  conflicts: SyncConflict[]
}

export interface PullResult {
  status: SyncStatus
  session: SyncSession | null
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

/** An empty branch means the remote's default branch; an empty name, the repository's name. */
export interface CloneInput {
  repository: string
  branch: string
  subfolder: string
  accountId: string
  clonePath: string
  name: string
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

function toEnvironmentVersion(dto: SyncEnvironmentVersionDto): SyncEnvironmentVersion {
  return {
    name: dto.name,
    color: dto.color,
    variables: (dto.variables ?? []).map((row) => ({ key: row.key, value: row.value, description: row.description, secret: row.secret, enabled: row.enabled })),
  }
}

function toItemVersion(dto: SyncItemVersionDto | null | undefined): SyncItemVersion | null {
  if (!dto) {
    return null
  }
  return {
    text: dto.text,
    request: dto.request ? toRequestDraft(dto.request) : null,
    environment: dto.environment ? toEnvironmentVersion(dto.environment) : null,
  }
}

function toFieldConflict(dto: SyncFieldConflictDto): SyncFieldConflict {
  return { path: dto.path, mine: dto.mine, theirs: dto.theirs, resolution: dto.resolution as ConflictResolution }
}

export function toSyncConflict(dto: SyncConflictItemDto): SyncConflict {
  return {
    itemId: dto.item_id,
    itemName: dto.item_name,
    itemKind: dto.item_kind,
    kind: dto.kind as ConflictKind,
    resolution: dto.resolution as ConflictResolution,
    path: dto.path,
    uncommitted: dto.uncommitted,
    base: toItemVersion(dto.base),
    mine: toItemVersion(dto.mine),
    theirs: toItemVersion(dto.theirs),
    fields: (dto.fields ?? []).map(toFieldConflict),
  }
}

export function toSyncSession(dto: SyncSessionDto): SyncSession {
  return { id: dto.id, localId: dto.local_id, conflicts: (dto.conflicts ?? []).map(toSyncConflict) }
}

export function toPullResult(dto: SyncPullResultDto): PullResult {
  return { status: toSyncStatus(dto.status), session: dto.session ? toSyncSession(dto.session) : null }
}
