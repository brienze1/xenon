import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyDraft } from '@/platform/mappers/request'
import { defaultUIState } from '@/platform/mappers/ui-state'

vi.mock('@/platform/services/snippets', () => ({ listSnippetLanguages: vi.fn(), generateSnippet: vi.fn(), copySnippet: vi.fn() }))
vi.mock('@/platform/services/requests', () => ({ autoHeaders: vi.fn(async () => []), applyUrl: vi.fn(async (draft) => draft), applyParams: vi.fn(async (draft) => draft), getRequest: vi.fn(), releaseResponse: vi.fn(async () => undefined) }))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn(async () => undefined) }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(async () => []), searchTree: vi.fn() }))
vi.mock('@/platform/services/environments', () => ({ listEnvironments: vi.fn(async () => ({ environments: [], activeId: null })), setActiveEnvironment: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({ inspectVariables: vi.fn(async () => null), suggestVariables: vi.fn(async () => []) }))

const snippets = await import('@/platform/services/snippets')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { useSidePanelStore } = await import('@/features/side-panel/store')
const { CodePanel } = await import('./CodePanel')

const LANGUAGES = [
  {
    id: 'curl',
    label: 'cURL',
    options: [
      { id: 'multiLine', label: 'Generate multiline snippet', description: 'Split cURL command across multiple lines', type: 'boolean' as const, defaultValue: 'true', values: [] },
      { id: 'quoteType', label: 'Quote type', description: 'Quote to use', type: 'enum' as const, defaultValue: 'single', values: ['single', 'double'] },
      { id: 'requestTimeoutInSeconds', label: 'Set request timeout (in seconds)', description: 'Use 0 for infinity', type: 'positiveInteger' as const, defaultValue: '0', values: [] },
    ],
  },
  { id: 'python_requests', label: 'Python - Requests', options: [] },
]

function lastInput() {
  return vi.mocked(snippets.generateSnippet).mock.calls.at(-1)![0]
}

describe('CodePanel', () => {
  let tabId = ''

  beforeEach(() => {
    vi.resetAllMocks()
    useSidePanelStore.getState().reset()
    useTabsStore.getState().reset()
    useRequestBuilderStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useTabsStore.getState().hydrate('ws', defaultUIState())
    tabId = useTabsStore.getState().activeTabId!
    useRequestBuilderStore.getState().loadDraft(tabId, { ...emptyDraft(), url: 'https://api.test/users' }, null, true)
    vi.mocked(snippets.listSnippetLanguages).mockResolvedValue(LANGUAGES)
    vi.mocked(snippets.generateSnippet).mockImplementation(async (input) => ({ text: `${input.language}: ${input.draft.url}`, notices: [] }))
    vi.mocked(snippets.copySnippet).mockResolvedValue()
  })

  afterEach(() => {
    useSidePanelStore.getState().reset()
    useTabsStore.getState().reset()
  })

  it('shows the current draft as cURL by default, masked, and follows edits', async () => {
    render(<CodePanel tabId={tabId} />)
    expect(await screen.findByRole('textbox', { name: 'cURL snippet' })).toHaveTextContent('curl: https://api.test/users')
    expect(lastInput()).toMatchObject({ localId: 'ws', tabId, language: 'curl', options: {}, includeSecrets: false })
    act(() => useRequestBuilderStore.getState().setUrl(tabId, 'https://api.test/orders'))
    await waitFor(() => expect(screen.getByRole('textbox', { name: 'cURL snippet' })).toHaveTextContent('curl: https://api.test/orders'))
  })

  it('switches language and remembers the options per language', async () => {
    render(<CodePanel tabId={tabId} />)
    await screen.findByRole('textbox', { name: 'cURL snippet' })
    await userEvent.click(screen.getByRole('button', { name: 'Snippet settings' }))
    const dialog = screen.getByRole('dialog', { name: 'Settings for cURL' })
    expect(dialog).toHaveTextContent('Split cURL command across multiple lines')
    expect(screen.getByRole('switch', { name: 'Generate multiline snippet' })).toBeChecked()
    await userEvent.click(screen.getByRole('switch', { name: 'Generate multiline snippet' }))
    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Quote type' }), 'double')
    await userEvent.click(screen.getByRole('button', { name: 'Done' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    await waitFor(() => expect(lastInput().options).toEqual({ multiLine: 'false', quoteType: 'double' }))
    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Language' }), 'python_requests')
    await waitFor(() => expect(screen.getByRole('textbox', { name: 'Python - Requests snippet' })).toHaveTextContent('python_requests: https://api.test/users'))
    expect(useTabsStore.getState().snippetOptions).toEqual([
      { language: 'curl', options: { multiLine: 'false', quoteType: 'double' } },
      { language: 'python_requests', options: {} },
    ])
    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Language' }), 'curl')
    await waitFor(() => expect(lastInput()).toMatchObject({ language: 'curl', options: { multiLine: 'false', quoteType: 'double' } }))
  })

  it('sets a numeric option', async () => {
    render(<CodePanel tabId={tabId} />)
    await screen.findByRole('textbox', { name: 'cURL snippet' })
    await userEvent.click(screen.getByRole('button', { name: 'Snippet settings' }))
    const timeout = screen.getByRole('spinbutton', { name: 'Set request timeout (in seconds)' })
    await userEvent.clear(timeout)
    await userEvent.type(timeout, '5')
    await waitFor(() => expect(lastInput().options).toEqual({ requestTimeoutInSeconds: '5' }))
  })

  it('includes secrets only on request and copies through Go', async () => {
    render(<CodePanel tabId={tabId} />)
    await screen.findByRole('textbox', { name: 'cURL snippet' })
    await userEvent.click(screen.getByRole('checkbox', { name: 'Include Secrets' }))
    await waitFor(() => expect(lastInput().includeSecrets).toBe(true))
    expect(screen.queryByRole('button', { name: 'Copy' })).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Copy snippet' }))
    expect(snippets.copySnippet).toHaveBeenCalledWith(expect.objectContaining({ language: 'curl', includeSecrets: true, tabId }))
    expect(screen.getByText('The snippet is on the clipboard.')).toBeInTheDocument()
  })

  it('shows notices and errors', async () => {
    vi.mocked(snippets.generateSnippet).mockResolvedValueOnce({ text: '# No OAuth 2.0 access token yet\ncurl x', notices: ['No OAuth 2.0 access token yet: get one in the Authorization tab.'] })
    render(<CodePanel tabId={tabId} />)
    expect(await screen.findByText('No OAuth 2.0 access token yet: get one in the Authorization tab.')).toBeInTheDocument()
    vi.mocked(snippets.copySnippet).mockRejectedValue({ cause: { error: { code: 'SNP-01400', description: 'bad' } } })
    await userEvent.click(screen.getByRole('button', { name: 'Copy snippet' }))
    expect(screen.getByText('The snippet could not be copied.')).toBeInTheDocument()
    vi.mocked(snippets.generateSnippet).mockRejectedValue({ cause: { error: { code: 'WKS-07404', description: 'The workspace is gone' } } })
    act(() => useRequestBuilderStore.getState().setUrl(tabId, 'https://other'))
    expect(await screen.findByRole('alert')).toHaveTextContent('The workspace is gone')
  })

  it('reports languages that cannot be loaded', async () => {
    vi.mocked(snippets.listSnippetLanguages).mockRejectedValue({ cause: { error: { code: 'UNK-00500', description: 'No languages' } } })
    render(<CodePanel tabId={tabId} />)
    expect(await screen.findByRole('alert')).toHaveTextContent('No languages')
  })
})
