import { createContext, use, useEffect, useId, useRef, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { cn } from '@/shared/lib/cn'
import styles from './Dialog.module.css'

const FOCUSABLE = 'button:not(:disabled), [href], input:not(:disabled), select, textarea, [tabindex]:not([tabindex="-1"])'

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
    dialogRef.current?.querySelector<HTMLElement>(FOCUSABLE)?.focus()
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        event.preventDefault()
        onOpenChange(false)
      }
    }
    document.addEventListener('keydown', onKeyDown)
    return () => {
      document.removeEventListener('keydown', onKeyDown)
      previous?.focus()
    }
  }, [open, onOpenChange])

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

function Title({ children }: { children: ReactNode }) {
  return (
    <h2 id={use(DialogTitleContext)} className={styles.title}>
      {children}
    </h2>
  )
}

function Body({ children }: { children: ReactNode }) {
  return <div className={styles.body}>{children}</div>
}

function Actions({ children }: { children: ReactNode }) {
  return <div className={styles.actions}>{children}</div>
}

export const Dialog = { Root, Title, Body, Actions }
