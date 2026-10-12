import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { Environment, EnvironmentList } from '@/platform/mappers/environment'
import type { VariableRow } from '@/platform/mappers/variable'
import type { VaultStatus } from '@/platform/mappers/vault'

vi.mock('@/platform/services/environments', () => ({
  listEnvironments: vi.fn(),
  setActiveEnvironment: vi.fn(),
  getEnvironment: vi.fn(),
  createEnvironment: vi.fn(),
  updateEnvironment: vi.fn(),
  duplicateEnvironment: vi.fn(),
  deleteEnvironment: vi.fn(),
}))
vi.mock('@/platform/services/variables', () => ({
  CONFIRM_REVEAL_REQUIRED: 'VAR-01409',
  getVariableTable: vi.fn(),
  saveVariableTable: vi.fn(),
  resetVariables: vi.fn(),
  persistVariables: vi.fn(),
}))
vi.mock('@/platform/services/vault', () => ({
  PASSPHRASE_SETUP_REQUIRED: 'VLT-05409',
  VAULT_LOCKED: 'VLT-04409',
  getVaultStatus: vi.fn(),
  setupPassphrase: vi.fn(),
  unlockVault: vi.fn(),
  revealSecret: vi.fn(),
  copySecret: vi.fn(),
  resetVault: vi.fn(),
}))

const services = await import('@/platform/services/environments')
const variables = await import('@/platform/services/variables')
const vault = await import('@/platform/services/vault')
const { isScopeDirty, useEnvironmentsStore } = await import('./store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')

const dev = { id: 'e1', name: 'dev', color: 'green', confirmBeforeSend: false, error: null } as const
const prod = { id: 'e2', name: 'prod', color: 'red', confirmBeforeSend: true, error: null } as const
const list: EnvironmentList = { environments: [dev, prod], activeId: 'e1' }
const devDetail: Environment = { id: 'e1', name: 'dev', description: 'Development', color: 'green', confirmBeforeSend: false, variables: [] }
const row: VariableRow = {
  key: 'base_url',
  initialValue: 'A',
  currentValue: 'B',
  hasCurrentOverride: true,
  secret: false,
  secretState: null,
  enabled: true,
  description: '',
  duplicate: false,
  surroundingSpaces: false,
}
const table = { scope: 'env:e1', rows: [row], warnings: [] }
const unlocked: VaultStatus = { mode: 'passphrase', needsPassphraseSetup: false, lockedLocalIds: [], brokenLocalIds: [] }

function store() {
  return useEnvironmentsStore.getState()
}

function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((done) => {
    resolve = done
  })
  return { promise, resolve }
}

describe('environments store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    store().reset()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
  })

  afterEach(() => useWorkspacesStore.setState({ activeLocalId: null }))

  it('loads the environments and the active one of the active workspace', async () => {
    vi.mocked(services.listEnvironments).mockResolvedValue(list)
    const loading = store().load()
    expect(store().status).toBe('loading')
    await loading
    expect(services.listEnvironments).toHaveBeenCalledWith('A')
    expect(store()).toMatchObject({ environments: [dev, prod], activeId: 'e1', status: 'ready', error: null })
  })

  it('does nothing without an active workspace', async () => {
    useWorkspacesStore.setState({ activeLocalId: null })
    await store().load()
    await store().setActive('e1')
    expect(services.listEnvironments).not.toHaveBeenCalled()
    expect(services.setActiveEnvironment).not.toHaveBeenCalled()
    expect(store().status).toBe('idle')
  })

  it('ignores answers that arrive after a workspace switch', async () => {
    const pending = deferred<EnvironmentList>()
    vi.mocked(services.listEnvironments).mockReturnValue(pending.promise)
    const loading = store().load()
    useWorkspacesStore.setState({ activeLocalId: 'G' })
    pending.resolve(list)
    await loading
    expect(store().environments).toEqual([])
    vi.mocked(services.listEnvironments).mockRejectedValue(new AppError('ENV-05500', 'disk'))
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    const failing = store().load()
    useWorkspacesStore.setState({ activeLocalId: 'G' })
    await failing
    expect(store().error).toBeNull()
  })

  it('keeps the error of a failed load', async () => {
    vi.mocked(services.listEnvironments).mockRejectedValue(new AppError('ENV-05500', 'disk'))
    await store().load()
    expect(store()).toMatchObject({ status: 'ready', error: { code: 'ENV-05500' } })
    store().clearError()
    expect(store().error).toBeNull()
  })

  it('sets the active environment and "No environment"', async () => {
    useEnvironmentsStore.setState({ environments: [dev, prod], activeId: 'e1' })
    vi.mocked(services.setActiveEnvironment).mockResolvedValueOnce('e2').mockResolvedValueOnce(null)
    await store().setActive('e2')
    expect(services.setActiveEnvironment).toHaveBeenCalledWith('A', 'e2')
    expect(store().activeId).toBe('e2')
    await store().setActive(null)
    expect(services.setActiveEnvironment).toHaveBeenLastCalledWith('A', null)
    expect(store().activeId).toBeNull()
  })

  it('keeps the previous active environment when setting fails or the workspace changed', async () => {
    useEnvironmentsStore.setState({ environments: [dev, prod], activeId: 'e1' })
    vi.mocked(services.setActiveEnvironment).mockRejectedValue(new AppError('ENV-01404', 'gone'))
    await store().setActive('e2')
    expect(store()).toMatchObject({ activeId: 'e1', error: { code: 'ENV-01404' } })
    const pending = deferred<string | null>()
    vi.mocked(services.setActiveEnvironment).mockReturnValue(pending.promise)
    const setting = store().setActive('e2')
    expect(store().error).toBeNull()
    useWorkspacesStore.setState({ activeLocalId: 'G' })
    pending.resolve('e2')
    await setting
    expect(store().activeId).toBe('e1')
  })

  it('opens and closes the switcher and resets everything', () => {
    useEnvironmentsStore.setState({ environments: [dev], activeId: 'e1', status: 'ready', error: new AppError('X', 'y') })
    store().openSwitcher()
    expect(store().switcherOpen).toBe(true)
    store().setSwitcherOpen(false)
    expect(store().switcherOpen).toBe(false)
    store().openSwitcher()
    store().reset()
    expect(store()).toMatchObject({ environments: [], activeId: null, status: 'idle', error: null, switcherOpen: false })
  })

  it('creates, renames, duplicates and deletes environments, reloading the list', async () => {
    vi.mocked(services.listEnvironments).mockResolvedValue(list)
    vi.mocked(services.createEnvironment).mockResolvedValue(devDetail)
    await expect(store().createEnvironment('dev')).resolves.toEqual({ ok: true, value: devDetail })
    expect(services.createEnvironment).toHaveBeenCalledWith('A', 'dev')
    expect(store().environments).toEqual([dev, prod])
    vi.mocked(services.getEnvironment).mockResolvedValue(devDetail)
    vi.mocked(services.updateEnvironment).mockResolvedValue({ ...devDetail, name: 'local' })
    await expect(store().renameEnvironment('e1', 'local')).resolves.toBeNull()
    expect(services.updateEnvironment).toHaveBeenCalledWith('A', {
      id: 'e1',
      name: 'local',
      description: 'Development',
      color: 'green',
      confirmBeforeSend: false,
    })
    vi.mocked(services.duplicateEnvironment).mockResolvedValue({ ...devDetail, id: 'e3', name: 'dev Copy' })
    await expect(store().duplicateEnvironment('e1')).resolves.toMatchObject({ ok: true, value: { id: 'e3' } })
    useEnvironmentsStore.setState({ tables: { 'env:e1': { ...table, status: 'ready', error: null }, globals: { ...table, status: 'ready', error: null } } })
    vi.mocked(services.deleteEnvironment).mockResolvedValue(null)
    await expect(store().deleteEnvironment('e1')).resolves.toBeNull()
    expect(services.deleteEnvironment).toHaveBeenCalledWith('A', 'e1')
    expect(Object.keys(store().tables)).toEqual(['globals'])
    expect(services.listEnvironments).toHaveBeenCalledTimes(4)
  })

  it('returns environment errors without reloading', async () => {
    vi.mocked(services.createEnvironment).mockRejectedValue(new AppError('ENV-02409', 'taken'))
    await expect(store().createEnvironment('dev')).resolves.toMatchObject({ ok: false, error: { code: 'ENV-02409' } })
    vi.mocked(services.getEnvironment).mockRejectedValue(new AppError('ENV-01404', 'gone'))
    await expect(store().renameEnvironment('e9', 'x')).resolves.toMatchObject({ code: 'ENV-01404' })
    vi.mocked(services.duplicateEnvironment).mockRejectedValue(new AppError('ENV-01404', 'gone'))
    await expect(store().duplicateEnvironment('e9')).resolves.toMatchObject({ ok: false })
    vi.mocked(services.deleteEnvironment).mockRejectedValue(new AppError('ENV-01404', 'gone'))
    await expect(store().deleteEnvironment('e9')).resolves.toMatchObject({ code: 'ENV-01404' })
    expect(services.listEnvironments).not.toHaveBeenCalled()
  })

  it('keeps a draft per loaded table and only marks it dirty on an edit', async () => {
    vi.mocked(variables.getVariableTable).mockResolvedValue(table)
    await store().loadTable('env:e1')
    expect(store().drafts['env:e1']).toMatchObject({ dirty: false, base: table.rows })
    expect(isScopeDirty(store(), 'env:e1')).toBe(false)
    store().editDraft('env:e1', (rows) => rows.map((row) => ({ ...row, description: 'edited' })))
    expect(isScopeDirty(store(), 'env:e1')).toBe(true)
    store().discardDraft('env:e1')
    expect(isScopeDirty(store(), 'env:e1')).toBe(false)
    expect(store().drafts['env:e1']?.rows[0]?.description).toBe(table.rows[0]?.description)
    store().editDraft('env:missing', (rows) => rows)
    expect(store().drafts['env:missing']).toBeUndefined()
    store().reset()
    expect(store().drafts).toEqual({})
  })

  it('keeps a dirty draft when the table reloads, and refreshes a clean one', async () => {
    vi.mocked(variables.getVariableTable).mockResolvedValue(table)
    await store().loadTable('env:e1')
    store().editDraft('env:e1', (rows) => rows.map((row) => ({ ...row, description: 'mine' })))
    const changed = { ...table, rows: table.rows.map((row) => ({ ...row, initialValue: 'theirs' })) }
    vi.mocked(variables.getVariableTable).mockResolvedValue(changed)
    await store().loadTable('env:e1')
    expect(store().drafts['env:e1']).toMatchObject({ dirty: true, base: table.rows })
    store().discardDraft('env:e1')
    expect(store().drafts['env:e1']?.base).toBe(store().tables['env:e1']?.rows)
  })

  it('saves a draft, or waits for a type switch to be confirmed', async () => {
    vi.mocked(variables.getVariableTable).mockResolvedValue(table)
    await store().loadTable('env:e1')
    await expect(store().saveDraft('env:e1')).resolves.toEqual({ ok: true, value: true })
    expect(variables.saveVariableTable).not.toHaveBeenCalled()
    store().editDraft('env:e1', (rows) => rows.map((row) => ({ ...row, secret: !row.secret })))
    await expect(store().saveDraft('env:e1')).resolves.toEqual({ ok: true, value: false })
    expect(store().drafts['env:e1']?.confirming).not.toBeNull()
    store().cancelDraftSave('env:e1')
    expect(store().drafts['env:e1']?.confirming).toBeNull()
    vi.mocked(variables.saveVariableTable).mockResolvedValue(table)
    await store().saveDraft('env:e1')
    await expect(store().confirmDraftSave('env:e1')).resolves.toEqual({ ok: true, value: true })
    expect(isScopeDirty(store(), 'env:e1')).toBe(false)
  })

  it('reports a failed draft save on the draft', async () => {
    vi.mocked(variables.getVariableTable).mockResolvedValue(table)
    await store().loadTable('env:e1')
    store().editDraft('env:e1', (rows) => rows.map((row) => ({ ...row, description: 'edited' })))
    vi.mocked(variables.saveVariableTable).mockRejectedValue(new AppError('VAR-01400', 'bad name'))
    await expect(store().saveDraft('env:e1')).resolves.toMatchObject({ ok: false })
    expect(store().drafts['env:e1']).toMatchObject({ dirty: true, saving: false, failure: { code: 'VAR-01400' } })
  })

  it('loads, saves, resets and persists a variable table', async () => {
    vi.mocked(variables.getVariableTable).mockResolvedValue(table)
    const loading = store().loadTable('env:e1')
    expect(store().tables['env:e1']?.status).toBe('loading')
    await loading
    expect(store().tables['env:e1']).toEqual({ rows: [row], warnings: [], status: 'ready', error: null })
    const edits = [{ key: 'base_url', previousKey: 'base_url', initialValue: 'A', currentValue: null, secret: true, secretValue: null, enabled: true, description: '' }]
    vi.mocked(variables.saveVariableTable).mockResolvedValue({ ...table, warnings: ['git_history'] })
    await expect(store().saveTable('env:e1', edits, false)).resolves.toBeNull()
    expect(variables.saveVariableTable).toHaveBeenCalledWith('A', 'env:e1', edits, false)
    expect(store().tables['env:e1']?.warnings).toEqual(['git_history'])
    vi.mocked(variables.resetVariables).mockResolvedValue({ ...table, rows: [{ ...row, currentValue: 'A', hasCurrentOverride: false }] })
    await expect(store().resetValues('env:e1', [])).resolves.toBeNull()
    expect(variables.resetVariables).toHaveBeenCalledWith('A', 'env:e1', [])
    expect(store().tables['env:e1']?.rows[0]?.hasCurrentOverride).toBe(false)
    vi.mocked(variables.persistVariables).mockRejectedValue(new AppError('VAR-02422', 'secret'))
    await expect(store().persistValues('env:e1', ['token'])).resolves.toMatchObject({ code: 'VAR-02422' })
  })

  it('keeps a failed table load and ignores tables without a workspace', async () => {
    vi.mocked(variables.getVariableTable).mockRejectedValue(new AppError('VAR-06422', 'folders have no variables'))
    await store().loadTable('collection:f1')
    expect(store().tables['collection:f1']).toMatchObject({ status: 'ready', error: { code: 'VAR-06422' } })
    useWorkspacesStore.setState({ activeLocalId: null })
    await expect(store().saveTable('globals', [], false)).resolves.toBeNull()
    expect(variables.saveVariableTable).not.toHaveBeenCalled()
  })

  it('reveals and copies secrets and asks for the passphrase when the vault needs it', async () => {
    vi.mocked(vault.revealSecret).mockResolvedValueOnce('s3cr3t')
    await expect(store().revealSecret('globals', 'token')).resolves.toEqual({ ok: true, value: 's3cr3t' })
    expect(vault.revealSecret).toHaveBeenCalledWith('A', 'globals', 'token')
    vi.mocked(vault.revealSecret).mockRejectedValueOnce(new AppError('VLT-04409', 'locked'))
    await expect(store().revealSecret('globals', 'token')).resolves.toMatchObject({ ok: false })
    expect(store().passphrasePrompt).toBe('unlock')
    vi.mocked(vault.copySecret).mockResolvedValueOnce()
    await expect(store().copySecret('globals', 'token')).resolves.toBeNull()
    expect(vault.copySecret).toHaveBeenCalledWith('A', 'globals', 'token')
    vi.mocked(variables.saveVariableTable).mockRejectedValueOnce(new AppError('VLT-05409', 'choose a passphrase'))
    await expect(store().saveTable('globals', [], false)).resolves.toMatchObject({ code: 'VLT-05409' })
    expect(store().passphrasePrompt).toBe('setup')
  })

  it('prompts from the vault status unless the prompt was dismissed', () => {
    store().applyVaultStatus({ ...unlocked, lockedLocalIds: ['G'] })
    expect(store().passphrasePrompt).toBeNull()
    store().applyVaultStatus({ ...unlocked, lockedLocalIds: ['A'] })
    expect(store().passphrasePrompt).toBe('unlock')
    store().dismissPassphrase()
    store().applyVaultStatus({ ...unlocked, lockedLocalIds: ['A'] })
    expect(store().passphrasePrompt).toBeNull()
    store().applyVaultStatus({ ...unlocked, needsPassphraseSetup: true })
    expect(store().passphrasePrompt).toBe('setup')
    store().openPassphrase('unlock')
    expect(store().passphrasePrompt).toBe('unlock')
    useEnvironmentsStore.setState({ dismissedPrompts: [], passphrasePrompt: null })
  })

  it('sets up, unlocks and resets the vault, refreshing the loaded tables', async () => {
    useEnvironmentsStore.setState({ passphrasePrompt: 'setup', tables: { globals: { ...table, status: 'ready', error: null } } })
    vi.mocked(variables.getVariableTable).mockResolvedValue({ ...table, scope: 'globals' })
    vi.mocked(vault.setupPassphrase).mockResolvedValue(unlocked)
    await expect(store().setupPassphrase('long enough', 'long enough')).resolves.toBeNull()
    expect(vault.setupPassphrase).toHaveBeenCalledWith('long enough', 'long enough')
    expect(store()).toMatchObject({ passphrasePrompt: null, vault: unlocked })
    expect(variables.getVariableTable).toHaveBeenCalledWith('A', 'globals')
    vi.mocked(vault.unlockVault).mockRejectedValue(new AppError('VLT-02401', 'wrong'))
    await expect(store().unlockVault('nope')).resolves.toMatchObject({ code: 'VLT-02401' })
    vi.mocked(vault.resetVault).mockResolvedValue(unlocked)
    await expect(store().resetVault()).resolves.toBeNull()
    expect(vault.resetVault).toHaveBeenCalledWith('A')
  })

  it('loads the vault status and keeps the last one when it cannot be read', async () => {
    vi.mocked(vault.getVaultStatus).mockResolvedValueOnce({ ...unlocked, brokenLocalIds: ['A'] })
    await store().loadVault()
    expect(store().vault.brokenLocalIds).toEqual(['A'])
    vi.mocked(vault.getVaultStatus).mockRejectedValueOnce(new AppError('VLT-07500', 'x'))
    await store().loadVault()
    expect(store().vault.brokenLocalIds).toEqual(['A'])
    store().applyVaultStatus(unlocked)
  })
})
