import { Events } from '@wailsio/runtime'
import { toConsoleEntriesEvent, type ConsoleEntriesEvent } from './mappers/console'
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
export const CONSOLE_ENTRIES = 'console:entries'
export const COOKIES_CHANGED = 'cookies:changed'

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

type Unsubscribe = () => void

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
