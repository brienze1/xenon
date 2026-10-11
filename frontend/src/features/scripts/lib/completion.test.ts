import { describe, expect, it } from 'vitest'
import type { ScriptMember } from '@/platform/mappers/scripts'
import { memberCompletions, variableCompletions, variableNameQuery } from './completion'

const members: ScriptMember[] = [
  { path: 'pm', signature: 'pm', description: 'The Postman object', kind: 'object' },
  { path: 'pm.request', signature: 'request', description: 'The request', kind: 'object' },
  { path: 'pm.response', signature: 'response', description: 'The response', kind: 'object' },
  { path: 'pm.test', signature: 'test(name, fn)', description: 'Defines a test', kind: 'function' },
  { path: 'pm.response.to.have.status', signature: 'status(code)', description: 'Asserts the status', kind: 'function' },
  { path: 'pm.response.to.have.header', signature: 'header(key)', description: 'Asserts a header', kind: 'function' },
  { path: 'console', signature: 'console', description: 'Writes to the console', kind: 'object' },
]

describe('script completion', () => {
  it('completes members of the object before the cursor', () => {
    expect(memberCompletions(members, 'pm.re', 5)).toEqual({
      from: 3,
      options: [
        { label: 'request', detail: 'The request', apply: 'request' },
        { label: 'response', detail: 'The response', apply: 'response' },
      ],
    })
    expect(memberCompletions(members, 'x = pm.response.to.have.', 24)?.options.map((option) => option.label)).toEqual(['status', 'header'])
    expect(memberCompletions(members, 'pm.Te', 5)?.options.map((option) => option.label)).toEqual(['test'])
  })

  it('completes globals by name and offers nothing elsewhere', () => {
    expect(memberCompletions(members, 'con', 3)?.options.map((option) => option.label)).toEqual(['console'])
    expect(memberCompletions(members, '', 0)).toBeNull()
    expect(memberCompletions(members, 'foo(', 4)).toBeNull()
    expect(memberCompletions(members, 'x.pm.re', 7)).toBeNull()
    expect(memberCompletions(members, 'pm.zz', 5)).toBeNull()
  })

  it('finds the variable name typed inside pm.<scope>.get/set/unset/has', () => {
    expect(variableNameQuery('pm.environment.get("ba', 22)).toEqual({ from: 20, prefix: 'ba' })
    expect(variableNameQuery("pm.globals.set('", 16)).toEqual({ from: 16, prefix: '' })
    expect(variableNameQuery('pm.collectionVariables.unset(`to', 32)).toEqual({ from: 30, prefix: 'to' })
    expect(variableNameQuery('pm.variables.has( "x', 20)).toEqual({ from: 19, prefix: 'x' })
    expect(variableNameQuery('pm.environment.get("a") + "', 27)).toBeNull()
    expect(variableNameQuery('pm.vault.get("a', 15)).toBeNull()
    expect(variableCompletions({ from: 20, prefix: 'ba' }, [{ name: 'base_url', detail: 'dev' }])).toEqual({
      from: 20,
      options: [{ label: 'base_url', detail: 'dev', apply: 'base_url' }],
    })
  })
})
