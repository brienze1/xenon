import { useRef } from 'react'
import { useFitHeight } from '@/shared/hooks/use-fit-height'
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

  useFitHeight(ref, value)

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
