import { describe, expect, it } from 'vitest'
import { AppError } from '@/platform/errors'
import { cancelledResult, type SendResult } from '@/platform/mappers/request'
import { EMPTY_ENTRY, withResult } from './entry'

function result(inlineBody: string | null): SendResult {
  return { ...cancelledResult('tab'), error: null, ok: true, statusCode: 200, inlineBody }
}

describe('withResult', () => {
  it('starts a fresh entry at version 1 and waits to fetch a spilled body', () => {
    const next = withResult(undefined, result(null))
    expect(next).toMatchObject({ version: 1, sending: false, failure: null, body: null, bodyStatus: 'idle' })
  })

  it('uses an inline body at once and bumps the version of an existing entry', () => {
    const previous = { ...EMPTY_ENTRY, version: 4, sending: true, failure: new AppError('X', 'boom') }
    const next = withResult(previous, result('{"a":1}'))
    expect(next).toMatchObject({ version: 5, sending: false, failure: null, body: '{"a":1}', bodyStatus: 'ready' })
  })
})
