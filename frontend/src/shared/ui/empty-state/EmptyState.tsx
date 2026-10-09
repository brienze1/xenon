import type { ComponentType, ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'
import styles from './EmptyState.module.css'
import { CollectionsArt } from './illustrations/CollectionsArt'
import { ConsoleArt } from './illustrations/ConsoleArt'
import { HistoryArt } from './illustrations/HistoryArt'
import { MissingArt } from './illustrations/MissingArt'
import { ResponseArt } from './illustrations/ResponseArt'
import { TabsArt } from './illustrations/TabsArt'
import { WorkspaceArt } from './illustrations/WorkspaceArt'

export type EmptyStateArt = 'workspace' | 'collections' | 'tabs' | 'response' | 'history' | 'console' | 'missing'

const ART: Record<EmptyStateArt, ComponentType> = {
  workspace: WorkspaceArt,
  collections: CollectionsArt,
  tabs: TabsArt,
  response: ResponseArt,
  history: HistoryArt,
  console: ConsoleArt,
  missing: MissingArt,
}

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

function Art({ variant }: { variant: EmptyStateArt }) {
  const Illustration = ART[variant]
  return (
    <div className={styles.art} data-art={variant}>
      <Illustration />
    </div>
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
