import * as WorkspaceService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/workspaceservice'
import { AppError, toAppError } from '../errors'
import {
  toWorkspaceList,
  toWorkspaceSnapshot,
  toWorkspaceSummary,
  type CreateWorkspaceInput,
  type WorkspaceList,
  type WorkspaceSnapshot,
  type WorkspaceSummary,
} from '../mappers/workspace'

const EMPTY_RESPONSE = 'UNK-00204'

export async function listWorkspaces(): Promise<WorkspaceList> {
  try {
    return toWorkspaceList(await WorkspaceService.List())
  } catch (error) {
    throw toAppError(error)
  }
}

export async function createWorkspace(input: CreateWorkspaceInput): Promise<WorkspaceSummary> {
  try {
    return toWorkspaceSummary(required(await WorkspaceService.Create({ name: input.name, path: input.path })))
  } catch (error) {
    throw toAppError(error)
  }
}

export async function activateWorkspace(localId: string): Promise<WorkspaceSnapshot> {
  try {
    return toWorkspaceSnapshot(required(await WorkspaceService.Activate({ local_id: localId })))
  } catch (error) {
    throw toAppError(error)
  }
}

export async function openWorkspace(path: string): Promise<WorkspaceSummary> {
  try {
    return toWorkspaceSummary(required(await WorkspaceService.Open({ path })))
  } catch (error) {
    throw toAppError(error)
  }
}

export async function renameWorkspace(localId: string, name: string): Promise<WorkspaceSummary> {
  try {
    return toWorkspaceSummary(required(await WorkspaceService.Rename({ local_id: localId, name })))
  } catch (error) {
    throw toAppError(error)
  }
}

export async function removeWorkspace(localId: string, moveToTrash: boolean): Promise<WorkspaceList> {
  try {
    return toWorkspaceList(await WorkspaceService.Remove({ local_id: localId, move_to_trash: moveToTrash }))
  } catch (error) {
    throw toAppError(error)
  }
}

export async function locateWorkspace(localId: string, path: string): Promise<WorkspaceSummary> {
  try {
    return toWorkspaceSummary(required(await WorkspaceService.Locate({ local_id: localId, path })))
  } catch (error) {
    throw toAppError(error)
  }
}

function required<T>(value: T | null): T {
  if (value === null) {
    throw new AppError(EMPTY_RESPONSE, 'The backend returned no data')
  }
  return value
}
