import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyAuth } from '@/platform/mappers/request'

vi.mock('@/platform/services/oauth', () => ({
  getNewToken: vi.fn(),
  refreshToken: vi.fn(),
  getToken: vi.fn(),
  revealToken: vi.fn(),
  copyToken: vi.fn(),
  chooseToken: vi.fn(),
  clearToken: vi.fn(),
}))
vi.mock('@/platform/services/requests', async (importOriginal) => ({ ...(await importOriginal<typeof import('@/platform/services/requests')>()), autoHeaders: vi.fn() }))

const oauth = await import('@/platform/services/oauth')
const { RequestAuth } = await import('./RequestAuth')
const { ItemAuth } = await import('./ItemAuth')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')

describe('auth composition', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(oauth.getToken).mockResolvedValue(null)
    useRequestBuilderStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
  })

  it('edits the request draft and keeps an unsaved request token by tab', async () => {
    const user = userEvent.setup()
    const { container, unmount } = render(<RequestAuth tabId="tab-9" />)
    expect(container).toBeEmptyDOMElement()
    unmount()
    useRequestBuilderStore.getState().ensureTab('tab-9')
    render(<RequestAuth tabId="tab-9" />)
    await user.selectOptions(screen.getByRole('combobox', { name: 'Auth type' }), 'oauth2')
    expect(useRequestBuilderStore.getState().tabs['tab-9']!.draft.auth.type).toBe('oauth2')
    await vi.waitFor(() => expect(oauth.getToken).toHaveBeenCalledWith('ws', 'tab-9'))
  })

  it('edits a collection auth owned by the collection', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<ItemAuth slot={{ itemId: 'c1', kind: 'collection', auth: { ...emptyAuth(), type: 'oauth2' }, onChange }} />)
    await vi.waitFor(() => expect(oauth.getToken).toHaveBeenCalledWith('ws', 'c1'))
    await user.selectOptions(screen.getByRole('combobox', { name: 'Auth type' }), 'none')
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ type: 'none' }))
  })
})
