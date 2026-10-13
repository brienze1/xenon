import { describe, expect, it } from 'vitest'
import { defaultLayout } from '@/platform/mappers/ui-state'
import { clampLayout, consoleMaxHeight } from './layout'

describe('layout limits', () => {
  it('clamps every size to its range', () => {
    const low = clampLayout({ ...defaultLayout(), sidebarWidth: 10, rightPanelWidth: 10, consoleHeight: 10, splitRatio: 0 }, 1000)
    expect(low).toMatchObject({ sidebarWidth: 200, rightPanelWidth: 240, consoleHeight: 120, splitRatio: 0.2 })
    const high = clampLayout({ ...defaultLayout(), sidebarWidth: 9000, rightPanelWidth: 9000, consoleHeight: 9000, splitRatio: 1 }, 1000)
    expect(high).toMatchObject({ sidebarWidth: 600, rightPanelWidth: 600, consoleHeight: 600, splitRatio: 0.8 })
    expect(clampLayout(defaultLayout(), 1000)).toEqual(defaultLayout())
  })

  it('never lets the console maximum drop below its minimum', () => {
    expect(consoleMaxHeight(1000)).toBe(600)
    expect(consoleMaxHeight(100)).toBe(120)
  })
})
