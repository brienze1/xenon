import { cn } from '@/shared/lib/cn'
import styles from './MethodBadge.module.css'

const STANDARD_METHODS = ['get', 'post', 'put', 'patch', 'delete', 'head', 'options'] as const

type StandardMethod = (typeof STANDARD_METHODS)[number]

export function methodTone(method: string): StandardMethod | 'custom' {
  const normalized = method.trim().toLowerCase()
  return (STANDARD_METHODS as readonly string[]).includes(normalized) ? (normalized as StandardMethod) : 'custom'
}

interface MethodBadgeProps {
  method: string
  className?: string
}

export function MethodBadge({ method, className }: MethodBadgeProps) {
  const tone = methodTone(method)
  return (
    <span className={cn(styles.badge, tone !== 'custom' && styles[tone], className)} data-tone={tone}>
      {method.trim().toUpperCase()}
    </span>
  )
}
