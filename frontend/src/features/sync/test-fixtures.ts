import type { SyncChange, SyncStatus } from '@/platform/mappers/sync'

export function syncStatus(overrides: Partial<SyncStatus> = {}): SyncStatus {
  return {
    state: 'up_to_date',
    localChanges: 0,
    ahead: 0,
    behind: 0,
    lastSyncAt: null,
    lastError: null,
    notice: '',
    settings: { repository: 'https://git.test/acme.git', branch: 'main', subfolder: 'api', accountId: 'a1', clonePath: '/c', autoPull: false, autoPush: false },
    ...overrides,
  }
}

export const pingChange: SyncChange = { itemId: 'r1', itemName: 'Ping', itemKind: 'request', change: 'modified', paths: ['ping.yaml'] }
export const usersChange: SyncChange = { itemId: 'r2', itemName: 'Users', itemKind: 'request', change: 'added', paths: ['users.yaml'] }
