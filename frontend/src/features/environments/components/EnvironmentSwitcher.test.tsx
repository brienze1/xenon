import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/environments', () => ({ listEnvironments: vi.fn(), setActiveEnvironment: vi.fn(), reorderEnvironment: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({ getVariableTable: vi.fn() }))
vi.mock('@/platform/services/vault', () => ({
  PASSPHRASE_SETUP_REQUIRED: 'VLT-05409',
  VAULT_LOCKED: 'VLT-04409',
  revealSecret: vi.fn(),
  copySecret: vi.fn(),
}))

const services = await import('@/platform/services/environments')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { EnvironmentSwitcher } = await import('./EnvironmentSwitcher')

const dev = { id: 'e1', name: 'dev', color: 'green', confirmBeforeSend: false, error: null } as const
const hml = { id: 'e2', name: 'hml', color: 'yellow', confirmBeforeSend: false, error: null } as const
const prod = { id: 'e3', name: 'prod', color: 'red', confirmBeforeSend: true, error: null } as const
const onOpenGlobals = vi.fn()
const onOpenEnvironment = vi.fn()
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
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onOpenEnvironment={onOpenEnvironment} onManage={onManage} />)
    await userEvent.click(screen.getByRole('button', { name: 'Environment: No environment' }))
    expect(screen.getByRole('menuitemradio', { name: 'No environment' })).toHaveAttribute('aria-checked', 'true')
    expect(screen.getAllByRole('menuitemradio').map((item) => item.textContent)).toEqual(['No environment', 'dev', 'hml', 'prod'])
    await userEvent.click(screen.getByRole('menuitemradio', { name: 'hml' }))
    expect(services.setActiveEnvironment).toHaveBeenCalledWith('A', 'e2')
    expect(screen.queryByRole('menu')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Environment: hml' })).toHaveTextContent('hml')
  })

  it('lists the environments in the order they were reordered into, not alphabetically', async () => {
    vi.mocked(services.reorderEnvironment).mockResolvedValue({ environments: [prod, dev, hml, broken], activeId: null })
    await act(async () => {
      await useEnvironmentsStore.getState().reorderEnvironment('e3', 0)
    })
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onOpenEnvironment={onOpenEnvironment} onManage={onManage} />)
    await userEvent.click(screen.getByRole('button', { name: 'Environment: No environment' }))
    expect(screen.getAllByRole('menuitemradio').map((item) => item.textContent)).toEqual(['No environment', 'prod', 'dev', 'hml'])
  })

  it('goes back to "No environment"', async () => {
    useEnvironmentsStore.setState({ activeId: 'e1' })
    vi.mocked(services.setActiveEnvironment).mockResolvedValue(null)
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onOpenEnvironment={onOpenEnvironment} onManage={onManage} />)
    await userEvent.click(screen.getByRole('button', { name: 'Environment: dev' }))
    expect(screen.getByRole('menuitemradio', { name: 'dev' })).toHaveAttribute('aria-checked', 'true')
    await userEvent.click(screen.getByRole('menuitemradio', { name: 'No environment' }))
    expect(services.setActiveEnvironment).toHaveBeenCalledWith('A', null)
    expect(screen.getByRole('button', { name: 'Environment: No environment' })).toBeInTheDocument()
  })

  it('lists a malformed environment with its error but does not let it be picked', async () => {
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onOpenEnvironment={onOpenEnvironment} onManage={onManage} />)
    await userEvent.click(screen.getByRole('button', { name: 'Environment: No environment' }))
    const item = screen.getByRole('menuitem', { name: /broken\.yaml/ })
    expect(item).toBeDisabled()
    expect(item).toHaveTextContent('broken.yaml cannot be read')
    await userEvent.click(item)
    expect(services.setActiveEnvironment).not.toHaveBeenCalled()
  })

  it('opens with keyboard focus when asked (Cmd/Ctrl+Shift+E) and picks by typing and Enter', async () => {
    vi.mocked(services.setActiveEnvironment).mockResolvedValue('e3')
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onOpenEnvironment={onOpenEnvironment} onManage={onManage} />)
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
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onOpenEnvironment={onOpenEnvironment} onManage={onManage} />)
    await userEvent.click(screen.getByRole('button', { name: 'Environment: No environment' }))
    await userEvent.click(screen.getByRole('menuitemradio', { name: 'dev' }))
    expect(screen.getByRole('alert')).toHaveTextContent(
      'Could not change the environment: the environment does not exist. Check that the workspace folder is writable, then pick the environment again.',
    )
    await userEvent.click(screen.getByRole('button', { name: 'Dismiss Error' }))
    expect(screen.queryByText(/Could not change the environment/)).not.toBeInTheDocument()
  })

  it('opens the globals tab and the sidebar environments from the menu', async () => {
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onOpenEnvironment={onOpenEnvironment} onManage={onManage} />)
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
    const { rerender } = render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onOpenEnvironment={onOpenEnvironment} onManage={onManage} />)
    expect(screen.getByRole('button', { name: 'Environment: No environment' })).toBeInTheDocument()
    act(() => useEnvironmentsStore.setState({ activeId: 'e3' }))
    rerender(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onOpenEnvironment={onOpenEnvironment} onManage={onManage} />)
    expect(screen.getByRole('button', { name: 'Environment: prod' })).toBeInTheDocument()
    act(() => useEnvironmentsStore.setState({ activeId: 'e1' }))
    expect(screen.getByRole('button', { name: 'Environment: dev' })).toBeInTheDocument()
  })

  it('opens the active environment in a tab from the eye', async () => {
    useEnvironmentsStore.setState({ activeId: 'e2' })
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onOpenEnvironment={onOpenEnvironment} onManage={onManage} />)
    const eye = screen.getByRole('button', { name: 'Open environment' })
    expect(eye).toHaveAttribute('title', 'Open environment')
    await userEvent.click(eye)
    expect(onOpenEnvironment).toHaveBeenCalledTimes(1)
    expect(onOpenEnvironment).toHaveBeenCalledWith('e2')
    expect(onOpenGlobals).not.toHaveBeenCalled()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('opens the globals from the eye when no environment is active', async () => {
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onOpenEnvironment={onOpenEnvironment} onManage={onManage} />)
    const eye = screen.getByRole('button', { name: 'Open globals' })
    expect(eye).toHaveAttribute('title', 'Open globals')
    await userEvent.click(eye)
    expect(onOpenGlobals).toHaveBeenCalledTimes(1)
    expect(onOpenEnvironment).not.toHaveBeenCalled()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('names what the eye opens after the active environment', () => {
    render(<EnvironmentSwitcher onOpenGlobals={onOpenGlobals} onOpenEnvironment={onOpenEnvironment} onManage={onManage} />)
    expect(screen.getByRole('button', { name: 'Open globals' })).toBeInTheDocument()
    act(() => useEnvironmentsStore.setState({ activeId: 'e1' }))
    expect(screen.getByRole('button', { name: 'Open environment' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Open globals' })).not.toBeInTheDocument()
  })
})
