import { describe, expect, it } from 'vitest'
import type { TreeNode } from '@/platform/mappers/tree'
import { planRow, runConfig, runRow } from '@/features/runner/test-fixtures'
import { curvePoints, defaultParam, usersAt, withProfileKind, withUsersAndDuration } from './load-profile'
import { filterCounts, isRunInProgress, matchesFilter, resultItems } from './results'
import { chooseAlias, configWithRows, moveRow, resetRows, selectedCount, selectionState, setAllSelected, toggleRow } from './run-order'

function node(id: string, kind: TreeNode['kind'], children: TreeNode[] = []): TreeNode {
  return { id, kind, name: id, method: null, url: null, error: null, warning: null, children }
}

describe('run order', () => {
  const rows = [planRow('r1', 1), planRow('r2', 2), planRow('r3', 3)]

  it('counts the selection and reports the select-all state', () => {
    expect(selectionState(rows)).toBe(true)
    const partial = toggleRow(rows, 'r2', false)
    expect(selectedCount(partial)).toBe(2)
    expect(selectionState(partial)).toBe('mixed')
    expect(selectionState(setAllSelected(rows, false))).toBe(false)
  })

  it('moves a row and renumbers the list', () => {
    expect(moveRow(rows, 'r3', 'r1').map((row) => [row.requestId, row.number])).toEqual([
      ['r3', 1],
      ['r1', 2],
      ['r2', 3],
    ])
    expect(moveRow(rows, 'r1', 'missing')).toBe(rows)
  })

  it('chooses an alias per row and builds the configuration rows', () => {
    const chosen = chooseAlias(rows, 'r2', 'a1')
    expect(configWithRows(runConfig(), toggleRow(chosen, 'r3', false)).rows).toEqual([
      { requestId: 'r1', aliasId: '', selected: true },
      { requestId: 'r2', aliasId: 'a1', selected: true },
      { requestId: 'r3', aliasId: '', selected: false },
    ])
  })

  it('resets to the sidebar order with every row checked, keeping alias choices', () => {
    const tree = [node('c1', 'collection', [node('f1', 'folder', [node('r2', 'request')]), node('r3', 'request'), node('r1', 'request')])]
    const shuffled = [planRow('r1', 1, { selected: false }), planRow('r3', 2, { aliasId: 'a1' }), planRow('r2', 3)]
    expect(resetRows(shuffled, tree, 'c1').map((row) => [row.requestId, row.number, row.selected, row.aliasId])).toEqual([
      ['r2', 1, true, ''],
      ['r3', 2, true, 'a1'],
      ['r1', 3, true, ''],
    ])
  })
})

describe('load profiles', () => {
  const performance = runConfig().performance

  it('fills FR-011 defaults when a profile is chosen', () => {
    expect(withProfileKind(performance, 'ramp_up').profile).toMatchObject({ kind: 'ramp_up', initialUsers: 1, rampSeconds: 150 })
    expect(withProfileKind(performance, 'spike').profile).toMatchObject({ baseUsers: 1, spikeSeconds: 60 })
    expect(withProfileKind(performance, 'peak').profile).toMatchObject({ baseUsers: 1, peakSeconds: 300 })
    expect(defaultParam('baseUsers', 50, 10)).toBe(5)
    expect(defaultParam('spikeSeconds', 50, 5)).toBe(1)
  })

  it('recomputes untouched defaults when users or duration change, keeping edited ones', () => {
    const ramp = withProfileKind(performance, 'ramp_up')
    expect(withUsersAndDuration(ramp, 10, 120).profile).toMatchObject({ initialUsers: 1, rampSeconds: 30 })
    const edited = { ...ramp, profile: { ...ramp.profile, rampSeconds: 99 } }
    expect(withUsersAndDuration(edited, 20, 120)).toMatchObject({ virtualUsers: 20, durationSeconds: 120, profile: { rampSeconds: 99 } })
    const spike = withProfileKind(performance, 'spike')
    expect(withUsersAndDuration(spike, 50, 600).profile.baseUsers).toBe(5)
  })

  it('shapes the user curve', () => {
    const settings = { ...performance, virtualUsers: 10, durationSeconds: 100 }
    expect(usersAt(settings, 50)).toBe(10)
    const ramp = { ...settings, profile: { ...settings.profile, kind: 'ramp_up' as const, initialUsers: 2, rampSeconds: 40 } }
    expect(usersAt(ramp, 20)).toBe(6)
    expect(usersAt(ramp, 60)).toBe(10)
    const spike = { ...settings, profile: { ...settings.profile, kind: 'spike' as const, baseUsers: 1, spikeSeconds: 10 } }
    expect([usersAt(spike, 10), usersAt(spike, 50), usersAt(spike, 90)]).toEqual([1, 10, 1])
    const peak = { ...settings, profile: { ...settings.profile, kind: 'peak' as const, baseUsers: 2, peakSeconds: 50 } }
    expect([usersAt(peak, 0), usersAt(peak, 50), usersAt(peak, 100)]).toEqual([2, 10, 2])
    expect(curvePoints(settings, 100, 20, 2)).toBe('0,0 50,0 100,0')
  })
})

describe('results', () => {
  const rows = [
    runRow(0, { outcome: 'passed' }),
    runRow(1, { outcome: 'failed' }),
    runRow(2, { outcome: 'error' }),
    runRow(3, { outcome: 'cancelled', iteration: 1 }),
    runRow(4, { outcome: 'skipped', iteration: 1 }),
  ]

  it('filters per FR-034 and counts each filter', () => {
    expect(rows.filter((row) => matchesFilter(row, 'failed')).map((row) => row.index)).toEqual([1, 2, 3])
    expect(filterCounts(rows)).toEqual({ all: 5, passed: 1, failed: 3, skipped: 1 })
  })

  it('groups the filtered rows by iteration', () => {
    expect(resultItems(rows, 'failed').map((item) => (item.kind === 'iteration' ? `#${item.iteration}:${item.count}` : item.row.index))).toEqual(['#0:2', 1, 2, '#1:1', 3])
  })
})

describe('isRunInProgress', () => {
  it.each([
    ['running', true],
    ['paused', true],
    ['completed', false],
    ['stopped', false],
    ['stopped_on_error', false],
  ] as const)('%s → %s', (status, expected) => {
    expect(isRunInProgress({ status })).toBe(expected)
  })
})
