import { describe, expect, it } from 'vitest'
import { toCompletions, variableCompletionQuery } from './variable-completion'

describe('variable completion', () => {
  it('finds the name typed after an open {{', () => {
    expect(variableCompletionQuery('{{ba', 4)).toEqual({ from: 2, prefix: 'ba', closed: false })
    expect(variableCompletionQuery('x/{{', 4)).toEqual({ from: 4, prefix: '', closed: false })
    expect(variableCompletionQuery('{{ba}}/x', 4)).toEqual({ from: 2, prefix: 'ba', closed: true })
    expect(variableCompletionQuery('{{a}}/{{b', 9)).toEqual({ from: 8, prefix: 'b', closed: false })
  })

  it('offers nothing outside an open reference', () => {
    expect(variableCompletionQuery('plain', 5)).toBeNull()
    expect(variableCompletionQuery('{{done}} x', 10)).toBeNull()
    expect(variableCompletionQuery('{a', 2)).toBeNull()
  })

  it('inserts the name and closes the reference unless it is closed already', () => {
    const suggestions = [
      { name: 'base_url', scope: 'environment' as const, description: '', secret: false, dynamic: false },
      { name: 'token', scope: 'globals' as const, description: '', secret: true, dynamic: false },
      { name: '$guid', scope: 'dynamic' as const, description: '', secret: false, dynamic: true },
    ]
    expect(toCompletions({ from: 2, prefix: '', closed: false }, suggestions, 'dev')).toEqual({
      from: 2,
      options: [
        { label: 'base_url', detail: 'Environment: dev', apply: 'base_url}}' },
        { label: 'token', detail: 'Globals · secret', apply: 'token}}' },
        { label: '$guid', detail: 'Dynamic', apply: '$guid}}' },
      ],
    })
    expect(toCompletions({ from: 2, prefix: 't', closed: true }, suggestions.slice(1, 2), null).options[0]?.apply).toBe('token')
  })
})
