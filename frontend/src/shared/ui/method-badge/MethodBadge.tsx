import { cn } from '@/shared/lib/cn'
import styles from './MethodBadge.module.css'

type StandardMethod = 'get' | 'post' | 'put' | 'patch' | 'delete' | 'head' | 'options'

const STANDARD_METHODS: ReadonlySet<string> = new Set<StandardMethod>(['get', 'post', 'put', 'patch', 'delete', 'head', 'options'])

export function methodTone(method: string): StandardMethod | 'custom' {
  const normalized = method.trim().toLowerCase()
  return STANDARD_METHODS.has(normalized) ? (normalized as StandardMethod) : 'custom'
}

interface MethodBadgeProps {
  method: string
  className?: string
}

export function MethodBadge({ method, className }: MethodBadgeProps) {
  const tone = methodTone(method)
  return (
    <span className={cn(styles.badge, tone !== 'custom' && styles[tone], className)} translate="no">
      {method.trim().toUpperCase()}
    </span>
  )
}
