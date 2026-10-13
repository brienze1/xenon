import { describe, expect, it } from 'vitest'
import { loadLanguage } from './languages'

describe('loadLanguage for code snippets', () => {
  it.each(['shell', 'go', 'python'] as const)('loads a %s grammar so snippets are coloured', async (language) => {
    expect(await loadLanguage(language)).not.toEqual([])
  })
})
