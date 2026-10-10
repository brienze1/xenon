import { describe, expect, it } from 'vitest'
import type { VariableInspection, VariableReference } from '@/platform/mappers/inspection'
import { decorationsByField, referenceTone, referenceTooltip, scopeLabel } from './variable-decorations'

function reference(patch: Partial<VariableReference>): VariableReference {
  return { field: 'url', name: 'base_url', start: 0, end: 12, status: 'resolved', scope: 'environment', displayValue: 'https://dev', secret: false, ...patch }
}

describe('variable decorations', () => {
  it('names the source scope', () => {
    expect(scopeLabel('environment', 'dev')).toBe('Environment: dev')
    expect(scopeLabel('environment', null)).toBe('Environment')
    expect(scopeLabel('collection', 'dev')).toBe('Collection')
    expect(scopeLabel('globals', null)).toBe('Globals')
    expect(scopeLabel('dynamic', null)).toBe('Dynamic')
    expect(scopeLabel('none', null)).toBe('Not defined')
  })

  it('colours resolved green, secrets accent, dynamic info and the rest red', () => {
    expect(referenceTone(reference({}))).toBe('success')
    expect(referenceTone(reference({ secret: true }))).toBe('accent')
    expect(referenceTone(reference({ status: 'dynamic' }))).toBe('info')
    for (const status of ['unresolved', 'missing_secret', 'locked_secret'] as const) {
      expect(referenceTone(reference({ status, secret: true }))).toBe('danger')
    }
  })

  it('shows the value and the source in the hover, never generating dynamic values', () => {
    expect(referenceTooltip(reference({}), 'dev')).toBe('https://dev — Environment: dev')
    expect(referenceTooltip(reference({ secret: true, displayValue: '•••••' }), 'dev')).toBe('••••• — Environment: dev')
    expect(referenceTooltip(reference({ status: 'dynamic', scope: 'dynamic', displayValue: '' }), null)).toBe('Generated at send — Dynamic variable')
    expect(referenceTooltip(reference({ status: 'missing_secret' }), 'dev')).toBe('Secret with no value on this machine — Environment: dev')
    expect(referenceTooltip(reference({ status: 'locked_secret', scope: 'globals' }), null)).toBe('Secret in a locked vault — Globals')
    expect(referenceTooltip(reference({ status: 'unresolved', scope: 'none' }), null)).toBe('Unresolved variable — sent as written')
  })

  it('groups the references per field', () => {
    const inspection: VariableInspection = {
      references: [reference({}), reference({ start: 13, end: 22, name: 'user_id', status: 'unresolved', scope: 'none' }), reference({ field: 'headers.0.value' })],
      variables: [],
      scopes: {},
    }
    const byField = decorationsByField(inspection, 'dev')
    expect(byField.get('url')).toEqual([
      { from: 0, to: 12, tone: 'success', tooltip: 'https://dev — Environment: dev' },
      { from: 13, to: 22, tone: 'danger', tooltip: 'Unresolved variable — sent as written' },
    ])
    expect(byField.get('headers.0.value')).toHaveLength(1)
    expect(decorationsByField(null, null).size).toBe(0)
  })
})
