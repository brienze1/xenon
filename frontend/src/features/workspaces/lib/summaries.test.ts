import { describe, expect, it } from 'vitest'
import type { WorkspaceSummary } from '@/platform/mappers/workspace'
import { replaceSummary, withMissing } from './summaries'

const acme: WorkspaceSummary = { localId: 'A', name: 'Acme', path: '/work/acme', missing: false }
const beta: WorkspaceSummary = { localId: 'B', name: 'Beta', path: '/work/beta', missing: false }

describe('replaceSummary', () => {
  it('swaps in the summary with the same local id and keeps the order', () => {
    const renamed = { ...acme, name: 'Acme Corp' }
    expect(replaceSummary([acme, beta], renamed)).toEqual([renamed, beta])
  })

  it('ignores a summary that is not listed', () => {
    expect(replaceSummary([acme], beta)).toEqual([acme])
  })
})

describe('withMissing', () => {
  it('flags only the given workspace as missing', () => {
    expect(withMissing([acme, beta], 'B')).toEqual([acme, { ...beta, missing: true }])
  })
})
