import { describe, expect, it } from 'vitest'
import { idleUpdateStatus, toUpdateStatus } from './update-status'

describe('update status mapper', () => {
  it('maps every field to camelCase', () => {
    expect(
      toUpdateStatus({
        state: 'failed',
        reason: 'not_writable',
        current_version: '0.1.4',
        latest_version: '0.1.7',
        release_url: 'https://github.com/brienze1/xenon/releases/tag/v0.1.7',
        folder: '/Applications',
        received: 10,
        total: 20,
        automatic: true,
      }),
    ).toEqual({
      state: 'failed',
      reason: 'not_writable',
      currentVersion: '0.1.4',
      latestVersion: '0.1.7',
      releaseUrl: 'https://github.com/brienze1/xenon/releases/tag/v0.1.7',
      folder: '/Applications',
      received: 10,
      total: 20,
      automatic: true,
    })
  })

  it('reads an empty reason as none and a missing status as idle', () => {
    expect(toUpdateStatus({ state: 'up_to_date', reason: '', current_version: '0.1.4', latest_version: '0.1.4', release_url: '', folder: '', received: 0, total: 0, automatic: false })).toMatchObject({
      state: 'up_to_date',
      reason: null,
    })
    expect(toUpdateStatus(null)).toEqual(idleUpdateStatus())
  })
})
