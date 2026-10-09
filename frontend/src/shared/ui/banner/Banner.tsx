import type { ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'
import styles from './Banner.module.css'

export type BannerTone = 'info' | 'warning' | 'error'

interface RootProps {
  tone: BannerTone
  className?: string
  children: ReactNode
}

function Root({ tone, className, children }: RootProps) {
  return (
    <div role={tone === 'error' ? 'alert' : 'status'} className={cn(styles.root, tone !== 'info' && styles[tone], className)}>
      {children}
    </div>
  )
}

function Message({ children }: { children: ReactNode }) {
  return <div className={styles.message}>{children}</div>
}

function Actions({ children }: { children: ReactNode }) {
  return <div className={styles.actions}>{children}</div>
}

export const Banner = { Root, Message, Actions }
