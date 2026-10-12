import { emptyDraft, type RequestDraft } from '@/platform/mappers/request'
import type { SyncChange, SyncConflict, SyncSession, SyncStatus } from '@/platform/mappers/sync'

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

export function requestDraft(method: string, url: string): RequestDraft {
  return { ...emptyDraft(), id: 'r1', name: 'Users', method, url, headers: [{ id: 'h1', key: 'Accept', value: 'application/json', description: '', enabled: true }] }
}

export function requestConflict(overrides: Partial<SyncConflict> = {}): SyncConflict {
  return {
    itemId: 'r1',
    itemName: 'Users',
    itemKind: 'request',
    kind: 'both_modified',
    resolution: 'unresolved',
    path: 'collections/api/users.yaml',
    uncommitted: false,
    base: { text: 'method: GET', request: requestDraft('GET', 'https://a.test/users'), environment: null },
    mine: { text: 'method: PATCH', request: requestDraft('PATCH', 'https://a.test/users'), environment: null },
    theirs: { text: 'method: POST', request: requestDraft('POST', 'https://a.test/users?v=2'), environment: null },
    fields: [],
    ...overrides,
  }
}

export const environmentConflict: SyncConflict = {
  itemId: 'e1',
  itemName: 'Dev',
  itemKind: 'environment',
  kind: 'delete_modify',
  resolution: 'unresolved',
  path: 'environments/dev.yaml',
  uncommitted: true,
  base: null,
  mine: null,
  theirs: {
    text: 'name: Dev',
    request: null,
    environment: {
      name: 'Dev',
      color: 'none',
      variables: [
        { key: 'base_url', value: 'https://theirs.test', description: '', secret: false, enabled: true },
        { key: 'token', value: '', description: '', secret: true, enabled: true },
      ],
    },
  },
  fields: [],
}

export const fieldConflict: SyncConflict = {
  itemId: 'c1',
  itemName: 'API',
  itemKind: 'collection',
  kind: 'field',
  resolution: 'unresolved',
  path: 'collections/api/collection.yaml',
  uncommitted: false,
  base: { text: 'description: The API', request: null, environment: null },
  mine: { text: 'description: Mine', request: null, environment: null },
  theirs: { text: 'description: Theirs', request: null, environment: null },
  fields: [
    { path: 'description', mine: 'Mine', theirs: 'Theirs', resolution: 'unresolved' },
    { path: 'auth.type', mine: null, theirs: 'bearer', resolution: 'unresolved' },
  ],
}

export function syncSession(conflicts: SyncConflict[] = [requestConflict()]): SyncSession {
  return { id: 's1', localId: 'L', conflicts }
}
