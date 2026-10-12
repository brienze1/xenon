import { useLayoutEffect, useRef } from 'react'
import { cn } from '@/shared/lib/cn'
import styles from './WrapTextCell.module.css'

interface WrapTextCellProps {
  name: string
  label: string
  value: string
  onChange: (value: string) => void
  description?: string
  className?: string
  translate?: 'no'
}

/** A table cell value that wraps and grows to show all of it (long URLs, tokens), like Postman's variable table. */
export function WrapTextCell({ name, label, value, onChange, description, className, translate }: WrapTextCellProps) {
  const ref = useRef<HTMLTextAreaElement>(null)

  // Fits the height to the wrapped text, again whenever the column width changes.
  useLayoutEffect(() => {
    const node = ref.current
    if (!node) {
      return undefined
    }
    function fit() {
      if (node) {
        node.style.height = 'auto'
        node.style.height = `${node.scrollHeight}px`
      }
    }
    fit()
    const observer = typeof ResizeObserver === 'undefined' ? null : new ResizeObserver(fit)
    observer?.observe(node)
    return () => observer?.disconnect()
  }, [value])

  return (
    <textarea
      ref={ref}
      name={name}
      rows={1}
      autoComplete="off"
      aria-label={label}
      aria-description={description}
      className={cn(styles.cell, className)}
      spellCheck={false}
      translate={translate}
      value={value}
      // A value is one line: Enter does not break it (pasted line breaks are kept).
      onKeyDown={(event) => {
        if (event.key === 'Enter' && !event.shiftKey) {
          event.preventDefault()
        }
      }}
      onChange={(event) => onChange(event.target.value)}
    />
  )
}
