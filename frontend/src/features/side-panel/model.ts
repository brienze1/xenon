export const SIDE_PANEL_VIEWS = ['variables', 'info'] as const

export type SidePanelView = (typeof SIDE_PANEL_VIEWS)[number]
