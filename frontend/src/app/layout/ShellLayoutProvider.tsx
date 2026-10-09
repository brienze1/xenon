import { useMemo, type ReactNode } from 'react'
import { ShellLayoutContext, type ShellLayoutContextValue } from './shell-layout'

interface ShellLayoutProviderProps extends ShellLayoutContextValue {
  children: ReactNode
}

export function ShellLayoutProvider({ state, actions, meta, children }: ShellLayoutProviderProps) {
  const value = useMemo(() => ({ state, actions, meta }), [state, actions, meta])
  return <ShellLayoutContext value={value}>{children}</ShellLayoutContext>
}
