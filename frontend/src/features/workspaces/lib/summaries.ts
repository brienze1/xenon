import type { WorkspaceSummary } from '@/platform/mappers/workspace'

export function replaceSummary(workspaces: WorkspaceSummary[], summary: WorkspaceSummary): WorkspaceSummary[] {
  return workspaces.map((workspace) => (workspace.localId === summary.localId ? summary : workspace))
}

export function withMissing(workspaces: WorkspaceSummary[], localId: string): WorkspaceSummary[] {
  return workspaces.map((workspace) => (workspace.localId === localId ? { ...workspace, missing: true } : workspace))
}
