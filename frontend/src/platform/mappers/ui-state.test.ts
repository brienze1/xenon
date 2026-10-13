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

function dto(tabs: object[], view = 'collections', historySource?: string) {
  return {
    local_id: 'ws',
    version: 1,
    active_tab_id: '',
    preview_tab_id: '',
    history_source: historySource,
    tabs,
    closed_tabs: [],
    snippet_options: [],
    sidebar: { expanded: [], selected_id: '', view, search: '' },
    layout: LAYOUT,
  } as unknown as Parameters<typeof toUIState>[0]
}

const TAB = { request_id: '', example_id: '', environment_id: '', item_id: '', base_revision: '', pinned: true }

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

  it('reads collection and folder tabs with their item id', () => {
    const state = toUIState(
      dto([
        { ...TAB, tab_id: 'c', kind: 'collection', item_id: 'col1' },
        { ...TAB, tab_id: 'f', kind: 'folder', item_id: 'fold1' },
        { ...TAB, tab_id: 'bad', kind: 'folder' },
        { ...TAB, tab_id: 'old', kind: 'request', request_id: 'r1', item_id: undefined },
      ]),
    )
    expect(state.tabs.map((tab) => [tab.kind, tab.itemId])).toEqual([
      ['collection', 'col1'],
      ['folder', 'fold1'],
      ['request', null],
      ['request', null],
    ])
  })

  it('writes the item id of a collection or folder tab', () => {
    const state = { ...defaultUIState(), tabs: [{ tabId: 'f', kind: 'folder' as const, itemId: 'fold1', requestId: null, pinned: true, baseRevision: null, draft: null }] }
    expect(toUIStateDto('ws', state).tabs?.[0]).toMatchObject({ kind: 'folder', item_id: 'fold1', request_id: '', environment_id: '' })
  })

  it('falls back to the collections view for an unknown view', () => {
    expect(toUIState(dto([], 'graph')).sidebar.view).toBe('collections')
  })

  it('writes the environment id of a tab', () => {
    const state = { ...defaultUIState(), tabs: [{ tabId: 'e', kind: 'environment' as const, environmentId: 'env1', requestId: null, pinned: false, baseRevision: null, draft: null }] }
    const written = toUIStateDto('ws', state)
    expect(written.tabs?.[0]).toMatchObject({ kind: 'environment', environment_id: 'env1', request_id: '' })
  })

  it('reads and writes the settings tab', () => {
    const state = toUIState(dto([{ ...TAB, tab_id: 's', kind: 'settings' }]))
    expect(state.tabs[0]).toMatchObject({ tabId: 's', kind: 'settings', requestId: null })
    expect(toUIStateDto('ws', state).tabs?.[0]).toMatchObject({ tab_id: 's', kind: 'settings' })
  })

  it('remembers the history source filter, defaulting to all', () => {
    expect(toUIState(dto([], 'history', 'ai')).sidebar.historySource).toBe('ai')
    expect(toUIState(dto([], 'history', 'robots')).sidebar.historySource).toBe('all')
    expect(toUIState(dto([])).sidebar.historySource).toBe('all')
    const state = { ...defaultUIState(), sidebar: { ...defaultUIState().sidebar, historySource: 'user' as const } }
    expect(toUIStateDto('ws', state).history_source).toBe('user')
  })
})
