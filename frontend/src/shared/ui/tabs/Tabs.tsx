import { Activity, createContext, use, useId, useMemo, type KeyboardEvent, type ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'
import { rovingIndex } from '@/shared/lib/roving-focus'
import styles from './Tabs.module.css'

interface TabsState {
  value: string
}

interface TabsActions {
  select: (value: string) => void
}

interface TabsMeta {
  baseId: string
}

interface TabsContextValue {
  state: TabsState
  actions: TabsActions
  meta: TabsMeta
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

interface ProviderProps extends TabsContextValue {
  children: ReactNode
}

function Provider({ state, actions, meta, children }: ProviderProps) {
  const value = useMemo(() => ({ state, actions, meta }), [state, actions, meta])
  return <TabsContext value={value}>{children}</TabsContext>
}

interface RootProps {
  value: string
  onValueChange: (value: string) => void
  className?: string
  children: ReactNode
}

function Root({ value, onValueChange, className, children }: RootProps) {
  const baseId = useId()
  const state = useMemo(() => ({ value }), [value])
  const actions = useMemo(() => ({ select: onValueChange }), [onValueChange])
  const meta = useMemo(() => ({ baseId }), [baseId])
  return (
    <Provider state={state} actions={actions} meta={meta}>
      <div className={cn(styles.root, className)}>{children}</div>
    </Provider>
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
  const { state, actions, meta } = useTabs()
  const selected = state.value === value
  return (
    <button
      type="button"
      role="tab"
      id={tabId(meta.baseId, value)}
      aria-selected={selected}
      aria-controls={panelId(meta.baseId, value)}
      tabIndex={selected ? 0 : -1}
      disabled={disabled}
      className={cn(styles.tab, className)}
      onClick={() => actions.select(value)}
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
  const { state, meta } = useTabs()
  const selected = state.value === value
  // Activity keeps a hidden panel's state and DOM (display: none, out of the accessibility
  // tree) so switching back is instant; `hidden` states the same thing in the markup.
  return (
    <Activity mode={selected ? 'visible' : 'hidden'}>
      <div
        role="tabpanel"
        id={panelId(meta.baseId, value)}
        aria-labelledby={tabId(meta.baseId, value)}
        hidden={!selected}
        className={cn(styles.panel, className)}
      >
        {children}
      </div>
    </Activity>
  )
}

export const Tabs = { Provider, Root, List, Tab, Panel }
