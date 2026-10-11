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

const LAST_TAB_POSITION = 9

/** The tab at a 1-based position; position 9 is always the last tab (Cmd/Ctrl+9). */
export function positionAt<T>(tabs: readonly T[], position: number): T | undefined {
  return position >= LAST_TAB_POSITION ? tabs.at(-1) : tabs[position - 1]
}

export function wrappedNeighbour<T extends { id: string }>(tabs: readonly T[], activeId: string | null, step: 1 | -1): T | undefined {
  if (tabs.length === 0) {
    return undefined
  }
  const index = tabs.findIndex((tab) => tab.id === activeId)
  return tabs[(index + step + tabs.length) % tabs.length]
}

export function tabsBefore<T extends { id: string }>(tabs: readonly T[], tabId: string): string[] {
  const index = tabs.findIndex((tab) => tab.id === tabId)
  return tabs.slice(0, Math.max(index, 0)).map((tab) => tab.id)
}

export function tabsAfter<T extends { id: string }>(tabs: readonly T[], tabId: string): string[] {
  const index = tabs.findIndex((tab) => tab.id === tabId)
  return index < 0 ? [] : tabs.slice(index + 1).map((tab) => tab.id)
}
