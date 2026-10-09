import { createContext, use, useEffect, useId, useRef, type KeyboardEvent, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { cn } from '@/shared/lib/cn'
import styles from './Dialog.module.css'

const FOCUSABLE = 'button:not(:disabled), [href], input:not(:disabled), select, textarea, [tabindex]:not([tabindex="-1"])'

function trapFocus(dialog: HTMLElement, event: KeyboardEvent<HTMLElement>) {
  const focusable = Array.from(dialog.querySelectorAll<HTMLElement>(FOCUSABLE))
  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  if (!first || !last) {
    event.preventDefault()
    return
  }
  const active = document.activeElement
  const inside = active instanceof Node && dialog.contains(active)
  if (event.shiftKey && (active === first || !inside)) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && (active === last || !inside)) {
    event.preventDefault()
    first.focus()
  }
}

const DialogTitleContext = createContext<string>('')

interface RootProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  className?: string
  children: ReactNode
}

function Root({ open, onOpenChange, className, children }: RootProps) {
  const titleId = useId()
  const dialogRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) {
      return
    }
    const previous = document.activeElement as HTMLElement | null
    const dialog = dialogRef.current
    const first = dialog?.querySelector<HTMLElement>(FOCUSABLE)
    if (first) {
      first.focus()
    } else {
      dialog?.focus()
    }
    return () => {
      previous?.focus()
    }
  }, [open])

  function handleKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.defaultPrevented) {
      return
    }
    if (event.key === 'Escape') {
      event.preventDefault()
      event.stopPropagation()
      onOpenChange(false)
      return
    }
    if (event.key === 'Tab') {
      trapFocus(event.currentTarget, event)
    }
  }

  if (!open) {
    return null
  }

  return createPortal(
    <div className={styles.backdrop} onMouseDown={(event) => event.target === event.currentTarget && onOpenChange(false)}>
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        tabIndex={-1}
        onKeyDown={handleKeyDown}
        className={cn(styles.dialog, className)}
      >
        <DialogTitleContext value={titleId}>{children}</DialogTitleContext>
      </div>
    </div>,
    document.body,
  )
}

interface TitleProps {
  children: ReactNode
}

function Title({ children }: TitleProps) {
  return (
    <h2 id={use(DialogTitleContext)} className={styles.title}>
      {children}
    </h2>
  )
}

interface BodyProps {
  children: ReactNode
}

function Body({ children }: BodyProps) {
  return <div className={styles.body}>{children}</div>
}

interface ActionsProps {
  children: ReactNode
}

function Actions({ children }: ActionsProps) {
  return <div className={styles.actions}>{children}</div>
}

export const Dialog = { Root, Title, Body, Actions }
