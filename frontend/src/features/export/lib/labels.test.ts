import { describe, expect, it } from 'vitest'
import { countsLabel, formatLabel, passphraseProblemLabel, secretsLabel, shapeLabel } from './labels'

describe('export labels', () => {
  it('names formats, shapes, secrets and passphrase problems', () => {
    expect(formatLabel('xenon_yaml')).toBe('Xenon YAML')
    expect(formatLabel('postman')).toBe('Postman')
    expect(shapeLabel('bundle')).toBe('Single YAML file')
    expect(shapeLabel('zip')).toBe('Zip of the folder')
    expect(secretsLabel('names_only')).toBe('Secret variables are exported by name with empty values.')
    expect(passphraseProblemLabel('mismatch')).toBe('The passphrases do not match.')
  })

  it('lists the non-zero counts in order', () => {
    expect(countsLabel({ collections: 1, folders: 0, requests: 1000, examples: 1, environments: 0, variables: 2 })).toBe('1 collection · 1,000 requests · 1 example · 2 variables')
    expect(countsLabel({ collections: 0, folders: 0, requests: 0, examples: 0, environments: 0, variables: 0 })).toBe('nothing')
  })
})
