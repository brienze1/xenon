import { createContext, use, useEffect, useEffectEvent, useMemo, useRef, useState, type KeyboardEvent, type MouseEvent, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { cn } from '@/shared/lib/cn'
import { moveFocus } from '@/shared/lib/roving-focus'
import styles from './ContextMenu.module.css'

interface Position {
  x: number
  y: number
}

interface OpenMenu {
  position: Position
  returnFocus: HTMLElement | null
}

interface ContextMenuState {
  menu: OpenMenu | null
}

interface ContextMenuActions {
  openAt: (position: Position, returnFocus: HTMLElement | null) => void
  close: () => void
}

type ContextMenuMeta = Record<string, never>

interface ContextMenuContextValue {
  state: ContextMenuState
  actions: ContextMenuActions
  meta: ContextMenuMeta
}

const ContextMenuContext = createContext<ContextMenuContextValue | null>(null)

function useContextMenu(): ContextMenuContextValue {
  const context = use(ContextMenuContext)
  if (!context) {
    throw new Error('ContextMenu parts must be rendered inside <ContextMenu.Root>')
  }
  return context
}

interface ProviderProps extends ContextMenuContextValue {
  children: ReactNode
}

function Provider({ state, actions, meta, children }: ProviderProps) {
  const value = useMemo(() => ({ state, actions, meta }), [state, actions, meta])
  return <ContextMenuContext value={value}>{children}</ContextMenuContext>
}

const NO_META: ContextMenuMeta = {}

interface RootProps {
  children: ReactNode
}

function Root({ children }: RootProps) {
  const [menu, setMenu] = useState<OpenMenu | null>(null)
  const state = useMemo(() => ({ menu }), [menu])
  const actions = useMemo<ContextMenuActions>(
    () => ({
      openAt: (position, returnFocus) => setMenu({ position, returnFocus }),
      close: () => setMenu(null),
    }),
    [],
  )
  return (
    <Provider state={state} actions={actions} meta={NO_META}>
      {children}
    </Provider>
  )
}

const FOCUSABLE = 'button, [href], input, select, textarea, [tabindex]'

function focusableWithin(trigger: HTMLElement, target: EventTarget): HTMLElement | null {
  if (!(target instanceof HTMLElement)) {
    return null
  }
  const focusable = target.closest<HTMLElement>(FOCUSABLE)
  return focusable && trigger.contains(focusable) ? focusable : null
}

function isMenuKey(event: KeyboardEvent<HTMLElement>): boolean {
  return event.key === 'ContextMenu' || (event.key === 'F10' && event.shiftKey)
}

interface TriggerProps {
  className?: string
  children: ReactNode
}

function Trigger({ className, children }: TriggerProps) {
  const { actions } = useContextMenu()
  function onContextMenu(event: MouseEvent<HTMLDivElement>) {
    event.preventDefault()
    actions.openAt({ x: event.clientX, y: event.clientY }, focusableWithin(event.currentTarget, event.target))
  }
  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (!isMenuKey(event) || !(event.target instanceof HTMLElement)) {
      return
    }
    event.preventDefault()
    const rect = event.target.getBoundingClientRect()
    actions.openAt({ x: rect.left, y: rect.bottom }, event.target)
  }
  return (
    <div className={className} onContextMenu={onContextMenu} onKeyDown={onKeyDown}>
      {children}
    </div>
  )
}

interface MenuProps {
  label: string
  children: ReactNode
}

function Menu({ label, children }: MenuProps) {
  const {
    state: { menu },
    actions: { close },
  } = useContextMenu()
  const menuRef = useRef<HTMLDivElement>(null)
  const open = menu !== null
  const returnFocus = menu?.returnFocus ?? null
  const dismiss = useEffectEvent(close)

  useEffect(() => {
    if (!open) {
      return
    }
    const previous = returnFocus ?? (document.activeElement as HTMLElement | null)
    const node = menuRef.current
    if (node) {
      moveFocus(node, 'ArrowDown', '[role="menuitem"]')
    }
    function onPointerDown(event: PointerEvent) {
      if (!node?.contains(event.target as Node)) {
        dismiss()
      }
    }
    document.addEventListener('pointerdown', onPointerDown)
    return () => {
      document.removeEventListener('pointerdown', onPointerDown)
      const active = document.activeElement
      if (!active || active === document.body || node?.contains(active)) {
        previous?.focus()
      }
    }
  }, [open, returnFocus])

  if (!menu) {
    return null
  }

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key === 'Escape' || event.key === 'Tab') {
      event.preventDefault()
      close()
      return
    }
    if (moveFocus(event.currentTarget, event.key, '[role="menuitem"]')) {
      event.preventDefault()
    }
  }

  return createPortal(
    <div
      ref={menuRef}
      role="menu"
      aria-label={label}
      className={styles.menu}
      style={{ left: menu.position.x, top: menu.position.y }}
      onKeyDown={onKeyDown}
    >
      {children}
    </div>,
    document.body,
  )
}

interface ItemProps {
  onSelect: () => void
  tone?: 'default' | 'danger'
  /** A disabled item stays focusable (WAI-ARIA menus) but does nothing. */
  disabled?: boolean
  children: ReactNode
}

function Item({ onSelect, tone = 'default', disabled = false, children }: ItemProps) {
  const {
    actions: { close },
  } = useContextMenu()
  return (
    <button
      type="button"
      role="menuitem"
      aria-disabled={disabled ? true : undefined}
      className={cn(styles.item, tone === 'danger' && styles.danger, disabled && styles.disabled)}
      onClick={() => {
        if (disabled) {
          return
        }
        close()
        onSelect()
      }}
    >
      {children}
    </button>
  )
}

function Separator() {
  return <div role="separator" className={styles.separator} />
}

export const ContextMenu = { Provider, Root, Trigger, Menu, Item, Separator }
