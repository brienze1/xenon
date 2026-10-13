import { Events } from '@wailsio/runtime'
import { toAIAccessRuntime, toPendingApproval, type AIAccessRuntime, type PendingApproval } from './mappers/ai-access'
import { toConsoleEntriesEvent, type ConsoleEntriesEvent } from './mappers/console'
import { toSyncStatus, type SyncStatus } from './mappers/sync'
import { toVaultStatus, type VaultStatus } from './mappers/vault'

export const WORKSPACE_TREE_CHANGED = 'workspace:tree-changed'
export const REQUEST_CHANGED_ON_DISK = 'request:changed-on-disk'
export const WORKSPACE_FOLDER_MISSING = 'workspace:folder-missing'
export const VARIABLES_CHANGED = 'variables:changed'
export const VAULT_STATUS_CHANGED = 'vault:status-changed'
export const ACTIVE_ENVIRONMENT_REMOVED = 'active_environment_removed'
export const IMPORT_PROGRESS = 'import:progress'
export const IMPORT_FILES_DROPPED = 'import:files-dropped'
export const IMPORT_REQUESTED = 'import:requested'
export const EXPORT_PROGRESS = 'export:progress'
export const EXPORT_REQUESTED = 'export:requested'
export const VIEW_ZOOM_REQUESTED = 'view:zoom-requested'
export const CONSOLE_ENTRIES = 'console:entries'
export const COOKIES_CHANGED = 'cookies:changed'
export const GIT_HOST_KEY_UNKNOWN = 'git:host-key-unknown'
export const GIT_STATUS_CHANGED = 'git:status-changed'
export const GIT_PROGRESS = 'git:progress'
export const AI_ACCESS_STATUS_CHANGED = 'ai:access-status-changed'
export const AI_APPROVAL_REQUESTED = 'ai:approval-requested'
export const AI_APPROVAL_RESOLVED = 'ai:approval-resolved'
export const HISTORY_CHANGED = 'history:changed'
export const WORKSPACES_CHANGED = 'workspaces:changed'

export interface WorkspaceTreeChanged {
  localId: string
}

export interface RequestChangedOnDisk {
  localId: string
  requestId: string
  deleted: boolean
  revision: string
}

export interface WorkspaceFolderMissing {
  localId: string
}

export interface VariablesChanged {
  localId: string
  scopes: string[]
  reason: string | null
}

export interface ImportProgress {
  sourceId: string
  currentItem: string
  done: number
  total: number
}

export interface ExportProgress {
  done: number
  total: number
}

export interface HostKeyUnknown {
  repository: string
  host: string
  fingerprint: string
  keyType: string
}

export interface GitStatusChanged {
  localId: string
  status: SyncStatus
}

export interface GitProgress {
  localId: string
  phase: string
  done: number
  total: number
}

export interface ApprovalResolved {
  id: string
  /** approved, denied, timed_out or cancelled (app quitting, client gone). */
  outcome: string
}

type Unsubscribe = () => void

/** The MCP server started, stopped, failed, or its client count changed. */
export function onAIAccessStatusChanged(callback: (status: AIAccessRuntime) => void): Unsubscribe {
  return Events.On(AI_ACCESS_STATUS_CHANGED, (event) => {
    callback(toAIAccessRuntime(event.data))
  })
}

/** An AI send to a confirm-before-send environment waits for Approve / Deny. */
export function onApprovalRequested(callback: (approval: PendingApproval) => void): Unsubscribe {
  return Events.On(AI_APPROVAL_REQUESTED, (event) => {
    callback(toPendingApproval(event.data))
  })
}

/** A pending approval was answered, timed out, or cancelled; its prompt goes away. */
export function onApprovalResolved(callback: (resolved: ApprovalResolved) => void): Unsubscribe {
  return Events.On(AI_APPROVAL_RESOLVED, (event) => {
    callback({ id: event.data.id, outcome: event.data.outcome })
  })
}

/** A workspace's history gained or lost entries (a user or AI send, an AI action, a clear). */
export function onHistoryChanged(callback: (localId: string) => void): Unsubscribe {
  return Events.On(HISTORY_CHANGED, (event) => {
    callback(event.data.local_id)
  })
}

/** MCP created, opened, renamed, removed or switched a workspace; `activeLocalId` is empty when none is active. */
export function onWorkspacesChanged(callback: (activeLocalId: string | null) => void): Unsubscribe {
  return Events.On(WORKSPACES_CHANGED, (event) => {
    callback(event.data.active_local_id ? event.data.active_local_id : null)
  })
}

export function onGitStatusChanged(callback: (event: GitStatusChanged) => void): Unsubscribe {
  return Events.On(GIT_STATUS_CHANGED, (event) => {
    callback({ localId: event.data.local_id, status: toSyncStatus(event.data.status) })
  })
}

export function onGitProgress(callback: (event: GitProgress) => void): Unsubscribe {
  return Events.On(GIT_PROGRESS, (event) => {
    callback({ localId: event.data.local_id, phase: event.data.phase, done: event.data.done, total: event.data.total })
  })
}

export function onGitHostKeyUnknown(callback: (event: HostKeyUnknown) => void): Unsubscribe {
  return Events.On(GIT_HOST_KEY_UNKNOWN, (event) => {
    callback({ repository: event.data.repository, host: event.data.host, fingerprint: event.data.fingerprint, keyType: event.data.key_type })
  })
}

export function onWorkspaceTreeChanged(callback: (event: WorkspaceTreeChanged) => void): Unsubscribe {
  return Events.On(WORKSPACE_TREE_CHANGED, (event) => {
    callback({ localId: event.data.local_id })
  })
}

export function onRequestChangedOnDisk(callback: (event: RequestChangedOnDisk) => void): Unsubscribe {
  return Events.On(REQUEST_CHANGED_ON_DISK, (event) => {
    callback({
      localId: event.data.local_id,
      requestId: event.data.request_id,
      deleted: event.data.deleted,
      revision: event.data.revision,
    })
  })
}

export function onWorkspaceFolderMissing(callback: (event: WorkspaceFolderMissing) => void): Unsubscribe {
  return Events.On(WORKSPACE_FOLDER_MISSING, (event) => {
    callback({ localId: event.data.local_id })
  })
}

export function onVariablesChanged(callback: (event: VariablesChanged) => void): Unsubscribe {
  return Events.On(VARIABLES_CHANGED, (event) => {
    callback({ localId: event.data.local_id, scopes: event.data.scopes ?? [], reason: event.data.reason ? event.data.reason : null })
  })
}

export function onVaultStatusChanged(callback: (status: VaultStatus) => void): Unsubscribe {
  return Events.On(VAULT_STATUS_CHANGED, (event) => {
    callback(toVaultStatus(event.data))
  })
}

export function onImportProgress(callback: (progress: ImportProgress) => void): Unsubscribe {
  return Events.On(IMPORT_PROGRESS, (event) => {
    callback({ sourceId: event.data.source_id, currentItem: event.data.current_item, done: event.data.done, total: event.data.total })
  })
}

/** Files or folders dropped on the window; Go delivers their absolute paths. */
export function onImportFilesDropped(callback: (paths: string[]) => void): Unsubscribe {
  return Events.On(IMPORT_FILES_DROPPED, (event) => {
    callback(event.data.paths ?? [])
  })
}

/** The native File → Import… menu item. */
export function onImportRequested(callback: () => void): Unsubscribe {
  return Events.On(IMPORT_REQUESTED, () => {
    callback()
  })
}

/** Files written by the running export, announced every 50 files when it writes more than 200. */
export function onExportProgress(callback: (progress: ExportProgress) => void): Unsubscribe {
  return Events.On(EXPORT_PROGRESS, (event) => {
    callback({ done: event.data.done, total: event.data.total })
  })
}

/** The native File → Export Workspace… menu item. */
export function onExportRequested(callback: () => void): Unsubscribe {
  return Events.On(EXPORT_REQUESTED, () => {
    callback()
  })
}

export type ZoomDirection = 'in' | 'out' | 'reset'

/** The native View → Zoom In / Zoom Out / Actual Size menu items (macOS). */
export function onViewZoomRequested(callback: (direction: ZoomDirection) => void): Unsubscribe {
  return Events.On(VIEW_ZOOM_REQUESTED, (event) => {
    callback(event.data.direction as ZoomDirection)
  })
}

/** New console entries, batched by Go at most every 50 ms; `cleared` means the console was emptied first. */
export function onConsoleEntries(callback: (event: ConsoleEntriesEvent) => void): Unsubscribe {
  return Events.On(CONSOLE_ENTRIES, (event) => {
    callback(toConsoleEntriesEvent(event.data))
  })
}

/** The workspace's cookie jar changed (a send stored cookies, a script or the Cookies dialog edited them). */
export function onCookiesChanged(callback: (localId: string) => void): Unsubscribe {
  return Events.On(COOKIES_CHANGED, (event) => {
    callback(event.data.local_id)
  })
}
