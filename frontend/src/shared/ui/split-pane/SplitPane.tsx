import { createContext, use, useRef, useState, type CSSProperties, type KeyboardEvent, type PointerEvent as ReactPointerEvent, type ReactNode, type RefObject } from 'react'
import { cn } from '@/shared/lib/cn'
import styles from './SplitPane.module.css'

export type SplitDirection = 'row' | 'column'

const KEYBOARD_STEP = 0.05

interface SplitPaneContextValue {
  direction: SplitDirection
  ratio: number
  min: number
  max: number
  rootRef: RefObject<HTMLDivElement | null>
  changeRatio: (ratio: number) => void
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
  const changeRatio = (next: number) => onRatioChange(clampRatio(next, min, max))
  return (
    <SplitPaneContext value={{ direction, ratio: value, min, max, rootRef, changeRatio }}>
      <div
        ref={rootRef}
        className={cn(styles.root, styles[direction], className)}
        style={{ '--split-ratio': value } as CSSProperties}
      >
        {children}
      </div>
    </SplitPaneContext>
  )
}

function Pane({ className, children }: { className?: string; children: ReactNode }) {
  return <div className={cn(styles.pane, className)}>{children}</div>
}

function Divider({ label }: { label: string }) {
  const split = useSplitPane()
  const [dragging, setDragging] = useState(false)

  function ratioAt(event: PointerEvent): number | null {
    const rect = split.rootRef.current?.getBoundingClientRect()
    if (!rect) {
      return null
    }
    return split.direction === 'row'
      ? (event.clientX - rect.left) / rect.width
      : (event.clientY - rect.top) / rect.height
  }

  function onPointerDown(event: ReactPointerEvent<HTMLDivElement>) {
    event.preventDefault()
    setDragging(true)
    function onPointerMove(move: PointerEvent) {
      const next = ratioAt(move)
      if (next !== null && Number.isFinite(next)) {
        split.changeRatio(next)
      }
    }
    function onPointerUp() {
      setDragging(false)
      window.removeEventListener('pointermove', onPointerMove)
      window.removeEventListener('pointerup', onPointerUp)
    }
    window.addEventListener('pointermove', onPointerMove)
    window.addEventListener('pointerup', onPointerUp)
  }

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const targets: Record<string, number> = {
      ArrowLeft: split.ratio - KEYBOARD_STEP,
      ArrowUp: split.ratio - KEYBOARD_STEP,
      ArrowRight: split.ratio + KEYBOARD_STEP,
      ArrowDown: split.ratio + KEYBOARD_STEP,
      Home: split.min,
      End: split.max,
    }
    const target = targets[event.key]
    if (target === undefined) {
      return
    }
    event.preventDefault()
    split.changeRatio(Math.round(target * 100) / 100)
  }

  return (
    <div
      role="separator"
      tabIndex={0}
      aria-label={label}
      aria-orientation={split.direction === 'row' ? 'vertical' : 'horizontal'}
      aria-valuemin={Math.round(split.min * 100)}
      aria-valuemax={Math.round(split.max * 100)}
      aria-valuenow={Math.round(split.ratio * 100)}
      className={cn(styles.divider, dragging && styles.dragging)}
      onPointerDown={onPointerDown}
      onKeyDown={onKeyDown}
    />
  )
}

export const SplitPane = { Root, Pane, Divider }
