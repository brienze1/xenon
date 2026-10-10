import { describe, expect, it } from 'vitest'
import { clashLabel, countsLabel, sourceLabel } from './labels'

describe('import labels', () => {
  it('names the source with its version', () => {
    expect(sourceLabel('postman_collection', '2.1')).toBe('Postman collection v2.1')
    expect(sourceLabel('bruno_bru', '')).toBe('Bruno collection (.bru)')
    expect(sourceLabel('other', '')).toBe('other')
  })

  it('lists the non-zero counts with singular and plural words', () => {
    expect(countsLabel({ collections: 1, folders: 3, requests: 14, examples: 1, variables: 0, scripts: 0, environments: 0 })).toBe(
      '1 collection · 3 folders · 14 requests · 1 example',
    )
    expect(countsLabel({ collections: 0, folders: 0, requests: 0, examples: 0, variables: 2, scripts: 1, environments: 1 })).toBe(
      '2 variables · 1 script · 1 environment',
    )
    expect(countsLabel({ collections: 0, folders: 0, requests: 0, examples: 0, variables: 0, scripts: 0, environments: 0 })).toBe('nothing')
  })

  it('describes a clash', () => {
    expect(clashLabel('request', 'Health')).toBe('A request named “Health” already exists')
  })
})
