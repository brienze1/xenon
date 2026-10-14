import { describe, expect, it } from 'vitest'
import type { HistoryActivity } from '@/platform/mappers/history'
import { activityName, activityOutcome, aiTag } from './history-labels'

const activity: HistoryActivity = { tool: 'tree', action: 'create', target: null, outcome: 'ok', code: null, durationMs: 12, args: null }

describe('history labels', () => {
  it('tags AI rows with the client name', () => {
    expect(aiTag('Claude Code')).toBe('AI · Claude Code')
    expect(aiTag(null)).toBe('AI')
  })

  it('names the tool and action', () => {
    expect(activityName(activity)).toBe('tree.create')
    expect(activityName({ ...activity, action: null })).toBe('tree')
    expect(activityName(null)).toBe('unknown')
  })

  it('labels each outcome', () => {
    expect(activityOutcome(activity)).toEqual({ label: 'ok', tone: 'success' })
    expect(activityOutcome({ ...activity, outcome: 'denied' })).toEqual({ label: 'denied', tone: 'redirect' })
    expect(activityOutcome({ ...activity, outcome: 'timed_out' })).toEqual({ label: 'timed out', tone: 'redirect' })
    expect(activityOutcome({ ...activity, outcome: 'error', code: 'TRE-01404' })).toEqual({ label: 'TRE-01404', tone: 'error' })
    expect(activityOutcome({ ...activity, outcome: 'error' })).toEqual({ label: 'error', tone: 'error' })
  })
})
