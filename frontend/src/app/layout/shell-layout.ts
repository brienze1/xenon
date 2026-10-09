import { createContext } from 'react'
import { defaultLayout, type LayoutState } from '@/platform/mappers/ui-state'

export const MAIN_CONTENT_ID = 'main-content'

export interface ShellLayout {
  layout: LayoutState
  onLayoutChange: (patch: Partial<LayoutState>) => void
}

export const ShellLayoutContext = createContext<ShellLayout>({ layout: defaultLayout(), onLayoutChange: () => undefined })
