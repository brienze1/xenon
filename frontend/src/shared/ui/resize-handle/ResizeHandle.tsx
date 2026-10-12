import { useState, type KeyboardEvent, type PointerEvent as ReactPointerEvent } from 'react'
import { cn } from '@/shared/lib/cn'
import { toPagePixels } from '@/shared/lib/page-zoom'
import { suspendTextSelection } from '@/shared/lib/text-selection'
import styles from './ResizeHandle.module.css'

export const RESIZE_KEYBOARD_STEP = 16

export type ResizeOrientation = 'vertical' | 'horizontal'
export type ResizeEdge = 'start' | 'end'

interface ResizeHandleProps {
  label: string
  orientation: ResizeOrientation
  edge: ResizeEdge
  value: number
  min: number
  max: number
  onChange: (value: number) => void
  className?: string
}

function clamp(value: number, min: number, max: number): number {
  return Math.round(Math.min(max, Math.max(min, value)))
}

export function ResizeHandle({ label, orientation, edge, value, min, max, onChange, className }: ResizeHandleProps) {
  const [dragging, setDragging] = useState(false)
  const growth = edge === 'end' ? 1 : -1

  function position(event: { clientX: number; clientY: number }): number {
    return orientation === 'vertical' ? event.clientX : event.clientY
  }

  function onPointerDown(event: ReactPointerEvent<HTMLDivElement>) {
    event.preventDefault()
    const origin = position(event)
    const initial = value
    const restoreSelection = suspendTextSelection()
    setDragging(true)
    function onPointerMove(move: PointerEvent) {
      onChange(clamp(initial + toPagePixels(position(move) - origin) * growth, min, max))
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
    const forward = orientation === 'vertical' ? 'ArrowRight' : 'ArrowDown'
    const backward = orientation === 'vertical' ? 'ArrowLeft' : 'ArrowUp'
    const targets: Record<string, number> = {
      [forward]: value + RESIZE_KEYBOARD_STEP * growth,
      [backward]: value - RESIZE_KEYBOARD_STEP * growth,
      Home: min,
      End: max,
    }
    const target = targets[event.key]
    if (target === undefined) {
      return
    }
    event.preventDefault()
    onChange(clamp(target, min, max))
  }

  return (
    <div
      role="separator"
      tabIndex={0}
      aria-label={label}
      aria-orientation={orientation}
      aria-valuemin={min}
      aria-valuemax={max}
      aria-valuenow={Math.round(value)}
      className={cn(styles.handle, styles[orientation], styles[edge], dragging && styles.dragging, className)}
      onPointerDown={onPointerDown}
      onKeyDown={onKeyDown}
    />
  )
}
