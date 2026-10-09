import { useId, useState, type ReactNode } from 'react'
import styles from './Tooltip.module.css'

interface TooltipProps {
  content: ReactNode
  children: ReactNode
}

export function Tooltip({ content, children }: TooltipProps) {
  const [open, setOpen] = useState(false)
  const id = useId()
  return (
    <span
      className={styles.anchor}
      aria-describedby={open ? id : undefined}
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => setOpen(false)}
      onFocus={() => setOpen(true)}
      onBlur={() => setOpen(false)}
    >
      {children}
      {open ? (
        <span role="tooltip" id={id} className={styles.tooltip}>
          {content}
        </span>
      ) : null}
    </span>
  )
}
