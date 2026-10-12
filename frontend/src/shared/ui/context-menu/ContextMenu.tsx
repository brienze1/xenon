import { createContext, use, useEffect, useEffectEvent, useMemo, useRef, useState, type KeyboardEvent, type MouseEvent, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { cn } from '@/shared/lib/cn'
import { toPagePixels } from '@/shared/lib/page-zoom'
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

interface ButtonProps {
  /** The accessible name; the visible content is usually an icon. */
  label: string
  title?: string
  className?: string
  /** -1 keeps it out of the Tab order where a composite widget (a tree) owns keyboard focus. */
  tabIndex?: number
  children: ReactNode
}

/** Opens the same menu from a click, below the button: for actions that would otherwise need a right click. */
function Button({ label, title, className, tabIndex, children }: ButtonProps) {
  const { state, actions } = useContextMenu()
  return (
    <button
      type="button"
      aria-label={label}
      aria-haspopup="menu"
      title={title}
      className={className}
      tabIndex={tabIndex}
      // Keeps the click away from what the button sits in (a row that opens, drags or toggles).
      onPointerDown={(event) => event.stopPropagation()}
      onDoubleClick={(event) => event.stopPropagation()}
      onClick={(event) => {
        event.stopPropagation()
        if (state.menu) {
          actions.close()
          return
        }
        const rect = event.currentTarget.getBoundingClientRect()
        actions.openAt({ x: rect.left, y: rect.bottom }, event.currentTarget)
      }}
    >
      {children}
    </button>
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
    // Capture phase: a press on another menu's button (which stops its own bubbling) still closes this menu.
    // A press on this menu's own button is left to the button, which toggles it.
    function onPointerDown(event: PointerEvent) {
      const target = event.target as Node
      if (!node?.contains(target) && !(returnFocus?.getAttribute('aria-haspopup') === 'menu' && returnFocus.contains(target))) {
        dismiss()
      }
    }
    document.addEventListener('pointerdown', onPointerDown, true)
    return () => {
      document.removeEventListener('pointerdown', onPointerDown, true)
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
      style={{ left: toPagePixels(menu.position.x), top: toPagePixels(menu.position.y) }}
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

export const ContextMenu = { Provider, Root, Trigger, Button, Menu, Item, Separator }
