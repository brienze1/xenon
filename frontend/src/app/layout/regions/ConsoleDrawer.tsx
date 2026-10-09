import type { ReactNode } from 'react'
import styles from './ConsoleDrawer.module.css'

interface ConsoleDrawerProps {
  open: boolean
  children?: ReactNode
}

export function ConsoleDrawer({ open, children }: ConsoleDrawerProps) {
  return (
    <section aria-label="Console" hidden={!open} className={styles.drawer}>
      {children}
    </section>
  )
}
