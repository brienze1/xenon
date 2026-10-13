import { describe, expect, it } from 'vitest'
import { splitProtectedTerms } from './protected-terms'

describe('splitProtectedTerms', () => {
  it('marks every occurrence of a protected term', () => {
    expect(splitProtectedTerms('Update Xenon to open workspace.yaml in Xenon.', ['Xenon', 'workspace.yaml'])).toEqual([
      { text: 'Update ', protected: false },
      { text: 'Xenon', protected: true },
      { text: ' to open ', protected: false },
      { text: 'workspace.yaml', protected: true },
      { text: ' in ', protected: false },
      { text: 'Xenon', protected: true },
      { text: '.', protected: false },
    ])
  })

  it('returns the text untouched without terms', () => {
    expect(splitProtectedTerms('plain', [])).toEqual([{ text: 'plain', protected: false }])
    expect(splitProtectedTerms('', ['Xenon'])).toEqual([])
  })
})
