import { describe, expect, it } from 'vitest'
import { AppError } from '@/platform/errors'
import { syncStatus } from '@/features/sync/test-fixtures'
import { errorText, lastErrorText } from './error-text'
import { changeLabel, NO_ACCOUNT, statusLabel, statusTone } from './status-label'

describe('sync status wording', () => {
  it('labels every state the footer shows', () => {
    expect(statusLabel(syncStatus({ state: 'not_linked' }))).toBeNull()
    expect(statusLabel(syncStatus())).toBe('Up to date')
    expect(statusLabel(syncStatus({ state: 'local_changes', localChanges: 3 }))).toBe('Local changes (3)')
    expect(statusLabel(syncStatus({ state: 'ahead', ahead: 2 }))).toBe('Ahead (2 to push)')
    expect(statusLabel(syncStatus({ state: 'behind', behind: 1 }))).toBe('Behind (1 to pull)')
    expect(statusLabel(syncStatus({ state: 'diverged', ahead: 1, behind: 2 }))).toBe('Ahead 1 · Behind 2')
    expect(statusLabel(syncStatus({ state: 'conflict' }))).toBe('Conflict')
    expect(statusLabel(syncStatus({ state: 'syncing' }))).toBe('Syncing…')
    expect(statusLabel(syncStatus({ state: 'error' }))).toBe('Error')
    expect(statusLabel(syncStatus({ state: 'error', lastError: { code: 'GIT-07428', message: NO_ACCOUNT, nextStep: '' } }))).toBe(NO_ACCOUNT)
  })

  it('tones the dot by urgency and capitalises change kinds', () => {
    expect(statusTone(syncStatus())).toBe('ok')
    expect(statusTone(syncStatus({ state: 'conflict' }))).toBe('attention')
    expect(statusTone(syncStatus({ state: 'error' }))).toBe('attention')
    expect(statusTone(syncStatus({ state: 'behind' }))).toBe('pending')
    expect(changeLabel('renamed')).toBe('Renamed')
  })

  it('joins an error with its next step', () => {
    expect(errorText(new AppError('GIT-10503', 'offline', [{ attribute: 'next_step', messages: ['Check the connection.'] }]))).toBe('offline. Check the connection.')
    expect(errorText(new AppError('GIT-1', ''))).toBe('GIT-1. Try again.')
    expect(lastErrorText({ code: 'GIT-11401', message: 'refused', nextStep: 'Sign in again.' })).toBe('refused. Sign in again.')
    expect(lastErrorText({ code: 'GIT-2', message: '', nextStep: '' })).toBe('GIT-2. Try again.')
  })
})
