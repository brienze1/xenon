import { beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyDraft } from '../mappers/request'
import { defaultUIState } from '../mappers/ui-state'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/uistateservice', () => ({ Get: vi.fn(), Save: vi.fn() }))

const UIStateService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/uistateservice')
const { getUIState, saveUIState } = await import('./ui-state')

describe('ui state service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('falls back to the defaults for an empty response', async () => {
    vi.mocked(UIStateService.Get).mockResolvedValue(null as never)
    await expect(getUIState('L1')).resolves.toEqual(defaultUIState())
    expect(UIStateService.Get).toHaveBeenCalledWith({ local_id: 'L1' })
  })

  it('round trips tabs, drafts, sidebar and layout through the DTO', async () => {
    const state = {
      ...defaultUIState(),
      tabs: [
        { tabId: 't1', requestId: null, pinned: true, baseRevision: null, draft: { ...emptyDraft(), url: 'https://u' } },
        { tabId: 't2', requestId: 'r2', pinned: false, baseRevision: 'rev', draft: null },
        { tabId: 't3', kind: 'example' as const, requestId: 'r2', exampleId: 'e1', pinned: true, baseRevision: null, draft: null },
      ],
      activeTabId: 't1',
      previewTabId: 't2',
      sidebar: { expanded: ['c1'], selectedId: 'r2', view: 'history' as const, search: 'q', historySource: 'ai' as const },
      layout: { ...defaultUIState().layout, split: 'side-by-side' as const, splitRatio: 0.6 },
    }
    vi.mocked(UIStateService.Save).mockResolvedValue(undefined as never)
    await saveUIState('L1', state)
    const dto = vi.mocked(UIStateService.Save).mock.calls[0]![0]
    expect(dto).toMatchObject({ local_id: 'L1', version: 1, active_tab_id: 't1', preview_tab_id: 't2', history_source: 'ai', sidebar: { selected_id: 'r2', view: 'history' } })
    expect(dto.tabs?.[0]).toMatchObject({ tab_id: 't1', request_id: '', base_revision: '', draft: { url: 'https://u' } })
    expect(dto.tabs?.[1]).not.toHaveProperty('draft')
    expect(dto.layout).toMatchObject({ split: 'side-by-side', split_ratio: 0.6 })
    vi.mocked(UIStateService.Get).mockResolvedValue(dto as never)
    const restored = await getUIState('L1')
    expect(restored.tabs[0]!.draft?.url).toBe('https://u')
    expect(dto.tabs?.[1]).toMatchObject({ kind: 'request', example_id: '' })
    expect(dto.tabs?.[2]).toMatchObject({ kind: 'example', request_id: 'r2', example_id: 'e1' })
    expect(restored.tabs[1]).toEqual({ tabId: 't2', kind: 'request', requestId: 'r2', exampleId: null, environmentId: null, itemId: null, pinned: false, baseRevision: 'rev', draft: null })
    expect(restored.tabs[2]).toMatchObject({ tabId: 't3', kind: 'example', requestId: 'r2', exampleId: 'e1' })
    expect(restored).toMatchObject({ activeTabId: 't1', previewTabId: 't2', sidebar: state.sidebar, layout: state.layout })
  })

  it('converts failures to AppError', async () => {
    const failure = Object.assign(new Error('x'), { cause: { error: { code: 'WKS-07404', description: 'unknown' } } })
    vi.mocked(UIStateService.Get).mockRejectedValue(failure)
    vi.mocked(UIStateService.Save).mockRejectedValue(failure)
    await expect(getUIState('L1')).rejects.toMatchObject({ code: 'WKS-07404' })
    await expect(saveUIState('L1', defaultUIState())).rejects.toMatchObject({ code: 'WKS-07404' })
  })
})
