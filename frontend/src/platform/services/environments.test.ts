import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/environmentservice', () => ({
  List: vi.fn(),
  SetActive: vi.fn(),
}))

const EnvironmentService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/environmentservice')
const { listEnvironments, setActiveEnvironment } = await import('./environments')

describe('environments service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('lists the environments with the active one and maps unknown colours to none', async () => {
    vi.mocked(EnvironmentService.List).mockResolvedValue({
      active_id: 'e1',
      environments: [
        { id: 'e1', name: 'dev', color: 'green', confirm_before_send: false },
        { id: 'e2', name: 'prod', color: 'red', confirm_before_send: true, error: '' },
        { id: 'error:broken.yaml', name: 'broken.yaml', color: 'purple', confirm_before_send: false, error: 'broken.yaml cannot be read' },
      ],
    } as never)
    await expect(listEnvironments('L1')).resolves.toEqual({
      activeId: 'e1',
      environments: [
        { id: 'e1', name: 'dev', color: 'green', confirmBeforeSend: false, error: null },
        { id: 'e2', name: 'prod', color: 'red', confirmBeforeSend: true, error: null },
        { id: 'error:broken.yaml', name: 'broken.yaml', color: 'none', confirmBeforeSend: false, error: 'broken.yaml cannot be read' },
      ],
    })
    expect(EnvironmentService.List).toHaveBeenCalledWith({ local_id: 'L1' })
  })

  it('treats an empty answer as no environments and no active one', async () => {
    vi.mocked(EnvironmentService.List).mockResolvedValue(null as never)
    await expect(listEnvironments('L1')).resolves.toEqual({ environments: [], activeId: null })
    vi.mocked(EnvironmentService.List).mockResolvedValue({ active_id: '', environments: null } as never)
    await expect(listEnvironments('L1')).resolves.toEqual({ environments: [], activeId: null })
  })

  it('sets the active environment, "No environment" being an empty id', async () => {
    vi.mocked(EnvironmentService.SetActive).mockResolvedValue({ active_id: 'e2' } as never)
    await expect(setActiveEnvironment('L1', 'e2')).resolves.toBe('e2')
    expect(EnvironmentService.SetActive).toHaveBeenCalledWith({ local_id: 'L1', id: 'e2' })
    vi.mocked(EnvironmentService.SetActive).mockResolvedValue({ active_id: '' } as never)
    await expect(setActiveEnvironment('L1', null)).resolves.toBeNull()
    expect(EnvironmentService.SetActive).toHaveBeenLastCalledWith({ local_id: 'L1', id: '' })
    vi.mocked(EnvironmentService.SetActive).mockResolvedValue(null as never)
    await expect(setActiveEnvironment('L1', null)).resolves.toBeNull()
  })

  it('maps backend errors', async () => {
    vi.mocked(EnvironmentService.SetActive).mockRejectedValue({ cause: { error: { code: 'ENV-01404', description: 'the environment does not exist' } } })
    await expect(setActiveEnvironment('L1', 'gone')).rejects.toMatchObject({ code: 'ENV-01404' })
    vi.mocked(EnvironmentService.List).mockRejectedValue({ cause: { error: { code: 'WKS-07404', description: 'workspace is not registered' } } })
    await expect(listEnvironments('L1')).rejects.toMatchObject({ code: 'WKS-07404' })
  })
})
