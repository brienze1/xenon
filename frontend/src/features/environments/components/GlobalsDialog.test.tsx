import { act, render, screen } from '@testing-library/react'
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
const { GlobalsDialog } = await import('./GlobalsDialog')

const row = {
  key: 'token',
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

describe('GlobalsDialog', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    vi.mocked(variables.getVariableTable).mockImplementation(async (_localId, scope) => ({ scope, rows: [row], warnings: [] }))
  })

  afterEach(() => {
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('closes straight away without unsaved edits', async () => {
    render(<GlobalsDialog />)
    act(() => useEnvironmentsStore.getState().openEditor('globals'))
    await screen.findByRole('table', { name: 'Global variables' })
    await userEvent.click(screen.getByRole('button', { name: 'Close' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('asks before closing with unsaved edits', async () => {
    render(<GlobalsDialog />)
    act(() => useEnvironmentsStore.getState().openEditor('globals'))
    await userEvent.type(await screen.findByRole('textbox', { name: 'Description of token' }), '!')
    await userEvent.click(screen.getByRole('button', { name: 'Close' }))
    expect(screen.getByText('Discard unsaved variable changes?')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Keep Editing' })).toHaveFocus()
    await userEvent.click(screen.getByRole('button', { name: 'Keep Editing' }))
    expect(screen.getByRole('textbox', { name: 'Description of token' })).toHaveValue('!')
    await userEvent.click(screen.getByRole('button', { name: 'Close' }))
    await userEvent.click(screen.getByRole('button', { name: 'Discard' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(useEnvironmentsStore.getState()).toMatchObject({ editor: null, dirtyScopes: [] })
  })
})
