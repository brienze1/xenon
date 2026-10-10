import { describe, expect, it } from 'vitest'
import { addableScopes, isEditableScope, isValidVariableName } from './variable-items'

describe('variable items', () => {
  it('offers only the tables the request can use, strongest first', () => {
    expect(addableScopes({ scopes: { globals: 'globals', environment: 'env:e1' } })).toEqual(['environment', 'globals'])
    expect(addableScopes({ scopes: {} })).toEqual([])
  })

  it('edits only environment, collection and globals values', () => {
    expect(['environment', 'collection', 'globals', 'dynamic', 'none'].map((scope) => isEditableScope(scope as never))).toEqual([
      true,
      true,
      true,
      false,
      false,
    ])
  })

  it('accepts names Go accepts', () => {
    expect(isValidVariableName('api_token')).toBe(true)
    expect(isValidVariableName(' spaced ')).toBe(true)
    expect(isValidVariableName('   ')).toBe(false)
    expect(isValidVariableName('a{b')).toBe(false)
    expect(isValidVariableName('a\nb')).toBe(false)
  })
})
