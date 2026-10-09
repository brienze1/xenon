import { createContext, use, useId, type KeyboardEvent, type ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'
import { rovingIndex } from '@/shared/lib/roving-focus'
import styles from './Tabs.module.css'

interface TabsContextValue {
  value: string
  baseId: string
  onValueChange: (value: string) => void
}

const TabsContext = createContext<TabsContextValue | null>(null)

function useTabs(): TabsContextValue {
  const context = use(TabsContext)
  if (!context) {
    throw new Error('Tabs parts must be rendered inside <Tabs.Root>')
  }
  return context
}

function tabId(baseId: string, value: string): string {
  return `${baseId}-tab-${value}`
}

function panelId(baseId: string, value: string): string {
  return `${baseId}-panel-${value}`
}

interface RootProps {
  value: string
  onValueChange: (value: string) => void
  className?: string
  children: ReactNode
}

function Root({ value, onValueChange, className, children }: RootProps) {
  const baseId = useId()
  return (
    <TabsContext value={{ value, baseId, onValueChange }}>
      <div className={cn(styles.root, className)}>{children}</div>
    </TabsContext>
  )
}

interface ListProps {
  label: string
  className?: string
  children: ReactNode
}

function List({ label, className, children }: ListProps) {
  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const tabs = Array.from(event.currentTarget.querySelectorAll<HTMLButtonElement>('[role="tab"]:not(:disabled)'))
    const current = tabs.indexOf(document.activeElement as HTMLButtonElement)
    const index = current < 0 ? null : rovingIndex(tabs.length, current, event.key, { orientation: 'horizontal' })
    if (index === null) {
      return
    }
    event.preventDefault()
    const next = tabs[index]
    next?.focus()
    next?.click()
  }

  return (
    <div role="tablist" aria-label={label} className={cn(styles.list, className)} onKeyDown={onKeyDown}>
      {children}
    </div>
  )
}

interface TabProps {
  value: string
  disabled?: boolean
  className?: string
  children: ReactNode
}

function Tab({ value, disabled, className, children }: TabProps) {
  const tabs = useTabs()
  const selected = tabs.value === value
  return (
    <button
      type="button"
      role="tab"
      id={tabId(tabs.baseId, value)}
      aria-selected={selected}
      aria-controls={panelId(tabs.baseId, value)}
      tabIndex={selected ? 0 : -1}
      disabled={disabled}
      className={cn(styles.tab, className)}
      onClick={() => tabs.onValueChange(value)}
    >
      {children}
    </button>
  )
}

interface PanelProps {
  value: string
  className?: string
  children: ReactNode
}

function Panel({ value, className, children }: PanelProps) {
  const tabs = useTabs()
  if (tabs.value !== value) {
    return null
  }
  return (
    <div
      role="tabpanel"
      id={panelId(tabs.baseId, value)}
      aria-labelledby={tabId(tabs.baseId, value)}
      className={cn(styles.panel, className)}
    >
      {children}
    </div>
  )
}

export const Tabs = { Root, List, Tab, Panel }
