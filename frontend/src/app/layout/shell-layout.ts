import { createContext } from 'react'
import { defaultLayout, type LayoutState } from '@/platform/mappers/ui-state'

export interface ShellLayout {
  layout: LayoutState
  onLayoutChange: (patch: Partial<LayoutState>) => void
}

export const ShellLayoutContext = createContext<ShellLayout>({ layout: defaultLayout(), onLayoutChange: () => undefined })
