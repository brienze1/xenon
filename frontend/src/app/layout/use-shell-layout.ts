import { use } from 'react'
import { ShellLayoutContext, type ShellLayoutContextValue } from '@/app/layout/shell-layout'

export function useShellLayout(): ShellLayoutContextValue {
  const context = use(ShellLayoutContext)
  if (!context) {
    throw new Error('Shell layout regions must be rendered inside <ShellLayoutProvider>')
  }
  return context
}
