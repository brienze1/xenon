import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/workspaces', () => ({
  listWorkspaces: vi.fn(),
  createWorkspace: vi.fn(),
  activateWorkspace: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn() }))

const services = await import('@/platform/services/workspaces')
const dialogs = await import('@/platform/services/dialogs')
const { useWorkspacesStore } = await import('./store')

const acme = { localId: 'A', name: 'Acme', path: '/w/acme', missing: false }
const globex = { localId: 'G', name: 'Globex', path: '/w/globex', missing: false }

describe('workspaces store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useWorkspacesStore.setState({ workspaces: [], activeLocalId: null, status: 'idle', error: null })
  })

  it('loads the registered workspaces', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme, globex], activeLocalId: 'G' })
    await useWorkspacesStore.getState().load()
    expect(useWorkspacesStore.getState()).toMatchObject({ workspaces: [acme, globex], activeLocalId: 'G', status: 'ready', error: null })
  })

  it('keeps the error when loading fails', async () => {
    vi.mocked(services.listWorkspaces).mockRejectedValue(new AppError('WKS-12500', 'bad app.json'))
    await useWorkspacesStore.getState().load()
    expect(useWorkspacesStore.getState()).toMatchObject({ status: 'ready', error: { code: 'WKS-12500' } })
  })

  it('creates a workspace and makes it active', async () => {
    useWorkspacesStore.setState({ workspaces: [acme], activeLocalId: 'A' })
    vi.mocked(services.createWorkspace).mockResolvedValue(globex)
    await expect(useWorkspacesStore.getState().create('Globex', '/w/globex')).resolves.toBe(true)
    expect(services.createWorkspace).toHaveBeenCalledWith({ name: 'Globex', path: '/w/globex' })
    expect(useWorkspacesStore.getState()).toMatchObject({ workspaces: [acme, globex], activeLocalId: 'G' })
  })

  it('reports a refused creation', async () => {
    vi.mocked(services.createWorkspace).mockRejectedValue(new AppError('WKS-01409', 'not empty'))
    await expect(useWorkspacesStore.getState().create('Acme', '/w/full')).resolves.toBe(false)
    expect(useWorkspacesStore.getState().error?.code).toBe('WKS-01409')
    useWorkspacesStore.getState().clearError()
    expect(useWorkspacesStore.getState().error).toBeNull()
  })

  it('activates a workspace and refreshes its summary', async () => {
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: 'G' })
    vi.mocked(services.activateWorkspace).mockResolvedValue({ summary: { ...acme, name: 'Acme Corp' } })
    await useWorkspacesStore.getState().activate('A')
    expect(useWorkspacesStore.getState()).toMatchObject({ activeLocalId: 'A', workspaces: [{ name: 'Acme Corp' }, globex] })
  })

  it('flags a missing folder when activation finds it gone', async () => {
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: 'A' })
    vi.mocked(services.activateWorkspace).mockRejectedValue(new AppError('WKS-06404', 'missing'))
    await useWorkspacesStore.getState().activate('A')
    expect(useWorkspacesStore.getState()).toMatchObject({
      activeLocalId: null,
      error: { code: 'WKS-06404' },
      workspaces: [{ localId: 'A', missing: true }, { localId: 'G', missing: false }],
    })
  })

  it('keeps the active workspace when another activation fails', async () => {
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: 'G' })
    vi.mocked(services.activateWorkspace).mockRejectedValue(new AppError('WKS-06404', 'missing'))
    await useWorkspacesStore.getState().activate('A')
    expect(useWorkspacesStore.getState().activeLocalId).toBe('G')
    vi.mocked(services.activateWorkspace).mockRejectedValue(new AppError('WKS-04422', 'newer'))
    await useWorkspacesStore.getState().activate('G')
    expect(useWorkspacesStore.getState()).toMatchObject({ activeLocalId: 'G', workspaces: [{ missing: true }, { missing: false }] })
  })

  it('chooses a folder through the native picker', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/w/new')
    await expect(useWorkspacesStore.getState().chooseFolder('Choose')).resolves.toBe('/w/new')
    vi.mocked(dialogs.pickFolder).mockRejectedValue(new AppError('DLG-00500', 'no dialogs'))
    await expect(useWorkspacesStore.getState().chooseFolder('Choose')).resolves.toBeNull()
    expect(useWorkspacesStore.getState().error?.code).toBe('DLG-00500')
  })

  it('marks a workspace whose folder disappeared', () => {
    useWorkspacesStore.setState({ workspaces: [acme, globex] })
    useWorkspacesStore.getState().markMissing('G')
    expect(useWorkspacesStore.getState().workspaces).toEqual([acme, { ...globex, missing: true }])
  })
})
