import { createContext, use, useMemo, useRef, useState, type CSSProperties, type KeyboardEvent, type PointerEvent as ReactPointerEvent, type ReactNode, type RefObject } from 'react'
import { cn } from '@/shared/lib/cn'
import { suspendTextSelection } from '@/shared/lib/text-selection'
import styles from './SplitPane.module.css'

export type SplitDirection = 'row' | 'column'

const KEYBOARD_STEP = 0.05

interface SplitPaneState {
  ratio: number
}

interface SplitPaneActions {
  changeRatio: (ratio: number) => void
}

interface SplitPaneMeta {
  direction: SplitDirection
  min: number
  max: number
  rootRef: RefObject<HTMLDivElement | null>
}

interface SplitPaneContextValue {
  state: SplitPaneState
  actions: SplitPaneActions
  meta: SplitPaneMeta
}

const SplitPaneContext = createContext<SplitPaneContextValue | null>(null)

function useSplitPane(): SplitPaneContextValue {
  const context = use(SplitPaneContext)
  if (!context) {
    throw new Error('SplitPane parts must be rendered inside <SplitPane.Root>')
  }
  return context
}

export function clampRatio(ratio: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, ratio))
}

interface ProviderProps extends SplitPaneContextValue {
  children: ReactNode
}

function Provider({ state, actions, meta, children }: ProviderProps) {
  const value = useMemo(() => ({ state, actions, meta }), [state, actions, meta])
  return <SplitPaneContext value={value}>{children}</SplitPaneContext>
}

interface RootProps {
  direction: SplitDirection
  ratio: number
  onRatioChange: (ratio: number) => void
  min?: number
  max?: number
  className?: string
  children: ReactNode
}

function Root({ direction, ratio, onRatioChange, min = 0.2, max = 0.8, className, children }: RootProps) {
  const rootRef = useRef<HTMLDivElement>(null)
  const value = clampRatio(ratio, min, max)
  const state = useMemo(() => ({ ratio: value }), [value])
  const actions = useMemo(() => ({ changeRatio: (next: number) => onRatioChange(clampRatio(next, min, max)) }), [onRatioChange, min, max])
  const meta = useMemo(() => ({ direction, min, max, rootRef }), [direction, min, max])
  return (
    <Provider state={state} actions={actions} meta={meta}>
      <div
        ref={rootRef}
        className={cn(styles.root, styles[direction], className)}
        style={{ '--split-ratio': value } as CSSProperties}
      >
        {children}
      </div>
    </Provider>
  )
}

interface PaneProps {
  className?: string
  children: ReactNode
}

function Pane({ className, children }: PaneProps) {
  return <div className={cn(styles.pane, className)}>{children}</div>
}

interface DividerProps {
  label: string
}

function Divider({ label }: DividerProps) {
  const { state, actions, meta } = useSplitPane()
  const [dragging, setDragging] = useState(false)

  function ratioAt(event: PointerEvent): number | null {
    const rect = meta.rootRef.current?.getBoundingClientRect()
    if (!rect) {
      return null
    }
    return meta.direction === 'row'
      ? (event.clientX - rect.left) / rect.width
      : (event.clientY - rect.top) / rect.height
  }

  function onPointerDown(event: ReactPointerEvent<HTMLDivElement>) {
    event.preventDefault()
    const restoreSelection = suspendTextSelection()
    setDragging(true)
    function onPointerMove(move: PointerEvent) {
      const next = ratioAt(move)
      if (next !== null && Number.isFinite(next)) {
        actions.changeRatio(next)
      }
    }
    function onPointerUp() {
      restoreSelection()
      setDragging(false)
      window.removeEventListener('pointermove', onPointerMove)
      window.removeEventListener('pointerup', onPointerUp)
    }
    window.addEventListener('pointermove', onPointerMove)
    window.addEventListener('pointerup', onPointerUp)
  }

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const targets: Record<string, number> = {
      ArrowLeft: state.ratio - KEYBOARD_STEP,
      ArrowUp: state.ratio - KEYBOARD_STEP,
      ArrowRight: state.ratio + KEYBOARD_STEP,
      ArrowDown: state.ratio + KEYBOARD_STEP,
      Home: meta.min,
      End: meta.max,
    }
    const target = targets[event.key]
    if (target === undefined) {
      return
    }
    event.preventDefault()
    actions.changeRatio(Math.round(target * 100) / 100)
  }

  return (
    <div
      role="separator"
      tabIndex={0}
      aria-label={label}
      aria-orientation={meta.direction === 'row' ? 'vertical' : 'horizontal'}
      aria-valuemin={Math.round(meta.min * 100)}
      aria-valuemax={Math.round(meta.max * 100)}
      aria-valuenow={Math.round(state.ratio * 100)}
      className={cn(styles.divider, dragging && styles.dragging)}
      onPointerDown={onPointerDown}
      onKeyDown={onKeyDown}
    />
  )
}

export const SplitPane = { Provider, Root, Pane, Divider }
