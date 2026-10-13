import { afterEach, describe, expect, it } from 'vitest'
import { pageZoom, toPagePixels } from './page-zoom'

describe('page zoom', () => {
  afterEach(() => document.body.style.removeProperty('zoom'))

  it('turns screen coordinates into page coordinates under the UI zoom', () => {
    expect(pageZoom()).toBe(1)
    expect(toPagePixels(100)).toBe(100)
    document.body.style.setProperty('zoom', '1.25')
    expect(pageZoom()).toBe(1.25)
    expect(toPagePixels(100)).toBe(80)
  })
})
