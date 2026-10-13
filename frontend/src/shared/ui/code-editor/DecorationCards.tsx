import { use, useCallback, useEffect, useLayoutEffect, useRef, useState, type CSSProperties, type RefObject } from 'react'
import { createPortal } from 'react-dom'
import { toPagePixels } from '@/shared/lib/page-zoom'
import { DecorationCardContext } from './decoration-card-context'
import styles from './DecorationCards.module.css'

interface DecorationCardsProps {
  /** The editor's container; its decorated ranges carry `data-card`. */
  host: RefObject<HTMLElement | null>
}

interface OpenCard {
  card: string
  anchor: HTMLElement
}

const OPEN_DELAY_MS = 250
const CLOSE_DELAY_MS = 250
const GAP_PX = 4
const EDGE_PX = 8

// Below the hovered range (above when there is no room), inside the window; measured in screen pixels, placed in page pixels.
function placement(anchor: DOMRect, card: DOMRect): CSSProperties {
  const below = anchor.bottom + GAP_PX
  const top = below + card.height <= window.innerHeight - EDGE_PX ? below : Math.max(EDGE_PX, anchor.top - GAP_PX - card.height)
  const left = Math.max(EDGE_PX, Math.min(anchor.left, window.innerWidth - card.width - EDGE_PX))
  return { top: toPagePixels(top), left: toPagePixels(left) }
}

function cardAnchor(host: HTMLElement, target: EventTarget | null): HTMLElement | null {
  if (!(target instanceof Element)) {
    return null
  }
  const anchor = target.closest<HTMLElement>('[data-card]')
  return anchor && host.contains(anchor) ? anchor : null
}

/**
 * Hover cards for an editor's decorations, like Postman's variable popover: it opens after a short hover, stays while
 * the pointer is on it or focus is inside it, and closes on Escape, a press elsewhere or typing in the editor.
 */
export function DecorationCards({ host }: DecorationCardsProps) {
  const render = use(DecorationCardContext)
  const [open, setOpen] = useState<OpenCard | null>(null)
  const [position, setPosition] = useState<CSSProperties>({ visibility: 'hidden' })
  const cardRef = useRef<HTMLDivElement>(null)
  const timer = useRef<number | undefined>(undefined)

  const schedule = useCallback((next: () => void, delay: number) => {
    window.clearTimeout(timer.current)
    timer.current = window.setTimeout(next, delay)
  }, [])

  // Leaves the card open while focus is inside it (the value is being typed).
  const closeUnlessFocused = useCallback(
    () => schedule(() => setOpen((current) => (current && cardRef.current?.contains(document.activeElement) ? current : null)), CLOSE_DELAY_MS),
    [schedule],
  )

  useEffect(() => {
    const node = host.current
    if (!node || render === null) {
      return undefined
    }
    function onPointerOver(event: PointerEvent) {
      const anchor = cardAnchor(node as HTMLElement, event.target)
      if (anchor) {
        const card = anchor.dataset.card as string
        schedule(() => setOpen((current) => (current?.card === card && current.anchor === anchor ? current : { card, anchor })), OPEN_DELAY_MS)
      }
    }
    function onPointerOut(event: PointerEvent) {
      if (cardAnchor(node as HTMLElement, event.target) && !cardAnchor(node as HTMLElement, event.relatedTarget)) {
        closeUnlessFocused()
      }
    }
    function onKeyDown() {
      window.clearTimeout(timer.current)
      setOpen(null)
    }
    node.addEventListener('pointerover', onPointerOver)
    node.addEventListener('pointerout', onPointerOut)
    node.addEventListener('keydown', onKeyDown)
    return () => {
      window.clearTimeout(timer.current)
      node.removeEventListener('pointerover', onPointerOver)
      node.removeEventListener('pointerout', onPointerOut)
      node.removeEventListener('keydown', onKeyDown)
    }
  }, [host, render, schedule, closeUnlessFocused])

  useEffect(() => {
    if (open === null) {
      return undefined
    }
    function onPointerDown(event: PointerEvent) {
      const target = event.target as Node
      if (!cardRef.current?.contains(target) && !open?.anchor.contains(target)) {
        setOpen(null)
      }
    }
    document.addEventListener('pointerdown', onPointerDown, true)
    return () => document.removeEventListener('pointerdown', onPointerDown, true)
  }, [open])

  useLayoutEffect(() => {
    if (open && cardRef.current) {
      setPosition(placement(open.anchor.getBoundingClientRect(), cardRef.current.getBoundingClientRect()))
    } else {
      setPosition({ visibility: 'hidden' })
    }
  }, [open])

  const close = useCallback(() => setOpen(null), [])

  if (open === null || render === null) {
    return null
  }

  return createPortal(
    <div
      ref={cardRef}
      className={styles.card}
      style={position}
      onPointerEnter={() => window.clearTimeout(timer.current)}
      onPointerLeave={closeUnlessFocused}
      onKeyDown={(event) => {
        if (event.key === 'Escape') {
          event.preventDefault()
          close()
        }
      }}
    >
      {render(open.card, close)}
    </div>,
    document.body,
  )
}
