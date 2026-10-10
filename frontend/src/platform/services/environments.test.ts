import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/environmentservice', () => ({
  List: vi.fn(),
  SetActive: vi.fn(),
  Get: vi.fn(),
  Create: vi.fn(),
  Update: vi.fn(),
  Duplicate: vi.fn(),
  Delete: vi.fn(),
}))

const EnvironmentService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/environmentservice')
const { listEnvironments, setActiveEnvironment, getEnvironment, createEnvironment, updateEnvironment, duplicateEnvironment, deleteEnvironment } =
  await import('./environments')

const environmentDto = {
  id: 'e1',
  name: 'dev',
  description: 'Development',
  color: 'cyan',
  confirm_before_send: false,
  variables: [
    {
      key: 'base_url',
      initial_value: 'https://dev',
      current_value: 'https://local',
      has_current_override: true,
      secret: false,
      enabled: true,
      description: '',
      duplicate: false,
      surrounding_spaces: false,
    },
    {
      key: 'api_key',
      initial_value: '',
      current_value: '',
      has_current_override: false,
      secret: true,
      secret_state: 'locked',
      enabled: true,
      description: 'token',
      duplicate: false,
      surrounding_spaces: false,
    },
  ],
}

const environment = {
  id: 'e1',
  name: 'dev',
  description: 'Development',
  color: 'cyan',
  confirmBeforeSend: false,
  variables: [
    {
      key: 'base_url',
      initialValue: 'https://dev',
      currentValue: 'https://local',
      hasCurrentOverride: true,
      secret: false,
      secretState: null,
      enabled: true,
      description: '',
      duplicate: false,
      surroundingSpaces: false,
    },
    {
      key: 'api_key',
      initialValue: '',
      currentValue: '',
      hasCurrentOverride: false,
      secret: true,
      secretState: 'locked',
      enabled: true,
      description: 'token',
      duplicate: false,
      surroundingSpaces: false,
    },
  ],
}

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

  it('gets, creates, duplicates and updates an environment with its rows', async () => {
    vi.mocked(EnvironmentService.Get).mockResolvedValue(environmentDto as never)
    await expect(getEnvironment('L1', 'e1')).resolves.toEqual(environment)
    expect(EnvironmentService.Get).toHaveBeenCalledWith({ local_id: 'L1', id: 'e1' })
    vi.mocked(EnvironmentService.Create).mockResolvedValue({ ...environmentDto, variables: null } as never)
    await expect(createEnvironment('L1', 'dev')).resolves.toEqual({ ...environment, variables: [] })
    expect(EnvironmentService.Create).toHaveBeenCalledWith({ local_id: 'L1', name: 'dev' })
    vi.mocked(EnvironmentService.Duplicate).mockResolvedValue({ ...environmentDto, id: 'e2', name: 'dev Copy' } as never)
    await expect(duplicateEnvironment('L1', 'e1')).resolves.toMatchObject({ id: 'e2', name: 'dev Copy' })
    expect(EnvironmentService.Duplicate).toHaveBeenCalledWith({ local_id: 'L1', id: 'e1' })
    vi.mocked(EnvironmentService.Update).mockResolvedValue({ ...environmentDto, name: 'prod', color: 'red' } as never)
    await expect(
      updateEnvironment('L1', { id: 'e1', name: 'prod', description: 'Development', color: 'red', confirmBeforeSend: true }),
    ).resolves.toMatchObject({ name: 'prod', color: 'red' })
    expect(EnvironmentService.Update).toHaveBeenCalledWith({
      local_id: 'L1',
      id: 'e1',
      name: 'prod',
      description: 'Development',
      color: 'red',
      confirm_before_send: true,
    })
  })

  it('reports an empty answer for a single environment as an error', async () => {
    vi.mocked(EnvironmentService.Get).mockResolvedValue(null as never)
    await expect(getEnvironment('L1', 'e1')).rejects.toMatchObject({ code: 'UNK-00204' })
  })

  it('deletes an environment and returns the active one afterwards', async () => {
    vi.mocked(EnvironmentService.Delete).mockResolvedValueOnce({ active_id: '' } as never).mockResolvedValueOnce({ active_id: 'e2' } as never)
    await expect(deleteEnvironment('L1', 'e1')).resolves.toBeNull()
    expect(EnvironmentService.Delete).toHaveBeenCalledWith({ local_id: 'L1', id: 'e1' })
    await expect(deleteEnvironment('L1', 'e3')).resolves.toBe('e2')
  })

  it('maps a duplicate name', async () => {
    vi.mocked(EnvironmentService.Create).mockRejectedValue({ cause: { error: { code: 'ENV-02409', description: 'name taken' } } })
    await expect(createEnvironment('L1', 'dev')).rejects.toMatchObject({ code: 'ENV-02409' })
  })
})
