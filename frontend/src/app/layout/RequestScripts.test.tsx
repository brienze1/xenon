import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyDraft } from '@/platform/mappers/request'

vi.mock('@/platform/services/scripts', () => ({ checkScript: vi.fn(), scriptReference: vi.fn(), scriptSnippets: vi.fn() }))

const scriptsService = await import('@/platform/services/scripts')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { RequestScripts, scopeDetail } = await import('./RequestScripts')

describe('RequestScripts', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(scriptsService.scriptReference).mockResolvedValue([])
    vi.mocked(scriptsService.scriptSnippets).mockResolvedValue([{ id: 'get-variable', title: 'Get a variable', phase: 'any', text: 'pm.variables.get("variable_key");' }])
    vi.mocked(scriptsService.checkScript).mockResolvedValue([])
    useRequestBuilderStore.getState().reset()
  })

  it('edits the scripts and the inherit toggle of the tab draft', async () => {
    const user = userEvent.setup()
    act(() => useRequestBuilderStore.getState().loadDraft('tab', { ...emptyDraft(), id: 'r1', scripts: { preRequest: 'one()', postResponse: '', disabled: [] } }, null, false))
    render(<RequestScripts tabId="tab" />)
    expect(await screen.findByLabelText('Pre-request script')).toHaveTextContent('one()')
    await user.click(screen.getByRole('checkbox', { name: 'Inherit parent scripts' }))
    expect(useRequestBuilderStore.getState().tabs.tab?.draft.inheritScripts).toBe(false)
    await user.click(screen.getByRole('button', { name: 'Snippets for the Pre-request script' }))
    await user.click(screen.getByRole('menuitem', { name: 'Get a variable' }))
    expect(useRequestBuilderStore.getState().tabs.tab?.draft.scripts.preRequest).toContain('pm.variables.get("variable_key");')
    expect(useRequestBuilderStore.getState().tabs.tab?.dirty).toBe(true)
  })

  it('renders nothing for an unknown tab and names the scope of a suggestion', () => {
    const { container } = render(<RequestScripts tabId="missing" />)
    expect(container).toBeEmptyDOMElement()
    expect(scopeDetail({ name: 'a', scope: 'environment', description: '', secret: true, dynamic: false }, 'dev')).toBe('dev · secret')
    expect(scopeDetail({ name: 'a', scope: 'globals', description: '', secret: false, dynamic: false }, null)).toBe('globals')
  })
})
