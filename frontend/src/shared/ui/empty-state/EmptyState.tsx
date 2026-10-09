import type { ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'
import styles from './EmptyState.module.css'

export type EmptyStateSize = 'regular' | 'compact'

interface RootProps {
  label?: string
  size?: EmptyStateSize
  className?: string
  children: ReactNode
}

function Root({ label, size = 'regular', className, children }: RootProps) {
  return (
    <section aria-label={label} className={cn(styles.root, styles[size], className)}>
      {children}
    </section>
  )
}

interface ArtProps {
  children: ReactNode
}

function Art({ children }: ArtProps) {
  return <div className={styles.art}>{children}</div>
}

interface TitleProps {
  children: ReactNode
}

function Title({ children }: TitleProps) {
  return <h2 className={styles.title}>{children}</h2>
}

interface DescriptionProps {
  children: ReactNode
}

function Description({ children }: DescriptionProps) {
  return <p className={styles.description}>{children}</p>
}

interface ActionsProps {
  children: ReactNode
}

function Actions({ children }: ActionsProps) {
  return <div className={styles.actions}>{children}</div>
}

export const EmptyState = { Root, Art, Title, Description, Actions }
