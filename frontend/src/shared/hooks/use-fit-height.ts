import { useLayoutEffect, type RefObject } from 'react'

// scrollHeight is rounded to whole pixels: at a fractional zoom a box exactly maxLines tall can read one pixel over.
const ROUNDING_PX = 1

function px(value: string): number {
  return Number.parseFloat(value) || 0
}

function fit(node: HTMLTextAreaElement, maxLines: number | undefined) {
  node.style.height = 'auto'
  const content = node.scrollHeight
  const style = window.getComputedStyle(node)
  const lineHeight = Number.parseFloat(style.lineHeight)
  const cap = maxLines !== undefined && Number.isFinite(lineHeight) ? maxLines * lineHeight + px(style.paddingTop) + px(style.paddingBottom) : undefined
  const capped = cap !== undefined && content - cap > ROUNDING_PX
  node.style.height = `${(capped ? cap : content) + px(style.borderTopWidth) + px(style.borderBottomWidth)}px`
  node.style.overflowX = 'hidden'
  node.style.overflowY = capped ? 'auto' : 'hidden'
}

/**
 * Fits a textarea's height to its wrapped text, again whenever its width changes, up to `maxLines` lines; only then
 * does it scroll, and only vertically.
 */
export function useFitHeight(ref: RefObject<HTMLTextAreaElement | null>, value: string, maxLines?: number): void {
  useLayoutEffect(() => {
    const node = ref.current
    if (!node) {
      return undefined
    }
    fit(node, maxLines)
    const observer = typeof ResizeObserver === 'undefined' ? null : new ResizeObserver(() => fit(node, maxLines))
    observer?.observe(node)
    return () => observer?.disconnect()
  }, [ref, value, maxLines])
}
