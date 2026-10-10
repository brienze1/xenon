import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { VariableInspection, VariableUsage } from '@/platform/mappers/inspection'

vi.mock('@/platform/services/environments', () => ({ listEnvironments: vi.fn(async () => ({ environments: [], activeId: null })), setActiveEnvironment: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({
  inspectVariables: vi.fn(),
  suggestVariables: vi.fn(),
  setCurrentValue: vi.fn(),
  addVariable: vi.fn(),
  getVariableTable: vi.fn(),
}))
vi.mock('@/platform/services/vault', () => ({
  PASSPHRASE_SETUP_REQUIRED: 'VLT-05409',
  VAULT_LOCKED: 'VLT-04409',
  revealSecret: vi.fn(),
  copySecret: vi.fn(),
}))
vi.mock('@/platform/services/requests', () => ({ autoHeaders: vi.fn(async () => []), applyUrl: vi.fn(), applyParams: vi.fn() }))

const variables = await import('@/platform/services/variables')
const vault = await import('@/platform/services/vault')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { VariablesPanel } = await import('./VariablesPanel')

const TAB = 'tab-1'

function usage(patch: Partial<VariableUsage>): VariableUsage {
  return { name: 'base_url', status: 'resolved', scope: 'environment', displayValue: 'https://dev', secret: false, usedIn: ['url'], ...patch }
}

const inspection: VariableInspection = {
  references: [],
  variables: [
    usage({}),
    usage({ name: 'api_key', secret: true, displayValue: '•••••', usedIn: ['headers.0.value'] }),
    usage({ name: 'user_id', status: 'unresolved', scope: 'none', displayValue: '' }),
    usage({ name: 'token', status: 'missing_secret', scope: 'globals', secret: true, displayValue: '•••••' }),
    usage({ name: 'password', status: 'locked_secret', scope: 'collection', secret: true }),
    usage({ name: '$guid', status: 'dynamic', scope: 'dynamic', displayValue: '', usedIn: [] }),
  ],
  scopes: { environment: 'env:e1', collection: 'collection:c1', globals: 'globals' },
}

function show(value: VariableInspection | null) {
  useRequestBuilderStore.getState().ensureTab(TAB)
  useRequestBuilderStore.setState((state) => ({ tabs: { ...state.tabs, [TAB]: { ...state.tabs[TAB]!, inspection: value } } }))
}

function item(name: string): HTMLElement {
  return screen.getByText(name, { selector: 'code' }).closest('li') as HTMLElement
}

describe('VariablesPanel', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useRequestBuilderStore.getState().reset()
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useEnvironmentsStore.setState({ environments: [{ id: 'e1', name: 'dev', color: 'none', confirmBeforeSend: false, error: null }], activeId: 'e1' })
    vi.mocked(variables.inspectVariables).mockResolvedValue(inspection)
  })

  afterEach(() => {
    useRequestBuilderStore.getState().reset()
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('explains how to use variables when the request has none', () => {
    show({ references: [], variables: [], scopes: {} })
    render(<VariablesPanel tabId={TAB} />)
    expect(screen.getByRole('region', { name: 'No variables' })).toBeInTheDocument()
  })

  it('lists each variable with status, source, value and the fields using it', () => {
    show(inspection)
    render(<VariablesPanel tabId={TAB} />)
    const base = item('base_url')
    expect(within(base).getByText('Resolved')).toBeInTheDocument()
    expect(within(base).getByText('Environment: dev')).toBeInTheDocument()
    expect(within(base).getByRole('textbox', { name: 'Value of base_url' })).toHaveValue('https://dev')
    expect(within(base).getByText('url')).toBeInTheDocument()
    expect(within(item('user_id')).getByText('Unresolved')).toBeInTheDocument()
    expect(within(item('user_id')).getByText('Not defined in any scope')).toBeInTheDocument()
    expect(within(item('token')).getByText('Missing secret')).toBeInTheDocument()
    expect(within(item('password')).getByText('Locked')).toBeInTheDocument()
    expect(within(item('$guid')).getByText('Generated at send')).toBeInTheDocument()
    expect(within(item('$guid')).queryByRole('textbox')).not.toBeInTheDocument()
  })

  it('edits a current value in its source scope and inspects again', async () => {
    show(inspection)
    vi.mocked(variables.setCurrentValue).mockResolvedValue({} as never)
    render(<VariablesPanel tabId={TAB} />)
    const base = item('base_url')
    const input = within(base).getByRole('textbox', { name: 'Value of base_url' })
    await userEvent.clear(input)
    await userEvent.type(input, 'https://local')
    await userEvent.click(within(base).getByRole('button', { name: 'Save' }))
    expect(variables.setCurrentValue).toHaveBeenCalledWith('ws', 'env:e1', 'base_url', 'https://local')
    expect(variables.inspectVariables).toHaveBeenCalled()
  })

  it('reports a failed edit', async () => {
    show(inspection)
    vi.mocked(variables.setCurrentValue).mockRejectedValue(new AppError('VAR-03404', 'gone'))
    render(<VariablesPanel tabId={TAB} />)
    const base = item('base_url')
    await userEvent.type(within(base).getByRole('textbox', { name: 'Value of base_url' }), 'x')
    await userEvent.click(within(base).getByRole('button', { name: 'Save' }))
    expect(within(base).getByRole('alert')).toHaveTextContent('gone')
  })

  it('masks secrets with reveal and copy', async () => {
    show(inspection)
    vi.mocked(vault.revealSecret).mockResolvedValue('s3cr3t')
    vi.mocked(vault.copySecret).mockResolvedValue()
    render(<VariablesPanel tabId={TAB} />)
    const secret = item('api_key')
    const input = within(secret).getByLabelText('Value of api_key')
    expect(input).toHaveAttribute('type', 'password')
    await userEvent.click(within(secret).getByRole('button', { name: 'Reveal api_key' }))
    expect(vault.revealSecret).toHaveBeenCalledWith('ws', 'env:e1', 'api_key')
    expect(input).toHaveAttribute('type', 'text')
    expect(input).toHaveValue('s3cr3t')
    await userEvent.click(within(secret).getByRole('button', { name: 'Hide api_key' }))
    expect(input).toHaveAttribute('type', 'password')
    await userEvent.click(within(secret).getByRole('button', { name: 'Copy api_key' }))
    expect(vault.copySecret).toHaveBeenCalledWith('ws', 'env:e1', 'api_key')
  })

  it('adds an unresolved variable to the chosen scope', async () => {
    show(inspection)
    vi.mocked(variables.addVariable).mockResolvedValue({} as never)
    render(<VariablesPanel tabId={TAB} />)
    const unresolved = item('user_id')
    const scope = within(unresolved).getByRole('combobox', { name: 'Add user_id to' })
    expect(within(scope).getAllByRole('option').map((option) => option.textContent)).toEqual(['Add to environment', 'Add to collection', 'Add to globals'])
    await userEvent.selectOptions(scope, 'collection')
    await userEvent.type(within(unresolved).getByRole('textbox', { name: 'Value of user_id' }), '42')
    await userEvent.click(within(unresolved).getByRole('button', { name: 'Add' }))
    expect(variables.addVariable).toHaveBeenCalledWith('ws', 'collection:c1', 'user_id', '42', false)
  })

  it('stores a missing secret and offers to unlock a locked one', async () => {
    show(inspection)
    vi.mocked(variables.setCurrentValue).mockResolvedValue({} as never)
    render(<VariablesPanel tabId={TAB} />)
    const missing = item('token')
    await userEvent.type(within(missing).getByLabelText('Value of token'), 'tok')
    await userEvent.click(within(missing).getByRole('button', { name: 'Save' }))
    expect(variables.setCurrentValue).toHaveBeenCalledWith('ws', 'globals', 'token', 'tok')
    await userEvent.click(within(item('password')).getByRole('button', { name: 'Unlock Vault' }))
    expect(useEnvironmentsStore.getState().passphrasePrompt).toBe('unlock')
  })

  it('moves a selected literal into a secret variable and references it', async () => {
    show(inspection)
    useRequestBuilderStore.getState().updateHeaders(TAB, [{ id: 'h1', key: 'Authorization', value: 'Bearer abc123', description: '', enabled: true }])
    useRequestBuilderStore.getState().setSelection(TAB, { field: 'headers.0.value', from: 7, to: 13, text: 'abc123' })
    vi.mocked(variables.addVariable).mockResolvedValue({} as never)
    render(<VariablesPanel tabId={TAB} />)
    const form = screen.getByRole('form', { name: 'Move to secret variable' })
    expect(form).toHaveTextContent('headers.0.value')
    const move = within(form).getByRole('button', { name: 'Move to Secret Variable' })
    expect(move).toBeDisabled()
    await userEvent.type(within(form).getByRole('textbox', { name: 'Secret variable name' }), 'api_token')
    await userEvent.selectOptions(within(form).getByRole('combobox', { name: 'Secret variable scope' }), 'globals')
    await userEvent.click(move)
    expect(variables.addVariable).toHaveBeenCalledWith('ws', 'globals', 'api_token', 'abc123', true)
    expect(useRequestBuilderStore.getState().tabs[TAB]!.draft.headers[0]!.value).toBe('Bearer {{api_token}}')
    expect(screen.queryByRole('form', { name: 'Move to secret variable' })).not.toBeInTheDocument()
  })

  it('keeps the selection when the secret cannot be created', async () => {
    show(inspection)
    useRequestBuilderStore.getState().setUrl(TAB, 'https://x/abcdefghijklmnopqrstuvwxyz0123')
    useRequestBuilderStore.getState().setSelection(TAB, { field: 'url', from: 10, to: 40, text: 'abcdefghijklmnopqrstuvwxyz0123' })
    vi.mocked(variables.addVariable).mockRejectedValue(new AppError('VAR-04409', 'exists'))
    render(<VariablesPanel tabId={TAB} />)
    const form = screen.getByRole('form', { name: 'Move to secret variable' })
    expect(form).toHaveTextContent('abcdefghijklmnopqrstuvwx…')
    await userEvent.type(within(form).getByRole('textbox', { name: 'Secret variable name' }), 'token')
    await userEvent.click(within(form).getByRole('button', { name: 'Move to Secret Variable' }))
    expect(within(form).getByRole('alert')).toHaveTextContent('exists')
    expect(useRequestBuilderStore.getState().tabs[TAB]!.draft.url).toBe('https://x/abcdefghijklmnopqrstuvwxyz0123')
  })
})
