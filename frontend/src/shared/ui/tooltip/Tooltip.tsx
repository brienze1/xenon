import { cloneElement, useId, useLayoutEffect, useRef, useState, type CSSProperties, type ReactElement, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import styles from './Tooltip.module.css'

interface DescribableProps {
  'aria-describedby'?: string
}

interface TooltipProps {
  content: ReactNode
  children: ReactElement<DescribableProps>
}

const GAP_PX = 4
const EDGE_PX = 8

// Drawn on document.body with fixed coordinates so scrolling or clipped panes (overflow: hidden) never cut it off.
// Prefers the space above the anchor, flips below when there is not enough room, and stays inside the window.
function placement(anchor: DOMRect, tip: DOMRect): CSSProperties {
  const above = anchor.top - GAP_PX - tip.height
  const top = above >= EDGE_PX ? above : Math.min(anchor.bottom + GAP_PX, window.innerHeight - tip.height - EDGE_PX)
  const centred = anchor.left + anchor.width / 2 - tip.width / 2
  const left = Math.max(EDGE_PX, Math.min(centred, window.innerWidth - tip.width - EDGE_PX))
  return { top, left }
}

export function Tooltip({ content, children }: TooltipProps) {
  const [open, setOpen] = useState(false)
  const [position, setPosition] = useState<CSSProperties>({ visibility: 'hidden' })
  const anchorRef = useRef<HTMLSpanElement>(null)
  const tipRef = useRef<HTMLSpanElement>(null)
  const id = useId()
  const describedBy = [children.props['aria-describedby'], open ? id : undefined].filter(Boolean).join(' ') || undefined

  useLayoutEffect(() => {
    if (open && anchorRef.current && tipRef.current) {
      setPosition(placement(anchorRef.current.getBoundingClientRect(), tipRef.current.getBoundingClientRect()))
    }
  }, [open, content])

  return (
    <span
      ref={anchorRef}
      className={styles.anchor}
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => {
        setOpen(false)
        setPosition({ visibility: 'hidden' })
      }}
      onFocus={() => setOpen(true)}
      onBlur={() => {
        setOpen(false)
        setPosition({ visibility: 'hidden' })
      }}
    >
      {cloneElement(children, { 'aria-describedby': describedBy })}
      {open
        ? createPortal(
            <span ref={tipRef} role="tooltip" id={id} className={styles.tooltip} style={position}>
              {content}
            </span>,
            document.body,
          )
        : null}
    </span>
  )
}
