export function moveTab<T extends { id: string }>(tabs: T[], fromId: string, toId: string): T[] {
  const from = tabs.findIndex((tab) => tab.id === fromId)
  const to = tabs.findIndex((tab) => tab.id === toId)
  if (from < 0 || to < 0 || from === to) {
    return tabs
  }
  const next = [...tabs]
  const [moved] = next.splice(from, 1)
  next.splice(to, 0, moved!)
  return next
}

export function neighbourOf<T extends { id: string }>(tabs: T[], tabId: string): string | null {
  const index = tabs.findIndex((tab) => tab.id === tabId)
  return tabs[index + 1]?.id ?? tabs[index - 1]?.id ?? null
}
