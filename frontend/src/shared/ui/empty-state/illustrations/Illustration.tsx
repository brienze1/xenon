import type { ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'
import styles from './Illustration.module.css'
import { Spray } from './Spray'

export type IllustrationTone = 'green' | 'cyan' | 'magenta' | 'yellow' | 'red'

interface CanvasProps {
  children: ReactNode
}

function Canvas({ children }: CanvasProps) {
  return (
    <svg className={styles.art} viewBox="0 0 120 96" aria-hidden="true">
      {children}
      <Spray />
    </svg>
  )
}

interface StrokeProps {
  tone: IllustrationTone
  d: string
}

function Stroke({ tone, d }: StrokeProps) {
  return <path className={styles[tone]} d={d} />
}

interface DripProps {
  tone: IllustrationTone
  d: string
}

function Drip({ tone, d }: DripProps) {
  return <path className={cn(styles[tone], styles.drip)} d={d} />
}

interface RingProps {
  tone: IllustrationTone
  cx: number
  cy: number
  r: number
}

function Ring({ tone, cx, cy, r }: RingProps) {
  return <circle className={styles[tone]} cx={cx} cy={cy} r={r} />
}

export const Illustration = { Canvas, Stroke, Drip, Ring }
