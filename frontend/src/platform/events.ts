import { Events } from '@wailsio/runtime'

export const WORKSPACE_TREE_CHANGED = 'workspace:tree-changed'
export const REQUEST_CHANGED_ON_DISK = 'request:changed-on-disk'
export const WORKSPACE_FOLDER_MISSING = 'workspace:folder-missing'

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
