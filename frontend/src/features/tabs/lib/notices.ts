import type { TabNotice } from '@/features/tabs/model'

export function withoutNotice(notices: Record<string, TabNotice>, tabId: string): Record<string, TabNotice> {
  const next = { ...notices }
  delete next[tabId]
  return next
}
