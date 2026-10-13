import { describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({
  CancelError: class CancelError extends Error {},
}))

const { CancelError } = await import('@wailsio/runtime')
const { AppError, CANCELLED_CODE, UNKNOWN_CODE, toAppError } = await import('./errors')

class RuntimeError extends Error {
  constructor(message: string, override readonly cause: unknown) {
    super(message)
  }
}

describe('toAppError', () => {
  it('reads code, description and details from the RuntimeError cause', () => {
    const error = toAppError(
      new RuntimeError('bad request', {
        error: {
          id: 'tx',
          code: 'WKS-00400',
          description: 'invalid request',
          error_details: [{ attribute: 'name', messages: ["failed on the 'required' rule"] }],
        },
      }),
    )
    expect(error).toBeInstanceOf(AppError)
    expect(error.code).toBe('WKS-00400')
    expect(error.description).toBe('invalid request')
    expect(error.message).toBe('invalid request')
    expect(error.details).toEqual([{ attribute: 'name', messages: ["failed on the 'required' rule"] }])
    expect(error.cancelled).toBe(false)
  })

  it('parses a cause given as a JSON string', () => {
    const error = toAppError(new RuntimeError('x', JSON.stringify({ error: { code: 'WKS-06404' } })))
    expect(error.code).toBe('WKS-06404')
    expect(error.description).toBe('')
    expect(error.details).toEqual([])
  })

  it('maps a cancelled call to the cancelled code', () => {
    const error = toAppError(new CancelError('cancelled'))
    expect(error.code).toBe(CANCELLED_CODE)
    expect(error.cancelled).toBe(true)
  })

  it('keeps an AppError as it is', () => {
    const original = new AppError('WKS-01409', 'not empty')
    expect(toAppError(original)).toBe(original)
  })

  it('falls back to the unknown code', () => {
    expect(toAppError(new Error('boom'))).toMatchObject({ code: UNKNOWN_CODE, description: 'boom' })
    expect(toAppError('text')).toMatchObject({ code: UNKNOWN_CODE, description: 'text' })
    expect(toAppError(new RuntimeError('x', 'not json')).code).toBe(UNKNOWN_CODE)
    expect(toAppError(new RuntimeError('x', null)).code).toBe(UNKNOWN_CODE)
    expect(toAppError(new RuntimeError('x', { error: { code: 42 } })).code).toBe(UNKNOWN_CODE)
  })

  it('tolerates malformed details', () => {
    const error = toAppError(new RuntimeError('x', { error: { code: 'A-1', error_details: [{ attribute: 1, messages: 'x' }] } }))
    expect(error.details).toEqual([{ attribute: '', messages: [] }])
  })
})
