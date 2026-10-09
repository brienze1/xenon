import { createContext, use, useEffect, useId, useMemo, useRef, useState, type KeyboardEvent, type ReactNode, type RefObject } from 'react'
import { cn } from '@/shared/lib/cn'
import { moveFocus } from '@/shared/lib/roving-focus'
import styles from './Dropdown.module.css'

interface DropdownState {
  open: boolean
}

interface DropdownActions {
  toggle: () => void
  close: () => void
}

interface DropdownMeta {
  menuId: string
  triggerRef: RefObject<HTMLButtonElement | null>
}

interface DropdownContextValue {
  state: DropdownState
  actions: DropdownActions
  meta: DropdownMeta
}

const DropdownContext = createContext<DropdownContextValue | null>(null)

function useDropdown(): DropdownContextValue {
  const context = use(DropdownContext)
  if (!context) {
    throw new Error('Dropdown parts must be rendered inside <Dropdown.Root>')
  }
  return context
}

interface ProviderProps extends DropdownContextValue {
  children: ReactNode
}

function Provider({ state, actions, meta, children }: ProviderProps) {
  const value = useMemo(() => ({ state, actions, meta }), [state, actions, meta])
  return <DropdownContext value={value}>{children}</DropdownContext>
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
  const state = useMemo(() => ({ open }), [open])
  const actions = useMemo(() => ({ toggle: () => setOpen((current) => !current), close: () => setOpen(false) }), [])
  const meta = useMemo(() => ({ menuId, triggerRef }), [menuId])

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
    <Provider state={state} actions={actions} meta={meta}>
      <div ref={rootRef} className={cn(styles.root, className)}>
        {children}
      </div>
    </Provider>
  )
}

interface TriggerProps {
  label?: string
  className?: string
  children: ReactNode
}

function Trigger({ label, className, children }: TriggerProps) {
  const {
    state,
    actions,
    meta: { menuId, triggerRef },
  } = useDropdown()
  return (
    <button
      ref={triggerRef}
      type="button"
      aria-label={label}
      aria-haspopup="menu"
      aria-expanded={state.open}
      aria-controls={state.open ? menuId : undefined}
      className={cn(styles.trigger, className)}
      onClick={actions.toggle}
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
  const { state, actions, meta } = useDropdown()
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (state.open && menuRef.current) {
      moveFocus(menuRef.current, 'ArrowDown', '[role^="menuitem"]')
    }
  }, [state.open])

  if (!state.open) {
    return null
  }

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key === 'Escape') {
      event.preventDefault()
      actions.close()
      return
    }
    if (moveFocus(event.currentTarget, event.key, '[role^="menuitem"]')) {
      event.preventDefault()
    }
  }

  return (
    <div
      ref={menuRef}
      id={meta.menuId}
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
  const { actions } = useDropdown()
  return (
    <button
      type="button"
      role={role}
      aria-checked={checked}
      className={cn(styles.item, className)}
      onClick={() => {
        actions.close()
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

export const Dropdown = { Provider, Root, Trigger, Menu, Item, RadioItem }
