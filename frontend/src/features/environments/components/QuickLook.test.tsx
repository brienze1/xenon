import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/platform/services/environments', () => ({ listEnvironments: vi.fn(), setActiveEnvironment: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({ getVariableTable: vi.fn() }))
vi.mock('@/platform/services/vault', () => ({
  PASSPHRASE_SETUP_REQUIRED: 'VLT-05409',
  VAULT_LOCKED: 'VLT-04409',
  revealSecret: vi.fn(),
  copySecret: vi.fn(),
}))

const variables = await import('@/platform/services/variables')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { QuickLook } = await import('./QuickLook')

const dev = { id: 'e1', name: 'dev', color: 'green', confirmBeforeSend: false, error: null } as const

function row(key: string) {
  return {
    key,
    initialValue: 'A',
    currentValue: 'A',
    hasCurrentOverride: false,
    secret: false,
    secretState: null,
    enabled: true,
    description: '',
    duplicate: false,
    surroundingSpaces: false,
  }
}

describe('QuickLook', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    useEnvironmentsStore.setState({ environments: [dev], activeId: 'e1', status: 'ready' })
    vi.mocked(variables.getVariableTable).mockImplementation(async (_localId, scope) => ({
      scope,
      rows: [row(scope === 'globals' ? 'token' : 'base_url')],
      warnings: [],
    }))
  })

  afterEach(() => {
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('shows the active environment and the globals, and asks before closing with unsaved edits', async () => {
    render(<QuickLook />)
    await userEvent.click(screen.getByRole('button', { name: 'Quick look at variables' }))
    expect(await screen.findByRole('table', { name: 'dev variables' })).toBeInTheDocument()
    await userEvent.type(await screen.findByRole('textbox', { name: 'Description of base_url' }), '!')
    await userEvent.click(screen.getByRole('button', { name: 'Close' }))
    expect(screen.getByText('Discard unsaved variable changes?')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Keep Editing' }))
    expect(screen.getByRole('dialog', { name: 'Quick Look' })).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Close' }))
    await userEvent.click(screen.getByRole('button', { name: 'Discard' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(useEnvironmentsStore.getState().dirtyScopes).toEqual([])
  })
})
