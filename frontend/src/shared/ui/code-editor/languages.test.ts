import { describe, expect, it, vi } from 'vitest'

vi.mock('@codemirror/lang-json', () => {
  throw new Error('Failed to fetch dynamically imported module')
})

const { loadLanguage } = await import('./languages')

describe('loadLanguage fallback', () => {
  it('falls back to plain text when a grammar fails to load', async () => {
    await expect(loadLanguage('json')).resolves.toEqual([])
  })
})
