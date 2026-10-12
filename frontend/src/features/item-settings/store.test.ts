import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyAuth } from '@/platform/mappers/request'
import type { ItemSettings } from '@/platform/mappers/tree'

vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(), searchTree: vi.fn(), getItemSettings: vi.fn(), saveItemSettings: vi.fn() }))
vi.mock('@/platform/services/requests', () => ({ parseBulk: vi.fn(), formatBulk: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const tree = await import('@/platform/services/tree')
const requests = await import('@/platform/services/requests')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { isItemDirty, useItemSettingsStore } = await import('./store')

const billing: ItemSettings = {
  id: 'c1',
  kind: 'collection',
  name: 'Billing',
  description: 'Billing API',
  auth: { ...emptyAuth(), type: 'inherit' },
  headers: [{ id: 'h1', key: 'X-Team', value: 'payments', description: '', enabled: true }],
  scripts: { preRequest: '', postResponse: '', disabled: [] },
  inheritScripts: null,
}

function store() {
  return useItemSettingsStore.getState()
}

describe('item settings store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    store().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    vi.mocked(tree.getItemSettings).mockResolvedValue(billing)
    vi.mocked(tree.saveItemSettings).mockImplementation(async (_localId, settings) => settings)
  })

  it('loads an item once and keeps its edits as a dirty draft', async () => {
    await store().open('c1')
    await store().open('c1')
    expect(tree.getItemSettings).toHaveBeenCalledTimes(1)
    expect(store().items.c1).toMatchObject({ status: 'ready', base: billing, draft: billing, dirty: false, section: 'overview' })
    store().edit('c1', { description: 'Edited' })
    expect(isItemDirty(store(), 'c1')).toBe(true)
    expect(store().items.c1?.base?.description).toBe('Billing API')
  })

  it('saves the draft and takes what was written as the new base', async () => {
    await store().open('c1')
    store().edit('c1', { description: 'Edited' })
    await expect(store().save('c1')).resolves.toEqual({ ok: true, value: true })
    expect(tree.saveItemSettings).toHaveBeenCalledWith('ws', expect.objectContaining({ description: 'Edited' }))
    expect(store().items.c1).toMatchObject({ dirty: false, saving: false, base: { description: 'Edited' } })
    await expect(store().save('c1')).resolves.toEqual({ ok: true, value: true })
    expect(tree.saveItemSettings).toHaveBeenCalledTimes(1)
    await expect(store().save('unknown')).resolves.toEqual({ ok: true, value: false })
  })

  it('keeps the edits and shows why saving failed', async () => {
    await store().open('c1')
    store().edit('c1', { description: 'Edited' })
    vi.mocked(tree.saveItemSettings).mockRejectedValue(new AppError('TRE-05500', 'The folder is read-only'))
    const saved = await store().save('c1')
    expect(saved).toMatchObject({ ok: false, error: { code: 'TRE-05500' } })
    expect(store().items.c1).toMatchObject({ dirty: true, saving: false, failure: { error: { code: 'TRE-05500' } } })
  })

  it('switches the headers to bulk text and back, saving the bulk text as headers', async () => {
    await store().open('c1')
    vi.mocked(requests.formatBulk).mockResolvedValue('X-Team:payments')
    await store().setBulk('c1', true)
    expect(store().items.c1?.bulkText).toBe('X-Team:payments')
    store().editBulkText('c1', 'X-Team:billing')
    const parsed = [{ id: 'h2', key: 'X-Team', value: 'billing', description: '', enabled: true }]
    vi.mocked(requests.parseBulk).mockResolvedValue(parsed)
    await store().save('c1')
    expect(tree.saveItemSettings).toHaveBeenCalledWith('ws', expect.objectContaining({ headers: parsed }))
    await store().setBulk('c1', false)
    expect(store().items.c1).toMatchObject({ bulkText: null, draft: { headers: parsed } })
  })

  it('shows bulk text that does not parse', async () => {
    await store().open('c1')
    vi.mocked(requests.formatBulk).mockResolvedValue('')
    await store().setBulk('c1', true)
    vi.mocked(requests.parseBulk).mockRejectedValue(new AppError('REQ-01400', 'Line 1 has no colon'))
    await store().setBulk('c1', false)
    expect(store().items.c1).toMatchObject({ bulkText: '', failure: { error: { code: 'REQ-01400' }, nextStep: expect.stringContaining('key:value') } })
  })

  it('takes a disk change when clean, and flags it as stale when there are unsaved edits', async () => {
    await store().open('c1')
    vi.mocked(tree.getItemSettings).mockResolvedValue({ ...billing, description: 'From disk' })
    await store().refreshAll()
    expect(store().items.c1).toMatchObject({ draft: { description: 'From disk' }, stale: false })
    store().edit('c1', { description: 'Mine' })
    vi.mocked(tree.getItemSettings).mockResolvedValue({ ...billing, description: 'Changed again', name: 'Billing API' })
    await store().refreshAll()
    expect(store().items.c1).toMatchObject({ draft: { description: 'Mine', name: 'Billing API' }, stale: true, dirty: true })
    store().discard('c1')
    expect(store().items.c1).toMatchObject({ draft: { description: 'Changed again' }, stale: false, dirty: false })
  })

  it('does not flag a rename as a conflict with unsaved edits', async () => {
    await store().open('c1')
    store().edit('c1', { description: 'Mine' })
    vi.mocked(tree.getItemSettings).mockResolvedValue({ ...billing, name: 'Payments' })
    await store().load('c1')
    expect(store().items.c1).toMatchObject({ stale: false, draft: { name: 'Payments', description: 'Mine' } })
    store().adoptName('c1', 'Finance')
    expect(store().items.c1).toMatchObject({ base: { name: 'Finance' }, draft: { name: 'Finance' } })
  })

  it('shows a load failure and forgets an item once its tab closes', async () => {
    vi.mocked(tree.getItemSettings).mockRejectedValueOnce(new AppError('TRE-02404', 'The item does not exist'))
    await store().open('c1')
    expect(store().items.c1).toMatchObject({ status: 'ready', loadError: { code: 'TRE-02404' }, draft: null })
    await store().load('c1')
    expect(store().items.c1).toMatchObject({ loadError: null, draft: billing })
    store().setSection('c1', 'scripts')
    expect(store().items.c1?.section).toBe('scripts')
    store().drop('c1')
    expect(store().items.c1).toBeUndefined()
  })
})
