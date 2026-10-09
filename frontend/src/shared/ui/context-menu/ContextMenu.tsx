import { createContext, use, useEffect, useEffectEvent, useRef, useState, type KeyboardEvent, type MouseEvent, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { cn } from '@/shared/lib/cn'
import { moveFocus } from '@/shared/lib/roving-focus'
import styles from './ContextMenu.module.css'

interface Position {
  x: number
  y: number
}

interface ContextMenuContextValue {
  position: Position | null
  setPosition: (position: Position | null) => void
}

const ContextMenuContext = createContext<ContextMenuContextValue | null>(null)

function useContextMenu(): ContextMenuContextValue {
  const context = use(ContextMenuContext)
  if (!context) {
    throw new Error('ContextMenu parts must be rendered inside <ContextMenu.Root>')
  }
  return context
}

interface RootProps {
  children: ReactNode
}

function Root({ children }: RootProps) {
  const [position, setPosition] = useState<Position | null>(null)
  return <ContextMenuContext value={{ position, setPosition }}>{children}</ContextMenuContext>
}

interface TriggerProps {
  className?: string
  children: ReactNode
}

function Trigger({ className, children }: TriggerProps) {
  const menu = useContextMenu()
  function onContextMenu(event: MouseEvent<HTMLDivElement>) {
    event.preventDefault()
    menu.setPosition({ x: event.clientX, y: event.clientY })
  }
  return (
    <div className={className} onContextMenu={onContextMenu}>
      {children}
    </div>
  )
}

interface MenuProps {
  label: string
  children: ReactNode
}

function Menu({ label, children }: MenuProps) {
  const menu = useContextMenu()
  const menuRef = useRef<HTMLDivElement>(null)
  const open = menu.position !== null
  const close = useEffectEvent(() => menu.setPosition(null))

  useEffect(() => {
    if (!open) {
      return
    }
    if (menuRef.current) {
      moveFocus(menuRef.current, 'ArrowDown', '[role="menuitem"]')
    }
    function onPointerDown(event: PointerEvent) {
      if (!menuRef.current?.contains(event.target as Node)) {
        close()
      }
    }
    document.addEventListener('pointerdown', onPointerDown)
    return () => document.removeEventListener('pointerdown', onPointerDown)
  }, [open])

  if (!menu.position) {
    return null
  }

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key === 'Escape') {
      event.preventDefault()
      menu.setPosition(null)
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
  children: ReactNode
}

function Item({ onSelect, tone = 'default', children }: ItemProps) {
  const menu = useContextMenu()
  return (
    <button
      type="button"
      role="menuitem"
      className={cn(styles.item, tone === 'danger' && styles.danger)}
      onClick={() => {
        menu.setPosition(null)
        onSelect()
      }}
    >
      {children}
    </button>
  )
}

export const ContextMenu = { Root, Trigger, Menu, Item }
