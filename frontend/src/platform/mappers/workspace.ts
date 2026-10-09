import type {
  WorkspaceList as WorkspaceListDto,
  WorkspaceSnapshot as WorkspaceSnapshotDto,
  WorkspaceSummary as WorkspaceSummaryDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import { toTreeNode, type TreeNode } from './tree'
import { toUIState, type UIState } from './ui-state'

export interface WorkspaceSummary {
  localId: string
  name: string
  path: string
  missing: boolean
}

export interface WorkspaceList {
  workspaces: WorkspaceSummary[]
  activeLocalId: string | null
}

export interface WorkspaceSnapshot {
  summary: WorkspaceSummary
  tree: TreeNode[]
  uiState: UIState
}

export interface CreateWorkspaceInput {
  name: string
  path: string
}

export function toWorkspaceSummary(dto: WorkspaceSummaryDto): WorkspaceSummary {
  return { localId: dto.local_id, name: dto.name, path: dto.path, missing: dto.missing }
}

export function toWorkspaceList(dto: WorkspaceListDto | null): WorkspaceList {
  return {
    workspaces: (dto?.workspaces ?? []).map(toWorkspaceSummary),
    activeLocalId: dto?.active_local_id ? dto.active_local_id : null,
  }
}

export function toWorkspaceSnapshot(dto: WorkspaceSnapshotDto): WorkspaceSnapshot {
  return {
    summary: toWorkspaceSummary(dto.summary),
    tree: (dto.tree ?? []).map(toTreeNode),
    uiState: toUIState(dto.ui_state ?? null),
  }
}
