import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyDraft, type SavedRequest } from '@/platform/mappers/request'

vi.mock('@/platform/services/requests', async (importOriginal) => {
  const { confirmationOf, expiredTokenOwnerOf, missingSecretsOf } = await importOriginal<typeof import('@/platform/services/requests')>()
  return { confirmationOf, expiredTokenOwnerOf, missingSecretsOf, getRequest: vi.fn(), autoHeaders: vi.fn(), sendRequest: vi.fn() }
})
vi.mock('@/platform/services/aliases', () => ({ createAlias: vi.fn(), saveAlias: vi.fn(), deleteAlias: vi.fn(), setAliasEnvironments: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({ inspectVariables: vi.fn(async () => ({ references: [], unresolved: [] })), suggestVariables: vi.fn() }))

const requests = await import('@/platform/services/requests')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { UnavailableAliasBanner } = await import('./UnavailableAliasBanner')
const { RequestBar } = await import('./RequestBar')

const TAB = 'tab-1'
const environments = [
  { id: 'hml', name: 'HML', color: 'green' as const, confirmBeforeSend: false, error: null },
  { id: 'dev', name: 'DEV', color: 'cyan' as const, confirmBeforeSend: false, error: null },
  { id: 'prod', name: 'Prod', color: 'red' as const, confirmBeforeSend: true, error: null },
]

function savedWith(environmentIds: string[]): SavedRequest {
  const draft = { ...emptyDraft(), id: 'r1', name: 'Login', url: 'https://x/v1' }
  return { draft, revision: 'rev-1', aliases: [{ id: 'a1', name: 'mx-test-user', environmentIds, draft: { ...draft, name: 'mx-test-user' } }] }
}

async function show(environmentIds: string[], aliasId: string | null = 'a1') {
  vi.mocked(requests.getRequest).mockResolvedValue(savedWith(environmentIds))
  await useRequestBuilderStore.getState().openSaved(TAB, 'r1', aliasId)
  render(
    <>
      <RequestBar tabId={TAB} />
      <UnavailableAliasBanner tabId={TAB} />
    </>,
  )
}

describe('UnavailableAliasBanner and the Send button', () => {
  const setActive = vi.fn(async (id: string | null) => useEnvironmentsStore.setState({ activeId: id }))

  beforeEach(() => {
    vi.resetAllMocks()
    useRequestBuilderStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useEnvironmentsStore.setState({ environments, activeId: 'prod', setActive })
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
  })

  afterEach(() => useEnvironmentsStore.setState({ environments: [], activeId: null }))

  it('disables Send and offers a switch per environment the alias is for, then clears once one is active', async () => {
    setActive.mockImplementation(async (id) => useEnvironmentsStore.setState({ activeId: id }))
    await show(['hml', 'dev', 'gone'])
    expect(screen.getByRole('button', { name: 'Send' })).toBeDisabled()
    expect(screen.getByRole('status')).toHaveTextContent('mx-test-user is only for HML, DEV, Deleted environment.')
    expect(screen.getAllByRole('button', { name: /^Switch to/ }).map((button) => button.textContent)).toEqual(['Switch to HML', 'Switch to DEV'])
    await userEvent.click(screen.getByRole('button', { name: 'Switch to HML' }))
    expect(setActive).toHaveBeenCalledWith('hml')
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Send' })).toBeEnabled()
  })

  it('points to the Aliases menu when every environment the alias was for is gone', async () => {
    await show(['gone'])
    expect(screen.getByRole('status')).toHaveTextContent('mx-test-user is only for environments that were deleted. Change its environments from the Aliases menu.')
    expect(screen.queryByRole('button', { name: /^Switch to/ })).not.toBeInTheDocument()
  })

  it('appears when the active environment changes while the alias is shown', async () => {
    useEnvironmentsStore.setState({ activeId: 'hml' })
    await show(['hml'])
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
    act(() => useEnvironmentsStore.setState({ activeId: null }))
    expect(screen.getByRole('status')).toHaveTextContent('mx-test-user is only for HML.')
    expect(screen.getByRole('button', { name: 'Send' })).toBeDisabled()
  })

  it('stays away for an alias for all environments and for the Default', async () => {
    await show([])
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Send' })).toBeEnabled()
  })

  it('stays away for the Default even when its aliases are linked elsewhere', async () => {
    await show(['hml'], null)
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Send' })).toBeEnabled()
  })
})
