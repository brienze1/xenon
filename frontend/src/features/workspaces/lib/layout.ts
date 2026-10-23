import type { LayoutState } from '@/platform/mappers/ui-state'

export const LAYOUT_LIMITS = {
  sidebar: { min: 200, max: 600 },
  rightPanel: { min: 240, max: 600 },
  console: { min: 120, maxShare: 0.6 },
  split: { min: 0.2, max: 0.8, minSize: 200 },
} as const

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value))
}

export function consoleMaxHeight(viewportHeight: number): number {
  return Math.max(LAYOUT_LIMITS.console.min, Math.round(viewportHeight * LAYOUT_LIMITS.console.maxShare))
}

export function clampLayout(layout: LayoutState, viewportHeight: number): LayoutState {
  return {
    ...layout,
    sidebarWidth: clamp(layout.sidebarWidth, LAYOUT_LIMITS.sidebar.min, LAYOUT_LIMITS.sidebar.max),
    rightPanelWidth: clamp(layout.rightPanelWidth, LAYOUT_LIMITS.rightPanel.min, LAYOUT_LIMITS.rightPanel.max),
    consoleHeight: clamp(layout.consoleHeight, LAYOUT_LIMITS.console.min, consoleMaxHeight(viewportHeight)),
    splitRatio: clamp(layout.splitRatio, LAYOUT_LIMITS.split.min, LAYOUT_LIMITS.split.max),
  }
}
