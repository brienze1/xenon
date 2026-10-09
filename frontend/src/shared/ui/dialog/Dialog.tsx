import { createContext, use, useEffect, useId, useMemo, useRef, type KeyboardEvent, type ReactNode } from 'react'
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

interface DialogState {
  open: boolean
}

interface DialogActions {
  close: () => void
}

interface DialogMeta {
  titleId: string
}

interface DialogContextValue {
  state: DialogState
  actions: DialogActions
  meta: DialogMeta
}

const DialogContext = createContext<DialogContextValue | null>(null)

function useDialog(): DialogContextValue {
  const context = use(DialogContext)
  if (!context) {
    throw new Error('Dialog parts must be rendered inside <Dialog.Root>')
  }
  return context
}

interface ProviderProps extends DialogContextValue {
  children: ReactNode
}

function Provider({ state, actions, meta, children }: ProviderProps) {
  const value = useMemo(() => ({ state, actions, meta }), [state, actions, meta])
  return <DialogContext value={value}>{children}</DialogContext>
}

interface RootProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  className?: string
  children: ReactNode
}

function Root({ open, onOpenChange, className, children }: RootProps) {
  const titleId = useId()
  const dialogRef = useRef<HTMLDivElement>(null)
  const state = useMemo(() => ({ open }), [open])
  const actions = useMemo(() => ({ close: () => onOpenChange(false) }), [onOpenChange])
  const meta = useMemo(() => ({ titleId }), [titleId])

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
      actions.close()
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
    <div className={styles.backdrop} onMouseDown={(event) => event.target === event.currentTarget && actions.close()}>
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        tabIndex={-1}
        onKeyDown={handleKeyDown}
        className={cn(styles.dialog, className)}
      >
        <Provider state={state} actions={actions} meta={meta}>
          {children}
        </Provider>
      </div>
    </div>,
    document.body,
  )
}

interface TitleProps {
  children: ReactNode
}

function Title({ children }: TitleProps) {
  const { meta } = useDialog()
  return (
    <h2 id={meta.titleId} className={styles.title}>
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

export const Dialog = { Provider, Root, Title, Body, Actions }
