import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { VariableRow, VariableTable as VariableTableData } from '@/platform/mappers/variable'

vi.mock('@/platform/services/environments', () => ({ listEnvironments: vi.fn(), setActiveEnvironment: vi.fn() }))
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
  revealSecret: vi.fn(),
  copySecret: vi.fn(),
}))

const variables = await import('@/platform/services/variables')
const vault = await import('@/platform/services/vault')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { VariableTable } = await import('./VariableTable')

const SCOPE = 'env:e1'
const LONG_URL = `https://${'very-long-subdomain.'.repeat(8)}example.test/api`

function plain(patch: Partial<VariableRow> = {}): VariableRow {
  return {
    key: 'base_url',
    initialValue: LONG_URL,
    currentValue: LONG_URL,
    hasCurrentOverride: false,
    secret: false,
    secretState: null,
    enabled: true,
    description: 'API root',
    duplicate: false,
    surroundingSpaces: false,
    ...patch,
  }
}

function secret(key: string, secretState: VariableRow['secretState']): VariableRow {
  return plain({ key, initialValue: '', currentValue: '', secret: true, secretState, description: '' })
}

function tableOf(rows: VariableRow[], warnings: string[] = []): VariableTableData {
  return { scope: SCOPE, rows, warnings }
}

async function renderTable(rows: VariableRow[]) {
  vi.mocked(variables.getVariableTable).mockResolvedValue(tableOf(rows))
  render(<VariableTable scope={SCOPE} label="dev variables" />)
  return screen.findByRole('table', { name: 'dev variables' })
}

const user = userEvent

describe('VariableTable', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
  })

  afterEach(() => {
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('shows every column and the inherited current value with distinct styling', async () => {
    const table = await renderTable([plain(), plain({ key: 'user', initialValue: 'ana', currentValue: 'bob', hasCurrentOverride: true })])
    expect(variables.getVariableTable).toHaveBeenCalledWith('A', SCOPE)
    expect(within(table).getAllByRole('columnheader').map((header) => header.textContent)).toEqual([
      'Name',
      'Initial Value',
      'Current Value',
      'Type',
      'Enabled',
      'Description',
      'Actions',
    ])
    const inherited = screen.getByRole('textbox', { name: 'Current value of base_url' })
    expect(inherited).toHaveValue(LONG_URL)
    expect(inherited).toHaveAccessibleDescription('Inherited from the initial value')
    const overridden = screen.getByRole('textbox', { name: 'Current value of user' })
    expect(overridden).toHaveValue('bob')
    expect(overridden).not.toHaveAccessibleDescription('Inherited from the initial value')
    expect(screen.getByRole('combobox', { name: 'Type of user' })).toHaveValue('plain')
    expect(screen.getByRole('checkbox', { name: 'Enable user' })).toBeChecked()
  })

  it('keeps long values whole on hover', async () => {
    await renderTable([plain()])
    expect(screen.getByRole('textbox', { name: 'Initial value of base_url' })).toHaveAttribute('title', LONG_URL)
    expect(screen.getByRole('textbox', { name: 'Current value of base_url' })).toHaveAttribute('title', LONG_URL)
  })

  it('flags duplicate names and names with surrounding spaces, also while editing', async () => {
    await renderTable([plain({ key: 'a', duplicate: true }), plain({ key: ' b', surroundingSpaces: true }), plain({ key: 'c' })])
    const rows = screen.getAllByRole('row').slice(1)
    expect(within(rows[0]!).getByText('Duplicate')).toBeInTheDocument()
    expect(within(rows[1]!).getByText('Spaces')).toBeInTheDocument()
    expect(within(rows[2]!).queryByText('Duplicate')).not.toBeInTheDocument()
    const name = screen.getByRole('textbox', { name: 'Name of c' })
    await user.clear(name)
    await user.type(name, 'a')
    expect(within(rows[2]!).getByText('Duplicate')).toBeInTheDocument()
  })

  it('resets and persists a row, or all of them, and hides both for secret rows', async () => {
    await renderTable([plain({ currentValue: 'B', hasCurrentOverride: true }), secret('token', 'set')])
    vi.mocked(variables.resetVariables).mockResolvedValue(tableOf([plain()]))
    await user.click(screen.getByRole('button', { name: 'Reset base_url' }))
    expect(variables.resetVariables).toHaveBeenCalledWith('A', SCOPE, ['base_url'])
    expect(await screen.findByRole('button', { name: 'Reset base_url' })).toBeDisabled()
    expect(screen.getByRole('textbox', { name: 'Current value of base_url' })).toHaveAccessibleDescription('Inherited from the initial value')
    expect(screen.queryByRole('button', { name: 'Persist token' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Reset token' })).not.toBeInTheDocument()
    vi.mocked(variables.getVariableTable).mockResolvedValue(tableOf([plain({ currentValue: 'B', hasCurrentOverride: true })]))
    await useEnvironmentsStore.getState().loadTable(SCOPE)
    vi.mocked(variables.persistVariables).mockResolvedValue(tableOf([plain({ initialValue: 'B', currentValue: 'B' })]))
    await user.click(await screen.findByRole('button', { name: 'Persist base_url' }))
    expect(variables.persistVariables).toHaveBeenCalledWith('A', SCOPE, ['base_url'])
    expect(await screen.findByRole('textbox', { name: 'Initial value of base_url' })).toHaveValue('B')
    vi.mocked(variables.getVariableTable).mockResolvedValue(tableOf([plain({ currentValue: 'B', hasCurrentOverride: true })]))
    await useEnvironmentsStore.getState().loadTable(SCOPE)
    vi.mocked(variables.resetVariables).mockResolvedValue(tableOf([plain()]))
    await user.click(await screen.findByRole('button', { name: 'Reset All' }))
    expect(screen.getByRole('status')).toHaveTextContent('Reset every current value to its initial value?')
    await user.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(variables.resetVariables).toHaveBeenCalledTimes(1)
    await user.click(screen.getByRole('button', { name: 'Reset All' }))
    await user.click(screen.getByRole('button', { name: 'Reset All Values' }))
    expect(variables.resetVariables).toHaveBeenLastCalledWith('A', SCOPE, [])
    vi.mocked(variables.getVariableTable).mockResolvedValue(tableOf([plain({ currentValue: 'B', hasCurrentOverride: true })]))
    await useEnvironmentsStore.getState().loadTable(SCOPE)
    vi.mocked(variables.persistVariables).mockRejectedValue(new AppError('VAR-02422', 'secret rows have no initial value'))
    await user.click(await screen.findByRole('button', { name: 'Persist All' }))
    expect(variables.persistVariables).toHaveBeenCalledTimes(1)
    await user.click(screen.getByRole('button', { name: 'Persist All Values' }))
    expect(variables.persistVariables).toHaveBeenLastCalledWith('A', SCOPE, [])
    expect(await screen.findByRole('alert')).toHaveTextContent('secret rows have no initial value')
  })

  it('waits with Reset and Persist while there are unsaved edits', async () => {
    await renderTable([plain({ currentValue: 'B', hasCurrentOverride: true })])
    await user.type(screen.getByRole('textbox', { name: 'Description of base_url' }), '!')
    expect(screen.getByRole('button', { name: 'Reset base_url' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Persist All' })).toBeDisabled()
    await user.click(screen.getByRole('button', { name: 'Discard Changes' }))
    expect(screen.getByRole('button', { name: 'Reset base_url' })).toBeEnabled()
    expect(screen.getByRole('textbox', { name: 'Description of base_url' })).toHaveValue('API root')
  })

  it('flags its scope as having unsaved edits until they are saved or discarded', async () => {
    await renderTable([plain()])
    expect(useEnvironmentsStore.getState().dirtyScopes).toEqual([])
    await user.type(screen.getByRole('textbox', { name: 'Description of base_url' }), '!')
    expect(useEnvironmentsStore.getState().dirtyScopes).toEqual([SCOPE])
    await user.click(screen.getByRole('button', { name: 'Discard Changes' }))
    expect(useEnvironmentsStore.getState().dirtyScopes).toEqual([])
    await user.type(screen.getByRole('textbox', { name: 'Description of base_url' }), '!')
    vi.mocked(variables.saveVariableTable).mockResolvedValue(tableOf([plain({ description: 'API root!' })]))
    await user.click(screen.getByRole('button', { name: 'Save Variables' }))
    expect(await screen.findByRole('textbox', { name: 'Description of base_url' })).toHaveValue('API root!')
    expect(useEnvironmentsStore.getState().dirtyScopes).toEqual([])
  })

  it('clears its unsaved-edits flag when it closes', async () => {
    vi.mocked(variables.getVariableTable).mockResolvedValue(tableOf([plain()]))
    const { unmount } = render(<VariableTable scope={SCOPE} label="dev variables" />)
    await user.type(await screen.findByRole('textbox', { name: 'Description of base_url' }), '!')
    expect(useEnvironmentsStore.getState().dirtyScopes).toEqual([SCOPE])
    unmount()
    expect(useEnvironmentsStore.getState().dirtyScopes).toEqual([])
  })

  it('adds, renames and removes rows, then saves the whole table', async () => {
    await renderTable([plain({ key: 'url', initialValue: 'A', currentValue: 'A' }), plain({ key: 'gone' })])
    const name = screen.getByRole('textbox', { name: 'Name of url' })
    await user.clear(name)
    await user.type(name, 'base_url')
    await user.type(screen.getByRole('textbox', { name: 'Current value of base_url' }), 'B')
    await user.click(screen.getByRole('button', { name: 'Remove gone' }))
    await user.click(screen.getByRole('button', { name: 'Add Variable' }))
    await user.type(screen.getByRole('textbox', { name: 'Name of row 2' }), 'user')
    await user.type(screen.getByRole('textbox', { name: 'Initial value of user' }), 'ana')
    await user.click(screen.getByRole('checkbox', { name: 'Enable user' }))
    vi.mocked(variables.saveVariableTable).mockResolvedValue(tableOf([plain({ key: 'base_url', initialValue: 'A', currentValue: 'AB', hasCurrentOverride: true })]))
    await user.click(screen.getByRole('button', { name: 'Save Variables' }))
    expect(variables.saveVariableTable).toHaveBeenCalledWith(
      'A',
      SCOPE,
      [
        { key: 'base_url', previousKey: 'url', initialValue: 'A', currentValue: 'AB', secret: false, secretValue: null, enabled: true, description: 'API root' },
        { key: 'user', previousKey: null, initialValue: 'ana', currentValue: null, secret: false, secretValue: null, enabled: false, description: '' },
      ],
      false,
    )
    expect(await screen.findByRole('button', { name: 'Save Variables' })).toBeEnabled()
    expect(screen.getAllByRole('row')).toHaveLength(2)
    await user.click(screen.getByRole('button', { name: 'Save Variables' }))
    expect(variables.saveVariableTable).toHaveBeenCalledTimes(1)
  })

  it('shows a save error inline and keeps the edits', async () => {
    await renderTable([plain()])
    await user.type(screen.getByRole('textbox', { name: 'Name of base_url' }), '{{')
    vi.mocked(variables.saveVariableTable).mockRejectedValue(new AppError('VAR-00400', 'the variable name is invalid'))
    await user.click(screen.getByRole('button', { name: 'Save Variables' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('the variable name is invalid')
    expect(screen.getByText('Invalid name')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Save Variables' })).toBeEnabled()
  })

  it('moves focus to the first invalid name after a failed save', async () => {
    await renderTable([plain({ key: 'ok' }), plain({ key: 'other' })])
    await user.type(screen.getByRole('textbox', { name: 'Name of other' }), '{{')
    vi.mocked(variables.saveVariableTable).mockRejectedValue(new AppError('VAR-00400', 'the variable name is invalid'))
    await user.click(screen.getByRole('button', { name: 'Save Variables' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('the variable name is invalid')
    expect(screen.getByRole('textbox', { name: 'Name of other{' })).toHaveFocus()
  })

  it('masks secret values, reveals one row on demand and hides it again', async () => {
    await renderTable([secret('api_key', 'set'), secret('token', 'set')])
    const value = screen.getByLabelText('Value of api_key')
    expect(value).toHaveAttribute('type', 'password')
    expect(value).toHaveValue('')
    expect(value).toHaveAttribute('placeholder', '•••••…')
    expect(screen.getAllByText('Not shared')).toHaveLength(2)
    vi.mocked(vault.revealSecret).mockResolvedValue('s3cr3t')
    await user.click(screen.getByRole('button', { name: 'Reveal api_key' }))
    expect(vault.revealSecret).toHaveBeenCalledWith('A', SCOPE, 'api_key')
    expect(value).toHaveAttribute('type', 'text')
    expect(value).toHaveValue('s3cr3t')
    expect(screen.getByLabelText('Value of token')).toHaveValue('')
    await user.click(screen.getByRole('button', { name: 'Hide api_key' }))
    expect(value).toHaveAttribute('type', 'password')
    expect(value).toHaveValue('')
  })

  it('copies a secret through Go and shows a failure next to it', async () => {
    await renderTable([secret('api_key', 'set')])
    vi.mocked(vault.copySecret).mockResolvedValueOnce()
    await user.click(screen.getByRole('button', { name: 'Copy api_key' }))
    expect(vault.copySecret).toHaveBeenCalledWith('A', SCOPE, 'api_key')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    vi.mocked(vault.revealSecret).mockRejectedValueOnce(new AppError('VLT-04409', 'the vault is locked'))
    await user.click(screen.getByRole('button', { name: 'Reveal api_key' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('the vault is locked')
    expect(useEnvironmentsStore.getState().passphrasePrompt).toBe('unlock')
  })

  it('marks missing and locked secrets and stores a typed value on save', async () => {
    await renderTable([secret('api_key', 'missing'), secret('token', 'locked')])
    expect(screen.getByText('Missing')).toBeInTheDocument()
    expect(screen.getByText('Locked')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Copy api_key' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Reveal token' })).toBeDisabled()
    await user.type(screen.getByLabelText('Value of api_key'), 'n3w')
    await user.click(screen.getByRole('button', { name: 'Reveal api_key' }))
    expect(vault.revealSecret).not.toHaveBeenCalled()
    expect(screen.getByLabelText('Value of api_key')).toHaveValue('n3w')
    vi.mocked(variables.saveVariableTable).mockResolvedValue(tableOf([secret('api_key', 'set'), secret('token', 'locked')]))
    await user.click(screen.getByRole('button', { name: 'Save Variables' }))
    expect(variables.saveVariableTable).toHaveBeenCalledWith(
      'A',
      SCOPE,
      [
        expect.objectContaining({ key: 'api_key', secret: true, secretValue: 'n3w' }),
        expect.objectContaining({ key: 'token', secret: true, secretValue: null }),
      ],
      false,
    )
  })

  it('warns about git history before making a row secret and after the save says so', async () => {
    await renderTable([plain({ key: 'password', initialValue: 'hunter2', currentValue: 'hunter2' })])
    await user.selectOptions(screen.getByRole('combobox', { name: 'Type of password' }), 'secret')
    await user.click(screen.getByRole('button', { name: 'Save Variables' }))
    expect(variables.saveVariableTable).not.toHaveBeenCalled()
    expect(screen.getByRole('status')).toHaveTextContent('may still be in git history')
    await user.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(screen.queryByText(/may still be in git history/)).not.toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Save Variables' }))
    vi.mocked(variables.saveVariableTable).mockResolvedValue(tableOf([secret('password', 'set')], ['git_history']))
    await user.click(screen.getByRole('button', { name: 'Make Secret' }))
    expect(variables.saveVariableTable).toHaveBeenCalledWith('A', SCOPE, [expect.objectContaining({ key: 'password', secret: true })], false)
    expect(await screen.findByText(/may still be in git history\. Rotate them/)).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Dismiss' }))
    expect(screen.queryByText(/Rotate them/)).not.toBeInTheDocument()
  })

  it('asks before writing a secret into the workspace files', async () => {
    await renderTable([secret('token', 'set')])
    await user.selectOptions(screen.getByRole('combobox', { name: 'Type of token' }), 'plain')
    await user.click(screen.getByRole('button', { name: 'Save Variables' }))
    expect(screen.getByText(/will be written to the workspace files/)).toBeInTheDocument()
    vi.mocked(variables.saveVariableTable).mockResolvedValue(tableOf([plain({ key: 'token', initialValue: 'abc', currentValue: 'abc' })]))
    await user.click(screen.getByRole('button', { name: 'Write Values to Files' }))
    expect(variables.saveVariableTable).toHaveBeenCalledWith('A', SCOPE, [expect.objectContaining({ key: 'token', secret: false })], true)
    expect(await screen.findByRole('textbox', { name: 'Initial value of token' })).toHaveValue('abc')
  })

  it('asks for the confirmation when Go requires it on save', async () => {
    await renderTable([plain()])
    await user.type(screen.getByRole('textbox', { name: 'Description of base_url' }), '!')
    vi.mocked(variables.saveVariableTable).mockRejectedValueOnce(new AppError('VAR-01409', 'confirm the reveal'))
    await user.click(screen.getByRole('button', { name: 'Save Variables' }))
    expect(await screen.findByText(/will be written to the workspace files/)).toBeInTheDocument()
    vi.mocked(variables.saveVariableTable).mockResolvedValue(tableOf([plain({ description: 'API root!' })]))
    await user.click(screen.getByRole('button', { name: 'Write Values to Files' }))
    expect(variables.saveVariableTable).toHaveBeenLastCalledWith('A', SCOPE, expect.any(Array), true)
  })

  it('offers to retry a table that cannot be loaded', async () => {
    vi.mocked(variables.getVariableTable).mockRejectedValueOnce(new AppError('VAR-06422', 'folders have no variables'))
    render(<VariableTable scope={SCOPE} label="dev variables" />)
    expect(await screen.findByRole('alert')).toHaveTextContent('folders have no variables')
    vi.mocked(variables.getVariableTable).mockResolvedValue(tableOf([]))
    await user.click(screen.getByRole('button', { name: 'Try Again' }))
    expect(await screen.findByText(/No variables yet/)).toBeInTheDocument()
  })
})
