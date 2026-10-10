import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/variableservice', () => ({
  GetTable: vi.fn(),
  SaveTable: vi.fn(),
  SetCurrent: vi.fn(),
  Reset: vi.fn(),
  Persist: vi.fn(),
  AddVariable: vi.fn(),
}))

const VariableService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/variableservice')
const { getVariableTable, saveVariableTable, setCurrentValue, resetVariables, persistVariables } = await import('./variables')

const plainDto = {
  key: 'base_url',
  initial_value: 'https://dev',
  current_value: 'https://dev',
  has_current_override: false,
  secret: false,
  enabled: true,
  description: 'API root',
  duplicate: false,
  surrounding_spaces: false,
}

const secretDto = {
  key: ' token',
  initial_value: 'leaked?',
  current_value: 'leaked?',
  has_current_override: false,
  secret: true,
  secret_state: 'bogus',
  enabled: false,
  description: '',
  duplicate: true,
  surrounding_spaces: true,
}

const plain = {
  key: 'base_url',
  initialValue: 'https://dev',
  currentValue: 'https://dev',
  hasCurrentOverride: false,
  secret: false,
  secretState: null,
  enabled: true,
  description: 'API root',
  duplicate: false,
  surroundingSpaces: false,
}

describe('variables service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('gets a table and never keeps a value on a secret row', async () => {
    vi.mocked(VariableService.GetTable).mockResolvedValue({ scope: 'env:e1', rows: [plainDto, secretDto], warnings: null } as never)
    await expect(getVariableTable('L1', 'env:e1')).resolves.toEqual({
      scope: 'env:e1',
      warnings: [],
      rows: [
        plain,
        {
          key: ' token',
          initialValue: '',
          currentValue: '',
          hasCurrentOverride: false,
          secret: true,
          secretState: 'missing',
          enabled: false,
          description: '',
          duplicate: true,
          surroundingSpaces: true,
        },
      ],
    })
    expect(VariableService.GetTable).toHaveBeenCalledWith({ local_id: 'L1', scope: 'env:e1' })
  })

  it('treats an empty table answer as no rows for the asked scope', async () => {
    vi.mocked(VariableService.GetTable).mockResolvedValue(null as never)
    await expect(getVariableTable('L1', 'globals')).resolves.toEqual({ scope: 'globals', rows: [], warnings: [] })
  })

  it('saves edits, sending only what changed for current and secret values', async () => {
    vi.mocked(VariableService.SaveTable).mockResolvedValue({ scope: 'env:e1', rows: [plainDto], warnings: ['git_history'] } as never)
    const saved = await saveVariableTable(
      'L1',
      'env:e1',
      [
        { key: 'base_url', previousKey: 'url', initialValue: 'https://dev', currentValue: null, secret: false, secretValue: null, enabled: true, description: 'API root' },
        { key: 'user', previousKey: null, initialValue: 'ana', currentValue: '', secret: false, secretValue: null, enabled: false, description: '' },
        { key: 'token', previousKey: 'token', initialValue: 'ignored', currentValue: 'ignored', secret: true, secretValue: null, enabled: true, description: '' },
        { key: 'password', previousKey: null, initialValue: '', currentValue: null, secret: true, secretValue: 's3cr3t', enabled: true, description: '' },
      ],
      true,
    )
    expect(saved.warnings).toEqual(['git_history'])
    expect(VariableService.SaveTable).toHaveBeenCalledWith({
      local_id: 'L1',
      scope: 'env:e1',
      confirm_reveal: true,
      rows: [
        { key: 'base_url', previous_key: 'url', initial_value: 'https://dev', description: 'API root', enabled: true, secret: false, set_current: false },
        { key: 'user', initial_value: 'ana', current_value: '', description: '', enabled: false, secret: false, set_current: true },
        { key: 'token', previous_key: 'token', initial_value: '', description: '', enabled: true, secret: true, set_current: false },
        { key: 'password', initial_value: '', secret_value: 's3cr3t', description: '', enabled: true, secret: true, set_current: false },
      ],
    })
  })

  it('sets a current value, resets and persists rows', async () => {
    vi.mocked(VariableService.SetCurrent).mockResolvedValue({ ...plainDto, current_value: 'x', has_current_override: true } as never)
    await expect(setCurrentValue('L1', 'globals', 'base_url', 'x')).resolves.toMatchObject({ currentValue: 'x', hasCurrentOverride: true })
    expect(VariableService.SetCurrent).toHaveBeenCalledWith({ local_id: 'L1', scope: 'globals', key: 'base_url', value: 'x' })
    vi.mocked(VariableService.Reset).mockResolvedValue({ scope: 'globals', rows: [plainDto], warnings: [] } as never)
    await expect(resetVariables('L1', 'globals', [])).resolves.toMatchObject({ rows: [plain] })
    expect(VariableService.Reset).toHaveBeenCalledWith({ local_id: 'L1', scope: 'globals', keys: [] })
    vi.mocked(VariableService.Persist).mockResolvedValue({ scope: 'globals', rows: [plainDto], warnings: [] } as never)
    await expect(persistVariables('L1', 'globals', ['base_url'])).resolves.toMatchObject({ rows: [plain] })
    expect(VariableService.Persist).toHaveBeenCalledWith({ local_id: 'L1', scope: 'globals', keys: ['base_url'] })
  })

  it('maps backend errors', async () => {
    vi.mocked(VariableService.SaveTable).mockRejectedValue({ cause: { error: { code: 'VAR-01409', description: 'confirm reveal' } } })
    await expect(saveVariableTable('L1', 'globals', [], false)).rejects.toMatchObject({ code: 'VAR-01409' })
    vi.mocked(VariableService.SetCurrent).mockResolvedValue(null as never)
    await expect(setCurrentValue('L1', 'globals', 'k', 'v')).rejects.toMatchObject({ code: 'UNK-00204' })
    vi.mocked(VariableService.Persist).mockRejectedValue({ cause: { error: { code: 'VAR-02422', description: 'secret' } } })
    await expect(persistVariables('L1', 'globals', ['token'])).rejects.toMatchObject({ code: 'VAR-02422' })
  })
})
