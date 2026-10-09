export type RovingOrientation = 'vertical' | 'horizontal'

interface RovingOptions {
  orientation?: RovingOrientation
  wrap?: boolean
}

const STEP_KEYS: Record<RovingOrientation, Record<string, number>> = {
  vertical: { ArrowDown: 1, ArrowUp: -1 },
  horizontal: { ArrowRight: 1, ArrowLeft: -1 },
}

export function isRovingKey(key: string, orientation: RovingOrientation = 'vertical'): boolean {
  return key === 'Home' || key === 'End' || STEP_KEYS[orientation][key] !== undefined
}

export function rovingIndex(count: number, current: number, key: string, options: RovingOptions = {}): number | null {
  const { orientation = 'vertical', wrap = true } = options
  if (count === 0 || !isRovingKey(key, orientation)) {
    return null
  }
  if (key === 'Home') {
    return 0
  }
  if (key === 'End') {
    return count - 1
  }
  const offset = STEP_KEYS[orientation][key] as number
  if (current < 0) {
    return offset > 0 ? 0 : count - 1
  }
  const next = current + offset
  if (wrap) {
    return (next + count) % count
  }
  return Math.min(Math.max(next, 0), count - 1)
}

export function moveFocus(container: HTMLElement, key: string, selector: string, orientation: RovingOrientation = 'vertical'): boolean {
  if (!isRovingKey(key, orientation)) {
    return false
  }
  const items = Array.from(container.querySelectorAll<HTMLElement>(selector))
  const next = rovingIndex(items.length, items.indexOf(document.activeElement as HTMLElement), key, { orientation })
  if (next !== null) {
    items[next]?.focus()
  }
  return true
}
