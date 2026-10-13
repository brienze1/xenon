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
vi.mock('@/platform/services/requests', () => ({ autoHeaders: vi.fn(async () => []), applyUrl: vi.fn(), applyParams: vi.fn() }))

const variables = await import('@/platform/services/variables')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { VariableCard } = await import('./VariableCard')

const TAB = 'tab-1'

function inspect(scopes: VariableInspection['scopes'], usages: VariableUsage[] = []) {
  useRequestBuilderStore.getState().ensureTab(TAB)
  useRequestBuilderStore.setState((state) => ({ tabs: { ...state.tabs, [TAB]: { ...state.tabs[TAB]!, inspection: { references: [], variables: usages, scopes } } } }))
}

function usage(patch: Partial<VariableUsage>): VariableUsage {
  return { name: 'creatorAdsUrl', status: 'resolved', scope: 'environment', displayValue: 'https://ads.dev', secret: false, usedIn: ['url'], ...patch }
}

describe('VariableCard', () => {
  beforeEach(() => {
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useEnvironmentsStore.setState({ environments: [{ id: 'e1', name: 'dev', color: 'none', confirmBeforeSend: false, error: null }], activeId: 'e1' })
    vi.mocked(variables.inspectVariables).mockResolvedValue({ references: [], variables: [], scopes: {} })
  })

  afterEach(() => {
    vi.clearAllMocks()
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('adds the variable with the typed value to the chosen scope', async () => {
    inspect({ environment: 'env:e1', collection: 'collection:c1', globals: 'globals' })
    vi.mocked(variables.addVariable).mockResolvedValue({} as never)
    const close = vi.fn()
    render(<VariableCard tabId={TAB} name="creatorAdsUrl" close={close} />)
    expect(screen.getByText('{{creatorAdsUrl}}')).toBeInTheDocument()
    await userEvent.type(screen.getByRole('textbox', { name: 'Value of creatorAdsUrl' }), 'https://ads.dev')
    const targets = screen.getByRole('list', { name: 'Add creatorAdsUrl to' })
    await userEvent.click(within(targets).getByRole('button', { name: /Environment: dev/ }))
    expect(variables.addVariable).toHaveBeenCalledWith('ws', 'env:e1', 'creatorAdsUrl', 'https://ads.dev', false)
    expect(variables.inspectVariables).toHaveBeenCalled()
  })

  it('shows a defined variable\'s current value and source, and saves an edit on Enter', async () => {
    inspect({ environment: 'env:e1' }, [usage({})])
    vi.mocked(variables.setCurrentValue).mockResolvedValue({} as never)
    render(<VariableCard tabId={TAB} name="creatorAdsUrl" close={vi.fn()} />)
    expect(screen.queryByText(/is not defined/)).not.toBeInTheDocument()
    expect(screen.getByText('Environment: dev')).toBeInTheDocument()
    const value = screen.getByRole('textbox', { name: 'Value of creatorAdsUrl' })
    expect(value).toHaveValue('https://ads.dev')
    await userEvent.clear(value)
    await userEvent.type(value, 'https://ads.prd{Enter}')
    expect(variables.setCurrentValue).toHaveBeenCalledWith('ws', 'env:e1', 'creatorAdsUrl', 'https://ads.prd')
  })

  it('keeps an edit when the card closes before Enter or blur', async () => {
    inspect({ environment: 'env:e1' }, [usage({})])
    vi.mocked(variables.setCurrentValue).mockResolvedValue({} as never)
    const { unmount } = render(<VariableCard tabId={TAB} name="creatorAdsUrl" close={vi.fn()} />)
    await userEvent.type(screen.getByRole('textbox', { name: 'Value of creatorAdsUrl' }), '/v2')
    unmount()
    expect(variables.setCurrentValue).toHaveBeenCalledTimes(1)
    expect(variables.setCurrentValue).toHaveBeenCalledWith('ws', 'env:e1', 'creatorAdsUrl', 'https://ads.dev/v2')
  })

  it('does not save an untouched value when the card closes', () => {
    inspect({ environment: 'env:e1' }, [usage({})])
    const { unmount } = render(<VariableCard tabId={TAB} name="creatorAdsUrl" close={vi.fn()} />)
    unmount()
    expect(variables.setCurrentValue).not.toHaveBeenCalled()
  })

  it('never shows a secret, and opens the request\'s variables in the side panel', async () => {
    inspect({ globals: 'globals' }, [usage({ scope: 'globals', secret: true, displayValue: '•••••' })])
    const close = vi.fn()
    render(<VariableCard tabId={TAB} name="creatorAdsUrl" close={close} />)
    expect(screen.getByRole('textbox', { name: 'Value of creatorAdsUrl' })).toHaveValue('')
    await userEvent.click(screen.getByRole('button', { name: 'Variables in request →' }))
    expect(useWorkspacesStore.getState().layout).toMatchObject({ rightPanelVisible: true, rightPanelView: 'variables' })
    expect(close).toHaveBeenCalled()
  })

  it('disables the scopes the request cannot use, like Postman', () => {
    inspect({ globals: 'globals' })
    render(<VariableCard tabId={TAB} name="x" close={vi.fn()} />)
    expect(screen.getByRole('button', { name: /No active environment/ })).toBeDisabled()
    expect(screen.getByRole('button', { name: /No parent collection/ })).toBeDisabled()
    expect(screen.getByRole('button', { name: /Globals/ })).toBeEnabled()
  })

  it('keeps the card open with the reason when adding fails', async () => {
    inspect({ globals: 'globals' })
    vi.mocked(variables.addVariable).mockRejectedValue(new AppError('VAR-04409', 'A variable with that name already exists.'))
    const close = vi.fn()
    render(<VariableCard tabId={TAB} name="x" close={close} />)
    await userEvent.click(screen.getByRole('button', { name: /Globals/ }))
    expect(await screen.findByRole('alert')).toHaveTextContent('A variable with that name already exists.')
    expect(close).not.toHaveBeenCalled()
  })
})
