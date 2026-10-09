import type { ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'
import styles from './EmptyState.module.css'

interface RootProps {
  label?: string
  className?: string
  children: ReactNode
}

function Root({ label, className, children }: RootProps) {
  return (
    <section aria-label={label} className={cn(styles.root, className)}>
      {children}
    </section>
  )
}

function Art() {
  return (
    <svg className={styles.art} viewBox="0 0 96 96" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" aria-hidden="true">
      <path d="M14 70c10-24 22-40 34-44s22 6 30 20" />
      <path d="M24 30l-6-12M48 22V8M72 30l8-11" />
      <circle cx="48" cy="58" r="10" />
      <path d="M38 82c6-4 14-4 20 0" />
    </svg>
  )
}

function Title({ children }: { children: ReactNode }) {
  return <h1 className={styles.title}>{children}</h1>
}

function Description({ children }: { children: ReactNode }) {
  return <p className={styles.description}>{children}</p>
}

function Actions({ children }: { children: ReactNode }) {
  return <div className={styles.actions}>{children}</div>
}

export const EmptyState = { Root, Art, Title, Description, Actions }
