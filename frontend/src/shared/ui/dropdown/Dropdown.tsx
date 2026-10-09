import { createContext, use, useEffect, useId, useRef, useState, type KeyboardEvent, type ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'
import { moveFocus } from '@/shared/lib/roving-focus'
import styles from './Dropdown.module.css'

interface DropdownContextValue {
  open: boolean
  menuId: string
  setOpen: (open: boolean) => void
}

const DropdownContext = createContext<DropdownContextValue | null>(null)

function useDropdown(): DropdownContextValue {
  const context = use(DropdownContext)
  if (!context) {
    throw new Error('Dropdown parts must be rendered inside <Dropdown.Root>')
  }
  return context
}

interface RootProps {
  className?: string
  children: ReactNode
}

function Root({ className, children }: RootProps) {
  const [open, setOpen] = useState(false)
  const menuId = useId()
  const rootRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) {
      return
    }
    function onPointerDown(event: PointerEvent) {
      if (!rootRef.current?.contains(event.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('pointerdown', onPointerDown)
    return () => document.removeEventListener('pointerdown', onPointerDown)
  }, [open])

  return (
    <DropdownContext value={{ open, menuId, setOpen }}>
      <div ref={rootRef} className={cn(styles.root, className)}>
        {children}
      </div>
    </DropdownContext>
  )
}

interface TriggerProps {
  label?: string
  className?: string
  children: ReactNode
}

function Trigger({ label, className, children }: TriggerProps) {
  const dropdown = useDropdown()
  return (
    <button
      type="button"
      aria-label={label}
      aria-haspopup="menu"
      aria-expanded={dropdown.open}
      aria-controls={dropdown.open ? dropdown.menuId : undefined}
      className={cn(styles.trigger, className)}
      onClick={() => dropdown.setOpen(!dropdown.open)}
    >
      {children}
    </button>
  )
}

interface MenuProps {
  label: string
  className?: string
  children: ReactNode
}

function Menu({ label, className, children }: MenuProps) {
  const dropdown = useDropdown()
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (dropdown.open && menuRef.current) {
      moveFocus(menuRef.current, 'ArrowDown', '[role^="menuitem"]')
    }
  }, [dropdown.open])

  if (!dropdown.open) {
    return null
  }

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key === 'Escape') {
      event.preventDefault()
      dropdown.setOpen(false)
      return
    }
    if (moveFocus(event.currentTarget, event.key, '[role^="menuitem"]')) {
      event.preventDefault()
    }
  }

  return (
    <div
      ref={menuRef}
      id={dropdown.menuId}
      role="menu"
      aria-label={label}
      className={cn(styles.menu, className)}
      onKeyDown={onKeyDown}
    >
      {children}
    </div>
  )
}

interface ItemProps {
  onSelect: () => void
  checked?: boolean
  className?: string
  children: ReactNode
}

function Item({ onSelect, checked, className, children }: ItemProps) {
  const dropdown = useDropdown()
  return (
    <button
      type="button"
      role={checked === undefined ? 'menuitem' : 'menuitemradio'}
      aria-checked={checked}
      className={cn(styles.item, className)}
      onClick={() => {
        dropdown.setOpen(false)
        onSelect()
      }}
    >
      {children}
    </button>
  )
}

export const Dropdown = { Root, Trigger, Menu, Item }
