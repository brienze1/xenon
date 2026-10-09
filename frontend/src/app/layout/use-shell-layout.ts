import { use } from 'react'
import { ShellLayoutContext, type ShellLayout } from '@/app/layout/shell-layout'

export function useShellLayout(): ShellLayout {
  return use(ShellLayoutContext)
}
