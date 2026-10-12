import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/environments', () => ({ listEnvironments: vi.fn(), setActiveEnvironment: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({ getVariableTable: vi.fn() }))
vi.mock('@/platform/services/vault', () => ({
  PASSPHRASE_SETUP_REQUIRED: 'VLT-05409',
  VAULT_LOCKED: 'VLT-04409',
  revealSecret: vi.fn(),
  copySecret: vi.fn(),
}))

const services = await import('@/platform/services/environments')
const variables = await import('@/platform/services/variables')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { EnvironmentSwitcher } = await import('./EnvironmentSwitcher')

const dev = { id: 'e1', name: 'dev', color: 'green', confirmBeforeSend: false, error: null } as const
const hml = { id: 'e2', name: 'hml', color: 'yellow', confirmBeforeSend: false, error: null } as const
const prod = { id: 'e3', name: 'prod', color: 'red', confirmBeforeSend: true, error: null } as const
const onOpenGlobals = vi.fn()
const onManage = vi.fn()
const broken = { id: 'error:broken.yaml', name: 'broken.yaml', color: 'none', confirmBeforeSend: false, error: 'broken.yaml cannot be read' } as const

describe('EnvironmentSwitcher', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    useEnvironmentsStore.setState({ environments: [dev, hml, prod, broken], activeId: null, status: 'ready' })
  })

  afterEach(() => {
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('shows "No environment" until one is picked, then the active one', async () => {
    vi.mocked(services.setActiveEnvironment).mockResolvedValue('e2')
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onManage={onManage} />)
    await userEvent.click(screen.getByRole('button', { name: 'Environment: No environment' }))
    expect(screen.getByRole('menuitemradio', { name: 'No environment' })).toHaveAttribute('aria-checked', 'true')
    expect(screen.getAllByRole('menuitemradio').map((item) => item.textContent)).toEqual(['No environment', 'dev', 'hml', 'prod'])
    await userEvent.click(screen.getByRole('menuitemradio', { name: 'hml' }))
    expect(services.setActiveEnvironment).toHaveBeenCalledWith('A', 'e2')
    expect(screen.queryByRole('menu')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Environment: hml' })).toHaveTextContent('hml')
  })

  it('goes back to "No environment"', async () => {
    useEnvironmentsStore.setState({ activeId: 'e1' })
    vi.mocked(services.setActiveEnvironment).mockResolvedValue(null)
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onManage={onManage} />)
    await userEvent.click(screen.getByRole('button', { name: 'Environment: dev' }))
    expect(screen.getByRole('menuitemradio', { name: 'dev' })).toHaveAttribute('aria-checked', 'true')
    await userEvent.click(screen.getByRole('menuitemradio', { name: 'No environment' }))
    expect(services.setActiveEnvironment).toHaveBeenCalledWith('A', null)
    expect(screen.getByRole('button', { name: 'Environment: No environment' })).toBeInTheDocument()
  })

  it('lists a malformed environment with its error but does not let it be picked', async () => {
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onManage={onManage} />)
    await userEvent.click(screen.getByRole('button', { name: 'Environment: No environment' }))
    const item = screen.getByRole('menuitem', { name: /broken\.yaml/ })
    expect(item).toBeDisabled()
    expect(item).toHaveTextContent('broken.yaml cannot be read')
    await userEvent.click(item)
    expect(services.setActiveEnvironment).not.toHaveBeenCalled()
  })

  it('opens with keyboard focus when asked (Cmd/Ctrl+Shift+E) and picks by typing and Enter', async () => {
    vi.mocked(services.setActiveEnvironment).mockResolvedValue('e3')
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onManage={onManage} />)
    act(() => useEnvironmentsStore.getState().openSwitcher())
    expect(screen.getByRole('menu', { name: 'Environments' })).toBeInTheDocument()
    expect(screen.getByRole('menuitemradio', { name: 'No environment' })).toHaveFocus()
    await userEvent.keyboard('pr')
    expect(screen.getByRole('menuitemradio', { name: 'prod' })).toHaveFocus()
    await userEvent.keyboard('{Enter}')
    expect(services.setActiveEnvironment).toHaveBeenCalledWith('A', 'e3')
    expect(useEnvironmentsStore.getState().switcherOpen).toBe(false)
  })

  it('shows a failed switch and lets it be dismissed', async () => {
    vi.mocked(services.setActiveEnvironment).mockRejectedValue(new AppError('ENV-01404', 'the environment does not exist'))
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onManage={onManage} />)
    await userEvent.click(screen.getByRole('button', { name: 'Environment: No environment' }))
    await userEvent.click(screen.getByRole('menuitemradio', { name: 'dev' }))
    expect(screen.getByRole('alert')).toHaveTextContent(
      'Could not change the environment: the environment does not exist. Check that the workspace folder is writable, then pick the environment again.',
    )
    await userEvent.click(screen.getByRole('button', { name: 'Dismiss Error' }))
    expect(screen.queryByText(/Could not change the environment/)).not.toBeInTheDocument()
  })

  it('opens the globals tab and the sidebar environments from the menu', async () => {
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onManage={onManage} />)
    await userEvent.click(screen.getByRole('button', { name: 'Environment: No environment' }))
    await userEvent.click(screen.getByRole('menuitem', { name: 'Globals' }))
    expect(screen.queryByRole('menu')).not.toBeInTheDocument()
    expect(onOpenGlobals).toHaveBeenCalledTimes(1)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Environment: No environment' }))
    await userEvent.click(screen.getByRole('menuitem', { name: 'Manage Environments…' }))
    expect(onManage).toHaveBeenCalledTimes(1)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('names the active environment on the trigger, none without one', async () => {
    const { rerender } = render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onManage={onManage} />)
    expect(screen.getByRole('button', { name: 'Environment: No environment' })).toBeInTheDocument()
    act(() => useEnvironmentsStore.setState({ activeId: 'e3' }))
    rerender(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onManage={onManage} />)
    expect(screen.getByRole('button', { name: 'Environment: prod' })).toBeInTheDocument()
    act(() => useEnvironmentsStore.setState({ activeId: 'e1' }))
    expect(screen.getByRole('button', { name: 'Environment: dev' })).toBeInTheDocument()
  })

  it('opens a quick look at the active environment and the globals', async () => {
    vi.mocked(variables.getVariableTable).mockImplementation(async (_localId, scope) => ({ scope, rows: [], warnings: [] }))
    useEnvironmentsStore.setState({ activeId: 'e3' })
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onManage={onManage} />)
    await userEvent.click(screen.getByRole('button', { name: 'Quick look at variables' }))
    const dialog = screen.getByRole('dialog', { name: 'Quick Look' })
    expect(within(dialog).getByRole('region', { name: 'Active environment variables' })).toHaveTextContent('prod')
    expect(within(dialog).getByRole('region', { name: 'Global variables' })).toBeInTheDocument()
    expect(variables.getVariableTable).toHaveBeenCalledWith('A', 'env:e3')
    expect(variables.getVariableTable).toHaveBeenCalledWith('A', 'globals')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Close' }))
    expect(screen.queryByRole('dialog', { name: 'Quick Look' })).not.toBeInTheDocument()
  })

  it('tells that no environment is active in the quick look', async () => {
    vi.mocked(variables.getVariableTable).mockImplementation(async (_localId, scope) => ({ scope, rows: [], warnings: [] }))
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onManage={onManage} />)
    await userEvent.click(screen.getByRole('button', { name: 'Quick look at variables' }))
    expect(screen.getByText(/No environment is active/)).toBeInTheDocument()
  })
})
