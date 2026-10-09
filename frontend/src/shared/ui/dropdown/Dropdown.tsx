import { createContext, use, useEffect, useId, useMemo, useRef, useState, type Dispatch, type KeyboardEvent, type ReactNode, type RefObject, type SetStateAction } from 'react'
import { cn } from '@/shared/lib/cn'
import { moveFocus } from '@/shared/lib/roving-focus'
import styles from './Dropdown.module.css'

interface DropdownContextValue {
  open: boolean
  menuId: string
  setOpen: Dispatch<SetStateAction<boolean>>
}

const DropdownContext = createContext<DropdownContextValue | null>(null)

const TriggerRefContext = createContext<RefObject<HTMLButtonElement | null> | null>(null)

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
  const triggerRef = useRef<HTMLButtonElement>(null)
  const value = useMemo(() => ({ open, menuId, setOpen }), [open, menuId])

  useEffect(() => {
    if (!open) {
      return
    }
    const root = rootRef.current
    const trigger = triggerRef.current
    function onPointerDown(event: PointerEvent) {
      if (!root?.contains(event.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('pointerdown', onPointerDown)
    return () => {
      document.removeEventListener('pointerdown', onPointerDown)
      const active = document.activeElement
      if (!active || active === document.body || root?.contains(active)) {
        trigger?.focus()
      }
    }
  }, [open])

  return (
    <DropdownContext value={value}>
      <TriggerRefContext value={triggerRef}>
        <div ref={rootRef} className={cn(styles.root, className)}>
          {children}
        </div>
      </TriggerRefContext>
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
  const triggerRef = use(TriggerRefContext)
  return (
    <button
      ref={triggerRef}
      type="button"
      aria-label={label}
      aria-haspopup="menu"
      aria-expanded={dropdown.open}
      aria-controls={dropdown.open ? dropdown.menuId : undefined}
      className={cn(styles.trigger, className)}
      onClick={() => dropdown.setOpen((open) => !open)}
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

interface MenuButtonProps {
  role: 'menuitem' | 'menuitemradio'
  checked?: boolean
  onSelect: () => void
  className?: string
  children: ReactNode
}

function MenuButton({ role, checked, onSelect, className, children }: MenuButtonProps) {
  const dropdown = useDropdown()
  return (
    <button
      type="button"
      role={role}
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

interface ItemProps {
  onSelect: () => void
  className?: string
  children: ReactNode
}

function Item({ onSelect, className, children }: ItemProps) {
  return (
    <MenuButton role="menuitem" onSelect={onSelect} className={className}>
      {children}
    </MenuButton>
  )
}

interface RadioItemProps {
  checked: boolean
  onSelect: () => void
  className?: string
  children: ReactNode
}

function RadioItem({ checked, onSelect, className, children }: RadioItemProps) {
  return (
    <MenuButton role="menuitemradio" checked={checked} onSelect={onSelect} className={className}>
      {children}
    </MenuButton>
  )
}

export const Dropdown = { Root, Trigger, Menu, Item, RadioItem }
