import type { SyncStatus } from '@/platform/mappers/sync'

export type StatusTone = 'ok' | 'pending' | 'attention'

export const NO_ACCOUNT = 'No account - choose one'

/** The footer wording of a sync status (spec FR-006); null when the workspace has no repository. */
export function statusLabel(status: SyncStatus): string | null {
  switch (status.state) {
    case 'not_linked':
      return null
    case 'up_to_date':
      return 'Up to date'
    case 'local_changes':
      return `Local changes (${status.localChanges})`
    case 'ahead':
      return `Ahead (${status.ahead} to push)`
    case 'behind':
      return `Behind (${status.behind} to pull)`
    case 'diverged':
      return `Ahead ${status.ahead} · Behind ${status.behind}`
    case 'conflict':
      return 'Conflict'
    case 'syncing':
      return 'Syncing…'
    default:
      return status.lastError?.message === NO_ACCOUNT ? NO_ACCOUNT : 'Error'
  }
}

export function statusTone(status: SyncStatus): StatusTone {
  switch (status.state) {
    case 'up_to_date':
      return 'ok'
    case 'conflict':
    case 'error':
      return 'attention'
    default:
      return 'pending'
  }
}

export function changeLabel(change: string): string {
  return change.charAt(0).toUpperCase() + change.slice(1)
}
