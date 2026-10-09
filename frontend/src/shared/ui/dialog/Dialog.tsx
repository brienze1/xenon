import { createContext, use, useEffect, useEffectEvent, useId, useRef, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { cn } from '@/shared/lib/cn'
import styles from './Dialog.module.css'

const FOCUSABLE = 'button:not(:disabled), [href], input:not(:disabled), select, textarea, [tabindex]:not([tabindex="-1"])'

function trapFocus(dialog: HTMLElement, event: KeyboardEvent) {
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
  const close = useEffectEvent(() => onOpenChange(false))

  useEffect(() => {
    if (!open) {
      return
    }
    const previous = document.activeElement as HTMLElement | null
    dialogRef.current?.querySelector<HTMLElement>(FOCUSABLE)?.focus()
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        event.preventDefault()
        close()
        return
      }
      if (event.key === 'Tab' && dialogRef.current) {
        trapFocus(dialogRef.current, event)
      }
    }
    document.addEventListener('keydown', onKeyDown)
    return () => {
      document.removeEventListener('keydown', onKeyDown)
      previous?.focus()
    }
  }, [open])

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
