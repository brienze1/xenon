import type { ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'
import styles from './PanelSection.module.css'

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

interface HeadingProps {
  children: ReactNode
}

function Heading({ children }: HeadingProps) {
  return <h3 className={styles.heading}>{children}</h3>
}

interface NoteProps {
  children: ReactNode
}

function Note({ children }: NoteProps) {
  return <p className={styles.note}>{children}</p>
}

export const PanelSection = { Root, Heading, Note }
