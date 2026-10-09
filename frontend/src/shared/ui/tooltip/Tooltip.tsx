import { cloneElement, useId, useState, type ReactElement, type ReactNode } from 'react'
import styles from './Tooltip.module.css'

interface DescribableProps {
  'aria-describedby'?: string
}

interface TooltipProps {
  content: ReactNode
  children: ReactElement<DescribableProps>
}

export function Tooltip({ content, children }: TooltipProps) {
  const [open, setOpen] = useState(false)
  const id = useId()
  const describedBy = [children.props['aria-describedby'], open ? id : undefined].filter(Boolean).join(' ') || undefined
  return (
    <span
      className={styles.anchor}
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => setOpen(false)}
      onFocus={() => setOpen(true)}
      onBlur={() => setOpen(false)}
    >
      {cloneElement(children, { 'aria-describedby': describedBy })}
      {open ? (
        <span role="tooltip" id={id} className={styles.tooltip}>
          {content}
        </span>
      ) : null}
    </span>
  )
}
