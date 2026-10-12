import { describe, expect, it } from 'vitest'
import { defaultUIState, toUIState, toUIStateDto } from './ui-state'

const LAYOUT = {
  sidebar_width: 280,
  sidebar_visible: true,
  right_panel_width: 320,
  right_panel_visible: true,
  right_panel_view: 'info',
  console_height: 200,
  console_visible: false,
  split: 'stacked',
  split_ratio: 0.5,
}

function dto(tabs: object[], view = 'collections') {
  return {
    local_id: 'ws',
    version: 1,
    active_tab_id: '',
    preview_tab_id: '',
    tabs,
    closed_tabs: [],
    snippet_options: [],
    sidebar: { expanded: [], selected_id: '', view, search: '' },
    layout: LAYOUT,
  } as unknown as Parameters<typeof toUIState>[0]
}

const TAB = { request_id: '', example_id: '', environment_id: '', base_revision: '', pinned: true }

describe('ui-state mapper', () => {
  it('reads environment and globals tabs and the environments sidebar view', () => {
    const state = toUIState(
      dto(
        [
          { ...TAB, tab_id: 'e', kind: 'environment', environment_id: 'env1' },
          { ...TAB, tab_id: 'g', kind: 'globals' },
          { ...TAB, tab_id: 'bad', kind: 'environment' },
        ],
        'environments',
      ),
    )
    expect(state.tabs.map((tab) => [tab.kind, tab.environmentId])).toEqual([
      ['environment', 'env1'],
      ['globals', null],
      ['request', null],
    ])
    expect(state.sidebar.view).toBe('environments')
  })

  it('falls back to the collections view for an unknown view', () => {
    expect(toUIState(dto([], 'graph')).sidebar.view).toBe('collections')
  })

  it('writes the environment id of a tab', () => {
    const state = { ...defaultUIState(), tabs: [{ tabId: 'e', kind: 'environment' as const, environmentId: 'env1', requestId: null, pinned: false, baseRevision: null, draft: null }] }
    const written = toUIStateDto('ws', state)
    expect(written.tabs?.[0]).toMatchObject({ kind: 'environment', environment_id: 'env1', request_id: '' })
  })
})
