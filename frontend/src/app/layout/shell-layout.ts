import { createContext } from 'react'
import type { LayoutState } from '@/platform/mappers/ui-state'

export const MAIN_CONTENT_ID = 'main-content'

export interface ShellLayoutState {
  layout: LayoutState
}

export interface ShellLayoutActions {
  update: (patch: Partial<LayoutState>) => void
}

export interface ShellLayoutMeta {
  mainContentId: string
}

export interface ShellLayoutContextValue {
  state: ShellLayoutState
  actions: ShellLayoutActions
  meta: ShellLayoutMeta
}

export const ShellLayoutContext = createContext<ShellLayoutContextValue | null>(null)
