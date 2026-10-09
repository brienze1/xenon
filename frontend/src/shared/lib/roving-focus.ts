const OFFSETS: Record<string, number> = { ArrowDown: 1, ArrowUp: -1 }

export function moveFocus(container: HTMLElement, key: string, selector: string): boolean {
  const offset = OFFSETS[key]
  if (offset === undefined) {
    return false
  }
  const items = Array.from(container.querySelectorAll<HTMLElement>(selector))
  if (items.length === 0) {
    return true
  }
  const current = items.indexOf(document.activeElement as HTMLElement)
  const next = current < 0 ? (offset > 0 ? 0 : items.length - 1) : (current + offset + items.length) % items.length
  items[next]?.focus()
  return true
}
